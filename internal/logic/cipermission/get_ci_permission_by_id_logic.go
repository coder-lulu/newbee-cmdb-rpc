package cipermission

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

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
	l.Logger.Infof("开始查询CI权限详情，ID: %d", in.Id)

	// 1. 数据验证
	if in.Id == 0 {
		return nil, fmt.Errorf("权限ID不能为空")
	}

	// 2. 查询权限记录
	permission, err := l.svcCtx.DB.CiPermission.Query().
		Where(cipermission.IDEQ(in.Id)).
		First(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CI权限时数据库操作失败: %v", err)
		return nil, fmt.Errorf("查询CI权限失败: %v", err)
	}

	l.Logger.Infof("成功查询CI权限，ID: %d, 权限ID: %s", permission.ID, permission.PermissionID)

	// 3. 转换数据格式
	result := &cmdb.CiPermissionInfo{
		Id:                     &permission.ID,
		CreatedAt:              pointy.GetPointer(permission.CreatedAt.Unix()),
		UpdatedAt:              pointy.GetPointer(permission.UpdatedAt.Unix()),
		PermissionId:           &permission.PermissionID,
		DepartmentId:           &permission.DepartmentID,
		ScopeType:              (*string)(&permission.ScopeType),
		FieldName:              pointy.GetPointer(permission.ScopeFieldName),
		SubjectType:            (*string)(&permission.SubjectType),
		SubjectId:              pointy.GetPointer(permission.SubjectID),
		SubjectName:            &permission.SubjectName,
		PermissionType:         (*string)(&permission.PermissionType),
		PermissionLevel:        (*string)(&permission.PermissionLevel),
		Priority:               pointy.GetPointer(int64(permission.Priority)),
		RequireApproval:        &permission.RequireApproval,
		Inheritable:            &permission.Inheritable,
		ParentPermissionId:     pointy.GetPointer(permission.ParentPermissionID),
		RiskLevel:              (*string)(&permission.RiskLevel),
		RequireMfa:             &permission.RequireMfa,
		IsTemporary:            &permission.IsTemporary,
		UsageCount:             pointy.GetPointer(int64(permission.UsageCount)),
		Description:            pointy.GetPointer(permission.Description),
		Comments:               pointy.GetPointer(permission.Comments),
		CreatedBy:              pointy.GetPointer(permission.CreatedBy),
		UpdatedBy:              pointy.GetPointer(permission.UpdatedBy),
	}

	// 设置CI类型、CI实例、属性相关字段
	if permission.ScopeTargetType == "ci_type_id" {
		result.CiTypeId = &permission.ScopeTargetID
	} else if permission.ScopeTargetType == "ci_id" {
		result.CiId = &permission.ScopeTargetID
	} else if permission.ScopeTargetType == "attribute_id" {
		result.AttributeId = &permission.ScopeTargetID
	}

	// 转换状态
	var statusNum uint32
	switch permission.Status {
	case "active":
		statusNum = 1
	case "inactive":
		statusNum = 0
	case "suspended":
		statusNum = 2
	case "revoked":
		statusNum = 3
	case "expired":
		statusNum = 4
	default:
		statusNum = 1
	}
	result.Status = &statusNum

	// 转换时间戳
	if !permission.EffectiveFrom.IsZero() {
		result.EffectiveFrom = pointy.GetPointer(permission.EffectiveFrom.Unix())
	}
	if !permission.EffectiveTo.IsZero() {
		result.EffectiveTo = pointy.GetPointer(permission.EffectiveTo.Unix())
	}
	if !permission.LastUsedAt.IsZero() {
		result.LastUsedAt = pointy.GetPointer(permission.LastUsedAt.Unix())
	}

	// 转换操作位掩码为操作列表
	if permission.OperationsMask > 0 {
		operations := &cmdb.CiPermissionOperations{
			Operations: []*cmdb.CiPermissionOperation{},
		}

		if permission.OperationsMask&1 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   "read",
				Description: "读取权限",
			})
		}
		if permission.OperationsMask&2 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   "write",
				Description: "写入权限",
			})
		}
		if permission.OperationsMask&4 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   "delete",
				Description: "删除权限",
			})
		}
		if permission.OperationsMask&8 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   "approve",
				Description: "审批权限",
			})
		}
		if permission.OperationsMask&16 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation:   "export",
				Description: "导出权限",
			})
		}

		result.Operations = operations
	}

	return result, nil
}