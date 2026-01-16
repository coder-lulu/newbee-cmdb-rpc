package cipermission

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cipermission"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiPermissionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiPermissionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiPermissionListLogic {
	return &GetCiPermissionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiPermissionListLogic) GetCiPermissionList(in *cmdb.CiPermissionListReq) (*cmdb.CiPermissionListResp, error) {
	l.Logger.Infof("开始查询CI权限列表，页码: %d, 页大小: %d", 
		in.Page, in.PageSize)

	// 调试：打印所有查询参数
	l.Logger.Infof("查询参数详情: PermissionId=%v, SubjectType=%v, ScopeType=%v, Status=%v", 
		pointy.GetPointer(in.PermissionId), pointy.GetPointer(in.SubjectType), 
		pointy.GetPointer(in.ScopeType), pointy.GetPointer(in.Status))

	// 1. 数据验证
	if err := l.validateListRequest(in); err != nil {
		l.Logger.Errorf("查询CI权限列表数据验证失败: %v", err)
		return nil, err
	}

	// 2. 构建查询条件
	query := l.svcCtx.DB.CiPermission.Query()

	// 应用过滤条件
	l.applyFilters(query, in)

	// 3. 查询总数
	total, err := query.Clone().Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CI权限总数时数据库操作失败: %v", err)
		return nil, fmt.Errorf("查询CI权限总数失败: %v", err)
	}

	l.Logger.Infof("查询到CI权限总数: %d", total)

	// 4. 应用分页和排序
	l.applyPaginationAndSort(query, in)

	// 5. 执行查询
	permissions, err := query.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CI权限列表时数据库操作失败: %v", err)
		return nil, fmt.Errorf("查询CI权限列表失败: %v", err)
	}

	l.Logger.Infof("实际查询到的权限数量: %d", len(permissions))

	// 6. 转换数据格式
	result := &cmdb.CiPermissionListResp{
		Total: uint64(total),
		Data:  make([]*cmdb.CiPermissionInfo, 0, len(permissions)),
	}

	for _, permission := range permissions {
		permissionInfo := l.convertPermissionToRPC(permission)
		result.Data = append(result.Data, permissionInfo)
	}

	l.Logger.Infof("成功查询CI权限列表，总数: %d, 返回数量: %d", total, len(result.Data))

	return result, nil
}

// validateListRequest 验证列表查询请求
func (l *GetCiPermissionListLogic) validateListRequest(in *cmdb.CiPermissionListReq) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	// 验证分页参数
	page := in.Page
	if page <= 0 {
		return fmt.Errorf("页码必须大于0")
	}

	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		return fmt.Errorf("页大小必须在1-1000之间")
	}

	return nil
}

// applyFilters 应用查询过滤条件
func (l *GetCiPermissionListLogic) applyFilters(query *ent.CiPermissionQuery, in *cmdb.CiPermissionListReq) {
	// 权限ID过滤
	if in.PermissionId != nil && *in.PermissionId != "" {
		query.Where(cipermission.PermissionIDContains(*in.PermissionId))
	}

	// 主体类型过滤
	if in.SubjectType != nil && *in.SubjectType != "" {
		query.Where(cipermission.SubjectTypeEQ(cipermission.SubjectType(*in.SubjectType)))
	}

	// 主体ID过滤
	if in.SubjectId != nil {
		query.Where(cipermission.SubjectIDEQ(*in.SubjectId))
	}

	// 主体名称过滤
	if in.SubjectName != nil && *in.SubjectName != "" {
		query.Where(cipermission.SubjectNameContains(*in.SubjectName))
	}

	// 权限范围类型过滤
	if in.ScopeType != nil && *in.ScopeType != "" {
		query.Where(cipermission.ScopeTypeEQ(cipermission.ScopeType(*in.ScopeType)))
	}

	// CI类型ID过滤
	if in.CiTypeId != nil {
		query.Where(cipermission.ScopeTargetTypeEQ("ci_type_id")).
			Where(cipermission.ScopeTargetIDEQ(*in.CiTypeId))
	}

	// CI实例ID过滤
	if in.CiId != nil {
		query.Where(cipermission.ScopeTargetTypeEQ("ci_id")).
			Where(cipermission.ScopeTargetIDEQ(*in.CiId))
	}

	// 属性ID过滤
	if in.AttributeId != nil {
		query.Where(cipermission.ScopeTargetTypeEQ("attribute_id")).
			Where(cipermission.ScopeTargetIDEQ(*in.AttributeId))
	}

	// 权限类型过滤
	if in.PermissionType != nil && *in.PermissionType != "" {
		query.Where(cipermission.PermissionTypeEQ(cipermission.PermissionType(*in.PermissionType)))
	}

	// 权限级别过滤
	if in.PermissionLevel != nil && *in.PermissionLevel != "" {
		query.Where(cipermission.PermissionLevelEQ(cipermission.PermissionLevel(*in.PermissionLevel)))
	}

	// 状态过滤
	if in.Status != nil {
		var statusStr string
		switch *in.Status {
		case "1":
			statusStr = "active"
		case "0":
			statusStr = "inactive"
		case "2":
			statusStr = "suspended"
		case "3":
			statusStr = "revoked"
		case "4":
			statusStr = "expired"
		}
		if statusStr != "" {
			query.Where(cipermission.StatusEQ(cipermission.Status(statusStr)))
		}
	}

	// 风险等级过滤
	if in.RiskLevel != nil && *in.RiskLevel != "" {
		query.Where(cipermission.RiskLevelEQ(cipermission.RiskLevel(*in.RiskLevel)))
	}

	// 是否需要审批过滤
	if in.RequireApproval != nil {
		query.Where(cipermission.RequireApprovalEQ(*in.RequireApproval))
	}

	// 是否可继承过滤
	if in.Inheritable != nil {
		query.Where(cipermission.InheritableEQ(*in.Inheritable))
	}

	// 是否需要MFA过滤
	if in.RequireMfa != nil {
		query.Where(cipermission.RequireMfaEQ(*in.RequireMfa))
	}

	// 是否临时权限过滤
	if in.IsTemporary != nil {
		query.Where(cipermission.IsTemporaryEQ(*in.IsTemporary))
	}
}

// applyPaginationAndSort 应用分页和排序
func (l *GetCiPermissionListLogic) applyPaginationAndSort(query *ent.CiPermissionQuery, in *cmdb.CiPermissionListReq) {
	// 默认按创建时间降序排序
	query.Order(ent.Desc(cipermission.FieldCreatedAt))

	// 应用分页
	page := in.Page
	pageSize := in.PageSize
	offset := (page - 1) * pageSize

	query.Offset(int(offset)).Limit(int(pageSize))
}

// convertPermissionToRPC 转换权限数据为RPC格式
func (l *GetCiPermissionListLogic) convertPermissionToRPC(permission *ent.CiPermission) *cmdb.CiPermissionInfo {
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
				Operation: "read",
			})
		}
		if permission.OperationsMask&2 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation: "write",
			})
		}
		if permission.OperationsMask&4 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation: "delete",
			})
		}
		if permission.OperationsMask&8 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation: "approve",
			})
		}
		if permission.OperationsMask&16 != 0 {
			operations.Operations = append(operations.Operations, &cmdb.CiPermissionOperation{
				Operation: "export",
			})
		}

		result.Operations = operations
	}

	return result
}