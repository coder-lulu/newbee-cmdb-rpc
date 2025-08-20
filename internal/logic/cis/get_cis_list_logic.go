package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/cis"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/ent/valuedatetime"
	"gitee.com/link234/cmdb-rpc/ent/valuefloat"
	"gitee.com/link234/cmdb-rpc/ent/valueindextext"
	"gitee.com/link234/cmdb-rpc/ent/valueinteger"
	"gitee.com/link234/cmdb-rpc/ent/valuejson"
	"gitee.com/link234/cmdb-rpc/ent/valuetext"
	"gitee.com/link234/cmdb-rpc/internal/consts"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/uuidx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCisListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCisListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCisListLogic {
	return &GetCisListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCisListLogic) GetCisList(in *cmdb.CisListReq) (*cmdb.CisListResp, error) {
	// 检查是否使用增强搜索引擎
	// 如果请求中包含高级搜索条件，优先使用增强搜索引擎
	useEnhancedSearch := l.shouldUseEnhancedSearch(in)

	if useEnhancedSearch {
		l.Logger.Info("使用增强搜索引擎")
		return l.getListWithEnhancedSearch(in)
	}

	l.Logger.Info("使用传统搜索引擎")
	return l.getListWithTraditionalSearch(in)
}

// shouldUseEnhancedSearch 判断是否应该使用增强搜索引擎
func (l *GetCisListLogic) shouldUseEnhancedSearch(in *cmdb.CisListReq) bool {
	// 强制使用增强搜索引擎
	if in.ForceEnhancedSearch != nil && *in.ForceEnhancedSearch {
		return true
	}

	// 如果有复杂的过滤条件，使用增强搜索引擎
	if len(in.FilterGroups) > 0 {
		return true
	}

	// 如果需要聚合统计，使用增强搜索引擎
	if (in.WithAggregations != nil && *in.WithAggregations) || len(in.AggregationFields) > 0 {
		return true
	}

	// 如果有多字段排序，使用增强搜索引擎
	if len(in.SortFields) > 1 {
		return true
	}

	// 如果有关系过滤，使用增强搜索引擎
	if len(in.RelationFilters) > 0 {
		return true
	}

	// 如果有继承搜索，使用增强搜索引擎
	if in.InheritanceSearch != nil {
		return true
	}

	// 如果有标签过滤，使用增强搜索引擎
	if len(in.TagFilters) > 0 {
		return true
	}

	// 如果有元数据过滤，使用增强搜索引擎
	if len(in.MetadataFilters) > 0 {
		return true
	}

	// 如果有时间范围搜索，使用增强搜索引擎
	if in.TimeRangeSearch != nil {
		return true
	}

	// 如果有分面搜索，使用增强搜索引擎
	if in.FacetSearch != nil {
		return true
	}

	// 默认使用传统搜索引擎
	return false
}

// getListWithEnhancedSearch 使用增强搜索引擎获取CI列表
func (l *GetCisListLogic) getListWithEnhancedSearch(in *cmdb.CisListReq) (*cmdb.CisListResp, error) {
	engine := NewEnhancedSearchEngine(l.ctx, l.svcCtx)

	// 构建搜索请求
	searchReq := &SearchRequest{
		CisListReq: in,
		// 这里可以扩展更多高级搜索条件
	}

	// 执行搜索
	result, err := engine.Search(searchReq)
	if err != nil {
		l.Logger.Errorf("增强搜索失败: %v", err)
		// 如果增强搜索失败，降级到传统搜索
		l.Logger.Info("降级到传统搜索引擎")
		return l.getListWithTraditionalSearch(in)
	}

	// 记录搜索性能
	l.Logger.Infof("增强搜索完成，耗时: %dms，结果数: %d", result.SearchTime, len(result.Data))

	// 扩展响应结果
	response := result.CisListResp

	// 添加增强搜索的额外信息到聚合数据中
	if result.Facets != nil || len(result.RelatedCis) > 0 {
		enhancedInfo := map[string]interface{}{
			"searchTime":  result.SearchTime,
			"totalPages":  result.TotalPages,
			"currentPage": result.CurrentPage,
		}

		if result.Facets != nil {
			enhancedInfo["facets"] = result.Facets
		}

		if len(result.RelatedCis) > 0 {
			enhancedInfo["relatedCis"] = result.RelatedCis
		}

		enhancedInfoJson, _ := json.Marshal(enhancedInfo)
		enhancedInfoStr := string(enhancedInfoJson)

		if response.Aggregations != nil {
			// 合并现有聚合信息
			var existingAgg map[string]interface{}
			if err := json.Unmarshal([]byte(*response.Aggregations), &existingAgg); err == nil {
				existingAgg["enhanced"] = enhancedInfo
				mergedJson, _ := json.Marshal(existingAgg)
				mergedStr := string(mergedJson)
				response.Aggregations = &mergedStr
			} else {
				response.Aggregations = &enhancedInfoStr
			}
		} else {
			response.Aggregations = &enhancedInfoStr
		}
	}

	return response, nil
}

// getListWithTraditionalSearch 使用传统搜索引擎获取CI列表（原有逻辑）
func (l *GetCisListLogic) getListWithTraditionalSearch(in *cmdb.CisListReq) (*cmdb.CisListResp, error) {
	l.Logger.Infof("使用传统搜索引擎: TypeId=%v, Search='%s'",
		func() interface{} {
			if in.TypeId != nil {
				return *in.TypeId
			} else {
				return "nil"
			}
		}(),
		func() string {
			if in.Search != nil {
				return *in.Search
			} else {
				return "nil"
			}
		}())

	// 构建基础查询条件
	var predicates []predicate.Cis
	predicates = l.buildBasicPredicates(in, predicates)
	l.Logger.Infof("基础查询条件数量: %d", len(predicates))

	// 构建属性过滤条件
	if len(in.AttributeFilters) > 0 || len(in.FilterGroups) > 0 {
		l.Logger.Infof("开始处理属性过滤条件...")
		attrPredicates, err := l.buildAttributePredicates(in)
		if err != nil {
			l.Logger.Errorf("构建属性过滤条件失败: %v", err)
			return nil, err
		}
		l.Logger.Infof("属性过滤条件数量: %d", len(attrPredicates))
		predicates = append(predicates, attrPredicates...)
	}

	// 构建全文搜索条件
	if in.Search != nil && *in.Search != "" {
		l.Logger.Infof("开始处理全文搜索条件: '%s'", *in.Search)
		searchPredicates, err := l.buildSearchPredicates(*in.Search, in.SearchFields)
		if err != nil {
			l.Logger.Errorf("构建全文搜索条件失败: %v", err)
			return nil, err
		}
		l.Logger.Infof("全文搜索条件数量: %d", len(searchPredicates))
		predicates = append(predicates, searchPredicates...)
	}

	l.Logger.Infof("总查询条件数量: %d", len(predicates))

	// 构建查询
	query := l.svcCtx.DB.Cis.Query().Where(predicates...)

	// 应用排序
	if len(in.SortFields) > 0 {
		l.Logger.Infof("应用排序条件数量: %d", len(in.SortFields))
		query = l.applySorting(query, in.SortFields)
	}

	// 执行分页查询
	l.Logger.Infof("执行分页查询: page=%d, pageSize=%d", in.Page, in.PageSize)
	result, err := query.Page(l.ctx, in.Page, in.PageSize)
	if err != nil {
		l.Logger.Errorf("分页查询执行失败: %v", err)
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	l.Logger.Infof("分页查询结果: 总数=%d, 当前页数据量=%d", result.PageDetails.Total, len(result.List))

	// 构建响应
	resp := &cmdb.CisListResp{
		Total: result.PageDetails.Total,
	}

	// 转换数据
	for _, v := range result.List {
		var attributes []*cmdb.CiAttributeValue

		// 根据请求决定是否加载属性
		if in.WithAttributes != nil && *in.WithAttributes {
			attributes, err = LoadCiAttributes(l.ctx, l.svcCtx.DB, v.ID, in.IncludeAttributes, in.ExcludeAttributes)
			if err != nil {
				l.Logger.Errorf("加载CI属性失败: CiID=%d, error=%v", v.ID, err)
				return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
			}
			l.Logger.Debugf("CI %d 加载了 %d 个属性", v.ID, len(attributes))
		}

		cisInfo := CisEntToProto(v, attributes)
		if cisInfo != nil {
			resp.Data = append(resp.Data, cisInfo)
		}
	}

	l.Logger.Infof("最终响应数据量: %d", len(resp.Data))

	// 处理聚合统计
	if in.WithAggregations != nil && *in.WithAggregations && len(in.AggregationFields) > 0 {
		l.Logger.Infof("开始处理聚合统计...")
		aggregations, err := l.buildAggregations(predicates, in.AggregationFields)
		if err != nil {
			l.Logger.Errorf("构建聚合统计失败: %v", err)
			return nil, err
		}
		if aggregations != "" {
			resp.Aggregations = &aggregations
			l.Logger.Infof("聚合统计完成")
		}
	}

	return resp, nil
}

// buildBasicPredicates 构建基础字段的查询条件
func (l *GetCisListLogic) buildBasicPredicates(in *cmdb.CisListReq, predicates []predicate.Cis) []predicate.Cis {
	if in.CreatedAt != nil {
		predicates = append(predicates, cis.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, cis.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, cis.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.TypeId != nil {
		predicates = append(predicates, cis.TypeIDEQ(*in.TypeId))
	}
	if in.Status != nil {
		predicates = append(predicates, cis.StatusEQ(*in.Status))
	}
	// 注意：schema中没有heartbeat字段，注释掉
	// if in.Heartbeat != nil {
	//     predicates = append(predicates, cis.HeartbeatGTE(time.UnixMilli(*in.Heartbeat)))
	// }
	// 注意：schema中没有available字段，注释掉
	// if in.Available != nil {
	//     predicates = append(predicates, cis.AvailableEQ(*in.Available))
	// }
	// 注意：schema中的created_by字段类型是UUID，不是updated_by
	if in.UpdatedBy != nil {
		predicates = append(predicates, cis.CreatedByEQ(uuidx.ParseUUIDString(*in.UpdatedBy)))
	}
	return predicates
}

// buildAttributePredicates 构建属性过滤条件
func (l *GetCisListLogic) buildAttributePredicates(in *cmdb.CisListReq) ([]predicate.Cis, error) {
	l.Logger.Infof("开始构建属性过滤条件, AttributeFilters数量: %d, FilterGroups数量: %d",
		len(in.AttributeFilters), len(in.FilterGroups))

	var ciIDs []uint64
	hasAttributeFilters := len(in.AttributeFilters) > 0 || len(in.FilterGroups) > 0

	// 处理简单属性过滤
	if len(in.AttributeFilters) > 0 {
		l.Logger.Infof("处理简单属性过滤...")
		ids, err := l.processAttributeFilters(in.AttributeFilters)
		if err != nil {
			l.Logger.Errorf("处理简单属性过滤失败: %v", err)
			return nil, err
		}
		l.Logger.Infof("简单属性过滤返回CI数量: %d", len(ids))
		ciIDs = l.intersectCiIDs(ciIDs, ids)
	}

	// 处理复合过滤组
	if len(in.FilterGroups) > 0 {
		l.Logger.Infof("处理复合过滤组...")
		ids, err := l.processFilterGroups(in.FilterGroups)
		if err != nil {
			l.Logger.Errorf("处理复合过滤组失败: %v", err)
			return nil, err
		}
		l.Logger.Infof("复合过滤组返回CI数量: %d", len(ids))
		ciIDs = l.intersectCiIDs(ciIDs, ids)
	}

	l.Logger.Infof("最终过滤后的CI数量: %d", len(ciIDs))

	// 如果有属性过滤条件但没有找到匹配的CI，返回一个永远不匹配的条件
	if hasAttributeFilters {
		if len(ciIDs) > 0 {
			l.Logger.Infof("返回匹配的CI ID列表, 数量: %d", len(ciIDs))
			return []predicate.Cis{cis.IDIn(ciIDs...)}, nil
		} else {
			l.Logger.Errorf("属性过滤条件没有找到匹配的CI，返回空结果条件")
			// 返回一个永远不匹配的条件，确保结果为空
			return []predicate.Cis{cis.IDEQ(0)}, nil
		}
	}

	l.Logger.Infof("没有属性过滤条件，返回空谓词")
	return nil, nil
}

// processAttributeFilters 处理属性过滤条件
func (l *GetCisListLogic) processAttributeFilters(filters []*cmdb.CiAttributeFilter) ([]uint64, error) {
	var allCiIDs []uint64

	for _, filter := range filters {
		ciIDs, err := l.processAttributeFilter(filter)
		if err != nil {
			return nil, err
		}
		allCiIDs = l.intersectCiIDs(allCiIDs, ciIDs)
	}

	return allCiIDs, nil
}

// processAttributeFilter 处理单个属性过滤条件
func (l *GetCisListLogic) processAttributeFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	switch filter.ValueType {
	case consts.ValueTypeShortText, consts.ValueTypeLongText, consts.ValueTypeLink:
		return l.processTextFilter(filter)
	case consts.ValueTypeInt:
		return l.processIntegerFilter(filter)
	case consts.ValueTypeFloat:
		return l.processFloatFilter(filter)
	case consts.ValueTypeDateTime, consts.ValueTypeDate, consts.ValueTypeTime:
		return l.processDatetimeFilter(filter)
	default:
		return l.processJSONFilter(filter)
	}
}

// processTextFilter 处理文本类型过滤
func (l *GetCisListLogic) processTextFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	l.Logger.Infof("处理文本过滤: AttrId=%d, Operator=%s, Value=%s", filter.AttrId, filter.Operator, func() string {
		if filter.Value != nil {
			return *filter.Value
		}
		return "nil"
	}())

	var allCiIDs []uint64

	// 查询ValueText表
	textQuery := l.svcCtx.DB.ValueText.Query().Where(valuetext.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		value := *filter.Value
		switch filter.Operator {
		case "eq":
			textQuery = textQuery.Where(valuetext.ValueEQ(value))
		case "ne":
			textQuery = textQuery.Where(valuetext.ValueNEQ(value))
		case "like":
			textQuery = textQuery.Where(valuetext.ValueContains(value))
		case "not_like":
			textQuery = textQuery.Where(valuetext.Not(valuetext.ValueContains(value)))
		case "empty":
			textQuery = textQuery.Where(valuetext.ValueEQ(""))
		case "not_empty":
			textQuery = textQuery.Where(valuetext.ValueNEQ(""))
		case "in":
			if filter.Values != nil {
				var values []string
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON: %w", err)
				}
				textQuery = textQuery.Where(valuetext.ValueIn(values...))
			}
		case "not_in":
			if filter.Values != nil {
				var values []string
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON: %w", err)
				}
				textQuery = textQuery.Where(valuetext.ValueNotIn(values...))
			}
		}
	}

	textResults, err := textQuery.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询ValueText表失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("ValueText表查询结果数量: %d", len(textResults))
	for _, result := range textResults {
		allCiIDs = append(allCiIDs, result.CiID)
		l.Logger.Debugf("ValueText结果: CiID=%d, Value=%s", result.CiID, result.Value)
	}

	// 同时查询ValueIndexText表（如果存在）
	indexTextQuery := l.svcCtx.DB.ValueIndexText.Query().Where(valueindextext.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		value := *filter.Value
		switch filter.Operator {
		case "eq":
			indexTextQuery = indexTextQuery.Where(valueindextext.ValueEQ(value))
		case "ne":
			indexTextQuery = indexTextQuery.Where(valueindextext.ValueNEQ(value))
		case "like":
			indexTextQuery = indexTextQuery.Where(valueindextext.ValueContains(value))
		case "not_like":
			indexTextQuery = indexTextQuery.Where(valueindextext.Not(valueindextext.ValueContains(value)))
		case "empty":
			indexTextQuery = indexTextQuery.Where(valueindextext.ValueEQ(""))
		case "not_empty":
			indexTextQuery = indexTextQuery.Where(valueindextext.ValueNEQ(""))
		case "in":
			if filter.Values != nil {
				var values []string
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON: %w", err)
				}
				indexTextQuery = indexTextQuery.Where(valueindextext.ValueIn(values...))
			}
		case "not_in":
			if filter.Values != nil {
				var values []string
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON: %w", err)
				}
				indexTextQuery = indexTextQuery.Where(valueindextext.ValueNotIn(values...))
			}
		}
	}

	indexTextResults, err := indexTextQuery.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询ValueIndexText表失败: %v", err)
		// 不返回错误，继续使用ValueText的结果
	} else {
		l.Logger.Infof("ValueIndexText表查询结果数量: %d", len(indexTextResults))
		for _, result := range indexTextResults {
			allCiIDs = append(allCiIDs, result.CiID)
			l.Logger.Debugf("ValueIndexText结果: CiID=%d, Value=%s", result.CiID, result.Value)
		}
	}

	// 去重CI ID
	ciIDMap := make(map[uint64]bool)
	var uniqueCiIDs []uint64
	for _, ciID := range allCiIDs {
		if !ciIDMap[ciID] {
			ciIDMap[ciID] = true
			uniqueCiIDs = append(uniqueCiIDs, ciID)
		}
	}

	l.Logger.Infof("文本过滤最终结果CI数量: %d", len(uniqueCiIDs))
	return uniqueCiIDs, nil
}

// processIntegerFilter 处理整数类型过滤
func (l *GetCisListLogic) processIntegerFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	query := l.svcCtx.DB.ValueInteger.Query().Where(valueinteger.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		value, err := strconv.Atoi(*filter.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value: %s", *filter.Value)
		}

		switch filter.Operator {
		case "eq":
			query = query.Where(valueinteger.ValueEQ(value))
		case "ne":
			query = query.Where(valueinteger.ValueNEQ(value))
		case "gt":
			query = query.Where(valueinteger.ValueGT(value))
		case "lt":
			query = query.Where(valueinteger.ValueLT(value))
		case "gte":
			query = query.Where(valueinteger.ValueGTE(value))
		case "lte":
			query = query.Where(valueinteger.ValueLTE(value))
		case "in":
			if filter.Values != nil {
				var values []int
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON for integer: %w", err)
				}
				query = query.Where(valueinteger.ValueIn(values...))
			}
		case "not_in":
			if filter.Values != nil {
				var values []int
				if err := json.Unmarshal([]byte(*filter.Values), &values); err != nil {
					return nil, fmt.Errorf("failed to parse Values JSON for integer: %w", err)
				}
				query = query.Where(valueinteger.ValueNotIn(values...))
			}
		}
	}

	results, err := query.All(l.ctx)
	if err != nil {
		return nil, err
	}

	var ciIDs []uint64
	for _, result := range results {
		ciIDs = append(ciIDs, result.CiID)
	}

	return ciIDs, nil
}

// processFloatFilter 处理浮点数类型过滤
func (l *GetCisListLogic) processFloatFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	query := l.svcCtx.DB.ValueFloat.Query().Where(valuefloat.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		value, err := strconv.ParseFloat(*filter.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", *filter.Value)
		}

		switch filter.Operator {
		case "eq":
			query = query.Where(valuefloat.ValueEQ(value))
		case "ne":
			query = query.Where(valuefloat.ValueNEQ(value))
		case "gt":
			query = query.Where(valuefloat.ValueGT(value))
		case "lt":
			query = query.Where(valuefloat.ValueLT(value))
		case "gte":
			query = query.Where(valuefloat.ValueGTE(value))
		case "lte":
			query = query.Where(valuefloat.ValueLTE(value))
		}
	}

	results, err := query.All(l.ctx)
	if err != nil {
		return nil, err
	}

	var ciIDs []uint64
	for _, result := range results {
		ciIDs = append(ciIDs, result.CiID)
	}

	return ciIDs, nil
}

// processDatetimeFilter 处理日期时间类型过滤
func (l *GetCisListLogic) processDatetimeFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	query := l.svcCtx.DB.ValueDatetime.Query().Where(valuedatetime.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		var timeValue time.Time
		var err error

		// 尝试解析时间
		if timestamp, parseErr := strconv.ParseInt(*filter.Value, 10, 64); parseErr == nil {
			timeValue = time.UnixMilli(timestamp)
		} else if timeValue, err = time.Parse(time.RFC3339, *filter.Value); err != nil {
			if timeValue, err = time.Parse("2006-01-02 15:04:05", *filter.Value); err != nil {
				return nil, fmt.Errorf("invalid datetime value: %s", *filter.Value)
			}
		}

		switch filter.Operator {
		case "eq":
			query = query.Where(valuedatetime.ValueEQ(timeValue))
		case "ne":
			query = query.Where(valuedatetime.ValueNEQ(timeValue))
		case "gt":
			query = query.Where(valuedatetime.ValueGT(timeValue))
		case "lt":
			query = query.Where(valuedatetime.ValueLT(timeValue))
		case "gte":
			query = query.Where(valuedatetime.ValueGTE(timeValue))
		case "lte":
			query = query.Where(valuedatetime.ValueLTE(timeValue))
		case "today":
			today := time.Now().Truncate(24 * time.Hour)
			tomorrow := today.Add(24 * time.Hour)
			query = query.Where(valuedatetime.ValueGTE(today), valuedatetime.ValueLT(tomorrow))
		case "yesterday":
			yesterday := time.Now().Truncate(24 * time.Hour).Add(-24 * time.Hour)
			today := yesterday.Add(24 * time.Hour)
			query = query.Where(valuedatetime.ValueGTE(yesterday), valuedatetime.ValueLT(today))
		}
	}

	results, err := query.All(l.ctx)
	if err != nil {
		return nil, err
	}

	var ciIDs []uint64
	for _, result := range results {
		ciIDs = append(ciIDs, result.CiID)
	}

	return ciIDs, nil
}

// processJSONFilter 处理JSON类型过滤
func (l *GetCisListLogic) processJSONFilter(filter *cmdb.CiAttributeFilter) ([]uint64, error) {
	query := l.svcCtx.DB.ValueJSON.Query().Where(valuejson.AttrIDEQ(filter.AttrId))

	if filter.Value != nil {
		value := *filter.Value
		switch filter.Operator {
		case "contains":
			// JSON包含查询，使用原生SQL
			query = query.Where(func(s *sql.Selector) {
				s.Where(sql.Contains(s.C("value"), value))
			})
		default:
			// 其他操作暂时不支持，返回所有该属性的记录
			// TODO: 实现更复杂的JSON查询
		}
	}

	results, err := query.All(l.ctx)
	if err != nil {
		return nil, err
	}

	var ciIDs []uint64
	for _, result := range results {
		ciIDs = append(ciIDs, result.CiID)
	}

	return ciIDs, nil
}

// processFilterGroups 处理复合过滤组
func (l *GetCisListLogic) processFilterGroups(groups []*cmdb.CiFilterGroup) ([]uint64, error) {
	var allCiIDs []uint64

	for _, group := range groups {
		ciIDs, err := l.processFilterGroup(group)
		if err != nil {
			return nil, err
		}

		if group.Logic == "or" {
			allCiIDs = l.unionCiIDs(allCiIDs, ciIDs)
		} else {
			allCiIDs = l.intersectCiIDs(allCiIDs, ciIDs)
		}
	}

	return allCiIDs, nil
}

// processFilterGroup 处理单个过滤组
func (l *GetCisListLogic) processFilterGroup(group *cmdb.CiFilterGroup) ([]uint64, error) {
	var groupCiIDs []uint64

	// 处理过滤条件
	if len(group.Filters) > 0 {
		filterCiIDs, err := l.processAttributeFilters(group.Filters)
		if err != nil {
			return nil, err
		}
		groupCiIDs = filterCiIDs
	}

	// 处理嵌套组
	if len(group.Groups) > 0 {
		nestedCiIDs, err := l.processFilterGroups(group.Groups)
		if err != nil {
			return nil, err
		}

		if group.Logic == "or" {
			groupCiIDs = l.unionCiIDs(groupCiIDs, nestedCiIDs)
		} else {
			groupCiIDs = l.intersectCiIDs(groupCiIDs, nestedCiIDs)
		}
	}

	return groupCiIDs, nil
}

// buildSearchPredicates 构建全文搜索条件
func (l *GetCisListLogic) buildSearchPredicates(search string, searchFields []uint64) ([]predicate.Cis, error) {
	l.Logger.Infof("开始全文搜索: search='%s', searchFields=%v", search, searchFields)

	var allCiIDs []uint64

	// 在ValueText表中搜索
	textQuery := l.svcCtx.DB.ValueText.Query().Where(valuetext.ValueContains(search))
	if len(searchFields) > 0 {
		textQuery = textQuery.Where(valuetext.AttrIDIn(searchFields...))
	}

	textResults, err := textQuery.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("ValueText表全文搜索失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("ValueText表搜索结果数量: %d", len(textResults))
	for _, result := range textResults {
		allCiIDs = append(allCiIDs, result.CiID)
		l.Logger.Debugf("ValueText搜索结果: CiID=%d, Value='%s'", result.CiID, result.Value)
	}

	// 同时在ValueIndexText表中搜索
	indexTextQuery := l.svcCtx.DB.ValueIndexText.Query().Where(valueindextext.ValueContains(search))
	if len(searchFields) > 0 {
		indexTextQuery = indexTextQuery.Where(valueindextext.AttrIDIn(searchFields...))
	}

	indexTextResults, err := indexTextQuery.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("ValueIndexText表全文搜索失败: %v", err)
		// 不返回错误，继续使用ValueText的结果
	} else {
		l.Logger.Infof("ValueIndexText表搜索结果数量: %d", len(indexTextResults))
		for _, result := range indexTextResults {
			allCiIDs = append(allCiIDs, result.CiID)
			l.Logger.Debugf("ValueIndexText搜索结果: CiID=%d, Value='%s'", result.CiID, result.Value)
		}
	}

	// 去重CI ID
	ciIDMap := make(map[uint64]bool)
	var uniqueCiIDs []uint64
	for _, ciID := range allCiIDs {
		if !ciIDMap[ciID] {
			ciIDMap[ciID] = true
			uniqueCiIDs = append(uniqueCiIDs, ciID)
		}
	}

	l.Logger.Infof("全文搜索最终结果CI数量: %d", len(uniqueCiIDs))

	// 如果有搜索条件但没有找到匹配的CI，返回一个永远不匹配的条件
	if len(uniqueCiIDs) > 0 {
		l.Logger.Infof("返回匹配的CI ID列表进行全文搜索")
		return []predicate.Cis{cis.IDIn(uniqueCiIDs...)}, nil
	} else {
		l.Logger.Errorf("全文搜索没有找到匹配的CI，返回空结果条件")
		// 返回一个永远不匹配的条件，确保搜索结果为空
		return []predicate.Cis{cis.IDEQ(0)}, nil
	}
}

// applySorting 应用排序
func (l *GetCisListLogic) applySorting(query *ent.CisQuery, sortFields []*cmdb.CiSortField) *ent.CisQuery {
	for _, sortField := range sortFields {
		if sortField.AttrId == 0 {
			// 基础字段排序
			if sortField.FieldName != nil {
				switch *sortField.FieldName {
				case "id":
					if sortField.Direction == "desc" {
						query = query.Order(ent.Desc(cis.FieldID))
					} else {
						query = query.Order(ent.Asc(cis.FieldID))
					}
				case "created_at":
					if sortField.Direction == "desc" {
						query = query.Order(ent.Desc(cis.FieldCreatedAt))
					} else {
						query = query.Order(ent.Asc(cis.FieldCreatedAt))
					}
				case "updated_at":
					if sortField.Direction == "desc" {
						query = query.Order(ent.Desc(cis.FieldUpdatedAt))
					} else {
						query = query.Order(ent.Asc(cis.FieldUpdatedAt))
					}
				}
			}
		}
		// 属性字段排序比较复杂，这里暂时跳过
	}
	return query
}

// buildAggregations 构建聚合统计
func (l *GetCisListLogic) buildAggregations(predicates []predicate.Cis, aggregationFields []uint64) (string, error) {
	// 简化的聚合实现，返回基础统计信息
	aggregations := make(map[string]interface{})

	// 统计总数
	total, err := l.svcCtx.DB.Cis.Query().Where(predicates...).Count(l.ctx)
	if err != nil {
		return "", err
	}
	aggregations["total"] = total

	// 按状态统计
	statusStats := make(map[string]int)
	statusResults, err := l.svcCtx.DB.Cis.Query().Where(predicates...).All(l.ctx)
	if err != nil {
		return "", err
	}

	for _, result := range statusResults {
		statusKey := fmt.Sprintf("status_%d", result.Status)
		statusStats[statusKey]++
	}
	aggregations["status_stats"] = statusStats

	jsonData, err := json.Marshal(aggregations)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// intersectCiIDs 计算CI ID的交集
func (l *GetCisListLogic) intersectCiIDs(a, b []uint64) []uint64 {
	// 如果是第一次调用（a为空），直接返回b
	if len(a) == 0 {
		return b
	}
	// 如果b为空，说明没有匹配的结果，返回空数组
	if len(b) == 0 {
		return []uint64{}
	}

	m := make(map[uint64]bool)
	for _, id := range a {
		m[id] = true
	}

	var result []uint64
	for _, id := range b {
		if m[id] {
			result = append(result, id)
		}
	}

	return result
}

// unionCiIDs 计算CI ID的并集
func (l *GetCisListLogic) unionCiIDs(a, b []uint64) []uint64 {
	m := make(map[uint64]bool)
	var result []uint64

	for _, id := range a {
		if !m[id] {
			m[id] = true
			result = append(result, id)
		}
	}

	for _, id := range b {
		if !m[id] {
			m[id] = true
			result = append(result, id)
		}
	}

	return result
}
