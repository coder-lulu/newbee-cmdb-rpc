package async_task

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/logic/async"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelTaskLogic {
	return &CancelTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelTaskLogic) CancelTask(in *cmdb.TaskCancelReq) (*cmdb.BaseResp, error) {
	// 创建异步任务管理logic实例
	asyncLogic := async.NewAsyncTaskLogic(l.ctx, l.svcCtx)

	// 调用实际的任务取消逻辑
	return asyncLogic.CancelTask(in)
}
