package citypediscoveryconfig

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypediscoveryconfig"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/utils/uuidx"
	"github.com/suyuan32/simple-admin-common/utils/pointy"
    "github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeDiscoveryConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeDiscoveryConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeDiscoveryConfigListLogic {
	return &GetCiTypeDiscoveryConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeDiscoveryConfigListLogic) GetCiTypeDiscoveryConfigList(in *cmdb.CiTypeDiscoveryConfigListReq) (*cmdb.CiTypeDiscoveryConfigListResp, error) {
	var predicates []predicate.CiTypeDiscoveryConfig
	if in.CreatedAt != nil {
		predicates = append(predicates, citypediscoveryconfig.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, citypediscoveryconfig.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.Status != nil {
		predicates = append(predicates, citypediscoveryconfig.StatusEQ(uint8(*in.Status)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, citypediscoveryconfig.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.DepartmentId != nil {
		predicates = append(predicates, citypediscoveryconfig.DepartmentIDEQ(*in.DepartmentId))
	}
	if in.CreatedBy != nil {
		predicates = append(predicates, citypediscoveryconfig.CreatedByEQ(uuidx.ParseUUIDString(*in.CreatedBy)))
	}
	if in.CiTypeId != nil {
		predicates = append(predicates, citypediscoveryconfig.CiTypeIDEQ(*in.CiTypeId))
	}
	if in.ConfigName != nil {
		predicates = append(predicates, citypediscoveryconfig.ConfigNameContains(*in.ConfigName))
	}
	if in.Description != nil {
		predicates = append(predicates, citypediscoveryconfig.DescriptionContains(*in.Description))
	}
	if in.DiscoveryMode != nil {
		predicates = append(predicates, citypediscoveryconfig.DiscoveryModeContains(*in.DiscoveryMode))
	}
	if in.DiscoveryType != nil {
		predicates = append(predicates, citypediscoveryconfig.DiscoveryTypeContains(*in.DiscoveryType))
	}
	if in.ProviderId != nil {
		predicates = append(predicates, citypediscoveryconfig.ProviderIDContains(*in.ProviderId))
	}
	// Skip JSON field queries as they're complex
	// if in.ProviderConfig != nil {
	//		predicates = append(predicates, citypediscoveryconfig.ProviderConfigEQ(*in.ProviderConfig))
	// }
	// if in.AttributeMappings != nil {
	//		predicates = append(predicates, citypediscoveryconfig.AttributeMappingsEQ(*in.AttributeMappings))
	// }
	// if in.DiscoveryRules != nil {
	//		predicates = append(predicates, citypediscoveryconfig.DiscoveryRulesEQ(*in.DiscoveryRules))
	// }
	// if in.FilterConditions != nil {
	//		predicates = append(predicates, citypediscoveryconfig.FilterConditionsEQ(*in.FilterConditions))
	// }
	if in.ExecutionMode != nil {
		predicates = append(predicates, citypediscoveryconfig.ExecutionModeContains(*in.ExecutionMode))
	}
	// if in.ScheduleConfig != nil {
	//		predicates = append(predicates, citypediscoveryconfig.ScheduleConfigEQ(*in.ScheduleConfig))
	// }
	if in.Priority != nil {
		predicates = append(predicates, citypediscoveryconfig.PriorityEQ(int(*in.Priority)))
	}
	if in.BatchSize != nil {
		predicates = append(predicates, citypediscoveryconfig.BatchSizeEQ(int(*in.BatchSize)))
	}
	if in.TimeoutSeconds != nil {
		predicates = append(predicates, citypediscoveryconfig.TimeoutSecondsEQ(int(*in.TimeoutSeconds)))
	}
	if in.ConflictResolution != nil {
		predicates = append(predicates, citypediscoveryconfig.ConflictResolutionContains(*in.ConflictResolution))
	}
	if in.AutoCreateCi != nil {
		predicates = append(predicates, citypediscoveryconfig.AutoCreateCiEQ(*in.AutoCreateCi))
	}
	if in.AutoUpdateAttributes != nil {
		predicates = append(predicates, citypediscoveryconfig.AutoUpdateAttributesEQ(*in.AutoUpdateAttributes))
	}
	// if in.NotificationConfig != nil {
	//		predicates = append(predicates, citypediscoveryconfig.NotificationConfigEQ(*in.NotificationConfig))
	// }
	if in.Enabled != nil {
		predicates = append(predicates, citypediscoveryconfig.EnabledEQ(*in.Enabled))
	}
	if in.ConfigStatus != nil {
		predicates = append(predicates, citypediscoveryconfig.ConfigStatusContains(*in.ConfigStatus))
	}
	if in.LastExecutedAt != nil {
		predicates = append(predicates, citypediscoveryconfig.LastExecutedAtGTE(time.UnixMilli(*in.LastExecutedAt)))
	}
	if in.NextExecutionAt != nil {
		predicates = append(predicates, citypediscoveryconfig.NextExecutionAtGTE(time.UnixMilli(*in.NextExecutionAt)))
	}
	// if in.ExecutionStats != nil {
	//		predicates = append(predicates, citypediscoveryconfig.ExecutionStatsEQ(*in.ExecutionStats))
	// }
	if in.LastError != nil {
		predicates = append(predicates, citypediscoveryconfig.LastErrorContains(*in.LastError))
	}
	result, err := l.svcCtx.DB.CiTypeDiscoveryConfig.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeDiscoveryConfigListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeDiscoveryConfigInfo{
			Id:          &v.ID,
			CreatedAt:   pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:   pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:	pointy.GetPointer(uint32(v.Status)),
			DepartmentId:	&v.DepartmentID,
			CreatedBy:	pointy.GetPointer(v.CreatedBy.String()),
			CiTypeId:	&v.CiTypeID,
			ConfigName:	&v.ConfigName,
			Description:	&v.Description,
			DiscoveryMode:	&v.DiscoveryMode,
			DiscoveryType:	&v.DiscoveryType,
			ProviderId:	&v.ProviderID,
			ProviderConfig:	jsonhelper.JSONToString(v.ProviderConfig),
			AttributeMappings:	jsonhelper.JSONToString(v.AttributeMappings),
			DiscoveryRules:	jsonhelper.JSONToString(v.DiscoveryRules),
			FilterConditions:	jsonhelper.JSONToString(v.FilterConditions),
			ExecutionMode:	&v.ExecutionMode,
			ScheduleConfig:	jsonhelper.JSONToString(v.ScheduleConfig),
			Priority:	pointy.GetPointer(int64(v.Priority)),
			BatchSize:	pointy.GetPointer(int64(v.BatchSize)),
			TimeoutSeconds:	pointy.GetPointer(int64(v.TimeoutSeconds)),
			ConflictResolution:	&v.ConflictResolution,
			AutoCreateCi:	&v.AutoCreateCi,
			AutoUpdateAttributes:	&v.AutoUpdateAttributes,
			NotificationConfig:	jsonhelper.JSONToString(v.NotificationConfig),
			Enabled:	&v.Enabled,
			ConfigStatus:	&v.ConfigStatus,
			LastExecutedAt:	pointy.GetUnixMilliPointer(v.LastExecutedAt.UnixMilli()),
			NextExecutionAt:	pointy.GetUnixMilliPointer(v.NextExecutionAt.UnixMilli()),
			ExecutionStats:	jsonhelper.JSONToString(v.ExecutionStats),
			LastError:	&v.LastError,
		})
	}

	return resp, nil
}
