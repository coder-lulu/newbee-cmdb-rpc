package citypegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeGroupItemListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeGroupItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeGroupItemListLogic {
	return &GetCiTypeGroupItemListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeGroupItemListLogic) GetCiTypeGroupItemList(in *cmdb.CiTypeGroupItemListReq) (*cmdb.CiTypeGroupItemListResp, error) {
	var predicates []predicate.CiTypeGroupItem
	if in.GroupId != nil {
		predicates = append(predicates, citypegroupitem.GroupIDEQ(*in.GroupId))
	}

	// 修改逻辑，即使CiTypeGroupItem不存在，也要返回CiTypeGroup

	result, err := l.svcCtx.DB.CiTypeGroupItem.Query().Where(predicates...).All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeGroupItemListResp{}

	for _, v := range result {
		resp.Data = append(resp.Data, &cmdb.CiTypeGroupItemInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Sort:      &v.Sort,
			GroupId:   &v.GroupID,
			TypeId:    &v.TypeID,
		})
	}

	return resp, nil
}
