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

// AttributeMappingRule 对应于数据库cmdb_attribute_mapping_rules
// 属性映射规则，定义源字段到CI属性的映射关系
type AttributeMappingRule struct {
	ent.Schema
}

func (AttributeMappingRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
	}
}

// Fields of the AttributeMappingRule.
func (AttributeMappingRule) Fields() []ent.Field {
	return []ent.Field{
		// 关联信息
		field.Uint64("discovery_config_id").Comment("发现配置ID"),
		field.Uint64("ci_attribute_id").Comment("CI属性ID"),
		
		// 映射配置
		field.String("source_field").Comment("源字段名"),
		field.String("source_path").Optional().Comment("源字段路径(支持JSON路径)"),
		field.String("target_attribute").Comment("目标属性名"),
		
		// 转换配置
		field.String("transform_type").Default("direct").Comment("转换类型: direct, lookup, script, template"),
		field.JSON("transform_config", map[string]interface{}{}).Optional().Comment("转换配置JSON"),
		field.String("default_value").Optional().Comment("默认值"),
		
		// 验证配置
		field.JSON("validation_rules", map[string]interface{}{}).Optional().Comment("验证规则JSON"),
		field.String("validation_regex").Optional().Comment("验证正则表达式"),
		field.Bool("is_required").Default(false).Comment("是否必填"),
		field.Bool("is_unique").Default(false).Comment("是否唯一"),
		
		// 处理配置
		field.Int("priority").Default(100).Comment("映射优先级"),
		field.Bool("enabled").Default(true).Comment("是否启用"),
		field.String("update_strategy").Default("overwrite").Comment("更新策略: overwrite, merge, append"),
		
		// 统计信息
		field.Int64("success_count").Default(0).Comment("成功次数"),
		field.Int64("failed_count").Default(0).Comment("失败次数"),
		field.Time("last_success_at").Optional().Comment("上次成功时间"),
		
		// 扩展信息
		field.Text("description").Optional().Comment("映射描述"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("扩展元数据JSON"),
	}
}

// Edges of the AttributeMappingRule.
func (AttributeMappingRule) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联发现配置
		edge.From("discovery_config", CiTypeDiscoveryConfig.Type).
			Ref("attribute_mapping_rules").
			Field("discovery_config_id").
			Unique().
			Required().
			Comment("关联cmdb_ci_type_discovery_configs.id"),
			
		// 关联CI属性
		edge.From("ci_attribute", Attribute.Type).
			Ref("mapping_rules").
			Field("ci_attribute_id").
			Unique().
			Required().
			Comment("关联cmdb_attributes.id"),
	}
}

// Annotations 返回表的注释
func (AttributeMappingRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_attribute_mapping_rules"},
	}
}