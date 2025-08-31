package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	mixins2 "github.com/coder-lulu/newbee-cmdb-rpc/ent/schema/mixins"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

// ImportRecord 对应于数据库cmdb_import_records
// 用于详细记录每条导入数据的处理结果和状态
type ImportRecord struct {
	ent.Schema
}

func (ImportRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the ImportRecord.
func (ImportRecord) Fields() []ent.Field {
	return []ent.Field{
		// 关联信息
		field.Uint64("task_id").Comment("关联导入任务ID"),
		field.String("batch_id").MaxLen(64).Optional().Comment("批次ID"),

		// 数据位置信息
		field.Int("row_number").Comment("行号（Excel/CSV中的行号）"),
		field.String("sheet_name").MaxLen(100).Optional().Comment("工作表名称（Excel专用）"),

		// 处理状态
		field.Enum("status").Values("pending", "processing", "success", "failed", "skipped").Comment("处理状态").Default("pending"),
		field.Enum("import_action").Values("create", "update", "skip").Optional().Comment("导入动作"),

		// 数据内容
		field.JSON("raw_data", map[string]interface{}{}).Optional().Comment("原始输入数据"),
		field.JSON("final_data", map[string]interface{}{}).Optional().Comment("最终处理后数据"),

		// CI相关信息
		field.Uint64("ci_id").Optional().Comment("生成的CI实例ID"),
		field.Uint64("ci_type_id").Optional().Comment("CI类型ID"),
		field.String("ci_unique_key").MaxLen(255).Optional().Comment("CI唯一标识"),

		// 错误信息
		field.Text("error_message").Optional().Comment("错误消息"),
		field.String("error_code").MaxLen(50).Optional().Comment("错误代码"),
		field.Enum("error_type").Values("validation", "transformation", "persistence", "business", "system").Optional().Comment("错误类型"),

		// 处理时间
		field.Time("start_time").Optional().Comment("开始处理时间"),
		field.Time("end_time").Optional().Comment("结束处理时间"),

		// 重试信息
		field.Int("retry_count").Comment("重试次数").Default(0),
		field.Int("max_retries").Comment("最大重试次数").Default(3),
	}
}

// Edges of the ImportRecord.
func (ImportRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联导入任务
		edge.From("task", ImportTask.Type).
			Ref("records").
			Field("task_id").
			Unique().
			Comment("关联导入任务").
			Required(),

		// 关联CI实例
		edge.From("ci", Cis.Type).
			Ref("import_records").
			Field("ci_id").
			Unique().
			Comment("关联生成的CI实例"),

		// 关联CI类型
		edge.From("ci_type", CiType.Type).
			Ref("import_records").
			Field("ci_type_id").
			Unique().
			Comment("关联CI类型"),

		// 关联导入错误
		edge.To("errors", ImportError.Type).
			Comment("关联导入错误"),
	}
}

// Indexes of the ImportRecord
func (ImportRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 基本查询索引
		index.Fields("task_id"),
		index.Fields("status"),
		index.Fields("ci_id"),
		index.Fields("row_number"),
		index.Fields("batch_id"),

		// 复合索引
		index.Fields("task_id", "status"),
		index.Fields("task_id", "row_number"),
		index.Fields("ci_type_id", "status"),
		index.Fields("ci_unique_key"),
	}
}

// Annotations 返回表的注释
func (ImportRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_import_records"},
	}
}
