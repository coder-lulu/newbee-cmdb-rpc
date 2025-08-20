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

type UpdateValueIntegerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateValueIntegerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateValueIntegerLogic {
	return &UpdateValueIntegerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateValueIntegerLogic) UpdateValueInteger(in *cmdb.ValueIntegerInfo) (*cmdb.BaseResp, error) {
	query := l.svcCtx.DB.ValueInteger.UpdateOneID(*in.Id).
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId)

	if in.Value != nil {
		query.SetNotNilValue(pointy.GetPointer(int(*in.Value)))
	}

	err := query.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
