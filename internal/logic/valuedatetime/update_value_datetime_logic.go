package valuedatetime

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateValueDatetimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateValueDatetimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateValueDatetimeLogic {
	return &UpdateValueDatetimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateValueDatetimeLogic) UpdateValueDatetime(in *cmdb.ValueDatetimeInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.ValueDatetime.UpdateOneID(*in.Id).
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(pointy.GetTimeMilliPointer(in.Value)).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
