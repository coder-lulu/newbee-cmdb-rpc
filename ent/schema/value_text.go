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

// ValueText 对应于数据库c_value_texts
type ValueText struct {
	ent.Schema
}

func (ValueText) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
	}
}

func (ValueText) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("ci_id").Comment("外键，关联cmdb_cis.id"),
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.Text("value").Comment("属性值"),
		field.Bool("is_cover").Comment("是否可被覆盖").Default(true),
	}
}

func (ValueText) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ci", Cis.Type).Ref("value_texts").Field("ci_id").Unique().Comment("CI实例").Required(),
		edge.From("attribute", Attribute.Type).Ref("value_texts").Field("attr_id").Unique().Comment("属性").Required(),
	}
}

// ValueText 返回表的注释
func (ValueText) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_value_texts"},
	}
}
