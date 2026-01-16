package citypegroupitem

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

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

	// 构建查询条件
	query := l.svcCtx.DB.CiTypeGroupItem.Query().WithCiType() // 预加载CI类型

	// 如果指定了GroupId，则过滤分组
	if in.GroupId != nil {
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.EQ("group_id", *in.GroupId))
		})
	}

	// 获取所有分组项和类型
	allItems, err := query.All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 按分组ID组织数据，同时应用过滤条件
	itemsByGroup := make(map[uint64][]*cmdb.CiTypeGroupItemInfo)
	for _, item := range allItems {
		ciType := item.Edges.CiType // 使用预加载的关联数据
		if item.Edges.CiType == nil {
			// 有可能删除了
			continue
		}

		// 应用过滤条件
		if !l.shouldIncludeItem(item, ciType, in) {
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

// shouldIncludeItem 判断是否应该包含该项
func (l *GetCiTypeGroupItemTreeListLogic) shouldIncludeItem(item *ent.CiTypeGroupItem, ciType *ent.CiType, in *cmdb.CiTypeGroupItemListReq) bool {
	// 检查包含的CI类型ID列表
	if len(in.IncludeTypeIds) > 0 {
		found := false
		for _, id := range in.IncludeTypeIds {
			if id == item.TypeID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查排除的CI类型ID列表
	if len(in.ExcludeTypeIds) > 0 {
		for _, id := range in.ExcludeTypeIds {
			if id == item.TypeID {
				return false
			}
		}
	}

	// 检查包含的CI类型名称列表
	if len(in.IncludeTypeNames) > 0 {
		found := false
		for _, name := range in.IncludeTypeNames {
			if name == ciType.Name || name == ciType.Alias {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查排除的CI类型名称列表
	if len(in.ExcludeTypeNames) > 0 {
		for _, name := range in.ExcludeTypeNames {
			if name == ciType.Name || name == ciType.Alias {
				return false
			}
		}
	}

	return true
}
