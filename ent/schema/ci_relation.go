package schema

import (
	"time"
	
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// AttributeMappingData 属性映射数据结构
type AttributeMappingData struct {
	SourceAttrID   uint64      `json:"sourceAttrId"`   // 源属性ID
	TargetAttrID   uint64      `json:"targetAttrId"`   // 目标属性ID
	SourceValue    interface{} `json:"sourceValue"`    // 源属性值
	TargetValue    interface{} `json:"targetValue"`    // 目标属性值
	LastSyncAt     *time.Time  `json:"lastSyncAt"`     // 最后同步时间
	SyncStatus     string      `json:"syncStatus"`     // 同步状态: synced, pending, failed
	ConflictReason string      `json:"conflictReason"` // 冲突原因
}

// ValidationResult 关系验证结果
type ValidationResult struct {
	IsValid        bool     `json:"isValid"`        // 是否有效
	Errors         []string `json:"errors"`         // 错误信息
	Warnings       []string `json:"warnings"`       // 警告信息
	ValidatedAt    time.Time `json:"validatedAt"`    // 验证时间
	ValidationType string   `json:"validationType"` // 验证类型: auto, manual
}

// SyncConfig 同步配置
type SyncConfig struct {
	SyncDirection   string        `json:"syncDirection"`   // 同步方向: bidirectional, source_to_target, target_to_source
	SyncFrequency   string        `json:"syncFrequency"`   // 同步频率: realtime, hourly, daily
	ConflictPolicy  string        `json:"conflictPolicy"`  // 冲突策略: manual, source_priority, target_priority, latest
	RetryAttempts   int           `json:"retryAttempts"`   // 重试次数
	RetryInterval   time.Duration `json:"retryInterval"`   // 重试间隔
	EnabledMappings []uint64      `json:"enabledMappings"` // 启用的映射ID列表
}

// CiRelation 对应于数据库cmdb_ci_relations
// 用于记录CI实例之间的关联关系
type CiRelation struct {
	ent.Schema
}

func (CiRelation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiRelation.
func (CiRelation) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("source_ci_id").Comment("外键，关联cmdb_cis.id，源CI"),
		field.Uint64("target_ci_id").Comment("外键，关联cmdb_cis.id，目标CI"),
		field.Uint64("relation_type_id").Comment("外键，关联cmdb_relation_types.id"),
		field.Uint64("more").Optional().Comment("更多CI，外键，关联cmdb_cis.id"),
		field.String("discovery_source").Optional().Comment("发现来源：manual, auto_discovery, import"),
		field.String("ancestor_ids").Optional().MaxLen(128).Comment("祖先关系ID链，用于关系路径追踪"),
		field.JSON("properties", map[string]interface{}{}).Optional().Comment("关系属性，扩展关系的自定义信息"),
		
		// 扩展字段：支持属性映射和关系状态管理
		field.JSON("attribute_mappings", []AttributeMappingData{}).Optional().Comment("属性映射数据，记录实际的映射值"),
		field.String("status").Optional().Default("active").Comment("关系状态：active, inactive, suspended"),
		field.JSON("validation_result", ValidationResult{}).Optional().Comment("关系验证结果"),
		field.Time("last_validated_at").Optional().Comment("最后验证时间"),
		field.Bool("auto_sync_enabled").Optional().Default(true).Comment("是否启用属性自动同步"),
		field.JSON("sync_config", SyncConfig{}).Optional().Comment("同步配置"),
		field.String("relation_strength").Optional().Default("normal").Comment("关系强度：weak, normal, strong"),
	}
}

// Edges of the CiRelation.
func (CiRelation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("source_ci", Cis.Type).Ref("source_relations").Field("source_ci_id").Unique().Comment("源CI").Required(),
		edge.From("target_ci", Cis.Type).Ref("target_relations").Field("target_ci_id").Unique().Comment("目标CI").Required(),
		edge.From("relation_type", RelationType.Type).Ref("ci_relations").Field("relation_type_id").Unique().Comment("关系类型").Required(),
		edge.From("more_ci", Cis.Type).Ref("more_relations").Field("more").Unique().Comment("更多CI"),
	}
}

// Annotations 返回表的注释
func (CiRelation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_relations"},
	}
}
