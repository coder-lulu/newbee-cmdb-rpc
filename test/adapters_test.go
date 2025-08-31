package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
)

// TestAdapterManager 测试适配器管理器
func TestAdapterManager(t *testing.T) {
	// 创建模拟的服务上下文
	svcCtx := &svc.ServiceContext{}

	// 创建适配器管理器
	manager := input.NewAdapterManager(svcCtx)

	// 测试获取可用适配器列表
	availableAdapters := manager.GetRegistry().GetAvailableAdapters()
	t.Logf("可用适配器: %v", availableAdapters)

	if len(availableAdapters) == 0 {
		t.Error("应该有可用的适配器")
	}

	// 验证预期的适配器类型
	expectedAdapters := []string{"excel", "api", "discovery"}
	for _, expected := range expectedAdapters {
		found := false
		for _, available := range availableAdapters {
			if available == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("缺少预期的适配器: %s", expected)
		}
	}
}

// TestExcelAdapter 测试Excel适配器
func TestExcelAdapter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	manager := input.NewAdapterManager(svcCtx)

	// 创建测试数据
	inputData := &input.InputData{
		ID:         "test_excel_001",
		Type:       "excel",
		Source:     "test",
		Data:       []byte("test excel data"),
		Config:     map[string]interface{}{},
		BatchID:    "batch_001",
		CreateTime: time.Now(),
		CreatedBy:  "test_user",
	}

	// 处理输入数据
	ctx := context.Background()
	result, err := manager.ProcessInput(ctx, inputData)
	if err != nil {
		t.Errorf("Excel适配器处理失败: %v", err)
		return
	}

	// 验证结果
	if result == nil {
		t.Error("处理结果不应为空")
		return
	}

	t.Logf("Excel处理结果: 处理%d条记录，%d条错误，耗时%v",
		result.ProcessedCount, result.TotalErrors, result.ProcessTime)

	if result.ProcessedData == nil {
		t.Error("处理后的数据不应为空")
	}
}

// TestAPIAdapter 测试API适配器
func TestAPIAdapter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	manager := input.NewAdapterManager(svcCtx)

	// 创建JSON测试数据
	testData := []map[string]interface{}{
		{
			"hostname":   "api-server-001",
			"ip_address": "192.168.2.100",
			"cpu_cores":  4,
			"memory_gb":  8,
		},
		{
			"hostname":   "api-server-002",
			"ip_address": "192.168.2.101",
			"cpu_cores":  8,
			"memory_gb":  16,
		},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	inputData := &input.InputData{
		ID:         "test_api_001",
		Type:       "api",
		Source:     "test",
		Data:       jsonData,
		Config:     map[string]interface{}{},
		BatchID:    "batch_002",
		CreateTime: time.Now(),
		CreatedBy:  "test_user",
	}

	// 处理输入数据
	ctx := context.Background()
	result, err := manager.ProcessInput(ctx, inputData)
	if err != nil {
		t.Errorf("API适配器处理失败: %v", err)
		return
	}

	// 验证结果
	if result == nil {
		t.Error("处理结果不应为空")
		return
	}

	t.Logf("API处理结果: 处理%d条记录，%d条错误，耗时%v",
		result.ProcessedCount, result.TotalErrors, result.ProcessTime)

	if result.ProcessedData == nil {
		t.Error("处理后的数据不应为空")
	}

	// 验证处理的记录数
	if result.ProcessedCount != len(testData) {
		t.Errorf("处理记录数不匹配: 期望%d，实际%d", len(testData), result.ProcessedCount)
	}
}

// TestDiscoveryAdapter 测试自动发现适配器
func TestDiscoveryAdapter(t *testing.T) {
	svcCtx := &svc.ServiceContext{}

	// 创建有发现规则的输入适配器
	discoveryConfig := &input.DiscoveryConfig{
		DiscoveryRules: []*input.DiscoveryRule{
			{
				ID:       "rule_001",
				Name:     "服务器发现",
				Type:     "server",
				CITypeID: 1,
				Enabled:  true,
			},
			{
				ID:       "rule_002",
				Name:     "服务发现",
				Type:     "service",
				CITypeID: 2,
				Enabled:  true,
			},
		},
		ScanInterval: 5 * time.Minute,
		MaxTargets:   100,
		Timeout:      30 * time.Second,
	}

	// 创建适配器
	adapter := input.NewDiscoveryInputAdapter(svcCtx, discoveryConfig)

	// 测试解析功能
	ctx := context.Background()
	assets, err := adapter.Parse(ctx, []byte{})
	if err != nil {
		t.Errorf("自动发现解析失败: %v", err)
		return
	}

	t.Logf("发现资产数量: %d", len(assets))

	if len(assets) == 0 {
		t.Error("应该发现一些资产")
	}

	// 验证发现的资产
	for _, asset := range assets {
		t.Logf("发现资产: ID=%s, 类型=%s, 属性=%v",
			asset.ID, asset.CITypeName, asset.Attributes)

		if asset.Source != "discovery" {
			t.Errorf("资产来源应为discovery，实际为: %s", asset.Source)
		}
	}
}

// TestAdapterConfig 测试适配器配置
func TestAdapterConfig(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	manager := input.NewAdapterManager(svcCtx)
	registry := manager.GetRegistry()

	// 测试获取Excel适配器配置
	excelConfig, err := registry.GetAdapterConfig("excel")
	if err != nil {
		t.Errorf("获取Excel适配器配置失败: %v", err)
		return
	}

	t.Logf("Excel适配器配置: %+v", excelConfig)

	if excelConfig.Type != "excel" {
		t.Errorf("配置类型错误: 期望excel，实际%s", excelConfig.Type)
	}

	// 测试配置验证
	testConfig := map[string]interface{}{
		"file_data": "test data",
	}

	err = registry.ValidateConfig("excel", testConfig)
	if err != nil {
		t.Errorf("配置验证失败: %v", err)
	}

	// 测试缺少必填配置的情况
	invalidConfig := map[string]interface{}{}
	err = registry.ValidateConfig("excel", invalidConfig)
	if err == nil {
		t.Error("应该检测到缺少必填配置")
	}
}

// BenchmarkAdapterCreation 性能测试：适配器创建
func BenchmarkAdapterCreation(b *testing.B) {
	svcCtx := &svc.ServiceContext{}
	manager := input.NewAdapterManager(svcCtx)

	config := map[string]interface{}{
		"max_file_size": int64(50 * 1024 * 1024),
		"max_rows":      50000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetAdapter("excel", config)
		if err != nil {
			b.Fatalf("适配器创建失败: %v", err)
		}
	}
}

// BenchmarkDataProcessing 性能测试：数据处理
func BenchmarkDataProcessing(b *testing.B) {
	svcCtx := &svc.ServiceContext{}
	manager := input.NewAdapterManager(svcCtx)

	// 创建大量测试数据
	testData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]interface{}{
			"hostname":   fmt.Sprintf("bench-server-%03d", i),
			"ip_address": fmt.Sprintf("192.168.%d.%d", i/254+1, i%254+1),
			"cpu_cores":  4,
			"memory_gb":  8,
		}
	}

	jsonData, _ := json.Marshal(testData)

	inputData := &input.InputData{
		ID:         "bench_test",
		Type:       "api",
		Source:     "benchmark",
		Data:       jsonData,
		Config:     map[string]interface{}{},
		BatchID:    "bench_batch",
		CreateTime: time.Now(),
		CreatedBy:  "bench_user",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ProcessInput(ctx, inputData)
		if err != nil {
			b.Fatalf("数据处理失败: %v", err)
		}
	}
}
