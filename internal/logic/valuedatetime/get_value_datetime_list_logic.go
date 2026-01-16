package valuedatetime

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuedatetime"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueDatetimeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueDatetimeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueDatetimeListLogic {
	return &GetValueDatetimeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueDatetimeListLogic) GetValueDatetimeList(in *cmdb.ValueDatetimeListReq) (*cmdb.ValueDatetimeListResp, error) {
	var predicates []predicate.ValueDatetime
	if in.CreatedAt != nil {
		predicates = append(predicates, valuedatetime.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, valuedatetime.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, valuedatetime.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.CiId != nil {
		predicates = append(predicates, valuedatetime.CiIDEQ(*in.CiId))
	}
	if in.AttrId != nil {
		predicates = append(predicates, valuedatetime.AttrIDEQ(*in.AttrId))
	}
	if in.Value != nil {
		predicates = append(predicates, valuedatetime.ValueGTE(time.UnixMilli(*in.Value)))
	}
	result, err := l.svcCtx.DB.ValueDatetime.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.ValueDatetimeListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.ValueDatetimeInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			CiId:      &v.CiID,
			AttrId:    &v.AttrID,
			Value:     pointy.GetPointer(v.Value.UnixMilli()),
		})
	}

	return resp, nil
}
