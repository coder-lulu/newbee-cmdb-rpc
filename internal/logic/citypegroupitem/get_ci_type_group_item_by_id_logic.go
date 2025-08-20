package citypegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeGroupItemByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeGroupItemByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeGroupItemByIdLogic {
	return &GetCiTypeGroupItemByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeGroupItemByIdLogic) GetCiTypeGroupItemById(in *cmdb.IDReq) (*cmdb.CiTypeGroupItemInfo, error) {
	result, err := l.svcCtx.DB.CiTypeGroupItem.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeGroupItemInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Sort:      &result.Sort,
		GroupId:   &result.GroupID,
		TypeId:    &result.TypeID,
	}, nil
}
