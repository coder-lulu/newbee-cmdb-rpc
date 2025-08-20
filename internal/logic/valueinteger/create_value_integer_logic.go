package valueinteger

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateValueIntegerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateValueIntegerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateValueIntegerLogic {
	return &CreateValueIntegerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateValueIntegerLogic) CreateValueInteger(in *cmdb.ValueIntegerInfo) (*cmdb.BaseIDResp, error) {
	query := l.svcCtx.DB.ValueInteger.Create().
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId)

	if in.Value != nil {
		query.SetNotNilValue(pointy.GetPointer(int(*in.Value)))
	}

	result, err := query.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
