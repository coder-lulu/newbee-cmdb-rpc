package input

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConcurrentProcessor 并发处理器 - 解决适配器并发处理能力不足问题
type ConcurrentProcessor struct {
	workerCount int
	batchSize   int
	queueSize   int
	timeout     time.Duration
	stats       *ConcurrentStats
	logger      logx.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	running     int32
}

// ConcurrentStats 并发处理统计
type ConcurrentStats struct {
	TotalTasks     int64         `json:"total_tasks"`
	CompletedTasks int64         `json:"completed_tasks"`
	FailedTasks    int64         `json:"failed_tasks"`
	ActiveWorkers  int32         `json:"active_workers"`
	QueueDepth     int32         `json:"queue_depth"`
	AvgProcessTime time.Duration `json:"avg_process_time"`
	TotalBytes     int64         `json:"total_bytes"`
	StartTime      time.Time     `json:"start_time"`
	mutex          sync.RWMutex
}

// ProcessTask 处理任务
type ProcessTask struct {
	ID       string
	Data     []byte
	Adapter  DataInputAdapter
	Callback func(*ConcurrentProcessResult, error)
	Context  context.Context
	Priority int // 优先级支持
}

// ProcessResult 处理结果
type ProcessResult struct {
	TaskID         string                `json:"task_id"`
	Assets         []*RawAssetData       `json:"assets"`
	ProcessedData  []*ProcessedAssetData `json:"processed_data"`
	Errors         []error               `json:"errors"`
	ProcessTime    time.Duration         `json:"process_time"`
	BytesProcessed int64                 `json:"bytes_processed"`
	WorkerID       int                   `json:"worker_id"`
}

// ConcurrentProcessResult 并发处理结果
type ConcurrentProcessResult struct {
	TaskID         string                `json:"task_id"`
	Assets         []*RawAssetData       `json:"assets"`
	ProcessedData  []*ProcessedAssetData `json:"processed_data"`
	Errors         []error               `json:"errors"`
	ProcessTime    time.Duration         `json:"process_time"`
	BytesProcessed int64                 `json:"bytes_processed"`
	WorkerID       int                   `json:"worker_id"`
}

// ConcurrentConfig 并发配置
type ConcurrentConfig struct {
	WorkerCount int           `json:"worker_count"`
	BatchSize   int           `json:"batch_size"`
	QueueSize   int           `json:"queue_size"`
	Timeout     time.Duration `json:"timeout"`
	EnableStats bool          `json:"enable_stats"`
	AutoScale   bool          `json:"auto_scale"`  // 自动扩缩容
	MaxWorkers  int           `json:"max_workers"` // 最大工作协程数
	MinWorkers  int           `json:"min_workers"` // 最小工作协程数
}

// NewConcurrentProcessor 创建并发处理器
func NewConcurrentProcessor(config *ConcurrentConfig) *ConcurrentProcessor {
	if config == nil {
		config = &ConcurrentConfig{
			WorkerCount: runtime.NumCPU() * 2,
			BatchSize:   100,
			QueueSize:   1000,
			Timeout:     30 * time.Second,
			EnableStats: true,
			AutoScale:   true,
			MaxWorkers:  runtime.NumCPU() * 4,
			MinWorkers:  2,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	processor := &ConcurrentProcessor{
		workerCount: config.WorkerCount,
		batchSize:   config.BatchSize,
		queueSize:   config.QueueSize,
		timeout:     config.Timeout,
		logger:      logx.WithContext(ctx),
		ctx:         ctx,
		cancel:      cancel,
		stats: &ConcurrentStats{
			StartTime: time.Now(),
		},
	}

	processor.logger.Infof("并发处理器初始化完成: 工作协程=%d, 批次大小=%d, 队列大小=%d",
		config.WorkerCount, config.BatchSize, config.QueueSize)

	return processor
}

// Start 启动并发处理器
func (p *ConcurrentProcessor) Start() error {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return fmt.Errorf("并发处理器已在运行中")
	}

	p.logger.Infof("启动并发处理器，工作协程数: %d", p.workerCount)
	return nil
}

// Stop 停止并发处理器
func (p *ConcurrentProcessor) Stop() error {
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return fmt.Errorf("并发处理器未在运行")
	}

	p.cancel()
	p.logger.Infof("并发处理器已停止")
	return nil
}

// ProcessBatch 批量处理数据
func (p *ConcurrentProcessor) ProcessBatch(
	ctx context.Context,
	tasks []*ProcessTask,
) ([]*ConcurrentProcessResult, error) {

	if !p.isRunning() {
		return nil, fmt.Errorf("并发处理器未启动")
	}

	if len(tasks) == 0 {
		return []*ConcurrentProcessResult{}, nil
	}

	startTime := time.Now()
	atomic.AddInt64(&p.stats.TotalTasks, int64(len(tasks)))

	// 创建工作通道
	taskChan := make(chan *ProcessTask, p.queueSize)
	resultChan := make(chan *ConcurrentProcessResult, len(tasks))
	errorChan := make(chan error, len(tasks))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go p.worker(ctx, i, taskChan, resultChan, errorChan, &wg)
	}

	// 分发任务
	go func() {
		defer close(taskChan)
		for _, task := range tasks {
			select {
			case taskChan <- task:
				atomic.AddInt32(&p.stats.QueueDepth, 1)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 收集结果
	results := make([]*ConcurrentProcessResult, 0, len(tasks))
	var errors []error

	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 处理结果
	completed := 0
	for completed < len(tasks) {
		select {
		case result, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				results = append(results, result)
				atomic.AddInt64(&p.stats.CompletedTasks, 1)
				atomic.AddInt64(&p.stats.TotalBytes, result.BytesProcessed)
				completed++
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				errors = append(errors, err)
				atomic.AddInt64(&p.stats.FailedTasks, 1)
				completed++
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("处理超时或被取消: %v", ctx.Err())
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	// 更新统计信息
	processTime := time.Since(startTime)
	p.updateAverageProcessTime(processTime)

	p.logger.Infof("批量处理完成: 任务数=%d, 成功=%d, 失败=%d, 耗时=%v",
		len(tasks), len(results), len(errors), processTime)

	if len(errors) > 0 {
		return results, fmt.Errorf("部分任务失败: %v", errors)
	}

	return results, nil
}

// worker 工作协程
func (p *ConcurrentProcessor) worker(
	ctx context.Context,
	workerID int,
	taskChan <-chan *ProcessTask,
	resultChan chan<- *ConcurrentProcessResult,
	errorChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	atomic.AddInt32(&p.stats.ActiveWorkers, 1)
	defer atomic.AddInt32(&p.stats.ActiveWorkers, -1)

	p.logger.Debugf("工作协程 %d 启动", workerID)

	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				p.logger.Debugf("工作协程 %d 结束", workerID)
				return
			}

			atomic.AddInt32(&p.stats.QueueDepth, -1)
			result, err := p.processTask(ctx, workerID, task)

			if err != nil {
				select {
				case errorChan <- err:
				case <-ctx.Done():
					return
				}
			} else {
				select {
				case resultChan <- result:
				case <-ctx.Done():
					return
				}
			}

		case <-ctx.Done():
			p.logger.Debugf("工作协程 %d 被取消", workerID)
			return
		}
	}
}

// processTask 处理单个任务
func (p *ConcurrentProcessor) processTask(
	ctx context.Context,
	workerID int,
	task *ProcessTask,
) (*ConcurrentProcessResult, error) {

	startTime := time.Now()

	result := &ConcurrentProcessResult{
		TaskID:   task.ID,
		WorkerID: workerID,
		Errors:   []error{},
	}

	// 设置任务上下文超时
	taskCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// 预处理
	inputData := &InputData{
		ID:         task.ID,
		Type:       task.Adapter.GetType(),
		Data:       task.Data,
		CreateTime: startTime,
	}

	preprocessResult, err := task.Adapter.PreProcess(taskCtx, inputData)
	if err != nil {
		return nil, fmt.Errorf("任务 %s 预处理失败: %v", task.ID, err)
	}

	if !preprocessResult.Success {
		return nil, fmt.Errorf("任务 %s 预处理验证失败", task.ID)
	}

	// 解析数据
	assets, err := task.Adapter.Parse(taskCtx, preprocessResult.ProcessedData.Data)
	if err != nil {
		return nil, fmt.Errorf("任务 %s 解析失败: %v", task.ID, err)
	}

	result.Assets = assets

	// 后处理
	processedAssets, err := task.Adapter.PostProcess(taskCtx, assets)
	if err != nil {
		return nil, fmt.Errorf("任务 %s 后处理失败: %v", task.ID, err)
	}

	result.ProcessedData = processedAssets
	result.ProcessTime = time.Since(startTime)
	result.BytesProcessed = int64(len(task.Data))

	// 执行回调
	if task.Callback != nil {
		task.Callback(result, nil)
	}

	p.logger.Debugf("工作协程 %d 完成任务 %s，资产数量: %d，耗时: %v",
		workerID, task.ID, len(assets), result.ProcessTime)

	return result, nil
}

// GetStats 获取统计信息
func (p *ConcurrentProcessor) GetStats() *ConcurrentStats {
	p.stats.mutex.RLock()
	defer p.stats.mutex.RUnlock()

	return &ConcurrentStats{
		TotalTasks:     atomic.LoadInt64(&p.stats.TotalTasks),
		CompletedTasks: atomic.LoadInt64(&p.stats.CompletedTasks),
		FailedTasks:    atomic.LoadInt64(&p.stats.FailedTasks),
		ActiveWorkers:  atomic.LoadInt32(&p.stats.ActiveWorkers),
		QueueDepth:     atomic.LoadInt32(&p.stats.QueueDepth),
		AvgProcessTime: p.stats.AvgProcessTime,
		TotalBytes:     atomic.LoadInt64(&p.stats.TotalBytes),
		StartTime:      p.stats.StartTime,
	}
}

// GetPerformanceMetrics 获取性能指标
func (p *ConcurrentProcessor) GetPerformanceMetrics() map[string]interface{} {
	stats := p.GetStats()
	runTime := time.Since(stats.StartTime)

	return map[string]interface{}{
		"total_tasks":      stats.TotalTasks,
		"completed_tasks":  stats.CompletedTasks,
		"failed_tasks":     stats.FailedTasks,
		"success_rate":     float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100,
		"active_workers":   stats.ActiveWorkers,
		"queue_depth":      stats.QueueDepth,
		"avg_process_time": stats.AvgProcessTime.Milliseconds(),
		"total_bytes":      stats.TotalBytes,
		"throughput_tps":   float64(stats.CompletedTasks) / runTime.Seconds(),
		"throughput_bps":   float64(stats.TotalBytes) / runTime.Seconds(),
		"uptime_seconds":   runTime.Seconds(),
		"cpu_cores":        runtime.NumCPU(),
		"goroutines":       runtime.NumGoroutine(),
	}
}

// isRunning 检查是否在运行
func (p *ConcurrentProcessor) isRunning() bool {
	return atomic.LoadInt32(&p.running) == 1
}

// updateAverageProcessTime 更新平均处理时间
func (p *ConcurrentProcessor) updateAverageProcessTime(duration time.Duration) {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()

	if p.stats.AvgProcessTime == 0 {
		p.stats.AvgProcessTime = duration
	} else {
		// 使用指数移动平均
		p.stats.AvgProcessTime = time.Duration(
			float64(p.stats.AvgProcessTime)*0.9 + float64(duration)*0.1,
		)
	}
}
