package choicefloat

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choicefloat"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChoiceFloatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChoiceFloatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChoiceFloatLogic {
	return &DeleteChoiceFloatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteChoiceFloatLogic) DeleteChoiceFloat(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.ChoiceFloat.Delete().Where(choicefloat.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
