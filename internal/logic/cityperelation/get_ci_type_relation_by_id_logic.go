package cityperelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeRelationByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeRelationByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeRelationByIdLogic {
	return &GetCiTypeRelationByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeRelationByIdLogic) GetCiTypeRelationById(in *cmdb.IDReq) (*cmdb.CiTypeRelationInfo, error) {
	result, err := l.svcCtx.DB.CiTypeRelation.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeRelationInfo{
		Id:             &result.ID,
		CreatedAt:      pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		ParentId:       &result.ParentID,
		ChildId:        &result.ChildID,
		RelationTypeId: &result.RelationTypeID,
		Constraint:     &result.Constraint,
		ParentAttrId:   &result.ParentAttrID,
		ChildAttrId:    &result.ChildAttrID,
		ParentAttrIds:  result.ParentAttrIds,
		ChildAttrIds:   result.ChildAttrIds,
	}, nil
}
