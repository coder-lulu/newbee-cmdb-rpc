package async_task

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/async"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskListLogic {
	return &GetTaskListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTaskListLogic) GetTaskList(in *cmdb.TaskListReq) (*cmdb.TaskListResp, error) {
	// 创建异步任务管理logic实例
	asyncLogic := async.NewAsyncTaskLogic(l.ctx, l.svcCtx)

	// 调用实际的任务列表查询逻辑
	return asyncLogic.GetTaskList(in)
}
