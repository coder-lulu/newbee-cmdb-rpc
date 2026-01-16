package types

import "time"

// ExecutionStatus 执行状态枚举
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// DiscoveryResult 发现结果
type DiscoveryResult struct {
	ExecutionID  string          `json:"execution_id"`
	ConfigID     uint64          `json:"config_id"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      *time.Time      `json:"end_time,omitempty"`
	Status       ExecutionStatus `json:"status"`
	TotalRecords int64           `json:"total_records"`
	SuccessCount int64           `json:"success_count"`
	FailedCount  int64           `json:"failed_count"`
	SkippedCount int64           `json:"skipped_count"`
	CreatedCIs   int64           `json:"created_cis"`
	UpdatedCIs   int64           `json:"updated_cis"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// TransformedCIData 转换后的CI数据
type TransformedCIData struct {
	CITypeID         uint64                 `json:"ci_type_id"`
	SourceID         string                 `json:"source_id"`         // 来源系统的ID
	UniqueKey        string                 `json:"unique_key"`        // 唯一标识
	Attributes       map[string]interface{} `json:"attributes"`        // 属性数据
	Metadata         map[string]interface{} `json:"metadata"`          // 元数据
	SourceData       map[string]interface{} `json:"source_data"`       // 原始源数据
	IsUpdate         bool                   `json:"is_update"`         // 是否为更新操作
	ExistingCIID     *uint64                `json:"existing_ci_id,omitempty"` // 已存在的CI ID
	TransformStats   TransformationStats    `json:"transform_stats"`   // 转换统计信息
	ValidationErrors []ValidationError     `json:"validation_errors"` // 验证错误
}

// TransformationStats 转换统计信息
type TransformationStats struct {
	TotalFields       int   `json:"total_fields"`
	MappedFields      int   `json:"mapped_fields"`
	TransformedFields int   `json:"transformed_fields"`
	ValidationErrors  int   `json:"validation_errors"`
	ProcessingTimeMs  int64 `json:"processing_time_ms"`
}

// PersistResult 持久化结果
type PersistResult struct {
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	SkippedCount int64   `json:"skipped_count"`
	CreatedCIs   int64   `json:"created_cis"`
	UpdatedCIs   int64   `json:"updated_cis"`
	Errors       []error `json:"errors,omitempty"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// TransformationError 转换错误
type TransformationError struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	Value       string `json:"value"`
	Error       string `json:"error"`
}

// DiscoveryError 发现过程中的错误
type DiscoveryError struct {
	Stage       string `json:"stage"`       // 阶段：connecting, discovering, transforming, persisting
	Component   string `json:"component"`   // 组件名称
	Message     string `json:"message"`     // 错误消息
	Details     string `json:"details"`     // 详细信息
	Recoverable bool   `json:"recoverable"` // 是否可恢复
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	ConnectTime    time.Duration `json:"connect_time"`
	DiscoveryTime  time.Duration `json:"discovery_time"`
	TransformTime  time.Duration `json:"transform_time"`
	PersistTime    time.Duration `json:"persist_time"`
	TotalTime      time.Duration `json:"total_time"`
	MemoryUsage    int64         `json:"memory_usage"`    // 内存使用量(字节)
	RecordsPerSec  float64       `json:"records_per_sec"` // 处理速度(记录/秒)
}