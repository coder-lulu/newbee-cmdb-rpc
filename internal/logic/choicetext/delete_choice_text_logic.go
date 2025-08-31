package choicetext

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/choicetext"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChoiceTextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteChoiceTextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChoiceTextLogic {
	return &DeleteChoiceTextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteChoiceTextLogic) DeleteChoiceText(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.ChoiceText.Delete().Where(choicetext.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
