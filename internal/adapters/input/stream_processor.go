package input

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// StreamProcessor 流式处理器 - 解决大文件内存问题
type StreamProcessor struct {
	chunkSize     int
	bufferSize    int
	maxGoroutines int
	memoryLimit   int64
	logger        logx.Logger
	stats         *StreamStats
}

// StreamStats 流式处理统计
type StreamStats struct {
	TotalChunks     int64         `json:"total_chunks"`
	ProcessedChunks int64         `json:"processed_chunks"`
	FailedChunks    int64         `json:"failed_chunks"`
	TotalBytes      int64         `json:"total_bytes"`
	ProcessedBytes  int64         `json:"processed_bytes"`
	StartTime       time.Time     `json:"start_time"`
	ProcessTime     time.Duration `json:"process_time"`
	AvgChunkTime    time.Duration `json:"avg_chunk_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	GCCount         uint32        `json:"gc_count"`
	mutex           sync.RWMutex
}

// StreamConfig 流式处理配置
type StreamConfig struct {
	ChunkSize     int   `json:"chunk_size"`     // 块大小 (字节)
	BufferSize    int   `json:"buffer_size"`    // 缓冲区大小
	MaxGoroutines int   `json:"max_goroutines"` // 最大协程数
	MemoryLimit   int64 `json:"memory_limit"`   // 内存限制 (字节)
}

// ChunkResult 块处理结果
type ChunkResult struct {
	ChunkID        int             `json:"chunk_id"`
	Assets         []*RawAssetData `json:"assets"`
	Errors         []error         `json:"errors"`
	ProcessTime    time.Duration   `json:"process_time"`
	BytesProcessed int64           `json:"bytes_processed"`
}

// StreamResult 流式处理结果
type StreamResult struct {
	TotalChunks   int            `json:"total_chunks"`
	SuccessChunks int            `json:"success_chunks"`
	FailedChunks  int            `json:"failed_chunks"`
	TotalAssets   int            `json:"total_assets"`
	TotalErrors   int            `json:"total_errors"`
	ProcessTime   time.Duration  `json:"process_time"`
	ChunkResults  []*ChunkResult `json:"chunk_results"`
}

// NewStreamProcessor 创建流式处理器
func NewStreamProcessor(config *StreamConfig) *StreamProcessor {
	if config == nil {
		config = &StreamConfig{
			ChunkSize:     1024 * 1024, // 1MB
			BufferSize:    64 * 1024,   // 64KB
			MaxGoroutines: runtime.NumCPU(),
			MemoryLimit:   100 * 1024 * 1024, // 100MB
		}
	}

	processor := &StreamProcessor{
		chunkSize:     config.ChunkSize,
		bufferSize:    config.BufferSize,
		maxGoroutines: config.MaxGoroutines,
		memoryLimit:   config.MemoryLimit,
		logger:        logx.WithContext(context.Background()),
		stats: &StreamStats{
			StartTime: time.Now(),
		},
	}

	processor.logger.Infof("流式处理器初始化: 块大小=%d KB, 协程数=%d, 内存限制=%d MB",
		config.ChunkSize/1024, config.MaxGoroutines, config.MemoryLimit/(1024*1024))

	return processor
}

// ProcessStream 流式处理数据
func (s *StreamProcessor) ProcessStream(
	ctx context.Context,
	reader io.Reader,
	adapter DataInputAdapter,
	resultCallback func(*ChunkResult),
) (*StreamResult, error) {

	startTime := time.Now()
	s.stats.mutex.Lock()
	s.stats.StartTime = startTime
	s.stats.mutex.Unlock()

	s.logger.Infof("开始流式处理")

	// 创建结果收集器
	resultChan := make(chan *ChunkResult, s.maxGoroutines)
	errorChan := make(chan error, s.maxGoroutines)

	// 启动并发处理
	var wg sync.WaitGroup
	if err := s.processStreamInChunks(ctx, reader, adapter, resultChan, errorChan); err != nil {
		return nil, fmt.Errorf("流式处理失败: %v", err)
	}

	// 收集结果
	var results []*ChunkResult
	var totalAssets int
	var totalErrors int
	successChunks := 0
	failedChunks := 0

	// 处理结果通道
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				results = append(results, result)
				totalAssets += len(result.Assets)
				totalErrors += len(result.Errors)

				if len(result.Errors) == 0 {
					successChunks++
				} else {
					failedChunks++
				}

				// 执行回调
				if resultCallback != nil {
					resultCallback(result)
				}

				// 更新统计
				s.stats.mutex.Lock()
				s.stats.ProcessedChunks++
				s.stats.ProcessedBytes += result.BytesProcessed
				s.stats.mutex.Unlock()
			}

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				s.logger.Errorf("块处理错误: %v", err)
				failedChunks++

				s.stats.mutex.Lock()
				s.stats.FailedChunks++
				s.stats.mutex.Unlock()
			}

		case <-ctx.Done():
			return nil, fmt.Errorf("流式处理被取消: %v", ctx.Err())
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	totalTime := time.Since(startTime)

	// 更新最终统计
	s.stats.mutex.Lock()
	s.stats.ProcessTime = totalTime
	if len(results) > 0 {
		s.stats.AvgChunkTime = time.Duration(int64(totalTime) / int64(len(results)))
	}
	s.stats.mutex.Unlock()

	result := &StreamResult{
		TotalChunks:   len(results),
		SuccessChunks: successChunks,
		FailedChunks:  failedChunks,
		TotalAssets:   totalAssets,
		TotalErrors:   totalErrors,
		ProcessTime:   totalTime,
		ChunkResults:  results,
	}

	s.logger.Infof("流式处理完成: 块数=%d, 成功=%d, 失败=%d, 资产数=%d, 耗时=%v",
		result.TotalChunks, result.SuccessChunks, result.FailedChunks, result.TotalAssets, totalTime)

	return result, nil
}

// processStreamInChunks 分块处理流数据
func (s *StreamProcessor) processStreamInChunks(
	ctx context.Context,
	reader io.Reader,
	adapter DataInputAdapter,
	resultChan chan<- *ChunkResult,
	errorChan chan<- error,
) error {

	// 创建缓冲读取器
	bufferedReader := bufio.NewReaderSize(reader, s.bufferSize)

	// 信号量控制并发数
	semaphore := make(chan struct{}, s.maxGoroutines)

	var wg sync.WaitGroup
	chunkID := 0

	for {
		// 读取数据块
		chunk := make([]byte, s.chunkSize)
		n, err := bufferedReader.Read(chunk)

		if n > 0 {
			chunkData := chunk[:n]

			// 更新统计
			s.stats.mutex.Lock()
			s.stats.TotalChunks++
			s.stats.TotalBytes += int64(n)
			s.stats.mutex.Unlock()

			// 启动协程处理块
			wg.Add(1)
			go func(id int, data []byte) {
				defer wg.Done()

				// 获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				result := s.processChunk(ctx, id, data, adapter)

				// 更新统计
				s.stats.mutex.Lock()
				s.stats.ProcessedChunks++
				s.stats.ProcessedBytes += int64(len(data))
				if len(result.Errors) > 0 {
					s.stats.FailedChunks++
				}
				s.stats.mutex.Unlock()

				resultChan <- result

			}(chunkID, chunkData)

			chunkID++
		}

		// 检查是否达到EOF
		if err == io.EOF {
			break
		}
		if err != nil {
			errorChan <- fmt.Errorf("读取块 %d 失败: %v", chunkID, err)
		}
	}

	// 等待所有协程完成
	wg.Wait()

	return nil
}

// processChunk 处理单个数据块
func (s *StreamProcessor) processChunk(
	ctx context.Context,
	chunkID int,
	data []byte,
	adapter DataInputAdapter,
) *ChunkResult {

	startTime := time.Now()

	result := &ChunkResult{
		ChunkID:        chunkID,
		Assets:         []*RawAssetData{},
		Errors:         []error{},
		BytesProcessed: int64(len(data)),
	}

	s.logger.Debugf("处理块 %d，大小: %d 字节", chunkID, len(data))

	// 检查数据是否为空或无效
	if len(data) == 0 {
		result.Errors = append(result.Errors, fmt.Errorf("块 %d 数据为空", chunkID))
		result.ProcessTime = time.Since(startTime)
		return result
	}

	// 预处理数据
	inputData := &InputData{
		ID:         fmt.Sprintf("chunk_%d", chunkID),
		Type:       adapter.GetType(),
		Data:       data,
		CreateTime: startTime,
	}

	preprocessResult, err := adapter.PreProcess(ctx, inputData)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("块 %d 预处理失败: %v", chunkID, err))
		result.ProcessTime = time.Since(startTime)
		return result
	}

	if !preprocessResult.Success {
		result.Errors = append(result.Errors, fmt.Errorf("块 %d 预处理验证失败: %v", chunkID, preprocessResult.Warnings))
		result.ProcessTime = time.Since(startTime)
		return result
	}

	// 解析数据
	assets, err := adapter.Parse(ctx, preprocessResult.ProcessedData.Data)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("块 %d 解析失败: %v", chunkID, err))
		result.ProcessTime = time.Since(startTime)
		return result
	}

	result.Assets = assets
	result.ProcessTime = time.Since(startTime)

	s.logger.Debugf("块 %d 处理完成，资产数量: %d，耗时: %v", chunkID, len(assets), result.ProcessTime)

	return result
}

// GetStats 获取统计信息
func (s *StreamProcessor) GetStats() *StreamStats {
	s.stats.mutex.RLock()
	defer s.stats.mutex.RUnlock()

	return &StreamStats{
		TotalChunks:     s.stats.TotalChunks,
		ProcessedChunks: s.stats.ProcessedChunks,
		FailedChunks:    s.stats.FailedChunks,
		TotalBytes:      s.stats.TotalBytes,
		ProcessedBytes:  s.stats.ProcessedBytes,
		StartTime:       s.stats.StartTime,
		ProcessTime:     s.stats.ProcessTime,
		AvgChunkTime:    s.stats.AvgChunkTime,
	}
}

// EstimateMemoryUsage 估算内存使用
func (s *StreamProcessor) EstimateMemoryUsage() int64 {
	// 基础内存：块大小 * 最大协程数 * 2 (读取缓冲 + 处理缓冲)
	baseMemory := int64(s.chunkSize * s.maxGoroutines * 2)

	// 缓冲区内存
	bufferMemory := int64(s.bufferSize)

	// 结果缓存内存 (估算)
	resultMemory := int64(s.maxGoroutines * 1024 * 100) // 每个结果约100KB

	return baseMemory + bufferMemory + resultMemory
}

// OptimizeForMemory 为内存受限环境优化配置
func (s *StreamProcessor) OptimizeForMemory(memoryLimitMB int64) {
	memoryLimit := memoryLimitMB * 1024 * 1024

	// 减少块大小
	if s.chunkSize > 256*1024 { // 如果大于256KB
		s.chunkSize = 256 * 1024 // 设置为256KB
	}

	// 减少协程数
	estimatedMemory := s.EstimateMemoryUsage()
	if estimatedMemory > memoryLimit {
		// 按比例减少协程数
		ratio := float64(memoryLimit) / float64(estimatedMemory)
		s.maxGoroutines = int(float64(s.maxGoroutines) * ratio)
		if s.maxGoroutines < 1 {
			s.maxGoroutines = 1
		}
	}

	s.logger.Infof("内存优化后配置: 块大小=%d, 协程数=%d, 预估内存=%d MB",
		s.chunkSize, s.maxGoroutines, s.EstimateMemoryUsage()/(1024*1024))
}

// ProcessLargeFile 处理大文件的便捷方法
func (s *StreamProcessor) ProcessLargeFile(
	ctx context.Context,
	filePath string,
	adapter DataInputAdapter,
	progressCallback func(processed, total int64),
) (*StreamResult, error) {

	// 这里应该打开文件，但为了避免文件系统依赖，我们返回一个示例实现
	// 实际使用时需要根据具体的文件系统操作来实现

	s.logger.Infof("开始处理大文件: %s", filePath)

	// 示例：创建一个模拟的文件读取器
	// file, err := os.Open(filePath)
	// if err != nil {
	//     return nil, fmt.Errorf("无法打开文件: %v", err)
	// }
	// defer file.Close()

	// 由于这是示例，我们返回一个错误提示需要实际实现
	return nil, fmt.Errorf("ProcessLargeFile需要实际的文件系统实现")
}

// SplitDataIntoLines 将数据按行分割（用于文本文件处理）
func (s *StreamProcessor) SplitDataIntoLines(data []byte) [][]byte {
	lines := [][]byte{}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			// 创建副本避免共享底层数组
			lineCopy := make([]byte, len(line))
			copy(lineCopy, line)
			lines = append(lines, lineCopy)
		}
	}

	return lines
}
