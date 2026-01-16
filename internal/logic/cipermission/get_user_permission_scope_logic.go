package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserPermissionScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionScopeLogic {
	return &GetUserPermissionScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserPermissionScopeLogic) GetUserPermissionScope(in *cmdb.UserPermissionScopeReq) (*cmdb.UserPermissionScopeResp, error) {
	// TODO: Implement actual scope retrieval
	return &cmdb.UserPermissionScopeResp{
		AllowedCiTypes: []uint64{},
	}, nil
}
