package initializer

import (
	"context"
	"go-noah/internal/model/insight"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InsightInitializer 负责 Insight 相关基础数据初始化（如 das_allowed_operations）
type InsightInitializer struct {
	logger *log.Logger
}

func NewInsightInitializer(logger *log.Logger) *InsightInitializer {
	return &InsightInitializer{logger: logger}
}

func (i *InsightInitializer) Name() string {
	return "insight"
}

func (i *InsightInitializer) Order() int {
	return InitOrderInsight
}

// MigrateTable 确保表结构存在（完整初始化场景使用）
func (i *InsightInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.AutoMigrate(&insight.DASAllowedOperation{})
}

func (i *InsightInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.Migrator().HasTable(&insight.DASAllowedOperation{})
}

// IsDataInitialized 判断 das_allowed_operations 是否已经有数据
func (i *InsightInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	var count int64
	if err := db.WithContext(ctx).
		Model(&insight.DASAllowedOperation{}).
		Count(&count).Error; err != nil {
		i.logger.Warn("检查 das_allowed_operations 数据是否已初始化失败", zap.Error(err))
		return false
	}
	return count > 0
}

// InitializeData 初始化 das_allowed_operations 默认数据
// 参考 goInsight 中的 initializeAllowedOperations
func (i *InsightInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	ops := []struct {
		Name     string
		IsEnable bool
		Remark   string
	}{
		// 基础查询和 EXPLAIN 能力
		{"SELECT", true, ""},
		{"UNION", true, ""},
		{"Use", true, ""},
		{"Desc", true, ""},
		{"ExplainSelect", true, ""},
		{"ExplainDelete", true, ""},
		{"ExplainInsert", true, ""},
		{"ExplainUpdate", true, ""},
		{"ExplainUnion", true, ""},
		{"ExplainFor", true, "ExplainForStmt is a statement to provite information about how is SQL statement executeing in connection #ConnectionID"},

		// 其余 Show* / 元信息相关语句，默认关闭（需要时可在管理界面或手动开启）
		{"ShowEngines", false, ""},
		{"ShowDatabases", false, ""},
		{"ShowTables", false, ""},
		{"ShowTableStatus", false, ""},
		{"ShowColumns", false, ""},
		{"ShowWarnings", false, ""},
		{"ShowCharset", false, ""},
		{"ShowVariables", false, ""},
		{"ShowStatus", false, ""},
		{"ShowCollation", false, ""},
		{"ShowCreateTable", false, ""},
		{"ShowCreateView", false, ""},
		{"ShowCreateUser", false, ""},
		{"ShowCreateSequence", false, ""},
		{"ShowCreatePlacementPolicy", false, ""},
		{"ShowGrants", false, ""},
		{"ShowTriggers", false, ""},
		{"ShowProcedureStatus", false, ""},
		{"ShowIndex", false, ""},
		{"ShowProcessList", false, ""},
		{"ShowCreateDatabase", false, ""},
		{"ShowConfig", false, ""},
		{"ShowEvents", false, ""},
		{"ShowStatsExtended", false, ""},
		{"ShowStatsMeta", false, ""},
		{"ShowStatsHistograms", false, ""},
		{"ShowStatsTopN", false, ""},
		{"ShowStatsBuckets", false, ""},
		{"ShowStatsHealthy", false, ""},
		{"ShowStatsLocked", false, ""},
		{"ShowHistogramsInFlight", false, ""},
		{"ShowColumnStatsUsage", false, ""},
		{"ShowPlugins", false, ""},
		{"ShowProfile", false, ""},
		{"ShowProfiles", false, ""},
		{"ShowMasterStatus", false, ""},
		{"ShowPrivileges", false, ""},
		{"ShowErrors", false, ""},
		{"ShowBindings", false, ""},
		{"ShowBindingCacheStatus", false, ""},
		{"ShowPumpStatus", false, ""},
		{"ShowDrainerStatus", false, ""},
		{"ShowOpenTables", false, ""},
		{"ShowAnalyzeStatus", false, ""},
		{"ShowRegions", false, ""},
		{"ShowBuiltins", false, ""},
		{"ShowTableNextRowId", false, ""},
		{"ShowBackups", false, ""},
		{"ShowRestores", false, ""},
		{"ShowPlacement", false, ""},
		{"ShowPlacementForDatabase", false, ""},
		{"ShowPlacementForTable", false, ""},
		{"ShowPlacementForPartition", false, ""},
		{"ShowPlacementLabels", false, ""},
		{"ShowSessionStates", false, ""},
	}

	for _, op := range ops {
		record := &insight.DASAllowedOperation{
			Name:     op.Name,
			IsEnable: op.IsEnable,
			Remark:   op.Remark,
		}
		if err := db.WithContext(ctx).
			Where("name = ?", op.Name).
			FirstOrCreate(record).Error; err != nil {
			return err
		}
	}

	i.logger.Info("初始化 das_allowed_operations 默认数据完成")
	return nil
}

