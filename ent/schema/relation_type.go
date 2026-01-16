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
		mixins.DepartmentMixin{},
	}
}

// Fields of the RelationType.
func (RelationType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(32).NotEmpty().Comment("关系类型名称"),
		field.String("code").MaxLen(32).NotEmpty().Comment("关系类型编码").Unique(),
		field.Enum("category").Values("physical", "logical", "business").Default("logical").Comment("关系分类"),
		field.Enum("direction").Values("unidirectional", "bidirectional").Default("bidirectional").Comment("关系方向"),
		field.Text("description").Optional().Comment("关系描述"),
		field.Bool("is_standard").Default(false).Comment("是否为标准关系类型"),
		field.Int("sort_order").Default(0).Comment("排序顺序"),
		field.Bool("is_enabled").Default(true).Comment("是否启用"),

		// 拓扑显示元数据字段
		field.String("display_color").MaxLen(16).Default("#666666").Comment("显示颜色(HEX格式)"),
		field.Enum("line_type").Values("solid", "dashed", "dotted").Default("solid").Comment("线条类型"),
		field.String("icon").MaxLen(64).Optional().Comment("关系图标"),
		field.Int("weight").Default(1).Comment("关系权重(1-10,影响显示粗细)"),
		field.String("display_label").MaxLen(64).Optional().Comment("显示标签"),
		field.Text("tooltip_template").Optional().Comment("悬浮提示模板"),
		field.JSON("display_style", map[string]interface{}{}).Optional().Comment("扩展显示样式配置"),
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
