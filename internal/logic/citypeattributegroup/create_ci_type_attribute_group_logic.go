package citypeattributegroup

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeAttributeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeAttributeGroupLogic {
	return &CreateCiTypeAttributeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeAttributeGroupLogic) CreateCiTypeAttributeGroup(in *cmdb.CiTypeAttributeGroupInfo) (*cmdb.BaseIDResp, error) {
	// 检查是否存在同名的组
	exist, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().Where(citypeattributegroup.Name(*in.Name), citypeattributegroup.TypeID(*in.TypeId)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if exist {
		return nil, errors.New("group name already exists")
	}
	result, err := l.svcCtx.DB.CiTypeAttributeGroup.Create().
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilTypeID(in.TypeId).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
