package citypediscoveryconfig

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProvidersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProvidersLogic {
	return &ListProvidersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProvidersLogic) ListProviders(in *cmdb.Empty) (*cmdb.ListProvidersResp, error) {
	// 获取发现服务实例
	discoveryService := discovery.NewService(l.svcCtx.DB)

	// 获取所有可用的提供者
	providersMap := discoveryService.ListProviders()

	var providers []*cmdb.ProviderInfo
	for providerType, providerIDs := range providersMap {
		for _, providerID := range providerIDs {
			providers = append(providers, &cmdb.ProviderInfo{
				ProviderType: providerType,
				ProviderId:   providerID,
				Name:         providerType + ":" + providerID,
				Description:  "Provider for " + providerType + " data sources",
			})
		}
	}

	l.Logger.Infof("Listed %d providers", len(providers))

	return &cmdb.ListProvidersResp{
		Providers: providers,
	}, nil
}
