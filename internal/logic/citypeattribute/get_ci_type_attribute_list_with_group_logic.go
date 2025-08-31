package citypeattribute

import (
	"context"
	"sort"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeinheritance"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/logic/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeListWithGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeListWithGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeListWithGroupLogic {
	return &GetCiTypeAttributeListWithGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeListWithGroupLogic) GetCiTypeAttributeListWithGroup(in *cmdb.CiTypeAttributeListWithGroupReq) (*cmdb.CiTypeAttributeListWithGroupResp, error) {
	// 获取当前CiType信息
	currentCiType, err := l.svcCtx.DB.CiType.Get(l.ctx, in.TypeId)
	if err != nil {
		l.Logger.Errorf("获取CiType失败: %v", err)
		return nil, err
	}

	// 构建完整的类型ID列表（包含继承链）
	allTypeIds := []uint64{in.TypeId} // 当前模型排在最后（优先级最高）

	if currentCiType.IsInherited != nil && *currentCiType.IsInherited {
		// 获取继承链（父类在前，优先级低）
		inheritanceChain, err := l.getInheritanceChain(in.TypeId)
		if err != nil {
			l.Logger.Errorf("获取继承链失败: %v", err)
			return nil, err
		}
		// 按继承层级排序：最顶层父类 -> 直接父类 -> 当前模型
		allTypeIds = append(inheritanceChain, in.TypeId)
		l.Logger.Infof("CiType %d 完整类型链: %v", in.TypeId, allTypeIds)
	}

	// 一次性获取所有分组和属性
	result, err := l.buildGroupsWithAttributes(allTypeIds, in.TypeId)
	if err != nil {
		return nil, err
	}

	return &cmdb.CiTypeAttributeListWithGroupResp{
		Data: result,
	}, nil
}

// buildGroupsWithAttributes 构建包含属性的分组列表
func (l *GetCiTypeAttributeListWithGroupLogic) buildGroupsWithAttributes(typeIds []uint64, currentTypeId uint64) ([]*cmdb.CiTypeAttributeListWithGroupInfo, error) {
	// 收集所有属性，按优先级处理冲突
	allAttributes := make(map[uint64]*cmdb.CiTypeAttributeItem)         // key: attributeId
	groupMap := make(map[string]*cmdb.CiTypeAttributeListWithGroupInfo) // key: groupName

	// 首先获取当前模型的所有分组（只有当前模型的分组才能被返回）
	currentGroups, err := l.getCiTypeAttributeGroups(currentTypeId)
	if err != nil {
		l.Logger.Errorf("获取当前类型 %d 的分组失败: %v", currentTypeId, err)
		return nil, err
	}

	// 初始化当前模型的分组映射
	for _, group := range currentGroups {
		groupMap[group.Name] = &cmdb.CiTypeAttributeListWithGroupInfo{
			GroupId:    group.GroupId,
			Name:       group.Name,
			Sort:       group.Sort,
			TypeId:     currentTypeId,
			Attributes: make([]*cmdb.CiTypeAttributeItem, 0),
		}
	}

	// 确保存在"其它"分组（用于放置无对应分组的继承属性）
	l.ensureOtherGroup(groupMap, currentTypeId)

	// 按类型优先级处理（父类优先级低，子类优先级高）
	for _, typeId := range typeIds {
		isCurrentType := typeId == currentTypeId

		// 获取该类型的分组
		groups, err := l.getCiTypeAttributeGroups(typeId)
		if err != nil {
			l.Logger.Errorf("获取类型 %d 的分组失败: %v", typeId, err)
			continue
		}

		for _, group := range groups {
			// 处理分组中的属性
			for _, attr := range group.Attributes {
				if attr.Attribute == nil || attr.Attribute.Id == nil {
					continue
				}

				attrId := *attr.Attribute.Id

				// 检查属性是否已存在（优先级处理）
				if _, exists := allAttributes[attrId]; exists {
					// 如果当前是子类属性，替换父类属性
					if isCurrentType {
						attr.IsInherited = pointy.GetPointer(false)
						allAttributes[attrId] = attr
						// 找到之前的属性并从分组中移除，然后添加新的
						l.removeAttrFromGroups(attrId, groupMap)
						// 子类属性放到原分组
						groupMap[group.Name].Attributes = append(groupMap[group.Name].Attributes, attr)
						l.Logger.Debugf("子类属性覆盖父类: AttrID=%d", attrId)
					} else {
						l.Logger.Debugf("跳过重复的父类属性: AttrID=%d", attrId)
					}
					continue
				}

				// 新属性，添加到映射中
				attr.IsInherited = pointy.GetPointer(!isCurrentType) // 非当前类型即为继承属性
				allAttributes[attrId] = attr

				// 决定属性应该放在哪个分组
				var targetGroupName string
				if isCurrentType {
					// 子类自己的属性，放在原分组
					targetGroupName = group.Name
					l.Logger.Debugf("添加子类属性: AttrID=%d, 分组=%s", attrId, targetGroupName)
				} else {
					// 继承属性，检查子类是否有同名分组
					if _, exists := groupMap[group.Name]; exists {
						// 子类有同名分组，放在同名分组
						targetGroupName = group.Name
						l.Logger.Debugf("添加继承属性到同名分组: AttrID=%d, 分组=%s, 来源类型=%d", attrId, targetGroupName, typeId)
					} else {
						// 子类没有同名分组，放在"其它"分组
						targetGroupName = l.getOtherGroupName(groupMap)
						l.Logger.Debugf("添加继承属性到其它分组: AttrID=%d, 原分组=%s, 目标分组=%s, 来源类型=%d", attrId, group.Name, targetGroupName, typeId)
					}
				}

				// 将属性添加到目标分组
				groupMap[targetGroupName].Attributes = append(groupMap[targetGroupName].Attributes, attr)
			}
		}
	}

	// 将属性分配到对应分组（属性已经在收集时就知道所属分组）
	// 属性在上面的循环中已经被分配到了正确的分组，这里不需要重新分配

	// 对每个分组内的属性进行排序（继承属性在前，自有属性在后）
	for _, group := range groupMap {
		l.sortGroupAttributes(group.Attributes)
	}

	// 转换为列表并按分组排序
	result := make([]*cmdb.CiTypeAttributeListWithGroupInfo, 0, len(groupMap))
	for _, group := range groupMap {
		// 所有分组都返回，包括空分组（方便前端拖拽排序）
		result = append(result, group)
	}

	// 按分组sort字段排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Sort < result[j].Sort
	})

	l.Logger.Infof("构建完成: %d个分组(包含空分组), %d个属性", len(result), len(allAttributes))
	return result, nil
}

// sortGroupAttributes 对分组内属性排序：继承属性在前，自有属性在后
func (l *GetCiTypeAttributeListWithGroupLogic) sortGroupAttributes(attributes []*cmdb.CiTypeAttributeItem) {
	sort.Slice(attributes, func(i, j int) bool {
		attrI, attrJ := attributes[i], attributes[j]

		// 第一优先级：继承属性排在前面
		if attrI.IsInherited != attrJ.IsInherited {
			return *attrI.IsInherited // true排在前面
		}

		// 第二优先级：按sort字段排序
		sortI := uint32(999)
		sortJ := uint32(999)
		if attrI.Sort != nil {
			sortI = *attrI.Sort
		}
		if attrJ.Sort != nil {
			sortJ = *attrJ.Sort
		}

		return sortI < sortJ
	})
}

// getCiTypeAttributeGroups 获取指定CiType的所有属性分组（简化版）
func (l *GetCiTypeAttributeListWithGroupLogic) getCiTypeAttributeGroups(typeId uint64) ([]*cmdb.CiTypeAttributeListWithGroupInfo, error) {
	// 获取分组
	ciTypeAttributeGroupList, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().
		Where(citypeattributegroup.TypeIDEQ(typeId)).
		Order(citypeattributegroup.BySort()).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	if len(ciTypeAttributeGroupList) == 0 {
		return []*cmdb.CiTypeAttributeListWithGroupInfo{}, nil
	}

	result := make([]*cmdb.CiTypeAttributeListWithGroupInfo, 0)

	for _, group := range ciTypeAttributeGroupList {
		if group == nil || group.TypeID == 0 || group.ID == 0 {
			continue
		}

		// 获取分组项
		groupItems, err := l.svcCtx.DB.CiTypeAttributeGroupItem.Query().
			Where(citypeattributegroupitem.GroupIDEQ(group.ID)).
			Order(citypeattributegroupitem.BySort()).
			All(l.ctx)
		if err != nil {
			return nil, err
		}

		attributes := make([]*cmdb.CiTypeAttributeItem, 0)

		for _, item := range groupItems {
			if item == nil || item.AttrID == 0 || item.ID == 0 {
				continue
			}

			// 获取属性信息
			attr, err := attribute.NewGetAttributeByIdLogic(l.ctx, l.svcCtx).GetAttributeById(&cmdb.IDReq{Id: item.AttrID})
			if err != nil || attr == nil || attr.Id == nil {
				continue
			}

			// 获取CiTypeAttribute信息
			ciTypeAttr, err := l.svcCtx.DB.CiTypeAttribute.Query().
				Where(citypeattribute.TypeIDEQ(group.TypeID), citypeattribute.AttrIDEQ(*attr.Id)).
				First(l.ctx)
			if err != nil || ciTypeAttr == nil {
				continue
			}

			attributes = append(attributes, &cmdb.CiTypeAttributeItem{
				Sort:              &item.Sort,
				Id:                &item.ID,
				GroupId:           &group.ID,
				TypeId:            group.TypeID,
				Attribute:         attr,
				IsRequired:        &ciTypeAttr.IsRequired,
				ListShow:          &ciTypeAttr.ListShow,
				IsEdit:            &ciTypeAttr.IsEdit,
				DetailShow:        &ciTypeAttr.DetailShow,
				IsUnique:          &ciTypeAttr.IsUnique,
				CiTypeAttributeId: &ciTypeAttr.ID,
				IsInherited:       pointy.GetPointer(false),
			})
		}

		result = append(result, &cmdb.CiTypeAttributeListWithGroupInfo{
			GroupId:    group.ID,
			Name:       group.Name,
			Sort:       group.Sort,
			TypeId:     group.TypeID,
			Attributes: attributes,
		})
	}

	return result, nil
}

// getInheritanceChain 获取完整的继承链（从最顶层父类到直接父类）
func (l *GetCiTypeAttributeListWithGroupLogic) getInheritanceChain(typeId uint64) ([]uint64, error) {
	visited := make(map[uint64]bool)
	var result []uint64

	// BFS获取所有父类
	queue := []uint64{typeId}
	visited[typeId] = true

	for len(queue) > 0 {
		currentId := queue[0]
		queue = queue[1:]

		parents, err := l.svcCtx.DB.CiTypeInheritance.Query().
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

	// 反转数组，让最顶层父类排在前面
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// removeAttrFromGroups 从所有分组中移除指定属性ID的属性
func (l *GetCiTypeAttributeListWithGroupLogic) removeAttrFromGroups(attrId uint64, groupMap map[string]*cmdb.CiTypeAttributeListWithGroupInfo) {
	for _, group := range groupMap {
		for i, attr := range group.Attributes {
			if attr.Attribute != nil && attr.Attribute.Id != nil && *attr.Attribute.Id == attrId {
				// 移除该属性
				group.Attributes = append(group.Attributes[:i], group.Attributes[i+1:]...)
				return
			}
		}
	}
}

// ensureOtherGroup 确保存在"其它"分组，如果不存在则创建
func (l *GetCiTypeAttributeListWithGroupLogic) ensureOtherGroup(groupMap map[string]*cmdb.CiTypeAttributeListWithGroupInfo, currentTypeId uint64) {
	otherGroupName := "其它"
	if _, exists := groupMap[otherGroupName]; !exists {
		// 查询数据库中是否已有"其它"分组
		otherGroup, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().
			Where(citypeattributegroup.TypeIDEQ(currentTypeId), citypeattributegroup.NameEQ(otherGroupName)).
			First(l.ctx)

		if err != nil {
			// 数据库中没有"其它"分组，创建一个虚拟的（仅用于显示）
			l.Logger.Infof("为类型 %d 创建虚拟其它分组", currentTypeId)
			groupMap[otherGroupName] = &cmdb.CiTypeAttributeListWithGroupInfo{
				GroupId:    0, // 虚拟分组ID为0
				Name:       otherGroupName,
				Sort:       999, // 排在最后
				TypeId:     currentTypeId,
				Attributes: make([]*cmdb.CiTypeAttributeItem, 0),
			}
		} else {
			// 数据库中有"其它"分组，使用真实数据
			groupMap[otherGroupName] = &cmdb.CiTypeAttributeListWithGroupInfo{
				GroupId:    otherGroup.ID,
				Name:       otherGroup.Name,
				Sort:       otherGroup.Sort,
				TypeId:     currentTypeId,
				Attributes: make([]*cmdb.CiTypeAttributeItem, 0),
			}
		}
	}
}

// getOtherGroupName 获取"其它"分组的名称
func (l *GetCiTypeAttributeListWithGroupLogic) getOtherGroupName(groupMap map[string]*cmdb.CiTypeAttributeListWithGroupInfo) string {
	// 优先查找"其它"
	if _, exists := groupMap["其它"]; exists {
		return "其它"
	}
	// 其次查找"其他"
	if _, exists := groupMap["其他"]; exists {
		return "其他"
	}
	// 如果都没有，返回"其它"（应该在ensureOtherGroup中已经创建）
	return "其它"
}
