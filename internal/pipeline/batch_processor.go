package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
)

// BatchProcessor 企业级批量处理器
type BatchProcessor struct {
	svcCtx          *svc.ServiceContext
	logger          logx.Logger
	config          *BatchConfig
	taskManager     *TaskManager
	dataPipeline    *DataPipeline
	concurrencyPool *ConcurrencyPool
}

// BatchConfig 批量处理配置
type BatchConfig struct {
	MaxBatchSize    int           `json:"max_batch_size"`    // 最大批量大小
	ChunkSize       int           `json:"chunk_size"`        // 分块大小
	MaxConcurrency  int           `json:"max_concurrency"`   // 最大并发数
	TimeoutPerChunk time.Duration `json:"timeout_per_chunk"` // 每块超时时间
	EnableProgress  bool          `json:"enable_progress"`   // 是否启用进度跟踪
	EnableRollback  bool          `json:"enable_rollback"`   // 是否启用回滚
	RetryAttempts   int           `json:"retry_attempts"`    // 重试次数
	RetryDelay      time.Duration `json:"retry_delay"`       // 重试延迟
	EnableNotify    bool          `json:"enable_notify"`     // 是否启用通知
}

// BatchTask 批量处理任务
type BatchTask struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`   // import/update/delete
	Status       string                `json:"status"` // pending/processing/completed/failed
	Priority     input.Priority        `json:"priority"`
	Data         *input.ProcessingData `json:"data"`
	Config       *BatchConfig          `json:"config"`
	Progress     *BatchProgress        `json:"progress"`
	Result       *BatchResult          `json:"result"`
	CreateTime   time.Time             `json:"create_time"`
	UpdateTime   time.Time             `json:"update_time"`
	StartTime    *time.Time            `json:"start_time"`
	EndTime      *time.Time            `json:"end_time"`
	CreatedBy    string                `json:"created_by"`
	ChunkResults []*ChunkResult        `json:"chunk_results"`
	ErrorDetails []*BatchError         `json:"error_details"`
}

// BatchProgress 批量处理进度
type BatchProgress struct {
	TotalItems         int           `json:"total_items"`
	ProcessedItems     int           `json:"processed_items"`
	SuccessItems       int           `json:"success_items"`
	FailedItems        int           `json:"failed_items"`
	SkippedItems       int           `json:"skipped_items"`
	CurrentChunk       int           `json:"current_chunk"`
	TotalChunks        int           `json:"total_chunks"`
	CurrentStep        string        `json:"current_step"`
	Percentage         float64       `json:"percentage"`
	StartTime          time.Time     `json:"start_time"`
	LastUpdateTime     time.Time     `json:"last_update_time"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
	ThroughputPerSec   float64       `json:"throughput_per_sec"`
}

// BatchResult 批量处理结果
type BatchResult struct {
	Success          bool                   `json:"success"`
	ProcessedCount   int                    `json:"processed_count"`
	SuccessCount     int                    `json:"success_count"`
	FailedCount      int                    `json:"failed_count"`
	SkippedCount     int                    `json:"skipped_count"`
	ProcessTime      time.Duration          `json:"process_time"`
	ThroughputPerSec float64                `json:"throughput_per_sec"`
	Summary          map[string]interface{} `json:"summary"`
	SuccessfulAssets []uint64               `json:"successful_assets"`
	FailedAssets     []string               `json:"failed_assets"`
}

// ChunkResult 分块处理结果
type ChunkResult struct {
	ChunkIndex     int                    `json:"chunk_index"`
	ChunkSize      int                    `json:"chunk_size"`
	Success        bool                   `json:"success"`
	ProcessedCount int                    `json:"processed_count"`
	SuccessCount   int                    `json:"success_count"`
	FailedCount    int                    `json:"failed_count"`
	ProcessTime    time.Duration          `json:"process_time"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	Summary        map[string]interface{} `json:"summary"`
}

// BatchError 批量处理错误
type BatchError struct {
	AssetID    string    `json:"asset_id"`
	ChunkIndex int       `json:"chunk_index"`
	ErrorType  string    `json:"error_type"`
	ErrorMsg   string    `json:"error_msg"`
	Timestamp  time.Time `json:"timestamp"`
	Retryable  bool      `json:"retryable"`
	RetryCount int       `json:"retry_count"`
}

// ConcurrencyPool 并发控制池
type ConcurrencyPool struct {
	semaphore   chan struct{}
	activeJobs  map[string]*JobInfo
	mutex       sync.RWMutex
	maxWorkers  int
	waitingJobs int
}

// JobInfo 作业信息
type JobInfo struct {
	JobID     string
	StartTime time.Time
	TaskType  string
	Status    string
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor(svcCtx *svc.ServiceContext) *BatchProcessor {
	config := &BatchConfig{
		MaxBatchSize:    10000,           // 最大批量大小
		ChunkSize:       100,             // 每块大小
		MaxConcurrency:  10,              // 最大并发数
		TimeoutPerChunk: 5 * time.Minute, // 每块超时时间
		EnableProgress:  true,            // 启用进度跟踪
		EnableRollback:  true,            // 启用回滚
		RetryAttempts:   3,               // 重试次数
		RetryDelay:      1 * time.Second, // 重试延迟
		EnableNotify:    true,            // 启用通知
	}

	return &BatchProcessor{
		svcCtx:          svcCtx,
		logger:          logx.WithContext(context.Background()),
		config:          config,
		taskManager:     NewTaskManager(svcCtx),
		dataPipeline:    NewDataPipeline(svcCtx),
		concurrencyPool: NewConcurrencyPool(config.MaxConcurrency),
	}
}

// NewConcurrencyPool 创建并发控制池
func NewConcurrencyPool(maxWorkers int) *ConcurrencyPool {
	return &ConcurrencyPool{
		semaphore:   make(chan struct{}, maxWorkers),
		activeJobs:  make(map[string]*JobInfo),
		maxWorkers:  maxWorkers,
		waitingJobs: 0,
	}
}

// SubmitBatchTask 提交批量处理任务
func (bp *BatchProcessor) SubmitBatchTask(ctx context.Context, data *input.ProcessingData, taskType string) (*BatchTask, error) {
	// 验证输入
	if data == nil || len(data.Assets) == 0 {
		return nil, fmt.Errorf("批量处理数据不能为空")
	}

	// 检查批量大小限制
	if len(data.Assets) > bp.config.MaxBatchSize {
		return nil, fmt.Errorf("批量大小超过限制: %d > %d", len(data.Assets), bp.config.MaxBatchSize)
	}

	// 创建批量任务
	task := &BatchTask{
		ID:         generateTaskID(),
		Type:       taskType,
		Status:     "pending",
		Priority:   input.PriorityNormal,
		Data:       data,
		Config:     bp.config,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		CreatedBy:  "system", // 可以从context获取用户信息
		Progress: &BatchProgress{
			TotalItems:     len(data.Assets),
			TotalChunks:    (len(data.Assets) + bp.config.ChunkSize - 1) / bp.config.ChunkSize,
			StartTime:      time.Now(),
			CurrentStep:    "initialization",
			LastUpdateTime: time.Now(),
		},
		ChunkResults: make([]*ChunkResult, 0),
		ErrorDetails: make([]*BatchError, 0),
	}

	// 提交任务到任务管理器
	if err := bp.taskManager.SubmitTask(ctx, task); err != nil {
		return nil, fmt.Errorf("提交任务失败: %v", err)
	}

	bp.logger.Infof("批量任务已提交: TaskID=%s, Type=%s, 数据量=%d",
		task.ID, taskType, len(data.Assets))

	return task, nil
}

// ProcessBatchTask 处理批量任务
func (bp *BatchProcessor) ProcessBatchTask(ctx context.Context, task *BatchTask) (*BatchResult, error) {
	startTime := time.Now()
	task.StartTime = &startTime
	task.Status = "processing"
	task.Progress.CurrentStep = "batch_processing"

	bp.logger.Infof("开始处理批量任务: TaskID=%s, 总数=%d", task.ID, task.Progress.TotalItems)

	// 初始化结果
	result := &BatchResult{
		ProcessedCount:   0,
		SuccessCount:     0,
		FailedCount:      0,
		SkippedCount:     0,
		SuccessfulAssets: make([]uint64, 0),
		FailedAssets:     make([]string, 0),
		Summary:          make(map[string]interface{}),
	}

	// 分块处理
	totalAssets := len(task.Data.Assets)
	chunkSize := bp.config.ChunkSize
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for i := 0; i < totalAssets; i += chunkSize {
		end := i + chunkSize
		if end > totalAssets {
			end = totalAssets
		}

		chunk := task.Data.Assets[i:end]
		chunkIndex := i / chunkSize

		wg.Add(1)
		go func(chunkIdx int, chunkData []*input.ProcessedAssetData) {
			defer wg.Done()

			// 获取并发许可
			bp.concurrencyPool.semaphore <- struct{}{}
			defer func() { <-bp.concurrencyPool.semaphore }()

			// 处理分块
			chunkResult := bp.processChunk(ctx, task, chunkIdx, chunkData)

			// 更新结果
			mutex.Lock()
			task.ChunkResults = append(task.ChunkResults, chunkResult)
			result.ProcessedCount += chunkResult.ProcessedCount
			result.SuccessCount += chunkResult.SuccessCount
			result.FailedCount += chunkResult.FailedCount

			// 更新进度
			task.Progress.ProcessedItems = result.ProcessedCount
			task.Progress.SuccessItems = result.SuccessCount
			task.Progress.FailedItems = result.FailedCount
			task.Progress.CurrentChunk = chunkIdx + 1
			task.Progress.Percentage = float64(result.ProcessedCount) / float64(task.Progress.TotalItems) * 100
			task.Progress.LastUpdateTime = time.Now()

			// 计算吞吐量
			elapsed := time.Since(task.Progress.StartTime)
			if elapsed.Seconds() > 0 {
				task.Progress.ThroughputPerSec = float64(result.ProcessedCount) / elapsed.Seconds()
			}

			mutex.Unlock()

			bp.logger.Infof("分块%d处理完成: 成功=%d, 失败=%d", chunkIdx, chunkResult.SuccessCount, chunkResult.FailedCount)

		}(chunkIndex, chunk)
	}

	// 等待所有分块完成
	wg.Wait()

	// 计算最终结果
	endTime := time.Now()
	task.EndTime = &endTime
	task.Status = "completed"
	task.Progress.CurrentStep = "completed"
	task.UpdateTime = endTime

	result.ProcessTime = endTime.Sub(startTime)
	if result.ProcessTime.Seconds() > 0 {
		result.ThroughputPerSec = float64(result.ProcessedCount) / result.ProcessTime.Seconds()
	}
	result.Success = result.FailedCount == 0

	// 更新任务结果
	task.Result = result

	// 构建汇总信息
	result.Summary = map[string]interface{}{
		"task_id":            task.ID,
		"task_type":          task.Type,
		"total_chunks":       len(task.ChunkResults),
		"processing_time":    result.ProcessTime.String(),
		"throughput_per_sec": result.ThroughputPerSec,
		"success_rate":       float64(result.SuccessCount) / float64(result.ProcessedCount) * 100,
	}

	bp.logger.Infof("批量任务处理完成: TaskID=%s, 成功=%d, 失败=%d, 耗时=%v",
		task.ID, result.SuccessCount, result.FailedCount, result.ProcessTime)

	return result, nil
}

// processChunk 处理单个分块
func (bp *BatchProcessor) processChunk(ctx context.Context, task *BatchTask, chunkIndex int, chunk []*input.ProcessedAssetData) *ChunkResult {
	startTime := time.Now()

	bp.logger.Infof("开始处理分块%d: 大小=%d", chunkIndex, len(chunk))

	chunkResult := &ChunkResult{
		ChunkIndex:     chunkIndex,
		ChunkSize:      len(chunk),
		ProcessedCount: 0,
		SuccessCount:   0,
		FailedCount:    0,
		Summary:        make(map[string]interface{}),
	}

	// 为分块创建处理数据
	chunkData := &input.ProcessingData{
		Assets:     chunk,
		BatchInfo:  task.Data.BatchInfo,
		Context:    task.Data.Context,
		ErrorCount: 0,
	}

	// 使用数据管道处理分块
	pipelineResult, err := bp.dataPipeline.Process(ctx, chunkData)
	if err != nil {
		chunkResult.Success = false
		chunkResult.ErrorMessage = err.Error()
		chunkResult.FailedCount = len(chunk)
		chunkResult.ProcessTime = time.Since(startTime)

		bp.logger.Errorf("分块%d处理失败: %v", chunkIndex, err)
		return chunkResult
	}

	// 统计结果
	chunkResult.ProcessedCount = pipelineResult.ProcessedCount
	chunkResult.SuccessCount = pipelineResult.PersistStats.SuccessCount
	chunkResult.FailedCount = pipelineResult.TotalErrors
	chunkResult.Success = chunkResult.FailedCount == 0
	chunkResult.ProcessTime = time.Since(startTime)

	// 构建分块汇总
	chunkResult.Summary = map[string]interface{}{
		"chunk_index":      chunkIndex,
		"validation_stats": pipelineResult.ValidationStats,
		"transform_stats":  pipelineResult.TransformStats,
		"persist_stats":    pipelineResult.PersistStats,
		"processing_time":  chunkResult.ProcessTime.String(),
	}

	return chunkResult
}

// GetTaskStatus 获取任务状态
func (bp *BatchProcessor) GetTaskStatus(taskID string) (*BatchTask, error) {
	return bp.taskManager.GetTask(taskID)
}

// CancelTask 取消任务
func (bp *BatchProcessor) CancelTask(ctx context.Context, taskID string) error {
	return bp.taskManager.CancelTask(ctx, taskID)
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("batch_%d", time.Now().UnixNano())
}

// GetInfo 获取批量处理器信息
func (bp *BatchProcessor) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "BatchProcessor",
		"version":     "v1.0.0",
		"description": "企业级批量处理器，支持大规模数据处理",
		"config":      bp.config,
		"features": []string{
			"并发处理",
			"进度跟踪",
			"错误处理",
			"事务回滚",
			"重试机制",
			"性能监控",
		},
	}
}
