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

// CiTypeDiscoveryConfig 对应于数据库cmdb_ci_type_discovery_configs
// CI类型发现配置，用于定义每个CI类型的自动发现方式
type CiTypeDiscoveryConfig struct {
	ent.Schema
}

func (CiTypeDiscoveryConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
		mixins.CreatedByMixin{},
	}
}

// Fields of the CiTypeDiscoveryConfig.
func (CiTypeDiscoveryConfig) Fields() []ent.Field {
	return []ent.Field{
		// 基础信息
		field.Uint64("ci_type_id").Comment("CI类型ID"),
		field.String("config_name").MaxLen(200).Comment("配置名称"),
		field.Text("description").Optional().Comment("配置描述"),
		
		// 发现配置
		field.String("discovery_mode").Default("attribute").Comment("发现模式: attribute, relation, dependency"),
		field.String("discovery_type").Comment("发现类型: file, api, sdk, builtin, agent"),
		field.String("provider_id").Comment("发现提供者ID"),
		field.JSON("provider_config", map[string]interface{}{}).Comment("提供者配置JSON"),
		
		// 属性映射配置
		field.JSON("attribute_mappings", []interface{}{}).Comment("属性映射配置JSON"),
		field.JSON("discovery_rules", map[string]interface{}{}).Comment("发现规则JSON"),
		field.JSON("filter_conditions", map[string]interface{}{}).Comment("过滤条件JSON"),
		
		// 执行配置
		field.String("execution_mode").Default("manual").Comment("执行模式: manual, scheduled, triggered"),
		field.JSON("schedule_config", map[string]interface{}{}).Optional().Comment("调度配置JSON"),
		field.Int("priority").Default(100).Comment("优先级(数字越小优先级越高)"),
		field.Int("batch_size").Default(100).Comment("批处理大小"),
		field.Int("timeout_seconds").Default(300).Comment("超时时间(秒)"),
		
		// 数据处理配置
		field.String("conflict_resolution").Default("priority").Comment("冲突解决策略: priority, merge, latest"),
		field.Bool("auto_create_ci").Default(false).Comment("是否自动创建CI实例"),
		field.Bool("auto_update_attributes").Default(true).Comment("是否自动更新属性"),
		field.JSON("notification_config", map[string]interface{}{}).Optional().Comment("通知配置JSON"),
		
		// 状态字段
		field.Bool("enabled").Default(true).Comment("是否启用"),
		field.String("config_status").Default("active").Comment("状态: active, inactive, error"),
		field.Time("last_executed_at").Optional().Comment("上次执行时间"),
		field.Time("next_execution_at").Optional().Comment("下次执行时间"),
		
		// 统计信息
		field.JSON("execution_stats", map[string]interface{}{}).Optional().Comment("执行统计JSON"),
		field.Text("last_error").Optional().Comment("最后错误信息"),
	}
}

// Edges of the CiTypeDiscoveryConfig.
func (CiTypeDiscoveryConfig) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联CI类型
		edge.From("ci_type", CiType.Type).
			Ref("discovery_configs").
			Field("ci_type_id").
			Unique().
			Required().
			Comment("关联cmdb_ci_types.id"),
			
		// 关联发现历史
		edge.To("execution_histories", DiscoveryExecutionHistory.Type).
			Comment("发现执行历史"),
		
		// 关联属性映射规则
		edge.To("attribute_mapping_rules", AttributeMappingRule.Type).
			Comment("属性映射规则"),
	}
}

// Annotations 返回表的注释
func (CiTypeDiscoveryConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_type_discovery_configs"},
	}
}