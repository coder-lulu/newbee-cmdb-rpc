package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

// CiTypeAttributeGroupItem 对应于数据库cmdb_ci_type_attribute_group_items
// 记录分组与属性的关系及顺序
type CiTypeAttributeGroupItem struct {
	ent.Schema
}

func (CiTypeAttributeGroupItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.SortMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiTypeAttributeGroupItem.
func (CiTypeAttributeGroupItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("group_id").Comment("外键，关联cmdb_ci_type_attribute_groups.id"),
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
	}
}

// Edges of the CiTypeAttributeGroupItem.
func (CiTypeAttributeGroupItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", CiTypeAttributeGroup.Type).Ref("group_items").Field("group_id").Unique().Comment("关联cmdb_ci_type_attribute_groups.id").Required(),
		edge.From("attribute", Attribute.Type).Ref("group_items").Field("attr_id").Unique().Comment("关联cmdb_attributes.id").Required(),
	}
}

// Annotations 返回表的注释
func (CiTypeAttributeGroupItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_attribute_group_items"},
	}
}
