package noah

import (
	"context"
	"go-noah/internal/job"
	"go-noah/internal/repository"
	insightRepo "go-noah/internal/repository/insight"
	"go-noah/internal/server"
	"go-noah/internal/task"
	"go-noah/pkg/app"
	"go-noah/pkg/global"
	"go-noah/pkg/jwt"
	"go-noah/pkg/log"
	"go-noah/pkg/sid"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewServerApp(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	// 只初始化基础设施（存入 global）
	global.Sid = sid.NewSid()
	global.JWT = jwt.NewJwt(conf)
	global.DB = repository.NewDB(conf, logger)
	global.Enforcer = repository.NewCasbinEnforcer(conf, logger, global.DB)
	global.Redis = repository.NewRedis(conf) // 初始化 Redis（用于 WebSocket 消息推送）
	global.Logger = logger
	global.Conf = conf

	// 自动迁移数据库表（创建不存在的表）
	if err := server.AutoMigrateTables(global.DB, logger); err != nil {
		logger.Error("自动迁移数据库表失败", zap.Error(err))
		// 不阻止服务启动，只记录错误
	}

	// 使用新的初始化管理器
	initManager := server.NewInitializerManager(global.DB, logger)
	
	// 注册所有初始化器
	server.RegisterAllInitializers(logger, global.Enforcer)
	
	// 按需初始化基础数据（如果不存在则自动初始化）
	// 初始化顺序由初始化器的 Order() 方法决定：Role -> User -> Menu -> RBAC
	ctx := context.Background()
	if err := initManager.InitializeIfNeeded(ctx); err != nil {
		logger.Warn("初始化数据失败", zap.Error(err))
		// 不阻止服务启动，只记录警告
	}

	// 初始化流程定义（如果不存在则自动初始化）
	if err := server.InitializeFlowDefinitionsIfNeeded(global.DB, logger); err != nil {
		logger.Error("初始化流程定义失败", zap.Error(err))
		// 不阻止服务启动，只记录错误
	}

	// 初始化审核参数（如果不存在则自动初始化）
	if err := server.InitializeInspectParamsIfNeeded(global.DB, logger); err != nil {
		logger.Error("初始化审核参数失败", zap.Error(err))
		// 不阻止服务启动，只记录错误
	}

	// 创建 Repository 和 Transaction（不存储在 global，避免循环导入）
	repo := repository.NewRepository(logger, global.DB, global.Enforcer)
	transaction := repository.NewTransaction(repo)

	httpServer := server.NewHTTPServer(logger, conf, global.JWT, global.Enforcer)

	jobBase := job.NewJob(transaction, global.Logger, global.Sid)
	userRepo := repository.NewUserRepository(repo)
	userJob := job.NewUserJob(jobBase, userRepo)
	jobServer := server.NewJobServer(global.Logger, userJob)

	// HTTP 服务器不处理定时任务，只负责处理 HTTP 请求
	// 定时任务的注册和执行由 TaskServer 负责

	a := app.NewApp(
		app.WithServer(httpServer, jobServer),
		app.WithName("noah-server"),
	)

	cleanup := func() {
		sqlDB, _ := global.DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		if global.Redis != nil {
			_ = global.Redis.Close()
		}
	}
	return a, cleanup, nil
}

func NewTaskApp(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	// 只初始化基础设施（存入 global）
	global.Sid = sid.NewSid()
	global.DB = repository.NewDB(conf, logger)
	global.Enforcer = repository.NewCasbinEnforcer(conf, logger, global.DB)
	global.Logger = logger
	global.Conf = conf

	// 创建 Repository 和 Transaction（不存储在 global，避免循环导入）
	repo := repository.NewRepository(logger, global.DB, global.Enforcer)
	transaction := repository.NewTransaction(repo)

	userRepo := repository.NewUserRepository(repo)
	insightRepo := insightRepo.NewInsightRepository(repo, logger, global.Enforcer)

	tk := task.NewTask(transaction, global.Logger, global.Sid)
	userTask := task.NewUserTask(tk, userRepo)
	insightTask := task.NewInsightTask(tk, insightRepo)
	taskServer := server.NewTaskServer(global.Logger, userTask, insightTask, conf)

	a := app.NewApp(
		app.WithServer(taskServer),
		app.WithName("noah-task"),
	)
	cleanup := func() {
		sqlDB, _ := global.DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return a, cleanup, nil
}

func NewMigrateApp(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	// 只初始化基础设施（存入 global）
	global.Sid = sid.NewSid()
	global.DB = repository.NewDB(conf, logger)
	global.Enforcer = repository.NewCasbinEnforcer(conf, logger, global.DB)
	global.Logger = logger
	global.Conf = conf

	migrateServer := server.NewMigrateServer(global.DB, global.Logger, global.Sid, global.Enforcer)

	a := app.NewApp(
		app.WithServer(migrateServer),
		app.WithName("noah-migrate"),
	)
	cleanup := func() {
		sqlDB, _ := global.DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
	return a, cleanup, nil
}
