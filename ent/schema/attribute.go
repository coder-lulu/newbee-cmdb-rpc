package schema

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "gitee.com/link234/cmdb-rpc/ent/schema/mixins"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
	"gitee.com/link234/newbee-backend-common/utils/validator"
	"github.com/gofrs/uuid/v5"
)

// Attribute 对应于数据库cmdb_attributes
// 用于定义CMDB的属性元数据
type Attribute struct {
	ent.Schema
}

func (Attribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the Attribute.
func (Attribute) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(64).NotEmpty().Comment("属性名").Unique(),
		field.String("alias").MaxLen(32).NotEmpty().Comment("属性别名"),
		field.Enum("value_type").Values("int", "float", "text", "longtext", "datetime", "date", "time", "json", "password", "link", "reference", "boolean", "image").Comment("值类型，枚举").Default("text"),
		field.Bool("is_choice").Optional().Comment("是否为选项").Default(false),
		field.Bool("is_list").Optional().Comment("是否为列表").Default(false),
		field.UUID("created_by", uuid.UUID{}).Optional().Nillable().Comment("用户ID/创建者"),
		field.Bool("is_computed").Optional().Comment("是否计算属性").Default(false),
		field.JSON("choice_web_hook", AttributeChoiceWebHookS{}).Optional().Comment("选项webhook"),
		field.JSON("option", AttributeOptionS{}).Optional().Comment("选项内容"),
		field.Bool("is_password").Optional().Comment("是否为密码").Default(false),
		field.String("compute_script").Optional().Comment("计算脚本"),
		field.String("compute_expr").Optional().Comment("计算表达式"),
		field.Bool("is_sortable").Optional().Comment("是否可排序").Default(false),
		field.JSON("default", AttributeDefaultValueS{}).Optional().Comment("默认值"),
		field.Bool("is_dynamic").Optional().Comment("是否动态").Default(false),
		field.Bool("is_reference").Optional().Comment("是否引用").Default(false),
		field.Uint64("reference_type_id").Optional().Comment("引用类型ID"),
		field.JSON("choice_other", AttributeChoiceOtherS{}).Optional().Comment("其他选项"),
		field.JSON("validator_rules", []validator.ValidationRule{}).Optional().Comment("正则校验"),
	}
}

type FontOptionS struct {
	Color          string `json:"color"`
	BgColor        string `json:"bgColor"`
	FontStyle      string `json:"fontStyle"`
	FontWeight     string `json:"fontWeight"`
	TextDecoration string `json:"textDecoration"`
}

type ImageOptionS struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Height uint32 `json:"height"`
	Width  uint32 `json:"width"`
}

type AttributeOptionS struct {
	FontOption  FontOptionS    `json:"fontOption"`
	ImageOption []ImageOptionS `json:"imageOptions"`
}

type ChoiceItemMetaS struct {
	Label string      `json:"label"`
	Icon  string      `json:"icon"`
	Style FontOptionS `json:"style"`
}

type ChoiceItemS struct {
	Id    uint64          `json:"id"`
	Value string          `json:"value"`
	Meta  ChoiceItemMetaS `json:"meta"`
}

type AttributeChoiceWebHookS struct {
	Url            string `json:"url"`
	Method         string `json:"method"`
	Body           string `json:"body"`
	Headers        string `json:"headers"`
	Params         string `json:"params"`
	ResponseType   string `json:"responseType"`
	ResponseFormat string `json:"responseFormat"`
}

type AttributeChoiceOtherS struct {
	Filter  string   `json:"filter"`
	AttrId  uint64   `json:"attrId"`
	TypeIds []uint64 `json:"typeIds"`
}

type AttributeDefaultValueS struct {
	Default interface{} `json:"default"`
}

// Edges of the Attribute
func (Attribute) Edges() []ent.Edge {
	return []ent.Edge{
		// 属性值关
		edge.To("value_texts", ValueText.Type),
		edge.To("value_index_texts", ValueIndexText.Type),
		edge.To("value_jsons", ValueJSON.Type),
		edge.To("value_integers", ValueInteger.Type),
		edge.To("value_floats", ValueFloat.Type),
		edge.To("value_datetimes", ValueDatetime.Type),

		// 选项关系
		edge.To("choice_texts", ChoiceText.Type),
		edge.To("choice_integers", ChoiceInteger.Type),
		edge.To("choice_floats", ChoiceFloat.Type),

		// CI类型关系
		edge.To("type_attributes", CiTypeAttribute.Type),

		// 属性组关系
		edge.To("group_items", CiTypeAttributeGroupItem.Type),
	}
}

// Annotations 返回表的注释
func (Attribute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_attributes"},
	}
}
