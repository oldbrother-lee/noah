package router

import (
	"go-noah/internal/handler/insight"
	"go-noah/internal/middleware"
	"go-noah/pkg/jwt"
	"go-noah/pkg/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// InitInsightRouter 初始化 goInsight 相关路由
func InitInsightRouter(r gin.IRouter, jwt *jwt.JWT, e *casbin.SyncedEnforcer, logger *log.Logger) {
	v1 := r.Group("/v1")
	{
		// 需要认证和权限的路由
		authRouter := v1.Group("/insight").Use(middleware.StrictAuth(jwt, logger), middleware.AuthMiddleware(e))
		{
			// ============ 环境管理 ============
			authRouter.GET("/environments", insight.EnvironmentHandlerApp.GetEnvironments)
			authRouter.POST("/environments", insight.EnvironmentHandlerApp.CreateEnvironment)
			authRouter.PUT("/environments/:id", insight.EnvironmentHandlerApp.UpdateEnvironment)
			authRouter.DELETE("/environments/:id", insight.EnvironmentHandlerApp.DeleteEnvironment)

			// ============ 数据库配置管理 ============
			authRouter.GET("/dbconfigs", insight.DBConfigHandlerApp.GetDBConfigs)
			authRouter.GET("/dbconfigs/:instance_id", insight.DBConfigHandlerApp.GetDBConfig)
			authRouter.POST("/dbconfigs", insight.DBConfigHandlerApp.CreateDBConfig)
			authRouter.PUT("/dbconfigs/:id", insight.DBConfigHandlerApp.UpdateDBConfig)
			authRouter.DELETE("/dbconfigs/:id", insight.DBConfigHandlerApp.DeleteDBConfig)
			authRouter.GET("/dbconfigs/:instance_id/schemas", insight.DBConfigHandlerApp.GetSchemas)

			// ============ DAS 数据查询 ============
			authRouter.POST("/das/query", insight.DASHandlerApp.ExecuteQuery)
			authRouter.GET("/das/schemas", insight.DASHandlerApp.GetUserSchemas)          // 获取用户授权的所有schemas
			authRouter.GET("/das/schemas/:instance_id", insight.DASHandlerApp.GetSchemas) // 获取指定实例的schemas
			authRouter.GET("/das/tables/:instance_id/:schema", insight.DASHandlerApp.GetTables)
			authRouter.GET("/das/columns/:instance_id/:schema/:table", insight.DASHandlerApp.GetTableColumns)

			// ============ DAS 执行记录 ============
			authRouter.GET("/das/records", insight.DASHandlerApp.GetRecords)

			// ============ DAS 收藏夹 ============
			authRouter.GET("/das/favorites", insight.DASHandlerApp.GetFavorites)
			authRouter.POST("/das/favorites", insight.DASHandlerApp.CreateFavorite)
			authRouter.DELETE("/das/favorites/:id", insight.DASHandlerApp.DeleteFavorite)

			// ============ DAS 权限管理 ============
			authRouter.GET("/das/permissions", insight.DASHandlerApp.GetUserPermissions)
			authRouter.POST("/das/permissions/schema", insight.DASHandlerApp.GrantSchemaPermission)
			authRouter.DELETE("/das/permissions/schema/:id", insight.DASHandlerApp.RevokeSchemaPermission)
			authRouter.POST("/das/permissions/table", insight.DASHandlerApp.GrantTablePermission)
			authRouter.DELETE("/das/permissions/table/:id", insight.DASHandlerApp.RevokeTablePermission)

			// ============ 权限模板管理 ============
			authRouter.GET("/das/permissions/templates", insight.DASHandlerApp.GetPermissionTemplates)
			authRouter.GET("/das/permissions/templates/:id", insight.DASHandlerApp.GetPermissionTemplate)
			authRouter.POST("/das/permissions/templates", insight.DASHandlerApp.CreatePermissionTemplate)
			authRouter.PUT("/das/permissions/templates/:id", insight.DASHandlerApp.UpdatePermissionTemplate)
			authRouter.DELETE("/das/permissions/templates/:id", insight.DASHandlerApp.DeletePermissionTemplate)

			// ============ 角色权限管理 ============
			authRouter.GET("/das/permissions/roles/:role", insight.DASHandlerApp.GetRolePermissions)
			authRouter.POST("/das/permissions/roles", insight.DASHandlerApp.CreateRolePermission)
			authRouter.DELETE("/das/permissions/roles/:id", insight.DASHandlerApp.DeleteRolePermission)

			// ============ 用户权限管理（与角色同构：object/template，无 rule）============
			authRouter.GET("/das/permissions/by-user", insight.DASHandlerApp.GetUserPermissionList)
			authRouter.POST("/das/permissions/user", insight.DASHandlerApp.CreateUserPermission)
			authRouter.DELETE("/das/permissions/user/:id", insight.DASHandlerApp.DeleteUserPermission)

			// ============ 用户权限查询 ============
			authRouter.GET("/das/permissions/users", insight.DASHandlerApp.GetUserEffectivePermissions)

			// ============ 组织管理 ============
			authRouter.GET("/organizations", insight.OrganizationHandlerApp.GetOrganizations)
			authRouter.GET("/organizations/tree", insight.OrganizationHandlerApp.GetOrganizationTree)
			authRouter.POST("/organizations", insight.OrganizationHandlerApp.CreateOrganization)
			authRouter.PUT("/organizations", insight.OrganizationHandlerApp.UpdateOrganization)
			authRouter.DELETE("/organizations/:id", insight.OrganizationHandlerApp.DeleteOrganization)
			authRouter.GET("/organizations/users", insight.OrganizationHandlerApp.GetOrganizationUsers)
			authRouter.POST("/organizations/users", insight.OrganizationHandlerApp.BindUser)
			authRouter.DELETE("/organizations/users/:uid", insight.OrganizationHandlerApp.UnbindUser)

			// ============ 工单管理 ============
			authRouter.GET("/orders", insight.OrderHandlerApp.GetOrders)
			authRouter.GET("/orders/my", insight.OrderHandlerApp.GetMyOrders)
			authRouter.GET("/orders/:order_id", insight.OrderHandlerApp.GetOrder)
			authRouter.POST("/orders", insight.OrderHandlerApp.CreateOrder)
			authRouter.POST("/orders/check-ddl", insight.OrderHandlerApp.CheckDDL)                       // DDL 预检：检查表是否有主键/唯一键（语法检查时调用）
			authRouter.GET("/orders/tables/:instance_id/:schema", insight.OrderHandlerApp.GetOrderTables) // 工单场景获取表列表（不检查DAS权限）
			authRouter.PUT("/orders/progress", insight.OrderHandlerApp.UpdateOrderProgress)
			authRouter.POST("/orders/approve", insight.OrderHandlerApp.ApproveOrder) // 审批工单
			authRouter.GET("/orders/:order_id/tasks", insight.OrderHandlerApp.GetOrderTasks)
			authRouter.GET("/orders/:order_id/tasks/:task_id/rollback-sql", insight.OrderHandlerApp.GetTaskRollbackSQL)
			authRouter.PUT("/orders/tasks/progress", insight.OrderHandlerApp.UpdateTaskProgress)
			authRouter.POST("/orders/tasks/execute", insight.OrderHandlerApp.ExecuteTask)
			authRouter.POST("/orders/ghost/control", insight.OrderHandlerApp.ControlGhost) // gh-ost 控制（暂停/取消/速度调节）
			authRouter.GET("/orders/:order_id/logs", insight.OrderHandlerApp.GetOrderLogs)
			authRouter.GET("/orders/:order_id/ghost-progress", insight.OrderHandlerApp.GetGhostProgress) // 获取 gh-ost 最新进度（从 Redis）

			// ============ SQL审核 ============
			authRouter.POST("/inspect/sql", insight.InspectHandlerApp.InspectSQL)
			authRouter.GET("/inspect/params", insight.InspectHandlerApp.GetInspectParams)
			authRouter.GET("/inspect/params/default", insight.InspectHandlerApp.GetDefaultInspectParams)
			authRouter.GET("/inspect/params/:id", insight.InspectHandlerApp.GetInspectParam)
			authRouter.POST("/inspect/params", insight.InspectHandlerApp.CreateInspectParams)
			authRouter.PUT("/inspect/params/:id", insight.InspectHandlerApp.UpdateInspectParams)
			authRouter.DELETE("/inspect/params/:id", insight.InspectHandlerApp.DeleteInspectParams)
		}
	}
}
