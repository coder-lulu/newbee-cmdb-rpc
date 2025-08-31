package cipermission

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiPermissionByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiPermissionByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiPermissionByIdLogic {
	return &GetCiPermissionByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiPermissionByIdLogic) GetCiPermissionById(in *cmdb.IDReq) (*cmdb.CiPermissionInfo, error) {
	// 暂时返回简化的响应，待权限系统完善后启用完整功能
	id := in.Id
	permissionId := "temp_disabled"
	subjectName := "权限功能暂时禁用"
	status := uint32(1)
	return &cmdb.CiPermissionInfo{
		Id:           &id,
		PermissionId: &permissionId,
		SubjectName:  &subjectName,
		Status:       &status,
	}, nil
}