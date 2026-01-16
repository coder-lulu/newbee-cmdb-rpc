package relationtype

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRelationTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRelationTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRelationTypeLogic {
	return &UpdateRelationTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRelationTypeLogic) UpdateRelationType(in *cmdb.RelationTypeInfo) (*cmdb.BaseResp, error) {
	updater := l.svcCtx.DB.RelationType.UpdateOneID(*in.Id).
		SetNotNilName(in.Name).
		SetNotNilCode(in.Code)

	// 处理category枚举
	if in.Category != nil {
		updater.SetCategory(relationtype.Category(*in.Category))
	}

	// 处理direction枚举
	if in.Direction != nil {
		updater.SetDirection(relationtype.Direction(*in.Direction))
	}

	err := updater.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
