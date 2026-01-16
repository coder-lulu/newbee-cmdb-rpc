package citypeattributegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiTypeAttributeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiTypeAttributeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiTypeAttributeGroupItemLogic {
	return &DeleteCiTypeAttributeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiTypeAttributeGroupItemLogic) DeleteCiTypeAttributeGroupItem(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.CiTypeAttributeGroupItem.Delete().Where(citypeattributegroupitem.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
