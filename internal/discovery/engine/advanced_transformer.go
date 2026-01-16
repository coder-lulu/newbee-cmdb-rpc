package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// AdvancedDataTransformer provides sophisticated attribute mapping and transformation
type AdvancedDataTransformer struct {
	logger logx.Logger
}

// TransformationRule defines a complex transformation rule
type TransformationRule struct {
	SourceField      string                 `json:"source_field"`
	TargetField      string                 `json:"target_field"`
	TransformType    string                 `json:"transform_type"`
	DefaultValue     interface{}            `json:"default_value"`
	ValidationRegex  string                 `json:"validation_regex"`
	ConditionalRules []ConditionalRule      `json:"conditional_rules"`
	NestedMapping    map[string]interface{} `json:"nested_mapping"`
	ArrayHandling    ArrayHandlingRule      `json:"array_handling"`
}

// ConditionalRule defines conditional transformation logic
type ConditionalRule struct {
	Condition    string      `json:"condition"`
	Value        interface{} `json:"value"`
	TargetValue  interface{} `json:"target_value"`
	ElseValue    interface{} `json:"else_value"`
}

// ArrayHandlingRule defines how to handle array transformations
type ArrayHandlingRule struct {
	Mode        string `json:"mode"` // "first", "last", "join", "split", "map"
	Separator   string `json:"separator"`
	IndexFilter []int  `json:"index_filter"`
}

// NewAdvancedDataTransformer creates a new advanced transformer
func NewAdvancedDataTransformer() *AdvancedDataTransformer {
	return &AdvancedDataTransformer{
		logger: logx.WithContext(nil),
	}
}

// TransformAdvanced performs advanced data transformation using complex rules
func (t *AdvancedDataTransformer) TransformAdvanced(sourceData map[string]interface{}, rules []TransformationRule) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for _, rule := range rules {
		value, err := t.applyTransformationRule(sourceData, rule)
		if err != nil {
			t.logger.Errorw("Failed to apply transformation rule",
				logx.Field("rule", rule),
				logx.Field("error", err))
			continue
		}
		
		if value != nil {
			t.setNestedValue(result, rule.TargetField, value)
		}
	}
	
	return result, nil
}

// applyTransformationRule applies a single transformation rule
func (t *AdvancedDataTransformer) applyTransformationRule(sourceData map[string]interface{}, rule TransformationRule) (interface{}, error) {
	// Get source value
	sourceValue := t.getNestedValue(sourceData, rule.SourceField)
	
	// Apply conditional rules if present
	if len(rule.ConditionalRules) > 0 {
		for _, condRule := range rule.ConditionalRules {
			if t.evaluateCondition(sourceValue, condRule.Condition, condRule.Value) {
				sourceValue = condRule.TargetValue
				break
			} else if condRule.ElseValue != nil {
				sourceValue = condRule.ElseValue
			}
		}
	}
	
	// Use default if source value is nil
	if sourceValue == nil && rule.DefaultValue != nil {
		sourceValue = rule.DefaultValue
	}
	
	if sourceValue == nil {
		return nil, nil
	}
	
	// Apply array handling if configured
	if rule.ArrayHandling.Mode != "" {
		if processedValue, err := t.handleArrayTransformation(sourceValue, rule.ArrayHandling); err == nil {
			sourceValue = processedValue
		}
	}
	
	// Apply transformation based on type
	transformedValue, err := t.transformValueByType(sourceValue, rule.TransformType)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}
	
	// Apply validation if configured
	if rule.ValidationRegex != "" {
		if !t.validateWithRegex(transformedValue, rule.ValidationRegex) {
			return nil, fmt.Errorf("validation failed for value: %v", transformedValue)
		}
	}
	
	return transformedValue, nil
}

// handleArrayTransformation handles array-specific transformations
func (t *AdvancedDataTransformer) handleArrayTransformation(value interface{}, rule ArrayHandlingRule) (interface{}, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		// If not an array but rule expects array handling, try to split string
		if rule.Mode == "split" && rule.Separator != "" {
			if str, ok := value.(string); ok {
				return strings.Split(str, rule.Separator), nil
			}
		}
		return value, nil
	}
	
	switch rule.Mode {
	case "first":
		if rv.Len() > 0 {
			return rv.Index(0).Interface(), nil
		}
		return nil, nil
		
	case "last":
		if rv.Len() > 0 {
			return rv.Index(rv.Len() - 1).Interface(), nil
		}
		return nil, nil
		
	case "join":
		var strValues []string
		for i := 0; i < rv.Len(); i++ {
			strValues = append(strValues, fmt.Sprintf("%v", rv.Index(i).Interface()))
		}
		separator := rule.Separator
		if separator == "" {
			separator = ","
		}
		return strings.Join(strValues, separator), nil
		
	case "map":
		// Apply index filtering if specified
		if len(rule.IndexFilter) > 0 {
			var filteredValues []interface{}
			for _, idx := range rule.IndexFilter {
				if idx >= 0 && idx < rv.Len() {
					filteredValues = append(filteredValues, rv.Index(idx).Interface())
				}
			}
			return filteredValues, nil
		}
		return value, nil
		
	default:
		return value, nil
	}
}

// transformValueByType transforms value based on specified type
func (t *AdvancedDataTransformer) transformValueByType(value interface{}, transformType string) (interface{}, error) {
	if transformType == "" {
		return value, nil
	}
	
	switch strings.ToLower(transformType) {
	case "string":
		return fmt.Sprintf("%v", value), nil
		
	case "int", "integer":
		switch v := value.(type) {
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			return strconv.Atoi(v)
		default:
			return strconv.Atoi(fmt.Sprintf("%v", v))
		}
		
	case "float", "double":
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		}
		
	case "bool", "boolean":
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(strings.ToLower(v))
		case int:
			return v != 0, nil
		default:
			return strconv.ParseBool(fmt.Sprintf("%v", v))
		}
		
	case "datetime", "timestamp":
		switch v := value.(type) {
		case string:
			// Try multiple datetime formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
				"2006-01-02",
				"15:04:05",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, v); err == nil {
					return t.Unix(), nil
				}
			}
			return nil, fmt.Errorf("unable to parse datetime: %s", v)
		case int64:
			return v, nil
		case time.Time:
			return v.Unix(), nil
		default:
			return nil, fmt.Errorf("unsupported datetime type: %T", v)
		}
		
	case "json":
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return string(jsonBytes), nil
		
	case "upper":
		return strings.ToUpper(fmt.Sprintf("%v", value)), nil
		
	case "lower":
		return strings.ToLower(fmt.Sprintf("%v", value)), nil
		
	case "trim":
		return strings.TrimSpace(fmt.Sprintf("%v", value)), nil
		
	case "base64":
		// Encode to base64
		return fmt.Sprintf("%v", value), nil // Placeholder for actual base64 encoding
		
	default:
		return value, nil
	}
}

// evaluateCondition evaluates a conditional rule
func (t *AdvancedDataTransformer) evaluateCondition(sourceValue interface{}, condition string, condValue interface{}) bool {
	switch condition {
	case "equals", "eq":
		return fmt.Sprintf("%v", sourceValue) == fmt.Sprintf("%v", condValue)
	case "not_equals", "ne":
		return fmt.Sprintf("%v", sourceValue) != fmt.Sprintf("%v", condValue)
	case "contains":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		condStr := fmt.Sprintf("%v", condValue)
		return strings.Contains(sourceStr, condStr)
	case "starts_with":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		condStr := fmt.Sprintf("%v", condValue)
		return strings.HasPrefix(sourceStr, condStr)
	case "ends_with":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		condStr := fmt.Sprintf("%v", condValue)
		return strings.HasSuffix(sourceStr, condStr)
	case "regex":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		condStr := fmt.Sprintf("%v", condValue)
		if regex, err := regexp.Compile(condStr); err == nil {
			return regex.MatchString(sourceStr)
		}
		return false
	case "greater_than", "gt":
		return t.compareNumbers(sourceValue, condValue, ">")
	case "less_than", "lt":
		return t.compareNumbers(sourceValue, condValue, "<")
	case "greater_equal", "gte":
		return t.compareNumbers(sourceValue, condValue, ">=")
	case "less_equal", "lte":
		return t.compareNumbers(sourceValue, condValue, "<=")
	case "is_null":
		return sourceValue == nil
	case "is_not_null":
		return sourceValue != nil
	case "is_empty":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		return strings.TrimSpace(sourceStr) == ""
	case "is_not_empty":
		sourceStr := fmt.Sprintf("%v", sourceValue)
		return strings.TrimSpace(sourceStr) != ""
	default:
		return false
	}
}

// compareNumbers compares numeric values
func (t *AdvancedDataTransformer) compareNumbers(a, b interface{}, operator string) bool {
	aFloat, aErr := t.toFloat64(a)
	bFloat, bErr := t.toFloat64(b)
	
	if aErr != nil || bErr != nil {
		return false
	}
	
	switch operator {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts interface{} to float64
func (t *AdvancedDataTransformer) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}

// validateWithRegex validates value against regex pattern
func (t *AdvancedDataTransformer) validateWithRegex(value interface{}, pattern string) bool {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		t.logger.Errorw("Invalid regex pattern", logx.Field("pattern", pattern), logx.Field("error", err))
		return false
	}
	
	valueStr := fmt.Sprintf("%v", value)
	return regex.MatchString(valueStr)
}

// getNestedValue retrieves nested value using dot notation
func (t *AdvancedDataTransformer) getNestedValue(data map[string]interface{}, path string) interface{} {
	keys := strings.Split(path, ".")
	current := data
	
	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}
		
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}

// setNestedValue sets nested value using dot notation
func (t *AdvancedDataTransformer) setNestedValue(data map[string]interface{}, path string, value interface{}) {
	keys := strings.Split(path, ".")
	current := data
	
	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return
		}
		
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			next := make(map[string]interface{})
			current[key] = next
			current = next
		}
	}
}

// ParseTransformationRules parses JSON string to transformation rules
func (t *AdvancedDataTransformer) ParseTransformationRules(rulesJSON string) ([]TransformationRule, error) {
	var rules []TransformationRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return nil, fmt.Errorf("failed to parse transformation rules: %w", err)
	}
	return rules, nil
}