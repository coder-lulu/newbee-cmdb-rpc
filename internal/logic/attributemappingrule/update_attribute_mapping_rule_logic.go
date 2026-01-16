package attributemappingrule

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

    "github.com/suyuan32/simple-admin-common/msg/errormsg"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAttributeMappingRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAttributeMappingRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAttributeMappingRuleLogic {
	return &UpdateAttributeMappingRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAttributeMappingRuleLogic) UpdateAttributeMappingRule(in *cmdb.AttributeMappingRuleInfo) (*cmdb.BaseResp, error) {
	query := l.svcCtx.DB.AttributeMappingRule.UpdateOneID(*in.Id).
		SetNotNilDiscoveryConfigID(in.DiscoveryConfigId).
		SetNotNilCiAttributeID(in.CiAttributeId).
		SetNotNilSourceField(in.SourceField).
		SetNotNilSourcePath(in.SourcePath).
		SetNotNilTargetAttribute(in.TargetAttribute).
		SetNotNilTransformType(in.TransformType).
		SetNotNilDefaultValue(in.DefaultValue).
		SetNotNilValidationRegex(in.ValidationRegex).
		SetNotNilIsRequired(in.IsRequired).
		SetNotNilIsUnique(in.IsUnique).
		SetNotNilEnabled(in.Enabled).
		SetNotNilUpdateStrategy(in.UpdateStrategy).
		SetNotNilSuccessCount(in.SuccessCount).
		SetNotNilFailedCount(in.FailedCount).
		SetNotNilLastSuccessAt(pointy.GetTimeMilliPointer(in.LastSuccessAt)).
		SetNotNilDescription(in.Description)

	// Handle JSON fields
	if in.TransformConfig != nil {
		transformConfig := jsonhelper.StringToJSONMap(in.TransformConfig)
		if transformConfig != nil {
			query.SetTransformConfig(transformConfig)
		}
	}
	
	if in.ValidationRules != nil {
		validationRules := jsonhelper.StringToJSONMap(in.ValidationRules)
		if validationRules != nil {
			query.SetValidationRules(validationRules)
		}
	}
	
	if in.Metadata != nil {
		metadata := jsonhelper.StringToJSONMap(in.Metadata)
		if metadata != nil {
			query.SetMetadata(metadata)
		}
	}

	if in.Status != nil {
		query.SetNotNilStatus(pointy.GetPointer(uint8(*in.Status)))
	}
	if in.Priority != nil {
		query.SetNotNilPriority(pointy.GetPointer(int(*in.Priority)))
	}

	err := query.Exec(l.ctx)

    if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

    return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess }, nil
}
