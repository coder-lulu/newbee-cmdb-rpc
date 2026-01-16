package discoveryexecutionhistory

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiscoveryExecutionHistoryByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDiscoveryExecutionHistoryByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiscoveryExecutionHistoryByIdLogic {
	return &GetDiscoveryExecutionHistoryByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDiscoveryExecutionHistoryByIdLogic) GetDiscoveryExecutionHistoryById(in *cmdb.IDReq) (*cmdb.DiscoveryExecutionHistoryInfo, error) {
	result, err := l.svcCtx.DB.DiscoveryExecutionHistory.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.DiscoveryExecutionHistoryInfo{
		Id:          &result.ID,
		CreatedAt:    pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:    pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		DiscoveryConfigId:	&result.DiscoveryConfigID,
		ExecutionId:	&result.ExecutionID,
		TriggerType:	&result.TriggerType,
		TriggeredBy:	&result.TriggeredBy,
		StartedAt:	pointy.GetPointer(result.StartedAt.UnixMilli()),
		CompletedAt:	pointy.GetUnixMilliPointer(result.CompletedAt.UnixMilli()),
		DurationSeconds:	pointy.GetPointer(int64(result.DurationSeconds)),
		ExecStatus:	&result.ExecStatus,
		Stage:	&result.Stage,
		Progress:	pointy.GetPointer(int64(result.Progress)),
		TotalRecords:	&result.TotalRecords,
		ProcessedRecords:	&result.ProcessedRecords,
		SuccessRecords:	&result.SuccessRecords,
		FailedRecords:	&result.FailedRecords,
		SkippedRecords:	&result.SkippedRecords,
		CreatedCis:	&result.CreatedCis,
		UpdatedCis:	&result.UpdatedCis,
		ErrorMessage:	&result.ErrorMessage,
		ErrorDetails:	jsonhelper.JSONToString(result.ErrorDetails),
		ValidationErrors:	jsonhelper.JSONToString(result.ValidationErrors),
		ExecutionResult:	jsonhelper.JSONToString(result.ExecutionResult),
		PerformanceMetrics:	jsonhelper.JSONToString(result.PerformanceMetrics),
		ConfigSnapshot:	jsonhelper.JSONToString(result.ConfigSnapshot),
		ProviderInfo:	jsonhelper.JSONToString(result.ProviderInfo),
		ExecutionLog:	&result.ExecutionLog,
		Metadata:	jsonhelper.JSONToString(result.Metadata),
		Tags:	jsonhelper.JSONToString(result.Tags),
	}, nil
}

