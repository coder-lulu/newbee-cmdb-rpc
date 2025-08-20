package cityperelation

import (
	"context"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/cityperelation"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeRelationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeRelationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeRelationListLogic {
	return &GetCiTypeRelationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeRelationListLogic) GetCiTypeRelationList(in *cmdb.CiTypeRelationListReq) (*cmdb.CiTypeRelationListResp, error) {
	var predicates []predicate.CiTypeRelation
	if in.CreatedAt != nil {
		predicates = append(predicates, cityperelation.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, cityperelation.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, cityperelation.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.ParentId != nil {
		predicates = append(predicates, cityperelation.ParentIDEQ(*in.ParentId))
	}
	if in.ChildId != nil {
		predicates = append(predicates, cityperelation.ChildIDEQ(*in.ChildId))
	}
	if in.RelationTypeId != nil {
		predicates = append(predicates, cityperelation.RelationTypeIDEQ(*in.RelationTypeId))
	}
	if in.Constraint != nil {
		predicates = append(predicates, cityperelation.ConstraintContains(*in.Constraint))
	}
	if in.ParentAttrId != nil {
		predicates = append(predicates, cityperelation.ParentAttrIDEQ(*in.ParentAttrId))
	}
	if in.ChildAttrId != nil {
		predicates = append(predicates, cityperelation.ChildAttrIDEQ(*in.ChildAttrId))
	}
	// 对于数组字段的查询条件，这里暂时不处理，因为需要更复杂的JSON查询逻辑

	result, err := l.svcCtx.DB.CiTypeRelation.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeRelationListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeRelationInfo{
			Id:             &v.ID,
			CreatedAt:      pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:      pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			ParentId:       &v.ParentID,
			ChildId:        &v.ChildID,
			RelationTypeId: &v.RelationTypeID,
			Constraint:     &v.Constraint,
			ParentAttrId:   &v.ParentAttrID,
			ChildAttrId:    &v.ChildAttrID,
			ParentAttrIds:  v.ParentAttrIds,
			ChildAttrIds:   v.ChildAttrIds,
		})
	}

	return resp, nil
}
