package citypediscoveryconfig

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestProviderConnectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestProviderConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestProviderConnectionLogic {
	return &TestProviderConnectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TestProviderConnectionLogic) TestProviderConnection(in *cmdb.TestProviderConnectionReq) (*cmdb.BaseResp, error) {
	// 获取发现服务实例
	discoveryService := discovery.NewService(l.svcCtx.DB)

	// 解析提供者配置
	config := make(map[string]interface{})
	if in.ProviderConfig != "" {
		if err := json.Unmarshal([]byte(in.ProviderConfig), &config); err != nil {
			l.Logger.Errorf("Failed to parse provider config: %v", err)
			return &cmdb.BaseResp{
				Msg: "Failed to parse provider config: " + err.Error(),
			}, nil
		}
	}

	// 测试连接
	err := discoveryService.TestConnection(l.ctx, in.ProviderType, in.ProviderId, config)
	if err != nil {
		l.Logger.Errorf("Provider connection test failed: %v", err)
		return &cmdb.BaseResp{
			Msg: "Connection test failed: " + err.Error(),
		}, nil
	}

	l.Logger.Infof("Provider connection test successful for %s:%s", in.ProviderType, in.ProviderId)

	return &cmdb.BaseResp{
		Msg: "Connection test successful",
	}, nil
}
