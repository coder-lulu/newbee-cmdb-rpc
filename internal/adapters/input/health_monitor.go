package input

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// HealthMonitor 健康监控器
type HealthMonitor struct {
	adapters      map[string]DataInputAdapter
	svcCtx        *svc.ServiceContext
	checkInterval time.Duration
	logger        logx.Logger
	metrics       *HealthMetrics
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
}

// HealthMetrics 健康指标
type HealthMetrics struct {
	SystemHealth     *SystemHealth                `json:"system_health"`
	AdapterHealth    map[string]*AdapterHealth    `json:"adapter_health"`
	DependencyHealth map[string]*DependencyHealth `json:"dependency_health"`
	LastCheckTime    time.Time                    `json:"last_check_time"`
	CheckDuration    time.Duration                `json:"check_duration"`
	OverallHealthy   bool                         `json:"overall_healthy"`
	AlertCount       int                          `json:"alert_count"`
	mutex            sync.RWMutex
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	MemoryTotal    uint64    `json:"memory_total"`
	MemoryUsed     uint64    `json:"memory_used"`
	GoroutineCount int       `json:"goroutine_count"`
	GCStats        *GCStats  `json:"gc_stats"`
	LastUpdated    time.Time `json:"last_updated"`
	Healthy        bool      `json:"healthy"`
}

// GCStats 垃圾回收统计
type GCStats struct {
	NumGC        uint32        `json:"num_gc"`
	PauseTotal   time.Duration `json:"pause_total"`
	LastGC       time.Time     `json:"last_gc"`
	NextGC       uint64        `json:"next_gc"`
	MemoryTarget uint64        `json:"memory_target"`
}

// AdapterHealth 适配器健康状态
type AdapterHealth struct {
	Type            string                 `json:"type"`
	Version         string                 `json:"version"`
	Healthy         bool                   `json:"healthy"`
	LastError       string                 `json:"last_error"`
	ErrorCount      int64                  `json:"error_count"`
	SuccessCount    int64                  `json:"success_count"`
	SuccessRate     float64                `json:"success_rate"`
	AvgResponseTime time.Duration          `json:"avg_response_time"`
	LastCheckTime   time.Time              `json:"last_check_time"`
	ConfigValid     bool                   `json:"config_valid"`
	Stats           map[string]interface{} `json:"stats"`
}

// DependencyHealth 依赖服务健康状态
type DependencyHealth struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"` // database, redis, external_api
	Healthy       bool          `json:"healthy"`
	LastError     string        `json:"last_error"`
	ResponseTime  time.Duration `json:"response_time"`
	LastCheckTime time.Time     `json:"last_check_time"`
	RetryCount    int           `json:"retry_count"`
	Timeout       time.Duration `json:"timeout"`
}

// HealthAlert 健康告警
type HealthAlert struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // error, warning, info
	Component  string                 `json:"component"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(svcCtx *svc.ServiceContext, checkInterval time.Duration) *HealthMonitor {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second // 默认30秒检查一次
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &HealthMonitor{
		adapters:      make(map[string]DataInputAdapter),
		svcCtx:        svcCtx,
		checkInterval: checkInterval,
		logger:        logx.WithContext(ctx),
		ctx:           ctx,
		cancel:        cancel,
		metrics: &HealthMetrics{
			SystemHealth:     &SystemHealth{},
			AdapterHealth:    make(map[string]*AdapterHealth),
			DependencyHealth: make(map[string]*DependencyHealth),
		},
	}

	return monitor
}

// RegisterAdapter 注册适配器
func (h *HealthMonitor) RegisterAdapter(id string, adapter DataInputAdapter) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.adapters[id] = adapter
	h.metrics.AdapterHealth[id] = &AdapterHealth{
		Type:        adapter.GetType(),
		Version:     adapter.GetVersion(),
		ConfigValid: true,
		Stats:       make(map[string]interface{}),
	}

	h.logger.Infof("注册适配器到健康监控: %s (%s)", id, adapter.GetType())
}

// UnregisterAdapter 注销适配器
func (h *HealthMonitor) UnregisterAdapter(id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.adapters, id)
	delete(h.metrics.AdapterHealth, id)

	h.logger.Infof("从健康监控注销适配器: %s", id)
}

// Start 启动健康监控
func (h *HealthMonitor) Start() {
	h.mutex.Lock()
	if h.running {
		h.mutex.Unlock()
		return
	}
	h.running = true
	h.mutex.Unlock()

	h.logger.Info("启动健康监控器")

	// 立即执行一次检查
	h.performHealthCheck()

	// 启动定期检查
	go h.runPeriodicCheck()
}

// Stop 停止健康监控
func (h *HealthMonitor) Stop() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.running {
		return
	}

	h.running = false
	h.cancel()
	h.logger.Info("健康监控器已停止")
}

// runPeriodicCheck 运行定期检查
func (h *HealthMonitor) runPeriodicCheck() {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (h *HealthMonitor) performHealthCheck() {
	startTime := time.Now()
	h.logger.Debug("开始健康检查")

	var wg sync.WaitGroup

	// 并行检查各个组件
	wg.Add(3)

	// 检查系统健康
	go func() {
		defer wg.Done()
		h.checkSystemHealth()
	}()

	// 检查适配器健康
	go func() {
		defer wg.Done()
		h.checkAdaptersHealth()
	}()

	// 检查依赖健康
	go func() {
		defer wg.Done()
		h.checkDependenciesHealth()
	}()

	wg.Wait()

	// 更新总体状态
	h.updateOverallHealth()

	checkDuration := time.Since(startTime)
	h.metrics.mutex.Lock()
	h.metrics.LastCheckTime = startTime
	h.metrics.CheckDuration = checkDuration
	h.metrics.mutex.Unlock()

	h.logger.Debugf("健康检查完成，耗时: %v", checkDuration)
}

// checkSystemHealth 检查系统健康
func (h *HealthMonitor) checkSystemHealth() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	gcStats := &GCStats{
		NumGC:        m.NumGC,
		PauseTotal:   time.Duration(m.PauseTotalNs),
		NextGC:       m.NextGC,
		MemoryTarget: m.HeapSys,
	}

	if m.LastGC > 0 {
		gcStats.LastGC = time.Unix(0, int64(m.LastGC))
	}

	memoryUsed := m.Alloc
	memoryTotal := m.Sys
	memoryUsage := float64(memoryUsed) / float64(memoryTotal) * 100

	goroutineCount := runtime.NumGoroutine()

	systemHealth := &SystemHealth{
		MemoryUsage:    memoryUsage,
		MemoryTotal:    memoryTotal,
		MemoryUsed:     memoryUsed,
		GoroutineCount: goroutineCount,
		GCStats:        gcStats,
		LastUpdated:    time.Now(),
		Healthy:        true,
	}

	// 健康检查阈值
	if memoryUsage > 90 {
		systemHealth.Healthy = false
		h.logger.Errorf("系统内存使用率过高: %.2f%%", memoryUsage)
	}

	if goroutineCount > 10000 {
		systemHealth.Healthy = false
		h.logger.Errorf("协程数量过多: %d", goroutineCount)
	}

	h.metrics.mutex.Lock()
	h.metrics.SystemHealth = systemHealth
	h.metrics.mutex.Unlock()
}

// checkAdaptersHealth 检查适配器健康
func (h *HealthMonitor) checkAdaptersHealth() {
	h.mutex.RLock()
	adapters := make(map[string]DataInputAdapter)
	for id, adapter := range h.adapters {
		adapters[id] = adapter
	}
	h.mutex.RUnlock()

	for id, adapter := range adapters {
		h.checkSingleAdapterHealth(id, adapter)
	}
}

// checkSingleAdapterHealth 检查单个适配器健康
func (h *HealthMonitor) checkSingleAdapterHealth(id string, adapter DataInputAdapter) {
	startTime := time.Now()

	health := &AdapterHealth{
		Type:          adapter.GetType(),
		Version:       adapter.GetVersion(),
		LastCheckTime: startTime,
		ConfigValid:   true,
		Stats:         make(map[string]interface{}),
	}

	// 执行健康检查
	err := adapter.HealthCheck()
	responseTime := time.Since(startTime)

	if err != nil {
		health.Healthy = false
		health.LastError = err.Error()
		health.ErrorCount++
		h.logger.Errorf("适配器 %s 健康检查失败: %v", id, err)
	} else {
		health.Healthy = true
		health.SuccessCount++
	}

	health.AvgResponseTime = responseTime

	// 计算成功率
	totalChecks := health.SuccessCount + health.ErrorCount
	if totalChecks > 0 {
		health.SuccessRate = float64(health.SuccessCount) / float64(totalChecks) * 100
	}

	// 获取适配器统计信息
	if baseAdapter, ok := adapter.(*ExcelInputAdapter); ok {
		health.Stats = baseAdapter.GetStats()
	} else if baseAdapter, ok := adapter.(*APIInputAdapter); ok {
		health.Stats = baseAdapter.GetStats()
	}

	h.metrics.mutex.Lock()
	h.metrics.AdapterHealth[id] = health
	h.metrics.mutex.Unlock()
}

// checkDependenciesHealth 检查依赖健康
func (h *HealthMonitor) checkDependenciesHealth() {
	// 检查数据库连接
	if h.svcCtx != nil && h.svcCtx.DB != nil {
		h.checkDatabaseHealth()
	}

	// 检查Redis连接
	// 注意：这里需要根据实际的ServiceContext结构调整
	// if h.svcCtx != nil && h.svcCtx.RedisClient != nil {
	//     h.checkRedisHealth()
	// }
}

// checkDatabaseHealth 检查数据库健康
func (h *HealthMonitor) checkDatabaseHealth() {
	startTime := time.Now()

	health := &DependencyHealth{
		Name:          "database",
		Type:          "database",
		LastCheckTime: startTime,
		Timeout:       5 * time.Second,
	}

	// 实际的数据库ping检查
	// 注意：这里需要根据实际的数据库客户端调整
	// err := h.svcCtx.DB.Ping()
	// 临时使用模拟检查
	var err error = nil // 在生产环境中应该实现真实的ping检查

	responseTime := time.Since(startTime)
	health.ResponseTime = responseTime

	if err != nil {
		health.Healthy = false
		health.LastError = err.Error()
		health.RetryCount++
		h.logger.Errorf("数据库健康检查失败: %v", err)
	} else {
		health.Healthy = true
		health.LastError = ""
	}

	h.metrics.mutex.Lock()
	h.metrics.DependencyHealth["database"] = health
	h.metrics.mutex.Unlock()
}

// updateOverallHealth 更新总体健康状态
func (h *HealthMonitor) updateOverallHealth() {
	h.metrics.mutex.Lock()
	defer h.metrics.mutex.Unlock()

	overallHealthy := true
	alertCount := 0

	// 检查系统健康
	if !h.metrics.SystemHealth.Healthy {
		overallHealthy = false
		alertCount++
	}

	// 检查适配器健康
	for _, health := range h.metrics.AdapterHealth {
		if !health.Healthy {
			overallHealthy = false
			alertCount++
		}
	}

	// 检查依赖健康
	for _, health := range h.metrics.DependencyHealth {
		if !health.Healthy {
			overallHealthy = false
			alertCount++
		}
	}

	h.metrics.OverallHealthy = overallHealthy
	h.metrics.AlertCount = alertCount
}

// GetHealth 获取健康状态
func (h *HealthMonitor) GetHealth() *HealthMetrics {
	h.metrics.mutex.RLock()
	defer h.metrics.mutex.RUnlock()

	// 深度复制以避免并发问题
	health := &HealthMetrics{
		SystemHealth:     h.copySystemHealth(),
		AdapterHealth:    h.copyAdapterHealth(),
		DependencyHealth: h.copyDependencyHealth(),
		LastCheckTime:    h.metrics.LastCheckTime,
		CheckDuration:    h.metrics.CheckDuration,
		OverallHealthy:   h.metrics.OverallHealthy,
		AlertCount:       h.metrics.AlertCount,
	}

	return health
}

// copySystemHealth 复制系统健康状态
func (h *HealthMonitor) copySystemHealth() *SystemHealth {
	if h.metrics.SystemHealth == nil {
		return nil
	}

	return &SystemHealth{
		CPUUsage:       h.metrics.SystemHealth.CPUUsage,
		MemoryUsage:    h.metrics.SystemHealth.MemoryUsage,
		MemoryTotal:    h.metrics.SystemHealth.MemoryTotal,
		MemoryUsed:     h.metrics.SystemHealth.MemoryUsed,
		GoroutineCount: h.metrics.SystemHealth.GoroutineCount,
		GCStats:        h.metrics.SystemHealth.GCStats,
		LastUpdated:    h.metrics.SystemHealth.LastUpdated,
		Healthy:        h.metrics.SystemHealth.Healthy,
	}
}

// copyAdapterHealth 复制适配器健康状态
func (h *HealthMonitor) copyAdapterHealth() map[string]*AdapterHealth {
	copied := make(map[string]*AdapterHealth)
	for id, health := range h.metrics.AdapterHealth {
		copied[id] = &AdapterHealth{
			Type:            health.Type,
			Version:         health.Version,
			Healthy:         health.Healthy,
			LastError:       health.LastError,
			ErrorCount:      health.ErrorCount,
			SuccessCount:    health.SuccessCount,
			SuccessRate:     health.SuccessRate,
			AvgResponseTime: health.AvgResponseTime,
			LastCheckTime:   health.LastCheckTime,
			ConfigValid:     health.ConfigValid,
			Stats:           health.Stats,
		}
	}
	return copied
}

// copyDependencyHealth 复制依赖健康状态
func (h *HealthMonitor) copyDependencyHealth() map[string]*DependencyHealth {
	copied := make(map[string]*DependencyHealth)
	for name, health := range h.metrics.DependencyHealth {
		copied[name] = &DependencyHealth{
			Name:          health.Name,
			Type:          health.Type,
			Healthy:       health.Healthy,
			LastError:     health.LastError,
			ResponseTime:  health.ResponseTime,
			LastCheckTime: health.LastCheckTime,
			RetryCount:    health.RetryCount,
			Timeout:       health.Timeout,
		}
	}
	return copied
}

// GetHealthSummary 获取健康状态摘要
func (h *HealthMonitor) GetHealthSummary() map[string]interface{} {
	health := h.GetHealth()

	summary := map[string]interface{}{
		"overall_healthy":   health.OverallHealthy,
		"last_check_time":   health.LastCheckTime,
		"check_duration_ms": health.CheckDuration.Milliseconds(),
		"alert_count":       health.AlertCount,
		"components": map[string]interface{}{
			"system": map[string]interface{}{
				"healthy":         health.SystemHealth.Healthy,
				"memory_usage":    fmt.Sprintf("%.2f%%", health.SystemHealth.MemoryUsage),
				"goroutine_count": health.SystemHealth.GoroutineCount,
			},
			"adapters": map[string]interface{}{
				"total":   len(health.AdapterHealth),
				"healthy": h.countHealthyAdapters(health.AdapterHealth),
			},
			"dependencies": map[string]interface{}{
				"total":   len(health.DependencyHealth),
				"healthy": h.countHealthyDependencies(health.DependencyHealth),
			},
		},
	}

	return summary
}

// countHealthyAdapters 计算健康的适配器数量
func (h *HealthMonitor) countHealthyAdapters(adapters map[string]*AdapterHealth) int {
	count := 0
	for _, health := range adapters {
		if health.Healthy {
			count++
		}
	}
	return count
}

// countHealthyDependencies 计算健康的依赖数量
func (h *HealthMonitor) countHealthyDependencies(dependencies map[string]*DependencyHealth) int {
	count := 0
	for _, health := range dependencies {
		if health.Healthy {
			count++
		}
	}
	return count
}

// IsHealthy 检查是否健康
func (h *HealthMonitor) IsHealthy() bool {
	health := h.GetHealth()
	return health.OverallHealthy
}
