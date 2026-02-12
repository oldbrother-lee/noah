package initializer

import (
	"context"
	"go-noah/internal/model/insight"
	"go-noah/pkg/log"

	"gorm.io/gorm"
)

type InspectInitializer struct {
	logger *log.Logger
}

func NewInspectInitializer(logger *log.Logger) *InspectInitializer {
	return &InspectInitializer{logger: logger}
}

func (i *InspectInitializer) Name() string {
	return "inspect"
}

func (i *InspectInitializer) Order() int {
	return InitOrderInspect
}

func (i *InspectInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.AutoMigrate(&insight.InspectParams{})
}

func (i *InspectInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.Migrator().HasTable(&insight.InspectParams{})
}

func (i *InspectInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	// InspectParams 表不需要初始化数据，只需要创建表结构
	return true
}

func (i *InspectInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	// InspectParams 表不需要初始化数据，只需要创建表结构
	return nil
}
