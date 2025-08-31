package svc

import (
	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	_ "github.com/coder-lulu/newbee-cmdb-rpc/ent/runtime"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/config"

	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client
	Redis  redis.UniversalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := ent.NewClient(
		ent.Log(logx.Info), // logger
		ent.Driver(c.DatabaseConf.NewNoCacheDriver()),
		ent.Debug(), // debug mode
	)

	// 注册租户Hook和Interceptor - 确保多租户数据隔离
	db.Use(hooks.TenantMutationHook())
	db.Intercept(hooks.TenantQueryInterceptor())

	// 注册数据权限拦截器 - 支持CMDB核心业务实体的数据权限控制
	db.Intercept(hooks.GetEnhancedDataPermissionInterceptor())

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  c.RedisConf.MustNewUniversalRedis(),
	}
}
