package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConflictResolutionStrategy 冲突解决策略
type ConflictResolutionStrategy string

const (
	StrategySkip      ConflictResolutionStrategy = "skip"      // 跳过冲突记录
	StrategyOverwrite ConflictResolutionStrategy = "overwrite" // 覆盖现有记录
	StrategyMerge     ConflictResolutionStrategy = "merge"     // 合并属性
	StrategyUpdate    ConflictResolutionStrategy = "update"    // 仅更新变化的属性
	StrategyAppend    ConflictResolutionStrategy = "append"    // 追加到数组字段
	StrategyCustom    ConflictResolutionStrategy = "custom"    // 自定义解决策略
)

// IncrementalUpdateMode 增量更新模式
type IncrementalUpdateMode string

const (
	ModeTimestamp   IncrementalUpdateMode = "timestamp"   // 基于时间戳
	ModeChecksum    IncrementalUpdateMode = "checksum"    // 基于数据校验和
	ModeVersioned   IncrementalUpdateMode = "versioned"   // 基于版本号
	ModeFieldLevel  IncrementalUpdateMode = "field_level" // 字段级别比较
)

// AdvancedCIPersister 高级CI数据持久化器
type AdvancedCIPersister struct {
	db                    *ent.Client
	logger                logx.Logger
	conflictResolver      *ConflictResolver
	incrementalProcessor  *IncrementalProcessor
	changeDetector        *ChangeDetector
}

// ConflictResolver 冲突解决器
type ConflictResolver struct {
	strategies map[string]ConflictResolutionStrategy
	rules      []ConflictRule
}

// ConflictRule 冲突解决规则
type ConflictRule struct {
	Field             string                     `json:"field"`
	Strategy          ConflictResolutionStrategy `json:"strategy"`
	Priority          int                        `json:"priority"`
	Condition         string                     `json:"condition"`         // 触发条件
	CustomHandler     string                     `json:"custom_handler"`    // 自定义处理器名称
	PreserveHistory   bool                       `json:"preserve_history"`  // 是否保留历史记录
	RequireConfirm    bool                       `json:"require_confirm"`   // 是否需要确认
}

// IncrementalProcessor 增量处理器
type IncrementalProcessor struct {
	mode              IncrementalUpdateMode
	timestampFields   []string
	checksumAlgorithm string
	batchSize         int
}

// ChangeDetector 变更检测器
type ChangeDetector struct {
	ignoredFields     []string
	sensitiveFields   []string
	customComparators map[string]func(interface{}, interface{}) bool
}

// ConflictContext 冲突上下文
type ConflictContext struct {
	ExistingData    map[string]interface{} `json:"existing_data"`
	IncomingData    map[string]interface{} `json:"incoming_data"`
	ConflictFields  []string               `json:"conflict_fields"`
	Resolution      ConflictResolution     `json:"resolution"`
	Timestamp       time.Time              `json:"timestamp"`
	SourceInfo      SourceInfo             `json:"source_info"`
}

// ConflictResolution 冲突解决结果
type ConflictResolution struct {
	Strategy      ConflictResolutionStrategy `json:"strategy"`
	ResolvedData  map[string]interface{}     `json:"resolved_data"`
	ChangedFields []string                   `json:"changed_fields"`
	Warnings      []string                   `json:"warnings"`
	RequiresReview bool                      `json:"requires_review"`
}

// SourceInfo 数据源信息
type SourceInfo struct {
	ProviderID   string    `json:"provider_id"`
	SourceSystem string    `json:"source_system"`
	LastSync     time.Time `json:"last_sync"`
	DataVersion  string    `json:"data_version"`
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Operation string      `json:"operation"` // create, update, delete
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

// NewAdvancedCIPersister 创建高级CI持久化器
func NewAdvancedCIPersister(db *ent.Client) *AdvancedCIPersister {
	return &AdvancedCIPersister{
		db:     db,
		logger: logx.WithContext(context.Background()),
		conflictResolver: &ConflictResolver{
			strategies: make(map[string]ConflictResolutionStrategy),
			rules:      []ConflictRule{},
		},
		incrementalProcessor: &IncrementalProcessor{
			mode:              ModeTimestamp,
			timestampFields:   []string{"updated_at", "last_modified", "sync_time"},
			checksumAlgorithm: "sha256",
			batchSize:         100,
		},
		changeDetector: &ChangeDetector{
			ignoredFields:     []string{"id", "created_at", "updated_at"},
			sensitiveFields:   []string{"password", "secret", "token"},
			customComparators: make(map[string]func(interface{}, interface{}) bool),
		},
	}
}

// PersistWithConflictResolution 使用冲突解决的持久化
func (p *AdvancedCIPersister) PersistWithConflictResolution(
	ctx context.Context, 
	config *ent.CiTypeDiscoveryConfig, 
	data []types.TransformedCIData,
) (*types.PersistResult, error) {
	
	result := &types.PersistResult{
		SuccessCount: 0,
		FailedCount:  0,
		SkippedCount: 0,
		CreatedCIs:   0,
		UpdatedCIs:   0,
		Errors:       []error{},
	}

	// 配置冲突解决策略
	if err := p.configureConflictResolution(config); err != nil {
		return nil, fmt.Errorf("failed to configure conflict resolution: %w", err)
	}

	// 处理批次数据
	for i := 0; i < len(data); i += p.incrementalProcessor.batchSize {
		end := i + p.incrementalProcessor.batchSize
		if end > len(data) {
			end = len(data)
		}
		
		batch := data[i:end]
		batchResult, err := p.processBatch(ctx, config, batch)
		if err != nil {
			p.logger.Errorw("Batch processing failed",
				logx.Field("batch_start", i),
				logx.Field("batch_size", len(batch)),
				logx.Field("error", err))
			return nil, err
		}
		
		// 累积结果
		result.SuccessCount += batchResult.SuccessCount
		result.FailedCount += batchResult.FailedCount
		result.SkippedCount += batchResult.SkippedCount
		result.CreatedCIs += batchResult.CreatedCIs
		result.UpdatedCIs += batchResult.UpdatedCIs
		result.Errors = append(result.Errors, batchResult.Errors...)
	}

	return result, nil
}

// processBatch 处理单个批次
func (p *AdvancedCIPersister) processBatch(
	ctx context.Context,
	config *ent.CiTypeDiscoveryConfig,
	batch []types.TransformedCIData,
) (*types.PersistResult, error) {
	
	result := &types.PersistResult{}
	
	// 开始事务
	tx, err := p.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for _, ciData := range batch {
		if err := p.persistWithIncrementalUpdate(ctx, tx, config, ciData, result); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, err)
			
			p.logger.Errorw("Failed to persist CI with incremental update",
				logx.Field("source_id", ciData.SourceID),
				logx.Field("error", err))
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// persistWithIncrementalUpdate 使用增量更新持久化
func (p *AdvancedCIPersister) persistWithIncrementalUpdate(
	ctx context.Context,
	tx *ent.Tx,
	config *ent.CiTypeDiscoveryConfig,
	ciData types.TransformedCIData,
	result *types.PersistResult,
) error {

	// 1. 检测是否存在现有记录
	existingData, err := p.findExistingCI(ctx, tx, ciData)
	if err != nil {
		return fmt.Errorf("failed to find existing CI: %w", err)
	}

	if existingData == nil {
		// 创建新记录
		return p.createNewCIWithValidation(ctx, tx, config, ciData, result)
	}

	// 2. 检测是否需要更新（增量检测）
	needsUpdate, changes := p.changeDetector.detectChanges(existingData, ciData.Attributes)
	if !needsUpdate {
		result.SkippedCount++
		p.logger.Debugw("No changes detected, skipping update",
			logx.Field("source_id", ciData.SourceID))
		return nil
	}

	// 3. 检测冲突并解决
	conflictContext := &ConflictContext{
		ExistingData: existingData,
		IncomingData: ciData.Attributes,
		ConflictFields: p.identifyConflictFields(existingData, ciData.Attributes),
		Timestamp:    time.Now(),
		SourceInfo: SourceInfo{
			ProviderID:   config.ProviderID,
			SourceSystem: config.DiscoveryType,
			LastSync:     time.Now(),
		},
	}

	resolution, err := p.resolveConflicts(ctx, conflictContext, config)
	if err != nil {
		return fmt.Errorf("failed to resolve conflicts: %w", err)
	}

	// 4. 应用解决方案并更新
	return p.applyResolutionAndUpdate(ctx, tx, config, ciData, existingData, resolution, changes, result)
}

// detectChanges 检测数据变更
func (p *ChangeDetector) detectChanges(existing, incoming map[string]interface{}) (bool, []ChangeRecord) {
	var changes []ChangeRecord
	hasChanges := false

	for field, newValue := range incoming {
		// 跳过忽略字段
		if p.isIgnoredField(field) {
			continue
		}

		oldValue, exists := existing[field]
		if !exists {
			// 新字段
			changes = append(changes, ChangeRecord{
				Field:     field,
				OldValue:  nil,
				NewValue:  newValue,
				Operation: "create",
				Timestamp: time.Now(),
			})
			hasChanges = true
			continue
		}

		// 使用自定义比较器（如果存在）
		if comparator, hasCustom := p.customComparators[field]; hasCustom {
			if !comparator(oldValue, newValue) {
				changes = append(changes, ChangeRecord{
					Field:     field,
					OldValue:  oldValue,
					NewValue:  newValue,
					Operation: "update",
					Timestamp: time.Now(),
				})
				hasChanges = true
			}
			continue
		}

		// 默认深度比较
		if !p.deepEqual(oldValue, newValue) {
			changes = append(changes, ChangeRecord{
				Field:     field,
				OldValue:  oldValue,
				NewValue:  newValue,
				Operation: "update",
				Timestamp: time.Now(),
			})
			hasChanges = true
		}
	}

	return hasChanges, changes
}

// resolveConflicts 解决冲突
func (p *AdvancedCIPersister) resolveConflicts(
	ctx context.Context,
	conflictCtx *ConflictContext,
	config *ent.CiTypeDiscoveryConfig,
) (*ConflictResolution, error) {
	
	resolution := &ConflictResolution{
		Strategy:      StrategyMerge, // 默认策略
		ResolvedData:  make(map[string]interface{}),
		ChangedFields: []string{},
		Warnings:      []string{},
		RequiresReview: false,
	}

	// 获取冲突解决策略
	strategy := p.getConflictStrategy(config)
	resolution.Strategy = strategy

	switch strategy {
	case StrategySkip:
		return p.resolveBySkipping(conflictCtx, resolution)
		
	case StrategyOverwrite:
		return p.resolveByOverwriting(conflictCtx, resolution)
		
	case StrategyMerge:
		return p.resolveByMerging(conflictCtx, resolution)
		
	case StrategyUpdate:
		return p.resolveByUpdating(conflictCtx, resolution)
		
	case StrategyAppend:
		return p.resolveByAppending(conflictCtx, resolution)
		
	case StrategyCustom:
		return p.resolveByCustomRules(ctx, conflictCtx, resolution, config)
		
	default:
		return p.resolveByMerging(conflictCtx, resolution)
	}
}

// getConflictStrategy 获取冲突解决策略
func (p *AdvancedCIPersister) getConflictStrategy(config *ent.CiTypeDiscoveryConfig) ConflictResolutionStrategy {
	if config.ConflictResolution == "" {
		return StrategyMerge
	}
	
	switch strings.ToLower(config.ConflictResolution) {
	case "skip":
		return StrategySkip
	case "overwrite":
		return StrategyOverwrite
	case "merge":
		return StrategyMerge
	case "update":
		return StrategyUpdate
	case "append":
		return StrategyAppend
	case "custom":
		return StrategyCustom
	default:
		return StrategyMerge
	}
}

// resolveByMerging 通过合并解决冲突
func (p *AdvancedCIPersister) resolveByMerging(
	conflictCtx *ConflictContext,
	resolution *ConflictResolution,
) (*ConflictResolution, error) {
	
	// 从现有数据开始
	for k, v := range conflictCtx.ExistingData {
		resolution.ResolvedData[k] = v
	}
	
	// 合并新数据，新数据优先
	for k, v := range conflictCtx.IncomingData {
		if existingValue, exists := resolution.ResolvedData[k]; exists {
			// 对于数组字段，尝试智能合并
			if p.isArrayField(k, v) {
				mergedValue := p.mergeArrays(existingValue, v)
				resolution.ResolvedData[k] = mergedValue
				resolution.ChangedFields = append(resolution.ChangedFields, k)
			} else {
				// 非数组字段直接覆盖
				resolution.ResolvedData[k] = v
				resolution.ChangedFields = append(resolution.ChangedFields, k)
			}
		} else {
			// 新字段直接添加
			resolution.ResolvedData[k] = v
			resolution.ChangedFields = append(resolution.ChangedFields, k)
		}
	}
	
	return resolution, nil
}

// resolveByOverwriting 通过覆盖解决冲突
func (p *AdvancedCIPersister) resolveByOverwriting(
	conflictCtx *ConflictContext,
	resolution *ConflictResolution,
) (*ConflictResolution, error) {
	
	// 完全使用新数据
	for k, v := range conflictCtx.IncomingData {
		resolution.ResolvedData[k] = v
		resolution.ChangedFields = append(resolution.ChangedFields, k)
	}
	
	return resolution, nil
}

// resolveByUpdating 通过更新解决冲突（只更新变化的字段）
func (p *AdvancedCIPersister) resolveByUpdating(
	conflictCtx *ConflictContext,
	resolution *ConflictResolution,
) (*ConflictResolution, error) {
	
	// 从现有数据开始
	for k, v := range conflictCtx.ExistingData {
		resolution.ResolvedData[k] = v
	}
	
	// 只更新实际变化的字段
	for k, newValue := range conflictCtx.IncomingData {
		if existingValue, exists := conflictCtx.ExistingData[k]; exists {
			if !p.changeDetector.deepEqual(existingValue, newValue) {
				resolution.ResolvedData[k] = newValue
				resolution.ChangedFields = append(resolution.ChangedFields, k)
			}
		} else {
			// 新字段
			resolution.ResolvedData[k] = newValue
			resolution.ChangedFields = append(resolution.ChangedFields, k)
		}
	}
	
	return resolution, nil
}

// Helper methods
func (p *AdvancedCIPersister) configureConflictResolution(config *ent.CiTypeDiscoveryConfig) error {
	// 从配置中加载冲突解决规则
	if config.FilterConditions != nil {
		if rules, ok := config.FilterConditions["conflict_rules"]; ok {
			if rulesBytes, err := json.Marshal(rules); err == nil {
				var conflictRules []ConflictRule
				if err := json.Unmarshal(rulesBytes, &conflictRules); err == nil {
					p.conflictResolver.rules = conflictRules
				}
			}
		}
	}
	return nil
}

func (p *AdvancedCIPersister) findExistingCI(ctx context.Context, tx *ent.Tx, ciData types.TransformedCIData) (map[string]interface{}, error) {
	// TODO: 实现基于唯一键的CI查找
	// 这里需要根据实际的CI表结构来实现
	return nil, nil
}

func (p *AdvancedCIPersister) createNewCIWithValidation(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData, result *types.PersistResult) error {
	// TODO: 实现新CI创建逻辑
	result.CreatedCIs++
	result.SuccessCount++
	return nil
}

func (p *AdvancedCIPersister) identifyConflictFields(existing, incoming map[string]interface{}) []string {
	var conflicts []string
	for field, newValue := range incoming {
		if existingValue, exists := existing[field]; exists {
			if !p.changeDetector.deepEqual(existingValue, newValue) {
				conflicts = append(conflicts, field)
			}
		}
	}
	return conflicts
}

func (p *AdvancedCIPersister) applyResolutionAndUpdate(ctx context.Context, tx *ent.Tx, config *ent.CiTypeDiscoveryConfig, ciData types.TransformedCIData, existingData map[string]interface{}, resolution *ConflictResolution, changes []ChangeRecord, result *types.PersistResult) error {
	// TODO: 实现更新逻辑
	result.UpdatedCIs++
	result.SuccessCount++
	return nil
}

func (p *ChangeDetector) isIgnoredField(field string) bool {
	for _, ignored := range p.ignoredFields {
		if field == ignored {
			return true
		}
	}
	return false
}

func (p *ChangeDetector) deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func (p *AdvancedCIPersister) isArrayField(field string, value interface{}) bool {
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Slice || v.Kind() == reflect.Array
}

func (p *AdvancedCIPersister) mergeArrays(existing, incoming interface{}) interface{} {
	// 简单的数组合并逻辑
	existingSlice, ok1 := existing.([]interface{})
	incomingSlice, ok2 := incoming.([]interface{})
	
	if !ok1 || !ok2 {
		return incoming // 如果不是数组，返回新值
	}
	
	// 去重合并
	merged := make([]interface{}, 0, len(existingSlice)+len(incomingSlice))
	seen := make(map[string]bool)
	
	for _, item := range existingSlice {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			merged = append(merged, item)
			seen[key] = true
		}
	}
	
	for _, item := range incomingSlice {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			merged = append(merged, item)
			seen[key] = true
		}
	}
	
	return merged
}

func (p *AdvancedCIPersister) resolveBySkipping(conflictCtx *ConflictContext, resolution *ConflictResolution) (*ConflictResolution, error) {
	// 保持现有数据不变
	for k, v := range conflictCtx.ExistingData {
		resolution.ResolvedData[k] = v
	}
	return resolution, nil
}

func (p *AdvancedCIPersister) resolveByAppending(conflictCtx *ConflictContext, resolution *ConflictResolution) (*ConflictResolution, error) {
	// 从现有数据开始
	for k, v := range conflictCtx.ExistingData {
		resolution.ResolvedData[k] = v
	}
	
	// 对于冲突字段，如果是数组则追加
	for _, field := range conflictCtx.ConflictFields {
		if newValue, exists := conflictCtx.IncomingData[field]; exists {
			if p.isArrayField(field, newValue) {
				existingValue := conflictCtx.ExistingData[field]
				resolution.ResolvedData[field] = p.mergeArrays(existingValue, newValue)
				resolution.ChangedFields = append(resolution.ChangedFields, field)
			} else {
				resolution.ResolvedData[field] = newValue
				resolution.ChangedFields = append(resolution.ChangedFields, field)
			}
		}
	}
	
	return resolution, nil
}

func (p *AdvancedCIPersister) resolveByCustomRules(ctx context.Context, conflictCtx *ConflictContext, resolution *ConflictResolution, config *ent.CiTypeDiscoveryConfig) (*ConflictResolution, error) {
	// 应用自定义规则
	for _, rule := range p.conflictResolver.rules {
		if p.matchesCondition(rule.Condition, conflictCtx) {
			if err := p.applyConflictRule(rule, conflictCtx, resolution); err != nil {
				return nil, fmt.Errorf("failed to apply conflict rule: %w", err)
			}
		}
	}
	return resolution, nil
}

func (p *AdvancedCIPersister) matchesCondition(condition string, conflictCtx *ConflictContext) bool {
	// 简单的条件匹配逻辑
	if condition == "" {
		return true
	}
	
	// TODO: 实现更复杂的条件匹配逻辑
	return true
}

func (p *AdvancedCIPersister) applyConflictRule(rule ConflictRule, conflictCtx *ConflictContext, resolution *ConflictResolution) error {
	// 应用特定字段的冲突规则
	if rule.Field != "" {
		if newValue, exists := conflictCtx.IncomingData[rule.Field]; exists {
			switch rule.Strategy {
			case StrategyOverwrite:
				resolution.ResolvedData[rule.Field] = newValue
			case StrategySkip:
				if existingValue, exists := conflictCtx.ExistingData[rule.Field]; exists {
					resolution.ResolvedData[rule.Field] = existingValue
				}
			case StrategyMerge:
				if existingValue, exists := conflictCtx.ExistingData[rule.Field]; exists {
					resolution.ResolvedData[rule.Field] = p.mergeValues(existingValue, newValue)
				} else {
					resolution.ResolvedData[rule.Field] = newValue
				}
			}
			resolution.ChangedFields = append(resolution.ChangedFields, rule.Field)
		}
	}
	return nil
}

func (p *AdvancedCIPersister) mergeValues(existing, incoming interface{}) interface{} {
	// 智能合并值
	if p.isArrayField("", incoming) {
		return p.mergeArrays(existing, incoming)
	}
	
	// 对于其他类型，优先使用新值
	return incoming
}