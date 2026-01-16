package input

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
)

// DiscoveryInputAdapter 自动发现输入适配器 (功能已迁移到Agent)
type DiscoveryInputAdapter struct {
	*BaseAdapter
	config *DiscoveryConfig
}

// NewDiscoveryInputAdapter 创建自动发现输入适配器
func NewDiscoveryInputAdapter(svcCtx *svc.ServiceContext, config *DiscoveryConfig) *DiscoveryInputAdapter {
	if config == nil {
		config = &DiscoveryConfig{
			DiscoveryRules: []*DiscoveryRule{},
			ScanInterval:   5 * time.Minute,
			MaxTargets:     1000,
			Timeout:        30 * time.Second,
			AgentConfig:    make(map[string]interface{}),
		}
	}

	base := NewBaseAdapter("discovery", "1.0.0", svcCtx)
	base.config = &ConfigSchema{
		Type:        "discovery",
		Version:     "1.0.0",
		Description: "自动发现适配器 (功能已迁移到Agent服务)",
		Properties: map[string]interface{}{
			"status":         "deprecated",
			"migration_note": "请使用Agent服务进行网络发现",
			"scan_interval":  config.ScanInterval.String(),
			"max_targets":    config.MaxTargets,
			"timeout":        config.Timeout.String(),
		},
		Required: []string{},
	}

	return &DiscoveryInputAdapter{
		BaseAdapter: base,
		config:      config,
	}
}

// PreProcess 预处理 - 返回迁移提示
func (a *DiscoveryInputAdapter) PreProcess(ctx context.Context, input *InputData) (*PreprocessResult, error) {
	a.logger.Infof("自动发现功能已迁移到Agent服务")

	return &PreprocessResult{
		Success:       false,
		ProcessedData: input,
		Warnings: []string{
			"自动发现功能已迁移到Agent服务",
			"请使用Agent服务进行网络发现和资产自动采集",
			"当前适配器仅保留兼容性，不提供实际发现功能",
		},
		Metadata: map[string]interface{}{
			"status":           "deprecated",
			"migration_target": "Agent服务",
			"migration_guide":  "请参考Agent服务文档配置网络发现",
			"last_check":       time.Now(),
		},
	}, fmt.Errorf("功能已迁移: 请使用Agent服务进行网络发现")
}

// Parse 解析数据 - 返回空结果和迁移提示
func (a *DiscoveryInputAdapter) Parse(ctx context.Context, data []byte) ([]*RawAssetData, error) {
	a.logger.Infof("自动发现解析功能已迁移到Agent服务")

	return []*RawAssetData{}, fmt.Errorf("功能已迁移: 自动发现功能已迁移到Agent服务，请使用Agent进行网络资产发现")
}

// PostProcess 后处理 - 返回空结果
func (a *DiscoveryInputAdapter) PostProcess(ctx context.Context, assets []*RawAssetData) ([]*ProcessedAssetData, error) {
	a.logger.Infof("自动发现后处理功能已迁移到Agent服务")

	return []*ProcessedAssetData{}, fmt.Errorf("功能已迁移: 请使用Agent服务")
}

// Validate 验证 - 返回迁移提示
func (a *DiscoveryInputAdapter) Validate(ctx context.Context, data *ParsedData, svcCtx *svc.ServiceContext) (*ValidationResult, error) {
	return &ValidationResult{
		Valid:      false,
		Errors:     []*ValidationError{},
		Warnings:   []*ValidationWarning{},
		ValidCount: 0,
		ErrorCount: 1,
		Summary: map[string]interface{}{
			"status":          "deprecated",
			"migration_note":  "功能已迁移到Agent服务",
			"migration_guide": "请使用Agent服务进行网络发现",
		},
	}, fmt.Errorf("功能已迁移: 自动发现功能已迁移到Agent服务")
}

// HealthCheck 健康检查 - 始终返回迁移状态
func (a *DiscoveryInputAdapter) HealthCheck() error {
	a.logger.Infof("自动发现适配器健康检查: 功能已迁移到Agent服务")

	// 返回nil表示适配器本身是"健康"的，只是功能已迁移
	return nil
}

// GetStats 获取统计信息
func (a *DiscoveryInputAdapter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type":             "discovery",
		"version":          "1.0.0",
		"status":           "deprecated",
		"migration_target": "Agent服务",
		"migration_note":   "功能已迁移到Agent服务，请使用Agent进行网络发现",
		"last_check":       time.Now().Format(time.RFC3339),
		"total_requests":   0,
		"success_requests": 0,
		"failed_requests":  0,
	}
}

// GetMigrationInfo 获取迁移信息
func (a *DiscoveryInputAdapter) GetMigrationInfo() map[string]interface{} {
	return map[string]interface{}{
		"status":           "migrated_to_agent",
		"migration_target": "Agent服务",
		"migration_reason": "网络发现功能需要更强的网络访问能力和资源管理",
		"agent_benefits": []string{
			"更高效的网络扫描性能",
			"更好的资源隔离和管理",
			"支持分布式发现",
			"更强的网络协议支持",
		},
		"migration_guide": map[string]interface{}{
			"step1": "部署Agent服务",
			"step2": "配置网络发现规则",
			"step3": "启动自动发现任务",
			"step4": "通过Agent API获取发现结果",
		},
		"agent_api_endpoint": "/agent/discovery",
		"documentation":      "请参考Agent服务文档",
		"support_contact":    "技术支持团队",
	}
}
