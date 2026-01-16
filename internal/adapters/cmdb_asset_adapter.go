package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/types"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
)

// CMDBAssetAdapter CMDB资产服务适配器
// 将CMDB的CI数据模型适配为通用资产模型
type CMDBAssetAdapter struct {
	svcCtx *svc.ServiceContext
	config *CMDBAdapterConfig
}

// CMDBAdapterConfig CMDB适配器配置
type CMDBAdapterConfig struct {
	// 字段映射配置：CI类型ID -> 字段映射规则
	FieldMappings map[string]*FieldMappingConfig `json:"fieldMappings"`

	// 类型映射配置：CI类型ID -> 资产类型配置
	TypeMappings map[string]*TypeMappingConfig `json:"typeMappings"`

	// 默认配置
	DefaultDisplayFields []string `json:"defaultDisplayFields"` // 默认显示字段
	DefaultSearchFields  []string `json:"defaultSearchFields"`  // 默认搜索字段
	PageSizeLimit        int      `json:"pageSizeLimit"`        // 分页大小限制
	CacheEnabled         bool     `json:"cacheEnabled"`         // 是否启用缓存
	CacheTTL             int      `json:"cacheTTL"`             // 缓存TTL（秒）
}

// FieldMappingConfig 字段映射配置
type FieldMappingConfig struct {
	// 显示名称映射
	DisplayNames map[string]string `json:"displayNames"`

	// 字段类型映射
	TypeMappings map[string]types.FieldType `json:"typeMappings"`

	// 显示配置
	DisplayInList map[string]bool `json:"displayInList"`
	DisplayInCard map[string]bool `json:"displayInCard"`

	// 搜索配置
	Searchable map[string]bool `json:"searchable"`

	// 排序配置
	Sortable map[string]bool `json:"sortable"`

	// 字段顺序
	FieldOrder map[string]int `json:"fieldOrder"`
}

// TypeMappingConfig 类型映射配置
type TypeMappingConfig struct {
	DisplayName string `json:"displayName"` // 显示名称
	Category    string `json:"category"`    // 分类
	Icon        string `json:"icon"`        // 图标
	Searchable  bool   `json:"searchable"`  // 是否可搜索
	Selectable  bool   `json:"selectable"`  // 是否可选择
}

// NewCMDBAssetAdapter 创建CMDB资产适配器
func NewCMDBAssetAdapter(svcCtx *svc.ServiceContext, config *CMDBAdapterConfig) *CMDBAssetAdapter {
	if config == nil {
		config = getDefaultCMDBAdapterConfig()
	}

	return &CMDBAssetAdapter{
		svcCtx: svcCtx,
		config: config,
	}
}

// GetAssetTypes 获取资产类型列表
func (a *CMDBAssetAdapter) GetAssetTypes(ctx context.Context, filter *types.AssetTypeFilter) ([]*types.AssetType, error) {
	// 构造CMDB查询请求
	req := &cmdb.CiTypeListReq{}

	// 应用过滤条件
	if filter != nil {
		if len(filter.IDs) > 0 {
			// CMDB API可能需要逐个查询或修改API支持批量ID查询
			// 这里暂时使用关键词搜索的方式
		}
		if filter.Keywords != "" {
			req.Name = &filter.Keywords
		}
	}

	// 调用CMDB CI类型查询逻辑
	logic := citype.NewGetCiTypeListLogic(ctx, a.svcCtx)
	resp, err := logic.GetCiTypeList(req)
	if err != nil {
		logx.Errorf("获取CI类型列表失败: %v", err)
		return nil, fmt.Errorf("获取CI类型列表失败: %w", err)
	}

	// 转换为通用资产类型
	assetTypes := make([]*types.AssetType, 0, len(resp.Data))
	for _, ciType := range resp.Data {
		assetType, err := a.mapCiTypeToAssetType(ctx, ciType)
		if err != nil {
			logx.Errorf("转换CI类型失败 [%d]: %v", ciType.Id, err)
			continue
		}

		// 应用过滤条件
		if filter != nil {
			if filter.Searchable != nil && *filter.Searchable != assetType.Searchable {
				continue
			}
			if filter.Selectable != nil && *filter.Selectable != assetType.Selectable {
				continue
			}
			if filter.Category != "" && filter.Category != assetType.Category {
				continue
			}
		}

		assetTypes = append(assetTypes, assetType)
	}

	return assetTypes, nil
}

// GetAssets 获取资产列表
func (a *CMDBAssetAdapter) GetAssets(ctx context.Context, query *types.AssetQuery) (*types.AssetListResponse, error) {
	// 构造CMDB查询请求
	req := &cmdb.CisListReq{
		Page:           uint64(query.Page),
		PageSize:       uint64(query.PageSize),
		WithAttributes: &[]bool{true}[0], // 包含属性值
	}

	// 限制分页大小
	if query.PageSize > a.config.PageSizeLimit {
		req.PageSize = uint64(a.config.PageSizeLimit)
	}

	// 应用类型过滤
	if len(query.TypeIDs) > 0 {
		if len(query.TypeIDs) == 1 {
			typeID, err := strconv.ParseUint(query.TypeIDs[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("无效的类型ID: %s", query.TypeIDs[0])
			}
			req.TypeId = &typeID
		}
		// 多个类型ID的情况需要多次查询或修改CMDB API支持
	}

	// 应用文本搜索
	if query.Search != "" {
		req.Search = &query.Search
	}

	// 应用属性过滤
	if len(query.AttributeFilters) > 0 {
		cmdbFilters := make([]*cmdb.CiAttributeFilter, 0, len(query.AttributeFilters))
		for _, filter := range query.AttributeFilters {
			cmdbFilter, err := a.mapAttributeFilterToCMDB(filter)
			if err != nil {
				logx.Errorf("转换属性过滤器失败: %v", err)
				continue
			}
			cmdbFilters = append(cmdbFilters, cmdbFilter)
		}
		req.AttributeFilters = cmdbFilters
	}

	// 应用排序
	if len(query.SortFields) > 0 {
		cmdbSorts := make([]*cmdb.CiSortField, 0, len(query.SortFields))
		for _, sort := range query.SortFields {
			cmdbSort, err := a.mapSortFieldToCMDB(sort)
			if err != nil {
				logx.Errorf("转换排序字段失败: %v", err)
				continue
			}
			cmdbSorts = append(cmdbSorts, cmdbSort)
		}
		req.SortFields = cmdbSorts
	}

	// 调用CMDB CI查询逻辑
	logic := cis.NewGetCisListLogic(ctx, a.svcCtx)
	resp, err := logic.GetCisList(req)
	if err != nil {
		logx.Errorf("获取CI列表失败: %v", err)
		return nil, fmt.Errorf("获取CI列表失败: %w", err)
	}

	// 转换为通用资产列表
	assets := make([]*types.Asset, 0, len(resp.Data))
	for _, cisInfo := range resp.Data {
		asset, err := a.mapCisInfoToAsset(ctx, cisInfo)
		if err != nil {
			logx.Errorf("转换CI信息失败 [%d]: %v", cisInfo.Id, err)
			continue
		}
		assets = append(assets, asset)
	}

	return &types.AssetListResponse{
		Total:  int64(resp.Total),
		Assets: assets,
	}, nil
}

// GetAssetDetail 获取资产详情
func (a *CMDBAssetAdapter) GetAssetDetail(ctx context.Context, id string) (*types.AssetDetail, error) {
	assetID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的资产ID: %s", id)
	}

	// 调用CMDB CI详情查询逻辑
	logic := cis.NewGetCisDetailByIdLogic(ctx, a.svcCtx)
	resp, err := logic.GetCisDetailById(&cmdb.IDReq{Id: assetID})
	if err != nil {
		logx.Errorf("获取CI详情失败 [%s]: %v", id, err)
		return nil, fmt.Errorf("获取CI详情失败: %w", err)
	}

	// 转换为通用资产
	asset, err := a.mapCisInfoToAsset(ctx, resp.CiInfo)
	if err != nil {
		return nil, fmt.Errorf("转换CI详情失败: %w", err)
	}

	// 获取资产类型信息
	assetType, err := a.getAssetTypeByID(ctx, asset.TypeID)
	if err != nil {
		logx.Errorf("获取资产类型失败 [%s]: %v", asset.TypeID, err)
	}

	return &types.AssetDetail{
		Asset: asset,
		Type:  assetType,
	}, nil
}

// GetSearchSuggestions 获取搜索建议
func (a *CMDBAssetAdapter) GetSearchSuggestions(ctx context.Context, input string, searchContext *types.SearchContext) ([]*types.Suggestion, error) {
	suggestions := make([]*types.Suggestion, 0)

	// 基于输入生成建议
	if len(input) < 2 {
		return suggestions, nil
	}

	// 搜索CI类型名称
	typeFilter := &types.AssetTypeFilter{
		Keywords: input,
	}
	assetTypes, err := a.GetAssetTypes(ctx, typeFilter)
	if err != nil {
		logx.Errorf("搜索资产类型失败: %v", err)
	} else {
		for _, assetType := range assetTypes {
			suggestions = append(suggestions, &types.Suggestion{
				Value:       fmt.Sprintf("type:%s", assetType.Name),
				Label:       assetType.DisplayName,
				Type:        "type",
				Description: fmt.Sprintf("搜索 %s 类型的资产", assetType.DisplayName),
			})
		}
	}

	// 限制建议数量
	maxResults := 10
	if searchContext != nil && searchContext.MaxResults > 0 {
		maxResults = searchContext.MaxResults
	}

	if len(suggestions) > maxResults {
		suggestions = suggestions[:maxResults]
	}

	return suggestions, nil
}

// ValidateAssetSelection 验证资产选择
func (a *CMDBAssetAdapter) ValidateAssetSelection(ctx context.Context, assets []*types.Asset) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		Valid:    true,
		Errors:   make([]*types.ValidationError, 0),
		Warnings: make([]*types.ValidationWarning, 0),
	}

	// 基本验证规则
	if len(assets) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, &types.ValidationError{
			Code:    "EMPTY_SELECTION",
			Message: "请至少选择一个资产",
		})
		return result, nil
	}

	// 验证资产状态
	for _, asset := range assets {
		if asset.Status == types.AssetStatusDeleted || asset.Status == types.AssetStatusRetired {
			result.Warnings = append(result.Warnings, &types.ValidationWarning{
				Code:    "INACTIVE_ASSET",
				Message: fmt.Sprintf("资产 %s 状态为 %s，可能无法正常使用", asset.DisplayName, asset.Status),
				AssetID: asset.ID,
			})
		}
	}

	return result, nil
}

// 私有方法：映射CI类型到资产类型
func (a *CMDBAssetAdapter) mapCiTypeToAssetType(ctx context.Context, ciType *cmdb.CiTypeInfo) (*types.AssetType, error) {
	typeID := strconv.FormatUint(*ciType.Id, 10)

	// 获取类型映射配置
	typeMapping := a.config.TypeMappings[typeID]
	if typeMapping == nil {
		typeMapping = &TypeMappingConfig{
			Searchable: true,
			Selectable: true,
		}
	}

	assetType := &types.AssetType{
		ID:          typeID,
		Name:        *ciType.Name,
		DisplayName: getStringValue(ciType.Alias, *ciType.Name),
		Searchable:  typeMapping.Searchable,
		Selectable:  typeMapping.Selectable,
		CreatedAt:   time.Unix(*ciType.CreatedAt/1000, 0),
		UpdatedAt:   time.Unix(*ciType.UpdatedAt/1000, 0),
	}

	// 应用映射配置
	if typeMapping.DisplayName != "" {
		assetType.DisplayName = typeMapping.DisplayName
	}
	if typeMapping.Category != "" {
		assetType.Category = typeMapping.Category
	}
	if typeMapping.Icon != "" {
		assetType.Icon = typeMapping.Icon
	} else if ciType.Icon != nil {
		assetType.Icon = *ciType.Icon
	}

	// 获取字段定义
	fields, err := a.getAssetFieldsForType(ctx, typeID)
	if err != nil {
		logx.Errorf("获取类型字段定义失败 [%s]: %v", typeID, err)
		fields = make([]*types.AssetField, 0)
	}
	assetType.Fields = fields

	return assetType, nil
}

// 私有方法：映射CI信息到资产
func (a *CMDBAssetAdapter) mapCisInfoToAsset(ctx context.Context, cisInfo *cmdb.CisInfo) (*types.Asset, error) {
	asset := &types.Asset{
		ID:         strconv.FormatUint(*cisInfo.Id, 10),
		TypeID:     strconv.FormatUint(*cisInfo.TypeId, 10),
		Status:     a.mapCisStatusToAssetStatus(cisInfo.Status),
		Attributes: make(map[string]*types.AssetAttributeValue),
		CreatedAt:  time.Unix(*cisInfo.CreatedAt/1000, 0),
		UpdatedAt:  time.Unix(*cisInfo.UpdatedAt/1000, 0),
	}

	// 映射状态
	if cisInfo.Available != nil && !*cisInfo.Available {
		asset.Status = types.AssetStatusInactive
	}

	// 映射属性值
	if len(cisInfo.Attributes) > 0 {
		for _, attr := range cisInfo.Attributes {
			attrValue := &types.AssetAttributeValue{
				FieldID:      strconv.FormatUint(attr.AttrId, 10),
				FieldName:    attr.AttrName,
				FieldType:    a.mapCmdbValueTypeToFieldType(attr.ValueType),
				Value:        attr.Value,
				DisplayValue: attr.Value,
			}

			// 如果有原始值，使用原始值
			if attr.RawValue != nil {
				attrValue.RawValue = *attr.RawValue
			}

			asset.Attributes[attr.AttrName] = attrValue
		}
	}

	// 生成显示名称
	asset.DisplayName = a.generateAssetDisplayName(asset)

	return asset, nil
}

// 私有方法：获取资产类型的字段定义
func (a *CMDBAssetAdapter) getAssetFieldsForType(ctx context.Context, typeID string) ([]*types.AssetField, error) {
	ciTypeID, err := strconv.ParseUint(typeID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的类型ID: %s", typeID)
	}

	// 获取属性组和属性定义
	logic := citypeattributegroup.NewListAttributeGroupWithAttributeLogic(ctx, a.svcCtx)
	resp, err := logic.ListAttributeGroupWithAttribute(&cmdb.IDReq{Id: ciTypeID})
	if err != nil {
		return nil, fmt.Errorf("获取属性组失败: %w", err)
	}

	fields := make([]*types.AssetField, 0)
	fieldMapping := a.config.FieldMappings[typeID]

	for _, group := range resp.Data {
		for _, item := range group.Items {
			if item.Attribute == nil {
				continue
			}

			attr := item.Attribute
			field := &types.AssetField{
				ID:            strconv.FormatUint(*attr.Id, 10),
				Name:          *attr.Name,
				DisplayName:   getStringValue(attr.Alias, *attr.Name),
				Type:          a.mapCmdbValueTypeToFieldType(*attr.ValueType),
				Required:      item.IsRequired != nil && *item.IsRequired,
				Searchable:    true, // 默认可搜索
				DisplayInList: item.DefaultShow != nil && *item.DefaultShow,
				Sortable:      true, // 默认可排序
				Order:         int(getUint32Value(item.Sort, 0)),
			}

			// 应用字段映射配置
			if fieldMapping != nil {
				if displayName, exists := fieldMapping.DisplayNames[field.Name]; exists {
					field.DisplayName = displayName
				}
				if fieldType, exists := fieldMapping.TypeMappings[field.Name]; exists {
					field.Type = fieldType
				}
				if searchable, exists := fieldMapping.Searchable[field.Name]; exists {
					field.Searchable = searchable
				}
				if displayInList, exists := fieldMapping.DisplayInList[field.Name]; exists {
					field.DisplayInList = displayInList
				}
				if sortable, exists := fieldMapping.Sortable[field.Name]; exists {
					field.Sortable = sortable
				}
				if order, exists := fieldMapping.FieldOrder[field.Name]; exists {
					field.Order = order
				}
			}

			// 处理选择类型字段的选项
			if attr.IsChoice != nil && *attr.IsChoice && len(attr.Choices) > 0 {
				field.Options = make([]*types.FieldOption, 0, len(attr.Choices))
				for _, choice := range attr.Choices {
					option := &types.FieldOption{
						Value: choice.Value,
						Label: choice.Value,
					}

					// 处理选项元数据
					if choice.Meta != nil {
						if choice.Meta.Label != "" {
							option.Label = choice.Meta.Label
						}

						if choice.Meta.Style != nil {
							style := make(map[string]interface{})
							if choice.Meta.Style.Color != nil {
								style["color"] = *choice.Meta.Style.Color
							}
							if choice.Meta.Style.BgColor != nil {
								style["bgColor"] = *choice.Meta.Style.BgColor
							}
							if choice.Meta.Style.FontStyle != nil {
								style["fontStyle"] = *choice.Meta.Style.FontStyle
							}
							if choice.Meta.Style.FontWeight != nil {
								style["fontWeight"] = *choice.Meta.Style.FontWeight
							}
							if choice.Meta.Style.TextDecoration != nil {
								style["textDecoration"] = *choice.Meta.Style.TextDecoration
							}
							option.Style = style
						}
					}

					field.Options = append(field.Options, option)
				}
			}

			fields = append(fields, field)
		}
	}

	return fields, nil
}

// 工具方法
func (a *CMDBAssetAdapter) mapAttributeFilterToCMDB(filter *types.AttributeFilter) (*cmdb.CiAttributeFilter, error) {
	attrID, err := strconv.ParseUint(filter.FieldID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的字段ID: %s", filter.FieldID)
	}

	cmdbFilter := &cmdb.CiAttributeFilter{
		AttrId:   attrID,
		AttrName: &filter.FieldName,
		Operator: filter.Operator,
	}

	// 处理值
	if filter.Value != nil {
		valueStr, err := json.Marshal(filter.Value)
		if err != nil {
			return nil, fmt.Errorf("序列化过滤值失败: %w", err)
		}
		value := string(valueStr)
		cmdbFilter.Value = &value
	}

	if len(filter.Values) > 0 {
		valuesStr, err := json.Marshal(filter.Values)
		if err != nil {
			return nil, fmt.Errorf("序列化过滤值列表失败: %w", err)
		}
		values := string(valuesStr)
		cmdbFilter.Values = &values
	}

	return cmdbFilter, nil
}

func (a *CMDBAssetAdapter) mapSortFieldToCMDB(sort *types.SortField) (*cmdb.CiSortField, error) {
	cmdbSort := &cmdb.CiSortField{
		Direction: sort.Direction,
	}

	if sort.FieldID != "" {
		attrID, err := strconv.ParseUint(sort.FieldID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的字段ID: %s", sort.FieldID)
		}
		cmdbSort.AttrId = attrID
	}

	if sort.FieldName != "" {
		cmdbSort.FieldName = &sort.FieldName
	}

	return cmdbSort, nil
}

func (a *CMDBAssetAdapter) mapCisStatusToAssetStatus(status *uint32) types.AssetStatus {
	if status == nil {
		return types.AssetStatusActive
	}

	switch *status {
	case 1:
		return types.AssetStatusActive
	case 2:
		return types.AssetStatusInactive
	default:
		return types.AssetStatusActive
	}
}

func (a *CMDBAssetAdapter) mapCmdbValueTypeToFieldType(valueType string) types.FieldType {
	switch strings.ToLower(valueType) {
	case "text", "varchar", "char", "string":
		return types.FieldTypeText
	case "int", "integer":
		return types.FieldTypeInteger
	case "float", "decimal", "number":
		return types.FieldTypeFloat
	case "bool", "boolean":
		return types.FieldTypeBoolean
	case "date":
		return types.FieldTypeDate
	case "datetime":
		return types.FieldTypeDateTime
	case "time":
		return types.FieldTypeTime
	case "json":
		return types.FieldTypeJSON
	default:
		return types.FieldTypeText
	}
}

func (a *CMDBAssetAdapter) generateAssetDisplayName(asset *types.Asset) string {
	// 优先使用常见的名称字段
	nameFields := []string{"name", "hostname", "ip", "title", "label"}

	for _, field := range nameFields {
		if attr, exists := asset.Attributes[field]; exists && attr.Value != nil {
			if str, ok := attr.Value.(string); ok && str != "" {
				return str
			}
		}
	}

	// 如果没有找到合适的字段，使用资产ID
	return fmt.Sprintf("Asset-%s", asset.ID)
}

func (a *CMDBAssetAdapter) getAssetTypeByID(ctx context.Context, typeID string) (*types.AssetType, error) {
	ciTypeID, err := strconv.ParseUint(typeID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的类型ID: %s", typeID)
	}

	logic := citype.NewGetCiTypeByIdLogic(ctx, a.svcCtx)
	ciType, err := logic.GetCiTypeById(&cmdb.IDReq{Id: ciTypeID})
	if err != nil {
		return nil, fmt.Errorf("获取CI类型失败: %w", err)
	}

	return a.mapCiTypeToAssetType(ctx, ciType)
}

// 默认配置
func getDefaultCMDBAdapterConfig() *CMDBAdapterConfig {
	return &CMDBAdapterConfig{
		FieldMappings:        make(map[string]*FieldMappingConfig),
		TypeMappings:         make(map[string]*TypeMappingConfig),
		DefaultDisplayFields: []string{"name", "ip", "hostname", "status"},
		DefaultSearchFields:  []string{"name", "ip", "hostname"},
		PageSizeLimit:        1000,
		CacheEnabled:         true,
		CacheTTL:             300, // 5分钟
	}
}

// 工具函数
func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func getUint32Value(ptr *uint32, defaultValue uint32) uint32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
