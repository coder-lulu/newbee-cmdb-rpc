package citypeinheritance

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeInheritanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeInheritanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeInheritanceLogic {
	return &UpdateCiTypeInheritanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeInheritanceLogic) UpdateCiTypeInheritance(in *cmdb.CiTypeInheritanceInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.CiTypeInheritance.UpdateOneID(*in.Id).
		SetNotNilParentID(in.ParentId).
		SetNotNilChildID(in.ChildId).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
