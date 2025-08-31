package citypeattributegroupitem

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeGroupItemListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeGroupItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeGroupItemListLogic {
	return &GetCiTypeAttributeGroupItemListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeGroupItemListLogic) GetCiTypeAttributeGroupItemList(in *cmdb.CiTypeAttributeGroupItemListReq) (*cmdb.CiTypeAttributeGroupItemListResp, error) {
	var predicates []predicate.CiTypeAttributeGroupItem
	if in.CreatedAt != nil {
		predicates = append(predicates, citypeattributegroupitem.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, citypeattributegroupitem.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, citypeattributegroupitem.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.Sort != nil {
		predicates = append(predicates, citypeattributegroupitem.SortEQ(*in.Sort))
	}
	if in.GroupId != nil {
		predicates = append(predicates, citypeattributegroupitem.GroupIDEQ(*in.GroupId))
	}
	if in.AttrId != nil {
		predicates = append(predicates, citypeattributegroupitem.AttrIDEQ(*in.AttrId))
	}
	result, err := l.svcCtx.DB.CiTypeAttributeGroupItem.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeAttributeGroupItemListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeAttributeGroupItemInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Sort:      &v.Sort,
			GroupId:   &v.GroupID,
			AttrId:    &v.AttrID,
		})
	}

	return resp, nil
}
