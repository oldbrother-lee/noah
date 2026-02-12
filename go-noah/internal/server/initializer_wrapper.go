package server

import (
	"context"
	"go-noah/internal/server/initializer"
	"go-noah/pkg/log"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

// NewInitializerManager 创建初始化管理器（包装函数，方便外部调用）
func NewInitializerManager(db *gorm.DB, logger *log.Logger) *initializer.Manager {
	return initializer.NewManager(db, logger)
}

// RegisterAllInitializers 注册所有初始化器（包装函数，方便外部调用）
func RegisterAllInitializers(logger *log.Logger, enforcer *casbin.SyncedEnforcer) {
	initializer.RegisterAll(logger, enforcer)
}

// InitializeAllData 执行所有数据初始化（用于迁移工具）
func InitializeAllData(ctx context.Context, db *gorm.DB, logger *log.Logger) error {
	manager := NewInitializerManager(db, logger)
	return manager.InitializeAll(ctx)
}

// InitializeDataIfNeeded 按需初始化数据（用于服务器启动）
func InitializeDataIfNeeded(ctx context.Context, db *gorm.DB, logger *log.Logger) error {
	manager := NewInitializerManager(db, logger)
	return manager.InitializeIfNeeded(ctx)
}
