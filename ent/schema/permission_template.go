package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

// PermissionTemplate 权限模板表 - 简化配置复杂度
type PermissionTemplate struct {
	ent.Schema
}

// Mixin of the PermissionTemplate.
func (PermissionTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
	}
}

// Fields of the PermissionTemplate.
func (PermissionTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("template_id").
			Unique().
			Comment("模板ID"),

		field.String("template_name").
			Comment("模板名称"),

		field.Text("template_description").
			Optional().
			Comment("模板描述"),

		field.String("category").
			Optional().
			Comment("模板分类"),

		// 模板配置
		field.Enum("scope_type").
			Values("global", "ci_type", "ci_instance", "attribute", "field").
			Optional().
			Comment("权限范围类型"),

		field.Enum("permission_level").
			Values("none", "read", "write", "admin", "super_admin").
			Optional().
			Comment("权限级别"),

		field.Uint64("operations_mask").
			Default(0).
			Comment("操作位掩码"),

		field.Enum("risk_level").
			Values("low", "medium", "high", "critical").
			Default("low").
			Comment("风险等级"),

		// 模板状态
		field.Bool("is_system_template").
			Default(false).
			Comment("是否系统模板"),

		field.Bool("is_active").
			Default(true).
			Comment("是否启用"),

		field.Int("sort_order").
			Default(0).
			Comment("排序序号"),

		field.String("created_by").
			Optional().
			Comment("创建人"),
	}
}

// Indexes of the PermissionTemplate.
func (PermissionTemplate) Indexes() []ent.Index {
	return []ent.Index{
		// 模板查询索引
		index.Fields("category", "is_active"),
		index.Fields("scope_type", "permission_level"),
		index.Fields("is_system_template", "is_active"),
		index.Fields("sort_order"),
	}
}

// Annotations 返回表的注释
func (PermissionTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_permission_templates"},
	}
}
