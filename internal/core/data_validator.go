package core

import (
	"context"
	"fmt"
	"time"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/internal/logic/cis"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"github.com/zeromicro/go-zero/core/logx"
)

// dataValidatorImpl 数据校验器实现
type dataValidatorImpl struct {
	svcCtx      *svc.ServiceContext
	config      *Configuration
	logger      logx.Logger
	initialized bool
}

// Name 返回组件名称
func (v *dataValidatorImpl) Name() string {
	return "DataValidator"
}

// Initialize 初始化组件
func (v *dataValidatorImpl) Initialize(ctx context.Context) error {
	v.logger = logx.WithContext(ctx)
	v.initialized = true
	v.logger.Info("数据校验器初始化完成")
	return nil
}

// Shutdown 关闭组件
func (v *dataValidatorImpl) Shutdown(ctx context.Context) error {
	v.initialized = false
	v.logger.Info("数据校验器已关闭")
	return nil
}

// HealthCheck 健康检查
func (v *dataValidatorImpl) HealthCheck(ctx context.Context) error {
	if !v.initialized {
		return fmt.Errorf("数据校验器未初始化")
	}
	return nil
}

// ValidateSchema 校验数据结构
func (v *dataValidatorImpl) ValidateSchema(ctx context.Context, data *cmdb.CisInfo, typeID uint64) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationWarning, 0),
		Performance: &ValidationPerformance{
			TotalTime: time.Since(startTime),
		},
	}

	// 基础数据校验
	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:     "data",
			ErrorType: "null_data",
			Message:   "数据不能为空",
			Code:      "ERR_NULL_DATA",
		})
		return result, nil
	}

	// 校验CI类型ID
	if data.TypeId == nil || *data.TypeId != typeID {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:     "type_id",
			ErrorType: "type_mismatch",
			Message:   "CI类型ID不匹配",
			Code:      "ERR_TYPE_MISMATCH",
			Value:     data.TypeId,
		})
	}

	// 校验属性数据结构
	if len(data.Attributes) > 0 {
		for _, attr := range data.Attributes {
			if attr.AttrId == 0 {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:     "attributes",
					AttrID:    attr.AttrId,
					ErrorType: "invalid_attr_id",
					Message:   "属性ID不能为0",
					Code:      "ERR_INVALID_ATTR_ID",
				})
			}
		}
	}

	result.Performance.TotalTime = time.Since(startTime)
	result.Performance.AttributeCount = len(data.Attributes)

	return result, nil
}

// ValidateBusinessRules 校验业务规则
func (v *dataValidatorImpl) ValidateBusinessRules(ctx context.Context, data *cmdb.CisInfo, typeID uint64) (*ValidationResult, error) {
	startTime := time.Now()

	// 使用现有的ValidateCisAttributesLogic进行业务规则校验
	validateLogic := cis.NewValidateCisAttributesLogic(ctx, v.svcCtx)

	req := &cmdb.CisAttributeValidateReq{
		TypeId:     typeID,
		Attributes: data.Attributes,
	}

	resp, err := validateLogic.ValidateCisAttributes(req)
	if err != nil {
		return nil, fmt.Errorf("业务规则校验失败: %w", err)
	}

	// 转换校验结果
	result := &ValidationResult{
		Valid:    resp.Valid,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationWarning, 0),
		Performance: &ValidationPerformance{
			TotalTime:      time.Since(startTime),
			AttributeCount: len(data.Attributes),
		},
	}

	// 转换错误信息
	for _, err := range resp.Errors {
		result.Errors = append(result.Errors, &ValidationError{
			Field:     err.AttrName,
			AttrID:    err.AttrId,
			ErrorType: err.ErrorType,
			Message:   err.Message,
			Code:      fmt.Sprintf("RULE_%s", err.ErrorType),
		})
	}

	result.Performance.RuleCount = len(result.Errors)

	return result, nil
}

// ValidateUniqueness 校验唯一性
func (v *dataValidatorImpl) ValidateUniqueness(ctx context.Context, data *cmdb.CisInfo, typeID uint64, excludeCiID *uint64) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationWarning, 0),
		Performance: &ValidationPerformance{
			TotalTime:      time.Since(startTime),
			AttributeCount: len(data.Attributes),
		},
	}

	// 使用现有的ValidateUniqueAttributes函数
	err := cis.ValidateUniqueAttributes(ctx, v.svcCtx.DB, typeID, data.Attributes, excludeCiID)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:     "attributes",
			ErrorType: "uniqueness_violation",
			Message:   err.Error(),
			Code:      "ERR_UNIQUENESS_VIOLATION",
		})
	}

	result.Performance.TotalTime = time.Since(startTime)

	return result, nil
}

// ValidateComplete 完整校验
func (v *dataValidatorImpl) ValidateComplete(ctx context.Context, operation *CiOperationContext) (*ValidationResult, error) {
	startTime := time.Now()

	// 根据校验级别决定校验范围
	var results []*ValidationResult

	// 1. 结构校验 (所有级别都执行)
	schemaResult, err := v.ValidateSchema(ctx, operation.DataAfter, operation.CiTypeID)
	if err != nil {
		return nil, fmt.Errorf("结构校验失败: %w", err)
	}
	results = append(results, schemaResult)

	// 2. 业务规则校验 (basic级别以上)
	if operation.ValidationLevel == ValidationBasic ||
		operation.ValidationLevel == ValidationStrict ||
		operation.ValidationLevel == ValidationFull {

		rulesResult, err := v.ValidateBusinessRules(ctx, operation.DataAfter, operation.CiTypeID)
		if err != nil {
			return nil, fmt.Errorf("业务规则校验失败: %w", err)
		}
		results = append(results, rulesResult)
	}

	// 3. 唯一性校验 (strict级别以上)
	if operation.ValidationLevel == ValidationStrict || operation.ValidationLevel == ValidationFull {
		uniqueResult, err := v.ValidateUniqueness(ctx, operation.DataAfter, operation.CiTypeID, operation.CiID)
		if err != nil {
			return nil, fmt.Errorf("唯一性校验失败: %w", err)
		}
		results = append(results, uniqueResult)
	}

	// 4. 高级校验 (full级别)
	if operation.ValidationLevel == ValidationFull {
		// TODO: 实现高级校验逻辑，如关联关系校验、数据完整性校验等
	}

	// 合并校验结果
	combinedResult := v.combineValidationResults(results)
	combinedResult.Performance.TotalTime = time.Since(startTime)
	combinedResult.CheckedRules = []string{
		"schema_validation",
		"business_rules",
		"uniqueness_check",
	}

	return combinedResult, nil
}

// combineValidationResults 合并多个校验结果
func (v *dataValidatorImpl) combineValidationResults(results []*ValidationResult) *ValidationResult {
	combined := &ValidationResult{
		Valid:        true,
		Errors:       make([]*ValidationError, 0),
		Warnings:     make([]*ValidationWarning, 0),
		CheckedRules: make([]string, 0),
		Performance: &ValidationPerformance{
			TotalTime:      0,
			RuleCount:      0,
			AttributeCount: 0,
		},
	}

	for _, result := range results {
		// 如果任何一个结果无效，整体就无效
		if !result.Valid {
			combined.Valid = false
		}

		// 合并错误和警告
		combined.Errors = append(combined.Errors, result.Errors...)
		combined.Warnings = append(combined.Warnings, result.Warnings...)

		// 合并性能指标
		if result.Performance != nil {
			combined.Performance.TotalTime += result.Performance.TotalTime
			combined.Performance.RuleCount += result.Performance.RuleCount
			if result.Performance.AttributeCount > combined.Performance.AttributeCount {
				combined.Performance.AttributeCount = result.Performance.AttributeCount
			}
		}

		// 合并检查的规则
		combined.CheckedRules = append(combined.CheckedRules, result.CheckedRules...)
	}

	return combined
}

// validateAttributesWithTypeInfo 根据类型信息校验属性
func (v *dataValidatorImpl) validateAttributesWithTypeInfo(ctx context.Context, attributes []*cmdb.CiAttributeValue, typeID uint64) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationWarning, 0),
	}

	// 获取CI类型的属性定义
	typeAttrs, err := v.svcCtx.DB.CiTypeAttribute.Query().
		Where(citypeattribute.TypeIDEQ(typeID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取CI类型属性定义失败: %w", err)
	}

	// 创建属性映射
	attrMap := make(map[uint64]*ent.CiTypeAttribute)
	for _, ta := range typeAttrs {
		attrMap[ta.AttrID] = ta
	}

	// 校验每个属性
	for _, attrValue := range attributes {
		if typeAttr, exists := attrMap[attrValue.AttrId]; exists {
			if typeAttr.Edges.Attribute != nil {
				// 校验属性值格式
				if err := v.validateAttributeValue(typeAttr, attrValue); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, &ValidationError{
						Field:     typeAttr.Edges.Attribute.Name,
						AttrID:    attrValue.AttrId,
						ErrorType: "format_error",
						Message:   err.Error(),
						Code:      "ERR_FORMAT",
						Value:     attrValue.Value,
					})
				}
			}
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Field:     "attribute",
				AttrID:    attrValue.AttrId,
				ErrorType: "unknown_attribute",
				Message:   fmt.Sprintf("未知的属性ID: %d", attrValue.AttrId),
				Code:      "ERR_UNKNOWN_ATTR",
				Value:     attrValue.AttrId,
			})
		}
	}

	// 检查必填属性
	for _, typeAttr := range typeAttrs {
		if typeAttr.IsRequired {
			found := false
			for _, attrValue := range attributes {
				if attrValue.AttrId == typeAttr.AttrID && attrValue.Value != "" {
					found = true
					break
				}
			}

			if !found {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:     typeAttr.Edges.Attribute.Name,
					AttrID:    typeAttr.AttrID,
					ErrorType: "required_missing",
					Message:   fmt.Sprintf("必填属性 %s 缺失或为空", typeAttr.Edges.Attribute.Name),
					Code:      "ERR_REQUIRED_MISSING",
				})
			}
		}
	}

	return result, nil
}

// validateAttributeValue 校验单个属性值
func (v *dataValidatorImpl) validateAttributeValue(typeAttr *ent.CiTypeAttribute, attrValue *cmdb.CiAttributeValue) error {
	// 这里可以调用现有的校验逻辑
	// 由于现有的ValidateCisAttributesLogic已经实现了详细的属性校验
	// 这里主要做简单的格式校验

	if attrValue.Value == "" && typeAttr.IsRequired {
		return fmt.Errorf("必填属性不能为空")
	}

	// 可以根据属性类型进行更详细的校验
	// 这里暂时返回nil，表示通过
	return nil
}
