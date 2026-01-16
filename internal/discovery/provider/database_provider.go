package provider

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseProvider 数据库提供者实现
type DatabaseProvider struct {
	name string
	providerType string
}

// NewDatabaseProvider 创建数据库提供者
func NewDatabaseProvider() *DatabaseProvider {
	return &DatabaseProvider{
		name:         "mysql",
		providerType: "database",
	}
}

// GetName 获取提供者名称
func (p *DatabaseProvider) GetName() string {
	return p.name
}

// GetType 获取提供者类型
func (p *DatabaseProvider) GetType() string {
	return p.providerType
}

// Connect 创建数据库连接
func (p *DatabaseProvider) Connect(ctx context.Context, config map[string]interface{}) (DataSource, error) {
	// 验证配置
	if err := p.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config["username"].(string),
		config["password"].(string),
		config["host"].(string),
		config["port"],
		config["database"].(string),
	)

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseDataSource{
		db:     db,
		config: config,
	}, nil
}

// ValidateConfig 验证配置
func (p *DatabaseProvider) ValidateConfig(config map[string]interface{}) error {
	required := []string{"host", "port", "username", "password", "database"}
	
	for _, field := range required {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// GetConfigSchema 获取配置结构定义
func (p *DatabaseProvider) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"host": map[string]interface{}{
				"type":        "string",
				"description": "数据库主机地址",
				"example":     "localhost",
			},
			"port": map[string]interface{}{
				"type":        "integer",
				"description": "数据库端口",
				"default":     3306,
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "数据库用户名",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "数据库密码",
				"format":      "password",
			},
			"database": map[string]interface{}{
				"type":        "string",
				"description": "数据库名称",
			},
		},
		"required": []string{"host", "port", "username", "password", "database"},
	}
}

// DatabaseDataSource 数据库数据源实现
type DatabaseDataSource struct {
	db     *sql.DB
	config map[string]interface{}
}

// Connect 连接到数据源（对于数据库已经在Provider中连接）
func (ds *DatabaseDataSource) Connect(ctx context.Context, config map[string]interface{}) error {
	// 数据库已经在Provider中连接，这里可以做额外的初始化
	return nil
}

// Discover 执行数据发现
func (ds *DatabaseDataSource) Discover(ctx context.Context, rules map[string]interface{}) ([]map[string]interface{}, error) {
	// 从规则中获取SQL查询
	query, exists := rules["sql"].(string)
	if !exists {
		return nil, fmt.Errorf("sql query not specified in discovery rules")
	}

	// 执行查询
	rows, err := ds.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	// 处理结果
	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 构建结果映射
		record := make(map[string]interface{})
		for i, col := range columns {
			record[col] = values[i]
		}

		results = append(results, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Close 关闭连接
func (ds *DatabaseDataSource) Close() error {
	return ds.db.Close()
}

// Validate 验证连接配置
func (ds *DatabaseDataSource) Validate(config map[string]interface{}) error {
	// 可以在这里执行额外的验证
	return nil
}

// GetSchema 获取数据源结构信息
func (ds *DatabaseDataSource) GetSchema(ctx context.Context) (map[string]interface{}, error) {
	// 查询所有表
	query := "SHOW TABLES"
	rows, err := ds.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	// 为每个表获取列信息
	schema := map[string]interface{}{
		"type":   "database",
		"tables": make(map[string]interface{}),
	}

	tablesMap := schema["tables"].(map[string]interface{})
	
	for _, table := range tables {
		columns, err := ds.getTableColumns(ctx, table)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table, err)
		}
		tablesMap[table] = columns
	}

	return schema, nil
}

// getTableColumns 获取表的列信息
func (ds *DatabaseDataSource) getTableColumns(ctx context.Context, tableName string) (map[string]interface{}, error) {
	query := "DESCRIBE " + tableName
	rows, err := ds.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]interface{})
	
	for rows.Next() {
		var field, dataType, null, key, defaultValue, extra sql.NullString
		
		if err := rows.Scan(&field, &dataType, &null, &key, &defaultValue, &extra); err != nil {
			return nil, err
		}

		columnInfo := map[string]interface{}{
			"type":     dataType.String,
			"nullable": null.String == "YES",
		}

		if key.Valid {
			columnInfo["key"] = key.String
		}
		if defaultValue.Valid {
			columnInfo["default"] = defaultValue.String
		}
		if extra.Valid {
			columnInfo["extra"] = extra.String
		}

		columns[field.String] = columnInfo
	}

	return columns, nil
}