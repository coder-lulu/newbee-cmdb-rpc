package citypeattributegroup

import (
	"context"
	"errors"

	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroup"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeAttributeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeAttributeGroupLogic {
	return &UpdateCiTypeAttributeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeAttributeGroupLogic) UpdateCiTypeAttributeGroup(in *cmdb.CiTypeAttributeGroupInfo) (*cmdb.BaseResp, error) {
	// 检查是否存在同名的组
	exist, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().Where(citypeattributegroup.Name(*in.Name), citypeattributegroup.TypeID(*in.TypeId)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if exist {
		return nil, errors.New("group name already exists")
	}
	// 如果改的分组名字为 其他 则不允许修改
	if *in.Name == "其他" {
		return nil, errors.New("group name cannot be changed to other")
	}
	// 如果改的分组名字为 其他 则不允许修改
	group, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().Where(citypeattributegroup.ID(*in.Id)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if group.Name == "其他" {
		return nil, errors.New("group name cannot be changed to other")
	}
	err = l.svcCtx.DB.CiTypeAttributeGroup.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilTypeID(in.TypeId).
		Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
