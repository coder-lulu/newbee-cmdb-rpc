package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/cis"
	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/ent/valuedatetime"
	"gitee.com/link234/cmdb-rpc/ent/valuefloat"
	"gitee.com/link234/cmdb-rpc/ent/valueindextext"
	"gitee.com/link234/cmdb-rpc/ent/valueinteger"
	"gitee.com/link234/cmdb-rpc/ent/valuejson"
	"gitee.com/link234/cmdb-rpc/ent/valuetext"
	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"gitee.com/link234/newbee-backend-common/utils/uuidx"
)

// CisEntToProto 将 ent.Cis 转为 cmdb.CisInfo
func CisEntToProto(v *ent.Cis, attributes []*cmdb.CiAttributeValue) *cmdb.CisInfo {
	// 空指针检查
	if v == nil {
		return nil
	}

	var createdAt, updatedAt *int64
	if !v.CreatedAt.IsZero() {
		createdAt = pointy.GetPointer(v.CreatedAt.UnixMilli())
	}
	if !v.UpdatedAt.IsZero() {
		updatedAt = pointy.GetPointer(v.UpdatedAt.UnixMilli())
	}

	var createdBy *string
	if v.CreatedBy != nil {
		s := v.CreatedBy.String()
		createdBy = &s
	}

	// 处理tags - 从JSON数组转换为字符串数组
	var tags []string
	if v.Tags != nil {
		for _, tag := range v.Tags {
			tags = append(tags, fmt.Sprintf("%s:%s", tag.Key, tag.Value))
		}
	}

	// 处理metadata - 从map转换为CisMetadata数组
	var metadata []*cmdb.CisMetadata
	if v.Metadata != nil {
		for key, value := range v.Metadata {
			valueStr := fmt.Sprintf("%v", value)
			metadata = append(metadata, &cmdb.CisMetadata{
				Key:   &key,
				Value: &valueStr,
			})
		}
	}

	// 处理custom_fields - 从map转换为CisMetadata数组
	var customFields []*cmdb.CisMetadata
	if v.CustomFields != nil {
		for key, value := range v.CustomFields {
			valueStr := fmt.Sprintf("%v", value)
			customFields = append(customFields, &cmdb.CisMetadata{
				Key:   &key,
				Value: &valueStr,
			})
		}
	}

	return &cmdb.CisInfo{
		Id:           &v.ID,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		TypeId:       &v.TypeID,
		Status:       &v.Status,
		CreatedBy:    createdBy,
		Tags:         tags,
		Metadata:     metadata,
		CustomFields: customFields,
		Attributes:   attributes,
	}
}

// CisCreateBuilderSetter 设置CI创建构建器的字段
func CisCreateBuilderSetter(builder *ent.CisCreate, in *cmdb.CisInfo) *ent.CisCreate {
	if in.TypeId != nil {
		builder = builder.SetTypeID(*in.TypeId)
	}
	if in.Status != nil {
		builder = builder.SetStatus(*in.Status)
	} else {
		builder = builder.SetStatus(1)
	}
	if in.CreatedBy != nil {
		builder = builder.SetCreatedBy(uuidx.ParseUUIDString(*in.CreatedBy))
	}
	if in.Tags != nil {
		tags := make([]schema.CiTag, 0)
		for _, tag := range in.Tags {
			parts := strings.Split(tag, ":")
			if len(parts) == 2 {
				tags = append(tags, schema.CiTag{Key: parts[0], Value: parts[1]})
			}
		}
		builder = builder.SetTags(tags)
	}

	// 处理metadata - 从CisMetadata数组转换为map
	if len(in.Metadata) > 0 {
		metadata := make(map[string]interface{})
		for _, meta := range in.Metadata {
			if meta.Key != nil && meta.Value != nil {
				metadata[*meta.Key] = *meta.Value
			}
		}
		builder = builder.SetMetadata(metadata)
	}

	// 处理custom_fields - 从CisMetadata数组转换为map
	if len(in.CustomFields) > 0 {
		customFields := make(map[string]interface{})
		for _, field := range in.CustomFields {
			if field.Key != nil && field.Value != nil {
				customFields[*field.Key] = *field.Value
			}
		}
		builder = builder.SetCustomFields(customFields)
	}

	return builder
}

// CisUpdateBuilderSetter 设置CI更新构建器的字段
func CisUpdateBuilderSetter(builder *ent.CisUpdateOne, in *cmdb.CisInfo) *ent.CisUpdateOne {
	if in.TypeId != nil {
		builder = builder.SetTypeID(*in.TypeId)
	}
	if in.Status != nil {
		builder = builder.SetStatus(*in.Status)
	}
	if in.CreatedBy != nil {
		builder = builder.SetCreatedBy(uuidx.ParseUUIDString(*in.CreatedBy))
	}
	if in.Tags != nil {
		tags := make([]schema.CiTag, 0)
		for _, tag := range in.Tags {
			parts := strings.Split(tag, ":")
			if len(parts) == 2 {
				tags = append(tags, schema.CiTag{Key: parts[0], Value: parts[1]})
			}
		}
		builder = builder.SetTags(tags)
	}

	// 处理metadata - 从CisMetadata数组转换为map
	if len(in.Metadata) > 0 {
		metadata := make(map[string]interface{})
		for _, meta := range in.Metadata {
			if meta.Key != nil && meta.Value != nil {
				metadata[*meta.Key] = *meta.Value
			}
		}
		builder = builder.SetMetadata(metadata)
	}

	// 处理custom_fields - 从CisMetadata数组转换为map
	if len(in.CustomFields) > 0 {
		customFields := make(map[string]interface{})
		for _, field := range in.CustomFields {
			if field.Key != nil && field.Value != nil {
				customFields[*field.Key] = *field.Value
			}
		}
		builder = builder.SetCustomFields(customFields)
	}

	return builder
}

// LoadCiAttributes 加载CI实例的属性值
func LoadCiAttributes(ctx context.Context, client *ent.Client, ciID uint64, includeAttrs, excludeAttrs []uint64) ([]*cmdb.CiAttributeValue, error) {
	var attributes []*cmdb.CiAttributeValue

	// 查询短文本类型属性值（ValueIndexText）
	indexTextValues, err := client.ValueIndexText.Query().
		Where(valueindextext.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range indexTextValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     v.Value,
			}
			// 设置原始值
			if rawValue := v.Value; rawValue != "" {
				attr.RawValue = &rawValue
			}
			attributes = append(attributes, attr)
		}
	}

	// 查询长文本类型属性值（ValueText）
	textValues, err := client.ValueText.Query().
		Where(valuetext.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range textValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     v.Value,
			}
			// 设置原始值
			if rawValue := v.Value; rawValue != "" {
				attr.RawValue = &rawValue
			}
			attributes = append(attributes, attr)
		}
	}

	// 查询整数类型属性值
	intValues, err := client.ValueInteger.Query().
		Where(valueinteger.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range intValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			valueStr := strconv.Itoa(v.Value)
			rawValue := strconv.Itoa(v.Value)
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     valueStr,
				RawValue:  &rawValue,
			}
			attributes = append(attributes, attr)
		}
	}

	// 查询浮点数类型属性值
	floatValues, err := client.ValueFloat.Query().
		Where(valuefloat.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range floatValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			valueStr := strconv.FormatFloat(v.Value, 'f', -1, 64)
			rawValue := strconv.FormatFloat(v.Value, 'f', -1, 64)
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     valueStr,
				RawValue:  &rawValue,
			}
			attributes = append(attributes, attr)
		}
	}

	// 查询日期时间类型属性值
	datetimeValues, err := client.ValueDatetime.Query().
		Where(valuedatetime.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range datetimeValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			valueStr := v.Value.Format(time.RFC3339)
			rawValue := strconv.FormatInt(v.Value.UnixMilli(), 10)
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     valueStr,
				RawValue:  &rawValue,
			}
			attributes = append(attributes, attr)
		}
	}

	// 查询JSON类型属性值
	jsonValues, err := client.ValueJSON.Query().
		Where(valuejson.CiID(ciID)).
		WithAttribute().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range jsonValues {
		if shouldIncludeAttribute(v.AttrID, includeAttrs, excludeAttrs) {
			valueStr := string(v.Value)
			rawValue := string(v.Value)
			attr := &cmdb.CiAttributeValue{
				AttrId:    v.AttrID,
				AttrName:  v.Edges.Attribute.Name,
				AttrAlias: v.Edges.Attribute.Alias,
				ValueType: string(v.Edges.Attribute.ValueType),
				Value:     valueStr,
				RawValue:  &rawValue,
			}
			attributes = append(attributes, attr)
		}
	}

	return attributes, nil
}

// shouldIncludeAttribute 判断是否应该包含该属性
func shouldIncludeAttribute(attrID uint64, includeAttrs, excludeAttrs []uint64) bool {
	// 如果在排除列表中，则不包含
	for _, excludeID := range excludeAttrs {
		if attrID == excludeID {
			return false
		}
	}

	// 如果包含列表为空，则包含所有（除了排除的）
	if len(includeAttrs) == 0 {
		return true
	}

	// 如果在包含列表中，则包含
	for _, includeID := range includeAttrs {
		if attrID == includeID {
			return true
		}
	}

	return false
}

// SaveCiAttributes 保存CI实例的属性值
func SaveCiAttributes(ctx context.Context, tx *ent.Tx, ciID uint64, attributes []*cmdb.CiAttributeValue) error {
	for _, attr := range attributes {
		if attr == nil {
			continue
		}

		attrID := attr.AttrId
		valueType := attr.ValueType
		value := attr.Value

		switch valueType {
		case consts.ValueTypeShortText:
			// 短文本使用ValueIndexText表（支持索引）
			// 删除现有值
			_, err := tx.ValueIndexText.Delete().
				Where(valueindextext.CiID(ciID), valueindextext.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				_, err = tx.ValueIndexText.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue(value).
					Save(ctx)
				if err != nil {
					return err
				}
			}

		case consts.ValueTypeLongText, consts.ValueTypeLink:
			// 长文本使用ValueText表（不支持索引）
			// 删除现有值
			_, err := tx.ValueText.Delete().
				Where(valuetext.CiID(ciID), valuetext.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				_, err = tx.ValueText.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue(value).
					Save(ctx)
				if err != nil {
					return err
				}
			}

		case consts.ValueTypeInt:
			// 删除现有值
			_, err := tx.ValueInteger.Delete().
				Where(valueinteger.CiID(ciID), valueinteger.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid integer value for attr %d: %s", attrID, value)
				}
				_, err = tx.ValueInteger.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue(intValue).
					Save(ctx)
				if err != nil {
					return err
				}
			}

		case consts.ValueTypeFloat:
			// 删除现有值
			_, err := tx.ValueFloat.Delete().
				Where(valuefloat.CiID(ciID), valuefloat.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				floatValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("invalid float value for attr %d: %s", attrID, value)
				}
				_, err = tx.ValueFloat.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue(floatValue).
					Save(ctx)
				if err != nil {
					return err
				}
			}

		case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
			// 删除现有值
			_, err := tx.ValueDatetime.Delete().
				Where(valuedatetime.CiID(ciID), valuedatetime.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				var timeValue time.Time
				var err error

				// 尝试解析不同的时间格式
				if timestamp, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil {
					timeValue = time.UnixMilli(timestamp)
				} else if timeValue, err = time.Parse(time.RFC3339, value); err != nil {
					if timeValue, err = time.Parse("2006-01-02 15:04:05", value); err != nil {
						return fmt.Errorf("invalid datetime value for attr %d: %s", attrID, value)
					}
				}

				_, err = tx.ValueDatetime.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue(timeValue).
					Save(ctx)
				if err != nil {
					return err
				}
			}

		default:
			// JSON类型或其他类型
			_, err := tx.ValueJSON.Delete().
				Where(valuejson.CiID(ciID), valuejson.AttrID(attrID)).
				Exec(ctx)
			if err != nil {
				return err
			}

			// 创建新值
			if value != "" {
				// 验证JSON格式
				var jsonData interface{}
				if err := json.Unmarshal([]byte(value), &jsonData); err != nil {
					// 如果不是有效JSON，则包装为字符串
					jsonBytes, _ := json.Marshal(value)
					value = string(jsonBytes)
				}

				_, err = tx.ValueJSON.Create().
					SetCiID(ciID).
					SetAttrID(attrID).
					SetValue([]byte(value)).
					Save(ctx)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ConvertAPIFilterToProto 将API过滤条件转换为Proto格式
func ConvertAPIFilterToProto(apiFilter interface{}) (string, error) {
	jsonBytes, err := json.Marshal(apiFilter)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ConvertProtoFilterToAPI 将Proto过滤条件转换为API格式
func ConvertProtoFilterToAPI(protoFilter string, target interface{}) error {
	if protoFilter == "" {
		return nil
	}
	return json.Unmarshal([]byte(protoFilter), target)
}

// ValidateUniqueAttributes 校验CI属性的唯一性
// excludeCiID: 排除的CI ID（用于更新时排除自身）
func ValidateUniqueAttributes(ctx context.Context, client *ent.Client, typeID uint64, attributes []*cmdb.CiAttributeValue, excludeCiID *uint64) error {
	if len(attributes) == 0 {
		return nil
	}

	// 获取所有属性ID
	var attrIDs []uint64
	attrValueMap := make(map[uint64]*cmdb.CiAttributeValue)
	for _, attr := range attributes {
		if attr != nil && attr.Value != "" {
			attrIDs = append(attrIDs, attr.AttrId)
			attrValueMap[attr.AttrId] = attr
		}
	}

	if len(attrIDs) == 0 {
		return nil
	}

	// 查询需要唯一性校验的属性，从CiTypeAttribute表查询
	uniqueTypeAttrs, err := client.CiTypeAttribute.Query().
		Where(
			citypeattribute.TypeIDEQ(typeID),
			citypeattribute.AttrIDIn(attrIDs...),
			citypeattribute.IsUniqueEQ(true),
		).
		WithAttribute().
		All(ctx)
	if err != nil {
		return fmt.Errorf("查询唯一属性失败: %w", err)
	}

	if len(uniqueTypeAttrs) == 0 {
		return nil
	}

	// 对每个唯一属性进行校验
	for _, uniqueTypeAttr := range uniqueTypeAttrs {
		if uniqueTypeAttr.Edges.Attribute == nil {
			continue
		}

		attrValue := attrValueMap[uniqueTypeAttr.AttrID]
		if attrValue == nil || attrValue.Value == "" {
			continue
		}

		// 根据属性值类型进行校验
		err := validateUniqueAttributeValue(ctx, client, typeID, uniqueTypeAttr.Edges.Attribute, attrValue.Value, excludeCiID)
		if err != nil {
			return fmt.Errorf("属性 '%s' 唯一性校验失败: %w", uniqueTypeAttr.Edges.Attribute.Alias, err)
		}
	}

	return nil
}

// validateUniqueAttributeValue 校验单个属性值的唯一性
func validateUniqueAttributeValue(ctx context.Context, client *ent.Client, typeID uint64, attr *ent.Attribute, value string, excludeCiID *uint64) error {
	switch attr.ValueType {
	case consts.ValueTypeShortText:
		return validateUniqueIndexTextValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	case consts.ValueTypeLongText, consts.ValueTypeLink:
		return validateUniqueTextValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	case consts.ValueTypeInt:
		return validateUniqueIntegerValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	case consts.ValueTypeFloat:
		return validateUniqueFloatValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
		return validateUniqueDatetimeValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	default:
		return validateUniqueJSONValue(ctx, client, typeID, attr.ID, value, excludeCiID)
	}
}

// validateUniqueIndexTextValue 校验短文本类型属性值的唯一性
func validateUniqueIndexTextValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	query := client.ValueIndexText.Query().
		Where(
			valueindextext.AttrIDEQ(attrID),
			valueindextext.ValueEQ(value),
		).
		WithCi(func(q *ent.CisQuery) {
			q.Where(cis.TypeIDEQ(typeID))
			if excludeCiID != nil {
				q.Where(cis.IDNEQ(*excludeCiID))
			}
		})

	exists, err := query.Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询短文本属性值失败: %w", err)
	}

	if exists {
		// 如果是更新操作，则不返回错误
		if excludeCiID != nil {
			return nil
		}
		return fmt.Errorf("属性值 '%s' 在当前CI类型下已存在", value)
	}

	return nil
}

// validateUniqueTextValue 校验长文本类型属性值的唯一性
func validateUniqueTextValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	query := client.ValueText.Query().
		Where(
			valuetext.AttrIDEQ(attrID),
			valuetext.ValueEQ(value),
		).
		WithCi(func(q *ent.CisQuery) {
			q.Where(cis.TypeIDEQ(typeID))
			if excludeCiID != nil {
				q.Where(cis.IDNEQ(*excludeCiID))
			}
		})

	exists, err := query.Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询文本属性值失败: %w", err)
	}

	if exists {
		// 如果是更新操作，则不返回错误
		if excludeCiID != nil {
			return nil
		}
		return fmt.Errorf("属性值 '%s' 在当前CI类型下已存在", value)
	}

	return nil
}

// validateUniqueIntegerValue 校验整数类型属性值的唯一性
func validateUniqueIntegerValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("无效的整数值: %s", value)
	}

	query := client.ValueInteger.Query().
		Where(
			valueinteger.AttrIDEQ(attrID),
			valueinteger.ValueEQ(intValue),
		).
		WithCi(func(q *ent.CisQuery) {
			q.Where(cis.TypeIDEQ(typeID))
			if excludeCiID != nil {
				q.Where(cis.IDNEQ(*excludeCiID))
			}
		})

	exists, err := query.Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询整数属性值失败: %w", err)
	}
	if exists {
		// 如果是更新操作，则不返回错误
		if excludeCiID != nil {
			return nil
		}
		return fmt.Errorf("属性值 '%s' 在当前CI类型下已存在", value)
	}

	return nil
}

// validateUniqueFloatValue 校验浮点数类型属性值的唯一性
func validateUniqueFloatValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("无效的浮点数值: %s", value)
	}

	query := client.ValueFloat.Query().
		Where(
			valuefloat.AttrIDEQ(attrID),
			valuefloat.ValueEQ(floatValue),
		).
		WithCi(func(q *ent.CisQuery) {
			q.Where(cis.TypeIDEQ(typeID))
			if excludeCiID != nil {
				q.Where(cis.IDNEQ(*excludeCiID))
			}
		})

	exists, err := query.Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询浮点数属性值失败: %w", err)
	}

	if exists {
		// 如果是更新操作，则不返回错误
		if excludeCiID != nil {
			return nil
		}
		return fmt.Errorf("属性值 '%s' 在当前CI类型下已存在", value)
	}

	return nil
}

// validateUniqueDatetimeValue 校验日期时间类型属性值的唯一性
func validateUniqueDatetimeValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	var timeValue time.Time
	var err error

	// 尝试解析不同的时间格式
	if timestamp, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil {
		timeValue = time.UnixMilli(timestamp)
	} else if timeValue, err = time.Parse(time.RFC3339, value); err != nil {
		if timeValue, err = time.Parse("2006-01-02 15:04:05", value); err != nil {
			return fmt.Errorf("无效的日期时间值: %s", value)
		}
	}

	query := client.ValueDatetime.Query().
		Where(
			valuedatetime.AttrIDEQ(attrID),
			valuedatetime.ValueEQ(timeValue),
		).
		WithCi(func(q *ent.CisQuery) {
			q.Where(cis.TypeIDEQ(typeID))
			if excludeCiID != nil {
				q.Where(cis.IDNEQ(*excludeCiID))
			}
		})

	exists, err := query.Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询日期时间属性值失败: %w", err)
	}

	if exists {
		// 如果是更新操作，则不返回错误
		if excludeCiID != nil {
			return nil
		}
		return fmt.Errorf("属性值 '%s' 在当前CI类型下已存在", value)
	}

	return nil
}

// validateUniqueJSONValue 校验JSON类型属性值的唯一性
// 注意：JSON类型字段不支持精确匹配查询，因此跳过唯一性校验
func validateUniqueJSONValue(ctx context.Context, client *ent.Client, typeID uint64, attrID uint64, value string, excludeCiID *uint64) error {
	// JSON类型字段通常包含复杂的结构化数据，不适合进行唯一性校验
	// 如果确实需要JSON字段的唯一性校验，建议：
	// 1. 将关键字段提取为独立的文本/数值属性
	// 2. 使用原始SQL查询进行比较
	// 3. 在应用层进行校验

	// 这里我们跳过JSON类型的唯一性校验
	return nil
}
