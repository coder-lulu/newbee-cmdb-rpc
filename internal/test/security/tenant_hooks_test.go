package security

import (
	"context"
	"testing"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/enttest"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestClient creates a test client with security hooks enabled
func setupTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Register tenant hooks - CRITICAL for security
	client.Use(hooks.TenantMutationHook())
	client.Intercept(hooks.TenantQueryInterceptor())
	
	// Register data permission interceptors
	hooks.RegisterDataPermissionInterceptorsWithTenant(client,
		"cis", "ci_types", "ci_permissions", "import_tasks", "import_records")
	
	return client
}

// TestTenantMutationHook tests that tenant hooks properly inject tenant_id
func TestTenantMutationHook(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	testCases := []struct {
		name     string
		tenantID uint64
		setup    func(ctx context.Context) error
		validate func(ctx context.Context) error
	}{
		{
			name:     "Create CI Type with Tenant Context",
			tenantID: 1,
			setup: func(ctx context.Context) error {
				_, err := client.CiType.Create().
					SetName("Server").
					SetAlias("server").
					SetIcon("server-icon").
					Save(ctx)
				return err
			},
			validate: func(ctx context.Context) error {
				ciTypes, err := client.CiType.Query().All(ctx)
				if err != nil {
					return err
				}
				
				assert.Len(t, ciTypes, 1, "Should have exactly one CI type")
				assert.Equal(t, uint64(1), ciTypes[0].TenantID, "CI type should have correct tenant ID")
				assert.Equal(t, "Server", ciTypes[0].Name, "CI type should have correct name")
				return nil
			},
		},
		{
			name:     "Create CI with Tenant Context",
			tenantID: 2,
			setup: func(ctx context.Context) error {
				// First create CI type
				ciType, err := client.CiType.Create().
					SetName("Application").
					SetAlias("app").
					SetIcon("app-icon").
					Save(ctx)
				if err != nil {
					return err
				}
				
				// Then create CI
				_, err = client.Cis.Create().
					SetTypeID(ciType.ID).
					SetStatus(1).
					SetCreatedBy(uuid.Must(uuid.NewV4())).
					Save(ctx)
				return err
			},
			validate: func(ctx context.Context) error {
				cis, err := client.Cis.Query().All(ctx)
				if err != nil {
					return err
				}
				
				assert.Len(t, cis, 1, "Should have exactly one CI")
				assert.Equal(t, uint64(2), cis[0].TenantID, "CI should have correct tenant ID")
				return nil
			},
		},
		{
			name:     "Update Operation with Tenant Context",
			tenantID: 1,
			setup: func(ctx context.Context) error {
				// Create CI type first
				ciType, err := client.CiType.Create().
					SetName("Database").
					SetAlias("db").
					SetIcon("db-icon").
					Save(ctx)
				if err != nil {
					return err
				}
				
				// Update the CI type
				return client.CiType.UpdateOneID(ciType.ID).
					SetDescription("Updated database description").
					Exec(ctx)
			},
			validate: func(ctx context.Context) error {
				ciTypes, err := client.CiType.Query().All(ctx)
				if err != nil {
					return err
				}
				
				assert.Len(t, ciTypes, 1, "Should have exactly one CI type")
				assert.Equal(t, uint64(1), ciTypes[0].TenantID, "CI type should maintain correct tenant ID after update")
				assert.Equal(t, "Updated database description", *ciTypes[0].Description, "Description should be updated")
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create tenant context
			ctx := context.WithValue(context.Background(), "tenantId", tc.tenantID)
			
			// Setup data
			err := tc.setup(ctx)
			require.NoError(t, err, "Setup should succeed")
			
			// Validate results
			err = tc.validate(ctx)
			require.NoError(t, err, "Validation should succeed")
			
			// Cleanup - use system context to bypass tenant isolation
			systemCtx := hooks.NewSystemContext(context.Background())
			_, err = client.CiType.Delete().Exec(systemCtx)
			require.NoError(t, err, "Cleanup should succeed")
			_, err = client.Cis.Delete().Exec(systemCtx)
			require.NoError(t, err, "Cleanup should succeed")
		})
	}
}

// TestTenantQueryInterceptor tests that query interceptor properly filters by tenant
func TestTenantQueryInterceptor(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup test data using system context
	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Create CI types for different tenants
	ciType1, err := client.CiType.Create().
		SetName("Server").
		SetAlias("server").
		SetIcon("server-icon").
		SetTenantID(1).
		Save(systemCtx)
	require.NoError(t, err)
	
	ciType2, err := client.CiType.Create().
		SetName("Application").
		SetAlias("app").
		SetIcon("app-icon").
		SetTenantID(2).
		Save(systemCtx)
	require.NoError(t, err)
	
	ciType3, err := client.CiType.Create().
		SetName("Database").
		SetAlias("db").
		SetIcon("db-icon").
		SetTenantID(1).
		Save(systemCtx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		tenantID       uint64
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:          "Tenant 1 should see 2 CI types",
			tenantID:      1,
			expectedCount: 2,
			expectedNames: []string{"Server", "Database"},
		},
		{
			name:          "Tenant 2 should see 1 CI type",
			tenantID:      2,
			expectedCount: 1,
			expectedNames: []string{"Application"},
		},
		{
			name:          "Tenant 3 should see 0 CI types",
			tenantID:      3,
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create tenant context
			ctx := context.WithValue(context.Background(), "tenantId", tc.tenantID)
			
			// Query CI types
			ciTypes, err := client.CiType.Query().All(ctx)
			require.NoError(t, err, "Query should succeed")
			
			// Verify count
			assert.Equal(t, tc.expectedCount, len(ciTypes), 
				"Should return correct number of CI types for tenant %d", tc.tenantID)
			
			// Verify names
			if tc.expectedCount > 0 {
				actualNames := make([]string, len(ciTypes))
				for i, ct := range ciTypes {
					actualNames[i] = ct.Name
					// Verify tenant ID is correct
					assert.Equal(t, tc.tenantID, ct.TenantID, 
						"All returned CI types should belong to tenant %d", tc.tenantID)
				}
				
				assert.ElementsMatch(t, tc.expectedNames, actualNames, 
					"Should return correct CI type names")
			}
		})
	}
}

// TestTenantIsolationInRelations tests tenant isolation in entity relationships
func TestTenantIsolationInRelations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup test data using system context
	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Create CI types for different tenants
	ciType1, err := client.CiType.Create().
		SetName("Server").
		SetAlias("server").
		SetIcon("server-icon").
		SetTenantID(1).
		Save(systemCtx)
	require.NoError(t, err)
	
	ciType2, err := client.CiType.Create().
		SetName("Application").
		SetAlias("app").
		SetIcon("app-icon").
		SetTenantID(2).
		Save(systemCtx)
	require.NoError(t, err)
	
	// Create CIs for different tenants
	ci1, err := client.Cis.Create().
		SetTypeID(ciType1.ID).
		SetStatus(1).
		SetCreatedBy(uuid.Must(uuid.NewV4())).
		SetTenantID(1).
		SetDepartmentID(1).
		Save(systemCtx)
	require.NoError(t, err)
	
	ci2, err := client.Cis.Create().
		SetTypeID(ciType2.ID).
		SetStatus(1).
		SetCreatedBy(uuid.Must(uuid.NewV4())).
		SetTenantID(2).
		SetDepartmentID(2).
		Save(systemCtx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		tenantID       uint64
		queryCI        func(ctx context.Context) (*ent.Cis, error)
		shouldSucceed  bool
		expectedTenant uint64
	}{
		{
			name:     "Query CI from same tenant via relation",
			tenantID: 1,
			queryCI: func(ctx context.Context) (*ent.Cis, error) {
				return client.CiType.Query().
					Where(citype.ID(ciType1.ID)).
					QueryCis().
					First(ctx)
			},
			shouldSucceed:  true,
			expectedTenant: 1,
		},
		{
			name:     "Query CI from different tenant should fail",
			tenantID: 1,
			queryCI: func(ctx context.Context) (*ent.Cis, error) {
				return client.CiType.Query().
					Where(citype.ID(ciType2.ID)).
					QueryCis().
					First(ctx)
			},
			shouldSucceed:  false,
			expectedTenant: 0,
		},
		{
			name:     "Query CI Type from CI relation",
			tenantID: 2,
			queryCI: func(ctx context.Context) (*ent.Cis, error) {
				return client.Cis.Query().
					Where(cis.ID(ci2.ID)).
					First(ctx)
			},
			shouldSucceed:  true,
			expectedTenant: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create tenant context
			ctx := context.WithValue(context.Background(), "tenantId", tc.tenantID)
			
			// Execute query
			ci, err := tc.queryCI(ctx)
			
			if tc.shouldSucceed {
				require.NoError(t, err, "Query should succeed for same tenant")
				require.NotNil(t, ci, "CI should not be nil")
				assert.Equal(t, tc.expectedTenant, ci.TenantID, 
					"Returned CI should have correct tenant ID")
			} else {
				// Query should either fail or return no results
				if err == nil {
					assert.Nil(t, ci, "Query should return no results for different tenant")
				} else {
					// Error is acceptable when accessing cross-tenant data
					t.Logf("Query failed as expected: %v", err)
				}
			}
		})
	}
}

// TestSystemContextBypassesTenantIsolation tests that system context can access all data
func TestSystemContextBypassesTenantIsolation(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup test data using system context
	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Create CI types for different tenants
	ciTypes := []*ent.CiType{}
	tenantIDs := []uint64{1, 2, 3}
	
	for _, tenantID := range tenantIDs {
		ciType, err := client.CiType.Create().
			SetName(fmt.Sprintf("Type-T%d", tenantID)).
			SetAlias(fmt.Sprintf("type-t%d", tenantID)).
			SetIcon("icon").
			SetTenantID(tenantID).
			Save(systemCtx)
		require.NoError(t, err)
		ciTypes = append(ciTypes, ciType)
	}

	t.Run("System context can see all tenants data", func(t *testing.T) {
		// Query with system context should return all CI types
		allCiTypes, err := client.CiType.Query().All(systemCtx)
		require.NoError(t, err, "System context query should succeed")
		
		assert.Equal(t, len(tenantIDs), len(allCiTypes), 
			"System context should see all CI types across tenants")
		
		// Verify all tenant IDs are present
		tenantIDsFound := make(map[uint64]bool)
		for _, ct := range allCiTypes {
			tenantIDsFound[ct.TenantID] = true
		}
		
		for _, tenantID := range tenantIDs {
			assert.True(t, tenantIDsFound[tenantID], 
				"System context should see data from tenant %d", tenantID)
		}
	})

	t.Run("Regular tenant context sees only own data", func(t *testing.T) {
		// Query with tenant context should return only tenant's data
		ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		tenant1CiTypes, err := client.CiType.Query().All(ctx)
		require.NoError(t, err, "Tenant context query should succeed")
		
		assert.Equal(t, 1, len(tenant1CiTypes), 
			"Tenant context should see only own CI type")
		assert.Equal(t, uint64(1), tenant1CiTypes[0].TenantID, 
			"CI type should belong to correct tenant")
	})

	t.Run("System context can perform cross-tenant operations", func(t *testing.T) {
		// System context should be able to update data across tenants
		err := client.CiType.Update().
			Where(citype.TenantID(2)).
			SetDescription("Updated by system").
			Exec(systemCtx)
		require.NoError(t, err, "System context should be able to update cross-tenant data")
		
		// Verify the update with system context
		updatedType, err := client.CiType.Query().
			Where(citype.TenantID(2)).
			First(systemCtx)
		require.NoError(t, err)
		assert.Equal(t, "Updated by system", *updatedType.Description)
		
		// Verify tenant context cannot see the update from different tenant
		ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		_, err = client.CiType.Query().
			Where(citype.TenantID(2)).
			First(ctx)
		// This should fail or return no results
		if err == nil {
			t.Error("Tenant context should not be able to see cross-tenant data")
		}
	})
}

// TestTenantHookErrorHandling tests error handling in tenant hooks
func TestTenantHookErrorHandling(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	testCases := []struct {
		name        string
		ctx         context.Context
		operation   func(ctx context.Context) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "Missing tenant context should fail",
			ctx:  context.Background(), // No tenant ID
			operation: func(ctx context.Context) error {
				_, err := client.CiType.Create().
					SetName("Test Type").
					SetAlias("test").
					SetIcon("icon").
					Save(ctx)
				return err
			},
			expectError: true,
			errorMsg:    "tenant",
		},
		{
			name: "Zero tenant ID should fail",
			ctx:  context.WithValue(context.Background(), "tenantId", uint64(0)),
			operation: func(ctx context.Context) error {
				_, err := client.CiType.Create().
					SetName("Test Type").
					SetAlias("test").
					SetIcon("icon").
					Save(ctx)
				return err
			},
			expectError: true,
			errorMsg:    "tenant",
		},
		{
			name: "Valid tenant context should succeed",
			ctx:  context.WithValue(context.Background(), "tenantId", uint64(1)),
			operation: func(ctx context.Context) error {
				_, err := client.CiType.Create().
					SetName("Valid Type").
					SetAlias("valid").
					SetIcon("icon").
					Save(ctx)
				return err
			},
			expectError: false,
			errorMsg:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation(tc.ctx)
			
			if tc.expectError {
				assert.Error(t, err, "Operation should fail")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, 
						"Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Operation should succeed")
				
				// Cleanup
				systemCtx := hooks.NewSystemContext(context.Background())
				client.CiType.Delete().Exec(systemCtx)
			}
		})
	}
}