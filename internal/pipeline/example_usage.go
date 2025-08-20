package pipeline

import (
	"context"
	"fmt"
	"time"

	"gitee.com/link234/cmdb-rpc/internal/adapters/input"
	"gitee.com/link234/cmdb-rpc/internal/svc"
)

// ExamplePipelineUsage 管道使用示例
func ExamplePipelineUsage(svcCtx *svc.ServiceContext) error {
	ctx := context.Background()

	// 1. 创建数据处理管道
	pipeline := NewDataPipeline(svcCtx)

	// 2. 准备测试数据
	testData := createSampleProcessingData()

	// 3. 执行完整的数据处理流程
	fmt.Println("=== 开始数据处理管道示例 ===")

	result, err := pipeline.Process(ctx, testData)
	if err != nil {
		return fmt.Errorf("管道处理失败: %v", err)
	}

	// 4. 输出处理结果
	printPipelineResult(result)

	// 5. 验证单独处理示例
	fmt.Println("\n=== 仅验证模式示例 ===")
	testData2 := createSampleProcessingData()

	validationResult, err := pipeline.ProcessValidationOnly(ctx, testData2)
	if err != nil {
		return fmt.Errorf("验证处理失败: %v", err)
	}

	printPipelineResult(validationResult)

	return nil
}

// createSampleProcessingData 创建示例处理数据
func createSampleProcessingData() *input.ProcessingData {
	// 创建示例资产数据
	assets := []*input.ProcessedAssetData{
		{
			RawAssetData: &input.RawAssetData{
				ID:         "server-001",
				CITypeID:   1,
				CITypeName: "服务器",
				Attributes: map[string]interface{}{
					"hostname": "web-server-01",
					"ip":       "192.168.1.100",
					"cpu":      "8核",
					"memory":   "16GB",
					"os":       "CentOS 7",
				},
				Source:     "excel_import",
				BatchID:    "batch-001",
				LineNumber: 1,
			},
			Status:         "raw",
			ProcessorChain: []string{},
			ProcessTime:    time.Now(),
		},
		{
			RawAssetData: &input.RawAssetData{
				ID:         "server-002",
				CITypeID:   1,
				CITypeName: "服务器",
				Attributes: map[string]interface{}{
					"hostname": "db-server-01",
					"ip":       "192.168.1.101",
					"cpu":      "16核",
					"memory":   "32GB",
					"os":       "Ubuntu 20.04",
				},
				Source:     "excel_import",
				BatchID:    "batch-001",
				LineNumber: 2,
			},
			Status:         "raw",
			ProcessorChain: []string{},
			ProcessTime:    time.Now(),
		},
	}

	// 创建批次信息
	batchInfo := &input.BatchInfo{
		ID:         "batch-001",
		Type:       "excel_import",
		Source:     "manual_upload",
		TotalCount: len(assets),
		CreateTime: time.Now(),
		CreatedBy:  "admin",
		Status:     "processing",
		Config: map[string]interface{}{
			"file_name": "server_assets.xlsx",
			"sheet":     "服务器资产",
		},
	}

	return &input.ProcessingData{
		Assets:     assets,
		BatchInfo:  batchInfo,
		Context:    map[string]interface{}{"example": "true"},
		ErrorCount: 0,
	}
}

// printPipelineResult 打印管道处理结果
func printPipelineResult(result *PipelineResult) {
	fmt.Printf("处理结果摘要:\n")
	fmt.Printf("  处理数量: %d\n", result.ProcessedCount)
	fmt.Printf("  错误数量: %d\n", result.TotalErrors)
	fmt.Printf("  处理时间: %v\n", result.ProcessTime)
	fmt.Printf("  状态消息: %s\n", result.Message)

	if result.ValidationStats != nil {
		fmt.Printf("验证统计:\n")
		printStats("  验证", result.ValidationStats)
	}

	if result.TransformStats != nil {
		fmt.Printf("转换统计:\n")
		printStats("  转换", result.TransformStats)
	}

	if result.PersistStats != nil {
		fmt.Printf("持久化统计:\n")
		printStats("  持久化", result.PersistStats)
	}

	fmt.Printf("最终资产状态:\n")
	for i, asset := range result.ProcessedData.Assets {
		fmt.Printf("  资产%d: %s -> %s\n", i+1, asset.ID, asset.Status)
		fmt.Printf("    处理链: %v\n", asset.ProcessorChain)
	}
}

// printStats 打印统计信息
func printStats(prefix string, stats *Stats) {
	fmt.Printf("%s总数: %d\n", prefix, stats.TotalCount)
	fmt.Printf("%s成功: %d\n", prefix, stats.SuccessCount)
	fmt.Printf("%s失败: %d\n", prefix, stats.FailedCount)
	fmt.Printf("%s跳过: %d\n", prefix, stats.SkippedCount)
	fmt.Printf("%s耗时: %v\n", prefix, stats.ProcessTime)
}

// QuickTestPipeline 快速测试管道功能
func QuickTestPipeline(svcCtx *svc.ServiceContext) {
	fmt.Println("=== 数据处理管道快速测试 ===")

	// 创建管道
	pipeline := NewDataPipeline(svcCtx)

	// 获取管道信息
	info := pipeline.GetInfo()
	fmt.Printf("管道信息: %+v\n", info)

	fmt.Println("管道组件初始化完成！")
}
