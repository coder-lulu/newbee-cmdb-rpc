package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
	"github.com/gofrs/uuid/v5"
)

// CiLifecycleState holds the schema definition for the CiLifecycleState entity.
// CI生命周期状态表 - 管理CI数据的生命周期状态
type CiLifecycleState struct {
	ent.Schema
}

// Mixin of the CiLifecycleState.
func (CiLifecycleState) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
		mixins.CreatedByMixin{},
	}
}

// Fields of the CiLifecycleState.
func (CiLifecycleState) Fields() []ent.Field {
	return []ent.Field{
		// 状态基础信息
		field.String("state_id").
			Comment("状态ID，全局唯一标识").
			Unique(),

		field.String("state_name").
			Comment("状态名称"),

		field.String("state_code").
			Comment("状态编码"),

		field.Text("state_description").
			Optional().
			Comment("状态描述"),

		// 关联CI信息
		field.Uint64("ci_id").
			Comment("CI实例ID"),

		field.Uint64("ci_type_id").
			Comment("CI类型ID"),

		// 状态类型和级别
		field.Enum("state_type").
			Values("draft", "active", "inactive", "pending", "approved", "rejected", "suspended", "archived", "deleted", "expired").
			Comment("状态类型"),

		field.Enum("state_category").
			Values("data", "approval", "lifecycle", "security", "business").
			Comment("状态分类"),

		field.Int("state_level").
			Default(0).
			Comment("状态级别，数值越大级别越高"),

		// 状态机制
		field.JSON("allowed_transitions", []string{}).
			Optional().
			Comment("允许的状态转换列表"),

		field.JSON("transition_conditions", map[string]interface{}{}).
			Optional().
			Comment("状态转换条件"),

		field.JSON("auto_transition_rules", map[string]interface{}{}).
			Optional().
			Comment("自动转换规则"),

		// 时间控制
		field.Time("entered_at").
			Comment("进入状态时间"),

		field.Time("expected_exit_at").
			Optional().
			Comment("预期退出时间"),

		field.Time("actual_exit_at").
			Optional().
			Comment("实际退出时间"),

		field.Int("duration_limit_hours").
			Optional().
			Comment("状态持续时间限制（小时）"),

		field.Bool("is_timeout").
			Default(false).
			Comment("是否已超时"),

		field.Time("timeout_at").
			Optional().
			Comment("超时时间"),

		// 状态触发信息
		field.Enum("trigger_type").
			Values("manual", "auto", "system", "scheduled", "event", "workflow").
			Comment("触发类型"),

		field.String("trigger_source").
			Optional().
			Comment("触发源"),

		field.JSON("trigger_context", map[string]interface{}{}).
			Optional().
			Comment("触发上下文"),

		field.UUID("triggered_by", uuid.UUID{}).
			Optional().
			Comment("触发人ID"),

		field.String("triggered_by_name").
			Optional().
			Comment("触发人姓名"),

		// 状态数据和配置
		field.JSON("state_data", map[string]interface{}{}).
			Optional().
			Comment("状态相关数据"),

		field.JSON("state_config", map[string]interface{}{}).
			Optional().
			Comment("状态配置参数"),

		field.JSON("validation_rules", map[string]interface{}{}).
			Optional().
			Comment("状态验证规则"),

		// 权限和访问控制
		field.JSON("required_permissions", []string{}).
			Optional().
			Comment("进入状态所需权限"),

		field.JSON("granted_permissions", []string{}).
			Optional().
			Comment("状态下获得的权限"),

		field.JSON("restricted_operations", []string{}).
			Optional().
			Comment("状态下限制的操作"),

		// 通知和提醒
		field.JSON("notification_config", map[string]interface{}{}).
			Optional().
			Comment("通知配置"),

		field.JSON("reminder_config", map[string]interface{}{}).
			Optional().
			Comment("提醒配置"),

		field.Time("last_notification_at").
			Optional().
			Comment("最后通知时间"),

		field.Int("notification_count").
			Default(0).
			Comment("通知次数"),

		// 审批相关
		field.Bool("require_approval").
			Default(false).
			Comment("是否需要审批"),

		field.String("approval_flow_id").
			Optional().
			Comment("审批流程ID"),

		field.Enum("approval_status").
			Values("pending", "approved", "rejected", "cancelled").
			Optional().
			Comment("审批状态"),

		field.UUID("approver_id", uuid.UUID{}).
			Optional().
			Comment("审批人ID"),

		field.String("approver_name").
			Optional().
			Comment("审批人姓名"),

		field.Time("approved_at").
			Optional().
			Comment("审批时间"),

		field.Text("approval_comment").
			Optional().
			Comment("审批意见"),

		// 错误和异常处理
		field.Bool("has_error").
			Default(false).
			Comment("是否有错误"),

		field.Text("error_message").
			Optional().
			Comment("错误信息"),

		field.JSON("error_details", map[string]interface{}{}).
			Optional().
			Comment("错误详情"),

		field.Int("retry_count").
			Default(0).
			Comment("重试次数"),

		field.Time("last_retry_at").
			Optional().
			Comment("最后重试时间"),

		// 性能监控
		field.Int("processing_duration").
			Optional().
			Comment("处理耗时（毫秒）"),

		field.JSON("performance_metrics", map[string]interface{}{}).
			Optional().
			Comment("性能指标"),

		field.JSON("resource_usage", map[string]interface{}{}).
			Optional().
			Comment("资源使用情况"),

		// 关联关系
		field.String("parent_state_id").
			Optional().
			Comment("父状态ID"),

		field.JSON("child_state_ids", []string{}).
			Optional().
			Comment("子状态ID列表"),

		field.String("related_operation_id").
			Optional().
			Comment("关联操作ID"),

		// 版本和历史
		field.Int("version").
			Default(1).
			Comment("状态版本"),

		field.JSON("change_history", []map[string]interface{}{}).
			Optional().
			Comment("变更历史"),

		field.Bool("is_rollback").
			Default(false).
			Comment("是否为回退状态"),

		field.String("rollback_from_state_id").
			Optional().
			Comment("回退前状态ID"),

		// 业务标记
		field.Bool("is_milestone").
			Default(false).
			Comment("是否为里程碑状态"),

		field.Bool("is_critical").
			Default(false).
			Comment("是否为关键状态"),

		field.Bool("is_reversible").
			Default(true).
			Comment("是否可逆转"),

		field.Bool("is_final").
			Default(false).
			Comment("是否为最终状态"),

		// 扩展信息
		field.JSON("custom_attributes", map[string]interface{}{}).
			Optional().
			Comment("自定义属性"),

		field.JSON("integration_data", map[string]interface{}{}).
			Optional().
			Comment("外部系统集成数据"),

		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("元数据信息"),

		field.JSON("tags", []string{}).
			Optional().
			Comment("标签列表"),

		field.Text("comments").
			Optional().
			Comment("备注信息"),

		// 审计字段
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Comment("最后更新人ID"),
	}
}

// Indexes of the CiLifecycleState.
func (CiLifecycleState) Indexes() []ent.Index {
	return []ent.Index{
		// CI相关索引
		index.Fields("ci_id", "state_type"),
		index.Fields("ci_id", "entered_at"),
		index.Fields("ci_type_id", "state_type"),

		// 状态类型索引
		index.Fields("state_type", "state_category"),
		index.Fields("state_code"),
		index.Fields("state_level", "state_type"),

		// 时间相关索引
		index.Fields("entered_at"),
		index.Fields("expected_exit_at"),
		index.Fields("timeout_at"),
		index.Fields("is_timeout", "timeout_at"),

		// 触发相关索引
		index.Fields("trigger_type", "entered_at"),
		index.Fields("triggered_by", "entered_at"),

		// 审批相关索引
		index.Fields("require_approval", "approval_status"),
		index.Fields("approval_flow_id", "approval_status"),
		index.Fields("approver_id", "approved_at"),

		// 错误和状态索引
		index.Fields("has_error", "retry_count"),
		index.Fields("state_type", "has_error"),

		// 关联关系索引
		index.Fields("parent_state_id"),
		index.Fields("related_operation_id"),

		// 业务标记索引
		index.Fields("is_milestone", "entered_at"),
		index.Fields("is_critical", "state_type"),
		index.Fields("is_final", "state_type"),

		// 复合查询索引
		index.Fields("ci_id", "state_type", "entered_at"),
		index.Fields("state_type", "approval_status", "entered_at"),
		index.Fields("trigger_type", "state_type", "entered_at"),
	}
}

// Annotations 返回表的注释
func (CiLifecycleState) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_lifecycle_states"},
	}
}
