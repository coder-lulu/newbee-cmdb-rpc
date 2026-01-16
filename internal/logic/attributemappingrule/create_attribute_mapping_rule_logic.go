package attributemappingrule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAttributeMappingRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAttributeMappingRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAttributeMappingRuleLogic {
	return &CreateAttributeMappingRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAttributeMappingRuleLogic) CreateAttributeMappingRule(in *cmdb.AttributeMappingRuleInfo) (*cmdb.BaseIDResp, error) {
	// Validate transformation config if provided
	if in.TransformConfig != nil && *in.TransformConfig != "" {
		if err := l.validateTransformConfig(*in.TransformConfig); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}
	}
	
	// Create query with basic fields
    query := l.svcCtx.DB.AttributeMappingRule.Create().
			SetNotNilDiscoveryConfigID(in.DiscoveryConfigId).
			SetNotNilCiAttributeID(in.CiAttributeId).
			SetNotNilSourceField(in.SourceField).
			SetNotNilSourcePath(in.SourcePath).
			SetNotNilTargetAttribute(in.TargetAttribute).
			SetNotNilTransformType(in.TransformType).
			SetNotNilDefaultValue(in.DefaultValue).
			SetNotNilValidationRegex(in.ValidationRegex).
			SetNotNilIsRequired(in.IsRequired).
			SetNotNilIsUnique(in.IsUnique).
			SetNotNilEnabled(in.Enabled).
			SetNotNilUpdateStrategy(in.UpdateStrategy).
			SetNotNilSuccessCount(in.SuccessCount).
			SetNotNilFailedCount(in.FailedCount).
			SetNotNilLastSuccessAt(pointy.GetTimeMilliPointer(in.LastSuccessAt)).
			SetNotNilDescription(in.Description)

	// Handle JSON fields with proper conversion
	if in.TransformConfig != nil {
		transformConfig := jsonhelper.StringToJSONMap(in.TransformConfig)
		if transformConfig != nil {
			query.SetTransformConfig(transformConfig)
		}
	}
	
	if in.ValidationRules != nil {
		validationRules := jsonhelper.StringToJSONMap(in.ValidationRules)
		if validationRules != nil {
			query.SetValidationRules(validationRules)
		}
	}
	
	if in.Metadata != nil {
		metadata := jsonhelper.StringToJSONMap(in.Metadata)
		if metadata != nil {
			query.SetMetadata(metadata)
		}
	}

	// Handle optional fields with type conversion
	if in.Status != nil {
		query.SetNotNilStatus(pointy.GetPointer(uint8(*in.Status)))
	}
	if in.Priority != nil {
		query.SetNotNilPriority(pointy.GetPointer(int(*in.Priority)))
	}

	// Save the rule
	result, err := query.Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// Log successful creation
	l.Logger.Infow("Attribute mapping rule created successfully",
		logx.Field("rule_id", result.ID),
		logx.Field("source_field", in.SourceField),
		logx.Field("target_attribute", in.TargetAttribute))

    return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess }, nil
}

// validateTransformConfig validates the transformation configuration JSON
func (l *CreateAttributeMappingRuleLogic) validateTransformConfig(configJSON string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	
	// Validate conditional rules if present
	if condRules, ok := config["conditional_rules"].([]interface{}); ok {
		for i, rule := range condRules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				return fmt.Errorf("conditional rule %d is not a valid object", i)
			}
			
			// Check required fields
			if _, ok := ruleMap["condition"]; !ok {
				return fmt.Errorf("conditional rule %d missing 'condition' field", i)
			}
			
			// Validate condition type
			condition, ok := ruleMap["condition"].(string)
			if !ok {
				return fmt.Errorf("conditional rule %d 'condition' must be a string", i)
			}
			
			validConditions := []string{"equals", "eq", "not_equals", "ne", "contains", "starts_with", "ends_with", "regex", "greater_than", "gt", "less_than", "lt", "greater_equal", "gte", "less_equal", "lte", "is_null", "is_not_null", "is_empty", "is_not_empty"}
			valid := false
			for _, validCond := range validConditions {
				if condition == validCond {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("conditional rule %d has invalid condition: %s", i, condition)
			}
		}
	}
	
	// Validate array handling if present
	if arrayConfig, ok := config["array_handling"].(map[string]interface{}); ok {
		if mode, ok := arrayConfig["mode"].(string); ok {
			validModes := []string{"first", "last", "join", "split", "map"}
			valid := false
			for _, validMode := range validModes {
				if mode == validMode {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid array handling mode: %s", mode)
			}
		}
	}
	
	return nil
}
