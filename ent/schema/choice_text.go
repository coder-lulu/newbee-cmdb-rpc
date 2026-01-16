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

// ChoiceText 对应于数据库cmdb_choice_texts
type ChoiceText struct {
	ent.Schema
}

func (ChoiceText) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

func (ChoiceText) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.String("value").Comment("选项内容"),
		field.JSON("option", ChoiceItemMetaS{}).Optional().Comment("选项内容"),
	}
}

func (ChoiceText) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attribute", Attribute.Type).Ref("choice_texts").Field("attr_id").Unique().Comment("属性").Required(),
	}
}

// Annotations 返回表的注释
func (ChoiceText) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_choice_texts"},
	}
}
