package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/schema"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

// ProcessCiRelations 处理CI关系数据（创建/更新）
func ProcessCiRelations(ctx context.Context, tx *ent.Tx, ciId uint64, relationsData *cmdb.CiRelationsData, logger logx.Logger) error {
	if relationsData == nil {
		return nil
	}

	// 处理要删除的关系
	if len(relationsData.DeleteRelationIds) > 0 {
		_, err := tx.CiRelation.Delete().Where(
			cirelation.And(
				cirelation.IDIn(relationsData.DeleteRelationIds...),
				cirelation.Or(
					cirelation.SourceCiIDEQ(ciId),
					cirelation.TargetCiIDEQ(ciId),
				),
			),
		).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete relations: %w", err)
		}
	}

	// 处理要创建的关系
	for _, createInfo := range relationsData.CreateRelations {
		err := CreateCiRelation(ctx, tx, ciId, createInfo, logger)
		if err != nil {
			if _, ok := err.(*errorx.CodeError); ok {
				return err
			}
			return fmt.Errorf("failed to create relation: %w", err)
		}
	}

	// 处理要更新的关系
	for _, updateInfo := range relationsData.UpdateRelations {
		err := UpdateCiRelation(ctx, tx, updateInfo, logger)
		if err != nil {
			if _, ok := err.(*errorx.CodeError); ok {
				return err
			}
			return fmt.Errorf("failed to update relation: %w", err)
		}
	}

	return nil
}

// CreateCiRelation 创建CI关系
func CreateCiRelation(ctx context.Context, tx *ent.Tx, ciId uint64, createInfo *cmdb.CiRelationCreateInfo, logger logx.Logger) error {
	// 确定源和目标CI ID
	var sourceCiId, targetCiId uint64
	switch createInfo.Direction {
	case "source_to_target":
		sourceCiId = ciId
		targetCiId = createInfo.TargetCiId
	case "target_to_source":
		sourceCiId = createInfo.TargetCiId
		targetCiId = ciId
	default:
		return fmt.Errorf("invalid relation direction: %s", createInfo.Direction)
	}

	// 创建验证器并验证关系约束
	validator := NewRelationValidator(ctx, tx.Client(), logger)
	validationResult, err := validator.ValidateRelationCreation(sourceCiId, targetCiId, createInfo.RelationTypeId)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !validationResult.IsValid {
		message := strings.Join(validationResult.Errors, "; ")
		if strings.Contains(strings.ToLower(message), "no relation definition") {
			message = "未找到匹配的模型关系，请先在模型关系中配置该关系类型"
		}
		return errorx.NewInvalidArgumentError(message)
	}

	// 记录警告信息
	if len(validationResult.Warnings) > 0 {
		logger.Infof("Relation creation warnings: %s", strings.Join(validationResult.Warnings, "; "))
	}

	// 创建关系构建器并设置源和目标CI
	builder := tx.CiRelation.Create().SetSourceCiID(sourceCiId).SetTargetCiID(targetCiId)

	// 设置关系类型
	builder = builder.SetRelationTypeID(createInfo.RelationTypeId)

	// 设置可选字段
	if createInfo.MoreCiId != nil {
		builder = builder.SetMore(*createInfo.MoreCiId)
	}
	if createInfo.DiscoverySource != nil {
		builder = builder.SetDiscoverySource(*createInfo.DiscoverySource)
	}
	if createInfo.Properties != nil {
		// 将属性JSON字符串转换为map
		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(*createInfo.Properties), &properties); err != nil {
			return fmt.Errorf("invalid properties JSON: %w", err)
		}
		builder = builder.SetProperties(properties)
	}
	if createInfo.Status != nil {
		builder = builder.SetStatus(*createInfo.Status)
	} else {
		builder = builder.SetStatus("active") // 默认状态
	}
	if createInfo.RelationStrength != nil {
		builder = builder.SetRelationStrength(*createInfo.RelationStrength)
	} else {
		builder = builder.SetRelationStrength("normal") // 默认强度
	}

	// 设置属性映射数据
	if len(createInfo.AttributeMappings) > 0 {
		// 转换为ent需要的格式
		var mappings []schema.AttributeMappingData
		for _, mapping := range createInfo.AttributeMappings {
			entMapping := schema.AttributeMappingData{
				SourceAttrID:   mapping.SourceAttrId,
				TargetAttrID:   mapping.TargetAttrId,
				SyncStatus:     *mapping.SyncStatus,
				ConflictReason: *mapping.ConflictReason,
			}
			if mapping.SourceValue != nil {
				entMapping.SourceValue = *mapping.SourceValue
			}
			if mapping.TargetValue != nil {
				entMapping.TargetValue = *mapping.TargetValue
			}
			if mapping.LastSyncAt != nil {
				syncTime := time.UnixMilli(*mapping.LastSyncAt)
				entMapping.LastSyncAt = &syncTime
			}
			mappings = append(mappings, entMapping)
		}
		builder = builder.SetAttributeMappings(mappings)
	}

	// 创建关系
	_, err = builder.Save(ctx)
	return err
}

// UpdateCiRelation 更新CI关系
func UpdateCiRelation(ctx context.Context, tx *ent.Tx, updateInfo *cmdb.CiRelationUpdateInfo, logger logx.Logger) error {
	// 验证关系更新约束
	validator := NewRelationValidator(ctx, tx.Client(), logger)
	validationResult, err := validator.ValidateRelationUpdate(updateInfo.RelationId, updateInfo)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !validationResult.IsValid {
		message := strings.Join(validationResult.Errors, "; ")
		if strings.Contains(strings.ToLower(message), "no relation definition") {
			message = "未找到匹配的模型关系，请先在模型关系中配置该关系类型"
		}
		return errorx.NewInvalidArgumentError(message)
	}

	// 记录警告信息
	if len(validationResult.Warnings) > 0 {
		logger.Infof("Relation update warnings: %s", strings.Join(validationResult.Warnings, "; "))
	}
	builder := tx.CiRelation.UpdateOneID(updateInfo.RelationId)

	// 更新可选字段
	if updateInfo.TargetCiId != nil {
		builder = builder.SetTargetCiID(*updateInfo.TargetCiId)
	}
	if updateInfo.RelationTypeId != nil {
		builder = builder.SetRelationTypeID(*updateInfo.RelationTypeId)
	}
	if updateInfo.MoreCiId != nil {
		builder = builder.SetMore(*updateInfo.MoreCiId)
	}
	if updateInfo.Properties != nil {
		// 将属性JSON字符串转换为map
		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(*updateInfo.Properties), &properties); err != nil {
			return fmt.Errorf("invalid properties JSON: %w", err)
		}
		builder = builder.SetProperties(properties)
	}
	if updateInfo.Status != nil {
		builder = builder.SetStatus(*updateInfo.Status)
	}
	if updateInfo.RelationStrength != nil {
		builder = builder.SetRelationStrength(*updateInfo.RelationStrength)
	}

	// 更新属性映射数据
	if len(updateInfo.AttributeMappings) > 0 {
		// 转换为ent需要的格式
		var mappings []schema.AttributeMappingData
		for _, mapping := range updateInfo.AttributeMappings {
			entMapping := schema.AttributeMappingData{
				SourceAttrID:   mapping.SourceAttrId,
				TargetAttrID:   mapping.TargetAttrId,
				SyncStatus:     *mapping.SyncStatus,
				ConflictReason: *mapping.ConflictReason,
			}
			if mapping.SourceValue != nil {
				entMapping.SourceValue = *mapping.SourceValue
			}
			if mapping.TargetValue != nil {
				entMapping.TargetValue = *mapping.TargetValue
			}
			if mapping.LastSyncAt != nil {
				syncTime := time.UnixMilli(*mapping.LastSyncAt)
				entMapping.LastSyncAt = &syncTime
			}
			mappings = append(mappings, entMapping)
		}
		builder = builder.SetAttributeMappings(mappings)
	}

	// 更新关系
	_, err = builder.Save(ctx)
	return err
}

// LoadCiRelations 加载CI的关系数据（用于查询返回）
func LoadCiRelations(ctx context.Context, db *ent.Client, ciId uint64) (*cmdb.CiRelationQueryResult, error) {
	// 查询作为源CI的关系
	sourceRelations, err := db.CiRelation.Query().
		Where(cirelation.SourceCiIDEQ(ciId)).
		WithTargetCi().     // 预加载目标CI
		WithRelationType(). // 预加载关系类型
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load source relations: %w", err)
	}

	// 查询作为目标CI的关系
	targetRelations, err := db.CiRelation.Query().
		Where(cirelation.TargetCiIDEQ(ciId)).
		WithSourceCi().     // 预加载源CI
		WithRelationType(). // 预加载关系类型
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load target relations: %w", err)
	}

	result := &cmdb.CiRelationQueryResult{
		SourceRelations: []*cmdb.CiRelationInfo{},
		TargetRelations: []*cmdb.CiRelationInfo{},
		TotalCount:      uint64(len(sourceRelations) + len(targetRelations)),
	}

	// 转换源关系数据
	for _, relation := range sourceRelations {
		relationInfo := convertEntToCiRelationInfo(relation)
		result.SourceRelations = append(result.SourceRelations, relationInfo)
	}

	// 转换目标关系数据
	for _, relation := range targetRelations {
		relationInfo := convertEntToCiRelationInfo(relation)
		result.TargetRelations = append(result.TargetRelations, relationInfo)
	}

	return result, nil
}

// convertEntToCiRelationInfo 将ent关系模型转换为proto格式
func convertEntToCiRelationInfo(relation *ent.CiRelation) *cmdb.CiRelationInfo {
	info := &cmdb.CiRelationInfo{
		Id:             &relation.ID,
		CreatedAt:      pointy.GetPointer(relation.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(relation.UpdatedAt.UnixMilli()),
		SourceCiId:     &relation.SourceCiID,
		TargetCiId:     &relation.TargetCiID,
		RelationTypeId: &relation.RelationTypeID,
	}

	// 设置可选字段
	if relation.More != 0 {
		info.More = &relation.More
	}
	if relation.DiscoverySource != "" {
		info.DiscoverySource = &relation.DiscoverySource
	}
	if relation.AncestorIds != "" {
		info.AncestorIds = &relation.AncestorIds
	}
	if relation.Status != "" {
		info.Status = &relation.Status
	}
	if relation.RelationStrength != "" {
		info.RelationStrength = &relation.RelationStrength
	}
	if relation.AutoSyncEnabled {
		info.AutoSyncEnabled = &relation.AutoSyncEnabled
	}

	// 处理JSON字段（属性映射、验证结果、同步配置等）
	// 这里可以根据需要进行序列化

	return info
}
