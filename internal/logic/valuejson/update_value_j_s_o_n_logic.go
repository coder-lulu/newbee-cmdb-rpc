package valuejson

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
	jsonx "github.com/coder-lulu/newbee-common/v2/utils/json"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateValueJSONLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateValueJSONLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateValueJSONLogic {
	return &UpdateValueJSONLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateValueJSONLogic) UpdateValueJSON(in *cmdb.ValueJSONInfo) (*cmdb.BaseResp, error) {
	err := l.svcCtx.DB.ValueJSON.UpdateOneID(*in.Id).
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(jsonx.StringToJSONRawMessage(*in.Value)).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
