package cipermission

import (
	"context"
	"fmt"
	"time"

	"gitee.com/link234/cmdb-rpc/ent"
	"gitee.com/link234/cmdb-rpc/ent/cipermission"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/gofrs/uuid/v5"

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
	l.Logger.Infof("开始查询CI权限列表，页码: %d, 每页数量: %d", in.Page, in.PageSize)

	// 1. 数据验证
	if err := l.validateRequest(in); err != nil {
		l.Logger.Errorf("查询CI权限列表数据验证失败: %v", err)
		return nil, err
	}

	// 2. 构建查询条件
	query := l.svcCtx.DB.CiPermission.Query()

	// 添加过滤条件
	l.applyFilters(query, in)

	// 3. 查询总数
	total, err := query.Clone().Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CI权限总数失败: %v", err)
		return nil, fmt.Errorf("查询权限总数失败: %v", err)
	}

	// 4. 分页查询
	offset := (in.Page - 1) * in.PageSize
	permissions, err := query.
		Order(ent.Desc(cipermission.FieldCreatedAt)).
		Offset(int(offset)).
		Limit(int(in.PageSize)).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询CI权限列表失败: %v", err)
		return nil, fmt.Errorf("查询权限列表失败: %v", err)
	}

	// 5. 转换数据格式
	result := &cmdb.CiPermissionListResp{
		Total: uint64(total),
		Data:  make([]*cmdb.CiPermissionInfo, 0, len(permissions)),
	}

	for _, permission := range permissions {
		permissionInfo, err := l.convertToPermissionInfo(permission)
		if err != nil {
			l.Logger.Errorf("转换权限数据失败，ID: %d, 错误: %v", permission.ID, err)
			continue // 跳过转换失败的数据，不中断整个列表查询
		}
		result.Data = append(result.Data, permissionInfo)
	}

	l.Logger.Infof("成功查询CI权限列表，总数: %d, 返回数量: %d", total, len(result.Data))

	return result, nil
}

// validateRequest 验证列表查询请求
func (l *GetCiPermissionListLogic) validateRequest(in *cmdb.CiPermissionListReq) error {
	if in == nil {
		return fmt.Errorf("请求数据不能为空")
	}

	if in.Page == 0 {
		in.Page = 1
	}

	if in.PageSize == 0 {
		in.PageSize = 10
	}

	// 限制单页查询数量，防止性能问题
	const maxPageSize = 1000
	if in.PageSize > maxPageSize {
		return fmt.Errorf("单页查询数量超过限制，最大允许 %d 条，当前请求 %d 条", maxPageSize, in.PageSize)
	}

	return nil
}

// applyFilters 应用查询过滤条件
func (l *GetCiPermissionListLogic) applyFilters(query *ent.CiPermissionQuery, req *cmdb.CiPermissionListReq) {
	// 基本字段过滤
	if req.DepartmentId != nil {
		query.Where(cipermission.DepartmentIDEQ(*req.DepartmentId))
	}

	if req.PermissionId != nil && *req.PermissionId != "" {
		query.Where(cipermission.PermissionIDContains(*req.PermissionId))
	}

	if req.ScopeType != nil && *req.ScopeType != "" {
		query.Where(cipermission.ScopeTypeEQ(cipermission.ScopeType(*req.ScopeType)))
	}

	if req.CiTypeId != nil {
		query.Where(cipermission.CiTypeIDEQ(*req.CiTypeId))
	}

	if req.CiId != nil {
		query.Where(cipermission.CiIDEQ(*req.CiId))
	}

	if req.AttributeId != nil {
		query.Where(cipermission.AttributeIDEQ(*req.AttributeId))
	}

	if req.FieldName != nil && *req.FieldName != "" {
		query.Where(cipermission.FieldNameContains(*req.FieldName))
	}

	// 权限主体过滤
	if req.SubjectType != nil && *req.SubjectType != "" {
		query.Where(cipermission.SubjectTypeEQ(cipermission.SubjectType(*req.SubjectType)))
	}

	if req.SubjectName != nil && *req.SubjectName != "" {
		query.Where(cipermission.SubjectNameContains(*req.SubjectName))
	}

	if req.SubjectCode != nil && *req.SubjectCode != "" {
		query.Where(cipermission.SubjectCodeContains(*req.SubjectCode))
	}

	// 权限类型过滤
	if req.PermissionType != nil && *req.PermissionType != "" {
		query.Where(cipermission.PermissionTypeEQ(cipermission.PermissionType(*req.PermissionType)))
	}

	if req.PermissionLevel != nil && *req.PermissionLevel != "" {
		query.Where(cipermission.PermissionLevelEQ(cipermission.PermissionLevel(*req.PermissionLevel)))
	}

	// 状态过滤
	if req.Status != nil && *req.Status != "" {
		query.Where(cipermission.StatusEQ(cipermission.Status(*req.Status)))
	}

	// 风险等级过滤
	if req.RiskLevel != nil && *req.RiskLevel != "" {
		query.Where(cipermission.RiskLevelEQ(cipermission.RiskLevel(*req.RiskLevel)))
	}

	// 审批相关过滤
	if req.RequireApproval != nil {
		query.Where(cipermission.RequireApprovalEQ(*req.RequireApproval))
	}

	// 继承相关过滤
	if req.Inheritable != nil {
		query.Where(cipermission.InheritableEQ(*req.Inheritable))
	}

	if req.ParentPermissionId != nil && *req.ParentPermissionId != "" {
		query.Where(cipermission.ParentPermissionIDEQ(*req.ParentPermissionId))
	}

	// MFA过滤
	if req.RequireMfa != nil {
		query.Where(cipermission.RequireMfaEQ(*req.RequireMfa))
	}

	// 时间过滤
	if req.IsTemporary != nil {
		query.Where(cipermission.IsTemporaryEQ(*req.IsTemporary))
	}

	// 时间范围过滤
	if req.CreatedAt != nil {
		query.Where(cipermission.CreatedAtGTE(time.Unix(*req.CreatedAt, 0)))
	}

	if req.UpdatedAt != nil {
		query.Where(cipermission.UpdatedAtGTE(time.Unix(*req.UpdatedAt, 0)))
	}
}

// convertToPermissionInfo 将数据库实体转换为响应格式（简化版本）
func (l *GetCiPermissionListLogic) convertToPermissionInfo(permission *ent.CiPermission) (*cmdb.CiPermissionInfo, error) {
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

	// 权限类型和级别
	result.PermissionType = pointy.GetPointer(string(permission.PermissionType))
	result.PermissionLevel = pointy.GetPointer(string(permission.PermissionLevel))
	result.Priority = pointy.GetPointer(int64(permission.Priority))

	// 时间相关
	if !permission.EffectiveFrom.IsZero() {
		result.EffectiveFrom = pointy.GetPointer(permission.EffectiveFrom.Unix())
	}
	if !permission.EffectiveTo.IsZero() {
		result.EffectiveTo = pointy.GetPointer(permission.EffectiveTo.Unix())
	}
	result.IsTemporary = pointy.GetPointer(permission.IsTemporary)

	// 审批相关
	result.RequireApproval = pointy.GetPointer(permission.RequireApproval)
	if permission.GrantedBy != zeroUUID {
		result.GrantedBy = pointy.GetPointer(permission.GrantedBy.String())
	}
	if permission.GrantedByName != "" {
		result.GrantedByName = pointy.GetPointer(permission.GrantedByName)
	}

	// 使用统计
	result.UsageCount = pointy.GetPointer(int64(permission.UsageCount))
	if !permission.LastUsedAt.IsZero() {
		result.LastUsedAt = pointy.GetPointer(permission.LastUsedAt.Unix())
	}

	// 权限状态
	result.Status = pointy.GetPointer(uint32(0)) // 简化处理
	if permission.StatusReason != "" {
		result.StatusReason = pointy.GetPointer(permission.StatusReason)
	}

	// 继承相关
	result.Inheritable = pointy.GetPointer(permission.Inheritable)
	if permission.ParentPermissionID != "" {
		result.ParentPermissionId = pointy.GetPointer(permission.ParentPermissionID)
	}

	// 风险控制
	result.RiskLevel = pointy.GetPointer(string(permission.RiskLevel))
	result.RequireMfa = pointy.GetPointer(permission.RequireMfa)

	// 描述信息
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

	return result, nil
}
