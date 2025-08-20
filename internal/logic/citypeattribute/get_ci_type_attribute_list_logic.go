package citypeattribute

import (
	"context"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/internal/logic/attribute"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeListLogic {
	return &GetCiTypeAttributeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeListLogic) GetCiTypeAttributeList(in *cmdb.CiTypeAttributeListReq) (*cmdb.CiTypeAttributeListResp, error) {
	var predicates []predicate.CiTypeAttribute
	if in.CreatedAt != nil {
		predicates = append(predicates, citypeattribute.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, citypeattribute.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, citypeattribute.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.Sort != nil {
		predicates = append(predicates, citypeattribute.SortEQ(*in.Sort))
	}
	if in.TypeId != nil {
		predicates = append(predicates, citypeattribute.TypeIDEQ(*in.TypeId))
	}
	if in.AttrId != nil {
		predicates = append(predicates, citypeattribute.AttrIDEQ(*in.AttrId))
	}
	if in.IsRequired != nil {
		predicates = append(predicates, citypeattribute.IsRequiredEQ(*in.IsRequired))
	}
	if in.ListShow != nil {
		predicates = append(predicates, citypeattribute.ListShowEQ(*in.ListShow))
	}
	result, err := l.svcCtx.DB.CiTypeAttribute.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeAttributeListResp{}
	resp.Total = result.PageDetails.Total
	for _, v := range result.List {
		attr, err := attribute.NewGetAttributeByIdLogic(l.ctx, l.svcCtx).GetAttributeById(&cmdb.IDReq{Id: v.AttrID})
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		resp.Data = append(resp.Data, &cmdb.CiTypeAttributeItem{
			Id:                &v.ID,
			TypeId:            v.TypeID,
			IsRequired:        &v.IsRequired,
			ListShow:          &v.ListShow,
			IsEdit:            &v.IsEdit,
			DetailShow:        &v.DetailShow,
			IsUnique:          &v.IsUnique,
			Sort:              &v.Sort,
			CiTypeAttributeId: &v.ID,
			Attribute:         attr,
		})
	}

	return resp, nil
}
