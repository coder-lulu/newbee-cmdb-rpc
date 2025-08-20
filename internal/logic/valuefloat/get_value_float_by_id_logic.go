package valuefloat

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueFloatByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueFloatByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueFloatByIdLogic {
	return &GetValueFloatByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueFloatByIdLogic) GetValueFloatById(in *cmdb.IDReq) (*cmdb.ValueFloatInfo, error) {
	result, err := l.svcCtx.DB.ValueFloat.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.ValueFloatInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		CiId:      &result.CiID,
		AttrId:    &result.AttrID,
		Value:     &result.Value,
	}, nil
}
