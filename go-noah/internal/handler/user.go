package handler

import (
	"go-noah/api"
	"go-noah/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户 Handler
type UserHandler struct{}

// UserHandlerApp 全局 Handler 实例
var UserHandlerApp = new(UserHandler)

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int true "页码" default(1)
// @Param page_size query int true "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Param organization_key query string false "组织架构key"
// @Param role_id query int false "角色ID"
// @Success 200 {object} api.GetUsersResponse
// @Router /v1/users [get]
func (h *UserHandler) GetUsers(ctx *gin.Context) {
	var req api.GetUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	data, err := service.UserServiceApp.GetUsers(ctx.Request.Context(), &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, data)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详情
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param uid path int true "用户ID"
// @Success 200 {object} api.GetUserResponse
// @Router /v1/users/:uid [get]
func (h *UserHandler) GetUser(ctx *gin.Context) {
	uid, err := strconv.ParseUint(ctx.Param("uid"), 10, 64)
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	user, err := service.UserServiceApp.GetUser(ctx.Request.Context(), uid)
	if err != nil {
		api.HandleError(ctx, http.StatusNotFound, api.ErrNotFound, nil)
		return
	}

	lastLogin := ""
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.Format("2006-01-02 15:04:05")
	}
	dateJoined := ""
	if user.DateJoined != nil {
		dateJoined = user.DateJoined.Format("2006-01-02 15:04:05")
	}

	api.HandleSuccess(ctx, api.GetUserResponseData{
		UserData: api.UserData{
			Uid:         user.Uid,
			Username:    user.Username,
			Email:       user.Email,
			NickName:    user.NickName,
			Mobile:      user.Mobile,
			AvatarFile:  user.AvatarFile,
			RoleID:      user.RoleID,
			IsSuperuser: user.IsSuperuser,
			IsActive:    user.IsActive,
			IsStaff:     user.IsStaff,
			IsTwoFA:     user.IsTwoFA,
			LastLogin:   lastLogin,
			DateJoined:  dateJoined,
			UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body api.CreateUserRequest true "用户信息"
// @Success 200 {object} api.Response
// @Router /v1/users [post]
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req api.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.UserServiceApp.CreateUser(ctx.Request.Context(), &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, nil)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param uid path int true "用户ID"
// @Param user body api.UpdateUserRequest true "用户信息"
// @Success 200 {object} api.Response
// @Router /v1/users/:uid [put]
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	uid, err := strconv.ParseUint(ctx.Param("uid"), 10, 64)
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	var req api.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.UserServiceApp.UpdateUser(ctx.Request.Context(), uid, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param uid path int true "用户ID"
// @Success 200 {object} api.Response
// @Router /v1/users/:uid [delete]
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	uid, err := strconv.ParseUint(ctx.Param("uid"), 10, 64)
	if err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.UserServiceApp.DeleteUser(ctx.Request.Context(), uid); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body api.ChangePasswordRequest true "密码信息"
// @Success 200 {object} api.Response
// @Router /v1/users/password [put]
func (h *UserHandler) ChangePassword(ctx *gin.Context) {
	var req api.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.UserServiceApp.ChangePassword(ctx.Request.Context(), &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, api.ErrInternalServerError, nil)
		return
	}

	api.HandleSuccess(ctx, nil)
}
