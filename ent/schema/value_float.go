package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// ValueFloat 对应于数据库c_value_floats
type ValueFloat struct {
	ent.Schema
}

func (ValueFloat) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

func (ValueFloat) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("ci_id").Comment("外键，关联cmdb_cis.id"),
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.Float("value").Comment("属性值"),
		field.Bool("is_cover").Comment("是否可被覆盖").Default(true),
	}
}

func (ValueFloat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ci", Cis.Type).Ref("value_floats").Field("ci_id").Unique().Comment("CI实例").Required(),
		edge.From("attribute", Attribute.Type).Ref("value_floats").Field("attr_id").Unique().Comment("属性").Required(),
	}
}

// Indexes of the ValueFloat
func (ValueFloat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("attr_id", "value"),
	}
}

// ValueFloat 浮点数值表
func (ValueFloat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_value_floats"},
	}
}
