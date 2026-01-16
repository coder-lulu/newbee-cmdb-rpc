package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
	"github.com/gofrs/uuid/v5"
)

// CiRecords 对应于数据库cmdb_ci_records
// 用于记录CI实例的所有变更历史（增删改查）
type CiRecords struct {
	ent.Schema
}

func (CiRecords) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the CiRecords.
func (CiRecords) Fields() []ent.Field {
	return []ent.Field{
		// CI基本信息
		field.Uint64("ci_id").Comment("CI实例ID"),
		field.Uint64("ci_type_id").Comment("CI类型ID"),
		field.String("ci_type_name").MaxLen(100).Comment("CI类型名称"),
		field.String("ci_unique_key").MaxLen(255).Optional().Comment("CI唯一标识"),

		// 操作信息
		field.Enum("operation_type").Values("create", "update", "delete", "restore").Comment("操作类型"),
		field.Time("operation_time").Comment("操作时间"),
		field.UUID("operation_user_id", uuid.UUID{}).Optional().Comment("操作用户ID"),
		field.String("operation_user_name").MaxLen(100).Optional().Comment("操作用户名"),

		// 变更来源
		field.Enum("source_type").Values("manual", "import", "api", "system", "sync").Comment("变更来源类型"),
		field.String("source_id").MaxLen(100).Optional().Comment("来源ID（如导入任务ID、API调用ID等）"),
		field.String("source_description").MaxLen(500).Optional().Comment("来源描述"),

		// 变更内容
		field.JSON("before_data", map[string]interface{}{}).Optional().Comment("变更前数据"),
		field.JSON("after_data", map[string]interface{}{}).Optional().Comment("变更后数据"),
		field.JSON("changed_fields", []string{}).Optional().Comment("变更字段列表"),
		field.JSON("change_summary", map[string]interface{}{}).Optional().Comment("变更摘要"),

		// 变更说明
		field.String("change_reason").MaxLen(500).Optional().Comment("变更原因"),
		field.Text("description").Optional().Comment("变更描述"),
		field.Text("comments").Optional().Comment("备注"),

		// 版本信息
		field.Int("version_number").Comment("版本号").Default(1),
		field.String("revision_id").MaxLen(64).Optional().Comment("修订ID"),

		// 审计信息
		field.String("client_ip").MaxLen(45).Optional().Comment("客户端IP"),
		field.String("user_agent").MaxLen(500).Optional().Comment("用户代理"),
		field.JSON("request_context", map[string]interface{}{}).Optional().Comment("请求上下文"),

		// 影响分析
		field.Int("affected_relations").Comment("影响的关系数量").Default(0),
		field.JSON("affected_relation_ids", []uint64{}).Optional().Comment("影响的关系ID列表"),
		field.Bool("cascade_changes").Comment("是否有级联变更").Default(false),

		// 状态标记
		field.Bool("is_major_change").Comment("是否为重大变更").Default(false),
		field.Bool("requires_approval").Comment("是否需要审批").Default(false),
		field.Enum("approval_status").Values("pending", "approved", "rejected", "auto_approved").Optional().Comment("审批状态"),
		field.UUID("approved_by", uuid.UUID{}).Optional().Comment("审批人ID"),
		field.Time("approved_at").Optional().Comment("审批时间"),
	}
}

// Edges of the CiRecords.
func (CiRecords) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联CI实例
		edge.From("ci", Cis.Type).
			Ref("records").
			Field("ci_id").
			Unique().
			Comment("关联CI实例").
			Required(),

		// 关联CI类型
		edge.From("ci_type", CiType.Type).
			Ref("ci_records").
			Field("ci_type_id").
			Unique().
			Comment("关联CI类型").
			Required(),
	}
}

// Indexes of the CiRecords
func (CiRecords) Indexes() []ent.Index {
	return []ent.Index{
		// 基本查询索引
		index.Fields("ci_id"),
		index.Fields("ci_type_id"),
		index.Fields("operation_type"),
		index.Fields("operation_time"),
		index.Fields("operation_user_id"),
		index.Fields("source_type"),

		// 复合索引
		index.Fields("ci_id", "operation_time"),
		index.Fields("ci_id", "operation_type"),
		index.Fields("ci_type_id", "operation_type"),
		index.Fields("operation_user_id", "operation_time"),
		index.Fields("source_type", "source_id"),

		// 业务索引
		index.Fields("ci_unique_key"),
		index.Fields("is_major_change", "operation_time"),
		index.Fields("requires_approval", "approval_status"),
		index.Fields("version_number", "ci_id"),
	}
}

// Annotations 返回表的注释
func (CiRecords) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_ci_records"},
	}
}
