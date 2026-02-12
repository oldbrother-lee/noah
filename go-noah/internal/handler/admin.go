package handler

import (
	"go-noah/api"
	"go-noah/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员 Handler
type AdminHandler struct{}

// AdminHandlerApp 全局 Handler 实例
var AdminHandlerApp = new(AdminHandler)

// Login godoc
// @Summary 账号登录
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body api.LoginRequest true "params"
// @Success 200 {object} api.LoginResponse
// @Router /v1/login [post]
func (h *AdminHandler) Login(ctx *gin.Context) {
	var req api.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	token, err := service.AdminServiceApp.Login(ctx, &req)
	if err != nil {
		// 直接使用 service 返回的错误，这样可以使用 ErrLoginFailed 显示友好的错误信息
		api.HandleError(ctx, http.StatusUnauthorized, err, nil)
		return
	}
	api.HandleSuccess(ctx, api.LoginResponseData{
		AccessToken: token,
	})
}

// GetMenus godoc
// @Summary 获取用户菜单
// @Schemes
// @Description 获取当前用户的菜单列表
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.GetMenuResponse
// @Router /v1/menus [get]
func (h *AdminHandler) GetMenus(ctx *gin.Context) {
	data, err := service.AdminServiceApp.GetMenus(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	// 过滤权限菜单
	api.HandleSuccess(ctx, data)
}

// GetAdminMenus godoc
// @Summary 获取管理员菜单
// @Schemes
// @Description 获取管理员菜单列表（Soybean-admin格式）
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.GetSoybeanMenuResponse
// @Router /v1/admin/menus [get]
func (h *AdminHandler) GetAdminMenus(ctx *gin.Context) {
	data, err := service.AdminServiceApp.GetAdminMenusSoybean(ctx)
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetUserRoutes godoc
// @Summary 获取用户动态路由
// @Schemes
// @Description 获取当前用户的动态路由（soybean-admin格式）
// @Tags 路由模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.GetUserRoutesResponse
// @Router /route/getUserRoutes [get]
func (h *AdminHandler) GetUserRoutes(ctx *gin.Context) {
	data, err := service.AdminServiceApp.GetUserRoutes(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetConstantRoutes godoc
// @Summary 获取常量路由
// @Schemes
// @Description 获取不需要权限的常量路由
// @Tags 路由模块
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /route/getConstantRoutes [get]
func (h *AdminHandler) GetConstantRoutes(ctx *gin.Context) {
	// 常量路由：登录页、404等不需要权限的路由
	constantRoutes := []map[string]interface{}{
		{
			"name":      "login",
			"path":      "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			"component": "layout.blank$view.login",
			"props":     true,
			"meta": map[string]interface{}{
				"title":      "login",
				"i18nKey":    "route.login",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"name":      "403",
			"path":      "/403",
			"component": "layout.blank$view.403",
			"meta": map[string]interface{}{
				"title":      "403",
				"i18nKey":    "route.403",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"name":      "404",
			"path":      "/404",
			"component": "layout.blank$view.404",
			"meta": map[string]interface{}{
				"title":      "404",
				"i18nKey":    "route.404",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"name":      "500",
			"path":      "/500",
			"component": "layout.blank$view.500",
			"meta": map[string]interface{}{
				"title":      "500",
				"i18nKey":    "route.500",
				"constant":   true,
				"hideInMenu": true,
			},
		},
	}
	api.HandleSuccess(ctx, constantRoutes)
}

// IsRouteExist godoc
// @Summary 检查路由是否存在
// @Schemes
// @Description 检查指定路由名称是否存在
// @Tags 路由模块
// @Accept json
// @Produce json
// @Param routeName query string true "路由名称"
// @Security Bearer
// @Success 200 {object} api.Response
// @Router /route/isRouteExist [get]
func (h *AdminHandler) IsRouteExist(ctx *gin.Context) {
	routeName := ctx.Query("routeName")
	// 简单实现：总是返回 true，表示路由存在
	// 实际应用中可以根据需求查询数据库
	api.HandleSuccess(ctx, routeName != "")
}

// GetUserPermissions godoc
// @Summary 获取用户权限
// @Schemes
// @Description 获取当前用户的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.GetUserPermissionsData
// @Router /v1/admin/user/permissions [get]
func (h *AdminHandler) GetUserPermissions(ctx *gin.Context) {
	data, err := service.AdminServiceApp.GetUserPermissions(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	// 过滤权限菜单
	api.HandleSuccess(ctx, data)
}

// GetRolePermissions godoc
// @Summary 获取角色权限
// @Schemes
// @Description 获取指定角色的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param role query string true "角色名称"
// @Success 200 {object} api.GetRolePermissionsData
// @Router /v1/admin/role/permissions [get]
func (h *AdminHandler) GetRolePermissions(ctx *gin.Context) {
	var req api.GetRolePermissionsRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.GetRolePermissions(ctx, req.Role)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// UpdateRolePermission godoc
// @Summary 更新角色权限
// @Schemes
// @Description 更新指定角色的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.UpdateRolePermissionRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/role/permissions [put]
func (h *AdminHandler) UpdateRolePermission(ctx *gin.Context) {
	var req api.UpdateRolePermissionRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	err := service.AdminServiceApp.UpdateRolePermission(ctx, &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// MenuUpdate godoc
// @Summary 更新菜单
// @Schemes
// @Description 更新菜单信息
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.MenuUpdateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/menu [put]
func (h *AdminHandler) MenuUpdate(ctx *gin.Context) {
	var req api.MenuUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.MenuUpdate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// MenuCreate godoc
// @Summary 创建菜单
// @Schemes
// @Description 创建新的菜单
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.MenuCreateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/menu [post]
func (h *AdminHandler) MenuCreate(ctx *gin.Context) {
	var req api.MenuCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.MenuCreate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// MenuDelete godoc
// @Summary 删除菜单
// @Schemes
// @Description 删除指定菜单
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "菜单ID"
// @Success 200 {object} api.Response
// @Router /v1/admin/menu [delete]
func (h *AdminHandler) MenuDelete(ctx *gin.Context) {
	var req api.MenuDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.MenuDelete(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return

	}
	api.HandleSuccess(ctx, nil)
}

// GetRoles godoc
// @Summary 获取角色列表
// @Schemes
// @Description 获取角色列表
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param sid query string false "角色ID"
// @Param name query string false "角色名称"
// @Success 200 {object} api.GetRolesResponse
// @Router /v1/admin/roles [get]
func (h *AdminHandler) GetRoles(ctx *gin.Context) {
	var req api.GetRoleListRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.GetRoles(ctx, &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, data)
}

// RoleCreate godoc
// @Summary 创建角色
// @Schemes
// @Description 创建新的角色
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.RoleCreateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/role [post]
func (h *AdminHandler) RoleCreate(ctx *gin.Context) {
	var req api.RoleCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.RoleCreate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// RoleUpdate godoc
// @Summary 更新角色
// @Schemes
// @Description 更新角色信息
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.RoleUpdateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/role [put]
func (h *AdminHandler) RoleUpdate(ctx *gin.Context) {
	var req api.RoleUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.RoleUpdate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// RoleDelete godoc
// @Summary 删除角色
// @Schemes
// @Description 删除指定角色
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "角色ID"
// @Success 200 {object} api.Response
// @Router /v1/admin/role [delete]
func (h *AdminHandler) RoleDelete(ctx *gin.Context) {
	var req api.RoleDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.RoleDelete(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// GetApis godoc
// @Summary 获取API列表
// @Schemes
// @Description 获取API列表
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param group query string false "API分组"
// @Param name query string false "API名称"
// @Param path query string false "API路径"
// @Param method query string false "请求方法"
// @Success 200 {object} api.GetApisResponse
// @Router /v1/admin/apis [get]
func (h *AdminHandler) GetApis(ctx *gin.Context) {
	var req api.GetApisRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.GetApis(ctx, &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, data)
}

// ApiCreate godoc
// @Summary 创建API
// @Schemes
// @Description 创建新的API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.ApiCreateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/api [post]
func (h *AdminHandler) ApiCreate(ctx *gin.Context) {
	var req api.ApiCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.ApiCreate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// ApiUpdate godoc
// @Summary 更新API
// @Schemes
// @Description 更新API信息
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.ApiUpdateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/api [put]
func (h *AdminHandler) ApiUpdate(ctx *gin.Context) {
	var req api.ApiUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.ApiUpdate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// ApiDelete godoc
// @Summary 删除API
// @Schemes
// @Description 删除指定API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "API ID"
// @Success 200 {object} api.Response
// @Router /v1/admin/api [delete]
func (h *AdminHandler) ApiDelete(ctx *gin.Context) {
	var req api.ApiDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.ApiDelete(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// SyncApi godoc
// @Summary 同步API（对比代码路由与数据库，返回新增/待删/忽略列表）
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.SyncApiResponse
// @Router /v1/admin/api/sync [get]
func (h *AdminHandler) SyncApi(ctx *gin.Context) {
	data, err := service.AdminServiceApp.SyncApi(ctx.Request.Context())
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// EnterSyncApi godoc
// @Summary 确认同步API（将对比结果写入/删除数据库并清理 Casbin）
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.EnterSyncApiRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/api/sync [post]
func (h *AdminHandler) EnterSyncApi(ctx *gin.Context) {
	var req api.EnterSyncApiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.EnterSyncApi(ctx.Request.Context(), &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// ApiAiFill godoc
// @Summary AI 自动填充 API 名称与分组
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.ApiAiFillRequest true "path+method 列表"
// @Success 200 {array} api.ApiAiFillItem
// @Router /v1/admin/api/ai-fill [post]
func (h *AdminHandler) ApiAiFill(ctx *gin.Context) {
	var req api.ApiAiFillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.ApiAiFill(ctx.Request.Context(), &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// IgnoreApi godoc
// @Summary 忽略/取消忽略API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.IgnoreApiRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/api/ignore [post]
func (h *AdminHandler) IgnoreApi(ctx *gin.Context) {
	var req api.IgnoreApiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.IgnoreApi(ctx.Request.Context(), &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// GetApiById godoc
// @Summary 按ID获取单条API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "API ID"
// @Success 200 {object} api.ApiDataItem
// @Router /v1/admin/api/detail [get]
func (h *AdminHandler) GetApiById(ctx *gin.Context) {
	var req api.GetApiByIdRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.GetApiById(ctx.Request.Context(), req.ID)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// DeleteApisByIds godoc
// @Summary 批量删除API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.DeleteApisByIdsRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/api/batch [delete]
func (h *AdminHandler) DeleteApisByIds(ctx *gin.Context) {
	var req api.DeleteApisByIdsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.DeleteApisByIds(ctx.Request.Context(), req.IDs); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// FreshCasbin godoc
// @Summary 刷新Casbin策略（从数据库重新加载）
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.Response
// @Router /v1/admin/casbin/fresh [get]
func (h *AdminHandler) FreshCasbin(ctx *gin.Context) {
	if err := service.AdminServiceApp.FreshCasbin(ctx.Request.Context()); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// AdminUserUpdate godoc
// @Summary 更新管理员用户
// @Schemes
// @Description 更新管理员用户信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.AdminUserUpdateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/user [put]
func (h *AdminHandler) AdminUserUpdate(ctx *gin.Context) {
	var req api.AdminUserUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.AdminUserUpdate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// AdminUserCreate godoc
// @Summary 创建管理员用户
// @Schemes
// @Description 创建新的管理员用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.AdminUserCreateRequest true "参数"
// @Success 200 {object} api.Response
// @Router /v1/admin/user [post]
func (h *AdminHandler) AdminUserCreate(ctx *gin.Context) {
	var req api.AdminUserCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.AdminUserCreate(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// AdminUserDelete godoc
// @Summary 删除管理员用户
// @Schemes
// @Description 删除指定管理员用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "用户ID"
// @Success 200 {object} api.Response
// @Router /v1/admin/user [delete]
func (h *AdminHandler) AdminUserDelete(ctx *gin.Context) {
	var req api.AdminUserDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	if err := service.AdminServiceApp.AdminUserDelete(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return

	}
	api.HandleSuccess(ctx, nil)
}

// GetAdminUsers godoc
// @Summary 获取管理员用户列表
// @Schemes
// @Description 获取管理员用户列表
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param username query string false "用户名"
// @Param nickname query string false "昵称"
// @Param phone query string false "手机号"
// @Param email query string false "邮箱"
// @Success 200 {object} api.GetAdminUsersResponse
// @Router /v1/admin/users [get]
func (h *AdminHandler) GetAdminUsers(ctx *gin.Context) {
	var req api.GetAdminUsersRequest
	if err := ctx.ShouldBind(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}
	data, err := service.AdminServiceApp.GetAdminUsers(ctx, &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, data)
}

// GetAdminUser godoc
// @Summary 获取管理用户信息
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.GetAdminUserResponse
// @Router /v1/admin/user [get]
func (h *AdminHandler) GetAdminUser(ctx *gin.Context) {
	data, err := service.AdminServiceApp.GetAdminUser(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, data)
}
