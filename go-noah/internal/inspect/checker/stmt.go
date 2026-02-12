package checker

import (
	"fmt"
	"regexp"
	"strings"

	"go-noah/internal/inspect"
	"go-noah/internal/inspect/config"
	"go-noah/internal/inspect/dao"
	"go-noah/internal/inspect/parser"
	"go-noah/internal/inspect/rules"
	"go-noah/pkg/kv"
	"go-noah/pkg/utils"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

// SyntaxInspectService 语法审核服务（适配规则引擎）
type SyntaxInspectService struct {
	DB            *dao.DB
	Audit         *parser.Audit
	InspectParams *config.InspectParams
	Charset       string
	Collation     string
}

// Stmt 语句检查器
type Stmt struct {
	*SyntaxInspectService
}

// CreateTableStmt 检查 CreateTable 语句
func (s *Stmt) CreateTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "CreateTable", Level: "INFO"}

	for _, rule := range rules.CreateTableRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// CreateViewStmt 检查 CreateView 语句
func (s *Stmt) CreateViewStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "CreateView", Level: "INFO"}

	for _, rule := range rules.CreateViewRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// CreateDatabaseStmt 检查 CreateDatabase 语句
func (s *Stmt) CreateDatabaseStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "CreateDatabase", Level: "INFO"}

	for _, rule := range rules.CreateDatabaseRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// RenameTableStmt 检查 RenameTable 语句
func (s *Stmt) RenameTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "RenameTable", Level: "INFO"}

	for _, rule := range rules.RenameTableRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// AnalyzeTableStmt 检查 AnalyzeTable 语句
func (s *Stmt) AnalyzeTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "AnalyzeTable", Level: "INFO"}

	for _, rule := range rules.AnalyzeTableRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// DropTableStmt 检查 DropTable 语句
func (s *Stmt) DropTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DropTable", Level: "INFO"}

	for _, rule := range rules.DropTableRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// AlterTableStmt 检查 AlterTable 语句
func (s *Stmt) AlterTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) (ReturnData, string) {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "AlterTable", Level: "INFO"}
	var mergeAlter string
	
	// 禁止使用ALTER TABLE...ADD CONSTRAINT...语法
	tmpCompile := regexp.MustCompile(`(?is:.*alter.*table.*add.*constraint.*)`)
	match := tmpCompile.MatchString(stmt.Text())
	if match {
		data.Level = "WARN"
		data.Summary = append(data.Summary, "禁止使用ALTER TABLE...ADD CONSTRAINT...语法")
		return data, mergeAlter
	}

	// 从 AST 中提取完整的表名（包含数据库名，如果存在）
	if alterStmt, ok := stmt.(*ast.AlterTableStmt); ok {
		if alterStmt.Table.Schema.String() != "" {
			// 如果指定了数据库名，使用 database.table 格式
			mergeAlter = fmt.Sprintf("%s.%s", alterStmt.Table.Schema.String(), alterStmt.Table.Name.String())
		} else {
			// 如果没有指定数据库名，使用当前连接的数据库名
			if s.DB != nil && s.DB.Database != "" {
				mergeAlter = fmt.Sprintf("%s.%s", s.DB.Database, alterStmt.Table.Name.String())
			} else {
				// 如果都没有，只使用表名（向后兼容）
				mergeAlter = alterStmt.Table.Name.String()
			}
		}
	}

	for _, rule := range rules.AlterTableRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)
		
		// 如果规则设置了 MergeAlter，但我们已经从 AST 提取了完整表名，优先使用 AST 提取的
		// 这样可以确保包含数据库名
		if len(rule.RuleHint.MergeAlter) > 0 && len(mergeAlter) == 0 {
			mergeAlter = rule.RuleHint.MergeAlter
		}
		if len(rule.RuleHint.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data, mergeAlter
}

// DMLStmt 检查 DML 语句
func (s *Stmt) DMLStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DML", Level: "INFO"}

	for _, rule := range rules.DMLRules() {
		var ruleHint *inspect.RuleHint = &inspect.RuleHint{
			DB:            s.DB,
			KV:            kv,
			Query:         stmt.Text(),
			InspectParams: s.InspectParams,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &stmt)

		// 当为DML语句时，赋值AffectedRows
		data.AffectedRows = rule.RuleHint.AffectedRows

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// mergeAlters 判断多条alter语句是否需要合并
func (s *SyntaxInspectService) mergeAlters(kv *kv.KVCache, mergeAlters []string) ReturnData {
	var data ReturnData = ReturnData{Level: "INFO"}
	dbVersion := kv.Get("dbVersion")
	if dbVersion != nil {
		versionStr := dbVersion.(string)
		// 如果不是 TiDB，则检查重复的 ALTER
		if !strings.Contains(strings.ToLower(versionStr), "tidb") {
			if s.InspectParams.ENABLE_MYSQL_MERGE_ALTER_TABLE {
				if ok, val := utils.IsRepeat(mergeAlters); ok {
					for _, v := range val {
						data.Summary = append(data.Summary, fmt.Sprintf("[MySQL数据库]表`%s`的多条ALTER操作，请合并为一条ALTER语句", v))
					}
				}
			}
		} else {
			// TiDB
			if s.InspectParams.ENABLE_TIDB_MERGE_ALTER_TABLE {
				if ok, val := utils.IsRepeat(mergeAlters); ok {
					for _, v := range val {
						data.Summary = append(data.Summary, fmt.Sprintf("[TiDB数据库]表`%s`的多条ALTER操作，请合并为一条ALTER语句", v))
					}
				}
			}
		}
	}
	if len(data.Summary) > 0 {
		data.Level = "WARN"
	}
	return data
}

