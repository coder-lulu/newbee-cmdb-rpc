package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
	"github.com/gofrs/uuid/v5"
)

// Ci 对应于数据库cmdb_cis
// 用于记录CI实例
type Cis struct {
	ent.Schema
}

func (Cis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the Ci.
func (Cis) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("type_id").Comment("外键，关联cmdb_ci_types.id"),
		field.Uint32("status").Optional().Comment("状态，枚举类型").Default(1),
		field.UUID("created_by", uuid.UUID{}).Optional().Nillable().Comment("创建者"),
		field.JSON("tags", []CiTag{}).Optional().Comment("标签"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("元数据"),
		field.JSON("custom_fields", map[string]interface{}{}).Optional().Comment("自定义字段"),
	}
}

type CiTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Edges of the Ci.
func (Cis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ci_type", CiType.Type).Ref("cis").Field("type_id").Unique().Comment("关联cmdb_ci_types.id").Required(),

		// 属性值关
		edge.To("value_texts", ValueText.Type),
		edge.To("value_index_texts", ValueIndexText.Type),
		edge.To("value_jsons", ValueJSON.Type),
		edge.To("value_integers", ValueInteger.Type),
		edge.To("value_floats", ValueFloat.Type),
		edge.To("value_datetimes", ValueDatetime.Type),

		// 关系
		edge.To("first_relations", CiRelation.Type),
		edge.To("second_relations", CiRelation.Type),
		edge.To("more_relations", CiRelation.Type),

		// 导入相关关系
		edge.To("import_records", ImportRecord.Type),

		// 变更记录关系
		edge.To("records", CiRecords.Type),
	}
}

// Annotations 返回表的注释
func (Cis) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_cis"},
	}
}
