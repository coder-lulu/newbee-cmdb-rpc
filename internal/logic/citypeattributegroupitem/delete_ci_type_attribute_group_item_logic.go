package citypeattributegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroupitem"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
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
