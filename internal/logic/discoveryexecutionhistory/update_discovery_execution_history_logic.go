package discoveryexecutionhistory

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDiscoveryExecutionHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDiscoveryExecutionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDiscoveryExecutionHistoryLogic {
	return &UpdateDiscoveryExecutionHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDiscoveryExecutionHistoryLogic) UpdateDiscoveryExecutionHistory(in *cmdb.DiscoveryExecutionHistoryInfo) (*cmdb.BaseResp, error) {
	query := l.svcCtx.DB.DiscoveryExecutionHistory.UpdateOneID(*in.Id).
		SetNotNilDiscoveryConfigID(in.DiscoveryConfigId).
		SetNotNilExecutionID(in.ExecutionId).
		SetNotNilTriggerType(in.TriggerType).
		SetNotNilTriggeredBy(in.TriggeredBy).
		SetNotNilStartedAt(pointy.GetTimeMilliPointer(in.StartedAt)).
		SetNotNilCompletedAt(pointy.GetTimeMilliPointer(in.CompletedAt)).
		SetNotNilExecStatus(in.ExecStatus).
		SetNotNilStage(in.Stage).
		SetNotNilTotalRecords(in.TotalRecords).
		SetNotNilProcessedRecords(in.ProcessedRecords).
		SetNotNilSuccessRecords(in.SuccessRecords).
		SetNotNilFailedRecords(in.FailedRecords).
		SetNotNilSkippedRecords(in.SkippedRecords).
		SetNotNilCreatedCis(in.CreatedCis).
		SetNotNilUpdatedCis(in.UpdatedCis).
		SetNotNilErrorMessage(in.ErrorMessage).
		SetNotNilExecutionLog(in.ExecutionLog)

	// Handle JSON fields
	if in.ErrorDetails != nil {
		errorDetails := jsonhelper.StringToJSONMap(in.ErrorDetails)
		if errorDetails != nil {
			query.SetErrorDetails(errorDetails)
		}
	}
	
	if in.ValidationErrors != nil {
		validationErrors := jsonhelper.StringToJSONArray(in.ValidationErrors)
		if validationErrors != nil {
			query.SetValidationErrors(validationErrors)
		}
	}
	
	if in.ExecutionResult != nil {
		executionResult := jsonhelper.StringToJSONMap(in.ExecutionResult)
		if executionResult != nil {
			query.SetExecutionResult(executionResult)
		}
	}
	
	if in.PerformanceMetrics != nil {
		performanceMetrics := jsonhelper.StringToJSONMap(in.PerformanceMetrics)
		if performanceMetrics != nil {
			query.SetPerformanceMetrics(performanceMetrics)
		}
	}
	
	if in.ConfigSnapshot != nil {
		configSnapshot := jsonhelper.StringToJSONMap(in.ConfigSnapshot)
		if configSnapshot != nil {
			query.SetConfigSnapshot(configSnapshot)
		}
	}
	
	if in.ProviderInfo != nil {
		providerInfo := jsonhelper.StringToJSONMap(in.ProviderInfo)
		if providerInfo != nil {
			query.SetProviderInfo(providerInfo)
		}
	}
	
	if in.Metadata != nil {
		metadata := jsonhelper.StringToJSONMap(in.Metadata)
		if metadata != nil {
			query.SetMetadata(metadata)
		}
	}
	
	if in.Tags != nil {
		tags := jsonhelper.StringToStringArray(in.Tags)
		if tags != nil {
			query.SetTags(tags)
		}
	}

	if in.DurationSeconds != nil {
		query.SetNotNilDurationSeconds(pointy.GetPointer(int(*in.DurationSeconds)))
	}
	if in.Progress != nil {
		query.SetNotNilProgress(pointy.GetPointer(int(*in.Progress)))
	}

	err := query.Exec(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess }, nil
}
