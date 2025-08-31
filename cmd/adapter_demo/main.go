package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
)

func main() {
	fmt.Println("=== CMDB 资产录入与输出适配器演示 ===")
	fmt.Println()

	// 创建服务上下文（模拟）
	svcCtx := &svc.ServiceContext{}

	// 创建适配器管理器
	manager := input.NewAdapterManager(svcCtx)

	// 1. 展示可用适配器
	fmt.Println("1. 可用适配器类型:")
	availableAdapters := manager.GetRegistry().GetAvailableAdapters()
	for i, adapter := range availableAdapters {
		fmt.Printf("   %d. %s\n", i+1, adapter)
	}
	fmt.Println()

	// 2. 演示API适配器
	fmt.Println("2. API适配器演示:")
	demoAPIAdapter(manager)
	fmt.Println()

	// 3. 演示自动发现适配器
	fmt.Println("3. 自动发现适配器演示:")
	demoDiscoveryAdapter(manager)
	fmt.Println()

	// 4. 展示适配器配置
	fmt.Println("4. 适配器配置信息:")
	showAdapterConfigs(manager)
	fmt.Println()

	fmt.Println("=== 演示完成 ===")
}

// demoAPIAdapter 演示API适配器
func demoAPIAdapter(manager *input.AdapterManager) {
	// 创建示例API数据
	testData := []map[string]interface{}{
		{
			"hostname":    "web-server-001",
			"ip_address":  "192.168.100.10",
			"cpu_cores":   8,
			"memory_gb":   32,
			"disk_gb":     500,
			"os":          "Ubuntu 20.04",
			"status":      "active",
			"owner":       "IT部门",
			"environment": "production",
		},
		{
			"hostname":    "db-server-001",
			"ip_address":  "192.168.100.20",
			"cpu_cores":   16,
			"memory_gb":   64,
			"disk_gb":     1000,
			"os":          "CentOS 7",
			"status":      "active",
			"owner":       "数据库团队",
			"environment": "production",
		},
		{
			"service_name": "nginx",
			"ip_address":   "192.168.100.10",
			"port":         80,
			"protocol":     "http",
			"version":      "1.20.1",
			"status":       "running",
			"owner":        "运维团队",
		},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		log.Printf("   JSON序列化失败: %v", err)
		return
	}

	inputData := &input.InputData{
		ID:     "demo_api_001",
		Type:   "api",
		Source: "demo_system",
		Data:   jsonData,
		Config: map[string]interface{}{
			"max_batch_size": 100,
		},
		BatchID:    "demo_batch_001",
		CreateTime: time.Now(),
		CreatedBy:  "demo_user",
	}

	fmt.Printf("   - 输入数据: %d条记录\n", len(testData))

	// 处理数据
	ctx := context.Background()
	result, err := manager.ProcessInput(ctx, inputData)
	if err != nil {
		log.Printf("   处理失败: %v", err)
		return
	}

	fmt.Printf("   - 处理结果: 成功处理%d条记录，%d条错误，耗时%v\n",
		result.ProcessedCount, result.TotalErrors, result.ProcessTime)

	// 展示处理后的资产
	if result.ProcessedData != nil && len(result.ProcessedData.Assets) > 0 {
		fmt.Printf("   - 资产示例:\n")
		for i, asset := range result.ProcessedData.Assets[:min(2, len(result.ProcessedData.Assets))] {
			fmt.Printf("     %d. ID: %s, 类型: %s\n", i+1, asset.ID, asset.CITypeName)
			fmt.Printf("        属性: hostname=%v, ip_address=%v\n",
				asset.Attributes["hostname"], asset.Attributes["ip_address"])
		}
	}
}

// demoDiscoveryAdapter 演示自动发现适配器
func demoDiscoveryAdapter(manager *input.AdapterManager) {
	// 创建发现规则配置
	discoveryConfig := &input.DiscoveryConfig{
		DiscoveryRules: []*input.DiscoveryRule{
			{
				ID:       "rule_server",
				Name:     "服务器发现规则",
				Type:     "server",
				CITypeID: 1,
				Enabled:  true,
			},
			{
				ID:       "rule_service",
				Name:     "服务发现规则",
				Type:     "service",
				CITypeID: 2,
				Enabled:  true,
			},
		},
		ScanInterval: 5 * time.Minute,
		MaxTargets:   1000,
		Timeout:      30 * time.Second,
	}

	// 创建自动发现适配器
	svcCtx := &svc.ServiceContext{}
	adapter := input.NewDiscoveryInputAdapter(svcCtx, discoveryConfig)

	fmt.Printf("   - 发现规则: %d条（%d条启用）\n",
		len(discoveryConfig.DiscoveryRules), countEnabledRules(discoveryConfig.DiscoveryRules))

	// 执行发现
	ctx := context.Background()
	assets, err := adapter.Parse(ctx, []byte{})
	if err != nil {
		log.Printf("   发现失败: %v", err)
		return
	}

	fmt.Printf("   - 发现结果: 发现%d个资产\n", len(assets))

	// 展示发现的资产
	if len(assets) > 0 {
		fmt.Printf("   - 发现资产示例:\n")
		for i, asset := range assets[:min(3, len(assets))] {
			fmt.Printf("     %d. ID: %s, 类型: %s, 来源: %s\n",
				i+1, asset.ID, asset.CITypeName, asset.Source)
			if hostname, ok := asset.Attributes["hostname"]; ok {
				fmt.Printf("        主机名: %v\n", hostname)
			}
			if serviceName, ok := asset.Attributes["service_name"]; ok {
				fmt.Printf("        服务名: %v\n", serviceName)
			}
		}
	}
}

// showAdapterConfigs 展示适配器配置
func showAdapterConfigs(manager *input.AdapterManager) {
	registry := manager.GetRegistry()
	adapters := registry.GetAvailableAdapters()

	for _, adapterType := range adapters {
		config, err := registry.GetAdapterConfig(adapterType)
		if err != nil {
			log.Printf("   获取%s配置失败: %v", adapterType, err)
			continue
		}

		fmt.Printf("   - %s适配器:\n", adapterType)
		fmt.Printf("     版本: %s\n", config.Version)
		fmt.Printf("     描述: %s\n", config.Description)
		fmt.Printf("     必填配置: %v\n", config.Required)

		if len(config.Properties) > 0 {
			fmt.Printf("     配置项: ")
			keys := make([]string, 0, len(config.Properties))
			for key := range config.Properties {
				keys = append(keys, key)
			}
			fmt.Printf("%v\n", keys)
		}
		fmt.Println()
	}
}

// countEnabledRules 统计启用的规则数量
func countEnabledRules(rules []*input.DiscoveryRule) int {
	count := 0
	for _, rule := range rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
