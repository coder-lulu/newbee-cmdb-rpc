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

// CiDimension CI统计维度表 - OLAP分析的核心维度表
type CiDimension struct {
	ent.Schema
}

func (CiDimension) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},     // 租户隔离
		mixins.DepartmentMixin{}, // 部门权限
	}
}

func (CiDimension) Fields() []ent.Field {
	return []ent.Field{
		// CI类型ID
		field.Uint64("ci_type_id").
			Comment("外键，关联cmdb_ci_types.id"),

		// CI类型名称（冗余存储，提升查询性能）
		field.String("ci_type_name").
			MaxLen(128).
			Comment("CI类型名称"),

		// CI类型别名
		field.String("ci_type_alias").
			MaxLen(64).
			Optional().
			Comment("CI类型别名"),

		// CI类型分类
		field.String("ci_category").
			MaxLen(64).
			Optional().
			Comment("CI类型分类：infrastructure, application, business等"),

		// 生命周期阶段
		field.Enum("lifecycle_stage").
			Values("development", "testing", "staging", "production", "retired").
			Default("production").
			Comment("生命周期阶段"),

		// 业务重要性等级
		field.Enum("business_criticality").
			Values("low", "medium", "high", "critical").
			Default("medium").
			Comment("业务重要性等级"),

		// 环境标识
		field.String("environment").
			MaxLen(32).
			Optional().
			Comment("环境标识：dev, test, uat, prod等"),

		// 地理位置/数据中心
		field.String("location").
			MaxLen(128).
			Optional().
			Comment("地理位置或数据中心"),

		// 组织架构维度（支持多层级）
		field.JSON("organization_hierarchy", map[string]interface{}{}).
			Optional().
			Comment("组织架构层级，格式: {JSON格式}"),

		// 标签维度（支持多标签分析）
		field.JSON("tag_dimensions", []string{}).
			Optional().
			Comment("标签维度列表，用于多维度标签分析"),

		// 成本中心
		field.String("cost_center").
			MaxLen(64).
			Optional().
			Comment("成本中心编码"),

		// 负责人信息（冗余存储）
		field.JSON("owner_info", map[string]interface{}{}).
			Optional().
			Comment("负责人信息，格式: {JSON格式}"),

		// 动态维度字段（支持扩展维度）
		field.JSON("custom_dimensions", map[string]interface{}{}).
			Optional().
			Comment("自定义维度字段，支持业务特定的维度分析"),

		// SLA等级
		field.Enum("sla_level").
			Values("bronze", "silver", "gold", "platinum").
			Optional().
			Comment("服务级别协议等级"),

		// 合规性标识
		field.JSON("compliance_tags", []string{}).
			Optional().
			Comment("合规性标识列表，如：GDPR, SOX, ISO27001等"),

		// 风险等级
		field.Enum("risk_level").
			Values("low", "medium", "high", "critical").
			Default("low").
			Comment("风险等级评估"),

		// 维度有效性标记
		field.Bool("is_active").
			Default(true).
			Comment("维度是否有效"),

		// 维度版本（支持维度变更历史）
		field.Uint64("dimension_version").
			Default(1).
			Comment("维度定义版本号"),

		// 最后更新维度时间
		field.Time("dimension_updated_at").
			Default(time.Now).
			Comment("维度信息最后更新时间"),
	}
}

func (CiDimension) Indexes() []ent.Index {
	return []ent.Index{
		// 租户 + CI类型的唯一索引
		index.Fields("tenant_id", "ci_type_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),

		// 租户 + CI类型名称查询索引
		index.Fields("tenant_id", "ci_type_name"),

		// 租户 + 分类查询索引
		index.Fields("tenant_id", "ci_category", "lifecycle_stage"),

		// 租户 + 业务重要性查询索引
		index.Fields("tenant_id", "business_criticality"),

		// 租户 + 环境查询索引
		index.Fields("tenant_id", "environment", "location"),

		// 租户 + SLA等级查询索引
		index.Fields("tenant_id", "sla_level", "risk_level"),

		// 部门权限查询索引
		index.Fields("tenant_id", "department_id", "ci_type_id"),

		// 成本中心查询索引
		index.Fields("tenant_id", "cost_center"),

		// 维度有效性查询索引
		index.Fields("tenant_id", "is_active", "dimension_updated_at").
			Annotations(entsql.DescColumns("dimension_updated_at")),

		// 版本查询索引（支持维度历史追踪）
		index.Fields("tenant_id", "ci_type_id", "dimension_version"),
	}
}

func (CiDimension) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cmdb_ci_dimension",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
			Options:   "ENGINE=InnoDB ROW_FORMAT=DYNAMIC COMMENT='CI统计维度表 - OLAP分析的核心维度表'",
		},
	}
}