package engine

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attributemappingrule"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// DataTransformer 数据转换器
type DataTransformer struct {
	db     *ent.Client
	logger logx.Logger
}

// NewDataTransformer 创建新的数据转换器
func NewDataTransformer(db *ent.Client) *DataTransformer {
	return &DataTransformer{
		db:     db,
		logger: logx.WithContext(context.Background()),
	}
}

// Transform 转换原始数据为CI数据
func (t *DataTransformer) Transform(ctx context.Context, config *ent.CiTypeDiscoveryConfig, rawData []map[string]interface{}) ([]types.TransformedCIData, error) {
	var transformedData []types.TransformedCIData

	// 加载属性映射规则
	mappingRules, err := t.loadMappingRules(ctx, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mapping rules: %w", err)
	}

	// 逐条处理原始数据
	for i, rawRecord := range rawData {
		transformed, err := t.transformRecord(ctx, config, rawRecord, mappingRules)
		if err != nil {
			t.logger.Errorf("Failed to transform record %d: %v", i, err)
			continue
		}

		if transformed != nil {
			transformedData = append(transformedData, *transformed)
		}
	}

	return transformedData, nil
}

// loadMappingRules 加载属性映射规则
func (t *DataTransformer) loadMappingRules(ctx context.Context, configID uint64) ([]*ent.AttributeMappingRule, error) {
	return t.db.AttributeMappingRule.Query().
		Where(
			attributemappingrule.DiscoveryConfigIDEQ(configID),
			attributemappingrule.EnabledEQ(true),
		).
		All(ctx)
}

// transformRecord 转换单条记录
func (t *DataTransformer) transformRecord(ctx context.Context, config *ent.CiTypeDiscoveryConfig, rawRecord map[string]interface{}, mappingRules []*ent.AttributeMappingRule) (*types.TransformedCIData, error) {
	// 生成唯一标识
	uniqueKey, err := t.generateUniqueKey(config, rawRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique key: %w", err)
	}

	// 检查是否已存在该CI
	existingCI, isUpdate, err := t.findExistingCI(ctx, config.CiTypeID, uniqueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing CI: %w", err)
	}

	// 转换属性
	attributes := make(map[string]interface{})
	metadata := make(map[string]interface{})

	for _, rule := range mappingRules {
		value, err := t.extractValue(rawRecord, rule)
		if err != nil {
			t.logger.Errorf("Failed to extract value for rule %s: %v", rule.TargetAttribute, err)
			continue
		}

		// 数据验证
		if err := t.validateValue(rule, value); err != nil {
			t.logger.Errorf("Validation failed for attribute %s: %v", rule.TargetAttribute, err)
			continue
		}

		// 数据转换
		transformedValue, err := t.transformValue(rule, value)
		if err != nil {
			t.logger.Errorf("Transformation failed for attribute %s: %v", rule.TargetAttribute, err)
			continue
		}

		attributes[rule.TargetAttribute] = transformedValue
	}

	// 添加系统元数据
	metadata["source_system"] = config.DiscoveryType
	metadata["discovery_time"] = time.Now().Unix()
	metadata["config_id"] = config.ID

	transformed := &types.TransformedCIData{
		CITypeID:     config.CiTypeID,
		SourceID:     fmt.Sprintf("%v", rawRecord["id"]), // 假设原始数据有id字段
		UniqueKey:    uniqueKey,
		Attributes:   attributes,
		Metadata:     metadata,
		IsUpdate:     isUpdate,
		ExistingCIID: existingCI,
	}

	return transformed, nil
}

// generateUniqueKey 生成唯一标识
func (t *DataTransformer) generateUniqueKey(config *ent.CiTypeDiscoveryConfig, rawRecord map[string]interface{}) (string, error) {
	// 暂时使用ID字段作为唯一键，实际应根据配置来确定
	// TODO: 根据配置的unique_fields来生成
	uniqueFields := []string{"id", "name"} // 默认字段
	
	if len(uniqueFields) == 0 {
		return "", fmt.Errorf("no unique fields configured")
	}

	var keyParts []string
	for _, field := range uniqueFields {
		value, exists := rawRecord[field]
		if !exists {
			continue // 跳过不存在的字段
		}
		keyParts = append(keyParts, fmt.Sprintf("%v", value))
	}

	if len(keyParts) == 0 {
		return "", fmt.Errorf("no unique field values found in raw data")
	}

	return strings.Join(keyParts, "|"), nil
}

// findExistingCI 查找已存在的CI
func (t *DataTransformer) findExistingCI(ctx context.Context, ciTypeID uint64, uniqueKey string) (*uint64, bool, error) {
	// 这里需要根据实际的CI存储结构来实现
	// 暂时返回不存在的情况
	return nil, false, nil
}

// extractValue 从原始数据中提取值
func (t *DataTransformer) extractValue(rawRecord map[string]interface{}, rule *ent.AttributeMappingRule) (interface{}, error) {
	// 支持路径访问，如 "user.name" 或 "config.database.host"
	if rule.SourcePath != "" {
		return t.extractValueByPath(rawRecord, rule.SourcePath)
	}

	// 直接字段访问
	value, exists := rawRecord[rule.SourceField]
	if !exists {
		// 使用默认值
		if rule.DefaultValue != "" {
			return rule.DefaultValue, nil
		}
		if rule.IsRequired {
			return nil, fmt.Errorf("required field %s not found", rule.SourceField)
		}
		return nil, nil
	}

	return value, nil
}

// extractValueByPath 按路径提取值
func (t *DataTransformer) extractValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("path %s not found", path)
		}

		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			// 到达叶子节点
			return value, nil
		}
	}

	return current, nil
}

// validateValue 验证值
func (t *DataTransformer) validateValue(rule *ent.AttributeMappingRule, value interface{}) error {
	if value == nil {
		if rule.IsRequired {
			return fmt.Errorf("required value is nil")
		}
		return nil
	}

	// 正则验证
	if rule.ValidationRegex != "" {
		// 实现正则验证逻辑
	}

	// 其他验证规则
	if rule.ValidationRules != nil {
		// 实现更复杂的验证逻辑
	}

	return nil
}

// transformValue 转换值
func (t *DataTransformer) transformValue(rule *ent.AttributeMappingRule, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch rule.TransformType {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "int":
		return t.toInt(value)
	case "float":
		return t.toFloat(value)
	case "bool":
		return t.toBool(value)
	case "datetime":
		return t.toDateTime(value)
	case "json":
		return value, nil // 保持原始JSON结构
	case "uppercase":
		return strings.ToUpper(fmt.Sprintf("%v", value)), nil
	case "lowercase":
		return strings.ToLower(fmt.Sprintf("%v", value)), nil
	case "trim":
		return strings.TrimSpace(fmt.Sprintf("%v", value)), nil
	default:
		return value, nil // 不转换
	}
}

// 类型转换辅助方法
func (t *DataTransformer) toInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (t *DataTransformer) toFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", value)
	}
}

func (t *DataTransformer) toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func (t *DataTransformer) toDateTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse datetime: %s", v)
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to datetime", value)
	}
}