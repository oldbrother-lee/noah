package handler

import (
	"go-noah/api"
	"go-noah/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DepartmentHandlerApp 全局 Handler 实例
var DepartmentHandlerApp = new(DepartmentHandler)

type DepartmentHandler struct{}

func NewDepartmentHandler() *DepartmentHandler {
	return &DepartmentHandler{}
}

// GetDepartmentTree 获取部门树
// @Summary 获取部门树
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.Response
// @Router /api/v1/admin/departments [get]
func (h *DepartmentHandler) GetDepartmentTree(ctx *gin.Context) {
	data, err := service.DepartmentServiceApp.GetDepartmentTree(ctx)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetDepartmentList 获取部门列表（扁平）
// @Summary 获取部门列表
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.Response
// @Router /api/v1/admin/departments/list [get]
func (h *DepartmentHandler) GetDepartmentList(ctx *gin.Context) {
	data, err := service.DepartmentServiceApp.GetDepartmentList(ctx)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetDepartment 获取部门详情
// @Summary 获取部门详情
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query int true "部门ID"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/department [get]
func (h *DepartmentHandler) GetDepartment(ctx *gin.Context) {
	var req api.GetDepartmentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	data, err := service.DepartmentServiceApp.GetDepartment(ctx, req.ID)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// CreateDepartment 创建部门
// @Summary 创建部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.CreateDepartmentRequest true "部门信息"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/department [post]
func (h *DepartmentHandler) CreateDepartment(ctx *gin.Context) {
	var req api.CreateDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.DepartmentServiceApp.CreateDepartment(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// UpdateDepartment 更新部门
// @Summary 更新部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.UpdateDepartmentRequest true "部门信息"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/department [put]
func (h *DepartmentHandler) UpdateDepartment(ctx *gin.Context) {
	var req api.UpdateDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.DepartmentServiceApp.UpdateDepartment(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// DeleteDepartment 删除部门
// @Summary 删除部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query int true "部门ID"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/department [delete]
func (h *DepartmentHandler) DeleteDepartment(ctx *gin.Context) {
	var req api.DeleteDepartmentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.DepartmentServiceApp.DeleteDepartment(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// GetDepartmentUsers 获取部门用户
// @Summary 获取部门用户
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param deptId query int true "部门ID"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/department/users [get]
func (h *DepartmentHandler) GetDepartmentUsers(ctx *gin.Context) {
	var req api.GetDepartmentUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	data, err := service.DepartmentServiceApp.GetDepartmentUsers(ctx, req.DeptID)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

