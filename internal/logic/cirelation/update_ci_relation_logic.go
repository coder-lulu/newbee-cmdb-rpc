package cirelation

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

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
	err := l.svcCtx.DB.CiRelation.UpdateOneID(*in.Id).
		SetNotNilFirstCiID(in.FirstCiId).
		SetNotNilSecondCiID(in.SecondCiId).
		SetNotNilRelationTypeID(in.RelationTypeId).
		SetNotNilMore(in.More).
		SetNotNilSource(in.Source).
		SetNotNilAncestorIds(in.AncestorIds).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
