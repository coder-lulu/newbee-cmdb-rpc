package input

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdapterFactory 适配器工厂函数
type AdapterFactory func(*svc.ServiceContext, map[string]interface{}) (DataInputAdapter, error)

// AdapterRegistry 适配器注册器
type AdapterRegistry struct {
	adapters map[string]AdapterFactory
	configs  map[string]*ConfigSchema
	mutex    sync.RWMutex
	logger   logx.Logger
}

// NewAdapterRegistry 创建新的适配器注册器
func NewAdapterRegistry() *AdapterRegistry {
	registry := &AdapterRegistry{
		adapters: make(map[string]AdapterFactory),
		configs:  make(map[string]*ConfigSchema),
		logger:   logx.WithContext(context.Background()),
	}

	// 注册默认适配器
	registry.registerDefaultAdapters()

	return registry
}

// registerDefaultAdapters 注册默认适配器
func (r *AdapterRegistry) registerDefaultAdapters() {
	// 注册Excel适配器
	r.RegisterAdapter("excel", func(svcCtx *svc.ServiceContext, config map[string]interface{}) (DataInputAdapter, error) {
		excelConfig := r.parseExcelConfig(config)
		return NewExcelInputAdapter(svcCtx, excelConfig), nil
	})

	// 注册API适配器
	r.RegisterAdapter("api", func(svcCtx *svc.ServiceContext, config map[string]interface{}) (DataInputAdapter, error) {
		apiConfig := r.parseAPIConfig(config)
		return NewAPIInputAdapter(svcCtx, apiConfig), nil
	})

	// 注册自动发现适配器
	r.RegisterAdapter("discovery", func(svcCtx *svc.ServiceContext, config map[string]interface{}) (DataInputAdapter, error) {
		discoveryConfig := r.parseDiscoveryConfig(config)
		return NewDiscoveryInputAdapter(svcCtx, discoveryConfig), nil
	})
}

// RegisterAdapter 注册适配器
func (r *AdapterRegistry) RegisterAdapter(adapterType string, factory AdapterFactory) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.adapters[adapterType] = factory
	r.logger.Infof("注册适配器: %s", adapterType)
}

// CreateAdapter 创建适配器实例
func (r *AdapterRegistry) CreateAdapter(adapterType string, svcCtx *svc.ServiceContext, config map[string]interface{}) (DataInputAdapter, error) {
	r.mutex.RLock()
	factory, exists := r.adapters[adapterType]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未找到适配器类型: %s", adapterType)
	}

	adapter, err := factory(svcCtx, config)
	if err != nil {
		return nil, fmt.Errorf("创建适配器失败 [%s]: %v", adapterType, err)
	}

	// 健康检查
	if err := adapter.HealthCheck(); err != nil {
		return nil, fmt.Errorf("适配器健康检查失败 [%s]: %v", adapterType, err)
	}

	r.logger.Infof("成功创建适配器: %s", adapterType)
	return adapter, nil
}

// GetAvailableAdapters 获取可用的适配器类型列表
func (r *AdapterRegistry) GetAvailableAdapters() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	adapters := make([]string, 0, len(r.adapters))
	for adapterType := range r.adapters {
		adapters = append(adapters, adapterType)
	}

	return adapters
}

// GetAdapterConfig 获取适配器配置模式
func (r *AdapterRegistry) GetAdapterConfig(adapterType string) (*ConfigSchema, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, exists := r.adapters[adapterType]; !exists {
		return nil, fmt.Errorf("未找到适配器类型: %s", adapterType)
	}

	// 如果有缓存的配置，直接返回
	if config, exists := r.configs[adapterType]; exists {
		return config, nil
	}

	// 创建临时适配器实例获取配置
	factory := r.adapters[adapterType]
	tempAdapter, err := factory(nil, make(map[string]interface{}))
	if err != nil {
		return nil, fmt.Errorf("获取适配器配置失败: %v", err)
	}

	config := tempAdapter.GetConfigSchema()
	r.configs[adapterType] = config

	return config, nil
}

// ValidateConfig 验证适配器配置
func (r *AdapterRegistry) ValidateConfig(adapterType string, config map[string]interface{}) error {
	schema, err := r.GetAdapterConfig(adapterType)
	if err != nil {
		return err
	}

	// 检查必填字段
	for _, required := range schema.Required {
		if _, exists := config[required]; !exists {
			return fmt.Errorf("缺少必填配置项: %s", required)
		}
	}

	// 根据适配器类型进行特定验证
	switch adapterType {
	case "excel":
		return r.validateExcelConfig(config)
	case "api":
		return r.validateAPIConfig(config)
	case "discovery":
		return r.validateDiscoveryConfig(config)
	default:
		return nil
	}
}

// validateExcelConfig 验证Excel配置
func (r *AdapterRegistry) validateExcelConfig(config map[string]interface{}) error {
	if maxFileSize, ok := config["max_file_size"].(int64); ok {
		if maxFileSize <= 0 || maxFileSize > 100*1024*1024 { // 限制100MB
			return fmt.Errorf("max_file_size必须在1到100MB之间")
		}
	}

	if maxRows, ok := config["max_rows"].(int); ok {
		if maxRows <= 0 || maxRows > 100000 { // 限制10万行
			return fmt.Errorf("max_rows必须在1到100,000之间")
		}
	}

	if headerRow, ok := config["header_row"].(int); ok {
		if headerRow < 1 {
			return fmt.Errorf("header_row必须大于0")
		}
	}

	if dataStartRow, ok := config["data_start_row"].(int); ok {
		if dataStartRow < 1 {
			return fmt.Errorf("data_start_row必须大于0")
		}

		if headerRow, ok := config["header_row"].(int); ok {
			if dataStartRow <= headerRow {
				return fmt.Errorf("data_start_row必须大于header_row")
			}
		}
	}

	return nil
}

// validateAPIConfig 验证API配置
func (r *AdapterRegistry) validateAPIConfig(config map[string]interface{}) error {
	if maxBatchSize, ok := config["max_batch_size"].(int); ok {
		if maxBatchSize <= 0 || maxBatchSize > 10000 {
			return fmt.Errorf("max_batch_size必须在1到10,000之间")
		}
	}

	if rateLimit, ok := config["rate_limit"].(int); ok {
		if rateLimit <= 0 || rateLimit > 1000 {
			return fmt.Errorf("rate_limit必须在1到1,000之间")
		}
	}

	if retryCount, ok := config["retry_count"].(int); ok {
		if retryCount < 0 || retryCount > 10 {
			return fmt.Errorf("retry_count必须在0到10之间")
		}
	}

	return nil
}

// validateDiscoveryConfig 验证自动发现配置
func (r *AdapterRegistry) validateDiscoveryConfig(config map[string]interface{}) error {
	if maxTargets, ok := config["max_targets"].(int); ok {
		if maxTargets <= 0 || maxTargets > 10000 {
			return fmt.Errorf("max_targets必须在1到10,000之间")
		}
	}

	return nil
}

// AdapterManager 适配器管理器
type AdapterManager struct {
	registry *AdapterRegistry
	svcCtx   *svc.ServiceContext
	cache    map[string]DataInputAdapter
	mutex    sync.RWMutex
	logger   logx.Logger
}

// NewAdapterManager 创建适配器管理器
func NewAdapterManager(svcCtx *svc.ServiceContext) *AdapterManager {
	return &AdapterManager{
		registry: NewAdapterRegistry(),
		svcCtx:   svcCtx,
		cache:    make(map[string]DataInputAdapter),
		logger:   logx.WithContext(context.Background()),
	}
}

// GetAdapter 获取适配器（支持缓存）
func (m *AdapterManager) GetAdapter(adapterType string, config map[string]interface{}) (DataInputAdapter, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("%s_%x", adapterType, m.hashConfig(config))

	// 检查缓存
	m.mutex.RLock()
	if adapter, exists := m.cache[cacheKey]; exists {
		m.mutex.RUnlock()
		return adapter, nil
	}
	m.mutex.RUnlock()

	// 创建新适配器
	adapter, err := m.registry.CreateAdapter(adapterType, m.svcCtx, config)
	if err != nil {
		return nil, err
	}

	// 缓存适配器
	m.mutex.Lock()
	m.cache[cacheKey] = adapter
	m.mutex.Unlock()

	return adapter, nil
}

// ProcessInput 处理输入数据的完整流程
func (m *AdapterManager) ProcessInput(ctx context.Context, input *InputData) (*AdapterProcessResult, error) {
	// 1. 获取适配器
	config := make(map[string]interface{})
	if input.Config != nil {
		config = input.Config
	}

	adapter, err := m.GetAdapter(input.Type, config)
	if err != nil {
		return nil, fmt.Errorf("获取适配器失败: %v", err)
	}

	startTime := time.Now()

	// 2. 预处理
	m.logger.Infof("开始预处理: %s", input.Type)
	preprocessResult, err := adapter.PreProcess(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("预处理失败: %v", err)
	}

	if !preprocessResult.Success {
		return &AdapterProcessResult{
			ProcessedData: &ProcessingData{
				Assets: []*ProcessedAssetData{},
			},
			TotalErrors: 1,
			ProcessTime: time.Since(startTime),
		}, fmt.Errorf("预处理失败: %v", preprocessResult.Warnings)
	}

	// 3. 解析数据
	m.logger.Infof("开始解析数据: %s", input.Type)
	rawAssets, err := adapter.Parse(ctx, preprocessResult.ProcessedData.Data)
	if err != nil {
		return nil, fmt.Errorf("数据解析失败: %v", err)
	}

	// 4. 后处理
	m.logger.Infof("开始后处理: %d条记录", len(rawAssets))
	processedAssets, err := adapter.PostProcess(ctx, rawAssets)
	if err != nil {
		return nil, fmt.Errorf("后处理失败: %v", err)
	}

	// 5. 数据验证
	m.logger.Infof("开始数据验证: %d条记录", len(processedAssets))
	parsedData := &ParsedData{
		Records:  rawAssets,
		Metadata: preprocessResult.Metadata,
	}

	validationResult, err := adapter.Validate(ctx, parsedData, m.svcCtx)
	if err != nil {
		return nil, fmt.Errorf("数据验证失败: %v", err)
	}

	// 6. 构建处理结果
	result := &AdapterProcessResult{
		ProcessedData: &ProcessingData{
			Assets: processedAssets,
			BatchInfo: &BatchInfo{
				ID:         input.BatchID,
				Type:       input.Type,
				Source:     input.Source,
				TotalCount: len(processedAssets),
				CreateTime: input.CreateTime,
				CreatedBy:  input.CreatedBy,
				Status:     "processed",
			},
			Context:    preprocessResult.Metadata,
			ErrorCount: validationResult.ErrorCount,
		},
		ProcessedCount: len(processedAssets),
		TotalErrors:    validationResult.ErrorCount,
		ProcessTime:    time.Since(startTime),
	}

	m.logger.Infof("数据处理完成: 处理%d条记录，%d条错误，耗时%v",
		result.ProcessedCount, result.TotalErrors, result.ProcessTime)

	return result, nil
}

// ClearCache 清理缓存
func (m *AdapterManager) ClearCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.cache = make(map[string]DataInputAdapter)
	m.logger.Info("适配器缓存已清理")
}

// GetRegistry 获取注册器
func (m *AdapterManager) GetRegistry() *AdapterRegistry {
	return m.registry
}

// GetAllAdapterStats 获取所有缓存适配器的统计信息
func (m *AdapterManager) GetAllAdapterStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	for cacheKey, adapter := range m.cache {
		if baseAdapter, ok := adapter.(*ExcelInputAdapter); ok {
			stats[cacheKey] = baseAdapter.GetStats()
		} else if baseAdapter, ok := adapter.(*APIInputAdapter); ok {
			stats[cacheKey] = baseAdapter.GetStats()
		} else if baseAdapter, ok := adapter.(*DiscoveryInputAdapter); ok {
			stats[cacheKey] = baseAdapter.GetStats()
		}
	}

	// 添加管理器级别的统计信息
	stats["manager_info"] = map[string]interface{}{
		"cached_adapters": len(m.cache),
		"available_types": m.registry.GetAvailableAdapters(),
		"last_updated":    time.Now().Format(time.RFC3339),
	}

	return stats
}

// GetAdapterByType 获取指定类型的适配器统计信息
func (m *AdapterManager) GetAdapterByType(adapterType string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, adapter := range m.cache {
		if adapter.GetType() == adapterType {
			if baseAdapter, ok := adapter.(*ExcelInputAdapter); ok {
				return baseAdapter.GetStats(), nil
			} else if baseAdapter, ok := adapter.(*APIInputAdapter); ok {
				return baseAdapter.GetStats(), nil
			} else if baseAdapter, ok := adapter.(*DiscoveryInputAdapter); ok {
				return baseAdapter.GetStats(), nil
			}
		}
	}

	return nil, fmt.Errorf("未找到类型为%s的适配器", adapterType)
}

// GetSystemHealth 获取适配器系统整体健康状态
func (m *AdapterManager) GetSystemHealth() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	health := map[string]interface{}{
		"total_adapters":     len(m.cache),
		"healthy_adapters":   0,
		"unhealthy_adapters": 0,
		"adapter_details":    make(map[string]interface{}),
		"check_time":         time.Now().Format(time.RFC3339),
	}

	healthyCount := 0
	unhealthyCount := 0
	adapterDetails := health["adapter_details"].(map[string]interface{})

	for cacheKey, adapter := range m.cache {
		err := adapter.HealthCheck()
		isHealthy := err == nil

		if isHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}

		adapterDetails[cacheKey] = map[string]interface{}{
			"type":    adapter.GetType(),
			"version": adapter.GetVersion(),
			"healthy": isHealthy,
			"error": func() string {
				if err != nil {
					return err.Error()
				} else {
					return ""
				}
			}(),
		}
	}

	health["healthy_adapters"] = healthyCount
	health["unhealthy_adapters"] = unhealthyCount
	health["overall_healthy"] = unhealthyCount == 0

	return health
}

// hashConfig 计算配置哈希值
func (m *AdapterManager) hashConfig(config map[string]interface{}) uint32 {
	// 使用更可靠的哈希算法

	// 将配置序列化为JSON以获得一致的哈希
	configJSON, err := json.Marshal(config)
	if err != nil {
		// 如果序列化失败，使用简单哈希作为备用
		var hash uint32 = 0
		for key, value := range config {
			hash += uint32(len(key))
			if str, ok := value.(string); ok {
				hash += uint32(len(str))
			}
		}
		return hash
	}

	// 计算MD5哈希
	hasher := md5.New()
	hasher.Write(configJSON)
	hashBytes := hasher.Sum(nil)

	// 将前4个字节转换为uint32
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
}

// parseExcelConfig 解析Excel配置
func (r *AdapterRegistry) parseExcelConfig(config map[string]interface{}) *ExcelConfig {
	excelConfig := &ExcelConfig{
		MaxFileSize:   50 * 1024 * 1024, // 默认50MB
		MaxRows:       50000,            // 默认50,000行
		SkipEmptyRows: true,
		HeaderRow:     1,
		DataStartRow:  2,
		SheetMapping:  make(map[string]string),
		ColumnMapping: make(map[string]string),
	}

	// 从配置中读取参数
	if maxFileSize, ok := config["max_file_size"].(int64); ok {
		excelConfig.MaxFileSize = maxFileSize
	}
	if maxRows, ok := config["max_rows"].(int); ok {
		excelConfig.MaxRows = maxRows
	}
	if skipEmptyRows, ok := config["skip_empty_rows"].(bool); ok {
		excelConfig.SkipEmptyRows = skipEmptyRows
	}
	if headerRow, ok := config["header_row"].(int); ok {
		excelConfig.HeaderRow = headerRow
	}
	if dataStartRow, ok := config["data_start_row"].(int); ok {
		excelConfig.DataStartRow = dataStartRow
	}

	return excelConfig
}

// parseAPIConfig 解析API配置
func (r *AdapterRegistry) parseAPIConfig(config map[string]interface{}) *APIConfig {
	apiConfig := &APIConfig{
		MaxBatchSize:   1000,             // 默认1000条
		RateLimit:      100,              // 默认100 req/min
		Timeout:        30 * time.Second, // 默认30秒
		RetryCount:     3,                // 默认重试3次
		RequiredFields: []string{},
		FieldMapping:   make(map[string]string),
	}

	// 从配置中读取参数
	if maxBatchSize, ok := config["max_batch_size"].(int); ok {
		apiConfig.MaxBatchSize = maxBatchSize
	}
	if rateLimit, ok := config["rate_limit"].(int); ok {
		apiConfig.RateLimit = rateLimit
	}
	if timeout, ok := config["timeout"].(time.Duration); ok {
		apiConfig.Timeout = timeout
	}
	if retryCount, ok := config["retry_count"].(int); ok {
		apiConfig.RetryCount = retryCount
	}

	return apiConfig
}

// parseDiscoveryConfig 解析自动发现配置
func (r *AdapterRegistry) parseDiscoveryConfig(config map[string]interface{}) *DiscoveryConfig {
	discoveryConfig := &DiscoveryConfig{
		DiscoveryRules: []*DiscoveryRule{},
		ScanInterval:   5 * time.Minute,  // 默认5分钟
		MaxTargets:     1000,             // 默认1000个目标
		Timeout:        30 * time.Second, // 默认30秒
		AgentConfig:    make(map[string]interface{}),
	}

	// 从配置中读取参数
	if scanInterval, ok := config["scan_interval"].(time.Duration); ok {
		discoveryConfig.ScanInterval = scanInterval
	}
	if maxTargets, ok := config["max_targets"].(int); ok {
		discoveryConfig.MaxTargets = maxTargets
	}
	if timeout, ok := config["timeout"].(time.Duration); ok {
		discoveryConfig.Timeout = timeout
	}

	return discoveryConfig
}
