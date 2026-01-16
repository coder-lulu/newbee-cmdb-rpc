package cirelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiRelationLogic {
	return &UpdateCiRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiRelationLogic) UpdateCiRelation(in *cmdb.CiRelationInfo) (*cmdb.BaseResp, error) {
	updater := l.svcCtx.DB.CiRelation.UpdateOneID(*in.Id)

	// 更新必要字段
	if in.SourceCiId != nil {
		updater.SetSourceCiID(*in.SourceCiId)
	}
	if in.TargetCiId != nil {
		updater.SetTargetCiID(*in.TargetCiId)
	}
	if in.RelationTypeId != nil {
		updater.SetRelationTypeID(*in.RelationTypeId)
	}

	// 更新可选字段
	if in.More != nil {
		updater.SetMore(*in.More)
	}
	if in.DiscoverySource != nil {
		updater.SetDiscoverySource(*in.DiscoverySource)
	}
	if in.AncestorIds != nil {
		updater.SetAncestorIds(*in.AncestorIds)
	}
	if in.Properties != nil {
		// 将properties JSON字符串解析为map
		updater.SetProperties(map[string]interface{}{
			"raw": *in.Properties,
		})
	}

	err := updater.Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
