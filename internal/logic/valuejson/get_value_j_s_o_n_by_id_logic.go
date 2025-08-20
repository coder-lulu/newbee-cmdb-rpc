package valuejson

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	jsonx "gitee.com/link234/newbee-backend-common/utils/json"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueJSONByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueJSONByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueJSONByIdLogic {
	return &GetValueJSONByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueJSONByIdLogic) GetValueJSONById(in *cmdb.IDReq) (*cmdb.ValueJSONInfo, error) {
	result, err := l.svcCtx.DB.ValueJSON.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.ValueJSONInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		CiId:      &result.CiID,
		AttrId:    &result.AttrID,
		Value:     pointy.GetPointer(jsonx.JSONRawMessageToString(&result.Value)),
	}, nil
}
