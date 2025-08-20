package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
)

// CiPermission holds the schema definition for the CiPermission entity.
// CI权限配置表 - 优化后的核心权限表，简化结构提升性能
type CiPermission struct {
	ent.Schema
}

// Mixin of the CiPermission.
func (CiPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiPermission.
func (CiPermission) Fields() []ent.Field {
	return []ent.Field{
		// 权限唯一标识
		field.String("permission_id").
			Comment("权限ID，全局唯一标识").
			Unique(),

		// 权限范围（优化：提取为独立字段，避免JSON查询）
		field.Enum("scope_type").
			Values("global", "ci_type", "ci_instance", "attribute", "field").
			Comment("权限范围类型"),

		field.String("scope_target_type").
			Optional().
			Comment("目标类型：ci_type_id/ci_id/attribute_id"),

		field.Uint64("scope_target_id").
			Optional().
			Comment("目标ID"),

		field.String("scope_field_name").
			Optional().
			Comment("字段名（仅field类型使用）"),

		// 权限主体
		field.Enum("subject_type").
			Values("user", "role", "department", "group", "system").
			Comment("权限主体类型"),

		field.String("subject_id").
			Optional().
			Comment("权限主体ID"),

		field.String("subject_name").
			Comment("权限主体名称"),

		// 权限配置（核心字段）
		field.Enum("permission_type").
			Values("allow", "deny").
			Default("allow").
			Comment("权限类型：允许或拒绝"),

		field.Enum("permission_level").
			Values("none", "read", "write", "admin", "super_admin").
			Default("none").
			Comment("权限级别"),

		// 操作位掩码（优化：用位掩码代替JSON数组）
		field.Uint64("operations_mask").
			Default(0).
			Comment("操作位掩码：1-read,2-write,4-delete,8-approve等"),

		// 时间控制
		field.Time("effective_from").
			Optional().
			Comment("权限生效开始时间"),

		field.Time("effective_to").
			Optional().
			Comment("权限生效结束时间"),

		field.Bool("is_temporary").
			Default(false).
			Comment("是否为临时权限"),

		// 优先级和状态
		field.Int("priority").
			Default(0).
			Comment("权限优先级，数值越大优先级越高"),

		field.Enum("status").
			Values("active", "inactive", "suspended", "revoked", "expired").
			Default("active").
			Comment("权限状态"),

		// 继承和审批
		field.String("parent_permission_id").
			Optional().
			Comment("父权限ID"),

		field.Bool("inheritable").
			Default(false).
			Comment("是否可继承给子级"),

		field.Bool("require_approval").
			Default(false).
			Comment("是否需要审批"),

		field.Bool("require_mfa").
			Default(false).
			Comment("是否需要多因素认证"),

		// 风险控制
		field.Enum("risk_level").
			Values("low", "medium", "high", "critical").
			Default("low").
			Comment("风险等级"),

		// 使用统计（高频字段）
		field.Int("usage_count").
			Default(0).
			Comment("权限使用次数"),

		field.Time("last_used_at").
			Optional().
			Comment("最后使用时间"),

		// 审计字段
		field.String("created_by").
			Optional().
			Comment("创建人ID"),

		field.String("updated_by").
			Optional().
			Comment("最后更新人ID"),

		field.Text("description").
			Optional().
			Comment("权限描述"),

		field.Text("comments").
			Optional().
			Comment("备注信息"),
	}
}

// Edges of the CiPermission.
func (CiPermission) Edges() []ent.Edge {
	return []ent.Edge{
		// 权限操作关联（替代JSON字段）
		edge.To("operations", PermissionOperation.Type).
			StorageKey(edge.Column("permission_id")),

		// 数据过滤规则关联（替代JSON字段）
		edge.To("data_filters", PermissionDataFilter.Type).
			StorageKey(edge.Column("permission_id")),

		// 字段掩码关联
		edge.To("field_masks", PermissionFieldMask.Type).
			StorageKey(edge.Column("permission_id")),
	}
}

// Indexes of the CiPermission.
func (CiPermission) Indexes() []ent.Index {
	return []ent.Index{
		// 核心查询性能优化索引
		index.Fields("scope_type", "subject_type", "status"),
		index.Fields("subject_id", "scope_type", "scope_target_type", "scope_target_id"),
		index.Fields("scope_target_type", "scope_target_id", "permission_level"),

		// 权限有效性检查索引
		index.Fields("effective_from", "effective_to", "status"),
		index.Fields("subject_id", "scope_type", "scope_target_type", "scope_target_id", "status", "effective_from", "effective_to"),

		// 权限继承查询优化
		index.Fields("parent_permission_id", "inheritable", "status"),

		// 使用统计查询优化
		index.Fields("last_used_at", "usage_count"),

		// 租户和部门级索引
		index.Fields("tenant_id", "subject_type", "subject_id"),

		// 权限类型和级别索引
		index.Fields("permission_type", "permission_level", "status"),
		index.Fields("priority", "status"),

		// 风险等级和审批索引
		index.Fields("risk_level", "require_mfa"),
		index.Fields("require_approval", "status"),

		// 临时权限优化索引
		index.Fields("is_temporary", "effective_to"),

		// 覆盖索引 - 减少回表查询
		index.Fields("subject_id", "scope_target_type", "scope_target_id", "permission_level", "operations_mask", "status"),
	}
}

// Annotations 返回表的注释
func (CiPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_permissions"},
	}
}
