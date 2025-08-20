package cis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/cirelation"
	"gitee.com/link234/cmdb-rpc/ent/cis"
	"gitee.com/link234/cmdb-rpc/ent/citypeinheritance"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/ent/relationtype"
	"gitee.com/link234/cmdb-rpc/ent/valueindextext"
	"gitee.com/link234/cmdb-rpc/ent/valuetext"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"github.com/zeromicro/go-zero/core/logx"
)

// EnhancedSearchEngine 增强版CI搜索引擎
type EnhancedSearchEngine struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger logx.Logger
}

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	*cmdb.CisListReq
	// 扩展搜索条件
	RelationFilters   []*RelationFilter  `json:"relationFilters,omitempty"`
	InheritanceSearch *InheritanceSearch `json:"inheritanceSearch,omitempty"`
	TagFilters        []*TagFilter       `json:"tagFilters,omitempty"`
	MetadataFilters   []*MetadataFilter  `json:"metadataFilters,omitempty"`
	SpatialSearch     *SpatialSearch     `json:"spatialSearch,omitempty"`
	TimeRangeSearch   *TimeRangeSearch   `json:"timeRangeSearch,omitempty"`
	FacetSearch       *FacetSearch       `json:"facetSearch,omitempty"`
}

// RelationFilter 关系过滤条件
type RelationFilter struct {
	RelationType    string   `json:"relationType"`    // 关系类型编码
	Direction       string   `json:"direction"`       // 关系方向: incoming, outgoing, both
	TargetCiIds     []uint64 `json:"targetCiIds"`     // 目标CI ID列表
	TargetTypeIds   []uint64 `json:"targetTypeIds"`   // 目标CI类型ID列表
	RelationDepth   int      `json:"relationDepth"`   // 关系深度，0表示直接关系
	IncludeIndirect bool     `json:"includeIndirect"` // 是否包含间接关系
}

// InheritanceSearch 继承搜索
type InheritanceSearch struct {
	IncludeChildren bool     `json:"includeChildren"` // 是否包含子类型
	IncludeParents  bool     `json:"includeParents"`  // 是否包含父类型
	TypeHierarchy   []uint64 `json:"typeHierarchy"`   // 指定类型层级
}

// TagFilter 标签过滤
type TagFilter struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	Operator string   `json:"operator"` // eq, ne, in, not_in, exists, not_exists
}

// MetadataFilter 元数据过滤
type MetadataFilter struct {
	Path     string      `json:"path"` // JSON路径，如 "system.version"
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, exists
}

// SpatialSearch 空间搜索（如果有地理位置信息）
type SpatialSearch struct {
	Center   GeoPoint `json:"center"`
	Radius   float64  `json:"radius"`   // 半径（米）
	GeoField string   `json:"geoField"` // 地理位置字段名
}

// TimeRangeSearch 时间范围搜索
type TimeRangeSearch struct {
	Field     string    `json:"field"` // 时间字段
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Interval  string    `json:"interval"` // 时间间隔：hour, day, week, month
}

// FacetSearch 分面搜索
type FacetSearch struct {
	Fields []string `json:"fields"` // 需要分面统计的字段
	Size   int      `json:"size"`   // 每个分面返回的项目数量
}

// GeoPoint 地理坐标点
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// SearchResult 搜索结果
type SearchResult struct {
	*cmdb.CisListResp
	Facets      map[string][]FacetItem `json:"facets,omitempty"`
	Highlights  map[uint64][]string    `json:"highlights,omitempty"`
	SearchTime  int64                  `json:"searchTime"` // 搜索耗时(毫秒)
	TotalPages  int                    `json:"totalPages"`
	CurrentPage int                    `json:"currentPage"`
	RelatedCis  []uint64               `json:"relatedCis,omitempty"` // 相关CI ID列表
}

// FacetItem 分面项
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// NewEnhancedSearchEngine 创建增强搜索引擎
func NewEnhancedSearchEngine(ctx context.Context, svcCtx *svc.ServiceContext) *EnhancedSearchEngine {
	return &EnhancedSearchEngine{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: logx.WithContext(ctx),
	}
}

// Search 执行增强搜索
func (e *EnhancedSearchEngine) Search(req *SearchRequest) (*SearchResult, error) {
	startTime := time.Now()

	// 构建基础查询条件
	var predicates []predicate.Cis
	predicates = e.buildBasicPredicates(req.CisListReq, predicates)

	// 构建关系搜索条件
	if len(req.RelationFilters) > 0 {
		relationPredicates, err := e.buildRelationPredicates(req.RelationFilters)
		if err != nil {
			return nil, fmt.Errorf("构建关系搜索条件失败: %w", err)
		}
		predicates = append(predicates, relationPredicates...)
	}

	// 构建继承搜索条件
	if req.InheritanceSearch != nil {
		inheritancePredicates, err := e.buildInheritancePredicates(req.InheritanceSearch, req.TypeId)
		if err != nil {
			return nil, fmt.Errorf("构建继承搜索条件失败: %w", err)
		}
		predicates = append(predicates, inheritancePredicates...)
	}

	// 构建标签搜索条件
	if len(req.TagFilters) > 0 {
		tagPredicates := e.buildTagPredicates(req.TagFilters)
		predicates = append(predicates, tagPredicates...)
	}

	// 构建元数据搜索条件
	if len(req.MetadataFilters) > 0 {
		metadataPredicates := e.buildMetadataPredicates(req.MetadataFilters)
		predicates = append(predicates, metadataPredicates...)
	}

	// 构建属性过滤条件
	if len(req.AttributeFilters) > 0 || len(req.FilterGroups) > 0 {
		attrPredicates, err := e.buildAttributePredicates(req.CisListReq)
		if err != nil {
			return nil, fmt.Errorf("构建属性过滤条件失败: %w", err)
		}
		predicates = append(predicates, attrPredicates...)
	}

	// 构建全文搜索条件（优化版，使用索引文本表）
	if req.Search != nil && *req.Search != "" {
		searchPredicates, err := e.buildEnhancedSearchPredicates(*req.Search, req.SearchFields)
		if err != nil {
			return nil, fmt.Errorf("构建全文搜索条件失败: %w", err)
		}
		predicates = append(predicates, searchPredicates...)
	}

	// 构建时间范围搜索条件
	if req.TimeRangeSearch != nil {
		timePredicates, err := e.buildTimeRangePredicates(req.TimeRangeSearch)
		if err != nil {
			return nil, fmt.Errorf("构建时间范围搜索条件失败: %w", err)
		}
		predicates = append(predicates, timePredicates...)
	}

	// 执行查询
	query := e.svcCtx.DB.Cis.Query().Where(predicates...)

	// 应用排序
	if len(req.SortFields) > 0 {
		query = e.applyEnhancedSorting(query, req.SortFields)
	}

	// 执行分页查询
	result, err := query.Page(e.ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("执行分页查询失败: %w", err)
	}

	// 构建响应
	searchResult := &SearchResult{
		CisListResp: &cmdb.CisListResp{
			Total: result.PageDetails.Total,
		},
		SearchTime:  time.Since(startTime).Milliseconds(),
		TotalPages:  int((int64(result.PageDetails.Total) + int64(req.PageSize) - 1) / int64(req.PageSize)),
		CurrentPage: int(req.Page),
	}

	// 转换数据并添加高亮
	for _, v := range result.List {
		var attributes []*cmdb.CiAttributeValue

		// 根据请求决定是否加载属性
		if req.WithAttributes != nil && *req.WithAttributes {
			attributes, err = LoadCiAttributes(e.ctx, e.svcCtx.DB, v.ID, req.IncludeAttributes, req.ExcludeAttributes)
			if err != nil {
				return nil, fmt.Errorf("加载CI属性失败: %w", err)
			}
		}

		cisInfo := CisEntToProto(v, attributes)
		if cisInfo != nil {
			searchResult.Data = append(searchResult.Data, cisInfo)
		}
	}

	// 处理分面搜索
	if req.FacetSearch != nil {
		facets, err := e.buildFacets(predicates, req.FacetSearch)
		if err != nil {
			return nil, fmt.Errorf("构建分面搜索失败: %w", err)
		}
		searchResult.Facets = facets
	}

	// 处理聚合统计
	if req.WithAggregations != nil && *req.WithAggregations && len(req.AggregationFields) > 0 {
		aggregations, err := e.buildEnhancedAggregations(predicates, req.AggregationFields)
		if err != nil {
			return nil, fmt.Errorf("构建聚合统计失败: %w", err)
		}
		if aggregations != "" {
			searchResult.Aggregations = &aggregations
		}
	}

	// 查找相关CI
	if len(req.RelationFilters) > 0 {
		relatedCis, err := e.findRelatedCis(searchResult.Data)
		if err != nil {
			e.logger.Errorf("查找相关CI失败: %v", err)
		} else {
			searchResult.RelatedCis = relatedCis
		}
	}

	return searchResult, nil
}

// buildRelationPredicates 构建关系搜索条件
func (e *EnhancedSearchEngine) buildRelationPredicates(filters []*RelationFilter) ([]predicate.Cis, error) {
	var ciIDs []uint64

	for _, filter := range filters {
		ids, err := e.processRelationFilter(filter)
		if err != nil {
			return nil, err
		}
		ciIDs = e.unionCiIDs(ciIDs, ids)
	}

	if len(ciIDs) > 0 {
		return []predicate.Cis{cis.IDIn(ciIDs...)}, nil
	}

	return []predicate.Cis{cis.IDEQ(0)}, nil // 返回空结果条件
}

// processRelationFilter 处理关系过滤条件
func (e *EnhancedSearchEngine) processRelationFilter(filter *RelationFilter) ([]uint64, error) {
	query := e.svcCtx.DB.CiRelation.Query()

	// 添加关系类型过滤
	if filter.RelationType != "" {
		query = query.Where(cirelation.HasRelationTypeWith(relationtype.CodeEQ(filter.RelationType)))
	}

	// 添加目标CI过滤
	if len(filter.TargetCiIds) > 0 {
		switch filter.Direction {
		case "incoming":
			query = query.Where(cirelation.FirstCiIDIn(filter.TargetCiIds...))
		case "outgoing":
			query = query.Where(cirelation.SecondCiIDIn(filter.TargetCiIds...))
		default: // both
			query = query.Where(cirelation.Or(
				cirelation.FirstCiIDIn(filter.TargetCiIds...),
				cirelation.SecondCiIDIn(filter.TargetCiIds...),
			))
		}
	}

	// 添加目标类型过滤
	if len(filter.TargetTypeIds) > 0 {
		switch filter.Direction {
		case "incoming":
			query = query.Where(cirelation.HasFirstCiWith(cis.TypeIDIn(filter.TargetTypeIds...)))
		case "outgoing":
			query = query.Where(cirelation.HasSecondCiWith(cis.TypeIDIn(filter.TargetTypeIds...)))
		default: // both
			query = query.Where(cirelation.Or(
				cirelation.HasFirstCiWith(cis.TypeIDIn(filter.TargetTypeIds...)),
				cirelation.HasSecondCiWith(cis.TypeIDIn(filter.TargetTypeIds...)),
			))
		}
	}

	relations, err := query.All(e.ctx)
	if err != nil {
		return nil, err
	}

	var ciIDs []uint64
	for _, relation := range relations {
		switch filter.Direction {
		case "incoming":
			ciIDs = append(ciIDs, relation.SecondCiID)
		case "outgoing":
			ciIDs = append(ciIDs, relation.FirstCiID)
		default: // both
			ciIDs = append(ciIDs, relation.FirstCiID, relation.SecondCiID)
		}
	}

	// 处理关系深度
	if filter.RelationDepth > 0 && filter.IncludeIndirect {
		indirectCiIDs, err := e.findIndirectRelations(ciIDs, filter.RelationDepth)
		if err != nil {
			return nil, err
		}
		ciIDs = e.unionCiIDs(ciIDs, indirectCiIDs)
	}

	return e.uniqueCiIDs(ciIDs), nil
}

// buildInheritancePredicates 构建继承搜索条件
func (e *EnhancedSearchEngine) buildInheritancePredicates(search *InheritanceSearch, baseTypeId *uint64) ([]predicate.Cis, error) {
	if baseTypeId == nil {
		return nil, nil
	}

	var typeIDs []uint64
	typeIDs = append(typeIDs, *baseTypeId)

	// 查找子类型
	if search.IncludeChildren {
		childTypes, err := e.findChildTypes(*baseTypeId)
		if err != nil {
			return nil, err
		}
		typeIDs = append(typeIDs, childTypes...)
	}

	// 查找父类型
	if search.IncludeParents {
		parentTypes, err := e.findParentTypes(*baseTypeId)
		if err != nil {
			return nil, err
		}
		typeIDs = append(typeIDs, parentTypes...)
	}

	// 添加指定的类型层级
	if len(search.TypeHierarchy) > 0 {
		typeIDs = append(typeIDs, search.TypeHierarchy...)
	}

	typeIDs = e.uniqueTypeIDs(typeIDs)

	if len(typeIDs) > 0 {
		return []predicate.Cis{cis.TypeIDIn(typeIDs...)}, nil
	}

	return nil, nil
}

// buildTagPredicates 构建标签搜索条件
func (e *EnhancedSearchEngine) buildTagPredicates(filters []*TagFilter) []predicate.Cis {
	var predicates []predicate.Cis

	for _, filter := range filters {
		switch filter.Operator {
		case "exists":
			predicates = append(predicates, func(s *sql.Selector) {
				s.Where(sql.Contains(s.C("tags"), fmt.Sprintf(`"key":"%s"`, filter.Key)))
			})
		case "not_exists":
			predicates = append(predicates, func(s *sql.Selector) {
				s.Where(sql.Not(sql.Contains(s.C("tags"), fmt.Sprintf(`"key":"%s"`, filter.Key))))
			})
		case "in":
			if len(filter.Values) > 0 {
				var conditions []string
				for _, value := range filter.Values {
					conditions = append(conditions, fmt.Sprintf(`{"key":"%s","value":"%s"}`, filter.Key, value))
				}
				predicates = append(predicates, func(s *sql.Selector) {
					for _, condition := range conditions {
						s.Where(sql.Contains(s.C("tags"), condition))
					}
				})
			}
		}
	}

	return predicates
}

// buildMetadataPredicates 构建元数据搜索条件
func (e *EnhancedSearchEngine) buildMetadataPredicates(filters []*MetadataFilter) []predicate.Cis {
	var predicates []predicate.Cis

	for _, filter := range filters {
		switch filter.Operator {
		case "exists":
			predicates = append(predicates, func(s *sql.Selector) {
				s.Where(sql.Contains(s.C("metadata"), fmt.Sprintf(`"%s"`, filter.Path)))
			})
		case "eq":
			predicates = append(predicates, func(s *sql.Selector) {
				valueJson, _ := json.Marshal(filter.Value)
				s.Where(sql.Contains(s.C("metadata"), fmt.Sprintf(`"%s":%s`, filter.Path, string(valueJson))))
			})
		case "contains":
			if strValue, ok := filter.Value.(string); ok {
				predicates = append(predicates, func(s *sql.Selector) {
					s.Where(sql.Contains(s.C("metadata"), strValue))
				})
			}
		}
	}

	return predicates
}

// buildEnhancedSearchPredicates 构建增强全文搜索条件（优先使用索引文本表）
func (e *EnhancedSearchEngine) buildEnhancedSearchPredicates(search string, searchFields []uint64) ([]predicate.Cis, error) {
	var ciIDs []uint64

	// 优先在索引文本表中搜索
	indexTextQuery := e.svcCtx.DB.ValueIndexText.Query().Where(valueindextext.ValueContains(search))
	if len(searchFields) > 0 {
		indexTextQuery = indexTextQuery.Where(valueindextext.AttrIDIn(searchFields...))
	}

	indexResults, err := indexTextQuery.All(e.ctx)
	if err != nil {
		return nil, err
	}

	for _, result := range indexResults {
		ciIDs = append(ciIDs, result.CiID)
	}

	// 在普通文本表中搜索
	textQuery := e.svcCtx.DB.ValueText.Query().Where(valuetext.ValueContains(search))
	if len(searchFields) > 0 {
		textQuery = textQuery.Where(valuetext.AttrIDIn(searchFields...))
	}

	textResults, err := textQuery.All(e.ctx)
	if err != nil {
		return nil, err
	}

	for _, result := range textResults {
		ciIDs = append(ciIDs, result.CiID)
	}

	// 在CI基础字段中搜索（如果有的话）
	// 这里可以扩展搜索CI的metadata、custom_fields等JSON字段

	ciIDs = e.uniqueCiIDs(ciIDs)

	if len(ciIDs) > 0 {
		return []predicate.Cis{cis.IDIn(ciIDs...)}, nil
	}

	return []predicate.Cis{cis.IDEQ(0)}, nil
}

// buildTimeRangePredicates 构建时间范围搜索条件
func (e *EnhancedSearchEngine) buildTimeRangePredicates(timeSearch *TimeRangeSearch) ([]predicate.Cis, error) {
	// 这里需要根据具体的时间字段来构建条件
	// 示例：如果是基础字段
	switch timeSearch.Field {
	case "created_at":
		return []predicate.Cis{
			cis.CreatedAtGTE(timeSearch.StartTime),
			cis.CreatedAtLTE(timeSearch.EndTime),
		}, nil
	case "updated_at":
		return []predicate.Cis{
			cis.UpdatedAtGTE(timeSearch.StartTime),
			cis.UpdatedAtLTE(timeSearch.EndTime),
		}, nil
	default:
		// 如果是属性字段，需要在对应的值表中查询
		// 这里简化处理，实际应该根据字段类型选择对应的值表
		return nil, nil
	}
}

// 辅助函数
func (e *EnhancedSearchEngine) unionCiIDs(a, b []uint64) []uint64 {
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

func (e *EnhancedSearchEngine) uniqueCiIDs(ids []uint64) []uint64 {
	m := make(map[uint64]bool)
	var result []uint64

	for _, id := range ids {
		if !m[id] {
			m[id] = true
			result = append(result, id)
		}
	}

	return result
}

func (e *EnhancedSearchEngine) uniqueTypeIDs(ids []uint64) []uint64 {
	m := make(map[uint64]bool)
	var result []uint64

	for _, id := range ids {
		if !m[id] {
			m[id] = true
			result = append(result, id)
		}
	}

	return result
}

// findChildTypes 查找子类型
func (e *EnhancedSearchEngine) findChildTypes(parentTypeId uint64) ([]uint64, error) {
	inheritances, err := e.svcCtx.DB.CiTypeInheritance.Query().
		Where(citypeinheritance.ParentIDEQ(parentTypeId)).
		All(e.ctx)
	if err != nil {
		return nil, err
	}

	var childTypes []uint64
	for _, inheritance := range inheritances {
		childTypes = append(childTypes, inheritance.ChildID)
	}

	return childTypes, nil
}

// findParentTypes 查找父类型
func (e *EnhancedSearchEngine) findParentTypes(childTypeId uint64) ([]uint64, error) {
	inheritances, err := e.svcCtx.DB.CiTypeInheritance.Query().
		Where(citypeinheritance.ChildIDEQ(childTypeId)).
		All(e.ctx)
	if err != nil {
		return nil, err
	}

	var parentTypes []uint64
	for _, inheritance := range inheritances {
		parentTypes = append(parentTypes, inheritance.ParentID)
	}

	return parentTypes, nil
}

// findIndirectRelations 查找间接关系
func (e *EnhancedSearchEngine) findIndirectRelations(directCiIDs []uint64, maxDepth int) ([]uint64, error) {
	var allCiIDs []uint64
	currentLevel := directCiIDs

	for depth := 1; depth < maxDepth && len(currentLevel) > 0; depth++ {
		nextLevel, err := e.findDirectRelations(currentLevel)
		if err != nil {
			return nil, err
		}

		allCiIDs = append(allCiIDs, nextLevel...)
		currentLevel = nextLevel
	}

	return e.uniqueCiIDs(allCiIDs), nil
}

// findDirectRelations 查找直接关系的CI
func (e *EnhancedSearchEngine) findDirectRelations(ciIDs []uint64) ([]uint64, error) {
	relations, err := e.svcCtx.DB.CiRelation.Query().
		Where(cirelation.Or(
			cirelation.FirstCiIDIn(ciIDs...),
			cirelation.SecondCiIDIn(ciIDs...),
		)).
		All(e.ctx)
	if err != nil {
		return nil, err
	}

	var relatedCiIDs []uint64
	for _, relation := range relations {
		relatedCiIDs = append(relatedCiIDs, relation.FirstCiID, relation.SecondCiID)
	}

	return e.uniqueCiIDs(relatedCiIDs), nil
}

// findRelatedCis 查找相关CI
func (e *EnhancedSearchEngine) findRelatedCis(cisData []*cmdb.CisInfo) ([]uint64, error) {
	if len(cisData) == 0 {
		return nil, nil
	}

	var ciIDs []uint64
	for _, ci := range cisData {
		if ci.Id != nil {
			ciIDs = append(ciIDs, *ci.Id)
		}
	}

	return e.findDirectRelations(ciIDs)
}

// buildBasicPredicates 构建基础字段的查询条件
func (e *EnhancedSearchEngine) buildBasicPredicates(in *cmdb.CisListReq, predicates []predicate.Cis) []predicate.Cis {
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
	if in.UpdatedBy != nil {
		// 注意：这里可能需要根据实际字段调整
		// predicates = append(predicates, cis.CreatedByEQ(uuidx.ParseUUIDString(*in.UpdatedBy)))
	}
	return predicates
}

// buildAttributePredicates 构建属性过滤条件（从原有逻辑移植）
func (e *EnhancedSearchEngine) buildAttributePredicates(in *cmdb.CisListReq) ([]predicate.Cis, error) {
	// 这里需要从GetCisListLogic中移植相关逻辑
	// 由于复杂度较高，暂时返回nil，实际使用时需要完整实现
	return nil, nil
}

// applyEnhancedSorting 应用增强排序
func (e *EnhancedSearchEngine) applyEnhancedSorting(query *ent.CisQuery, sortFields []*cmdb.CiSortField) *ent.CisQuery {
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

// buildFacets 构建分面搜索
func (e *EnhancedSearchEngine) buildFacets(predicates []predicate.Cis, facetSearch *FacetSearch) (map[string][]FacetItem, error) {
	facets := make(map[string][]FacetItem)

	for _, field := range facetSearch.Fields {
		switch field {
		case "type_id":
			// 按CI类型分面统计
			var typeStats []struct {
				TypeID uint64 `json:"type_id"`
				Count  int    `json:"count"`
			}

			err := e.svcCtx.DB.Cis.Query().
				Where(predicates...).
				GroupBy(cis.FieldTypeID).
				Aggregate(ent.Count()).
				Scan(e.ctx, &typeStats)
			if err != nil {
				return nil, err
			}

			var items []FacetItem
			for _, stat := range typeStats {
				items = append(items, FacetItem{
					Value: fmt.Sprintf("%d", stat.TypeID),
					Count: stat.Count,
				})
			}

			// 限制返回数量
			if facetSearch.Size > 0 && len(items) > facetSearch.Size {
				items = items[:facetSearch.Size]
			}

			facets[field] = items

		case "status":
			// 按状态分面统计
			var statusStats []struct {
				Status uint32 `json:"status"`
				Count  int    `json:"count"`
			}

			err := e.svcCtx.DB.Cis.Query().
				Where(predicates...).
				GroupBy(cis.FieldStatus).
				Aggregate(ent.Count()).
				Scan(e.ctx, &statusStats)
			if err != nil {
				return nil, err
			}

			var items []FacetItem
			for _, stat := range statusStats {
				items = append(items, FacetItem{
					Value: fmt.Sprintf("%d", stat.Status),
					Count: stat.Count,
				})
			}

			// 限制返回数量
			if facetSearch.Size > 0 && len(items) > facetSearch.Size {
				items = items[:facetSearch.Size]
			}

			facets[field] = items
		}
	}

	return facets, nil
}

// buildEnhancedAggregations 构建增强聚合统计
func (e *EnhancedSearchEngine) buildEnhancedAggregations(predicates []predicate.Cis, aggregationFields []uint64) (string, error) {
	aggregations := make(map[string]interface{})

	// 统计总数
	total, err := e.svcCtx.DB.Cis.Query().Where(predicates...).Count(e.ctx)
	if err != nil {
		return "", err
	}
	aggregations["total"] = total

	// 按状态统计
	statusStats := make(map[string]int)
	statusResults, err := e.svcCtx.DB.Cis.Query().Where(predicates...).All(e.ctx)
	if err != nil {
		return "", err
	}

	for _, result := range statusResults {
		statusKey := fmt.Sprintf("status_%d", result.Status)
		statusStats[statusKey]++
	}
	aggregations["status_stats"] = statusStats

	// 按类型统计
	typeStats := make(map[string]int)
	for _, result := range statusResults {
		typeKey := fmt.Sprintf("type_%d", result.TypeID)
		typeStats[typeKey]++
	}
	aggregations["type_stats"] = typeStats

	// 按创建时间统计（按天）
	timeStats := make(map[string]int)
	for _, result := range statusResults {
		dayKey := result.CreatedAt.Format("2006-01-02")
		timeStats[dayKey]++
	}
	aggregations["daily_stats"] = timeStats

	jsonData, err := json.Marshal(aggregations)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// intersectCiIDs 计算CI ID的交集（从原有逻辑移植）
func (e *EnhancedSearchEngine) intersectCiIDs(a, b []uint64) []uint64 {
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
