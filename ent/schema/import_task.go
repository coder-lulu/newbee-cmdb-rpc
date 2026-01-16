package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/mixins"
)

// ImportTask 对应于数据库cmdb_import_tasks
// 用于记录资产导入任务的基本信息和状态
type ImportTask struct {
	ent.Schema
}

func (ImportTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
		mixins.CreatedByMixin{},
	}
}

// Fields of the ImportTask.
func (ImportTask) Fields() []ent.Field {
	return []ent.Field{
		// 基本信息
		field.String("task_id").MaxLen(64).NotEmpty().Unique().Comment("任务唯一标识"),
		field.String("name").MaxLen(255).NotEmpty().Comment("任务名称"),
		field.String("description").MaxLen(500).Optional().Comment("任务描述"),

		// 任务类型和状态
		field.Enum("type").Values("excel", "api", "csv", "json", "auto_discovery").Comment("导入类型").Default("excel"),
		field.Enum("status").Values("pending", "processing", "completed", "failed", "cancelled").Comment("任务状态").Default("pending"),
		field.Enum("priority").Values("low", "normal", "high").Comment("任务优先级").Default("normal"),

		// 数据来源
		field.String("source_path").MaxLen(500).Optional().Comment("数据源路径"),
		field.String("source_format").MaxLen(32).Optional().Comment("数据格式"),
		field.Int64("source_size").Optional().Comment("数据源大小（字节）"),

		// 模板配置
		field.Uint64("template_id").Optional().Comment("关联导入模板ID"),
		field.JSON("mapping_config", map[string]interface{}{}).Optional().Comment("字段映射配置"),

		// 处理配置
		field.Int("batch_size").Optional().Comment("批处理大小").Default(100),
		field.Int("max_errors").Optional().Comment("最大错误数").Default(100),
		field.Bool("dry_run").Comment("是否为预演模式").Default(false),

		// 处理统计
		field.Int("total_count").Comment("总记录数").Default(0),
		field.Int("processed_count").Comment("已处理记录数").Default(0),
		field.Int("success_count").Comment("成功记录数").Default(0),
		field.Int("failed_count").Comment("失败记录数").Default(0),
		field.Float("progress_percentage").Comment("进度百分比").Default(0.0),

		// 时间信息
		field.Time("start_time").Optional().Comment("开始时间"),
		field.Time("end_time").Optional().Comment("结束时间"),

		// 错误和结果
		field.Text("error_message").Optional().Comment("错误信息"),
		field.String("result_file_path").MaxLen(500).Optional().Comment("结果文件路径"),

		// 用户信息
		field.String("created_by_name").MaxLen(100).Optional().Comment("创建者姓名"),
	}
}

// Edges of the ImportTask.
func (ImportTask) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联导入模板
		edge.From("template", ImportTemplate.Type).
			Ref("tasks").
			Field("template_id").
			Unique().
			Comment("关联导入模板"),

		// 关联导入记录
		edge.To("records", ImportRecord.Type).
			Comment("关联导入记录"),

		// 关联导入错误
		edge.To("errors", ImportError.Type).
			Comment("关联导入错误"),
	}
}

// Indexes of the ImportTask
func (ImportTask) Indexes() []ent.Index {
	return []ent.Index{
		// 基本查询索引
		index.Fields("task_id").Unique(),
		index.Fields("status"),
		index.Fields("type"),
		index.Fields("created_by"),

		// 复合索引
		index.Fields("status", "priority", "created_at"),
		index.Fields("type", "status"),
		index.Fields("created_by", "status"),
	}
}

// Annotations 返回表的注释
func (ImportTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_import_tasks"},
	}
}
