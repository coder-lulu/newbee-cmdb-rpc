package citypegroup

import (
	"context"
	"errors"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeGroupLogic {
	return &CreateCiTypeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeGroupLogic) CreateCiTypeGroup(in *cmdb.CiTypeGroupInfo) (*cmdb.BaseIDResp, error) {
	// 检查是否存在同名的组
	exist, err := l.svcCtx.DB.CiTypeGroup.Query().Where(citypegroup.Name(*in.Name)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if exist {
		return nil, errors.New("分组名称已存在")
	}
	// 如果组名是 其他 或者 基类 则不允许创建
	if *in.Name == "其它" || *in.Name == "基类" {
		return nil, errors.New("分组名称不能为其它或基类")
	}
	result, err := l.svcCtx.DB.CiTypeGroup.Create().
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilDescription(in.Description).
		SetNotNilIcon(in.Icon).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
