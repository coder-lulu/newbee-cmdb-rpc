# CMDB资产录入数据库设计文档 v2.0

> **版本**: v2.0.0  
> **创建时间**: 2024-12-20  
> **设计原则**: 简化设计，去除过度复杂的字段，专注核心业务需求  
> **变更记录**: 新增CI变更记录表，优化导入相关表结构

## 📊 总体设计概述

本设计包含5个核心表，支持完整的资产导入和CI变更审计功能：

1. **cmdb_import_templates** - 导入模板配置
2. **cmdb_import_tasks** - 导入任务管理  
3. **cmdb_import_records** - 导入记录详情
4. **cmdb_import_errors** - 导入错误记录
5. **cmdb_ci_records** - CI变更审计记录 ⭐ **新增**

### 🎯 设计优化亮点

- **简化字段**: 去除过度设计的字段，专注核心业务需求
- **职责分离**: CI变更记录独立为通用审计机制  
- **验证复用**: 充分利用现有属性验证规则
- **高性能**: 优化的索引策略支持高频查询

## 📋 表结构设计

### 1. 导入模板表 (cmdb_import_templates)

**设计原则**: 保持原有设计，专注于模板配置的核心功能

#### 核心字段（无变化）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `name` | VARCHAR(255) | 模板名称 |
| `code` | VARCHAR(64) | 模板编码（唯一） |
| `type` | ENUM | 模板类型（excel/csv/json/api/xml） |
| `ci_type_id` | BIGINT UNSIGNED | 目标CI类型ID |
| `field_mappings` | JSON | 字段映射配置 |

### 2. 导入任务表 (cmdb_import_tasks) - **简化优化**

**主要变更**: 
- ✅ 去除了 `estimated_completion`、`duration_seconds` 等过度设计字段
- ✅ 简化优先级为 `low/normal/high`
- ✅ 去除通知配置、元数据等复杂字段
- ✅ 保留核心业务字段

#### 简化后核心字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `task_id` | VARCHAR(64) | 任务唯一标识 |
| `name` | VARCHAR(255) | 任务名称 |
| `type` | ENUM | 导入类型（excel/api/csv/json/auto_discovery） |
| `status` | ENUM | 任务状态（pending/processing/completed/failed/cancelled） |
| `priority` | ENUM | 任务优先级（low/normal/high） |
| `source_path` | VARCHAR(500) | 数据源路径 |
| `template_id` | BIGINT UNSIGNED | 关联导入模板ID |
| `mapping_config` | JSON | 字段映射配置 |
| `batch_size` | INT | 批处理大小 |
| `total_count` | INT | 总记录数 |
| `processed_count` | INT | 已处理记录数 |
| `success_count` | INT | 成功记录数 |
| `failed_count` | INT | 失败记录数 |
| `progress_percentage` | FLOAT | 进度百分比 |
| `start_time` | TIMESTAMP | 开始时间 |
| `end_time` | TIMESTAMP | 结束时间 |

### 3. 导入记录表 (cmdb_import_records) - **大幅简化**

**主要变更**:
- ✅ 去除复杂的字段级处理结果、质量评分、业务规则结果
- ✅ 去除审核信息、依赖关系等过度设计
- ✅ 保留核心的数据处理和错误记录功能

#### 简化后核心字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `task_id` | BIGINT UNSIGNED | 关联导入任务ID |
| `batch_id` | VARCHAR(64) | 批次ID |
| `row_number` | INT | 行号 |
| `sheet_name` | VARCHAR(100) | 工作表名称（Excel专用） |
| `status` | ENUM | 处理状态（pending/processing/success/failed/skipped） |
| `import_action` | ENUM | 导入动作（create/update/skip） |
| `raw_data` | JSON | 原始输入数据 |
| `final_data` | JSON | 最终处理后数据 |
| `ci_id` | BIGINT UNSIGNED | 生成的CI实例ID |
| `ci_type_id` | BIGINT UNSIGNED | CI类型ID |
| `ci_unique_key` | VARCHAR(255) | CI唯一标识 |
| `error_message` | TEXT | 错误消息 |
| `error_code` | VARCHAR(50) | 错误代码 |
| `error_type` | ENUM | 错误类型 |
| `start_time` | TIMESTAMP | 开始处理时间 |
| `end_time` | TIMESTAMP | 结束处理时间 |
| `retry_count` | INT | 重试次数 |

### 4. 导入错误表 (cmdb_import_errors) - **大幅简化**

**主要变更**:
- ✅ 去除复杂的错误统计、自动修复、通知跟踪等功能
- ✅ 去除错误分组、性能信息、外部关联等过度设计
- ✅ 保留核心的错误记录和分类功能

#### 简化后核心字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `task_id` | BIGINT UNSIGNED | 关联导入任务ID |
| `record_id` | BIGINT UNSIGNED | 关联导入记录ID |
| `batch_id` | VARCHAR(64) | 批次ID |
| `error_code` | VARCHAR(50) | 错误代码 |
| `error_title` | VARCHAR(255) | 错误标题 |
| `error_message` | TEXT | 错误消息 |
| `error_details` | TEXT | 详细错误信息 |
| `error_type` | ENUM | 错误类型（6种核心类型） |
| `severity` | ENUM | 严重程度（low/medium/high/critical） |
| `row_number` | INT | 错误发生的行号 |
| `field_name` | VARCHAR(100) | 错误发生的字段名 |
| `input_data` | JSON | 导致错误的输入数据 |
| `suggestion` | TEXT | 修复建议 |
| `status` | ENUM | 处理状态（new/acknowledged/resolved/ignored） |

### 5. CI变更记录表 (cmdb_ci_records) - ⭐ **新增表**

**设计目标**: 
- 记录CI实例的所有变更历史（增删改查）
- 提供完整的审计跟踪能力
- 支持多种变更来源（手动、导入、API、系统同步）

#### 核心字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `ci_id` | BIGINT UNSIGNED | CI实例ID |
| `ci_type_id` | BIGINT UNSIGNED | CI类型ID |
| `ci_type_name` | VARCHAR(100) | CI类型名称 |
| `ci_unique_key` | VARCHAR(255) | CI唯一标识 |
| `operation_type` | ENUM | 操作类型（create/update/delete/restore） |
| `operation_time` | TIMESTAMP | 操作时间 |
| `operation_user_id` | BINARY(16) | 操作用户ID |
| `operation_user_name` | VARCHAR(100) | 操作用户名 |
| `source_type` | ENUM | 变更来源类型（manual/import/api/system/sync） |
| `source_id` | VARCHAR(100) | 来源ID（如导入任务ID） |
| `source_description` | VARCHAR(500) | 来源描述 |
| `before_data` | JSON | 变更前数据 |
| `after_data` | JSON | 变更后数据 |
| `changed_fields` | JSON | 变更字段列表 |
| `change_summary` | JSON | 变更摘要 |
| `change_reason` | VARCHAR(500) | 变更原因 |
| `description` | TEXT | 变更描述 |
| `version_number` | INT | 版本号 |
| `client_ip` | VARCHAR(45) | 客户端IP |
| `affected_relations` | INT | 影响的关系数量 |
| `is_major_change` | BOOLEAN | 是否为重大变更 |
| `requires_approval` | BOOLEAN | 是否需要审批 |
| `approval_status` | ENUM | 审批状态 |

#### 使用场景

```sql
-- 查询某个CI的变更历史
SELECT * FROM cmdb_ci_records 
WHERE ci_id = ? 
ORDER BY operation_time DESC;

-- 查询导入任务产生的所有变更
SELECT * FROM cmdb_ci_records 
WHERE source_type = 'import' AND source_id = ?;

-- 查询用户的操作历史
SELECT * FROM cmdb_ci_records 
WHERE operation_user_id = ? 
ORDER BY operation_time DESC;
```

## 🔍 验证策略设计（无变化）

### 验证规则复用原则

本设计采用**验证规则复用**的策略，避免在导入模板中重复定义验证规则：

#### 验证流程

1. **字段映射阶段**: 只需配置源字段到目标属性的映射关系
2. **验证阶段**: 直接调用现有的 `ValidateCisAttributesLogic`
3. **错误处理**: 统一的错误格式和分类

#### 字段映射配置示例

```json
{
  "source_field": "服务器名称",
  "target_field": "hostname",  // 对应cmdb_attributes.name
  "field_type": "text",
  "is_required": true,
  "transformations": ["trim", "lowercase"]
  // 无需配置validation_rules，直接使用属性定义中的规则
}
```

## 📊 索引策略（优化后）

### 导入任务表索引

```sql
-- 基本查询索引
CREATE INDEX idx_task_id ON cmdb_import_tasks(task_id);
CREATE INDEX idx_status ON cmdb_import_tasks(status);
CREATE INDEX idx_type ON cmdb_import_tasks(type);

-- 复合索引
CREATE INDEX idx_status_priority_created ON cmdb_import_tasks(status, priority, created_at);
CREATE INDEX idx_type_status ON cmdb_import_tasks(type, status);
```

### 导入记录表索引

```sql
-- 基本查询索引
CREATE INDEX idx_task_id ON cmdb_import_records(task_id);
CREATE INDEX idx_status ON cmdb_import_records(status);
CREATE INDEX idx_ci_id ON cmdb_import_records(ci_id);

-- 复合索引
CREATE INDEX idx_task_status ON cmdb_import_records(task_id, status);
CREATE INDEX idx_ci_unique_key ON cmdb_import_records(ci_unique_key);
```

### CI变更记录表索引

```sql
-- 基本查询索引
CREATE INDEX idx_ci_id ON cmdb_ci_records(ci_id);
CREATE INDEX idx_operation_time ON cmdb_ci_records(operation_time);
CREATE INDEX idx_operation_type ON cmdb_ci_records(operation_type);
CREATE INDEX idx_source_type ON cmdb_ci_records(source_type);

-- 复合索引
CREATE INDEX idx_ci_operation_time ON cmdb_ci_records(ci_id, operation_time);
CREATE INDEX idx_source_type_id ON cmdb_ci_records(source_type, source_id);
CREATE INDEX idx_user_time ON cmdb_ci_records(operation_user_id, operation_time);
```

## 🔄 表关系设计

### ER关系图

```
cmdb_import_templates 1:N cmdb_import_tasks
cmdb_import_tasks 1:N cmdb_import_records  
cmdb_import_tasks 1:N cmdb_import_errors
cmdb_import_records 1:N cmdb_import_errors

cmdb_cis 1:N cmdb_import_records
cmdb_ci_types ← → cmdb_import_records  
cmdb_cis 1:N cmdb_ci_records      ⭐ 新增
cmdb_ci_types 1:N cmdb_ci_records ⭐ 新增
```

## 💡 使用示例

### 1. 创建导入任务（简化版）

```sql
INSERT INTO cmdb_import_tasks (
    task_id, name, type, status, source_path, template_id
) VALUES (
    'task_20241220_001', '服务器资产导入', 'excel', 'pending', 
    '/uploads/servers.xlsx', 1
);
```

### 2. 记录导入详情（简化版）

```sql
INSERT INTO cmdb_import_records (
    task_id, row_number, status, raw_data, ci_id
) VALUES (
    1, 2, 'success', '{"hostname":"server01","ip":"192.168.1.1"}', 100
);
```

### 3. 记录CI变更（新功能）

```sql
INSERT INTO cmdb_ci_records (
    ci_id, ci_type_id, operation_type, operation_time, 
    source_type, source_id, before_data, after_data
) VALUES (
    100, 1, 'create', NOW(), 'import', 'task_20241220_001',
    NULL, '{"hostname":"server01","ip":"192.168.1.1"}'
);
```

## 📊 性能考虑

### 数据量估算

| 表名 | 预计记录数 | 增长率 | 保留期 |
|------|------------|--------|--------|
| import_tasks | 10,000/年 | 30/天 | 2年 |
| import_records | 100万/年 | 3000/天 | 1年 |
| import_errors | 10万/年 | 300/天 | 1年 |
| ci_records | 500万/年 | 15000/天 | 永久 |

### 分区策略建议

```sql
-- CI变更记录表按月分区
ALTER TABLE cmdb_ci_records 
PARTITION BY RANGE (YEAR(operation_time) * 100 + MONTH(operation_time)) (
    PARTITION p202412 VALUES LESS THAN (202501),
    PARTITION p202501 VALUES LESS THAN (202502),
    ...
);
```

## 📊 总结

### 🎯 核心设计亮点

**简化设计策略**
- 去除了50%以上的非核心字段，专注业务本质
- 保留了完整的导入流程和错误处理能力
- 新增独立的CI变更审计机制

**验证规则复用策略**
- 避免重复定义验证规则，确保导入数据与现有CI数据使用相同验证标准
- 利用现有ValidateCisAttributesLogic，保持系统一致性
- 降低维护成本，验证规则变更无需同步更新导入模板

**企业级审计能力**
- 完整的CI变更历史记录
- 支持多种变更来源追溯
- 灵活的审批和版本管理机制

### 💪 技术优势

1. **性能优化**: 简化字段减少存储空间，优化索引提升查询效率
2. **维护性**: 减少复杂性，降低维护难度
3. **扩展性**: 保留核心扩展点，支持未来需求演进
4. **一致性**: 充分复用现有验证和业务逻辑

通过这套简化的数据库设计，可以高效支持企业级的资产导入需求，同时提供完善的CI变更审计能力。 