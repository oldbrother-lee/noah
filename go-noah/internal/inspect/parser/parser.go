package parser

import (
	"fmt"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
)

// Audit SQL审核结构
type Audit struct {
	Query  string         // 原始SQL
	TiStmt []ast.StmtNode // 通过TiDB解析出的抽象语法树
}

// NewParse 解析SQL
func NewParse(sqltext, charset, collation string) (*Audit, []error, error) {
	q := &Audit{Query: sqltext}

	// TiDB parser 语法解析
	var warns []error
	var err error
	q.TiStmt, warns, err = parser.New().Parse(sqltext, charset, collation)
	return q, warns, err
}

// ParseSQL 简化版解析（使用默认字符集）
func ParseSQL(sqltext string) (*Audit, []error, error) {
	return NewParse(sqltext, "", "")
}

// GetStatementType 获取语句类型
func GetStatementType(stmt ast.StmtNode) string {
	switch stmt.(type) {
	case *ast.CreateTableStmt:
		return "CREATE_TABLE"
	case *ast.AlterTableStmt:
		return "ALTER_TABLE"
	case *ast.DropTableStmt:
		return "DROP_TABLE"
	case *ast.TruncateTableStmt:
		return "TRUNCATE_TABLE"
	case *ast.RenameTableStmt:
		return "RENAME_TABLE"
	case *ast.CreateIndexStmt:
		return "CREATE_INDEX"
	case *ast.DropIndexStmt:
		return "DROP_INDEX"
	case *ast.CreateDatabaseStmt:
		return "CREATE_DATABASE"
	case *ast.DropDatabaseStmt:
		return "DROP_DATABASE"
	case *ast.CreateViewStmt:
		return "CREATE_VIEW"
	case *ast.InsertStmt:
		return "INSERT"
	case *ast.UpdateStmt:
		return "UPDATE"
	case *ast.DeleteStmt:
		return "DELETE"
	case *ast.SelectStmt:
		return "SELECT"
	case *ast.SetStmt:
		return "SET"
	case *ast.UseStmt:
		return "USE"
	case *ast.ShowStmt:
		return "SHOW"
	case *ast.ExplainStmt:
		return "EXPLAIN"
	case *ast.AnalyzeTableStmt:
		return "ANALYZE"
	default:
		return "UNKNOWN"
	}
}

// IsDDL 判断是否为DDL语句
func IsDDL(stmt ast.StmtNode) bool {
	switch stmt.(type) {
	case *ast.CreateTableStmt, *ast.AlterTableStmt, *ast.DropTableStmt,
		*ast.TruncateTableStmt, *ast.RenameTableStmt, *ast.CreateIndexStmt,
		*ast.DropIndexStmt, *ast.CreateDatabaseStmt, *ast.DropDatabaseStmt,
		*ast.CreateViewStmt:
		return true
	}
	return false
}

// IsDML 判断是否为DML语句
func IsDML(stmt ast.StmtNode) bool {
	switch stmt.(type) {
	case *ast.InsertStmt, *ast.UpdateStmt, *ast.DeleteStmt:
		return true
	}
	return false
}

// IsSelect 判断是否为SELECT语句
func IsSelect(stmt ast.StmtNode) bool {
	_, ok := stmt.(*ast.SelectStmt)
	return ok
}

// CheckSqlType 检查SQL类型是否匹配工单类型
// sqltext: SQL文本
// sqltype: 工单类型 (DDL/DML/EXPORT)
func CheckSqlType(sqltext, sqltype string) error {
	audit, warns, err := ParseSQL(sqltext)
	if err != nil {
		return err
	}
	if len(warns) > 0 {
		// 有警告但不影响类型判断
	}

	for _, stmt := range audit.TiStmt {
		var st string
		switch stmt.(type) {
		case *ast.SelectStmt, *ast.SetOprStmt:
			st = "EXPORT"
		case *ast.DeleteStmt, *ast.InsertStmt, *ast.UpdateStmt:
			st = "DML"
		case *ast.AlterTableStmt, *ast.AlterSequenceStmt, *ast.AlterPlacementPolicyStmt:
			st = "DDL"
		case *ast.CreateDatabaseStmt, *ast.CreateIndexStmt, *ast.CreateTableStmt, *ast.CreateViewStmt, *ast.CreateSequenceStmt, *ast.CreatePlacementPolicyStmt:
			st = "DDL"
		case *ast.DropDatabaseStmt, *ast.DropIndexStmt, *ast.DropTableStmt, *ast.DropSequenceStmt, *ast.DropPlacementPolicyStmt:
			st = "DDL"
		case *ast.RenameTableStmt:
			st = "DDL"
		case *ast.TruncateTableStmt:
			st = "DDL"
		default:
			// 未知类型，跳过检查
			continue
		}
		if st != sqltype {
			if sqltype == "DML" {
				return fmt.Errorf("DML模式下，不允许提交%s语句", st)
			}
			if sqltype == "DDL" {
				return fmt.Errorf("DDL模式下，不允许提交%s语句", st)
			}
		}
	}
	return nil
}

// SplitSQLText 拆分SQL文本为多个SQL语句
func SplitSQLText(sqltext string) ([]string, error) {
	audit, warns, err := ParseSQL(sqltext)
	if err != nil {
		return nil, err
	}
	if len(warns) > 0 {
		// 有警告但不影响拆分
	}

	var sqls []string
	for _, stmt := range audit.TiStmt {
		sqls = append(sqls, stmt.Text())
	}
	return sqls, nil
}

// GetSqlStatement 获取SQL语句类型（用于执行器判断执行方式）
// 返回：CreateDatabase, CreateTable, CreateView, DropTable, DropIndex, TruncateTable,
//      RenameTable, CreateIndex, DropDatabase, AlterTable 等
func GetSqlStatement(sqltext string) (string, error) {
	stmt, err := parser.New().ParseOneStmt(sqltext, "", "")
	if err != nil {
		return "", fmt.Errorf("SQL解析错误:%s", err.Error())
	}

	switch stmt.(type) {
	case *ast.AlterTableStmt:
		return "AlterTable", nil
	case *ast.CreateDatabaseStmt:
		return "CreateDatabase", nil
	case *ast.CreateIndexStmt:
		return "CreateIndex", nil
	case *ast.CreateTableStmt:
		return "CreateTable", nil
	case *ast.CreateViewStmt:
		return "CreateView", nil
	case *ast.DropIndexStmt:
		return "DropIndex", nil
	case *ast.DropTableStmt:
		return "DropTable", nil
	case *ast.RenameTableStmt:
		return "RenameTable", nil
	case *ast.TruncateTableStmt:
		return "TruncateTable", nil
	case *ast.DropDatabaseStmt:
		return "DropDatabase", nil
	default:
		return "", fmt.Errorf("当前SQL未匹配到规则，执行失败")
	}
}

// GetTableNameFromAlterStatement 从 ALTER TABLE 语句中提取表名
func GetTableNameFromAlterStatement(sqltext string) (string, error) {
	stmt, err := parser.New().ParseOneStmt(sqltext, "", "")
	if err != nil {
		return "", fmt.Errorf("SQL解析错误:%s", err.Error())
	}

	switch s := stmt.(type) {
	case *ast.AlterTableStmt:
		if s.Table.Schema.String() != "" {
			return fmt.Sprintf("%s.%s", s.Table.Schema.String(), s.Table.Name.String()), nil
		}
		return s.Table.Name.String(), nil
	default:
		return "", fmt.Errorf("未提取到表名，当前SQL不是ALTER TABLE语句")
	}
}

