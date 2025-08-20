package input

import (
	"time"
)

// InputData 输入数据统一格式
type InputData struct {
	ID         string                 `json:"id"`       // 请求ID
	Type       string                 `json:"type"`     // 输入类型(excel/api/discovery)
	Source     string                 `json:"source"`   // 数据源标识
	Data       []byte                 `json:"data"`     // 原始数据
	Config     map[string]interface{} `json:"config"`   // 配置参数
	Metadata   map[string]string      `json:"metadata"` // 元数据
	BatchID    string                 `json:"batch_id"` // 批次ID
	CreateTime time.Time              `json:"create_time"`
	CreatedBy  string                 `json:"created_by"` // 创建者
}

// RawAssetData 原始资产数据
type RawAssetData struct {
	ID         string                 `json:"id"`           // 临时ID
	CITypeID   uint64                 `json:"ci_type_id"`   // CI类型ID
	CITypeName string                 `json:"ci_type_name"` // CI类型名称
	Attributes map[string]interface{} `json:"attributes"`   // 原始属性数据
	Relations  []*RelationData        `json:"relations"`    // 关系数据
	Tags       []string               `json:"tags"`         // 标签
	Source     string                 `json:"source"`       // 数据来源
	BatchID    string                 `json:"batch_id"`     // 批次ID
	LineNumber int                    `json:"line_number"`  // 行号(错误定位)
	Metadata   map[string]interface{} `json:"metadata"`     // 元数据
}

// ProcessedAssetData 处理后的资产数据
type ProcessedAssetData struct {
	*RawAssetData
	ValidationResult *ValidationResult `json:"validation_result"`
	TransformResult  *TransformResult  `json:"transform_result"`
	PersistResult    *PersistResult    `json:"persist_result"`
	ProcessTime      time.Time         `json:"process_time"`
	Status           string            `json:"status"` // validated/transformed/cleaned/stored
	ProcessorChain   []string          `json:"processor_chain"`
}

// ParsedData 解析后数据结构
type ParsedData struct {
	Records   []*RawAssetData        `json:"records"`   // 资产记录
	Relations []*RelationData        `json:"relations"` // 关系记录
	Metadata  map[string]interface{} `json:"metadata"`  // 解析元数据
	Errors    []ParseError           `json:"errors"`    // 解析错误
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Errors      []*ValidationError     `json:"errors"`
	Warnings    []*ValidationWarning   `json:"warnings"`
	ValidCount  int                    `json:"valid_count"`
	ErrorCount  int                    `json:"error_count"`
	ProcessTime time.Duration          `json:"process_time"`
	Summary     map[string]interface{} `json:"summary"`
}

// ValidationError 验证错误
type ValidationError struct {
	AssetID       string `json:"asset_id"`
	LineNumber    int    `json:"line_number"`
	ColumnName    string `json:"column_name"`
	ErrorType     string `json:"error_type"`
	ErrorMsg      string `json:"error_msg"`
	ExpectedValue string `json:"expected_value"`
	ActualValue   string `json:"actual_value"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	AssetID    string `json:"asset_id"`
	LineNumber int    `json:"line_number"`
	ColumnName string `json:"column_name"`
	WarningMsg string `json:"warning_msg"`
}

// RelationData 关系数据
type RelationData struct {
	ID             string                 `json:"id"`
	SourceAssetID  string                 `json:"source_asset_id"`
	TargetAssetID  string                 `json:"target_asset_id"`
	RelationType   string                 `json:"relation_type"`
	RelationTypeID uint64                 `json:"relation_type_id"`
	Attributes     map[string]interface{} `json:"attributes"`
	Source         string                 `json:"source"`
	LineNumber     int                    `json:"line_number"`
}

// ParseError 解析错误
type ParseError struct {
	LineNumber   int    `json:"line_number"`
	ColumnNumber int    `json:"column_number"`
	ErrorType    string `json:"error_type"`
	ErrorMsg     string `json:"error_msg"`
	RawData      string `json:"raw_data"`
}

// ProcessingData 处理管道中的数据
type ProcessingData struct {
	Assets     []*ProcessedAssetData  `json:"assets"`
	BatchInfo  *BatchInfo             `json:"batch_info"`
	Context    map[string]interface{} `json:"context"`
	ErrorCount int                    `json:"error_count"`
}

// BatchInfo 批次信息
type BatchInfo struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Source     string                 `json:"source"`
	TotalCount int                    `json:"total_count"`
	CreateTime time.Time              `json:"create_time"`
	CreatedBy  string                 `json:"created_by"`
	Config     map[string]interface{} `json:"config"`
	Status     string                 `json:"status"`
}

// ExcelConfig Excel导入配置
type ExcelConfig struct {
	MaxFileSize    int64             `json:"max_file_size"`   // 最大文件大小 (MB)
	MaxRows        int               `json:"max_rows"`        // 最大行数
	SheetMapping   map[string]string `json:"sheet_mapping"`   // 工作表映射
	ColumnMapping  map[string]string `json:"column_mapping"`  // 列映射
	RequiredFields []string          `json:"required_fields"` // 必填字段
	SkipEmptyRows  bool              `json:"skip_empty_rows"` // 跳过空行
	HeaderRow      int               `json:"header_row"`      // 表头行号
	DataStartRow   int               `json:"data_start_row"`  // 数据起始行号
}

// APIConfig API导入配置
type APIConfig struct {
	MaxBatchSize   int               `json:"max_batch_size"`  // 最大批量大小
	RateLimit      int               `json:"rate_limit"`      // 限流配置
	Timeout        time.Duration     `json:"timeout"`         // 超时时间
	RetryCount     int               `json:"retry_count"`     // 重试次数
	RequiredFields []string          `json:"required_fields"` // 必填字段
	FieldMapping   map[string]string `json:"field_mapping"`   // 字段映射
}

// DiscoveryConfig 自动发现配置
type DiscoveryConfig struct {
	DiscoveryRules []*DiscoveryRule       `json:"discovery_rules"` // 发现规则
	ScanInterval   time.Duration          `json:"scan_interval"`   // 扫描间隔
	MaxTargets     int                    `json:"max_targets"`     // 最大目标数
	Timeout        time.Duration          `json:"timeout"`         // 超时时间
	AgentConfig    map[string]interface{} `json:"agent_config"`    // Agent配置
}

// DiscoveryRule 发现规则
type DiscoveryRule struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`       // server/network/service
	Condition map[string]interface{} `json:"condition"`  // 发现条件
	Mapping   map[string]string      `json:"mapping"`    // 属性映射
	CITypeID  uint64                 `json:"ci_type_id"` // 目标CI类型
	Enabled   bool                   `json:"enabled"`
	Priority  int                    `json:"priority"`
}

// AssetImportTask 资产导入任务
type AssetImportTask struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"`     // import/export/batch_operation
	Status        string        `json:"status"`   // pending/processing/completed/failed
	Priority      Priority      `json:"priority"` // high/normal/low
	InputData     *InputData    `json:"input_data"`
	Config        *TaskConfig   `json:"config"`
	Progress      *TaskProgress `json:"progress"`
	Result        *TaskResult   `json:"result"`
	CreateTime    time.Time     `json:"create_time"`
	UpdateTime    time.Time     `json:"update_time"`
	CreateBy      string        `json:"create_by"`
	EstimatedTime time.Duration `json:"estimated_time"`
	ActualTime    time.Duration `json:"actual_time"`
}

// Priority 优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
)

// TaskConfig 任务配置
type TaskConfig struct {
	BatchSize       int                    `json:"batch_size"`
	MaxRetries      int                    `json:"max_retries"`
	EnableNotify    bool                   `json:"enable_notify"`
	NotifyConfig    map[string]interface{} `json:"notify_config"`
	ProcessorConfig map[string]interface{} `json:"processor_config"`
}

// TaskProgress 任务进度
type TaskProgress struct {
	TotalItems         int           `json:"total_items"`
	ProcessedItems     int           `json:"processed_items"`
	SuccessItems       int           `json:"success_items"`
	FailedItems        int           `json:"failed_items"`
	CurrentStep        string        `json:"current_step"`
	Percentage         float64       `json:"percentage"`
	StartTime          time.Time     `json:"start_time"`
	LastUpdateTime     time.Time     `json:"last_update_time"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
}

// TaskResult 任务结果
type TaskResult struct {
	Success        bool                   `json:"success"`
	ProcessedCount int                    `json:"processed_count"`
	SuccessCount   int                    `json:"success_count"`
	FailedCount    int                    `json:"failed_count"`
	Errors         []*TaskError           `json:"errors"`
	Warnings       []*TaskWarning         `json:"warnings"`
	Summary        map[string]interface{} `json:"summary"`
	OutputData     interface{}            `json:"output_data"`
	ProcessTime    time.Duration          `json:"process_time"`
}

// TaskError 任务错误
type TaskError struct {
	AssetID    string    `json:"asset_id"`
	LineNumber int       `json:"line_number"`
	ErrorType  string    `json:"error_type"`
	ErrorMsg   string    `json:"error_msg"`
	Timestamp  time.Time `json:"timestamp"`
}

// TaskWarning 任务警告
type TaskWarning struct {
	AssetID    string    `json:"asset_id"`
	LineNumber int       `json:"line_number"`
	WarningMsg string    `json:"warning_msg"`
	Timestamp  time.Time `json:"timestamp"`
}

// AdapterProcessResult 适配器处理结果 (用于registry.go)
type AdapterProcessResult struct {
	ProcessedData  *ProcessingData `json:"processed_data"`
	ProcessedCount int             `json:"processed_count"`
	TotalErrors    int             `json:"total_errors"`
	ProcessTime    time.Duration   `json:"process_time"`
}

// TransformResult 转换结果
type TransformResult struct {
	Success         bool                   `json:"success"`
	TransformedCIS  interface{}            `json:"transformed_cis"` // 转换后的CIS对象
	AttributesCount int                    `json:"attributes_count"`
	ErrorCount      int                    `json:"error_count"`
	ProcessTime     time.Duration          `json:"process_time"`
	Summary         map[string]interface{} `json:"summary"`
	Errors          []*TransformError      `json:"errors"`
}

// TransformError 转换错误
type TransformError struct {
	AssetID       string `json:"asset_id"`
	AttributeName string `json:"attribute_name"`
	ErrorType     string `json:"error_type"`
	ErrorMsg      string `json:"error_msg"`
	OriginalValue string `json:"original_value"`
}

// PersistResult 持久化结果
type PersistResult struct {
	Success     bool                   `json:"success"`
	CIID        uint64                 `json:"ci_id"` // 保存后的CI ID
	ErrorCount  int                    `json:"error_count"`
	ProcessTime time.Duration          `json:"process_time"`
	Summary     map[string]interface{} `json:"summary"`
	Errors      []*PersistError        `json:"errors"`
}

// PersistError 持久化错误
type PersistError struct {
	AssetID       string `json:"asset_id"`
	ErrorType     string `json:"error_type"`
	ErrorMsg      string `json:"error_msg"`
	TableName     string `json:"table_name"`
	AttributeName string `json:"attribute_name"`
}
