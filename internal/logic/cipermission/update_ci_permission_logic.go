package cipermission

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

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
	l.Logger.Infof("开始更新CI权限，ID: %d, 权限ID: %s", 
		pointy.GetPointer(in.Id), pointy.GetPointer(in.PermissionId))

	// 1. 数据验证
	if err := l.validateUpdateRequest(in); err != nil {
		l.Logger.Errorf("更新CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 检查权限是否存在
	if in.Id == nil {
		return nil, fmt.Errorf("权限ID不能为空")
	}

	exists, err := l.svcCtx.DB.CiPermission.Query().
		Where(cipermission.IDEQ(*in.Id)).
		Exist(l.ctx)
	if err != nil {
		l.Logger.Errorf("检查权限存在性时数据库查询失败: %v", err)
		return nil, fmt.Errorf("检查权限存在性失败: %v", err)
	}
	if !exists {
		err := fmt.Errorf("权限ID [%d] 不存在", *in.Id)
		l.Logger.Errorf("权限不存在: %v", err)
		return nil, err
	}

	// 3. 验证业务规则
	if err := l.validateBusinessRules(in); err != nil {
		l.Logger.Errorf("业务规则验证失败: %v", err)
		return nil, err
	}

	// 4. 构建更新请求
	updateBuilder := l.svcCtx.DB.CiPermission.UpdateOneID(*in.Id)

	// 权限范围设置
	if in.ScopeType != nil {
		updateBuilder.SetScopeType(cipermission.ScopeType(*in.ScopeType))
	}
	if in.CiTypeId != nil {
		updateBuilder.SetScopeTargetType("ci_type_id")
		updateBuilder.SetNillableScopeTargetID(in.CiTypeId)
	}
	if in.CiId != nil {
		updateBuilder.SetScopeTargetType("ci_id")
		updateBuilder.SetNillableScopeTargetID(in.CiId)
	}
	if in.AttributeId != nil {
		updateBuilder.SetScopeTargetType("attribute_id")
		updateBuilder.SetNillableScopeTargetID(in.AttributeId)
	}
	if in.FieldName != nil {
		updateBuilder.SetNillableScopeFieldName(in.FieldName)
	}

	// 权限主体设置
	if in.SubjectType != nil {
		updateBuilder.SetSubjectType(cipermission.SubjectType(*in.SubjectType))
	}
	if in.SubjectId != nil {
		updateBuilder.SetNillableSubjectID(in.SubjectId)
	}
	if in.SubjectName != nil {
		updateBuilder.SetSubjectName(*in.SubjectName)
	}

	// 权限类型和操作
	if in.PermissionType != nil {
		updateBuilder.SetPermissionType(cipermission.PermissionType(*in.PermissionType))
	}
	if in.Operations != nil {
		// 将操作转换为位掩码
		var operationsMask uint64 = 0
		for _, op := range in.Operations.Operations {
			switch op.Operation {
			case "read":
				operationsMask |= 1
			case "write":
				operationsMask |= 2
			case "delete":
				operationsMask |= 4
			case "approve":
				operationsMask |= 8
			case "export":
				operationsMask |= 16
			}
		}
		updateBuilder.SetOperationsMask(operationsMask)
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

	// 审批信息
	if in.RequireApproval != nil {
		updateBuilder.SetRequireApproval(*in.RequireApproval)
	}

	// 权限状态
	if in.Status != nil {
		// 将数字状态转换为字符串枚举
		var statusStr string
		switch *in.Status {
		case 1:
			statusStr = "active"
		case 0:
			statusStr = "inactive"
		case 2:
			statusStr = "suspended"
		case 3:
			statusStr = "revoked"
		case 4:
			statusStr = "expired"
		default:
			statusStr = "active"
		}
		updateBuilder.SetStatus(cipermission.Status(statusStr))
	}

	// 权限继承
	if in.Inheritable != nil {
		updateBuilder.SetInheritable(*in.Inheritable)
	}
	if in.ParentPermissionId != nil {
		updateBuilder.SetNillableParentPermissionID(in.ParentPermissionId)
	}

	// 风险等级和安全控制
	if in.RiskLevel != nil {
		updateBuilder.SetRiskLevel(cipermission.RiskLevel(*in.RiskLevel))
	}
	if in.RequireMfa != nil {
		updateBuilder.SetRequireMfa(*in.RequireMfa)
	}

	// 扩展信息
	if in.Description != nil {
		updateBuilder.SetNillableDescription(in.Description)
	}
	if in.Comments != nil {
		updateBuilder.SetNillableComments(in.Comments)
	}

	// 审计字段
	if in.UpdatedBy != nil {
		updateBuilder.SetNillableUpdatedBy(in.UpdatedBy)
	}

	// 5. 执行更新操作
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

	// ID是更新操作的必填字段
	if in.Id == nil {
		return fmt.Errorf("权限ID不能为空")
	}

	// 验证更新字段的有效性
	if in.SubjectType != nil && *in.SubjectType != "" {
		validSubjectTypes := map[string]bool{
			"user": true, "role": true, "department": true, "group": true, "system": true,
		}
		if !validSubjectTypes[*in.SubjectType] {
			return fmt.Errorf("无效的权限主体类型: %s", *in.SubjectType)
		}
	}

	if in.ScopeType != nil && *in.ScopeType != "" {
		validScopeTypes := map[string]bool{
			"global": true, "ci_type": true, "ci_instance": true, "attribute": true, "field": true,
		}
		if !validScopeTypes[*in.ScopeType] {
			return fmt.Errorf("无效的权限范围类型: %s", *in.ScopeType)
		}
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