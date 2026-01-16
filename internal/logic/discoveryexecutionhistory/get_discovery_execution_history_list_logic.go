package discoveryexecutionhistory

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/discoveryexecutionhistory"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
    "github.com/zeromicro/go-zero/core/logx"
)

type GetDiscoveryExecutionHistoryListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDiscoveryExecutionHistoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiscoveryExecutionHistoryListLogic {
	return &GetDiscoveryExecutionHistoryListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDiscoveryExecutionHistoryListLogic) GetDiscoveryExecutionHistoryList(in *cmdb.DiscoveryExecutionHistoryListReq) (*cmdb.DiscoveryExecutionHistoryListResp, error) {
	var predicates []predicate.DiscoveryExecutionHistory
	if in.CreatedAt != nil {
		predicates = append(predicates, discoveryexecutionhistory.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, discoveryexecutionhistory.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, discoveryexecutionhistory.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.DiscoveryConfigId != nil {
		predicates = append(predicates, discoveryexecutionhistory.DiscoveryConfigIDEQ(*in.DiscoveryConfigId))
	}
	if in.ExecutionId != nil {
		predicates = append(predicates, discoveryexecutionhistory.ExecutionIDContains(*in.ExecutionId))
	}
	if in.TriggerType != nil {
		predicates = append(predicates, discoveryexecutionhistory.TriggerTypeContains(*in.TriggerType))
	}
	if in.TriggeredBy != nil {
		predicates = append(predicates, discoveryexecutionhistory.TriggeredByEQ(*in.TriggeredBy))
	}
	if in.StartedAt != nil {
		predicates = append(predicates, discoveryexecutionhistory.StartedAtGTE(time.UnixMilli(*in.StartedAt)))
	}
	if in.CompletedAt != nil {
		predicates = append(predicates, discoveryexecutionhistory.CompletedAtGTE(time.UnixMilli(*in.CompletedAt)))
	}
	if in.DurationSeconds != nil {
		predicates = append(predicates, discoveryexecutionhistory.DurationSecondsEQ(int(*in.DurationSeconds)))
	}
	if in.ExecStatus != nil {
		predicates = append(predicates, discoveryexecutionhistory.ExecStatusContains(*in.ExecStatus))
	}
	if in.Stage != nil {
		predicates = append(predicates, discoveryexecutionhistory.StageContains(*in.Stage))
	}
	if in.Progress != nil {
		predicates = append(predicates, discoveryexecutionhistory.ProgressEQ(int(*in.Progress)))
	}
	if in.TotalRecords != nil {
		predicates = append(predicates, discoveryexecutionhistory.TotalRecordsEQ(*in.TotalRecords))
	}
	if in.ProcessedRecords != nil {
		predicates = append(predicates, discoveryexecutionhistory.ProcessedRecordsEQ(*in.ProcessedRecords))
	}
	if in.SuccessRecords != nil {
		predicates = append(predicates, discoveryexecutionhistory.SuccessRecordsEQ(*in.SuccessRecords))
	}
	if in.FailedRecords != nil {
		predicates = append(predicates, discoveryexecutionhistory.FailedRecordsEQ(*in.FailedRecords))
	}
	if in.SkippedRecords != nil {
		predicates = append(predicates, discoveryexecutionhistory.SkippedRecordsEQ(*in.SkippedRecords))
	}
	if in.CreatedCis != nil {
		predicates = append(predicates, discoveryexecutionhistory.CreatedCisEQ(*in.CreatedCis))
	}
	if in.UpdatedCis != nil {
		predicates = append(predicates, discoveryexecutionhistory.UpdatedCisEQ(*in.UpdatedCis))
	}
	if in.ErrorMessage != nil {
		predicates = append(predicates, discoveryexecutionhistory.ErrorMessageContains(*in.ErrorMessage))
	}
	// Skip JSON field queries as they're complex
	// if in.ErrorDetails != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.ErrorDetailsEQ(*in.ErrorDetails))
	// }
	// if in.ValidationErrors != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.ValidationErrorsEQ(*in.ValidationErrors))
	// }
	// if in.ExecutionResult != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.ExecutionResultEQ(*in.ExecutionResult))
	// }
	// if in.PerformanceMetrics != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.PerformanceMetricsEQ(*in.PerformanceMetrics))
	// }
	// if in.ConfigSnapshot != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.ConfigSnapshotEQ(*in.ConfigSnapshot))
	// }
	// if in.ProviderInfo != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.ProviderInfoEQ(*in.ProviderInfo))
	// }
	if in.ExecutionLog != nil {
		predicates = append(predicates, discoveryexecutionhistory.ExecutionLogContains(*in.ExecutionLog))
	}
	// if in.Metadata != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.MetadataEQ(*in.Metadata))
	// }
	// if in.Tags != nil {
	//     predicates = append(predicates, discoveryexecutionhistory.TagsEQ(*in.Tags))
	// }
	result, err := l.svcCtx.DB.DiscoveryExecutionHistory.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.DiscoveryExecutionHistoryListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.DiscoveryExecutionHistoryInfo{
			Id:          &v.ID,
			CreatedAt:   pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:   pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			DiscoveryConfigId:	&v.DiscoveryConfigID,
			ExecutionId:	&v.ExecutionID,
			TriggerType:	&v.TriggerType,
			TriggeredBy:	&v.TriggeredBy,
			StartedAt:	pointy.GetPointer(v.StartedAt.UnixMilli()),
			CompletedAt:	pointy.GetUnixMilliPointer(v.CompletedAt.UnixMilli()),
			DurationSeconds:	pointy.GetPointer(int64(v.DurationSeconds)),
			ExecStatus:	&v.ExecStatus,
			Stage:	&v.Stage,
			Progress:	pointy.GetPointer(int64(v.Progress)),
			TotalRecords:	&v.TotalRecords,
			ProcessedRecords:	&v.ProcessedRecords,
			SuccessRecords:	&v.SuccessRecords,
			FailedRecords:	&v.FailedRecords,
			SkippedRecords:	&v.SkippedRecords,
			CreatedCis:	&v.CreatedCis,
			UpdatedCis:	&v.UpdatedCis,
			ErrorMessage:	&v.ErrorMessage,
			ErrorDetails:	jsonhelper.JSONToString(v.ErrorDetails),
			ValidationErrors:	jsonhelper.JSONToString(v.ValidationErrors),
			ExecutionResult:	jsonhelper.JSONToString(v.ExecutionResult),
			PerformanceMetrics:	jsonhelper.JSONToString(v.PerformanceMetrics),
			ConfigSnapshot:	jsonhelper.JSONToString(v.ConfigSnapshot),
			ProviderInfo:	jsonhelper.JSONToString(v.ProviderInfo),
			ExecutionLog:	&v.ExecutionLog,
			Metadata:	jsonhelper.JSONToString(v.Metadata),
			Tags:	jsonhelper.JSONToString(v.Tags),
		})
	}

	return resp, nil
}
