package valuedatetime

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuedatetime"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteValueDatetimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteValueDatetimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteValueDatetimeLogic {
	return &DeleteValueDatetimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteValueDatetimeLogic) DeleteValueDatetime(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.ValueDatetime.Delete().Where(valuedatetime.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
