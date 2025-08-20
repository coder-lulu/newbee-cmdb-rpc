package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
	"github.com/gofrs/uuid/v5"
)

// CiOperation holds the schema definition for the CiOperation entity.
// CI操作记录表 - 记录所有CI数据操作的详细信息
type CiOperation struct {
	ent.Schema
}

// Mixin of the CiOperation.
func (CiOperation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiOperation.
func (CiOperation) Fields() []ent.Field {
	return []ent.Field{
		// 操作基础信息
		field.String("operation_id").
			Comment("操作ID，全局唯一标识").
			Unique(),

		field.Enum("operation_type").
			Values("create", "update", "delete", "batch_create", "batch_update", "batch_delete", "import", "sync").
			Comment("操作类型"),

		field.Enum("operation_status").
			Values("pending", "validating", "processing", "waiting_approval", "approved", "rejected", "completed", "failed").
			Default("pending").
			Comment("操作状态"),

		// CI相关信息
		field.Uint64("ci_id").
			Optional().
			Comment("CI实例ID，批量操作时可为空"),

		field.Uint64("ci_type_id").
			Comment("CI类型ID"),

		// 操作者信息
		field.UUID("operator_id", uuid.UUID{}).
			Comment("操作者ID"),

		field.String("operator_name").
			Comment("操作者姓名"),

		field.String("operator_role").
			Optional().
			Comment("操作者角色"),

		field.String("operator_department").
			Optional().
			Comment("操作者部门"),

		// 操作来源和上下文
		field.Enum("operation_source").
			Values("manual", "api", "import", "sync", "automation", "system").
			Default("manual").
			Comment("操作来源"),

		field.String("source_detail").
			Optional().
			Comment("来源详情，如文件名、API调用方等"),

		field.String("operation_reason").
			Optional().
			Comment("操作原因说明"),

		field.JSON("operation_context", map[string]interface{}{}).
			Optional().
			Comment("操作上下文信息"),

		// 数据内容
		field.JSON("data_before", map[string]interface{}{}).
			Optional().
			Comment("操作前数据快照"),

		field.JSON("data_after", map[string]interface{}{}).
			Optional().
			Comment("操作后数据快照"),

		field.JSON("affected_attributes", []uint64{}).
			Optional().
			Comment("受影响的属性ID列表"),

		// 批量操作信息
		field.JSON("batch_ci_ids", []uint64{}).
			Optional().
			Comment("批量操作涉及的CI ID列表"),

		field.Int("batch_total").
			Optional().
			Default(0).
			Comment("批量操作总数"),

		field.Int("batch_success").
			Optional().
			Default(0).
			Comment("批量操作成功数"),

		field.Int("batch_failed").
			Optional().
			Default(0).
			Comment("批量操作失败数"),

		// 审批流程信息
		field.Bool("require_approval").
			Default(false).
			Comment("是否需要审批"),

		field.String("approval_flow_id").
			Optional().
			Comment("审批流程ID"),

		field.UUID("approver_id", uuid.UUID{}).
			Optional().
			Comment("审批人ID"),

		field.String("approver_name").
			Optional().
			Comment("审批人姓名"),

		field.Time("approved_at").
			Optional().
			Comment("审批时间"),

		field.String("approval_comment").
			Optional().
			Comment("审批意见"),

		// 执行结果
		field.Time("started_at").
			Optional().
			Comment("开始执行时间"),

		field.Time("completed_at").
			Optional().
			Comment("完成时间"),

		field.Int("execution_duration").
			Optional().
			Comment("执行耗时（毫秒）"),

		field.JSON("execution_result", map[string]interface{}{}).
			Optional().
			Comment("执行结果详情"),

		field.Text("error_message").
			Optional().
			Comment("错误信息"),

		field.JSON("error_details", map[string]interface{}{}).
			Optional().
			Comment("错误详情"),

		// 生命周期管理
		field.Enum("lifecycle_stage").
			Values("draft", "submitted", "validated", "approved", "executed", "completed", "cancelled", "expired").
			Default("draft").
			Comment("生命周期阶段"),

		field.Time("expires_at").
			Optional().
			Comment("过期时间"),

		// 权限和安全
		field.JSON("required_permissions", []string{}).
			Optional().
			Comment("所需权限列表"),

		field.JSON("permission_check_result", map[string]interface{}{}).
			Optional().
			Comment("权限检查结果"),

		field.Bool("is_sensitive").
			Default(false).
			Comment("是否包含敏感数据"),

		// 追溯和关联
		field.String("parent_operation_id").
			Optional().
			Comment("父操作ID，用于关联相关操作"),

		field.JSON("child_operation_ids", []string{}).
			Optional().
			Comment("子操作ID列表"),

		field.String("transaction_id").
			Optional().
			Comment("事务ID，用于关联同一事务的多个操作"),

		// 扩展字段
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("元数据信息"),

		field.JSON("tags", []string{}).
			Optional().
			Comment("标签列表"),
	}
}

// Indexes of the CiOperation.
func (CiOperation) Indexes() []ent.Index {
	return []ent.Index{
		// 主要查询索引
		index.Fields("operation_type", "operation_status", "created_at"),
		index.Fields("ci_id", "operation_type", "created_at"),
		index.Fields("ci_type_id", "operation_type", "created_at"),
		index.Fields("operator_id", "created_at"),

		// 状态和生命周期索引
		index.Fields("operation_status", "lifecycle_stage"),
		index.Fields("require_approval", "operation_status"),

		// 时间范围查询索引
		index.Fields("started_at"),
		index.Fields("completed_at"),
		index.Fields("expires_at"),

		// 批量操作索引
		index.Fields("operation_source", "created_at"),
		index.Fields("transaction_id"),

		// 关联操作索引
		index.Fields("parent_operation_id"),

		// 复合索引用于复杂查询
		index.Fields("operation_type", "ci_type_id", "operation_status", "created_at"),
		index.Fields("operator_id", "operation_type", "created_at"),
	}
}
