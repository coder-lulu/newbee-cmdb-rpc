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

// CiRelation 对应于数据库cmdb_ci_relations
// 用于记录CI实例之间的关联关系
type CiRelation struct {
	ent.Schema
}

func (CiRelation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiRelation.
func (CiRelation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("first_ci_id").Comment("外键，关联cmdb_cis.id，第一CI"),
		field.Uint64("second_ci_id").Comment("外键，关联cmdb_cis.id，第二CI"),
		field.Uint64("relation_type_id").Comment("外键，关联cmdb_relation_types.id"),
		field.Uint64("more").Optional().Comment("更多CI，外键，关联cmdb_cis.id"),
		field.String("source").Optional().Comment("来源，枚举类"),
		field.String("ancestor_ids").Optional().MaxLen(128).Comment("祖先ID"),
	}
}

// Edges of the CiRelation.
func (CiRelation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("first_ci", Cis.Type).Ref("first_relations").Field("first_ci_id").Unique().Comment("第一CI").Required(),
		edge.From("second_ci", Cis.Type).Ref("second_relations").Field("second_ci_id").Unique().Comment("第二CI").Required(),
		edge.From("relation_type", RelationType.Type).Ref("ci_relations").Field("relation_type_id").Unique().Comment("关系类型").Required(),
		edge.From("more_ci", Cis.Type).Ref("more_relations").Field("more").Unique().Comment("更多CI"),
	}
}

// Annotations 返回表的注释
func (CiRelation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_relations"},
	}
}
