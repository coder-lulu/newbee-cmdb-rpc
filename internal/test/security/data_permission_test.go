package security

import (
	"context"
	"fmt"
	"testing"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/enttest"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// DataPermissionLevel represents the 5-level data permission model
type DataPermissionLevel string

const (
	DataPermAll           DataPermissionLevel = "All"           // 全部数据权限
	DataPermCustomDept    DataPermissionLevel = "CustomDept"    // 自定义部门权限  
	DataPermOwnDeptAndSub DataPermissionLevel = "OwnDeptAndSub" // 本部门及子部门权限
	DataPermOwnDept       DataPermissionLevel = "OwnDept"       // 本部门权限
	DataPermSelf          DataPermissionLevel = "Self"          // 仅本人数据权限
)

// TestUser represents a test user with data permissions
type TestDataPermUser struct {
	ID           uuid.UUID
	Username     string
	TenantID     uint64
	DepartmentID uint64
	DataPerm     DataPermissionLevel
}

// TestDepartment represents a department hierarchy for testing
type TestDataPermDept struct {
	ID       uint64
	Name     string
	TenantID uint64
	ParentID *uint64
}

// setupDataPermissionTestClient creates a test client with data permission interceptors
func setupDataPermissionTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Register tenant hooks first
	client.Use(hooks.TenantMutationHook())
	client.Intercept(hooks.TenantQueryInterceptor())
	
	// Register data permission interceptors for CMDB entities
	hooks.RegisterDataPermissionInterceptorsWithTenant(client,
		"cis", "ci_types", "ci_permissions", "import_tasks", "import_records")
	
	return client
}

// setupDataPermissionTestData creates comprehensive test data for data permission testing
func setupDataPermissionTestData(t *testing.T, client *ent.Client) (
	[]TestDataPermUser, []TestDataPermDept, []*ent.CiType, []*ent.Cis) {
	
	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Setup departments with hierarchy
	departments := []TestDataPermDept{
		{ID: 1, Name: "IT Department", TenantID: 1, ParentID: nil},
		{ID: 2, Name: "Security Team", TenantID: 1, ParentID: &[]uint64{1}[0]},
		{ID: 3, Name: "DevOps Team", TenantID: 1, ParentID: &[]uint64{1}[0]},
		{ID: 4, Name: "Infrastructure", TenantID: 1, ParentID: &[]uint64{3}[0]},
		{ID: 5, Name: "HR Department", TenantID: 1, ParentID: nil},
		{ID: 6, Name: "Tech Department", TenantID: 2, ParentID: nil},
	}
	
	// Setup users with different permission levels
	users := []TestDataPermUser{
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "admin",
			TenantID:     1,
			DepartmentID: 1,
			DataPerm:     DataPermAll,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "manager_it",
			TenantID:     1,
			DepartmentID: 1,
			DataPerm:     DataPermOwnDeptAndSub,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "lead_security",
			TenantID:     1,
			DepartmentID: 2,
			DataPerm:     DataPermOwnDept,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "user_devops",
			TenantID:     1,
			DepartmentID: 3,
			DataPerm:     DataPermSelf,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "user_infra",
			TenantID:     1,
			DepartmentID: 4,
			DataPerm:     DataPermSelf,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "user_hr",
			TenantID:     1,
			DepartmentID: 5,
			DataPerm:     DataPermOwnDept,
		},
		{
			ID:           uuid.Must(uuid.NewV4()),
			Username:     "user_tenant2",
			TenantID:     2,
			DepartmentID: 6,
			DataPerm:     DataPermAll,
		},
	}
	
	// Create CI types
	ciTypes := []*ent.CiType{}
	ciTypeNames := []string{"Server", "Application", "Database", "Network"}
	
	for i, name := range ciTypeNames {
		ciType, err := client.CiType.Create().
			SetName(name).
			SetAlias(fmt.Sprintf("%s_%d", name, i)).
			SetIcon(fmt.Sprintf("%s-icon", name)).
			SetTenantID(1).
			Save(systemCtx)
		require.NoError(t, err)
		ciTypes = append(ciTypes, ciType)
	}
	
	// Create one CI type for tenant 2
	ciTypeTenant2, err := client.CiType.Create().
		SetName("Cloud Service").
		SetAlias("cloud_service").
		SetIcon("cloud-icon").
		SetTenantID(2).
		Save(systemCtx)
	require.NoError(t, err)
	ciTypes = append(ciTypes, ciTypeTenant2)
	
	// Create CIs owned by different users/departments
	cis := []*ent.Cis{}
	
	// CIs for different departments and users in tenant 1
	ciOwnershipData := []struct {
		typeIndex    int
		departmentID uint64
		userID       uuid.UUID
		tenantID     uint64
	}{
		{0, 1, users[0].ID, 1}, // admin's server
		{0, 1, users[1].ID, 1}, // manager_it's server
		{1, 2, users[2].ID, 1}, // lead_security's app
		{1, 2, users[2].ID, 1}, // lead_security's second app
		{2, 3, users[3].ID, 1}, // user_devops's database
		{3, 4, users[4].ID, 1}, // user_infra's network
		{0, 5, users[5].ID, 1}, // user_hr's server
		{4, 6, users[6].ID, 2}, // tenant2 user's cloud service
	}
	
	for i, data := range ciOwnershipData {
		ci, err := client.Cis.Create().
			SetTypeID(ciTypes[data.typeIndex].ID).
			SetStatus(1).
			SetCreatedBy(data.userID).
			SetTenantID(data.tenantID).
			SetDepartmentID(data.departmentID).
			Save(systemCtx)
		require.NoError(t, err)
		cis = append(cis, ci)
	}
	
	return users, departments, ciTypes, cis
}

// createUserContext creates context with user and data permission information
func createUserContext(user TestDataPermUser) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenantId", user.TenantID)
	ctx = context.WithValue(ctx, "userId", user.ID)
	ctx = context.WithValue(ctx, "departmentId", user.DepartmentID)
	ctx = context.WithValue(ctx, "dataPerm", string(user.DataPerm))
	return ctx
}

// TestDataPermissionAll tests "All" level data permission
func TestDataPermissionAll(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, cis := setupDataPermissionTestData(t, client)
	
	// Find admin user (DataPermAll)
	var adminUser TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermAll && user.TenantID == 1 {
			adminUser = user
			break
		}
	}
	
	t.Run("Admin can see all CIs in tenant", func(t *testing.T) {
		ctx := createUserContext(adminUser)
		
		allCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Should see all CIs in tenant 1 (7 CIs total, 7 in tenant 1)
		tenant1CisCount := 0
		for _, ci := range cis {
			if ci.TenantID == 1 {
				tenant1CisCount++
			}
		}
		
		assert.Equal(t, tenant1CisCount, len(allCis), 
			"Admin should see all CIs in tenant")
		
		// Verify all returned CIs belong to same tenant
		for _, ci := range allCis {
			assert.Equal(t, adminUser.TenantID, ci.TenantID,
				"All CIs should belong to admin's tenant")
		}
	})
	
	t.Run("Admin cannot see CIs from other tenants", func(t *testing.T) {
		ctx := createUserContext(adminUser)
		
		allCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Should not see any CIs from tenant 2
		for _, ci := range allCis {
			assert.NotEqual(t, uint64(2), ci.TenantID,
				"Admin should not see CIs from other tenants")
		}
	})
}

// TestDataPermissionOwnDeptAndSub tests "OwnDeptAndSub" level data permission
func TestDataPermissionOwnDeptAndSub(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, departments, _, cis := setupDataPermissionTestData(t, client)
	
	// Find IT manager (department 1, OwnDeptAndSub permission)
	var managerUser TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermOwnDeptAndSub && user.DepartmentID == 1 {
			managerUser = user
			break
		}
	}
	
	t.Run("Manager can see own department and sub-departments data", func(t *testing.T) {
		ctx := createUserContext(managerUser)
		
		visibleCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Manager in department 1 should see:
		// - Department 1 (IT) CIs: admin's + manager's CIs  
		// - Department 2 (Security, child of IT): lead_security's CIs
		// - Department 3 (DevOps, child of IT): user_devops's CIs
		// - Department 4 (Infrastructure, child of DevOps): user_infra's CIs
		// Should NOT see:
		// - Department 5 (HR): user_hr's CI
		// - Tenant 2 CIs
		
		expectedDepts := map[uint64]bool{1: true, 2: true, 3: true, 4: true}
		
		for _, ci := range visibleCis {
			assert.True(t, expectedDepts[ci.DepartmentID],
				"Manager should only see CIs from own department and sub-departments, got dept %d", ci.DepartmentID)
			assert.Equal(t, managerUser.TenantID, ci.TenantID,
				"All visible CIs should be in same tenant")
		}
		
		// Count expected CIs
		expectedCICount := 0
		for _, ci := range cis {
			if ci.TenantID == 1 && expectedDepts[ci.DepartmentID] {
				expectedCICount++
			}
		}
		
		assert.Equal(t, expectedCICount, len(visibleCis),
			"Manager should see correct number of CIs")
	})
}

// TestDataPermissionOwnDept tests "OwnDept" level data permission  
func TestDataPermissionOwnDept(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, cis := setupDataPermissionTestData(t, client)
	
	// Find security lead (department 2, OwnDept permission)
	var securityLead TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermOwnDept && user.DepartmentID == 2 {
			securityLead = user
			break
		}
	}
	
	t.Run("User can see only own department data", func(t *testing.T) {
		ctx := createUserContext(securityLead)
		
		visibleCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Should only see CIs from department 2
		for _, ci := range visibleCis {
			assert.Equal(t, uint64(2), ci.DepartmentID,
				"User should only see CIs from own department")
			assert.Equal(t, securityLead.TenantID, ci.TenantID,
				"All visible CIs should be in same tenant")
		}
		
		// Count expected CIs from department 2
		expectedCICount := 0
		for _, ci := range cis {
			if ci.TenantID == 1 && ci.DepartmentID == 2 {
				expectedCICount++
			}
		}
		
		assert.Equal(t, expectedCICount, len(visibleCis),
			"User should see correct number of CIs from own department")
	})
}

// TestDataPermissionSelf tests "Self" level data permission
func TestDataPermissionSelf(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, cis := setupDataPermissionTestData(t, client)
	
	// Find devops user (Self permission)
	var devopsUser TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermSelf && user.DepartmentID == 3 {
			devopsUser = user
			break
		}
	}
	
	t.Run("User can see only own data", func(t *testing.T) {
		ctx := createUserContext(devopsUser)
		
		visibleCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Should only see CIs created by this user
		for _, ci := range visibleCis {
			assert.Equal(t, devopsUser.ID, ci.CreatedBy,
				"User should only see CIs created by themselves")
			assert.Equal(t, devopsUser.TenantID, ci.TenantID,
				"All visible CIs should be in same tenant")
		}
		
		// Count expected CIs created by this user
		expectedCICount := 0
		for _, ci := range cis {
			if ci.TenantID == 1 && ci.CreatedBy == devopsUser.ID {
				expectedCICount++
			}
		}
		
		assert.Equal(t, expectedCICount, len(visibleCis),
			"User should see correct number of own CIs")
	})
}

// TestDataPermissionCrossTenanIsolation tests that data permissions respect tenant boundaries
func TestDataPermissionCrossTenantIsolation(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, _ := setupDataPermissionTestData(t, client)
	
	// Find tenant 2 user
	var tenant2User TestDataPermUser
	for _, user := range users {
		if user.TenantID == 2 {
			tenant2User = user
			break
		}
	}
	
	t.Run("User can only see data from own tenant regardless of data permission", func(t *testing.T) {
		ctx := createUserContext(tenant2User)
		
		visibleCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Should only see CIs from tenant 2
		for _, ci := range visibleCis {
			assert.Equal(t, uint64(2), ci.TenantID,
				"User should only see CIs from own tenant")
		}
		
		// Should see exactly 1 CI (the one we created for tenant 2)
		assert.Equal(t, 1, len(visibleCis),
			"Tenant 2 user should see exactly 1 CI")
	})
}

// TestDataPermissionInheritance tests department hierarchy in data permissions
func TestDataPermissionInheritance(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, departments, _, _ := setupDataPermissionTestData(t, client)
	
	// Test department hierarchy: IT(1) -> DevOps(3) -> Infrastructure(4)
	// Manager in IT with OwnDeptAndSub should see all three levels
	var itManager TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermOwnDeptAndSub && user.DepartmentID == 1 {
			itManager = user
			break
		}
	}
	
	t.Run("Department hierarchy is respected in data permissions", func(t *testing.T) {
		ctx := createUserContext(itManager)
		
		visibleCis, err := client.Cis.Query().All(ctx)
		require.NoError(t, err)
		
		// Build expected department hierarchy
		departmentHierarchy := make(map[uint64][]uint64)
		for _, dept := range departments {
			if dept.ParentID != nil {
				departmentHierarchy[*dept.ParentID] = append(departmentHierarchy[*dept.ParentID], dept.ID)
			}
		}
		
		// For department 1, should include: 1, 2, 3, 4 (but not 5 - HR)
		expectedDepts := map[uint64]bool{1: true, 2: true, 3: true, 4: true}
		
		visibleDepts := make(map[uint64]bool)
		for _, ci := range visibleCis {
			visibleDepts[ci.DepartmentID] = true
		}
		
		for deptID := range expectedDepts {
			assert.True(t, visibleDepts[deptID] || !cisExistForDepartment(t, client, deptID),
				"Should see CIs from department %d or no CIs exist", deptID)
		}
		
		// Should NOT see HR department (5)
		assert.False(t, visibleDepts[5],
			"Should not see CIs from HR department")
	})
}

// Helper function to check if CIs exist for a department
func cisExistForDepartment(t *testing.T, client *ent.Client, deptID uint64) bool {
	systemCtx := hooks.NewSystemContext(context.Background())
	count, err := client.Cis.Query().Where(cis.DepartmentID(deptID)).Count(systemCtx)
	require.NoError(t, err)
	return count > 0
}

// TestDataPermissionUpdate tests that data permission applies to update operations
func TestDataPermissionUpdate(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, cis := setupDataPermissionTestData(t, client)
	
	// Find a user with Self permission
	var selfPermUser TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermSelf {
			selfPermUser = user
			break
		}
	}
	
	// Find a CI belonging to this user and another CI belonging to someone else
	var ownCI, otherCI *ent.Cis
	for _, ci := range cis {
		if ci.TenantID == selfPermUser.TenantID {
			if ci.CreatedBy == selfPermUser.ID {
				ownCI = ci
			} else if otherCI == nil {
				otherCI = ci
			}
		}
	}
	
	require.NotNil(t, ownCI, "Should have own CI for testing")
	require.NotNil(t, otherCI, "Should have other user's CI for testing")
	
	t.Run("User can update own CI", func(t *testing.T) {
		ctx := createUserContext(selfPermUser)
		
		err := client.Cis.UpdateOneID(ownCI.ID).
			SetStatus(2). // Change status
			Exec(ctx)
		
		// Should succeed for own CI
		assert.NoError(t, err, "User should be able to update own CI")
	})
	
	t.Run("User cannot update other user's CI", func(t *testing.T) {
		ctx := createUserContext(selfPermUser)
		
		err := client.Cis.UpdateOneID(otherCI.ID).
			SetStatus(2). // Try to change status
			Exec(ctx)
		
		// Should fail for other user's CI
		assert.Error(t, err, "User should not be able to update other user's CI")
	})
}

// TestDataPermissionDelete tests that data permission applies to delete operations
func TestDataPermissionDelete(t *testing.T) {
	client := setupDataPermissionTestClient(t)
	defer client.Close()
	
	users, _, _, _ := setupDataPermissionTestData(t, client)
	
	// Create additional test CIs for deletion testing
	systemCtx := hooks.NewSystemContext(context.Background())
	
	// Find users for testing
	var selfPermUser, adminUser TestDataPermUser
	for _, user := range users {
		if user.DataPerm == DataPermSelf && selfPermUser.ID == uuid.Nil {
			selfPermUser = user
		} else if user.DataPerm == DataPermAll && adminUser.ID == uuid.Nil {
			adminUser = user
		}
	}
	
	// Create CI type for testing
	ciType, err := client.CiType.Create().
		SetName("Test Type").
		SetAlias("test_type").
		SetIcon("test-icon").
		SetTenantID(1).
		Save(systemCtx)
	require.NoError(t, err)
	
	// Create CIs for deletion testing
	selfUserCI, err := client.Cis.Create().
		SetTypeID(ciType.ID).
		SetStatus(1).
		SetCreatedBy(selfPermUser.ID).
		SetTenantID(selfPermUser.TenantID).
		SetDepartmentID(selfPermUser.DepartmentID).
		Save(systemCtx)
	require.NoError(t, err)
	
	adminUserCI, err := client.Cis.Create().
		SetTypeID(ciType.ID).
		SetStatus(1).
		SetCreatedBy(adminUser.ID).
		SetTenantID(adminUser.TenantID).
		SetDepartmentID(adminUser.DepartmentID).
		Save(systemCtx)
	require.NoError(t, err)
	
	t.Run("User can delete own CI", func(t *testing.T) {
		ctx := createUserContext(selfPermUser)
		
		err := client.Cis.DeleteOneID(selfUserCI.ID).Exec(ctx)
		assert.NoError(t, err, "User should be able to delete own CI")
	})
	
	t.Run("User cannot delete other user's CI", func(t *testing.T) {
		ctx := createUserContext(selfPermUser)
		
		err := client.Cis.DeleteOneID(adminUserCI.ID).Exec(ctx)
		assert.Error(t, err, "User should not be able to delete other user's CI")
	})
}