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

// DiscoveryExecutionHistory 对应于数据库cmdb_discovery_execution_histories
// 发现执行历史记录，跟踪每次发现任务的执行情况
type DiscoveryExecutionHistory struct {
	ent.Schema
}

func (DiscoveryExecutionHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
	}
}

// Fields of the DiscoveryExecutionHistory.
func (DiscoveryExecutionHistory) Fields() []ent.Field {
	return []ent.Field{
		// 关联信息
		field.Uint64("discovery_config_id").Comment("发现配置ID"),
		field.String("execution_id").Comment("执行ID(UUID)"),
		
		// 执行信息
		field.String("trigger_type").Comment("触发类型: manual, scheduled, api"),
		field.Uint64("triggered_by").Optional().Comment("触发人ID"),
		field.Time("started_at").Comment("开始时间"),
		field.Time("completed_at").Optional().Comment("完成时间"),
		field.Int("duration_seconds").Optional().Comment("执行时长(秒)"),
		
		// 执行状态
		field.String("exec_status").Comment("状态: running, completed, failed, cancelled"),
		field.String("stage").Optional().Comment("当前阶段: connecting, discovering, transforming, persisting"),
		field.Int("progress").Default(0).Comment("进度百分比(0-100)"),
		
		// 统计信息
		field.Int64("total_records").Default(0).Comment("总记录数"),
		field.Int64("processed_records").Default(0).Comment("已处理记录数"),
		field.Int64("success_records").Default(0).Comment("成功记录数"),
		field.Int64("failed_records").Default(0).Comment("失败记录数"),
		field.Int64("skipped_records").Default(0).Comment("跳过记录数"),
		field.Int64("created_cis").Default(0).Comment("创建的CI数量"),
		field.Int64("updated_cis").Default(0).Comment("更新的CI数量"),
		
		// 错误信息
		field.Text("error_message").Optional().Comment("错误信息"),
		field.JSON("error_details", map[string]interface{}{}).Optional().Comment("错误详情JSON"),
		field.JSON("validation_errors", []interface{}{}).Optional().Comment("验证错误JSON"),
		
		// 执行结果
		field.JSON("execution_result", map[string]interface{}{}).Optional().Comment("执行结果JSON"),
		field.JSON("performance_metrics", map[string]interface{}{}).Optional().Comment("性能指标JSON"),
		
		// 执行配置快照
		field.JSON("config_snapshot", map[string]interface{}{}).Optional().Comment("执行时的配置快照JSON"),
		field.JSON("provider_info", map[string]interface{}{}).Optional().Comment("提供者信息JSON"),
		
		// 扩展信息
		field.Text("execution_log").Optional().Comment("执行日志"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("扩展元数据JSON"),
		field.JSON("tags", []string{}).Optional().Comment("标签"),
	}
}

// Edges of the DiscoveryExecutionHistory.
func (DiscoveryExecutionHistory) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联发现配置
		edge.From("discovery_config", CiTypeDiscoveryConfig.Type).
			Ref("execution_histories").
			Field("discovery_config_id").
			Unique().
			Required().
			Comment("关联cmdb_ci_type_discovery_configs.id"),
	}
}

// Annotations 返回表的注释
func (DiscoveryExecutionHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_discovery_execution_histories"},
	}
}