package citypeattribute

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeAttributeLogic {
	return &UpdateCiTypeAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeAttributeLogic) UpdateCiTypeAttribute(in *cmdb.CiTypeAttributeItem) (*cmdb.BaseResp, error) {
	// 1. 校验ID
	if in.Id == nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, errors.New("ID不能为空"), in)
	}

	// 2. 如果有Attribute，先更新Attribute
	if in.Attribute != nil && in.Attribute.Id != nil {
		_, err := attribute.NewUpdateAttributeLogic(l.ctx, l.svcCtx).UpdateAttribute(in.Attribute)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 3. 更新CiTypeAttribute表
	updateBuilder := l.svcCtx.DB.CiTypeAttribute.UpdateOneID(*in.Id)

	if in.IsRequired != nil {
		updateBuilder = updateBuilder.SetIsRequired(*in.IsRequired)
	}
	if in.ListShow != nil {
		updateBuilder = updateBuilder.SetListShow(*in.ListShow)
	}
	if in.IsUnique != nil {
		updateBuilder = updateBuilder.SetIsUnique(*in.IsUnique)
	}
	if in.IsEdit != nil {
		updateBuilder = updateBuilder.SetIsEdit(*in.IsEdit)
	}

	err := updateBuilder.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
