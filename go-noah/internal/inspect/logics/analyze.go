/*
@Time    :   2023/04/19 15:09:38
@Author  :   xff
@Desc    :
*/

package logics

import (
	"go-noah/internal/inspect"
	"go-noah/internal/inspect/dao"
	"go-noah/internal/inspect/process"
	"go-noah/internal/inspect/traverses"
)

// LogicRenameTable
func LogicAnalyzeTable(v *traverses.TraverseAnalyzeTable, r *inspect.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	dbVersionIns := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	if !dbVersionIns.IsTiDB() {
		r.Summary = append(r.Summary, "仅允许TiDB提交Analyze table语法")
		return
	}
	// 表必须存在
	for _, table := range v.TableNames {
		if msg, err := dao.CheckIfTableExists(table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
