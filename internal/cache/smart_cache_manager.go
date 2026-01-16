package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/types"
)

// SmartCacheManager 智能多层缓存管理器
// 提供L1本地缓存 + L2分布式缓存的两级缓存策略
type SmartCacheManager struct {
	localCache       *LocalCacheManager
	distributedCache *DistributedCacheManager
	config           *CacheConfig
	mu               sync.RWMutex
	stats            *CacheStats
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 本地缓存配置
	LocalCacheEnabled bool          `json:"localCacheEnabled"`
	LocalCacheSize    int           `json:"localCacheSize"` // 最大条目数
	LocalCacheTTL     time.Duration `json:"localCacheTTL"`  // 本地缓存TTL

	// 分布式缓存配置
	DistributedCacheEnabled bool          `json:"distributedCacheEnabled"`
	DistributedCacheTTL     time.Duration `json:"distributedCacheTTL"` // 分布式缓存TTL

	// 缓存策略配置
	CacheStrategy    string        `json:"cacheStrategy"`    // write-through, write-back, write-around
	EvictionPolicy   string        `json:"evictionPolicy"`   // lru, lfu, fifo
	RefreshThreshold time.Duration `json:"refreshThreshold"` // 主动刷新阈值

	// 性能配置
	BatchSize           int    `json:"batchSize"`           // 批量操作大小
	CompressEnabled     bool   `json:"compressEnabled"`     // 是否启用压缩
	SerializationFormat string `json:"serializationFormat"` // json, msgpack, protobuf
}

// CacheStats 缓存统计信息
type CacheStats struct {
	LocalHits         int64     `json:"localHits"`
	LocalMisses       int64     `json:"localMisses"`
	DistributedHits   int64     `json:"distributedHits"`
	DistributedMisses int64     `json:"distributedMisses"`
	TotalRequests     int64     `json:"totalRequests"`
	LastResetTime     time.Time `json:"lastResetTime"`
	mu                sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Key         string        `json:"key"`
	Value       interface{}   `json:"value"`
	TTL         time.Duration `json:"ttl"`
	CreatedAt   time.Time     `json:"createdAt"`
	AccessedAt  time.Time     `json:"accessedAt"`
	AccessCount int64         `json:"accessCount"`
}

// NewSmartCacheManager 创建智能缓存管理器
func NewSmartCacheManager(redisClient *redis.Redis, config *CacheConfig) *SmartCacheManager {
	if config == nil {
		config = getDefaultCacheConfig()
	}

	manager := &SmartCacheManager{
		config: config,
		stats:  &CacheStats{LastResetTime: time.Now()},
	}

	// 初始化本地缓存
	if config.LocalCacheEnabled {
		manager.localCache = NewLocalCacheManager(&LocalCacheConfig{
			MaxSize:        config.LocalCacheSize,
			DefaultTTL:     config.LocalCacheTTL,
			EvictionPolicy: config.EvictionPolicy,
		})
	}

	// 初始化分布式缓存
	if config.DistributedCacheEnabled && redisClient != nil {
		manager.distributedCache = NewDistributedCacheManager(redisClient, &DistributedCacheConfig{
			DefaultTTL:          config.DistributedCacheTTL,
			CompressEnabled:     config.CompressEnabled,
			SerializationFormat: config.SerializationFormat,
		})
	}

	return manager
}

// Get 获取缓存数据（实现多层缓存策略）
func (m *SmartCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	m.stats.incrementTotalRequests()

	cacheKey := m.buildCacheKey(key)

	// L1: 检查本地缓存
	if m.config.LocalCacheEnabled && m.localCache != nil {
		if value, err := m.localCache.Get(ctx, cacheKey); err == nil && value != nil {
			m.stats.incrementLocalHits()
			logx.Debugf("缓存命中 [L1]: %s", cacheKey)
			return value, nil
		}
		m.stats.incrementLocalMisses()
	}

	// L2: 检查分布式缓存
	if m.config.DistributedCacheEnabled && m.distributedCache != nil {
		if value, err := m.distributedCache.Get(ctx, cacheKey); err == nil && value != nil {
			m.stats.incrementDistributedHits()
			logx.Debugf("缓存命中 [L2]: %s", cacheKey)

			// 回填到本地缓存
			if m.config.LocalCacheEnabled && m.localCache != nil {
				_ = m.localCache.Set(ctx, cacheKey, value, m.config.LocalCacheTTL)
			}

			return value, nil
		}
		m.stats.incrementDistributedMisses()
	}

	logx.Debugf("缓存未命中: %s", cacheKey)
	return nil, fmt.Errorf("缓存未命中: %s", key)
}

// Set 设置缓存数据
func (m *SmartCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cacheKey := m.buildCacheKey(key)

	var errs []error

	// 设置分布式缓存
	if m.config.DistributedCacheEnabled && m.distributedCache != nil {
		distributedTTL := ttl
		if distributedTTL == 0 {
			distributedTTL = m.config.DistributedCacheTTL
		}

		if err := m.distributedCache.Set(ctx, cacheKey, value, distributedTTL); err != nil {
			logx.Errorf("设置分布式缓存失败 [%s]: %v", cacheKey, err)
			errs = append(errs, err)
		}
	}

	// 设置本地缓存
	if m.config.LocalCacheEnabled && m.localCache != nil {
		localTTL := ttl
		if localTTL == 0 {
			localTTL = m.config.LocalCacheTTL
		}
		// 本地缓存TTL应该小于等于分布式缓存TTL
		if localTTL > m.config.LocalCacheTTL {
			localTTL = m.config.LocalCacheTTL
		}

		if err := m.localCache.Set(ctx, cacheKey, value, localTTL); err != nil {
			logx.Errorf("设置本地缓存失败 [%s]: %v", cacheKey, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分缓存设置失败: %v", errs)
	}

	return nil
}

// Delete 删除缓存数据
func (m *SmartCacheManager) Delete(ctx context.Context, key string) error {
	cacheKey := m.buildCacheKey(key)

	var errs []error

	// 删除本地缓存
	if m.config.LocalCacheEnabled && m.localCache != nil {
		if err := m.localCache.Delete(ctx, cacheKey); err != nil {
			logx.Errorf("删除本地缓存失败 [%s]: %v", cacheKey, err)
			errs = append(errs, err)
		}
	}

	// 删除分布式缓存
	if m.config.DistributedCacheEnabled && m.distributedCache != nil {
		if err := m.distributedCache.Delete(ctx, cacheKey); err != nil {
			logx.Errorf("删除分布式缓存失败 [%s]: %v", cacheKey, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分缓存删除失败: %v", errs)
	}

	return nil
}

// InvalidatePattern 批量删除匹配模式的缓存
func (m *SmartCacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	cachePattern := m.buildCacheKey(pattern)

	var errs []error

	// 清理本地缓存
	if m.config.LocalCacheEnabled && m.localCache != nil {
		if err := m.localCache.InvalidatePattern(ctx, cachePattern); err != nil {
			logx.Errorf("清理本地缓存模式失败 [%s]: %v", cachePattern, err)
			errs = append(errs, err)
		}
	}

	// 清理分布式缓存
	if m.config.DistributedCacheEnabled && m.distributedCache != nil {
		if err := m.distributedCache.InvalidatePattern(ctx, cachePattern); err != nil {
			logx.Errorf("清理分布式缓存模式失败 [%s]: %v", cachePattern, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分缓存模式清理失败: %v", errs)
	}

	return nil
}

// GetOrSet 获取缓存，如果不存在则设置
func (m *SmartCacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (interface{}, error)) (interface{}, error) {
	// 先尝试获取缓存
	if value, err := m.Get(ctx, key); err == nil {
		return value, nil
	}

	// 缓存不存在，使用加载器获取数据
	value, err := loader()
	if err != nil {
		return nil, fmt.Errorf("数据加载失败: %w", err)
	}

	// 设置缓存
	if setErr := m.Set(ctx, key, value, ttl); setErr != nil {
		logx.Errorf("设置缓存失败 [%s]: %v", key, setErr)
		// 不影响返回结果
	}

	return value, nil
}

// 资产选择器专用缓存方法

// GetAssetTypes 获取资产类型缓存
func (m *SmartCacheManager) GetAssetTypes(ctx context.Context, filter *types.AssetTypeFilter) ([]*types.AssetType, error) {
	key := m.buildAssetTypesKey(filter)

	value, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if assetTypes, ok := value.([]*types.AssetType); ok {
		return assetTypes, nil
	}

	return nil, fmt.Errorf("缓存数据类型错误")
}

// SetAssetTypes 设置资产类型缓存
func (m *SmartCacheManager) SetAssetTypes(ctx context.Context, filter *types.AssetTypeFilter, assetTypes []*types.AssetType) error {
	key := m.buildAssetTypesKey(filter)
	ttl := 10 * time.Minute // 资产类型缓存较长时间
	return m.Set(ctx, key, assetTypes, ttl)
}

// GetAssets 获取资产列表缓存
func (m *SmartCacheManager) GetAssets(ctx context.Context, query *types.AssetQuery) (*types.AssetListResponse, error) {
	key := m.buildAssetsKey(query)

	value, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if response, ok := value.(*types.AssetListResponse); ok {
		return response, nil
	}

	return nil, fmt.Errorf("缓存数据类型错误")
}

// SetAssets 设置资产列表缓存
func (m *SmartCacheManager) SetAssets(ctx context.Context, query *types.AssetQuery, response *types.AssetListResponse) error {
	key := m.buildAssetsKey(query)
	ttl := 5 * time.Minute // 资产列表缓存较短时间
	return m.Set(ctx, key, response, ttl)
}

// InvalidateAssetCache 清理资产相关缓存
func (m *SmartCacheManager) InvalidateAssetCache(ctx context.Context, assetTypeID string) error {
	patterns := []string{
		fmt.Sprintf("asset_types:*"),
		fmt.Sprintf("assets:type_%s:*", assetTypeID),
		fmt.Sprintf("asset_detail:%s:*", assetTypeID),
	}

	for _, pattern := range patterns {
		if err := m.InvalidatePattern(ctx, pattern); err != nil {
			logx.Errorf("清理资产缓存失败 [%s]: %v", pattern, err)
		}
	}

	return nil
}

// GetStats 获取缓存统计信息
func (m *SmartCacheManager) GetStats() *CacheStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	// 复制统计信息
	stats := *m.stats
	return &stats
}

// ResetStats 重置缓存统计信息
func (m *SmartCacheManager) ResetStats() {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats.LocalHits = 0
	m.stats.LocalMisses = 0
	m.stats.DistributedHits = 0
	m.stats.DistributedMisses = 0
	m.stats.TotalRequests = 0
	m.stats.LastResetTime = time.Now()
}

// GetHitRate 获取缓存命中率
func (m *SmartCacheManager) GetHitRate() (float64, float64, float64) {
	stats := m.GetStats()

	if stats.TotalRequests == 0 {
		return 0, 0, 0
	}

	localHitRate := float64(stats.LocalHits) / float64(stats.TotalRequests)
	distributedHitRate := float64(stats.DistributedHits) / float64(stats.TotalRequests)
	totalHitRate := float64(stats.LocalHits+stats.DistributedHits) / float64(stats.TotalRequests)

	return localHitRate, distributedHitRate, totalHitRate
}

// 私有方法

// buildCacheKey 构建缓存键
func (m *SmartCacheManager) buildCacheKey(key string) string {
	return fmt.Sprintf("asset_selector:%s", key)
}

// buildAssetTypesKey 构建资产类型缓存键
func (m *SmartCacheManager) buildAssetTypesKey(filter *types.AssetTypeFilter) string {
	filterHash := m.hashObject(filter)
	return fmt.Sprintf("asset_types:%s", filterHash)
}

// buildAssetsKey 构建资产列表缓存键
func (m *SmartCacheManager) buildAssetsKey(query *types.AssetQuery) string {
	queryHash := m.hashObject(query)
	return fmt.Sprintf("assets:%s", queryHash)
}

// hashObject 计算对象哈希值
func (m *SmartCacheManager) hashObject(obj interface{}) string {
	data, _ := json.Marshal(obj)
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// 统计方法
func (s *CacheStats) incrementTotalRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
}

func (s *CacheStats) incrementLocalHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LocalHits++
}

func (s *CacheStats) incrementLocalMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LocalMisses++
}

func (s *CacheStats) incrementDistributedHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DistributedHits++
}

func (s *CacheStats) incrementDistributedMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DistributedMisses++
}

// 默认配置
func getDefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		LocalCacheEnabled:       true,
		LocalCacheSize:          1000,
		LocalCacheTTL:           5 * time.Minute,
		DistributedCacheEnabled: true,
		DistributedCacheTTL:     15 * time.Minute,
		CacheStrategy:           "write-through",
		EvictionPolicy:          "lru",
		RefreshThreshold:        2 * time.Minute,
		BatchSize:               100,
		CompressEnabled:         true,
		SerializationFormat:     "json",
	}
}
