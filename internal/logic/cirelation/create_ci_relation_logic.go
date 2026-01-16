package cirelation

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cirelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cityperelation"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/relationtype"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiRelationLogic {
	return &CreateCiRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiRelationLogic) CreateCiRelation(in *cmdb.CiRelationInfo) (*cmdb.BaseIDResp, error) {
	// 1. 基础参数验证
	if in.SourceCiId == nil || in.TargetCiId == nil || in.RelationTypeId == nil {
		return nil, fmt.Errorf("源CI、目标CI和关系类型不能为空")
	}

	sourceCiID := *in.SourceCiId
	targetCiID := *in.TargetCiId
	relationTypeID := *in.RelationTypeId

	// 2. 验证源CI是否存在
	sourceCi, err := l.svcCtx.DB.Cis.Query().
		Where(cis.IDEQ(sourceCiID)).
		WithCiType().
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("源CI [%d] 不存在", sourceCiID)
		}
		return nil, fmt.Errorf("查询源CI失败: %w", err)
	}

	// 3. 验证目标CI是否存在
	targetCi, err := l.svcCtx.DB.Cis.Query().
		Where(cis.IDEQ(targetCiID)).
		WithCiType().
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("目标CI [%d] 不存在", targetCiID)
		}
		return nil, fmt.Errorf("查询目标CI失败: %w", err)
	}

	// 4. 验证关系类型是否存在且启用
	relationType, err := l.svcCtx.DB.RelationType.Query().
		Where(relationtype.IDEQ(relationTypeID)).
		First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("关系类型 [%d] 不存在", relationTypeID)
		}
		return nil, fmt.Errorf("查询关系类型失败: %w", err)
	}

	if !relationType.IsEnabled {
		return nil, fmt.Errorf("关系类型 [%s] 已禁用", relationType.Name)
	}

	// 5. 验证CI类型兼容性 - 检查是否存在允许的类型关系
	sourceCiTypeID := sourceCi.Edges.CiType.ID
	targetCiTypeID := targetCi.Edges.CiType.ID

	// 查询CI类型关系配置
	typeRelationExists, err := l.svcCtx.DB.CiTypeRelation.Query().
		Where(
			cityperelation.ParentIDEQ(sourceCiTypeID),
			cityperelation.ChildIDEQ(targetCiTypeID),
			cityperelation.RelationTypeIDEQ(relationTypeID),
		).
		Exist(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("查询CI类型关系配置失败: %w", err)
	}

	// 如果是双向关系，也检查反向关系
	if relationType.Direction == relationtype.DirectionBidirectional && !typeRelationExists {
		typeRelationExists, err = l.svcCtx.DB.CiTypeRelation.Query().
			Where(
				cityperelation.ParentIDEQ(targetCiTypeID),
				cityperelation.ChildIDEQ(sourceCiTypeID),
				cityperelation.RelationTypeIDEQ(relationTypeID),
			).
			Exist(l.ctx)
		if err != nil {
			return nil, fmt.Errorf("查询CI类型关系配置失败: %w", err)
		}
	}

	if !typeRelationExists {
		return nil, fmt.Errorf("CI类型 [%s] 和 [%s] 之间不支持关系类型 [%s]",
			sourceCi.Edges.CiType.Name, targetCi.Edges.CiType.Name, relationType.Name)
	}

	// 6. 验证关系是否已存在（防止重复创建）
	existingRelation, err := l.svcCtx.DB.CiRelation.Query().
		Where(
			cirelation.SourceCiIDEQ(sourceCiID),
			cirelation.TargetCiIDEQ(targetCiID),
			cirelation.RelationTypeIDEQ(relationTypeID),
		).
		First(l.ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询现有关系失败: %w", err)
	}
	if existingRelation != nil {
		return nil, fmt.Errorf("关系已存在: [%s] -> [%s] (%s)",
			sourceCi.Edges.CiType.Name, targetCi.Edges.CiType.Name, relationType.Name)
	}

	// 如果是双向关系，也检查反向关系是否存在
	if relationType.Direction == relationtype.DirectionBidirectional {
		existingReverseRelation, err := l.svcCtx.DB.CiRelation.Query().
			Where(
				cirelation.SourceCiIDEQ(targetCiID),
				cirelation.TargetCiIDEQ(sourceCiID),
				cirelation.RelationTypeIDEQ(relationTypeID),
			).
			First(l.ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询反向关系失败: %w", err)
		}
		if existingReverseRelation != nil {
			return nil, fmt.Errorf("反向关系已存在: [%s] -> [%s] (%s)",
				targetCi.Edges.CiType.Name, sourceCi.Edges.CiType.Name, relationType.Name)
		}
	}

	// 7. 创建关系
	creator := l.svcCtx.DB.CiRelation.Create().
		SetSourceCiID(sourceCiID).
		SetTargetCiID(targetCiID).
		SetRelationTypeID(relationTypeID)

	// 处理可选字段
	if in.More != nil {
		creator.SetMore(*in.More)
	}

	if in.DiscoverySource != nil {
		creator.SetDiscoverySource(*in.DiscoverySource)
	} else {
		creator.SetDiscoverySource("manual") // 默认为手动创建
	}

	if in.AncestorIds != nil {
		creator.SetAncestorIds(*in.AncestorIds)
	}

	if in.Properties != nil {
		// 将properties JSON字符串解析为map
		creator.SetProperties(map[string]interface{}{
			"raw": *in.Properties,
		})
	}

	result, err := creator.Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	l.Logger.Infof("成功创建CI关系: 源CI[%d] -> 目标CI[%d], 关系类型[%s]", 
		sourceCiID, targetCiID, relationType.Name)

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
