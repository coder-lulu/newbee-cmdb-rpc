package citypegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeGroupItemLogic {
	return &UpdateCiTypeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeGroupItemLogic) UpdateCiTypeGroupItem(in *cmdb.CiTypeGroupItemInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.CiTypeGroupItem.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilGroupID(in.GroupId).
		SetNotNilTypeID(in.TypeId).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
