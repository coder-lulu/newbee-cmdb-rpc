package citypegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeGroupItemLogic {
	return &CreateCiTypeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeGroupItemLogic) CreateCiTypeGroupItem(in *cmdb.CiTypeGroupItemInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.CiTypeGroupItem.Create().
		SetNotNilSort(in.Sort).
		SetNotNilGroupID(in.GroupId).
		SetNotNilTypeID(in.TypeId).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
