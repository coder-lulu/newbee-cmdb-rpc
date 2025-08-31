package citypeattributegroupitem

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SortCiTypeAttributeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSortCiTypeAttributeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SortCiTypeAttributeGroupItemLogic {
	return &SortCiTypeAttributeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SortCiTypeAttributeGroupItemLogic) SortCiTypeAttributeGroupItem(in *cmdb.CiTypeAttributeGroupItemSortReq) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		l.Logger.Errorf("开启事务失败: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	processedCount := 0
	skippedCount := 0

	for _, item := range in.Data {
		// 首先验证分组是否存在并获取其对应的typeId
		group, err := tx.CiTypeAttributeGroup.Get(l.ctx, item.GroupId)
		if err != nil {
			l.Logger.Errorf("查询属性分组失败: GroupId=%d, error=%v", item.GroupId, err)
			continue
		}

		for _, item2 := range item.Items {
			// 查询属性分组项以获取attributeId
			groupItem, err := tx.CiTypeAttributeGroupItem.Get(l.ctx, item2.Id)
			if err != nil {
				l.Logger.Errorf("查询属性分组项失败: ItemId=%d, error=%v", item2.Id, err)
				skippedCount++
				continue
			}

			// 验证该属性是否属于这个模型
			exists, err := tx.CiTypeAttribute.Query().
				Where(
					citypeattribute.TypeIDEQ(group.TypeID),
					citypeattribute.AttrIDEQ(groupItem.AttrID),
				).
				Exist(l.ctx)
			if err != nil {
				l.Logger.Errorf("验证属性归属失败: TypeId=%d, AttrId=%d, error=%v",
					group.TypeID, groupItem.AttrID, err)
				skippedCount++
				continue
			}

			if !exists {
				l.Logger.Errorf("属性不属于指定模型，跳过排序: TypeId=%d, AttrId=%d, ItemId=%d",
					group.TypeID, groupItem.AttrID, item2.Id)
				skippedCount++
				continue
			}

			// 验证通过，执行排序操作
			err = tx.CiTypeAttributeGroupItem.UpdateOneID(item2.Id).
				SetSort(item2.Sort).
				SetGroupID(item.GroupId).
				Exec(l.ctx)
			if err != nil {
				l.Logger.Errorf("更新排序失败: ItemId=%d, error=%v", item2.Id, err)
				return nil, err
			}

			processedCount++
			l.Logger.Debugf("成功更新排序: TypeId=%d, AttrId=%d, ItemId=%d, Sort=%d",
				group.TypeID, groupItem.AttrID, item2.Id, item2.Sort)
		}
	}

	err = tx.Commit()
	if err != nil {
		l.Logger.Errorf("提交事务失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("属性排序完成: 成功处理=%d, 跳过=%d", processedCount, skippedCount)
	return &cmdb.BaseResp{}, nil
}
