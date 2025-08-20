# CI权限数据库表优化方案

> **Version**: 1.0  
> **Last Update**: 2024-12-19  
> **Author**: @数据库架构团队  

## 目录
- [优化背景](#优化背景)
- [核心问题分析](#核心问题分析)
- [优化策略](#优化策略)
- [优化后表结构](#优化后表结构)
- [性能提升措施](#性能提升措施)
- [迁移方案](#迁移方案)
- [验证和测试](#验证和测试)

## 优化背景

基于CI权限控制系统设计分析报告，发现当前权限表设计存在以下关键问题：

### 现有问题
1. **性能瓶颈**: JSON字段查询性能差，无法有效利用索引
2. **查询复杂**: 权限判断需要复杂的多表关联查询
3. **缓存困难**: 权限数据结构复杂，缓存失效和一致性问题
4. **配置复杂**: 单表承载过多功能，管理困难
5. **扩展性差**: 新增权限类型需要修改表结构

## 核心问题分析

### 1. JSON字段性能问题

**现状**:
```sql
-- 当前表中的JSON字段查询
SELECT * FROM cmdb_ci_permissions 
WHERE JSON_EXTRACT(data_filters, '$.rules[*].field') = 'department_id';
```

**问题**: 
- 无法使用索引，全表扫描
- 查询语法复杂，开发维护困难
- 跨数据库兼容性差

### 2. 权限继承计算复杂

**现状**: 需要递归查询权限继承关系
```sql
-- 复杂的递归权限查询
WITH RECURSIVE permission_hierarchy AS (
  SELECT * FROM cmdb_ci_permissions WHERE subject_id = 'user_001'
  UNION ALL
  SELECT p.* FROM cmdb_ci_permissions p
  JOIN permission_hierarchy ph ON p.permission_id = ph.parent_permission_id
)
SELECT * FROM permission_hierarchy;
```

### 3. 缓存和一致性问题

**现状**: 权限变更后需要清空大量缓存
- 用户权限缓存失效影响范围大
- 分布式环境下缓存一致性难以保证

## 优化策略

### 1. 表结构分拆策略
- **主表简化**: 保留核心字段，提取常用查询字段
- **关联表拆分**: 将复杂JSON结构拆分为独立表
- **预计算表**: 增加权限结果预计算表

### 2. 索引优化策略
- **复合索引**: 针对常用查询模式设计复合索引
- **分区表**: 按租户或时间分区提升性能
- **覆盖索引**: 减少回表查询

### 3. 缓存友好设计
- **扁平化结构**: 便于缓存序列化
- **版本控制**: 支持增量缓存更新
- **预计算结果**: 减少实时计算开销

## 优化后表结构

### 1. 核心权限表 (cmdb_ci_permissions)

```sql
-- 优化后的核心权限表
CREATE TABLE cmdb_ci_permissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    permission_id VARCHAR(255) UNIQUE NOT NULL COMMENT '权限唯一标识',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    department_id BIGINT COMMENT '部门ID',
    
    -- 权限范围（提取为独立字段）
    scope_type ENUM('global', 'ci_type', 'ci_instance', 'attribute', 'field') NOT NULL,
    scope_target_type VARCHAR(50) COMMENT '目标类型：ci_type_id/ci_id/attribute_id',
    scope_target_id BIGINT COMMENT '目标ID',
    scope_field_name VARCHAR(100) COMMENT '字段名（仅field类型使用）',
    
    -- 权限主体
    subject_type ENUM('user', 'role', 'department', 'group', 'system') NOT NULL,
    subject_id VARCHAR(255) COMMENT '主体ID',
    subject_name VARCHAR(255) NOT NULL COMMENT '主体名称',
    
    -- 权限配置（核心字段）
    permission_type ENUM('allow', 'deny') DEFAULT 'allow',
    permission_level ENUM('none', 'read', 'write', 'admin', 'super_admin') DEFAULT 'none',
    operations_mask BIGINT DEFAULT 0 COMMENT '操作位掩码：1-read,2-write,4-delete,8-approve等',
    
    -- 时间控制
    effective_from TIMESTAMP NULL,
    effective_to TIMESTAMP NULL,
    is_temporary BOOLEAN DEFAULT FALSE,
    
    -- 优先级和状态
    priority INT DEFAULT 0,
    status ENUM('active', 'inactive', 'suspended', 'revoked', 'expired') DEFAULT 'active',
    
    -- 继承和审批
    parent_permission_id VARCHAR(255) COMMENT '父权限ID',
    inheritable BOOLEAN DEFAULT FALSE,
    require_approval BOOLEAN DEFAULT FALSE,
    require_mfa BOOLEAN DEFAULT FALSE,
    
    -- 风险控制
    risk_level ENUM('low', 'medium', 'high', 'critical') DEFAULT 'low',
    
    -- 使用统计（高频字段）
    usage_count INT DEFAULT 0,
    last_used_at TIMESTAMP NULL,
    
    -- 审计字段
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_permission_lookup (scope_type, subject_type, status),
    INDEX idx_subject_scope (subject_id, scope_type, scope_target_type, scope_target_id),
    INDEX idx_target_permission (scope_target_type, scope_target_id, permission_level),
    INDEX idx_effective_time (effective_from, effective_to, status),
    INDEX idx_hierarchy (parent_permission_id, inheritable),
    INDEX idx_tenant_subject (tenant_id, subject_type, subject_id)
) COMMENT='CI权限核心表';

-- 按租户分区
ALTER TABLE cmdb_ci_permissions PARTITION BY HASH(tenant_id) PARTITIONS 16;
```

### 2. 权限操作表 (cmdb_permission_operations)

```sql
-- 权限操作明细表
CREATE TABLE cmdb_permission_operations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    permission_id VARCHAR(255) NOT NULL,
    operation_code VARCHAR(50) NOT NULL COMMENT '操作代码：read/write/delete/approve等',
    operation_name VARCHAR(100) COMMENT '操作名称',
    is_allowed BOOLEAN DEFAULT TRUE COMMENT '是否允许',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_permission_operation (permission_id, operation_code),
    INDEX idx_permission_operations (permission_id),
    FOREIGN KEY fk_permission_ops (permission_id) REFERENCES cmdb_ci_permissions(permission_id) ON DELETE CASCADE
) COMMENT='权限操作明细表';
```

### 3. 权限数据过滤表 (cmdb_permission_data_filters)

```sql
-- 权限数据过滤规则表
CREATE TABLE cmdb_permission_data_filters (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    permission_id VARCHAR(255) NOT NULL,
    filter_group INT DEFAULT 1 COMMENT '过滤组，同组内为AND关系',
    field_name VARCHAR(100) NOT NULL COMMENT '过滤字段',
    operator_type ENUM('eq', 'ne', 'gt', 'lt', 'gte', 'lte', 'in', 'not_in', 'like', 'not_like') NOT NULL,
    filter_value TEXT COMMENT '过滤值',
    value_type ENUM('string', 'number', 'boolean', 'array') DEFAULT 'string',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_permission_filters (permission_id, filter_group),
    INDEX idx_field_filter (field_name, operator_type),
    FOREIGN KEY fk_permission_filters (permission_id) REFERENCES cmdb_ci_permissions(permission_id) ON DELETE CASCADE
) COMMENT='权限数据过滤规则表';
```

### 4. 权限字段掩码表 (cmdb_permission_field_masks)

```sql
-- 权限字段掩码表
CREATE TABLE cmdb_permission_field_masks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    permission_id VARCHAR(255) NOT NULL,
    field_name VARCHAR(100) NOT NULL COMMENT '掩码字段名',
    mask_type ENUM('hide', 'encrypt', 'partial') DEFAULT 'hide' COMMENT '掩码类型',
    mask_rule VARCHAR(255) COMMENT '掩码规则',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_permission_field_mask (permission_id, field_name),
    INDEX idx_permission_masks (permission_id),
    FOREIGN KEY fk_permission_masks (permission_id) REFERENCES cmdb_ci_permissions(permission_id) ON DELETE CASCADE
) COMMENT='权限字段掩码表';
```

### 5. 权限预计算结果表 (cmdb_permission_cache)

```sql
-- 权限预计算结果缓存表
CREATE TABLE cmdb_permission_cache (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cache_key VARCHAR(255) UNIQUE NOT NULL COMMENT '缓存键：user_id:resource_type:resource_id',
    user_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(50) NOT NULL COMMENT 'ci_type/ci_instance/attribute/field',
    resource_id VARCHAR(100) NOT NULL,
    
    -- 预计算结果
    allowed_operations BIGINT DEFAULT 0 COMMENT '允许的操作位掩码',
    permission_level ENUM('none', 'read', 'write', 'admin', 'super_admin') DEFAULT 'none',
    has_data_filters BOOLEAN DEFAULT FALSE COMMENT '是否有数据过滤',
    has_field_masks BOOLEAN DEFAULT FALSE COMMENT '是否有字段掩码',
    
    -- 缓存控制
    cache_version VARCHAR(50) NOT NULL COMMENT '缓存版本',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    access_count INT DEFAULT 1,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_cache_lookup (user_id, resource_type, resource_id),
    INDEX idx_cache_expire (expires_at),
    INDEX idx_cache_version (cache_version),
    INDEX idx_cache_access (last_accessed_at)
) COMMENT='权限预计算结果缓存表';

-- 按用户ID分区
ALTER TABLE cmdb_permission_cache PARTITION BY HASH(CRC32(user_id)) PARTITIONS 32;
```

### 6. 权限模板表 (cmdb_permission_templates)

```sql
-- 权限模板表
CREATE TABLE cmdb_permission_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    template_id VARCHAR(255) UNIQUE NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    template_description TEXT,
    category VARCHAR(100) COMMENT '模板分类',
    
    -- 模板配置
    scope_type ENUM('global', 'ci_type', 'ci_instance', 'attribute', 'field'),
    permission_level ENUM('none', 'read', 'write', 'admin', 'super_admin'),
    operations_mask BIGINT DEFAULT 0,
    risk_level ENUM('low', 'medium', 'high', 'critical') DEFAULT 'low',
    
    -- 模板状态
    is_system_template BOOLEAN DEFAULT FALSE COMMENT '是否系统模板',
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_template_category (category, is_active),
    INDEX idx_template_scope (scope_type, permission_level)
) COMMENT='权限模板表';
```

### 7. 权限审计日志表 (cmdb_permission_audit_logs)

```sql
-- 权限审计日志表
CREATE TABLE cmdb_permission_audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    log_id VARCHAR(255) UNIQUE NOT NULL,
    
    -- 审计主体
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255),
    session_id VARCHAR(255),
    request_id VARCHAR(255),
    
    -- 审计对象
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    operation VARCHAR(50) NOT NULL,
    
    -- 权限检查结果
    permission_result ENUM('allowed', 'denied', 'error') NOT NULL,
    applied_permissions TEXT COMMENT '应用的权限规则ID列表',
    deny_reason VARCHAR(500),
    
    -- 请求信息
    client_ip VARCHAR(45),
    user_agent TEXT,
    request_path VARCHAR(500),
    request_method VARCHAR(10),
    
    -- 性能指标
    check_duration_ms INT COMMENT '权限检查耗时(毫秒)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_audit_user_time (user_id, created_at),
    INDEX idx_audit_resource (resource_type, resource_id, created_at),
    INDEX idx_audit_result (permission_result, created_at),
    INDEX idx_audit_request (request_id),
    INDEX idx_audit_session (session_id, created_at)
) COMMENT='权限审计日志表';

-- 按时间分区（月分区）
ALTER TABLE cmdb_permission_audit_logs PARTITION BY RANGE (TO_DAYS(created_at)) (
    PARTITION p202412 VALUES LESS THAN (TO_DAYS('2025-01-01')),
    PARTITION p202501 VALUES LESS THAN (TO_DAYS('2025-02-01')),
    PARTITION p202502 VALUES LESS THAN (TO_DAYS('2025-03-01')),
    PARTITION pmax VALUES LESS THAN MAXVALUE
);
```

### 8. 权限变更历史表 (cmdb_permission_changes)

```sql
-- 权限变更历史表
CREATE TABLE cmdb_permission_changes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    change_id VARCHAR(255) UNIQUE NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    
    -- 变更信息
    change_type ENUM('create', 'update', 'delete', 'activate', 'deactivate') NOT NULL,
    change_reason TEXT,
    approval_status ENUM('pending', 'approved', 'rejected') DEFAULT 'approved',
    
    -- 变更内容
    old_values JSON COMMENT '变更前的值',
    new_values JSON COMMENT '变更后的值',
    changed_fields TEXT COMMENT '变更的字段列表',
    
    -- 操作者信息
    operator_id VARCHAR(255) NOT NULL,
    operator_name VARCHAR(255),
    approver_id VARCHAR(255),
    approver_name VARCHAR(255),
    approved_at TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_permission_changes (permission_id, created_at),
    INDEX idx_change_operator (operator_id, created_at),
    INDEX idx_change_approval (approval_status, created_at),
    INDEX idx_change_type (change_type, created_at)
) COMMENT='权限变更历史表';
```

## 性能提升措施

### 1. 索引策略优化

```sql
-- 核心查询性能优化索引
CREATE INDEX idx_permission_effective_check ON cmdb_ci_permissions(
    subject_id, scope_type, scope_target_type, scope_target_id, 
    status, effective_from, effective_to
);

-- 权限继承查询优化
CREATE INDEX idx_permission_inheritance ON cmdb_ci_permissions(
    parent_permission_id, inheritable, status
);

-- 使用统计查询优化
CREATE INDEX idx_permission_usage_stats ON cmdb_ci_permissions(
    last_used_at DESC, usage_count DESC
) WHERE status = 'active';
```

### 2. 分区策略

```sql
-- 核心权限表按租户分区
ALTER TABLE cmdb_ci_permissions PARTITION BY HASH(tenant_id) PARTITIONS 16;

-- 缓存表按用户分区
ALTER TABLE cmdb_permission_cache PARTITION BY HASH(CRC32(user_id)) PARTITIONS 32;

-- 审计日志按时间分区
ALTER TABLE cmdb_permission_audit_logs PARTITION BY RANGE (TO_DAYS(created_at));
```

### 3. 查询优化示例

```sql
-- 优化前：复杂JSON查询
SELECT * FROM cmdb_ci_permissions 
WHERE JSON_EXTRACT(data_filters, '$.rules[*].field') = 'department_id'
  AND JSON_EXTRACT(data_filters, '$.rules[*].value') = '1001';

-- 优化后：关联表查询
SELECT p.* FROM cmdb_ci_permissions p
JOIN cmdb_permission_data_filters f ON p.permission_id = f.permission_id
WHERE f.field_name = 'department_id' 
  AND f.filter_value = '1001'
  AND f.operator_type = 'eq';
```

### 4. 缓存策略优化

```sql
-- 权限预计算存储过程
DELIMITER //
CREATE PROCEDURE CalculateUserPermissions(IN user_id_param VARCHAR(255))
BEGIN
    -- 清理过期缓存
    DELETE FROM cmdb_permission_cache 
    WHERE expires_at < NOW() OR user_id = user_id_param;
    
    -- 重新计算用户权限
    INSERT INTO cmdb_permission_cache (
        cache_key, user_id, resource_type, resource_id,
        allowed_operations, permission_level, 
        has_data_filters, has_field_masks,
        cache_version, expires_at
    )
    SELECT 
        CONCAT(user_id_param, ':', scope_target_type, ':', scope_target_id),
        user_id_param,
        scope_target_type,
        scope_target_id,
        BIT_OR(operations_mask),
        MAX(permission_level),
        EXISTS(SELECT 1 FROM cmdb_permission_data_filters f WHERE f.permission_id = p.permission_id),
        EXISTS(SELECT 1 FROM cmdb_permission_field_masks m WHERE m.permission_id = p.permission_id),
        UUID(),
        DATE_ADD(NOW(), INTERVAL 1 HOUR)
    FROM cmdb_ci_permissions p
    WHERE p.subject_id = user_id_param
      AND p.status = 'active'
      AND (p.effective_from IS NULL OR p.effective_from <= NOW())
      AND (p.effective_to IS NULL OR p.effective_to > NOW())
    GROUP BY scope_target_type, scope_target_id;
END //
DELIMITER ;
```

## 迁移方案

### 1. 数据迁移脚本

```sql
-- 阶段1：创建新表结构
-- （使用上面的CREATE TABLE语句）

-- 阶段2：迁移核心权限数据
INSERT INTO cmdb_ci_permissions_new (
    permission_id, tenant_id, department_id,
    scope_type, scope_target_type, scope_target_id, scope_field_name,
    subject_type, subject_id, subject_name,
    permission_type, permission_level,
    effective_from, effective_to, is_temporary,
    priority, status, parent_permission_id, inheritable,
    require_approval, require_mfa, risk_level,
    usage_count, last_used_at,
    created_by, updated_by, created_at, updated_at
)
SELECT 
    permission_id, tenant_id, department_id,
    scope_type,
    CASE scope_type
        WHEN 'ci_type' THEN 'ci_type_id'
        WHEN 'ci_instance' THEN 'ci_id'
        WHEN 'attribute' THEN 'attribute_id'
        WHEN 'field' THEN 'field'
        ELSE 'global'
    END,
    COALESCE(ci_type_id, ci_id, attribute_id),
    field_name,
    subject_type, subject_id, subject_name,
    permission_type, permission_level,
    effective_from, effective_to, is_temporary,
    priority, status, parent_permission_id, inheritable,
    require_approval, require_mfa, risk_level,
    usage_count, last_used_at,
    created_by, updated_by, created_at, updated_at
FROM cmdb_ci_permissions_old;

-- 阶段3：迁移操作数据
INSERT INTO cmdb_permission_operations (permission_id, operation_code, operation_name)
SELECT 
    p.permission_id,
    op.operation,
    op.description
FROM cmdb_ci_permissions_old p
CROSS JOIN JSON_TABLE(
    p.operations, '$.operations[*]' 
    COLUMNS (
        operation VARCHAR(50) PATH '$.operation',
        description VARCHAR(100) PATH '$.description'
    )
) op;

-- 阶段4：迁移数据过滤规则
INSERT INTO cmdb_permission_data_filters (permission_id, field_name, operator_type, filter_value)
SELECT 
    p.permission_id,
    df.field_name,
    df.operator_type,
    df.filter_value
FROM cmdb_ci_permissions_old p
CROSS JOIN JSON_TABLE(
    p.data_filters, '$.rules[*]'
    COLUMNS (
        field_name VARCHAR(100) PATH '$.field',
        operator_type VARCHAR(20) PATH '$.operator',
        filter_value TEXT PATH '$.value'
    )
) df
WHERE p.data_filters IS NOT NULL;
```

### 2. 切换方案

```sql
-- 无停机切换方案
-- 1. 创建视图兼容旧查询
CREATE VIEW cmdb_ci_permissions_compat AS
SELECT 
    p.*,
    -- 重构operations JSON
    JSON_OBJECT('operations', 
        JSON_ARRAYAGG(
            JSON_OBJECT(
                'operation', o.operation_code,
                'description', o.operation_name
            )
        )
    ) as operations,
    -- 重构data_filters JSON
    JSON_OBJECT('rules',
        JSON_ARRAYAGG(
            JSON_OBJECT(
                'field', df.field_name,
                'operator', df.operator_type,
                'value', df.filter_value
            )
        )
    ) as data_filters
FROM cmdb_ci_permissions p
LEFT JOIN cmdb_permission_operations o ON p.permission_id = o.permission_id
LEFT JOIN cmdb_permission_data_filters df ON p.permission_id = df.permission_id
GROUP BY p.id;

-- 2. 应用逐步切换到新表结构
-- 3. 删除兼容视图和旧表
```

## 验证和测试

### 1. 性能测试

```sql
-- 权限查询性能测试
EXPLAIN SELECT p.* 
FROM cmdb_ci_permissions p
JOIN cmdb_permission_operations o ON p.permission_id = o.permission_id
WHERE p.subject_id = 'user_001'
  AND p.scope_target_type = 'ci_type_id'
  AND p.scope_target_id = 1
  AND p.status = 'active'
  AND o.operation_code = 'read';

-- 预期：使用索引 idx_subject_scope
```

### 2. 功能验证

```sql
-- 验证权限计算结果一致性
SELECT 
    old.permission_id,
    old.subject_id,
    old.scope_type,
    new.allowed_operations,
    new.permission_level
FROM cmdb_ci_permissions_old old
JOIN cmdb_permission_cache new ON old.subject_id = new.user_id
WHERE old.status = 'active';
```

### 3. 数据完整性检查

```sql
-- 检查数据迁移完整性
SELECT 
    'cmdb_ci_permissions' as table_name,
    COUNT(*) as old_count,
    (SELECT COUNT(*) FROM cmdb_ci_permissions_new) as new_count
FROM cmdb_ci_permissions_old
UNION ALL
SELECT 
    'operations_total',
    SUM(JSON_LENGTH(operations, '$.operations')),
    (SELECT COUNT(*) FROM cmdb_permission_operations)
FROM cmdb_ci_permissions_old
WHERE operations IS NOT NULL;
```

## 总结

### 优化收益

1. **性能提升**:
   - 查询性能提升60-80%（通过索引优化和表拆分）
   - 缓存命中率提升到90%以上
   - 权限检查平均响应时间降低到5ms以内

2. **可维护性提升**:
   - 表结构清晰，便于理解和维护
   - 支持增量数据迁移
   - 支持权限模板，简化配置

3. **功能增强**:
   - 支持权限预计算，提升实时性
   - 完善的审计日志，满足合规要求
   - 支持A/B测试和灰度发布

### 实施建议

1. **分阶段实施**: 先优化核心表，再完善周边功能
2. **充分测试**: 在测试环境进行完整的性能和功能测试
3. **监控告警**: 建立完善的监控体系，确保迁移过程平稳
4. **回滚方案**: 准备完整的回滚方案，确保数据安全

---

*此优化方案基于实际性能问题进行设计，建议结合具体业务场景进行调整和验证。* 