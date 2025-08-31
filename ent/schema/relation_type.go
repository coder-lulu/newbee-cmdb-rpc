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

// RelationType 对应于数据库cmdb_relation_types
// 用于定义关系类型
type RelationType struct {
	ent.Schema
}

func (RelationType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
	}
}

// Fields of the RelationType.
func (RelationType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(32).NotEmpty().Comment("关系类型名称"),
		field.String("code").MaxLen(32).NotEmpty().Comment("关系类型编码"),
		field.Enum("category").Values("physical", "logic", "business").Default("logic"),
		field.Enum("direction").Values("unidirectional", "bidirectional").Default("bidirectional"),
	}
}

// Edges of the RelationType.
func (RelationType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ci_relations", CiRelation.Type),
		edge.To("ci_type_relations", CiTypeRelation.Type),
	}
}

// Annotations 返回表的注释
func (RelationType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_relation_types"},
	}
}
