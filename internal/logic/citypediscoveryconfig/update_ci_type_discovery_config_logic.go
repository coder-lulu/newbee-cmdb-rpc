package citypediscoveryconfig

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"
	"github.com/suyuan32/simple-admin-common/utils/uuidx"
	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeDiscoveryConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeDiscoveryConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeDiscoveryConfigLogic {
	return &UpdateCiTypeDiscoveryConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeDiscoveryConfigLogic) UpdateCiTypeDiscoveryConfig(in *cmdb.CiTypeDiscoveryConfigInfo) (*cmdb.BaseResp, error) {
	query:= l.svcCtx.DB.CiTypeDiscoveryConfig.UpdateOneID(*in.Id).
			SetNotNilDepartmentID(in.DepartmentId).
			SetNotNilCreatedBy(uuidx.ParseUUIDStringToPointer(in.CreatedBy)).
			SetNotNilCiTypeID(in.CiTypeId).
			SetNotNilConfigName(in.ConfigName).
			SetNotNilDescription(in.Description).
			SetNotNilDiscoveryMode(in.DiscoveryMode).
			SetNotNilDiscoveryType(in.DiscoveryType).
			SetNotNilProviderID(in.ProviderId).
			SetNotNilExecutionMode(in.ExecutionMode).
			SetNotNilConflictResolution(in.ConflictResolution).
			SetNotNilAutoCreateCi(in.AutoCreateCi).
			SetNotNilAutoUpdateAttributes(in.AutoUpdateAttributes).
			SetNotNilEnabled(in.Enabled).
			SetNotNilConfigStatus(in.ConfigStatus).
			SetNotNilLastExecutedAt(pointy.GetTimeMilliPointer(in.LastExecutedAt)).
			SetNotNilNextExecutionAt(pointy.GetTimeMilliPointer(in.NextExecutionAt)).
			SetNotNilLastError(in.LastError)

	// Handle JSON fields
	if in.ProviderConfig != nil {
		providerConfig := jsonhelper.StringToJSONMap(in.ProviderConfig)
		if providerConfig != nil {
			query.SetProviderConfig(providerConfig)
		}
	}
	
	if in.AttributeMappings != nil {
		attributeMappings := jsonhelper.StringToJSONArray(in.AttributeMappings)
		if attributeMappings != nil {
			query.SetAttributeMappings(attributeMappings)
		}
	}
	
	if in.DiscoveryRules != nil {
		discoveryRules := jsonhelper.StringToJSONMap(in.DiscoveryRules)
		if discoveryRules != nil {
			query.SetDiscoveryRules(discoveryRules)
		}
	}
	
	if in.FilterConditions != nil {
		filterConditions := jsonhelper.StringToJSONMap(in.FilterConditions)
		if filterConditions != nil {
			query.SetFilterConditions(filterConditions)
		}
	}
	
	if in.ScheduleConfig != nil {
		scheduleConfig := jsonhelper.StringToJSONMap(in.ScheduleConfig)
		if scheduleConfig != nil {
			query.SetScheduleConfig(scheduleConfig)
		}
	}
	
	if in.NotificationConfig != nil {
		notificationConfig := jsonhelper.StringToJSONMap(in.NotificationConfig)
		if notificationConfig != nil {
			query.SetNotificationConfig(notificationConfig)
		}
	}
	
	if in.ExecutionStats != nil {
		executionStats := jsonhelper.StringToJSONMap(in.ExecutionStats)
		if executionStats != nil {
			query.SetExecutionStats(executionStats)
		}
	}

	if in.Status != nil {
		query.SetNotNilStatus(pointy.GetPointer(uint8(*in.Status)))
	}
	if in.Priority != nil {
		query.SetNotNilPriority(pointy.GetPointer(int(*in.Priority)))
	}
	if in.BatchSize != nil {
		query.SetNotNilBatchSize(pointy.GetPointer(int(*in.BatchSize)))
	}
	if in.TimeoutSeconds != nil {
		query.SetNotNilTimeoutSeconds(pointy.GetPointer(int(*in.TimeoutSeconds)))
	}

	 err := query.Exec(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess }, nil
}
