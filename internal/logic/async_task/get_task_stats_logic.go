package async_task

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/async"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskStatsLogic {
	return &GetTaskStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTaskStatsLogic) GetTaskStats(in *cmdb.TaskStatsReq) (*cmdb.TaskStatsResp, error) {
	// 创建异步任务管理logic实例
	asyncLogic := async.NewAsyncTaskLogic(l.ctx, l.svcCtx)

	// 调用实际的任务统计逻辑
	return asyncLogic.GetTaskStats(in)
}
