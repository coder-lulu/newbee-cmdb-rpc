package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/adapters/input"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// BatchProcessingExamples 批量处理使用示例
type BatchProcessingExamples struct {
	svcCtx           *svc.ServiceContext
	batchProcessor   *BatchProcessor
	batchOpProcessor *BatchOperationProcessor
	taskManager      *TaskManager
}

// NewBatchProcessingExamples 创建示例实例
func NewBatchProcessingExamples(svcCtx *svc.ServiceContext) *BatchProcessingExamples {
	return &BatchProcessingExamples{
		svcCtx:           svcCtx,
		batchProcessor:   NewBatchProcessor(svcCtx),
		batchOpProcessor: NewBatchOperationProcessor(svcCtx),
		taskManager:      NewTaskManager(svcCtx),
	}
}

// ExampleBasicBatchProcessing 基础批量处理示例
func (e *BatchProcessingExamples) ExampleBasicBatchProcessing() {
	fmt.Println("========== 基础批量处理示例 ==========")

	// 创建示例数据
	processingData := e.createSampleProcessingData(50)

	ctx := context.Background()

	// 提交批量任务
	task, err := e.batchProcessor.SubmitBatchTask(ctx, processingData, "basic_import")
	if err != nil {
		log.Printf("提交批量任务失败: %v", err)
		return
	}

	fmt.Printf("批量任务已提交: TaskID=%s\n", task.ID)
	fmt.Printf("预期处理数量: %d\n", task.Progress.TotalItems)
	fmt.Printf("预计分块数: %d\n", task.Progress.TotalChunks)

	// 等待任务完成
	e.waitForTaskCompletion(task.ID)
}

// ExampleLargeBatchProcessing 大批量处理示例
func (e *BatchProcessingExamples) ExampleLargeBatchProcessing() {
	fmt.Println("========== 大批量处理示例 ==========")

	// 创建大量示例数据
	processingData := e.createSampleProcessingData(1000)

	ctx := context.Background()

	// 提交大批量任务
	task, err := e.batchProcessor.SubmitBatchTask(ctx, processingData, "large_import")
	if err != nil {
		log.Printf("提交大批量任务失败: %v", err)
		return
	}

	fmt.Printf("大批量任务已提交: TaskID=%s\n", task.ID)
	fmt.Printf("预期处理数量: %d\n", task.Progress.TotalItems)
	fmt.Printf("预计分块数: %d\n", task.Progress.TotalChunks)

	// 监控任务进度
	e.monitorTaskProgress(task.ID)
}

// ExampleBatchOperations 批量操作示例
func (e *BatchProcessingExamples) ExampleBatchOperations() {
	fmt.Println("========== 批量操作示例 ==========")

	ctx := context.Background()

	// 示例1: 小批量删除 (同步处理)
	smallBatchReq := &cmdb.CisBatchOperationReq{
		Operation: "delete",
		CiIds:     []uint64{1, 2, 3, 4, 5}, // 5个CI ID
	}

	fmt.Println("执行小批量删除操作...")
	response, err := e.batchOpProcessor.ProcessBatchOperation(ctx, smallBatchReq)
	if err != nil {
		log.Printf("小批量操作失败: %v", err)
	} else {
		fmt.Printf("小批量操作结果: Success=%v, Message=%s\n", response.Success, response.Message)
		if response.Summary != nil {
			fmt.Printf("操作汇总: %+v\n", response.Summary)
		}
	}

	// 示例2: 大批量状态更新 (异步处理)
	largeBatchReq := &cmdb.CisBatchOperationReq{
		Operation: "update_status",
		CiIds:     make([]uint64, 500), // 500个CI ID
	}

	// 填充CI ID
	for i := range largeBatchReq.CiIds {
		largeBatchReq.CiIds[i] = uint64(i + 1)
	}

	// 设置状态参数
	statusParam := `{"status": 2}`
	largeBatchReq.Params = &statusParam

	fmt.Println("执行大批量状态更新操作...")
	response, err = e.batchOpProcessor.ProcessBatchOperation(ctx, largeBatchReq)
	if err != nil {
		log.Printf("大批量操作失败: %v", err)
	} else {
		fmt.Printf("大批量操作结果: Success=%v, Message=%s\n", response.Success, response.Message)
		if response.TaskID != "" {
			fmt.Printf("异步任务ID: %s\n", response.TaskID)
			// 监控异步任务进度
			e.monitorBatchOperationTask(response.TaskID)
		}
	}
}

// ExampleTaskManagement 任务管理示例
func (e *BatchProcessingExamples) ExampleTaskManagement() {
	fmt.Println("========== 任务管理示例 ==========")

	// 获取任务统计
	stats := e.taskManager.GetTaskStats()
	fmt.Printf("当前任务统计: %+v\n", stats)

	// 获取所有任务
	allTasks := e.taskManager.GetAllTasks()
	fmt.Printf("当前活跃任务数: %d\n", len(allTasks))

	for taskID, task := range allTasks {
		fmt.Printf("任务: ID=%s, Type=%s, Status=%s, Progress=%.2f%%\n",
			taskID, task.Type, task.Status, task.Progress.Percentage)
	}

	// 获取不同状态的任务
	pendingTasks := e.taskManager.GetTasksByStatus("pending")
	processingTasks := e.taskManager.GetTasksByStatus("processing")
	completedTasks := e.taskManager.GetTasksByStatus("completed")

	fmt.Printf("等待任务: %d个\n", len(pendingTasks))
	fmt.Printf("处理中任务: %d个\n", len(processingTasks))
	fmt.Printf("完成任务: %d个\n", len(completedTasks))
}

// ExampleBatchConfiguration 批量配置示例
func (e *BatchProcessingExamples) ExampleBatchConfiguration() {
	fmt.Println("========== 批量配置示例 ==========")

	// 获取当前配置
	config := e.batchOpProcessor.GetBatchConfig()
	fmt.Printf("当前批量操作配置: %+v\n", config)

	// 更新配置
	newConfig := &BatchOperationConfig{
		SyncThreshold:   50,              // 降低同步处理阈值
		AsyncThreshold:  2000,            // 提高异步处理阈值
		ChunkSize:       25,              // 减小分块大小
		MaxConcurrency:  8,               // 增加并发数
		TimeoutPerChunk: 3 * time.Minute, // 增加超时时间
		EnableProgress:  true,
		EnableRollback:  true,
		RetryAttempts:   5,               // 增加重试次数
		RetryDelay:      2 * time.Second, // 增加重试延迟
	}

	e.batchOpProcessor.UpdateBatchConfig(newConfig)
	fmt.Println("批量操作配置已更新")

	// 获取批量统计信息
	stats := e.batchOpProcessor.GetBatchStatistics()
	fmt.Printf("批量处理统计: %+v\n", stats)
}

// createSampleProcessingData 创建示例处理数据
func (e *BatchProcessingExamples) createSampleProcessingData(count int) *input.ProcessingData {
	assets := make([]*input.ProcessedAssetData, 0, count)

	for i := 0; i < count; i++ {
		rawAsset := &input.RawAssetData{
			ID:         fmt.Sprintf("asset_%d", i+1),
			CITypeID:   1, // 假设服务器类型ID为1
			CITypeName: "服务器",
			Attributes: map[string]interface{}{
				"hostname":   fmt.Sprintf("server%03d", i+1),
				"ip_address": fmt.Sprintf("192.168.1.%d", i+1),
				"os_type":    "Linux",
				"cpu_cores":  8,
				"memory_gb":  16,
				"disk_gb":    500,
			},
			Tags:   []string{"production", "web-server"},
			Source: "batch_example",
		}

		asset := &input.ProcessedAssetData{
			RawAssetData:   rawAsset,
			Status:         "pending",
			ProcessorChain: []string{},
		}

		assets = append(assets, asset)
	}

	return &input.ProcessingData{
		Assets: assets,
		BatchInfo: &input.BatchInfo{
			ID:         fmt.Sprintf("example_batch_%d", time.Now().Unix()),
			Type:       "example_import",
			Source:     "example",
			TotalCount: count,
			CreateTime: time.Now(),
			CreatedBy:  "example_user",
			Status:     "processing",
		},
		Context: map[string]interface{}{
			"example": true,
			"note":    "这是一个批量处理示例",
		},
		ErrorCount: 0,
	}
}

// waitForTaskCompletion 等待任务完成
func (e *BatchProcessingExamples) waitForTaskCompletion(taskID string) {
	fmt.Printf("等待任务完成: %s\n", taskID)

	for {
		task, err := e.batchProcessor.GetTaskStatus(taskID)
		if err != nil {
			log.Printf("获取任务状态失败: %v", err)
			break
		}

		fmt.Printf("任务状态: %s, 进度: %.2f%%, 成功: %d, 失败: %d\n",
			task.Status, task.Progress.Percentage,
			task.Progress.SuccessItems, task.Progress.FailedItems)

		if task.Status == "completed" || task.Status == "failed" {
			if task.Result != nil {
				fmt.Printf("任务完成，最终结果: 成功=%d, 失败=%d, 耗时=%v\n",
					task.Result.SuccessCount, task.Result.FailedCount, task.Result.ProcessTime)
			}
			break
		}

		time.Sleep(2 * time.Second)
	}
}

// monitorTaskProgress 监控任务进度
func (e *BatchProcessingExamples) monitorTaskProgress(taskID string) {
	fmt.Printf("监控任务进度: %s\n", taskID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			task, err := e.batchProcessor.GetTaskStatus(taskID)
			if err != nil {
				log.Printf("获取任务状态失败: %v", err)
				return
			}

			elapsed := time.Since(startTime)
			fmt.Printf("[%v] 任务进度: %s %.2f%% (%d/%d) 当前步骤: %s 吞吐量: %.2f/秒\n",
				elapsed.Truncate(time.Second),
				task.Status,
				task.Progress.Percentage,
				task.Progress.ProcessedItems,
				task.Progress.TotalItems,
				task.Progress.CurrentStep,
				task.Progress.ThroughputPerSec)

			if task.Status == "completed" || task.Status == "failed" {
				if task.Result != nil {
					fmt.Printf("监控完成，最终结果: 成功=%d, 失败=%d, 总耗时=%v, 平均吞吐量=%.2f/秒\n",
						task.Result.SuccessCount, task.Result.FailedCount,
						task.Result.ProcessTime, task.Result.ThroughputPerSec)
				}
				return
			}
		}
	}
}

// monitorBatchOperationTask 监控批量操作任务
func (e *BatchProcessingExamples) monitorBatchOperationTask(taskID string) {
	fmt.Printf("监控批量操作任务: %s\n", taskID)

	for {
		task, err := e.batchOpProcessor.GetBatchTaskStatus(taskID)
		if err != nil {
			log.Printf("获取批量操作任务状态失败: %v", err)
			break
		}

		fmt.Printf("批量操作任务状态: %s, 进度: %.2f%%\n", task.Status, task.Progress.Percentage)

		if task.Status == "completed" || task.Status == "failed" {
			fmt.Printf("批量操作任务完成: %s\n", task.Status)
			break
		}

		time.Sleep(3 * time.Second)
	}
}

// RunAllExamples 运行所有示例
func (e *BatchProcessingExamples) RunAllExamples() {
	fmt.Println("开始运行批量处理功能示例...")
	fmt.Println()

	// 1. 基础批量处理示例
	e.ExampleBasicBatchProcessing()
	fmt.Println()

	// 2. 任务管理示例
	e.ExampleTaskManagement()
	fmt.Println()

	// 3. 批量操作示例
	e.ExampleBatchOperations()
	fmt.Println()

	// 4. 批量配置示例
	e.ExampleBatchConfiguration()
	fmt.Println()

	// 注释掉大批量处理示例，避免占用过多资源
	// e.ExampleLargeBatchProcessing()

	fmt.Println("所有批量处理功能示例运行完成!")
}
