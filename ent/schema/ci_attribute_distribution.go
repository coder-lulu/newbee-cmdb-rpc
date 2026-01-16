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

// CiAttributeDistribution CI属性分布统计表 - 动态属性统计分析
type CiAttributeDistribution struct {
	ent.Schema
}

func (CiAttributeDistribution) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离
		mixins.DepartmentMixin{}, // 部门权限
	}
}

func (CiAttributeDistribution) Fields() []ent.Field {
	return []ent.Field{
		// CI类型ID
		field.Uint64("ci_type_id").
			Comment("外键，关联cmdb_ci_types.id"),

		// 属性ID
		field.Uint64("attribute_id").
			Comment("外键，关联cmdb_attributes.id"),

		// 属性名称（冗余字段，便于查询）
		field.String("attribute_name").
			MaxLen(128).
			Comment("属性名称"),

		// 属性别名
		field.String("attribute_alias").
			MaxLen(64).
			Optional().
			Comment("属性别名"),

		// 属性值类型
		field.Enum("value_type").
			Values("text", "integer", "float", "datetime", "json", "boolean").
			Comment("属性值类型"),

		// 值分布详情（支持各种数据类型的分布）
		field.JSON("value_distribution", map[string]interface{}{}).
			Comment("值分布详情，格式根据类型变化：\n" +
				"数值型: {\"min\": 1, \"max\": 100, \"avg\": 50, \"median\": 45, \"buckets\": [...]})\n" +
				"文本型: {\"top_values\": [{\"value\": \"linux\", \"count\": 80}], \"unique_count\": 10}\n" +
				"枚举型: {\"enum_distribution\": {\"active\": 70, \"inactive\": 30}}"),

		// 统计特征值
		field.JSON("statistical_metrics", map[string]interface{}{}).
			Optional().
			Comment("统计特征值，格式: {\"variance\": 12.5, \"std_dev\": 3.54, \"skewness\": 0.1}"),

		// 质量评分
		field.Float("quality_score").
			Default(0.0).
			Comment("数据质量评分 (0-100)，基于完整性、一致性、准确性"),

		// 完整性统计
		field.JSON("completeness_stats", map[string]interface{}{}).
			Optional().
			Comment("完整性统计，格式: {\"total_count\": 1000, \"null_count\": 50, \"completeness_rate\": 0.95}"),

		// 异常值检测
		field.JSON("outlier_analysis", map[string]interface{}{}).
			Optional().
			Comment("异常值分析，格式: {\"outlier_count\": 5, \"outlier_threshold\": 3.0, \"outliers\": [...]}"),

		// 总记录数
		field.Uint64("total_count").
			Default(0).
			Comment("总记录数"),

		// 唯一值数量
		field.Uint64("unique_count").
			Default(0).
			Comment("唯一值数量"),

		// 最后分析时间
		field.Time("last_analyzed_at").
			Default(time.Now).
			Comment("最后分析时间"),

		// 数据采样率
		field.Float("sample_rate").
			Default(1.0).
			Comment("数据采样率 (0.0-1.0)，1.0表示全量分析"),

		// 分析状态
		field.Enum("analysis_status").
			Values("pending", "analyzing", "completed", "failed").
			Default("pending").
			Comment("分析状态"),

		// 趋势变化（相比上一次分析）
		field.JSON("trend_analysis", map[string]interface{}{}).
			Optional().
			Comment("趋势分析，格式: {\"trend_direction\": \"up\", \"change_rate\": 0.05}"),
	}
}

func (CiAttributeDistribution) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + CI类型 + 属性的唯一索引
		index.Fields("tenant_id", "ci_type_id", "attribute_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),

		// 租户 + 属性名称查询索引
		index.Fields("tenant_id", "attribute_name"),

		// 租户 + 值类型查询索引
		index.Fields("tenant_id", "value_type", "last_analyzed_at"),

		// 质量评分查询索引
		index.Fields("tenant_id", "quality_score").
			Annotations(entsql.DescColumns("quality_score")),

		// 分析状态查询索引
		index.Fields("tenant_id", "analysis_status", "created_at"),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "ci_type_id"),

		// 最后分析时间排序索引
		index.Fields("tenant_id", "last_analyzed_at").
			Annotations(entsql.DescColumns("last_analyzed_at")),

		// 记录数量排序索引（用于找到高频属性）
		index.Fields("tenant_id", "total_count").
			Annotations(entsql.DescColumns("total_count")),
	}
}

func (CiAttributeDistribution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_ci_attribute_distribution",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='CI属性分布统计表 - 支持动态属性统计分析'",
		},
	}
}
