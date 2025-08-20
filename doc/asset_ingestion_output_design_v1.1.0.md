# CMDB资产录入与输出架构设计

> **版本**: v1.1.0  
> **创建时间**: 2024-12-20  
> **基于**: CMDB v1.0.0 架构设计  
> **维护团队**: NewBee CMDB Team  
> **设计目标**: 企业级、高性能、高扩展性的资产录入与输出机制

## 📋 设计概述

基于现有CMDB基础架构，设计统一的资产录入与输出引擎，支持多种数据输入方式、智能数据处理、高性能存储和灵活的数据输出能力。该架构采用插件化设计，确保高扩展性和高可用性。

### 核心设计原则
- **统一接入**: 所有数据输入通过统一的数据接入层处理
- **智能处理**: 自动数据验证、转换、清洗和去重
- **异步解耦**: 输入处理与存储分离，提高系统响应性
- **插件扩展**: 支持自定义输入输出插件
- **性能优先**: 批量处理、并发处理、缓存优化
- **可观测性**: 完整的监控、日志和告警机制

## 🏗️ 整体架构图 
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ CMDB资产录入与输出引擎架构 │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据输入层 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ Excel导入 │ │ API接口 │ │ 自动发现 │ │ 第三方推送 │ │ 文件上传 │ │
│ │ 插件 │ │ 插件 │ │ 插件 │ │ 插件 │ │ 插件 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 统一数据接入网关 │
│ ┌─────────────────────────────────────────────────────────────────────────────────┐ │
│ │ 路由分发 │ 格式解析 │ 权限验证 │ 限流控制 │ 数据预处理 │ │ │
│ └─────────────────────────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据处理引擎 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 数据验证 │ │ 数据转换 │ │ 数据清洗 │ │ 数据去重 │ │ 数据映射 │ │
│ │ 引擎 │ │ 引擎 │ │ 引擎 │ │ 引擎 │ │ 引擎 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 异步处理队列 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 优先级队列 │ │ 批量处理 │ │ 失败重试 │ │ 死信队列 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据存储引擎 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 事务管理 │ │ 批量写入 │ │ 索引维护 │ │ 缓存同步 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据输出层 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ REST API │ │ GraphQL │ │ gRPC │ │ WebSocket │ │ 文件导出 │ │
│ │ 输出 │ │ 输出 │ │ 输出 │ │ 实时推送 │ │ 输出 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 监控与可观测性 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 性能监控 │ │ 错误追踪 │ │ 日志聚合 │ │ 告警通知 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────────┘
继续

...
我来继续完成CMDB资产录入与输出架构设计文档：
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ CMDB资产录入与输出引擎架构 │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据输入层 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ Excel导入 │ │ API接口 │ │ 自动发现 │ │ 第三方推送 │ │ 文件上传 │ │
│ │ 插件 │ │ 插件 │ │ 插件 │ │ 插件 │ │ 插件 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 统一数据接入网关 │
│ ┌─────────────────────────────────────────────────────────────────────────────────┐ │
│ │ 路由分发 │ 格式解析 │ 权限验证 │ 限流控制 │ 数据预处理 │ │ │
│ └─────────────────────────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据处理引擎 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 数据验证 │ │ 数据转换 │ │ 数据清洗 │ │ 数据去重 │ │ 数据映射 │ │
│ │ 引擎 │ │ 引擎 │ │ 引擎 │ │ 引擎 │ │ 引擎 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 异步处理队列 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 优先级队列 │ │ 批量处理 │ │ 失败重试 │ │ 死信队列 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据存储引擎 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 事务管理 │ │ 批量写入 │ │ 索引维护 │ │ 缓存同步 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 数据输出层 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ REST API │ │ GraphQL │ │ gRPC │ │ WebSocket │ │ 文件导出 │ │
│ │ 输出 │ │ 输出 │ │ 输出 │ │ 实时推送 │ │ 输出 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ 监控与可观测性 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ 性能监控 │ │ 错误追踪 │ │ 日志聚合 │ │ 告警通知 │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────────┘

## 🔌 数据输入引擎设计

### 1. 插件化输入架构

基于现有的logic层架构，设计统一的输入插件接口：

```go
// 输入插件接口定义
type DataInputPlugin interface {
    // 插件基本信息
    Name() string
    Version() string
    Description() string
    SupportedFormats() []string
    
    // 数据处理能力
    Parse(ctx context.Context, input *InputData) (*ParsedData, error)
    Validate(ctx context.Context, data *ParsedData) (*ValidationResult, error)
    Transform(ctx context.Context, data *ParsedData) (*TransformedData, error)
    
    // 插件配置
    Configure(config map[string]interface{}) error
    HealthCheck() error
}

// 统一输入数据结构
type InputData struct {
    ID          string                 `json:"id"`          // 请求ID
    Source      string                 `json:"source"`      // 数据来源
    Format      string                 `json:"format"`      // 数据格式
    Content     []byte                 `json:"content"`     // 原始内容
    Metadata    map[string]interface{} `json:"metadata"`    // 元数据信息
    Headers     map[string]string      `json:"headers"`     // 请求头信息
    Timestamp   time.Time              `json:"timestamp"`   // 接收时间
}

// 解析后数据结构
type ParsedData struct {
    Records   []*AssetRecord         `json:"records"`   // 资产记录
    Relations []*RelationRecord      `json:"relations"` // 关系记录
    Metadata  map[string]interface{} `json:"metadata"`  // 解析元数据
    Errors    []ParseError           `json:"errors"`    // 解析错误
}

// 资产记录结构
type AssetRecord struct {
    ID          string                 `json:"id"`           // 记录ID
    TypeID      uint64                 `json:"type_id"`      // CI类型ID
    TypeName    string                 `json:"type_name"`    // CI类型名称
    Attributes  map[string]interface{} `json:"attributes"`   // 属性值
    Tags        []string               `json:"tags"`         // 标签
    Metadata    map[string]interface{} `json:"metadata"`     // 元数据
    Source      string                 `json:"source"`       // 数据源
    LineNumber  int                    `json:"line_number"`  // 行号（用于错误定位）
}
```

### 2. Excel导入插件

```go
// Excel导入插件实现
type ExcelInputPlugin struct {
    config *ExcelConfig
    logger logx.Logger
}

type ExcelConfig struct {
    MaxFileSize    int64             `json:"max_file_size"`    // 最大文件大小 (MB)
    MaxRows        int               `json:"max_rows"`         // 最大行数
    SheetMapping   map[string]string `json:"sheet_mapping"`    // 工作表映射
    ColumnMapping  map[string]string `json:"column_mapping"`   // 列映射
    RequiredFields []string          `json:"required_fields"`  // 必填字段
}

func (p *ExcelInputPlugin) Parse(ctx context.Context, input *InputData) (*ParsedData, error) {
    // 打开Excel文件
    file, err := excelize.OpenReader(bytes.NewReader(input.Content))
    if err != nil {
        return nil, fmt.Errorf("打开Excel文件失败: %v", err)
    }
    defer file.Close()

    result := &ParsedData{
        Records:  make([]*AssetRecord, 0),
        Metadata: map[string]interface{}{
            "sheet_count": len(file.GetSheetList()),
            "file_size":   len(input.Content),
        },
    }

    // 遍历工作表
    for _, sheetName := range file.GetSheetList() {
        records, err := p.parseSheet(file, sheetName)
        if err != nil {
            result.Errors = append(result.Errors, ParseError{
                Type:    "sheet_error",
                Sheet:   sheetName,
                Message: err.Error(),
            })
            continue
        }
        result.Records = append(result.Records, records...)
    }

    return result, nil
}

func (p *ExcelInputPlugin) parseSheet(file *excelize.File, sheetName string) ([]*AssetRecord, error) {
    rows, err := file.GetRows(sheetName)
    if err != nil {
        return nil, err
    }

    if len(rows) == 0 {
        return nil, fmt.Errorf("工作表 %s 为空", sheetName)
    }

    // 获取表头映射
    headerMap := make(map[int]string)
    if len(rows) > 0 {
        for i, header := range rows[0] {
            if mappedField, exists := p.config.ColumnMapping[header]; exists {
                headerMap[i] = mappedField
            } else {
                headerMap[i] = header
            }
        }
    }

    records := make([]*AssetRecord, 0)
    
    // 从第二行开始解析数据
    for rowIndex, row := range rows[1:] {
        record := &AssetRecord{
            ID:         fmt.Sprintf("%s_%d", sheetName, rowIndex+2),
            Attributes: make(map[string]interface{}),
            LineNumber: rowIndex + 2,
            Source:     "excel_import",
        }

        // 解析每个单元格
        for colIndex, cellValue := range row {
            if fieldName, exists := headerMap[colIndex]; exists && cellValue != "" {
                record.Attributes[fieldName] = cellValue
            }
        }

        // 验证必填字段
        if err := p.validateRequiredFields(record); err != nil {
            // 记录错误但继续处理
            continue
        }

        records = append(records, record)
    }

    return records, nil
}
```

### 3. API接口输入插件

```go
// API输入插件
type APIInputPlugin struct {
    config *APIConfig
    logger logx.Logger
}

type APIConfig struct {
    BatchSize      int               `json:"batch_size"`       // 批量大小
    RateLimit      int               `json:"rate_limit"`       // 限流配置
    AuthRequired   bool              `json:"auth_required"`    // 是否需要认证
    AllowedSources []string          `json:"allowed_sources"`  // 允许的数据源
}

// API输入数据结构
type APIAssetRequest struct {
    Source    string        `json:"source"`    // 数据源标识
    BatchID   string        `json:"batch_id"`  // 批次ID
    Assets    []APIAsset    `json:"assets"`    // 资产列表
    Relations []APIRelation `json:"relations"` // 关系列表
}

type APIAsset struct {
    ExternalID string                 `json:"external_id"` // 外部系统ID
    TypeName   string                 `json:"type_name"`   // 类型名称
    Attributes map[string]interface{} `json:"attributes"`  // 属性值
    Tags       []string               `json:"tags"`        // 标签
}

func (p *APIInputPlugin) Parse(ctx context.Context, input *InputData) (*ParsedData, error) {
    var apiRequest APIAssetRequest
    if err := json.Unmarshal(input.Content, &apiRequest); err != nil {
        return nil, fmt.Errorf("解析API请求失败: %v", err)
    }

    result := &ParsedData{
        Records: make([]*AssetRecord, 0, len(apiRequest.Assets)),
        Metadata: map[string]interface{}{
            "batch_id":    apiRequest.BatchID,
            "source":      apiRequest.Source,
            "asset_count": len(apiRequest.Assets),
        },
    }

    // 转换API资产为内部格式
    for i, asset := range apiRequest.Assets {
        record := &AssetRecord{
            ID:         fmt.Sprintf("%s_%s_%d", apiRequest.Source, apiRequest.BatchID, i),
            TypeName:   asset.TypeName,
            Attributes: asset.Attributes,
            Tags:       asset.Tags,
            Source:     apiRequest.Source,
            Metadata: map[string]interface{}{
                "external_id": asset.ExternalID,
                "batch_id":    apiRequest.BatchID,
            },
        }
        result.Records = append(result.Records, record)
    }

    return result, nil
}
```

### 4. 自动发现插件

```go
// 自动发现插件
type DiscoveryInputPlugin struct {
    config     *DiscoveryConfig
    agentPool  *AgentPool
    scheduler  *TaskScheduler
    logger     logx.Logger
}

type DiscoveryConfig struct {
    AgentTimeout    time.Duration     `json:"agent_timeout"`     // Agent超时时间
    DiscoveryRules  []DiscoveryRule   `json:"discovery_rules"`   // 发现规则
    UpdateStrategy  string            `json:"update_strategy"`   // 更新策略：merge/replace
    ConflictPolicy  string            `json:"conflict_policy"`   // 冲突处理策略
}

type DiscoveryRule struct {
    Name        string            `json:"name"`         // 规则名称
    TargetType  string            `json:"target_type"`  // 目标类型：server/network/application
    Filters     map[string]string `json:"filters"`      // 过滤条件
    Mappings    []AttributeMapping `json:"mappings"`    // 属性映射
    Enabled     bool              `json:"enabled"`      // 是否启用
}

type AttributeMapping struct {
    SourceField string      `json:"source_field"` // 源字段
    TargetField string      `json:"target_field"` // 目标字段
    Transform   string      `json:"transform"`    // 转换规则
    DefaultValue interface{} `json:"default_value"` // 默认值
}

func (p *DiscoveryInputPlugin) Parse(ctx context.Context, input *InputData) (*ParsedData, error) {
    var discoveryData AgentDiscoveryData
    if err := json.Unmarshal(input.Content, &discoveryData); err != nil {
        return nil, fmt.Errorf("解析Agent发现数据失败: %v", err)
    }

    result := &ParsedData{
        Records: make([]*AssetRecord, 0),
        Metadata: map[string]interface{}{
            "agent_id":       discoveryData.AgentID,
            "discovery_time": discoveryData.Timestamp,
            "target_count":   len(discoveryData.Targets),
        },
    }

    // 应用发现规则
    for _, target := range discoveryData.Targets {
        for _, rule := range p.config.DiscoveryRules {
            if !rule.Enabled {
                continue
            }

            if p.matchRule(target, rule) {
                record := p.transformTarget(target, rule)
                if record != nil {
                    result.Records = append(result.Records, record)
                }
            }
        }
    }

    return result, nil
}
```

## 🔄 数据处理引擎设计

### 1. 统一数据处理管道

```go
// 数据处理管道
type DataProcessingPipeline struct {
    validators  []DataValidator
    transformers []DataTransformer
    cleaners    []DataCleaner
    deduplicators []DataDeduplicator
    enrichers   []DataEnricher
    logger      logx.Logger
}

// 数据验证器接口
type DataValidator interface {
    Name() string
    Validate(ctx context.Context, data *ParsedData) (*ValidationResult, error)
}

// 基于现有验证器的实现
type AttributeValidator struct {
    svcCtx *svc.ServiceContext
    logger logx.Logger
}

func (v *AttributeValidator) Validate(ctx context.Context, data *ParsedData) (*ValidationResult, error) {
    result := &ValidationResult{
        Valid:   true,
        Errors:  make([]ValidationError, 0),
        Records: make([]*ValidatedRecord, 0),
    }

    for _, record := range data.Records {
        // 获取CI类型信息
        ciType, err := v.getCITypeByName(ctx, record.TypeName)
        if err != nil {
            result.Valid = false
            result.Errors = append(result.Errors, ValidationError{
                RecordID: record.ID,
                Type:     "type_not_found",
                Message:  fmt.Sprintf("CI类型 %s 不存在", record.TypeName),
            })
            continue
        }

        // 构建验证请求
        validateReq := &cmdb.CisAttributeValidateReq{
            TypeId:     ciType.ID,
            Attributes: v.buildAttributeValues(record.Attributes),
        }

        // 调用现有的验证逻辑
        validateLogic := cis.NewValidateCisAttributesLogic(ctx, v.svcCtx)
        validateResp, err := validateLogic.ValidateCisAttributes(validateReq)
        if err != nil {
            result.Valid = false
            result.Errors = append(result.Errors, ValidationError{
                RecordID: record.ID,
                Type:     "validation_error",
                Message:  err.Error(),
            })
            continue
        }

        if !validateResp.Valid {
            result.Valid = false
            for _, attrError := range validateResp.Errors {
                result.Errors = append(result.Errors, ValidationError{
                    RecordID:  record.ID,
                    Type:      attrError.ErrorType,
                    Message:   attrError.Message,
                    AttrName:  attrError.AttrName,
                })
            }
        }

        // 记录验证通过的数据
        validatedRecord := &ValidatedRecord{
            Original:   record,
            CIType:     ciType,
            Valid:      validateResp.Valid,
            Errors:     validateResp.Errors,
        }
        result.Records = append(result.Records, validatedRecord)
    }

    return result, nil
}
```

### 2. 数据转换引擎

```go
// 数据转换器
type DataTransformer interface {
    Name() string
    Transform(ctx context.Context, record *ValidatedRecord) (*TransformedRecord, error)
}

// 属性转换器
type AttributeTransformer struct {
    typeCache  map[string]*ent.CiType
    attrCache  map[uint64]*ent.Attribute
    logger     logx.Logger
}

func (t *AttributeTransformer) Transform(ctx context.Context, record *ValidatedRecord) (*TransformedRecord, error) {
    transformed := &TransformedRecord{
        TypeID:     record.CIType.ID,
        Attributes: make([]*cmdb.CiAttributeValue, 0),
        Metadata:   record.Original.Metadata,
        Source:     record.Original.Source,
    }

    // 获取类型的所有属性定义
    typeAttrs, err := t.getTypeAttributes(ctx, record.CIType.ID)
    if err != nil {
        return nil, err
    }

    // 转换每个属性值
    for _, typeAttr := range typeAttrs {
        attrName := typeAttr.Edges.Attribute.Name
        if rawValue, exists := record.Original.Attributes[attrName]; exists {
            // 根据属性类型转换值
            convertedValue, err := t.convertAttributeValue(typeAttr.Edges.Attribute, rawValue)
            if err != nil {
                return nil, fmt.Errorf("属性 %s 转换失败: %v", attrName, err)
            }

            transformed.Attributes = append(transformed.Attributes, &cmdb.CiAttributeValue{
                AttrId: typeAttr.AttrID,
                Value:  convertedValue,
            })
        }
    }

    return transformed, nil
}

func (t *AttributeTransformer) convertAttributeValue(attr *ent.Attribute, rawValue interface{}) (string, error) {
    switch attr.ValueType {
    case consts.ValueTypeInt:
        if intVal, ok := rawValue.(int); ok {
            return strconv.Itoa(intVal), nil
        }
        if strVal, ok := rawValue.(string); ok {
            if _, err := strconv.Atoi(strVal); err != nil {
                return "", fmt.Errorf("无法转换为整数: %s", strVal)
            }
            return strVal, nil
        }
    case consts.ValueTypeFloat:
        if floatVal, ok := rawValue.(float64); ok {
            return strconv.FormatFloat(floatVal, 'f', -1, 64), nil
        }
        if strVal, ok := rawValue.(string); ok {
            if _, err := strconv.ParseFloat(strVal, 64); err != nil {
                return "", fmt.Errorf("无法转换为浮点数: %s", strVal)
            }
            return strVal, nil
        }
    case consts.ValueTypeShortText, consts.ValueTypeLongText:
        return fmt.Sprintf("%v", rawValue), nil
    case consts.ValueTypeJSON:
        jsonBytes, err := json.Marshal(rawValue)
        if err != nil {
            return "", fmt.Errorf("无法转换为JSON: %v", err)
        }
        return string(jsonBytes), nil
    case consts.ValueTypeBool:
        if boolVal, ok := rawValue.(bool); ok {
            return strconv.FormatBool(boolVal), nil
        }
        if strVal, ok := rawValue.(string); ok {
            if boolVal, err := strconv.ParseBool(strVal); err == nil {
                return strconv.FormatBool(boolVal), nil
            }
        }
    }
    
    return fmt.Sprintf("%v", rawValue), nil
}
```

### 3. 数据去重引擎

```go
// 数据去重器  
type DataDeduplicator struct {
    svcCtx      *svc.ServiceContext
    strategy    DeduplicationStrategy
    logger      logx.Logger
}

type DeduplicationStrategy string

const (
    StrategySkip     DeduplicationStrategy = "skip"     // 跳过重复数据
    StrategyMerge    DeduplicationStrategy = "merge"    // 合并数据
    StrategyReplace  DeduplicationStrategy = "replace"  // 替换数据
    StrategyError    DeduplicationStrategy = "error"    // 报错
)

func (d *DataDeduplicator) Deduplicate(ctx context.Context, records []*TransformedRecord) ([]*DeduplicatedRecord, error) {
    result := make([]*DeduplicatedRecord, 0)
    
    for _, record := range records {
        // 查找是否存在重复数据
        existing, err := d.findExistingCI(ctx, record)
        if err != nil {
            return nil, err
        }

        deduplicatedRecord := &DeduplicatedRecord{
            Record:   record,
            Action:   ActionCreate,
            Existing: existing,
        }

        if existing != nil {
            // 根据策略处理重复数据
            switch d.strategy {
            case StrategySkip:
                deduplicatedRecord.Action = ActionSkip
            case StrategyMerge:
                deduplicatedRecord.Action = ActionUpdate
                deduplicatedRecord.MergedAttributes = d.mergeAttributes(existing, record)
            case StrategyReplace:
                deduplicatedRecord.Action = ActionReplace
            case StrategyError:
                return nil, fmt.Errorf("发现重复数据: CI类型=%d", record.TypeID)
            }
        }

        result = append(result, deduplicatedRecord)
    }

    return result, nil
}

func (d *DataDeduplicator) findExistingCI(ctx context.Context, record *TransformedRecord) (*ent.Cis, error) {
    // 获取唯一性约束
    ciType, err := d.svcCtx.DB.CiType.Get(ctx, record.TypeID)
    if err != nil {
        return nil, err
    }

    // 根据唯一性约束查找
    for _, uniqueConst := range ciType.UniqueConst {
        query := d.svcCtx.DB.Cis.Query().Where(cis.TypeIDEQ(record.TypeID))
        
        // 构建唯一性查询条件
        hasAllFields := true
        for _, attrID := range uniqueConst.AttrIds {
            found := false
            for _, attr := range record.Attributes {
                if attr.AttrId == attrID {
                    found = true
                    break
                }
            }
            if !found {
                hasAllFields = false
                break
            }
        }

        if hasAllFields {
            // 执行复杂查询以检查唯一性
            existing, err := d.queryByUniqueConstraint(ctx, query, uniqueConst, record)
            if err != nil {
                return nil, err
            }
            if existing != nil {
                return existing, nil
            }
        }
    }

    return nil, nil
}
```

## 🗄️ 异步处理队列设计

### 1. 优先级队列系统

```go
// 消息队列管理器
type MessageQueueManager struct {
    producer   *Producer
    consumers  map[string]*Consumer
    config     *QueueConfig
    logger     logx.Logger
}

type QueueConfig struct {
    HighPriorityQueue    string `json:"high_priority_queue"`    // 高优先级队列
    NormalPriorityQueue  string `json:"normal_priority_queue"`  // 普通优先级队列
    LowPriorityQueue     string `json:"low_priority_queue"`     // 低优先级队列
    DeadLetterQueue      string `json:"dead_letter_queue"`      // 死信队列
    BatchSize            int    `json:"batch_size"`             // 批处理大小
    RetryLimit           int    `json:"retry_limit"`            // 重试次数限制
    RetryDelay           int    `json:"retry_delay"`            // 重试延迟(秒)
}

// 处理消息结构
type ProcessingMessage struct {
    ID          string                 `json:"id"`           // 消息ID
    Type        string                 `json:"type"`         // 消息类型
    Priority    Priority               `json:"priority"`     // 优先级
    Source      string                 `json:"source"`       // 数据源
    Data        []*DeduplicatedRecord  `json:"data"`         // 处理数据
    Metadata    map[string]interface{} `json:"metadata"`     // 元数据
    RetryCount  int                    `json:"retry_count"`  // 重试次数
    Timestamp   time.Time              `json:"timestamp"`    // 时间戳
}

type Priority int

const (
    PriorityHigh   Priority = 1 // 高优先级：实时同步、紧急数据
    PriorityNormal Priority = 2 // 普通优先级：API录入、Excel导入
    PriorityLow    Priority = 3 // 低优先级：自动发现、定时同步
)

// 消息生产者
func (m *MessageQueueManager) PublishProcessingMessage(ctx context.Context, msg *ProcessingMessage) error {
    // 根据优先级选择队列
    queueName := m.getQueueByPriority(msg.Priority)
    
    // 序列化消息
    msgData, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("序列化消息失败: %v", err)
    }

    // 发送到消息队列
    return m.producer.Publish(ctx, queueName, msgData)
}

// 批量消息处理器
type BatchMessageProcessor struct {
    svcCtx       *svc.ServiceContext
    batchSize    int
    flushTimeout time.Duration
    buffer       []*ProcessingMessage
    mutex        sync.Mutex
    logger       logx.Logger
}

func (p *BatchMessageProcessor) ProcessMessage(ctx context.Context, msg *ProcessingMessage) error {
    p.mutex.Lock()
    defer p.mutex.Unlock()

    p.buffer = append(p.buffer, msg)

    // 检查是否需要批量处理
    if len(p.buffer) >= p.batchSize {
        return p.flushBuffer(ctx)
    }

    return nil
}

func (p *BatchMessageProcessor) flushBuffer(ctx context.Context) error {
    if len(p.buffer) == 0 {
        return nil
    }

    // 开启数据库事务
    tx, err := p.svcCtx.DB.Tx(ctx)
    if err != nil {
        return fmt.Errorf("开启事务失败: %v", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    successCount := 0
    errorCount := 0

    // 批量处理消息
    for _, msg := range p.buffer {
        if err := p.processRecords(ctx, tx, msg.Data); err != nil {
            p.logger.Errorf("处理消息失败: %s, 错误: %v", msg.ID, err)
            errorCount++
            // 发送到重试队列或死信队列
            if msg.RetryCount >= 3 {
                p.sendToDeadLetterQueue(msg, err)
            } else {
                p.sendToRetryQueue(msg)
            }
        } else {
            successCount++
        }
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("提交事务失败: %v", err)
    }

    p.logger.Infof("批量处理完成: 成功=%d, 失败=%d", successCount, errorCount)
    
    // 清空缓冲区
    p.buffer = p.buffer[:0]
    
    return nil
}
```

### 2. 失败重试机制

```go
// 重试处理器
type RetryProcessor struct {
    svcCtx      *svc.ServiceContext
    maxRetries  int
    retryDelay  time.Duration
    logger      logx.Logger
}

func (r *RetryProcessor) ProcessRetryMessage(ctx context.Context, msg *ProcessingMessage) error {
    msg.RetryCount++
    
    if msg.RetryCount > r.maxRetries {
        // 超过最大重试次数，发送到死信队列
        return r.sendToDeadLetterQueue(msg, fmt.Errorf("超过最大重试次数: %d", r.maxRetries))
    }

    // 计算退避延迟 (指数退避)
    delay := time.Duration(math.Pow(2, float64(msg.RetryCount))) * r.retryDelay

    r.logger.Infof("消息重试: ID=%s, 重试次数=%d, 延迟=%v", msg.ID, msg.RetryCount, delay)

    // 延迟后重新处理
    time.Sleep(delay)
    
    return r.processMessage(ctx, msg)
}

func (r *RetryProcessor) sendToDeadLetterQueue(msg *ProcessingMessage, err error) error {
    deadMsg := &DeadLetterMessage{
        OriginalMessage: msg,
        FailureReason:   err.Error(),
        FailureTime:     time.Now(),
        RequiresManual:  true,
    }

    deadMsgData, _ := json.Marshal(deadMsg)
    return r.publishToQueue("dead_letter_queue", deadMsgData)
}
```

## 💾 数据存储引擎优化

### 1. 批量写入优化

```go
// 批量存储管理器
type BatchStorageManager struct {
    svcCtx      *svc.ServiceContext
    batchSize   int
    ciBuffer    []*cmdb.CisInfo
    valueBuffer map[string][]*AttributeValue
    mutex       sync.RWMutex
    logger      logx.Logger
}

type AttributeValue struct {
    CIID    uint64
    AttrID  uint64
    Value   string
    Type    string
}

func (b *BatchStorageManager) StoreBatch(ctx context.Context, records []*DeduplicatedRecord) error {
    // 开启事务
    tx, err := b.svcCtx.DB.Tx(ctx)
    if err != nil {
        return fmt.Errorf("开启事务失败: %v", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // 分组处理：创建、更新、跳过
    createRecords := make([]*DeduplicatedRecord, 0)
    updateRecords := make([]*DeduplicatedRecord, 0)

    for _, record := range records {
        switch record.Action {
        case ActionCreate:
            createRecords = append(createRecords, record)
        case ActionUpdate, ActionReplace:
            updateRecords = append(updateRecords, record)
        case ActionSkip:
            continue
        }
    }

    // 批量创建CI
    if len(createRecords) > 0 {
        if err := b.batchCreateCIs(ctx, tx, createRecords); err != nil {
            tx.Rollback()
            return err
        }
    }

    // 批量更新CI
    if len(updateRecords) > 0 {
        if err := b.batchUpdateCIs(ctx, tx, updateRecords); err != nil {
            tx.Rollback()
            return err
        }
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("提交事务失败: %v", err)
    }

    b.logger.Infof("批量存储完成: 创建=%d, 更新=%d", len(createRecords), len(updateRecords))
    return nil
}

func (b *BatchStorageManager) batchCreateCIs(ctx context.Context, tx *ent.Tx, records []*DeduplicatedRecord) error {
    // 批量创建CI基础信息
    ciBuilders := make([]*ent.CisCreate, 0, len(records))
    
    for _, record := range records {
        cisInfo := &cmdb.CisInfo{
            TypeId: &record.Record.TypeID,
            Status: pointy.GetPointer(uint32(1)), // 默认状态：运行中
        }
        
        builder := tx.Cis.Create()
        builder = cis.CisCreateBuilderSetter(builder, cisInfo)
        ciBuilders = append(ciBuilders, builder)
    }

    // 执行批量创建
    createdCIs, err := tx.Cis.CreateBulk(ciBuilders...).Save(ctx)
    if err != nil {
        return fmt.Errorf("批量创建CI失败: %v", err)
    }

    // 批量创建属性值
    for i, record := range records {
        ciID := createdCIs[i].ID
        if err := b.batchCreateAttributes(ctx, tx, ciID, record.Record.Attributes); err != nil {
            return fmt.Errorf("批量创建属性失败: CI ID=%d, 错误=%v", ciID, err)
        }
    }

    return nil
}

func (b *BatchStorageManager) batchCreateAttributes(ctx context.Context, tx *ent.Tx, ciID uint64, attributes []*cmdb.CiAttributeValue) error {
    // 根据属性类型分组
    intValues := make([]*ent.ValueIntegerCreate, 0)
    floatValues := make([]*ent.ValueFloatCreate, 0)
    textValues := make([]*ent.ValueTextCreate, 0)
    jsonValues := make([]*ent.ValueJsonCreate, 0)
    datetimeValues := make([]*ent.ValueDatetimeCreate, 0)

    for _, attr := range attributes {
        // 获取属性定义以确定类型
        attrDef, err := tx.Attribute.Get(ctx, attr.AttrId)
        if err != nil {
            continue // 跳过未找到的属性
        }

        switch attrDef.ValueType {
        case consts.ValueTypeInt:
            if intVal, err := strconv.Atoi(attr.Value); err == nil {
                intValues = append(intValues, tx.ValueInteger.Create().
                    SetCiID(ciID).
                    SetAttrID(attr.AttrId).
                    SetValue(intVal))
            }
        case consts.ValueTypeFloat:
            if floatVal, err := strconv.ParseFloat(attr.Value, 64); err == nil {
                floatValues = append(floatValues, tx.ValueFloat.Create().
                    SetCiID(ciID).
                    SetAttrID(attr.AttrId).
                    SetValue(floatVal))
            }
        case consts.ValueTypeShortText, consts.ValueTypeLongText:
            textValues = append(textValues, tx.ValueText.Create().
                SetCiID(ciID).
                SetAttrID(attr.AttrId).
                SetValue(attr.Value))
        case consts.ValueTypeJSON:
            jsonValues = append(jsonValues, tx.ValueJson.Create().
                SetCiID(ciID).
                SetAttrID(attr.AttrId).
                SetValue(json.RawMessage(attr.Value)))
        case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
            if timeVal, err := parseTimeValue(attr.Value); err == nil {
                datetimeValues = append(datetimeValues, tx.ValueDatetime.Create().
                    SetCiID(ciID).
                    SetAttrID(attr.AttrId).
                    SetValue(timeVal))
            }
        }
    }

    // 执行批量创建
    if len(intValues) > 0 {
        if _, err := tx.ValueInteger.CreateBulk(intValues...).Save(ctx); err != nil {
            return fmt.Errorf("批量创建整数值失败: %v", err)
        }
    }
    
    if len(floatValues) > 0 {
        if _, err := tx.ValueFloat.CreateBulk(floatValues...).Save(ctx); err != nil {
            return fmt.Errorf("批量创建浮点数值失败: %v", err)
        }
    }
    
    if len(textValues) > 0 {
        if _, err := tx.ValueText.CreateBulk(textValues...).Save(ctx); err != nil {
            return fmt.Errorf("批量创建文本值失败: %v", err)
        }
    }

    if len(jsonValues) > 0 {
        if _, err := tx.ValueJson.CreateBulk(jsonValues...).Save(ctx); err != nil {
            return fmt.Errorf("批量创建JSON值失败: %v", err)
        }
    }

    if len(datetimeValues) > 0 {
        if _, err := tx.ValueDatetime.CreateBulk(datetimeValues...).Save(ctx); err != nil {
            return fmt.Errorf("批量创建日期时间值失败: %v", err)
        }
    }

    return nil
}
```

### 2. 缓存同步策略

```go
// 缓存同步管理器
type CacheSyncManager struct {
    redisClient *redis.Client
    svcCtx      *svc.ServiceContext
    config      *CacheConfig
    logger      logx.Logger
}

type CacheConfig struct {
    CITypeTTL       time.Duration `json:"ci_type_ttl"`       // CI类型缓存TTL
    CITTL           time.Duration `json:"ci_ttl"`            // CI缓存TTL
    AttributeTTL    time.Duration `json:"attribute_ttl"`     // 属性缓存TTL
    RelationTTL     time.Duration `json:"relation_ttl"`      // 关系缓存TTL
    BatchUpdateSize int           `json:"batch_update_size"` // 批量更新大小
}

func (c *CacheSyncManager) SyncAfterBatchOperation(ctx context.Context, records []*DeduplicatedRecord) error {
    // 收集需要更新的缓存键
    ciTypeIDs := make(map[uint64]bool)
    ciIDs := make([]uint64, 0)

    for _, record := range records {
        ciTypeIDs[record.Record.TypeID] = true
        if record.Action == ActionUpdate || record.Action == ActionReplace {
            if record.Existing != nil {
                ciIDs = append(ciIDs, record.Existing.ID)
            }
        }
    }

    // 批量删除相关缓存
    pipeline := c.redisClient.Pipeline()

    // 删除CI类型相关缓存
    for typeID := range ciTypeIDs {
        cacheKey := fmt.Sprintf("cmdb:ci_type:%d", typeID)
        pipeline.Del(ctx, cacheKey)
        
        // 删除类型属性缓存
        attrCacheKey := fmt.Sprintf("cmdb:ci_type_attrs:%d", typeID)
        pipeline.Del(ctx, attrCacheKey)
    }

    // 删除CI实例缓存
    for _, ciID := range ciIDs {
        ciCacheKey := fmt.Sprintf("cmdb:ci:%d", ciID)
        pipeline.Del(ctx, ciCacheKey)
        
        // 删除CI关系缓存
        relationCacheKey := fmt.Sprintf("cmdb:ci_relations:%d", ciID)
        pipeline.Del(ctx, relationCacheKey)
    }

    // 执行批量删除
    _, err := pipeline.Exec(ctx)
    if err != nil {
        c.logger.Errorf("批量删除缓存失败: %v", err)
        return err
    }

    c.logger.Infof("缓存同步完成: CI类型=%d个, CI实例=%d个", len(ciTypeIDs), len(ciIDs))
    return nil
}

// 预热关键缓存
func (c *CacheSyncManager) WarmupCache(ctx context.Context, typeIDs []uint64) error {
    for _, typeID := range typeIDs {
        // 预热CI类型信息
        if err := c.warmupCIType(ctx, typeID); err != nil {
            c.logger.Errorf("预热CI类型缓存失败: typeID=%d, error=%v", typeID, err)
        }
        
        // 预热属性定义
        if err := c.warmupTypeAttributes(ctx, typeID); err != nil {
            c.logger.Errorf("预热类型属性缓存失败: typeID=%d, error=%v", typeID, err)
        }
    }
    return nil
}

func (c *CacheSyncManager) warmupCIType(ctx context.Context, typeID uint64) error {
    cacheKey := fmt.Sprintf("cmdb:ci_type:%d", typeID)
    
    // 检查缓存是否存在
    exists, err := c.redisClient.Exists(ctx, cacheKey).Result()
    if err != nil {
        return err
    }
    
    if exists == 0 {
        // 从数据库加载并缓存
        ciType, err := c.svcCtx.DB.CiType.Get(ctx, typeID)
        if err != nil {
            return err
        }
        
        ciTypeData, _ := json.Marshal(ciType)
        return c.redisClient.Set(ctx, cacheKey, ciTypeData, c.config.CITypeTTL).Err()
    }
    
    return nil
}
```

## 📤 数据输出引擎设计

### 1. 统一输出接口

```go
// 数据输出接口
type DataOutputPlugin interface {
    Name() string
    Version() string
    SupportedFormats() []string
    
    // 输出方法
    Export(ctx context.Context, request *ExportRequest) (*ExportResult, error)
    Stream(ctx context.Context, request *StreamRequest) (<-chan *StreamData, error)
    
    // 配置和健康检查
    Configure(config map[string]interface{}) error
    HealthCheck() error
}

// 导出请求结构
type ExportRequest struct {
    ID          string                 `json:"id"`           // 请求ID
    Type        string                 `json:"type"`         // 导出类型
    Format      string                 `json:"format"`       // 输出格式
    Query       *QueryCriteria         `json:"query"`        // 查询条件
    Fields      []string               `json:"fields"`       // 导出字段
    Options     map[string]interface{} `json:"options"`      // 导出选项
    Metadata    map[string]interface{} `json:"metadata"`     // 元数据
    RequestBy   string                 `json:"request_by"`   // 请求人
    Timestamp   time.Time              `json:"timestamp"`    // 请求时间
}

// 查询条件
type QueryCriteria struct {
    CITypes     []uint64               `json:"ci_types"`     // CI类型过滤
    Filters     []QueryFilter          `json:"filters"`      // 属性过滤
    Relations   []RelationFilter       `json:"relations"`    // 关系过滤
    DateRange   *DateRange             `json:"date_range"`   // 时间范围
    Pagination  *PaginationInfo        `json:"pagination"`   // 分页信息
    Sorting     []SortField            `json:"sorting"`      // 排序规则
}

// 导出结果
type ExportResult struct {
    ID           string                 `json:"id"`            // 结果ID
    Status       string                 `json:"status"`        // 导出状态
    Format       string                 `json:"format"`        // 输出格式
    Size         int64                  `json:"size"`          // 文件大小
    RecordCount  int                    `json:"record_count"`  // 记录数量
    FilePath     string                 `json:"file_path"`     // 文件路径
    DownloadURL  string                 `json:"download_url"`  // 下载链接
    Metadata     map[string]interface{} `json:"metadata"`      // 结果元数据
    CreateTime   time.Time              `json:"create_time"`   // 创建时间
    ExpireTime   time.Time              `json:"expire_time"`   // 过期时间
}
```

### 2. REST API输出插件

```go
// REST API输出插件
type RESTOutputPlugin struct {
    svcCtx     *svc.ServiceContext
    queryCache *QueryCache
    config     *RESTOutputConfig
    logger     logx.Logger
}

type RESTOutputConfig struct {
    MaxPageSize     int           `json:"max_page_size"`     // 最大分页大小
    DefaultPageSize int           `json:"default_page_size"` // 默认分页大小
    CacheTTL        time.Duration `json:"cache_ttl"`         // 缓存TTL
    EnableCache     bool          `json:"enable_cache"`      // 是否启用缓存
    RateLimit       int           `json:"rate_limit"`        // 限流配置
}

func (p *RESTOutputPlugin) Export(ctx context.Context, request *ExportRequest) (*ExportResult, error) {
    // 构建查询
    query := p.buildQuery(request.Query)
    
    // 执行查询
    results, totalCount, err := p.executeQuery(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("查询执行失败: %v", err)
    }

    // 根据请求格式序列化数据
    var data []byte
    switch request.Format {
    case "json":
        data, err = p.serializeToJSON(results, request.Fields)
    case "xml":
        data, err = p.serializeToXML(results, request.Fields)
    case "csv":
        data, err = p.serializeToCSV(results, request.Fields)
    default:
        return nil, fmt.Errorf("不支持的格式: %s", request.Format)
    }

    if err != nil {
        return nil, fmt.Errorf("数据序列化失败: %v", err)
    }

    // 生成结果
    result := &ExportResult{
        ID:          request.ID,
        Status:      "completed",
        Format:      request.Format,
        Size:        int64(len(data)),
        RecordCount: len(results),
        CreateTime:  time.Now(),
        ExpireTime:  time.Now().Add(24 * time.Hour), // 24小时后过期
        Metadata: map[string]interface{}{
            "total_count": totalCount,
            "page_size":   len(results),
            "query_time":  time.Since(request.Timestamp).Milliseconds(),
        },
    }

    // 如果是大文件，保存到临时文件
    if len(data) > 10*1024*1024 { // 超过10MB
        filePath, downloadURL, err := p.saveToFile(request.ID, request.Format, data)
        if err != nil {
            return nil, fmt.Errorf("保存文件失败: %v", err)
        }
        result.FilePath = filePath
        result.DownloadURL = downloadURL
    } else {
        // 小文件直接返回数据
        result.Metadata["data"] = string(data)
    }

    return result, nil
}

func (p *RESTOutputPlugin) serializeToJSON(results []*EnrichedCI, fields []string) ([]byte, error) {
    output := make([]map[string]interface{}, 0, len(results))
    
    for _, ci := range results {
        item := make(map[string]interface{})
        
        // 基础字段
        if len(fields) == 0 || p.containsField(fields, "id") {
            item["id"] = ci.ID
        }
        if len(fields) == 0 || p.containsField(fields, "type_name") {
            item["type_name"] = ci.TypeName
        }
        if len(fields) == 0 || p.containsField(fields, "status") {
            item["status"] = ci.Status
        }
        if len(fields) == 0 || p.containsField(fields, "create_time") {
            item["create_time"] = ci.CreateTime.Format(time.RFC3339)
        }
        if len(fields) == 0 || p.containsField(fields, "update_time") {
            item["update_time"] = ci.UpdateTime.Format(time.RFC3339)
        }

        // 动态属性
        if len(fields) == 0 || p.containsField(fields, "attributes") {
            attributes := make(map[string]interface{})
            for _, attr := range ci.Attributes {
                attributes[attr.Name] = attr.Value
            }
            item["attributes"] = attributes
        }

        // 关系信息
        if len(fields) == 0 || p.containsField(fields, "relations") {
            relations := make([]map[string]interface{}, 0)
            for _, rel := range ci.Relations {
                relations = append(relations, map[string]interface{}{
                    "type":      rel.Type,
                    "target_id": rel.TargetID,
                    "direction": rel.Direction,
                })
            }
            item["relations"] = relations
        }

        output = append(output, item)
    }

    return json.MarshalIndent(output, "", "  ")
}
```

### 3. Excel导出插件

```go
// Excel导出插件
type ExcelOutputPlugin struct {
    svcCtx    *svc.ServiceContext
    config    *ExcelOutputConfig
    templates map[string]*ExcelTemplate
    logger    logx.Logger
}

type ExcelOutputConfig struct {
    MaxRows        int               `json:"max_rows"`         // 最大行数
    Templates      map[string]string `json:"templates"`        // 模板路径
    DefaultStyle   *ExcelStyle       `json:"default_style"`    // 默认样式
    OutputPath     string            `json:"output_path"`      // 输出路径
}

type ExcelTemplate struct {
    Name          string            `json:"name"`           // 模板名称
    FilePath      string            `json:"file_path"`      // 模板文件路径
    SheetMappings []SheetMapping    `json:"sheet_mappings"` // 工作表映射
    Styles        map[string]*ExcelStyle `json:"styles"`   // 样式定义
}

type SheetMapping struct {
    SheetName     string            `json:"sheet_name"`     // 工作表名称
    CITypes       []uint64          `json:"ci_types"`       // 支持的CI类型
    HeaderRow     int               `json:"header_row"`     // 表头行号
    DataStartRow  int               `json:"data_start_row"` // 数据开始行号
    ColumnMap     map[string]string `json:"column_map"`     // 列映射
    MaxRows       int               `json:"max_rows"`       // 最大行数
}

func (p *ExcelOutputPlugin) Export(ctx context.Context, request *ExportRequest) (*ExportResult, error) {
    // 获取模板
    templateName := "default"
    if tpl, exists := request.Options["template"].(string); exists {
        templateName = tpl
    }

    template, exists := p.templates[templateName]
    if !exists {
        return nil, fmt.Errorf("模板不存在: %s", templateName)
    }

    // 创建Excel文件
    file := excelize.NewFile()
    defer file.Close()

    // 查询数据
    results, err := p.queryData(ctx, request.Query)
    if err != nil {
        return nil, fmt.Errorf("查询数据失败: %v", err)
    }

    // 按CI类型分组数据
    groupedData := p.groupByCIType(results)

    // 生成工作表
    for _, sheetMapping := range template.SheetMappings {
        if err := p.generateSheet(file, sheetMapping, groupedData, template); err != nil {
            return nil, fmt.Errorf("生成工作表失败: %s, 错误: %v", sheetMapping.SheetName, err)
        }
    }

    // 保存文件
    fileName := fmt.Sprintf("cmdb_export_%s_%s.xlsx", request.ID, time.Now().Format("20060102_150405"))
    filePath := filepath.Join(p.config.OutputPath, fileName)
    
    if err := file.SaveAs(filePath); err != nil {
        return nil, fmt.Errorf("保存Excel文件失败: %v", err)
    }

    // 获取文件信息
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("获取文件信息失败: %v", err)
    }

    result := &ExportResult{
        ID:          request.ID,
        Status:      "completed",
        Format:      "xlsx",
        Size:        fileInfo.Size(),
        RecordCount: len(results),
        FilePath:    filePath,
        CreateTime:  time.Now(),
        ExpireTime:  time.Now().Add(7 * 24 * time.Hour), // 7天后过期
        Metadata: map[string]interface{}{
            "template":    templateName,
            "sheet_count": len(template.SheetMappings),
            "file_name":   fileName,
        },
    }

    return result, nil
}

func (p *ExcelOutputPlugin) generateSheet(file *excelize.File, mapping SheetMapping, 
    groupedData map[uint64][]*EnrichedCI, template *ExcelTemplate) error {
    
    // 创建工作表
    sheetIndex, err := file.NewSheet(mapping.SheetName)
    if err != nil {
        return err
    }

    // 设置表头
    headers := make([]string, 0)
    for _, field := range mapping.ColumnMap {
        headers = append(headers, field)
    }

    for col, header := range headers {
        cell := fmt.Sprintf("%s%d", p.getColumnName(col), mapping.HeaderRow)
        file.SetCellValue(mapping.SheetName, cell, header)
        
        // 应用表头样式
        if style, exists := template.Styles["header"]; exists {
            file.SetCellStyle(mapping.SheetName, cell, cell, p.convertStyle(style))
        }
    }

    // 填充数据
    currentRow := mapping.DataStartRow
    for _, ciTypeID := range mapping.CITypes {
        if ciList, exists := groupedData[ciTypeID]; exists {
            for _, ci := range ciList {
                if currentRow-mapping.DataStartRow >= mapping.MaxRows {
                    break // 超出最大行数限制
                }

                col := 0
                for sourceField, _ := range mapping.ColumnMap {
                    cell := fmt.Sprintf("%s%d", p.getColumnName(col), currentRow)
                    value := p.getFieldValue(ci, sourceField)
                    file.SetCellValue(mapping.SheetName, cell, value)
                    col++
                }
                currentRow++
            }
        }
    }

    // 设置工作表为活跃状态
    file.SetActiveSheet(sheetIndex)
    
    return nil
}
```

### 4. WebSocket实时推送插件

```go
// WebSocket实时推送插件
type WebSocketOutputPlugin struct {
    hub     *WebSocketHub
    config  *WebSocketConfig
    logger  logx.Logger
}

type WebSocketConfig struct {
    MaxConnections    int           `json:"max_connections"`     // 最大连接数
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`  // 心跳间隔
    MessageBuffer     int           `json:"message_buffer"`      // 消息缓冲区大小
    EnableCompression bool          `json:"enable_compression"`  // 是否启用压缩
}

type WebSocketHub struct {
    clients    map[*WebSocketClient]bool
    broadcast  chan []byte
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    mutex      sync.RWMutex
}

type WebSocketClient struct {
    ID           string
    conn         *websocket.Conn
    send         chan []byte
    subscriptions map[string]*Subscription
    lastPing     time.Time
}

type Subscription struct {
    ID        string         `json:"id"`         // 订阅ID
    Type      string         `json:"type"`       // 订阅类型：ci_change/relation_change/discovery
    CITypes   []uint64       `json:"ci_types"`   // 关注的CI类型
    Filters   []QueryFilter  `json:"filters"`    // 过滤条件
    CreatedAt time.Time      `json:"created_at"` // 创建时间
}

func (p *WebSocketOutputPlugin) Stream(ctx context.Context, request *StreamRequest) (<-chan *StreamData, error) {
    dataChan := make(chan *StreamData, p.config.MessageBuffer)
    
    // 创建订阅
    subscription := &Subscription{
        ID:        request.ID,
        Type:      request.Type,
        CITypes:   request.CITypes,
        Filters:   request.Filters,
        CreatedAt: time.Now(),
    }

    // 启动数据流
    go p.startDataStream(ctx, subscription, dataChan)
    
    return dataChan, nil
}

func (p *WebSocketOutputPlugin) startDataStream(ctx context.Context, sub *Subscription, dataChan chan<- *StreamData) {
    defer close(dataChan)

    // 创建变更监听器
    changeListener := p.createChangeListener(sub)
    defer changeListener.Close()

    for {
        select {
        case <-ctx.Done():
            return
        case change := <-changeListener.Changes():
            if p.matchesSubscription(change, sub) {
                streamData := &StreamData{
                    ID:        uuid.New().String(),
                    Type:      change.Type,
                    Action:    change.Action,
                    Data:      change.Data,
                    Timestamp: time.Now(),
                }
                
                select {
                case dataChan <- streamData:
                case <-ctx.Done():
                    return
                }
            }
        }
    }
}

// 变更监听器
type ChangeListener struct {
    subscription *Subscription
    changes      chan *ChangeEvent
    stopChan     chan struct{}
    watcher      *DatabaseWatcher
}

type ChangeEvent struct {
    Type      string                 `json:"type"`       // 变更类型
    Action    string                 `json:"action"`     // 操作类型：create/update/delete
    CIID      uint64                 `json:"ci_id"`      // CI ID
    CITypeID  uint64                 `json:"ci_type_id"` // CI类型ID
    Data      map[string]interface{} `json:"data"`       // 变更数据
    Timestamp time.Time              `json:"timestamp"`  // 变更时间
}

func (p *WebSocketOutputPlugin) createChangeListener(sub *Subscription) *ChangeListener {
    listener := &ChangeListener{
        subscription: sub,
        changes:      make(chan *ChangeEvent, 100),
        stopChan:     make(chan struct{}),
    }

    // 启动数据库变更监听
    go listener.startWatching()
    
    return listener
}
```

## 📊 监控与可观测性

### 1. 性能监控

```go
// 性能监控管理器
type PerformanceMonitor struct {
    metrics     *MetricsCollector
    config      *MonitorConfig
    alerts      *AlertManager
    logger      logx.Logger
}

type MonitorConfig struct {
    MetricsInterval   time.Duration `json:"metrics_interval"`    // 指标收集间隔
    AlertThresholds   AlertThresholds `json:"alert_thresholds"` // 告警阈值
    RetentionPeriod   time.Duration `json:"retention_period"`   // 数据保留期
    EnableDetailLogs  bool          `json:"enable_detail_logs"` // 是否启用详细日志
}

type AlertThresholds struct {
    ErrorRate          float64 `json:"error_rate"`           // 错误率阈值
    ResponseTime       int64   `json:"response_time"`        // 响应时间阈值(ms)
    QueueDepth         int     `json:"queue_depth"`          // 队列深度阈值
    MemoryUsage        float64 `json:"memory_usage"`         // 内存使用率阈值
    DatabaseConnUsage  float64 `json:"database_conn_usage"`  // 数据库连接使用率阈值
}

// 核心指标结构
type IngestionMetrics struct {
    // 输入指标
    TotalRequests        int64     `json:"total_requests"`         // 总请求数
    SuccessfulRequests   int64     `json:"successful_requests"`    // 成功请求数
    FailedRequests       int64     `json:"failed_requests"`        // 失败请求数
    RequestRate          float64   `json:"request_rate"`           // 请求速率(req/s)
    
    // 处理指标
    ProcessingLatency    int64     `json:"processing_latency"`     // 处理延迟(ms)
    ValidationErrors     int64     `json:"validation_errors"`      // 验证错误数
    TransformationErrors int64     `json:"transformation_errors"`  // 转换错误数
    DeduplicationHits    int64     `json:"deduplication_hits"`     // 去重命中数
    
    // 存储指标
    StorageLatency       int64     `json:"storage_latency"`        // 存储延迟(ms)
    DatabaseConnections  int       `json:"database_connections"`   // 数据库连接数
    BatchSize            int       `json:"batch_size"`             // 批处理大小
    
    // 队列指标
    QueueDepth           int       `json:"queue_depth"`            // 队列深度
    MessageProcessRate   float64   `json:"message_process_rate"`   // 消息处理速率
    DeadLetterCount      int64     `json:"dead_letter_count"`      // 死信数量
    
    // 系统指标
    MemoryUsage          float64   `json:"memory_usage"`           // 内存使用率
    CPUUsage             float64   `json:"cpu_usage"`              // CPU使用率
    DiskUsage            float64   `json:"disk_usage"`             // 磁盘使用率
    
    Timestamp            time.Time `json:"timestamp"`              // 指标时间戳
}

func (m *PerformanceMonitor) CollectMetrics(ctx context.Context) (*IngestionMetrics, error) {
    metrics := &IngestionMetrics{
        Timestamp: time.Now(),
    }

    // 收集输入指标
    if err := m.collectInputMetrics(ctx, metrics); err != nil {
        return nil, fmt.Errorf("收集输入指标失败: %v", err)
    }

    // 收集处理指标
    if err := m.collectProcessingMetrics(ctx, metrics); err != nil {
        return nil, fmt.Errorf("收集处理指标失败: %v", err)
    }

    // 收集存储指标
    if err := m.collectStorageMetrics(ctx, metrics); err != nil {
        return nil, fmt.Errorf("收集存储指标失败: %v", err)
    }

    // 收集队列指标
    if err := m.collectQueueMetrics(ctx, metrics); err != nil {
        return nil, fmt.Errorf("收集队列指标失败: %v", err)
    }

    // 收集系统指标
    if err := m.collectSystemMetrics(ctx, metrics); err != nil {
        return nil, fmt.Errorf("收集系统指标失败: %v", err)
    }

    // 检查告警条件
    m.checkAlerts(metrics)

    return metrics, nil
}

func (m *PerformanceMonitor) checkAlerts(metrics *IngestionMetrics) {
    // 错误率告警
    if metrics.TotalRequests > 0 {
        errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
        if errorRate > m.config.AlertThresholds.ErrorRate {
            m.alerts.TriggerAlert("high_error_rate", map[string]interface{}{
                "current_rate": errorRate,
                "threshold":    m.config.AlertThresholds.ErrorRate,
                "failed_count": metrics.FailedRequests,
                "total_count":  metrics.TotalRequests,
            })
        }
    }

    // 响应时间告警
    if metrics.ProcessingLatency > m.config.AlertThresholds.ResponseTime {
        m.alerts.TriggerAlert("high_response_time", map[string]interface{}{
            "current_latency": metrics.ProcessingLatency,
            "threshold":       m.config.AlertThresholds.ResponseTime,
        })
    }

    // 队列深度告警
    if metrics.QueueDepth > m.config.AlertThresholds.QueueDepth {
        m.alerts.TriggerAlert("high_queue_depth", map[string]interface{}{
            "current_depth": metrics.QueueDepth,
            "threshold":     m.config.AlertThresholds.QueueDepth,
        })
    }

    // 内存使用率告警
    if metrics.MemoryUsage > m.config.AlertThresholds.MemoryUsage {
        m.alerts.TriggerAlert("high_memory_usage", map[string]interface{}{
            "current_usage": metrics.MemoryUsage,
            "threshold":     m.config.AlertThresholds.MemoryUsage,
        })
    }
}
```

### 2. 错误追踪与日志聚合

```go
// 错误追踪管理器
type ErrorTracker struct {
    logAggregator *LogAggregator
    errorStore    *ErrorStore
    config        *ErrorTrackingConfig
    logger        logx.Logger
}

type ErrorTrackingConfig struct {
    EnableDetailTrace bool          `json:"enable_detail_trace"` // 是否启用详细追踪
    ErrorRetention    time.Duration `json:"error_retention"`     // 错误保留期
    SampleRate        float64       `json:"sample_rate"`         // 采样率
    AlertRules        []AlertRule   `json:"alert_rules"`         // 告警规则
}

type AlertRule struct {
    Name        string        `json:"name"`         // 规则名称
    Condition   string        `json:"condition"`    // 告警条件
    Threshold   int           `json:"threshold"`    // 阈值
    Window      time.Duration `json:"window"`       // 时间窗口
    Severity    string        `json:"severity"`     // 严重级别
    Enabled     bool          `json:"enabled"`      // 是否启用
}

// 错误记录结构
type ErrorRecord struct {
    ID           string                 `json:"id"`            // 错误ID
    Type         string                 `json:"type"`          // 错误类型
    Message      string                 `json:"message"`       // 错误消息
    StackTrace   string                 `json:"stack_trace"`   // 堆栈跟踪
    Context      map[string]interface{} `json:"context"`       // 错误上下文
    Source       string                 `json:"source"`        // 错误源
    Severity     string                 `json:"severity"`      // 严重级别
    RequestID    string                 `json:"request_id"`    // 请求ID
    UserID       string                 `json:"user_id"`       // 用户ID
    Timestamp    time.Time              `json:"timestamp"`     // 发生时间
    Fingerprint  string                 `json:"fingerprint"`   // 错误指纹
    Count        int                    `json:"count"`         // 错误次数
    FirstSeen    time.Time              `json:"first_seen"`    // 首次出现
    LastSeen     time.Time              `json:"last_seen"`     // 最后出现
}

func (e *ErrorTracker) TrackError(ctx context.Context, err error, context map[string]interface{}) {
    // 生成错误记录
    record := &ErrorRecord{
        ID:          uuid.New().String(),
        Type:        reflect.TypeOf(err).String(),
        Message:     err.Error(),
        StackTrace:  e.getStackTrace(),
        Context:     context,
        Source:      e.getSource(ctx),
        Severity:    e.determineSeverity(err),
        RequestID:   e.getRequestID(ctx),
        UserID:      e.getUserID(ctx),
        Timestamp:   time.Now(),
        Fingerprint: e.generateFingerprint(err, context),
    }

    // 检查是否是重复错误
    existing, err := e.errorStore.FindByFingerprint(record.Fingerprint)
    if err == nil && existing != nil {
        // 更新重复错误统计
        existing.Count++
        existing.LastSeen = record.Timestamp
        e.errorStore.Update(existing)
    } else {
        // 新错误
        record.Count = 1
        record.FirstSeen = record.Timestamp
        record.LastSeen = record.Timestamp
        e.errorStore.Store(record)
    }

    // 检查告警规则
    e.checkErrorAlerts(record)

    // 发送到日志聚合器
    e.logAggregator.SendError(record)
}

// 日志聚合器
type LogAggregator struct {
    elasticClient *elasticsearch.Client
    config        *LogConfig
    buffer        []*LogEntry
    mutex         sync.Mutex
    logger        logx.Logger
}

type LogConfig struct {
    IndexPrefix     string        `json:"index_prefix"`      // 索引前缀
    BufferSize      int           `json:"buffer_size"`       // 缓冲区大小
    FlushInterval   time.Duration `json:"flush_interval"`    // 刷新间隔
    RetentionDays   int           `json:"retention_days"`    // 日志保留天数
    EnableSampling  bool          `json:"enable_sampling"`   // 是否启用采样
    SampleRate      float64       `json:"sample_rate"`       // 采样率
}

type LogEntry struct {
    Level       string                 `json:"level"`        // 日志级别
    Message     string                 `json:"message"`      // 日志消息
    Component   string                 `json:"component"`    // 组件名称
    RequestID   string                 `json:"request_id"`   // 请求ID
    UserID      string                 `json:"user_id"`      // 用户ID
    Source      string                 `json:"source"`       // 日志源
    Context     map[string]interface{} `json:"context"`      // 上下文信息
    Timestamp   time.Time              `json:"timestamp"`    // 时间戳
    TraceID     string                 `json:"trace_id"`     // 链路追踪ID
    SpanID      string                 `json:"span_id"`      // Span ID
}

func (l *LogAggregator) SendError(record *ErrorRecord) {
    entry := &LogEntry{
        Level:     "ERROR",
        Message:   record.Message,
        Component: "cmdb-ingestion",
        RequestID: record.RequestID,
        UserID:    record.UserID,
        Source:    record.Source,
        Context: map[string]interface{}{
            "error_id":     record.ID,
            "error_type":   record.Type,
            "fingerprint":  record.Fingerprint,
            "stack_trace":  record.StackTrace,
            "severity":     record.Severity,
            "count":        record.Count,
        },
        Timestamp: record.Timestamp,
    }

    l.addToBuffer(entry)
}

func (l *LogAggregator) addToBuffer(entry *LogEntry) {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    l.buffer = append(l.buffer, entry)
    
    if len(l.buffer) >= l.config.BufferSize {
        go l.flushBuffer()
    }
}

func (l *LogAggregator) flushBuffer() {
    l.mutex.Lock()
    entries := make([]*LogEntry, len(l.buffer))
    copy(entries, l.buffer)
    l.buffer = l.buffer[:0]
    l.mutex.Unlock()

    if len(entries) == 0 {
        return
    }

    // 批量发送到Elasticsearch
    indexName := fmt.Sprintf("%s-%s", l.config.IndexPrefix, time.Now().Format("2006.01.02"))
    
    for _, entry := range entries {
        // 应用采样
        if l.config.EnableSampling && rand.Float64() > l.config.SampleRate {
            continue
        }

        doc, _ := json.Marshal(entry)
        _, err := l.elasticClient.Index(
            indexName,
            bytes.NewReader(doc),
        )
        
        if err != nil {
            l.logger.Errorf("发送日志到Elasticsearch失败: %v", err)
        }
    }

    l.logger.Infof("批量发送日志完成: %d条", len(entries))
}
```

## 🚀 性能优化策略

### 1. 并发处理优化

```go
// 并发处理管理器
type ConcurrencyManager struct {
    workerPool    *WorkerPool
    rateLimiter   *RateLimiter
    loadBalancer  *LoadBalancer
    config        *ConcurrencyConfig
    logger        logx.Logger
}

type ConcurrencyConfig struct {
    MaxWorkers      int           `json:"max_workers"`       // 最大工作协程数
    QueueSize       int           `json:"queue_size"`        // 队列大小
    BatchSize       int           `json:"batch_size"`        // 批处理大小
    MaxConcurrency  int           `json:"max_concurrency"`   // 最大并发数
    RateLimit       int           `json:"rate_limit"`        // 速率限制(req/s)
    WorkerTimeout   time.Duration `json:"worker_timeout"`    // 工作协程超时
    HealthCheckInt  time.Duration `json:"health_check_int"`  // 健康检查间隔
}

// 工作协程池
type WorkerPool struct {
    workers   []*Worker
    taskQueue chan *Task
    resultCh  chan *TaskResult
    stopCh    chan struct{}
    config    *ConcurrencyConfig
    logger    logx.Logger
    wg        sync.WaitGroup
}

type Worker struct {
    ID       int
    taskCh   chan *Task
    resultCh chan *TaskResult
    stopCh   chan struct{}
    status   WorkerStatus
    lastTask time.Time
}

type WorkerStatus int

const (
    WorkerIdle    WorkerStatus = 0 // 空闲
    WorkerBusy    WorkerStatus = 1 // 忙碌
    WorkerStopped WorkerStatus = 2 // 已停止
)

// 任务定义
type Task struct {
    ID        string                 `json:"id"`         // 任务ID
    Type      string                 `json:"type"`       // 任务类型
    Priority  Priority               `json:"priority"`   // 优先级
    Data      interface{}            `json:"data"`       // 任务数据
    Context   map[string]interface{} `json:"context"`    // 任务上下文
    Timeout   time.Duration          `json:"timeout"`    // 超时时间
    CreatedAt time.Time              `json:"created_at"` // 创建时间
}

type TaskResult struct {
    TaskID    string                 `json:"task_id"`    // 任务ID
    Status    string                 `json:"status"`     // 执行状态
    Result    interface{}            `json:"result"`     // 执行结果
    Error     error                  `json:"error"`      // 错误信息
    Duration  time.Duration          `json:"duration"`   // 执行时长
    Timestamp time.Time              `json:"timestamp"`  // 完成时间
}

func NewWorkerPool(config *ConcurrencyConfig, logger logx.Logger) *WorkerPool {
    pool := &WorkerPool{
        workers:   make([]*Worker, config.MaxWorkers),
        taskQueue: make(chan *Task, config.QueueSize),
        resultCh:  make(chan *TaskResult, config.QueueSize),
        stopCh:    make(chan struct{}),
        config:    config,
        logger:    logger,
    }

    // 创建工作协程
    for i := 0; i < config.MaxWorkers; i++ {
        worker := &Worker{
            ID:       i,
            taskCh:   make(chan *Task, 1),
            resultCh: pool.resultCh,
            stopCh:   make(chan struct{}),
            status:   WorkerIdle,
        }
        pool.workers[i] = worker
        
        pool.wg.Add(1)
        go pool.runWorker(worker)
    }

    // 启动任务分发器
    go pool.taskDispatcher()

    return pool
}

func (p *WorkerPool) runWorker(worker *Worker) {
    defer p.wg.Done()

    for {
        select {
        case task := <-worker.taskCh:
            worker.status = WorkerBusy
            worker.lastTask = time.Now()
            
            result := p.executeTask(worker, task)
            
            worker.status = WorkerIdle
            worker.resultCh <- result
            
        case <-worker.stopCh:
            worker.status = WorkerStopped
            return
        }
    }
}

func (p *WorkerPool) executeTask(worker *Worker, task *Task) *TaskResult {
    start := time.Now()
    
    result := &TaskResult{
        TaskID:    task.ID,
        Timestamp: time.Now(),
    }

    // 设置超时上下文
    ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
    defer cancel()

    // 执行任务
    switch task.Type {
    case "data_validation":
        result.Result, result.Error = p.executeValidation(ctx, task.Data)
    case "data_transformation":
        result.Result, result.Error = p.executeTransformation(ctx, task.Data)
    case "data_storage":
        result.Result, result.Error = p.executeStorage(ctx, task.Data)
    case "cache_update":
        result.Result, result.Error = p.executeCacheUpdate(ctx, task.Data)
    default:
        result.Error = fmt.Errorf("未知任务类型: %s", task.Type)
    }

    result.Duration = time.Since(start)
    
    if result.Error != nil {
        result.Status = "failed"
        p.logger.Errorf("工作协程 %d 执行任务失败: %s, 错误: %v", worker.ID, task.ID, result.Error)
    } else {
        result.Status = "completed"
        p.logger.Infof("工作协程 %d 执行任务成功: %s, 耗时: %v", worker.ID, task.ID, result.Duration)
    }

    return result
}

func (p *WorkerPool) taskDispatcher() {
    for {
        select {
        case task := <-p.taskQueue:
            // 查找空闲工作协程
            worker := p.findIdleWorker()
            if worker != nil {
                select {
                case worker.taskCh <- task:
                    // 任务分发成功
                case <-time.After(1 * time.Second):
                    // 分发超时，重新放回队列
                    go func() {
                        p.taskQueue <- task
                    }()
                }
            } else {
                // 没有空闲工作协程，重新放回队列
                go func() {
                    time.Sleep(100 * time.Millisecond)
                    p.taskQueue <- task
                }()
            }
        case <-p.stopCh:
            return
        }
    }
}

func (p *WorkerPool) findIdleWorker() *Worker {
    for _, worker := range p.workers {
        if worker.status == WorkerIdle {
            return worker
        }
    }
    return nil
}
```

### 2. 负载均衡与限流

```go
// 负载均衡器
type LoadBalancer struct {
    nodes    []*ProcessingNode
    strategy BalanceStrategy
    metrics  *LoadMetrics
    config   *LoadBalancerConfig
    logger   logx.Logger
}

type BalanceStrategy string

const (
    RoundRobin    BalanceStrategy = "round_robin"    // 轮询
    LeastLoad     BalanceStrategy = "least_load"     // 最少负载
    WeightedRound BalanceStrategy = "weighted_round" // 加权轮询
    IPHash        BalanceStrategy = "ip_hash"        // IP哈希
)

type ProcessingNode struct {
    ID          string        `json:"id"`           // 节点ID
    Address     string        `json:"address"`      // 节点地址
    Weight      int           `json:"weight"`       // 权重
    MaxLoad     int           `json:"max_load"`     // 最大负载
    CurrentLoad int           `json:"current_load"` // 当前负载
    Status      NodeStatus    `json:"status"`       // 节点状态
    LastCheck   time.Time     `json:"last_check"`   // 最后检查时间
    ResponseTime time.Duration `json:"response_time"` // 响应时间
}

type NodeStatus int

const (
    NodeActive   NodeStatus = 1 // 活跃
    NodeInactive NodeStatus = 0 // 不活跃
    NodeFailed   NodeStatus = -1 // 失败
)

type LoadBalancerConfig struct {
    Strategy        BalanceStrategy `json:"strategy"`         // 负载均衡策略
    HealthCheckInt  time.Duration   `json:"health_check_int"` // 健康检查间隔
    FailureThreshold int            `json:"failure_threshold"` // 失败阈值
    RecoveryTime    time.Duration   `json:"recovery_time"`    // 恢复时间
}

func (lb *LoadBalancer) SelectNode(ctx context.Context, request *ProcessingRequest) (*ProcessingNode, error) {
    activeNodes := lb.getActiveNodes()
    if len(activeNodes) == 0 {
        return nil, fmt.Errorf("没有可用的处理节点")
    }

    var selectedNode *ProcessingNode
    
    switch lb.strategy {
    case RoundRobin:
        selectedNode = lb.roundRobinSelect(activeNodes)
    case LeastLoad:
        selectedNode = lb.leastLoadSelect(activeNodes)
    case WeightedRound:
        selectedNode = lb.weightedRoundSelect(activeNodes)
    case IPHash:
        selectedNode = lb.ipHashSelect(activeNodes, request.ClientIP)
    default:
        selectedNode = activeNodes[0]
    }

    // 更新负载统计
    selectedNode.CurrentLoad++
    
    return selectedNode, nil
}

func (lb *LoadBalancer) leastLoadSelect(nodes []*ProcessingNode) *ProcessingNode {
    var selected *ProcessingNode
    minLoad := int(^uint(0) >> 1) // 最大整数值

    for _, node := range nodes {
        if node.CurrentLoad < minLoad {
            minLoad = node.CurrentLoad
            selected = node
        }
    }

    return selected
}

// 限流器
type RateLimiter struct {
    limiters map[string]*TokenBucket
    config   *RateLimitConfig
    mutex    sync.RWMutex
    logger   logx.Logger
}

type RateLimitConfig struct {
    GlobalRateLimit int           `json:"global_rate_limit"` // 全局限流
    UserRateLimit   int           `json:"user_rate_limit"`   // 用户限流
    IPRateLimit     int           `json:"ip_rate_limit"`     // IP限流
    WindowSize      time.Duration `json:"window_size"`       // 时间窗口
    BurstSize       int           `json:"burst_size"`        // 突发大小
}

type TokenBucket struct {
    capacity     int           // 桶容量
    tokens       int           // 当前令牌数
    refillRate   int           // 填充速率
    lastRefill   time.Time     // 最后填充时间
    mutex        sync.Mutex
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, rateType string) bool {
    rl.mutex.RLock()
    bucket, exists := rl.limiters[key]
    rl.mutex.RUnlock()

    if !exists {
        // 创建新的令牌桶
        bucket = rl.createTokenBucket(rateType)
        rl.mutex.Lock()
        rl.limiters[key] = bucket
        rl.mutex.Unlock()
    }

    return bucket.TakeToken()
}

func (tb *TokenBucket) TakeToken() bool {
    tb.mutex.Lock()
    defer tb.mutex.Unlock()

    now := time.Now()
    
    // 计算需要添加的令牌数
    elapsed := now.Sub(tb.lastRefill)
    tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
    
    if tokensToAdd > 0 {
        tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
        tb.lastRefill = now
    }

    // 尝试消费令牌
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }

    return false
}
```

### 3. 缓存同步策略

```go
// 缓存同步管理器
type CacheSyncManager struct {
    redisClient *redis.Client
    svcCtx      *svc.ServiceContext
    config      *CacheConfig
    logger      logx.Logger
}

type CacheConfig struct {
    CITypeTTL       time.Duration `json:"ci_type_ttl"`       // CI类型缓存TTL
    CITTL           time.Duration `json:"ci_ttl"`            // CI缓存TTL
    AttributeTTL    time.Duration `json:"attribute_ttl"`     // 属性缓存TTL
    RelationTTL     time.Duration `json:"relation_ttl"`      // 关系缓存TTL
    BatchUpdateSize int           `json:"batch_update_size"` // 批量更新大小
}

func (c *CacheSyncManager) SyncAfterBatchOperation(ctx context.Context, records []*DeduplicatedRecord) error {
    // 收集需要更新的缓存键
    ciTypeIDs := make(map[uint64]bool)
    ciIDs := make([]uint64, 0)

    for _, record := range records {
        ciTypeIDs[record.Record.TypeID] = true
        if record.Action == ActionUpdate || record.Action == ActionReplace {
            if record.Existing != nil {
                ciIDs = append(ciIDs, record.Existing.ID)
            }
        }
    }

    // 批量删除相关缓存
    pipeline := c.redisClient.Pipeline()

    // 删除CI类型相关缓存
    for typeID := range ciTypeIDs {
        cacheKey := fmt.Sprintf("cmdb:ci_type:%d", typeID)
        pipeline.Del(ctx, cacheKey)
        
        // 删除类型属性缓存
        attrCacheKey := fmt.Sprintf("cmdb:ci_type_attrs:%d", typeID)
        pipeline.Del(ctx, attrCacheKey)
    }

    // 删除CI实例缓存
    for _, ciID := range ciIDs {
        ciCacheKey := fmt.Sprintf("cmdb:ci:%d", ciID)
        pipeline.Del(ctx, ciCacheKey)
        
        // 删除CI关系缓存
        relationCacheKey := fmt.Sprintf("cmdb:ci_relations:%d", ciID)
        pipeline.Del(ctx, relationCacheKey)
    }

    // 执行批量删除
    _, err := pipeline.Exec(ctx)
    if err != nil {
        c.logger.Errorf("批量删除缓存失败: %v", err)
        return err
    }

    c.logger.Infof("缓存同步完成: CI类型=%d个, CI实例=%d个", len(ciTypeIDs), len(ciIDs))
    return nil
}

// 预热关键缓存
func (c *CacheSyncManager) WarmupCache(ctx context.Context, typeIDs []uint64) error {
    for _, typeID := range typeIDs {
        // 预热CI类型信息
        if err := c.warmupCIType(ctx, typeID); err != nil {
            c.logger.Errorf("预热CI类型缓存失败: typeID=%d, error=%v", typeID, err)
        }
        
        // 预热属性定义
        if err := c.warmupTypeAttributes(ctx, typeID); err != nil {
            c.logger.Errorf("预热类型属性缓存失败: typeID=%d, error=%v", typeID, err)
        }
    }
    return nil
}

func (c *CacheSyncManager) warmupCIType(ctx context.Context, typeID uint64) error {
    cacheKey := fmt.Sprintf("cmdb:ci_type:%d", typeID)
    
    // 检查缓存是否存在
    exists, err := c.redisClient.Exists(ctx, cacheKey).Result()
    if err != nil {
        return err
    }
    
    if exists == 0 {
        // 从数据库加载并缓存
        ciType, err := c.svcCtx.DB.CiType.Get(ctx, typeID)
        if err != nil {
            return err
        }
        
        ciTypeData, _ := json.Marshal(ciType)
        return c.redisClient.Set(ctx, cacheKey, ciTypeData, c.config.CITypeTTL).Err()
    }
    
    return nil
}
```

---
📅 实现计划与里程碑

### 第一阶段：基础架构搭建 (4周)

**目标**: 建立核心架构和基础功能

**里程碑**:

#### Week 1
- [ ] 完成插件化输入接口设计
- [ ] 实现Excel导入插件
- [ ] 建立统一数据接入网关

#### Week 2
- [ ] 实现API接口输入插件
- [ ] 完成数据验证引擎集成
- [ ] 建立基础的异步处理队列

#### Week 3
- [ ] 实现数据转换引擎
- [ ] 完成数据去重功能
- [ ] 建立批量存储机制

#### Week 4
- [ ] 实现基础的REST API输出
- [ ] 建立监控指标收集
- [ ] 完成基础功能端到端测试

**关键交付物**:
- 基础插件架构代码
- Excel和API输入插件
- 数据处理管道
- 基础监控面板

### 第二阶段：高级功能开发 (4周)

**目标**: 实现高级处理功能和性能优化

**里程碑**:

#### Week 5-6
- [ ] 实现自动发现插件
- [ ] 完成WebSocket实时推送
- [ ] 实现Excel导出功能
- [ ] 优化数据库批量操作

#### Week 7-8
- [ ] 实现并发处理优化
- [ ] 完成负载均衡机制
- [ ] 建立缓存同步策略
- [ ] 实现失败重试机制

**关键交付物**:
- 自动发现插件
- 完整的输出引擎
- 性能优化方案
- 高可用架构

### 第三阶段：企业级特性 (3周)

**目标**: 实现企业级安全、监控和运维特性

**里程碑**:

#### Week 9
- [ ] 实现完整的错误追踪系统
- [ ] 建立日志聚合机制
- [ ] 完成告警规则配置

#### Week 10
- [ ] 实现限流和熔断机制
- [ ] 完成容器化部署配置
- [ ] 建立CI/CD流水线

#### Week 11
- [ ] 完成性能基准测试
- [ ] 建立灾备方案
- [ ] 完成文档和培训材料

**关键交付物**:
- 完整的监控告警体系
- 容器化部署方案
- 性能测试报告
- 运维手册

### 第四阶段：生产部署与优化 (2周)

**目标**: 生产环境部署和性能调优

**里程碑**:

#### Week 12
- [ ] 生产环境部署
- [ ] 性能调优和压力测试
- [ ] 建立运维流程

#### Week 13
- [ ] 用户培训和文档完善
- [ ] 生产环境稳定性验证
- [ ] 项目交付和知识转移

**关键交付物**:
- 生产环境运行系统
- 性能调优报告
- 完整的项目文档
- 运维培训材料

## 🎯 性能基准与测试策略

### 性能目标

| 指标类型 | 目标值 | 测试条件 |
|---------|--------|---------|
| 吞吐量 | 10,000 records/min | Excel导入，1000行/文件 |
| 响应时间 | P95 < 2s | API接口单条记录录入 |
| 并发处理 | 1000 req/s | API批量录入 |
| 错误率 | < 0.1% | 正常负载下 |
| 可用性 | 99.9% | 7×24小时运行 |
| 数据一致性 | 100% | 并发写入场景 |


## 📋 总结与展望

### 设计亮点

1. **高扩展性**: 插件化架构支持灵活的功能扩展
2. **高性能**: 异步处理、批量操作、并发优化
3. **高可用**: 负载均衡、故障恢复、监控告警
4. **企业级**: 完整的安全、审计、运维体系

### 技术创新

1. **统一数据接入**: 多种输入方式的统一处理管道
2. **智能去重**: 基于业务规则的数据去重策略
3. **实时处理**: WebSocket实时推送和变更通知
4. **监控体系**: 全链路的性能监控和错误追踪

### 下一步规划

#### 短期目标 (3个月)
- [ ] 完成基础功能开发和测试
- [ ] 建立生产环境部署能力
- [ ] 实现核心性能指标
- [ ] 建立基础的运维能力

#### 中期目标 (6个月)
- [ ] 支持更多数据源和格式
- [ ] 实现AI驱动的数据质量检查
- [ ] 建立数据血缘追踪能力
- [ ] 支持多租户和权限管理

#### 长期目标 (12个月)
- [ ] 实现跨数据中心的分布式部署
- [ ] 支持流式数据处理
- [ ] 建立智能运维能力
- [ ] 实现与主流ITSM系统的深度集成

### 风险与应对

| 风险类型 | 风险描述 | 应对策略 |
|---------|---------|---------|
| 技术风险 | 大数据量处理性能瓶颈 | 分片处理、缓存优化、异步队列 |
| 业务风险 | 数据一致性问题 | 事务管理、幂等设计、补偿机制 |
| 运维风险 | 系统复杂度过高 | 自动化部署、监控告警、文档完善 |
| 安全风险 | 数据泄露和权限问题 | 数据脱敏、访问控制、审计日志 |


















# 多阶段构建优化镜像大小
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmdb-ingestion-engine .

# 最终镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 cmdb && \
    adduser -D -s /bin/sh -u 1000 -G cmdb cmdb

WORKDIR /home/cmdb

# 复制构建产物
COPY --from=builder /app/cmdb-ingestion-engine .
COPY --from=builder /app/etc/ ./etc/
COPY --from=builder /app/templates/ ./templates/

# 设置权限
RUN chown -R cmdb:cmdb /home/cmdb

# 切换到非root用户
USER cmdb

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD ./cmdb-ingestion-engine healthcheck || exit 1

# 暴露端口
EXPOSE 8080 9090

# 启动命令
CMD ["./cmdb-ingestion-engine", "-f", "etc/config.yaml"]

---
# cmdb-ingestion-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cmdb-ingestion-engine
  namespace: cmdb-system
  labels:
    app: cmdb-ingestion-engine
    component: data-processing
    version: v1.1.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: cmdb-ingestion-engine
  template:
    metadata:
      labels:
        app: cmdb-ingestion-engine
        component: data-processing
        version: v1.1.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: cmdb-ingestion-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
      containers:
      - name: cmdb-ingestion-engine
        image: newbee/cmdb-ingestion-engine:v1.1.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP
        env:
        - name: CONFIG_PATH
          value: "/etc/cmdb/config.yaml"
        - name: LOG_LEVEL
          value: "info"
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: cmdb-mysql-secret
              key: password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: cmdb-redis-secret
              key: password
        - name: MQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: cmdb-mq-secret
              key: password
        volumeMounts:
        - name: config
          mountPath: /etc/cmdb
          readOnly: true
        - name: templates
          mountPath: /app/templates
          readOnly: true
        - name: temp-storage
          mountPath: /tmp/cmdb
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
      volumes:
      - name: config
        configMap:
          name: cmdb-ingestion-config
      - name: templates
        configMap:
          name: cmdb-export-templates
      - name: temp-storage
        emptyDir:
          sizeLimit: 10Gi

---
# cmdb-ingestion-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: cmdb-ingestion-engine-svc
  namespace: cmdb-system
  labels:
    app: cmdb-ingestion-engine
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  selector:
    app: cmdb-ingestion-engine

---
# cmdb-ingestion-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmdb-ingestion-config
  namespace: cmdb-system
data:
  config.yaml: |
    # CMDB数据录入引擎配置
    name: cmdb-ingestion-engine
    host: 0.0.0.0
    port: 8080
    
    # 数据库配置
    mysql:
      master:
        host: mysql-master-svc
        port: 3306
        database: cmdb
        username: cmdb_user
        password: ${MYSQL_PASSWORD}
        max_open_conns: 100
        max_idle_conns: 10
        conn_max_lifetime: 3600
      slave:
        host: mysql-slave-svc
        port: 3306
        database: cmdb
        username: cmdb_readonly
        password: ${MYSQL_PASSWORD}
        max_open_conns: 50
        max_idle_conns: 5
        conn_max_lifetime: 3600
    
    # Redis配置
    redis:
      cluster:
        - redis-cluster-0.redis-cluster-svc:6379
        - redis-cluster-1.redis-cluster-svc:6379
        - redis-cluster-2.redis-cluster-svc:6379
      password: ${REDIS_PASSWORD}
      pool_size: 20
      dial_timeout: 5s
      read_timeout: 3s
      write_timeout: 3s
    
    # 消息队列配置
    message_queue:
      type: rabbitmq
      address: amqp://cmdb_user:${MQ_PASSWORD}@rabbitmq-svc:5672/
      high_priority_queue: cmdb_high_priority
      normal_priority_queue: cmdb_normal_priority
      low_priority_queue: cmdb_low_priority
      dead_letter_queue: cmdb_dead_letter
      batch_size: 100
      retry_limit: 3
      retry_delay: 5
    
    # 输入插件配置
    input_plugins:
      excel:
        max_file_size: 50 # MB
        max_rows: 50000
        required_fields: ["name", "type"]
      api:
        batch_size: 1000
        rate_limit: 100
        auth_required: true
      discovery:
        agent_timeout: 30s
        update_strategy: merge
        conflict_policy: merge
    
    # 处理引擎配置
    processing:
      max_workers: 10
      queue_size: 10000
      batch_size: 100
      max_concurrency: 20
      rate_limit: 1000
      worker_timeout: 30s
    
    # 缓存配置
    cache:
      ci_type_ttl: 24h
      ci_ttl: 1h
      attribute_ttl: 6h
      relation_ttl: 30m
      batch_update_size: 1000
    
    # 监控配置
    monitoring:
      metrics_interval: 30s
      enable_detail_logs: true
      alert_thresholds:
        error_rate: 0.05
        response_time: 5000 # ms
        queue_depth: 5000
        memory_usage: 0.8
        database_conn_usage: 0.9
    
    # 日志配置
    logging:
      level: info
      output: stdout
      enable_sampling: true
      sample_rate: 0.1
      
---
# cmdb-ingestion-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: cmdb-ingestion-engine-hpa
  namespace: cmdb-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cmdb-ingestion-engine
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max

---
# prometheus-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cmdb-ingestion-alerts
  namespace: cmdb-system
  labels:
    app: cmdb-ingestion-engine
    prometheus: kube-prometheus
    role: alert-rules
spec:
  groups:
  - name: cmdb-ingestion.rules
    rules:
    # 高错误率告警
    - alert: CMDBIngestionHighErrorRate
      expr: |
        (
          rate(cmdb_ingestion_requests_failed_total[5m]) /
          rate(cmdb_ingestion_requests_total[5m])
        ) > 0.05
      for: 2m
      labels:
        severity: warning
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入引擎错误率过高"
        description: "CMDB录入引擎在过去5分钟的错误率为 {{ $value | humanizePercentage }}，超过了5%的阈值"
    
    # 高响应时间告警
    - alert: CMDBIngestionHighLatency
      expr: |
        histogram_quantile(0.95,
          rate(cmdb_ingestion_request_duration_seconds_bucket[5m])
        ) > 5
      for: 3m
      labels:
        severity: warning
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入引擎响应时间过高"
        description: "CMDB录入引擎95%响应时间为 {{ $value }}s，超过了5s的阈值"
    
    # 队列积压告警
    - alert: CMDBIngestionQueueBacklog
      expr: |
        cmdb_ingestion_queue_depth > 5000
      for: 5m
      labels:
        severity: critical
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入队列严重积压"
        description: "CMDB录入队列深度为 {{ $value }}，超过了5000的阈值，可能影响数据处理时效性"
    
    # 内存使用率告警
    - alert: CMDBIngestionHighMemoryUsage
      expr: |
        (
          container_memory_working_set_bytes{pod=~"cmdb-ingestion-engine-.*"} /
          container_spec_memory_limit_bytes{pod=~"cmdb-ingestion-engine-.*"}
        ) > 0.85
      for: 5m
      labels:
        severity: warning
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入引擎内存使用率过高"
        description: "Pod {{ $labels.pod }} 内存使用率为 {{ $value | humanizePercentage }}，超过了85%的阈值"
    
    # 数据库连接池告警
    - alert: CMDBIngestionDBConnectionHigh
      expr: |
        cmdb_ingestion_db_connections_active / cmdb_ingestion_db_connections_max > 0.9
      for: 3m
      labels:
        severity: critical
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入引擎数据库连接池使用率过高"
        description: "数据库连接池使用率为 {{ $value | humanizePercentage }}，超过了90%的阈值"
    
    # 死信队列告警
    - alert: CMDBIngestionDeadLetterMessages
      expr: |
        increase(cmdb_ingestion_dead_letter_messages_total[10m]) > 10
      for: 0m
      labels:
        severity: critical
        component: cmdb-ingestion-engine
      annotations:
        summary: "CMDB录入引擎死信消息增加"
        description: "在过去10分钟内有 {{ $value }} 条消息进入死信队列，需要人工介入处理"

---
# grafana-dashboard.json
{
  "dashboard": {
    "id": null,
    "title": "CMDB录入与输出引擎监控",
    "tags": ["cmdb", "ingestion", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "请求总览",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(cmdb_ingestion_requests_total[5m])",
            "legendFormat": "请求速率 (req/s)"
          },
          {
            "expr": "cmdb_ingestion_requests_total",
            "legendFormat": "总请求数"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "reqps"
          }
        }
      },
      {
        "id": 2,
        "title": "错误率趋势",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(cmdb_ingestion_requests_failed_total[5m]) / rate(cmdb_ingestion_requests_total[5m])",
            "legendFormat": "错误率"
          }
        ],
        "yAxes": [
          {
            "unit": "percentunit",
            "max": 1,
            "min": 0
          }
        ]
      },
      {
        "id": 3,
        "title": "响应时间分布",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(cmdb_ingestion_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(cmdb_ingestion_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(cmdb_ingestion_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P99"
          }
        ],
        "yAxes": [
          {
            "unit": "s"
          }
        ]
      },
      {
        "id": 4,
        "title": "队列状态",
        "type": "graph",
        "targets": [
          {
            "expr": "cmdb_ingestion_queue_depth",
            "legendFormat": "队列深度"
          },
          {
            "expr": "rate(cmdb_ingestion_messages_processed_total[5m])",
            "legendFormat": "处理速率"
          }
        ]
      },
      {
        "id": 5,
        "title": "资源使用情况",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{pod=~\"cmdb-ingestion-engine-.*\"}[5m])",
            "legendFormat": "CPU使用率 - {{ $labels.pod }}"
          },
          {
            "expr": "container_memory_working_set_bytes{pod=~\"cmdb-ingestion-engine-.*\"} / 1024 / 1024",
            "legendFormat": "内存使用 (MB) - {{ $labels.pod }}"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}