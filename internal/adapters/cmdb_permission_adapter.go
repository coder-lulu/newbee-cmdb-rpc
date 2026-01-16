package adapters

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/types"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// CMDBPermissionAdapter CMDB权限服务适配器
// 将CMDB的权限控制系统适配为通用权限接口
type CMDBPermissionAdapter struct {
	svcCtx *svc.ServiceContext
	config *CMDBPermissionConfig
}

// CMDBPermissionConfig CMDB权限适配器配置
type CMDBPermissionConfig struct {
	// 是否启用权限检查
	EnablePermissionCheck bool `json:"enablePermissionCheck"`

	// 默认权限策略
	DefaultPolicy string `json:"defaultPolicy"` // allow, deny

	// 权限缓存配置
	CacheEnabled bool `json:"cacheEnabled"`
	CacheTTL     int  `json:"cacheTTL"` // 缓存TTL（秒）

	// 管理员用户ID列表（跳过权限检查）
	AdminUserIDs []string `json:"adminUserIds"`

	// 类型级别的权限配置
	TypePermissions map[string]*TypePermissionConfig `json:"typePermissions"`
}

// TypePermissionConfig 类型权限配置
type TypePermissionConfig struct {
	AllowRead        bool     `json:"allowRead"`        // 是否允许读取
	AllowSelect      bool     `json:"allowSelect"`      // 是否允许选择
	RestrictedFields []string `json:"restrictedFields"` // 受限制的字段
}

// NewCMDBPermissionAdapter 创建CMDB权限适配器
func NewCMDBPermissionAdapter(svcCtx *svc.ServiceContext, config *CMDBPermissionConfig) *CMDBPermissionAdapter {
	if config == nil {
		config = getDefaultCMDBPermissionConfig()
	}

	return &CMDBPermissionAdapter{
		svcCtx: svcCtx,
		config: config,
	}
}

// CheckAssetTypeAccess 检查用户对特定资产类型的访问权限
func (a *CMDBPermissionAdapter) CheckAssetTypeAccess(ctx context.Context, userID string, assetTypeID string) (bool, error) {
	// 如果禁用权限检查，直接返回允许
	if !a.config.EnablePermissionCheck {
		return true, nil
	}

	// 检查是否为管理员用户
	if a.isAdminUser(userID) {
		return true, nil
	}

	// 检查类型级别的权限配置
	if typeConfig, exists := a.config.TypePermissions[assetTypeID]; exists {
		if !typeConfig.AllowRead {
			return false, nil
		}
	}

	// 转换参数
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("无效的用户ID: %s", userID)
	}

	typeIDUint, err := strconv.ParseUint(assetTypeID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("无效的类型ID: %s", assetTypeID)
	}

	// 调用CMDB权限检查逻辑
	logic := cipermission.NewCheckCiTypePermissionLogic(ctx, a.svcCtx)
	req := &cmdb.CiPermissionReq{
		UserId:   userIDUint,
		CiTypeId: typeIDUint,
		Action:   "read", // 读取权限
	}

	resp, err := logic.CheckCiTypePermission(req)
	if err != nil {
		logx.Errorf("检查CI类型权限失败 [userID=%s, typeID=%s]: %v", userID, assetTypeID, err)
		// 如果权限检查失败，根据默认策略决定
		return a.config.DefaultPolicy == "allow", nil
	}

	return resp.HasPermission, nil
}

// FilterVisibleAssets 过滤用户可见的资产列表
func (a *CMDBPermissionAdapter) FilterVisibleAssets(ctx context.Context, userID string, assets []*types.Asset) ([]*types.Asset, error) {
	// 如果禁用权限检查，返回所有资产
	if !a.config.EnablePermissionCheck {
		return assets, nil
	}

	// 如果是管理员用户，返回所有资产
	if a.isAdminUser(userID) {
		return assets, nil
	}

	// 按类型分组检查权限
	typePermissions := make(map[string]bool)
	filteredAssets := make([]*types.Asset, 0, len(assets))

	for _, asset := range assets {
		// 检查是否已经检查过该类型的权限
		hasPermission, checked := typePermissions[asset.TypeID]

		if !checked {
			// 检查类型权限
			var err error
			hasPermission, err = a.CheckAssetTypeAccess(ctx, userID, asset.TypeID)
			if err != nil {
				logx.Errorf("检查资产类型权限失败 [userID=%s, typeID=%s]: %v", userID, asset.TypeID, err)
				hasPermission = a.config.DefaultPolicy == "allow"
			}
			typePermissions[asset.TypeID] = hasPermission
		}

		if hasPermission {
			// 检查数据级别权限
			if a.checkDataLevelPermission(ctx, userID, asset) {
				// 过滤敏感字段
				filteredAsset := a.filterSensitiveFields(asset)
				filteredAssets = append(filteredAssets, filteredAsset)
			}
		}
	}

	return filteredAssets, nil
}

// GetUserAssetScope 获取用户的资产访问范围
func (a *CMDBPermissionAdapter) GetUserAssetScope(ctx context.Context, userID string) (*types.AssetScope, error) {
	scope := &types.AssetScope{
		AllowedTypes:  make([]string, 0),
		DepartmentIDs: make([]string, 0),
		CustomRules:   make(map[string]interface{}),
	}

	// 如果禁用权限检查，返回空范围（表示无限制）
	if !a.config.EnablePermissionCheck {
		return scope, nil
	}

	// 如果是管理员用户，返回空范围（表示无限制）
	if a.isAdminUser(userID) {
		return scope, nil
	}

	// 转换用户ID
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %s", userID)
	}

	// 获取用户权限范围
	logic := cipermission.NewGetUserPermissionScopeLogic(ctx, a.svcCtx)
	resp, err := logic.GetUserPermissionScope(&cmdb.UserPermissionScopeReq{
		UserId: userIDUint,
	})
	if err != nil {
		logx.Errorf("获取用户权限范围失败 [userID=%s]: %v", userID, err)
		return scope, nil
	}

	// 转换允许的类型
	for _, typeID := range resp.AllowedCiTypes {
		scope.AllowedTypes = append(scope.AllowedTypes, strconv.FormatUint(typeID, 10))
	}

	// 转换部门权限
	for _, deptID := range resp.AllowedDepartmentIds {
		scope.DepartmentIDs = append(scope.DepartmentIDs, strconv.FormatUint(deptID, 10))
	}

	// 处理数据权限范围
	if resp.DataScope != nil {
		scope.CustomRules["dataScope"] = *resp.DataScope
	}

	// 处理自定义权限规则
	if resp.CustomRules != nil {
		scope.CustomRules["customRules"] = resp.CustomRules
	}

	return scope, nil
}

// checkDataLevelPermission 检查数据级别权限
func (a *CMDBPermissionAdapter) checkDataLevelPermission(ctx context.Context, userID string, asset *types.Asset) bool {
	// 获取用户上下文信息
	userCtx := a.getUserContext(ctx)

	// 检查租户权限
	if userCtx.TenantID != "" && asset.TenantID != "" && userCtx.TenantID != asset.TenantID {
		return false
	}

	// 检查部门权限
	if userCtx.DepartmentID != "" && asset.DepartmentID != "" {
		// 检查是否在同一部门或子部门
		if !a.isInDepartmentScope(userCtx.DepartmentID, asset.DepartmentID, userCtx.DataScope) {
			return false
		}
	}

	// 检查所有者权限
	if userCtx.DataScope == "self" && asset.OwnerID != "" && asset.OwnerID != userID {
		return false
	}

	return true
}

// filterSensitiveFields 过滤敏感字段
func (a *CMDBPermissionAdapter) filterSensitiveFields(asset *types.Asset) *types.Asset {
	// 检查类型权限配置
	typeConfig, exists := a.config.TypePermissions[asset.TypeID]
	if !exists || len(typeConfig.RestrictedFields) == 0 {
		return asset
	}

	// 复制资产对象
	filteredAsset := *asset
	filteredAsset.Attributes = make(map[string]*types.AssetAttributeValue)

	// 过滤属性
	for fieldName, attrValue := range asset.Attributes {
		if !a.isRestrictedField(fieldName, typeConfig.RestrictedFields) {
			filteredAsset.Attributes[fieldName] = attrValue
		}
	}

	return &filteredAsset
}

// isAdminUser 检查是否为管理员用户
func (a *CMDBPermissionAdapter) isAdminUser(userID string) bool {
	for _, adminID := range a.config.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// isRestrictedField 检查是否为受限制字段
func (a *CMDBPermissionAdapter) isRestrictedField(fieldName string, restrictedFields []string) bool {
	for _, restricted := range restrictedFields {
		if fieldName == restricted {
			return true
		}
	}
	return false
}

// isInDepartmentScope 检查是否在部门权限范围内
func (a *CMDBPermissionAdapter) isInDepartmentScope(userDeptID, assetDeptID, dataScope string) bool {
	switch dataScope {
	case "all":
		return true
	case "dept":
		return userDeptID == assetDeptID
	case "dept_and_sub":
		// 检查是否为同一部门或子部门
		return userDeptID == assetDeptID || strings.HasPrefix(assetDeptID, userDeptID+".")
	case "custom_dept":
		// 自定义部门权限，需要额外的逻辑
		return a.checkCustomDepartmentPermission(userDeptID, assetDeptID)
	default:
		return false
	}
}

// checkCustomDepartmentPermission 检查自定义部门权限
func (a *CMDBPermissionAdapter) checkCustomDepartmentPermission(userDeptID, assetDeptID string) bool {
	// 这里可以实现自定义的部门权限逻辑
	// 例如：查询用户有权限的部门列表
	return userDeptID == assetDeptID
}

// getUserContext 获取用户上下文信息
func (a *CMDBPermissionAdapter) getUserContext(ctx context.Context) *UserContext {
	userCtx := &UserContext{}

	// 从上下文中提取用户信息
	if tenantID := ctx.Value("tenantId"); tenantID != nil {
		if tid, ok := tenantID.(string); ok {
			userCtx.TenantID = tid
		}
	}

	if deptID := ctx.Value("departmentId"); deptID != nil {
		if did, ok := deptID.(string); ok {
			userCtx.DepartmentID = did
		}
	}

	if dataScope := ctx.Value("dataScope"); dataScope != nil {
		if ds, ok := dataScope.(string); ok {
			userCtx.DataScope = ds
		}
	}

	return userCtx
}

// UserContext 用户上下文信息
type UserContext struct {
	TenantID     string `json:"tenantId"`
	DepartmentID string `json:"departmentId"`
	DataScope    string `json:"dataScope"`
}

// 默认权限配置
func getDefaultCMDBPermissionConfig() *CMDBPermissionConfig {
	return &CMDBPermissionConfig{
		EnablePermissionCheck: true,
		DefaultPolicy:         "deny", // 默认拒绝访问
		CacheEnabled:          true,
		CacheTTL:              300, // 5分钟
		AdminUserIDs:          make([]string, 0),
		TypePermissions:       make(map[string]*TypePermissionConfig),
	}
}
