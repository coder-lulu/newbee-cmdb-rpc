package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
	"go.openly.dev/pointy"
)

// insertCmdbMenuData 插入CMDB菜单数据到Core服务菜单表
func (l *InitDatabaseLogic) insertCmdbMenuData(ctx context.Context) error {
	if l.svcCtx.CoreRpc == nil {
		logx.Info("Core RPC client is not configured, skipping menu insertion")
		return nil
	}

	// 为Core RPC调用创建包含租户信息的上下文
	// 使用默认租户ID 1，这是系统初始化时使用的标准租户
	// 使用hooks.SetTenantIDToContext确保gRPC metadata也被正确设置
	tenantCtx := hooks.SetTenantIDToContext(ctx, uint64(1))

	// 检查菜单是否已经存在，防止重复插入
	// 构建已存在菜单的映射表，用于增量插入检查
	existingMenus := make(map[uint64]bool)
	menuListResp, err := l.svcCtx.CoreRpc.GetMenuList(tenantCtx, &core.PageInfoReq{
		Page:     1,
		PageSize: 1000, // 获取足够多的菜单
	})
	if err != nil {
		logx.Errorw("Failed to check existing CMDB menus, but continuing with insertion", logx.Field("error", err.Error()))
		// 如果Core服务暂时不可用，我们仍然继续尝试菜单插入
		// 这样当Core服务恢复后，菜单就已经准备好了
		logx.Info("Attempting to insert CMDB menus despite Core service connection issues")
	} else {
		cmdbMenuCount := 0
		if menuListResp.Data != nil {
			for _, menu := range menuListResp.Data {
				if menu.ServiceName != nil && *menu.ServiceName == "Cmdb" {
					if menu.Id != nil {
						existingMenus[*menu.Id] = true
					}
					cmdbMenuCount++
				}
			}
		}
		logx.Infof("Found %d existing CMDB menus, will perform incremental insertion", cmdbMenuCount)
	}

	// 定义CMDB菜单数据，参考core服务的菜单结构
	cmdbMenus := []*core.MenuInfo{
		// 配置管理根菜单 ID 2000
		{
			Id:          pointy.Uint64(2000),
			Level:       pointy.Uint32(1),
			MenuType:    pointy.Uint32(0),                      // 目录类型
			ParentId:    pointy.Uint64(common.DefaultParentId), // 根菜单标准父ID
			Path:        pointy.String("/cmdb"),
			Name:        pointy.String("配置管理"),
			Component:   pointy.String("Layout"),
			Sort:        pointy.Uint32(800),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1), // 设置租户ID
			Meta: &core.Meta{
				Title:    pointy.String("配置管理"),
				Icon:     pointy.String("mdi:server-network"),
				HideMenu: pointy.Bool(false),
			},
		},
	}

	// 使用预定义的根菜单ID
	rootMenuId := uint64(2000)

	// 先插入根菜单，获取实际的根菜单ID
	var actualRootMenuId uint64
	rootMenu := cmdbMenus[0]
	
	if existingMenus[*rootMenu.Id] {
		logx.Infof("CMDB root menu (ID: %d) already exists, skipping", *rootMenu.Id)
		actualRootMenuId = *rootMenu.Id
	} else {
		createResp, err := l.svcCtx.CoreRpc.CreateMenu(tenantCtx, rootMenu)
		if err != nil {
			logx.Errorw("Failed to create CMDB root menu due to Core service issues",
				logx.Field("menu", rootMenu.Name),
				logx.Field("expected_id", *rootMenu.Id),
				logx.Field("error", err.Error()))
			logx.Info("CMDB initialization will continue without menu insertion. Menus can be added later when Core service is available.")
			return nil // 不返回错误，允许其他初始化继续
		}

		if createResp != nil && createResp.Id != 0 {
			actualRootMenuId = createResp.Id
			if createResp.Id == *rootMenu.Id {
				logx.Infof("Created CMDB root menu with expected ID: %d", createResp.Id)
			} else {
				logx.Infof("Created CMDB root menu with auto-generated ID: %d (expected: %d)", createResp.Id, *rootMenu.Id)
			}
			// 更新存在映射
			existingMenus[createResp.Id] = true
		} else {
			logx.Error("Failed to get root menu ID from create response")
			return nil
		}
	}

	// 定义子菜单，使用实际的根菜单ID
	logx.Infof("Using actual root menu ID: %d for sub-menus", actualRootMenuId)
	subMenus := []*core.MenuInfo{
		// 资产实例 ID 2001
		{
			Id:          pointy.Uint64(2001),
			Level:       pointy.Uint32(2),
			MenuType:    pointy.Uint32(1),    // 菜单类型
			ParentId:    pointy.Uint64(actualRootMenuId), // 使用实际的根菜单ID
			Path:        pointy.String("cis"),
			Name:        pointy.String("资产实例"),
			Component:   pointy.String("cmdb/cis/index"),
			Sort:        pointy.Uint32(1),
			Permission:  pointy.String("cmdb:cis:list"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1), // 设置租户ID
			Meta: &core.Meta{
				Title:    pointy.String("资产实例"),
				Icon:     pointy.String("material-symbols-light:azm-outline"),
				HideMenu: pointy.Bool(false),
			},
		},
		// 资产分类 ID 2002
		{
			Id:          pointy.Uint64(2002),
			Level:       pointy.Uint32(2),
			MenuType:    pointy.Uint32(1),    // 菜单类型
			ParentId:    pointy.Uint64(actualRootMenuId), // 使用实际的根菜单ID
			Path:        pointy.String("types"),
			Name:        pointy.String("资产分类"),
			Component:   pointy.String("cmdb/ci_types/index"),
			Sort:        pointy.Uint32(9),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1), // 设置租户ID
			Meta: &core.Meta{
				Title:    pointy.String("资产分类"),
				Icon:     pointy.String("material-symbols-light:group-work-outline"),
				HideMenu: pointy.Bool(false),
			},
		},
		// CI权限管理 ID 2003
		{
			Id:          pointy.Uint64(2003),
			Level:       pointy.Uint32(2),
			MenuType:    pointy.Uint32(1),    // 菜单类型
			ParentId:    pointy.Uint64(actualRootMenuId), // 使用实际的根菜单ID
			Path:        pointy.String("permissions"),
			Name:        pointy.String("CI权限管理"),
			Component:   pointy.String("cmdb/ci_permission/index"),
			Sort:        pointy.Uint32(8),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1), // 设置租户ID
			Meta: &core.Meta{
				Title:    pointy.String("CI权限管理"),
				Icon:     pointy.String("material-symbols-light:security-outline"),
				HideMenu: pointy.Bool(false),
			},
		},
		// CIS资产实例权限按钮 - 父ID为资产实例(2001)
		{
			Id:          pointy.Uint64(2004),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2001), // 资产实例菜单ID
			Path:        pointy.String("/cmdb/cis/list"),
			Name:        pointy.String("查询资产实例"),
			Sort:        pointy.Uint32(1),
			Permission:  pointy.String("cmdb:cis:list"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("查询资产实例"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2005),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2001), // 资产实例菜单ID
			Path:        pointy.String("/cmdb/cis/create"),
			Name:        pointy.String("创建资产实例"),
			Sort:        pointy.Uint32(2),
			Permission:  pointy.String("cmdb:cis:create"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("创建资产实例"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2006),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2001), // 资产实例菜单ID
			Path:        pointy.String("/cmdb/cis/update"),
			Name:        pointy.String("更新资产实例"),
			Sort:        pointy.Uint32(3),
			Permission:  pointy.String("cmdb:cis:update"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("更新资产实例"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2007),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2001), // 资产实例菜单ID
			Path:        pointy.String("/cmdb/cis/delete"),
			Name:        pointy.String("删除资产实例"),
			Sort:        pointy.Uint32(4),
			Permission:  pointy.String("cmdb:cis:delete"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("删除资产实例"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2008),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2001), // 资产实例菜单ID
			Path:        pointy.String("/cmdb/cis/detail"),
			Name:        pointy.String("资产实例详情"),
			Sort:        pointy.Uint32(5),
			Permission:  pointy.String("cmdb:cis:detail"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("资产实例详情"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		// CI权限管理按钮 - 父ID为CI权限管理(2003)
		{
			Id:          pointy.Uint64(2015),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2003), // CI权限管理菜单ID
			Path:        pointy.String("/cmdb/ci_permission/list"),
			Name:        pointy.String("查询CI权限"),
			Sort:        pointy.Uint32(1),
			Permission:  pointy.String("cmdb:ci_permission:list"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("查询CI权限"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2016),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2003), // CI权限管理菜单ID
			Path:        pointy.String("/cmdb/ci_permission/create"),
			Name:        pointy.String("创建CI权限"),
			Sort:        pointy.Uint32(2),
			Permission:  pointy.String("cmdb:ci_permission:create"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("创建CI权限"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2017),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2003), // CI权限管理菜单ID
			Path:        pointy.String("/cmdb/ci_permission/update"),
			Name:        pointy.String("更新CI权限"),
			Sort:        pointy.Uint32(3),
			Permission:  pointy.String("cmdb:ci_permission:update"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("更新CI权限"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2018),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2003), // CI权限管理菜单ID
			Path:        pointy.String("/cmdb/ci_permission/delete"),
			Name:        pointy.String("删除CI权限"),
			Sort:        pointy.Uint32(4),
			Permission:  pointy.String("cmdb:ci_permission:delete"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("删除CI权限"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2019),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2003), // CI权限管理菜单ID
			Path:        pointy.String("/cmdb/ci_permission/detail"),
			Name:        pointy.String("CI权限详情"),
			Sort:        pointy.Uint32(5),
			Permission:  pointy.String("cmdb:ci_permission:detail"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("CI权限详情"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		// CI类型管理权限按钮 - 父ID为资产分类(2002)
		{
			Id:          pointy.Uint64(2010),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2002), // 资产分类菜单ID
			Path:        pointy.String("/cmdb/ci_type/list"),
			Name:        pointy.String("查询CI类型"),
			Sort:        pointy.Uint32(1),
			Permission:  pointy.String("cmdb:ci_type:list"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("查询CI类型"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2011),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2002), // 资产分类菜单ID
			Path:        pointy.String("/cmdb/ci_type/create"),
			Name:        pointy.String("创建CI类型"),
			Sort:        pointy.Uint32(2),
			Permission:  pointy.String("cmdb:ci_type:create"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("创建CI类型"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2012),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2002), // 资产分类菜单ID
			Path:        pointy.String("/cmdb/ci_type/update"),
			Name:        pointy.String("更新CI类型"),
			Sort:        pointy.Uint32(3),
			Permission:  pointy.String("cmdb:ci_type:update"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("更新CI类型"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2013),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2002), // 资产分类菜单ID
			Path:        pointy.String("/cmdb/ci_type/delete"),
			Name:        pointy.String("删除CI类型"),
			Sort:        pointy.Uint32(4),
			Permission:  pointy.String("cmdb:ci_type:delete"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("删除CI类型"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
		{
			Id:          pointy.Uint64(2014),
			Level:       pointy.Uint32(3),
			MenuType:    pointy.Uint32(2),    // 按钮类型
			ParentId:    pointy.Uint64(2002), // 资产分类菜单ID
			Path:        pointy.String("/cmdb/ci_type"),
			Name:        pointy.String("CI类型详情"),
			Sort:        pointy.Uint32(6),
			Permission:  pointy.String("cmdb:ci_type:detail"),
			ServiceName: pointy.String("Cmdb"),
			TenantId:    pointy.Uint64(1),
			Meta: &core.Meta{
				Title:    pointy.String("CI类型详情"),
				Icon:     pointy.String("#"),
				HideMenu: pointy.Bool(true),
			},
		},
	}

	// 分层插入：先插入子菜单，再插入按钮权限
	// 分离菜单和按钮权限
	var menus []*core.MenuInfo
	var buttons []*core.MenuInfo
	
	for _, item := range subMenus {
		if item.MenuType != nil && *item.MenuType == 2 {
			buttons = append(buttons, item)
		} else {
			menus = append(menus, item)
		}
	}
	
	logx.Infof("Starting CMDB menu insertion: %d menus, %d buttons", len(menus), len(buttons))
	
	successCount := 0
	skippedCount := 0
	
	// 建立子菜单ID映射（计划ID -> 实际ID）
	subMenuIdMap := make(map[uint64]uint64)
	
	// 第一步：插入子菜单（MenuType = 1）
	logx.Info("Step 1: Inserting CMDB sub-menus...")
	for _, menu := range menus {
		// 检查菜单是否已存在
		if existingMenus[*menu.Id] {
			logx.Infof("CMDB menu '%s' (ID: %d) already exists, skipping", *menu.Name, *menu.Id)
			subMenuIdMap[*menu.Id] = *menu.Id // 已存在的菜单保持原ID
			skippedCount++
			continue
		}

		// 菜单不存在，执行插入
		createResp, err := l.svcCtx.CoreRpc.CreateMenu(tenantCtx, menu)
		if err != nil {
			logx.Errorw("Failed to create CMDB menu due to Core service issues",
				logx.Field("menu", menu.Name),
				logx.Field("expected_id", menu.Id),
				logx.Field("menu_type", menu.MenuType),
				logx.Field("error", err.Error()))
			continue
		}

		if createResp != nil && createResp.Id != 0 {
			subMenuIdMap[*menu.Id] = createResp.Id // 记录实际ID
			if createResp.Id == *menu.Id {
				logx.Infof("Created CMDB menu '%s' with expected ID: %d", *menu.Name, createResp.Id)
			} else {
				logx.Infof("Created CMDB menu '%s' with auto-generated ID: %d (expected: %d)", *menu.Name, createResp.Id, *menu.Id)
			}
			successCount++
			// 将新创建的菜单加入已存在映射，供按钮权限使用
			existingMenus[createResp.Id] = true
		}
	}
	
	// 第二步：插入按钮权限（MenuType = 2）
	logx.Info("Step 2: Inserting CMDB button permissions...")
	for _, button := range buttons {
		// 检查按钮是否已存在
		if existingMenus[*button.Id] {
			logx.Infof("CMDB button '%s' (ID: %d) already exists, skipping", *button.Name, *button.Id)
			skippedCount++
			continue
		}

		// 更新按钮的父菜单ID为实际ID
		originalParentId := *button.ParentId
		if actualParentId, exists := subMenuIdMap[originalParentId]; exists {
			button.ParentId = pointy.Uint64(actualParentId)
			logx.Infof("Updated button '%s' parent ID from %d to %d", *button.Name, originalParentId, actualParentId)
		} else {
			logx.Errorw("Cannot find actual parent ID for button, skipping",
				logx.Field("button", button.Name),
				logx.Field("planned_parent_id", originalParentId))
			continue
		}

		// 按钮不存在，执行插入
		createResp, err := l.svcCtx.CoreRpc.CreateMenu(tenantCtx, button)
		if err != nil {
			logx.Errorw("Failed to create CMDB button due to Core service issues",
				logx.Field("button", button.Name),
				logx.Field("expected_id", button.Id),
				logx.Field("planned_parent_id", originalParentId),
				logx.Field("actual_parent_id", button.ParentId),
				logx.Field("menu_type", button.MenuType),
				logx.Field("error", err.Error()))
			continue
		}

		if createResp != nil && createResp.Id != 0 {
			if createResp.Id == *button.Id {
				logx.Infof("Created CMDB button '%s' with expected ID: %d (parent: %d)", *button.Name, createResp.Id, *button.ParentId)
			} else {
				logx.Infof("Created CMDB button '%s' with auto-generated ID: %d (expected: %d, parent: %d)", *button.Name, createResp.Id, *button.Id, *button.ParentId)
			}
			successCount++
		}
	}

	totalMenus := len(subMenus) + 1 // 所有子菜单 + 1个根菜单
	rootMenuSkipped := existingMenus[actualRootMenuId]
	
	logx.Infow("CMDB menu insertion completed",
		logx.Field("total_items", totalMenus),
		logx.Field("menu_count", len(menus)),
		logx.Field("button_count", len(buttons)),
		logx.Field("created_count", successCount),
		logx.Field("skipped_existing", skippedCount + (func() int { if rootMenuSkipped { return 1 } else { return 0 } }())),
		logx.Field("failed_count", totalMenus - successCount - skippedCount - (func() int { if rootMenuSkipped { return 1 } else { return 0 } }())),
		logx.Field("actual_root_menu_id", actualRootMenuId),
		logx.Field("planned_root_menu_id", rootMenuId))
	return nil
}