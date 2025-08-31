package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiPermissionLogic {
	return &DeleteCiPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiPermissionLogic) DeleteCiPermission(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	// 暂时禁用复杂权限功能，返回简化响应
	return &cmdb.BaseResp{Msg: "权限功能暂时禁用"}, nil
}