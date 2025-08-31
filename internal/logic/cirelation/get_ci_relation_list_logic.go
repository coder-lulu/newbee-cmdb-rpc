package cirelation

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiRelationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiRelationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiRelationListLogic {
	return &GetCiRelationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiRelationListLogic) GetCiRelationList(in *cmdb.CiRelationListReq) (*cmdb.CiRelationListResp, error) {
	var predicates []predicate.CiRelation
	if in.CreatedAt != nil {
		predicates = append(predicates, cirelation.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, cirelation.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, cirelation.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.FirstCiId != nil {
		predicates = append(predicates, cirelation.FirstCiIDEQ(*in.FirstCiId))
	}
	if in.SecondCiId != nil {
		predicates = append(predicates, cirelation.SecondCiIDEQ(*in.SecondCiId))
	}
	if in.RelationTypeId != nil {
		predicates = append(predicates, cirelation.RelationTypeIDEQ(*in.RelationTypeId))
	}
	if in.More != nil {
		predicates = append(predicates, cirelation.MoreEQ(*in.More))
	}
	if in.Source != nil {
		predicates = append(predicates, cirelation.SourceContains(*in.Source))
	}
	if in.AncestorIds != nil {
		predicates = append(predicates, cirelation.AncestorIdsContains(*in.AncestorIds))
	}
	result, err := l.svcCtx.DB.CiRelation.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiRelationListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiRelationInfo{
			Id:             &v.ID,
			CreatedAt:      pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:      pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			FirstCiId:      &v.FirstCiID,
			SecondCiId:     &v.SecondCiID,
			RelationTypeId: &v.RelationTypeID,
			More:           &v.More,
			Source:         &v.Source,
			AncestorIds:    &v.AncestorIds,
		})
	}

	return resp, nil
}
