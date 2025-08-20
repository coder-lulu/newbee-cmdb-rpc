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

// ChoiceFloat 对应于数据库cmdb_choice_floats
type ChoiceFloat struct {
	ent.Schema
}

func (ChoiceFloat) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

func (ChoiceFloat) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.Float("value").Comment("选项值"),
		field.JSON("option", ChoiceItemMetaS{}).Optional().Comment("选项内容"),
	}
}

func (ChoiceFloat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attribute", Attribute.Type).Ref("choice_floats").Field("attr_id").Unique().Comment("属性").Required(),
	}
}

// Annotations 返回表的注释
func (ChoiceFloat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_choice_floats"},
	}
}
