package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
	"go.openly.dev/pointy"
)

// insertCmdbDictData 插入CMDB字典数据到Core服务字典表
func (l *InitDatabaseLogic) insertCmdbDictData(ctx context.Context) error {
	if l.svcCtx.CoreRpc == nil {
		logx.Info("Core RPC client is not configured, skipping dictionary insertion")
		return nil
	}

	// 为Core RPC调用创建包含租户信息的上下文
	tenantCtx := hooks.SetTenantIDToContext(ctx, uint64(1))

	// 检查字典是否已经存在，防止重复插入
	existingDicts := make(map[string]bool)
	dictListResp, err := l.svcCtx.CoreRpc.GetDictionaryList(tenantCtx, &core.DictionaryListReq{
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		logx.Errorw("Failed to check existing CMDB dictionaries, but continuing with insertion", logx.Field("error", err.Error()))
	} else {
		cmdbDictCount := 0
		if dictListResp.Data != nil {
			for _, dict := range dictListResp.Data {
				if dict.Name != nil {
					existingDicts[*dict.Name] = true
					// 检查是否为CMDB相关字典
					if containsCmdbKeyword(*dict.Name) {
						cmdbDictCount++
					}
				}
			}
		}
		logx.Infof("Found %d existing CMDB dictionaries, will perform incremental insertion", cmdbDictCount)
	}

	// 定义CMDB相关字典数据
	cmdbDictionaries := []*core.DictionaryInfo{
		// CI状态字典
		{
			Name:     pointy.String("ci_status"),
			Title:    pointy.String("CI状态"),
			Desc:     pointy.String("配置项状态枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// 属性值类型字典
		{
			Name:     pointy.String("attribute_value_type"),
			Title:    pointy.String("属性值类型"),
			Desc:     pointy.String("属性值类型枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// 关系类型字典
		{
			Name:     pointy.String("relation_type_category"),
			Title:    pointy.String("关系类型分类"),
			Desc:     pointy.String("关系类型分类枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// 关系方向字典
		{
			Name:     pointy.String("relation_direction"),
			Title:    pointy.String("关系方向"),
			Desc:     pointy.String("关系方向枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// CI权限级别字典
		{
			Name:     pointy.String("ci_permission_level"),
			Title:    pointy.String("CI权限级别"),
			Desc:     pointy.String("CI权限级别枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// CI权限主体类型字典
		{
			Name:     pointy.String("ci_permission_subject_type"),
			Title:    pointy.String("CI权限主体类型"),
			Desc:     pointy.String("CI权限主体类型枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// CI权限范围类型字典
		{
			Name:     pointy.String("ci_permission_scope_type"),
			Title:    pointy.String("CI权限范围类型"),
			Desc:     pointy.String("CI权限范围类型枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
		// CI风险级别字典
		{
			Name:     pointy.String("ci_risk_level"),
			Title:    pointy.String("CI风险级别"),
			Desc:     pointy.String("CI风险级别枚举"),
			Status:   pointy.Uint32(1),
			TenantId: pointy.Uint64(1),
		},
	}

	// 定义字典详情数据（字典项）
	dictDetails := map[string][]*core.DictionaryDetailInfo{
		"ci_status": {
			{
				Title:    pointy.String("正常"),
				Value:    pointy.String("normal"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("维护中"),
				Value:    pointy.String("maintenance"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("已下线"),
				Value:    pointy.String("offline"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("故障"),
				Value:    pointy.String("fault"),
				Sort:     pointy.Uint32(4),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"attribute_value_type": {
			{
				Title:    pointy.String("文本"),
				Value:    pointy.String("text"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("整数"),
				Value:    pointy.String("int"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("浮点数"),
				Value:    pointy.String("float"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("布尔值"),
				Value:    pointy.String("boolean"),
				Sort:     pointy.Uint32(4),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("日期时间"),
				Value:    pointy.String("datetime"),
				Sort:     pointy.Uint32(5),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("JSON对象"),
				Value:    pointy.String("json"),
				Sort:     pointy.Uint32(6),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"relation_type_category": {
			{
				Title:    pointy.String("逻辑关系"),
				Value:    pointy.String("logical"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("物理关系"),
				Value:    pointy.String("physical"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("业务关系"),
				Value:    pointy.String("business"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"relation_direction": {
			{
				Title:    pointy.String("双向"),
				Value:    pointy.String("bidirectional"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("单向"),
				Value:    pointy.String("unidirectional"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"ci_permission_level": {
			{
				Title:    pointy.String("只读"),
				Value:    pointy.String("read"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("读写"),
				Value:    pointy.String("write"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("管理"),
				Value:    pointy.String("admin"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"ci_permission_subject_type": {
			{
				Title:    pointy.String("用户"),
				Value:    pointy.String("user"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("角色"),
				Value:    pointy.String("role"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("部门"),
				Value:    pointy.String("department"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"ci_permission_scope_type": {
			{
				Title:    pointy.String("全部"),
				Value:    pointy.String("all"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("指定CI类型"),
				Value:    pointy.String("ci_type"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("指定CI实例"),
				Value:    pointy.String("ci_instance"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
		"ci_risk_level": {
			{
				Title:    pointy.String("低"),
				Value:    pointy.String("low"),
				Sort:     pointy.Uint32(1),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("中"),
				Value:    pointy.String("medium"),
				Sort:     pointy.Uint32(2),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("高"),
				Value:    pointy.String("high"),
				Sort:     pointy.Uint32(3),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
			{
				Title:    pointy.String("极高"),
				Value:    pointy.String("critical"),
				Sort:     pointy.Uint32(4),
				Status:   pointy.Uint32(1),
				TenantId: pointy.Uint64(1),
			},
		},
	}

	// 插入字典主表数据
	successCount := 0
	skippedCount := 0
	totalDicts := len(cmdbDictionaries)

	for _, dict := range cmdbDictionaries {
		// 检查字典是否已存在
		if existingDicts[*dict.Name] {
			logx.Infof("Dictionary '%s' already exists, skipping", *dict.Name)
			skippedCount++
			continue
		}

		// 字典不存在，执行插入
		createResp, err := l.svcCtx.CoreRpc.CreateDictionary(tenantCtx, dict)
		if err != nil {
			logx.Errorw("Failed to create CMDB dictionary due to Core service issues",
				logx.Field("dict", dict.Name),
				logx.Field("error", err.Error()))
			continue
		}

		if createResp != nil && createResp.Id != 0 {
			logx.Infof("Created CMDB dictionary '%s' with ID: %d", *dict.Name, createResp.Id)
			successCount++

			// 插入对应的字典项
			if details, exists := dictDetails[*dict.Name]; exists {
				detailSuccessCount := 0
				for _, detail := range details {
					detail.DictionaryId = pointy.Uint64(createResp.Id)
					detailResp, err := l.svcCtx.CoreRpc.CreateDictionaryDetail(tenantCtx, detail)
					if err != nil {
						logx.Errorw("Failed to create dictionary detail",
							logx.Field("dict", dict.Name),
							logx.Field("detail", detail.Title),
							logx.Field("error", err.Error()))
						continue
					}
					if detailResp != nil && detailResp.Id != 0 {
						detailSuccessCount++
					}
				}
				logx.Infof("Created %d dictionary details for '%s'", detailSuccessCount, *dict.Name)
			}
		}
	}

	// 输出统计信息
	logx.Infow("CMDB dictionary insertion completed",
		logx.Field("total_dicts", totalDicts),
		logx.Field("created_count", successCount),
		logx.Field("skipped_existing", skippedCount),
		logx.Field("failed_count", totalDicts-successCount-skippedCount))

	return nil
}

// containsCmdbKeyword 检查字典名称是否包含CMDB关键词
func containsCmdbKeyword(name string) bool {
	cmdbKeywords := []string{"ci_", "cmdb_", "attribute_", "relation_"}
	for _, keyword := range cmdbKeywords {
		if len(name) >= len(keyword) && name[:len(keyword)] == keyword {
			return true
		}
	}
	return false
}