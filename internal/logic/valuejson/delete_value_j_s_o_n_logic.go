package valuejson

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/valuejson"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteValueJSONLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteValueJSONLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteValueJSONLogic {
	return &DeleteValueJSONLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteValueJSONLogic) DeleteValueJSON(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.ValueJSON.Delete().Where(valuejson.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
