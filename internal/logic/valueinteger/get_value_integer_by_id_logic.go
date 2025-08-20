package valueinteger

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueIntegerByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueIntegerByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueIntegerByIdLogic {
	return &GetValueIntegerByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueIntegerByIdLogic) GetValueIntegerById(in *cmdb.IDReq) (*cmdb.ValueIntegerInfo, error) {
	result, err := l.svcCtx.DB.ValueInteger.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.ValueIntegerInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		CiId:      &result.CiID,
		AttrId:    &result.AttrID,
		Value:     pointy.GetPointer(int64(result.Value)),
	}, nil
}
