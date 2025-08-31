package async

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/async"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/pipeline"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// EnhancedAsyncTaskLogic 增强版异步任务管理logic
type EnhancedAsyncTaskLogic struct {
	ctx              context.Context
	svcCtx           *svc.ServiceContext
	logger           logx.Logger
	batchProcessor   *pipeline.BatchProcessor
	batchOpProcessor *pipeline.BatchOperationProcessor
	taskManager      *pipeline.TaskManager

	// 新增：增强版异步任务管理器
	enhancedTaskManager *async.EnhancedAsyncTaskManager
	redisTaskStorage    *async.RedisTaskStorage
}

// NewEnhancedAsyncTaskLogic 创建增强版异步任务logic
func NewEnhancedAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnhancedAsyncTaskLogic {
	// 创建增强版任务管理器
	enhancedTaskManager := async.NewEnhancedAsyncTaskManager(svcCtx.Redis, nil)

	// 启动任务管理器
	go func() {
		if err := enhancedTaskManager.Start(ctx); err != nil {
			logx.Errorf("启动增强版任务管理器失败: %v", err)
		}
	}()

	return &EnhancedAsyncTaskLogic{
		ctx:                 ctx,
		svcCtx:              svcCtx,
		logger:              logx.WithContext(ctx),
		batchProcessor:      pipeline.NewBatchProcessor(svcCtx),
		batchOpProcessor:    pipeline.NewBatchOperationProcessor(svcCtx),
		taskManager:         pipeline.NewTaskManager(svcCtx),
		enhancedTaskManager: enhancedTaskManager,
		redisTaskStorage:    async.NewRedisTaskStorage(svcCtx.Redis),
	}
}

// SubmitAsyncTask 提交异步任务
func (l *EnhancedAsyncTaskLogic) SubmitAsyncTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
	l.logger.Infof("提交异步任务: Type=%s, Operation=%s", in.Type, in.Operation)

	// 验证请求参数
	if err := l.validateAsyncTaskReq(in); err != nil {
		return &cmdb.AsyncTaskResp{
			Success: false,
			Message: fmt.Sprintf("请求参数验证失败: %v", err),
		}, nil
	}

	// 根据任务类型处理
	switch in.Type {
	case "batch_delete", "batch_update", "batch_operation":
		return l.submitBatchOperationTask(in)
	default:
		return &cmdb.AsyncTaskResp{
			Success: false,
			Message: fmt.Sprintf("不支持的任务类型: %s", in.Type),
		}, nil
	}
}

// GetTaskStatus 获取任务状态
func (l *EnhancedAsyncTaskLogic) GetTaskStatus(in *cmdb.TaskStatusReq) (*cmdb.TaskStatusResp, error) {
	l.logger.Infof("查询任务状态: TaskID=%s", in.TaskId)

	// 首先尝试从Redis存储获取
	if task, err := l.redisTaskStorage.GetTask(l.ctx, in.TaskId); err == nil {
		taskStatus := l.convertImportTaskToStatusInfo(task)
		return &cmdb.TaskStatusResp{
			Task: taskStatus,
		}, nil
	}

	// 然后尝试从BatchProcessor获取（向后兼容）
	if task, err := l.batchProcessor.GetTaskStatus(in.TaskId); err == nil {
		taskStatus := l.convertBatchTaskToStatusInfo(task)
		return &cmdb.TaskStatusResp{
			Task: taskStatus,
		}, nil
	}

	// 最后尝试从TaskManager获取（向后兼容）
	if task, err := l.taskManager.GetTask(in.TaskId); err == nil {
		taskStatus := l.convertPipelineTaskToStatusInfo(task)
		return &cmdb.TaskStatusResp{
			Task: taskStatus,
		}, nil
	}

	return nil, fmt.Errorf("任务不存在: %s", in.TaskId)
}

// CancelTask 取消任务
func (l *EnhancedAsyncTaskLogic) CancelTask(in *cmdb.TaskCancelReq) (*cmdb.BaseResp, error) {
	l.logger.Infof("取消任务: TaskID=%s, Reason=%s", in.TaskId, in.GetReason())

	// 尝试从增强版任务管理器取消
	if err := l.enhancedTaskManager.CancelTask(l.ctx, in.TaskId); err == nil {
		l.logger.Infof("成功取消增强版任务: %s", in.TaskId)
		return &cmdb.BaseResp{
			Msg: "任务已取消",
		}, nil
	}

	// 尝试取消BatchProcessor中的任务（向后兼容）
	if err := l.batchProcessor.CancelTask(l.ctx, in.TaskId); err == nil {
		l.logger.Infof("成功取消批量处理任务: %s", in.TaskId)
		return &cmdb.BaseResp{
			Msg: "任务已取消",
		}, nil
	}

	// 尝试取消TaskManager中的任务（向后兼容）
	if err := l.taskManager.CancelTask(l.ctx, in.TaskId); err == nil {
		l.logger.Infof("成功取消管道任务: %s", in.TaskId)
		return &cmdb.BaseResp{
			Msg: "任务已取消",
		}, nil
	}

	return nil, fmt.Errorf("取消任务失败，任务可能不存在或无法取消: %s", in.TaskId)
}

// GetTaskList 获取任务列表
func (l *EnhancedAsyncTaskLogic) GetTaskList(in *cmdb.TaskListReq) (*cmdb.TaskListResp, error) {
	l.logger.Infof("获取任务列表: Page=%d, PageSize=%d", in.Page, in.PageSize)

	// 从Redis存储获取所有任务
	allTasks, err := l.redisTaskStorage.GetAllTasks(l.ctx)
	if err != nil {
		l.logger.Errorf("获取Redis任务列表失败: %v", err)
		// 降级到原有逻辑
		return l.getTaskListFallback(in)
	}

	// 转换为protobuf格式并应用过滤
	var tasks []*cmdb.TaskStatusInfo
	for _, task := range allTasks {
		// 简单的类型过滤
		if in.Type != nil && task.Type != *in.Type {
			continue
		}
		// 简单的状态过滤
		if in.Status != nil && task.Status != *in.Status {
			continue
		}

		taskStatus := l.convertImportTaskToStatusInfo(task)
		tasks = append(tasks, taskStatus)
	}

	// 简单分页处理
	total := uint64(len(tasks))
	start := (in.Page - 1) * in.PageSize
	end := start + in.PageSize

	if start > uint64(len(tasks)) {
		tasks = []*cmdb.TaskStatusInfo{}
	} else if end > uint64(len(tasks)) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	return &cmdb.TaskListResp{
		Total: total,
		Tasks: tasks,
	}, nil
}

// GetTaskStats 获取任务统计信息
func (l *EnhancedAsyncTaskLogic) GetTaskStats(in *cmdb.TaskStatsReq) (*cmdb.TaskStatsResp, error) {
	l.logger.Infof("获取任务统计信息")

	// 从Redis存储获取统计信息
	stats, err := l.enhancedTaskManager.GetTaskStats(l.ctx)
	if err != nil {
		l.logger.Errorf("获取Redis任务统计失败: %v", err)
		// 降级到原有逻辑
		return l.getTaskStatsFallback(in)
	}

	// 计算成功率
	var successRate float32 = 0
	if stats.TotalTasks > 0 {
		successRate = float32(stats.CompletedTasks) / float32(stats.TotalTasks) * 100
	}

	// 转换为protobuf格式
	taskStats := &cmdb.TaskStatistics{
		TotalTasks:            int32(stats.TotalTasks),
		PendingTasks:          int32(stats.PendingTasks),
		ProcessingTasks:       int32(stats.ProcessingTasks),
		CompletedTasks:        int32(stats.CompletedTasks),
		FailedTasks:           int32(stats.FailedTasks),
		CancelledTasks:        0, // 暂时设为0，可以后续扩展
		AverageProcessingTime: 0, // 暂时设为0，可以后续扩展
		SuccessRate:           successRate,
		TypeStats:             []*cmdb.TaskTypeStats{},
	}

	return &cmdb.TaskStatsResp{
		Stats: taskStats,
	}, nil
}

// submitBatchOperationTask 提交批量操作任务
func (l *EnhancedAsyncTaskLogic) submitBatchOperationTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
	// 使用增强版任务管理器
	task, err := l.enhancedTaskManager.SubmitBatchTask(l.ctx, in.Operation, in.CiIds, in.Params)
	if err != nil {
		return &cmdb.AsyncTaskResp{
			Success: false,
			Message: fmt.Sprintf("提交批量操作任务失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	initialStatus := l.convertImportTaskToStatusInfo(task)

	return &cmdb.AsyncTaskResp{
		TaskId:        task.ID,
		Success:       true,
		Message:       "任务已成功提交",
		InitialStatus: initialStatus,
	}, nil
}

// 转换函数：ImportTask -> TaskStatusInfo
func (l *EnhancedAsyncTaskLogic) convertImportTaskToStatusInfo(task *async.ImportTask) *cmdb.TaskStatusInfo {
	var result *cmdb.TaskResultInfo
	var startTime, endTime *int64

	// 转换时间
	createTime := task.CreateTime.Unix()
	updateTime := task.UpdateTime.Unix()

	// 如果有开始时间
	if !task.CreateTime.IsZero() {
		st := task.CreateTime.Unix()
		startTime = &st
	}

	// 如果任务已完成，设置结束时间
	if task.Status == "completed" || task.Status == "failed" {
		et := task.UpdateTime.Unix()
		endTime = &et
	}

	// 构建进度信息
	progress := &cmdb.TaskProgressInfo{
		TotalItems:     int32(task.Progress.TotalItems),
		ProcessedItems: int32(task.Progress.ProcessedItems),
		CurrentStep:    task.Progress.CurrentStep,
		Percentage:     float32(task.Progress.Percentage),
	}

	// 构建结果信息（如果任务已完成）
	if task.Status == "completed" || task.Status == "failed" {
		success := task.Status == "completed"
		result = &cmdb.TaskResultInfo{
			Success:        success,
			ProcessedCount: int32(task.Progress.ProcessedItems),
			SuccessCount:   int32(task.Progress.ProcessedItems),
			FailedCount:    0,
		}
	}

	// 构建错误信息
	var errorMsg *string
	if task.Error != "" {
		errorMsg = &task.Error
	}

	return &cmdb.TaskStatusInfo{
		Id:         task.ID,
		Type:       task.Type,
		Status:     task.Status,
		Priority:   1, // 默认优先级
		Progress:   progress,
		Result:     result,
		CreateTime: createTime,
		UpdateTime: updateTime,
		StartTime:  startTime,
		EndTime:    endTime,
		CreatedBy:  "system",
		Error:      errorMsg,
	}
}

// 其他转换函数和工具函数（向后兼容）
func (l *EnhancedAsyncTaskLogic) convertBatchTaskToStatusInfo(task *pipeline.BatchTask) *cmdb.TaskStatusInfo {
	// 原有的转换逻辑...
	return &cmdb.TaskStatusInfo{
		Id:         task.ID,
		Type:       task.Type,
		Status:     task.Status,
		CreateTime: task.CreateTime.Unix(),
		UpdateTime: task.UpdateTime.Unix(),
	}
}

func (l *EnhancedAsyncTaskLogic) convertPipelineTaskToStatusInfo(task *pipeline.BatchTask) *cmdb.TaskStatusInfo {
	// 原有的转换逻辑...
	return &cmdb.TaskStatusInfo{
		Id:         task.ID,
		Type:       task.Type,
		Status:     task.Status,
		CreateTime: task.CreateTime.Unix(),
		UpdateTime: task.UpdateTime.Unix(),
	}
}

// 降级处理函数
func (l *EnhancedAsyncTaskLogic) getTaskListFallback(in *cmdb.TaskListReq) (*cmdb.TaskListResp, error) {
	// 使用原有的TaskManager逻辑
	stats := l.taskManager.GetTaskStats()
	tasks := []*cmdb.TaskStatusInfo{}

	allTasks := l.taskManager.GetAllTasks()
	for _, task := range allTasks {
		if in.Type != nil && task.Type != *in.Type {
			continue
		}
		if in.Status != nil && task.Status != *in.Status {
			continue
		}

		taskStatus := l.convertBatchTaskToStatusInfo(task)
		tasks = append(tasks, taskStatus)
	}

	return &cmdb.TaskListResp{
		Total: uint64(stats.TotalTasks),
		Tasks: tasks,
	}, nil
}

func (l *EnhancedAsyncTaskLogic) getTaskStatsFallback(in *cmdb.TaskStatsReq) (*cmdb.TaskStatsResp, error) {
	// 使用原有的TaskManager逻辑
	stats := l.taskManager.GetTaskStats()

	var successRate float32 = 0
	if stats.TotalTasks > 0 {
		successRate = float32(stats.CompletedTasks) / float32(stats.TotalTasks) * 100
	}

	taskStats := &cmdb.TaskStatistics{
		TotalTasks:      int32(stats.TotalTasks),
		PendingTasks:    int32(stats.PendingTasks),
		ProcessingTasks: int32(stats.ProcessingTasks),
		CompletedTasks:  int32(stats.CompletedTasks),
		FailedTasks:     int32(stats.FailedTasks),
		SuccessRate:     successRate,
	}

	return &cmdb.TaskStatsResp{
		Stats: taskStats,
	}, nil
}

// 工具函数
func (l *EnhancedAsyncTaskLogic) validateAsyncTaskReq(in *cmdb.AsyncTaskReq) error {
	if in.Type == "" {
		return fmt.Errorf("任务类型不能为空")
	}

	switch in.Type {
	case "batch_operation", "batch_delete", "batch_update":
		if len(in.CiIds) == 0 {
			return fmt.Errorf("批量操作需要提供CI ID列表")
		}
		if in.Operation == "" {
			return fmt.Errorf("批量操作需要指定操作类型")
		}
	}

	return nil
}
