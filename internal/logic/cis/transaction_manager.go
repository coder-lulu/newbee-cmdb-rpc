package cis

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

// TransactionManager 事务管理器，处理复杂的CI和关系操作
type TransactionManager struct {
	ctx    context.Context
	db     *ent.Client
	logger logx.Logger
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(ctx context.Context, db *ent.Client, logger logx.Logger) *TransactionManager {
	return &TransactionManager{
		ctx:    ctx,
		db:     db,
		logger: logger,
	}
}

// CreateCiWithRelationsResult 创建CI和关系的结果
type CreateCiWithRelationsResult struct {
	CiID              uint64    `json:"ciId"`
	CreatedRelations  []uint64  `json:"createdRelations"`
	FailedRelations   []string  `json:"failedRelations"`
	ValidationErrors  []string  `json:"validationErrors"`
}

// CreateCiWithRelations 事务性创建CI和关系
func (tm *TransactionManager) CreateCiWithRelations(cisInfo *cmdb.CisInfo) (*CreateCiWithRelationsResult, error) {
	// 开启事务
	tx, err := tm.db.Tx(tm.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	result := &CreateCiWithRelationsResult{
		CreatedRelations: []uint64{},
		FailedRelations:  []string{},
		ValidationErrors: []string{},
	}

	// 1. 预验证阶段 - 验证所有数据完整性
	if err := tm.validateCiData(cisInfo); err != nil {
		tx.Rollback()
		result.ValidationErrors = append(result.ValidationErrors, err.Error())
		return result, fmt.Errorf("validation failed: %w", err)
	}

	// 2. 验证关系数据完整性
	if cisInfo.Relations != nil {
		if err := tm.validateRelationsData(cisInfo.Relations); err != nil {
			tx.Rollback()
			result.ValidationErrors = append(result.ValidationErrors, err.Error())
			return result, fmt.Errorf("relation validation failed: %w", err)
		}
	}

	// 3. 校验属性唯一性
	if len(cisInfo.Attributes) > 0 && cisInfo.TypeId != nil {
		err = ValidateUniqueAttributes(tm.ctx, tm.db, *cisInfo.TypeId, cisInfo.Attributes, nil)
		if err != nil {
			tx.Rollback()
			result.ValidationErrors = append(result.ValidationErrors, err.Error())
			return result, fmt.Errorf("attribute validation failed: %w", err)
		}
	}

	// 4. 创建CI实例
	cisBuilder := tx.Cis.Create()
	cisBuilder = CisCreateBuilderSetter(cisBuilder, cisInfo)

	ciResult, err := cisBuilder.Save(tm.ctx)
	if err != nil {
		tx.Rollback()
		return result, fmt.Errorf("failed to create CI: %w", err)
	}
	result.CiID = ciResult.ID

	// 5. 保存动态属性值
	if len(cisInfo.Attributes) > 0 {
		err = SaveCiAttributes(tm.ctx, tx, ciResult.ID, cisInfo.Attributes)
		if err != nil {
			tx.Rollback()
			return result, fmt.Errorf("failed to save CI attributes: %w", err)
		}
	}

	// 6. 创建关系（包含详细的错误处理）
	if cisInfo.Relations != nil && len(cisInfo.Relations.CreateRelations) > 0 {
		createdRelations, failedRelations := tm.createRelationsWithErrorHandling(tx, ciResult.ID, cisInfo.Relations.CreateRelations)
		result.CreatedRelations = createdRelations
		result.FailedRelations = failedRelations

		// 如果所有关系都失败了，可以选择回滚事务
		if len(createdRelations) == 0 && len(cisInfo.Relations.CreateRelations) > 0 {
			tx.Rollback()
			return result, fmt.Errorf("all relations failed to create")
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tm.logger.Infof("Successfully created CI %d with %d relations", result.CiID, len(result.CreatedRelations))
	return result, nil
}

// UpdateCiWithRelationsResult 更新CI和关系的结果
type UpdateCiWithRelationsResult struct {
	CiID              uint64    `json:"ciId"`
	UpdatedRelations  []uint64  `json:"updatedRelations"`
	CreatedRelations  []uint64  `json:"createdRelations"`
	DeletedRelations  []uint64  `json:"deletedRelations"`
	FailedOperations  []string  `json:"failedOperations"`
	ValidationErrors  []string  `json:"validationErrors"`
}

// UpdateCiWithRelations 事务性更新CI和关系
func (tm *TransactionManager) UpdateCiWithRelations(cisInfo *cmdb.CisInfo) (*UpdateCiWithRelationsResult, error) {
	if cisInfo.Id == nil {
		return nil, fmt.Errorf("CI ID is required for update")
	}

	// 开启事务
	tx, err := tm.db.Tx(tm.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	result := &UpdateCiWithRelationsResult{
		CiID:             *cisInfo.Id,
		UpdatedRelations: []uint64{},
		CreatedRelations: []uint64{},
		DeletedRelations: []uint64{},
		FailedOperations: []string{},
		ValidationErrors: []string{},
	}

	// 1. 验证CI存在性
	exists, err := tx.Cis.Query().Where(cis.IDEQ(*cisInfo.Id)).Exist(tm.ctx)
	if err != nil {
		tx.Rollback()
		return result, fmt.Errorf("failed to check CI existence: %w", err)
	}
	if !exists {
		tx.Rollback()
		return result, fmt.Errorf("CI with ID %d not found", *cisInfo.Id)
	}

	// 2. 获取当前CI的类型ID用于唯一性校验
	var typeID uint64
	if cisInfo.TypeId != nil {
		typeID = *cisInfo.TypeId
	} else {
		currentCi, err := tm.db.Cis.Get(tm.ctx, *cisInfo.Id)
		if err != nil {
			tx.Rollback()
			return result, fmt.Errorf("failed to get current CI: %w", err)
		}
		typeID = currentCi.TypeID
	}

	// 3. 校验属性唯一性
	if len(cisInfo.Attributes) > 0 {
		err = ValidateUniqueAttributes(tm.ctx, tm.db, typeID, cisInfo.Attributes, cisInfo.Id)
		if err != nil {
			tx.Rollback()
			result.ValidationErrors = append(result.ValidationErrors, err.Error())
			return result, fmt.Errorf("attribute validation failed: %w", err)
		}
	}

	// 4. 更新CI基础信息
	cisBuilder := tx.Cis.UpdateOneID(*cisInfo.Id)
	cisBuilder = CisUpdateBuilderSetter(cisBuilder, cisInfo)

	_, err = cisBuilder.Save(tm.ctx)
	if err != nil {
		tx.Rollback()
		return result, fmt.Errorf("failed to update CI: %w", err)
	}

	// 5. 更新动态属性值
	if len(cisInfo.Attributes) > 0 {
		err = SaveCiAttributes(tm.ctx, tx, *cisInfo.Id, cisInfo.Attributes)
		if err != nil {
			tx.Rollback()
			return result, fmt.Errorf("failed to save CI attributes: %w", err)
		}
	}

	// 6. 处理关系操作
	if cisInfo.Relations != nil {
		err = tm.processRelationOperations(tx, *cisInfo.Id, cisInfo.Relations, result)
		if err != nil {
			tx.Rollback()
			return result, fmt.Errorf("failed to process relations: %w", err)
		}
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tm.logger.Infof("Successfully updated CI %d", result.CiID)
	return result, nil
}

// validateCiData 验证CI数据完整性
func (tm *TransactionManager) validateCiData(cisInfo *cmdb.CisInfo) error {
	if cisInfo.TypeId == nil {
		return fmt.Errorf("CI type ID is required")
	}

	// 验证CI类型是否存在
	exists, err := tm.db.CiType.Query().Where(citype.IDEQ(*cisInfo.TypeId)).Exist(tm.ctx)
	if err != nil {
		return fmt.Errorf("failed to check CI type existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("CI type with ID %d not found", *cisInfo.TypeId)
	}

	// 验证必填属性等其他业务规则
	// 这里可以添加更多的验证逻辑

	return nil
}

// validateRelationsData 验证关系数据完整性
func (tm *TransactionManager) validateRelationsData(relationsData *cmdb.CiRelationsData) error {
	// 验证要创建的关系
	for i, createInfo := range relationsData.CreateRelations {
		if createInfo.TargetCiId == 0 {
			return fmt.Errorf("create relation %d: target CI ID is required", i)
		}
		if createInfo.RelationTypeId == 0 {
			return fmt.Errorf("create relation %d: relation type ID is required", i)
		}
		if createInfo.Direction == "" {
			return fmt.Errorf("create relation %d: direction is required", i)
		}
		if createInfo.Direction != "source_to_target" && createInfo.Direction != "target_to_source" {
			return fmt.Errorf("create relation %d: invalid direction %s", i, createInfo.Direction)
		}

		// 验证目标CI是否存在
		exists, err := tm.db.Cis.Query().Where(cis.IDEQ(createInfo.TargetCiId)).Exist(tm.ctx)
		if err != nil {
			return fmt.Errorf("create relation %d: failed to check target CI existence: %w", i, err)
		}
		if !exists {
			return fmt.Errorf("create relation %d: target CI with ID %d not found", i, createInfo.TargetCiId)
		}

		// 验证关系类型是否存在
		exists, err = tm.db.RelationType.Query().Where(relationtype.IDEQ(createInfo.RelationTypeId)).Exist(tm.ctx)
		if err != nil {
			return fmt.Errorf("create relation %d: failed to check relation type existence: %w", i, err)
		}
		if !exists {
			return fmt.Errorf("create relation %d: relation type with ID %d not found", i, createInfo.RelationTypeId)
		}
	}

	// 验证要更新的关系
	for i, updateInfo := range relationsData.UpdateRelations {
		if updateInfo.RelationId == 0 {
			return fmt.Errorf("update relation %d: relation ID is required", i)
		}

		// 验证关系是否存在
		exists, err := tm.db.CiRelation.Query().Where(cirelation.IDEQ(updateInfo.RelationId)).Exist(tm.ctx)
		if err != nil {
			return fmt.Errorf("update relation %d: failed to check relation existence: %w", i, err)
		}
		if !exists {
			return fmt.Errorf("update relation %d: relation with ID %d not found", i, updateInfo.RelationId)
		}
	}

	return nil
}

// createRelationsWithErrorHandling 创建关系并处理错误
func (tm *TransactionManager) createRelationsWithErrorHandling(tx *ent.Tx, ciId uint64, createInfos []*cmdb.CiRelationCreateInfo) ([]uint64, []string) {
	var createdRelations []uint64
	var failedRelations []string

	for i, createInfo := range createInfos {
		if err := CreateCiRelation(tm.ctx, tx, ciId, createInfo, tm.logger); err != nil {
			failedRelations = append(failedRelations, fmt.Sprintf("relation %d: %v", i, err))
			tm.logger.Errorf("Failed to create relation %d: %v", i, err)
		} else {
			// 这里应该获取创建的关系ID，但由于CreateCiRelation没有返回ID，
			// 我们暂时记录成功数量
			createdRelations = append(createdRelations, uint64(i))
		}
	}

	return createdRelations, failedRelations
}

// processRelationOperations 处理关系操作（删除、创建、更新）
func (tm *TransactionManager) processRelationOperations(tx *ent.Tx, ciId uint64, relationsData *cmdb.CiRelationsData, result *UpdateCiWithRelationsResult) error {
	// 删除关系
	if len(relationsData.DeleteRelationIds) > 0 {
		deletedCount, err := tm.deleteRelations(tx, ciId, relationsData.DeleteRelationIds)
		if err != nil {
			result.FailedOperations = append(result.FailedOperations, fmt.Sprintf("delete relations: %v", err))
		} else {
			result.DeletedRelations = relationsData.DeleteRelationIds[:deletedCount]
		}
	}

	// 创建关系
	if len(relationsData.CreateRelations) > 0 {
		createdRelations, failedRelations := tm.createRelationsWithErrorHandling(tx, ciId, relationsData.CreateRelations)
		result.CreatedRelations = createdRelations
		result.FailedOperations = append(result.FailedOperations, failedRelations...)
	}

	// 更新关系
	if len(relationsData.UpdateRelations) > 0 {
		updatedRelations, failedUpdates := tm.updateRelationsWithErrorHandling(tx, relationsData.UpdateRelations)
		result.UpdatedRelations = updatedRelations
		result.FailedOperations = append(result.FailedOperations, failedUpdates...)
	}

	return nil
}

// deleteRelations 删除关系
func (tm *TransactionManager) deleteRelations(tx *ent.Tx, ciId uint64, relationIds []uint64) (int, error) {
	deletedCount, err := tx.CiRelation.Delete().Where(
		cirelation.And(
			cirelation.IDIn(relationIds...),
			cirelation.Or(
				cirelation.SourceCiIDEQ(ciId),
				cirelation.TargetCiIDEQ(ciId),
			),
		),
	).Exec(tm.ctx)
	return deletedCount, err
}

// updateRelationsWithErrorHandling 更新关系并处理错误
func (tm *TransactionManager) updateRelationsWithErrorHandling(tx *ent.Tx, updateInfos []*cmdb.CiRelationUpdateInfo) ([]uint64, []string) {
	var updatedRelations []uint64
	var failedUpdates []string

	for _, updateInfo := range updateInfos {
		if err := UpdateCiRelation(tm.ctx, tx, updateInfo, tm.logger); err != nil {
			failedUpdates = append(failedUpdates, fmt.Sprintf("relation %d: %v", updateInfo.RelationId, err))
			tm.logger.Errorf("Failed to update relation %d: %v", updateInfo.RelationId, err)
		} else {
			updatedRelations = append(updatedRelations, updateInfo.RelationId)
		}
	}

	return updatedRelations, failedUpdates
}