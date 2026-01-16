package config

import (
	"time"
)

// AdapterConfig 适配器配置结构 - 从cmdb.yaml读取
type AdapterConfig struct {
	HealthCheck *HealthCheckConfig `yaml:"HealthCheck" json:"HealthCheck,optional"`
	Performance *PerformanceConfig `yaml:"Performance" json:"Performance,optional"`
	Excel       *ExcelSettings     `yaml:"Excel" json:"Excel,optional"`
	API         *APISettings       `yaml:"API" json:"API,optional"`
	Discovery   *DiscoverySettings `yaml:"Discovery" json:"Discovery,optional"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enable          bool          `yaml:"Enable" json:"Enable,optional"`
	Interval        time.Duration `yaml:"Interval" json:"Interval,optional"`
	DatabaseCheck   bool          `yaml:"DatabaseCheck" json:"DatabaseCheck,optional"`
	RedisCheck      bool          `yaml:"RedisCheck" json:"RedisCheck,optional"`
	MemoryThreshold int           `yaml:"MemoryThreshold" json:"MemoryThreshold,optional"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	ConcurrentWorkers int           `yaml:"ConcurrentWorkers" json:"ConcurrentWorkers,optional"`
	BatchSize         int           `yaml:"BatchSize" json:"BatchSize,optional"`
	MemoryLimit       string        `yaml:"MemoryLimit" json:"MemoryLimit,optional"`
	ProcessTimeout    time.Duration `yaml:"ProcessTimeout" json:"ProcessTimeout,optional"`
}

// ExcelSettings Excel适配器设置
type ExcelSettings struct {
	MaxRows       int    `yaml:"MaxRows" json:"MaxRows,optional"`
	MaxFileSize   string `yaml:"MaxFileSize" json:"MaxFileSize,optional"`
	StreamProcess bool   `yaml:"StreamProcess" json:"StreamProcess,optional"`
	ChunkSize     string `yaml:"ChunkSize" json:"ChunkSize,optional"`
}

// APISettings API适配器设置
type APISettings struct {
	MaxBatchSize  int           `yaml:"MaxBatchSize" json:"MaxBatchSize,optional"`
	Timeout       time.Duration `yaml:"Timeout" json:"Timeout,optional"`
	MaxConcurrent int           `yaml:"MaxConcurrent" json:"MaxConcurrent,optional"`
}

// DiscoverySettings 自动发现适配器设置
type DiscoverySettings struct {
	Disabled       bool   `yaml:"Disabled" json:"Disabled,optional"`
	MigrationGuide string `yaml:"MigrationGuide" json:"MigrationGuide,optional"`
}

// DefaultAdapterConfig 返回默认配置
func DefaultAdapterConfig() *AdapterConfig {
	return &AdapterConfig{
		HealthCheck: &HealthCheckConfig{
			Enable:          true,
			Interval:        30 * time.Second,
			DatabaseCheck:   true,
			RedisCheck:      true,
			MemoryThreshold: 90,
		},
		Performance: &PerformanceConfig{
			ConcurrentWorkers: 8,
			BatchSize:         1000,
			MemoryLimit:       "100MB",
			ProcessTimeout:    30 * time.Second,
		},
		Excel: &ExcelSettings{
			MaxRows:       100000,
			MaxFileSize:   "50MB",
			StreamProcess: true,
			ChunkSize:     "1MB",
		},
		API: &APISettings{
			MaxBatchSize:  5000,
			Timeout:       30 * time.Second,
			MaxConcurrent: 16,
		},
		Discovery: &DiscoverySettings{
			Disabled:       true,
			MigrationGuide: "请使用Agent服务进行网络发现",
		},
	}
}

// IsHealthCheckEnabled 检查是否启用健康检查
func (c *AdapterConfig) IsHealthCheckEnabled() bool {
	return c.HealthCheck != nil && c.HealthCheck.Enable
}

// IsStreamProcessEnabled 检查是否启用流式处理
func (c *AdapterConfig) IsStreamProcessEnabled() bool {
	return c.Excel != nil && c.Excel.StreamProcess
}

// IsDiscoveryDisabled 检查自动发现是否被禁用
func (c *AdapterConfig) IsDiscoveryDisabled() bool {
	return c.Discovery != nil && c.Discovery.Disabled
}
