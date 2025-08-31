package async_task

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/async"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitAsyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitAsyncTaskLogic {
	return &SubmitAsyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitAsyncTaskLogic) SubmitAsyncTask(in *cmdb.AsyncTaskReq) (*cmdb.AsyncTaskResp, error) {
	// 创建异步任务管理logic实例
	asyncLogic := async.NewAsyncTaskLogic(l.ctx, l.svcCtx)

	// 调用实际的异步任务提交逻辑
	return asyncLogic.SubmitAsyncTask(in)
}
