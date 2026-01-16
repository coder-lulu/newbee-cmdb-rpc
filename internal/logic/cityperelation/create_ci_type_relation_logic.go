package cityperelation

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cityperelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"
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
	// 检查是否已存在相同的父子关系组合
	if in.ParentId != nil && in.ChildId != nil && in.RelationTypeId != nil {
		exists, err := l.svcCtx.DB.CiTypeRelation.Query().
			Where(
				cityperelation.ParentIDEQ(*in.ParentId),
				cityperelation.ChildIDEQ(*in.ChildId),
				cityperelation.RelationTypeIDEQ(*in.RelationTypeId),
			).
			Exist(l.ctx)
		
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		
		if exists {
			return nil, errors.New("该父子模型关系已存在，不能重复创建")
		}
	}

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
