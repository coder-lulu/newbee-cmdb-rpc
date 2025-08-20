package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "gitee.com/link234/cmdb-rpc/ent/schema/mixins"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
)

// CiTypeGroupItem 对应于数据库cmdb_ci_type_group_items
// 用于配置项类型分组与类型的关联关系
type CiTypeGroupItem struct {
	ent.Schema
}

func (CiTypeGroupItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.SortMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiTypeGroupItem.
func (CiTypeGroupItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("group_id").Comment("外键，关联cmdb_ci_type_groups.id"),
		field.Uint64("type_id").Comment("外键，关联cmdb_ci_types.id"),
	}
}

// Edges of the CiTypeGroupItem.
func (CiTypeGroupItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", CiTypeGroup.Type).Ref("group_items").Field("group_id").Unique().Comment("关联cmdb_ci_type_groups.id").Required(),
		edge.From("ci_type", CiType.Type).Ref("group_items").Field("type_id").Unique().Comment("关联cmdb_ci_types.id").Required(),
	}
}

// Annotations 返回表的注释
func (CiTypeGroupItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_group_items"},
	}
}
