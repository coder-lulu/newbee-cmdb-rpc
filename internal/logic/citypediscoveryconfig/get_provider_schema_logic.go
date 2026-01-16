package citypediscoveryconfig

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProviderSchemaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProviderSchemaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderSchemaLogic {
	return &GetProviderSchemaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProviderSchemaLogic) GetProviderSchema(in *cmdb.TestProviderConnectionReq) (*cmdb.ProviderSchemaResp, error) {
	// 获取发现服务实例
	discoveryService := discovery.NewService(l.svcCtx.DB)

	// 获取提供者配置结构
	schema, err := discoveryService.GetProviderSchema(in.ProviderType, in.ProviderId)
	if err != nil {
		l.Logger.Errorf("Failed to get provider schema: %v", err)
		return &cmdb.ProviderSchemaResp{
			ProviderType: in.ProviderType,
			ProviderId:   in.ProviderId,
			SchemaJson:   "{}",
		}, nil
	}

	// 转换为JSON字符串
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		l.Logger.Errorf("Failed to marshal schema: %v", err)
		return &cmdb.ProviderSchemaResp{
			ProviderType: in.ProviderType,
			ProviderId:   in.ProviderId,
			SchemaJson:   "{}",
		}, nil
	}

	l.Logger.Infof("Retrieved schema for provider %s:%s", in.ProviderType, in.ProviderId)

	return &cmdb.ProviderSchemaResp{
		ProviderType: in.ProviderType,
		ProviderId:   in.ProviderId,
		SchemaJson:   string(schemaBytes),
	}, nil
}
