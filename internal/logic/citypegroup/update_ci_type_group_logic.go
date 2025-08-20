package citypegroup

import (
	"context"
	"errors"

	"gitee.com/link234/cmdb-rpc/ent/citypegroup"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeGroupLogic {
	return &UpdateCiTypeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeGroupLogic) UpdateCiTypeGroup(in *cmdb.CiTypeGroupInfo) (*cmdb.BaseResp, error) {
	// 检查是否存在同名的组
	exist, err := l.svcCtx.DB.CiTypeGroup.Query().Where(citypegroup.Name(*in.Name), citypegroup.IDNEQ(*in.Id)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if exist {
		return nil, errors.New("分组名称已存在")
	}
	// 如果组名是 其他 或者 基类 则不允许修改
	if *in.Name == "其他" || *in.Name == "基类" {
		return nil, errors.New("分组名称不能为其它或基类")
	}
	group, err := l.svcCtx.DB.CiTypeGroup.Query().Where(citypegroup.ID(*in.Id)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if group.Name == "其他" || group.Name == "基类" {
		return nil, errors.New("分组名称不能为其它或基类")
	}
	err = l.svcCtx.DB.CiTypeGroup.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
