package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	mixins2 "gitee.com/link234/cmdb-rpc/ent/schema/mixins"
	"gitee.com/link234/newbee-backend-common/orm/ent/mixins"
)

// ImportError 对应于数据库cmdb_import_errors
// 用于记录导入过程中的错误信息
type ImportError struct {
	ent.Schema
}

func (ImportError) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the ImportError.
func (ImportError) Fields() []ent.Field {
	return []ent.Field{
		// 关联信息
		field.Uint64("task_id").Comment("关联导入任务ID"),
		field.Uint64("record_id").Optional().Comment("关联导入记录ID"),
		field.String("batch_id").MaxLen(64).Optional().Comment("批次ID"),

		// 错误基本信息
		field.String("error_code").MaxLen(50).Comment("错误代码"),
		field.String("error_title").MaxLen(255).Comment("错误标题"),
		field.Text("error_message").Comment("错误消息"),
		field.Text("error_details").Optional().Comment("详细错误信息"),

		// 错误分类
		field.Enum("error_type").Values(
			"validation",     // 验证错误
			"transformation", // 转换错误
			"persistence",    // 持久化错误
			"business",       // 业务规则错误
			"system",         // 系统错误
			"format",         // 格式错误
		).Comment("错误类型"),

		field.Enum("severity").Values("low", "medium", "high", "critical").Comment("严重程度").Default("medium"),

		// 数据位置信息
		field.Int("row_number").Optional().Comment("错误发生的行号"),
		field.String("field_name").MaxLen(100).Optional().Comment("错误发生的字段名"),
		field.String("sheet_name").MaxLen(100).Optional().Comment("工作表名称（Excel专用）"),

		// 错误数据
		field.JSON("input_data", map[string]interface{}{}).Optional().Comment("导致错误的输入数据"),
		field.JSON("error_context", map[string]interface{}{}).Optional().Comment("错误上下文"),

		// 修复建议
		field.Text("suggestion").Optional().Comment("修复建议"),

		// 处理状态
		field.Enum("status").Values("new", "acknowledged", "resolved", "ignored").Comment("处理状态").Default("new"),
		field.String("resolved_by").MaxLen(100).Optional().Comment("解决人"),
		field.Time("resolved_at").Optional().Comment("解决时间"),
	}
}

// Edges of the ImportError.
func (ImportError) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联导入任务
		edge.From("task", ImportTask.Type).
			Ref("errors").
			Field("task_id").
			Unique().
			Comment("关联导入任务").
			Required(),

		// 关联导入记录
		edge.From("record", ImportRecord.Type).
			Ref("errors").
			Field("record_id").
			Unique().
			Comment("关联导入记录"),
	}
}

// Indexes of the ImportError
func (ImportError) Indexes() []ent.Index {
	return []ent.Index{
		// 基本查询索引
		index.Fields("task_id"),
		index.Fields("record_id"),
		index.Fields("error_type"),
		index.Fields("severity"),
		index.Fields("status"),

		// 复合索引
		index.Fields("task_id", "error_type"),
		index.Fields("task_id", "severity"),
		index.Fields("error_type", "severity"),
		index.Fields("row_number", "task_id"),
	}
}

// Annotations 返回表的注释
func (ImportError) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_import_errors"},
	}
}
