package handler

import (
	"go-noah/api"
	"go-noah/internal/service"
	"go-noah/pkg/jwt"
	"net/http"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/gin-gonic/gin"
)

// FlowHandlerApp 全局 Handler 实例
var FlowHandlerApp = new(FlowHandler)

type FlowHandler struct{}

func NewFlowHandler() *FlowHandler {
	return &FlowHandler{}
}

// ============= 流程定义 =============

// GetFlowDefinitionList 获取流程定义列表
// @Summary 获取流程定义列表
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flows [get]
func (h *FlowHandler) GetFlowDefinitionList(ctx *gin.Context) {
	data, err := service.FlowServiceApp.GetFlowDefinitionList(ctx)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetFlowDefinition 获取流程定义详情
// @Summary 获取流程定义详情
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query int true "流程定义ID"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flow [get]
func (h *FlowHandler) GetFlowDefinition(ctx *gin.Context) {
	var req api.GetFlowDefinitionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	data, err := service.FlowServiceApp.GetFlowDefinition(ctx, req.ID)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// CreateFlowDefinition 创建流程定义
// @Summary 创建流程定义
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.CreateFlowDefinitionRequest true "流程定义信息"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flow [post]
func (h *FlowHandler) CreateFlowDefinition(ctx *gin.Context) {
	var req api.CreateFlowDefinitionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.FlowServiceApp.CreateFlowDefinition(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// UpdateFlowDefinition 更新流程定义
// @Summary 更新流程定义
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.UpdateFlowDefinitionRequest true "流程定义信息"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flow [put]
func (h *FlowHandler) UpdateFlowDefinition(ctx *gin.Context) {
	var req api.UpdateFlowDefinitionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.FlowServiceApp.UpdateFlowDefinition(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// DeleteFlowDefinition 删除流程定义
// @Summary 删除流程定义
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query int true "流程定义ID"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flow [delete]
func (h *FlowHandler) DeleteFlowDefinition(ctx *gin.Context) {
	var req api.DeleteFlowDefinitionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.FlowServiceApp.DeleteFlowDefinition(ctx, req.ID); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// SaveFlowNodes 保存流程节点
// @Summary 保存流程节点
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.SaveFlowNodesRequest true "流程节点信息"
// @Success 200 {object} api.Response
// @Router /api/v1/admin/flow/nodes [put]
func (h *FlowHandler) SaveFlowNodes(ctx *gin.Context) {
	var req api.SaveFlowNodesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.FlowServiceApp.SaveFlowNodes(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// ============= 流程实例 =============

// StartFlow 发起流程
// @Summary 发起流程
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.StartFlowRequest true "发起流程信息"
// @Success 200 {object} api.Response
// @Router /api/v1/flow/start [post]
func (h *FlowHandler) StartFlow(ctx *gin.Context) {
	var req api.StartFlowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取当前用户信息
	claims, _ := ctx.Get("claims")
	if claims != nil {
		userClaims := claims.(*jwt.MyCustomClaims)
		req.InitiatorID = uint(userClaims.UserId)
		// Initiator 由前端传递或通过 Service 查询
		if req.Initiator == "" {
			req.Initiator = convertor.ToString(userClaims.UserId)
		}
	}

	data, err := service.FlowServiceApp.StartFlow(ctx, &req)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// GetFlowInstance 获取流程实例详情
// @Summary 获取流程实例详情
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query int true "流程实例ID"
// @Success 200 {object} api.Response
// @Router /api/v1/flow/instance [get]
func (h *FlowHandler) GetFlowInstance(ctx *gin.Context) {
	var req api.GetFlowInstanceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	data, err := service.FlowServiceApp.GetFlowInstanceDetail(ctx, req.ID)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// ============= 审批任务 =============

// GetMyPendingTasks 获取我的待办任务
// @Summary 获取我的待办任务
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} api.Response
// @Router /api/v1/flow/tasks/pending [get]
func (h *FlowHandler) GetMyPendingTasks(ctx *gin.Context) {
	var req api.GetMyPendingTasksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	claims, _ := ctx.Get("claims")
	userID := uint(claims.(*jwt.MyCustomClaims).UserId)

	data, err := service.FlowServiceApp.GetMyPendingTasks(ctx, userID, req.Page, req.PageSize)
	if err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, data)
}

// ApproveTask 审批通过
// @Summary 审批通过
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.ApproveTaskRequest true "审批信息"
// @Success 200 {object} api.Response
// @Router /api/v1/flow/task/approve [post]
func (h *FlowHandler) ApproveTask(ctx *gin.Context) {
	var req api.ApproveTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取当前用户信息
	claims, _ := ctx.Get("claims")
	if claims != nil {
		userClaims := claims.(*jwt.MyCustomClaims)
		req.OperatorID = uint(userClaims.UserId)
		if req.Operator == "" {
			req.Operator = convertor.ToString(userClaims.UserId)
		}
	}

	if err := service.FlowServiceApp.ApproveTask(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}

// RejectTask 审批驳回
// @Summary 审批驳回
// @Tags 审批流程
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body api.RejectTaskRequest true "驳回信息"
// @Success 200 {object} api.Response
// @Router /api/v1/flow/task/reject [post]
func (h *FlowHandler) RejectTask(ctx *gin.Context) {
	var req api.RejectTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.HandleError(ctx, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取当前用户信息
	claims, _ := ctx.Get("claims")
	if claims != nil {
		userClaims := claims.(*jwt.MyCustomClaims)
		req.OperatorID = uint(userClaims.UserId)
		if req.Operator == "" {
			req.Operator = convertor.ToString(userClaims.UserId)
		}
	}

	if err := service.FlowServiceApp.RejectTask(ctx, &req); err != nil {
		api.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(ctx, nil)
}
