package svc

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	_ "github.com/coder-lulu/newbee-cmdb-rpc/ent/runtime"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/config"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/discovery/engine"
	initpkg "github.com/coder-lulu/newbee-cmdb-rpc/internal/init"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entenum"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	DB              *ent.Client
	Redis           redis.UniversalClient
	CoreRpc         coreclient.Core         // Core服务RPC客户端
	DiscoveryEngine *engine.DiscoveryEngine // Discovery引擎实例
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := ent.NewClient(
		ent.Log(logx.Info), // logger
		ent.Driver(c.DatabaseConf.NewNoCacheDriver()),
		ent.Debug(), // debug mode
	)

	// 🎯 使用统一Hook系统 - 一键设置租户和部门Hook
	// 配置CMDB服务的租户过滤规则 - 添加CMDB特有的系统表
	hooks.AddExcludedTable("cmdb_asset_types")           // 资产类型表是系统级数据
	hooks.AddExcludedTable("cmdb_templates")             // 模板表是系统级数据
	hooks.AddExcludedTable("cmdb_attribute_definitions") // 属性定义表是系统级数据

	// 一键设置：初始化配置 + 注册所有hooks (租户Hook + 部门Hook)
	if err := hooks.QuickSetup(db); err != nil {
		logx.Errorw("Failed to setup unified hooks", logx.Field("error", err.Error()))
		panic("统一Hook初始化失败: " + err.Error())
	}
	logx.Infow("✅ CMDB service: Unified hooks initialized successfully")

	// TODO: 【后续实施】注册数据权限拦截器 - 符合CLAUDE.md规范3.1节
	// 原架构设计假设"RPC层不承担权限职责"是错误的
	// 实际情况：Unified-IO等服务直接调用CMDB RPC，必须在RPC层执行数据权限控制
	// 否则存在跨部门数据泄露的高危风险
	//
	// 实施步骤：
	// 1. 验证所有业务表是否包含department_id字段
	// 2. 创建数据库索引：CREATE INDEX idx_xxx_tenant_dept ON cmdb_xxx(tenant_id, department_id);
	// 3. 解除下面代码的注释，启用数据权限拦截器
	// 4. 测试验证：
	//    - 同租户不同部门用户隔离
	//    - SystemContext操作不受影响（自动发现、初始化等）
	//    - 跨服务调用（Unified-IO → CMDB RPC）正确过滤数据
	//
	// db.Intercept(hooks.GetDataPermissionInterceptor())
	// logx.Infow("✅ CMDB service: Data permission interceptor registered for all tables with department_id field")

	// 初始化标准关系类型 - 在系统启动时自动创建预设的关系类型
	tenantCtx := hooks.SetTenantIDToContext(context.Background(), entenum.TenantDefaultId)
	if err := initpkg.InitStandardRelationTypes(tenantCtx, db); err != nil {
		logx.Errorf("初始化标准关系类型失败: %v", err)
		// 注意：这里不终止启动，只记录错误，避免影响系统正常启动
	}

	// 初始化Core RPC客户端 - 参考Core服务的实现模式
	var coreRpc coreclient.Core
	if c.CoreRpc.Endpoints != nil && len(c.CoreRpc.Endpoints) > 0 {
		// 创建RPC客户端，使用SystemContext拦截器支持系统级操作
		rpcClient, err := zrpc.NewClient(c.CoreRpc, zrpc.WithUnaryClientInterceptor(hooks.SystemContextClientInterceptor()))
		if err != nil {
			logx.Errorf("Failed to create Core RPC client: %v", err)
		} else {
			coreRpc = coreclient.NewCore(rpcClient)
		}
	}

	// 初始化Discovery引擎
	discoveryEngine := engine.NewDiscoveryEngine(db)
	logx.Infow("✅ CMDB service: Discovery engine initialized successfully")

	return &ServiceContext{
		Config:          c,
		DB:              db,
		Redis:           c.RedisConf.MustNewUniversalRedis(),
		CoreRpc:         coreRpc,
		DiscoveryEngine: discoveryEngine,
	}
}
