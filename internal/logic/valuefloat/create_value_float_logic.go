package valuefloat

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateValueFloatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateValueFloatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateValueFloatLogic {
	return &CreateValueFloatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateValueFloatLogic) CreateValueFloat(in *cmdb.ValueFloatInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.ValueFloat.Create().
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(in.Value).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
