package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateCisAttributesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateCisAttributesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateCisAttributesLogic {
	return &ValidateCisAttributesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateCisAttributesLogic) ValidateCisAttributes(in *cmdb.CisAttributeValidateReq) (*cmdb.CisAttributeValidateResp, error) {
	resp := &cmdb.CisAttributeValidateResp{
		Valid:  true,
		Errors: make([]*cmdb.CiAttributeError, 0),
	}

	// 获取CI类型的属性定义
	typeAttrs, err := l.svcCtx.DB.CiTypeAttribute.Query().
		Where(citypeattribute.TypeIDEQ(in.TypeId)).
		WithAttribute().
		All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 创建属性映射
	attrMap := make(map[uint64]*ent.CiTypeAttribute)
	for _, ta := range typeAttrs {
		attrMap[ta.AttrID] = ta
	}

	// 验证每个属性值
	for _, attrValue := range in.Attributes {
		if typeAttr, exists := attrMap[attrValue.AttrId]; exists {
			if err := l.validateAttributeValue(typeAttr, attrValue); err != nil {
				resp.Valid = false
				resp.Errors = append(resp.Errors, &cmdb.CiAttributeError{
					AttrId:    attrValue.AttrId,
					AttrName:  typeAttr.Edges.Attribute.Name,
					ErrorType: "validation_error",
					Message:   err.Error(),
				})
			}
		}
	}

	// 检查必填属性
	for _, typeAttr := range typeAttrs {
		if typeAttr.IsRequired {
			found := false
			for _, attrValue := range in.Attributes {
				if attrValue.AttrId == typeAttr.AttrID {
					found = true
					break
				}
			}
			if !found {
				resp.Valid = false
				resp.Errors = append(resp.Errors, &cmdb.CiAttributeError{
					AttrId:    typeAttr.AttrID,
					AttrName:  typeAttr.Edges.Attribute.Name,
					ErrorType: "required_error",
					Message:   fmt.Sprintf("属性 %s 是必填的", typeAttr.Edges.Attribute.Name),
				})
			}
		}
	}

	return resp, nil
}

// validateAttributeValue 验证单个属性值
func (l *ValidateCisAttributesLogic) validateAttributeValue(typeAttr *ent.CiTypeAttribute, attrValue *cmdb.CiAttributeValue) error {
	attr := typeAttr.Edges.Attribute
	if attr == nil {
		return fmt.Errorf("属性定义不存在")
	}

	// 检查值是否为空
	if attrValue.Value == "" {
		if typeAttr.IsRequired {
			return fmt.Errorf("属性 %s 是必填的", attr.Name)
		}
		return nil
	}

	// 根据值类型验证
	switch attr.ValueType {
	case consts.ValueTypeInt:
		return l.validateIntegerValue(attr, attrValue.Value)
	case consts.ValueTypeFloat:
		return l.validateFloatValue(attr, attrValue.Value)
	case consts.ValueTypeShortText:
		return l.validateShortTextValue(attr, attrValue.Value)
	case consts.ValueTypeLongText:
		return l.validateLongTextValue(attr, attrValue.Value)
	case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
		return l.validateDateTimeValue(attr, attrValue.Value)
	case consts.ValueTypeJSON:
		return l.validateJSONValue(attr, attrValue.Value)
	case consts.ValueTypeLink:
		return l.validateLinkValue(attr, attrValue.Value)
	case consts.ValueTypeBool:
		return l.validateBoolValue(attr, attrValue.Value)
	}

	return nil
}

// validateIntegerValue 验证整数值
func (l *ValidateCisAttributesLogic) validateIntegerValue(attr *ent.Attribute, value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("属性 %s 的值 '%s' 不是有效的整数", attr.Name, value)
	}
	return nil
}

// validateFloatValue 验证浮点数值
func (l *ValidateCisAttributesLogic) validateFloatValue(attr *ent.Attribute, value string) error {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("属性 %s 的值 '%s' 不是有效的浮点数", attr.Name, value)
	}
	return nil
}

// validateShortTextValue 验证短文本值
func (l *ValidateCisAttributesLogic) validateShortTextValue(attr *ent.Attribute, value string) error {
	if len(value) > 255 {
		return fmt.Errorf("属性 %s 的值长度不能超过255个字符", attr.Name)
	}
	return l.validateTextPattern(attr, value)
}

// validateLongTextValue 验证长文本值
func (l *ValidateCisAttributesLogic) validateLongTextValue(attr *ent.Attribute, value string) error {
	return l.validateTextPattern(attr, value)
}

// validateTextPattern 验证文本模式
func (l *ValidateCisAttributesLogic) validateTextPattern(attr *ent.Attribute, value string) error {
	// 使用ValidatorRules进行正则校验
	if len(attr.ValidatorRules) > 0 {
		for _, rule := range attr.ValidatorRules {
			if !rule.Enabled {
				continue
			}

			if rule.Type == "pattern" || rule.Type == "regex" {
				// 从Params中获取正则表达式
				if rule.Params != nil {
					if pattern, ok := rule.Params["pattern"].(string); ok && pattern != "" {
						matched, err := regexp.MatchString(pattern, value)
						if err != nil {
							return fmt.Errorf("属性 %s 的正则表达式无效: %v", attr.Name, err)
						}
						if !matched {
							message := fmt.Sprintf("属性 %s 的值 '%s' 不符合正则表达式 '%s'", attr.Name, value, pattern)
							if rule.Message != "" {
								message = fmt.Sprintf("属性 %s 的值 '%s' %s", attr.Name, value, rule.Message)
							}
							return fmt.Errorf("%s", message)
						}
					}
				}
			}
		}
	}
	return nil
}

// validateDateTimeValue 验证日期时间值
func (l *ValidateCisAttributesLogic) validateDateTimeValue(attr *ent.Attribute, value string) error {
	// 尝试解析时间戳
	if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
		_ = time.UnixMilli(timestamp)
		return nil
	}

	// 尝试解析RFC3339格式
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return nil
	}

	// 尝试解析标准格式
	if _, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return nil
	}

	// 尝试解析日期格式
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return nil
	}

	return fmt.Errorf("属性 %s 的值 '%s' 不是有效的日期时间格式", attr.Name, value)
}

// validateJSONValue 验证JSON值
func (l *ValidateCisAttributesLogic) validateJSONValue(attr *ent.Attribute, value string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		return fmt.Errorf("属性 %s 的值 '%s' 不是有效的JSON格式: %v", attr.Name, value, err)
	}
	return nil
}

// validateLinkValue 验证链接值
func (l *ValidateCisAttributesLogic) validateLinkValue(attr *ent.Attribute, value string) error {
	// 简单的URL格式验证
	urlPattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, err := regexp.MatchString(urlPattern, value)
	if err != nil {
		return fmt.Errorf("链接验证正则表达式错误: %v", err)
	}
	if !matched {
		return fmt.Errorf("属性 %s 的值 '%s' 不是有效的链接格式", attr.Name, value)
	}
	return nil
}

// validateBoolValue 验证布尔值
func (l *ValidateCisAttributesLogic) validateBoolValue(attr *ent.Attribute, value string) error {
	if value != "true" && value != "false" && value != "1" && value != "0" {
		return fmt.Errorf("属性 %s 的值 '%s' 不是有效的布尔值(true/false/1/0)", attr.Name, value)
	}
	return nil
}
