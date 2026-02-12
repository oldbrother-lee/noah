package initializer

import (
	"go-noah/pkg/log"

	"github.com/casbin/casbin/v2"
)

// RegisterAll 注册所有初始化器
// 需要在 logger 和 enforcer 初始化后调用
func RegisterAll(logger *log.Logger, enforcer *casbin.SyncedEnforcer) {
	// 注册基础初始化器
	Register(NewRoleInitializer(logger))
	Register(NewUserInitializer(logger))
	Register(NewMenuInitializer(logger))

	// 注册 RBAC 初始化器（需要 enforcer）
	if enforcer != nil {
		Register(NewRBACInitializer(logger, enforcer))
	}

	// 注册 Inspect 初始化器
	Register(NewInspectInitializer(logger))

	// TODO: 注册其他初始化器
	// Register(NewAPIInitializer(logger))
	// Register(NewFlowInitializer(logger))
}
