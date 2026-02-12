/*
@Time    :   2022/06/29 15:30:31
@Author  :   xff
*/

package inspect

import (
	"go-noah/internal/inspect/config"
	"go-noah/internal/inspect/dao"
	"go-noah/pkg/kv"
)

type RuleHint struct {
	Summary        []string `json:"summary"`       // 摘要
	AffectedRows   int      `json:"affected_rows"` // 默认为0
	IsSkipNextStep bool     // 是否跳过接下来的检查步骤
	DB             *dao.DB
	KV             *kv.KVCache
	Query          string // 原始SQL
	MergeAlter     string
	InspectParams  *config.InspectParams
}

