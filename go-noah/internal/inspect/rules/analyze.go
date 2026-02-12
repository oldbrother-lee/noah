/*
@Time    :   2023/04/19 15:10:11
@Author  :   xff
@Desc    :
*/

package rules

import (
	"go-noah/internal/inspect/logics"
	"go-noah/internal/inspect/traverses"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

func AnalyzeTableRules() []Rule {
	return []Rule{
		{
			Hint:      "AnalyzeTable#检查",
			CheckFunc: (*Rule).RuleAnalyzeTable,
		},
	}
}

// RuleAnalyzeTable
func (r *Rule) RuleAnalyzeTable(tistmt *ast.StmtNode) {
	v := &traverses.TraverseAnalyzeTable{}
	(*tistmt).Accept(v)
	logics.LogicAnalyzeTable(v, r.RuleHint)
}
