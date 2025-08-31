package citypeattributegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeGroupItemByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeGroupItemByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeGroupItemByIdLogic {
	return &GetCiTypeAttributeGroupItemByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeGroupItemByIdLogic) GetCiTypeAttributeGroupItemById(in *cmdb.IDReq) (*cmdb.CiTypeAttributeGroupItemInfo, error) {
	result, err := l.svcCtx.DB.CiTypeAttributeGroupItem.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeAttributeGroupItemInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Sort:      &result.Sort,
		GroupId:   &result.GroupID,
		AttrId:    &result.AttrID,
	}, nil
}
