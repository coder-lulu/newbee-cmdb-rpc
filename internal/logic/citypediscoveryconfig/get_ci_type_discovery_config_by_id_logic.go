package citypediscoveryconfig

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeDiscoveryConfigByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeDiscoveryConfigByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeDiscoveryConfigByIdLogic {
	return &GetCiTypeDiscoveryConfigByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeDiscoveryConfigByIdLogic) GetCiTypeDiscoveryConfigById(in *cmdb.IDReq) (*cmdb.CiTypeDiscoveryConfigInfo, error) {
	result, err := l.svcCtx.DB.CiTypeDiscoveryConfig.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeDiscoveryConfigInfo{
		Id:          &result.ID,
		CreatedAt:    pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:    pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:	pointy.GetPointer(uint32(result.Status)),
		DepartmentId:	&result.DepartmentID,
		CreatedBy:	pointy.GetPointer(result.CreatedBy.String()),
		CiTypeId:	&result.CiTypeID,
		ConfigName:	&result.ConfigName,
		Description:	&result.Description,
		DiscoveryMode:	&result.DiscoveryMode,
		DiscoveryType:	&result.DiscoveryType,
		ProviderId:	&result.ProviderID,
		ProviderConfig:	jsonhelper.JSONToString(result.ProviderConfig),
		AttributeMappings:	jsonhelper.JSONToString(result.AttributeMappings),
		DiscoveryRules:	jsonhelper.JSONToString(result.DiscoveryRules),
		FilterConditions:	jsonhelper.JSONToString(result.FilterConditions),
		ExecutionMode:	&result.ExecutionMode,
		ScheduleConfig:	jsonhelper.JSONToString(result.ScheduleConfig),
		Priority:	pointy.GetPointer(int64(result.Priority)),
		BatchSize:	pointy.GetPointer(int64(result.BatchSize)),
		TimeoutSeconds:	pointy.GetPointer(int64(result.TimeoutSeconds)),
		ConflictResolution:	&result.ConflictResolution,
		AutoCreateCi:	&result.AutoCreateCi,
		AutoUpdateAttributes:	&result.AutoUpdateAttributes,
		NotificationConfig:	jsonhelper.JSONToString(result.NotificationConfig),
		Enabled:	&result.Enabled,
		ConfigStatus:	&result.ConfigStatus,
		LastExecutedAt:	pointy.GetUnixMilliPointer(result.LastExecutedAt.UnixMilli()),
		NextExecutionAt:	pointy.GetUnixMilliPointer(result.NextExecutionAt.UnixMilli()),
		ExecutionStats:	jsonhelper.JSONToString(result.ExecutionStats),
		LastError:	&result.LastError,
	}, nil
}

