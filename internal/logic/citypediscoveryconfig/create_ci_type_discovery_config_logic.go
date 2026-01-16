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

type CreateCiTypeDiscoveryConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeDiscoveryConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeDiscoveryConfigLogic {
	return &CreateCiTypeDiscoveryConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeDiscoveryConfigLogic) CreateCiTypeDiscoveryConfig(in *cmdb.CiTypeDiscoveryConfigInfo) (*cmdb.BaseIDResp, error) {
	query := l.svcCtx.DB.CiTypeDiscoveryConfig.Create().
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
		SetNotNilLastError(in.LastError)

	// Handle numeric fields with type conversion
	if in.BatchSize != nil {
		query.SetNotNilBatchSize(pointy.GetPointer(int(*in.BatchSize)))
	}
	if in.TimeoutSeconds != nil {
		query.SetNotNilTimeoutSeconds(pointy.GetPointer(int(*in.TimeoutSeconds)))
	}

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
	if in.LastExecutedAt != nil {
		query.SetNotNilLastExecutedAt(pointy.GetTimeMilliPointer(in.LastExecutedAt))
	}
	if in.NextExecutionAt != nil {
		query.SetNotNilNextExecutionAt(pointy.GetTimeMilliPointer(in.NextExecutionAt))
	}

	result, err := query.Save(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess }, nil
}