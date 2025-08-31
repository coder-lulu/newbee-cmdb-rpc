package relationtype

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRelationTypeByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRelationTypeByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRelationTypeByIdLogic {
	return &GetRelationTypeByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRelationTypeByIdLogic) GetRelationTypeById(in *cmdb.IDReq) (*cmdb.RelationTypeInfo, error) {
	result, err := l.svcCtx.DB.RelationType.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.RelationTypeInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Name:      &result.Name,
		Code:      &result.Code,
		Category:  pointy.GetPointer(string(result.Category)),
		Direction: pointy.GetPointer(string(result.Direction)),
	}, nil
}
