package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisPermissionCache Redis权限缓存实现
type RedisPermissionCache struct {
	client redis.UniversalClient
	prefix string
	logger *logrus.Logger
}

// NewRedisPermissionCache 创建Redis权限缓存
func NewRedisPermissionCache(client redis.UniversalClient, prefix string, logger *logrus.Logger) *RedisPermissionCache {
	return &RedisPermissionCache{
		client: client,
		prefix: prefix,
		logger: logger,
	}
}

// Get 获取权限检查结果缓存
func (c *RedisPermissionCache) Get(ctx context.Context, key string) (*PermissionResult, error) {
	fullKey := c.buildKey(key)

	data, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("获取缓存失败: %w", err)
	}

	var result PermissionResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("解析缓存数据失败: %w", err)
	}

	return &result, nil
}

// Set 设置权限检查结果缓存
func (c *RedisPermissionCache) Set(ctx context.Context, key string, result *PermissionResult, ttl time.Duration) error {
	fullKey := c.buildKey(key)

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (c *RedisPermissionCache) Delete(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)

	// 获取匹配的键
	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("获取匹配键失败: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 批量删除
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("批量删除缓存失败: %w", err)
	}

	c.logger.Infof("删除缓存键数量: %d", len(keys))
	return nil
}

// buildKey 构建完整的缓存键
func (c *RedisPermissionCache) buildKey(key string) string {
	return fmt.Sprintf("%s:perm:%s", c.prefix, key)
}

// GetUserPermissions 获取用户权限缓存
func (c *RedisPermissionCache) GetUserPermissions(ctx context.Context, userID string) ([]*EffectivePermission, error) {
	key := fmt.Sprintf("user_perms:%s", userID)
	fullKey := c.buildKey(key)

	data, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("获取用户权限缓存失败: %w", err)
	}

	var permissions []*EffectivePermission
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		return nil, fmt.Errorf("解析用户权限数据失败: %w", err)
	}

	return permissions, nil
}

// SetUserPermissions 设置用户权限缓存
func (c *RedisPermissionCache) SetUserPermissions(ctx context.Context, userID string, perms []*EffectivePermission, ttl time.Duration) error {
	key := fmt.Sprintf("user_perms:%s", userID)
	fullKey := c.buildKey(key)

	data, err := json.Marshal(perms)
	if err != nil {
		return fmt.Errorf("序列化用户权限数据失败: %w", err)
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("设置用户权限缓存失败: %w", err)
	}

	return nil
}

// InvalidateUserCache 失效用户相关缓存
func (c *RedisPermissionCache) InvalidateUserCache(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("*:%s:*", userID),
		fmt.Sprintf("user_perms:%s", userID),
	}

	for _, pattern := range patterns {
		if err := c.Delete(ctx, pattern); err != nil {
			c.logger.Errorf("删除用户缓存失败 %s: %v", pattern, err)
		}
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *RedisPermissionCache) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	info, err := c.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis内存信息失败: %w", err)
	}

	keyCount, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis键数量失败: %w", err)
	}

	return &CacheStats{
		KeyCount:   keyCount,
		MemoryInfo: info,
		Timestamp:  time.Now(),
	}, nil
}

// CacheStats 缓存统计信息
type CacheStats struct {
	KeyCount   int64     `json:"key_count"`
	MemoryInfo string    `json:"memory_info"`
	Timestamp  time.Time `json:"timestamp"`
}

// EffectivePermission 有效权限
type EffectivePermission struct {
	ResourceType    string    `json:"resource_type"`
	ResourceID      string    `json:"resource_id"`
	AllowedOps      uint64    `json:"allowed_ops"`
	PermissionLevel string    `json:"permission_level"`
	HasDataFilters  bool      `json:"has_data_filters"`
	HasFieldMasks   bool      `json:"has_field_masks"`
	ExpiresAt       time.Time `json:"expires_at"`
}
