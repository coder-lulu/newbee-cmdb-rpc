package attribute

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/attribute"
	"gitee.com/link234/cmdb-rpc/ent/choicefloat"
	"gitee.com/link234/cmdb-rpc/ent/choiceinteger"
	"gitee.com/link234/cmdb-rpc/ent/choicetext"
	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/uuidx"
	"gitee.com/link234/newbee-backend-common/utils/validator"
)

// isZeroAttributeOptionS 检查 AttributeOptionS 是否为零值
func isZeroAttributeOptionS(opt schema.AttributeOptionS) bool {
	return reflect.DeepEqual(opt, schema.AttributeOptionS{})
}

// isZeroAttributeDefaultValueS 检查 AttributeDefaultValueS 是否为零值
func isZeroAttributeDefaultValueS(def schema.AttributeDefaultValueS) bool {
	return reflect.DeepEqual(def, schema.AttributeDefaultValueS{})
}

// isZeroAttributeChoiceWebHookS 检查 AttributeChoiceWebHookS 是否为零值
func isZeroAttributeChoiceWebHookS(hook schema.AttributeChoiceWebHookS) bool {
	return reflect.DeepEqual(hook, schema.AttributeChoiceWebHookS{})
}

// isZeroAttributeChoiceOtherS 检查 AttributeChoiceOtherS 是否为零值
func isZeroAttributeChoiceOtherS(other schema.AttributeChoiceOtherS) bool {
	return reflect.DeepEqual(other, schema.AttributeChoiceOtherS{})
}

// convertValidatorRulesToProto 将 []validator.ValidationRule 转换为 []*cmdb.ValidationRule
func convertValidatorRulesToProto(rules []validator.ValidationRule) []*cmdb.ValidationRule {
	if len(rules) == 0 {
		return nil
	}
	protoRules := make([]*cmdb.ValidationRule, 0, len(rules))
	for _, rule := range rules {
		protoRule := &cmdb.ValidationRule{
			Type:    &rule.Type,
			Message: &rule.Message,
			Enabled: &rule.Enabled,
		}

		if rule.CustomValidator != "" {
			protoRule.CustomValidator = &rule.CustomValidator
		}

		// 转换 Params: map[string]interface{} → []*cmdb.ValidationRuleParams
		if rule.Params != nil && len(rule.Params) > 0 {
			params := make([]*cmdb.ValidationRuleParams, 0, len(rule.Params))
			for key, value := range rule.Params {
				valueStr := fmt.Sprintf("%v", value)
				params = append(params, &cmdb.ValidationRuleParams{
					Key:   &key,
					Value: &valueStr,
				})
			}
			protoRule.Params = params
		}

		protoRules = append(protoRules, protoRule)
	}

	return protoRules
}

// convertProtoToValidatorRules 将 []*cmdb.ValidationRule 转换为 []validator.ValidationRule
func convertProtoToValidatorRules(protoRules []*cmdb.ValidationRule) []validator.ValidationRule {
	if len(protoRules) == 0 {
		return nil
	}

	rules := make([]validator.ValidationRule, 0, len(protoRules))
	for _, protoRule := range protoRules {
		if protoRule == nil {
			continue
		}

		rule := validator.ValidationRule{
			Type:    getStringValue(protoRule.Type),
			Message: getStringValue(protoRule.Message),
			Enabled: getBoolValue(protoRule.Enabled),
		}

		if protoRule.CustomValidator != nil {
			rule.CustomValidator = *protoRule.CustomValidator
		}

		// 转换 Params: []*cmdb.ValidationRuleParams → map[string]interface{}
		if len(protoRule.Params) > 0 {
			params := make(map[string]interface{})
			for _, param := range protoRule.Params {
				if param != nil && param.Key != nil && param.Value != nil {
					// 尝试转换为合适的类型
					value := convertParamValue(*param.Value)
					params[*param.Key] = value
				}
			}
			rule.Params = params
		}

		rules = append(rules, rule)
	}

	return rules
}

// convertParamValue 尝试将字符串值转换为合适的类型
func convertParamValue(valueStr string) interface{} {
	// 尝试转换为数字
	if intVal, err := strconv.Atoi(valueStr); err == nil {
		return intVal
	}

	// 尝试转换为浮点数
	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal
	}

	// 尝试转换为布尔值
	if boolVal, err := strconv.ParseBool(valueStr); err == nil {
		return boolVal
	}

	// 默认返回字符串
	return valueStr
}

// getBoolValue 安全获取布尔值
func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// AttributeEntToProto 将 ent.Attribute 转为 cmdb.AttributeInfo
func AttributeEntToProto(
	v *ent.Attribute,
	choices []*cmdb.AttributeChoiceItem,
) *cmdb.AttributeInfo {
	var createdBy *string
	if v.CreatedBy != nil {
		s := v.CreatedBy.String()
		createdBy = &s
	}

	// 处理 Option 字段
	var optionObj *cmdb.AttributeOption
	if !isZeroAttributeOptionS(v.Option) {
		optionObj = &cmdb.AttributeOption{}
		if !reflect.DeepEqual(v.Option.FontOption, schema.FontOptionS{}) {
			optionObj.FontOption = &cmdb.AttributeFontOption{
				Color:          &v.Option.FontOption.Color,
				BgColor:        &v.Option.FontOption.BgColor,
				FontStyle:      &v.Option.FontOption.FontStyle,
				FontWeight:     &v.Option.FontOption.FontWeight,
				TextDecoration: &v.Option.FontOption.TextDecoration,
			}
		}
		for _, img := range v.Option.ImageOption {
			optionObj.ImageOptions = append(optionObj.ImageOptions, &cmdb.AttributeImageOption{
				Path:   img.Path,
				Name:   img.Name,
				Height: img.Height,
				Width:  img.Width,
			})
		}
	}

	// 处理 Default 字段
	var defaultObj *cmdb.AttributeDefault
	if !isZeroAttributeDefaultValueS(v.Default) {
		defaultObj = &cmdb.AttributeDefault{}
		if v.Default.Default != nil {
			defaultStr := fmt.Sprintf("%v", v.Default.Default)
			defaultObj.Default = &defaultStr
		}
	}

	// 处理 ChoiceWebHook 字段
	var choiceWebHookObj *cmdb.AttributeChoiceWebHook
	if !isZeroAttributeChoiceWebHookS(v.ChoiceWebHook) {
		choiceWebHookObj = &cmdb.AttributeChoiceWebHook{
			Url:            &v.ChoiceWebHook.Url,
			Method:         &v.ChoiceWebHook.Method,
			Body:           &v.ChoiceWebHook.Body,
			Headers:        &v.ChoiceWebHook.Headers,
			Params:         &v.ChoiceWebHook.Params,
			ResponseType:   &v.ChoiceWebHook.ResponseType,
			ResponseFormat: &v.ChoiceWebHook.ResponseFormat,
		}
	}

	// 处理 ChoiceOther 字段
	var choiceOtherObj *cmdb.AttributeChoiceOther
	if !isZeroAttributeChoiceOtherS(v.ChoiceOther) {
		choiceOtherObj = &cmdb.AttributeChoiceOther{
			Filter:  &v.ChoiceOther.Filter,
			AttrId:  &v.ChoiceOther.AttrId,
			TypeIds: v.ChoiceOther.TypeIds,
		}
	}

	// 转换 ValueType 枚举为字符串
	valueTypeStr := string(v.ValueType)
	// 转换 ValidatorRules
	validatorRules := convertValidatorRulesToProto(v.ValidatorRules)

	return &cmdb.AttributeInfo{
		Id:              &v.ID,
		Name:            &v.Name,
		Alias:           &v.Alias,
		ValueType:       &valueTypeStr,
		IsChoice:        &v.IsChoice,
		IsList:          &v.IsList,
		CreatedBy:       createdBy,
		IsComputed:      &v.IsComputed,
		ChoiceWebHook:   choiceWebHookObj,
		Option:          optionObj,
		IsPassword:      &v.IsPassword,
		ComputeScript:   &v.ComputeScript,
		ComputeExpr:     &v.ComputeExpr,
		IsSortable:      &v.IsSortable,
		Default:         defaultObj,
		IsDynamic:       &v.IsDynamic,
		ReferenceTypeId: &v.ReferenceTypeID,
		ChoiceOther:     choiceOtherObj,
		Choices:         choices,
		ValidatorRules:  validatorRules,
	}
}

// AttributeCreateBuilderSetter 用于设置 ent.AttributeCreate 的通用字段
func AttributeCreateBuilderSetter(builder *ent.AttributeCreate, in *cmdb.AttributeInfo) *ent.AttributeCreate {
	if in == nil {
		return builder
	}

	name := strings.TrimSpace(*in.Name)
	alias := strings.TrimSpace(*in.Alias)

	// 转换 ValueType 字符串为枚举
	var valueType *attribute.ValueType
	if in.ValueType != nil {
		vt := attribute.ValueType(*in.ValueType)
		valueType = &vt
	}

	// 处理 Option 字段
	var optionS schema.AttributeOptionS
	if in.Option != nil {
		if in.Option.FontOption != nil {
			optionS.FontOption = schema.FontOptionS{
				Color:          getStringValue(in.Option.FontOption.Color),
				BgColor:        getStringValue(in.Option.FontOption.BgColor),
				FontStyle:      getStringValue(in.Option.FontOption.FontStyle),
				FontWeight:     getStringValue(in.Option.FontOption.FontWeight),
				TextDecoration: getStringValue(in.Option.FontOption.TextDecoration),
			}
		}
		for _, img := range in.Option.ImageOptions {
			optionS.ImageOption = append(optionS.ImageOption, schema.ImageOptionS{
				Path:   img.Path,
				Name:   img.Name,
				Height: img.Height,
				Width:  img.Width,
			})
		}
	}

	// 处理 Default 字段
	var defaultS schema.AttributeDefaultValueS
	if in.Default != nil && in.Default.Default != nil {
		defaultS.Default = *in.Default.Default
	}

	// 处理 ChoiceWebHook 字段
	var choiceWebHookS schema.AttributeChoiceWebHookS
	if in.ChoiceWebHook != nil {
		choiceWebHookS = schema.AttributeChoiceWebHookS{
			Url:            getStringValue(in.ChoiceWebHook.Url),
			Method:         getStringValue(in.ChoiceWebHook.Method),
			Body:           getStringValue(in.ChoiceWebHook.Body),
			Headers:        getStringValue(in.ChoiceWebHook.Headers),
			Params:         getStringValue(in.ChoiceWebHook.Params),
			ResponseType:   getStringValue(in.ChoiceWebHook.ResponseType),
			ResponseFormat: getStringValue(in.ChoiceWebHook.ResponseFormat),
		}
	}

	// 处理 ChoiceOther 字段
	var choiceOtherS schema.AttributeChoiceOtherS
	if in.ChoiceOther != nil {
		choiceOtherS = schema.AttributeChoiceOtherS{
			Filter:  getStringValue(in.ChoiceOther.Filter),
			AttrId:  getUint64Value(in.ChoiceOther.AttrId),
			TypeIds: in.ChoiceOther.TypeIds,
		}
	}

	// 转换 ValidatorRules
	validatorRules := convertProtoToValidatorRules(in.ValidatorRules)

	builder = builder.
		SetNotNilName(&name).
		SetNotNilAlias(&alias).
		SetNotNilValueType(valueType).
		SetNotNilIsChoice(in.IsChoice).
		SetNotNilIsList(in.IsList).
		SetNotNilCreatedBy(uuidx.ParseUUIDStringToPointer(in.CreatedBy)).
		SetNotNilIsComputed(in.IsComputed).
		SetNotNilChoiceWebHook(&choiceWebHookS).
		SetNotNilOption(&optionS).
		SetNotNilIsPassword(in.IsPassword).
		SetNotNilComputeScript(in.ComputeScript).
		SetNotNilComputeExpr(in.ComputeExpr).
		SetNotNilIsSortable(in.IsSortable).
		SetNotNilDefault(&defaultS).
		SetNotNilIsDynamic(in.IsDynamic).
		SetNotNilIsReference(in.IsReference).
		SetNotNilReferenceTypeID(in.ReferenceTypeId).
		SetNotNilChoiceOther(&choiceOtherS)

	// 设置 ValidatorRules（如果不为空）
	if len(validatorRules) > 0 {
		builder = builder.SetValidatorRules(validatorRules)
	}

	return builder
}

// AttributeUpdateBuilderSetter 用于设置 ent.AttributeUpdateOne 的通用字段
func AttributeUpdateBuilderSetter(builder *ent.AttributeUpdateOne, in *cmdb.AttributeInfo, alias *string) *ent.AttributeUpdateOne {
	if in == nil {
		return builder
	}

	// 处理 Option 字段
	var optionS schema.AttributeOptionS
	if in.Option != nil {
		if in.Option.FontOption != nil {
			optionS.FontOption = schema.FontOptionS{
				Color:          getStringValue(in.Option.FontOption.Color),
				BgColor:        getStringValue(in.Option.FontOption.BgColor),
				FontStyle:      getStringValue(in.Option.FontOption.FontStyle),
				FontWeight:     getStringValue(in.Option.FontOption.FontWeight),
				TextDecoration: getStringValue(in.Option.FontOption.TextDecoration),
			}
		}
		for _, img := range in.Option.ImageOptions {
			optionS.ImageOption = append(optionS.ImageOption, schema.ImageOptionS{
				Path:   img.Path,
				Name:   img.Name,
				Height: img.Height,
				Width:  img.Width,
			})
		}
	}

	// 处理 Default 字段
	var defaultS schema.AttributeDefaultValueS
	if in.Default != nil && in.Default.Default != nil {
		defaultS.Default = *in.Default.Default
	}

	// 处理 ChoiceWebHook 字段
	var choiceWebHookS schema.AttributeChoiceWebHookS
	if in.ChoiceWebHook != nil {
		choiceWebHookS = schema.AttributeChoiceWebHookS{
			Url:            getStringValue(in.ChoiceWebHook.Url),
			Method:         getStringValue(in.ChoiceWebHook.Method),
			Body:           getStringValue(in.ChoiceWebHook.Body),
			Headers:        getStringValue(in.ChoiceWebHook.Headers),
			Params:         getStringValue(in.ChoiceWebHook.Params),
			ResponseType:   getStringValue(in.ChoiceWebHook.ResponseType),
			ResponseFormat: getStringValue(in.ChoiceWebHook.ResponseFormat),
		}
	}

	// 处理 ChoiceOther 字段
	var choiceOtherS schema.AttributeChoiceOtherS
	if in.ChoiceOther != nil {
		choiceOtherS = schema.AttributeChoiceOtherS{
			Filter:  getStringValue(in.ChoiceOther.Filter),
			AttrId:  getUint64Value(in.ChoiceOther.AttrId),
			TypeIds: in.ChoiceOther.TypeIds,
		}
	}

	// 转换 ValidatorRules
	validatorRules := convertProtoToValidatorRules(in.ValidatorRules)

	builder = builder.
		SetNotNilAlias(alias).
		SetNotNilIsChoice(in.IsChoice).
		SetNotNilIsComputed(in.IsComputed).
		SetNotNilChoiceWebHook(&choiceWebHookS).
		SetNotNilOption(&optionS).
		SetNotNilIsPassword(in.IsPassword).
		SetNotNilComputeScript(in.ComputeScript).
		SetNotNilComputeExpr(in.ComputeExpr).
		SetNotNilIsSortable(in.IsSortable).
		SetNotNilDefault(&defaultS).
		SetNotNilIsDynamic(in.IsDynamic).
		SetNotNilIsReference(in.IsReference).
		SetNotNilReferenceTypeID(in.ReferenceTypeId).
		SetNotNilChoiceOther(&choiceOtherS)

	// 设置 ValidatorRules（如果不为空）
	if len(validatorRules) > 0 {
		builder = builder.SetValidatorRules(validatorRules)
	}

	return builder
}

// CreateOrUpdateChoices 统一处理属性选项的创建、更新和删除
func CreateOrUpdateChoices(ctx context.Context, tx *ent.Tx, attrID *uint64, valueType *string, choices []*cmdb.AttributeChoiceItem) error {
	if attrID == nil || valueType == nil {
		return nil
	}

	// 1. 获取现有选项
	var existingValues []interface{}
	switch *valueType {
	case consts.ValueTypeShortText, consts.ValueTypeLongText,
		consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime,
		consts.ValueTypeLink, consts.ValueTypeImage:
		existingChoices, err := tx.ChoiceText.Query().
			Where(choicetext.AttrID(*attrID)).
			All(ctx)
		if err != nil {
			return err
		}
		for _, c := range existingChoices {
			existingValues = append(existingValues, c.Value)
		}
	case consts.ValueTypeInt:
		existingChoices, err := tx.ChoiceInteger.Query().
			Where(choiceinteger.AttrID(*attrID)).
			All(ctx)
		if err != nil {
			return err
		}
		for _, c := range existingChoices {
			existingValues = append(existingValues, c.Value)
		}
	case consts.ValueTypeFloat:
		existingChoices, err := tx.ChoiceFloat.Query().
			Where(choicefloat.AttrID(*attrID)).
			All(ctx)
		if err != nil {
			return err
		}
		for _, c := range existingChoices {
			existingValues = append(existingValues, c.Value)
		}
	}

	// 2. 找出需要删除的选项
	newValues := make(map[string]bool)
	for _, choice := range choices {
		newValues[choice.Value] = true
	}

	// 3. 删除不存在的选项
	for _, existingValue := range existingValues {
		strValue := fmt.Sprintf("%v", existingValue)
		if !newValues[strValue] {
			switch *valueType {
			case consts.ValueTypeShortText, consts.ValueTypeLongText,
				consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime,
				consts.ValueTypeLink, consts.ValueTypeImage:
				_, err := tx.ChoiceText.Delete().
					Where(choicetext.AttrID(*attrID), choicetext.ValueEQ(strValue)).
					Exec(ctx)
				if err != nil {
					return err
				}
			case consts.ValueTypeInt:
				intValue, err := strconv.Atoi(strValue)
				if err != nil {
					return err
				}
				_, err = tx.ChoiceInteger.Delete().
					Where(choiceinteger.AttrID(*attrID), choiceinteger.ValueEQ(intValue)).
					Exec(ctx)
				if err != nil {
					return err
				}
			case consts.ValueTypeFloat:
				floatValue, err := strconv.ParseFloat(strValue, 64)
				if err != nil {
					return err
				}
				_, err = tx.ChoiceFloat.Delete().
					Where(choicefloat.AttrID(*attrID), choicefloat.ValueEQ(floatValue)).
					Exec(ctx)
				if err != nil {
					return err
				}
			}
		}
	}

	// 4. 创建或更新选项
	for _, choice := range choices {
		// 构造 ChoiceItemMetaS
		metaS := schema.ChoiceItemMetaS{
			Label: choice.Meta.Label,
			Icon:  choice.Meta.Icon,
		}
		if choice.Meta.Style != nil {
			metaS.Style = schema.FontOptionS{
				Color:          getStringValue(choice.Meta.Style.Color),
				BgColor:        getStringValue(choice.Meta.Style.BgColor),
				FontStyle:      getStringValue(choice.Meta.Style.FontStyle),
				FontWeight:     getStringValue(choice.Meta.Style.FontWeight),
				TextDecoration: getStringValue(choice.Meta.Style.TextDecoration),
			}
		}

		switch *valueType {
		case consts.ValueTypeShortText, consts.ValueTypeLongText,
			consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime,
			consts.ValueTypeLink, consts.ValueTypeImage:
			isExist, err := tx.ChoiceText.Query().
				Where(choicetext.AttrID(*attrID), choicetext.ValueEQ(choice.Value)).
				Exist(ctx)
			if err != nil {
				return err
			}
			if isExist {
				// 更新现有选项
				err = tx.ChoiceText.Update().
					Where(choicetext.AttrID(*attrID), choicetext.ValueEQ(choice.Value)).
					SetOption(metaS).
					Exec(ctx)
				if err != nil {
					return err
				}
			} else {
				// 创建新选项
				_, err = tx.ChoiceText.Create().
					SetNotNilAttrID(attrID).
					SetNotNilValue(&choice.Value).
					SetNotNilOption(&metaS).
					Save(ctx)
				if err != nil {
					return err
				}
			}
		case consts.ValueTypeInt:
			intValue, err := strconv.Atoi(choice.Value)
			if err != nil {
				return err
			}
			isExist, err := tx.ChoiceInteger.Query().
				Where(choiceinteger.AttrID(*attrID), choiceinteger.ValueEQ(intValue)).
				Exist(ctx)
			if err != nil {
				return err
			}
			if isExist {
				// 更新现有选项
				err = tx.ChoiceInteger.Update().
					Where(choiceinteger.AttrID(*attrID), choiceinteger.ValueEQ(intValue)).
					SetOption(metaS).
					Exec(ctx)
				if err != nil {
					return err
				}
			} else {
				// 创建新选项
				_, err = tx.ChoiceInteger.Create().
					SetNotNilAttrID(attrID).
					SetNotNilValue(&intValue).
					SetNotNilOption(&metaS).
					Save(ctx)
				if err != nil {
					return err
				}
			}
		case consts.ValueTypeFloat:
			floatValue, err := strconv.ParseFloat(choice.Value, 64)
			if err != nil {
				return err
			}
			isExist, err := tx.ChoiceFloat.Query().
				Where(choicefloat.AttrID(*attrID), choicefloat.ValueEQ(floatValue)).
				Exist(ctx)
			if err != nil {
				return err
			}
			if isExist {
				// 更新现有选项
				err = tx.ChoiceFloat.Update().
					Where(choicefloat.AttrID(*attrID), choicefloat.ValueEQ(floatValue)).
					SetOption(metaS).
					Exec(ctx)
				if err != nil {
					return err
				}
			} else {
				// 创建新选项
				_, err = tx.ChoiceFloat.Create().
					SetNotNilAttrID(attrID).
					SetNotNilValue(&floatValue).
					SetNotNilOption(&metaS).
					Save(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 辅助函数：安全获取字符串值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// 辅助函数：安全获取 uint64 值
func getUint64Value(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}
