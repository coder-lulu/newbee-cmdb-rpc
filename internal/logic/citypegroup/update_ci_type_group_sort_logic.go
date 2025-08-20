package citypegroup

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/citypegroup"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeGroupSortLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeGroupSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeGroupSortLogic {
	return &UpdateCiTypeGroupSortLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeGroupSortLogic) UpdateCiTypeGroupSort(in *cmdb.CiTypeGroupSortReq) (*cmdb.BaseResp, error) {
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
	// 更新排序
	for _, sort := range in.Data {
		err := l.svcCtx.DB.CiTypeGroup.UpdateOneID(sort.GroupId).SetSort(sort.Sort).Exec(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}
	// 设置基类和其他的排序永远为999 和1000
	err = tx.CiTypeGroup.Update().Where(citypegroup.Name("基类")).SetSort(999).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	err = tx.CiTypeGroup.Update().Where(citypegroup.Name("其他")).SetSort(1000).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)

	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{}, nil
}
