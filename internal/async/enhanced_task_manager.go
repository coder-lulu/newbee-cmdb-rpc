package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// EnhancedAsyncTaskManager 增强版异步任务管理器
type EnhancedAsyncTaskManager struct {
	taskQueue   chan *ImportTask
	taskStorage *RedisTaskStorage
	workersPool chan struct{}
	logger      logx.Logger
	started     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
	config      *TaskManagerConfig
}

// TaskManagerConfig 任务管理器配置
type TaskManagerConfig struct {
	MaxWorkers        int           `json:"max_workers"`        // 最大工作协程数
	QueueSize         int           `json:"queue_size"`         // 队列缓冲区大小
	CleanupInterval   time.Duration `json:"cleanup_interval"`   // 清理间隔
	RecoveryOnStart   bool          `json:"recovery_on_start"`  // 启动时是否恢复任务
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // 心跳间隔
}

// DefaultTaskManagerConfig 默认配置
func DefaultTaskManagerConfig() *TaskManagerConfig {
	return &TaskManagerConfig{
		MaxWorkers:        10,
		QueueSize:         1000,
		CleanupInterval:   1 * time.Hour,
		RecoveryOnStart:   true,
		HeartbeatInterval: 30 * time.Second,
	}
}

// NewEnhancedAsyncTaskManager 创建增强版异步任务管理器
func NewEnhancedAsyncTaskManager(redisClient redis.UniversalClient, config *TaskManagerConfig) *EnhancedAsyncTaskManager {
	if config == nil {
		config = DefaultTaskManagerConfig()
	}

	manager := &EnhancedAsyncTaskManager{
		taskQueue:   make(chan *ImportTask, config.QueueSize),
		taskStorage: NewRedisTaskStorage(redisClient),
		workersPool: make(chan struct{}, config.MaxWorkers),
		logger:      logx.WithContext(context.Background()),
		stopCh:      make(chan struct{}),
		config:      config,
	}

	return manager
}

// Start 启动任务管理器
func (m *EnhancedAsyncTaskManager) Start(ctx context.Context) error {
	if m.started {
		return fmt.Errorf("任务管理器已启动")
	}

	m.started = true
	m.logger.Info("启动增强版异步任务管理器...")

	// 恢复待处理任务
	if m.config.RecoveryOnStart {
		if err := m.recoverTasks(ctx); err != nil {
			m.logger.Errorf("任务恢复失败: %v", err)
		}
	}

	// 启动工作协程
	for i := 0; i < m.config.MaxWorkers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}

	// 启动清理协程
	m.wg.Add(1)
	go m.cleanupWorker()

	// 启动心跳协程
	m.wg.Add(1)
	go m.heartbeatWorker()

	m.logger.Infof("任务管理器启动完成，工作协程数: %d, 队列大小: %d",
		m.config.MaxWorkers, m.config.QueueSize)
	return nil
}

// Stop 停止任务管理器
func (m *EnhancedAsyncTaskManager) Stop() {
	if !m.started {
		return
	}

	m.logger.Info("停止异步任务管理器...")
	m.started = false

	close(m.stopCh)
	m.wg.Wait()

	m.logger.Info("异步任务管理器已停止")
}

// SubmitBatchTask 提交批量任务
func (m *EnhancedAsyncTaskManager) SubmitBatchTask(ctx context.Context, operation string, ciIds []uint64, params *string) (*ImportTask, error) {
	if !m.started {
		return nil, fmt.Errorf("任务管理器未启动")
	}

	task := &ImportTask{
		ID:         generateTaskID(),
		Type:       "batch_operation",
		Status:     "pending",
		Progress:   &TaskProgress{TotalItems: len(ciIds)},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 保存到Redis
	err := m.taskStorage.SaveTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("保存任务失败: %v", err)
	}

	// 提交到队列
	select {
	case m.taskQueue <- task:
		m.logger.Infof("任务已提交: TaskID=%s, Type=%s, CIs=%d", task.ID, task.Type, len(ciIds))
		return task, nil
	case <-time.After(5 * time.Second):
		// 队列满，更新任务状态为失败
		task.Status = "failed"
		task.Error = "任务队列已满，提交超时"
		m.taskStorage.UpdateTask(ctx, task)
		return nil, fmt.Errorf("任务队列已满，提交超时")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetTaskStatus 获取任务状态
func (m *EnhancedAsyncTaskManager) GetTaskStatus(ctx context.Context, taskID string) (*ImportTask, error) {
	return m.taskStorage.GetTask(ctx, taskID)
}

// CancelTask 取消任务
func (m *EnhancedAsyncTaskManager) CancelTask(ctx context.Context, taskID string) error {
	task, err := m.taskStorage.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
		return fmt.Errorf("任务已完成，无法取消: %s", task.Status)
	}

	task.Status = "cancelled"
	task.Error = "用户取消"
	return m.taskStorage.UpdateTask(ctx, task)
}

// GetTaskStats 获取任务统计信息
func (m *EnhancedAsyncTaskManager) GetTaskStats(ctx context.Context) (*TaskStats, error) {
	stats, err := m.taskStorage.GetTaskStats(ctx)
	if err != nil {
		return nil, err
	}

	// 添加活跃工作协程数
	stats.ActiveWorkers = m.config.MaxWorkers - len(m.workersPool)
	return stats, nil
}

// GetTasksByStatus 根据状态获取任务列表
func (m *EnhancedAsyncTaskManager) GetTasksByStatus(ctx context.Context, status string) ([]*ImportTask, error) {
	return m.taskStorage.GetTasksByStatus(ctx, status)
}

// recoverTasks 恢复待处理任务
func (m *EnhancedAsyncTaskManager) recoverTasks(ctx context.Context) error {
	m.logger.Info("开始恢复待处理任务...")

	tasks, err := m.taskStorage.RecoverPendingTasks(ctx)
	if err != nil {
		return fmt.Errorf("恢复任务失败: %v", err)
	}

	// 重新提交到队列
	recoveredCount := 0
	for _, task := range tasks {
		select {
		case m.taskQueue <- task:
			recoveredCount++
		default:
			m.logger.Errorf("队列已满，无法恢复任务: TaskID=%s", task.ID)
		}
	}

	m.logger.Infof("任务恢复完成，恢复任务数: %d/%d", recoveredCount, len(tasks))
	return nil
}

// worker 工作协程
func (m *EnhancedAsyncTaskManager) worker(workerID int) {
	defer m.wg.Done()
	m.logger.Infof("工作协程-%d 已启动", workerID)

	for {
		select {
		case task := <-m.taskQueue:
			// 获取工作许可
			m.workersPool <- struct{}{}

			// 处理任务
			m.processTask(workerID, task)

			// 释放工作许可
			<-m.workersPool

		case <-m.stopCh:
			m.logger.Infof("工作协程-%d 收到停止信号", workerID)
			return
		}
	}
}

// processTask 处理任务
func (m *EnhancedAsyncTaskManager) processTask(workerID int, task *ImportTask) {
	ctx := context.Background()
	startTime := time.Now()

	m.logger.Infof("工作协程-%d 开始处理任务: TaskID=%s", workerID, task.ID)

	// 更新任务状态为处理中
	task.Status = "processing"
	task.UpdateTime = time.Now()
	err := m.taskStorage.UpdateTask(ctx, task)
	if err != nil {
		m.logger.Errorf("更新任务状态失败: %v", err)
		return
	}

	// 检查任务是否被取消
	currentTask, err := m.taskStorage.GetTask(ctx, task.ID)
	if err != nil {
		m.logger.Errorf("获取任务状态失败: %v", err)
		return
	}

	if currentTask.Status == "cancelled" {
		m.logger.Infof("任务已被取消: TaskID=%s", task.ID)
		return
	}

	// 模拟任务处理 (实际应该调用具体的业务逻辑)
	err = m.executeTask(ctx, task)

	// 更新任务状态
	task.UpdateTime = time.Now()
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		m.logger.Errorf("任务处理失败: TaskID=%s, Error=%v", task.ID, err)
	} else {
		task.Status = "completed"
		task.Progress.Percentage = 100
		task.Progress.CurrentStep = "完成"
		m.logger.Infof("任务处理成功: TaskID=%s", task.ID)
	}

	// 保存最终状态
	err = m.taskStorage.UpdateTask(ctx, task)
	if err != nil {
		m.logger.Errorf("保存任务最终状态失败: %v", err)
	}

	duration := time.Since(startTime)
	m.logger.Infof("工作协程-%d 完成任务: TaskID=%s, 耗时=%v", workerID, task.ID, duration)
}

// executeTask 执行具体的任务逻辑
func (m *EnhancedAsyncTaskManager) executeTask(ctx context.Context, task *ImportTask) error {
	// 这里应该根据任务类型调用具体的业务逻辑
	// 目前只是模拟处理过程

	// 模拟处理步骤
	steps := []string{"验证数据", "处理数据", "保存结果"}
	for i, step := range steps {
		// 检查任务是否被取消
		currentTask, err := m.taskStorage.GetTask(ctx, task.ID)
		if err != nil {
			return fmt.Errorf("获取任务状态失败: %v", err)
		}
		if currentTask.Status == "cancelled" {
			return fmt.Errorf("任务已被取消")
		}

		// 更新进度
		task.Progress.CurrentStep = step
		task.Progress.Percentage = float64((i + 1) * 100 / len(steps))
		m.taskStorage.UpdateTask(ctx, task)

		// 模拟处理时间
		time.Sleep(1 * time.Second)
	}

	return nil
}

// cleanupWorker 清理工作协程
func (m *EnhancedAsyncTaskManager) cleanupWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			err := m.taskStorage.CleanupExpiredTasks(ctx)
			if err != nil {
				m.logger.Errorf("清理过期任务失败: %v", err)
			}

		case <-m.stopCh:
			m.logger.Info("清理协程收到停止信号")
			return
		}
	}
}

// heartbeatWorker 心跳工作协程
func (m *EnhancedAsyncTaskManager) heartbeatWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			stats, err := m.GetTaskStats(ctx)
			if err != nil {
				m.logger.Errorf("获取任务统计失败: %v", err)
				continue
			}

			m.logger.Infof("任务管理器心跳 - 总任务=%d, 等待=%d, 处理中=%d, 完成=%d, 失败=%d, 活跃工作协程=%d",
				stats.TotalTasks, stats.PendingTasks, stats.ProcessingTasks,
				stats.CompletedTasks, stats.FailedTasks, stats.ActiveWorkers)

		case <-m.stopCh:
			m.logger.Info("心跳协程收到停止信号")
			return
		}
	}
}
