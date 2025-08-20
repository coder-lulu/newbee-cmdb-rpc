package pipeline

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"gitee.com/link234/cmdb-rpc/internal/adapters/input"
	"gitee.com/link234/cmdb-rpc/internal/svc"
)

// PipelineResult 管道处理结果
type PipelineResult struct {
	ProcessedData  *input.ProcessingData `json:"processed_data"`
	ProcessedCount int                   `json:"processed_count"`
	TotalErrors    int                   `json:"total_errors"`
	ProcessTime    time.Duration         `json:"process_time"`
	Message        string                `json:"message"`
	// 详细统计
	ValidationStats *Stats `json:"validation_stats"`
	TransformStats  *Stats `json:"transform_stats"`
	PersistStats    *Stats `json:"persist_stats"`
}

// Stats 处理器统计信息
type Stats struct {
	TotalCount   int           `json:"total_count"`
	SuccessCount int           `json:"success_count"`
	FailedCount  int           `json:"failed_count"`
	SkippedCount int           `json:"skipped_count"`
	ProcessTime  time.Duration `json:"process_time"`
}

// DataPipeline 数据处理管道
type DataPipeline struct {
	svcCtx              *svc.ServiceContext
	logger              logx.Logger
	validationProcessor *ValidationProcessor
	transformProcessor  *TransformProcessor
	persistProcessor    *PersistProcessor
}

// NewDataPipeline 创建数据处理管道
func NewDataPipeline(svcCtx *svc.ServiceContext) *DataPipeline {
	return &DataPipeline{
		svcCtx:              svcCtx,
		logger:              logx.WithContext(context.Background()),
		validationProcessor: NewValidationProcessor(svcCtx),
		transformProcessor:  NewTransformProcessor(svcCtx),
		persistProcessor:    NewPersistProcessor(svcCtx),
	}
}

// Process 处理数据 - 完整的管道流程
func (dp *DataPipeline) Process(ctx context.Context, processingData *input.ProcessingData) (*PipelineResult, error) {
	startTime := time.Now()

	dp.logger.Infof("开始数据处理管道，输入资产数量: %d", len(processingData.Assets))

	result := &PipelineResult{
		ProcessedData:   processingData,
		ProcessedCount:  len(processingData.Assets),
		TotalErrors:     processingData.ErrorCount,
		ValidationStats: &Stats{},
		TransformStats:  &Stats{},
		PersistStats:    &Stats{},
	}

	// 第一阶段：数据验证
	validationStart := time.Now()
	validatedAssets, err := dp.validationProcessor.ProcessAssets(ctx, processingData.Assets)
	if err != nil {
		dp.logger.Errorf("数据验证阶段失败: %v", err)
		result.Message = "数据验证阶段失败"
		result.ProcessTime = time.Since(startTime)
		return result, err
	}

	// 统计验证结果
	result.ValidationStats = dp.calculateStats(validatedAssets, "validated", time.Since(validationStart))
	dp.logger.Infof("验证阶段完成: 成功=%d, 失败=%d",
		result.ValidationStats.SuccessCount, result.ValidationStats.FailedCount)

	// 第二阶段：数据转换
	transformStart := time.Now()
	transformedAssets, err := dp.transformProcessor.ProcessAssets(ctx, validatedAssets)
	if err != nil {
		dp.logger.Errorf("数据转换阶段失败: %v", err)
		result.Message = "数据转换阶段失败"
		result.ProcessTime = time.Since(startTime)
		return result, err
	}

	// 统计转换结果
	result.TransformStats = dp.calculateStats(transformedAssets, "transformed", time.Since(transformStart))
	dp.logger.Infof("转换阶段完成: 成功=%d, 失败=%d",
		result.TransformStats.SuccessCount, result.TransformStats.FailedCount)

	// 第三阶段：数据持久化
	persistStart := time.Now()
	persistedAssets, err := dp.persistProcessor.ProcessAssets(ctx, transformedAssets)
	if err != nil {
		dp.logger.Errorf("数据持久化阶段失败: %v", err)
		result.Message = "数据持久化阶段失败"
		result.ProcessTime = time.Since(startTime)
		return result, err
	}

	// 统计持久化结果
	result.PersistStats = dp.calculateStats(persistedAssets, "stored", time.Since(persistStart))
	dp.logger.Infof("持久化阶段完成: 成功=%d, 失败=%d",
		result.PersistStats.SuccessCount, result.PersistStats.FailedCount)

	// 更新最终结果
	processingData.Assets = persistedAssets
	result.ProcessedData = processingData
	result.ProcessTime = time.Since(startTime)
	result.Message = "数据处理管道完成"

	// 计算总体错误数
	result.TotalErrors = result.ValidationStats.FailedCount +
		result.TransformStats.FailedCount +
		result.PersistStats.FailedCount

	dp.logger.Infof("数据处理管道完成，总耗时: %v，最终存储成功: %d",
		result.ProcessTime, result.PersistStats.SuccessCount)

	return result, nil
}

// ProcessValidationOnly 仅处理数据验证
func (dp *DataPipeline) ProcessValidationOnly(ctx context.Context, processingData *input.ProcessingData) (*PipelineResult, error) {
	startTime := time.Now()

	dp.logger.Infof("开始数据验证，输入资产数量: %d", len(processingData.Assets))

	validatedAssets, err := dp.validationProcessor.ProcessAssets(ctx, processingData.Assets)
	if err != nil {
		return nil, err
	}

	processingData.Assets = validatedAssets
	result := &PipelineResult{
		ProcessedData:   processingData,
		ProcessedCount:  len(processingData.Assets),
		ProcessTime:     time.Since(startTime),
		Message:         "数据验证完成",
		ValidationStats: dp.calculateStats(validatedAssets, "validated", time.Since(startTime)),
	}

	return result, nil
}

// calculateStats 计算处理器统计信息
func (dp *DataPipeline) calculateStats(assets []*input.ProcessedAssetData, targetStatus string, processTime time.Duration) *Stats {
	stats := &Stats{
		TotalCount:  len(assets),
		ProcessTime: processTime,
	}

	for _, asset := range assets {
		if asset.Status == targetStatus {
			stats.SuccessCount++
		} else if asset.Status == "validation_failed" ||
			asset.Status == "transform_failed" ||
			asset.Status == "persist_failed" {
			stats.FailedCount++
		} else {
			stats.SkippedCount++
		}
	}

	return stats
}

// GetInfo 获取管道信息
func (dp *DataPipeline) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "DataPipeline",
		"version":     "v2.0.0",
		"description": "完整版数据处理管道，集成验证、转换、持久化三个阶段",
		"features": []string{
			"数据验证集成",
			"数据转换集成",
			"数据持久化集成",
			"错误统计",
			"阶段统计",
		},
		"processors": []map[string]interface{}{
			dp.validationProcessor.GetInfo(),
			dp.transformProcessor.GetInfo(),
			dp.persistProcessor.GetInfo(),
		},
	}
}
