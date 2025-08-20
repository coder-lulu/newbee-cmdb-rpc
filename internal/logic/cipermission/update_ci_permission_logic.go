package cipermission

import (
	"context"
	"fmt"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/cipermission"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCiPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCiPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCiPermissionLogic {
	return &UpdateCiPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCiPermissionLogic) UpdateCiPermission(in *cmdb.CiPermissionInfo) (*cmdb.BaseResp, error) {
	l.Logger.Infof("开始更新CI权限，ID: %d", pointy.GetPointer(in.Id))

	// 1. 数据验证
	if err := l.validateUpdateRequest(in); err != nil {
		l.Logger.Errorf("更新CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 检查权限是否存在
	existingPermission, err := l.svcCtx.DB.CiPermission.Get(l.ctx, *in.Id)
	if err != nil {
		l.Logger.Errorf("查询要更新的权限失败，ID: %d, 错误: %v", *in.Id, err)
		return nil, fmt.Errorf("权限不存在或查询失败: %v", err)
	}

	// 3. 检查权限ID唯一性（如果要更新权限ID）
	if in.PermissionId != nil && existingPermission.PermissionID != *in.PermissionId {
		exists, err := l.svcCtx.DB.CiPermission.Query().
			Where(cipermission.PermissionIDEQ(*in.PermissionId)).
			Where(cipermission.IDNEQ(*in.Id)).
			Exist(l.ctx)
		if err != nil {
			l.Logger.Errorf("检查权限ID唯一性时数据库查询失败: %v", err)
			return nil, fmt.Errorf("检查权限ID唯一性失败: %v", err)
		}
		if exists {
			err := fmt.Errorf("权限ID [%s] 已存在", *in.PermissionId)
			l.Logger.Errorf("权限ID重复: %v", err)
			return nil, err
		}
	}

	// 4. 验证业务规则
	if err := l.validateBusinessRules(in); err != nil {
		l.Logger.Errorf("业务规则验证失败: %v", err)
		return nil, err
	}

	// 5. 构建更新请求
	updateBuilder := l.svcCtx.DB.CiPermission.UpdateOneID(*in.Id)

	// 设置基本字段
	if in.PermissionId != nil {
		updateBuilder.SetPermissionID(*in.PermissionId)
	}
	if in.DepartmentId != nil {
		updateBuilder.SetDepartmentID(*in.DepartmentId)
	}

	// 权限范围设置
	if in.ScopeType != nil {
		updateBuilder.SetScopeType(cipermission.ScopeType(*in.ScopeType))
	}
	if in.CiTypeId != nil {
		updateBuilder.SetNillableCiTypeID(in.CiTypeId)
	}
	if in.CiId != nil {
		updateBuilder.SetNillableCiID(in.CiId)
	}
	if in.AttributeId != nil {
		updateBuilder.SetNillableAttributeID(in.AttributeId)
	}
	if in.FieldName != nil {
		updateBuilder.SetNillableFieldName(in.FieldName)
	}

	// 权限主体设置
	if in.SubjectType != nil {
		updateBuilder.SetSubjectType(cipermission.SubjectType(*in.SubjectType))
	}
	if in.SubjectId != nil {
		if subjectUUID, err := uuid.FromString(*in.SubjectId); err == nil {
			updateBuilder.SetNillableSubjectID(&subjectUUID)
		}
	}
	if in.SubjectName != nil {
		updateBuilder.SetSubjectName(*in.SubjectName)
	}
	if in.SubjectCode != nil {
		updateBuilder.SetNillableSubjectCode(in.SubjectCode)
	}

	// 权限类型和操作
	if in.PermissionType != nil {
		updateBuilder.SetPermissionType(cipermission.PermissionType(*in.PermissionType))
	}
	if in.Operations != nil {
		operationStrings := make([]string, 0)
		for _, op := range in.Operations.Operations {
			operationStrings = append(operationStrings, op.Operation)
		}
		updateBuilder.SetOperations(operationStrings)
	}
	if in.Conditions != nil {
		conditionsMap := map[string]interface{}{
			"field":    in.Conditions.Field,
			"operator": in.Conditions.Operator,
			"value":    in.Conditions.Value,
			"logic":    in.Conditions.Logic,
		}
		updateBuilder.SetConditions(conditionsMap)
	}

	// 权限级别和优先级
	if in.Priority != nil {
		updateBuilder.SetPriority(int(*in.Priority))
	}
	if in.PermissionLevel != nil {
		updateBuilder.SetPermissionLevel(cipermission.PermissionLevel(*in.PermissionLevel))
	}

	// 时间限制
	if in.EffectiveFrom != nil {
		updateBuilder.SetNillableEffectiveFrom(pointy.GetPointer(time.Unix(*in.EffectiveFrom, 0)))
	}
	if in.EffectiveTo != nil {
		updateBuilder.SetNillableEffectiveTo(pointy.GetPointer(time.Unix(*in.EffectiveTo, 0)))
	}
	if in.IsTemporary != nil {
		updateBuilder.SetIsTemporary(*in.IsTemporary)
	}

	// 数据过滤和访问控制
	if in.DataFilters != nil {
		dataFiltersMap := map[string]interface{}{
			"rules": in.DataFilters.Rules,
			"logic": in.DataFilters.Logic,
		}
		updateBuilder.SetDataFilters(dataFiltersMap)
	}
	if in.FieldMasks != nil {
		updateBuilder.SetFieldMasks(in.FieldMasks.Fields)
	}
	if in.AllowedValues != nil {
		allowedValuesMap := map[string]interface{}{}
		for k, v := range in.AllowedValues.FieldValues {
			allowedValuesMap[k] = v
		}
		updateBuilder.SetAllowedValues(allowedValuesMap)
	}

	// 审批和授权信息
	if in.RequireApproval != nil {
		updateBuilder.SetRequireApproval(*in.RequireApproval)
	}
	if in.GrantedBy != nil {
		if grantedByUUID, err := uuid.FromString(*in.GrantedBy); err == nil {
			updateBuilder.SetNillableGrantedBy(&grantedByUUID)
		}
	}
	if in.GrantedByName != nil {
		updateBuilder.SetNillableGrantedByName(in.GrantedByName)
	}
	if in.GrantedAt != nil {
		updateBuilder.SetNillableGrantedAt(pointy.GetPointer(time.Unix(*in.GrantedAt, 0)))
	}
	if in.GrantReason != nil {
		updateBuilder.SetNillableGrantReason(in.GrantReason)
	}

	// 使用情况统计
	if in.UsageCount != nil {
		updateBuilder.SetUsageCount(int(*in.UsageCount))
	}
	if in.LastUsedAt != nil {
		updateBuilder.SetNillableLastUsedAt(pointy.GetPointer(time.Unix(*in.LastUsedAt, 0)))
	}
	if in.UsageStatistics != nil {
		usageStatsMap := map[string]interface{}{
			"daily_usage":    in.UsageStatistics.DailyUsage,
			"weekly_usage":   in.UsageStatistics.WeeklyUsage,
			"monthly_usage":  in.UsageStatistics.MonthlyUsage,
			"top_operations": in.UsageStatistics.TopOperations,
		}
		updateBuilder.SetUsageStatistics(usageStatsMap)
	}

	// 权限状态
	if in.Status != nil {
		updateBuilder.SetStatus(cipermission.Status(fmt.Sprintf("%d", *in.Status)))
	}
	if in.StatusReason != nil {
		updateBuilder.SetNillableStatusReason(in.StatusReason)
	}

	// 权限继承和传播
	if in.Inheritable != nil {
		updateBuilder.SetInheritable(*in.Inheritable)
	}
	if in.ParentPermissionId != nil {
		updateBuilder.SetNillableParentPermissionID(in.ParentPermissionId)
	}
	if in.InheritedFrom != nil {
		updateBuilder.SetInheritedFrom(in.InheritedFrom.PermissionIds)
	}

	// 风险等级和安全控制
	if in.RiskLevel != nil {
		updateBuilder.SetRiskLevel(cipermission.RiskLevel(*in.RiskLevel))
	}
	if in.RequireMfa != nil {
		updateBuilder.SetRequireMfa(*in.RequireMfa)
	}
	if in.SecurityConstraints != nil {
		securityConstraintsMap := map[string]interface{}{
			"require_vpn":           in.SecurityConstraints.RequireVpn,
			"allowed_ips":           in.SecurityConstraints.AllowedIps,
			"blocked_ips":           in.SecurityConstraints.BlockedIps,
			"time_restrictions":     in.SecurityConstraints.TimeRestrictions,
			"location_restrictions": in.SecurityConstraints.LocationRestrictions,
		}
		updateBuilder.SetSecurityConstraints(securityConstraintsMap)
	}

	// 扩展信息
	if in.Metadata != nil {
		metadataMap := map[string]interface{}{
			"business_owner":  in.Metadata.BusinessOwner,
			"technical_owner": in.Metadata.TechnicalOwner,
			"data_class":      in.Metadata.DataClass,
			"compliance_reqs": in.Metadata.ComplianceReqs,
			"custom_fields":   in.Metadata.CustomFields,
		}
		updateBuilder.SetMetadata(metadataMap)
	}
	if in.Tags != nil {
		tagStrings := make([]string, len(in.Tags))
		for i, tag := range in.Tags {
			tagStrings[i] = fmt.Sprintf("%s:%s", tag.Key, tag.Value)
		}
		updateBuilder.SetTags(tagStrings)
	}
	if in.Description != nil {
		updateBuilder.SetNillableDescription(in.Description)
	}
	if in.Comments != nil {
		updateBuilder.SetNillableComments(in.Comments)
	}

	// 审计字段
	if in.UpdatedBy != nil {
		if updatedByUUID, err := uuid.FromString(*in.UpdatedBy); err == nil {
			updateBuilder.SetNillableUpdatedBy(&updatedByUUID)
		}
	}
	if in.LastReviewedAt != nil {
		updateBuilder.SetNillableLastReviewedAt(pointy.GetPointer(time.Unix(*in.LastReviewedAt, 0)))
	}
	if in.LastReviewedBy != nil {
		if lastReviewedByUUID, err := uuid.FromString(*in.LastReviewedBy); err == nil {
			updateBuilder.SetNillableLastReviewedBy(&lastReviewedByUUID)
		}
	}

	// 6. 执行更新操作
	result, err := updateBuilder.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("更新CI权限时数据库操作失败: %v", err)
		return nil, fmt.Errorf("更新CI权限失败: %v", err)
	}

	l.Logger.Infof("成功更新CI权限，ID: %d, 权限ID: %s", result.ID, result.PermissionID)

	return &cmdb.BaseResp{
		Msg: "权限更新成功",
	}, nil
}

// validateUpdateRequest 验证更新请求的基本数据
func (l *UpdateCiPermissionLogic) validateUpdateRequest(in *cmdb.CiPermissionInfo) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	if in.Id == nil {
		return fmt.Errorf("权限ID不能为空")
	}

	return nil
}

// validateBusinessRules 验证业务规则
func (l *UpdateCiPermissionLogic) validateBusinessRules(in *cmdb.CiPermissionInfo) error {
	// 根据权限范围类型验证相关字段
	if in.ScopeType != nil {
		switch *in.ScopeType {
		case "ci_type":
			if in.CiTypeId == nil {
				return fmt.Errorf("权限范围为ci_type时，CI类型ID不能为空")
			}
		case "ci_instance":
			if in.CiId == nil {
				return fmt.Errorf("权限范围为ci_instance时，CI实例ID不能为空")
			}
		case "attribute":
			if in.AttributeId == nil {
				return fmt.Errorf("权限范围为attribute时，属性ID不能为空")
			}
		case "field":
			if in.FieldName == nil || *in.FieldName == "" {
				return fmt.Errorf("权限范围为field时，字段名称不能为空")
			}
		}
	}

	// 验证时间范围
	if in.EffectiveFrom != nil && in.EffectiveTo != nil {
		if *in.EffectiveFrom >= *in.EffectiveTo {
			return fmt.Errorf("权限生效开始时间不能晚于或等于结束时间")
		}
	}

	// 验证权限级别
	if in.PermissionLevel != nil {
		validLevels := map[string]bool{
			"none": true, "read": true, "write": true, "admin": true, "super_admin": true,
		}
		if !validLevels[*in.PermissionLevel] {
			return fmt.Errorf("无效的权限级别: %s", *in.PermissionLevel)
		}
	}

	// 验证权限类型
	if in.PermissionType != nil {
		validTypes := map[string]bool{"allow": true, "deny": true}
		if !validTypes[*in.PermissionType] {
			return fmt.Errorf("无效的权限类型: %s", *in.PermissionType)
		}
	}

	// 验证风险等级
	if in.RiskLevel != nil {
		validRiskLevels := map[string]bool{
			"low": true, "medium": true, "high": true, "critical": true,
		}
		if !validRiskLevels[*in.RiskLevel] {
			return fmt.Errorf("无效的风险等级: %s", *in.RiskLevel)
		}
	}

	return nil
}
