package init

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/consts"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/logx"
)

// InitStandardRelationTypes 初始化标准关系类型
// 在系统启动时自动创建预设的标准关系类型
func InitStandardRelationTypes(ctx context.Context, db *ent.Client) error {
	// 使用SystemContext确保可以跳过租户限制进行系统级操作
	systemCtx := hooks.NewSystemContext(ctx)
	
	logx.Info("开始初始化标准关系类型...")

	for i, stdType := range consts.StandardRelationTypes {
		// 检查关系类型是否已存在
		exists, err := db.RelationType.Query().
			Where(relationtype.CodeEQ(stdType.Code)).
			Exist(systemCtx)
		if err != nil {
			logx.Errorf("检查关系类型 %s 是否存在失败: %v", stdType.Code, err)
			return err
		}

		if exists {
			logx.Infof("关系类型 %s 已存在，跳过创建", stdType.Code)
			continue
		}

		// 创建标准关系类型
		_, err = db.RelationType.Create().
			SetName(stdType.Name).
			SetCode(stdType.Code).
			SetCategory(relationtype.Category(stdType.Category)).
			SetDirection(relationtype.Direction(stdType.Direction)).
			SetDescription(stdType.Description).
			SetIsStandard(true).
			SetSortOrder(i + 1). // 按定义顺序排序
			SetIsEnabled(true).
			Save(systemCtx)

		if err != nil {
			logx.Errorf("创建关系类型 %s 失败: %v", stdType.Code, err)
			return err
		}

		logx.Infof("成功创建标准关系类型: %s (%s)", stdType.Name, stdType.Code)
	}

	logx.Info("标准关系类型初始化完成")
	return nil
}

// UpdateStandardRelationTypes 更新标准关系类型
// 用于升级时更新已有的标准关系类型
func UpdateStandardRelationTypes(ctx context.Context, db *ent.Client) error {
	systemCtx := hooks.NewSystemContext(ctx)
	
	logx.Info("开始更新标准关系类型...")

	for i, stdType := range consts.StandardRelationTypes {
		// 查找现有的关系类型
		existing, err := db.RelationType.Query().
			Where(relationtype.CodeEQ(stdType.Code)).
			First(systemCtx)
		if err != nil {
			if ent.IsNotFound(err) {
				// 如果不存在，创建新的
				_, err = db.RelationType.Create().
					SetName(stdType.Name).
					SetCode(stdType.Code).
					SetCategory(relationtype.Category(stdType.Category)).
					SetDirection(relationtype.Direction(stdType.Direction)).
					SetDescription(stdType.Description).
					SetIsStandard(true).
					SetSortOrder(i + 1).
					SetIsEnabled(true).
					Save(systemCtx)
				if err != nil {
					logx.Errorf("创建关系类型 %s 失败: %v", stdType.Code, err)
					return err
				}
				logx.Infof("创建新的标准关系类型: %s (%s)", stdType.Name, stdType.Code)
			} else {
				logx.Errorf("查询关系类型 %s 失败: %v", stdType.Code, err)
				return err
			}
			continue
		}

		// 更新现有的关系类型（只更新标准类型）
		if existing.IsStandard {
			_, err = existing.Update().
				SetName(stdType.Name).
				SetCategory(relationtype.Category(stdType.Category)).
				SetDirection(relationtype.Direction(stdType.Direction)).
				SetDescription(stdType.Description).
				SetSortOrder(i + 1).
				Save(systemCtx)
			if err != nil {
				logx.Errorf("更新关系类型 %s 失败: %v", stdType.Code, err)
				return err
			}
			logx.Infof("更新标准关系类型: %s (%s)", stdType.Name, stdType.Code)
		}
	}

	logx.Info("标准关系类型更新完成")
	return nil
}

// GetStandardRelationTypeList 获取所有标准关系类型列表
func GetStandardRelationTypeList(ctx context.Context, db *ent.Client) ([]*ent.RelationType, error) {
	return db.RelationType.Query().
		Where(relationtype.IsStandardEQ(true)).
		Where(relationtype.IsEnabledEQ(true)).
		Order(ent.Asc(relationtype.FieldSortOrder)).
		All(ctx)
}

// IsValidStandardRelationType 验证是否为有效的标准关系类型
func IsValidStandardRelationType(code string) bool {
	return consts.IsStandardRelationType(code)
}