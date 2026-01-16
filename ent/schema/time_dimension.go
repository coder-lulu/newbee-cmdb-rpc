package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// TimeDimension 时间维度表 - 支持多粒度时间分析的维度表
type TimeDimension struct {
	ent.Schema
}

func (TimeDimension) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离（时间维度也需要租户隔离以支持不同时区）
		mixins.DepartmentMixin{}, // 维度数据需要部门隔离
	}
}

func (TimeDimension) Fields() []ent.Field {
	return []ent.Field{
		// 完整日期时间（主键性质）
		field.Time("date_time").
			Comment("完整的日期时间"),

		// 日期（YYYY-MM-DD格式）
		field.String("date").
			MaxLen(10).
			Comment("日期字符串，格式：YYYY-MM-DD"),

		// 年份
		field.Int("year").
			Comment("年份，如：2024"),

		// 季度
		field.Int("quarter").
			Range(1, 4).
			Comment("季度，1-4"),

		// 月份
		field.Int("month").
			Range(1, 12).
			Comment("月份，1-12"),

		// 周数（年内第几周）
		field.Int("week_of_year").
			Range(1, 53).
			Comment("年内第几周，1-53"),

		// 月内第几周
		field.Int("week_of_month").
			Range(1, 6).
			Comment("月内第几周，1-6"),

		// 日期（月内第几天）
		field.Int("day_of_month").
			Range(1, 31).
			Comment("月内第几天，1-31"),

		// 年内第几天
		field.Int("day_of_year").
			Range(1, 366).
			Comment("年内第几天，1-366"),

		// 星期几（1=Monday, 7=Sunday）
		field.Int("day_of_week").
			Range(1, 7).
			Comment("星期几，1=周一，7=周日"),

		// 小时（0-23）
		field.Int("hour").
			Range(0, 23).
			Comment("小时，0-23"),

		// 分钟（0-59）
		field.Int("minute").
			Range(0, 59).
			Comment("分钟，0-59"),

		// 时间段标识
		field.Enum("time_period").
			Values("morning", "afternoon", "evening", "night").
			Comment("时间段：morning(6-12), afternoon(12-18), evening(18-22), night(22-6)"),

		// 工作日标识
		field.Bool("is_weekday").
			Comment("是否工作日"),

		// 周末标识
		field.Bool("is_weekend").
			Comment("是否周末"),

		// 节假日标识
		field.Bool("is_holiday").
			Default(false).
			Comment("是否节假日"),

		// 节假日名称
		field.String("holiday_name").
			MaxLen(64).
			Optional().
			Comment("节假日名称，如：春节、国庆节等"),

		// 季度名称
		field.String("quarter_name").
			MaxLen(16).
			Comment("季度名称，如：Q1, Q2, Q3, Q4"),

		// 月份名称
		field.String("month_name").
			MaxLen(16).
			Comment("月份名称，如：January, 一月等"),

		// 星期名称
		field.String("weekday_name").
			MaxLen(16).
			Comment("星期名称，如：Monday, 周一等"),

		// 时区标识
		field.String("timezone").
			MaxLen(32).
			Default("Asia/Shanghai").
			Comment("时区标识，如：Asia/Shanghai, UTC等"),

		// UTC偏移量（小时）
		field.Int("utc_offset").
			Default(8).
			Comment("相对UTC的偏移小时数"),

		// 财务年度
		field.Int("fiscal_year").
			Optional().
			Comment("财务年度（可能与自然年不同）"),

		// 财务季度
		field.Int("fiscal_quarter").
			Optional().
			Range(1, 4).
			Comment("财务季度，1-4"),

		// 财务月份
		field.Int("fiscal_month").
			Optional().
			Range(1, 12).
			Comment("财务月份，1-12"),

		// 业务周期标识（支持自定义业务周期）
		field.String("business_cycle").
			MaxLen(32).
			Optional().
			Comment("业务周期标识，支持自定义业务场景"),

		// 相对时间标识
		field.JSON("relative_time_flags", map[string]bool{}).
			Optional().
			Comment("相对时间标识，格式: {JSON格式}"),

		// 时间维度版本（支持时间维度定义变更）
		field.Uint64("dimension_version").
			Default(1).
			Comment("时间维度版本号"),
	}
}

func (TimeDimension) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + 日期时间唯一索引
		index.Fields("tenant_id", "date_time").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),

		// 租户 + 日期查询索引
		index.Fields("tenant_id", "date"),

		// 租户 + 年月查询索引
		index.Fields("tenant_id", "year", "month"),

		// 租户 + 年季度查询索引
		index.Fields("tenant_id", "year", "quarter"),

		// 租户 + 周查询索引
		index.Fields("tenant_id", "year", "week_of_year"),

		// 租户 + 工作日查询索引
		index.Fields("tenant_id", "is_weekday", "date"),

		// 租户 + 节假日查询索引
		index.Fields("tenant_id", "is_holiday", "date"),

		// 租户 + 时间段查询索引
		index.Fields("tenant_id", "time_period", "date"),

		// 租户 + 小时查询索引（用于小时级统计）
		index.Fields("tenant_id", "year", "month", "day_of_month", "hour"),

		// 财务年度查询索引
		index.Fields("tenant_id", "fiscal_year", "fiscal_quarter"),

		// 业务周期查询索引
		index.Fields("tenant_id", "business_cycle", "date"),

		// 时区查询索引
		index.Fields("tenant_id", "timezone", "date"),

		// 年份范围查询优化索引
		index.Fields("tenant_id", "year").
			Annotations(entsql.DescColumns("year")),

		// 日期时间范围查询优化索引（单独的降序索引）
		index.Fields("date_time").
			Annotations(entsql.DescColumns("date_time")),
	}
}

func (TimeDimension) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_time_dimension",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='时间维度表 - 支持多粒度时间分析'",
		},
		// 建议按年份分区（MySQL 8.0+）
		entsql.WithComments(true),
	}
}
