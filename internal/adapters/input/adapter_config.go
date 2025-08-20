package input

import (
	"time"
)

// AdapterConfig 适配器配置结构 - 从cmdb.yaml读取
type AdapterConfig struct {
	HealthCheck *HealthCheckConfig `yaml:"HealthCheck"`
	Performance *PerformanceConfig `yaml:"Performance"`
	Excel       *ExcelSettings     `yaml:"Excel"`
	API         *APISettings       `yaml:"API"`
	Discovery   *DiscoverySettings `yaml:"Discovery"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enable          bool          `yaml:"Enable"`          // 是否启用健康检查
	Interval        time.Duration `yaml:"Interval"`        // 检查间隔
	DatabaseCheck   bool          `yaml:"DatabaseCheck"`   // 是否检查数据库
	RedisCheck      bool          `yaml:"RedisCheck"`      // 是否检查Redis
	MemoryThreshold int           `yaml:"MemoryThreshold"` // 内存告警阈值(%)
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	ConcurrentWorkers int           `yaml:"ConcurrentWorkers"` // 并发工作协程数
	BatchSize         int           `yaml:"BatchSize"`         // 批处理大小
	MemoryLimit       string        `yaml:"MemoryLimit"`       // 内存限制(如: 100MB)
	ProcessTimeout    time.Duration `yaml:"ProcessTimeout"`    // 处理超时
}

// ExcelSettings Excel适配器设置
type ExcelSettings struct {
	MaxRows       int    `yaml:"MaxRows"`       // 最大行数
	MaxFileSize   string `yaml:"MaxFileSize"`   // 最大文件大小(如: 50MB)
	StreamProcess bool   `yaml:"StreamProcess"` // 是否启用流式处理
	ChunkSize     string `yaml:"ChunkSize"`     // 流处理块大小(如: 1MB)
}

// APISettings API适配器设置
type APISettings struct {
	MaxBatchSize  int           `yaml:"MaxBatchSize"`  // 最大批量处理数
	Timeout       time.Duration `yaml:"Timeout"`       // 请求超时
	MaxConcurrent int           `yaml:"MaxConcurrent"` // 最大并发连接数
}

// DiscoverySettings 自动发现适配器设置
type DiscoverySettings struct {
	Disabled       bool   `yaml:"Disabled"`       // 是否禁用
	MigrationGuide string `yaml:"MigrationGuide"` // 迁移指南
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

// GetConcurrentWorkers 获取并发工作协程数
func (c *AdapterConfig) GetConcurrentWorkers() int {
	if c.Performance == nil || c.Performance.ConcurrentWorkers <= 0 {
		return 8 // 默认值
	}
	return c.Performance.ConcurrentWorkers
}

// GetBatchSize 获取批处理大小
func (c *AdapterConfig) GetBatchSize() int {
	if c.Performance == nil || c.Performance.BatchSize <= 0 {
		return 1000 // 默认值
	}
	return c.Performance.BatchSize
}

// GetMaxRows 获取Excel最大行数
func (c *AdapterConfig) GetMaxRows() int {
	if c.Excel == nil || c.Excel.MaxRows <= 0 {
		return 100000 // 默认值
	}
	return c.Excel.MaxRows
}

// GetMaxBatchSize 获取API最大批量处理数
func (c *AdapterConfig) GetMaxBatchSize() int {
	if c.API == nil || c.API.MaxBatchSize <= 0 {
		return 5000 // 默认值
	}
	return c.API.MaxBatchSize
}

// GetMaxConcurrent 获取最大并发连接数
func (c *AdapterConfig) GetMaxConcurrent() int {
	if c.API == nil || c.API.MaxConcurrent <= 0 {
		return 16 // 默认值
	}
	return c.API.MaxConcurrent
}

// Validate 验证配置的有效性
func (c *AdapterConfig) Validate() error {
	// 设置默认值
	if c.HealthCheck == nil {
		c.HealthCheck = &HealthCheckConfig{
			Enable:          true,
			Interval:        30 * time.Second,
			DatabaseCheck:   true,
			RedisCheck:      true,
			MemoryThreshold: 90,
		}
	}

	if c.Performance == nil {
		c.Performance = &PerformanceConfig{
			ConcurrentWorkers: 8,
			BatchSize:         1000,
			MemoryLimit:       "100MB",
			ProcessTimeout:    30 * time.Second,
		}
	}

	if c.Excel == nil {
		c.Excel = &ExcelSettings{
			MaxRows:       100000,
			MaxFileSize:   "50MB",
			StreamProcess: true,
			ChunkSize:     "1MB",
		}
	}

	if c.API == nil {
		c.API = &APISettings{
			MaxBatchSize:  5000,
			Timeout:       30 * time.Second,
			MaxConcurrent: 16,
		}
	}

	if c.Discovery == nil {
		c.Discovery = &DiscoverySettings{
			Disabled:       true,
			MigrationGuide: "请使用Agent服务进行网络发现",
		}
	}

	// 验证健康检查配置
	if c.HealthCheck.Interval < time.Second {
		c.HealthCheck.Interval = 30 * time.Second
	}

	if c.HealthCheck.MemoryThreshold <= 0 || c.HealthCheck.MemoryThreshold > 100 {
		c.HealthCheck.MemoryThreshold = 90
	}

	// 验证性能配置
	if c.Performance.ConcurrentWorkers <= 0 {
		c.Performance.ConcurrentWorkers = 8
	}

	if c.Performance.BatchSize <= 0 {
		c.Performance.BatchSize = 1000
	}

	if c.Performance.ProcessTimeout < time.Second {
		c.Performance.ProcessTimeout = 30 * time.Second
	}

	// 验证Excel配置
	if c.Excel.MaxRows <= 0 {
		c.Excel.MaxRows = 100000
	}

	// 验证API配置
	if c.API.MaxBatchSize <= 0 {
		c.API.MaxBatchSize = 5000
	}

	if c.API.Timeout < time.Second {
		c.API.Timeout = 30 * time.Second
	}

	if c.API.MaxConcurrent <= 0 {
		c.API.MaxConcurrent = 16
	}

	return nil
}
