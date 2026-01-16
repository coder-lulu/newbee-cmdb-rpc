package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// CiTypeAttributeGroup 对应于数据库cmdb_ci_type_attribute_groups
// 用于定义属性分组
type CiTypeAttributeGroup struct {
	ent.Schema
}

func (CiTypeAttributeGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.SortMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}

}

// Fields of the CiTypeAttributeGroup.
func (CiTypeAttributeGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(64).NotEmpty().Comment("分组名称"),
		field.Uint64("type_id").Comment("外键，关联cmdb_ci_types.id"),
	}
}

// Edges of the CiTypeAttributeGroup.
func (CiTypeAttributeGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ci_type", CiType.Type).Ref("attribute_groups").Field("type_id").Unique().Comment("关联cmdb_ci_types.id").Required(),
		edge.To("group_items", CiTypeAttributeGroupItem.Type),
	}
}

// Annotations 返回表的注释
func (CiTypeAttributeGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_attribute_groups"},
	}
}
