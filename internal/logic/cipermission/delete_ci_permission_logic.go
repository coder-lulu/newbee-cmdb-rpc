package cipermission

import (
	"context"
	"fmt"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/cipermission"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCiPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCiPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCiPermissionLogic {
	return &DeleteCiPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCiPermissionLogic) DeleteCiPermission(in *cmdb.IDsReq) (*cmdb.BaseResp, error) {
	l.Logger.Infof("开始删除CI权限，权限IDs: %v", in.Ids)

	// 1. 数据验证
	if err := l.validateDeleteRequest(in); err != nil {
		l.Logger.Errorf("删除CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 检查权限是否存在
	existingPermissions, err := l.svcCtx.DB.CiPermission.Query().
		Where(cipermission.IDIn(in.Ids...)).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询要删除的权限失败，IDs: %v, 错误: %v", in.Ids, err)
		return nil, fmt.Errorf("查询权限失败: %v", err)
	}

	if len(existingPermissions) != len(in.Ids) {
		// 找出不存在的权限ID
		existingIds := make(map[uint64]bool)
		for _, permission := range existingPermissions {
			existingIds[permission.ID] = true
		}

		var notFoundIds []uint64
		for _, id := range in.Ids {
			if !existingIds[id] {
				notFoundIds = append(notFoundIds, id)
			}
		}

		l.Logger.Errorf("部分权限不存在，不存在的权限IDs: %v", notFoundIds)
		return nil, fmt.Errorf("权限不存在，IDs: %v", notFoundIds)
	}

	// 3. 检查权限是否可以删除
	if err := l.validateDeletion(existingPermissions); err != nil {
		l.Logger.Errorf("权限删除验证失败: %v", err)
		return nil, err
	}

	// 4. 检查是否有子权限依赖
	for _, permission := range existingPermissions {
		childPermissions, err := l.svcCtx.DB.CiPermission.Query().
			Where(cipermission.ParentPermissionIDEQ(permission.PermissionID)).
			Count(l.ctx)
		if err != nil {
			l.Logger.Errorf("检查子权限依赖失败，权限ID: %s, 错误: %v", permission.PermissionID, err)
			return nil, fmt.Errorf("检查权限依赖失败: %v", err)
		}

		if childPermissions > 0 {
			l.Logger.Errorf("权限存在子权限依赖，无法删除，权限ID: %s, 子权限数量: %d",
				permission.PermissionID, childPermissions)
			return nil, fmt.Errorf("权限 [%s] 存在 %d 个子权限，无法删除",
				permission.PermissionID, childPermissions)
		}
	}

	// 5. 记录要删除的权限信息（用于日志）
	var deletedPermissionInfo []string
	for _, permission := range existingPermissions {
		deletedPermissionInfo = append(deletedPermissionInfo,
			fmt.Sprintf("ID:%d,PermissionID:%s,SubjectType:%s,SubjectName:%s",
				permission.ID, permission.PermissionID, permission.SubjectType, permission.SubjectName))
	}

	// 6. 执行删除操作
	deletedCount, err := l.svcCtx.DB.CiPermission.Delete().
		Where(cipermission.IDIn(in.Ids...)).
		Exec(l.ctx)
	if err != nil {
		l.Logger.Errorf("删除CI权限时数据库操作失败: %v", err)
		return nil, fmt.Errorf("删除CI权限失败: %v", err)
	}

	l.Logger.Infof("成功删除CI权限，删除数量: %d, 删除的权限信息: %v",
		deletedCount, deletedPermissionInfo)

	return &cmdb.BaseResp{
		Msg: fmt.Sprintf("成功删除 %d 个权限", deletedCount),
	}, nil
}

// validateDeleteRequest 验证删除请求的基本数据
func (l *DeleteCiPermissionLogic) validateDeleteRequest(in *cmdb.IDsReq) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	if len(in.Ids) == 0 {
		return fmt.Errorf("权限ID列表不能为空")
	}

	// 检查是否有重复的ID
	idMap := make(map[uint64]bool)
	for _, id := range in.Ids {
		if id == 0 {
			return fmt.Errorf("权限ID不能为0")
		}
		if idMap[id] {
			return fmt.Errorf("权限ID重复: %d", id)
		}
		idMap[id] = true
	}

	// 限制批量删除数量，防止误操作
	const maxBatchDeleteCount = 100
	if len(in.Ids) > maxBatchDeleteCount {
		return fmt.Errorf("批量删除数量超过限制，最大允许删除 %d 个权限，当前请求删除 %d 个",
			maxBatchDeleteCount, len(in.Ids))
	}

	return nil
}

// validateDeletion 验证权限是否可以删除
func (l *DeleteCiPermissionLogic) validateDeletion(permissions []*ent.CiPermission) error {
	for _, permission := range permissions {
		// 检查是否是系统级别的权限
		if permission.SubjectType == cipermission.SubjectTypeSystem {
			return fmt.Errorf("不能删除系统级权限，权限ID: %s", permission.PermissionID)
		}

		// 检查是否是超级管理员权限
		if permission.PermissionLevel == cipermission.PermissionLevelSuperAdmin {
			return fmt.Errorf("不能删除超级管理员权限，权限ID: %s", permission.PermissionID)
		}

		// 检查权限状态
		if permission.Status == cipermission.StatusActive && permission.RequireApproval {
			return fmt.Errorf("活跃状态的审批权限不能直接删除，请先禁用，权限ID: %s", permission.PermissionID)
		}

		// 检查是否是高风险权限
		if permission.RiskLevel == cipermission.RiskLevelCritical {
			return fmt.Errorf("不能删除高风险权限，请联系系统管理员，权限ID: %s", permission.PermissionID)
		}
	}

	return nil
}
