package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypediscoveryconfig"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/discoveryexecutionhistory"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/provider"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// DiscoveryEngine CI属性发现引擎
type DiscoveryEngine struct {
	db                    *ent.Client
	providerRegistry      *provider.Registry
	mappingService        *AttributeMappingService
	advancedPersister     *AdvancedCIPersister
	incrementalService    *IncrementalUpdateService
	logger                logx.Logger
}

// NewDiscoveryEngine 创建新的发现引擎实例
func NewDiscoveryEngine(db *ent.Client) *DiscoveryEngine {
	return &DiscoveryEngine{
		db:                 db,
		providerRegistry:   provider.NewRegistry(),
		mappingService:     NewAttributeMappingService(db),
		advancedPersister:  NewAdvancedCIPersister(db),
		incrementalService: NewIncrementalUpdateService(db),
		logger:             logx.WithContext(context.Background()),
	}
}

// ExecuteDiscovery 执行发现任务
func (e *DiscoveryEngine) ExecuteDiscovery(ctx context.Context, configID uint64) (*types.DiscoveryResult, error) {
	// 1. 加载发现配置
	config, err := e.loadDiscoveryConfig(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to load discovery config: %w", err)
	}

	// 2. 创建执行历史记录
	executionID, err := e.createExecutionHistory(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution history: %w", err)
	}

	// 3. 开始执行发现
	result := &types.DiscoveryResult{
		ExecutionID:   executionID,
		ConfigID:      configID,
		StartTime:     time.Now(),
		Status:        types.StatusRunning,
		TotalRecords:  0,
		SuccessCount:  0,
		FailedCount:   0,
		SkippedCount:  0,
		CreatedCIs:    0,
		UpdatedCIs:    0,
	}

	// 4. 异步执行发现任务
	go e.executeDiscoveryAsync(ctx, config, result)

	return result, nil
}

// loadDiscoveryConfig 加载发现配置
func (e *DiscoveryEngine) loadDiscoveryConfig(ctx context.Context, configID uint64) (*ent.CiTypeDiscoveryConfig, error) {
	config, err := e.db.CiTypeDiscoveryConfig.Query().
		Where(
			citypediscoveryconfig.IDEQ(configID),
			citypediscoveryconfig.EnabledEQ(true),
			citypediscoveryconfig.ConfigStatusEQ("active"),
		).
		WithCiType().
		WithAttributeMappingRules().
		First(ctx)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// createExecutionHistory 创建执行历史记录
func (e *DiscoveryEngine) createExecutionHistory(ctx context.Context, config *ent.CiTypeDiscoveryConfig) (string, error) {
	executionID := fmt.Sprintf("exec_%d_%d", config.ID, time.Now().Unix())
	
	history, err := e.db.DiscoveryExecutionHistory.Create().
		SetDiscoveryConfigID(config.ID).
		SetExecutionID(executionID).
		SetTriggerType("manual").
		SetStartedAt(time.Now()).
		SetExecStatus("running").
		SetStage("connecting").
		SetProgress(0).
		Save(ctx)

	if err != nil {
		return "", err
	}

	return history.ExecutionID, nil
}

// executeDiscoveryAsync 异步执行发现任务
func (e *DiscoveryEngine) executeDiscoveryAsync(ctx context.Context, config *ent.CiTypeDiscoveryConfig, result *types.DiscoveryResult) {
	defer e.updateExecutionResult(ctx, result)

	// 1. 获取发现提供者
	discoveryProvider, err := e.providerRegistry.GetProvider(config.DiscoveryType, config.ProviderID)
	if err != nil {
		result.Status = types.StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to get provider: %v", err)
		return
	}

	// 2. 连接到数据源
	e.updateStage(ctx, result.ExecutionID, "connecting", 10)
	dataSource, err := discoveryProvider.Connect(ctx, config.ProviderConfig)
	if err != nil {
		result.Status = types.StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to connect to data source: %v", err)
		return
	}
	defer dataSource.Close()

	// 3. 执行数据发现
	e.updateStage(ctx, result.ExecutionID, "discovering", 30)
	rawData, err := dataSource.Discover(ctx, config.DiscoveryRules)
	if err != nil {
		result.Status = types.StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to discover data: %v", err)
		return
	}

	result.TotalRecords = int64(len(rawData))

	// 4. 数据转换和映射
	e.updateStage(ctx, result.ExecutionID, "transforming", 60)
	transformedData, err := e.transformData(ctx, config, rawData)
	if err != nil {
		result.Status = types.StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to transform data: %v", err)
		return
	}

	// 5. 持久化CI数据
	e.updateStage(ctx, result.ExecutionID, "persisting", 80)
	persistResult, err := e.persistCIData(ctx, config, transformedData)
	if err != nil {
		result.Status = types.StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to persist CI data: %v", err)
		return
	}

	// 6. 更新结果统计
	result.SuccessCount = persistResult.SuccessCount
	result.FailedCount = persistResult.FailedCount
	result.SkippedCount = persistResult.SkippedCount
	result.CreatedCIs = persistResult.CreatedCIs
	result.UpdatedCIs = persistResult.UpdatedCIs
	result.Status = types.StatusCompleted
	result.EndTime = &time.Time{}
	*result.EndTime = time.Now()

	e.updateStage(ctx, result.ExecutionID, "completed", 100)
}

// updateStage 更新执行阶段
func (e *DiscoveryEngine) updateStage(ctx context.Context, executionID, stage string, progress int) {
	_, err := e.db.DiscoveryExecutionHistory.Update().
		Where(discoveryexecutionhistory.ExecutionIDEQ(executionID)).
		SetStage(stage).
		SetProgress(progress).
		Save(ctx)

	if err != nil {
		e.logger.Errorf("Failed to update execution stage: %v", err)
	}
}

// updateExecutionResult 更新执行结果
func (e *DiscoveryEngine) updateExecutionResult(ctx context.Context, result *types.DiscoveryResult) {
	execResult := map[string]interface{}{
		"total_records":  result.TotalRecords,
		"success_count":  result.SuccessCount,
		"failed_count":   result.FailedCount,
		"skipped_count":  result.SkippedCount,
		"created_cis":    result.CreatedCIs,
		"updated_cis":    result.UpdatedCIs,
		"start_time":     result.StartTime,
		"end_time":       result.EndTime,
	}

	updateQuery := e.db.DiscoveryExecutionHistory.Update().
		Where(discoveryexecutionhistory.ExecutionIDEQ(result.ExecutionID)).
		SetExecStatus(string(result.Status)).
		SetTotalRecords(result.TotalRecords).
		SetSuccessRecords(result.SuccessCount).
		SetFailedRecords(result.FailedCount).
		SetSkippedRecords(result.SkippedCount).
		SetCreatedCis(result.CreatedCIs).
		SetUpdatedCis(result.UpdatedCIs).
		SetExecutionResult(execResult)

	if result.EndTime != nil {
		updateQuery.SetCompletedAt(*result.EndTime)
		duration := result.EndTime.Sub(result.StartTime)
		updateQuery.SetDurationSeconds(int(duration.Seconds()))
	}

	if result.ErrorMessage != "" {
		updateQuery.SetErrorMessage(result.ErrorMessage)
	}

	_, err := updateQuery.Save(ctx)
	if err != nil {
		e.logger.Errorf("Failed to update execution result: %v", err)
	}
}

// transformData 转换数据 - 使用高级属性映射服务
func (e *DiscoveryEngine) transformData(ctx context.Context, config *ent.CiTypeDiscoveryConfig, rawData []map[string]interface{}) ([]types.TransformedCIData, error) {
	var transformedData []types.TransformedCIData

	// 创建映射上下文
	mappingContext := MappingContext{
		CiTypeID:       config.CiTypeID,
		ProviderID:     config.ProviderID,
		SourceSystem:   config.DiscoveryType,
		GlobalSettings: make(map[string]interface{}),
	}

	// 如果有全局转换设置，从配置中提取
	if config.AttributeMappings != nil {
		// AttributeMappings is []interface{}, so iterate through it
		for _, mapping := range config.AttributeMappings {
			if mappingMap, ok := mapping.(map[string]interface{}); ok {
				if globalSettings, ok := mappingMap["global_settings"].(map[string]interface{}); ok {
					mappingContext.GlobalSettings = globalSettings
					break
				}
			}
		}
	}

	// 处理每条原始数据
	for i, sourceData := range rawData {
		// 使用高级属性映射服务转换数据
		mappingResult, err := e.mappingService.MapAttributes(ctx, sourceData, mappingContext)
		if err != nil {
			e.logger.Errorw("Failed to map attributes for record",
				logx.Field("record_index", i),
				logx.Field("error", err))
			continue
		}

		// 检查是否有验证错误
		if len(mappingResult.ValidationErrors) > 0 {
			e.logger.Errorw("Validation errors in attribute mapping",
				logx.Field("record_index", i),
				logx.Field("validation_errors", mappingResult.ValidationErrors))
		}

		// 创建转换后的CI数据
		transformedCI := types.TransformedCIData{
			SourceID:         fmt.Sprintf("%v", sourceData["id"]),
			CITypeID:         config.CiTypeID,
			Attributes:       mappingResult.MappedData,
			SourceData:       sourceData,
			TransformStats:   types.TransformationStats(mappingResult.TransformStats),
			ValidationErrors: convertValidationErrors(mappingResult.ValidationErrors),
		}

		// 记录未映射的字段用于调试
		if len(mappingResult.UnmappedFields) > 0 {
			e.logger.Debugw("Unmapped fields detected",
				logx.Field("record_index", i),
				logx.Field("unmapped_fields", mappingResult.UnmappedFields))
		}

		transformedData = append(transformedData, transformedCI)
	}

	e.logger.Infow("Data transformation completed",
		logx.Field("total_records", len(rawData)),
		logx.Field("transformed_records", len(transformedData)),
		logx.Field("config_id", config.ID))

	return transformedData, nil
}

// persistCIData 持久化CI数据 - 使用高级持久化器和增量更新
func (e *DiscoveryEngine) persistCIData(ctx context.Context, config *ent.CiTypeDiscoveryConfig, data []types.TransformedCIData) (*types.PersistResult, error) {
	// 1. 检查是否启用增量更新
	if e.isIncrementalUpdateEnabled(config) {
		return e.persistWithIncrementalUpdate(ctx, config, data)
	}
	
	// 2. 使用高级持久化器进行冲突解决
	return e.advancedPersister.PersistWithConflictResolution(ctx, config, data)
}

// persistWithIncrementalUpdate 使用增量更新进行持久化
func (e *DiscoveryEngine) persistWithIncrementalUpdate(ctx context.Context, config *ent.CiTypeDiscoveryConfig, data []types.TransformedCIData) (*types.PersistResult, error) {
	e.logger.Infow("Starting incremental update process",
		logx.Field("config_id", config.ID),
		logx.Field("total_records", len(data)))
	
	// 1. 处理增量更新
	deltaResult, err := e.incrementalService.ProcessIncrementalUpdate(ctx, config, data)
	if err != nil {
		return nil, fmt.Errorf("incremental update failed: %w", err)
	}
	
	e.logger.Infow("Incremental update analysis completed",
		logx.Field("new_items", deltaResult.Statistics.NewCount),
		logx.Field("updated_items", deltaResult.Statistics.UpdatedCount),
		logx.Field("deleted_items", deltaResult.Statistics.DeletedCount),
		logx.Field("conflicts", deltaResult.Statistics.ConflictCount))
	
	// 2. 处理需要持久化的数据
	var dataToProcess []types.TransformedCIData
	dataToProcess = append(dataToProcess, deltaResult.NewItems...)
	dataToProcess = append(dataToProcess, deltaResult.UpdatedItems...)
	
	if len(dataToProcess) == 0 {
		// 没有需要更新的数据
		return &types.PersistResult{
			SuccessCount: int64(deltaResult.Statistics.SkippedCount),
			SkippedCount: int64(deltaResult.Statistics.SkippedCount),
		}, nil
	}
	
	// 3. 使用高级持久化器处理数据
	persistResult, err := e.advancedPersister.PersistWithConflictResolution(ctx, config, dataToProcess)
	if err != nil {
		return nil, fmt.Errorf("advanced persistence failed: %w", err)
	}
	
	// 4. 处理冲突数据
	if len(deltaResult.Conflicts) > 0 {
		conflictResult, err := e.handleConflictData(ctx, config, deltaResult.Conflicts)
		if err != nil {
			e.logger.Errorw("Failed to handle conflict data",
				logx.Field("conflicts_count", len(deltaResult.Conflicts)),
				logx.Field("error", err))
		} else {
			// 合并冲突处理结果
			persistResult.SuccessCount += conflictResult.SuccessCount
			persistResult.FailedCount += conflictResult.FailedCount
			persistResult.SkippedCount += conflictResult.SkippedCount
		}
	}
	
	// 5. 处理删除的数据
	if len(deltaResult.DeletedItems) > 0 {
		deleteResult, err := e.handleDeletedItems(ctx, config, deltaResult.DeletedItems)
		if err != nil {
			e.logger.Errorw("Failed to handle deleted items",
				logx.Field("deleted_count", len(deltaResult.DeletedItems)),
				logx.Field("error", err))
		} else {
			persistResult.UpdatedCIs += deleteResult.UpdatedCIs
		}
	}
	
	e.logger.Infow("Incremental update completed",
		logx.Field("success_count", persistResult.SuccessCount),
		logx.Field("failed_count", persistResult.FailedCount),
		logx.Field("skipped_count", persistResult.SkippedCount))
	
	return persistResult, nil
}

// isIncrementalUpdateEnabled 检查是否启用增量更新
func (e *DiscoveryEngine) isIncrementalUpdateEnabled(config *ent.CiTypeDiscoveryConfig) bool {
	// 检查配置中的执行模式
	return config.ExecutionMode == "incremental" || config.ExecutionMode == "delta"
}

// handleConflictData 处理冲突数据
func (e *DiscoveryEngine) handleConflictData(ctx context.Context, config *ent.CiTypeDiscoveryConfig, conflicts []ConflictContext) (*types.PersistResult, error) {
	result := &types.PersistResult{}
	
	for _, conflict := range conflicts {
		// 根据配置的冲突解决策略处理
		resolution, err := e.advancedPersister.resolveConflicts(ctx, &conflict, config)
		if err != nil {
			result.FailedCount++
			e.logger.Errorw("Failed to resolve conflict",
				logx.Field("conflict_fields", conflict.ConflictFields),
				logx.Field("error", err))
			continue
		}
		
		if resolution.RequiresReview {
			// 需要人工审核的冲突，记录到待审核队列
			if err := e.queueForReview(ctx, config, &conflict, resolution); err != nil {
				e.logger.Errorw("Failed to queue conflict for review",
					logx.Field("error", err))
			}
			result.SkippedCount++
		} else {
			// 自动解决的冲突，继续处理
			result.SuccessCount++
		}
	}
	
	return result, nil
}

// handleDeletedItems 处理删除的数据项
func (e *DiscoveryEngine) handleDeletedItems(ctx context.Context, config *ent.CiTypeDiscoveryConfig, deletedItems []string) (*types.PersistResult, error) {
	result := &types.PersistResult{}
	
	// 根据配置决定如何处理删除的项目
	deleteStrategy := e.getDeleteStrategy(config)
	
	for _, sourceID := range deletedItems {
		switch deleteStrategy {
		case "soft_delete":
			// 软删除：标记为已删除但保留数据
			if err := e.softDeleteCI(ctx, sourceID); err != nil {
				result.FailedCount++
				e.logger.Errorw("Failed to soft delete CI",
					logx.Field("source_id", sourceID),
					logx.Field("error", err))
			} else {
				result.UpdatedCIs++
			}
			
		case "hard_delete":
			// 硬删除：完全移除数据
			if err := e.hardDeleteCI(ctx, sourceID); err != nil {
				result.FailedCount++
				e.logger.Errorw("Failed to hard delete CI",
					logx.Field("source_id", sourceID),
					logx.Field("error", err))
			} else {
				result.UpdatedCIs++
			}
			
		case "ignore":
			// 忽略：不处理删除的项目
			result.SkippedCount++
			
		default:
			// 默认：软删除
			if err := e.softDeleteCI(ctx, sourceID); err != nil {
				result.FailedCount++
			} else {
				result.UpdatedCIs++
			}
		}
	}
	
	return result, nil
}

// queueForReview 将冲突数据加入审核队列
func (e *DiscoveryEngine) queueForReview(ctx context.Context, config *ent.CiTypeDiscoveryConfig, conflict *ConflictContext, resolution *ConflictResolution) error {
	// TODO: 实现审核队列逻辑
	// 可以将冲突数据保存到专门的审核表中
	e.logger.Infow("Conflict queued for manual review",
		logx.Field("config_id", config.ID),
		logx.Field("conflict_fields", conflict.ConflictFields))
	return nil
}

// getDeleteStrategy 获取删除策略
func (e *DiscoveryEngine) getDeleteStrategy(config *ent.CiTypeDiscoveryConfig) string {
	// 从配置中获取删除策略
	if config.FilterConditions != nil {
		if strategy, ok := config.FilterConditions["delete_strategy"].(string); ok {
			return strategy
		}
	}
	return "soft_delete" // 默认软删除
}

// softDeleteCI 软删除CI
func (e *DiscoveryEngine) softDeleteCI(ctx context.Context, sourceID string) error {
	// TODO: 实现软删除逻辑
	// 例如：更新deleted_at字段或status字段
	e.logger.Debugw("Soft deleting CI", logx.Field("source_id", sourceID))
	return nil
}

// hardDeleteCI 硬删除CI  
func (e *DiscoveryEngine) hardDeleteCI(ctx context.Context, sourceID string) error {
	// TODO: 实现硬删除逻辑
	// 例如：从数据库中完全删除记录
	e.logger.Debugw("Hard deleting CI", logx.Field("source_id", sourceID))
	return nil
}

// RegisterProvider 注册提供者
func (e *DiscoveryEngine) RegisterProvider(provider provider.Provider) error {
	return e.providerRegistry.RegisterProvider(provider)
}

// GetProvider 获取提供者
func (e *DiscoveryEngine) GetProvider(providerType, providerID string) (provider.Provider, error) {
	return e.providerRegistry.GetProvider(providerType, providerID)
}

// ListProviders 列出提供者
func (e *DiscoveryEngine) ListProviders() map[string][]string {
	return e.providerRegistry.ListProviders()
}

// ValidateProviderConfig 验证提供者配置
func (e *DiscoveryEngine) ValidateProviderConfig(providerType, providerID string, config map[string]interface{}) error {
	return e.providerRegistry.ValidateProvider(providerType, providerID, config)
}

// convertValidationErrors converts engine ValidationErrors to types ValidationErrors
func convertValidationErrors(errors []ValidationError) []types.ValidationError {
	var result []types.ValidationError
	for _, err := range errors {
		result = append(result, types.ValidationError{
			Field:   err.Field,
			Value:   fmt.Sprintf("%v", err.Value),
			Rule:    err.Rule,
			Message: err.Message,
		})
	}
	return result
}