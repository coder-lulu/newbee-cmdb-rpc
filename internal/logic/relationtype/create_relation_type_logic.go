package relationtype

import (
	"context"

	"gitee.com/link234/cmdb-rpc/ent/relationtype"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

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
