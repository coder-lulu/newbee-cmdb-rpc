package attributemappingrule

import (
	"context"

    "github.com/coder-lulu/newbee-cmdb-rpc/ent/attributemappingrule"
    "github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
    "github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
    "github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"
    "github.com/zeromicro/go-zero/core/logx"
)

type DeleteAttributeMappingRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAttributeMappingRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAttributeMappingRuleLogic {
	return &DeleteAttributeMappingRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAttributeMappingRuleLogic) DeleteAttributeMappingRule(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	_, err := l.svcCtx.DB.AttributeMappingRule.Delete().Where(attributemappingrule.IDIn(in.Ids...)).Exec(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess }, nil
}
