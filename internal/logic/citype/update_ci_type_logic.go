package citype

import (
	"context"
	"errors"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/attribute"
	"gitee.com/link234/cmdb-rpc/ent/cis"
	"gitee.com/link234/cmdb-rpc/ent/citype"
	"gitee.com/link234/cmdb-rpc/ent/citypeattribute"
	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroup"
	"gitee.com/link234/cmdb-rpc/ent/citypeattributegroupitem"
	"gitee.com/link234/cmdb-rpc/ent/citypeinheritance"
	"gitee.com/link234/cmdb-rpc/ent/schema"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiTypeLogic {
	return &UpdateCiTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiTypeLogic) UpdateCiType(in *cmdb.CiTypeInfo) (*cmdb.BaseResp, error) {
	tx, err := l.svcCtx.DB.BeginTx(l.ctx, nil)
	if err != nil {
		l.Logger.Errorf("开启事务失败: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// 查询现有的CiType
	oldCiType, err := tx.CiType.Get(l.ctx, *in.Id)
	if err != nil {
		l.Logger.Errorf("查询CiType失败: %v", err)
		return nil, err
	}

	// 检查uniqueId是否发生变化
	if oldCiType.UniqueID != *in.UniqueId {
		l.Logger.Infof("检测到uniqueId变更: 旧值=%d, 新值=%d", oldCiType.UniqueID, *in.UniqueId)

		// 验证新的uniqueId是否存在
		isExist, err := tx.Attribute.Query().Where(attribute.IDEQ(*in.UniqueId)).Exist(l.ctx)
		if err != nil {
			l.Logger.Errorf("查询新uniqueId属性是否存在失败: %v", err)
			return nil, err
		}
		if !isExist {
			l.Logger.Errorf("新uniqueId属性不存在: %d", *in.UniqueId)
			return nil, errors.New("新uniqueId属性不存在")
		}

		// 处理uniqueId变更逻辑
		err = l.handleUniqueIdChange(tx, oldCiType, *in.UniqueId)
		if err != nil {
			l.Logger.Errorf("处理uniqueId变更失败: %v", err)
			return nil, err
		}
	}

	// 处理继承关系变更
	if in.InheritedModels != nil {
		err = l.handleInheritanceChange(tx, oldCiType, in.IsInherited, in.InheritedModels)
		if err != nil {
			l.Logger.Errorf("处理继承关系变更失败: %v", err)
			return nil, err
		}
	}

	// 更新CiType基本信息
	query := tx.CiType.UpdateOneID(*in.Id).
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilAlias(in.Alias).
		SetNotNilUniqueID(in.UniqueId).
		SetNotNilIcon(in.Icon).
		SetNotNilShowID(in.ShowId).
		SetNotNilDefaultOrderAttrID(in.DefaultOrderAttr).
		SetNotNilIsInherited(in.IsInherited)

	// 设置是否继承标志
	if in.IsInherited != nil {
		query.SetNotNilIsInherited(in.IsInherited)
	}

	// 处理唯一性约束
	if len(in.UniqueConst) > 0 {
		uniqueConsts := []schema.CiTypeUniqueConstS{}
		for _, v := range in.UniqueConst {
			uniqueConsts = append(uniqueConsts, schema.CiTypeUniqueConstS{
				AttrIds: v.AttrIds,
			})
		}
		query.SetUniqueConst(uniqueConsts)
	}

	err = query.Exec(l.ctx)
	if err != nil {
		l.Logger.Errorf("更新CiType失败: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		l.Logger.Errorf("提交事务失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("更新CiType成功: ID=%d", *in.Id)
	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}

// handleInheritanceChange 处理继承关系变更逻辑
func (l *UpdateCiTypeLogic) handleInheritanceChange(tx *ent.Tx, oldCiType *ent.CiType, isInherited *bool, newInheritedModels []uint64) error {
	l.Logger.Infof("开始处理继承关系变更: CiType ID=%d", oldCiType.ID)

	// 获取当前的继承关系
	currentInheritances, err := tx.CiTypeInheritance.Query().
		Where(citypeinheritance.ChildIDEQ(oldCiType.ID)).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询当前继承关系失败: %v", err)
		return err
	}

	// 提取当前父模型ID列表
	currentParentIds := make(map[uint64]bool)
	for _, inheritance := range currentInheritances {
		currentParentIds[inheritance.ParentID] = true
	}

	// 提取新的父模型ID列表
	newParentIds := make(map[uint64]bool)
	for _, parentId := range newInheritedModels {
		newParentIds[parentId] = true
	}

	// 查找需要删除的继承关系
	toDelete := make([]uint64, 0)
	for parentId := range currentParentIds {
		if !newParentIds[parentId] {
			toDelete = append(toDelete, parentId)
		}
	}

	// 查找需要添加的继承关系
	toAdd := make([]uint64, 0)
	for parentId := range newParentIds {
		if !currentParentIds[parentId] {
			toAdd = append(toAdd, parentId)
		}
	}

	l.Logger.Infof("继承关系变更分析: 删除=%v, 添加=%v", toDelete, toAdd)

	// 检查是否存在资产实例，如果存在则不允许删除继承关系
	if len(toDelete) > 0 {
		hasInstances, err := tx.Cis.Query().Where(cis.TypeIDEQ(oldCiType.ID)).Exist(l.ctx)
		if err != nil {
			l.Logger.Errorf("检查资产实例是否存在失败: %v", err)
			return err
		}
		if hasInstances {
			l.Logger.Errorf("模型存在资产实例，不允许删除继承关系: CiType ID=%d", oldCiType.ID)
			return errors.New("模型存在资产实例，不允许删除继承关系")
		}
	}

	// 如果从继承改变为不继承则删除所有继承关系
	if isInherited != nil && !*isInherited {
		_, err := tx.CiTypeInheritance.Delete().
			Where(citypeinheritance.ChildIDEQ(oldCiType.ID)).
			Exec(l.ctx)
		if err != nil {
			l.Logger.Errorf("删除继承关系失败: %v", err)
			return err
		}
		l.Logger.Infof("成功删除继承关系: CiType ID=%d", oldCiType.ID)
	}

	if isInherited != nil && *isInherited {
		// 删除继承关系
		for _, parentId := range toDelete {
			deletedCount, err := tx.CiTypeInheritance.Delete().
				Where(citypeinheritance.ParentIDEQ(parentId), citypeinheritance.ChildIDEQ(oldCiType.ID)).
				Exec(l.ctx)
			if err != nil {
				l.Logger.Errorf("删除继承关系失败: parent_id=%d, child_id=%d, error=%v", parentId, oldCiType.ID, err)
				return err
			}
			l.Logger.Infof("成功删除继承关系: parent_id=%d, child_id=%d, 删除数量=%d", parentId, oldCiType.ID, deletedCount)
		}

		// 验证新的父模型是否存在
		for _, parentId := range toAdd {
			exists, err := tx.CiType.Query().Where(citype.IDEQ(parentId)).Exist(l.ctx)
			if err != nil {
				l.Logger.Errorf("查询父模型失败: parent_id=%d, error=%v", parentId, err)
				return err
			}
			if !exists {
				l.Logger.Errorf("父模型不存在: parent_id=%d", parentId)
				return errors.New("继承的父模型不存在")
			}
		}

		// 添加新的继承关系
		for _, parentId := range toAdd {
			_, err := tx.CiTypeInheritance.Create().
				SetParentID(parentId).
				SetChildID(oldCiType.ID).
				Save(l.ctx)
			if err != nil {
				l.Logger.Errorf("创建继承关系失败: parent_id=%d, child_id=%d, error=%v", parentId, oldCiType.ID, err)
				return err
			}
			l.Logger.Infof("成功创建继承关系: parent_id=%d, child_id=%d", parentId, oldCiType.ID)
		}
	}

	l.Logger.Infof("继承关系变更处理完成: CiType ID=%d, 删除数量=%d, 添加数量=%d", oldCiType.ID, len(toDelete), len(toAdd))
	return nil
}

// handleUniqueIdChange 处理uniqueId变更逻辑
func (l *UpdateCiTypeLogic) handleUniqueIdChange(tx *ent.Tx, oldCiType *ent.CiType, newUniqueId uint64) error {
	l.Logger.Infof("开始处理uniqueId变更: 从%d变更为%d", oldCiType.UniqueID, newUniqueId)

	// 1. 查询所有属性分组
	groupIds, err := tx.CiTypeAttributeGroup.Query().
		Where(citypeattributegroup.TypeIDEQ(oldCiType.ID)).
		Select(citypeattributegroup.FieldID).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询属性分组失败: %v", err)
		return err
	}

	var ids []uint64
	for _, group := range groupIds {
		ids = append(ids, group.ID)
	}

	// 4. 检查新uniqueId是否已经存在于CiTypeAttribute中
	existingCitAttr, err := tx.CiTypeAttribute.Query().
		Where(citypeattribute.TypeIDEQ(oldCiType.ID), citypeattribute.AttrIDEQ(newUniqueId)).
		First(l.ctx)
	if err != nil && !ent.IsNotFound(err) {
		l.Logger.Errorf("查询新uniqueId是否已关联CiType失败: %v", err)
		return err
	}

	if existingCitAttr == nil {
		// 5. 添加新uniqueId到CiTypeAttribute
		_, err = tx.CiTypeAttribute.Create().
			SetNotNilTypeID(pointy.GetPointer(oldCiType.ID)).
			SetNotNilAttrID(pointy.GetPointer(newUniqueId)).
			SetNotNilIsRequired(pointy.GetPointer(true)).
			SetNotNilSort(pointy.GetPointer(uint32(1))).
			Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("添加新uniqueId到CiTypeAttribute失败: %v", err)
			return err
		}
		l.Logger.Infof("成功添加新uniqueId到CiTypeAttribute: %d", newUniqueId)
	} else {
		// 6. 如果已存在，确保设置为必填
		_, err = tx.CiTypeAttribute.UpdateOneID(existingCitAttr.ID).
			SetIsRequired(true).
			SetSort(1).
			Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("更新新uniqueId为必填失败: %v", err)
			return err
		}
		l.Logger.Infof("更新现有uniqueId为必填属性: %d", newUniqueId)
	}

	// 7. 检查新uniqueId是否已在分组中，如果不在则添加到默认分组
	isInGroup, err := tx.CiTypeAttributeGroupItem.Query().
		Where(citypeattributegroupitem.AttrIDEQ(newUniqueId), citypeattributegroupitem.GroupIDIn(ids...)).
		Exist(l.ctx)
	if err != nil {
		l.Logger.Errorf("检查新uniqueId是否在分组中失败: %v", err)
		return err
	}

	if !isInGroup {
		// 8. 查找默认分组并添加新uniqueId
		defaultGroup, err := tx.CiTypeAttributeGroup.Query().
			Where(citypeattributegroup.TypeIDEQ(oldCiType.ID), citypeattributegroup.NameEQ("其他")).
			First(l.ctx)
		if err != nil {
			l.Logger.Errorf("查询默认分组失败: %v", err)
			return err
		}

		_, err = tx.CiTypeAttributeGroupItem.Create().
			SetNotNilAttrID(pointy.GetPointer(newUniqueId)).
			SetNotNilGroupID(pointy.GetPointer(defaultGroup.ID)).
			Save(l.ctx)
		if err != nil {
			l.Logger.Errorf("添加新uniqueId到默认分组失败: %v", err)
			return err
		}
		l.Logger.Infof("成功添加新uniqueId到默认分组: %d", newUniqueId)
	} else {
		l.Logger.Infof("新uniqueId已存在于分组中: %d", newUniqueId)
	}

	l.Logger.Infof("uniqueId变更处理完成: 从%d变更为%d", oldCiType.UniqueID, newUniqueId)
	return nil
}
