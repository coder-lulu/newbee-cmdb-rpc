package citypegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
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
