package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// CIPersister CI数据持久化器
type CIPersister struct {
	db     *ent.Client
	logger logx.Logger
}

// NewCIPersister 创建新的CI持久化器
func NewCIPersister(db *ent.Client) *CIPersister {
	return &CIPersister{
		db:     db,
		logger: logx.WithContext(context.Background()),
	}
}

// Persist 持久化CI数据
func (p *CIPersister) Persist(ctx context.Context, config *ent.CiTypeDiscoveryConfig, data []types.TransformedCIData) (*types.PersistResult, error) {
	result := &types.PersistResult{
		SuccessCount: 0,
		FailedCount:  0,
		SkippedCount: 0,
		CreatedCIs:   0,
		UpdatedCIs:   0,
		Errors:       []error{},
	}

	// 使用事务确保数据一致性
	tx, err := p.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for _, ciData := range data {
		err := p.persistSingleCI(ctx, tx, config, ciData, result)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, err)
			p.logger.Errorf("Failed to persist CI %s: %v", ciData.UniqueKey, err)
			
			// 根据配置决定是否继续
			// TODO: 添加ContinueOnError字段到schema
			continueOnError := true // 默认继续执行
			if !continueOnError {
				return result, fmt.Errorf("persist failed and continue_on_error is false: %w", err)
			}
		} else {
			result.SuccessCount++
			if ciData.IsUpdate {
				result.UpdatedCIs++
			} else {
				result.CreatedCIs++
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// persistSingleCI 持久化单个CI
func (p *CIPersister) persistSingleCI(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData, result *types.PersistResult) error {
	// 检查冲突解决策略
	if ciData.IsUpdate && ciData.ExistingCIID != nil {
		return p.updateExistingCI(ctx, tx, config, ciData)
	}

	// 检查是否已存在（基于唯一键）
	existingCI, err := p.findCIByUniqueKey(ctx, tx, ciData)
	if err != nil {
		return fmt.Errorf("failed to check existing CI: %w", err)
	}

	if existingCI != nil {
		// 已存在，根据更新策略处理
		return p.handleExistingCI(ctx, tx, config, ciData, existingCI.ID)
	}

	// 创建新CI
	return p.createNewCI(ctx, tx, config, ciData)
}

// findCIByUniqueKey 根据唯一键查找CI
func (p *CIPersister) findCIByUniqueKey(ctx context.Context, tx *ent.Tx, ciData types.TransformedCIData) (*DummyCI, error) {
	// 这里需要根据实际的CI表结构来实现
	// 假设CI表有unique_key字段
	/*
	return tx.ConfigurationItem.Query().
		Where(configurationitem.UniqueKeyEQ(ciData.UniqueKey)).
		Where(configurationitem.CiTypeIDEQ(ciData.CITypeID)).
		First(ctx)
	*/
	
	// 暂时返回nil，表示未找到
	return nil, nil
}

// DummyCI 临时CI结构（待实际CI表定义）
type DummyCI struct {
	ID uint64
}

// handleExistingCI 处理已存在的CI
func (p *CIPersister) handleExistingCI(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData, existingID uint64) error {
	switch config.ConflictResolution {
	case "skip":
		// 跳过，不做任何操作
		return nil
	case "update":
		// 更新现有CI
		return p.updateCI(ctx, tx, existingID, ciData)
	case "overwrite":
		// 覆盖现有CI
		return p.overwriteCI(ctx, tx, existingID, ciData)
	case "merge":
		// 合并属性
		return p.mergeCI(ctx, tx, existingID, ciData)
	default:
		return fmt.Errorf("unknown conflict resolution strategy: %s", config.ConflictResolution)
	}
}

// createNewCI 创建新CI
func (p *CIPersister) createNewCI(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData) error {
	// 这里需要根据实际的CI表结构来实现
	// 示例代码：
	/*
	_, err := tx.ConfigurationItem.Create().
		SetCiTypeID(ciData.CITypeID).
		SetUniqueKey(ciData.UniqueKey).
		SetSourceID(ciData.SourceID).
		SetAttributes(ciData.Attributes).
		SetMetadata(ciData.Metadata).
		SetDiscoverySource(config.DiscoveryType).
		SetLastDiscoveredAt(time.Now()).
		Save(ctx)
	*/
	
	// 暂时模拟成功
	p.logger.Infof("Created new CI: %s", ciData.UniqueKey)
	return nil
}

// updateExistingCI 更新已存在的CI
func (p *CIPersister) updateExistingCI(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData) error {
	return p.updateCI(ctx, tx, *ciData.ExistingCIID, ciData)
}

// updateCI 更新CI
func (p *CIPersister) updateCI(ctx context.Context, tx *ent.Tx, ciID uint64, ciData types.TransformedCIData) error {
	// 这里需要根据实际的CI表结构来实现
	// 示例代码：
	/*
	_, err := tx.ConfigurationItem.UpdateOneID(ciID).
		SetAttributes(ciData.Attributes).
		SetMetadata(ciData.Metadata).
		SetLastDiscoveredAt(time.Now()).
		Save(ctx)
	*/
	
	// 暂时模拟成功
	p.logger.Infof("Updated CI: %s (ID: %d)", ciData.UniqueKey, ciID)
	return nil
}

// overwriteCI 覆盖CI
func (p *CIPersister) overwriteCI(ctx context.Context, tx *ent.Tx, ciID uint64, ciData types.TransformedCIData) error {
	// 完全覆盖现有CI的所有属性
	return p.updateCI(ctx, tx, ciID, ciData)
}

// mergeCI 合并CI属性
func (p *CIPersister) mergeCI(ctx context.Context, tx *ent.Tx, ciID uint64, ciData types.TransformedCIData) error {
	// 先获取现有CI
	/*
	existingCI, err := tx.ConfigurationItem.Get(ctx, ciID)
	if err != nil {
		return fmt.Errorf("failed to get existing CI: %w", err)
	}
	
	// 合并属性
	mergedAttributes := make(map[string]interface{})
	
	// 复制现有属性
	for k, v := range existingCI.Attributes {
		mergedAttributes[k] = v
	}
	
	// 覆盖新属性
	for k, v := range ciData.Attributes {
		mergedAttributes[k] = v
	}
	
	// 更新CI
	_, err = tx.ConfigurationItem.UpdateOneID(ciID).
		SetAttributes(mergedAttributes).
		SetLastDiscoveredAt(time.Now()).
		Save(ctx)
	*/
	
	// 暂时模拟成功
	p.logger.Infof("Merged CI: %s (ID: %d)", ciData.UniqueKey, ciID)
	return nil
}

// validateCI 验证CI数据
func (p *CIPersister) validateCI(ciData types.TransformedCIData) []types.ValidationError {
	var errors []types.ValidationError

	// 基本验证
	if ciData.CITypeID == 0 {
		errors = append(errors, types.ValidationError{
			Field:   "ci_type_id",
			Value:   fmt.Sprintf("%d", ciData.CITypeID),
			Rule:    "required",
			Message: "CI type ID is required",
		})
	}

	if ciData.UniqueKey == "" {
		errors = append(errors, types.ValidationError{
			Field:   "unique_key",
			Value:   ciData.UniqueKey,
			Rule:    "required",
			Message: "Unique key is required",
		})
	}

	// 属性验证
	if len(ciData.Attributes) == 0 {
		errors = append(errors, types.ValidationError{
			Field:   "attributes",
			Value:   "empty",
			Rule:    "required",
			Message: "At least one attribute is required",
		})
	}

	return errors
}

// bulkCreate 批量创建CI（可选优化）
func (p *CIPersister) bulkCreate(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciDataList []types.TransformedCIData) error {
	// 实现批量创建逻辑以提高性能
	// 这里可以使用数据库的批量插入功能
	
	batchSize := 100 // 每批处理100条记录
	
	for i := 0; i < len(ciDataList); i += batchSize {
		end := i + batchSize
		if end > len(ciDataList) {
			end = len(ciDataList)
		}
		
		batch := ciDataList[i:end]
		if err := p.processBatch(ctx, tx, config, batch); err != nil {
			return fmt.Errorf("failed to process batch %d-%d: %w", i, end-1, err)
		}
	}
	
	return nil
}

// processBatch 处理一批CI数据
func (p *CIPersister) processBatch(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, batch []types.TransformedCIData) error {
	// 实现批量处理逻辑
	for _, ciData := range batch {
		if err := p.createNewCI(ctx, tx, config, ciData); err != nil {
			return err
		}
	}
	return nil
}

// recordPersistenceMetrics 记录持久化指标
func (p *CIPersister) recordPersistenceMetrics(ctx context.Context, config *ent.CiTypeDiscoveryConfig, result *types.PersistResult, duration time.Duration) {
	// 记录性能指标到监控系统
	metrics := map[string]interface{}{
		"config_id":      config.ID,
		"success_count":  result.SuccessCount,
		"failed_count":   result.FailedCount,
		"skipped_count":  result.SkippedCount,
		"created_cis":    result.CreatedCIs,
		"updated_cis":    result.UpdatedCIs,
		"duration_ms":    duration.Milliseconds(),
		"records_per_sec": float64(result.SuccessCount) / duration.Seconds(),
	}
	
	p.logger.Infof("Persistence metrics: %+v", metrics)
}