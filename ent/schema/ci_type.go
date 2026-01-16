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

// CiType 对应于数据库cmdb_ci_types
// 代表CMDB中的配置项类型（CI类型）
type CiType struct {
	ent.Schema
}

func (CiType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
		mixins.CreatedByMixin{},
	}
}

// Fields of the CiType.
func (CiType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(32).NotEmpty().Comment("名称"),
		field.String("alias").MaxLen(32).NotEmpty().Comment("别名"),
		field.Uint64("unique_id").Comment("外键，关联c_attributes.id"),
		field.Bool("is_inherited").Comment("是否继承").Optional().Nillable().Default(false),
		field.String("icon").Optional().Nillable().Comment("图标"),
		field.Uint64("default_order_attr_id").Optional().Nillable().Comment("默认排序属性"),
		field.Uint64("show_id").Optional().Nillable().Comment("展示ID"),
		field.JSON("unique_const", []CiTypeUniqueConstS{}).Optional().Comment("唯一性约束"),
	}
}

type CiTypeUniqueConstS struct {
	AttrIds []uint64 `json:"attrIds"`
}

// Edges of the CiType.
func (CiType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("attributes", Attribute.Type).Field("unique_id").Unique().Comment("关联c_attributes.id").Required(),

		// CI实例关系
		edge.To("cis", Cis.Type),

		// 类型属性关系
		edge.To("type_attributes", CiTypeAttribute.Type),

		// 类型分组关系
		edge.To("attribute_groups", CiTypeAttributeGroup.Type),
		edge.To("group_items", CiTypeGroupItem.Type),

		// 类型继承关系
		edge.To("children", CiTypeInheritance.Type),
		edge.To("parents", CiTypeInheritance.Type),

		// 类型关系定义
		edge.To("child_relations", CiTypeRelation.Type),
		edge.To("parent_relations", CiTypeRelation.Type),

		// 导入相关关系
		edge.To("import_templates", ImportTemplate.Type),
		edge.To("import_records", ImportRecord.Type),

		// 变更记录关系
		edge.To("ci_records", CiRecords.Type),

		// 自动发现配置关系
		edge.To("discovery_configs", CiTypeDiscoveryConfig.Type),
	}
}

// Annotations 返回表的注释
func (CiType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_types"},
	}
}
