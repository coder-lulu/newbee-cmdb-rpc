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

// CiTypeGroup 对应于数据库cmdb_ci_type_groups
// 用于配置项类型的分组管理
type CiTypeGroup struct {
	ent.Schema
}

func (CiTypeGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.SortMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
		mixins2.SoftDeleteMixin{},
	}
}

// Fields of the CiTypeGroup.
func (CiTypeGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(32).Unique().NotEmpty().Comment("分组名称"),
		field.String("description").MaxLen(255).Comment("分组描述"),
		field.String("icon").MaxLen(255).Comment("分组图标").Optional(),
	}
}

// Edges of the CiTypeGroup.
func (CiTypeGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group_items", CiTypeGroupItem.Type),
	}
}

// Annotations 返回表的注释
func (CiTypeGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_groups"},
	}
}
