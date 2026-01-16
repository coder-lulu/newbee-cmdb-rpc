package citypeinheritance

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeInheritanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeInheritanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeInheritanceLogic {
	return &CreateCiTypeInheritanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeInheritanceLogic) CreateCiTypeInheritance(in *cmdb.CiTypeInheritanceInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.CiTypeInheritance.Create().
		SetNotNilParentID(in.ParentId).
		SetNotNilChildID(in.ChildId).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
