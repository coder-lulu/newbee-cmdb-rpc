package attributemappingrule

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/jsonhelper"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttributeMappingRuleByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAttributeMappingRuleByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttributeMappingRuleByIdLogic {
	return &GetAttributeMappingRuleByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAttributeMappingRuleByIdLogic) GetAttributeMappingRuleById(in *cmdb.IDReq) (*cmdb.AttributeMappingRuleInfo, error) {
	result, err := l.svcCtx.DB.AttributeMappingRule.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.AttributeMappingRuleInfo{
		Id:          &result.ID,
		CreatedAt:    pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:    pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:	pointy.GetPointer(uint32(result.Status)),
		DiscoveryConfigId:	&result.DiscoveryConfigID,
		CiAttributeId:	&result.CiAttributeID,
		SourceField:	&result.SourceField,
		SourcePath:	&result.SourcePath,
		TargetAttribute:	&result.TargetAttribute,
		TransformType:	&result.TransformType,
		TransformConfig:	jsonhelper.JSONToString(result.TransformConfig),
		DefaultValue:	&result.DefaultValue,
		ValidationRules:	jsonhelper.JSONToString(result.ValidationRules),
		ValidationRegex:	&result.ValidationRegex,
		IsRequired:	&result.IsRequired,
		IsUnique:	&result.IsUnique,
		Priority:	pointy.GetPointer(int64(result.Priority)),
		Enabled:	&result.Enabled,
		UpdateStrategy:	&result.UpdateStrategy,
		SuccessCount:	&result.SuccessCount,
		FailedCount:	&result.FailedCount,
		LastSuccessAt:	pointy.GetUnixMilliPointer(result.LastSuccessAt.UnixMilli()),
		Description:	&result.Description,
		Metadata:	jsonhelper.JSONToString(result.Metadata),
	}, nil
}

