package security

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

// setupDatabaseSecurityTestClient creates test client with all security features
func setupDatabaseSecurityTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Enable all security features
	client.Use(hooks.TenantMutationHook())
	client.Intercept(hooks.TenantQueryInterceptor())
	hooks.RegisterDataPermissionInterceptorsWithTenant(client,
		"cis", "ci_types", "ci_permissions", "import_tasks", "import_records")
	
	return client
}

// TestDatabaseTenantIsolationAtSchemaLevel tests tenant isolation at database schema level
func TestDatabaseTenantIsolationAtSchemaLevel(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())

	t.Run("Tenant ID is automatically set on creation", func(t *testing.T) {
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
		
		ciType, err := client.CiType.Create().
			SetName("Test Type").
			SetAlias("test").
			SetIcon("icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		// Verify tenant ID was automatically set
		assert.Equal(t, uint64(1), ciType.TenantID, 
			"Tenant ID should be automatically set by hook")
	})

	t.Run("Queries are automatically filtered by tenant", func(t *testing.T) {
		// Create data for multiple tenants using system context
		for tenantID := uint64(1); tenantID <= 3; tenantID++ {
			_, err := client.CiType.Create().
				SetName(fmt.Sprintf("Type T%d", tenantID)).
				SetAlias(fmt.Sprintf("type_t%d", tenantID)).
				SetIcon("icon").
				SetTenantID(tenantID).
				Save(systemCtx)
			require.NoError(t, err)
		}
		
		// Query with tenant context should only return tenant's data
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(2))
		ciTypes, err := client.CiType.Query().All(tenantCtx)
		require.NoError(t, err)
		
		assert.Len(t, ciTypes, 1, "Should return only tenant 2's data")
		assert.Equal(t, uint64(2), ciTypes[0].TenantID)
		assert.Equal(t, "Type T2", ciTypes[0].Name)
	})

	t.Run("Updates respect tenant isolation", func(t *testing.T) {
		// Create data for different tenants
		ciType1, err := client.CiType.Create().
			SetName("Tenant 1 Type").
			SetAlias("t1_type").
			SetIcon("icon").
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)

		ciType2, err := client.CiType.Create().
			SetName("Tenant 2 Type").
			SetAlias("t2_type").
			SetIcon("icon").
			SetTenantID(2).
			Save(systemCtx)
		require.NoError(t, err)

		// Try to update tenant 2's data from tenant 1 context
		tenant1Ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		
		err = client.CiType.UpdateOneID(ciType2.ID).
			SetDescription("Attempted update").
			Exec(tenant1Ctx)
		
		// Should fail or have no effect
		assert.Error(t, err, "Cross-tenant update should be blocked")

		// Verify tenant 2's data wasn't modified
		unchanged, err := client.CiType.GetX(systemCtx, ciType2.ID)
		require.NoError(t, err)
		assert.Nil(t, unchanged.Description, "Description should remain unchanged")

		// But updating own tenant's data should work
		err = client.CiType.UpdateOneID(ciType1.ID).
			SetDescription("Valid update").
			Exec(tenant1Ctx)
		assert.NoError(t, err, "Same-tenant update should succeed")
	})

	t.Run("Deletes respect tenant isolation", func(t *testing.T) {
		// Create test data for different tenants
		ciType1, err := client.CiType.Create().
			SetName("Delete Test T1").
			SetAlias("delete_t1").
			SetIcon("icon").
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)

		ciType2, err := client.CiType.Create().
			SetName("Delete Test T2").
			SetAlias("delete_t2").
			SetIcon("icon").
			SetTenantID(2).
			Save(systemCtx)
		require.NoError(t, err)

		// Try to delete tenant 2's data from tenant 1 context
		tenant1Ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		
		err = client.CiType.DeleteOneID(ciType2.ID).Exec(tenant1Ctx)
		assert.Error(t, err, "Cross-tenant delete should be blocked")

		// Verify tenant 2's data still exists
		exists, err := client.CiType.Query().
			Where(citype.ID(ciType2.ID)).
			Exist(systemCtx)
		require.NoError(t, err)
		assert.True(t, exists, "Tenant 2 data should still exist")

		// But deleting own tenant's data should work
		err = client.CiType.DeleteOneID(ciType1.ID).Exec(tenant1Ctx)
		assert.NoError(t, err, "Same-tenant delete should succeed")
	})
}

// TestSoftDeleteSecurity tests that soft delete respects security constraints
func TestSoftDeleteSecurity(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())

	t.Run("Soft deleted records are filtered by tenant", func(t *testing.T) {
		// Create CI type and CI
		ciType, err := client.CiType.Create().
			SetName("Test Type").
			SetAlias("test_type").
			SetIcon("icon").
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)

		ci, err := client.Cis.Create().
			SetTypeID(ciType.ID).
			SetStatus(1).
			SetCreatedBy(uuid.Must(uuid.NewV4())).
			SetTenantID(1).
			SetDepartmentID(1).
			Save(systemCtx)
		require.NoError(t, err)

		// Soft delete the CI using system context
		err = client.Cis.UpdateOneID(ci.ID).
			SetDeletedAt(time.Now()).
			Exec(systemCtx)
		require.NoError(t, err)

		// Query with tenant context should not return soft deleted record
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
		cisCount, err := client.Cis.Query().Count(tenantCtx)
		require.NoError(t, err)
		
		assert.Equal(t, 0, cisCount, "Soft deleted records should be filtered out")

		// But system context should see it
		systemCisCount, err := client.Cis.Query().Count(systemCtx)
		require.NoError(t, err)
		assert.Equal(t, 1, systemCisCount, "System context should see soft deleted records")
	})

	t.Run("Cannot access soft deleted records from other tenants", func(t *testing.T) {
		// Create and soft delete CI for tenant 2
		ciType, err := client.CiType.Create().
			SetName("Tenant 2 Type").
			SetAlias("t2_type").
			SetIcon("icon").
			SetTenantID(2).
			Save(systemCtx)
		require.NoError(t, err)

		ci, err := client.Cis.Create().
			SetTypeID(ciType.ID).
			SetStatus(1).
			SetCreatedBy(uuid.Must(uuid.NewV4())).
			SetTenantID(2).
			SetDepartmentID(2).
			Save(systemCtx)
		require.NoError(t, err)

		// Soft delete
		err = client.Cis.UpdateOneID(ci.ID).
			SetDeletedAt(time.Now()).
			Exec(systemCtx)
		require.NoError(t, err)

		// Tenant 1 context should not be able to access it even if it bypassed soft delete
		tenant1Ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		
		// Even with WithDeleted (if available), should not see other tenant's data
		cisCount, err := client.Cis.Query().Count(tenant1Ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, cisCount, "Should not see other tenant's soft deleted records")
	})
}

// TestDatabaseIndexSecurity tests that database indexes support security features
func TestDatabaseIndexSecurity(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())

	t.Run("Tenant ID indexes are utilized for performance", func(t *testing.T) {
		// Create large dataset across multiple tenants
		const recordsPerTenant = 10
		const numTenants = 5

		// Create CI type first
		ciType, err := client.CiType.Create().
			SetName("Performance Test Type").
			SetAlias("perf_test").
			SetIcon("icon").
			SetTenantID(1). // Just use tenant 1 for the type
			Save(systemCtx)
		require.NoError(t, err)

		// Create CIs for multiple tenants
		for tenantID := uint64(1); tenantID <= numTenants; tenantID++ {
			for i := 0; i < recordsPerTenant; i++ {
				_, err := client.Cis.Create().
					SetTypeID(ciType.ID).
					SetStatus(1).
					SetCreatedBy(uuid.Must(uuid.NewV4())).
					SetTenantID(tenantID).
					SetDepartmentID(tenantID). // Use tenantID as departmentID for simplicity
					Save(systemCtx)
				require.NoError(t, err)
			}
		}

		// Query with tenant context should be fast and return correct results
		start := time.Now()
		
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(3))
		tenant3Cis, err := client.Cis.Query().All(tenantCtx)
		require.NoError(t, err)
		
		queryDuration := time.Since(start)

		// Verify results
		assert.Len(t, tenant3Cis, recordsPerTenant, 
			"Should return correct number of records for tenant 3")
		
		for _, ci := range tenant3Cis {
			assert.Equal(t, uint64(3), ci.TenantID, 
				"All returned CIs should belong to tenant 3")
		}

		// Query should be reasonably fast (adjust threshold as needed)
		assert.Less(t, queryDuration, time.Second, 
			"Tenant-filtered query should be performant")

		t.Logf("Query returned %d records in %v", len(tenant3Cis), queryDuration)
	})

	t.Run("Compound indexes with tenant_id work correctly", func(t *testing.T) {
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))

		// Query with multiple conditions that should use compound indexes
		start := time.Now()
		
		cisWithStatus, err := client.Cis.Query().
			Where(cis.Status(1)).
			All(tenantCtx)
		require.NoError(t, err)

		queryDuration := time.Since(start)

		// All results should belong to tenant 1 and have status 1
		for _, ci := range cisWithStatus {
			assert.Equal(t, uint64(1), ci.TenantID)
			assert.Equal(t, uint32(1), ci.Status)
		}

		t.Logf("Compound index query returned %d records in %v", 
			len(cisWithStatus), queryDuration)
	})
}

// TestDatabaseConstraintSecurity tests that database constraints enforce security
func TestDatabaseConstraintSecurity(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())

	t.Run("Foreign key constraints respect tenant boundaries", func(t *testing.T) {
		// Create CI type in tenant 1
		ciType1, err := client.CiType.Create().
			SetName("Type T1").
			SetAlias("type_t1").
			SetIcon("icon").
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)

		// Try to create CI in tenant 2 referencing CI type from tenant 1
		// This should be prevented by application logic, but let's test
		tenant2Ctx := context.WithValue(context.Background(), "tenantId", uint64(2))
		
		_, err = client.Cis.Create().
			SetTypeID(ciType1.ID). // Cross-tenant reference
			SetStatus(1).
			SetCreatedBy(uuid.Must(uuid.NewV4())).
			Save(tenant2Ctx)
		
		// Should fail due to tenant isolation
		assert.Error(t, err, "Cross-tenant foreign key reference should be blocked")
	})

	t.Run("Unique constraints work within tenant scope", func(t *testing.T) {
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))

		// Create CI type with unique alias within tenant
		_, err := client.CiType.Create().
			SetName("Unique Test").
			SetAlias("unique_test").
			SetIcon("icon").
			Save(tenantCtx)
		require.NoError(t, err)

		// Try to create another CI type with same alias in same tenant
		_, err = client.CiType.Create().
			SetName("Unique Test 2").
			SetAlias("unique_test"). // Duplicate alias
			SetIcon("icon").
			Save(tenantCtx)
		
		// Should fail due to unique constraint within tenant
		assert.Error(t, err, "Duplicate alias within tenant should be blocked")

		// But creating with same alias in different tenant should work
		tenant2Ctx := context.WithValue(context.Background(), "tenantId", uint64(2))
		
		_, err = client.CiType.Create().
			SetName("Unique Test T2").
			SetAlias("unique_test"). // Same alias, different tenant
			SetIcon("icon").
			Save(tenant2Ctx)
		
		assert.NoError(t, err, "Same alias in different tenant should be allowed")
	})
}

// TestDatabaseTransactionSecurity tests transaction-level security
func TestDatabaseTransactionSecurity(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	t.Run("Transactions respect tenant isolation", func(t *testing.T) {
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))

		// Start transaction
		tx, err := client.Tx(tenantCtx)
		require.NoError(t, err)

		// Create CI type in transaction
		ciType, err := tx.CiType.Create().
			SetName("Transaction Test").
			SetAlias("tx_test").
			SetIcon("icon").
			Save(tenantCtx)
		require.NoError(t, err)

		// Create CI in transaction
		_, err = tx.Cis.Create().
			SetTypeID(ciType.ID).
			SetStatus(1).
			SetCreatedBy(uuid.Must(uuid.NewV4())).
			Save(tenantCtx)
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit()
		require.NoError(t, err)

		// Verify data exists and belongs to correct tenant
		ciTypes, err := client.CiType.Query().All(tenantCtx)
		require.NoError(t, err)
		
		found := false
		for _, ct := range ciTypes {
			if ct.Name == "Transaction Test" {
				found = true
				assert.Equal(t, uint64(1), ct.TenantID, 
					"Transaction-created data should have correct tenant ID")
				break
			}
		}
		assert.True(t, found, "Transaction-created CI type should exist")
	})

	t.Run("Transaction rollback respects tenant isolation", func(t *testing.T) {
		tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))

		// Count existing records
		initialCount, err := client.CiType.Query().Count(tenantCtx)
		require.NoError(t, err)

		// Start transaction
		tx, err := client.Tx(tenantCtx)
		require.NoError(t, err)

		// Create data in transaction
		_, err = tx.CiType.Create().
			SetName("Rollback Test").
			SetAlias("rollback_test").
			SetIcon("icon").
			Save(tenantCtx)
		require.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback()
		require.NoError(t, err)

		// Verify data was not persisted
		finalCount, err := client.CiType.Query().Count(tenantCtx)
		require.NoError(t, err)
		
		assert.Equal(t, initialCount, finalCount, 
			"Rolled back transaction should not persist data")
	})
}

// TestDatabaseConnectionSecurity tests connection-level security features
func TestDatabaseConnectionSecurity(t *testing.T) {
	client := setupDatabaseSecurityTestClient(t)
	defer client.Close()

	t.Run("Connection pooling respects tenant context", func(t *testing.T) {
		// Simulate concurrent requests from different tenants
		tenant1Ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
		tenant2Ctx := context.WithValue(context.Background(), "tenantId", uint64(2))

		// Create data for both tenants concurrently
		done1 := make(chan error, 1)
		done2 := make(chan error, 1)

		go func() {
			_, err := client.CiType.Create().
				SetName("Concurrent T1").
				SetAlias("concurrent_t1").
				SetIcon("icon").
				Save(tenant1Ctx)
			done1 <- err
		}()

		go func() {
			_, err := client.CiType.Create().
				SetName("Concurrent T2").
				SetAlias("concurrent_t2").
				SetIcon("icon").
				Save(tenant2Ctx)
			done2 <- err
		}()

		// Wait for both operations
		err1 := <-done1
		err2 := <-done2

		assert.NoError(t, err1, "Tenant 1 operation should succeed")
		assert.NoError(t, err2, "Tenant 2 operation should succeed")

		// Verify data is properly isolated
		tenant1Data, err := client.CiType.Query().All(tenant1Ctx)
		require.NoError(t, err)

		tenant2Data, err := client.CiType.Query().All(tenant2Ctx)
		require.NoError(t, err)

		// Each tenant should only see their own data
		tenant1HasT1 := false
		tenant1HasT2 := false
		for _, ct := range tenant1Data {
			if ct.Name == "Concurrent T1" {
				tenant1HasT1 = true
			}
			if ct.Name == "Concurrent T2" {
				tenant1HasT2 = true
			}
		}

		tenant2HasT1 := false
		tenant2HasT2 := false
		for _, ct := range tenant2Data {
			if ct.Name == "Concurrent T1" {
				tenant2HasT1 = true
			}
			if ct.Name == "Concurrent T2" {
				tenant2HasT2 = true
			}
		}

		assert.True(t, tenant1HasT1, "Tenant 1 should see its own data")
		assert.False(t, tenant1HasT2, "Tenant 1 should not see tenant 2 data")
		assert.False(t, tenant2HasT1, "Tenant 2 should not see tenant 1 data")
		assert.True(t, tenant2HasT2, "Tenant 2 should see its own data")
	})
}

// BenchmarkTenantFilteringPerformance benchmarks the performance of tenant filtering
func BenchmarkTenantFilteringPerformance(b *testing.B) {
	client := setupDatabaseSecurityTestClient(&testing.T{})
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())

	// Setup test data
	ciType, err := client.CiType.Create().
		SetName("Benchmark Type").
		SetAlias("bench_type").
		SetIcon("icon").
		SetTenantID(1).
		Save(systemCtx)
	require.NoError(b, err)

	// Create test CIs across multiple tenants
	const numTenants = 10
	const cisPerTenant = 100

	for tenantID := uint64(1); tenantID <= numTenants; tenantID++ {
		for i := 0; i < cisPerTenant; i++ {
			_, err := client.Cis.Create().
				SetTypeID(ciType.ID).
				SetStatus(1).
				SetCreatedBy(uuid.Must(uuid.NewV4())).
				SetTenantID(tenantID).
				SetDepartmentID(tenantID).
				Save(systemCtx)
			require.NoError(b, err)
		}
	}

	// Benchmark tenant-filtered queries
	tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(5))

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cis, err := client.Cis.Query().All(tenantCtx)
		require.NoError(b, err)
		assert.Len(b, cis, cisPerTenant, "Should return correct number of CIs")
		
		// Verify all belong to correct tenant
		for _, ci := range cis {
			assert.Equal(b, uint64(5), ci.TenantID)
		}
	}
}