package relationtype

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRelationTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRelationTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRelationTypeLogic {
	return &CreateRelationTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRelationTypeLogic) CreateRelationType(in *cmdb.RelationTypeInfo) (*cmdb.BaseIDResp, error) {
	creator := l.svcCtx.DB.RelationType.Create().
		SetNotNilName(in.Name).
		SetNotNilCode(in.Code)

	// 处理category枚举
	if in.Category != nil {
		creator.SetCategory(relationtype.Category(*in.Category))
	}

	// 处理direction枚举
	if in.Direction != nil {
		creator.SetDirection(relationtype.Direction(*in.Direction))
	}

	result, err := creator.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
