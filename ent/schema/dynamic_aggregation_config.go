package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// DynamicAggregationConfig 动态聚合配置表 - 支持灵活配置统计分析维度和指标
type DynamicAggregationConfig struct {
	ent.Schema
}

func (DynamicAggregationConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离
		mixins.DepartmentMixin{}, // 部门权限
	}
}

func (DynamicAggregationConfig) Fields() []ent.Field {
	return []ent.Field{
		// 配置名称
		field.String("config_name").
			MaxLen(128).
			Comment("聚合配置名称"),

		// 配置别名
		field.String("config_alias").
			MaxLen(64).
			Optional().
			Comment("配置别名，用于显示"),

		// 配置描述
		field.String("description").
			MaxLen(512).
			Optional().
			Comment("配置描述"),

		// 配置类型
		field.Enum("config_type").
			Values("dashboard", "report", "alert", "custom").
			Comment("配置类型：仪表板、报表、告警、自定义"),

		// 数据源配置（支持多表关联）
		field.JSON("data_sources", []map[string]interface{}{}).
			Comment("数据源配置，格式: [{JSON格式}"),

		// 维度配置（支持多维度分析）
		field.JSON("dimensions", []map[string]interface{}{}).
			Comment("维度配置，格式: [{JSON格式}"),

		// 度量配置（支持多种聚合函数）
		field.JSON("measures", []map[string]interface{}{}).
			Comment("度量配置，格式: [{JSON格式}"),

		// 过滤条件（支持复杂过滤逻辑）
		field.JSON("filters", map[string]interface{}{}).
			Optional().
			Comment("过滤条件配置，支持动态属性和关系过滤"),

		// 排序配置
		field.JSON("sort_config", []map[string]interface{}{}).
			Optional().
			Comment("排序配置，格式: [{JSON格式}"),

		// 分页配置
		field.JSON("pagination_config", map[string]interface{}{}).
			Optional().
			Comment("分页配置，格式: {JSON格式}"),

		// 时间维度配置
		field.JSON("time_dimension_config", map[string]interface{}{}).
			Optional().
			Comment("时间维度配置，格式: {JSON格式}"),

		// 钻取配置（支持多级钻取）
		field.JSON("drill_down_config", []map[string]interface{}{}).
			Optional().
			Comment("钻取配置，支持从汇总数据钻取到明细数据"),

		// 输出格式配置
		field.JSON("output_format", map[string]interface{}{}).
			Optional().
			Comment("输出格式配置，支持表格、图表、JSON等多种格式"),

		// 缓存配置
		field.JSON("cache_config", map[string]interface{}{}).
			Optional().
			Comment("缓存配置，格式: {JSON格式}"),

		// 权限配置
		field.JSON("permission_config", map[string]interface{}{}).
			Optional().
			Comment("权限配置，控制不同角色的数据可见性"),

		// 动态属性映射（EAV模式支持）
		field.JSON("dynamic_attribute_mapping", map[string]interface{}{}).
			Optional().
			Comment("动态属性映射配置，支持EAV模式的属性查询"),

		// 计算字段配置（支持公式计算）
		field.JSON("calculated_fields", []map[string]interface{}{}).
			Optional().
			Comment("计算字段配置，支持基于现有字段的公式计算"),

		// 预聚合配置
		field.JSON("pre_aggregation_config", map[string]interface{}{}).
			Optional().
			Comment("预聚合配置，支持定时预计算常用聚合结果"),

		// 实时性要求
		field.Enum("freshness_requirement").
			Values("real_time", "near_real_time", "batch", "historical").
			Default("near_real_time").
			Comment("数据实时性要求"),

		// 查询复杂度评估
		field.Enum("complexity_level").
			Values("simple", "medium", "complex", "very_complex").
			Default("medium").
			Comment("查询复杂度等级"),

		// 预计查询时间（秒）
		field.Float("estimated_query_time").
			Optional().
			Comment("预计查询执行时间（秒）"),

		// 数据量估算
		field.Uint64("estimated_data_size").
			Optional().
			Comment("预计处理的数据量"),

		// 配置状态
		field.Enum("status").
			Values("active", "inactive", "testing", "deprecated").
			Default("testing").
			Comment("配置状态"),

		// 配置版本
		field.Uint64("config_version").
			Default(1).
			Comment("配置版本号"),

		// 最后执行时间
		field.Time("last_executed_at").
			Optional().
			Comment("最后执行时间"),

		// 执行次数统计
		field.Uint64("execution_count").
			Default(0).
			Comment("配置被执行的次数"),

		// 平均执行时间（毫秒）
		field.Float("avg_execution_time").
			Optional().
			Comment("平均执行时间（毫秒）"),

		// 最大执行时间（毫秒）
		field.Float("max_execution_time").
			Optional().
			Comment("最大执行时间（毫秒）"),

		// 错误次数
		field.Uint64("error_count").
			Default(0).
			Comment("执行错误次数"),

		// 最后错误信息
		field.String("last_error_message").
			MaxLen(1024).
			Optional().
			Comment("最后一次错误信息"),

		// 创建者ID
		field.Uint64("creator_id").
			Comment("配置创建者ID"),

		// 创建者名称（冗余存储）
		field.String("creator_name").
			MaxLen(128).
			Comment("配置创建者姓名"),

		// 标签（用于分类和搜索）
		field.JSON("tags", []string{}).
			Optional().
			Comment("配置标签"),

		// 是否为系统预定义配置
		field.Bool("is_system_config").
			Default(false).
			Comment("是否为系统预定义配置"),

		// 配置的有效期
		field.Time("valid_until").
			Optional().
			Comment("配置有效期截止时间"),

		// 监控配置
		field.JSON("monitoring_config", map[string]interface{}{}).
			Optional().
			Comment("监控配置，如性能阈值、告警规则等"),
	}
}

func (DynamicAggregationConfig) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + 配置名称唯一索引
		index.Fields("tenant_id", "config_name").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),

		// 租户 + 配置类型查询索引
		index.Fields("tenant_id", "config_type", "status"),

		// 租户 + 状态查询索引
		index.Fields("tenant_id", "status", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "config_type"),

		// 创建者查询索引
		index.Fields("tenant_id", "creator_id", "created_at").
			Annotations(entsql.DescColumns("created_at")),

		// 实时性要求查询索引
		index.Fields("tenant_id", "freshness_requirement"),

		// 复杂度查询索引
		index.Fields("tenant_id", "complexity_level"),

		// 执行频率分析索引
		index.Fields("tenant_id", "execution_count").
			Annotations(entsql.DescColumns("execution_count")),

		// 最后执行时间索引
		index.Fields("tenant_id", "last_executed_at").
			Annotations(entsql.DescColumns("last_executed_at")).
			Annotations(entsql.IndexWhere("last_executed_at IS NOT NULL")),

		// 性能分析索引
		index.Fields("tenant_id", "avg_execution_time").
			Annotations(entsql.DescColumns("avg_execution_time")).
			Annotations(entsql.IndexWhere("avg_execution_time IS NOT NULL")),

		// 错误统计索引
		index.Fields("tenant_id", "error_count").
			Annotations(entsql.DescColumns("error_count")).
			Annotations(entsql.IndexWhere("error_count > 0")),

		// 系统配置查询索引
		index.Fields("tenant_id", "is_system_config", "status"),

		// 有效期查询索引
		index.Fields("tenant_id", "valid_until").
			Annotations(entsql.IndexWhere("valid_until IS NOT NULL")),

		// 配置版本索引
		index.Fields("tenant_id", "config_name", "config_version").
			Annotations(entsql.DescColumns("config_version")),
	}
}

func (DynamicAggregationConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_dynamic_aggregation_config",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='动态聚合配置表 - 支持灵活配置统计分析'",
		},
	}
}