package cis

import (
	"context"
	"fmt"
	"strings"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cityperelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

// RelationValidator 关系约束验证器
type RelationValidator struct {
	ctx    context.Context
	db     *ent.Client
	logger logx.Logger
}

// NewRelationValidator 创建关系验证器
func NewRelationValidator(ctx context.Context, db *ent.Client, logger logx.Logger) *RelationValidator {
	return &RelationValidator{
		ctx:    ctx,
		db:     db,
		logger: logger,
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid      bool     `json:"isValid"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	ConstraintId uint64   `json:"constraintId,omitempty"`
}

// ValidateRelationCreation 验证关系创建
func (rv *RelationValidator) ValidateRelationCreation(sourceCiId, targetCiId, relationTypeId uint64) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 1. 获取源CI和目标CI的类型
	sourceCi, err := rv.db.Cis.Query().Where(cis.IDEQ(sourceCiId)).WithCiType().Only(rv.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get source CI: %w", err)
	}

	targetCi, err := rv.db.Cis.Query().Where(cis.IDEQ(targetCiId)).WithCiType().Only(rv.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get target CI: %w", err)
	}

	sourceCiTypeId := sourceCi.TypeID
	targetCiTypeId := targetCi.TypeID

	// 2. 查找对应的CI类型关系定义
	ciTypeRelations, err := rv.db.CiTypeRelation.Query().Where(
		cityperelation.And(
			cityperelation.ParentIDEQ(sourceCiTypeId),
			cityperelation.ChildIDEQ(targetCiTypeId),
			cityperelation.RelationTypeIDEQ(relationTypeId),
		),
	).All(rv.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query CI type relations: %w", err)
	}

	if len(ciTypeRelations) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, 
			fmt.Sprintf("No relation definition found between CI types %d and %d for relation type %d", 
				sourceCiTypeId, targetCiTypeId, relationTypeId))
		return result, nil
	}

	// 3. 验证约束条件
	ciTypeRelation := ciTypeRelations[0]
	result.ConstraintId = ciTypeRelation.ID

	if ciTypeRelation.Constraint != "" {
		constraintErr := rv.validateConstraint(ciTypeRelation.Constraint, sourceCiId, targetCiId, relationTypeId)
		if constraintErr != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, constraintErr.Error())
		}
	}

	// 4. 验证属性映射完整性
	if len(ciTypeRelation.ParentAttrIds) > 0 || len(ciTypeRelation.ChildAttrIds) > 0 {
		mappingWarnings := rv.validateAttributeMapping(ciTypeRelation, sourceCi, targetCi)
		result.Warnings = append(result.Warnings, mappingWarnings...)
	}

	// 5. 检查重复关系
	duplicateErr := rv.checkDuplicateRelation(sourceCiId, targetCiId, relationTypeId)
	if duplicateErr != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, duplicateErr.Error())
	}

	return result, nil
}

// ValidateRelationUpdate 验证关系更新
func (rv *RelationValidator) ValidateRelationUpdate(relationId uint64, updateInfo *cmdb.CiRelationUpdateInfo) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 获取现有关系
	existingRelation, err := rv.db.CiRelation.Query().Where(cirelation.IDEQ(relationId)).Only(rv.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing relation: %w", err)
	}

	// 确定更新后的目标CI和关系类型
	targetCiId := existingRelation.TargetCiID
	relationTypeId := existingRelation.RelationTypeID

	if updateInfo.TargetCiId != nil {
		targetCiId = *updateInfo.TargetCiId
	}
	if updateInfo.RelationTypeId != nil {
		relationTypeId = *updateInfo.RelationTypeId
	}

	// 如果目标CI或关系类型发生变化，需要重新验证约束
	if updateInfo.TargetCiId != nil || updateInfo.RelationTypeId != nil {
		validationResult, err := rv.ValidateRelationCreation(existingRelation.SourceCiID, targetCiId, relationTypeId)
		if err != nil {
			return nil, err
		}
		
		// 检查是否会产生重复关系（排除当前关系）
		if validationResult.IsValid {
			duplicateErr := rv.checkDuplicateRelationExcluding(existingRelation.SourceCiID, targetCiId, relationTypeId, relationId)
			if duplicateErr != nil {
				result.IsValid = false
				result.Errors = append(result.Errors, duplicateErr.Error())
			}
		} else {
			result = validationResult
		}
	}

	return result, nil
}

// validateConstraint 验证约束条件
func (rv *RelationValidator) validateConstraint(constraint string, sourceCiId, targetCiId, relationTypeId uint64) error {
	switch strings.ToLower(constraint) {
	case "one_to_one":
		return rv.validateOneToOne(sourceCiId, targetCiId, relationTypeId)
	case "one_to_many":
		return rv.validateOneToMany(sourceCiId, targetCiId, relationTypeId)
	case "many_to_one":
		return rv.validateManyToOne(sourceCiId, targetCiId, relationTypeId)
	case "many_to_many":
		// many_to_many 通常没有特殊约束
		return nil
	default:
		rv.logger.Errorf("Unknown constraint type: %s", constraint)
		return nil
	}
}

// validateOneToOne 验证一对一约束
func (rv *RelationValidator) validateOneToOne(sourceCiId, targetCiId, relationTypeId uint64) error {
	// 检查源CI是否已经有这种类型的关系
	sourceCount, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.SourceCiIDEQ(sourceCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to count source relations: %w", err)
	}
	if sourceCount > 0 {
		return fmt.Errorf("one-to-one constraint violated: source CI %d already has a relation of type %d", sourceCiId, relationTypeId)
	}

	// 检查目标CI是否已经有这种类型的关系
	targetCount, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.TargetCiIDEQ(targetCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to count target relations: %w", err)
	}
	if targetCount > 0 {
		return fmt.Errorf("one-to-one constraint violated: target CI %d already has a relation of type %d", targetCiId, relationTypeId)
	}

	return nil
}

// validateOneToMany 验证一对多约束
func (rv *RelationValidator) validateOneToMany(sourceCiId, targetCiId, relationTypeId uint64) error {
	// 检查目标CI是否已经有这种类型的入关系
	targetCount, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.TargetCiIDEQ(targetCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to count target relations: %w", err)
	}
	if targetCount > 0 {
		return fmt.Errorf("one-to-many constraint violated: target CI %d already has a relation of type %d", targetCiId, relationTypeId)
	}

	return nil
}

// validateManyToOne 验证多对一约束
func (rv *RelationValidator) validateManyToOne(sourceCiId, targetCiId, relationTypeId uint64) error {
	// 检查源CI是否已经有这种类型的出关系
	sourceCount, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.SourceCiIDEQ(sourceCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to count source relations: %w", err)
	}
	if sourceCount > 0 {
		return fmt.Errorf("many-to-one constraint violated: source CI %d already has a relation of type %d", sourceCiId, relationTypeId)
	}

	return nil
}

// validateAttributeMapping 验证属性映射完整性
func (rv *RelationValidator) validateAttributeMapping(ciTypeRelation *ent.CiTypeRelation, sourceCi, targetCi *ent.Cis) []string {
	var warnings []string

	// 检查单属性映射
	if ciTypeRelation.ParentAttrID != 0 && ciTypeRelation.ChildAttrID != 0 {
		// 这里可以验证属性是否存在、类型是否兼容等
		warnings = append(warnings, "Single attribute mapping defined - ensure attributes are compatible")
	}

	// 检查多属性映射
	if len(ciTypeRelation.ParentAttrIds) > 0 && len(ciTypeRelation.ChildAttrIds) > 0 {
		if len(ciTypeRelation.ParentAttrIds) != len(ciTypeRelation.ChildAttrIds) {
			warnings = append(warnings, "Attribute mapping arrays have different lengths")
		}
		warnings = append(warnings, "Multiple attribute mappings defined - ensure all attributes are compatible")
	}

	return warnings
}

// checkDuplicateRelation 检查重复关系
func (rv *RelationValidator) checkDuplicateRelation(sourceCiId, targetCiId, relationTypeId uint64) error {
	count, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.SourceCiIDEQ(sourceCiId),
			cirelation.TargetCiIDEQ(targetCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to check duplicate relations: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("duplicate relation: relation already exists between CI %d and CI %d with type %d", sourceCiId, targetCiId, relationTypeId)
	}
	return nil
}

// checkDuplicateRelationExcluding 检查重复关系（排除指定关系）
func (rv *RelationValidator) checkDuplicateRelationExcluding(sourceCiId, targetCiId, relationTypeId, excludeRelationId uint64) error {
	count, err := rv.db.CiRelation.Query().Where(
		cirelation.And(
			cirelation.SourceCiIDEQ(sourceCiId),
			cirelation.TargetCiIDEQ(targetCiId),
			cirelation.RelationTypeIDEQ(relationTypeId),
			cirelation.IDNEQ(excludeRelationId),
		),
	).Count(rv.ctx)
	if err != nil {
		return fmt.Errorf("failed to check duplicate relations: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("duplicate relation: relation already exists between CI %d and CI %d with type %d", sourceCiId, targetCiId, relationTypeId)
	}
	return nil
}

// ValidateBatchRelations 批量验证关系
func (rv *RelationValidator) ValidateBatchRelations(relations []*cmdb.CiRelationCreateInfo, sourceCiId uint64) ([]*ValidationResult, error) {
	results := make([]*ValidationResult, len(relations))
	
	for i, relation := range relations {
		var targetCiId uint64
		switch relation.Direction {
		case "source_to_target":
			targetCiId = relation.TargetCiId
		case "target_to_source":
			targetCiId = sourceCiId
			sourceCiId = relation.TargetCiId
		default:
			results[i] = &ValidationResult{
				IsValid: false,
				Errors:  []string{fmt.Sprintf("invalid direction: %s", relation.Direction)},
			}
			continue
		}

		result, err := rv.ValidateRelationCreation(sourceCiId, targetCiId, relation.RelationTypeId)
		if err != nil {
			return nil, fmt.Errorf("failed to validate relation %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// GetConstraintDescription 获取约束描述
func (rv *RelationValidator) GetConstraintDescription(constraint string) string {
	switch strings.ToLower(constraint) {
	case "one_to_one":
		return "一对一关系：每个CI只能有一个此类型的关系"
	case "one_to_many":
		return "一对多关系：源CI可以有多个目标CI，但目标CI只能有一个源CI"
	case "many_to_one":
		return "多对一关系：多个源CI可以指向同一个目标CI，但源CI只能有一个此类型关系"
	case "many_to_many":
		return "多对多关系：没有数量限制"
	default:
		return "未知约束类型"
	}
}