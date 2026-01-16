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

// AggregationCache 聚合结果缓存表 - 存储预计算的聚合结果，提升查询性能
type AggregationCache struct {
	ent.Schema
}

func (AggregationCache) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{}, // 租户隔离
	}
}

func (AggregationCache) Fields() []ent.Field {
	return []ent.Field{
		// 聚合配置ID
		field.Uint64("config_id").
			Comment("关联dynamic_aggregation_config表的ID"),

		// 缓存键（用于快速查找）
		field.String("cache_key").
			MaxLen(512).
			Comment("缓存键，基于查询参数生成的唯一标识"),

		// 查询参数哈希
		field.String("params_hash").
			MaxLen(64).
			Comment("查询参数的MD5哈希值"),

		// 聚合结果（JSON格式存储）
		field.JSON("aggregation_result", map[string]interface{}{}).
			Comment("聚合计算结果，JSON格式存储"),

		// 结果元数据
		field.JSON("result_metadata", map[string]interface{}{}).
			Optional().
			Comment("结果元数据，如：记录数量、计算时间、数据来源等"),

		// 维度值映射（便于钻取查询）
		field.JSON("dimension_values", map[string]interface{}{}).
			Optional().
			Comment("维度值映射，用于支持钻取操作"),

		// 数据快照时间戳
		field.Time("data_snapshot_time").
			Default(time.Now).
			Comment("数据快照时间，表示缓存数据的时间点"),

		// 缓存创建时间（计算开始时间）
		field.Time("cache_created_at").
			Default(time.Now).
			Comment("缓存创建时间"),

		// 缓存过期时间
		field.Time("expires_at").
			Comment("缓存过期时间"),

		// 缓存状态
		field.Enum("cache_status").
			Values("building", "ready", "expired", "invalid", "error").
			Default("building").
			Comment("缓存状态"),

		// 计算耗时（毫秒）
		field.Uint64("computation_time_ms").
			Optional().
			Comment("聚合计算耗时（毫秒）"),

		// 访问次数
		field.Uint64("access_count").
			Default(0).
			Comment("缓存被访问的次数"),

		// 最后访问时间
		field.Time("last_accessed_at").
			Optional().
			Comment("最后访问时间"),

		// 命中率（针对相似查询）
		field.Float("hit_rate").
			Optional().
			Comment("缓存命中率"),

		// 数据源版本（用于检测数据变更）
		field.JSON("data_source_versions", map[string]interface{}{}).
			Optional().
			Comment("数据源版本信息，用于检测底层数据是否发生变更"),

		// 依赖的基础表列表
		field.JSON("dependent_tables", []string{}).
			Optional().
			Comment("依赖的基础表列表，用于缓存失效判断"),

		// 缓存大小（字节）
		field.Uint64("cache_size_bytes").
			Optional().
			Comment("缓存数据大小（字节）"),

		// 压缩标识
		field.Bool("is_compressed").
			Default(false).
			Comment("结果数据是否经过压缩"),

		// 压缩算法
		field.String("compression_algorithm").
			MaxLen(32).
			Optional().
			Comment("压缩算法：gzip, lz4, snappy等"),

		// 缓存层级
		field.Enum("cache_level").
			Values("l1_memory", "l2_redis", "l3_database").
			Default("l3_database").
			Comment("缓存层级：L1内存、L2Redis、L3数据库"),

		// 刷新策略
		field.Enum("refresh_strategy").
			Values("manual", "scheduled", "on_demand", "incremental").
			Default("scheduled").
			Comment("缓存刷新策略"),

		// 下次刷新时间
		field.Time("next_refresh_time").
			Optional().
			Comment("下次计划刷新时间"),

		// 刷新频率（秒）
		field.Uint64("refresh_interval_seconds").
			Optional().
			Comment("刷新间隔（秒）"),

		// 缓存优先级
		field.Enum("cache_priority").
			Values("low", "normal", "high", "critical").
			Default("normal").
			Comment("缓存优先级，影响内存分配和清理策略"),

		// 部门ID（继承自查询权限）
		field.Uint64("department_id").
			Optional().
			Comment("部门ID，用于权限隔离"),

		// 用户ID（缓存创建者）
		field.Uint64("user_id").
			Optional().
			Comment("缓存创建者用户ID"),

		// 共享级别
		field.Enum("sharing_level").
			Values("private", "department", "tenant", "public").
			Default("department").
			Comment("缓存共享级别"),

		// 错误信息（如果构建失败）
		field.String("error_message").
			MaxLen(1024).
			Optional().
			Comment("缓存构建错误信息"),

		// 错误堆栈
		field.Text("error_stack").
			Optional().
			Comment("错误堆栈信息"),

		// 重试次数
		field.Uint64("retry_count").
			Default(0).
			Comment("缓存构建重试次数"),

		// 最大重试次数
		field.Uint64("max_retry_count").
			Default(3).
			Comment("最大重试次数"),

		// 缓存标签（用于批量管理）
		field.JSON("cache_tags", []string{}).
			Optional().
			Comment("缓存标签，用于分类和批量操作"),

		// 是否启用增量更新
		field.Bool("incremental_update_enabled").
			Default(false).
			Comment("是否启用增量更新"),

		// 最后增量更新时间
		field.Time("last_incremental_update").
			Optional().
			Comment("最后增量更新时间"),

		// 增量更新检查点
		field.JSON("incremental_checkpoint", map[string]interface{}{}).
			Optional().
			Comment("增量更新检查点，用于跟踪数据变更位置"),
	}
}

func (AggregationCache) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + 缓存键唯一索引
		index.Fields("tenant_id", "cache_key").
			Unique(),

		// 租户 + 配置ID + 参数哈希唯一索引
		index.Fields("tenant_id", "config_id", "params_hash").
			Unique(),

		// 租户 + 缓存状态查询索引
		index.Fields("tenant_id", "cache_status", "expires_at"),

		// 租户 + 过期时间索引（用于清理过期缓存）
		index.Fields("tenant_id", "expires_at").
			Annotations(entsql.DescColumns("expires_at")),

		// 租户 + 配置ID查询索引
		index.Fields("tenant_id", "config_id", "cache_created_at").
			Annotations(entsql.DescColumns("cache_created_at")),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "sharing_level"),

		// 用户缓存查询索引
		index.Fields("tenant_id", "user_id", "cache_created_at").
			Annotations(entsql.DescColumns("cache_created_at")),

		// 访问频率分析索引
		index.Fields("tenant_id", "access_count").
			Annotations(entsql.DescColumns("access_count")),

		// 最后访问时间索引
		index.Fields("tenant_id", "last_accessed_at").
			Annotations(entsql.DescColumns("last_accessed_at")).
			Annotations(entsql.IndexWhere("last_accessed_at IS NOT NULL")),

		// 缓存层级查询索引
		index.Fields("tenant_id", "cache_level", "cache_priority"),

		// 刷新策略查询索引
		index.Fields("tenant_id", "refresh_strategy", "next_refresh_time"),

		// 下次刷新时间索引（用于定时任务）
		index.Fields("tenant_id", "next_refresh_time").
			Annotations(entsql.IndexWhere("next_refresh_time IS NOT NULL")),

		// 错误缓存查询索引
		index.Fields("tenant_id", "cache_status", "retry_count").
			Annotations(entsql.IndexWhere("cache_status = 'error'")),

		// 缓存大小索引（用于内存管理）
		index.Fields("tenant_id", "cache_size_bytes").
			Annotations(entsql.DescColumns("cache_size_bytes")).
			Annotations(entsql.IndexWhere("cache_size_bytes IS NOT NULL")),

		// 增量更新索引
		index.Fields("tenant_id", "incremental_update_enabled", "last_incremental_update").
			Annotations(entsql.IndexWhere("incremental_update_enabled = true")),

		// 计算性能分析索引
		index.Fields("tenant_id", "computation_time_ms").
			Annotations(entsql.DescColumns("computation_time_ms")).
			Annotations(entsql.IndexWhere("computation_time_ms IS NOT NULL")),

		// 数据快照时间索引
		index.Fields("tenant_id", "data_snapshot_time").
			Annotations(entsql.DescColumns("data_snapshot_time")),
	}
}

func (AggregationCache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_aggregation_cache",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='聚合结果缓存表 - 存储预计算聚合结果'",
		},
		// 建议按过期时间分区，便于自动清理
		entsql.WithComments(true),
	}
}