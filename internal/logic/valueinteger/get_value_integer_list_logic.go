package valueinteger

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valueinteger"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueIntegerListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueIntegerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueIntegerListLogic {
	return &GetValueIntegerListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueIntegerListLogic) GetValueIntegerList(in *cmdb.ValueIntegerListReq) (*cmdb.ValueIntegerListResp, error) {
	var predicates []predicate.ValueInteger
	if in.CreatedAt != nil {
		predicates = append(predicates, valueinteger.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, valueinteger.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, valueinteger.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.CiId != nil {
		predicates = append(predicates, valueinteger.CiIDEQ(*in.CiId))
	}
	if in.AttrId != nil {
		predicates = append(predicates, valueinteger.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, valueinteger.ValueEQ(int(*in.Value)))
	}
	result, err := l.svcCtx.DB.ValueInteger.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ValueIntegerListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.ValueIntegerInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			CiId:      &v.CiID,
			AttrId:    &v.AttrID,
			Value:     pointy.GetPointer(int64(v.Value)),
		})
	}

	return resp, nil
}
