# CMDB资产录入工作流程设计

> **版本**: v1.0  
> **创建时间**: 2024-12-20  
> **设计目标**: 用户友好的资产录入完整流程  
> **涵盖范围**: 前端UI → API层 → RPC层 → 数据库

## 🎯 设计目标

### 用户视角需求
1. **简单易用**: 用户能够轻松导入资产数据
2. **模板驱动**: 提供标准化的导入模板
3. **实时反馈**: 导入过程和结果的实时展示
4. **错误处理**: 清晰的错误提示和修复建议
5. **批量处理**: 支持大批量数据导入

### 技术架构需求
1. **数据一致性**: 确保入库数据的完整性
2. **性能优化**: 大批量数据的高效处理
3. **可追溯性**: 完整的操作记录和审计
4. **可扩展性**: 支持多种数据源和格式

## 🏗️ 整体架构流程

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                           CMDB资产录入完整工作流程                                    │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 1. 用户交互层 (Frontend)                                                              │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 模板管理页面 │ │ Excel导入页面│ │ API导入页面  │ │ 进度监控页面 │ │ 结果查看页面 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 2. API网关层 (cmdb-api)                                                               │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 模板管理API  │ │ 文件上传API  │ │ 数据导入API  │ │ 任务状态API  │ │ 结果查询API  │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 3. 业务处理层 (cmdb-rpc)                                                              │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 模板引擎     │ │ 数据解析器   │ │ 验证引擎     │ │ 转换引擎     │ │ 入库引擎     │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 4. 数据存储层                                                                         │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ CI实例表     │ │ 属性值表     │ │ 关系表       │ │ 导入记录表   │ │ 错误日志表   │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

## 📋 详细工作流程设计

### 流程一: 模板管理流程

#### 1.1 模板定义阶段
```
用户操作:
1. 选择CI类型 (如: 服务器、网络设备、应用系统)
2. 配置必填字段和可选字段
3. 设置字段验证规则
4. 定义字段映射关系
5. 保存模板配置

系统处理:
1. 验证CI类型有效性
2. 检查字段配置合理性
3. 生成模板元数据
4. 保存到模板库
```

#### 1.2 Excel模板生成
```
输入: 模板ID + 用户需求
处理流程:
1. 读取模板配置
2. 生成Excel标准格式
   - Sheet1: 数据录入区
   - Sheet2: 字段说明
   - Sheet3: 数据验证规则
   - Sheet4: 示例数据
3. 设置Excel数据验证
4. 生成下载链接

输出: 可下载的Excel模板文件
```

#### 1.3 API模板生成
```
输入: 模板ID
处理流程:
1. 读取模板配置
2. 生成JSON Schema
3. 生成API文档
4. 提供Postman Collection

输出: 
- JSON Schema文件
- API接口文档
- 示例请求数据
```

### 流程二: Excel导入流程

#### 2.1 文件上传阶段
```
前端操作:
1. 用户选择Excel文件
2. 选择对应的模板
3. 上传文件到服务器

API层处理:
1. 文件格式验证
2. 文件大小检查
3. 保存到临时存储
4. 返回文件ID

数据结构:
{
  "file_id": "upload_123456789",
  "original_name": "server_import.xlsx",
  "template_id": "template_server_v1",
  "upload_time": "2024-12-20T10:00:00Z",
  "file_size": 2048576,
  "status": "uploaded"
}
```

#### 2.2 数据预览阶段
```
用户操作:
1. 查看解析后的数据预览
2. 确认字段映射正确性
3. 查看数据验证结果
4. 修正错误数据或重新上传

系统处理:
1. Excel文件解析
2. 数据格式转换
3. 基础验证检查
4. 生成预览报告

预览数据结构:
{
  "preview_id": "preview_123456789",
  "total_rows": 1000,
  "valid_rows": 950,
  "invalid_rows": 50,
  "columns": [
    {
      "excel_column": "A",
      "field_name": "hostname",
      "data_type": "string",
      "sample_values": ["web-01", "web-02", "db-01"]
    }
  ],
  "errors": [
    {
      "row": 15,
      "column": "B",
      "error": "IP地址格式不正确",
      "value": "192.168.1"
    }
  ]
}
```

#### 2.3 导入确认阶段
```
用户操作:
1. 确认导入设置
2. 选择错误处理策略
3. 提交导入任务

导入配置:
{
  "import_mode": "insert|update|upsert",
  "error_strategy": "stop_on_error|skip_errors|partial_import",
  "batch_size": 100,
  "duplicate_strategy": "skip|overwrite|merge",
  "enable_notification": true,
  "notification_email": "user@example.com"
}
```

#### 2.4 异步处理阶段
```
RPC层处理:
1. 创建导入任务
2. 数据分块处理
3. 逐步验证和转换
4. 批量入库操作
5. 实时进度更新

数据流转:
Excel数据 → 解析器 → 验证器 → 转换器 → 入库器 → 数据库

任务状态跟踪:
{
  "task_id": "import_task_123456789",
  "status": "processing",
  "progress": {
    "total_items": 1000,
    "processed_items": 250,
    "success_items": 240,
    "failed_items": 10,
    "current_step": "数据入库",
    "percentage": 25.0
  },
  "estimated_remaining": "00:03:45"
}
```

### 流程三: API导入流程

#### 3.1 接口定义
```
POST /api/v1/assets/import
Content-Type: application/json

Request Body:
{
  "template_id": "template_server_v1",
  "import_mode": "insert",
  "data": [
    {
      "hostname": "web-01",
      "ip_address": "192.168.1.100",
      "cpu_cores": 8,
      "memory_gb": 32,
      "disk_gb": 500,
      "os": "CentOS 7",
      "department": "IT部门",
      "owner": "张三"
    }
  ],
  "options": {
    "validate_only": false,
    "batch_size": 100,
    "error_strategy": "stop_on_error"
  }
}

Response:
{
  "success": true,
  "task_id": "api_import_123456789",
  "message": "导入任务已创建",
  "summary": {
    "total_items": 1,
    "estimated_time": "00:00:30"
  }
}
```

#### 3.2 批量API导入
```
POST /api/v1/assets/batch-import
Content-Type: application/json

支持多种数据格式:
1. JSON数组格式
2. CSV文本格式  
3. 压缩包格式 (多个文件)

处理流程:
1. 数据格式检测
2. 异步任务创建
3. 分批处理
4. 进度跟踪
5. 结果通知
```

## 🗄️ 数据库设计

### 核心入库表设计

#### 1. 导入任务表 (import_tasks)
```sql
CREATE TABLE import_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) UNIQUE NOT NULL COMMENT '任务唯一标识',
    template_id VARCHAR(64) NOT NULL COMMENT '模板ID',
    import_type ENUM('excel', 'api', 'csv') NOT NULL COMMENT '导入类型',
    import_mode ENUM('insert', 'update', 'upsert') NOT NULL COMMENT '导入模式',
    status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    
    -- 数据统计
    total_items INT NOT NULL DEFAULT 0 COMMENT '总条数',
    processed_items INT NOT NULL DEFAULT 0 COMMENT '已处理条数',
    success_items INT NOT NULL DEFAULT 0 COMMENT '成功条数',
    failed_items INT NOT NULL DEFAULT 0 COMMENT '失败条数',
    
    -- 文件信息
    source_file VARCHAR(255) COMMENT '源文件名',
    file_size BIGINT COMMENT '文件大小',
    file_path VARCHAR(500) COMMENT '文件存储路径',
    
    -- 配置信息
    import_config JSON COMMENT '导入配置',
    error_strategy ENUM('stop_on_error', 'skip_errors', 'partial_import') DEFAULT 'stop_on_error',
    
    -- 时间信息
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    start_time TIMESTAMP NULL,
    end_time TIMESTAMP NULL,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 用户信息
    created_by VARCHAR(64) NOT NULL COMMENT '创建人',
    
    INDEX idx_task_id (task_id),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time),
    INDEX idx_created_by (created_by)
) COMMENT='导入任务表';
```

#### 2. 导入记录详情表 (import_records)
```sql
CREATE TABLE import_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL COMMENT '任务ID',
    batch_no INT NOT NULL COMMENT '批次号',
    row_no INT NOT NULL COMMENT '行号',
    
    -- 处理状态
    status ENUM('pending', 'processing', 'success', 'failed', 'skipped') NOT NULL DEFAULT 'pending',
    error_message TEXT COMMENT '错误信息',
    error_code VARCHAR(32) COMMENT '错误代码',
    
    -- 原始数据
    raw_data JSON NOT NULL COMMENT '原始数据',
    validated_data JSON COMMENT '验证后数据',
    
    -- CI信息
    ci_id BIGINT COMMENT '生成的CI ID',
    ci_type_id BIGINT COMMENT 'CI类型ID',
    
    -- 处理时间
    process_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_task_id (task_id),
    INDEX idx_status (status),
    INDEX idx_ci_id (ci_id),
    INDEX idx_batch_no (task_id, batch_no)
) COMMENT='导入记录详情表';
```

#### 3. 模板配置表 (import_templates)
```sql
CREATE TABLE import_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    template_id VARCHAR(64) UNIQUE NOT NULL COMMENT '模板唯一标识',
    template_name VARCHAR(128) NOT NULL COMMENT '模板名称',
    template_type ENUM('excel', 'api', 'csv') NOT NULL COMMENT '模板类型',
    ci_type_id BIGINT NOT NULL COMMENT '关联CI类型ID',
    
    -- 模板配置
    field_mapping JSON NOT NULL COMMENT '字段映射配置',
    validation_rules JSON COMMENT '验证规则',
    default_values JSON COMMENT '默认值',
    
    -- 模板文件
    excel_template_path VARCHAR(500) COMMENT 'Excel模板文件路径',
    json_schema TEXT COMMENT 'JSON Schema定义',
    
    -- 版本信息
    version VARCHAR(16) NOT NULL DEFAULT '1.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 时间信息
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(64) NOT NULL,
    
    INDEX idx_template_id (template_id),
    INDEX idx_ci_type_id (ci_type_id),
    INDEX idx_active (is_active)
) COMMENT='导入模板配置表';
```

#### 4. 导入错误日志表 (import_errors)
```sql
CREATE TABLE import_errors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(64) NOT NULL COMMENT '任务ID',
    record_id BIGINT COMMENT '记录ID',
    error_type ENUM('validation', 'conversion', 'database', 'business') NOT NULL COMMENT '错误类型',
    error_level ENUM('warning', 'error', 'fatal') NOT NULL DEFAULT 'error' COMMENT '错误级别',
    
    -- 错误详情
    error_code VARCHAR(32) COMMENT '错误代码',
    error_message TEXT NOT NULL COMMENT '错误信息',
    field_name VARCHAR(128) COMMENT '出错字段',
    field_value TEXT COMMENT '出错值',
    
    -- 位置信息
    row_no INT COMMENT '行号',
    column_no INT COMMENT '列号',
    
    -- 建议信息
    suggestion TEXT COMMENT '修复建议',
    
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_task_id (task_id),
    INDEX idx_error_type (error_type),
    INDEX idx_error_level (error_level)
) COMMENT='导入错误日志表';
```

## 🔄 模板管理设计

### 模板类型定义

#### Excel模板结构
```
工作表结构:
├── Sheet1: 数据录入区
│   ├── 第1行: 字段标题 (中文显示名)
│   ├── 第2行: 字段名称 (英文字段名)
│   ├── 第3行: 数据类型 (string/int/float/date)
│   ├── 第4行: 是否必填 (必填/可选)
│   ├── 第5行: 示例数据
│   └── 第6行开始: 用户录入数据
│
├── Sheet2: 字段说明
│   ├── 字段名称
│   ├── 字段描述
│   ├── 数据格式要求
│   ├── 取值范围
│   └── 填写示例
│
├── Sheet3: 数据验证规则
│   ├── 必填字段检查
│   ├── 数据格式验证
│   ├── 取值范围限制
│   └── 业务规则验证
│
└── Sheet4: 示例数据
    └── 完整的示例记录
```

#### API模板结构
```json
{
  "template_info": {
    "template_id": "template_server_v1",
    "template_name": "服务器资产导入模板",
    "version": "1.0",
    "ci_type": "server"
  },
  "field_definitions": [
    {
      "field_name": "hostname",
      "display_name": "主机名",
      "data_type": "string",
      "required": true,
      "max_length": 64,
      "pattern": "^[a-zA-Z0-9\\-]+$",
      "description": "服务器主机名，只能包含字母、数字和连字符"
    },
    {
      "field_name": "ip_address",
      "display_name": "IP地址",
      "data_type": "string",
      "required": true,
      "pattern": "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$",
      "description": "服务器主IP地址"
    }
  ],
  "validation_rules": [
    {
      "rule_type": "unique",
      "fields": ["hostname"],
      "message": "主机名不能重复"
    },
    {
      "rule_type": "reference",
      "field": "department",
      "reference_table": "departments",
      "message": "部门必须在系统中存在"
    }
  ],
  "default_values": {
    "status": "运行中",
    "environment": "生产"
  }
}
```

### 模板生成引擎

#### Excel模板生成流程
```
输入: 模板配置 + CI类型信息
处理步骤:
1. 读取CI类型的属性定义
2. 生成Excel工作簿结构
3. 设置数据验证规则
4. 添加样式和格式
5. 插入示例数据
6. 生成下载文件

输出: 标准化Excel模板文件
```

#### JSON Schema生成流程
```
输入: 模板配置 + CI类型信息
处理步骤:
1. 分析字段类型和验证规则
2. 生成JSON Schema定义
3. 添加验证约束
4. 生成示例数据
5. 输出API文档

输出: 
- JSON Schema文件
- OpenAPI规范
- 示例请求数据
```

## 🎛️ 用户交互设计

### 前端页面结构

#### 1. 模板管理页面
```
功能模块:
├── 模板列表展示
│   ├── 模板基本信息
│   ├── 使用统计
│   ├── 操作按钮 (编辑/删除/复制)
│   └── 模板下载链接
│
├── 模板创建向导
│   ├── 步骤1: 选择CI类型
│   ├── 步骤2: 配置字段映射
│   ├── 步骤3: 设置验证规则
│   ├── 步骤4: 预览模板
│   └── 步骤5: 保存并生成
│
└── 模板编辑器
    ├── 字段配置区
    ├── 验证规则区
    ├── 预览区
    └── 保存操作区
```

#### 2. 数据导入页面
```
导入流程:
├── 步骤1: 选择导入方式
│   ├── Excel文件导入
│   ├── API批量导入
│   └── CSV文件导入
│
├── 步骤2: 选择模板
│   ├── 模板选择器
│   ├── 模板预览
│   └── 模板下载
│
├── 步骤3: 上传文件/输入数据
│   ├── 文件拖拽上传
│   ├── 文件格式验证
│   └── 上传进度显示
│
├── 步骤4: 数据预览确认
│   ├── 数据表格展示
│   ├── 错误标记
│   ├── 统计信息
│   └── 修正操作
│
├── 步骤5: 确认导入
│   ├── 导入设置
│   ├── 错误处理策略
│   └── 提交导入
│
└── 步骤6: 进度监控
    ├── 实时进度条
    ├── 处理状态
    ├── 错误信息
    └── 结果预览
```

#### 3. 任务监控页面
```
监控功能:
├── 任务列表
│   ├── 任务状态过滤
│   ├── 时间范围筛选
│   ├── 创建人筛选
│   └── 分页显示
│
├── 任务详情
│   ├── 基本信息
│   ├── 进度统计
│   ├── 错误详情
│   └── 操作日志
│
└── 实时监控
    ├── 系统负载
    ├── 处理性能
    ├── 错误率统计
    └── 资源使用率
```

## 📊 API接口设计

### 核心API接口

#### 1. 模板管理API
```
# 获取模板列表
GET /api/v1/templates
Response: {
  "templates": [
    {
      "template_id": "template_server_v1",
      "template_name": "服务器资产模板",
      "ci_type": "server",
      "version": "1.0",
      "create_time": "2024-12-20T10:00:00Z",
      "usage_count": 150
    }
  ]
}

# 下载Excel模板
GET /api/v1/templates/{template_id}/excel
Response: Excel文件下载

# 获取JSON Schema
GET /api/v1/templates/{template_id}/schema
Response: JSON Schema定义

# 创建模板
POST /api/v1/templates
Request: 模板配置信息
Response: 创建结果
```

#### 2. 数据导入API
```
# 文件上传
POST /api/v1/import/upload
Content-Type: multipart/form-data
Response: {
  "file_id": "upload_123456789",
  "preview_url": "/api/v1/import/preview/123456789"
}

# 数据预览
GET /api/v1/import/preview/{file_id}
Response: {
  "preview_data": [...],
  "validation_result": {...},
  "statistics": {...}
}

# 提交导入
POST /api/v1/import/submit
Request: {
  "file_id": "upload_123456789",
  "template_id": "template_server_v1",
  "import_config": {...}
}
Response: {
  "task_id": "import_task_123456789"
}

# 查询任务状态
GET /api/v1/import/tasks/{task_id}
Response: 任务详细状态

# 获取导入结果
GET /api/v1/import/tasks/{task_id}/result
Response: 导入结果详情
```

#### 3. 批量API导入
```
# 批量数据导入
POST /api/v1/assets/batch-import
Request: {
  "template_id": "template_server_v1",
  "data": [...],
  "options": {...}
}
Response: 异步任务信息

# 实时API导入
POST /api/v1/assets/import
Request: 单条或小批量数据
Response: 同步处理结果
```

## 🔐 权限和安全设计

### 权限控制
```
权限分级:
├── 模板管理权限
│   ├── 查看模板
│   ├── 创建模板
│   ├── 编辑模板
│   └── 删除模板
│
├── 数据导入权限
│   ├── 上传文件
│   ├── 预览数据
│   ├── 执行导入
│   └── 查看结果
│
└── 任务管理权限
    ├── 查看任务
    ├── 取消任务
    ├── 重试任务
    └── 删除任务
```

### 安全措施
```
安全控制:
├── 文件安全
│   ├── 文件类型检查
│   ├── 文件大小限制
│   ├── 病毒扫描
│   └── 内容安全检查
│
├── 数据安全
│   ├── 敏感数据脱敏
│   ├── 数据访问控制
│   ├── 操作审计日志
│   └── 数据加密存储
│
└── 接口安全
    ├── 访问频率限制
    ├── IP白名单控制
    ├── Token有效期管理
    └── 异常访问监控
```

## 🚀 总结

这个完整的资产录入工作流程设计涵盖了：

### ✅ 核心功能
1. **用户友好的模板管理** - 可视化模板创建和维护
2. **多种导入方式支持** - Excel、API、CSV等多种格式
3. **实时进度监控** - 完整的任务状态跟踪
4. **完善的错误处理** - 详细的错误信息和修复建议
5. **数据完整性保证** - 从上传到入库的全程数据验证

### 🎯 用户价值
1. **简化操作流程** - 向导式操作，降低使用门槛
2. **提高导入效率** - 批量处理，异步执行
3. **保证数据质量** - 多层验证，错误预警
4. **增强可视化** - 实时监控，结果展示

### 📈 技术优势
1. **可扩展架构** - 支持多种数据源和格式
2. **高性能处理** - 异步任务，批量入库
3. **完整审计** - 全程记录，可追溯
4. **企业级稳定** - 错误恢复，数据一致性

下一步我们可以根据这个设计开始实现具体的功能模块。您希望从哪个部分开始实现呢？ 