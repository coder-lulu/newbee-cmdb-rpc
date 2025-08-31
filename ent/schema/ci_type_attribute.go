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

// CiTypeAttribute 对应于数据库cmdb_ci_type_attributes
// 记录CI类型与属性的关系及顺序、必填、默认展示等
type CiTypeAttribute struct {
	ent.Schema
}

func (CiTypeAttribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.SortMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiTypeAttribute.
func (CiTypeAttribute) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("type_id").Comment("外键，关联cmdb_ci_types.id"),
		field.Uint64("attr_id").Comment("外键，关联cmdb_attributes.id"),
		field.Bool("is_required").Optional().Comment("是否必填").Default(false),
		field.Bool("is_unique").Optional().Comment("是否唯一").Default(false),
		field.Bool("is_list").Optional().Comment("是否列表").Default(false),
		field.Bool("list_show").Optional().Comment("是否列表展示").Default(false),
		field.Bool("detail_show").Optional().Comment("是否详情展示").Default(false),
		field.Bool("is_edit").Optional().Comment("是否可编辑").Default(true),
	}
}

// Edges of the CiTypeAttribute.
func (CiTypeAttribute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ci_type", CiType.Type).Ref("type_attributes").Field("type_id").Unique().Comment("关联cmdb_ci_types.id").Required(),
		edge.From("attribute", Attribute.Type).Ref("type_attributes").Field("attr_id").Unique().Comment("关联cmdb_attributes.id").Required(),
	}
}

// Annotations 返回表的注释
func (CiTypeAttribute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_attributes"},
	}
}
