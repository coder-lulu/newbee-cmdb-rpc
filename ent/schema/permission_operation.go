package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// PermissionOperation 权限操作明细表 - 将JSON操作拆分为独立表
type PermissionOperation struct {
	ent.Schema
}

// Mixin of the PermissionOperation.
func (PermissionOperation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the PermissionOperation.
func (PermissionOperation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("permission_id").
			Comment("权限ID"),

		field.String("operation_code").
			Comment("操作代码：read/write/delete/approve等"),

		field.String("operation_name").
			Optional().
			Comment("操作名称"),

		field.Bool("is_allowed").
			Default(true).
			Comment("是否允许"),
	}
}

// Edges of the PermissionOperation.
func (PermissionOperation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission", CiPermission.Type).
			Ref("operations").
			Field("permission_id").
			Unique().
			Required(),
	}
}

// Indexes of the PermissionOperation.
func (PermissionOperation) Indexes() []ent.Index {
	return []ent.Index{
		// 权限操作的唯一性约束
		index.Fields("permission_id", "operation_code").Unique(),

		// 常用查询索引
		index.Fields("permission_id"),
		index.Fields("operation_code"),
		index.Fields("is_allowed"),
	}
}

// Annotations 返回表的注释
func (PermissionOperation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_permission_operations"},
	}
}
