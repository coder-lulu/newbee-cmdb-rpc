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

// CiTypeRelation 对应于数据库cmdb_ci_type_relations
// 用于定义CI类型之间的关系类型及约束
type CiTypeRelation struct {
	ent.Schema
}

func (CiTypeRelation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
	}
}

// Fields of the CiTypeRelation.
func (CiTypeRelation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("parent_id").Comment("外键，关联cmdb_ci_types.id，父类型"),
		field.Uint64("child_id").Comment("外键，关联cmdb_ci_types.id，子类型"),
		field.Uint64("relation_type_id").Comment("外键，关联cmdb_relation_types.id"),
		field.String("constraint").Optional().Comment("约束，枚举类"),
		field.Uint64("parent_attr_id").Optional().Comment("父类型属性ID"),
		field.Uint64("child_attr_id").Optional().Comment("子类型属性ID"),
		field.JSON("parent_attr_ids", []uint64{}).Optional().Comment("父类型属性ID数组"),
		field.JSON("child_attr_ids", []uint64{}).Optional().Comment("子类型属性ID数组"),
	}
}

// Edges of the CiTypeRelation.
func (CiTypeRelation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parent", CiType.Type).Ref("child_relations").Field("parent_id").Unique().Comment("父类").Required(),
		edge.From("child", CiType.Type).Ref("parent_relations").Field("child_id").Unique().Comment("子类").Required(),
		edge.From("relation_type", RelationType.Type).Ref("ci_type_relations").Field("relation_type_id").Unique().Comment("关系类型").Required(),
	}
}

// Annotations 返回表的注释
func (CiTypeRelation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_relations"},
	}
}
