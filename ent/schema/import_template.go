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
	"github.com/gofrs/uuid/v5"
)

// ImportTemplate 对应于数据库cmdb_import_templates
// 用于存储可重复使用的导入模板配置
type ImportTemplate struct {
	ent.Schema
}

func (ImportTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins2.SoftDeleteMixin{},
		mixins.TenantMixin{},
		mixins.DepartmentMixin{},
	}
}

// Fields of the ImportTemplate.
func (ImportTemplate) Fields() []ent.Field {
	return []ent.Field{
		// 基本信息
		field.String("name").MaxLen(255).NotEmpty().Comment("模板名称"),
		field.String("code").MaxLen(64).NotEmpty().Unique().Comment("模板编码"),
		field.String("description").MaxLen(500).Optional().Comment("模板描述"),
		field.String("version").MaxLen(32).Comment("模板版本").Default("1.0.0"),

		// 模板类型
		field.Enum("type").Values("excel", "csv", "json", "api", "xml").Comment("模板类型").Default("excel"),
		field.Enum("import_mode").Values("create_only", "update_only", "upsert", "merge").Comment("导入模式").Default("upsert"),

		// CI类型配置
		field.Uint64("ci_type_id").Optional().Comment("目标CI类型ID"),
		field.String("ci_type_name").MaxLen(100).Optional().Comment("目标CI类型名称"),
		field.Bool("auto_create_ci_type").Comment("是否自动创建CI类型").Default(false),

		// 字段映射配置
		field.JSON("field_mappings", []FieldMappingConfig{}).Optional().Comment("字段映射配置"),
		field.JSON("header_mappings", map[string]string{}).Optional().Comment("表头映射配置"),
		field.JSON("default_values", map[string]interface{}{}).Optional().Comment("默认值配置"),
		field.JSON("computed_fields", []ComputedFieldConfig{}).Optional().Comment("计算字段配置"),

		// 验证规则将直接使用CI类型属性中定义的验证规则

		// 数据处理配置
		field.JSON("data_transformations", []DataTransformationConfig{}).Optional().Comment("数据转换规则"),
		field.JSON("data_filters", []DataFilterConfig{}).Optional().Comment("数据过滤规则"),
		field.JSON("data_cleaners", []DataCleanerConfig{}).Optional().Comment("数据清洗规则"),

		// Excel专用配置
		field.String("excel_sheet_name").MaxLen(100).Optional().Comment("Excel工作表名称"),
		field.Int("excel_header_row").Optional().Comment("Excel表头行号").Default(1),
		field.Int("excel_data_start_row").Optional().Comment("Excel数据开始行号").Default(2),
		field.JSON("excel_column_mappings", map[string]string{}).Optional().Comment("Excel列映射"),

		// API专用配置
		field.String("api_endpoint").MaxLen(500).Optional().Comment("API端点"),
		field.Enum("api_method").Values("GET", "POST", "PUT", "PATCH").Optional().Comment("API方法"),
		field.JSON("api_headers", map[string]string{}).Optional().Comment("API请求头"),
		field.JSON("api_params", map[string]interface{}{}).Optional().Comment("API参数"),
		field.String("api_response_path").MaxLen(200).Optional().Comment("API响应数据路径"),

		// 错误处理配置
		field.Int("max_errors").Optional().Comment("最大允许错误数").Default(100),
		field.Bool("stop_on_first_error").Comment("遇到第一个错误时停止").Default(false),
		field.Bool("skip_invalid_rows").Comment("跳过无效行").Default(true),
		field.Enum("error_handling_mode").Values("strict", "lenient", "custom").Comment("错误处理模式").Default("lenient"),

		// 批处理配置
		field.Int("batch_size").Optional().Comment("批处理大小").Default(100),
		field.Int("max_parallel_jobs").Optional().Comment("最大并行任务数").Default(5),
		field.Bool("enable_transaction").Comment("是否启用事务").Default(true),

		// 模板元数据
		field.JSON("tags", []string{}).Optional().Comment("标签"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("扩展元数据"),
		field.String("icon").MaxLen(200).Optional().Comment("模板图标"),
		field.String("category").MaxLen(100).Optional().Comment("模板分类"),

		// 使用统计
		field.Int("usage_count").Comment("使用次数").Default(0),
		field.Int("success_count").Comment("成功次数").Default(0),
		field.Int("error_count").Comment("错误次数").Default(0),
		field.Float("success_rate").Comment("成功率").Default(0.0),

		// 权限和共享
		field.Bool("is_public").Comment("是否公开").Default(false),
		field.Bool("is_system").Comment("是否系统模板").Default(false),
		field.JSON("shared_with", []string{}).Optional().Comment("共享给用户列表"),

		// 审核信息
		field.UUID("created_by", uuid.UUID{}).Optional().Comment("创建者ID"),
		field.String("created_by_name").MaxLen(100).Optional().Comment("创建者姓名"),
		field.UUID("approved_by", uuid.UUID{}).Optional().Comment("审核者ID"),
		field.String("approved_by_name").MaxLen(100).Optional().Comment("审核者姓名"),
		field.Time("approved_at").Optional().Comment("审核时间"),
	}
}

// 字段映射配置
type FieldMappingConfig struct {
	SourceField     string                 `json:"source_field"`    // 源字段名
	TargetField     string                 `json:"target_field"`    // 目标字段名
	FieldType       string                 `json:"field_type"`      // 字段类型
	IsRequired      bool                   `json:"is_required"`     // 是否必填
	DefaultValue    interface{}            `json:"default_value"`   // 默认值
	Transformations []string               `json:"transformations"` // 转换规则
	Metadata        map[string]interface{} `json:"metadata"`        // 元数据
	// 注意：验证规则将直接使用CI类型属性中定义的规则
}

// 计算字段配置
type ComputedFieldConfig struct {
	FieldName    string                 `json:"field_name"`   // 字段名称
	Expression   string                 `json:"expression"`   // 计算表达式
	Dependencies []string               `json:"dependencies"` // 依赖字段
	FieldType    string                 `json:"field_type"`   // 字段类型
	Metadata     map[string]interface{} `json:"metadata"`     // 元数据
}

// 验证规则将直接使用CI类型属性中定义的验证规则，无需重复定义

// 数据转换配置
type DataTransformationConfig struct {
	Name        string                 `json:"name"`         // 转换名称
	Type        string                 `json:"type"`         // 转换类型
	SourceField string                 `json:"source_field"` // 源字段
	TargetField string                 `json:"target_field"` // 目标字段
	Expression  string                 `json:"expression"`   // 转换表达式
	Parameters  map[string]interface{} `json:"parameters"`   // 转换参数
}

// 数据过滤配置
type DataFilterConfig struct {
	Name       string                 `json:"name"`       // 过滤器名称
	Condition  string                 `json:"condition"`  // 过滤条件
	Parameters map[string]interface{} `json:"parameters"` // 过滤参数
	Action     string                 `json:"action"`     // 过滤动作(skip/include/transform)
}

// 数据清洗配置
type DataCleanerConfig struct {
	Name       string                 `json:"name"`       // 清洗器名称
	Type       string                 `json:"type"`       // 清洗类型
	Fields     []string               `json:"fields"`     // 作用字段
	Parameters map[string]interface{} `json:"parameters"` // 清洗参数
}

// Edges of the ImportTemplate.
func (ImportTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联CI类型
		edge.From("ci_type", CiType.Type).
			Ref("import_templates").
			Field("ci_type_id").
			Unique().
			Comment("关联CI类型"),

		// 关联导入任务
		edge.To("tasks", ImportTask.Type).
			Comment("关联导入任务"),
	}
}

// Indexes of the ImportTemplate
func (ImportTemplate) Indexes() []ent.Index {
	return []ent.Index{
		// 基本查询索引
		index.Fields("code").Unique(),
		index.Fields("name"),
		index.Fields("type"),
		index.Fields("ci_type_id"),
		index.Fields("created_by"),
		index.Fields("is_public"),
		index.Fields("is_system"),

		// 复合索引
		index.Fields("type", "ci_type_id"),
		index.Fields("created_by", "type"),
		index.Fields("is_public", "type"),
		index.Fields("category", "type"),
	}
}

// Annotations 返回表的注释
func (ImportTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cmdb_import_templates"},
	}
}
