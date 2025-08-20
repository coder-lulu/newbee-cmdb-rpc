# CI权限逻辑函数优化方案

> **Version**: 1.0  
> **Last Update**: 2024-12-19  
> **Author**: @逻辑架构团队  

## 目录
- [优化背景](#优化背景)
- [原有逻辑分析](#原有逻辑分析)
- [优化策略](#优化策略)
- [核心优化逻辑](#核心优化逻辑)
- [性能优化措施](#性能优化措施)
- [缓存策略](#缓存策略)
- [实施方案](#实施方案)

## 优化背景

基于数据库表结构优化，现需要对权限逻辑函数进行相应优化，以充分发挥新表结构的性能优势。

### 优化目标
1. **性能提升**: 权限检查响应时间从50ms降低到5ms以内
2. **缓存命中率**: 从60%提升到90%以上
3. **代码复杂度**: 降低权限逻辑的复杂度，提升可维护性
4. **扩展性**: 支持新的权限类型和业务场景

## 原有逻辑分析

### 现有问题
1. **JSON字段查询性能差**: 大量使用JSON_EXTRACT导致全表扫描
2. **权限计算复杂**: 每次都需要实时计算权限继承和合并
3. **缓存策略不当**: 缓存粒度过大，失效频繁
4. **代码重复**: 权限检查逻辑分散在各个业务模块

### 现有权限检查流程
```go
// 原有复杂的权限检查逻辑
func (l *Logic) CheckPermission(userID string, resourceType string, operation string) (bool, error) {
    // 1. 查询用户所有权限（包含JSON字段）
    permissions, err := l.queryUserPermissions(userID)
    if err != nil {
        return false, err
    }
    
    // 2. 实时计算权限继承
    effectivePerms := l.calculateInheritedPermissions(permissions)
    
    // 3. 遍历JSON字段检查操作权限
    for _, perm := range effectivePerms {
        operations := perm.Operations // JSON字段
        if l.checkOperationInJSON(operations, operation) {
            return true, nil
        }
    }
    
    return false, nil
}
```

## 优化策略

### 1. 权限预计算策略
- **用户权限预计算**: 定期计算用户的有效权限并缓存
- **位掩码优化**: 使用位运算替代JSON遍历
- **分层缓存**: 多层缓存减少数据库查询

### 2. 查询优化策略
- **索引利用**: 充分利用新建的复合索引
- **关联查询优化**: 减少N+1查询问题
- **分页查询**: 大数据量场景下的分页处理

### 3. 缓存优化策略
- **细粒度缓存**: 按用户+资源类型+资源ID缓存
- **缓存预热**: 系统启动时预热核心用户权限
- **增量更新**: 权限变更时增量更新缓存

## 核心优化逻辑

### 1. 权限服务接口定义

```go
// PermissionService 权限服务接口
type PermissionService interface {
    // 核心权限检查方法
    CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionResult, error)
    
    // 批量权限检查
    CheckBatchPermissions(ctx context.Context, reqs []*PermissionCheckRequest) ([]*PermissionResult, error)
    
    // 获取用户有效权限
    GetUserEffectivePermissions(ctx context.Context, userID string) ([]*EffectivePermission, error)
    
    // 权限预计算
    PreCalculateUserPermissions(ctx context.Context, userID string) error
    
    // 获取数据过滤条件
    GetDataFilters(ctx context.Context, userID, resourceType string) (*DataFilterCondition, error)
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
    UserID       string `json:"user_id"`
    ResourceType string `json:"resource_type"`
    ResourceID   string `json:"resource_id"`
    Operation    string `json:"operation"`
    Context      map[string]interface{} `json:"context,omitempty"`
}

// PermissionResult 权限检查结果
type PermissionResult struct {
    Allowed      bool                   `json:"allowed"`
    Reason       string                 `json:"reason,omitempty"`
    DataFilters  *DataFilterCondition   `json:"data_filters,omitempty"`
    FieldMasks   []string               `json:"field_masks,omitempty"`
    CacheHit     bool                   `json:"cache_hit"`
    Duration     time.Duration          `json:"duration"`
}

// EffectivePermission 有效权限
type EffectivePermission struct {
    ResourceType     string    `json:"resource_type"`
    ResourceID       string    `json:"resource_id"`
    AllowedOps       uint64    `json:"allowed_ops"`
    PermissionLevel  string    `json:"permission_level"`
    HasDataFilters   bool      `json:"has_data_filters"`
    HasFieldMasks    bool      `json:"has_field_masks"`
    ExpiresAt        time.Time `json:"expires_at"`
}
```

### 2. 优化后的权限检查逻辑

```go
// PermissionServiceImpl 权限服务实现
type PermissionServiceImpl struct {
    db       *ent.Client
    cache    PermissionCache
    logger   *logrus.Logger
    metrics  PermissionMetrics
}

// CheckPermission 优化后的权限检查主方法
func (s *PermissionServiceImpl) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionResult, error) {
    startTime := time.Now()
    result := &PermissionResult{}
    
    // 1. 快速缓存检查
    if cachedResult := s.checkCache(ctx, req); cachedResult != nil {
        result = cachedResult
        result.CacheHit = true
        result.Duration = time.Since(startTime)
        s.metrics.RecordCacheHit()
        return result, nil
    }
    
    // 2. 位掩码快速检查
    allowed, err := s.checkWithBitmask(ctx, req)
    if err != nil {
        return nil, err
    }
    
    result.Allowed = allowed
    result.Duration = time.Since(startTime)
    
    // 3. 如果允许，获取数据过滤和字段掩码
    if allowed {
        result.DataFilters, err = s.getDataFilters(ctx, req.UserID, req.ResourceType)
        if err != nil {
            s.logger.Errorf("获取数据过滤失败: %v", err)
        }
        
        result.FieldMasks, err = s.getFieldMasks(ctx, req.UserID, req.ResourceType)
        if err != nil {
            s.logger.Errorf("获取字段掩码失败: %v", err)
        }
    }
    
    // 4. 更新缓存
    s.updateCache(ctx, req, result)
    
    // 5. 记录审计日志
    s.recordAuditLog(ctx, req, result)
    
    return result, nil
}

// checkWithBitmask 使用位掩码进行快速权限检查
func (s *PermissionServiceImpl) checkWithBitmask(ctx context.Context, req *PermissionCheckRequest) (bool, error) {
    // 操作位掩码映射
    operationMask := s.getOperationMask(req.Operation)
    if operationMask == 0 {
        return false, fmt.Errorf("无效的操作: %s", req.Operation)
    }
    
    // 查询用户对指定资源的权限掩码
    query := s.db.CiPermissions.Query().
        Where(
            cipermissions.SubjectIDEQ(req.UserID),
            cipermissions.ScopeTargetTypeEQ(req.ResourceType),
            cipermissions.ScopeTargetIDEQ(req.ResourceID),
            cipermissions.StatusEQ(cipermissions.StatusActive),
            cipermissions.PermissionTypeEQ(cipermissions.PermissionTypeAllow),
        )
    
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
    
    permissions, err := query.All(ctx)
    if err != nil {
        return false, err
    }
    
    // 计算最终权限掩码
    var finalMask uint64
    for _, perm := range permissions {
        finalMask |= perm.OperationsMask
    }
    
    // 检查是否有权限
    return (finalMask & operationMask) != 0, nil
}

// getOperationMask 获取操作对应的位掩码
func (s *PermissionServiceImpl) getOperationMask(operation string) uint64 {
    operationMasks := map[string]uint64{
        "read":    1,  // 0001
        "write":   2,  // 0010
        "delete":  4,  // 0100
        "approve": 8,  // 1000
        "admin":   16, // 10000
    }
    return operationMasks[operation]
}
```

### 3. 权限预计算逻辑

```go
// PreCalculateUserPermissions 权限预计算主方法
func (s *PermissionServiceImpl) PreCalculateUserPermissions(ctx context.Context, userID string) error {
    s.logger.Infof("开始预计算用户权限: %s", userID)
    
    // 1. 清理旧缓存
    err := s.cleanOldCache(ctx, userID)
    if err != nil {
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

// getUserAllPermissions 获取用户所有有效权限
func (s *PermissionServiceImpl) getUserAllPermissions(ctx context.Context, userID string) ([]*ent.CiPermissions, error) {
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
func (s *PermissionServiceImpl) groupPermissionsByResource(permissions []*ent.CiPermissions) map[string]*ResourcePermissionGroup {
    groups := make(map[string]*ResourcePermissionGroup)
    
    for _, perm := range permissions {
        key := fmt.Sprintf("%s:%s", perm.ScopeTargetType, perm.ScopeTargetID)
        
        if group, exists := groups[key]; exists {
            group.Permissions = append(group.Permissions, perm)
            group.CombinedMask |= perm.OperationsMask
        } else {
            groups[key] = &ResourcePermissionGroup{
                ResourceType: perm.ScopeTargetType,
                ResourceID:   perm.ScopeTargetID,
                Permissions:  []*ent.CiPermissions{perm},
                CombinedMask: perm.OperationsMask,
            }
        }
    }
    
    return groups
}
```

### 4. 数据过滤逻辑优化

```go
// getDataFilters 获取数据过滤条件
func (s *PermissionServiceImpl) getDataFilters(ctx context.Context, userID, resourceType string) (*DataFilterCondition, error) {
    // 1. 查询用户对该资源类型的数据过滤规则
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
        return nil, err
    }
    
    // 2. 按过滤组分组
    filterGroups := s.groupFiltersByGroup(filters)
    
    // 3. 构建SQL WHERE条件
    return s.buildSQLCondition(filterGroups), nil
}

// buildSQLCondition 构建SQL WHERE条件
func (s *PermissionServiceImpl) buildSQLCondition(groups map[int][]*ent.PermissionDataFilters) *DataFilterCondition {
    if len(groups) == 0 {
        return &DataFilterCondition{}
    }
    
    var conditions []string
    var params []interface{}
    
    for groupID, filters := range groups {
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
func (s *PermissionServiceImpl) buildSingleCondition(filter *ent.PermissionDataFilters) (string, interface{}) {
    switch filter.OperatorType {
    case permissiondatafilters.OperatorTypeEq:
        return fmt.Sprintf("%s = ?", filter.FieldName), filter.FilterValue
    case permissiondatafilters.OperatorTypeIn:
        return fmt.Sprintf("%s IN (?)", filter.FieldName), filter.FilterValue
    case permissiondatafilters.OperatorTypeLike:
        return fmt.Sprintf("%s LIKE ?", filter.FieldName), fmt.Sprintf("%%%s%%", filter.FilterValue)
    case permissiondatafilters.OperatorTypeGt:
        return fmt.Sprintf("%s > ?", filter.FieldName), filter.FilterValue
    case permissiondatafilters.OperatorTypeLt:
        return fmt.Sprintf("%s < ?", filter.FieldName), filter.FilterValue
    default:
        return "", nil
    }
}
```

## 性能优化措施

### 1. 查询优化

```go
// 批量权限检查优化
func (s *PermissionServiceImpl) CheckBatchPermissions(ctx context.Context, reqs []*PermissionCheckRequest) ([]*PermissionResult, error) {
    if len(reqs) == 0 {
        return nil, nil
    }
    
    // 1. 缓存批量检查
    results := make([]*PermissionResult, len(reqs))
    uncachedReqs := make([]*PermissionCheckRequest, 0)
    uncachedIndexes := make([]int, 0)
    
    for i, req := range reqs {
        if cached := s.checkCache(ctx, req); cached != nil {
            results[i] = cached
            results[i].CacheHit = true
        } else {
            uncachedReqs = append(uncachedReqs, req)
            uncachedIndexes = append(uncachedIndexes, i)
        }
    }
    
    if len(uncachedReqs) == 0 {
        return results, nil
    }
    
    // 2. 批量数据库查询
    batchResults, err := s.batchCheckPermissions(ctx, uncachedReqs)
    if err != nil {
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
func (s *PermissionServiceImpl) batchCheckPermissions(ctx context.Context, reqs []*PermissionCheckRequest) ([]*PermissionResult, error) {
    // 构建批量查询条件
    userIDs := make([]string, 0, len(reqs))
    resourceKeys := make(map[string]bool)
    
    for _, req := range reqs {
        userIDs = append(userIDs, req.UserID)
        key := fmt.Sprintf("%s:%s", req.ResourceType, req.ResourceID)
        resourceKeys[key] = true
    }
    
    // 批量查询权限
    permissions, err := s.db.CiPermissions.Query().
        Where(
            cipermissions.SubjectIDIn(userIDs...),
            cipermissions.StatusEQ(cipermissions.StatusActive),
            cipermissions.PermissionTypeEQ(cipermissions.PermissionTypeAllow),
        ).
        All(ctx)
    
    if err != nil {
        return nil, err
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
```

### 2. 缓存策略优化

```go
// PermissionCache 权限缓存接口
type PermissionCache interface {
    Get(ctx context.Context, key string) (*PermissionResult, error)
    Set(ctx context.Context, key string, result *PermissionResult, ttl time.Duration) error
    Delete(ctx context.Context, pattern string) error
    GetUserPermissions(ctx context.Context, userID string) ([]*EffectivePermission, error)
    SetUserPermissions(ctx context.Context, userID string, perms []*EffectivePermission, ttl time.Duration) error
}

// RedisPermissionCache Redis权限缓存实现
type RedisPermissionCache struct {
    client redis.UniversalClient
    prefix string
}

// Get 获取权限检查结果缓存
func (c *RedisPermissionCache) Get(ctx context.Context, key string) (*PermissionResult, error) {
    fullKey := fmt.Sprintf("%s:perm:%s", c.prefix, key)
    
    data, err := c.client.Get(ctx, fullKey).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, nil // 缓存未命中
        }
        return nil, err
    }
    
    var result PermissionResult
    err = json.Unmarshal([]byte(data), &result)
    return &result, err
}

// Set 设置权限检查结果缓存
func (c *RedisPermissionCache) Set(ctx context.Context, key string, result *PermissionResult, ttl time.Duration) error {
    fullKey := fmt.Sprintf("%s:perm:%s", c.prefix, key)
    
    data, err := json.Marshal(result)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, fullKey, data, ttl).Err()
}

// 权限缓存键生成
func (s *PermissionServiceImpl) buildCacheKey(req *PermissionCheckRequest) string {
    return fmt.Sprintf("%s:%s:%s:%s", req.UserID, req.ResourceType, req.ResourceID, req.Operation)
}

// 缓存失效策略
func (s *PermissionServiceImpl) InvalidateUserCache(ctx context.Context, userID string) error {
    pattern := fmt.Sprintf("*:perm:%s:*", userID)
    return s.cache.Delete(ctx, pattern)
}
```

### 3. 监控和指标

```go
// PermissionMetrics 权限检查指标
type PermissionMetrics struct {
    CheckCount     prometheus.Counter
    CacheHitCount  prometheus.Counter
    CacheMissCount prometheus.Counter
    ErrorCount     prometheus.Counter
    Duration       prometheus.Histogram
}

// 指标初始化
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
```

## 缓存策略

### 1. 多层缓存架构

```go
// MultiLevelCache 多层缓存
type MultiLevelCache struct {
    l1Cache *ristretto.Cache    // 本地缓存
    l2Cache redis.UniversalClient // Redis缓存
    l3Cache *ent.Client         // 数据库预计算缓存
}

// 缓存策略配置
type CacheConfig struct {
    L1TTL time.Duration // 本地缓存TTL: 5分钟
    L2TTL time.Duration // Redis缓存TTL: 30分钟  
    L3TTL time.Duration // DB缓存TTL: 2小时
}
```

### 2. 缓存预热策略

```go
// WarmupCache 缓存预热
func (s *PermissionServiceImpl) WarmupCache(ctx context.Context) error {
    // 1. 预热核心用户权限
    coreUsers, err := s.getCoreUsers(ctx)
    if err != nil {
        return err
    }
    
    // 2. 并发预热
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // 限制并发数
    
    for _, userID := range coreUsers {
        wg.Add(1)
        go func(uid string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            err := s.PreCalculateUserPermissions(ctx, uid)
            if err != nil {
                s.logger.Errorf("预热用户权限失败 %s: %v", uid, err)
            }
        }(userID)
    }
    
    wg.Wait()
    return nil
}
```

## 实施方案

### 1. 分阶段实施计划

#### 阶段1：核心优化（1-2周）
- [ ] 实现权限服务基础接口
- [ ] 完成位掩码权限检查逻辑
- [ ] 部署Redis缓存层
- [ ] 基础性能监控

#### 阶段2：缓存优化（2-3周）
- [ ] 实现多层缓存架构
- [ ] 权限预计算逻辑
- [ ] 缓存预热和失效策略
- [ ] 批量权限检查

#### 阶段3：高级功能（3-4周）
- [ ] 数据过滤优化
- [ ] 字段掩码处理
- [ ] 权限审计日志
- [ ] 性能调优

### 2. 性能目标

| 指标 | 当前值 | 目标值 | 验证方法 |
|------|--------|--------|----------|
| 权限检查响应时间 | 50ms | <5ms | 压力测试 |
| 缓存命中率 | 60% | >90% | 监控统计 |
| 数据库查询次数 | 每次检查3-5次 | 每次检查<1次 | SQL监控 |
| 并发处理能力 | 500 QPS | >2000 QPS | 负载测试 |

### 3. 监控和告警

```go
// 关键指标监控
func (s *PermissionServiceImpl) setupMonitoring() {
    // 权限检查延迟告警
    if s.avgResponseTime > 10*time.Millisecond {
        s.alertManager.SendAlert("权限检查延迟过高")
    }
    
    // 缓存命中率告警
    if s.cacheHitRate < 0.8 {
        s.alertManager.SendAlert("权限缓存命中率过低")
    }
    
    // 错误率告警
    if s.errorRate > 0.01 {
        s.alertManager.SendAlert("权限检查错误率过高")
    }
}
```

### 4. 回滚方案

```go
// 特性开关控制
type PermissionFeatureFlags struct {
    EnableNewPermissionLogic bool
    EnablePermissionCache    bool
    EnableBatchCheck        bool
    FallbackToOldLogic      bool
}

// 逐步迁移策略
func (s *PermissionServiceImpl) CheckPermissionWithFallback(ctx context.Context, req *PermissionCheckRequest) (*PermissionResult, error) {
    if s.featureFlags.EnableNewPermissionLogic {
        result, err := s.CheckPermission(ctx, req)
        if err != nil && s.featureFlags.FallbackToOldLogic {
            // 降级到旧逻辑
            return s.checkPermissionOld(ctx, req)
        }
        return result, err
    }
    
    return s.checkPermissionOld(ctx, req)
}
```

## 总结

通过以上优化方案，我们将实现：

1. **性能提升60-80%**: 通过位掩码、索引优化和多层缓存
2. **代码复杂度降低50%**: 通过服务化和标准化接口
3. **扩展性增强**: 支持新的权限类型和业务场景
4. **可维护性提升**: 清晰的模块划分和监控体系

该方案采用渐进式优化策略，确保业务稳定性的同时逐步提升性能。

---

*此优化方案基于数据库表结构优化设计，建议与数据库优化方案同步实施以获得最佳效果。* 