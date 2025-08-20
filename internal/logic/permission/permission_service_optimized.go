package permission

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"cmdb-rpc/ent"
	"cmdb-rpc/ent/cipermissions"
	"cmdb-rpc/ent/permissiondatafilters"
	"cmdb-rpc/ent/permissionfieldmasks"
)

// PermissionServiceOptimized 优化后的权限服务实现
type PermissionServiceOptimized struct {
	db      *ent.Client
	cache   PermissionCache
	logger  *logrus.Logger
	metrics *PermissionMetrics
	config  *PermissionConfig

	// 操作位掩码映射
	operationMasks map[string]uint64

	// 特性开关
	featureFlags *PermissionFeatureFlags
}

// PermissionConfig 权限配置
type PermissionConfig struct {
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	PreCalculateTTL time.Duration `json:"pre_calculate_ttl"`
	MaxBatchSize    int           `json:"max_batch_size"`
	EnableAuditLog  bool          `json:"enable_audit_log"`
	EnableMetrics   bool          `json:"enable_metrics"`
}

// PermissionFeatureFlags 特性开关
type PermissionFeatureFlags struct {
	EnableNewPermissionLogic bool `json:"enable_new_permission_logic"`
	EnablePermissionCache    bool `json:"enable_permission_cache"`
	EnableBatchCheck         bool `json:"enable_batch_check"`
	FallbackToOldLogic       bool `json:"fallback_to_old_logic"`
}

// NewPermissionServiceOptimized 创建优化后的权限服务
func NewPermissionServiceOptimized(db *ent.Client, cache PermissionCache, logger *logrus.Logger, config *PermissionConfig) *PermissionServiceOptimized {
	service := &PermissionServiceOptimized{
		db:      db,
		cache:   cache,
		logger:  logger,
		config:  config,
		metrics: NewPermissionMetrics(),
		operationMasks: map[string]uint64{
			"read":    1,  // 0001
			"write":   2,  // 0010
			"delete":  4,  // 0100
			"approve": 8,  // 1000
			"admin":   16, // 10000
		},
		featureFlags: &PermissionFeatureFlags{
			EnableNewPermissionLogic: true,
			EnablePermissionCache:    true,
			EnableBatchCheck:         true,
			FallbackToOldLogic:       false,
		},
	}

	// 注册Prometheus指标
	if config.EnableMetrics {
		prometheus.MustRegister(service.metrics.CheckCount)
		prometheus.MustRegister(service.metrics.CacheHitCount)
		prometheus.MustRegister(service.metrics.CacheMissCount)
		prometheus.MustRegister(service.metrics.ErrorCount)
		prometheus.MustRegister(service.metrics.Duration)
	}

	return service
}

// CheckPermission 核心权限检查方法 [[memory:569635]]
func (s *PermissionServiceOptimized) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.metrics.Duration.Observe(duration.Seconds())
		s.metrics.CheckCount.Inc()
	}()

	result := &PermissionResult{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Operation:    req.Operation,
	}

	// 1. 快速缓存检查
	if s.config.CacheEnabled && s.featureFlags.EnablePermissionCache {
		if cachedResult := s.checkCache(ctx, req); cachedResult != nil {
			result = cachedResult
			result.CacheHit = true
			result.Duration = time.Since(startTime)
			s.metrics.CacheHitCount.Inc()
			return result, nil
		}
		s.metrics.CacheMissCount.Inc()
	}

	// 2. 位掩码快速检查
	allowed, err := s.checkWithBitmask(ctx, req)
	if err != nil {
		s.metrics.ErrorCount.Inc()
		s.logger.Errorf("位掩码权限检查失败: %v", err)
		return nil, err
	}

	result.Allowed = allowed
	result.Duration = time.Since(startTime)

	// 3. 如果允许，获取数据过滤和字段掩码
	if allowed {
		// 获取数据过滤条件
		if result.DataFilters, err = s.getDataFilters(ctx, req.UserID, req.ResourceType, req.ResourceID); err != nil {
			s.logger.Errorf("获取数据过滤失败: %v", err)
		}

		// 获取字段掩码
		if result.FieldMasks, err = s.getFieldMasks(ctx, req.UserID, req.ResourceType, req.ResourceID); err != nil {
			s.logger.Errorf("获取字段掩码失败: %v", err)
		}
	} else {
		result.Reason = "权限不足"
	}

	// 4. 更新缓存
	if s.config.CacheEnabled && s.featureFlags.EnablePermissionCache {
		if err := s.updateCache(ctx, req, result); err != nil {
			s.logger.Errorf("更新缓存失败: %v", err)
		}
	}

	// 5. 记录审计日志
	if s.config.EnableAuditLog {
		if err := s.recordAuditLog(ctx, req, result); err != nil {
			s.logger.Errorf("记录审计日志失败: %v", err)
		}
	}

	return result, nil
}

// checkWithBitmask 使用位掩码进行快速权限检查
func (s *PermissionServiceOptimized) checkWithBitmask(ctx context.Context, req *PermissionCheckRequest) (bool, error) {
	// 获取操作位掩码
	operationMask := s.getOperationMask(req.Operation)
	if operationMask == 0 {
		return false, fmt.Errorf("无效的操作: %s", req.Operation)
	}

	// 构建查询条件
	query := s.db.CiPermissions.Query().
		Where(
			cipermissions.SubjectIDEQ(req.UserID),
			cipermissions.ScopeTargetTypeEQ(req.ResourceType),
			cipermissions.StatusEQ(cipermissions.StatusActive),
			cipermissions.PermissionTypeEQ(cipermissions.PermissionTypeAllow),
		)

	// 添加资源ID过滤（如果指定）
	if req.ResourceID != "" {
		resourceIDInt, err := strconv.ParseUint(req.ResourceID, 10, 64)
		if err != nil {
			return false, fmt.Errorf("无效的资源ID: %s", req.ResourceID)
		}
		query = query.Where(cipermissions.ScopeTargetIDEQ(resourceIDInt))
	}

	// 添加时间有效性检查
	now := time.Now()
	query = query.Where(
		cipermissions.Or(
			cipermissions.EffectiveFromIsNil(),
			cipermissions.EffectiveFromLTE(now),
		),
		cipermissions.Or(
			cipermissions.EffectiveToIsNil(),
			cipermissions.EffectiveToGTE(now),
		),
	)

	// 执行查询
	permissions, err := query.All(ctx)
	if err != nil {
		return false, fmt.Errorf("查询权限失败: %w", err)
	}

	// 计算最终权限掩码
	var finalMask uint64
	for _, perm := range permissions {
		finalMask |= perm.OperationsMask
	}

	// 检查是否有权限
	hasPermission := (finalMask & operationMask) != 0

	s.logger.Debugf("权限检查结果: user=%s, resource=%s:%s, operation=%s, mask=%d, allowed=%t",
		req.UserID, req.ResourceType, req.ResourceID, req.Operation, finalMask, hasPermission)

	return hasPermission, nil
}

// getOperationMask 获取操作对应的位掩码
func (s *PermissionServiceOptimized) getOperationMask(operation string) uint64 {
	return s.operationMasks[operation]
}

// getDataFilters 获取数据过滤条件
func (s *PermissionServiceOptimized) getDataFilters(ctx context.Context, userID, resourceType, resourceID string) (*DataFilterCondition, error) {
	// 查询用户对该资源类型的数据过滤规则
	filters, err := s.db.PermissionDataFilters.Query().
		Where(
			permissiondatafilters.HasPermissionWith(
				cipermissions.SubjectIDEQ(userID),
				cipermissions.ScopeTargetTypeEQ(resourceType),
				cipermissions.StatusEQ(cipermissions.StatusActive),
			),
		).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("查询数据过滤规则失败: %w", err)
	}

	if len(filters) == 0 {
		return &DataFilterCondition{}, nil
	}

	// 按过滤组分组
	filterGroups := s.groupFiltersByGroup(filters)

	// 构建SQL WHERE条件
	return s.buildSQLCondition(filterGroups), nil
}

// groupFiltersByGroup 按过滤组分组
func (s *PermissionServiceOptimized) groupFiltersByGroup(filters []*ent.PermissionDataFilters) map[int][]*ent.PermissionDataFilters {
	groups := make(map[int][]*ent.PermissionDataFilters)

	for _, filter := range filters {
		groupID := filter.FilterGroup
		if _, exists := groups[groupID]; !exists {
			groups[groupID] = make([]*ent.PermissionDataFilters, 0)
		}
		groups[groupID] = append(groups[groupID], filter)
	}

	return groups
}

// buildSQLCondition 构建SQL WHERE条件
func (s *PermissionServiceOptimized) buildSQLCondition(groups map[int][]*ent.PermissionDataFilters) *DataFilterCondition {
	if len(groups) == 0 {
		return &DataFilterCondition{}
	}

	var conditions []string
	var params []interface{}

	for _, filters := range groups {
		groupConditions := make([]string, 0, len(filters))

		for _, filter := range filters {
			condition, param := s.buildSingleCondition(filter)
			if condition != "" {
				groupConditions = append(groupConditions, condition)
				if param != nil {
					params = append(params, param)
				}
			}
		}

		if len(groupConditions) > 0 {
			// 同组内用AND连接
			groupCondition := fmt.Sprintf("(%s)", strings.Join(groupConditions, " AND "))
			conditions = append(conditions, groupCondition)
		}
	}

	// 不同组间用OR连接
	finalCondition := strings.Join(conditions, " OR ")

	return &DataFilterCondition{
		WhereClause: finalCondition,
		Parameters:  params,
	}
}

// buildSingleCondition 构建单个过滤条件
func (s *PermissionServiceOptimized) buildSingleCondition(filter *ent.PermissionDataFilters) (string, interface{}) {
	switch filter.OperatorType {
	case permissiondatafilters.OperatorTypeEq:
		return fmt.Sprintf("%s = ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeNe:
		return fmt.Sprintf("%s != ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeGt:
		return fmt.Sprintf("%s > ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeLt:
		return fmt.Sprintf("%s < ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeGte:
		return fmt.Sprintf("%s >= ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeLte:
		return fmt.Sprintf("%s <= ?", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeIn:
		return fmt.Sprintf("%s IN (?)", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeNotIn:
		return fmt.Sprintf("%s NOT IN (?)", filter.FieldName), filter.FilterValue
	case permissiondatafilters.OperatorTypeLike:
		return fmt.Sprintf("%s LIKE ?", filter.FieldName), fmt.Sprintf("%%%s%%", filter.FilterValue)
	case permissiondatafilters.OperatorTypeNotLike:
		return fmt.Sprintf("%s NOT LIKE ?", filter.FieldName), fmt.Sprintf("%%%s%%", filter.FilterValue)
	default:
		return "", nil
	}
}

// getFieldMasks 获取字段掩码
func (s *PermissionServiceOptimized) getFieldMasks(ctx context.Context, userID, resourceType, resourceID string) ([]string, error) {
	// 查询用户对该资源类型的字段掩码
	masks, err := s.db.PermissionFieldMasks.Query().
		Where(
			permissionfieldmasks.HasPermissionWith(
				cipermissions.SubjectIDEQ(userID),
				cipermissions.ScopeTargetTypeEQ(resourceType),
				cipermissions.StatusEQ(cipermissions.StatusActive),
			),
		).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("查询字段掩码失败: %w", err)
	}

	fieldMasks := make([]string, 0, len(masks))
	for _, mask := range masks {
		fieldMasks = append(fieldMasks, mask.FieldName)
	}

	return fieldMasks, nil
}

// CheckBatchPermissions 批量权限检查
func (s *PermissionServiceOptimized) CheckBatchPermissions(ctx context.Context, reqs []*PermissionCheckRequest) ([]*PermissionResult, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	if len(reqs) > s.config.MaxBatchSize {
		return nil, fmt.Errorf("批量检查数量超过限制: %d > %d", len(reqs), s.config.MaxBatchSize)
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.metrics.Duration.Observe(duration.Seconds())
		s.metrics.CheckCount.Add(float64(len(reqs)))
	}()

	// 1. 缓存批量检查
	results := make([]*PermissionResult, len(reqs))
	uncachedReqs := make([]*PermissionCheckRequest, 0)
	uncachedIndexes := make([]int, 0)

	if s.config.CacheEnabled && s.featureFlags.EnablePermissionCache {
		for i, req := range reqs {
			if cached := s.checkCache(ctx, req); cached != nil {
				results[i] = cached
				results[i].CacheHit = true
				s.metrics.CacheHitCount.Inc()
			} else {
				uncachedReqs = append(uncachedReqs, req)
				uncachedIndexes = append(uncachedIndexes, i)
				s.metrics.CacheMissCount.Inc()
			}
		}
	} else {
		uncachedReqs = reqs
		for i := range reqs {
			uncachedIndexes = append(uncachedIndexes, i)
		}
	}

	if len(uncachedReqs) == 0 {
		return results, nil
	}

	// 2. 批量数据库查询
	batchResults, err := s.batchCheckPermissions(ctx, uncachedReqs)
	if err != nil {
		s.metrics.ErrorCount.Inc()
		return nil, err
	}

	// 3. 合并结果
	for i, result := range batchResults {
		originalIndex := uncachedIndexes[i]
		results[originalIndex] = result
	}

	return results, nil
}

// batchCheckPermissions 批量权限检查
func (s *PermissionServiceOptimized) batchCheckPermissions(ctx context.Context, reqs []*PermissionCheckRequest) ([]*PermissionResult, error) {
	// 构建批量查询条件
	userIDs := make([]string, 0, len(reqs))
	resourceTypes := make(map[string]bool)

	for _, req := range reqs {
		userIDs = append(userIDs, req.UserID)
		resourceTypes[req.ResourceType] = true
	}

	// 去重用户ID
	uniqueUserIDs := make([]string, 0, len(userIDs))
	userIDMap := make(map[string]bool)
	for _, userID := range userIDs {
		if !userIDMap[userID] {
			uniqueUserIDs = append(uniqueUserIDs, userID)
			userIDMap[userID] = true
		}
	}

	// 批量查询权限
	now := time.Now()
	permissions, err := s.db.CiPermissions.Query().
		Where(
			cipermissions.SubjectIDIn(uniqueUserIDs...),
			cipermissions.StatusEQ(cipermissions.StatusActive),
			cipermissions.PermissionTypeEQ(cipermissions.PermissionTypeAllow),
			cipermissions.Or(
				cipermissions.EffectiveFromIsNil(),
				cipermissions.EffectiveFromLTE(now),
			),
			cipermissions.Or(
				cipermissions.EffectiveToIsNil(),
				cipermissions.EffectiveToGTE(now),
			),
		).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("批量查询权限失败: %w", err)
	}

	// 构建用户权限映射
	userPermMap := s.buildUserPermissionMap(permissions)

	// 批量计算结果
	results := make([]*PermissionResult, len(reqs))
	for i, req := range reqs {
		results[i] = s.calculatePermissionResult(req, userPermMap[req.UserID])
	}

	return results, nil
}

// buildUserPermissionMap 构建用户权限映射
func (s *PermissionServiceOptimized) buildUserPermissionMap(permissions []*ent.CiPermissions) map[string][]*ent.CiPermissions {
	userPermMap := make(map[string][]*ent.CiPermissions)

	for _, perm := range permissions {
		userID := perm.SubjectID
		if userPermMap[userID] == nil {
			userPermMap[userID] = make([]*ent.CiPermissions, 0)
		}
		userPermMap[userID] = append(userPermMap[userID], perm)
	}

	return userPermMap
}

// calculatePermissionResult 计算权限结果
func (s *PermissionServiceOptimized) calculatePermissionResult(req *PermissionCheckRequest, permissions []*ent.CiPermissions) *PermissionResult {
	operationMask := s.getOperationMask(req.Operation)

	result := &PermissionResult{
		UserID:       req.UserID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Operation:    req.Operation,
		Allowed:      false,
		CacheHit:     false,
	}

	// 计算有效权限掩码
	var effectiveMask uint64
	for _, perm := range permissions {
		// 检查资源类型匹配
		if perm.ScopeTargetType != req.ResourceType {
			continue
		}

		// 检查资源ID匹配（如果指定）
		if req.ResourceID != "" {
			resourceIDInt, err := strconv.ParseUint(req.ResourceID, 10, 64)
			if err != nil {
				continue
			}
			if perm.ScopeTargetID != resourceIDInt {
				continue
			}
		}

		effectiveMask |= perm.OperationsMask
	}

	// 检查是否有权限
	result.Allowed = (effectiveMask & operationMask) != 0

	if !result.Allowed {
		result.Reason = "权限不足"
	}

	return result
}

// PreCalculateUserPermissions 权限预计算
func (s *PermissionServiceOptimized) PreCalculateUserPermissions(ctx context.Context, userID string) error {
	s.logger.Infof("开始预计算用户权限: %s", userID)

	// 1. 清理旧缓存
	if err := s.cleanOldCache(ctx, userID); err != nil {
		return fmt.Errorf("清理旧缓存失败: %w", err)
	}

	// 2. 查询用户所有有效权限
	permissions, err := s.getUserAllPermissions(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户权限失败: %w", err)
	}

	// 3. 按资源类型分组并计算
	resourcePerms := s.groupPermissionsByResource(permissions)

	// 4. 批量插入预计算结果
	return s.batchInsertPermissionCache(ctx, userID, resourcePerms)
}

// cleanOldCache 清理旧缓存
func (s *PermissionServiceOptimized) cleanOldCache(ctx context.Context, userID string) error {
	// 清理Redis缓存
	if s.cache != nil {
		pattern := fmt.Sprintf("*:%s:*", userID)
		if err := s.cache.Delete(ctx, pattern); err != nil {
			return fmt.Errorf("清理Redis缓存失败: %w", err)
		}
	}

	// 清理数据库预计算缓存
	_, err := s.db.PermissionCache.Delete().
		Where(permissioncache.UserIDEQ(userID)).
		Exec(ctx)

	return err
}

// getUserAllPermissions 获取用户所有有效权限
func (s *PermissionServiceOptimized) getUserAllPermissions(ctx context.Context, userID string) ([]*ent.CiPermissions, error) {
	now := time.Now()

	return s.db.CiPermissions.Query().
		Where(
			cipermissions.SubjectIDEQ(userID),
			cipermissions.StatusEQ(cipermissions.StatusActive),
			cipermissions.PermissionTypeEQ(cipermissions.PermissionTypeAllow),
			cipermissions.Or(
				cipermissions.EffectiveFromIsNil(),
				cipermissions.EffectiveFromLTE(now),
			),
			cipermissions.Or(
				cipermissions.EffectiveToIsNil(),
				cipermissions.EffectiveToGTE(now),
			),
		).
		WithOperations().
		WithDataFilters().
		WithFieldMasks().
		All(ctx)
}

// groupPermissionsByResource 按资源分组权限
func (s *PermissionServiceOptimized) groupPermissionsByResource(permissions []*ent.CiPermissions) map[string]*ResourcePermissionGroup {
	groups := make(map[string]*ResourcePermissionGroup)

	for _, perm := range permissions {
		key := fmt.Sprintf("%s:%d", perm.ScopeTargetType, perm.ScopeTargetID)

		if group, exists := groups[key]; exists {
			group.Permissions = append(group.Permissions, perm)
			group.CombinedMask |= perm.OperationsMask
		} else {
			groups[key] = &ResourcePermissionGroup{
				ResourceType: perm.ScopeTargetType,
				ResourceID:   strconv.FormatUint(perm.ScopeTargetID, 10),
				Permissions:  []*ent.CiPermissions{perm},
				CombinedMask: perm.OperationsMask,
			}
		}
	}

	return groups
}

// batchInsertPermissionCache 批量插入权限缓存
func (s *PermissionServiceOptimized) batchInsertPermissionCache(ctx context.Context, userID string, groups map[string]*ResourcePermissionGroup) error {
	now := time.Now()
	expiresAt := now.Add(s.config.PreCalculateTTL)

	// 构建批量插入数据
	creates := make([]*ent.PermissionCacheCreate, 0, len(groups))

	for _, group := range groups {
		cacheKey := fmt.Sprintf("%s:%s:%s", userID, group.ResourceType, group.ResourceID)

		// 检查是否有数据过滤
		hasDataFilters := false
		for _, perm := range group.Permissions {
			if len(perm.Edges.DataFilters) > 0 {
				hasDataFilters = true
				break
			}
		}

		// 检查是否有字段掩码
		hasFieldMasks := false
		for _, perm := range group.Permissions {
			if len(perm.Edges.FieldMasks) > 0 {
				hasFieldMasks = true
				break
			}
		}

		create := s.db.PermissionCache.Create().
			SetCacheKey(cacheKey).
			SetUserID(userID).
			SetResourceType(group.ResourceType).
			SetResourceID(group.ResourceID).
			SetAllowedOperations(group.CombinedMask).
			SetHasDataFilters(hasDataFilters).
			SetHasFieldMasks(hasFieldMasks).
			SetCacheVersion(fmt.Sprintf("%d", now.Unix())).
			SetExpiresAt(expiresAt)

		creates = append(creates, create)
	}

	// 批量插入
	_, err := s.db.PermissionCache.CreateBulk(creates...).Save(ctx)
	return err
}

// 辅助方法：缓存相关
func (s *PermissionServiceOptimized) checkCache(ctx context.Context, req *PermissionCheckRequest) *PermissionResult {
	if s.cache == nil {
		return nil
	}

	key := s.buildCacheKey(req)
	result, err := s.cache.Get(ctx, key)
	if err != nil {
		s.logger.Errorf("获取缓存失败: %v", err)
		return nil
	}

	return result
}

func (s *PermissionServiceOptimized) updateCache(ctx context.Context, req *PermissionCheckRequest, result *PermissionResult) error {
	if s.cache == nil {
		return nil
	}

	key := s.buildCacheKey(req)
	return s.cache.Set(ctx, key, result, s.config.CacheTTL)
}

func (s *PermissionServiceOptimized) buildCacheKey(req *PermissionCheckRequest) string {
	return fmt.Sprintf("%s:%s:%s:%s", req.UserID, req.ResourceType, req.ResourceID, req.Operation)
}

func (s *PermissionServiceOptimized) recordAuditLog(ctx context.Context, req *PermissionCheckRequest, result *PermissionResult) error {
	// 实现审计日志记录逻辑
	s.logger.Infof("权限检查审计: user=%s, resource=%s:%s, operation=%s, allowed=%t, duration=%v",
		req.UserID, req.ResourceType, req.ResourceID, req.Operation, result.Allowed, result.Duration)
	return nil
}

// 类型定义
type PermissionCheckRequest struct {
	UserID       string                 `json:"user_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Operation    string                 `json:"operation"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type PermissionResult struct {
	UserID       string               `json:"user_id"`
	ResourceType string               `json:"resource_type"`
	ResourceID   string               `json:"resource_id"`
	Operation    string               `json:"operation"`
	Allowed      bool                 `json:"allowed"`
	Reason       string               `json:"reason,omitempty"`
	DataFilters  *DataFilterCondition `json:"data_filters,omitempty"`
	FieldMasks   []string             `json:"field_masks,omitempty"`
	CacheHit     bool                 `json:"cache_hit"`
	Duration     time.Duration        `json:"duration"`
}

type DataFilterCondition struct {
	WhereClause string        `json:"where_clause"`
	Parameters  []interface{} `json:"parameters"`
}

type ResourcePermissionGroup struct {
	ResourceType string               `json:"resource_type"`
	ResourceID   string               `json:"resource_id"`
	Permissions  []*ent.CiPermissions `json:"permissions"`
	CombinedMask uint64               `json:"combined_mask"`
}

type PermissionCache interface {
	Get(ctx context.Context, key string) (*PermissionResult, error)
	Set(ctx context.Context, key string, result *PermissionResult, ttl time.Duration) error
	Delete(ctx context.Context, pattern string) error
}

type PermissionMetrics struct {
	CheckCount     prometheus.Counter
	CacheHitCount  prometheus.Counter
	CacheMissCount prometheus.Counter
	ErrorCount     prometheus.Counter
	Duration       prometheus.Histogram
}

func NewPermissionMetrics() *PermissionMetrics {
	return &PermissionMetrics{
		CheckCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "permission_check_total",
			Help: "Total number of permission checks",
		}),
		CacheHitCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "permission_cache_hits_total",
			Help: "Total number of permission cache hits",
		}),
		CacheMissCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "permission_cache_misses_total",
			Help: "Total number of permission cache misses",
		}),
		ErrorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "permission_check_errors_total",
			Help: "Total number of permission check errors",
		}),
		Duration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "permission_check_duration_seconds",
			Help:    "Duration of permission checks",
			Buckets: prometheus.DefBuckets,
		}),
	}
}
