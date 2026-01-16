package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckCiTypePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckCiTypePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckCiTypePermissionLogic {
	return &CheckCiTypePermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckCiTypePermissionLogic) CheckCiTypePermission(in *cmdb.CiPermissionReq) (*cmdb.CiPermissionResp, error) {
	// TODO: Implement actual permission check
	// For now, return true to unblock the adapter
	return &cmdb.CiPermissionResp{
		HasPermission: true,
	}, nil
}
