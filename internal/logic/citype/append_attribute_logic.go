package citype

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeinheritance"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppendAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppendAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppendAttributeLogic {
	return &AppendAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AppendAttributeLogic) AppendAttribute(in *cmdb.CiTypeAppendAttributeReq) (*cmdb.BaseResp, error) {
	// 启用事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取事务失败: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// 1. 验证模型是否存在
	ciType, err := tx.CiType.Query().Where(citype.IDEQ(*in.TypeId)).First(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取模型失败: %v", err)
		return nil, fmt.Errorf("模型不存在")
	}
	l.Logger.Infof("验证模型成功: ID=%d, Name=%s", ciType.ID, ciType.Name)

	// 2. 验证分组是否存在且属于该模型
	group, err := tx.CiTypeAttributeGroup.Query().
		Where(citypeattributegroup.IDEQ(*in.GroupId), citypeattributegroup.TypeIDEQ(*in.TypeId)).
		First(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取分组失败: %v", err)
		return nil, fmt.Errorf("分组不存在或不属于该模型")
	}
	l.Logger.Infof("验证分组成功: ID=%d, Name=%s", group.ID, group.Name)

	// 3. 获取模型已有的属性ID（包括继承的属性）
	existingAttrIds, err := l.getModelAllAttributeIds(tx, *in.TypeId, ciType)
	if err != nil {
		l.Logger.Errorf("获取模型现有属性失败: %v", err)
		return nil, err
	}
	l.Logger.Infof("模型 %d 现有属性数量: %d", *in.TypeId, len(existingAttrIds))

	// 4. 处理每个要追加的属性
	var successCount, skipCount int
	for _, attrId := range in.AttrIds {
		// 验证属性是否存在
		attr, err := tx.Attribute.Get(l.ctx, attrId)
		if err != nil {
			l.Logger.Errorf("属性 %d 不存在: %v", attrId, err)
			continue
		}

		// 检查属性是否已经被模型拥有（包括继承）
		if _, exists := existingAttrIds[attrId]; exists {
			l.Logger.Infof("跳过已存在的属性: ID=%d, Name=%s", attr.ID, attr.Name)
			skipCount++
			continue
		}

		// 添加属性到模型
		err = l.addAttributeToModel(tx, *in.TypeId, *in.GroupId, attrId)
		if err != nil {
			l.Logger.Errorf("添加属性 %d 到模型失败: %v", attrId, err)
			continue
		}

		l.Logger.Infof("成功添加属性: ID=%d, Name=%s", attr.ID, attr.Name)
		successCount++
	}

	// 5. 提交事务
	err = tx.Commit()
	if err != nil {
		l.Logger.Errorf("提交事务失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("追加属性完成: 成功=%d, 跳过=%d", successCount, skipCount)
	return &cmdb.BaseResp{
		Msg: fmt.Sprintf("追加属性完成，成功添加 %d 个属性，跳过 %d 个已存在属性", successCount, skipCount),
	}, nil
}

// getModelAllAttributeIds 获取模型拥有的所有属性ID（包括继承的属性）
func (l *AppendAttributeLogic) getModelAllAttributeIds(tx *ent.Tx, typeId uint64, ciType *ent.CiType) (map[uint64]bool, error) {
	attrIds := make(map[uint64]bool)

	// 收集需要检查的所有类型ID
	var allTypeIds []uint64
	allTypeIds = append(allTypeIds, typeId) // 当前模型

	// 如果启用了继承，获取继承链
	if ciType.IsInherited != nil && *ciType.IsInherited {
		inheritanceChain, err := l.getInheritanceChain(tx, typeId)
		if err != nil {
			l.Logger.Errorf("获取继承链失败: %v", err)
			return nil, err
		}
		allTypeIds = append(allTypeIds, inheritanceChain...)
		l.Logger.Infof("模型 %d 启用继承，检查类型链: %v", typeId, allTypeIds)
	}

	// 获取所有类型的属性ID
	for _, tId := range allTypeIds {
		attrs, err := tx.CiTypeAttribute.Query().
			Where(citypeattribute.TypeIDEQ(tId)).
			Select(citypeattribute.FieldAttrID).
			All(l.ctx)
		if err != nil {
			l.Logger.Errorf("获取类型 %d 的属性失败: %v", tId, err)
			continue
		}

		for _, attr := range attrs {
			attrIds[attr.AttrID] = true
		}
	}

	return attrIds, nil
}

// getInheritanceChain 获取完整的继承链（从最顶层父类到直接父类）
func (l *AppendAttributeLogic) getInheritanceChain(tx *ent.Tx, typeId uint64) ([]uint64, error) {
	visited := make(map[uint64]bool)
	var result []uint64

	// BFS获取所有父类
	queue := []uint64{typeId}
	visited[typeId] = true

	for len(queue) > 0 {
		currentId := queue[0]
		queue = queue[1:]

		parents, err := tx.CiTypeInheritance.Query().
			Where(citypeinheritance.ChildIDEQ(currentId)).
			All(l.ctx)
		if err != nil {
			return nil, err
		}

		for _, parent := range parents {
			if !visited[parent.ParentID] {
				visited[parent.ParentID] = true
				result = append(result, parent.ParentID)
				queue = append(queue, parent.ParentID)
			}
		}
	}

	return result, nil
}

// addAttributeToModel 添加属性到模型
func (l *AppendAttributeLogic) addAttributeToModel(tx *ent.Tx, typeId, groupId, attrId uint64) error {
	// 1. 创建 CiTypeAttribute 关联（使用默认值）
	_, err := tx.CiTypeAttribute.Create().
		SetTypeID(typeId).
		SetAttrID(attrId).
		SetIsRequired(false). // 默认非必填
		SetListShow(true).    // 默认列表显示
		SetIsEdit(true).      // 默认可编辑
		SetDetailShow(true).  // 默认详情显示
		SetIsUnique(false).   // 默认非唯一
		Save(l.ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	// 2. 创建 CiTypeAttributeGroupItem 关联
	_, err = tx.CiTypeAttributeGroupItem.Create().
		SetGroupID(groupId).
		SetAttrID(attrId).
		SetSort(999). // 默认排序值，放在最后
		Save(l.ctx)
	if err != nil {
		return dberrorhandler.DefaultEntError(l.Logger, err, nil)
	}

	return nil
}
