package attribute

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeinheritance"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeSimpleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeSimpleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeSimpleListLogic {
	return &GetAttributeSimpleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeSimpleListLogic) GetAttributeSimpleList(in *cmdb.AttributeSimpleListReq) (*cmdb.AttributeSimpleListResp, error) {
	query := l.svcCtx.DB.Attribute.Query()
	if in.Name != nil {
		query = query.Where(attribute.NameContains(*in.Name))
	}
	if in.Alias != nil {
		query = query.Where(attribute.AliasContains(*in.Alias))
	}
	var err error
	var attrs []*ent.Attribute
	if in.ExcludeCiType != nil {
		// 获取需要排除的所有属性ID（包括继承的属性）
		excludeIds, err := l.getExcludeAttributeIds(*in.ExcludeCiType)
		if err != nil {
			l.Logger.Errorf("获取排除属性ID失败: %v", err)
			return nil, err
		}

		if len(excludeIds) > 0 {
			attrs, err = query.Where(attribute.IDNotIn(excludeIds...)).All(l.ctx)
			if err != nil {
				return nil, err
			}
		} else {
			// 没有需要排除的属性，返回所有属性
			attrs, err = query.All(l.ctx)
			if err != nil {
				return nil, err
			}
		}
	} else {
		attrs, err = query.All(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	simpleAttrs := make([]*cmdb.AttributeSimple, 0, len(attrs))
	for _, attr := range attrs {
		valueTypeStr := string(attr.ValueType)
		simpleAttrs = append(simpleAttrs, &cmdb.AttributeSimple{
			Id:        &attr.ID,
			Name:      &attr.Name,
			Alias:     &attr.Alias,
			ValueType: &valueTypeStr,
			IsChoice:  &attr.IsChoice,
			IsList:    &attr.IsList,
		})
	}

	return &cmdb.AttributeSimpleListResp{
		Total: uint64(len(simpleAttrs)),
		Data:  simpleAttrs,
	}, nil
}

// getExcludeAttributeIds 获取需要排除的属性ID列表（包括继承的属性）
func (l *GetAttributeSimpleListLogic) getExcludeAttributeIds(excludeCiTypeId uint64) ([]uint64, error) {
	// 首先检查该模型是否启用了继承
	ciType, err := l.svcCtx.DB.CiType.Get(l.ctx, excludeCiTypeId)
	if err != nil {
		return nil, err
	}

	// 收集所有需要排除的类型ID
	var allTypeIds []uint64
	allTypeIds = append(allTypeIds, excludeCiTypeId) // 当前模型

	// 如果启用了继承，获取继承链
	if ciType.IsInherited != nil && *ciType.IsInherited {
		inheritanceChain, err := l.getInheritanceChain(excludeCiTypeId)
		if err != nil {
			l.Logger.Errorf("获取继承链失败: %v", err)
			return nil, err
		}
		allTypeIds = append(allTypeIds, inheritanceChain...)
		l.Logger.Infof("模型 %d 启用继承，排除类型链: %v", excludeCiTypeId, allTypeIds)
	} else {
		l.Logger.Infof("模型 %d 未启用继承，仅排除自身属性", excludeCiTypeId)
	}

	// 获取所有类型的属性ID
	var excludeIds []uint64
	for _, typeId := range allTypeIds {
		excludeAttrIds, err := l.svcCtx.DB.CiTypeAttribute.Query().
			Where(citypeattribute.TypeIDEQ(typeId)).
			Select(citypeattribute.FieldAttrID).
			All(l.ctx)
		if err != nil {
			l.Logger.Errorf("获取类型 %d 的属性失败: %v", typeId, err)
			continue
		}

		for _, cta := range excludeAttrIds {
			excludeIds = append(excludeIds, cta.AttrID)
		}
	}

	// 去重
	excludeIds = l.removeDuplicateIds(excludeIds)
	l.Logger.Infof("总共需要排除 %d 个属性", len(excludeIds))

	return excludeIds, nil
}

// getInheritanceChain 获取完整的继承链（从最顶层父类到直接父类）
func (l *GetAttributeSimpleListLogic) getInheritanceChain(typeId uint64) ([]uint64, error) {
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

	return result, nil
}

// removeDuplicateIds 去除重复的ID
func (l *GetAttributeSimpleListLogic) removeDuplicateIds(ids []uint64) []uint64 {
	seen := make(map[uint64]bool)
	var result []uint64

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}
