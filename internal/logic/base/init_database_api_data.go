package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
	"go.openly.dev/pointy"
)

// insertCmdbApiData 插入CMDB API数据到Core服务API表
func (l *InitDatabaseLogic) insertCmdbApiData(ctx context.Context) error {
	if l.svcCtx.CoreRpc == nil {
		logx.Info("Core RPC client is not configured, skipping API insertion")
		return nil
	}

	// 为Core RPC调用创建包含租户信息的上下文
	// 使用默认租户ID 1，这是系统初始化时使用的标准租户
	// 使用hooks.SetTenantIDToContext确保gRPC metadata也被正确设置
	tenantCtx := hooks.SetTenantIDToContext(ctx, uint64(1))

	// 获取已存在的CMDB API列表，用于增量插入检查
	var existingCmdbApis map[string]bool
	apiListResp, err := l.svcCtx.CoreRpc.GetApiList(tenantCtx, &core.ApiListReq{
		Page:     1,
		PageSize: 1000, // 获取足够多的API
	})
	if err != nil {
		logx.Errorw("Failed to check existing CMDB APIs, proceeding with full insertion", logx.Field("error", err.Error()))
		existingCmdbApis = make(map[string]bool) // 如果无法获取，则全部插入
	} else {
		// 构建已存在API的映射表，用于快速查询
		existingCmdbApis = make(map[string]bool)
		cmdbApiCount := 0
		if apiListResp.Data != nil {
			for _, api := range apiListResp.Data {
				if api.ServiceName != nil && *api.ServiceName == "Cmdb" {
					// 使用Path+Method作为唯一键
					key := *api.Path + "|" + *api.Method
					existingCmdbApis[key] = true
					cmdbApiCount++
				}
			}
		}
		logx.Infof("Found %d existing CMDB APIs, will perform incremental insertion", cmdbApiCount)
	}

	// 定义CMDB API数据，基于之前分析的完整API清单
	cmdbApis := []*core.ApiInfo{
		// 属性管理API
		{
			Path:        pointy.String("/attribute/create"),
			Description: pointy.String("创建属性"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/attribute/update"),
			Description: pointy.String("更新属性"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/attribute/delete"),
			Description: pointy.String("删除属性"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/attribute/list"),
			Description: pointy.String("获取属性列表"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/attribute/simple-list"),
			Description: pointy.String("获取简要属性列表"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/attribute"),
			Description: pointy.String("通过ID获取属性"),
			ApiGroup:    pointy.String("attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},

		// 配置项管理API
		{
			Path:        pointy.String("/cis/create"),
			Description: pointy.String("创建配置项"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/cis/update"),
			Description: pointy.String("更新配置项"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/cis/delete"),
			Description: pointy.String("删除配置项"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/cis/list"),
			Description: pointy.String("获取配置项列表"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/cis"),
			Description: pointy.String("通过ID获取配置项"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/cis/detail"),
			Description: pointy.String("通过ID获取配置项详情"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/cis/batch"),
			Description: pointy.String("批量配置项操作"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/cis/validate"),
			Description: pointy.String("验证配置项属性"),
			ApiGroup:    pointy.String("cis"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// 配置项类型管理API
		{
			Path:        pointy.String("/ci_type/create"),
			Description: pointy.String("创建配置项类型"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type/update"),
			Description: pointy.String("更新配置项类型"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type/delete"),
			Description: pointy.String("删除配置项类型"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type/list"),
			Description: pointy.String("获取配置项类型列表"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type/list/by_group_id"),
			Description: pointy.String("按组获取配置项类型列表"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type"),
			Description: pointy.String("通过ID获取配置项类型"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type/append-attribute"),
			Description: pointy.String("为配置项类型添加属性"),
			ApiGroup:    pointy.String("ci_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// CI类型分组管理API
		{
			Path:        pointy.String("/ci_type_group/create"),
			Description: pointy.String("创建CI类型分组"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group/update"),
			Description: pointy.String("更新CI类型分组"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group/delete"),
			Description: pointy.String("删除CI类型分组"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group/list"),
			Description: pointy.String("获取CI类型分组列表"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_group"),
			Description: pointy.String("通过ID获取CI类型分组"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_group/sort"),
			Description: pointy.String("CI类型分组排序"),
			ApiGroup:    pointy.String("ci_type_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// CI类型分组项管理API
		{
			Path:        pointy.String("/ci_type_group_item/create"),
			Description: pointy.String("创建CI类型分组项"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group_item/update"),
			Description: pointy.String("更新CI类型分组项"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group_item/delete"),
			Description: pointy.String("删除CI类型分组项"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_group_item/list"),
			Description: pointy.String("获取CI类型分组项列表"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_group_item/tree"),
			Description: pointy.String("获取CI类型分组项树形结构"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_group_item"),
			Description: pointy.String("通过ID获取CI类型分组项"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_group_item/sort"),
			Description: pointy.String("CI类型分组项排序"),
			ApiGroup:    pointy.String("ci_type_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// 配置项关系管理API
		{
			Path:        pointy.String("/ci_relation/create"),
			Description: pointy.String("创建配置项关系"),
			ApiGroup:    pointy.String("ci_relation"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_relation/update"),
			Description: pointy.String("更新配置项关系"),
			ApiGroup:    pointy.String("ci_relation"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_relation/delete"),
			Description: pointy.String("删除配置项关系"),
			ApiGroup:    pointy.String("ci_relation"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_relation/list"),
			Description: pointy.String("获取配置项关系列表"),
			ApiGroup:    pointy.String("ci_relation"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_relation"),
			Description: pointy.String("通过ID获取配置项关系"),
			ApiGroup:    pointy.String("ci_relation"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},

		// 配置项权限管理API
		{
			Path:        pointy.String("/ci_permission/create"),
			Description: pointy.String("创建配置项权限"),
			ApiGroup:    pointy.String("ci_permission"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_permission/update"),
			Description: pointy.String("更新配置项权限"),
			ApiGroup:    pointy.String("ci_permission"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_permission/delete"),
			Description: pointy.String("删除配置项权限"),
			ApiGroup:    pointy.String("ci_permission"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_permission/list"),
			Description: pointy.String("获取配置项权限列表"),
			ApiGroup:    pointy.String("ci_permission"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_permission"),
			Description: pointy.String("通过ID获取配置项权限"),
			ApiGroup:    pointy.String("ci_permission"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},

		// 值类型管理API
		{
			Path:        pointy.String("/value_text/create"),
			Description: pointy.String("创建文本值"),
			ApiGroup:    pointy.String("value_text"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/value_integer/create"),
			Description: pointy.String("创建整数值"),
			ApiGroup:    pointy.String("value_integer"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/value_float/create"),
			Description: pointy.String("创建浮点值"),
			ApiGroup:    pointy.String("value_float"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/value_datetime/create"),
			Description: pointy.String("创建日期时间值"),
			ApiGroup:    pointy.String("value_datetime"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/value_json/create"),
			Description: pointy.String("创建JSON值"),
			ApiGroup:    pointy.String("value_json"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// 选项管理API
		{
			Path:        pointy.String("/choice_text/create"),
			Description: pointy.String("创建文本选项"),
			ApiGroup:    pointy.String("choice_text"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/choice_integer/create"),
			Description: pointy.String("创建整数选项"),
			ApiGroup:    pointy.String("choice_integer"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/choice_float/create"),
			Description: pointy.String("创建浮点选项"),
			ApiGroup:    pointy.String("choice_float"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},

		// 关系类型管理API
		{
			Path:        pointy.String("/relation_type/create"),
			Description: pointy.String("创建关系类型"),
			ApiGroup:    pointy.String("relation_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/relation_type/update"),
			Description: pointy.String("更新关系类型"),
			ApiGroup:    pointy.String("relation_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/relation_type/delete"),
			Description: pointy.String("删除关系类型"),
			ApiGroup:    pointy.String("relation_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/relation_type/list"),
			Description: pointy.String("获取关系类型列表"),
			ApiGroup:    pointy.String("relation_type"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		// 属性管理API
		{
			Path:        pointy.String("/ci_type_attribute/create"),
			Description: pointy.String("创建属性"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute/update"),
			Description: pointy.String("更新属性"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute/delete"),
			Description: pointy.String("删除属性"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute/list"),
			Description: pointy.String("获取属性列表"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute"),
			Description: pointy.String("获取属性详情"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(true),
		},
		{
			Path:        pointy.String("/ci_type_attribute/change-list-show"),
			Description: pointy.String("更新属性是否列表展示"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute/list_with_group"),
			Description: pointy.String("按分组展示属性"),
			ApiGroup:    pointy.String("ci_type_attribute"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group/create"),
			Description: pointy.String("创建属性分组"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group/update"),
			Description: pointy.String("更新属性分组"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group/delete"),
			Description: pointy.String("删除属性分组"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group/list"),
			Description: pointy.String("获取属性分组列表"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group"),
			Description: pointy.String("获取属性分组详情"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group/sort"),
			Description: pointy.String("属性分组排序"),
			ApiGroup:    pointy.String("ci_type_attribute_group"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
		{
			Path:        pointy.String("/ci_type_attribute_group_item/sort"),
			Description: pointy.String("属性分组项排序"),
			ApiGroup:    pointy.String("ci_type_attribute_group_item"),
			Method:      pointy.String("POST"),
			ServiceName: pointy.String("Cmdb"),
			IsRequired:  pointy.Bool(false),
		},
	}

	// 逐个检查并插入缺失的API
	successCount := 0
	skippedCount := 0
	totalApis := len(cmdbApis)

	for _, api := range cmdbApis {
		// 构建API的唯一标识符
		apiKey := *api.Path + "|" + *api.Method

		// 检查API是否已存在
		if existingCmdbApis[apiKey] {
			logx.Infof("API '%s %s' already exists, skipping", *api.Method, *api.Path)
			skippedCount++
			continue
		}

		// API不存在，执行插入
		createResp, err := l.svcCtx.CoreRpc.CreateApi(tenantCtx, api)
		if err != nil {
			logx.Errorw("Failed to create CMDB API due to Core service issues",
				logx.Field("api", api.Path),
				logx.Field("method", api.Method),
				logx.Field("error", err.Error()))
			// 继续处理其他API，不中断整个过程
			continue
		}

		if createResp != nil && createResp.Id != 0 {
			logx.Infof("Created CMDB API '%s %s' with ID: %d", *api.Method, *api.Path, createResp.Id)
			successCount++
		}
	}

	// 输出详细的插入统计信息
	logx.Infow("CMDB API insertion completed",
		logx.Field("total_apis", totalApis),
		logx.Field("created_count", successCount),
		logx.Field("skipped_existing", skippedCount),
		logx.Field("failed_count", totalApis-successCount-skippedCount))
	return nil
}