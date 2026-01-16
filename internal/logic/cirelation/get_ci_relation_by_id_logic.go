package cirelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiRelationByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiRelationByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiRelationByIdLogic {
	return &GetCiRelationByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiRelationByIdLogic) GetCiRelationById(in *cmdb.IDReq) (*cmdb.CiRelationInfo, error) {
	result, err := l.svcCtx.DB.CiRelation.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 构建properties JSON字符串
	var propertiesStr *string
	if result.Properties != nil {
		if raw, ok := result.Properties["raw"].(string); ok {
			propertiesStr = &raw
		}
	}

	return &cmdb.CiRelationInfo{
		Id:              &result.ID,
		CreatedAt:       pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:       pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		SourceCiId:      &result.SourceCiID,
		TargetCiId:      &result.TargetCiID,
		RelationTypeId:  &result.RelationTypeID,
		More:            &result.More,
		DiscoverySource: &result.DiscoverySource,
		AncestorIds:     &result.AncestorIds,
		Properties:      propertiesStr,
	}, nil
}
