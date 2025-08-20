package cipermission

import (
	"context"
	"fmt"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/gofrs/uuid/v5"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiPermissionByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiPermissionByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiPermissionByIdLogic {
	return &GetCiPermissionByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiPermissionByIdLogic) GetCiPermissionById(in *cmdb.IDReq) (*cmdb.CiPermissionInfo, error) {
	l.Logger.Infof("开始查询CI权限，ID: %d", in.Id)

	// 1. 数据验证
	if err := l.validateRequest(in); err != nil {
		l.Logger.Errorf("查询CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 查询权限
	permission, err := l.svcCtx.DB.CiPermission.Get(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("查询CI权限失败，ID: %d, 错误: %v", in.Id, err)
		return nil, fmt.Errorf("权限不存在或查询失败: %v", err)
	}

	// 3. 转换数据格式
	result, err := l.convertToResponse(permission)
	if err != nil {
		l.Logger.Errorf("转换权限数据格式失败: %v", err)
		return nil, fmt.Errorf("数据转换失败: %v", err)
	}

	l.Logger.Infof("成功查询CI权限，ID: %d, 权限ID: %s", permission.ID, permission.PermissionID)

	return result, nil
}

// validateRequest 验证查询请求
func (l *GetCiPermissionByIdLogic) validateRequest(in *cmdb.IDReq) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	if in.Id == 0 {
		return fmt.Errorf("权限ID不能为0")
	}

	return nil
}

// convertToResponse 将数据库实体转换为响应格式
func (l *GetCiPermissionByIdLogic) convertToResponse(permission *ent.CiPermission) (*cmdb.CiPermissionInfo, error) {
	var zeroUUID uuid.UUID

	result := &cmdb.CiPermissionInfo{
		Id:        pointy.GetPointer(permission.ID),
		CreatedAt: pointy.GetPointer(permission.CreatedAt.Unix()),
		UpdatedAt: pointy.GetPointer(permission.UpdatedAt.Unix()),
	}

	// 基本字段
	if permission.DepartmentID != 0 {
		result.DepartmentId = pointy.GetPointer(permission.DepartmentID)
	}
	result.PermissionId = pointy.GetPointer(permission.PermissionID)
	result.ScopeType = pointy.GetPointer(string(permission.ScopeType))

	// 权限范围相关字段
	if permission.CiTypeID != 0 {
		result.CiTypeId = pointy.GetPointer(permission.CiTypeID)
	}
	if permission.CiID != 0 {
		result.CiId = pointy.GetPointer(permission.CiID)
	}
	if permission.AttributeID != 0 {
		result.AttributeId = pointy.GetPointer(permission.AttributeID)
	}
	if permission.FieldName != "" {
		result.FieldName = pointy.GetPointer(permission.FieldName)
	}

	// 权限主体字段
	result.SubjectType = pointy.GetPointer(string(permission.SubjectType))
	if permission.SubjectID != zeroUUID {
		result.SubjectId = pointy.GetPointer(permission.SubjectID.String())
	}
	result.SubjectName = pointy.GetPointer(permission.SubjectName)
	if permission.SubjectCode != "" {
		result.SubjectCode = pointy.GetPointer(permission.SubjectCode)
	}

	// 权限类型和操作
	result.PermissionType = pointy.GetPointer(string(permission.PermissionType))

	// 转换操作列表
	if len(permission.Operations) > 0 {
		operations := &cmdb.CiPermissionOperations{
			Operations: make([]*cmdb.CiPermissionOperation, 0, len(permission.Operations)),
		}
		for _, op := range permission.Operations {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   op,
				Description: "",
			})
		}
		result.Operations = operations
	}

	// 转换条件
	if permission.Conditions != nil {
		conditions := &cmdb.CiPermissionConditions{}
		if field, ok := permission.Conditions["field"].(string); ok {
			conditions.Field = field
		}
		if operator, ok := permission.Conditions["operator"].(string); ok {
			conditions.Operator = operator
		}
		if value, ok := permission.Conditions["value"].(string); ok {
			conditions.Value = value
		}
		if logic, ok := permission.Conditions["logic"].(string); ok {
			conditions.Logic = logic
		}
		result.Conditions = conditions
	}

	// 权限级别和优先级
	result.Priority = pointy.GetPointer(int64(permission.Priority))
	result.PermissionLevel = pointy.GetPointer(string(permission.PermissionLevel))

	// 时间限制
	if !permission.EffectiveFrom.IsZero() {
		result.EffectiveFrom = pointy.GetPointer(permission.EffectiveFrom.Unix())
	}
	if !permission.EffectiveTo.IsZero() {
		result.EffectiveTo = pointy.GetPointer(permission.EffectiveTo.Unix())
	}
	result.IsTemporary = pointy.GetPointer(permission.IsTemporary)

	// 数据过滤和访问控制
	if permission.DataFilters != nil {
		dataFilters := &cmdb.CiPermissionDataFilters{}
		if logic, ok := permission.DataFilters["logic"].(string); ok {
			dataFilters.Logic = logic
		}
		// 这里可以进一步解析rules，现在简化处理
		result.DataFilters = dataFilters
	}

	if permission.FieldMasks != nil && len(permission.FieldMasks) > 0 {
		fieldMasks := &cmdb.CiPermissionFieldMasks{
			Fields: permission.FieldMasks,
		}
		result.FieldMasks = fieldMasks
	}

	if permission.AllowedValues != nil {
		allowedValues := &cmdb.CiPermissionAllowedValues{
			FieldValues: make(map[string]string),
		}
		for k, v := range permission.AllowedValues {
			if valueStr, ok := v.(string); ok {
				allowedValues.FieldValues[k] = valueStr
			}
		}
		result.AllowedValues = allowedValues
	}

	// 审批和授权信息
	result.RequireApproval = pointy.GetPointer(permission.RequireApproval)
	if permission.GrantedBy != zeroUUID {
		result.GrantedBy = pointy.GetPointer(permission.GrantedBy.String())
	}
	if permission.GrantedByName != "" {
		result.GrantedByName = pointy.GetPointer(permission.GrantedByName)
	}
	if !permission.GrantedAt.IsZero() {
		result.GrantedAt = pointy.GetPointer(permission.GrantedAt.Unix())
	}
	if permission.GrantReason != "" {
		result.GrantReason = pointy.GetPointer(permission.GrantReason)
	}

	// 使用情况统计
	result.UsageCount = pointy.GetPointer(int64(permission.UsageCount))
	if !permission.LastUsedAt.IsZero() {
		result.LastUsedAt = pointy.GetPointer(permission.LastUsedAt.Unix())
	}

	if permission.UsageStatistics != nil {
		usageStats := &cmdb.CiPermissionUsageStatistics{}
		if dailyUsage, ok := permission.UsageStatistics["daily_usage"].(map[string]interface{}); ok {
			usageStats.DailyUsage = make(map[string]int64)
			for k, v := range dailyUsage {
				if count, ok := v.(float64); ok {
					usageStats.DailyUsage[k] = int64(count)
				}
			}
		}
		result.UsageStatistics = usageStats
	}

	// 权限状态
	result.Status = pointy.GetPointer(uint32(0)) // 简化处理，需要根据实际状态枚举转换
	if permission.StatusReason != "" {
		result.StatusReason = pointy.GetPointer(permission.StatusReason)
	}

	// 权限继承和传播
	result.Inheritable = pointy.GetPointer(permission.Inheritable)
	if permission.ParentPermissionID != "" {
		result.ParentPermissionId = pointy.GetPointer(permission.ParentPermissionID)
	}

	if permission.InheritedFrom != nil && len(permission.InheritedFrom) > 0 {
		inheritedFrom := &cmdb.CiPermissionInheritedFrom{
			PermissionIds: permission.InheritedFrom,
		}
		result.InheritedFrom = inheritedFrom
	}

	// 风险等级和安全控制
	result.RiskLevel = pointy.GetPointer(string(permission.RiskLevel))
	result.RequireMfa = pointy.GetPointer(permission.RequireMfa)

	if permission.SecurityConstraints != nil {
		securityConstraints := &cmdb.CiPermissionSecurityConstraints{}
		if requireVpn, ok := permission.SecurityConstraints["require_vpn"].(bool); ok {
			securityConstraints.RequireVpn = requireVpn
		}
		// 可以进一步解析其他安全约束字段
		result.SecurityConstraints = securityConstraints
	}

	// 扩展信息
	if permission.Metadata != nil {
		metadata := &cmdb.CiPermissionMetadata{}
		if businessOwner, ok := permission.Metadata["business_owner"].(string); ok {
			metadata.BusinessOwner = businessOwner
		}
		if technicalOwner, ok := permission.Metadata["technical_owner"].(string); ok {
			metadata.TechnicalOwner = technicalOwner
		}
		result.Metadata = metadata
	}

	if len(permission.Tags) > 0 {
		tags := make([]*cmdb.CiPermissionTag, 0, len(permission.Tags))
		for _, tagStr := range permission.Tags {
			// 简化处理，假设格式为 "key:value"
			if len(tagStr) > 0 {
				parts := []string{tagStr, ""}
				if len(parts) >= 2 {
					tags = append(tags, &cmdb.CiPermissionTag{
						Key:   parts[0],
						Value: parts[1],
					})
				}
			}
		}
		result.Tags = tags
	}

	if permission.Description != "" {
		result.Description = pointy.GetPointer(permission.Description)
	}
	if permission.Comments != "" {
		result.Comments = pointy.GetPointer(permission.Comments)
	}

	// 审计字段
	if permission.CreatedBy != zeroUUID {
		result.CreatedBy = pointy.GetPointer(permission.CreatedBy.String())
	}
	if permission.UpdatedBy != zeroUUID {
		result.UpdatedBy = pointy.GetPointer(permission.UpdatedBy.String())
	}
	if !permission.LastReviewedAt.IsZero() {
		result.LastReviewedAt = pointy.GetPointer(permission.LastReviewedAt.Unix())
	}
	if permission.LastReviewedBy != zeroUUID {
		result.LastReviewedBy = pointy.GetPointer(permission.LastReviewedBy.String())
	}

	return result, nil
}
