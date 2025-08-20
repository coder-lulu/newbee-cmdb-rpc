package input

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitee.com/link234/cmdb-rpc/internal/svc"
)

// APIInputAdapter API输入适配器
type APIInputAdapter struct {
	*BaseAdapter
	config *APIConfig
}

// NewAPIInputAdapter 创建API输入适配器
func NewAPIInputAdapter(svcCtx *svc.ServiceContext, config *APIConfig) *APIInputAdapter {
	if config == nil {
		config = &APIConfig{
			MaxBatchSize:   1000,
			RateLimit:      100,
			Timeout:        30 * time.Second,
			RetryCount:     3,
			RequiredFields: []string{},
			FieldMapping:   make(map[string]string),
		}
	}

	base := NewBaseAdapter("api", "1.0.0", svcCtx)
	base.config = &ConfigSchema{
		Type:        "api",
		Version:     "1.0.0",
		Description: "API接口导入适配器，支持JSON格式数据导入",
		Properties: map[string]interface{}{
			"maxBatchSize":   config.MaxBatchSize,
			"rateLimit":      config.RateLimit,
			"timeout":        config.Timeout.String(),
			"retryCount":     config.RetryCount,
			"requiredFields": config.RequiredFields,
			"fieldMapping":   config.FieldMapping,
		},
		Required: []string{"api_data"},
	}

	return &APIInputAdapter{
		BaseAdapter: base,
		config:      config,
	}
}

// PreProcess 预处理API数据
func (a *APIInputAdapter) PreProcess(ctx context.Context, input *InputData) (*PreprocessResult, error) {
	startTime := time.Now()
	a.logger.Infof("开始API数据预处理: 数据大小=%d字节", len(input.Data))

	result := &PreprocessResult{
		Success:       true,
		ProcessedData: input,
		Warnings:      make([]string, 0),
		Metadata:      make(map[string]interface{}),
	}

	// 1. 验证数据不为空
	if len(input.Data) == 0 {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), 0, 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{"输入数据为空"},
		}, fmt.Errorf("输入数据为空")
	}

	// 2. 解析和验证JSON格式
	records, err := a.parseJSONData(input.Data)
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(input.Data)), 0)
		return &PreprocessResult{
			Success:  false,
			Warnings: []string{fmt.Sprintf("JSON解析失败: %v", err)},
		}, fmt.Errorf("JSON格式错误: %v", err)
	}

	// 3. 检查数据量
	recordCount := len(records)
	if recordCount == 0 {
		result.Warnings = append(result.Warnings, "未发现有效的数据记录")
	}

	if recordCount > a.config.MaxBatchSize {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("数据量超过建议值: %d > %d，建议分批处理", recordCount, a.config.MaxBatchSize))
	}

	// 4. 检查字段完整性
	fieldStats := a.analyzeFieldUsage(records)
	missingFieldRecords := a.checkRequiredFields(records)
	if len(missingFieldRecords) > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("有%d条记录缺少必填字段", len(missingFieldRecords)))
	}

	// 5. 设置元数据
	result.Metadata = map[string]interface{}{
		"record_count":           recordCount,
		"data_type":              "api_json",
		"field_statistics":       fieldStats,
		"missing_required_count": len(missingFieldRecords),
		"data_size_bytes":        len(input.Data),
		"processed_at":           time.Now(),
		"estimated_process_time": a.estimateProcessTime(recordCount),
	}

	// 记录统计信息
	a.BaseAdapter.stats.RecordRequest(true, time.Since(startTime), int64(len(input.Data)), int64(recordCount))

	a.logger.Infof("API预处理完成: %d条记录，%d个警告", recordCount, len(result.Warnings))

	return result, nil
}

// Parse 解析API数据
func (a *APIInputAdapter) Parse(ctx context.Context, data []byte) ([]*RawAssetData, error) {
	startTime := time.Now()
	a.logger.Infof("开始解析API数据")

	// 解析JSON数据
	records, err := a.parseJSONData(data)
	if err != nil {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	if len(records) == 0 {
		a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
		return nil, fmt.Errorf("未发现可解析的记录")
	}

	var assets []*RawAssetData
	var parseErrors []string

	for i, record := range records {
		lineNumber := i + 1

		// 验证必填字段
		if missingFields := a.validateRequiredFields(record); len(missingFields) > 0 {
			parseErrors = append(parseErrors,
				fmt.Sprintf("第%d条记录缺少必填字段: %s", lineNumber, strings.Join(missingFields, ", ")))
			continue
		}

		// 应用字段映射
		mappedRecord := a.applyFieldMapping(record)

		// 数据类型转换和验证
		validatedRecord, err := a.validateAndConvertTypes(mappedRecord, lineNumber)
		if err != nil {
			parseErrors = append(parseErrors,
				fmt.Sprintf("第%d条记录数据验证失败: %v", lineNumber, err))
			continue
		}

		// 创建资产数据
		asset := &RawAssetData{
			ID:         a.generateAssetID(lineNumber, validatedRecord),
			LineNumber: lineNumber,
			Attributes: validatedRecord,
			Source:     "api",
			Metadata: map[string]interface{}{
				"parsed_at":     time.Now(),
				"record_index":  lineNumber,
				"original_size": len(fmt.Sprintf("%v", record)),
				"field_count":   len(validatedRecord),
			},
		}

		// 确定CI类型
		ciTypeID, ciTypeName := a.determineCIType(asset.Attributes)
		asset.CITypeID = ciTypeID
		asset.CITypeName = ciTypeName

		assets = append(assets, asset)
	}

	// 处理解析错误
	if len(parseErrors) > 0 {
		a.logger.Errorf("解析过程中发现%d个错误: %s", len(parseErrors), strings.Join(parseErrors, "; "))
		// 如果所有记录都失败，返回错误
		if len(assets) == 0 {
			a.BaseAdapter.stats.RecordRequest(false, time.Since(startTime), int64(len(data)), 0)
			return nil, fmt.Errorf("所有记录解析失败: %s", strings.Join(parseErrors, "; "))
		}
		// 如果部分失败，记录警告但继续处理
		a.logger.Infof("警告: 部分记录解析失败，成功解析%d条，失败%d条", len(assets), len(parseErrors))
	}

	// 记录统计信息
	a.BaseAdapter.stats.RecordRequest(true, time.Since(startTime), int64(len(data)), int64(len(assets)))

	a.logger.Infof("API解析完成: 生成%d个资产记录", len(assets))

	return assets, nil
}

// PostProcess 后处理解析的资产数据
func (a *APIInputAdapter) PostProcess(ctx context.Context, assets []*RawAssetData) ([]*ProcessedAssetData, error) {
	a.logger.Infof("开始API数据后处理: %d个资产", len(assets))

	processedAssets := make([]*ProcessedAssetData, 0, len(assets))

	for _, asset := range assets {
		processed := &ProcessedAssetData{
			RawAssetData:   asset,
			ProcessTime:    time.Now(),
			Status:         "parsed",
			ProcessorChain: []string{"api_parser"},
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
			processed.Status = "error"
		} else {
			processed.Status = "cleaned"
		}

		processed.ProcessorChain = append(processed.ProcessorChain, "data_cleaner")
		processedAssets = append(processedAssets, processed)
	}

	a.logger.Infof("API后处理完成: %d个资产", len(processedAssets))

	return processedAssets, nil
}

// Validate 验证解析的数据
func (a *APIInputAdapter) Validate(ctx context.Context, data *ParsedData, svcCtx *svc.ServiceContext) (*ValidationResult, error) {
	return a.ValidateWithExistingLogic(ctx, data.Records)
}

// parseJSONData 解析JSON数据（统一的解析逻辑）
func (a *APIInputAdapter) parseJSONData(data []byte) ([]map[string]interface{}, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("无效的JSON格式: %v", err)
	}

	var records []map[string]interface{}

	switch data := jsonData.(type) {
	case []interface{}:
		// 直接的数组格式
		for i, item := range data {
			if record, ok := item.(map[string]interface{}); ok {
				records = append(records, record)
			} else {
				return nil, fmt.Errorf("数组索引%d的元素不是有效的对象格式", i)
			}
		}

	case map[string]interface{}:
		// 检查是否包含data字段
		if dataArray, exists := data["data"]; exists {
			if dataSlice, ok := dataArray.([]interface{}); ok {
				for i, item := range dataSlice {
					if record, ok := item.(map[string]interface{}); ok {
						records = append(records, record)
					} else {
						return nil, fmt.Errorf("data数组索引%d的元素不是有效的对象格式", i)
					}
				}
			} else {
				return nil, fmt.Errorf("data字段不是数组格式")
			}
		} else {
			// 单个对象
			records = append(records, data)
		}

	default:
		return nil, fmt.Errorf("不支持的JSON数据格式，期望对象或对象数组")
	}

	return records, nil
}

// validateRequiredFields 验证必填字段
func (a *APIInputAdapter) validateRequiredFields(record map[string]interface{}) []string {
	var missingFields []string

	for _, field := range a.config.RequiredFields {
		if _, exists := record[field]; !exists {
			missingFields = append(missingFields, field)
		} else {
			// 检查字段值是否为空
			if value := record[field]; value == nil || value == "" {
				missingFields = append(missingFields, field+"(值为空)")
			}
		}
	}

	return missingFields
}

// validateAndConvertTypes 验证和转换数据类型
func (a *APIInputAdapter) validateAndConvertTypes(record map[string]interface{}, lineNumber int) (map[string]interface{}, error) {
	validated := make(map[string]interface{})

	for key, value := range record {
		convertedValue, err := a.convertAndValidateValue(key, value)
		if err != nil {
			return nil, fmt.Errorf("字段'%s'验证失败: %v", key, err)
		}
		validated[key] = convertedValue
	}

	return validated, nil
}

// convertAndValidateValue 转换和验证单个值
func (a *APIInputAdapter) convertAndValidateValue(fieldName string, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 根据字段名进行特定验证
	switch strings.ToLower(fieldName) {
	case "ip_address", "ip", "ipaddress":
		return a.validateIP(value)
	case "port":
		return a.validatePort(value)
	case "hostname", "host_name", "server_name":
		return a.validateHostname(value)
	case "cpu_cores", "cpu", "cores":
		return a.validatePositiveInteger(value, "CPU核数")
	case "memory_gb", "memory", "ram":
		return a.validatePositiveNumber(value, "内存大小")
	case "disk_gb", "disk", "storage":
		return a.validatePositiveNumber(value, "磁盘大小")
	case "status", "state":
		return a.validateStatus(value)
	default:
		// 通用类型转换
		return a.convertCommonTypes(value), nil
	}
}

// validateIP 验证IP地址
func (a *APIInputAdapter) validateIP(value interface{}) (string, error) {
	ipStr := fmt.Sprintf("%v", value)
	ipStr = strings.TrimSpace(ipStr)

	if ipStr == "" {
		return "", fmt.Errorf("IP地址不能为空")
	}

	if net.ParseIP(ipStr) == nil {
		return "", fmt.Errorf("无效的IP地址格式: %s", ipStr)
	}

	return ipStr, nil
}

// validatePort 验证端口号
func (a *APIInputAdapter) validatePort(value interface{}) (int64, error) {
	port, err := a.convertToInt64(value)
	if err != nil {
		return 0, fmt.Errorf("无效的端口号: %v", value)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口号必须在1-65535范围内: %d", port)
	}

	return port, nil
}

// validateHostname 验证主机名
func (a *APIInputAdapter) validateHostname(value interface{}) (string, error) {
	hostname := strings.TrimSpace(fmt.Sprintf("%v", value))

	if hostname == "" {
		return "", fmt.Errorf("主机名不能为空")
	}

	// 简单的主机名格式验证
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$`)
	if len(hostname) > 1 && !hostnameRegex.MatchString(hostname) {
		return "", fmt.Errorf("无效的主机名格式: %s", hostname)
	}

	return hostname, nil
}

// validatePositiveInteger 验证正整数
func (a *APIInputAdapter) validatePositiveInteger(value interface{}, fieldDesc string) (int64, error) {
	intVal, err := a.convertToInt64(value)
	if err != nil {
		return 0, fmt.Errorf("%s必须为正整数: %v", fieldDesc, value)
	}

	if intVal <= 0 {
		return 0, fmt.Errorf("%s必须大于0: %d", fieldDesc, intVal)
	}

	return intVal, nil
}

// validatePositiveNumber 验证正数
func (a *APIInputAdapter) validatePositiveNumber(value interface{}, fieldDesc string) (float64, error) {
	floatVal, err := a.convertToFloat64(value)
	if err != nil {
		return 0, fmt.Errorf("%s必须为正数: %v", fieldDesc, value)
	}

	if floatVal <= 0 {
		return 0, fmt.Errorf("%s必须大于0: %f", fieldDesc, floatVal)
	}

	return floatVal, nil
}

// validateStatus 验证状态字段
func (a *APIInputAdapter) validateStatus(value interface{}) (string, error) {
	status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value)))

	// 状态标准化映射
	statusMap := map[string]string{
		"active":      "active",
		"running":     "active",
		"online":      "active",
		"正常":          "active",
		"运行中":         "active",
		"运行":          "active",
		"启动":          "active",
		"inactive":    "inactive",
		"stopped":     "inactive",
		"offline":     "inactive",
		"停止":          "inactive",
		"下线":          "inactive",
		"关闭":          "inactive",
		"maintenance": "maintenance",
		"maintaining": "maintenance",
		"维护中":         "maintenance",
		"维护":          "maintenance",
	}

	if normalizedStatus, exists := statusMap[status]; exists {
		return normalizedStatus, nil
	}

	return status, nil // 返回原始值，允许其他状态
}

// convertToInt64 转换为int64
func (a *APIInputAdapter) convertToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	default:
		return strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	}
}

// convertToFloat64 转换为float64
func (a *APIInputAdapter) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	}
}

// convertCommonTypes 通用类型转换
func (a *APIInputAdapter) convertCommonTypes(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		// 尝试转换为数字
		if intVal, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return intVal
		}
		if floatVal, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return floatVal
		}
		// 尝试转换为布尔值
		if boolVal, err := strconv.ParseBool(strings.ToLower(trimmed)); err == nil {
			return boolVal
		}
		// 处理中文布尔值
		switch strings.ToLower(trimmed) {
		case "是", "对", "yes", "true", "1":
			return true
		case "否", "错", "no", "false", "0":
			return false
		}
		return trimmed
	default:
		return value
	}
}

// applyFieldMapping 应用字段映射
func (a *APIInputAdapter) applyFieldMapping(record map[string]interface{}) map[string]interface{} {
	if len(a.config.FieldMapping) == 0 {
		return record
	}

	mapped := make(map[string]interface{})
	for key, value := range record {
		if mappedKey, exists := a.config.FieldMapping[key]; exists {
			mapped[mappedKey] = value
		} else {
			mapped[key] = value
		}
	}

	return mapped
}

// generateAssetID 生成资产ID
func (a *APIInputAdapter) generateAssetID(index int, attributes map[string]interface{}) string {
	// 尝试使用id字段
	if id, exists := attributes["id"]; exists {
		if idStr := fmt.Sprintf("%v", id); idStr != "" && idStr != "<nil>" {
			return fmt.Sprintf("api_%s", idStr)
		}
	}

	// 尝试使用name或hostname字段
	nameFields := []string{"name", "hostname", "host_name", "server_name"}
	for _, field := range nameFields {
		if name, exists := attributes[field]; exists {
			if nameStr := fmt.Sprintf("%v", name); nameStr != "" && nameStr != "<nil>" {
				return fmt.Sprintf("api_%s_%d", nameStr, index)
			}
		}
	}

	// 尝试使用IP地址
	if ip, exists := attributes["ip_address"]; exists {
		if ipStr := fmt.Sprintf("%v", ip); ipStr != "" && ipStr != "<nil>" {
			return fmt.Sprintf("api_%s_%d", strings.ReplaceAll(ipStr, ".", "_"), index)
		}
	}

	// 默认使用索引
	return fmt.Sprintf("api_record_%d", index)
}

// determineCIType 确定CI类型
func (a *APIInputAdapter) determineCIType(attributes map[string]interface{}) (uint64, string) {
	// 1. 检查是否有明确的CI类型字段
	ciTypeFields := []string{"ci_type", "type", "asset_type", "category"}
	for _, field := range ciTypeFields {
		if ciType, exists := attributes[field]; exists {
			if ciTypeStr := strings.ToLower(fmt.Sprintf("%v", ciType)); ciTypeStr != "" {
				switch ciTypeStr {
				case "server", "服务器", "主机", "host":
					return 1, "服务器"
				case "service", "服务", "应用":
					return 2, "服务"
				case "network", "网络设备", "交换机", "路由器", "switch", "router":
					return 3, "网络设备"
				case "database", "数据库", "db":
					return 4, "数据库"
				case "middleware", "中间件":
					return 5, "中间件"
				}
			}
		}
	}

	// 2. 基于属性内容智能推断

	// 服务器特征：有IP地址和主机名
	hasIP := false
	hasHostname := false
	hasPort := false
	hasServiceName := false

	ipFields := []string{"ip_address", "ip", "ipaddress"}
	for _, field := range ipFields {
		if _, exists := attributes[field]; exists {
			hasIP = true
			break
		}
	}

	hostnameFields := []string{"hostname", "host_name", "server_name", "name"}
	for _, field := range hostnameFields {
		if _, exists := attributes[field]; exists {
			hasHostname = true
			break
		}
	}

	if _, exists := attributes["port"]; exists {
		hasPort = true
	}

	serviceFields := []string{"service_name", "service", "application", "app_name"}
	for _, field := range serviceFields {
		if _, exists := attributes[field]; exists {
			hasServiceName = true
			break
		}
	}

	// 推断逻辑
	if hasIP && hasHostname {
		// 检查是否有CPU、内存等硬件信息
		hardwareFields := []string{"cpu_cores", "memory_gb", "disk_gb", "os", "operating_system"}
		for _, field := range hardwareFields {
			if _, exists := attributes[field]; exists {
				return 1, "服务器"
			}
		}
		return 1, "服务器" // 有IP和主机名，默认认为是服务器
	}

	if hasPort && hasServiceName {
		return 2, "服务"
	}

	if hasServiceName {
		return 2, "服务"
	}

	// 检查数据库特征
	dbFields := []string{"database_name", "db_name", "schema", "connection_string"}
	for _, field := range dbFields {
		if _, exists := attributes[field]; exists {
			return 4, "数据库"
		}
	}

	// 检查网络设备特征
	networkFields := []string{"vlan", "interface", "switch_port", "mac_address"}
	for _, field := range networkFields {
		if _, exists := attributes[field]; exists {
			return 3, "网络设备"
		}
	}

	// 默认CI类型
	return 0, "通用资产"
}

// cleanAndNormalizeData 清洗和规范化数据
func (a *APIInputAdapter) cleanAndNormalizeData(processed *ProcessedAssetData) error {
	attributes := processed.Attributes

	// 1. 清理字符串值
	for key, value := range attributes {
		if strVal, ok := value.(string); ok {
			// 去除前后空格
			cleaned := strings.TrimSpace(strVal)
			// 处理换行符和回车符
			cleaned = strings.ReplaceAll(cleaned, "\n", " ")
			cleaned = strings.ReplaceAll(cleaned, "\r", " ")
			// 合并多余空格
			cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
			attributes[key] = cleaned
		}
	}

	// 2. 特殊字段处理
	if status, exists := attributes["status"]; exists {
		if statusStr, ok := status.(string); ok {
			normalized, _ := a.validateStatus(statusStr)
			attributes["status"] = normalized
		}
	}

	// 3. 数据完整性检查
	requiredSystemFields := []string{"id", "source"}
	for _, field := range requiredSystemFields {
		if field == "id" && processed.ID == "" {
			return fmt.Errorf("资产ID不能为空")
		}
		if field == "source" && processed.Source == "" {
			processed.Source = "api"
		}
	}

	processed.Status = "cleaned"
	processed.ProcessorChain = append(processed.ProcessorChain, "data_cleaner")

	return nil
}

// analyzeFieldUsage 分析字段使用情况
func (a *APIInputAdapter) analyzeFieldUsage(records []map[string]interface{}) map[string]interface{} {
	fieldCount := make(map[string]int)
	totalRecords := len(records)

	for _, record := range records {
		for field := range record {
			fieldCount[field]++
		}
	}

	stats := map[string]interface{}{
		"total_records": totalRecords,
		"total_fields":  len(fieldCount),
		"field_usage":   make(map[string]interface{}),
	}

	fieldUsage := stats["field_usage"].(map[string]interface{})
	for field, count := range fieldCount {
		fieldUsage[field] = map[string]interface{}{
			"count":      count,
			"percentage": float64(count) / float64(totalRecords) * 100,
		}
	}

	return stats
}

// checkRequiredFields 检查必填字段缺失情况
func (a *APIInputAdapter) checkRequiredFields(records []map[string]interface{}) []int {
	var missingRecords []int

	for i, record := range records {
		if missing := a.validateRequiredFields(record); len(missing) > 0 {
			missingRecords = append(missingRecords, i+1)
		}
	}

	return missingRecords
}

// estimateProcessTime 预估处理时间
func (a *APIInputAdapter) estimateProcessTime(recordCount int) time.Duration {
	// 基于经验数据：平均每条记录处理时间约1ms
	baseTime := time.Duration(recordCount) * time.Millisecond

	// 考虑批量处理的效率提升
	if recordCount > 100 {
		baseTime = baseTime * 80 / 100 // 20%的效率提升
	}

	return baseTime
}
