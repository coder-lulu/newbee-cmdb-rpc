package citypegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeGroupItemTreeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeGroupItemTreeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeGroupItemTreeListLogic {
	return &GetCiTypeGroupItemTreeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeGroupItemTreeListLogic) GetCiTypeGroupItemTreeList(in *cmdb.CiTypeGroupItemListReq) (*cmdb.CiTypeGroupItemTreeListResp, error) {
	// 获取所有分组
	groups, err := l.svcCtx.DB.CiTypeGroup.Query().All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	// 获取所有分组项和类型（一次性查询）
	allItems, err := l.svcCtx.DB.CiTypeGroupItem.Query().
		WithCiType(). // 预加载CI类型
		All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	// 按分组ID组织数据
	itemsByGroup := make(map[uint64][]*cmdb.CiTypeGroupItemInfo)
	for _, item := range allItems {
		ciType := item.Edges.CiType // 使用预加载的关联数据
		if item.Edges.CiType == nil {
			// 有可能删除了
			continue
		}
		itemInfo := &cmdb.CiTypeGroupItemInfo{
			Id:      &item.ID,
			Sort:    &item.Sort,
			GroupId: &item.GroupID,
			TypeId:  &item.TypeID,
			Name:    &ciType.Alias,
		}
		itemsByGroup[item.GroupID] = append(itemsByGroup[item.GroupID], itemInfo)
	}
	// 构建最终数据
	data := make([]*cmdb.CiTypeGroupItemTreeListInfo, 0, len(groups))
	for _, group := range groups {
		data = append(data, &cmdb.CiTypeGroupItemTreeListInfo{
			GroupId: &group.ID,
			Name:    &group.Name,
			Sort:    &group.Sort,
			Items:   itemsByGroup[group.ID],
		})
	}

	return &cmdb.CiTypeGroupItemTreeListResp{Data: data}, nil
}
