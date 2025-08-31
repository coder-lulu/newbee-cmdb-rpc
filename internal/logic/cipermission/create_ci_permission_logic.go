package cipermission

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiPermissionLogic {
	return &CreateCiPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CiPermission management
func (l *CreateCiPermissionLogic) CreateCiPermission(in *cmdb.CiPermissionInfo) (*cmdb.BaseIDResp, error) {
	l.Logger.Infof("开始创建CI权限，权限ID: %s, 主体类型: %s",
		pointy.GetPointer(in.PermissionId), pointy.GetPointer(in.SubjectType))

	// 1. 数据验证
	if err := l.validateCreateRequest(in); err != nil {
		l.Logger.Errorf("创建CI权限数据验证失败: %v", err)
		return nil, err
	}

	// 2. 检查权限ID唯一性
	if in.PermissionId != nil {
		exists, err := l.svcCtx.DB.CiPermission.Query().
			Where(cipermission.PermissionIDEQ(*in.PermissionId)).
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

	// 3. 验证业务规则
	if err := l.validateBusinessRules(in); err != nil {
		l.Logger.Errorf("业务规则验证失败: %v", err)
		return nil, err
	}

	// 4. 构建创建请求
	createBuilder := l.svcCtx.DB.CiPermission.Create()

	// 设置基本字段
	if in.PermissionId != nil {
		createBuilder.SetPermissionID(*in.PermissionId)
	} else {
		// 自动生成权限ID
		permissionID := fmt.Sprintf("perm_%d_%s", time.Now().Unix(), uuid.Must(uuid.NewV4()).String()[:8])
		createBuilder.SetPermissionID(permissionID)
	}

	if in.DepartmentId != nil {
		createBuilder.SetDepartmentID(*in.DepartmentId)
	}

	// 权限范围设置
	if in.ScopeType != nil {
		createBuilder.SetScopeType(cipermission.ScopeType(*in.ScopeType))
	}
	if in.CiTypeId != nil {
		createBuilder.SetScopeTargetType("ci_type_id")
		createBuilder.SetNillableScopeTargetID(in.CiTypeId)
	}
	if in.CiId != nil {
		createBuilder.SetScopeTargetType("ci_id")
		createBuilder.SetNillableScopeTargetID(in.CiId)
	}
	if in.AttributeId != nil {
		createBuilder.SetScopeTargetType("attribute_id")
		createBuilder.SetNillableScopeTargetID(in.AttributeId)
	}
	if in.FieldName != nil {
		createBuilder.SetNillableScopeFieldName(in.FieldName)
	}

	// 权限主体设置
	if in.SubjectType != nil {
		createBuilder.SetSubjectType(cipermission.SubjectType(*in.SubjectType))
	}
	if in.SubjectId != nil {
		createBuilder.SetNillableSubjectID(in.SubjectId)
	}
	if in.SubjectName != nil {
		createBuilder.SetSubjectName(*in.SubjectName)
	}

	// 权限类型和操作
	if in.PermissionType != nil {
		createBuilder.SetPermissionType(cipermission.PermissionType(*in.PermissionType))
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
		createBuilder.SetOperationsMask(operationsMask)
	}

	// 权限级别和优先级
	if in.Priority != nil {
		createBuilder.SetPriority(int(*in.Priority))
	}
	if in.PermissionLevel != nil {
		createBuilder.SetPermissionLevel(cipermission.PermissionLevel(*in.PermissionLevel))
	}

	// 时间限制
	if in.EffectiveFrom != nil {
		createBuilder.SetNillableEffectiveFrom(pointy.GetPointer(time.Unix(*in.EffectiveFrom, 0)))
	}
	if in.EffectiveTo != nil {
		createBuilder.SetNillableEffectiveTo(pointy.GetPointer(time.Unix(*in.EffectiveTo, 0)))
	}
	if in.IsTemporary != nil {
		createBuilder.SetIsTemporary(*in.IsTemporary)
	}

	// 注意：数据过滤和字段掩码现在通过边关系处理，不直接设置字段
	// 这部分逻辑将在关联表中处理

	// 审批信息
	if in.RequireApproval != nil {
		createBuilder.SetRequireApproval(*in.RequireApproval)
	}

	// 使用情况统计
	if in.UsageCount != nil {
		createBuilder.SetUsageCount(int(*in.UsageCount))
	}
	if in.LastUsedAt != nil {
		createBuilder.SetNillableLastUsedAt(pointy.GetPointer(time.Unix(*in.LastUsedAt, 0)))
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
		createBuilder.SetStatus(cipermission.Status(statusStr))
	}

	// 权限继承
	if in.Inheritable != nil {
		createBuilder.SetInheritable(*in.Inheritable)
	}
	if in.ParentPermissionId != nil {
		createBuilder.SetNillableParentPermissionID(in.ParentPermissionId)
	}

	// 风险等级和安全控制
	if in.RiskLevel != nil {
		createBuilder.SetRiskLevel(cipermission.RiskLevel(*in.RiskLevel))
	}
	if in.RequireMfa != nil {
		createBuilder.SetRequireMfa(*in.RequireMfa)
	}

	// 扩展信息
	if in.Description != nil {
		createBuilder.SetNillableDescription(in.Description)
	}
	if in.Comments != nil {
		createBuilder.SetNillableComments(in.Comments)
	}

	// 审计字段
	if in.CreatedBy != nil {
		createBuilder.SetNillableCreatedBy(in.CreatedBy)
	}
	if in.UpdatedBy != nil {
		createBuilder.SetNillableUpdatedBy(in.UpdatedBy)
	}

	// 5. 执行创建操作
	result, err := createBuilder.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("创建CI权限时数据库操作失败: %v", err)
		return nil, fmt.Errorf("创建CI权限失败: %v", err)
	}

	l.Logger.Infof("成功创建CI权限，ID: %d, 权限ID: %s", result.ID, result.PermissionID)

	return &cmdb.BaseIDResp{
		Id:  result.ID,
		Msg: "权限创建成功",
	}, nil
}

// validateCreateRequest 验证创建请求的基本数据
func (l *CreateCiPermissionLogic) validateCreateRequest(in *cmdb.CiPermissionInfo) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	// 验证必填字段
	if in.SubjectType == nil || *in.SubjectType == "" {
		return fmt.Errorf("权限主体类型不能为空")
	}

	if in.SubjectName == nil || *in.SubjectName == "" {
		return fmt.Errorf("权限主体名称不能为空")
	}

	if in.ScopeType == nil || *in.ScopeType == "" {
		return fmt.Errorf("权限范围类型不能为空")
	}

	// 验证枚举值
	validSubjectTypes := map[string]bool{
		"user": true, "role": true, "department": true, "group": true, "system": true,
	}
	if !validSubjectTypes[*in.SubjectType] {
		return fmt.Errorf("无效的权限主体类型: %s", *in.SubjectType)
	}

	validScopeTypes := map[string]bool{
		"global": true, "ci_type": true, "ci_instance": true, "attribute": true, "field": true,
	}
	if !validScopeTypes[*in.ScopeType] {
		return fmt.Errorf("无效的权限范围类型: %s", *in.ScopeType)
	}

	return nil
}

// validateBusinessRules 验证业务规则
func (l *CreateCiPermissionLogic) validateBusinessRules(in *cmdb.CiPermissionInfo) error {
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
