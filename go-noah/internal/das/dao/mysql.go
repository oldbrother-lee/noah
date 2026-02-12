package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// MySQLDB MySQL数据库连接
type MySQLDB struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Params   map[string]string
	Ctx      context.Context
}

// Open 打开数据库连接
func (d *MySQLDB) Open() (*sql.DB, error) {
	config := mysql.Config{
		User:                 d.User,
		Passwd:               d.Password,
		Addr:                 fmt.Sprintf("%s:%d", d.Host, d.Port),
		Net:                  "tcp",
		DBName:               d.Database,
		AllowNativePasswords: true,
		Params:               d.Params,
		Timeout:              3 * time.Second,
	}

	DSN := config.FormatDSN()
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// Query 执行查询并返回结果
func (d *MySQLDB) Query(query string) ([]string, []map[string]interface{}, error) {
	db, err := d.Open()
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(d.Ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	vals := make([]interface{}, len(columns))
	for i := range columns {
		vals[i] = new(sql.RawBytes)
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err := rows.Scan(vals...); err != nil {
			return nil, nil, err
		}

		vmap := make(map[string]interface{}, len(vals))
		for i, c := range vals {
			// 处理列名重复的问题
			if _, ok := vmap[columns[i]]; ok {
				columns[i] = fmt.Sprintf("%s_%s[别名]", columns[i], uuid.New())
			}

			switch v := c.(type) {
			case *sql.RawBytes:
				if *v == nil {
					vmap[columns[i]] = nil
				} else {
					vmap[columns[i]] = string(*v)
				}
			}
		}
		result = append(result, vmap)
	}

	if err = rows.Close(); err != nil {
		return nil, nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, result, nil
}

// Execute 执行非查询语句
func (d *MySQLDB) Execute(query string) error {
	db, err := d.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(d.Ctx, query)
	return err
}

// TestConnection 测试数据库连接
func (d *MySQLDB) TestConnection() error {
	db, err := d.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(d.Ctx)
}

// GetSchemas 获取数据库列表
func (d *MySQLDB) GetSchemas() ([]string, error) {
	db, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(d.Ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		// 过滤系统库
		if schema != "information_schema" && schema != "performance_schema" && schema != "mysql" && schema != "sys" {
			schemas = append(schemas, schema)
		}
	}

	return schemas, nil
}

// TableMetaData 表元数据结构
type TableMetaData struct {
	TableSchema string `json:"table_schema"`
	TableName   string `json:"table_name"`
	Columns     string `json:"columns"`
}

// GetTables 获取指定库的表和字段元数据
func (d *MySQLDB) GetTables(schema string) ([]TableMetaData, error) {
	// 验证参数，防止 SQL 注入和 undefined 值
	if schema == "" || schema == "undefined" {
		return nil, fmt.Errorf("schema 参数不能为空")
	}

	db, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 设置 group_concat_max_len 以支持大量列
	_, err = db.ExecContext(d.Ctx, "SET SESSION group_concat_max_len = 4194304")
	if err != nil {
		return nil, err
	}

	// 查询表和字段信息，排除 gh-ost 相关临时表
	query := fmt.Sprintf(`
		SELECT 
			table_schema as table_schema,
			table_name as table_name, 
			group_concat(concat(column_name, '$$', column_type, '$$', IFNULL(column_comment, '')) SEPARATOR '@@') as columns
		FROM 
			information_schema.columns
		WHERE 
			table_schema='%s' AND table_name NOT REGEXP '^_(.*)[_ghc|_gho|_del]$'
		GROUP BY 
			table_schema, table_name 
		ORDER BY table_name
	`, schema)

	rows, err := db.QueryContext(d.Ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableMetaData
	for rows.Next() {
		var t TableMetaData
		if err := rows.Scan(&t.TableSchema, &t.TableName, &t.Columns); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}

	return tables, nil
}

// GetTableColumns 获取表结构
func (d *MySQLDB) GetTableColumns(schema, table string) ([]map[string]interface{}, error) {
	// 验证参数，防止 SQL 注入和 undefined 值
	if schema == "" || table == "" || schema == "undefined" || table == "undefined" {
		return nil, fmt.Errorf("schema 和 table 参数不能为空")
	}

	db, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf("DESC `%s`.`%s`", schema, table)
	rows, err := db.QueryContext(d.Ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		vals := make([]interface{}, len(columns))
		for i := range columns {
			vals[i] = new(sql.RawBytes)
		}

		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}

		vmap := make(map[string]interface{}, len(vals))
		for i, c := range vals {
			switch v := c.(type) {
			case *sql.RawBytes:
				if *v == nil {
					vmap[columns[i]] = nil
				} else {
					vmap[columns[i]] = string(*v)
				}
			}
		}
		result = append(result, vmap)
	}

	return result, nil
}
