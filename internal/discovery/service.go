package discovery

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/engine"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/provider"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// Service 发现服务
type Service struct {
	engine *engine.DiscoveryEngine
	logger logx.Logger
}

// NewService 创建发现服务
func NewService(db *ent.Client) *Service {
	discoveryEngine := engine.NewDiscoveryEngine(db)
	
	// 注册默认提供者
	registerDefaultProviders(discoveryEngine)

	return &Service{
		engine: discoveryEngine,
		logger: logx.WithContext(context.Background()),
	}
}

// registerDefaultProviders 注册默认提供者
func registerDefaultProviders(engine *engine.DiscoveryEngine) {
	// 注册数据库提供者
	dbProvider := provider.NewDatabaseProvider()
	if err := engine.RegisterProvider(dbProvider); err != nil {
		logx.Errorf("Failed to register database provider: %v", err)
	}

	// 这里可以注册更多提供者
	// - HTTP API提供者
	// - LDAP提供者
	// - 文件系统提供者
	// - 云服务提供者等
}

// ExecuteDiscovery 执行发现任务
func (s *Service) ExecuteDiscovery(ctx context.Context, configID uint64) (*types.DiscoveryResult, error) {
	s.logger.Infof("Starting discovery execution for config ID: %d", configID)
	
	result, err := s.engine.ExecuteDiscovery(ctx, configID)
	if err != nil {
		s.logger.Errorf("Discovery execution failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Discovery execution started with ID: %s", result.ExecutionID)
	return result, nil
}

// GetDiscoveryStatus 获取发现状态
func (s *Service) GetDiscoveryStatus(ctx context.Context, executionID string) (*types.DiscoveryResult, error) {
	// 从数据库查询执行状态
	// 这里需要实现从execution_history表查询状态的逻辑
	return nil, fmt.Errorf("not implemented")
}

// CancelDiscovery 取消发现任务
func (s *Service) CancelDiscovery(ctx context.Context, executionID string) error {
	// 实现取消逻辑
	s.logger.Infof("Cancelling discovery execution: %s", executionID)
	return fmt.Errorf("not implemented")
}

// ListProviders 列出可用的提供者
func (s *Service) ListProviders() map[string][]string {
	return s.engine.ListProviders()
}

// ValidateProviderConfig 验证提供者配置
func (s *Service) ValidateProviderConfig(providerType, providerID string, config map[string]interface{}) error {
	return s.engine.ValidateProviderConfig(providerType, providerID, config)
}

// GetProviderSchema 获取提供者配置结构
func (s *Service) GetProviderSchema(providerType, providerID string) (map[string]interface{}, error) {
	provider, err := s.engine.GetProvider(providerType, providerID)
	if err != nil {
		return nil, err
	}
	
	return provider.GetConfigSchema(), nil
}

// TestConnection 测试数据源连接
func (s *Service) TestConnection(ctx context.Context, providerType, providerID string, config map[string]interface{}) error {
	provider, err := s.engine.GetProvider(providerType, providerID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// 创建临时连接进行测试
	dataSource, err := provider.Connect(ctx, config)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer dataSource.Close()

	// 验证连接
	if err := dataSource.Validate(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// GetDataSourceSchema 获取数据源结构信息
func (s *Service) GetDataSourceSchema(ctx context.Context, providerType, providerID string, config map[string]interface{}) (map[string]interface{}, error) {
	provider, err := s.engine.GetProvider(providerType, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	dataSource, err := provider.Connect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer dataSource.Close()

	return dataSource.GetSchema(ctx)
}