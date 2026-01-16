package citypeattributegroup

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiTypeAttributeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiTypeAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiTypeAttributeGroupLogic {
	return &DeleteCiTypeAttributeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiTypeAttributeGroupLogic) DeleteCiTypeAttributeGroup(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if len(in.Ids) == 0 {
		return nil, errors.New("ids is required")
	}
	for _, id := range in.Ids {
		ciTypeAttributeGroup, err := tx.CiTypeAttributeGroup.Query().Where(citypeattributegroup.ID(id)).First(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		if ciTypeAttributeGroup == nil {
			return nil, errors.New("ci type attribute group not found")
		}
		if ciTypeAttributeGroup.Name == "其他" {
			return nil, errors.New("default ci type attribute group cannot be deleted")
		}
		ciTypeAttributeGroupItems, err := tx.CiTypeAttributeGroupItem.Query().Where(citypeattributegroupitem.GroupID(id)).Count(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		if ciTypeAttributeGroupItems > 0 {
			return nil, errors.New("ci type attribute group has items, cannot be deleted")
		}
		_, err = tx.CiTypeAttributeGroup.Delete().Where(citypeattributegroup.ID(id)).Exec(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{Msg: errormsg.DeleteSuccess}, nil
}
