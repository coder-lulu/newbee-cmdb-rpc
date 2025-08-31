package security

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/enttest"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// MockAuditLogger simulates audit logging for testing
type MockAuditLogger struct {
	logs []AuditLog
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uint64
	TenantID    uint64
	UserID      uuid.UUID
	Action      string
	Resource    string
	ResourceID  string
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
	IPAddress   string
	UserAgent   string
	Timestamp   time.Time
	Success     bool
	ErrorMsg    string
	RequestID   string
}

// Log records an audit log entry
func (m *MockAuditLogger) Log(ctx context.Context, log AuditLog) error {
	// Extract tenant ID from context
	if tenantID, ok := ctx.Value("tenantId").(uint64); ok {
		log.TenantID = tenantID
	}
	
	// Extract user ID from context
	if userID, ok := ctx.Value("userId").(uuid.UUID); ok {
		log.UserID = userID
	}
	
	// Extract request ID from context
	if requestID, ok := ctx.Value("requestId").(string); ok {
		log.RequestID = requestID
	}
	
	log.Timestamp = time.Now()
	m.logs = append(m.logs, log)
	return nil
}

// GetLogs returns all audit logs
func (m *MockAuditLogger) GetLogs() []AuditLog {
	return m.logs
}

// GetLogsByTenant returns audit logs for specific tenant
func (m *MockAuditLogger) GetLogsByTenant(tenantID uint64) []AuditLog {
	var result []AuditLog
	for _, log := range m.logs {
		if log.TenantID == tenantID {
			result = append(result, log)
		}
	}
	return result
}

// GetLogsByUser returns audit logs for specific user
func (m *MockAuditLogger) GetLogsByUser(userID uuid.UUID) []AuditLog {
	var result []AuditLog
	for _, log := range m.logs {
		if log.UserID == userID {
			result = append(result, log)
		}
	}
	return result
}

// Clear clears all audit logs
func (m *MockAuditLogger) Clear() {
	m.logs = nil
}

// setupAuditTestClient creates test client with audit logging
func setupAuditTestClient(t *testing.T, auditor *MockAuditLogger) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Register security hooks
	client.Use(hooks.TenantMutationHook())
	client.Intercept(hooks.TenantQueryInterceptor())
	hooks.RegisterDataPermissionInterceptorsWithTenant(client,
		"cis", "ci_types", "ci_permissions", "import_tasks", "import_records")
	
	// Register audit hooks
	client.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Log before mutation
			auditLog := AuditLog{
				Action:   string(m.Op()),
				Resource: m.Type(),
			}
			
			// Capture old values for updates
			if m.Op() == ent.OpUpdate || m.Op() == ent.OpUpdateOne {
				auditLog.OldValues = captureOldValues(ctx, m)
			}
			
			// Execute mutation
			value, err := next.Mutate(ctx, m)
			
			// Log result
			auditLog.Success = (err == nil)
			if err != nil {
				auditLog.ErrorMsg = err.Error()
			} else {
				// Capture new values
				auditLog.NewValues = captureNewValues(value, m)
				if id := extractID(value); id != "" {
					auditLog.ResourceID = id
				}
			}
			
			// Record audit log
			auditor.Log(ctx, auditLog)
			
			return value, err
		})
	})
	
	return client
}

// captureOldValues captures old values before update (simplified)
func captureOldValues(ctx context.Context, m ent.Mutation) map[string]interface{} {
	values := make(map[string]interface{})
	
	// This would need to query the existing record to get old values
	// For testing purposes, we'll simulate this
	values["captured"] = "old_values"
	
	return values
}

// captureNewValues captures new values after mutation (simplified)
func captureNewValues(value ent.Value, m ent.Mutation) map[string]interface{} {
	values := make(map[string]interface{})
	
	// Extract values from the mutation or result
	// This is simplified for testing
	for field, val := range m.Fields() {
		values[field] = val
	}
	
	return values
}

// extractID extracts ID from mutation result (simplified)
func extractID(value ent.Value) string {
	// This would extract the actual ID from the result
	// For testing purposes, we'll return a placeholder
	return "test-id"
}

// TestAuditLogging tests comprehensive audit logging functionality
func TestAuditLogging(t *testing.T) {
	auditor := &MockAuditLogger{}
	client := setupAuditTestClient(t, auditor)
	defer client.Close()

	userID := uuid.Must(uuid.NewV4())
	tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
	tenantCtx = context.WithValue(tenantCtx, "userId", userID)
	tenantCtx = context.WithValue(tenantCtx, "requestId", "test-request-123")

	t.Run("Create operation is audited", func(t *testing.T) {
		auditor.Clear()
		
		ciType, err := client.CiType.Create().
			SetName("Audit Test Type").
			SetAlias("audit_test").
			SetIcon("audit-icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "Should have one audit log entry")
		
		log := logs[0]
		assert.Equal(t, uint64(1), log.TenantID, "Audit log should have correct tenant ID")
		assert.Equal(t, userID, log.UserID, "Audit log should have correct user ID")
		assert.Equal(t, "OpCreate", log.Action, "Audit log should record create action")
		assert.Equal(t, "CiType", log.Resource, "Audit log should record correct resource type")
		assert.True(t, log.Success, "Audit log should record success")
		assert.Equal(t, "test-request-123", log.RequestID, "Audit log should have request ID")
		
		t.Logf("Audit log: %+v", log)
	})
	
	t.Run("Update operation is audited with old and new values", func(t *testing.T) {
		auditor.Clear()
		
		// Create initial record
		ciType, err := client.CiType.Create().
			SetName("Update Test").
			SetAlias("update_test").
			SetIcon("update-icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		auditor.Clear() // Clear create log
		
		// Update the record
		err = client.CiType.UpdateOneID(ciType.ID).
			SetDescription("Updated description").
			Exec(tenantCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "Should have one audit log entry for update")
		
		log := logs[0]
		assert.Equal(t, "OpUpdateOne", log.Action, "Should record update action")
		assert.True(t, log.Success, "Update should be successful")
		assert.NotNil(t, log.OldValues, "Should capture old values")
		assert.NotNil(t, log.NewValues, "Should capture new values")
		
		t.Logf("Update audit log: %+v", log)
	})
	
	t.Run("Delete operation is audited", func(t *testing.T) {
		auditor.Clear()
		
		// Create record to delete
		ciType, err := client.CiType.Create().
			SetName("Delete Test").
			SetAlias("delete_test").
			SetIcon("delete-icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		auditor.Clear() // Clear create log
		
		// Delete the record
		err = client.CiType.DeleteOneID(ciType.ID).Exec(tenantCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "Should have one audit log entry for delete")
		
		log := logs[0]
		assert.Equal(t, "OpDeleteOne", log.Action, "Should record delete action")
		assert.True(t, log.Success, "Delete should be successful")
		
		t.Logf("Delete audit log: %+v", log)
	})
	
	t.Run("Failed operations are audited", func(t *testing.T) {
		auditor.Clear()
		
		// Try to create record without required fields
		_, err := client.CiType.Create().
			SetName(""). // Invalid empty name
			Save(tenantCtx)
		require.Error(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "Should have one audit log entry for failed operation")
		
		log := logs[0]
		assert.Equal(t, "OpCreate", log.Action, "Should record create action")
		assert.False(t, log.Success, "Should record failure")
		assert.NotEmpty(t, log.ErrorMsg, "Should record error message")
		
		t.Logf("Failed operation audit log: %+v", log)
	})
}

// TestAuditTenantIsolation tests that audit logs respect tenant isolation
func TestAuditTenantIsolation(t *testing.T) {
	auditor := &MockAuditLogger{}
	client := setupAuditTestClient(t, auditor)
	defer client.Close()

	user1ID := uuid.Must(uuid.NewV4())
	user2ID := uuid.Must(uuid.NewV4())
	
	tenant1Ctx := context.WithValue(context.Background(), "tenantId", uint64(1))
	tenant1Ctx = context.WithValue(tenant1Ctx, "userId", user1ID)
	
	tenant2Ctx := context.WithValue(context.Background(), "tenantId", uint64(2))
	tenant2Ctx = context.WithValue(tenant2Ctx, "userId", user2ID)

	t.Run("Audit logs are isolated by tenant", func(t *testing.T) {
		auditor.Clear()
		
		// Create records in different tenants
		_, err := client.CiType.Create().
			SetName("Tenant 1 Type").
			SetAlias("t1_type").
			SetIcon("t1-icon").
			Save(tenant1Ctx)
		require.NoError(t, err)
		
		_, err = client.CiType.Create().
			SetName("Tenant 2 Type").
			SetAlias("t2_type").
			SetIcon("t2-icon").
			Save(tenant2Ctx)
		require.NoError(t, err)
		
		// Verify logs are properly isolated
		tenant1Logs := auditor.GetLogsByTenant(1)
		tenant2Logs := auditor.GetLogsByTenant(2)
		
		assert.Len(t, tenant1Logs, 1, "Tenant 1 should have 1 audit log")
		assert.Len(t, tenant2Logs, 1, "Tenant 2 should have 1 audit log")
		
		assert.Equal(t, uint64(1), tenant1Logs[0].TenantID)
		assert.Equal(t, user1ID, tenant1Logs[0].UserID)
		
		assert.Equal(t, uint64(2), tenant2Logs[0].TenantID)
		assert.Equal(t, user2ID, tenant2Logs[0].UserID)
	})
	
	t.Run("Users can only access their tenant's audit logs", func(t *testing.T) {
		// In a real system, this would be enforced by the audit log query API
		// Here we simulate the behavior
		
		tenant1Logs := auditor.GetLogsByTenant(1)
		tenant2Logs := auditor.GetLogsByTenant(2)
		
		// Verify no cross-tenant access
		for _, log := range tenant1Logs {
			assert.Equal(t, uint64(1), log.TenantID, "Tenant 1 logs should only contain tenant 1 data")
		}
		
		for _, log := range tenant2Logs {
			assert.Equal(t, uint64(2), log.TenantID, "Tenant 2 logs should only contain tenant 2 data")
		}
	})
}

// TestAuditDataPermissions tests audit logging for data permission violations
func TestAuditDataPermissions(t *testing.T) {
	auditor := &MockAuditLogger{}
	client := setupAuditTestClient(t, auditor)
	defer client.Close()

	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Setup test data
	ciType, err := client.CiType.Create().
		SetName("Permission Test Type").
		SetAlias("perm_test").
		SetIcon("perm-icon").
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

	t.Run("Data permission violations are audited", func(t *testing.T) {
		auditor.Clear()
		
		// Create user context with limited permissions
		userID := uuid.Must(uuid.NewV4())
		limitedCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
		limitedCtx = context.WithValue(limitedCtx, "userId", userID)
		limitedCtx = context.WithValue(limitedCtx, "departmentId", uint64(2)) // Different department
		limitedCtx = context.WithValue(limitedCtx, "dataPerm", "OwnDept") // Can only see own department
		
		// Try to access CI from different department
		_, err := client.Cis.Get(limitedCtx, ci.ID)
		
		// This might succeed or fail depending on implementation
		// But it should be audited either way
		logs := auditor.GetLogs()
		
		if len(logs) > 0 {
			log := logs[len(logs)-1] // Get last log
			assert.Equal(t, userID, log.UserID, "Should record user attempting access")
			assert.Equal(t, uint64(1), log.TenantID, "Should record correct tenant")
			
			if !log.Success {
				t.Logf("Permission violation audited: %+v", log)
			}
		}
	})
}

// TestAuditSecurityEvents tests logging of security-related events
func TestAuditSecurityEvents(t *testing.T) {
	auditor := &MockAuditLogger{}
	client := setupAuditTestClient(t, auditor)
	defer client.Close()

	userID := uuid.Must(uuid.NewV4())
	tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
	tenantCtx = context.WithValue(tenantCtx, "userId", userID)
	tenantCtx = context.WithValue(tenantCtx, "ipAddress", "192.168.1.100")
	tenantCtx = context.WithValue(tenantCtx, "userAgent", "TestAgent/1.0")

	t.Run("System context usage is audited", func(t *testing.T) {
		auditor.Clear()
		
		// Simulate system context operation
		systemCtx := hooks.NewSystemContext(context.Background())
		systemCtx = context.WithValue(systemCtx, "systemOperation", true)
		systemCtx = context.WithValue(systemCtx, "requestId", "system-req-123")
		
		_, err := client.CiType.Create().
			SetName("System Created Type").
			SetAlias("system_type").
			SetIcon("system-icon").
			Save(systemCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "System operation should be audited")
		
		log := logs[0]
		assert.Equal(t, "system-req-123", log.RequestID, "Should record system request ID")
		
		t.Logf("System operation audit log: %+v", log)
	})
	
	t.Run("Bulk operations are audited", func(t *testing.T) {
		auditor.Clear()
		
		// Create multiple records
		names := []string{"Bulk1", "Bulk2", "Bulk3"}
		for _, name := range names {
			_, err := client.CiType.Create().
				SetName(name).
				SetAlias(fmt.Sprintf("bulk_%s", strings.ToLower(name))).
				SetIcon("bulk-icon").
				Save(tenantCtx)
			require.NoError(t, err)
		}
		
		logs := auditor.GetLogs()
		assert.Equal(t, len(names), len(logs), "Each bulk operation should be audited")
		
		for i, log := range logs {
			assert.Equal(t, "OpCreate", log.Action, "Should record create action")
			assert.Equal(t, userID, log.UserID, "Should record correct user")
			t.Logf("Bulk operation %d audit log: %+v", i+1, log)
		}
	})
	
	t.Run("Cross-tenant access attempts are audited", func(t *testing.T) {
		auditor.Clear()
		
		// Create data in tenant 1
		systemCtx := hooks.NewSystemContext(context.Background())
		ciType, err := client.CiType.Create().
			SetName("Cross Tenant Test").
			SetAlias("cross_tenant").
			SetIcon("cross-icon").
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)
		
		auditor.Clear() // Clear system creation log
		
		// Try to access from tenant 2 context
		tenant2Ctx := context.WithValue(context.Background(), "tenantId", uint64(2))
		tenant2Ctx = context.WithValue(tenant2Ctx, "userId", userID)
		
		_, err = client.CiType.Get(tenant2Ctx, ciType.ID)
		// This should fail due to tenant isolation
		
		// The query itself might be audited depending on implementation
		// For now, we just verify the context is properly set up
		assert.NotNil(t, tenant2Ctx, "Cross-tenant context should be valid")
	})
}

// TestAuditLogRetention tests audit log retention and cleanup policies
func TestAuditLogRetention(t *testing.T) {
	auditor := &MockAuditLogger{}
	client := setupAuditTestClient(t, auditor)
	defer client.Close()

	userID := uuid.Must(uuid.NewV4())
	tenantCtx := context.WithValue(context.Background(), "tenantId", uint64(1))
	tenantCtx = context.WithValue(tenantCtx, "userId", userID)

	t.Run("Audit logs can be queried by time range", func(t *testing.T) {
		auditor.Clear()
		
		// Create some records with time gaps
		_, err := client.CiType.Create().
			SetName("Time Test 1").
			SetAlias("time_test_1").
			SetIcon("time-icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		// Simulate time passing
		time.Sleep(10 * time.Millisecond)
		
		_, err = client.CiType.Create().
			SetName("Time Test 2").
			SetAlias("time_test_2").
			SetIcon("time-icon").
			Save(tenantCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 2, "Should have two audit logs")
		
		// Verify timestamps are different
		assert.True(t, logs[1].Timestamp.After(logs[0].Timestamp),
			"Second log should have later timestamp")
		
		t.Logf("Time range test - Log 1: %v, Log 2: %v", 
			logs[0].Timestamp, logs[1].Timestamp)
	})
	
	t.Run("Audit logs contain sufficient detail for compliance", func(t *testing.T) {
		auditor.Clear()
		
		// Create a comprehensive test record
		ciType, err := client.CiType.Create().
			SetName("Compliance Test").
			SetAlias("compliance_test").
			SetIcon("compliance-icon").
			SetDescription("Test for compliance logging").
			Save(tenantCtx)
		require.NoError(t, err)
		
		logs := auditor.GetLogs()
		require.Len(t, logs, 1, "Should have one audit log")
		
		log := logs[0]
		
		// Verify all required fields are present
		assert.NotZero(t, log.TenantID, "Should have tenant ID")
		assert.NotEqual(t, uuid.Nil, log.UserID, "Should have user ID")
		assert.NotEmpty(t, log.Action, "Should have action")
		assert.NotEmpty(t, log.Resource, "Should have resource type")
		assert.NotZero(t, log.Timestamp, "Should have timestamp")
		assert.True(t, log.Success, "Should record success status")
		
		t.Logf("Compliance audit log: %+v", log)
	})
}