package input

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/xuri/excelize/v2"
)

// ExcelInputAdapter Excel输入适配器
type ExcelInputAdapter struct {
	*BaseAdapter
	config *ExcelConfig
}

// NewExcelInputAdapter 创建Excel输入适配器
func NewExcelInputAdapter(svcCtx *svc.ServiceContext, config *ExcelConfig) *ExcelInputAdapter {
	if config == nil {
		config = &ExcelConfig{
			MaxFileSize:    50 * 1024 * 1024, // 50MB
			MaxRows:        50000,            // 50,000行
			SkipEmptyRows:  true,
			HeaderRow:      1,
			DataStartRow:   2,
			SheetMapping:   make(map[string]string),
			ColumnMapping:  make(map[string]string),
			RequiredFields: []string{},
		}
	}

	base := NewBaseAdapter("excel", "1.0.0", svcCtx)
	base.config = &ConfigSchema{
		Type:        "excel",
		Version:     "1.0.0",
		Description: "Excel文件导入适配器",
		Properties: map[string]interface{}{
			"maxFileSize":   config.MaxFileSize,
			"maxRows":       config.MaxRows,
			"skipEmptyRows": config.SkipEmptyRows,
			"headerRow":     config.HeaderRow,
			"dataStartRow":  config.DataStartRow,
		},
		Required: []string{"file_data"},
	}

	return &ExcelInputAdapter{
		BaseAdapter: base,
		config:      config,
	}
}

// PreProcess 预处理Excel数据
func (a *ExcelInputAdapter) PreProcess(ctx context.Context, input *InputData) (*PreprocessResult, error) {
	startTime := time.Now()

	result := &PreprocessResult{
		Success:       true,
		ProcessedData: input,
		Warnings:      make([]string, 0),
		Metadata:      make(map[string]interface{}),
	}

	// 1. 检查文件大小
	if len(input.Data) > int(a.config.MaxFileSize) {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success: false,
			Warnings: []string{
				fmt.Sprintf("文件大小超过限制: %d bytes > %d bytes", len(input.Data), a.config.MaxFileSize),
			},
		}, fmt.Errorf("文件大小超过限制")
	}

	// 2. 验证Excel文件格式
	xlsx, err := excelize.OpenReader(bytes.NewReader(input.Data))
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{fmt.Sprintf("无法解析Excel文件: %v", err)},
		}, fmt.Errorf("Excel文件格式错误: %v", err)
	}
	defer func() {
		if closeErr := xlsx.Close(); closeErr != nil {
			a.logger.Errorf("关闭Excel文件失败: %v", closeErr)
		}
	}()

	// 3. 获取工作表信息
	sheetList := xlsx.GetSheetList()
	if len(sheetList) == 0 {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{"Excel文件没有工作表"},
		}, fmt.Errorf("Excel文件没有工作表")
	}

	// 4. 检查数据行数
	firstSheet := sheetList[0]
	rows, err := xlsx.GetRows(firstSheet)
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{fmt.Sprintf("无法读取工作表数据: %v", err)},
		}, fmt.Errorf("无法读取工作表数据: %v", err)
	}

	if len(rows) > a.config.MaxRows {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("数据行数超过建议值: %d > %d，可能影响处理性能", len(rows), a.config.MaxRows))
	}

	// 5. 验证表头
	if len(rows) < a.config.HeaderRow {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{fmt.Sprintf("文件行数不足，无法找到表头行(第%d行)", a.config.HeaderRow)},
		}, fmt.Errorf("文件行数不足")
	}

	headerRow := rows[a.config.HeaderRow-1]
	if len(headerRow) == 0 {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{"表头行为空"},
		}, fmt.Errorf("表头行为空")
	}

	// 6. 设置元数据
	result.Metadata = map[string]interface{}{
		"file_size":           len(input.Data),
		"file_type":           "excel",
		"sheet_count":         len(sheetList),
		"sheet_names":         sheetList,
		"primary_sheet":       firstSheet,
		"total_rows":          len(rows),
		"estimated_data_rows": len(rows) - a.config.DataStartRow + 1,
		"header_columns":      len(headerRow),
		"processed_at":        time.Now(),
	}

	// 记录统计信息
	a.BaseAdapter.stats.RecordRequest(true, time.Since(startTime), int64(len(input.Data)), int64(len(rows)))

	a.logger.Infof("Excel预处理完成: %d个工作表, %d行数据, %d列", len(sheetList), len(rows), len(headerRow))

	return result, nil
}

// Parse 解析Excel数据
func (a *ExcelInputAdapter) Parse(ctx context.Context, data []byte) ([]*RawAssetData, error) {
	startTime := time.Now()

	xlsx, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer func() {
		if closeErr := xlsx.Close(); closeErr != nil {
			a.logger.Errorf("关闭Excel文件失败: %v", closeErr)
		}
	}()

	// 获取第一个工作表（支持后续扩展多工作表）
	sheetList := xlsx.GetSheetList()
	if len(sheetList) == 0 {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	primarySheet := sheetList[0]
	// 如果配置了工作表映射，使用映射的工作表
	if mappedSheet, exists := a.config.SheetMapping[primarySheet]; exists {
		primarySheet = mappedSheet
	}

	rows, err := xlsx.GetRows(primarySheet)
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("读取工作表数据失败: %v", err)
	}

	if len(rows) < a.config.DataStartRow {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("数据行数不足，需要至少%d行", a.config.DataStartRow)
	}

	// 验证表头行索引
	if a.config.HeaderRow < 1 || a.config.HeaderRow > len(rows) {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("无效的表头行号: %d，文件只有%d行", a.config.HeaderRow, len(rows))
	}

	// 解析表头
	headerRow := rows[a.config.HeaderRow-1] // 转换为0基索引
	headers := a.processHeaders(headerRow)

	a.logger.Infof("开始解析Excel数据: 工作表=%s, 表头列数=%d", primarySheet, len(headers))

	// 解析数据行
	var assets []*RawAssetData
	for i := a.config.DataStartRow - 1; i < len(rows); i++ { // 转换为0基索引
		row := rows[i]

		// 跳过空行
		if a.config.SkipEmptyRows && a.isEmptyRow(row) {
			continue
		}

		asset, err := a.parseRow(row, headers, i+1) // 传递实际行号
		if err != nil {
			a.logger.Errorf("解析第%d行失败: %v", i+1, err)
			// 创建错误记录，但继续处理其他行
			asset = &RawAssetData{
				ID:         fmt.Sprintf("row_%d_error", i+1),
				LineNumber: i + 1,
				Attributes: make(map[string]interface{}),
				Source:     "excel",
				Metadata: map[string]interface{}{
					"parse_error": err.Error(),
					"raw_data":    strings.Join(row, ","),
					"parsed_at":   time.Now(),
				},
			}
		}

		assets = append(assets, asset)
	}

	// 记录统计信息
	a.BaseAdapter.stats.RecordRequest(true, time.Since(startTime), int64(len(data)), int64(len(assets)))

	a.logger.Infof("Excel解析完成: 共解析%d行数据，生成%d个资产记录", len(rows)-a.config.DataStartRow+1, len(assets))

	return assets, nil
}

// PostProcess 后处理解析的资产数据
func (a *ExcelInputAdapter) PostProcess(ctx context.Context, assets []*RawAssetData) ([]*ProcessedAssetData, error) {
	processedAssets := make([]*ProcessedAssetData, 0, len(assets))

	for _, asset := range assets {
		processed := &ProcessedAssetData{
			RawAssetData:   asset,
			ProcessTime:    time.Now(),
			Status:         "parsed",
			ProcessorChain: []string{"excel_parser"},
		}

		// 数据清洗和规范化
		err := a.cleanAndNormalizeData(processed)
		if err != nil {
			a.logger.Errorf("数据清洗失败，资产ID: %s, 错误: %v", asset.ID, err)
			processed.ValidationResult = &ValidationResult{
				Valid: false,
				Errors: []*ValidationError{{
					AssetID:    asset.ID,
					LineNumber: asset.LineNumber,
					ErrorType:  "data_cleaning_error",
					ErrorMsg:   err.Error(),
				}},
				ErrorCount: 1,
			}
		}

		processedAssets = append(processedAssets, processed)
	}

	return processedAssets, nil
}

// Validate 验证解析的数据
func (a *ExcelInputAdapter) Validate(ctx context.Context, data *ParsedData, svcCtx *svc.ServiceContext) (*ValidationResult, error) {
	return a.ValidateWithExistingLogic(ctx, data.Records)
}

// processHeaders 处理表头
func (a *ExcelInputAdapter) processHeaders(headerRow []string) []string {
	headers := make([]string, len(headerRow))
	for i, header := range headerRow {
		// 清理表头：去除空格、特殊字符等
		cleanHeader := strings.TrimSpace(header)

		// 应用列映射
		if mappedHeader, exists := a.config.ColumnMapping[cleanHeader]; exists {
			cleanHeader = mappedHeader
		}

		headers[i] = cleanHeader
	}
	return headers
}

// parseRow 解析单行数据
func (a *ExcelInputAdapter) parseRow(row []string, headers []string, lineNumber int) (*RawAssetData, error) {
	attributes := make(map[string]interface{})

	// 确保行数据长度与表头匹配
	maxLen := len(headers)
	if len(row) > maxLen {
		maxLen = len(row)
	}

	for i := 0; i < maxLen; i++ {
		var headerName string
		var cellValue string

		if i < len(headers) {
			headerName = headers[i]
		} else {
			headerName = fmt.Sprintf("column_%d", i+1)
		}

		if i < len(row) {
			cellValue = strings.TrimSpace(row[i])
		}

		// 跳过空的表头和值
		if headerName == "" && cellValue == "" {
			continue
		}

		// 如果表头为空但有值，使用默认列名
		if headerName == "" && cellValue != "" {
			headerName = fmt.Sprintf("column_%d", i+1)
		}

		// 尝试类型转换
		parsedValue := a.parseValue(cellValue)
		attributes[headerName] = parsedValue
	}

	// 检查必填字段
	for _, requiredField := range a.config.RequiredFields {
		if _, exists := attributes[requiredField]; !exists {
			return nil, fmt.Errorf("缺少必填字段: %s", requiredField)
		}
		if val, ok := attributes[requiredField].(string); ok && val == "" {
			return nil, fmt.Errorf("必填字段不能为空: %s", requiredField)
		}
	}

	asset := &RawAssetData{
		ID:         a.generateAssetID(lineNumber, attributes),
		LineNumber: lineNumber,
		Attributes: attributes,
		Source:     "excel",
		Metadata: map[string]interface{}{
			"parsed_at": time.Now(),
			"row_index": lineNumber,
		},
	}

	// 尝试确定CI类型
	ciTypeID, ciTypeName := a.determineCIType(attributes)
	asset.CITypeID = ciTypeID
	asset.CITypeName = ciTypeName

	return asset, nil
}

// parseValue 解析单元格值并进行类型转换
func (a *ExcelInputAdapter) parseValue(value string) interface{} {
	if value == "" {
		return ""
	}

	// 尝试解析为数字
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// 尝试解析为布尔值
	if boolVal, err := strconv.ParseBool(strings.ToLower(value)); err == nil {
		return boolVal
	}

	// 尝试解析JSON
	if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
		var jsonVal interface{}
		if err := json.Unmarshal([]byte(value), &jsonVal); err == nil {
			return jsonVal
		}
	}

	// 处理特殊值
	switch strings.ToLower(value) {
	case "null", "nil", "none":
		return nil
	case "true", "yes", "是", "1":
		return true
	case "false", "no", "否", "0":
		return false
	}

	// 默认返回字符串
	return value
}

// isEmptyRow 检查是否为空行
func (a *ExcelInputAdapter) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// generateAssetID 生成资产ID
func (a *ExcelInputAdapter) generateAssetID(lineNumber int, attributes map[string]interface{}) string {
	// 尝试使用name字段作为ID
	if name, exists := attributes["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			return fmt.Sprintf("excel_%s_%d", nameStr, lineNumber)
		}
	}

	// 尝试使用hostname字段
	if hostname, exists := attributes["hostname"]; exists {
		if hostnameStr, ok := hostname.(string); ok && hostnameStr != "" {
			return fmt.Sprintf("excel_%s_%d", hostnameStr, lineNumber)
		}
	}

	// 尝试使用id字段
	if id, exists := attributes["id"]; exists {
		if idStr, ok := id.(string); ok && idStr != "" {
			return fmt.Sprintf("excel_%s_%d", idStr, lineNumber)
		}
	}

	// 默认使用行号
	return fmt.Sprintf("excel_row_%d", lineNumber)
}

// determineCIType 确定CI类型
func (a *ExcelInputAdapter) determineCIType(attributes map[string]interface{}) (uint64, string) {
	// 检查是否有明确的CI类型字段
	if ciType, exists := attributes["ci_type"]; exists {
		if ciTypeStr, ok := ciType.(string); ok {
			switch strings.ToLower(ciTypeStr) {
			case "server", "服务器", "主机":
				return 1, "服务器"
			case "service", "服务", "应用":
				return 2, "服务"
			case "network", "网络设备", "交换机", "路由器":
				return 3, "网络设备"
			case "database", "数据库":
				return 4, "数据库"
			}
		}
	}

	// 基于属性内容推断CI类型
	if _, hasIP := attributes["ip_address"]; hasIP {
		if _, hasHostname := attributes["hostname"]; hasHostname {
			return 1, "服务器"
		}
	}

	if _, hasPort := attributes["port"]; hasPort {
		if _, hasService := attributes["service_name"]; hasService {
			return 2, "服务"
		}
	}

	// 检查数据库相关字段
	if _, hasDB := attributes["database_name"]; hasDB {
		return 4, "数据库"
	}

	// 默认CI类型
	return 0, "未知"
}

// cleanAndNormalizeData 清洗和规范化数据
func (a *ExcelInputAdapter) cleanAndNormalizeData(processed *ProcessedAssetData) error {
	attributes := processed.Attributes

	// 1. 清理字符串值
	for key, value := range attributes {
		if strVal, ok := value.(string); ok {
			// 去除前后空格
			cleaned := strings.TrimSpace(strVal)
			// 去除特殊字符（根据需要定制）
			cleaned = strings.ReplaceAll(cleaned, "\n", " ")
			cleaned = strings.ReplaceAll(cleaned, "\r", " ")
			// 处理多余空格
			cleaned = strings.Join(strings.Fields(cleaned), " ")
			attributes[key] = cleaned
		}
	}

	// 2. 验证IP地址格式
	if ipAddr, exists := attributes["ip_address"]; exists {
		if ipStr, ok := ipAddr.(string); ok && ipStr != "" {
			if !a.isValidIP(ipStr) {
				return fmt.Errorf("无效的IP地址格式: %s", ipStr)
			}
		}
	}

	// 3. 标准化主机名
	if hostname, exists := attributes["hostname"]; exists {
		if hostnameStr, ok := hostname.(string); ok {
			attributes["hostname"] = strings.ToLower(hostnameStr)
		}
	}

	// 4. 标准化状态字段
	if status, exists := attributes["status"]; exists {
		if statusStr, ok := status.(string); ok {
			switch strings.ToLower(statusStr) {
			case "active", "running", "online", "正常", "运行中":
				attributes["status"] = "active"
			case "inactive", "stopped", "offline", "停止", "下线":
				attributes["status"] = "inactive"
			case "maintenance", "maintaining", "维护中":
				attributes["status"] = "maintenance"
			default:
				attributes["status"] = "unknown"
			}
		}
	}

	// 5. 数值字段验证
	numericFields := []string{"cpu_cores", "memory_gb", "disk_gb", "port"}
	for _, field := range numericFields {
		if val, exists := attributes[field]; exists {
			if numVal, ok := val.(int64); ok && numVal < 0 {
				return fmt.Errorf("数值字段%s不能为负数: %d", field, numVal)
			}
		}
	}

	processed.Status = "cleaned"
	processed.ProcessorChain = append(processed.ProcessorChain, "data_cleaner")

	return nil
}

// isValidIP 简单的IP地址验证
func (a *ExcelInputAdapter) isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}
