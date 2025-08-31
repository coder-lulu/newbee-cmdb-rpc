package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiPermissionLogic {
	return &UpdateCiPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiPermissionLogic) UpdateCiPermission(in *cmdb.CiPermissionInfo) (*cmdb.BaseResp, error) {
	// 暂时禁用复杂权限功能，返回简化响应
	return &cmdb.BaseResp{Msg: "权限功能暂时禁用"}, nil
}