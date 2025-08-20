package cirelation

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiRelationLogic {
	return &CreateCiRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiRelationLogic) CreateCiRelation(in *cmdb.CiRelationInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.CiRelation.Create().
		SetNotNilFirstCiID(in.FirstCiId).
		SetNotNilSecondCiID(in.SecondCiId).
		SetNotNilRelationTypeID(in.RelationTypeId).
		SetNotNilMore(in.More).
		SetNotNilSource(in.Source).
		SetNotNilAncestorIds(in.AncestorIds).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
