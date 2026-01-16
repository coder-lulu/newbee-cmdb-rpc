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

// CiTypeInheritance 对应于数据库cmdb_ci_type_inheritance
// 用于配置项类型的继承关系（父类型-子类型）
type CiTypeInheritance struct {
	ent.Schema
}

func (CiTypeInheritance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiTypeInheritance.
func (CiTypeInheritance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("parent_id").Comment("外键，关联cmdb_ci_types.id，父类型"),
		field.Uint64("child_id").Comment("外键，关联cmdb_ci_types.id，子类型"),
	}
}

// Edges of the CiTypeInheritance.
func (CiTypeInheritance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parent", CiType.Type).Ref("children").Field("parent_id").Unique().Comment("父类").Required(),
		edge.From("child", CiType.Type).Ref("parents").Field("child_id").Unique().Comment("子类").Required(),
	}
}

// Annotations 返回表的注释
func (CiTypeInheritance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_inheritance"},
	}
}
