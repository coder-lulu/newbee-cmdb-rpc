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

// PermissionDataFilter 权限数据过滤规则表 - 将JSON过滤规则拆分为独立表
type PermissionDataFilter struct {
	ent.Schema
}

// Mixin of the PermissionDataFilter.
func (PermissionDataFilter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the PermissionDataFilter.
func (PermissionDataFilter) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("permission_id").
			Comment("权限ID"),

		field.Int("filter_group").
			Default(1).
			Comment("过滤组，同组内为AND关系"),

		field.String("field_name").
			Comment("过滤字段"),

		field.Enum("operator_type").
			Values("eq", "ne", "gt", "lt", "gte", "lte", "contains", "not_contains", "like", "not_like").
			Comment("操作符类型"),

		field.Text("filter_value").
			Optional().
			Comment("过滤值"),

		field.Enum("value_type").
			Values("string", "number", "boolean", "array").
			Default("string").
			Comment("值类型"),
	}
}

// Edges of the PermissionDataFilter.
func (PermissionDataFilter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission", CiPermission.Type).
			Ref("data_filters").
			Field("permission_id").
			Unique().
			Required(),
	}
}

// Indexes of the PermissionDataFilter.
func (PermissionDataFilter) Indexes() []ent.Index {
	return []ent.Index{
		// 权限过滤查询索引
		index.Fields("permission_id", "filter_group"),
		index.Fields("field_name", "operator_type"),
		index.Fields("permission_id"),
	}
}

// Annotations 返回表的注释
func (PermissionDataFilter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_permission_data_filters"},
	}
}
