package valuetext

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetValueTextByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetValueTextByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetValueTextByIdLogic {
	return &GetValueTextByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetValueTextByIdLogic) GetValueTextById(in *cmdb.IDReq) (*cmdb.ValueTextInfo, error) {
	result, err := l.svcCtx.DB.ValueText.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.ValueTextInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		CiId:      &result.CiID,
		AttrId:    &result.AttrID,
		Value:     &result.Value,
	}, nil
}
