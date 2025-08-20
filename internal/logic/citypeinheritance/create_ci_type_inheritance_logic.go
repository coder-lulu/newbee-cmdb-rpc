package citypeinheritance

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

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
