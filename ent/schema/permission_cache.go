package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

// PermissionCache 权限预计算结果缓存表 - 提升查询性能的核心表
type PermissionCache struct {
	ent.Schema
}

// Mixin of the PermissionCache.
func (PermissionCache) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
	}
}

// Fields of the PermissionCache.
func (PermissionCache) Fields() []ent.Field {
	return []ent.Field{
		field.String("cache_key").
			Unique().
			Comment("缓存键：user_id:resource_type:resource_id"),

		field.String("user_id").
			Comment("用户ID"),

		field.String("resource_type").
			Comment("资源类型：ci_type/ci_instance/attribute/field"),

		field.String("resource_id").
			Comment("资源ID"),

		// 预计算结果
		field.Uint64("allowed_operations").
			Default(0).
			Comment("允许的操作位掩码"),

		field.Enum("permission_level").
			Values("none", "read", "write", "admin", "super_admin").
			Default("none").
			Comment("权限级别"),

		field.Bool("has_data_filters").
			Default(false).
			Comment("是否有数据过滤"),

		field.Bool("has_field_masks").
			Default(false).
			Comment("是否有字段掩码"),

		// 缓存控制
		field.String("cache_version").
			Comment("缓存版本"),

		field.Time("expires_at").
			Comment("过期时间"),

		field.Time("last_accessed_at").
			Default(time.Now).
			Comment("最后访问时间"),

		field.Int("access_count").
			Default(1).
			Comment("访问次数"),
	}
}

// Indexes of the PermissionCache.
func (PermissionCache) Indexes() []ent.Index {
	return []ent.Index{
		// 核心查询索引
		index.Fields("user_id", "resource_type", "resource_id"),

		// 缓存管理索引
		index.Fields("expires_at"),
		index.Fields("cache_version"),
		index.Fields("last_accessed_at"),

		// 复合查询索引
		index.Fields("user_id", "expires_at"),
		index.Fields("resource_type", "resource_id"),
	}
}

// Annotations 返回表的注释
func (PermissionCache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_permission_cache"},
	}
}
