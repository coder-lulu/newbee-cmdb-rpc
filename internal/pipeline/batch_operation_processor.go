package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"gitee.com/link234/cmdb-rpc/internal/adapters/input"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
)

// BatchOperationProcessor 批量操作处理器
type BatchOperationProcessor struct {
	svcCtx         *svc.ServiceContext
	logger         logx.Logger
	batchProcessor *BatchProcessor
	config         *BatchOperationConfig
}

// BatchOperationConfig 批量操作配置
type BatchOperationConfig struct {
	SyncThreshold   int           `json:"sync_threshold"`    // 同步处理阈值
	AsyncThreshold  int           `json:"async_threshold"`   // 异步处理阈值
	ChunkSize       int           `json:"chunk_size"`        // 分块大小
	MaxConcurrency  int           `json:"max_concurrency"`   // 最大并发数
	TimeoutPerChunk time.Duration `json:"timeout_per_chunk"` // 每块超时时间
	EnableProgress  bool          `json:"enable_progress"`   // 是否启用进度跟踪
	EnableRollback  bool          `json:"enable_rollback"`   // 是否启用回滚
	RetryAttempts   int           `json:"retry_attempts"`    // 重试次数
	RetryDelay      time.Duration `json:"retry_delay"`       // 重试延迟
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	Success        bool                   `json:"success"`
	Message        string                 `json:"message"`
	TaskID         string                 `json:"task_id,omitempty"`         // 异步任务ID
	ProcessedCount int                    `json:"processed_count,omitempty"` // 同步处理时的数量
	Summary        map[string]interface{} `json:"summary,omitempty"`
}

// NewBatchOperationProcessor 创建批量操作处理器
func NewBatchOperationProcessor(svcCtx *svc.ServiceContext) *BatchOperationProcessor {
	config := &BatchOperationConfig{
		SyncThreshold:   100,             // 100条以下同步处理
		AsyncThreshold:  1000,            // 1000条以上使用异步处理
		ChunkSize:       50,              // 每块50条
		MaxConcurrency:  5,               // 最大5个并发
		TimeoutPerChunk: 2 * time.Minute, // 每块2分钟超时
		EnableProgress:  true,            // 启用进度跟踪
		EnableRollback:  true,            // 启用回滚
		RetryAttempts:   3,               // 重试3次
		RetryDelay:      1 * time.Second, // 重试延迟1秒
	}

	return &BatchOperationProcessor{
		svcCtx:         svcCtx,
		logger:         logx.WithContext(context.Background()),
		batchProcessor: NewBatchProcessor(svcCtx),
		config:         config,
	}
}

// ProcessBatchOperation 处理批量操作
func (p *BatchOperationProcessor) ProcessBatchOperation(ctx context.Context, in *cmdb.CisBatchOperationReq) (*BatchOperationResponse, error) {
	// 参数验证
	if len(in.CiIds) == 0 {
		return nil, fmt.Errorf("操作需要提供CI实例ID列表")
	}

	p.logger.Infof("开始批量操作: 操作类型=%s, CI数量=%d", in.Operation, len(in.CiIds))

	// 根据数据量选择处理策略
	if len(in.CiIds) <= p.config.SyncThreshold {
		// 小批量：使用同步处理
		return p.processSyncBatch(ctx, in)
	} else {
		// 大批量：使用异步处理
		return p.processAsyncBatch(ctx, in)
	}
}

// processSyncBatch 处理同步批量操作
func (p *BatchOperationProcessor) processSyncBatch(ctx context.Context, in *cmdb.CisBatchOperationReq) (*BatchOperationResponse, error) {
	p.logger.Infof("执行同步批量操作: CI数量=%d", len(in.CiIds))

	startTime := time.Now()

	// 创建批量任务数据
	taskData, err := p.createBatchTaskData(in)
	if err != nil {
		return nil, fmt.Errorf("创建任务数据失败: %v", err)
	}

	// 直接处理批量任务
	task := &BatchTask{
		ID:     generateTaskID(),
		Type:   fmt.Sprintf("sync_batch_%s", in.Operation),
		Status: "processing",
		Data:   taskData,
		Config: &BatchConfig{
			ChunkSize:      p.config.ChunkSize,
			MaxConcurrency: p.config.MaxConcurrency,
		},
		CreateTime: time.Now(),
	}

	result, err := p.batchProcessor.ProcessBatchTask(ctx, task)
	if err != nil {
		p.logger.Errorf("同步批量操作失败: %v", err)
		return &BatchOperationResponse{
			Success: false,
			Message: fmt.Sprintf("批量操作失败: %v", err),
		}, nil
	}

	processingTime := time.Since(startTime)
	p.logger.Infof("同步批量操作完成: 耗时=%v, 成功=%d, 失败=%d",
		processingTime, result.SuccessCount, result.FailedCount)

	return &BatchOperationResponse{
		Success:        result.Success,
		Message:        "批量操作完成",
		ProcessedCount: result.ProcessedCount,
		Summary: map[string]interface{}{
			"success_count":      result.SuccessCount,
			"failed_count":       result.FailedCount,
			"processing_time":    processingTime.String(),
			"throughput_per_sec": result.ThroughputPerSec,
		},
	}, nil
}

// processAsyncBatch 处理异步批量操作
func (p *BatchOperationProcessor) processAsyncBatch(ctx context.Context, in *cmdb.CisBatchOperationReq) (*BatchOperationResponse, error) {
	p.logger.Infof("执行异步批量操作: CI数量=%d", len(in.CiIds))

	// 创建异步任务数据
	taskData, err := p.createBatchTaskData(in)
	if err != nil {
		return nil, fmt.Errorf("创建异步任务数据失败: %v", err)
	}

	// 提交批量处理任务
	task, err := p.batchProcessor.SubmitBatchTask(ctx, taskData, fmt.Sprintf("async_batch_%s", in.Operation))
	if err != nil {
		return nil, fmt.Errorf("提交异步批量任务失败: %v", err)
	}

	p.logger.Infof("异步批量任务已提交: TaskID=%s", task.ID)

	return &BatchOperationResponse{
		Success: true,
		Message: fmt.Sprintf("批量操作已提交异步处理，任务ID: %s", task.ID),
		TaskID:  task.ID,
		Summary: map[string]interface{}{
			"task_id":        task.ID,
			"total_items":    len(in.CiIds),
			"estimated_time": task.Progress.EstimatedRemaining.String(),
		},
	}, nil
}

// createBatchTaskData 创建批量任务数据
func (p *BatchOperationProcessor) createBatchTaskData(in *cmdb.CisBatchOperationReq) (*input.ProcessingData, error) {
	// 将批量操作请求转换为处理数据格式
	assets := make([]*input.ProcessedAssetData, 0, len(in.CiIds))

	for _, ciID := range in.CiIds {
		// 创建RawAssetData
		rawAsset := &input.RawAssetData{
			ID:       fmt.Sprintf("ci_%d", ciID),
			CITypeID: 0, // 在实际操作中需要根据CI ID查询CI类型
			Source:   "batch_operation",
			Metadata: map[string]interface{}{
				"operation": in.Operation,
				"ci_id":     ciID,
				"params":    in.Params,
			},
		}

		// 创建ProcessedAssetData
		asset := &input.ProcessedAssetData{
			RawAssetData:   rawAsset,
			Status:         "pending",
			ProcessorChain: []string{},
		}

		assets = append(assets, asset)
	}

	processingData := &input.ProcessingData{
		Assets: assets,
		BatchInfo: &input.BatchInfo{
			ID:         fmt.Sprintf("batch_%d", time.Now().UnixNano()),
			Type:       fmt.Sprintf("batch_%s", in.Operation),
			Source:     "batch_operation",
			TotalCount: len(assets),
			CreateTime: time.Now(),
			CreatedBy:  "system",
			Status:     "processing",
		},
		Context: map[string]interface{}{
			"operation":    in.Operation,
			"original_req": in,
		},
		ErrorCount: 0,
	}

	return processingData, nil
}

// ProcessBatchOperationChunk 处理批量操作分块
func (p *BatchOperationProcessor) ProcessBatchOperationChunk(ctx context.Context, assets []*input.ProcessedAssetData) ([]*input.ProcessedAssetData, error) {
	p.logger.Infof("处理批量操作分块: 数量=%d", len(assets))

	processedAssets := make([]*input.ProcessedAssetData, 0, len(assets))

	for i, asset := range assets {
		// 从元数据中获取操作信息
		operation, ok := asset.Metadata["operation"].(string)
		if !ok {
			p.logger.Errorf("第%d条资产缺少操作类型信息", i+1)
			asset.Status = "failed"
			processedAssets = append(processedAssets, asset)
			continue
		}

		ciID, ok := asset.Metadata["ci_id"].(uint64)
		if !ok {
			p.logger.Errorf("第%d条资产缺少CI ID信息", i+1)
			asset.Status = "failed"
			processedAssets = append(processedAssets, asset)
			continue
		}

		// 处理单个操作
		err := p.processSingleOperation(ctx, operation, ciID, asset.Metadata["params"])
		if err != nil {
			p.logger.Errorf("第%d条资产操作失败: CI ID=%d, 操作=%s, 错误=%v", i+1, ciID, operation, err)
			asset.Status = "failed"
			asset.Metadata["error"] = err.Error()
		} else {
			asset.Status = "completed"
			p.logger.Debugf("第%d条资产操作成功: CI ID=%d, 操作=%s", i+1, ciID, operation)
		}

		processedAssets = append(processedAssets, asset)
	}

	return processedAssets, nil
}

// processSingleOperation 处理单个操作
func (p *BatchOperationProcessor) processSingleOperation(ctx context.Context, operation string, ciID uint64, params interface{}) error {
	// 构建单个操作请求
	req := &cmdb.CisBatchOperationReq{
		Operation: operation,
		CiIds:     []uint64{ciID},
	}

	// 如果有参数，序列化为字符串
	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("参数序列化失败: %v", err)
		}
		paramsStr := string(paramsBytes)
		req.Params = &paramsStr
	}

	// 这里需要调用实际的CIS操作逻辑
	// 但是为了避免循环引用，我们需要通过接口的方式
	// 暂时模拟操作成功
	p.logger.Infof("模拟执行操作: operation=%s, ci_id=%d", operation, ciID)

	// TODO: 实际的操作应该通过接口调用 CisBatchOperationLogic
	// 这里需要重新设计架构来避免循环引用

	return nil
}

// GetBatchTaskStatus 获取批量任务状态
func (p *BatchOperationProcessor) GetBatchTaskStatus(taskID string) (*BatchTask, error) {
	return p.batchProcessor.GetTaskStatus(taskID)
}

// CancelBatchTask 取消批量任务
func (p *BatchOperationProcessor) CancelBatchTask(ctx context.Context, taskID string) error {
	return p.batchProcessor.CancelTask(ctx, taskID)
}

// GetBatchConfig 获取批量操作配置
func (p *BatchOperationProcessor) GetBatchConfig() *BatchOperationConfig {
	return p.config
}

// UpdateBatchConfig 更新批量操作配置
func (p *BatchOperationProcessor) UpdateBatchConfig(config *BatchOperationConfig) {
	p.config = config
	p.logger.Infof("批量操作配置已更新: %+v", config)
}

// GetBatchStatistics 获取批量操作统计信息
func (p *BatchOperationProcessor) GetBatchStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"config":         p.config,
		"processor_info": p.batchProcessor.GetInfo(),
	}

	return stats
}
