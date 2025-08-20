package valuedatetime

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateValueDatetimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateValueDatetimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateValueDatetimeLogic {
	return &CreateValueDatetimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateValueDatetimeLogic) CreateValueDatetime(in *cmdb.ValueDatetimeInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.ValueDatetime.Create().
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(pointy.GetTimeMilliPointer(in.Value)).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
