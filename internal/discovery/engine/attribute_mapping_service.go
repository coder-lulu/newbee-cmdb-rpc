package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attributemappingrule"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/zeromicro/go-zero/core/logx"
)

// AttributeMappingService manages attribute mapping rules and transformations
type AttributeMappingService struct {
	db                  *ent.Client
	advancedTransformer *AdvancedDataTransformer
	logger              logx.Logger
}

// MappingContext contains context information for attribute mapping
type MappingContext struct {
	CiTypeID        uint64
	ProviderID      string
	SourceSystem    string
	TransformRules  []TransformationRule
	GlobalSettings  map[string]interface{}
}

// MappingResult contains the result of attribute mapping
type MappingResult struct {
	MappedData      map[string]interface{} `json:"mapped_data"`
	UnmappedFields  []string               `json:"unmapped_fields"`
	ValidationErrors []ValidationError      `json:"validation_errors"`
	TransformStats  TransformationStats    `json:"transform_stats"`
}

// ValidationError represents a validation error during mapping
type ValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// TransformationStats provides statistics about the transformation process
type TransformationStats struct {
	TotalFields      int `json:"total_fields"`
	MappedFields     int `json:"mapped_fields"`
	TransformedFields int `json:"transformed_fields"`
	ValidationErrors int `json:"validation_errors"`
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// NewAttributeMappingService creates a new attribute mapping service
func NewAttributeMappingService(db *ent.Client) *AttributeMappingService {
	return &AttributeMappingService{
		db:                  db,
		advancedTransformer: NewAdvancedDataTransformer(),
		logger:              logx.WithContext(nil),
	}
}

// LoadMappingRules loads mapping rules for a specific CI type and provider
func (s *AttributeMappingService) LoadMappingRules(ctx context.Context, ciTypeID uint64, providerID string) ([]TransformationRule, error) {
	rules, err := s.db.AttributeMappingRule.Query().
		Where(
			attributemappingrule.DiscoveryConfigIDIn(ciTypeID), // Use discovery config ID instead
			attributemappingrule.EnabledEQ(true),
		).
		Order(ent.Asc(attributemappingrule.FieldPriority)).
		All(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to load mapping rules: %w", err)
	}
	
	var transformRules []TransformationRule
	for _, rule := range rules {
		transformRule, err := s.convertToTransformationRule(rule)
		if err != nil {
			s.logger.Errorw("Failed to convert mapping rule",
				logx.Field("rule_id", rule.ID),
				logx.Field("error", err))
			continue
		}
		transformRules = append(transformRules, transformRule)
	}
	
	return transformRules, nil
}

// MapAttributes performs comprehensive attribute mapping for source data
func (s *AttributeMappingService) MapAttributes(ctx context.Context, sourceData map[string]interface{}, mappingCtx MappingContext) (*MappingResult, error) {
	startTime := getCurrentTimeMs()
	
	result := &MappingResult{
		MappedData:       make(map[string]interface{}),
		UnmappedFields:   []string{},
		ValidationErrors: []ValidationError{},
		TransformStats: TransformationStats{
			TotalFields: len(sourceData),
		},
	}
	
	// Load mapping rules if not provided
	if len(mappingCtx.TransformRules) == 0 {
		rules, err := s.LoadMappingRules(ctx, mappingCtx.CiTypeID, mappingCtx.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("failed to load mapping rules: %w", err)
		}
		mappingCtx.TransformRules = rules
	}
	
	// Track mapped fields
	mappedFields := make(map[string]bool)
	
	// Apply transformation rules
	for _, rule := range mappingCtx.TransformRules {
		if err := s.applyMappingRule(sourceData, result, rule, mappedFields); err != nil {
			result.ValidationErrors = append(result.ValidationErrors, ValidationError{
				Field:   rule.SourceField,
				Rule:    rule.TransformType,
				Message: err.Error(),
			})
		}
	}
	
	// Identify unmapped fields
	for field := range sourceData {
		if !mappedFields[field] {
			result.UnmappedFields = append(result.UnmappedFields, field)
		}
	}
	
	// Calculate statistics
	result.TransformStats.MappedFields = len(mappedFields)
	result.TransformStats.ValidationErrors = len(result.ValidationErrors)
	result.TransformStats.ProcessingTimeMs = getCurrentTimeMs() - startTime
	
	// Apply global transformations if configured
	if len(mappingCtx.GlobalSettings) > 0 {
		s.applyGlobalTransformations(result.MappedData, mappingCtx.GlobalSettings)
	}
	
	return result, nil
}

// applyMappingRule applies a single mapping rule
func (s *AttributeMappingService) applyMappingRule(sourceData map[string]interface{}, result *MappingResult, rule TransformationRule, mappedFields map[string]bool) error {
	// Mark source field as mapped
	mappedFields[rule.SourceField] = true
	
	// Apply transformation
	transformedValue, err := s.advancedTransformer.applyTransformationRule(sourceData, rule)
	if err != nil {
		return fmt.Errorf("transformation failed for field %s: %w", rule.SourceField, err)
	}
	
	if transformedValue != nil {
		result.MappedData[rule.TargetField] = transformedValue
		result.TransformStats.TransformedFields++
	}
	
	return nil
}

// applyGlobalTransformations applies global transformation settings
func (s *AttributeMappingService) applyGlobalTransformations(data map[string]interface{}, settings map[string]interface{}) {
	// Apply global prefix/suffix if configured
	if prefix, ok := settings["field_prefix"].(string); ok && prefix != "" {
		newData := make(map[string]interface{})
		for k, v := range data {
			newData[prefix+k] = v
		}
		// Replace data content
		for k := range data {
			delete(data, k)
		}
		for k, v := range newData {
			data[k] = v
		}
	}
	
	// Apply field name transformations
	if nameCase, ok := settings["field_name_case"].(string); ok {
		switch nameCase {
		case "lower":
			s.transformFieldNames(data, func(name string) string {
				return strings.ToLower(name)
			})
		case "upper":
			s.transformFieldNames(data, func(name string) string {
				return strings.ToUpper(name)
			})
		case "camelCase":
			s.transformFieldNames(data, s.toCamelCase)
		case "snake_case":
			s.transformFieldNames(data, s.toSnakeCase)
		}
	}
}

// transformFieldNames transforms all field names using the provided function
func (s *AttributeMappingService) transformFieldNames(data map[string]interface{}, transformer func(string) string) {
	newData := make(map[string]interface{})
	for k, v := range data {
		newKey := transformer(k)
		newData[newKey] = v
	}
	
	// Replace data content
	for k := range data {
		delete(data, k)
	}
	for k, v := range newData {
		data[k] = v
	}
}

// toCamelCase converts string to camelCase
func (s *AttributeMappingService) toCamelCase(str string) string {
	words := strings.FieldsFunc(str, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})
	
	if len(words) == 0 {
		return str
	}
	
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += strings.Title(strings.ToLower(words[i]))
	}
	
	return result
}

// toSnakeCase converts string to snake_case
func (s *AttributeMappingService) toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// convertToTransformationRule converts database rule to transformation rule
func (s *AttributeMappingService) convertToTransformationRule(rule *ent.AttributeMappingRule) (TransformationRule, error) {
	transformRule := TransformationRule{
		SourceField:     rule.SourceField,
		TargetField:     rule.TargetAttribute,
		TransformType:   rule.TransformType,
		DefaultValue:    rule.DefaultValue,
		ValidationRegex: rule.ValidationRegex,
	}
	
	// Parse transform config JSON
	if rule.TransformConfig != nil {
		var config map[string]interface{}
		configStr := jsonhelper.JSONToString(rule.TransformConfig)
		if configStr != nil && *configStr != "" {
			if err := json.Unmarshal([]byte(*configStr), &config); err == nil {
				// Extract conditional rules if present
				if condRules, ok := config["conditional_rules"].([]interface{}); ok {
					for _, condRule := range condRules {
						if condMap, ok := condRule.(map[string]interface{}); ok {
							transformRule.ConditionalRules = append(transformRule.ConditionalRules, ConditionalRule{
								Condition:   getString(condMap, "condition"),
								Value:       condMap["value"],
								TargetValue: condMap["target_value"],
								ElseValue:   condMap["else_value"],
							})
						}
					}
				}
				
				// Extract array handling if present
				if arrayConfig, ok := config["array_handling"].(map[string]interface{}); ok {
					transformRule.ArrayHandling = ArrayHandlingRule{
						Mode:      getString(arrayConfig, "mode"),
						Separator: getString(arrayConfig, "separator"),
					}
					
					if indices, ok := arrayConfig["index_filter"].([]interface{}); ok {
						for _, idx := range indices {
							if intIdx, ok := idx.(float64); ok {
								transformRule.ArrayHandling.IndexFilter = append(transformRule.ArrayHandling.IndexFilter, int(intIdx))
							}
						}
					}
				}
			}
		}
	}
	
	return transformRule, nil
}

// getString safely gets string value from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// CreateMappingRule creates a new attribute mapping rule
func (s *AttributeMappingService) CreateMappingRule(ctx context.Context, rule *ent.AttributeMappingRule) (*ent.AttributeMappingRule, error) {
	// Validate rule before creation
	if err := s.validateMappingRule(rule); err != nil {
		return nil, fmt.Errorf("rule validation failed: %w", err)
	}
	
	createdRule, err := s.db.AttributeMappingRule.Create().
		SetDiscoveryConfigID(rule.DiscoveryConfigID).
		SetCiAttributeID(rule.CiAttributeID).
		SetSourceField(rule.SourceField).
		SetTargetAttribute(rule.TargetAttribute).
		SetTransformType(rule.TransformType).
		SetNotNilTransformConfig(&rule.TransformConfig).
		SetNotNilValidationRegex(&rule.ValidationRegex).
		SetNotNilDefaultValue(&rule.DefaultValue).
		SetNotNilPriority(&rule.Priority).
		SetNotNilEnabled(&rule.Enabled).
		SetNotNilDescription(&rule.Description).
		Save(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create mapping rule: %w", err)
	}
	
	return createdRule, nil
}

// validateMappingRule validates a mapping rule
func (s *AttributeMappingService) validateMappingRule(rule *ent.AttributeMappingRule) error {
	if rule.DiscoveryConfigID == 0 {
		return fmt.Errorf("discovery_config_id is required")
	}
	
	if rule.CiAttributeID == 0 {
		return fmt.Errorf("ci_attribute_id is required")
	}
	
	if rule.SourceField == "" {
		return fmt.Errorf("source_field is required")
	}
	
	if rule.TargetAttribute == "" {
		return fmt.Errorf("target_attribute is required")
	}
	
	// Validate transformation type
	validTypes := []string{"direct", "lookup", "script", "template", "string", "int", "float", "bool", "datetime", "json", "upper", "lower", "trim"}
	if rule.TransformType != "" {
		valid := false
		for _, validType := range validTypes {
			if rule.TransformType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid transformation type: %s", rule.TransformType)
		}
	}
	
	return nil
}

// getCurrentTimeMs returns current time in milliseconds
func getCurrentTimeMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}