package citypegroup

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeGroupByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeGroupByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeGroupByIdLogic {
	return &GetCiTypeGroupByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeGroupByIdLogic) GetCiTypeGroupById(in *cmdb.IDReq) (*cmdb.CiTypeGroupInfo, error) {
	result, err := l.svcCtx.DB.CiTypeGroup.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeGroupInfo{
		Id:          &result.ID,
		CreatedAt:   pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:   pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Sort:        &result.Sort,
		Name:        &result.Name,
		Description: &result.Description,
		Icon:        &result.Icon,
	}, nil
}
