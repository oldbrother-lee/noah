package logics

import (
	"go-noah/internal/inspect"
	"go-noah/internal/inspect/dao"
	"go-noah/internal/inspect/traverses"
)

// LogicCreateDatabaseIsExist
func LogicCreateDatabaseIsExist(v *traverses.TraverseCreateDatabaseIsExist, r *inspect.RuleHint) {
	if msg, err := dao.CheckIfDatabaseExists(v.Name, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
		r.IsSkipNextStep = true
	}
}
