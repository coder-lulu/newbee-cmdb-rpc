package citypegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/citypegroupitem"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeGroupItemSortLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeGroupItemSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeGroupItemSortLogic {
	return &UpdateCiTypeGroupItemSortLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeGroupItemSortLogic) UpdateCiTypeGroupItemSort(in *cmdb.CiTypeGroupItemSortReq) (*cmdb.BaseResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 延迟提交或回滚事务
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	totalUpdated := 0
	for _, item := range in.SortItems {

		// 获取要更新的记录
		record, err := tx.CiTypeGroupItem.Query().
			Where(citypegroupitem.TypeIDEQ(*item.TypeId)).
			Only(l.ctx)

		if err != nil {
			l.Logger.Errorf("查询记录失败: %v", err)
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// 更新排序
		record, err = record.Update().SetSort(*item.Sort).SetGroupID(*item.GroupId).Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("更新记录失败: %v", err)
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		totalUpdated++
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	l.Logger.Infof("成功更新 %d 条记录", totalUpdated)
	return &cmdb.BaseResp{}, nil

}
