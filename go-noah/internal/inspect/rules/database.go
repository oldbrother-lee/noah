package rules

import (
	"go-noah/internal/inspect/logics"
	"go-noah/internal/inspect/traverses"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

func CreateDatabaseRules() []Rule {
	return []Rule{
		{
			Hint:      "CreateDatabase#检查DB是否存在",
			CheckFunc: (*Rule).RuleCreateDatabaseIsExist,
		},
	}
}

// RuleCreateDatabaseIsExist
func (r *Rule) RuleCreateDatabaseIsExist(tistmt *ast.StmtNode) {
	v := &traverses.TraverseCreateDatabaseIsExist{}
	(*tistmt).Accept(v)
	logics.LogicCreateDatabaseIsExist(v, r.RuleHint)
}
