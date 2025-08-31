package citype

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeinheritance"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiTypeLogic {
	return &DeleteCiTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiTypeLogic) DeleteCiType(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 查询是否有实例依赖citype
	ciType, err := tx.Cis.Query().Where(cis.TypeIDIn(in.Ids...)).Count(l.ctx)
	if err != nil {
		return nil, err
	}
	if ciType > 0 {
		return nil, errors.New("模型存在资产实例，不允许删除")
	}

	// 查询是否有继承关系
	inheritance, err := tx.CiTypeInheritance.Query().Where(citypeinheritance.ChildIDIn(in.Ids...)).Count(l.ctx)
	if err != nil {
		return nil, err
	}
	if inheritance > 0 {
		return nil, errors.New("模型存在继承关系，不允许删除")
	}
	// 查询所有分组
	groupQuery := tx.CiTypeAttributeGroup.Query().Where(citypeattributegroup.TypeIDIn(in.Ids...))
	groupResult, err := groupQuery.All(l.ctx)
	if err != nil {
		return nil, err
	}
	// 删除分组
	for _, group := range groupResult {
		// 删除分组属性
		_, err := tx.CiTypeAttributeGroupItem.Delete().Where(citypeattributegroupitem.GroupIDEQ(group.ID)).Exec(l.ctx)
		if err != nil {
			return nil, err
		}
		// 删除分组
		_, err = tx.CiTypeAttributeGroup.Delete().Where(citypeattributegroup.IDEQ(group.ID)).Exec(l.ctx)
		if err != nil {
			return nil, err
		}
	}
	// 删除属性分组记录
	_, err = tx.CiTypeAttribute.Delete().Where(citypeattribute.TypeIDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, err
	}

	// 从模型分组中删除
	_, err = tx.CiTypeGroupItem.Delete().Where(citypegroupitem.TypeIDIn(in.Ids...)).Exec(l.ctx)
	// 删除CITYPE
	_, err = tx.CiType.Delete().Where(citype.IDIn(in.Ids...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
