package citypeattributegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeAttributeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeAttributeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeAttributeGroupItemLogic {
	return &UpdateCiTypeAttributeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeAttributeGroupItemLogic) UpdateCiTypeAttributeGroupItem(in *cmdb.CiTypeAttributeGroupItemInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.CiTypeAttributeGroupItem.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilGroupID(in.GroupId).
		SetNotNilAttrID(in.AttrId).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
