package valuetext

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateValueTextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateValueTextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateValueTextLogic {
	return &UpdateValueTextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateValueTextLogic) UpdateValueText(in *cmdb.ValueTextInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.ValueText.UpdateOneID(*in.Id).
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(in.Value).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
