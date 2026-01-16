package citypeinheritance

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

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
