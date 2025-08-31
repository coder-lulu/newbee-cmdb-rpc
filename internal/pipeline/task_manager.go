package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
)

// TaskManager 批量处理任务管理器
type TaskManager struct {
	svcCtx      *svc.ServiceContext
	logger      logx.Logger
	taskStorage map[string]*BatchTask
	taskQueue   chan *BatchTask
	workerPool  chan struct{}
	mutex       sync.RWMutex
	workerCount int
	started     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// TaskStats 任务统计信息
type TaskStats struct {
	TotalTasks      int `json:"total_tasks"`
	PendingTasks    int `json:"pending_tasks"`
	ProcessingTasks int `json:"processing_tasks"`
	CompletedTasks  int `json:"completed_tasks"`
	FailedTasks     int `json:"failed_tasks"`
	ActiveWorkers   int `json:"active_workers"`
}

// NewTaskManager 创建任务管理器
func NewTaskManager(svcCtx *svc.ServiceContext) *TaskManager {
	tm := &TaskManager{
		svcCtx:      svcCtx,
		logger:      logx.WithContext(context.Background()),
		taskStorage: make(map[string]*BatchTask),
		taskQueue:   make(chan *BatchTask, 1000), // 任务队列缓冲区
		workerPool:  make(chan struct{}, 10),     // 最多10个并发worker
		workerCount: 10,
		stopCh:      make(chan struct{}),
	}

	// 启动任务管理器
	tm.Start()
	return tm
}

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	if tm.started {
		return
	}

	tm.started = true
	tm.logger.Info("启动批量处理任务管理器...")

	// 启动工作协程
	for i := 0; i < tm.workerCount; i++ {
		tm.wg.Add(1)
		go tm.worker(i)
	}

	// 启动监控协程
	tm.wg.Add(1)
	go tm.monitor()

	tm.logger.Infof("任务管理器启动完成，工作协程数: %d", tm.workerCount)
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	if !tm.started {
		return
	}

	tm.logger.Info("停止批量处理任务管理器...")
	tm.started = false

	close(tm.stopCh)
	tm.wg.Wait()

	tm.logger.Info("任务管理器已停止")
}

// SubmitTask 提交任务
func (tm *TaskManager) SubmitTask(ctx context.Context, task *BatchTask) error {
	if !tm.started {
		return fmt.Errorf("任务管理器未启动")
	}

	// 验证任务
	if err := tm.validateTask(task); err != nil {
		return fmt.Errorf("任务验证失败: %v", err)
	}

	// 存储任务
	tm.mutex.Lock()
	tm.taskStorage[task.ID] = task
	tm.mutex.Unlock()

	// 提交到队列
	select {
	case tm.taskQueue <- task:
		tm.logger.Infof("任务已提交到队列: TaskID=%s, Type=%s", task.ID, task.Type)
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("任务队列已满，提交超时")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*BatchTask, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	task, exists := tm.taskStorage[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	return task, nil
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(ctx context.Context, taskID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task, exists := tm.taskStorage[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status == "processing" {
		// 如果任务正在处理中，标记为取消
		task.Status = "cancelled"
		task.UpdateTime = time.Now()
		tm.logger.Infof("任务已标记为取消: TaskID=%s", taskID)
		return nil
	}

	if task.Status == "pending" {
		// 如果任务还在等待，直接标记为取消
		task.Status = "cancelled"
		task.UpdateTime = time.Now()
		tm.logger.Infof("等待中的任务已取消: TaskID=%s", taskID)
		return nil
	}

	return fmt.Errorf("任务状态不允许取消: %s", task.Status)
}

// GetTaskStats 获取任务统计
func (tm *TaskManager) GetTaskStats() *TaskStats {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	stats := &TaskStats{
		TotalTasks:    len(tm.taskStorage),
		ActiveWorkers: tm.workerCount,
	}

	for _, task := range tm.taskStorage {
		switch task.Status {
		case "pending":
			stats.PendingTasks++
		case "processing":
			stats.ProcessingTasks++
		case "completed":
			stats.CompletedTasks++
		case "failed":
			stats.FailedTasks++
		}
	}

	return stats
}

// worker 工作协程
func (tm *TaskManager) worker(workerID int) {
	defer tm.wg.Done()

	tm.logger.Infof("工作协程-%d 已启动", workerID)

	for {
		select {
		case task := <-tm.taskQueue:
			// 获取工作许可
			tm.workerPool <- struct{}{}

			// 处理任务
			tm.processTask(workerID, task)

			// 释放工作许可
			<-tm.workerPool

		case <-tm.stopCh:
			tm.logger.Infof("工作协程-%d 收到停止信号", workerID)
			return
		}
	}
}

// processTask 处理任务
func (tm *TaskManager) processTask(workerID int, task *BatchTask) {
	startTime := time.Now()
	tm.logger.Infof("工作协程-%d 开始处理任务: TaskID=%s", workerID, task.ID)

	// 更新任务状态
	tm.updateTaskStatus(task.ID, "processing")

	// 创建处理上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 创建批量处理器进行实际处理
	processor := &BatchProcessor{
		svcCtx:          tm.svcCtx,
		logger:          tm.logger,
		config:          task.Config,
		dataPipeline:    NewDataPipeline(tm.svcCtx),
		concurrencyPool: NewConcurrencyPool(task.Config.MaxConcurrency),
	}

	// 执行批量处理
	result, err := processor.ProcessBatchTask(ctx, task)
	if err != nil {
		tm.logger.Errorf("工作协程-%d 任务处理失败: TaskID=%s, Error=%v", workerID, task.ID, err)
		tm.updateTaskStatus(task.ID, "failed")

		// 记录错误
		tm.mutex.Lock()
		if storedTask, exists := tm.taskStorage[task.ID]; exists {
			storedTask.ErrorDetails = append(storedTask.ErrorDetails, &BatchError{
				AssetID:   task.ID,
				ErrorType: "processing_error",
				ErrorMsg:  err.Error(),
				Timestamp: time.Now(),
				Retryable: true,
			})
		}
		tm.mutex.Unlock()
		return
	}

	// 更新任务结果
	tm.mutex.Lock()
	if storedTask, exists := tm.taskStorage[task.ID]; exists {
		storedTask.Result = result
		storedTask.Status = "completed"
		storedTask.UpdateTime = time.Now()
		endTime := time.Now()
		storedTask.EndTime = &endTime
	}
	tm.mutex.Unlock()

	processingTime := time.Since(startTime)
	tm.logger.Infof("工作协程-%d 任务处理完成: TaskID=%s, 耗时=%v, 成功=%d, 失败=%d",
		workerID, task.ID, processingTime, result.SuccessCount, result.FailedCount)
}

// updateTaskStatus 更新任务状态
func (tm *TaskManager) updateTaskStatus(taskID, status string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.taskStorage[taskID]; exists {
		task.Status = status
		task.UpdateTime = time.Now()

		if status == "processing" && task.StartTime == nil {
			startTime := time.Now()
			task.StartTime = &startTime
		}
	}
}

// monitor 监控协程
func (tm *TaskManager) monitor() {
	defer tm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tm.performMonitoring()
		case <-tm.stopCh:
			return
		}
	}
}

// performMonitoring 执行监控
func (tm *TaskManager) performMonitoring() {
	stats := tm.GetTaskStats()
	tm.logger.Infof("任务管理器状态 - 总任务=%d, 等待=%d, 处理中=%d, 完成=%d, 失败=%d",
		stats.TotalTasks, stats.PendingTasks, stats.ProcessingTasks,
		stats.CompletedTasks, stats.FailedTasks)

	// 清理完成的任务 (保留最近24小时的任务)
	tm.cleanupOldTasks()
}

// cleanupOldTasks 清理旧任务
func (tm *TaskManager) cleanupOldTasks() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	deletedCount := 0

	for taskID, task := range tm.taskStorage {
		if (task.Status == "completed" || task.Status == "failed") &&
			task.UpdateTime.Before(cutoff) {
			delete(tm.taskStorage, taskID)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		tm.logger.Infof("清理完成，删除旧任务数: %d", deletedCount)
	}
}

// validateTask 验证任务
func (tm *TaskManager) validateTask(task *BatchTask) error {
	if task.ID == "" {
		return fmt.Errorf("任务ID不能为空")
	}
	if task.Type == "" {
		return fmt.Errorf("任务类型不能为空")
	}
	if task.Data == nil {
		return fmt.Errorf("任务数据不能为空")
	}
	if len(task.Data.Assets) == 0 {
		return fmt.Errorf("任务资产数据不能为空")
	}
	if task.Config == nil {
		return fmt.Errorf("任务配置不能为空")
	}
	return nil
}

// GetAllTasks 获取所有任务
func (tm *TaskManager) GetAllTasks() map[string]*BatchTask {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	result := make(map[string]*BatchTask)
	for k, v := range tm.taskStorage {
		result[k] = v
	}
	return result
}

// GetTasksByStatus 根据状态获取任务
func (tm *TaskManager) GetTasksByStatus(status string) []*BatchTask {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var result []*BatchTask
	for _, task := range tm.taskStorage {
		if task.Status == status {
			result = append(result, task)
		}
	}
	return result
}
