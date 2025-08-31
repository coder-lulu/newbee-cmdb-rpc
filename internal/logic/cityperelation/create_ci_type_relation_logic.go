package cityperelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeRelationLogic {
	return &CreateCiTypeRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeRelationLogic) CreateCiTypeRelation(in *cmdb.CiTypeRelationInfo) (*cmdb.BaseIDResp, error) {
	creator := l.svcCtx.DB.CiTypeRelation.Create().
		SetNotNilParentID(in.ParentId).
		SetNotNilChildID(in.ChildId).
		SetNotNilRelationTypeID(in.RelationTypeId).
		SetNotNilConstraint(in.Constraint).
		SetNotNilParentAttrID(in.ParentAttrId).
		SetNotNilChildAttrID(in.ChildAttrId)

	// 处理parent_attr_ids数组
	if len(in.ParentAttrIds) > 0 {
		creator.SetParentAttrIds(in.ParentAttrIds)
	}

	// 处理child_attr_ids数组
	if len(in.ChildAttrIds) > 0 {
		creator.SetChildAttrIds(in.ChildAttrIds)
	}

	result, err := creator.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
