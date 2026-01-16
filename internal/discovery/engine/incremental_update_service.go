package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// IncrementalUpdateService 增量更新服务
type IncrementalUpdateService struct {
	db     *ent.Client
	logger logx.Logger
}

// UpdateCheckpoint 更新检查点
type UpdateCheckpoint struct {
	ConfigID        uint64    `json:"config_id"`
	LastUpdateTime  time.Time `json:"last_update_time"`
	LastDataVersion string    `json:"last_data_version"`
	LastChecksum    string    `json:"last_checksum"`
	ProcessedCount  int64     `json:"processed_count"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// IncrementalBatch 增量批次
type IncrementalBatch struct {
	BatchID      string                  `json:"batch_id"`
	ConfigID     uint64                  `json:"config_id"`
	Items        []types.TransformedCIData `json:"items"`
	StartTime    time.Time               `json:"start_time"`
	EndTime      time.Time               `json:"end_time"`
	TotalRecords int                     `json:"total_records"`
	NewRecords   int                     `json:"new_records"`
	UpdatedRecords int                   `json:"updated_records"`
	SkippedRecords int                   `json:"skipped_records"`
}

// DataFingerprint 数据指纹
type DataFingerprint struct {
	ID          string                 `json:"id"`
	CITypeID    uint64                 `json:"ci_type_id"`
	SourceID    string                 `json:"source_id"`
	Checksum    string                 `json:"checksum"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Attributes  map[string]interface{} `json:"attributes"`
	FieldHashes map[string]string      `json:"field_hashes"`
}

// DeltaResult 增量结果
type DeltaResult struct {
	NewItems     []types.TransformedCIData `json:"new_items"`
	UpdatedItems []types.TransformedCIData `json:"updated_items"`
	DeletedItems []string                  `json:"deleted_items"`
	UnchangedItems []string                `json:"unchanged_items"`
	Conflicts    []ConflictContext         `json:"conflicts"`
	Statistics   DeltaStatistics           `json:"statistics"`
}

// DeltaStatistics 增量统计
type DeltaStatistics struct {
	TotalProcessed int           `json:"total_processed"`
	NewCount       int           `json:"new_count"`
	UpdatedCount   int           `json:"updated_count"`
	DeletedCount   int           `json:"deleted_count"`
	SkippedCount   int           `json:"skipped_count"`
	ConflictCount  int           `json:"conflict_count"`
	ProcessingTime time.Duration `json:"processing_time"`
	ThroughputPerSecond float64  `json:"throughput_per_second"`
}

// NewIncrementalUpdateService 创建增量更新服务
func NewIncrementalUpdateService(db *ent.Client) *IncrementalUpdateService {
	return &IncrementalUpdateService{
		db:     db,
		logger: logx.WithContext(context.Background()),
	}
}

// ProcessIncrementalUpdate 处理增量更新
func (s *IncrementalUpdateService) ProcessIncrementalUpdate(
	ctx context.Context,
	config *ent.CiTypeDiscoveryConfig,
	currentData []types.TransformedCIData,
) (*DeltaResult, error) {
	
	startTime := time.Now()
	
	// 1. 获取上次更新的检查点
	checkpoint, err := s.getLastCheckpoint(ctx, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last checkpoint: %w", err)
	}
	
	// 2. 生成当前数据的指纹
	currentFingerprints, err := s.generateDataFingerprints(currentData, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate current fingerprints: %w", err)
	}
	
	// 3. 获取上次的数据指纹
	lastFingerprints, err := s.getLastFingerprints(ctx, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last fingerprints: %w", err)
	}
	
	// 4. 计算增量差异
	deltaResult, err := s.calculateDelta(currentFingerprints, lastFingerprints, currentData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate delta: %w", err)
	}
	
	// 5. 更新统计信息
	deltaResult.Statistics.ProcessingTime = time.Since(startTime)
	deltaResult.Statistics.ThroughputPerSecond = float64(deltaResult.Statistics.TotalProcessed) / deltaResult.Statistics.ProcessingTime.Seconds()
	
	// 6. 保存新的检查点和指纹
	if err := s.saveCheckpointAndFingerprints(ctx, config.ID, checkpoint, currentFingerprints, deltaResult); err != nil {
		s.logger.Errorw("Failed to save checkpoint and fingerprints",
			logx.Field("config_id", config.ID),
			logx.Field("error", err))
	}
	
	return deltaResult, nil
}

// calculateDelta 计算增量差异
func (s *IncrementalUpdateService) calculateDelta(
	current, last map[string]*DataFingerprint,
	currentData []types.TransformedCIData,
) (*DeltaResult, error) {
	
	result := &DeltaResult{
		NewItems:       []types.TransformedCIData{},
		UpdatedItems:   []types.TransformedCIData{},
		DeletedItems:   []string{},
		UnchangedItems: []string{},
		Conflicts:      []ConflictContext{},
		Statistics:     DeltaStatistics{},
	}
	
	// 创建当前数据的映射以便快速查找
	currentDataMap := make(map[string]types.TransformedCIData)
	for _, data := range currentData {
		currentDataMap[data.SourceID] = data
	}
	
	// 1. 检查新增和更新的项目
	for sourceID, currentFP := range current {
		lastFP, existsInLast := last[sourceID]
		
		if !existsInLast {
			// 新增项目
			if data, exists := currentDataMap[sourceID]; exists {
				result.NewItems = append(result.NewItems, data)
				result.Statistics.NewCount++
			}
		} else {
			// 检查是否有更新
			if s.hasDataChanged(currentFP, lastFP) {
				if data, exists := currentDataMap[sourceID]; exists {
					// 检测字段级别的变更
					fieldChanges := s.detectFieldLevelChanges(currentFP, lastFP)
					
					// 如果有冲突，记录冲突
					if s.hasConflicts(fieldChanges) {
						conflict := ConflictContext{
							ExistingData:   lastFP.Attributes,
							IncomingData:   currentFP.Attributes,
							ConflictFields: s.extractConflictFields(fieldChanges),
							Timestamp:      time.Now(),
						}
						result.Conflicts = append(result.Conflicts, conflict)
						result.Statistics.ConflictCount++
					}
					
					result.UpdatedItems = append(result.UpdatedItems, data)
					result.Statistics.UpdatedCount++
				}
			} else {
				// 无变化
				result.UnchangedItems = append(result.UnchangedItems, sourceID)
			}
		}
	}
	
	// 2. 检查删除的项目
	for sourceID := range last {
		if _, existsInCurrent := current[sourceID]; !existsInCurrent {
			result.DeletedItems = append(result.DeletedItems, sourceID)
			result.Statistics.DeletedCount++
		}
	}
	
	result.Statistics.TotalProcessed = len(current)
	result.Statistics.SkippedCount = len(result.UnchangedItems)
	
	return result, nil
}

// generateDataFingerprints 生成数据指纹
func (s *IncrementalUpdateService) generateDataFingerprints(
	data []types.TransformedCIData,
	config *ent.CiTypeDiscoveryConfig,
) (map[string]*DataFingerprint, error) {
	
	fingerprints := make(map[string]*DataFingerprint)
	
	for _, item := range data {
		fp, err := s.createFingerprint(item, config)
		if err != nil {
			s.logger.Errorw("Failed to create fingerprint",
				logx.Field("source_id", item.SourceID),
				logx.Field("error", err))
			continue
		}
		fingerprints[item.SourceID] = fp
	}
	
	return fingerprints, nil
}

// createFingerprint 创建单个数据指纹
func (s *IncrementalUpdateService) createFingerprint(
	data types.TransformedCIData,
	config *ent.CiTypeDiscoveryConfig,
) (*DataFingerprint, error) {
	
	// 1. 计算整体校验和
	checksum, err := s.calculateChecksum(data.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// 2. 计算字段级别的哈希
	fieldHashes, err := s.calculateFieldHashes(data.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate field hashes: %w", err)
	}
	
	// 3. 生成版本号
	version := s.generateVersion(data.Attributes, checksum)
	
	fingerprint := &DataFingerprint{
		ID:          fmt.Sprintf("%d_%s", config.ID, data.SourceID),
		CITypeID:    data.CITypeID,
		SourceID:    data.SourceID,
		Checksum:    checksum,
		Version:     version,
		Timestamp:   time.Now(),
		Attributes:  data.Attributes,
		FieldHashes: fieldHashes,
	}
	
	return fingerprint, nil
}

// calculateChecksum 计算数据校验和
func (s *IncrementalUpdateService) calculateChecksum(data map[string]interface{}) (string, error) {
	// 1. 将数据序列化为JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}
	
	// 2. 计算SHA256哈希
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:]), nil
}

// calculateFieldHashes 计算字段级别哈希
func (s *IncrementalUpdateService) calculateFieldHashes(data map[string]interface{}) (map[string]string, error) {
	fieldHashes := make(map[string]string)
	
	for field, value := range data {
		// 序列化单个字段的值
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			s.logger.Errorw("Failed to marshal field value",
				logx.Field("field", field),
				logx.Field("error", err))
			continue
		}
		
		// 计算字段哈希
		hash := sha256.Sum256(jsonBytes)
		fieldHashes[field] = hex.EncodeToString(hash[:])
	}
	
	return fieldHashes, nil
}

// generateVersion 生成版本号
func (s *IncrementalUpdateService) generateVersion(data map[string]interface{}, checksum string) string {
	// 使用时间戳和校验和的组合作为版本号
	timestamp := time.Now().Unix()
	return fmt.Sprintf("v%d_%s", timestamp, checksum[:8])
}

// hasDataChanged 检查数据是否有变化
func (s *IncrementalUpdateService) hasDataChanged(current, last *DataFingerprint) bool {
	return current.Checksum != last.Checksum
}

// detectFieldLevelChanges 检测字段级别的变更
func (s *IncrementalUpdateService) detectFieldLevelChanges(current, last *DataFingerprint) map[string]FieldChange {
	changes := make(map[string]FieldChange)
	
	// 检查当前字段
	for field, currentHash := range current.FieldHashes {
		lastHash, exists := last.FieldHashes[field]
		
		if !exists {
			// 新增字段
			changes[field] = FieldChange{
				Type:     "added",
				OldValue: nil,
				NewValue: current.Attributes[field],
				OldHash:  "",
				NewHash:  currentHash,
			}
		} else if currentHash != lastHash {
			// 修改字段
			changes[field] = FieldChange{
				Type:     "modified",
				OldValue: last.Attributes[field],
				NewValue: current.Attributes[field],
				OldHash:  lastHash,
				NewHash:  currentHash,
			}
		}
	}
	
	// 检查删除的字段
	for field, lastHash := range last.FieldHashes {
		if _, exists := current.FieldHashes[field]; !exists {
			changes[field] = FieldChange{
				Type:     "deleted",
				OldValue: last.Attributes[field],
				NewValue: nil,
				OldHash:  lastHash,
				NewHash:  "",
			}
		}
	}
	
	return changes
}

// FieldChange 字段变更信息
type FieldChange struct {
	Type     string      `json:"type"`      // added, modified, deleted
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	OldHash  string      `json:"old_hash"`
	NewHash  string      `json:"new_hash"`
}

// hasConflicts 检查是否有冲突
func (s *IncrementalUpdateService) hasConflicts(changes map[string]FieldChange) bool {
	// 定义冲突检测规则
	conflictFields := []string{"name", "status", "owner", "critical_attribute"}
	
	for field := range changes {
		for _, conflictField := range conflictFields {
			if field == conflictField {
				return true
			}
		}
	}
	
	return false
}

// extractConflictFields 提取冲突字段
func (s *IncrementalUpdateService) extractConflictFields(changes map[string]FieldChange) []string {
	var conflicts []string
	conflictFields := []string{"name", "status", "owner", "critical_attribute"}
	
	for field := range changes {
		for _, conflictField := range conflictFields {
			if field == conflictField {
				conflicts = append(conflicts, field)
			}
		}
	}
	
	return conflicts
}

// getLastCheckpoint 获取上次更新检查点
func (s *IncrementalUpdateService) getLastCheckpoint(ctx context.Context, configID uint64) (*UpdateCheckpoint, error) {
	// TODO: 从数据库加载检查点
	// 这里需要一个专门的表来存储检查点信息
	
	// 返回默认检查点
	return &UpdateCheckpoint{
		ConfigID:        configID,
		LastUpdateTime:  time.Time{},
		LastDataVersion: "",
		LastChecksum:    "",
		ProcessedCount:  0,
		Metadata:        make(map[string]interface{}),
	}, nil
}

// getLastFingerprints 获取上次的数据指纹
func (s *IncrementalUpdateService) getLastFingerprints(ctx context.Context, configID uint64) (map[string]*DataFingerprint, error) {
	// TODO: 从数据库加载上次的指纹数据
	// 这里需要一个专门的表来存储指纹信息
	
	return make(map[string]*DataFingerprint), nil
}

// saveCheckpointAndFingerprints 保存检查点和指纹
func (s *IncrementalUpdateService) saveCheckpointAndFingerprints(
	ctx context.Context,
	configID uint64,
	checkpoint *UpdateCheckpoint,
	fingerprints map[string]*DataFingerprint,
	result *DeltaResult,
) error {
	
	// 更新检查点
	checkpoint.LastUpdateTime = time.Now()
	checkpoint.ProcessedCount += int64(result.Statistics.TotalProcessed)
	
	// 计算新的整体校验和
	sortedKeys := make([]string, 0, len(fingerprints))
	for key := range fingerprints {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	
	var allChecksums []string
	for _, key := range sortedKeys {
		allChecksums = append(allChecksums, fingerprints[key].Checksum)
	}
	
	overallChecksum := s.calculateOverallChecksum(allChecksums)
	checkpoint.LastChecksum = overallChecksum
	
	// TODO: 保存到数据库
	// 1. 保存检查点到 discovery_checkpoints 表
	// 2. 保存指纹到 discovery_fingerprints 表
	
	s.logger.Infow("Checkpoint and fingerprints saved",
		logx.Field("config_id", configID),
		logx.Field("processed_count", checkpoint.ProcessedCount),
		logx.Field("fingerprints_count", len(fingerprints)))
	
	return nil
}

// calculateOverallChecksum 计算整体校验和
func (s *IncrementalUpdateService) calculateOverallChecksum(checksums []string) string {
	allChecksums := strings.Join(checksums, "|")
	hash := sha256.Sum256([]byte(allChecksums))
	return hex.EncodeToString(hash[:])
}

// BatchProcessIncrementalUpdates 批量处理增量更新
func (s *IncrementalUpdateService) BatchProcessIncrementalUpdates(
	ctx context.Context,
	config *ent.CiTypeDiscoveryConfig,
	data []types.TransformedCIData,
	batchSize int,
) ([]*DeltaResult, error) {
	
	var results []*DeltaResult
	
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		
		batch := data[i:end]
		batchID := fmt.Sprintf("batch_%d_%d", config.ID, time.Now().Unix())
		
		s.logger.Infow("Processing incremental batch",
			logx.Field("batch_id", batchID),
			logx.Field("batch_size", len(batch)),
			logx.Field("start_index", i))
		
		result, err := s.ProcessIncrementalUpdate(ctx, config, batch)
		if err != nil {
			s.logger.Errorw("Failed to process incremental batch",
				logx.Field("batch_id", batchID),
				logx.Field("error", err))
			return nil, fmt.Errorf("failed to process batch %s: %w", batchID, err)
		}
		
		results = append(results, result)
	}
	
	return results, nil
}