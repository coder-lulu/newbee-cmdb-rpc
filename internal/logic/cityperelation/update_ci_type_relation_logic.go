package cityperelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeRelationLogic {
	return &UpdateCiTypeRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeRelationLogic) UpdateCiTypeRelation(in *cmdb.CiTypeRelationInfo) (*cmdb.BaseResp, error) {
	updater := l.svcCtx.DB.CiTypeRelation.UpdateOneID(*in.Id).
		SetNotNilParentID(in.ParentId).
		SetNotNilChildID(in.ChildId).
		SetNotNilRelationTypeID(in.RelationTypeId).
		SetNotNilConstraint(in.Constraint).
		SetNotNilParentAttrID(in.ParentAttrId).
		SetNotNilChildAttrID(in.ChildAttrId)

	// 处理parent_attr_ids数组
	if len(in.ParentAttrIds) > 0 {
		updater.SetParentAttrIds(in.ParentAttrIds)
	}

	// 处理child_attr_ids数组
	if len(in.ChildAttrIds) > 0 {
		updater.SetChildAttrIds(in.ChildAttrIds)
	}

	err := updater.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
