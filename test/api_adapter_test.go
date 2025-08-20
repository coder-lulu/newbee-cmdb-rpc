package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitee.com/link234/cmdb-rpc/internal/adapters/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIAdapterComprehensive 综合测试API适配器
func TestAPIAdapterComprehensive(t *testing.T) {
	// 创建API适配器配置
	config := &input.APIConfig{
		MaxBatchSize:   100,
		RateLimit:      50,
		Timeout:        10 * time.Second,
		RetryCount:     2,
		RequiredFields: []string{}, // 不设置必填字段，让所有测试数据都能通过
		FieldMapping: map[string]string{
			"host":   "hostname",
			"ip":     "ip_address",
			"cpu":    "cpu_cores",
			"memory": "memory_gb",
		},
	}

	// 创建适配器
	adapter := input.NewAPIInputAdapter(nil, config)

	t.Run("PreProcess", func(t *testing.T) {
		// 测试数据
		testData := createTestAPIData()

		inputData := &input.InputData{
			ID:      "test-api-001",
			Type:    "api",
			Source:  "test_api",
			Data:    testData,
			BatchID: "batch-001",
		}

		result, err := adapter.PreProcess(context.Background(), inputData)
		require.NoError(t, err)
		assert.True(t, result.Success)

		// 验证元数据
		assert.NotNil(t, result.Metadata)
		assert.Equal(t, 4, result.Metadata["record_count"])
		assert.Equal(t, "api_json", result.Metadata["data_type"])
		assert.NotNil(t, result.Metadata["field_statistics"])

		t.Logf("预处理元数据: %+v", result.Metadata)
	})

	t.Run("Parse", func(t *testing.T) {
		testData := createTestAPIData()

		assets, err := adapter.Parse(context.Background(), testData)
		require.NoError(t, err)
		require.Len(t, assets, 4)

		// 验证第一个资产
		asset1 := assets[0]
		assert.NotEmpty(t, asset1.ID)
		assert.Equal(t, "api", asset1.Source)
		assert.Equal(t, 1, asset1.LineNumber)
		assert.Equal(t, uint64(1), asset1.CITypeID)
		assert.Equal(t, "服务器", asset1.CITypeName)

		// 验证字段映射效果
		assert.Equal(t, "web-server-01", asset1.Attributes["hostname"])
		assert.Equal(t, "192.168.1.100", asset1.Attributes["ip_address"])
		assert.Equal(t, int64(8), asset1.Attributes["cpu_cores"])
		assert.Equal(t, float64(16), asset1.Attributes["memory_gb"])

		// 验证类型转换
		assert.IsType(t, int64(0), asset1.Attributes["cpu_cores"])
		assert.IsType(t, float64(0), asset1.Attributes["memory_gb"])
		assert.IsType(t, int64(0), asset1.Attributes["port"])

		t.Logf("解析结果 - 资产1: ID=%s, 类型=%s, 属性数量=%d",
			asset1.ID, asset1.CITypeName, len(asset1.Attributes))
	})

	t.Run("PostProcess", func(t *testing.T) {
		testData := createTestAPIData()
		rawAssets, err := adapter.Parse(context.Background(), testData)
		require.NoError(t, err)

		processedAssets, err := adapter.PostProcess(context.Background(), rawAssets)
		require.NoError(t, err)
		require.Len(t, processedAssets, 4)

		// 验证处理状态
		for i, processed := range processedAssets {
			assert.NotNil(t, processed.ProcessTime)
			assert.Contains(t, []string{"cleaned", "error"}, processed.Status)
			assert.Contains(t, processed.ProcessorChain, "api_parser")
			assert.Contains(t, processed.ProcessorChain, "data_cleaner")

			t.Logf("资产 %d: 状态=%s, 处理链=%v",
				i+1, processed.Status, processed.ProcessorChain)
		}
	})
}

// TestAPIAdapterDataValidation 测试数据验证功能
func TestAPIAdapterDataValidation(t *testing.T) {
	config := &input.APIConfig{
		RequiredFields: []string{}, // 简化测试，不设置必填字段
		FieldMapping: map[string]string{
			"host": "hostname",
			"ip":   "ip_address",
		},
	}

	adapter := input.NewAPIInputAdapter(nil, config)

	t.Run("ValidData", func(t *testing.T) {
		validData := []map[string]interface{}{
			{
				"host":   "server1",
				"ip":     "192.168.1.1",
				"cpu":    "4",
				"memory": "8.5",
				"status": "运行中",
				"port":   "22",
			},
		}

		jsonData, _ := json.Marshal(validData)
		assets, err := adapter.Parse(context.Background(), jsonData)

		require.NoError(t, err)
		require.Len(t, assets, 1)

		asset := assets[0]
		// 验证字段映射
		assert.Equal(t, "server1", asset.Attributes["hostname"])
		assert.Equal(t, "192.168.1.1", asset.Attributes["ip_address"])

		// 验证类型转换
		assert.Equal(t, int64(4), asset.Attributes["cpu"])
		assert.Equal(t, float64(8.5), asset.Attributes["memory"])
		assert.Equal(t, int64(22), asset.Attributes["port"])

		// 验证状态标准化
		assert.Equal(t, "active", asset.Attributes["status"])
	})

	t.Run("InvalidIP", func(t *testing.T) {
		invalidData := []map[string]interface{}{
			{
				"hostname":   "server1",
				"ip_address": "invalid.ip.address",
			},
		}

		jsonData, _ := json.Marshal(invalidData)
		_, err := adapter.Parse(context.Background(), jsonData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的IP地址格式")
	})

	t.Run("InvalidPort", func(t *testing.T) {
		invalidData := []map[string]interface{}{
			{
				"hostname":   "server1",
				"ip_address": "192.168.1.1",
				"port":       "99999", // 超出范围
			},
		}

		jsonData, _ := json.Marshal(invalidData)
		_, err := adapter.Parse(context.Background(), jsonData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "端口号必须在1-65535范围内")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		// 重新配置必填字段进行测试
		strictAdapter := input.NewAPIInputAdapter(nil, &input.APIConfig{
			RequiredFields: []string{"hostname", "ip_address"},
		})

		invalidData := []map[string]interface{}{
			{
				"hostname": "server1",
				// 缺少ip_address
			},
		}

		jsonData, _ := json.Marshal(invalidData)
		_, err := strictAdapter.Parse(context.Background(), jsonData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "缺少必填字段")
	})
}

// TestAPIAdapterCITypeDetection 测试CI类型检测
func TestAPIAdapterCITypeDetection(t *testing.T) {
	adapter := input.NewAPIInputAdapter(nil, nil)

	testCases := []struct {
		name           string
		attributes     map[string]interface{}
		expectedTypeID uint64
		expectedName   string
	}{
		{
			name: "服务器-明确类型",
			attributes: map[string]interface{}{
				"ci_type":    "server",
				"hostname":   "web-01",
				"ip_address": "192.168.1.1",
			},
			expectedTypeID: 1,
			expectedName:   "服务器",
		},
		{
			name: "服务器-硬件特征",
			attributes: map[string]interface{}{
				"hostname":   "db-01",
				"ip_address": "192.168.1.2",
				"cpu_cores":  8,
				"memory_gb":  32,
			},
			expectedTypeID: 1,
			expectedName:   "服务器",
		},
		{
			name: "服务-端口和服务名",
			attributes: map[string]interface{}{
				"service_name": "nginx",
				"port":         80,
			},
			expectedTypeID: 2,
			expectedName:   "服务",
		},
		{
			name: "数据库-数据库特征",
			attributes: map[string]interface{}{
				"database_name": "mysql_prod",
				"db_name":       "inventory",
			},
			expectedTypeID: 4,
			expectedName:   "数据库",
		},
		{
			name: "网络设备-VLAN特征",
			attributes: map[string]interface{}{
				"vlan":      "100",
				"interface": "eth0",
			},
			expectedTypeID: 3,
			expectedName:   "网络设备",
		},
		{
			name: "通用资产-无明确特征",
			attributes: map[string]interface{}{
				"name":        "unknown-asset",
				"description": "some asset",
			},
			expectedTypeID: 0,
			expectedName:   "通用资产",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 使用反射调用私有方法的方式不可行，这里直接通过Parse来测试
			testData := []map[string]interface{}{tc.attributes}
			jsonData, _ := json.Marshal(testData)

			assets, err := adapter.Parse(context.Background(), jsonData)
			require.NoError(t, err)
			require.Len(t, assets, 1)

			asset := assets[0]
			assert.Equal(t, tc.expectedTypeID, asset.CITypeID,
				"CI类型ID不匹配，属性: %+v", tc.attributes)
			assert.Equal(t, tc.expectedName, asset.CITypeName,
				"CI类型名称不匹配，属性: %+v", tc.attributes)
		})
	}
}

// TestAPIAdapterJSONFormats 测试不同JSON格式
func TestAPIAdapterJSONFormats(t *testing.T) {
	adapter := input.NewAPIInputAdapter(nil, nil)

	t.Run("DirectArray", func(t *testing.T) {
		data := []map[string]interface{}{
			{"name": "asset1", "value": "test1"},
			{"name": "asset2", "value": "test2"},
		}
		jsonData, _ := json.Marshal(data)

		assets, err := adapter.Parse(context.Background(), jsonData)
		require.NoError(t, err)
		assert.Len(t, assets, 2)
	})

	t.Run("WrappedInDataField", func(t *testing.T) {
		wrapper := map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{"name": "asset1", "value": "test1"},
				{"name": "asset2", "value": "test2"},
			},
		}
		jsonData, _ := json.Marshal(wrapper)

		assets, err := adapter.Parse(context.Background(), jsonData)
		require.NoError(t, err)
		assert.Len(t, assets, 2)
	})

	t.Run("SingleObject", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "single-asset",
			"value": "test",
		}
		jsonData, _ := json.Marshal(data)

		assets, err := adapter.Parse(context.Background(), jsonData)
		require.NoError(t, err)
		assert.Len(t, assets, 1)
	})
}

// TestAPIAdapterErrorHandling 测试错误处理
func TestAPIAdapterErrorHandling(t *testing.T) {
	adapter := input.NewAPIInputAdapter(nil, nil)

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)

		_, err := adapter.Parse(context.Background(), invalidJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JSON解析失败")
	})

	t.Run("EmptyData", func(t *testing.T) {
		emptyData := []byte(``)

		inputData := &input.InputData{Data: emptyData}
		result, err := adapter.PreProcess(context.Background(), inputData)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, err.Error(), "输入数据为空")
	})

	t.Run("UnsupportedFormat", func(t *testing.T) {
		unsupportedData := []byte(`"just a string"`)

		_, err := adapter.Parse(context.Background(), unsupportedData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不支持的JSON数据格式")
	})

	t.Run("InvalidArrayElement", func(t *testing.T) {
		invalidArray := []byte(`[{"valid": "object"}, "invalid element", {"another": "valid"}]`)

		_, err := adapter.Parse(context.Background(), invalidArray)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不是有效的对象格式")
	})
}

// TestAPIAdapterTypeConversions 测试类型转换
func TestAPIAdapterTypeConversions(t *testing.T) {
	adapter := input.NewAPIInputAdapter(nil, nil)

	testData := []map[string]interface{}{
		{
			"hostname":    "test-server",
			"ip_address":  "192.168.1.1",
			"cpu_cores":   "8",        // 字符串转数字
			"memory_gb":   "16.5",     // 字符串转浮点数
			"enabled":     "true",     // 字符串转布尔值
			"status":      "运行中",      // 中文状态标准化
			"port":        22,         // 保持数字
			"description": "  test  ", // 字符串清理
		},
	}

	jsonData, _ := json.Marshal(testData)
	assets, err := adapter.Parse(context.Background(), jsonData)

	require.NoError(t, err)
	require.Len(t, assets, 1)

	asset := assets[0]
	attrs := asset.Attributes

	// 验证类型转换
	assert.Equal(t, int64(8), attrs["cpu_cores"])
	assert.Equal(t, float64(16.5), attrs["memory_gb"])
	assert.Equal(t, true, attrs["enabled"])
	assert.Equal(t, "active", attrs["status"])
	assert.Equal(t, int64(22), attrs["port"])
	assert.Equal(t, "test", attrs["description"])

	t.Logf("类型转换结果: %+v", attrs)
}

// TestAPIAdapterPerformance 性能测试
func TestAPIAdapterPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	adapter := input.NewAPIInputAdapter(nil, nil)

	// 创建大量测试数据
	largeData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]interface{}{
			"id":         i,
			"hostname":   fmt.Sprintf("server-%03d", i),
			"ip_address": fmt.Sprintf("192.168.%d.%d", (i/250)+1, (i%250)+1),
			"cpu_cores":  i%16 + 1,
			"memory_gb":  float64(i%64 + 4),
			"status":     "active",
		}
	}

	jsonData, _ := json.Marshal(largeData)

	start := time.Now()
	assets, err := adapter.Parse(context.Background(), jsonData)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, assets, 1000)

	t.Logf("性能测试结果: 解析1000条记录耗时 %v", duration)
	t.Logf("平均每条记录耗时: %v", duration/1000)

	// 性能基准：平均每条记录应该在1ms以内
	avgPerRecord := duration / 1000
	assert.Less(t, avgPerRecord, 1*time.Millisecond,
		"性能不达标：平均每条记录耗时超过1ms")
}

// createTestAPIData 创建测试API数据
func createTestAPIData() []byte {
	testData := []map[string]interface{}{
		{
			"host":    "web-server-01",
			"ip":      "192.168.1.100",
			"cpu":     "8",
			"memory":  "16",
			"status":  "运行中",
			"port":    "80",
			"ci_type": "server",
		},
		{
			"hostname":   "db-server-01",
			"ip_address": "192.168.1.101",
			"cpu_cores":  4,
			"memory_gb":  32.0,
			"status":     "active",
			"port":       3306,
			"type":       "database",
		},
		{
			"service_name": "nginx",
			"hostname":     "web-01",
			"ip_address":   "192.168.1.102",
			"port":         80,
			"status":       "running",
		},
		{
			"name":       "storage-device",
			"ip_address": "192.168.1.103",
			"type":       "network",
			"vlan":       "100",
			"interface":  "eth0",
		},
	}

	jsonData, _ := json.Marshal(testData)
	return jsonData
}

// BenchmarkAPIAdapterParse API适配器解析性能基准测试
func BenchmarkAPIAdapterParse(b *testing.B) {
	adapter := input.NewAPIInputAdapter(nil, nil)
	testData := createBenchmarkAPIData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Parse(context.Background(), testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// createBenchmarkAPIData 创建基准测试数据
func createBenchmarkAPIData() []byte {
	// 创建适中大小的测试数据集
	data := make([]map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		data[i] = map[string]interface{}{
			"hostname":   fmt.Sprintf("benchmark-server-%d", i),
			"ip_address": fmt.Sprintf("10.0.0.%d", i+1),
			"cpu_cores":  i%8 + 1,
			"memory_gb":  float64(i%16 + 4),
			"status":     "active",
		}
	}

	jsonData, _ := json.Marshal(data)
	return jsonData
}
