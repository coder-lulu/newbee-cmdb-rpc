package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
)

// PermissionFieldMask 权限字段掩码表
type PermissionFieldMask struct {
	ent.Schema
}

// Mixin of the PermissionFieldMask.
func (PermissionFieldMask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
	}
}

// Fields of the PermissionFieldMask.
func (PermissionFieldMask) Fields() []ent.Field {
	return []ent.Field{
		field.String("permission_id").
			Comment("权限ID"),

		field.String("field_name").
			Comment("掩码字段名"),

		field.Enum("mask_type").
			Values("hide", "encrypt", "partial").
			Default("hide").
			Comment("掩码类型"),

		field.String("mask_rule").
			Optional().
			Comment("掩码规则"),

		field.Time("created_at").
			Default(time.Now).
			Comment("创建时间"),
	}
}

// Edges of the PermissionFieldMask.
func (PermissionFieldMask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission", CiPermission.Type).
			Ref("field_masks").
			Field("permission_id").
			Unique().
			Required(),
	}
}

// Indexes of the PermissionFieldMask.
func (PermissionFieldMask) Indexes() []ent.Index {
	return []ent.Index{
		// 权限字段掩码的唯一性约束
		index.Fields("permission_id", "field_name").Unique(),
		index.Fields("permission_id"),
		index.Fields("field_name", "mask_type"),
	}
}

// Annotations 返回表的注释
func (PermissionFieldMask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_permission_field_masks"},
	}
}
