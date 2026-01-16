package cirelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiRelationsBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiRelationsBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiRelationsBatchLogic {
	return &GetCiRelationsBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiRelationsBatchLogic) GetCiRelationsBatch(in *cmdb.CiRelationBatchQueryReq) (*cmdb.CiRelationBatchQueryResp, error) {
	if len(in.CiIds) == 0 {
		return &cmdb.CiRelationBatchQueryResp{
			RelationsByCi:  make(map[uint64]*cmdb.CiRelationGroup),
			TotalRelations: 0,
			TotalCis:       0,
		}, nil
	}

	resp := &cmdb.CiRelationBatchQueryResp{
		RelationsByCi: make(map[uint64]*cmdb.CiRelationGroup),
	}

	// 默认方向为both
	direction := "both"
	if in.Direction != nil {
		direction = *in.Direction
	}

	var totalRelations uint64

	// 为每个CI ID创建关系分组
	for _, ciId := range in.CiIds {
		group := &cmdb.CiRelationGroup{
			CiId:            ciId,
			SourceRelations: []*cmdb.CiRelationInfo{},
			TargetRelations: []*cmdb.CiRelationInfo{},
			TotalCount:      0,
		}

		// 查询作为源CI的关系 (source_ci_id = ci_id)
		if direction == "source" || direction == "both" {
			sourceRelations, err := l.queryRelationsBySourceCi(ciId, in)
			if err != nil {
				return nil, err
			}
			
			for _, relation := range sourceRelations {
				relationInfo := l.convertToRelationInfo(relation, in.IncludeAttributeMappings)
				group.SourceRelations = append(group.SourceRelations, relationInfo)
			}
			group.TotalCount += uint64(len(sourceRelations))
			totalRelations += uint64(len(sourceRelations))
		}

		// 查询作为目标CI的关系 (target_ci_id = ci_id)
		if direction == "target" || direction == "both" {
			targetRelations, err := l.queryRelationsByTargetCi(ciId, in)
			if err != nil {
				return nil, err
			}
			
			for _, relation := range targetRelations {
				relationInfo := l.convertToRelationInfo(relation, in.IncludeAttributeMappings)
				group.TargetRelations = append(group.TargetRelations, relationInfo)
			}
			group.TotalCount += uint64(len(targetRelations))
			totalRelations += uint64(len(targetRelations))
		}

		resp.RelationsByCi[ciId] = group
	}

	resp.TotalRelations = totalRelations
	resp.TotalCis = uint64(len(in.CiIds))

	return resp, nil
}

// queryRelationsBySourceCi 查询作为源CI的关系
func (l *GetCiRelationsBatchLogic) queryRelationsBySourceCi(ciId uint64, in *cmdb.CiRelationBatchQueryReq) ([]*ent.CiRelation, error) {
	var predicates []predicate.CiRelation
	
	// 基础条件：该CI作为源CI
	predicates = append(predicates, cirelation.SourceCiIDEQ(ciId))
	
	// 应用过滤条件
	predicates = append(predicates, l.buildFilterPredicates(in)...)

	return l.svcCtx.DB.CiRelation.Query().
		Where(predicates...).
		WithSourceCi().   // 预加载源CI信息
		WithTargetCi().   // 预加载目标CI信息
		WithRelationType(). // 预加载关系类型信息
		All(l.ctx)
}

// queryRelationsByTargetCi 查询作为目标CI的关系
func (l *GetCiRelationsBatchLogic) queryRelationsByTargetCi(ciId uint64, in *cmdb.CiRelationBatchQueryReq) ([]*ent.CiRelation, error) {
	var predicates []predicate.CiRelation
	
	// 基础条件：该CI作为目标CI
	predicates = append(predicates, cirelation.TargetCiIDEQ(ciId))
	
	// 应用过滤条件
	predicates = append(predicates, l.buildFilterPredicates(in)...)

	return l.svcCtx.DB.CiRelation.Query().
		Where(predicates...).
		WithSourceCi().   // 预加载源CI信息
		WithTargetCi().   // 预加载目标CI信息
		WithRelationType(). // 预加载关系类型信息
		All(l.ctx)
}

// buildFilterPredicates 构建过滤条件
func (l *GetCiRelationsBatchLogic) buildFilterPredicates(in *cmdb.CiRelationBatchQueryReq) []predicate.CiRelation {
	var predicates []predicate.CiRelation

	// 关系类型ID过滤
	if len(in.RelationTypeIds) > 0 {
		predicates = append(predicates, cirelation.RelationTypeIDIn(in.RelationTypeIds...))
	}

	// 状态过滤
	if len(in.StatusFilter) > 0 {
		predicates = append(predicates, cirelation.StatusIn(in.StatusFilter...))
	}

	// 发现来源过滤
	if len(in.DiscoverySourceFilter) > 0 {
		predicates = append(predicates, cirelation.DiscoverySourceIn(in.DiscoverySourceFilter...))
	}

	// 只查询激活状态的关系
	if in.ActiveOnly != nil && *in.ActiveOnly {
		predicates = append(predicates, cirelation.StatusEQ("active"))
	}

	return predicates
}

// convertToRelationInfo 将ent模型转换为Proto响应格式
func (l *GetCiRelationsBatchLogic) convertToRelationInfo(relation *ent.CiRelation, includeAttributeMappings *bool) *cmdb.CiRelationInfo {
	info := &cmdb.CiRelationInfo{
		Id:             &relation.ID,
		CreatedAt:      pointy.GetPointer(relation.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(relation.UpdatedAt.UnixMilli()),
		SourceCiId:     &relation.SourceCiID,
		TargetCiId:     &relation.TargetCiID,
		RelationTypeId: &relation.RelationTypeID,
	}

	// 可选字段
	if relation.More != 0 {
		info.More = &relation.More
	}
	if relation.DiscoverySource != "" {
		info.DiscoverySource = &relation.DiscoverySource
	}
	if relation.AncestorIds != "" {
		info.AncestorIds = &relation.AncestorIds
	}
	if relation.Properties != nil {
		// 这里需要处理JSON字段到字符串的转换
		// 暂时跳过，或者根据实际需要进行序列化
	}
	if relation.Status != "" {
		info.Status = &relation.Status
	}
	if relation.RelationStrength != "" {
		info.RelationStrength = &relation.RelationStrength
	}

	// 扩展字段
	if relation.AutoSyncEnabled {
		info.AutoSyncEnabled = &relation.AutoSyncEnabled
	}

	// 如果需要包含属性映射信息
	if includeAttributeMappings == nil || *includeAttributeMappings {
		// 这里需要处理AttributeMappings等JSON字段
		// 根据实际需要进行处理
	}

	return info
}
