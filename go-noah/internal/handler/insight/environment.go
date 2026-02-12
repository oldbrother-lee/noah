package insight

import (
	"go-noah/api"
	"go-noah/internal/model/insight"
	"go-noah/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// EnvironmentHandlerApp 全局 Handler 实例
var EnvironmentHandlerApp = new(EnvironmentHandler)

// EnvironmentHandler 环境管理Handler
type EnvironmentHandler struct{}

// GetEnvironments 获取环境列表
// @Summary 获取环境列表
// @Tags 环境管理
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /api/v1/insight/environments [get]
func (h *EnvironmentHandler) GetEnvironments(c *gin.Context) {
	envs, err := service.InsightServiceApp.GetEnvironments(c.Request.Context())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, envs)
}

// CreateEnvironmentRequest 创建环境请求
type CreateEnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateEnvironment 创建环境
// @Summary 创建环境
// @Tags 环境管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body CreateEnvironmentRequest true "环境信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/environments [post]
func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	env := &insight.DBEnvironment{
		Name: req.Name,
	}

	if err := service.InsightServiceApp.CreateEnvironment(c.Request.Context(), env); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, env)
}

// UpdateEnvironmentRequest 更新环境请求
type UpdateEnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateEnvironment 更新环境
// @Summary 更新环境
// @Tags 环境管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "环境ID"
// @Param request body UpdateEnvironmentRequest true "环境信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/environments/{id} [put]
func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	if err := service.InsightServiceApp.UpdateEnvironment(c.Request.Context(), uint(id), req.Name); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, nil)
}

// DeleteEnvironment 删除环境
// @Summary 删除环境
// @Tags 环境管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "环境ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/environments/{id} [delete]
func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteEnvironment(c.Request.Context(), uint(id)); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, nil)
}
