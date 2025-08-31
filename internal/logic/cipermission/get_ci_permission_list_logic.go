package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiPermissionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiPermissionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiPermissionListLogic {
	return &GetCiPermissionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiPermissionListLogic) GetCiPermissionList(in *cmdb.CiPermissionListReq) (*cmdb.CiPermissionListResp, error) {
	// 暂时禁用复杂权限功能，返回空列表
	return &cmdb.CiPermissionListResp{
		Total: 0,
		Data:  []*cmdb.CiPermissionInfo{},
	}, nil
}