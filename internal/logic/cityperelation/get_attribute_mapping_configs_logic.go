package cityperelation

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cityperelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeMappingConfigsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeMappingConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeMappingConfigsLogic {
	return &GetAttributeMappingConfigsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeMappingConfigsLogic) GetAttributeMappingConfigs(in *cmdb.AttributeMappingConfigReq) (*cmdb.AttributeMappingConfigResp, error) {
	var predicates []predicate.CiTypeRelation

	// 源CI类型ID过滤
	if in.SourceCiTypeId != nil {
		predicates = append(predicates, cityperelation.ParentIDEQ(*in.SourceCiTypeId))
	}

	// 目标CI类型ID过滤
	if in.TargetCiTypeId != nil {
		predicates = append(predicates, cityperelation.ChildIDEQ(*in.TargetCiTypeId))
	}

	// 关系类型ID过滤
	if in.RelationTypeId != nil {
		predicates = append(predicates, cityperelation.RelationTypeIDEQ(*in.RelationTypeId))
	}

	// 如果只返回已配置映射的关系
	if in.ConfiguredOnly != nil && *in.ConfiguredOnly {
		// 添加条件：至少有一个属性映射配置
		predicates = append(predicates, cityperelation.Or(
			cityperelation.ParentAttrIDNEQ(0),    // 有单属性映射
			cityperelation.ChildAttrIDNEQ(0),     // 有单属性映射
			cityperelation.ParentAttrIdsNotNil(), // 有多属性映射
			cityperelation.ChildAttrIdsNotNil(),  // 有多属性映射
		))
	}

	// 查询关系定义并预加载相关信息
	relations, err := l.svcCtx.DB.CiTypeRelation.Query().
		Where(predicates...).
		WithParent().       // 预加载父类型（源CI类型）
		WithChild().        // 预加载子类型（目标CI类型）
		WithRelationType(). // 预加载关系类型
		All(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.AttributeMappingConfigResp{
		Configs:    []*cmdb.AttributeMappingConfigInfo{},
		TotalCount: uint64(len(relations)),
	}

	// 转换为响应格式
	for _, relation := range relations {
		config, err := l.convertToConfigInfo(relation, in.IncludeAttributeDetails)
		if err != nil {
			l.Logger.Errorf("Failed to convert relation to config info: %v", err)
			continue
		}
		resp.Configs = append(resp.Configs, config)
	}

	return resp, nil
}

// convertToConfigInfo 将关系定义转换为属性映射配置信息
func (l *GetAttributeMappingConfigsLogic) convertToConfigInfo(relation *ent.CiTypeRelation, includeDetails *bool) (*cmdb.AttributeMappingConfigInfo, error) {
	config := &cmdb.AttributeMappingConfigInfo{
		RelationId: relation.ID,
		Constraint: &relation.Constraint,
	}

	// 填充源CI类型信息
	if relation.Edges.Parent != nil {
		config.SourceCiType = &cmdb.CiTypeBasicInfo{
			Id:    relation.Edges.Parent.ID,
			Name:  relation.Edges.Parent.Name,
			Alias: relation.Edges.Parent.Alias,
		}
	}

	// 填充目标CI类型信息
	if relation.Edges.Child != nil {
		config.TargetCiType = &cmdb.CiTypeBasicInfo{
			Id:    relation.Edges.Child.ID,
			Name:  relation.Edges.Child.Name,
			Alias: relation.Edges.Child.Alias,
		}
	}

	// 填充关系类型信息
	if relation.Edges.RelationType != nil {
		config.RelationType = &cmdb.RelationTypeBasicInfo{
			Id:   relation.Edges.RelationType.ID,
			Name: relation.Edges.RelationType.Name,
			Code: relation.Edges.RelationType.Code,
		}
	}

	// 处理单属性映射
	if relation.ParentAttrID != 0 && relation.ChildAttrID != 0 {
		singleMapping := &cmdb.AttributeMappingPair{}
		
		// 如果需要包含详细的属性信息
		if includeDetails == nil || *includeDetails {
			// 这里需要查询属性详细信息
			// 由于属性可能来自不同的表，这里先用基础信息代替
			singleMapping.SourceAttribute = &cmdb.AttributeBasicInfo{
				Id: relation.ParentAttrID,
				// Name, Alias, DataType 需要从属性表查询
			}
			singleMapping.TargetAttribute = &cmdb.AttributeBasicInfo{
				Id: relation.ChildAttrID,
				// Name, Alias, DataType 需要从属性表查询
			}
		} else {
			// 只包含ID信息
			singleMapping.SourceAttribute = &cmdb.AttributeBasicInfo{Id: relation.ParentAttrID}
			singleMapping.TargetAttribute = &cmdb.AttributeBasicInfo{Id: relation.ChildAttrID}
		}
		
		config.SingleMapping = singleMapping
	}

	// 处理多属性映射
	if len(relation.ParentAttrIds) > 0 && len(relation.ChildAttrIds) > 0 {
		// 确保两个数组长度一致，否则无法正确配对
		minLen := len(relation.ParentAttrIds)
		if len(relation.ChildAttrIds) < minLen {
			minLen = len(relation.ChildAttrIds)
		}

		for i := 0; i < minLen; i++ {
			mapping := &cmdb.AttributeMappingPair{}
			
			if includeDetails == nil || *includeDetails {
				// 包含详细信息时需要查询属性表
				mapping.SourceAttribute = &cmdb.AttributeBasicInfo{
					Id: relation.ParentAttrIds[i],
					// Name, Alias, DataType 需要从属性表查询
				}
				mapping.TargetAttribute = &cmdb.AttributeBasicInfo{
					Id: relation.ChildAttrIds[i],
					// Name, Alias, DataType 需要从属性表查询
				}
			} else {
				// 只包含ID信息
				mapping.SourceAttribute = &cmdb.AttributeBasicInfo{Id: relation.ParentAttrIds[i]}
				mapping.TargetAttribute = &cmdb.AttributeBasicInfo{Id: relation.ChildAttrIds[i]}
			}
			
			config.MultipleMappings = append(config.MultipleMappings, mapping)
		}
	}

	return config, nil
}
