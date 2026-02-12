package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-noah/internal/inspect/parser"
	"go-noah/pkg/kv"
	"go-noah/pkg/utils"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// DB 数据库连接结构
type DB struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

// Open 打开数据库连接
func (d *DB) Open() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=3s&readTimeout=3s&writeTimeout=3s",
		d.User, d.Password, d.Host, d.Port, d.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(5 * time.Second)
	return db, nil
}

// Execute 执行SQL语句（不返回结果）
func (d *DB) Execute(query string) error {
	db, err := d.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(query)
	return err
}

// Query 执行SQL查询并返回结果
func (d *DB) Query(query string) (*[]map[string]interface{}, error) {
	db, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 执行查询
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Make a slice
	vals := make([]interface{}, len(columns))
	for i := range columns {
		vals[i] = new(sql.RawBytes)
	}

	// Fetch rows
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}

		vmap := make(map[string]interface{}, len(vals))
		for i, c := range vals {
			// 类型断言
			switch v := c.(type) {
			case *sql.RawBytes:
				if *v == nil {
					// nil在前端解析的是null，符合预期
					vmap[columns[i]] = "NULL"
				} else {
					vmap[columns[i]] = string(*v)
				}
			}
		}
		result = append(result, vmap)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckIfTableExists 检查表是否存在
// 返回值: (消息, 错误)
// - 如果表不存在: ("表或视图`xxx`不存在", error)
// - 如果表存在: ("表或视图`xxx`已存在", nil)
// - 如果数据库不存在: ("数据库`xxx`不存在", error)
// - 如果连接失败: ("", error)
func (d *DB) CheckIfTableExists(table string) (string, error) {
	db, err := d.Open()
	if err != nil {
		return "", fmt.Errorf("数据库连接失败: %w", err)
	}
	defer db.Close()

	// 尝试连接（Ping）来捕获数据库不存在的错误
	if err := db.Ping(); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			// MySQL error 1049: Unknown database
			if mysqlErr.Number == 1049 {
				return fmt.Sprintf("数据库`%s`不存在", d.Database), err
			}
		}
		// 检查错误消息中是否包含 "Unknown database" 或 "1049"
		errStr := err.Error()
		if strings.Contains(errStr, "Unknown database") || strings.Contains(errStr, "1049") {
			return fmt.Sprintf("数据库`%s`不存在", d.Database), err
		}
		// 其他连接错误，静默失败（不阻止审核）
		return "", err
	}

	// 连接成功，检查表是否存在
	query := fmt.Sprintf("DESC `%s`", table)
	_, err = db.Exec(query)
	if err != nil {
		// 检查是否是表不存在的错误 (MySQL error 1146)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1146 {
				return fmt.Sprintf("表或视图`%s`不存在", table), err
			}
		}
		// 检查错误消息中是否包含 "doesn't exist" 或 "1146"
		errStr := err.Error()
		if strings.Contains(errStr, "doesn't exist") || strings.Contains(errStr, "1146") {
			return fmt.Sprintf("表或视图`%s`不存在", table), err
		}
		// 其他错误，静默失败（不阻止审核）
		return "", fmt.Errorf("检查表存在性失败: %w", err)
	}
	return fmt.Sprintf("表或视图`%s`已存在", table), nil
}

// CheckIfTableExists 检查表是否存在（兼容老代码的函数形式）
func CheckIfTableExists(table string, db *DB) (string, error) {
	return db.CheckIfTableExists(table)
}

// CheckIfDatabaseExists 检查数据库是否存在
func CheckIfDatabaseExists(database string, db *DB) (string, error) {
	// Query the information_schema.schemata to check if the database exists
	result, err := db.Query(fmt.Sprintf("SELECT COUNT(*) as count FROM information_schema.schemata WHERE schema_name='%s'", database))
	if err != nil {
		return fmt.Sprintf("执行SQL失败, 主机: %s:%d, 错误: %s", db.Host, db.Port, err.Error()), err
	}
	var count int
	for _, row := range *result {
		count, _ = strconv.Atoi(row["count"].(string))
		break
	}
	if count == 0 {
		// Database does not exist
		return fmt.Sprintf("数据库`%s`不存在", database), errors.New("error")
	}
	// Database exists
	return fmt.Sprintf("数据库`%s`已存在", database), nil
}

// CheckIfTableExistsCrossDB 跨数据库检查表是否存在
func CheckIfTableExistsCrossDB(table string, db *DB) (string, error) {
	// Check if the table exists using information_schema.tables, suitable for cross-database checks
	result, err := db.Query(fmt.Sprintf("SELECT COUNT(*) as count FROM information_schema.tables WHERE table_name='%s'", table))
	if err != nil {
		return fmt.Sprintf("执行SQL失败, 主机: %s:%d, 错误: %s", db.Host, db.Port, err.Error()), err
	}
	var count int
	for _, row := range *result {
		count, _ = strconv.Atoi(row["count"].(string))
		break
	}
	if count == 0 {
		// Table does not exist
		return fmt.Sprintf("表或视图`%s`不存在", table), errors.New("error")
	}
	// Table exists
	return fmt.Sprintf("表或视图`%s`已存在", table), nil
}

// ShowCreateTable 获取表的创建语句
func ShowCreateTable(table string, db *DB, kv *kv.KVCache) (data interface{}, err error) {
	// Return table structure from cache if available
	data = kv.Get(table)
	if data != nil {
		return data, nil
	}
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", table)
	result, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var createStatement string
	for _, sql := range *result {
		// Table
		if _, ok := sql["Create Table"]; ok {
			createStatement = sql["Create Table"].(string)
		}
		// View
		if _, ok := sql["Create View"]; ok {
			createStatement = sql["Create View"].(string)
		}
	}

	var warns []error
	data, warns, err = parser.NewParse(createStatement, "", "")
	if len(warns) > 0 {
		return nil, fmt.Errorf("解析警告: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return nil, fmt.Errorf("SQL语法解析错误: %s", err.Error())
	}
	kv.Put(table, data)
	return data, nil
}

// GetDBVars 获取数据库变量
func GetDBVars(db *DB) (map[string]string, error) {
	result, err := db.Query(`SHOW VARIABLES WHERE Variable_name IN ('innodb_large_prefix','version','character_set_database','innodb_default_row_format')`)
	if err != nil {
		return nil, err
	}

	var data map[string]string = map[string]string{
		"dbVersion":              "",
		"dbCharset":              "utf8",
		"largePrefix":            "OFF",
		"innodbDefaultRowFormat": "dynamic",
	}

	// [map[Value:utf8 Variable_name:character_set_database] map[Value:5.7.35-log Variable_name:version]]
	for _, row := range *result {
		variableName, ok := row["Variable_name"].(string)
		if !ok {
			return nil, fmt.Errorf("Variable_name类型意外")
		}

		value, ok := row["Value"].(string)
		if !ok {
			return nil, fmt.Errorf("行中Value类型意外")
		}

		switch variableName {
		case "version":
			data["dbVersion"] = value
		case "character_set_database":
			data["dbCharset"] = value
		case "innodb_large_prefix":
			switch value {
			case "0":
				data["largePrefix"] = "OFF"
			case "1":
				data["largePrefix"] = "ON"
			default:
				data["largePrefix"] = strings.ToUpper(value)
			}
		case "innodb_default_row_format":
			data["innodbDefaultRowFormat"] = value
		}
	}
	return data, nil
}
