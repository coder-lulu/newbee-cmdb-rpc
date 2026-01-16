package citypeattributegroup

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeGroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeGroupListLogic {
	return &GetCiTypeAttributeGroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeGroupListLogic) GetCiTypeAttributeGroupList(in *cmdb.CiTypeAttributeGroupListReq) (*cmdb.CiTypeAttributeGroupListResp, error) {
	var predicates []predicate.CiTypeAttributeGroup
	if in.CreatedAt != nil {
		predicates = append(predicates, citypeattributegroup.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, citypeattributegroup.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, citypeattributegroup.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.Sort != nil {
		predicates = append(predicates, citypeattributegroup.SortEQ(*in.Sort))
	}
	if in.Name != nil {
		predicates = append(predicates, citypeattributegroup.NameContains(*in.Name))
	}
	if in.TypeId != nil {
		predicates = append(predicates, citypeattributegroup.TypeIDEQ(*in.TypeId))
	}
	result, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeAttributeGroupListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeAttributeGroupInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Sort:      &v.Sort,
			Name:      &v.Name,
			TypeId:    &v.TypeID,
		})
	}

	return resp, nil
}
