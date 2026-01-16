package citypediscoveryconfig

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteDiscoveryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteDiscoveryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteDiscoveryLogic {
	return &ExecuteDiscoveryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteDiscovery 执行CI发现任务
func (l *ExecuteDiscoveryLogic) ExecuteDiscovery(in *cmdb.IDReq) (*cmdb.DiscoveryExecutionResp, error) {
	// 验证配置是否存在
	config, err := l.svcCtx.DB.CiTypeDiscoveryConfig.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 检查配置状态
	if !config.Enabled {
		return nil, fmt.Errorf("discovery config %d is disabled", in.Id)
	}

	if config.ConfigStatus != "active" {
		return nil, fmt.Errorf("discovery config %d is not active, current status: %s", in.Id, config.ConfigStatus)
	}

	// 获取发现服务实例
	discoveryService := discovery.NewService(l.svcCtx.DB)

	// 执行发现任务
	result, err := discoveryService.ExecuteDiscovery(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("Failed to execute discovery for config %d: %v", in.Id, err)
		return nil, fmt.Errorf("failed to execute discovery: %w", err)
	}

	// 构建响应
	response := &cmdb.DiscoveryExecutionResp{
		ExecutionId:  result.ExecutionID,
		ConfigId:     result.ConfigID,
		Status:       string(result.Status),
		StartTime:    pointy.GetPointer(result.StartTime.UnixMilli()),
		TotalRecords: pointy.GetPointer(result.TotalRecords),
		Message:      pointy.GetPointer("Discovery task started successfully"),
	}

	l.Logger.Infof("Discovery execution started for config %d with execution ID: %s", in.Id, result.ExecutionID)

	return response, nil
}