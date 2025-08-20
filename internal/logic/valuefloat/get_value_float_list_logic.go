package valuefloat

import (
	"context"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/ent/valuefloat"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueFloatListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueFloatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueFloatListLogic {
	return &GetValueFloatListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueFloatListLogic) GetValueFloatList(in *cmdb.ValueFloatListReq) (*cmdb.ValueFloatListResp, error) {
	var predicates []predicate.ValueFloat
	if in.CreatedAt != nil {
		predicates = append(predicates, valuefloat.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, valuefloat.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, valuefloat.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.CiId != nil {
		predicates = append(predicates, valuefloat.CiIDEQ(*in.CiId))
	}
	if in.AttrId != nil {
		predicates = append(predicates, valuefloat.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, valuefloat.ValueEQ(*in.Value))
	}
	result, err := l.svcCtx.DB.ValueFloat.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ValueFloatListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.ValueFloatInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			CiId:      &v.CiID,
			AttrId:    &v.AttrID,
			Value:     &v.Value,
		})
	}

	return resp, nil
}
