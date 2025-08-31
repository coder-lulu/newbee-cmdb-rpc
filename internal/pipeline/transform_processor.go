package pipeline

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/consts"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// TransformProcessor 数据转换处理器
type TransformProcessor struct {
	svcCtx *svc.ServiceContext
	logger logx.Logger
}

// NewTransformProcessor 创建数据转换处理器
func NewTransformProcessor(svcCtx *svc.ServiceContext) *TransformProcessor {
	return &TransformProcessor{
		svcCtx: svcCtx,
		logger: logx.WithContext(context.Background()),
	}
}

// ProcessAssets 处理资产转换
func (p *TransformProcessor) ProcessAssets(ctx context.Context, assets []*input.ProcessedAssetData) ([]*input.ProcessedAssetData, error) {
	startTime := time.Now()
	transformedAssets := make([]*input.ProcessedAssetData, 0)
	errorCount := 0

	p.logger.Infof("开始数据转换，待转换资产数量: %d", len(assets))

	for i, asset := range assets {
		// 只处理验证通过的资产
		if asset.Status != "validated" {
			p.logger.Infof("第%d条资产状态非validated，跳过转换: %s", i+1, asset.Status)
			transformedAssets = append(transformedAssets, asset)
			continue
		}

		// 转换资产数据
		transformedAsset, err := p.transformSingleAsset(ctx, asset)
		if err != nil {
			p.logger.Errorf("第%d条资产转换失败: %v", i+1, err)
			errorCount++

			// 设置转换失败状态
			asset.Status = "transform_failed"
			asset.TransformResult = &input.TransformResult{
				Success:     false,
				ErrorCount:  1,
				ProcessTime: time.Since(startTime),
				Summary:     map[string]interface{}{"error": err.Error()},
			}
			asset.ProcessorChain = append(asset.ProcessorChain, "transform_processor")

			transformedAssets = append(transformedAssets, asset)
			continue
		}

		transformedAssets = append(transformedAssets, transformedAsset)
		p.logger.Debugf("第%d条资产转换成功", i+1)
	}

	processingTime := time.Since(startTime)
	successCount := len(transformedAssets) - errorCount

	p.logger.Infof("数据转换完成: 总数=%d, 成功=%d, 失败=%d, 耗时=%v",
		len(assets), successCount, errorCount, processingTime)

	return transformedAssets, nil
}

// transformSingleAsset 转换单个资产
func (p *TransformProcessor) transformSingleAsset(ctx context.Context, asset *input.ProcessedAssetData) (*input.ProcessedAssetData, error) {
	startTime := time.Now()

	// 创建CIS信息对象
	cisInfo := &cmdb.CisInfo{
		TypeId:    &asset.CITypeID,
		CreatedBy: &asset.Source,
		Tags:      asset.Tags,
	}

	// 转换属性数据
	transformedAttributes, err := p.transformAttributes(asset.Attributes, asset.CITypeID)
	if err != nil {
		return nil, fmt.Errorf("属性转换失败: %v", err)
	}

	cisInfo.Attributes = transformedAttributes

	// 转换元数据
	if asset.Metadata != nil {
		metadata := p.transformMetadata(asset.Metadata)
		cisInfo.Metadata = metadata
	}

	// 创建转换结果
	transformResult := &input.TransformResult{
		Success:         true,
		TransformedCIS:  cisInfo,
		AttributesCount: len(transformedAttributes),
		ErrorCount:      0,
		ProcessTime:     time.Since(startTime),
		Summary: map[string]interface{}{
			"ci_type_id":       asset.CITypeID,
			"attributes_count": len(transformedAttributes),
			"tags_count":       len(asset.Tags),
			"has_metadata":     asset.Metadata != nil,
		},
	}

	// 更新资产状态
	asset.TransformResult = transformResult
	asset.Status = "transformed"
	asset.ProcessorChain = append(asset.ProcessorChain, "transform_processor")

	return asset, nil
}

// transformAttributes 转换属性数据
func (p *TransformProcessor) transformAttributes(attributes map[string]interface{}, ciTypeID uint64) ([]*cmdb.CiAttributeValue, error) {
	var result []*cmdb.CiAttributeValue

	// 获取CI类型的属性定义 (简化处理，实际应该从数据库获取)
	for attrName, attrValue := range attributes {
		// 推断属性类型
		valueType := p.inferValueType(attrValue)
		valueStr := p.convertToString(attrValue, valueType)

		ciAttrValue := &cmdb.CiAttributeValue{
			AttrName:  attrName,
			ValueType: valueType,
			Value:     valueStr,
		}

		result = append(result, ciAttrValue)
	}

	return result, nil
}

// inferValueType 推断属性值类型
func (p *TransformProcessor) inferValueType(value interface{}) string {
	if value == nil {
		return consts.ValueTypeShortText
	}

	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return consts.ValueTypeInt
	case float32, float64:
		return consts.ValueTypeFloat
	case bool:
		return consts.ValueTypeBool
	case string:
		if len(v) > 128 {
			return consts.ValueTypeLongText
		}
		return consts.ValueTypeShortText
	case time.Time:
		return consts.ValueTypeDateTime
	default:
		// JSON 或其他复杂类型
		return consts.ValueTypeJSON
	}
}

// convertToString 将值转换为字符串
func (p *TransformProcessor) convertToString(value interface{}, valueType string) string {
	if value == nil {
		return ""
	}

	switch valueType {
	case consts.ValueTypeInt:
		switch v := value.(type) {
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		case uint64:
			return strconv.FormatUint(v, 10)
		default:
			return fmt.Sprintf("%d", v)
		}
	case consts.ValueTypeFloat:
		switch v := value.(type) {
		case float32:
			return strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		default:
			return fmt.Sprintf("%f", v)
		}
	case consts.ValueTypeBool:
		if b, ok := value.(bool); ok {
			return strconv.FormatBool(b)
		}
		return "false"
	case consts.ValueTypeDateTime:
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
		if str, ok := value.(string); ok {
			return str
		}
		return ""
	case consts.ValueTypeJSON:
		// 简化处理，实际应该使用JSON序列化
		return fmt.Sprintf("%v", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// transformMetadata 转换元数据
func (p *TransformProcessor) transformMetadata(metadata map[string]interface{}) []*cmdb.CisMetadata {
	var result []*cmdb.CisMetadata

	for key, value := range metadata {
		valueStr := fmt.Sprintf("%v", value)
		cisMetadata := &cmdb.CisMetadata{
			Key:   &key,
			Value: &valueStr,
		}
		result = append(result, cisMetadata)
	}

	return result
}

// GetInfo 获取处理器信息
func (p *TransformProcessor) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "TransformProcessor",
		"version":     "v1.0.0",
		"description": "数据转换处理器，将验证后的资产数据转换为CIS业务模型格式",
		"features": []string{
			"属性类型推断",
			"数据格式转换",
			"元数据转换",
			"CIS模型适配",
		},
	}
}
