package cipermission

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

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
	l.Logger.Infof("开始删除CI权限，IDs: %v", in.Ids)

	// 1. 数据验证
	if err := l.validateDeleteRequest(in); err != nil {
		l.Logger.Errorf("删除CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 检查权限是否存在以及是否可以删除
	for _, id := range in.Ids {
		exists, err := l.svcCtx.DB.CiPermission.Query().
			Where(cipermission.IDEQ(id)).
			Exist(l.ctx)
		if err != nil {
			l.Logger.Errorf("检查权限存在性时数据库查询失败，ID: %d, 错误: %v", id, err)
			return nil, fmt.Errorf("检查权限存在性失败: %v", err)
		}
		if !exists {
			err := fmt.Errorf("权限ID [%d] 不存在", id)
			l.Logger.Errorf("权限不存在: %v", err)
			return nil, err
		}

		// 检查权限是否有子权限依赖
		hasChildren, err := l.svcCtx.DB.CiPermission.Query().
			Where(cipermission.ParentPermissionIDEQ(fmt.Sprintf("%d", id))).
			Exist(l.ctx)
		if err != nil {
			l.Logger.Errorf("检查子权限依赖时数据库查询失败，ID: %d, 错误: %v", id, err)
			return nil, fmt.Errorf("检查子权限依赖失败: %v", err)
		}
		if hasChildren {
			err := fmt.Errorf("权限ID [%d] 存在子权限依赖，无法删除", id)
			l.Logger.Errorf("存在子权限依赖: %v", err)
			return nil, err
		}
	}

	// 3. 执行删除操作
	deletedCount, err := l.svcCtx.DB.CiPermission.Delete().
		Where(cipermission.IDIn(in.Ids...)).
		Exec(l.ctx)
	if err != nil {
		l.Logger.Errorf("删除CI权限时数据库操作失败: %v", err)
		return nil, fmt.Errorf("删除CI权限失败: %v", err)
	}

	l.Logger.Infof("成功删除CI权限，删除数量: %d, IDs: %v", deletedCount, in.Ids)

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
		return fmt.Errorf("删除的权限ID列表不能为空")
	}

	// 检查ID的有效性
	for _, id := range in.Ids {
		if id <= 0 {
			return fmt.Errorf("无效的权限ID: %d", id)
		}
	}

	// 限制批量删除的数量，防止误操作
	if len(in.Ids) > 100 {
		return fmt.Errorf("批量删除数量不能超过100个，当前数量: %d", len(in.Ids))
	}

	return nil
}