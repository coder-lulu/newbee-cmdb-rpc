package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"gitee.com/link234/cmdb-rpc/internal/adapters/input"
	"gitee.com/link234/cmdb-rpc/internal/logic/cis"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
)

// ValidationProcessor 数据验证处理器 (集成现有验证逻辑)
type ValidationProcessor struct {
	svcCtx *svc.ServiceContext
	logger logx.Logger
}

// NewValidationProcessor 创建数据验证处理器
func NewValidationProcessor(svcCtx *svc.ServiceContext) *ValidationProcessor {
	return &ValidationProcessor{
		svcCtx: svcCtx,
		logger: logx.WithContext(context.Background()),
	}
}

// ProcessAssets 处理资产验证
func (p *ValidationProcessor) ProcessAssets(ctx context.Context, assets []*input.ProcessedAssetData) ([]*input.ProcessedAssetData, error) {
	startTime := time.Now()
	validatedAssets := make([]*input.ProcessedAssetData, 0)
	errorCount := 0

	p.logger.Infof("开始数据验证，待验证资产数量: %d", len(assets))

	for i, asset := range assets {
		// 构建验证请求 (复用现有逻辑)
		validateReq := &cmdb.CisAttributeValidateReq{
			TypeId:     asset.CITypeID,
			Attributes: p.convertToValidationFormat(asset.Attributes),
		}

		// 调用现有验证逻辑
		validateLogic := cis.NewValidateCisAttributesLogic(ctx, p.svcCtx)
		result, err := validateLogic.ValidateCisAttributes(validateReq)
		if err != nil {
			p.logger.Errorf("第%d条资产验证失败: %v", i+1, err)
			errorCount++

			// 创建验证失败的ProcessedAssetData
			asset.ValidationResult = &input.ValidationResult{
				Valid:       false,
				ErrorCount:  1,
				ProcessTime: time.Since(startTime),
				Summary:     map[string]interface{}{"error": err.Error()},
			}
			asset.Status = "validation_failed"
			asset.ProcessorChain = append(asset.ProcessorChain, "validation_processor")

			validatedAssets = append(validatedAssets, asset)
			continue
		}

		if !result.Valid {
			p.logger.Errorf("第%d条资产数据验证失败: %v", i+1, result.Errors)
			errorCount++

			// 转换验证结果
			validationResult := p.convertValidationResult(result)
			asset.ValidationResult = validationResult
			asset.Status = "validation_failed"
			asset.ProcessorChain = append(asset.ProcessorChain, "validation_processor")

			validatedAssets = append(validatedAssets, asset)
			continue
		}

		// 验证通过，更新资产状态
		validationResult := p.convertValidationResult(result)
		asset.ValidationResult = validationResult
		asset.Status = "validated"
		asset.ProcessorChain = append(asset.ProcessorChain, "validation_processor")

		validatedAssets = append(validatedAssets, asset)
		p.logger.Debugf("第%d条资产验证通过", i+1)
	}

	processingTime := time.Since(startTime)
	successCount := len(validatedAssets) - errorCount

	p.logger.Infof("数据验证完成: 总数=%d, 成功=%d, 失败=%d, 耗时=%v",
		len(assets), successCount, errorCount, processingTime)

	return validatedAssets, nil
}

// convertToValidationFormat 将属性转换为验证格式
func (p *ValidationProcessor) convertToValidationFormat(attributes map[string]interface{}) []*cmdb.CiAttributeValue {
	var result []*cmdb.CiAttributeValue
	for key, value := range attributes {
		result = append(result, &cmdb.CiAttributeValue{
			AttrName: key,
			Value:    fmt.Sprintf("%v", value),
		})
	}
	return result
}

// convertValidationResult 转换验证结果格式
func (p *ValidationProcessor) convertValidationResult(result *cmdb.CisAttributeValidateResp) *input.ValidationResult {
	validationResult := &input.ValidationResult{
		Valid:       result.Valid,
		ValidCount:  0, // 成功的属性数量 (从API返回中获取)
		ErrorCount:  len(result.Errors),
		ProcessTime: time.Duration(0), // 这里可以传入实际的处理时间
		Summary:     make(map[string]interface{}),
	}

	// 转换错误信息
	if len(result.Errors) > 0 {
		validationErrors := make([]*input.ValidationError, 0)
		for _, err := range result.Errors {
			validationError := &input.ValidationError{
				AssetID:       "", // 可以从上下文获取
				LineNumber:    0,  // 可以从原始数据获取
				ColumnName:    err.AttrName,
				ErrorType:     err.ErrorType,
				ErrorMsg:      err.Message,
				ExpectedValue: "", // 可以从验证规则获取
				ActualValue:   "", // CiAttributeError没有Value字段
			}
			validationErrors = append(validationErrors, validationError)
		}
		validationResult.Errors = validationErrors
	}

	// 设置汇总信息
	validationResult.Summary["total_attributes"] = len(result.Errors) + validationResult.ValidCount
	validationResult.Summary["valid_attributes"] = validationResult.ValidCount
	validationResult.Summary["invalid_attributes"] = len(result.Errors)

	return validationResult
}

// GetInfo 获取处理器信息
func (p *ValidationProcessor) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "ValidationProcessor",
		"version":     "v1.0.0",
		"description": "数据验证处理器，集成现有ValidateCisAttributesLogic",
		"features": []string{
			"CI类型属性验证",
			"验证错误详细信息",
			"验证结果统计",
			"集成现有验证逻辑",
		},
	}
}
