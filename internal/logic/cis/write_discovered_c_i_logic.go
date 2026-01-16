package cis

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuedatetime"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuefloat"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valueindextext"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valueinteger"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuejson"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuetext"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type WriteDiscoveredCILogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteDiscoveredCILogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteDiscoveredCILogic {
	return &WriteDiscoveredCILogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *WriteDiscoveredCILogic) WriteDiscoveredCI(in *cmdb.DiscoveredCIData) (*cmdb.WriteDiscoveredCIResp, error) {
	// 1. 验证 CI 类型是否存在
	if in.CiTypeId == nil {
		return nil, errorx.NewCodeError(400, "ci_type_id is required")
	}

	ciTypeExists, err := l.svcCtx.DB.CiType.Query().
		Where(citype.IDEQ(*in.CiTypeId)).
		Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if !ciTypeExists {
		return nil, fmt.Errorf("CI类型ID %d 不存在", *in.CiTypeId)
	}

	// 2. 根据 unique_key 查找已存在的 CI
	var existingCI *ent.Cis
	if in.UniqueKey != nil && *in.UniqueKey != "" {
		// 使用 metadata 中的 unique_key 字段查找
		cis, err := l.svcCtx.DB.Cis.Query().
			Where(
				cis.TypeIDEQ(*in.CiTypeId),
			).
			All(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// 在内存中过滤匹配 unique_key 的 CI
		for _, ci := range cis {
			if ci.Metadata != nil {
				if uk, ok := ci.Metadata["unique_key"].(string); ok && uk == *in.UniqueKey {
					existingCI = ci
					break
				}
			}
		}
	}

	// 3. 决定操作类型：创建、更新还是跳过
	if existingCI == nil {
		// CI 不存在
		if in.AutoCreate == nil || !*in.AutoCreate {
			return &cmdb.WriteDiscoveredCIResp{
				Action:  pointy.GetPointer("skipped"),
				Message: pointy.GetPointer("CI不存在且未启用自动创建"),
			}, nil
		}

		// 执行创建操作
		return l.createDiscoveredCI(in)
	}

	// CI 已存在
	if in.AutoUpdate == nil || !*in.AutoUpdate {
		return &cmdb.WriteDiscoveredCIResp{
			CiId:    pointy.GetPointer(existingCI.ID),
			Action:  pointy.GetPointer("skipped"),
			Message: pointy.GetPointer("CI已存在且未启用自动更新"),
		}, nil
	}

	// 执行更新操作
	return l.updateDiscoveredCI(existingCI.ID, in)
}

// createDiscoveredCI 创建发现的 CI
func (l *WriteDiscoveredCILogic) createDiscoveredCI(in *cmdb.DiscoveredCIData) (*cmdb.WriteDiscoveredCIResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 准备 metadata：添加发现来源信息和 unique_key
	metadata := make(map[string]interface{})
	if in.Source != nil {
		metadata["discovery_source"] = *in.Source
	}
	if in.SourceType != nil {
		metadata["discovery_source_type"] = *in.SourceType
	}
	if in.UniqueKey != nil {
		metadata["unique_key"] = *in.UniqueKey
	}
	metadata["discovered_at"] = time.Now().Format(time.RFC3339)

	// 创建 CI 实例
	cisBuilder := tx.Cis.Create().
		SetTypeID(*in.CiTypeId).
		SetStatus(1).
		SetMetadata(metadata)

	result, err := cisBuilder.Save(l.ctx)
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 保存属性值
	if in.Attributes != nil && len(in.Attributes) > 0 {
		err = l.saveDiscoveredAttributes(l.ctx, tx, result.ID, *in.CiTypeId, in.Attributes)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.WriteDiscoveredCIResp{
		CiId:    pointy.GetPointer(result.ID),
		Action:  pointy.GetPointer("created"),
		Message: pointy.GetPointer(fmt.Sprintf("成功创建CI，ID: %d", result.ID)),
	}, nil
}

// updateDiscoveredCI 更新发现的 CI
func (l *WriteDiscoveredCILogic) updateDiscoveredCI(ciID uint64, in *cmdb.DiscoveredCIData) (*cmdb.WriteDiscoveredCIResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 获取现有 CI 的 metadata
	existingCI, err := l.svcCtx.DB.Cis.Get(l.ctx, ciID)
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 更新 metadata：记录最后发现时间
	metadata := existingCI.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if in.Source != nil {
		metadata["discovery_source"] = *in.Source
	}
	if in.SourceType != nil {
		metadata["discovery_source_type"] = *in.SourceType
	}
	metadata["last_discovered_at"] = time.Now().Format(time.RFC3339)

	// 更新 CI 实例的 metadata
	_, err = tx.Cis.UpdateOneID(ciID).
		SetMetadata(metadata).
		Save(l.ctx)
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 根据冲突解决策略更新属性
	if in.Attributes != nil && len(in.Attributes) > 0 {
		conflictResolution := "merge" // 默认策略
		if in.ConflictResolution != nil {
			conflictResolution = *in.ConflictResolution
		}

		switch conflictResolution {
		case "update":
			// 完全覆盖：删除所有旧属性，写入新属性
			err = l.deleteAllAttributes(l.ctx, tx, ciID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			err = l.saveDiscoveredAttributes(l.ctx, tx, ciID, *in.CiTypeId, in.Attributes)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

		case "merge":
			// 合并：只更新新属性提供的字段，保留未提供的字段
			err = l.saveDiscoveredAttributes(l.ctx, tx, ciID, *in.CiTypeId, in.Attributes)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

		case "skip":
			// 跳过更新属性，只更新 metadata
			// 不做任何操作

		default:
			tx.Rollback()
			return nil, fmt.Errorf("不支持的冲突解决策略: %s", conflictResolution)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.WriteDiscoveredCIResp{
		CiId:    pointy.GetPointer(ciID),
		Action:  pointy.GetPointer("updated"),
		Message: pointy.GetPointer(fmt.Sprintf("成功更新CI，ID: %d", ciID)),
	}, nil
}

// saveDiscoveredAttributes 保存发现的属性值
func (l *WriteDiscoveredCILogic) saveDiscoveredAttributes(ctx context.Context, tx *ent.Tx, ciID uint64, ciTypeID uint64, attributes map[string]string) error {
	// 查询 CI 类型的所有属性定义（包含 attribute 边）
	typeAttrs, err := l.svcCtx.DB.CiTypeAttribute.Query().
		Where(citypeattribute.TypeIDEQ(ciTypeID)).
		WithAttribute(). // 加载关联的属性详情
		All(ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, attributes)
	}

	// 构建属性名称到属性定义的映射
	attrMap := make(map[string]*ent.CiTypeAttribute)
	for _, typeAttr := range typeAttrs {
		if typeAttr.Edges.Attribute != nil {
			attrMap[typeAttr.Edges.Attribute.Name] = typeAttr
		}
	}

	// 转换为 CiAttributeValue 列表
	var ciAttributes []*cmdb.CiAttributeValue
	for attrName, attrValue := range attributes {
		typeAttr, ok := attrMap[attrName]
		if !ok {
			// 属性不存在，记录警告但继续处理
			logx.WithContext(ctx).Infow("发现的属性在CI类型中不存在",
				logx.Field("ci_type_id", ciTypeID),
				logx.Field("attr_name", attrName))
			continue
		}

		attr := typeAttr.Edges.Attribute
		ciAttributes = append(ciAttributes, &cmdb.CiAttributeValue{
			AttrId:    attr.ID,
			AttrName:  attr.Name,
			AttrAlias: attr.Alias,
			ValueType: string(attr.ValueType),
			Value:     attrValue,
		})
	}

	// 使用现有的 SaveCiAttributes 函数保存属性
	if len(ciAttributes) > 0 {
		err = SaveCiAttributes(ctx, tx, ciID, ciAttributes)
		if err != nil {
			return dberrorhandler.DefaultEntError(l.Logger, err, ciAttributes)
		}
	}

	return nil
}

// deleteAllAttributes 删除 CI 的所有属性值
func (l *WriteDiscoveredCILogic) deleteAllAttributes(ctx context.Context, tx *ent.Tx, ciID uint64) error {
	// 删除所有类型的属性值
	if _, err := tx.ValueIndexText.Delete().Where(valueindextext.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}
	if _, err := tx.ValueText.Delete().Where(valuetext.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}
	if _, err := tx.ValueInteger.Delete().Where(valueinteger.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}
	if _, err := tx.ValueFloat.Delete().Where(valuefloat.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}
	if _, err := tx.ValueDatetime.Delete().Where(valuedatetime.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}
	if _, err := tx.ValueJSON.Delete().Where(valuejson.CiIDEQ(ciID)).Exec(ctx); err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, ciID)
	}

	return nil
}
