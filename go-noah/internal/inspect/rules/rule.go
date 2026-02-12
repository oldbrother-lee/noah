/*
@Time    :   2022/06/29 15:30:31
@Author  :   xff
*/

package rules

import (
	"go-noah/internal/inspect"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

type Rule struct {
	*inspect.RuleHint
	Hint      string                     `json:"hint"` // 规则说明
	CheckFunc func(*Rule, *ast.StmtNode) // 函数名
}
