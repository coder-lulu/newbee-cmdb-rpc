package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
	"github.com/gofrs/uuid/v5"
)

// CiApprovalFlow holds the schema definition for the CiApprovalFlow entity.
// CI审批流程表 - 管理CI数据变更的审批流程
type CiApprovalFlow struct {
	ent.Schema
}

// Mixin of the CiApprovalFlow.
func (CiApprovalFlow) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiApprovalFlow.
func (CiApprovalFlow) Fields() []ent.Field {
	return []ent.Field{
		// 流程基础信息
		field.String("flow_id").
			Comment("流程ID，全局唯一标识").
			Unique(),

		field.String("flow_name").
			Comment("流程名称"),

		field.String("flow_code").
			Optional().
			Comment("流程编码"),

		field.Text("flow_description").
			Optional().
			Comment("流程描述"),

		// 流程范围和触发条件
		field.Enum("scope_type").
			Values("global", "ci_type", "operation_type", "risk_level", "value_range", "custom").
			Comment("流程适用范围类型"),

		field.JSON("scope_config", map[string]interface{}{}).
			Optional().
			Comment("范围配置：CI类型、操作类型、风险等级等"),

		field.JSON("trigger_conditions", map[string]interface{}{}).
			Optional().
			Comment("触发条件：数据变更阈值、字段变更等"),

		// 流程配置
		field.Enum("flow_type").
			Values("sequential", "parallel", "hybrid", "auto").
			Default("sequential").
			Comment("流程类型：顺序、并行、混合、自动"),

		field.JSON("approval_stages", []map[string]interface{}{}).
			Comment("审批阶段配置"),

		field.Int("total_stages").
			Comment("总阶段数"),

		field.Bool("allow_skip_stages").
			Default(false).
			Comment("是否允许跳过阶段"),

		field.Bool("allow_rollback").
			Default(true).
			Comment("是否允许回退"),

		// 审批人配置
		field.JSON("approver_config", map[string]interface{}{}).
			Comment("审批人配置：角色、部门、指定人员等"),

		field.JSON("fallback_approvers", []map[string]interface{}{}).
			Optional().
			Comment("备用审批人配置"),

		field.Bool("require_all_approvers").
			Default(false).
			Comment("是否需要所有审批人同意"),

		// 时间限制
		field.Int("timeout_hours").
			Default(24).
			Comment("超时时间（小时）"),

		field.JSON("stage_timeouts", map[string]interface{}{}).
			Optional().
			Comment("各阶段超时配置"),

		field.Enum("timeout_action").
			Values("auto_approve", "auto_reject", "escalate", "notify").
			Default("notify").
			Comment("超时处理动作"),

		// 通知配置
		field.JSON("notification_config", map[string]interface{}{}).
			Optional().
			Comment("通知配置：邮件、短信、系统消息等"),

		field.Bool("notify_on_submit").
			Default(true).
			Comment("提交时是否通知"),

		field.Bool("notify_on_approve").
			Default(true).
			Comment("审批通过时是否通知"),

		field.Bool("notify_on_reject").
			Default(true).
			Comment("审批拒绝时是否通知"),

		// 流程状态
		field.Enum("status").
			Values("active", "inactive", "draft", "archived").
			Default("draft").
			Comment("流程状态"),

		field.String("status_reason").
			Optional().
			Comment("状态变更原因"),

		// 版本控制
		field.String("version").
			Default("1.0").
			Comment("流程版本"),

		field.String("parent_flow_id").
			Optional().
			Comment("父流程ID，用于版本追踪"),

		field.Bool("is_default").
			Default(false).
			Comment("是否为默认流程"),

		// 流程统计
		field.Int("usage_count").
			Default(0).
			Comment("使用次数"),

		field.Int("approval_rate").
			Default(0).
			Comment("通过率（百分比）"),

		field.Float("avg_approval_time").
			Default(0).
			Comment("平均审批时间（小时）"),

		field.Time("last_used_at").
			Optional().
			Comment("最后使用时间"),

		// 审计信息
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Comment("创建人ID"),

		field.String("created_by_name").
			Optional().
			Comment("创建人姓名"),

		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Comment("最后更新人ID"),

		field.String("updated_by_name").
			Optional().
			Comment("最后更新人姓名"),

		field.Time("published_at").
			Optional().
			Comment("发布时间"),

		field.UUID("published_by", uuid.UUID{}).
			Optional().
			Comment("发布人ID"),

		// 扩展配置
		field.JSON("custom_fields", map[string]interface{}{}).
			Optional().
			Comment("自定义字段配置"),

		field.JSON("integration_config", map[string]interface{}{}).
			Optional().
			Comment("外部系统集成配置"),

		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("元数据信息"),

		field.JSON("tags", []string{}).
			Optional().
			Comment("标签列表"),

		field.Text("comments").
			Optional().
			Comment("备注信息"),
	}
}

// Indexes of the CiApprovalFlow.
func (CiApprovalFlow) Indexes() []ent.Index {
	return []ent.Index{
		// 基础查询索引
		index.Fields("flow_code"),
		index.Fields("scope_type", "status"),
		index.Fields("flow_type", "status"),
		index.Fields("status", "is_default"),

		// 使用情况索引
		index.Fields("usage_count", "status"),
		index.Fields("last_used_at"),
		index.Fields("approval_rate"),

		// 版本控制索引
		index.Fields("parent_flow_id", "version"),
		index.Fields("version", "status"),

		// 审计索引
		index.Fields("created_by", "created_at"),
		index.Fields("updated_by", "updated_at"),
		index.Fields("published_by", "published_at"),

		// 复合查询索引
		index.Fields("scope_type", "is_default", "status"),
		index.Fields("flow_type", "total_stages", "status"),
	}
}

// Annotations 返回表的注释
func (CiApprovalFlow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_approval_flows"},
	}
}
