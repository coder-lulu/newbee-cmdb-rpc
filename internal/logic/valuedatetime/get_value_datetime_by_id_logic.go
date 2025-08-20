package valuedatetime

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueDatetimeByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueDatetimeByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueDatetimeByIdLogic {
	return &GetValueDatetimeByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueDatetimeByIdLogic) GetValueDatetimeById(in *cmdb.IDReq) (*cmdb.ValueDatetimeInfo, error) {
	result, err := l.svcCtx.DB.ValueDatetime.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.ValueDatetimeInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		CiId:      &result.CiID,
		AttrId:    &result.AttrID,
		Value:     pointy.GetPointer(result.Value.UnixMilli()),
	}, nil
}
