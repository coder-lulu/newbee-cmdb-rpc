package attributemappingrule

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attributemappingrule"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
    "github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeMappingRuleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeMappingRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeMappingRuleListLogic {
	return &GetAttributeMappingRuleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeMappingRuleListLogic) GetAttributeMappingRuleList(in *cmdb.AttributeMappingRuleListReq) (*cmdb.AttributeMappingRuleListResp, error) {
	var predicates []predicate.AttributeMappingRule
	if in.CreatedAt != nil {
		predicates = append(predicates, attributemappingrule.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, attributemappingrule.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.Status != nil {
		predicates = append(predicates, attributemappingrule.StatusEQ(uint8(*in.Status)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, attributemappingrule.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.DiscoveryConfigId != nil {
		predicates = append(predicates, attributemappingrule.DiscoveryConfigIDEQ(*in.DiscoveryConfigId))
	}
	if in.CiAttributeId != nil {
		predicates = append(predicates, attributemappingrule.CiAttributeIDEQ(*in.CiAttributeId))
	}
	if in.SourceField != nil {
		predicates = append(predicates, attributemappingrule.SourceFieldContains(*in.SourceField))
	}
	if in.SourcePath != nil {
		predicates = append(predicates, attributemappingrule.SourcePathContains(*in.SourcePath))
	}
	if in.TargetAttribute != nil {
		predicates = append(predicates, attributemappingrule.TargetAttributeContains(*in.TargetAttribute))
	}
	if in.TransformType != nil {
		predicates = append(predicates, attributemappingrule.TransformTypeContains(*in.TransformType))
	}
	// Skip TransformConfig query as it's a JSON field
	if in.DefaultValue != nil {
		predicates = append(predicates, attributemappingrule.DefaultValueContains(*in.DefaultValue))
	}
	// Skip ValidationRules query as it's a JSON field
	if in.ValidationRegex != nil {
		predicates = append(predicates, attributemappingrule.ValidationRegexContains(*in.ValidationRegex))
	}
	if in.IsRequired != nil {
		predicates = append(predicates, attributemappingrule.IsRequiredEQ(*in.IsRequired))
	}
	if in.IsUnique != nil {
		predicates = append(predicates, attributemappingrule.IsUniqueEQ(*in.IsUnique))
	}
	if in.Priority != nil {
		predicates = append(predicates, attributemappingrule.PriorityEQ(int(*in.Priority)))
	}
	if in.Enabled != nil {
		predicates = append(predicates, attributemappingrule.EnabledEQ(*in.Enabled))
	}
	if in.UpdateStrategy != nil {
		predicates = append(predicates, attributemappingrule.UpdateStrategyContains(*in.UpdateStrategy))
	}
	if in.SuccessCount != nil {
		predicates = append(predicates, attributemappingrule.SuccessCountEQ(*in.SuccessCount))
	}
	if in.FailedCount != nil {
		predicates = append(predicates, attributemappingrule.FailedCountEQ(*in.FailedCount))
	}
	if in.LastSuccessAt != nil {
		predicates = append(predicates, attributemappingrule.LastSuccessAtGTE(time.UnixMilli(*in.LastSuccessAt)))
	}
	if in.Description != nil {
		predicates = append(predicates, attributemappingrule.DescriptionContains(*in.Description))
	}
	// Skip Metadata query as it's a JSON field
	result, err := l.svcCtx.DB.AttributeMappingRule.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.AttributeMappingRuleListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.AttributeMappingRuleInfo{
			Id:          &v.ID,
			CreatedAt:   pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:   pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:	pointy.GetPointer(uint32(v.Status)),
			DiscoveryConfigId:	&v.DiscoveryConfigID,
			CiAttributeId:	&v.CiAttributeID,
			SourceField:	&v.SourceField,
			SourcePath:	&v.SourcePath,
			TargetAttribute:	&v.TargetAttribute,
			TransformType:	&v.TransformType,
			TransformConfig:	jsonhelper.JSONToString(v.TransformConfig),
			DefaultValue:	&v.DefaultValue,
			ValidationRules:	jsonhelper.JSONToString(v.ValidationRules),
			ValidationRegex:	&v.ValidationRegex,
			IsRequired:	&v.IsRequired,
			IsUnique:	&v.IsUnique,
			Priority:	pointy.GetPointer(int64(v.Priority)),
			Enabled:	&v.Enabled,
			UpdateStrategy:	&v.UpdateStrategy,
			SuccessCount:	&v.SuccessCount,
			FailedCount:	&v.FailedCount,
			LastSuccessAt:	pointy.GetUnixMilliPointer(v.LastSuccessAt.UnixMilli()),
			Description:	&v.Description,
			Metadata:	jsonhelper.JSONToString(v.Metadata),
		})
	}

	return resp, nil
}
