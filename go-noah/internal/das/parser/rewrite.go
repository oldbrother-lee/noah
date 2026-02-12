/*
@Time    :   2023/03/24 10:06:56
@Author  :   xff
@Desc    :   重写sql
*/

package parser

import (
	"fmt"
	"go-noah/pkg/global"
	"strings"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/model"
	driver "github.com/pingcap/tidb/pkg/types/parser_driver"
	utilparser "github.com/pingcap/tidb/pkg/util/parser"
)

type Rewrite struct {
	Stmt      ast.StmtNode
	RequestID string
	DbType    string
}

func (r *Rewrite) BindHints(name string, value interface{}) (hints []*ast.TableOptimizerHint) {
	// 绑定hints
	var hint *ast.TableOptimizerHint = &ast.TableOptimizerHint{
		HintName: model.CIStr{O: name, L: name},
		HintData: value,
	}
	hints = append(hints, hint)
	return hints
}

func (r *Rewrite) AddHintsForMaxExecutionTime() {
	// 增加max_execution_time
	switch stmt := r.Stmt.(type) {
	case *ast.SelectStmt:
		// 从配置读取 MaxExecutionTime，默认 600000 毫秒（10分钟）
		maxTime := global.Conf.GetUint64("das.max_execution_time")
		if maxTime == 0 {
			maxTime = 600000 // 默认 10 分钟
		}
		hints := r.BindHints("max_execution_time", maxTime)
		hints = append(hints, stmt.TableHints...)
		stmt.TableHints = hints
		r.Stmt = stmt
	}
}

func (r *Rewrite) RewriteLimitSetCount(stmt ast.StmtNode, value uint64) {
	// 重写count
	switch stmt := r.Stmt.(type) {
	case *ast.SelectStmt:
		if stmt.Limit == nil {
			stmt.Limit = &ast.Limit{}
			stmt.Limit.Count = &driver.ValueExpr{}
		}
		switch ex := stmt.Limit.Count.(type) {
		case *driver.ValueExpr:
			ex.SetValue(value)
		}
	case *ast.SetOprStmt:
		if stmt.Limit == nil {
			stmt.Limit = &ast.Limit{}
			stmt.Limit.Count = &driver.ValueExpr{}
		}
		switch ex := stmt.Limit.Count.(type) {
		case *driver.ValueExpr:
			ex.SetValue(value)
		}
	}
}

func (r *Rewrite) RewriteLimit() {
	// 重写limit
	switch r.Stmt.(type) {
	case *ast.SelectStmt, *ast.SetOprStmt:
		// 遍历
		v := &Limit{}
		r.Stmt.Accept(v)

		// 从配置读取限制值
		defaultReturnRows := global.Conf.GetUint64("das.default_return_rows")
		if defaultReturnRows == 0 {
			defaultReturnRows = 1000 // 默认 1000 行
		}
		maxReturnRows := global.Conf.GetUint64("das.max_return_rows")
		if maxReturnRows == 0 {
			maxReturnRows = 10000 // 默认 10000 行
		}

		// SQL语句没有limit子句，增加limit N
		if v.Count == 0 {
			r.RewriteLimitSetCount(r.Stmt, defaultReturnRows)
		} else {
			// SQL语句有limit N子句
			// 只有当N大于MaxReturnRows时，才改写为limit MaxReturnRows
			// 如果N在defaultReturnRows和maxReturnRows之间，保持不变
			if v.Count > maxReturnRows {
				r.RewriteLimitSetCount(r.Stmt, maxReturnRows)
			}
			// 如果 v.Count <= maxReturnRows，保持不变，不需要重写
		}
	}
}

func (r *Rewrite) RewriteExplain() {
	// explain
	switch stmt := r.Stmt.(type) {
	case *ast.ExplainStmt:
		switch stmt.Stmt.(type) {
		case *ast.SelectStmt, *ast.SetOprStmt:
			// mysql没有row格式，仅有traditional，json格式
			if strings.EqualFold(r.DbType, "mysql") && strings.EqualFold(stmt.Format, "row") {
				stmt.Format = "traditional"
			}
		}
		r.Stmt = stmt
	}
}

func (r *Rewrite) ReplaceClickHouseExplain(sql string) string {
	// 如果是clickhouse explain，移除format子句
	switch stmt := r.Stmt.(type) {
	case *ast.ExplainStmt:
		switch stmt.Stmt.(type) {
		case *ast.SelectStmt, *ast.SetOprStmt:
			if strings.EqualFold(r.DbType, "clickhouse") {
				return strings.Replace(sql, "FORMAT = 'row'", "", 1)
			}
		}
	}
	return sql
}

func (r *Rewrite) RestoreSQL() string {
	// 从ast还原SQL
	return utilparser.RestoreWithDefaultDB(r.Stmt, "", "")
}

func (r *Rewrite) AddCommentForRequestID(sql string) string {
	// SQL增加request_id注释和系统标识
	if r.RequestID != "" {
		return strings.Join([]string{sql, fmt.Sprintf("/* [go-noah] %s */", r.RequestID)}, " ")
	}
	return sql
}

func (r *Rewrite) Run() string {
	r.AddHintsForMaxExecutionTime()
	r.RewriteLimit()
	r.RewriteExplain()
	restoreSQL := r.RestoreSQL()
	replaceRestoreSQL := r.ReplaceClickHouseExplain(restoreSQL)
	return r.AddCommentForRequestID(replaceRestoreSQL)
}
