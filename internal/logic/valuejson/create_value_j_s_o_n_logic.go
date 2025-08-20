package valuejson

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	jsonx "gitee.com/link234/newbee-backend-common/utils/json"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateValueJSONLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateValueJSONLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateValueJSONLogic {
	return &CreateValueJSONLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateValueJSONLogic) CreateValueJSON(in *cmdb.ValueJSONInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.ValueJSON.Create().
		SetNotNilCiID(in.CiId).
		SetNotNilAttrID(in.AttrId).
		SetNotNilValue(jsonx.StringToJSONRawMessage(*in.Value)).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
