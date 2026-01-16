package provider

import (
	"context"
	"fmt"
	"sync"
)

// DataSource 数据源接口
type DataSource interface {
	// Connect 连接到数据源
	Connect(ctx context.Context, config map[string]interface{}) error
	// Discover 执行数据发现
	Discover(ctx context.Context, rules map[string]interface{}) ([]map[string]interface{}, error)
	// Close 关闭连接
	Close() error
	// Validate 验证连接配置
	Validate(config map[string]interface{}) error
	// GetSchema 获取数据源结构信息
	GetSchema(ctx context.Context) (map[string]interface{}, error)
}

// Provider 发现提供者接口
type Provider interface {
	// GetName 获取提供者名称
	GetName() string
	// GetType 获取提供者类型
	GetType() string
	// Connect 创建数据源连接
	Connect(ctx context.Context, config map[string]interface{}) (DataSource, error)
	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error
	// GetConfigSchema 获取配置结构定义
	GetConfigSchema() map[string]interface{}
}

// Registry 提供者注册中心
type Registry struct {
	providers map[string]map[string]Provider // [type][id]Provider
	mu        sync.RWMutex
}

// NewRegistry 创建新的注册中心
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]map[string]Provider),
	}
}

// RegisterProvider 注册提供者
func (r *Registry) RegisterProvider(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	providerType := provider.GetType()
	providerID := provider.GetName()

	if r.providers[providerType] == nil {
		r.providers[providerType] = make(map[string]Provider)
	}

	if _, exists := r.providers[providerType][providerID]; exists {
		return fmt.Errorf("provider %s:%s already registered", providerType, providerID)
	}

	r.providers[providerType][providerID] = provider
	return nil
}

// GetProvider 获取提供者
func (r *Registry) GetProvider(providerType, providerID string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeProviders, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider type %s not found", providerType)
	}

	provider, exists := typeProviders[providerID]
	if !exists {
		return nil, fmt.Errorf("provider %s:%s not found", providerType, providerID)
	}

	return provider, nil
}

// ListProviders 列出所有提供者
func (r *Registry) ListProviders() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string)
	for providerType, providers := range r.providers {
		var providerIDs []string
		for providerID := range providers {
			providerIDs = append(providerIDs, providerID)
		}
		result[providerType] = providerIDs
	}

	return result
}

// GetProvidersByType 按类型获取提供者列表
func (r *Registry) GetProvidersByType(providerType string) ([]Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeProviders, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider type %s not found", providerType)
	}

	var providers []Provider
	for _, provider := range typeProviders {
		providers = append(providers, provider)
	}

	return providers, nil
}

// UnregisterProvider 注销提供者
func (r *Registry) UnregisterProvider(providerType, providerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	typeProviders, exists := r.providers[providerType]
	if !exists {
		return fmt.Errorf("provider type %s not found", providerType)
	}

	if _, exists := typeProviders[providerID]; !exists {
		return fmt.Errorf("provider %s:%s not found", providerType, providerID)
	}

	delete(typeProviders, providerID)

	// 如果该类型下没有提供者了，删除整个类型
	if len(typeProviders) == 0 {
		delete(r.providers, providerType)
	}

	return nil
}

// ValidateProvider 验证提供者配置
func (r *Registry) ValidateProvider(providerType, providerID string, config map[string]interface{}) error {
	provider, err := r.GetProvider(providerType, providerID)
	if err != nil {
		return err
	}

	return provider.ValidateConfig(config)
}