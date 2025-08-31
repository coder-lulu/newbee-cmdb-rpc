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

// ChoiceInteger 对应于数据库cmdb_choice_integers
type ChoiceInteger struct {
	ent.Schema
}

func (ChoiceInteger) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

func (ChoiceInteger) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.Int("value").Comment("选项值"),
		field.JSON("option", ChoiceItemMetaS{}).Optional().Comment("选项内容"),
	}
}

func (ChoiceInteger) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attribute", Attribute.Type).Ref("choice_integers").Field("attr_id").Unique().Comment("属性").Required(),
	}
}

// Annotations 返回表的注释
func (ChoiceInteger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_choice_integers"},
	}
}
