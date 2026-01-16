package discoveryexecutionhistory

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDiscoveryExecutionHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDiscoveryExecutionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDiscoveryExecutionHistoryLogic {
	return &CreateDiscoveryExecutionHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDiscoveryExecutionHistoryLogic) CreateDiscoveryExecutionHistory(in *cmdb.DiscoveryExecutionHistoryInfo) (*cmdb.BaseIDResp, error) {
    query := l.svcCtx.DB.DiscoveryExecutionHistory.Create().
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
	if in.ErrorDetails != nil && *in.ErrorDetails != "" {
		var errorDetails map[string]interface{}
		if err := json.Unmarshal([]byte(*in.ErrorDetails), &errorDetails); err == nil {
			query.SetErrorDetails(errorDetails)
		}
	}
	
	if in.ValidationErrors != nil && *in.ValidationErrors != "" {
		var validationErrors []interface{}
		if err := json.Unmarshal([]byte(*in.ValidationErrors), &validationErrors); err == nil {
			query.SetValidationErrors(validationErrors)
		}
	}
	
	if in.ExecutionResult != nil && *in.ExecutionResult != "" {
		var executionResult map[string]interface{}
		if err := json.Unmarshal([]byte(*in.ExecutionResult), &executionResult); err == nil {
			query.SetExecutionResult(executionResult)
		}
	}
	
	if in.PerformanceMetrics != nil && *in.PerformanceMetrics != "" {
		var performanceMetrics map[string]interface{}
		if err := json.Unmarshal([]byte(*in.PerformanceMetrics), &performanceMetrics); err == nil {
			query.SetPerformanceMetrics(performanceMetrics)
		}
	}
	
	if in.ConfigSnapshot != nil && *in.ConfigSnapshot != "" {
		var configSnapshot map[string]interface{}
		if err := json.Unmarshal([]byte(*in.ConfigSnapshot), &configSnapshot); err == nil {
			query.SetConfigSnapshot(configSnapshot)
		}
	}
	
	if in.ProviderInfo != nil && *in.ProviderInfo != "" {
		var providerInfo map[string]interface{}
		if err := json.Unmarshal([]byte(*in.ProviderInfo), &providerInfo); err == nil {
			query.SetProviderInfo(providerInfo)
		}
	}
	
	if in.Metadata != nil && *in.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*in.Metadata), &metadata); err == nil {
			query.SetMetadata(metadata)
		}
	}
	
	if in.Tags != nil && *in.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(*in.Tags), &tags); err == nil {
			query.SetTags(tags)
		}
	}

	if in.DurationSeconds != nil {
		query.SetNotNilDurationSeconds(pointy.GetPointer(int(*in.DurationSeconds)))
	}
	if in.Progress != nil {
		query.SetNotNilProgress(pointy.GetPointer(int(*in.Progress)))
	}

	result, err := query.Save(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess }, nil
}
