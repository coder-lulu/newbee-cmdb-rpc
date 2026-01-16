package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// UserActivityFact 用户活动统计事实表 - 追踪用户在CMDB系统中的操作行为
type UserActivityFact struct {
	ent.Schema
}

func (UserActivityFact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离
		mixins.DepartmentMixin{}, // 部门权限
	}
}

func (UserActivityFact) Fields() []ent.Field {
	return []ent.Field{
		// 用户ID
		field.Uint64("user_id").
			Comment("操作用户ID"),

		// 用户名（冗余存储，提升查询性能）
		field.String("username").
			MaxLen(128).
			Comment("用户名"),

		// 用户角色信息
		field.JSON("user_roles", []string{}).
			Optional().
			Comment("用户角色列表"),

		// 操作类型
		field.Enum("operation_type").
			Values("create", "read", "update", "delete", "batch_update", "batch_delete", "import", "export", "search", "login", "logout").
			Comment("操作类型"),

		// 资源类型（操作的对象类型）
		field.String("resource_type").
			MaxLen(64).
			Comment("资源类型：ci, ci_type, attribute, relation等"),

		// 资源ID（可选，某些操作如search没有具体资源ID）
		field.Uint64("resource_id").
			Optional().
			Comment("操作的资源ID"),

		// CI类型ID（如果操作涉及CI）
		field.Uint64("ci_type_id").
			Optional().
			Comment("相关CI类型ID"),

		// 操作详情（JSON格式存储操作参数）
		field.JSON("operation_details", map[string]interface{}{}).
			Optional().
			Comment("操作详情，如：修改的字段、搜索条件、批量操作的数量等"),

		// 操作结果状态
		field.Enum("status").
			Values("success", "failed", "partial", "unauthorized", "forbidden").
			Comment("操作结果状态"),

		// 错误信息（如果操作失败）
		field.String("error_message").
			MaxLen(512).
			Optional().
			Comment("错误信息"),

		// 响应时间（毫秒）
		field.Uint64("response_time_ms").
			Optional().
			Comment("操作响应时间（毫秒）"),

		// 影响的记录数量
		field.Uint64("affected_records").
			Default(0).
			Comment("操作影响的记录数量"),

		// IP地址
		field.String("ip_address").
			MaxLen(45). // IPv6最大长度
			Optional().
			Comment("客户端IP地址"),

		// User Agent信息
		field.String("user_agent").
			MaxLen(512).
			Optional().
			Comment("客户端User Agent信息"),

		// 会话ID
		field.String("session_id").
			MaxLen(128).
			Optional().
			Comment("会话标识"),

		// API路径
		field.String("api_path").
			MaxLen(256).
			Optional().
			Comment("API请求路径"),

		// HTTP方法
		field.Enum("http_method").
			Values("GET", "POST", "PUT", "DELETE", "PATCH").
			Optional().
			Comment("HTTP请求方法"),

		// 数据变更前快照（重要操作的数据快照）
		field.JSON("before_snapshot", map[string]interface{}{}).
			Optional().
			Comment("变更前的数据快照（仅记录关键操作）"),

		// 数据变更后快照
		field.JSON("after_snapshot", map[string]interface{}{}).
			Optional().
			Comment("变更后的数据快照（仅记录关键操作）"),

		// 操作时间（精确到秒的时间戳）
		field.Time("operation_time").
			Default(time.Now).
			Comment("操作发生时间"),

		// 操作时间维度ID（关联时间维度表）
		field.Uint64("time_dimension_id").
			Optional().
			Comment("关联时间维度表ID"),

		// 业务标签（支持业务层面的操作分类）
		field.JSON("business_tags", []string{}).
			Optional().
			Comment("业务标签，用于业务层面的操作分类和统计"),

		// 风险等级评估
		field.Enum("risk_level").
			Values("low", "medium", "high", "critical").
			Default("low").
			Comment("操作风险等级评估"),

		// 合规性标记
		field.JSON("compliance_flags", []string{}).
			Optional().
			Comment("合规性相关标记，如：audit_required, sensitive_data等"),

		// 操作来源
		field.Enum("operation_source").
			Values("web_ui", "api", "batch_job", "system", "import", "sync").
			Default("web_ui").
			Comment("操作来源渠道"),

		// 地理位置信息（基于IP解析）
		field.JSON("geo_location", map[string]interface{}{}).
			Optional().
			Comment("地理位置信息，格式: {JSON格式}"),

		// 设备信息
		field.JSON("device_info", map[string]interface{}{}).
			Optional().
			Comment("设备信息，格式: {JSON格式}"),

		// 批次ID（用于关联批量操作）
		field.String("batch_id").
			MaxLen(64).
			Optional().
			Comment("批量操作批次标识"),

		// 父操作ID（用于追踪关联操作）
		field.Uint64("parent_operation_id").
			Optional().
			Comment("父操作ID，用于追踪操作链"),

		// 数据版本号（用于追踪数据变更历史）
		field.Uint64("data_version").
			Default(1).
			Comment("数据版本号"),
	}
}

func (UserActivityFact) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + 用户 + 操作时间索引
		index.Fields("tenant_id", "user_id", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 租户 + 操作类型 + 操作时间索引
		index.Fields("tenant_id", "operation_type", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 租户 + 资源类型 + 操作时间索引
		index.Fields("tenant_id", "resource_type", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 租户 + CI类型 + 操作时间索引
		index.Fields("tenant_id", "ci_type_id", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 租户 + 状态 + 操作时间索引
		index.Fields("tenant_id", "status", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 用户活动分析索引
		index.Fields("tenant_id", "user_id", "operation_type", "status"),

		// 资源访问分析索引
		index.Fields("tenant_id", "resource_type", "resource_id", "operation_time"),

		// 时间维度关联索引
		index.Fields("tenant_id", "time_dimension_id"),

		// 风险等级查询索引
		index.Fields("tenant_id", "risk_level", "operation_time").
			Annotations(entsql.DescColumns("operation_time")),

		// 操作来源分析索引
		index.Fields("tenant_id", "operation_source", "operation_time"),

		// IP地址追踪索引
		index.Fields("tenant_id", "ip_address", "operation_time"),

		// 会话分析索引
		index.Fields("tenant_id", "session_id", "operation_time"),

		// 批量操作追踪索引
		index.Fields("tenant_id", "batch_id").
			Annotations(entsql.IndexWhere("batch_id IS NOT NULL")),

		// 父子操作关联索引
		index.Fields("tenant_id", "parent_operation_id").
			Annotations(entsql.IndexWhere("parent_operation_id IS NOT NULL")),

		// API路径分析索引
		index.Fields("tenant_id", "api_path", "http_method", "operation_time"),

		// 响应时间性能分析索引
		index.Fields("tenant_id", "response_time_ms").
			Annotations(entsql.DescColumns("response_time_ms")).
			Annotations(entsql.IndexWhere("response_time_ms IS NOT NULL")),

		// 错误操作分析索引
		index.Fields("tenant_id", "status", "error_message").
			Annotations(entsql.IndexWhere("status IN ('failed', 'partial', 'unauthorized', 'forbidden') AND error_message IS NOT NULL")),
	}
}

func (UserActivityFact) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_user_activity_fact",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='用户活动统计事实表 - 追踪CMDB系统操作行为'",
		},
		// 建议按操作时间分区（MySQL 8.0+）
		entsql.WithComments(true),
	}
}