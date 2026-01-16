package cityperelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cityperelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeRelationDefinitionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeRelationDefinitionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeRelationDefinitionsLogic {
	return &GetCiTypeRelationDefinitionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeRelationDefinitionsLogic) GetCiTypeRelationDefinitions(in *cmdb.CiTypeRelationDefinitionReq) (*cmdb.CiTypeRelationDefinitionResp, error) {
	resp := &cmdb.CiTypeRelationDefinitionResp{}
	var totalCount uint64

	// 默认方向为both
	direction := "both"
	if in.Direction != nil {
		direction = *in.Direction
	}

	// 查询作为源CI类型的关系定义 (parent_id = ci_type_id)
	if direction == "source" || direction == "both" {
		var sourcePredicates []predicate.CiTypeRelation
		sourcePredicates = append(sourcePredicates, cityperelation.ParentIDEQ(in.CiTypeId))
		
		// 如果只查询有效的关系定义，可以在这里添加额外的过滤条件
		if in.ActiveOnly != nil && *in.ActiveOnly {
			// 这里可以添加状态过滤，比如constraint不为空等
			sourcePredicates = append(sourcePredicates, cityperelation.ConstraintNEQ(""))
		}

		sourceRelations, err := l.svcCtx.DB.CiTypeRelation.Query().
			Where(sourcePredicates...).
			WithParent().   // 预加载父类型信息
			WithChild().    // 预加载子类型信息
			WithRelationType(). // 预加载关系类型信息
			All(l.ctx)

		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// 转换为响应格式
		for _, relation := range sourceRelations {
			relationInfo := l.convertToRelationInfo(relation, in.IncludeAttributeMapping)
			resp.SourceRelations = append(resp.SourceRelations, relationInfo)
		}
		totalCount += uint64(len(sourceRelations))
	}

	// 查询作为目标CI类型的关系定义 (child_id = ci_type_id)
	if direction == "target" || direction == "both" {
		var targetPredicates []predicate.CiTypeRelation
		targetPredicates = append(targetPredicates, cityperelation.ChildIDEQ(in.CiTypeId))
		
		// 如果只查询有效的关系定义
		if in.ActiveOnly != nil && *in.ActiveOnly {
			targetPredicates = append(targetPredicates, cityperelation.ConstraintNEQ(""))
		}

		targetRelations, err := l.svcCtx.DB.CiTypeRelation.Query().
			Where(targetPredicates...).
			WithParent().   // 预加载父类型信息
			WithChild().    // 预加载子类型信息
			WithRelationType(). // 预加载关系类型信息
			All(l.ctx)

		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// 转换为响应格式
		for _, relation := range targetRelations {
			relationInfo := l.convertToRelationInfo(relation, in.IncludeAttributeMapping)
			resp.TargetRelations = append(resp.TargetRelations, relationInfo)
		}
		totalCount += uint64(len(targetRelations))
	}

	resp.TotalCount = totalCount
	return resp, nil
}

// convertToRelationInfo 将ent模型转换为Proto响应格式
func (l *GetCiTypeRelationDefinitionsLogic) convertToRelationInfo(relation *ent.CiTypeRelation, includeAttributeMapping *bool) *cmdb.CiTypeRelationInfo {
	info := &cmdb.CiTypeRelationInfo{
		Id:             &relation.ID,
		CreatedAt:      pointy.GetPointer(relation.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(relation.UpdatedAt.UnixMilli()),
		ParentId:       &relation.ParentID,
		ChildId:        &relation.ChildID,
		RelationTypeId: &relation.RelationTypeID,
		Constraint:     &relation.Constraint,
	}

	// 如果需要包含属性映射信息
	if includeAttributeMapping == nil || *includeAttributeMapping {
		if relation.ParentAttrID != 0 {
			info.ParentAttrId = &relation.ParentAttrID
		}
		if relation.ChildAttrID != 0 {
			info.ChildAttrId = &relation.ChildAttrID
		}
		info.ParentAttrIds = relation.ParentAttrIds
		info.ChildAttrIds = relation.ChildAttrIds
	}

	return info
}
