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

// CiStatisticsFact CI统计事实表 - 支持动态字段的OLAP分析
type CiStatisticsFact struct {
	ent.Schema
}

func (CiStatisticsFact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离
		mixins.DepartmentMixin{}, // 部门权限
	}
}

func (CiStatisticsFact) Fields() []ent.Field {
	return []ent.Field{
		// CI类型ID
		field.Uint64("ci_type_id").
			Comment("外键，关联cmdb_ci_types.id"),

		// CI统计数量
		field.Uint64("ci_count").
			Default(0).
			Comment("CI实例数量"),

		// 属性分布统计（JSONB格式，支持动态属性）
		field.JSON("attribute_distribution", map[string]interface{}{}).
			Optional().
			Comment("属性分布统计，格式: {\\\"memory\\\": {\\\"avg\\\": 16, \\\"max\\\": 128, \\\"min\\\": 4, \\\"count\\\": 100}}"),

		// 关系网络统计
		field.JSON("relation_metrics", map[string]interface{}{}).
			Optional().
			Comment("关系统计指标，格式: {\\\"host_to_vm\\\": 10, \\\"vm_to_storage\\\": 5}"),

		// 生命周期阶段
		field.Enum("lifecycle_stage").
			Values("development", "testing", "staging", "production", "retired").
			Default("production").
			Comment("生命周期阶段"),

		// 状态分布
		field.JSON("status_distribution", map[string]interface{}{}).
			Optional().
			Comment("状态分布统计，格式: {\\\"active\\\": 80, \\\"inactive\\\": 20}"),

		// 部门分布
		field.JSON("department_distribution", map[string]interface{}{}).
			Optional().
			Comment("部门分布统计"),

		// 用户行为统计
		field.JSON("user_activity", map[string]interface{}{}).
			Optional().
			Comment("用户操作活动统计"),

		// 统计时间窗口
		field.Enum("time_window").
			Values("hourly", "daily", "weekly", "monthly", "yearly").
			Default("daily").
			Comment("统计时间窗口"),

		// 统计基准时间
		field.Time("stat_time").
			Default(time.Now).
			Comment("统计数据的基准时间"),

		// 数据版本号（支持增量更新）
		field.Uint64("version").
			Default(1).
			Comment("统计数据版本号"),

		// 计算状态
		field.Enum("compute_status").
			Values("pending", "computing", "completed", "failed").
			Default("pending").
			Comment("统计计算状态"),

		// 原始数据Hash（用于检测数据变化）
		field.String("data_hash").
			Optional().
			Comment("原始数据Hash值，用于增量更新检测"),
	}
}

func (CiStatisticsFact) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + CI类型 + 统计时间的唯一索引（防重复统计）
		index.Fields("tenant_id", "ci_type_id", "time_window", "stat_time").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),

		// 租户 + 生命周期阶段查询索引
		index.Fields("tenant_id", "lifecycle_stage", "stat_time"),

		// 租户 + 计算状态查询索引
		index.Fields("tenant_id", "compute_status", "created_at"),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "ci_type_id", "stat_time"),

		// 时间范围查询索引
		index.Fields("tenant_id", "stat_time").
			Annotations(entsql.DescColumns("stat_time")),

		// 版本查询索引（支持增量更新）
		index.Fields("tenant_id", "ci_type_id", "version"),
	}
}

func (CiStatisticsFact) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_ci_statistics_fact",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='CI统计事实表 - 支持动态属性的OLAP分析'",
		},
		// 支持按时间分区（如果使用MySQL 8.0+）
		entsql.WithComments(true),
	}
}