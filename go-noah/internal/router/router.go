package router

import (
	"go-noah/internal/handler"
	insight "go-noah/internal/handler/insight"
	"go-noah/internal/middleware"
	"go-noah/pkg/jwt"
	"go-noah/pkg/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter(r *gin.Engine, jwt *jwt.JWT, e *casbin.SyncedEnforcer, logger *log.Logger) {
	// 添加 /api 前缀的路由组（兼容前端请求）
	api := r.Group("/api")
	{
		InitAdminRouter(api, jwt, e, logger)
		InitUserRouter(api, jwt, e, logger)
		InitInsightRouter(api, jwt, e, logger) // goInsight 功能路由
	}

	// 保持原有的 /v1 路由（向后兼容）
	InitAdminRouter(r, jwt, e, logger)
	InitUserRouter(r, jwt, e, logger)
	InitInsightRouter(r, jwt, e, logger)

	// WebSocket 路由（不需要认证，因为 WebSocket 升级在 handler 中处理）
	r.GET("/ws/:channel", insight.WebSocketHandlerApp.HandleWebSocket)
}

// InitAdminRouter 初始化管理员相关路由
func InitAdminRouter(r gin.IRouter, jwt *jwt.JWT, e *casbin.SyncedEnforcer, logger *log.Logger) {
	v1 := r.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/login", handler.AdminHandlerApp.Login)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger), middleware.AuthMiddleware(e))
		{
			strictAuthRouter.GET("/menus", handler.AdminHandlerApp.GetMenus)
			strictAuthRouter.GET("/admin/menus", handler.AdminHandlerApp.GetAdminMenus)
			strictAuthRouter.POST("/admin/menu", handler.AdminHandlerApp.MenuCreate)
			strictAuthRouter.PUT("/admin/menu", handler.AdminHandlerApp.MenuUpdate)
			strictAuthRouter.DELETE("/admin/menu", handler.AdminHandlerApp.MenuDelete)

			// 获取当前用户信息（兼容前端调用）
			strictAuthRouter.GET("/user", handler.AdminHandlerApp.GetAdminUser)

			strictAuthRouter.GET("/admin/users", handler.AdminHandlerApp.GetAdminUsers)
			strictAuthRouter.GET("/admin/user", handler.AdminHandlerApp.GetAdminUser)
			strictAuthRouter.PUT("/admin/user", handler.AdminHandlerApp.AdminUserUpdate)
			strictAuthRouter.POST("/admin/user", handler.AdminHandlerApp.AdminUserCreate)
			strictAuthRouter.DELETE("/admin/user", handler.AdminHandlerApp.AdminUserDelete)
			strictAuthRouter.GET("/admin/user/permissions", handler.AdminHandlerApp.GetUserPermissions)
			strictAuthRouter.GET("/admin/das/user-permissions", handler.AdminHandlerApp.GetDASUserPermissionsForAdmin)
			strictAuthRouter.GET("/admin/das/user-effective-permissions", handler.AdminHandlerApp.GetDASUserEffectivePermissionsForAdmin)
			strictAuthRouter.GET("/admin/role/permissions", handler.AdminHandlerApp.GetRolePermissions)
			strictAuthRouter.PUT("/admin/role/permission", handler.AdminHandlerApp.UpdateRolePermission)
			strictAuthRouter.GET("/admin/roles", handler.AdminHandlerApp.GetRoles)
			strictAuthRouter.POST("/admin/role", handler.AdminHandlerApp.RoleCreate)
			strictAuthRouter.PUT("/admin/role", handler.AdminHandlerApp.RoleUpdate)
			strictAuthRouter.DELETE("/admin/role", handler.AdminHandlerApp.RoleDelete)

			strictAuthRouter.GET("/admin/apis", handler.AdminHandlerApp.GetApis)
			strictAuthRouter.POST("/admin/api", handler.AdminHandlerApp.ApiCreate)
			strictAuthRouter.PUT("/admin/api", handler.AdminHandlerApp.ApiUpdate)
			strictAuthRouter.DELETE("/admin/api", handler.AdminHandlerApp.ApiDelete)
			strictAuthRouter.GET("/admin/api/sync", handler.AdminHandlerApp.SyncApi)
			strictAuthRouter.POST("/admin/api/sync", handler.AdminHandlerApp.EnterSyncApi)
			strictAuthRouter.POST("/admin/api/ai-fill", handler.AdminHandlerApp.ApiAiFill)
			strictAuthRouter.POST("/admin/api/ignore", handler.AdminHandlerApp.IgnoreApi)
			strictAuthRouter.GET("/admin/api/detail", handler.AdminHandlerApp.GetApiById)
			strictAuthRouter.DELETE("/admin/api/batch", handler.AdminHandlerApp.DeleteApisByIds)
			strictAuthRouter.GET("/admin/casbin/fresh", handler.AdminHandlerApp.FreshCasbin)

			// 部门管理
			strictAuthRouter.GET("/admin/departments", handler.DepartmentHandlerApp.GetDepartmentTree)
			strictAuthRouter.GET("/admin/departments/list", handler.DepartmentHandlerApp.GetDepartmentList)
			strictAuthRouter.GET("/admin/department", handler.DepartmentHandlerApp.GetDepartment)
			strictAuthRouter.POST("/admin/department", handler.DepartmentHandlerApp.CreateDepartment)
			strictAuthRouter.PUT("/admin/department", handler.DepartmentHandlerApp.UpdateDepartment)
			strictAuthRouter.DELETE("/admin/department", handler.DepartmentHandlerApp.DeleteDepartment)
			strictAuthRouter.GET("/admin/department/users", handler.DepartmentHandlerApp.GetDepartmentUsers)

			// 审批流程定义管理
			strictAuthRouter.GET("/admin/flows", handler.FlowHandlerApp.GetFlowDefinitionList)
			strictAuthRouter.GET("/admin/flow", handler.FlowHandlerApp.GetFlowDefinition)
			strictAuthRouter.POST("/admin/flow", handler.FlowHandlerApp.CreateFlowDefinition)
			strictAuthRouter.PUT("/admin/flow", handler.FlowHandlerApp.UpdateFlowDefinition)
			strictAuthRouter.DELETE("/admin/flow", handler.FlowHandlerApp.DeleteFlowDefinition)
			strictAuthRouter.PUT("/admin/flow/nodes", handler.FlowHandlerApp.SaveFlowNodes)

			// 审批流程实例和任务
			strictAuthRouter.POST("/flow/start", handler.FlowHandlerApp.StartFlow)
			strictAuthRouter.GET("/flow/instance", handler.FlowHandlerApp.GetFlowInstance)
			strictAuthRouter.GET("/flow/tasks/pending", handler.FlowHandlerApp.GetMyPendingTasks)
			strictAuthRouter.POST("/flow/task/approve", handler.FlowHandlerApp.ApproveTask)
			strictAuthRouter.POST("/flow/task/reject", handler.FlowHandlerApp.RejectTask)
		}
	}

	// 动态路由接口（soybean-admin格式）
	// 不需要认证的路由
	routeNoAuth := r.Group("/route")
	{
		routeNoAuth.GET("/getConstantRoutes", handler.AdminHandlerApp.GetConstantRoutes)
	}
	// 需要认证的路由
	routeAuth := r.Group("/route").Use(middleware.StrictAuth(jwt, logger))
	{
		routeAuth.GET("/getUserRoutes", handler.AdminHandlerApp.GetUserRoutes)
		routeAuth.GET("/isRouteExist", handler.AdminHandlerApp.IsRouteExist)
	}
}

// InitUserRouter 初始化用户相关路由
func InitUserRouter(r gin.IRouter, jwt *jwt.JWT, e *casbin.SyncedEnforcer, logger *log.Logger) {
	v1 := r.Group("/v1")
	{
		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger), middleware.AuthMiddleware(e))
		{
			// 用户管理路由
			strictAuthRouter.GET("/users", handler.UserHandlerApp.GetUsers)
			strictAuthRouter.GET("/users/:uid", handler.UserHandlerApp.GetUser)
			strictAuthRouter.POST("/users", handler.UserHandlerApp.CreateUser)
			strictAuthRouter.PUT("/users/:uid", handler.UserHandlerApp.UpdateUser)
			strictAuthRouter.DELETE("/users/:uid", handler.UserHandlerApp.DeleteUser)
			strictAuthRouter.PUT("/users/password", handler.UserHandlerApp.ChangePassword)
		}
	}
}
