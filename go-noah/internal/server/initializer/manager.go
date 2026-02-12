package initializer

import (
	"context"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager 初始化管理器
type Manager struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewManager 创建初始化管理器
func NewManager(db *gorm.DB, logger *log.Logger) *Manager {
	return &Manager{
		db:     db,
		logger: logger,
	}
}

// InitializeAll 执行所有初始化（完整初始化，用于迁移工具）
func (m *Manager) InitializeAll(ctx context.Context) error {
	initializers := GetAll()

	// 第一步：创建所有表
	m.logger.Info("开始创建数据库表...")
	for _, init := range initializers {
		if init.IsTableCreated(ctx, m.db) {
			m.logger.Debug("表已存在，跳过", zap.String("name", init.Name()))
			continue
		}

		if err := init.MigrateTable(ctx, m.db); err != nil {
			m.logger.Error("创建表失败", zap.String("name", init.Name()), zap.Error(err))
			return err
		}
		m.logger.Info("创建表成功", zap.String("name", init.Name()))
	}

	// 第二步：初始化数据
	m.logger.Info("开始初始化数据...")
	for _, init := range initializers {
		if init.IsDataInitialized(ctx, m.db) {
			m.logger.Debug("数据已存在，跳过", zap.String("name", init.Name()))
			continue
		}

		if err := init.InitializeData(ctx, m.db); err != nil {
			m.logger.Error("初始化数据失败", zap.String("name", init.Name()), zap.Error(err))
			// 可以选择继续或返回错误
			// return err
		} else {
			m.logger.Info("初始化数据成功", zap.String("name", init.Name()))
		}
	}

	m.logger.Info("所有初始化完成")
	return nil
}

// InitializeIfNeeded 按需初始化（服务器启动时调用，只初始化数据，不创建表）
func (m *Manager) InitializeIfNeeded(ctx context.Context) error {
	initializers := GetAll()

	// 只初始化基础数据（表结构已在 AutoMigrate 中创建）
	for _, init := range initializers {
		if init.IsDataInitialized(ctx, m.db) {
			continue
		}

		if err := init.InitializeData(ctx, m.db); err != nil {
			m.logger.Warn("初始化数据失败", zap.String("name", init.Name()), zap.Error(err))
			// 不阻止服务启动
		} else {
			m.logger.Debug("初始化数据成功", zap.String("name", init.Name()))
		}
	}

	return nil
}

// MigrateTablesOnly 只创建表结构，不初始化数据
func (m *Manager) MigrateTablesOnly(ctx context.Context) error {
	initializers := GetAll()

	m.logger.Info("开始创建数据库表...")
	for _, init := range initializers {
		if init.IsTableCreated(ctx, m.db) {
			m.logger.Debug("表已存在，跳过", zap.String("name", init.Name()))
			continue
		}

		if err := init.MigrateTable(ctx, m.db); err != nil {
			m.logger.Error("创建表失败", zap.String("name", init.Name()), zap.Error(err))
			return err
		}
		m.logger.Info("创建表成功", zap.String("name", init.Name()))
	}

	return nil
}
