package async

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/pipeline"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// AsyncTaskLogic 异步任务管理logic
type AsyncTaskLogic struct {
	ctx              context.Context
	svcCtx           *svc.ServiceContext
	logger           logx.Logger
	batchProcessor   *pipeline.BatchProcessor
	batchOpProcessor *pipeline.BatchOperationProcessor
	taskManager      *pipeline.TaskManager
}

// NewAsyncTaskLogic 创建异步任务logic
func NewAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AsyncTaskLogic {
	return &AsyncTaskLogic{
		ctx:              ctx,
		svcCtx:           svcCtx,
		logger:           logx.WithContext(ctx),
		batchProcessor:   pipeline.NewBatchProcessor(svcCtx),
		batchOpProcessor: pipeline.NewBatchOperationProcessor(svcCtx),
		taskManager:      pipeline.NewTaskManager(svcCtx),
	}
}

// SubmitAsyncTask 提交异步任务
func (l *AsyncTaskLogic) SubmitAsyncTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
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
	case "batch_import":
		return l.submitBatchImportTask(in)
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
func (l *AsyncTaskLogic) GetTaskStatus(in *cmdb.TaskStatusReq) (*cmdb.TaskStatusResp, error) {
	l.logger.Infof("查询任务状态: TaskID=%s", in.TaskId)

	// 首先尝试从BatchProcessor获取
	if task, err := l.batchProcessor.GetTaskStatus(in.TaskId); err == nil {
		taskStatus := l.convertBatchTaskToStatusInfo(task)
		return &cmdb.TaskStatusResp{
			Task: taskStatus,
		}, nil
	}

	// 然后尝试从TaskManager获取
	if task, err := l.taskManager.GetTask(in.TaskId); err == nil {
		taskStatus := l.convertPipelineTaskToStatusInfo(task)
		return &cmdb.TaskStatusResp{
			Task: taskStatus,
		}, nil
	}

	return nil, fmt.Errorf("任务不存在: %s", in.TaskId)
}

// CancelTask 取消任务
func (l *AsyncTaskLogic) CancelTask(in *cmdb.TaskCancelReq) (*cmdb.BaseResp, error) {
	l.logger.Infof("取消任务: TaskID=%s, Reason=%s", in.TaskId, in.GetReason())

	// 尝试取消BatchProcessor中的任务
	if err := l.batchProcessor.CancelTask(l.ctx, in.TaskId); err == nil {
		l.logger.Infof("成功取消批量处理任务: %s", in.TaskId)
		return &cmdb.BaseResp{
			Msg: "任务已取消",
		}, nil
	}

	// 尝试取消TaskManager中的任务
	if err := l.taskManager.CancelTask(l.ctx, in.TaskId); err == nil {
		l.logger.Infof("成功取消管道任务: %s", in.TaskId)
		return &cmdb.BaseResp{
			Msg: "任务已取消",
		}, nil
	}

	return nil, fmt.Errorf("取消任务失败，任务可能不存在或无法取消: %s", in.TaskId)
}

// submitBatchImportTask 提交批量导入任务
func (l *AsyncTaskLogic) submitBatchImportTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
	// 解析数据
	var importData *input.ProcessingData
	if in.Data != nil {
		if err := json.Unmarshal([]byte(*in.Data), &importData); err != nil {
			return &cmdb.AsyncTaskResp{
				Success: false,
				Message: fmt.Sprintf("解析导入数据失败: %v", err),
			}, nil
		}
	} else {
		// 创建默认的导入数据（用于测试）
		importData = l.createDefaultImportData(in)
	}

	// 提交到批量处理器
	task, err := l.batchProcessor.SubmitBatchTask(l.ctx, importData, in.Type)
	if err != nil {
		return &cmdb.AsyncTaskResp{
			Success: false,
			Message: fmt.Sprintf("提交批量导入任务失败: %v", err),
		}, nil
	}

	// 转换任务状态
	initialStatus := l.convertBatchTaskToStatusInfo(task)

	return &cmdb.AsyncTaskResp{
		TaskId:        task.ID,
		Success:       true,
		Message:       "批量导入任务已提交",
		InitialStatus: initialStatus,
	}, nil
}

// submitBatchOperationTask 提交批量操作任务
func (l *AsyncTaskLogic) submitBatchOperationTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
	// 构建批量操作请求
	batchOpReq := &cmdb.CisBatchOperationReq{
		Operation: in.Operation,
		CiIds:     in.CiIds,
		Params:    in.Params,
	}

	// 提交到批量操作处理器
	response, err := l.batchOpProcessor.ProcessBatchOperation(l.ctx, batchOpReq)
	if err != nil {
		return &cmdb.AsyncTaskResp{
			Success: false,
			Message: fmt.Sprintf("提交批量操作任务失败: %v", err),
		}, nil
	}

	// 如果是异步任务，返回任务ID
	if response.TaskID != "" {
		// 获取任务状态详情
		var initialStatus *cmdb.TaskStatusInfo
		if task, err := l.batchOpProcessor.GetBatchTaskStatus(response.TaskID); err == nil {
			initialStatus = l.convertBatchTaskToStatusInfo(task)
		}

		return &cmdb.AsyncTaskResp{
			TaskId:        response.TaskID,
			Success:       response.Success,
			Message:       response.Message,
			InitialStatus: initialStatus,
		}, nil
	}

	// 同步处理完成
	return &cmdb.AsyncTaskResp{
		TaskId:  "", // 同步任务没有TaskID
		Success: response.Success,
		Message: response.Message,
	}, nil
}

// createDefaultImportData 创建默认的导入数据（用于测试）
func (l *AsyncTaskLogic) createDefaultImportData(in *cmdb.AsyncTaskReq) *input.ProcessingData {
	assets := make([]*input.ProcessedAssetData, 0, len(in.CiIds))

	for i, ciID := range in.CiIds {
		rawAsset := &input.RawAssetData{
			ID:         fmt.Sprintf("ci_%d", ciID),
			CITypeID:   1, // 默认CI类型
			CITypeName: "默认类型",
			Attributes: map[string]interface{}{
				"name":        fmt.Sprintf("Asset_%d", i+1),
				"description": "通过异步任务创建",
			},
			Source: "async_task",
		}

		asset := &input.ProcessedAssetData{
			RawAssetData:   rawAsset,
			Status:         "pending",
			ProcessorChain: []string{},
		}

		assets = append(assets, asset)
	}

	return &input.ProcessingData{
		Assets: assets,
		BatchInfo: &input.BatchInfo{
			ID:         fmt.Sprintf("async_batch_%d", time.Now().UnixNano()),
			Type:       in.Type,
			Source:     "async_task",
			TotalCount: len(assets),
			CreateTime: time.Now(),
			CreatedBy:  in.GetCreatedBy(),
			Status:     "processing",
		},
		Context: map[string]interface{}{
			"async_task": true,
			"operation":  in.Operation,
		},
		ErrorCount: 0,
	}
}

// convertBatchTaskToStatusInfo 转换BatchTask为TaskStatusInfo
func (l *AsyncTaskLogic) convertBatchTaskToStatusInfo(task *pipeline.BatchTask) *cmdb.TaskStatusInfo {
	status := &cmdb.TaskStatusInfo{
		Id:         task.ID,
		Type:       task.Type,
		Status:     task.Status,
		Priority:   int32(task.Priority),
		CreatedBy:  task.CreatedBy,
		CreateTime: task.CreateTime.Unix(),
		UpdateTime: task.UpdateTime.Unix(),
	}

	if task.StartTime != nil {
		startTime := task.StartTime.Unix()
		status.StartTime = &startTime
	}

	if task.EndTime != nil {
		endTime := task.EndTime.Unix()
		status.EndTime = &endTime
	}

	// 转换进度信息
	if task.Progress != nil {
		status.Progress = &cmdb.TaskProgressInfo{
			TotalItems:     int32(task.Progress.TotalItems),
			ProcessedItems: int32(task.Progress.ProcessedItems),
			SuccessItems:   int32(task.Progress.SuccessItems),
			FailedItems:    int32(task.Progress.FailedItems),
			Percentage:     float32(task.Progress.Percentage),
			CurrentStep:    task.Progress.CurrentStep,
		}

		if task.Progress.ThroughputPerSec != 0 {
			throughputPerSec := float32(task.Progress.ThroughputPerSec)
			status.Progress.ThroughputPerSec = &throughputPerSec
		}

		if task.Progress.EstimatedRemaining > 0 {
			estimatedRemaining := int64(task.Progress.EstimatedRemaining.Seconds())
			status.Progress.EstimatedRemaining = &estimatedRemaining
		}
	}

	// 转换结果信息
	if task.Result != nil {
		status.Result = &cmdb.TaskResultInfo{
			Success:          task.Result.Success,
			ProcessedCount:   int32(task.Result.ProcessedCount),
			SuccessCount:     int32(task.Result.SuccessCount),
			FailedCount:      int32(task.Result.FailedCount),
			SkippedCount:     int32(task.Result.SkippedCount),
			ProcessTimeMs:    task.Result.ProcessTime.Milliseconds(),
			ThroughputPerSec: float32(task.Result.ThroughputPerSec),
		}

		// 转换成功和失败的资产ID
		for _, assetID := range task.Result.SuccessfulAssets {
			status.Result.SuccessfulAssets = append(status.Result.SuccessfulAssets, strconv.FormatUint(assetID, 10))
		}

		for _, assetID := range task.Result.FailedAssets {
			status.Result.FailedAssets = append(status.Result.FailedAssets, assetID)
		}

		// 转换汇总信息
		if summaryBytes, err := json.Marshal(task.Result.Summary); err == nil {
			summary := string(summaryBytes)
			status.Result.Summary = &summary
		}
	}

	// 处理错误信息
	if len(task.ErrorDetails) > 0 && task.Status == "failed" {
		errorMsg := fmt.Sprintf("任务执行失败，错误数量: %d", len(task.ErrorDetails))
		status.Error = &errorMsg
	}

	return status
}

// convertPipelineTaskToStatusInfo 转换pipeline.BatchTask为TaskStatusInfo
func (l *AsyncTaskLogic) convertPipelineTaskToStatusInfo(task *pipeline.BatchTask) *cmdb.TaskStatusInfo {
	// 由于pipeline.BatchTask和我们的BatchTask结构相同，直接复用转换函数
	return l.convertBatchTaskToStatusInfo(task)
}

// GetTaskList 获取任务列表
func (l *AsyncTaskLogic) GetTaskList(in *cmdb.TaskListReq) (*cmdb.TaskListResp, error) {
	l.logger.Infof("获取任务列表: Page=%d, PageSize=%d", in.Page, in.PageSize)

	// 获取任务统计信息
	stats := l.taskManager.GetTaskStats()

	// 模拟任务列表数据（实际应该从任务管理器获取）
	tasks := []*cmdb.TaskStatusInfo{}

	// 从TaskManager获取所有任务
	allTasks := l.taskManager.GetAllTasks()

	// 转换为protobuf格式并应用过滤
	for _, task := range allTasks {
		// 简单的类型过滤
		if in.Type != nil && task.Type != *in.Type {
			continue
		}
		// 简单的状态过滤
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

// GetTaskStats 获取任务统计信息
func (l *AsyncTaskLogic) GetTaskStats(in *cmdb.TaskStatsReq) (*cmdb.TaskStatsResp, error) {
	l.logger.Infof("获取任务统计信息")

	// 获取任务统计信息
	stats := l.taskManager.GetTaskStats()

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
		CancelledTasks:        0, // TaskStats中没有这个字段，设为0
		AverageProcessingTime: 0, // 暂时设为0，后续可以扩展
		SuccessRate:           successRate,
		TypeStats:             []*cmdb.TaskTypeStats{},
	}

	return &cmdb.TaskStatsResp{
		Stats: taskStats,
	}, nil
}

// validateAsyncTaskReq 验证异步任务请求
func (l *AsyncTaskLogic) validateAsyncTaskReq(in *cmdb.AsyncTaskReq) error {
	if in.Type == "" {
		return fmt.Errorf("任务类型不能为空")
	}

	switch in.Type {
	case "batch_import":
		// 导入任务可以没有CI ID，但需要有数据
		if in.Data == nil && len(in.CiIds) == 0 {
			return fmt.Errorf("批量导入任务需要提供数据或CI ID")
		}
	case "batch_delete", "batch_update", "batch_operation":
		// 操作任务必须有CI ID和操作类型
		if len(in.CiIds) == 0 {
			return fmt.Errorf("批量操作任务需要提供CI ID列表")
		}
		if in.Operation == "" {
			return fmt.Errorf("批量操作任务需要指定操作类型")
		}
	default:
		return fmt.Errorf("不支持的任务类型: %s", in.Type)
	}

	return nil
}
