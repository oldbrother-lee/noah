package insight

import (
	"context"
	"encoding/json"
	"fmt"
	"go-noah/api"
	"go-noah/internal/das/dao"
	"go-noah/internal/handler"
	"go-noah/internal/inspect/parser"
	"go-noah/internal/model"
	"go-noah/internal/model/insight"
	"go-noah/internal/orders/executor"
	"go-noah/internal/repository"
	insightRepo "go-noah/internal/repository/insight"
	"go-noah/internal/service"
	"go-noah/pkg/global"
	"go-noah/pkg/notifier"
	"go-noah/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrderHandlerApp 全局 Handler 实例
var OrderHandlerApp = new(OrderHandler)

// OrderHandler 工单管理 Handler
type OrderHandler struct{}

// checkOrderAccess 检查用户是否有权限访问工单
// 返回 (是否有权限, 错误)
func (h *OrderHandler) checkOrderAccess(c *gin.Context, orderID string) (bool, error) {
	// 获取工单信息
	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		return false, err
	}

	// 如果工单不限制访问，所有人都可以查看
	if !order.IsRestrictAccess {
		return true, nil
	}

	// 获取当前用户信息
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	if username == "" {
		return false, nil
	}

	// 获取流程实例信息（用于获取流程引擎中的审核人和执行人）
	var flowInstance *api.FlowInstanceDetail
	if order.FlowInstanceID > 0 {
		flowInstance, _ = service.FlowServiceApp.GetFlowInstanceDetail(c.Request.Context(), order.FlowInstanceID)
	}

	// 首先检查用户是否是审核人或执行人（审核人和执行人始终有权限，不受限制访问控制）
	// 从流程实例中检查审核人和执行人（检查所有状态的任务，包括已审批的）
	if flowInstance != nil && flowInstance.Tasks != nil {
		for _, task := range flowInstance.Tasks {
			// 检查是否是审核人（nodeCode 包含 "approval" 或 nodeName 包含 "审批"）
			if (strings.Contains(task.NodeCode, "approval") || strings.Contains(task.NodeName, "审批")) && task.Assignee == username {
				global.Logger.Debug("用户是审核人，允许访问",
					zap.String("order_id", orderID),
					zap.String("username", username),
					zap.String("node_code", task.NodeCode),
					zap.String("task_status", task.Status),
				)
				return true, nil // 审核人始终有权限
			}
			// 检查是否是执行人（nodeCode 包含 "execute" 或 nodeName 包含 "执行"）
			if (strings.Contains(task.NodeCode, "execute") || strings.Contains(task.NodeName, "执行")) && task.Assignee == username {
				global.Logger.Debug("用户是执行人，允许访问",
					zap.String("order_id", orderID),
					zap.String("username", username),
					zap.String("node_code", task.NodeCode),
					zap.String("task_status", task.Status),
				)
				return true, nil // 执行人始终有权限
			}
		}
	}

	// 如果流程实例存在但任务列表为空，尝试从流程节点定义中获取审核人和执行人
	if flowInstance != nil && len(flowInstance.Tasks) == 0 {
		// 从流程节点定义中获取审核人和执行人（用于任务还未创建的情况）
		// 需要先获取流程实例的 FlowDefID
		baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
		flowRepo := repository.NewFlowRepository(baseRepo)
		instance, err := flowRepo.GetFlowInstance(c.Request.Context(), order.FlowInstanceID)
		if err == nil && instance != nil {
			nodes, _ := flowRepo.GetFlowNodes(c.Request.Context(), instance.FlowDefID)
			for _, node := range nodes {
				// 检查审核节点
				if (strings.Contains(node.NodeCode, "approval") || strings.Contains(node.NodeName, "审批")) && node.ApproverType != "" {
					// 根据审批人类型获取用户列表
					switch node.ApproverType {
					case "role":
						roles := strings.Split(node.ApproverIDs, ",")
						for _, role := range roles {
							role = strings.TrimSpace(role)
							if role == "" {
								continue
							}
							userIDs, _ := global.Enforcer.GetUsersForRole(role)
							for _, uidStr := range userIDs {
								uidInt, err := strconv.ParseUint(uidStr, 10, 64)
								if err != nil {
									continue
								}
								if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
									if adminUser.Username == username {
										global.Logger.Debug("用户是审核人（从节点定义获取），允许访问",
											zap.String("order_id", orderID),
											zap.String("username", username),
											zap.String("node_code", node.NodeCode),
										)
										return true, nil
									}
								}
							}
						}
					case "user":
						userIDs := strings.Split(node.ApproverIDs, ",")
						for _, uidStr := range userIDs {
							uidStr = strings.TrimSpace(uidStr)
							if uidStr == "" {
								continue
							}
							uidInt, err := strconv.ParseUint(uidStr, 10, 64)
							if err != nil {
								continue
							}
							if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
								if adminUser.Username == username {
									global.Logger.Debug("用户是审核人（从节点定义获取），允许访问",
										zap.String("order_id", orderID),
										zap.String("username", username),
										zap.String("node_code", node.NodeCode),
									)
									return true, nil
								}
							}
						}
					}
				}
				// 检查执行节点
				if (strings.Contains(node.NodeCode, "execute") || strings.Contains(node.NodeName, "执行")) && node.ApproverType != "" {
					// 根据审批人类型获取用户列表
					switch node.ApproverType {
					case "role":
						roles := strings.Split(node.ApproverIDs, ",")
						for _, role := range roles {
							role = strings.TrimSpace(role)
							if role == "" {
								continue
							}
							userIDs, _ := global.Enforcer.GetUsersForRole(role)
							for _, uidStr := range userIDs {
								uidInt, err := strconv.ParseUint(uidStr, 10, 64)
								if err != nil {
									continue
								}
								if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
									if adminUser.Username == username {
										global.Logger.Debug("用户是执行人（从节点定义获取），允许访问",
											zap.String("order_id", orderID),
											zap.String("username", username),
											zap.String("node_code", node.NodeCode),
										)
										return true, nil
									}
								}
							}
						}
					case "user":
						userIDs := strings.Split(node.ApproverIDs, ",")
						for _, uidStr := range userIDs {
							uidStr = strings.TrimSpace(uidStr)
							if uidStr == "" {
								continue
							}
							uidInt, err := strconv.ParseUint(uidStr, 10, 64)
							if err != nil {
								continue
							}
							if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
								if adminUser.Username == username {
									global.Logger.Debug("用户是执行人（从节点定义获取），允许访问",
										zap.String("order_id", orderID),
										zap.String("username", username),
										zap.String("node_code", node.NodeCode),
									)
									return true, nil
								}
							}
						}
					}
				}
			}
		}
	}

	// 从工单字段中检查审核人和执行人
	// 解析审批人 JSON
	if len(order.Approver) > 0 {
		var approvers []map[string]interface{}
		if err := json.Unmarshal(order.Approver, &approvers); err == nil {
			for _, approver := range approvers {
				if user, ok := approver["user"].(string); ok && user == username {
					return true, nil // 审核人始终有权限
				}
			}
		}
	}

	// 解析执行人 JSON
	if len(order.Executor) > 0 {
		var executors []string
		if err := json.Unmarshal(order.Executor, &executors); err == nil {
			for _, executor := range executors {
				if executor == username {
					return true, nil // 执行人始终有权限
				}
			}
		} else {
			// 如果解析失败，尝试解析为对象数组
			var executorObjs []map[string]interface{}
			if err := json.Unmarshal(order.Executor, &executorObjs); err == nil {
				for _, executor := range executorObjs {
					if user, ok := executor["user"].(string); ok && user == username {
						return true, nil // 执行人始终有权限
					}
				}
			}
		}
	}

	// 检查是否是其他相关人员（申请人、复核人、抄送人）
	var relatedUsers []string

	// 添加申请人
	if order.Applicant != "" {
		relatedUsers = append(relatedUsers, order.Applicant)
	}

	// 解析复核人 JSON
	if len(order.Reviewer) > 0 {
		var reviewers []map[string]interface{}
		if err := json.Unmarshal(order.Reviewer, &reviewers); err == nil {
			for _, reviewer := range reviewers {
				if user, ok := reviewer["user"].(string); ok && user != "" {
					relatedUsers = append(relatedUsers, user)
				}
			}
		}
	}

	// 解析抄送人 JSON
	if len(order.CC) > 0 {
		var ccs []map[string]interface{}
		if err := json.Unmarshal(order.CC, &ccs); err == nil {
			for _, cc := range ccs {
				if user, ok := cc["user"].(string); ok && user != "" {
					relatedUsers = append(relatedUsers, user)
				}
			}
		}
	}

	// 检查当前用户是否在其他相关人员列表中
	for _, user := range relatedUsers {
		if user == username {
			return true, nil
		}
	}

	return false, nil
}

// GetOrdersRequest 获取工单列表请求
type GetOrdersRequest struct {
	Page        int    `form:"current"` // 前端用 current
	PageSize    int    `form:"size"`    // 前端用 size
	OnlyMyOrder int    `form:"only_my_orders"`
	Applicant   string `form:"applicant"`
	Progress    string `form:"progress"`
	Environment int    `form:"environment"`
	SQLType     string `form:"sql_type"`
	DBType      string `form:"db_type"`
	Title       string `form:"title"`
}

// GetOrders 获取工单列表
// @Summary 获取工单列表
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Param applicant query string false "申请人"
// @Param progress query string false "进度"
// @Param environment query int false "环境"
// @Param sql_type query string false "SQL类型"
// @Param db_type query string false "DB类型"
// @Param title query string false "标题"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	var req GetOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 设置默认分页值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 如果 only_my_orders=1，获取当前用户名
	var applicant string
	if req.OnlyMyOrder == 1 {
		userId := handler.GetUserIdFromCtx(c)
		if userId > 0 {
			user, err := service.AdminServiceApp.GetAdminUser(c.Request.Context(), userId)
			if err == nil && user != nil {
				applicant = user.Username
			}
		}
	} else {
		applicant = req.Applicant
	}

	params := &insightRepo.OrderQueryParams{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Applicant:   applicant,
		Progress:    req.Progress,
		Environment: req.Environment,
		SQLType:     req.SQLType,
		DBType:      req.DBType,
		Title:       req.Title,
	}

	orders, total, err := service.InsightServiceApp.GetOrders(c.Request.Context(), params)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, gin.H{
		"list":  orders,
		"total": total,
	})
}

// GetOrder 获取工单详情
// @Summary 获取工单详情
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param order_id path string true "工单ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/{order_id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 获取当前用户信息
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 获取流程实例信息（如果有，用于权限检查）
	var flowInstance *api.FlowInstanceDetail
	if order.FlowInstanceID > 0 {
		flowInstance, _ = service.FlowServiceApp.GetFlowInstanceDetail(c.Request.Context(), order.FlowInstanceID)
	}

	// 限制访问权限检查
	isRelatedUser := true // 默认允许访问
	if order.IsRestrictAccess && username != "" {
		// 首先检查用户是否是审核人或执行人（审核人和执行人始终有权限，不受限制访问控制）
		isApproverOrExecutor := false

		// 从流程实例中检查审核人和执行人（检查所有状态的任务，包括已审批的）
		if flowInstance != nil && flowInstance.Tasks != nil {
			for _, task := range flowInstance.Tasks {
				// 检查是否是审核人（nodeCode 包含 "approval" 或 nodeName 包含 "审批"）
				if (strings.Contains(task.NodeCode, "approval") || strings.Contains(task.NodeName, "审批")) && task.Assignee == username {
					isApproverOrExecutor = true
					break
				}
				// 检查是否是执行人（nodeCode 包含 "execute" 或 nodeName 包含 "执行"）
				if (strings.Contains(task.NodeCode, "execute") || strings.Contains(task.NodeName, "执行")) && task.Assignee == username {
					isApproverOrExecutor = true
					break
				}
			}
		}

		// 从工单字段中检查审核人和执行人
		if !isApproverOrExecutor {
			// 解析审批人 JSON
			if len(order.Approver) > 0 {
				var approvers []map[string]interface{}
				if err := json.Unmarshal(order.Approver, &approvers); err == nil {
					for _, approver := range approvers {
						if user, ok := approver["user"].(string); ok && user == username {
							isApproverOrExecutor = true
							break
						}
					}
				}
			}

			// 解析执行人 JSON
			if !isApproverOrExecutor && len(order.Executor) > 0 {
				var executors []string
				if err := json.Unmarshal(order.Executor, &executors); err == nil {
					for _, executor := range executors {
						if executor == username {
							isApproverOrExecutor = true
							break
						}
					}
				} else {
					// 如果解析失败，尝试解析为对象数组
					var executorObjs []map[string]interface{}
					if err := json.Unmarshal(order.Executor, &executorObjs); err == nil {
						for _, executor := range executorObjs {
							if user, ok := executor["user"].(string); ok && user == username {
								isApproverOrExecutor = true
								break
							}
						}
					}
				}
			}
		}

		// 如果流程实例存在但任务列表为空，尝试从流程节点定义中获取审核人和执行人
		if !isApproverOrExecutor && flowInstance != nil && len(flowInstance.Tasks) == 0 {
			// 从流程节点定义中获取审核人和执行人（用于任务还未创建的情况）
			baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
			flowRepo := repository.NewFlowRepository(baseRepo)
			instance, err := flowRepo.GetFlowInstance(c.Request.Context(), order.FlowInstanceID)
			if err == nil && instance != nil {
				nodes, _ := flowRepo.GetFlowNodes(c.Request.Context(), instance.FlowDefID)
				for _, node := range nodes {
					// 检查审核节点
					if (strings.Contains(node.NodeCode, "approval") || strings.Contains(node.NodeName, "审批")) && node.ApproverType != "" {
						// 根据审批人类型获取用户列表
						switch node.ApproverType {
						case "role":
							roles := strings.Split(node.ApproverIDs, ",")
							for _, role := range roles {
								role = strings.TrimSpace(role)
								if role == "" {
									continue
								}
								userIDs, _ := global.Enforcer.GetUsersForRole(role)
								for _, uidStr := range userIDs {
									uidInt, err := strconv.ParseUint(uidStr, 10, 64)
									if err != nil {
										continue
									}
									if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
										if adminUser.Username == username {
											isApproverOrExecutor = true
											global.Logger.Debug("用户是审核人（从节点定义获取），允许访问",
												zap.String("order_id", orderID),
												zap.String("username", username),
												zap.String("node_code", node.NodeCode),
											)
											break
										}
									}
								}
								if isApproverOrExecutor {
									break
								}
							}
						case "user":
							userIDs := strings.Split(node.ApproverIDs, ",")
							for _, uidStr := range userIDs {
								uidStr = strings.TrimSpace(uidStr)
								if uidStr == "" {
									continue
								}
								uidInt, err := strconv.ParseUint(uidStr, 10, 64)
								if err != nil {
									continue
								}
								if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
									if adminUser.Username == username {
										isApproverOrExecutor = true
										global.Logger.Debug("用户是审核人（从节点定义获取），允许访问",
											zap.String("order_id", orderID),
											zap.String("username", username),
											zap.String("node_code", node.NodeCode),
										)
										break
									}
								}
							}
						}
						if isApproverOrExecutor {
							break
						}
					}
					// 检查执行节点
					if (strings.Contains(node.NodeCode, "execute") || strings.Contains(node.NodeName, "执行")) && node.ApproverType != "" {
						// 根据审批人类型获取用户列表
						switch node.ApproverType {
						case "role":
							roles := strings.Split(node.ApproverIDs, ",")
							for _, role := range roles {
								role = strings.TrimSpace(role)
								if role == "" {
									continue
								}
								userIDs, _ := global.Enforcer.GetUsersForRole(role)
								for _, uidStr := range userIDs {
									uidInt, err := strconv.ParseUint(uidStr, 10, 64)
									if err != nil {
										continue
									}
									if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
										if adminUser.Username == username {
											isApproverOrExecutor = true
											global.Logger.Debug("用户是执行人（从节点定义获取），允许访问",
												zap.String("order_id", orderID),
												zap.String("username", username),
												zap.String("node_code", node.NodeCode),
											)
											break
										}
									}
								}
								if isApproverOrExecutor {
									break
								}
							}
						case "user":
							userIDs := strings.Split(node.ApproverIDs, ",")
							for _, uidStr := range userIDs {
								uidStr = strings.TrimSpace(uidStr)
								if uidStr == "" {
									continue
								}
								uidInt, err := strconv.ParseUint(uidStr, 10, 64)
								if err != nil {
									continue
								}
								if adminUser, err := service.AdminServiceApp.GetAdminUser(c, uint(uidInt)); err == nil {
									if adminUser.Username == username {
										isApproverOrExecutor = true
										global.Logger.Debug("用户是执行人（从节点定义获取），允许访问",
											zap.String("order_id", orderID),
											zap.String("username", username),
											zap.String("node_code", node.NodeCode),
										)
										break
									}
								}
							}
						}
						if isApproverOrExecutor {
							break
						}
					}
				}
			}
		}

		// 如果是审核人或执行人，直接允许访问
		if isApproverOrExecutor {
			isRelatedUser = true
		} else {
			// 检查是否是其他相关人员（申请人、复核人、抄送人）
			var relatedUsers []string

			// 添加申请人
			if order.Applicant != "" {
				relatedUsers = append(relatedUsers, order.Applicant)
			}

			// 解析复核人 JSON
			if len(order.Reviewer) > 0 {
				var reviewers []map[string]interface{}
				if err := json.Unmarshal(order.Reviewer, &reviewers); err == nil {
					for _, reviewer := range reviewers {
						if user, ok := reviewer["user"].(string); ok && user != "" {
							relatedUsers = append(relatedUsers, user)
						}
					}
				}
			}

			// 解析抄送人 JSON
			if len(order.CC) > 0 {
				var ccs []map[string]interface{}
				if err := json.Unmarshal(order.CC, &ccs); err == nil {
					for _, cc := range ccs {
						if user, ok := cc["user"].(string); ok && user != "" {
							relatedUsers = append(relatedUsers, user)
						}
					}
				}
			}

			// 检查当前用户是否在其他相关人员列表中
			isRelatedUser = false
			for _, user := range relatedUsers {
				if user == username {
					isRelatedUser = true
					break
				}
			}
		}

		// 如果不是相关人员，隐藏 SQL 内容、执行结果和执行日志
		if !isRelatedUser {
			order.Content = "您没有权限查看当前工单内容"
		}
	}

	// 获取申请人昵称
	if order.Applicant != "" {
		// 根据用户名查询用户信息获取昵称
		baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
		adminRepo := repository.NewAdminRepository(baseRepo)
		if adminUser, err := adminRepo.GetAdminUserByUsername(c.Request.Context(), order.Applicant); err == nil {
			order.ApplicantNickname = adminUser.Nickname
		}
	}

	// 获取任务列表
	var tasks []insight.OrderTask
	if isRelatedUser {
		// 只有相关人员才能查看执行结果
		tasks, err = service.InsightServiceApp.GetOrderTasks(c.Request.Context(), orderID)
		if err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}
	} else {
		// 非相关人员返回空数组
		tasks = []insight.OrderTask{}
	}

	// 获取操作日志
	var logs []insight.OrderOpLog
	if isRelatedUser {
		// 只有相关人员才能查看操作日志
		logs, err = service.InsightServiceApp.GetOpLogs(c.Request.Context(), orderID)
		if err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}
	} else {
		// 非相关人员返回空数组
		logs = []insight.OrderOpLog{}
	}

	api.HandleSuccess(c, gin.H{
		"order":        order,
		"tasks":        tasks,
		"logs":         logs,
		"flowInstance": flowInstance,
	})
}

// FlexibleTime 灵活的时间类型，支持多种时间格式
type FlexibleTime struct {
	*time.Time
}

// UnmarshalJSON 自定义 JSON 解析，支持 ISO 8601 和 "2006-01-02 15:04:05" 格式
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		// 如果解析失败，可能是 null
		ft.Time = nil
		return nil
	}

	if timeStr == "" {
		ft.Time = nil
		return nil
	}

	// 先尝试解析 "2006-01-02 15:04:05" 格式
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local); err == nil {
		ft.Time = &t
		return nil
	}

	// 再尝试解析 ISO 8601 格式
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		ft.Time = &t
		return nil
	}

	// 尝试解析其他常见格式
	if t, err := time.Parse("2006-01-02T15:04:05", timeStr); err == nil {
		ft.Time = &t
		return nil
	}

	return fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// CreateOrderRequest 创建工单请求
type CreateOrderRequest struct {
	Title              string       `json:"title" binding:"required"`
	Remark             string       `json:"remark"`
	IsRestrictAccess   bool         `json:"is_restrict_access"`
	DBType             string       `json:"db_type" binding:"required"`
	SQLType            string       `json:"sql_type" binding:"required"`
	Environment        int          `json:"environment" binding:"required"`
	InstanceID         string       `json:"instance_id" binding:"required"`
	Schema             string       `json:"schema" binding:"required"`
	Content            string       `json:"content" binding:"required"`
	Approver           []string     `json:"approver"`
	Executor           []string     `json:"executor"`
	Reviewer           []string     `json:"reviewer"`
	CC                 []string     `json:"cc"`
	GhostOkToDropTable bool         `json:"ghost_ok_to_drop_table"` // gh-ost执行成功后自动删除旧表
	ScheduleTime       FlexibleTime `json:"schedule_time"`
	FixVersion         string       `json:"fix_version"`
	ExportFileFormat   string       `json:"export_file_format"`
	GenerateRollback   *bool        `json:"generate_rollback"` // DML工单是否生成回滚语句，默认 true
}

// CreateOrder 创建工单
// @Summary 创建工单
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "工单信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 解析 InstanceID
	instanceUUID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// DML 工单：仅当明确传 generate_rollback 时使用传值，否则默认 true；非 DML 忽略该字段
	var generateRollback bool
	if req.SQLType == "DML" {
		if req.GenerateRollback != nil {
			generateRollback = *req.GenerateRollback
		} else {
			generateRollback = true // 默认生成回滚语句
		}
	} else {
		generateRollback = true // 非 DML 不使用该字段，占位即可
	}
	order := &insight.OrderRecord{
		Title:              req.Title,
		Remark:             req.Remark,
		IsRestrictAccess:   req.IsRestrictAccess,
		DBType:             insight.DbType(req.DBType),
		SQLType:            insight.SQLType(req.SQLType),
		Environment:        req.Environment,
		InstanceID:         instanceUUID,
		Schema:             req.Schema,
		Content:            req.Content,
		Applicant:          username,
		Progress:           insight.ProgressPending,
		ScheduleTime:       req.ScheduleTime.Time,
		FixVersion:         req.FixVersion,
		ExportFileFormat:   insight.ExportFileFormat(req.ExportFileFormat),
		GhostOkToDropTable: false, // 默认false，由审核人在审核时设置
		GenerateRollback:   &generateRollback,
	}

	// 转换 JSON 字段
	if len(req.Approver) > 0 {
		order.Approver, _ = jsonMarshal(req.Approver)
	}
	if len(req.Executor) > 0 {
		order.Executor, _ = jsonMarshal(req.Executor)
	}
	if len(req.Reviewer) > 0 {
		order.Reviewer, _ = jsonMarshal(req.Reviewer)
	}
	if len(req.CC) > 0 {
		order.CC, _ = jsonMarshal(req.CC)
	}

	// 创建工单
	if err := service.InsightServiceApp.CreateOrder(c.Request.Context(), order); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 启动流程引擎（必须，如果流程引擎未配置，返回错误）
	businessType := fmt.Sprintf("order_%s", strings.ToLower(string(order.SQLType)))
	flowResp, err := service.FlowServiceApp.StartFlow(c.Request.Context(), &api.StartFlowRequest{
		BusinessType: businessType,
		BusinessID:   order.OrderID.String(),
		Title:        order.Title,
		InitiatorID:  userId,
		Initiator:    username,
	})

	if err != nil || flowResp == nil {
		// 流程引擎未配置，删除刚创建的工单，返回错误
		// 直接使用 repository 删除工单
		baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
		repo := insightRepo.NewInsightRepository(baseRepo, global.Logger, global.Enforcer)
		_ = repo.DeleteOrder(c.Request.Context(), order.OrderID.String())
		api.HandleError(c, http.StatusBadRequest, fmt.Errorf("流程引擎未配置，请先为业务类型 %s 配置流程定义", businessType), nil)
		return
	}

	// 关联流程实例ID到工单
	order.FlowInstanceID = flowResp.FlowInstanceID
	if err := service.InsightServiceApp.UpdateOrder(c.Request.Context(), order); err != nil {
		// 如果更新失败，记录日志但不影响流程
		global.Logger.Warn("关联流程实例ID失败", zap.Error(err))
	}

	// 记录操作日志
	_ = service.InsightServiceApp.CreateOpLog(c.Request.Context(), &insight.OrderOpLog{
		Username: username,
		OrderID:  order.OrderID,
		Msg:      "创建工单",
	})

	// 发送通知：通知申请人、审批人、复核人、执行人、抄送人
	go func() {
		ctx := context.Background()

		// 获取环境名称
		envs, _ := service.InsightServiceApp.GetEnvironments(ctx)
		envName := ""
		for _, env := range envs {
			if int(env.ID) == req.Environment {
				envName = env.Name
				break
			}
		}

		// 从流程实例中获取审批人、执行人等信息
		var approvers, reviewers, executors []string
		var receivers []string

		if flowResp != nil && flowResp.FlowInstanceID > 0 {
			// 获取流程实例的待处理任务（审批人）
			flowDetail, _ := service.FlowServiceApp.GetFlowInstanceDetail(ctx, flowResp.FlowInstanceID)
			if flowDetail != nil && flowDetail.Tasks != nil {
				for _, task := range flowDetail.Tasks {
					if task.Assignee == "" {
						continue
					}
					// 获取显示名称：优先使用昵称，如果没有则使用用户名
					displayName := task.AssigneeNickname
					if displayName == "" {
						displayName = task.Assignee
					}
					// 根据节点代码分类
					if strings.Contains(task.NodeCode, "approval") || strings.Contains(task.NodeName, "审批") {
						approvers = append(approvers, displayName)
					} else if strings.Contains(task.NodeCode, "review") || strings.Contains(task.NodeName, "复核") {
						reviewers = append(reviewers, displayName)
					} else if strings.Contains(task.NodeCode, "execute") || strings.Contains(task.NodeName, "执行") {
						executors = append(executors, displayName)
					}
					// receivers 使用用户名（用于通知系统）
					receivers = append(receivers, task.Assignee)
				}
			}

			// 从流程定义中获取审批节点和执行节点配置（如果任务还未创建）
			baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
			flowRepo := repository.NewFlowRepository(baseRepo)
			instance, _ := flowRepo.GetFlowInstance(ctx, flowResp.FlowInstanceID)
			if instance != nil {
				// 获取流程定义
				flowDef, _ := flowRepo.GetFlowDefinition(ctx, instance.FlowDefID)
				if flowDef != nil {
					// 如果审批人列表为空，从流程定义中获取审批节点配置
					if len(approvers) == 0 {
						approvalNode, _ := flowRepo.GetFlowNodeByCode(ctx, flowDef.ID, "dba_approval")
						if approvalNode != nil {
							approverUsers := h.getUsersFromFlowNode(ctx, approvalNode)
							approvers = append(approvers, approverUsers...)
							// receivers 需要用户名，根据昵称查询用户名
							for _, nickname := range approverUsers {
								// 根据昵称查询用户名（用于通知系统）
								var adminUser model.AdminUser
								if err := global.DB.WithContext(ctx).Where("nickname = ?", nickname).First(&adminUser).Error; err == nil {
									receivers = append(receivers, adminUser.Username)
								}
							}
						}
					}
					// 如果执行人列表为空，从流程定义中获取执行节点配置
					if len(executors) == 0 {
						executeNode, _ := flowRepo.GetFlowNodeByCode(ctx, flowDef.ID, "dba_execute")
						if executeNode != nil {
							// 根据执行节点的配置获取执行人列表（返回昵称）
							executorUsers := h.getUsersFromFlowNode(ctx, executeNode)
							executors = append(executors, executorUsers...)
							// receivers 需要用户名，根据昵称查询用户名
							for _, nickname := range executorUsers {
								// 根据昵称查询用户名（用于通知系统）
								var adminUser model.AdminUser
								if err := global.DB.WithContext(ctx).Where("nickname = ?", nickname).First(&adminUser).Error; err == nil {
									receivers = append(receivers, adminUser.Username)
								}
							}
						}
					}
				}
			}
		}

		// 如果流程中没有找到，尝试从请求参数中获取（兼容旧数据）
		if len(approvers) == 0 {
			approvers = req.Approver
		}
		if len(reviewers) == 0 {
			reviewers = req.Reviewer
		}
		if len(executors) == 0 {
			executors = req.Executor
		}

		// 添加抄送人
		receivers = append(receivers, req.CC...)
		receivers = append(receivers, approvers...)
		receivers = append(receivers, reviewers...)
		receivers = append(receivers, executors...)

		// 格式化显示（如果为空显示"未指定"）
		approverStr := strings.Join(approvers, ",")
		if approverStr == "" {
			approverStr = "未指定"
		}
		reviewerStr := strings.Join(reviewers, ",")
		if reviewerStr == "" {
			reviewerStr = "未指定"
		}
		executorStr := strings.Join(executors, ",")
		if executorStr == "" {
			executorStr = "未指定"
		}
		ccStr := strings.Join(req.CC, ",")
		if ccStr == "" {
			ccStr = "无"
		}

		// 获取申请人昵称（用于消息显示）
		applicantDisplayName := username
		baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
		adminRepo := repository.NewAdminRepository(baseRepo)
		if adminUser, err := adminRepo.GetAdminUserByUsername(ctx, username); err == nil && adminUser.Nickname != "" {
			applicantDisplayName = adminUser.Nickname
		}

		// 构建通知消息
		msg := fmt.Sprintf(
			"您好，用户%s提交了工单\n"+
				">工单标题：%s\n"+
				">备注：%s\n"+
				">审核人：%s\n"+
				">执行人：%s\n"+
				">环境：%s\n"+
				">数据库类型：%s\n"+
				">工单类型：%s\n"+
				">库名：%s",
			applicantDisplayName, order.Title, order.Remark,
			approverStr, executorStr,
			envName, order.DBType, order.SQLType, order.Schema,
		)

		notifier.SendOrderNotification(
			order.OrderID.String(),
			order.Title,
			username,
			receivers,
			msg,
		)
	}()

	api.HandleSuccess(c, order)
}

// UpdateOrderProgressRequest 更新工单进度请求
type UpdateOrderProgressRequest struct {
	OrderID  string `json:"order_id" binding:"required"`
	Progress string `json:"progress" binding:"required"`
	Remark   string `json:"remark"`
}

// UpdateOrderProgress 更新工单进度
// @Summary 更新工单进度
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UpdateOrderProgressRequest true "进度信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/progress [put]
func (h *OrderHandler) UpdateOrderProgress(c *gin.Context) {
	var req UpdateOrderProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 获取工单
	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), req.OrderID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 更新进度
	if err := service.InsightServiceApp.UpdateOrderProgress(c.Request.Context(), req.OrderID, insight.Progress(req.Progress)); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 记录操作日志
	msg := "更新工单进度为: " + req.Progress
	if req.Remark != "" {
		msg += ", 备注: " + req.Remark
	}
	_ = service.InsightServiceApp.CreateOpLog(c.Request.Context(), &insight.OrderOpLog{
		Username: username,
		OrderID:  order.OrderID,
		Msg:      msg,
	})

	api.HandleSuccess(c, nil)
}

// ApproveOrderRequest 审批工单请求
type ApproveOrderRequest struct {
	OrderID            string `json:"order_id" binding:"required"`
	Status             string `json:"status" binding:"required,oneof=pass reject"` // pass: 通过, reject: 驳回
	Msg                string `json:"msg"`                                         // 审批意见
	GhostOkToDropTable bool   `json:"ghost_ok_to_drop_table"`                      // gh-ost执行成功后自动删除旧表（仅DDL工单有效）
}

// ApproveOrder 审批工单
// @Summary 审批工单
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body ApproveOrderRequest true "审批信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/approve [post]
func (h *OrderHandler) ApproveOrder(c *gin.Context) {
	var req ApproveOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 获取工单
	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), req.OrderID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 检查工单状态
	if order.Progress != insight.ProgressPending {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "工单状态不允许审批")
		return
	}

	// 检查是否有流程实例（新工单必须有）
	if order.FlowInstanceID == 0 {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "工单未关联流程实例，无法审批")
		return
	}

	// 使用流程引擎审批
	// 获取当前用户的待办任务
	flowRepo := repository.NewFlowRepository(repository.NewRepository(global.Logger, global.DB, global.Enforcer))
	tasks, err := flowRepo.GetPendingTasksByInstance(c.Request.Context(), order.FlowInstanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	var myTask *model.FlowTask
	for i := range tasks {
		if tasks[i].AssigneeID == userId {
			myTask = &tasks[i]
			break
		}
	}

	if myTask == nil {
		api.HandleError(c, http.StatusForbidden, api.ErrForbidden, "您没有当前工单的审批权限")
		return
	}

	// 如果是DDL工单且审核通过，更新 gh-ost 参数
	if req.Status == "pass" && order.SQLType == insight.SQLTypeDDL {
		// 使用 UpdateDBConfigFields 的方式更新单个字段
		updates := map[string]interface{}{
			"ghost_ok_to_drop_table": req.GhostOkToDropTable,
		}
		if err := service.InsightServiceApp.UpdateOrderFields(c.Request.Context(), req.OrderID, updates); err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}
	}

	// 调用流程引擎审批
	if req.Status == "pass" {
		if err := service.FlowServiceApp.ApproveTask(c.Request.Context(), &api.ApproveTaskRequest{
			TaskID:     myTask.ID,
			Comment:    req.Msg,
			OperatorID: userId,
			Operator:   username,
		}); err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}
	} else {
		if err := service.FlowServiceApp.RejectTask(c.Request.Context(), &api.RejectTaskRequest{
			TaskID:     myTask.ID,
			Comment:    req.Msg,
			OperatorID: userId,
			Operator:   username,
		}); err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}
	}

	api.HandleSuccess(c, nil)
}

// GetOrderTasks 获取工单任务列表
// @Summary 获取工单任务列表
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param order_id path string true "工单ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/{order_id}/tasks [get]
func (h *OrderHandler) GetOrderTasks(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 检查工单访问权限
	isRelatedUser, err := h.checkOrderAccess(c, orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 如果不是相关人员，返回空数组
	if !isRelatedUser {
		api.HandleSuccess(c, []insight.OrderTask{})
		return
	}

	tasks, err := service.InsightServiceApp.GetOrderTasks(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, tasks)
}

// GetTaskRollbackSQL 获取任务回滚SQL
// @Summary 获取任务回滚SQL
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param order_id path string true "工单ID"
// @Param task_id path string true "任务ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/{order_id}/tasks/{task_id}/rollback-sql [get]
func (h *OrderHandler) GetTaskRollbackSQL(c *gin.Context) {
	orderID := c.Param("order_id")
	taskID := c.Param("task_id")
	if orderID == "" || taskID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	rollbackSQL, err := service.InsightServiceApp.GetTaskRollbackSQL(c.Request.Context(), taskID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, gin.H{
		"rollback_sql": rollbackSQL,
	})
}

// UpdateTaskProgressRequest 更新任务进度请求
type UpdateTaskProgressRequest struct {
	TaskID   string `json:"task_id" binding:"required"`
	Progress string `json:"progress" binding:"required"`
}

// UpdateTaskProgress 更新任务进度
// @Summary 更新任务进度
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UpdateTaskProgressRequest true "进度信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/tasks/progress [put]
func (h *OrderHandler) UpdateTaskProgress(c *gin.Context) {
	var req UpdateTaskProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	if err := service.InsightServiceApp.UpdateTaskProgress(c.Request.Context(), req.TaskID, insight.TaskProgress(req.Progress), nil); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// GetOrderLogs 获取工单操作日志
// @Summary 获取工单操作日志
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param order_id path string true "工单ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/{order_id}/logs [get]
func (h *OrderHandler) GetOrderLogs(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 检查工单访问权限
	isRelatedUser, err := h.checkOrderAccess(c, orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 如果不是相关人员，返回空数组
	if !isRelatedUser {
		api.HandleSuccess(c, []insight.OrderMessage{})
		return
	}

	logs, err := service.InsightServiceApp.GetOpLogs(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, logs)
}

// GetGhostProgress 获取 gh-ost 最新进度信息（从 Redis 缓存）
// @Summary 获取 gh-ost 最新进度
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param order_id path string true "工单ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/{order_id}/ghost-progress [get]
func (h *OrderHandler) GetGhostProgress(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 从 Redis 获取最新进度
	progress, err := utils.GetGhostProgressFromRedis(orderID)
	if err != nil {
		global.Logger.Warn("Failed to get ghost progress from Redis",
			zap.String("order_id", orderID),
			zap.Error(err),
		)
		// 获取失败不影响，返回空数据（不是错误）
		api.HandleSuccess(c, nil)
		return
	}

	// 如果没有进度数据，返回 null（不是错误）
	if progress == nil {
		api.HandleSuccess(c, nil)
		return
	}

	api.HandleSuccess(c, progress)
}

// GetMyOrders 获取我的工单
// @Summary 获取我的工单
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Param progress query string false "进度"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/my [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	username := ""
	user, err := service.AdminServiceApp.GetAdminUser(c, userId)
	if err == nil {
		username = user.Username
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	progress := c.Query("progress")

	params := &insightRepo.OrderQueryParams{
		Page:      page,
		PageSize:  pageSize,
		Applicant: username,
		Progress:  progress,
	}

	orders, total, err := service.InsightServiceApp.GetOrders(c.Request.Context(), params)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, gin.H{
		"list":  orders,
		"total": total,
	})
}

// jsonMarshal JSON序列化辅助函数
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// getUsersFromFlowNode 从流程节点配置中获取用户列表（返回昵称，用于展示）
func (h *OrderHandler) getUsersFromFlowNode(ctx context.Context, node *model.FlowNode) []string {
	var users []string

	switch node.ApproverType {
	case model.ApproverTypeRole:
		// 获取角色下的所有用户
		roles := strings.Split(node.ApproverIDs, ",")
		for _, role := range roles {
			role = strings.TrimSpace(role)
			if role == "" {
				continue
			}
			userIDs, _ := global.Enforcer.GetUsersForRole(role)
			for _, uidStr := range userIDs {
				uidInt, err := strconv.ParseUint(uidStr, 10, 64)
				if err != nil {
					continue
				}
				// 根据用户ID查询真实用户名和昵称，优先使用昵称展示
				if adminUser, err := service.AdminServiceApp.GetAdminUser(ctx, uint(uidInt)); err == nil {
					if adminUser.Nickname != "" {
						users = append(users, adminUser.Nickname)
					} else {
						users = append(users, adminUser.Username)
					}
				}
			}
		}
	case model.ApproverTypeUser:
		// 指定用户
		userIDs := strings.Split(node.ApproverIDs, ",")
		for _, uidStr := range userIDs {
			uidStr = strings.TrimSpace(uidStr)
			if uidStr == "" {
				continue
			}
			uidInt, err := strconv.ParseUint(uidStr, 10, 64)
			if err != nil {
				continue
			}
			// 根据用户ID查询真实用户名和昵称，优先使用昵称展示
			if adminUser, err := service.AdminServiceApp.GetAdminUser(ctx, uint(uidInt)); err == nil {
				if adminUser.Nickname != "" {
					users = append(users, adminUser.Nickname)
				} else {
					users = append(users, adminUser.Username)
				}
			}
		}
	}

	return users
}

// ExecuteTaskRequest 执行任务请求
type ExecuteTaskRequest struct {
	TaskID  string `json:"task_id"`  // 执行单个任务时使用
	OrderID string `json:"order_id"` // 执行全部任务时使用
}

// ExecuteTask 执行工单任务（支持单个任务和全部任务）
// @Summary 执行工单任务
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body ExecuteTaskRequest true "任务信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/tasks/execute [post]
func (h *OrderHandler) ExecuteTask(c *gin.Context) {
	var req ExecuteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 如果提供了 order_id，执行全部任务
	if req.OrderID != "" {
		h.executeAllTasks(c, req.OrderID, username, userId)
		return
	}

	// 如果提供了 task_id，执行单个任务
	if req.TaskID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "task_id 或 order_id 必须提供其中一个")
		return
	}

	// 获取任务信息
	task, err := service.InsightServiceApp.GetTaskByID(c.Request.Context(), req.TaskID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 获取工单信息
	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), task.OrderID.String())
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 检查执行权限
	if err := h.checkOrderStatus(c.Request.Context(), task.OrderID.String(), username, userId); err != nil {
		api.HandleError(c, http.StatusForbidden, err, nil)
		return
	}

	// 检查任务状态（避免重复执行）
	if task.Progress == insight.TaskProgressCompleted {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "当前任务已完成，请勿重复执行")
		return
	}
	if task.Progress == insight.TaskProgressExecuting {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "当前任务正在执行中，请勿重复执行")
		return
	}

	// 检查是否有其他任务正在执行中
	noExecutingTasks, err := service.InsightServiceApp.CheckTasksProgressIsDoing(c.Request.Context(), task.OrderID.String())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	if !noExecutingTasks {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "当前有任务正在执行中，请先等待执行完成")
		return
	}

	// 获取数据库配置
	dbConfig, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), order.InstanceID.String())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 使用事务同时更新任务状态和工单状态
	err = service.InsightServiceApp.UpdateTaskAndOrderProgress(c.Request.Context(), req.TaskID, task.OrderID.String(), insight.TaskProgressExecuting, insight.ProgressExecuting)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 记录操作日志
	_ = service.InsightServiceApp.CreateOpLog(c.Request.Context(), &insight.OrderOpLog{
		Username: username,
		OrderID:  order.OrderID,
		Msg:      "开始执行任务: " + req.TaskID,
	})

	// DML 是否生成回滚：nil 视为 true（兼容旧数据），否则用工单配置
	generateRollback := true
	if order.GenerateRollback != nil {
		generateRollback = *order.GenerateRollback
	}
	// 创建执行器配置
	global.Logger.Info("创建执行器配置",
		zap.String("order_id", order.OrderID.String()),
		zap.Bool("ghost_ok_to_drop_table", order.GhostOkToDropTable),
		zap.Bool("generate_rollback", generateRollback),
		zap.String("sql_type", string(task.SQLType)),
	)
	execConfig := &executor.DBConfig{
		Hostname:           dbConfig.Hostname,
		Port:               dbConfig.Port,
		UserName:           dbConfig.UserName,
		Password:           dbConfig.Password,
		Schema:             order.Schema,
		DBType:             string(dbConfig.DbType),
		SQLType:            string(task.SQLType),
		SQL:                task.SQL,
		OrderID:            order.OrderID.String(),
		TaskID:             task.TaskID.String(),
		ExportFileFormat:   string(order.ExportFileFormat),
		GhostOkToDropTable: order.GhostOkToDropTable,
		GenerateRollback:   generateRollback,
	}

	// 记录执行开始日志（用于调试）
	global.Logger.Info("Starting task execution (async)",
		zap.String("order_id", order.OrderID.String()),
		zap.String("task_id", task.TaskID.String()),
		zap.String("sql_type", string(task.SQLType)),
	)

	// 立即返回 HTTP 响应，避免超时
	api.HandleSuccess(c, gin.H{
		"message":  "任务已开始执行，请通过工单详情页查看执行进度",
		"task_id":  req.TaskID,
		"order_id": task.OrderID.String(),
	})

	// 在后台 goroutine 中异步执行任务
	// 使用独立的 context，不依赖 HTTP 请求的 context（避免 HTTP 超时导致执行中断）
	go func() {
		ctx := context.Background() // 使用独立的 context

		// 创建执行器
		exec, err := executor.NewExecuteSQL(execConfig)
		if err != nil {
			global.Logger.Error("Failed to create executor",
				zap.String("task_id", req.TaskID),
				zap.String("order_id", task.OrderID.String()),
				zap.Error(err),
			)
			_ = service.InsightServiceApp.UpdateTaskProgress(ctx, req.TaskID, insight.TaskProgressFailed, nil)
			_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
				Username: username,
				OrderID:  order.OrderID,
				Msg:      "创建执行器失败: " + err.Error(),
			})
			return
		}

		// 执行SQL（可能耗时很长，比如 gh-ost 执行几小时）
		global.Logger.Info("Executing task in background",
			zap.String("task_id", req.TaskID),
			zap.String("order_id", task.OrderID.String()),
		)
		result, err := exec.Run()

		// 保存执行结果
		resultJSON, _ := json.Marshal(result)
		if err != nil {
			global.Logger.Error("Task execution failed",
				zap.String("task_id", req.TaskID),
				zap.String("order_id", task.OrderID.String()),
				zap.Error(err),
			)
			_ = service.InsightServiceApp.UpdateTaskProgress(ctx, req.TaskID, insight.TaskProgressFailed, resultJSON)
			_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
				Username: username,
				OrderID:  order.OrderID,
				Msg:      "任务执行失败: " + err.Error(),
			})
			// 更新工单状态
			h.updateOrderStatusToFinish(ctx, order.OrderID.String())
			return
		}

		// 更新任务状态为完成
		global.Logger.Info("Task execution completed",
			zap.String("task_id", req.TaskID),
			zap.String("order_id", task.OrderID.String()),
			zap.Int64("affected_rows", result.AffectedRows),
		)
		_ = service.InsightServiceApp.UpdateTaskProgress(ctx, req.TaskID, insight.TaskProgressCompleted, resultJSON)
		_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
			Username: username,
			OrderID:  order.OrderID,
			Msg:      "任务执行成功，影响行数: " + strconv.FormatInt(result.AffectedRows, 10),
		})

		// 更新工单状态为已完成（如果所有任务都完成）
		h.updateOrderStatusToFinish(ctx, order.OrderID.String())
	}()
}

// checkOrderStatus 检查工单状态和执行权限
func (h *OrderHandler) checkOrderStatus(ctx context.Context, orderID string, username string, userID uint) error {
	order, err := service.InsightServiceApp.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 检查用户是否是 admin（用户ID=1 或拥有 admin 角色）
	isAdmin := false
	if userID == 1 {
		// 用户ID=1，是超级管理员
		isAdmin = true
	} else {
		// 检查用户是否有 admin 角色
		roles, err := global.Enforcer.GetRolesForUser(string(rune(userID)))
		if err == nil {
			for _, role := range roles {
				if role == model.AdminRole {
					isAdmin = true
					break
				}
			}
		}
	}

	// 如果不是 admin，检查执行权限
	if !isAdmin {
		var executorList []string
		if order.Executor != nil {
			if err := json.Unmarshal(order.Executor, &executorList); err != nil {
				return err
			}
		}
		// 如果 executor 列表不为空且当前用户不在列表中，则拒绝
		if len(executorList) > 0 && !utils.IsContain(executorList, username) {
			return api.ErrForbidden
		}
	}

	// 检查工单状态
	if order.Progress != insight.ProgressApproved && order.Progress != insight.ProgressExecuting {
		return api.ErrBadRequest
	}

	return nil
}

// updateOrderStatusToFinish 检查所有任务是否完成，如果完成则更新工单状态
func (h *OrderHandler) updateOrderStatusToFinish(ctx context.Context, orderID string) {
	allCompleted, err := service.InsightServiceApp.CheckAllTasksCompleted(ctx, orderID)
	if err != nil || !allCompleted {
		return
	}

	// 更新工单状态为已完成
	_ = service.InsightServiceApp.UpdateOrderProgress(ctx, orderID, insight.ProgressCompleted)

	// 同步流程引擎：自动完成执行节点任务，推进到结束节点
	go func() {
		order, err := service.InsightServiceApp.GetOrderByID(context.Background(), orderID)
		if err != nil || order == nil {
			return
		}

		// 如果工单有关联的流程实例，自动完成执行节点任务
		hasFlowInstance := false
		if order.FlowInstanceID > 0 {
			// 获取流程实例
			flowInstance, err := service.FlowServiceApp.GetFlowInstanceDetail(context.Background(), order.FlowInstanceID)
			if err != nil || flowInstance == nil {
				return
			}

			// 查找执行节点的待处理任务
			for _, task := range flowInstance.Tasks {
				// 查找执行节点（nodeCode 包含 "execute" 或 nodeName 包含 "执行"）
				if (strings.Contains(task.NodeCode, "execute") || strings.Contains(task.NodeName, "执行")) &&
					task.Status == "pending" {
					// 自动完成执行节点任务
					// 获取执行人信息（优先使用任务分配的执行人）
					operator := task.Assignee
					var operatorID uint

					// 如果没有分配执行人，尝试从工单中获取
					if operator == "" {
						// 如果工单执行人是 JSON 数组，解析第一个
						if len(order.Executor) > 0 {
							var executors []string
							if err := json.Unmarshal(order.Executor, &executors); err == nil && len(executors) > 0 {
								operator = executors[0]
							} else {
								// 如果解析失败，尝试解析为对象数组
								var executorObjs []map[string]interface{}
								if err := json.Unmarshal(order.Executor, &executorObjs); err == nil && len(executorObjs) > 0 {
									if user, ok := executorObjs[0]["user"].(string); ok {
										operator = user
									}
								}
							}
						}
					}

					// 如果仍然没有执行人，使用申请人
					if operator == "" {
						operator = order.Applicant
					}

					// 调用流程引擎审批接口，自动完成执行节点任务
					_ = service.FlowServiceApp.ApproveTask(context.Background(), &api.ApproveTaskRequest{
						TaskID:     task.ID,
						Comment:    "工单执行完成，自动完成执行节点",
						OperatorID: operatorID,
						Operator:   operator,
					})

					global.Logger.Info("工单执行完成，自动完成流程引擎执行节点任务",
						zap.String("order_id", orderID),
						zap.Uint("flow_instance_id", order.FlowInstanceID),
						zap.Uint("task_id", task.ID),
						zap.String("operator", operator),
					)
					hasFlowInstance = true
					break // 只处理第一个执行节点任务
				}
			}
		}

		// 如果工单没有关联流程实例，或者流程实例中没有执行节点任务，才发送通知
		// 如果有流程实例且已调用 ApproveTask，syncOrderStatusOnFlowCompleted 已经发送了通知，这里不再重复发送
		if !hasFlowInstance {
			msg := fmt.Sprintf("您好，工单已经执行完成，请悉知\n>工单标题：%s", order.Title)
			notifier.SendOrderNotification(order.OrderID.String(), order.Title, order.Applicant, []string{}, msg)
		}
	}()
}

// executeAllTasks 执行工单的所有任务
func (h *OrderHandler) executeAllTasks(c *gin.Context, orderID string, username string, userID uint) {
	// 获取工单信息
	order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 检查执行权限和工单状态
	if err := h.checkOrderStatus(c.Request.Context(), orderID, username, userID); err != nil {
		api.HandleError(c, http.StatusForbidden, err, nil)
		return
	}

	// 检查是否有任务正在执行中
	noExecutingTasks, err := service.InsightServiceApp.CheckTasksProgressIsDoing(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	if !noExecutingTasks {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "当前有任务正在执行中，请先等待执行完成")
		return
	}

	// 检查是否有已暂停的任务
	noPausedTasks, err := service.InsightServiceApp.CheckTasksProgressIsPause(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	if !noPausedTasks {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "当前有任务已暂停，可手动执行单个任务")
		return
	}

	// 获取工单的所有任务
	tasks, err := service.InsightServiceApp.GetOrderTasks(c.Request.Context(), orderID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	if len(tasks) == 0 {
		api.HandleSuccess(c, gin.H{
			"type":    "warning",
			"message": "没有需要执行的任务",
		})
		return
	}

	// 更新工单状态为执行中
	_ = service.InsightServiceApp.UpdateOrderProgress(c.Request.Context(), orderID, insight.ProgressExecuting)

	// 记录操作日志
	_ = service.InsightServiceApp.CreateOpLog(c.Request.Context(), &insight.OrderOpLog{
		Username: username,
		OrderID:  order.OrderID,
		Msg:      "开始批量执行任务",
	})

	// 立即返回 HTTP 响应，避免超时
	api.HandleSuccess(c, gin.H{
		"type":       "info",
		"message":    "任务已开始批量执行，请通过工单详情页查看执行进度",
		"order_id":   orderID,
		"task_count": len(tasks),
	})

	// 在后台 goroutine 中异步执行所有任务
	// 使用独立的 context，不依赖 HTTP 请求的 context（避免 HTTP 超时导致执行中断）
	go func() {
		ctx := context.Background() // 使用独立的 context

		var executedCount, successCount, failCount int
		var typeResult string

		// 获取数据库配置
		dbConfig, err := service.InsightServiceApp.GetDBConfigByInstanceID(ctx, order.InstanceID.String())
		if err != nil {
			global.Logger.Error("Failed to get DB config for batch execution",
				zap.String("order_id", orderID),
				zap.Error(err),
			)
			return
		}

		// 逐个执行任务
		for _, task := range tasks {
			// 跳过已完成的任务
			if task.Progress == insight.TaskProgressCompleted {
				continue
			}

			executedCount++

			// 更新任务状态为执行中
			_ = service.InsightServiceApp.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressExecuting, nil)

			// DML 是否生成回滚：nil 视为 true，否则用工单配置
			generateRollbackBatch := true
			if order.GenerateRollback != nil {
				generateRollbackBatch = *order.GenerateRollback
			}
			// 创建执行器配置
			global.Logger.Info("创建执行器配置（批量执行）",
				zap.String("order_id", order.OrderID.String()),
				zap.Bool("ghost_ok_to_drop_table", order.GhostOkToDropTable),
				zap.Bool("generate_rollback", generateRollbackBatch),
				zap.String("sql_type", string(task.SQLType)),
			)
			execConfig := &executor.DBConfig{
				Hostname:           dbConfig.Hostname,
				Port:               dbConfig.Port,
				UserName:           dbConfig.UserName,
				Password:           dbConfig.Password,
				Schema:             order.Schema,
				DBType:             string(dbConfig.DbType),
				SQLType:            string(task.SQLType),
				SQL:                task.SQL,
				OrderID:            order.OrderID.String(),
				TaskID:             task.TaskID.String(),
				ExportFileFormat:   string(order.ExportFileFormat),
				GhostOkToDropTable: order.GhostOkToDropTable,
				GenerateRollback:   generateRollbackBatch,
			}

			// 创建执行器
			exec, err := executor.NewExecuteSQL(execConfig)
			if err != nil {
				failCount++
				global.Logger.Error("Failed to create executor for task",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Error(err),
				)
				resultJSON, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
				_ = service.InsightServiceApp.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressFailed, resultJSON)
				_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
					Username: username,
					OrderID:  order.OrderID,
					Msg:      "任务执行失败: " + err.Error(),
				})
				continue
			}

			// 执行SQL（可能耗时很长，比如 gh-ost 执行几小时）
			global.Logger.Info("Executing task in background",
				zap.String("task_id", task.TaskID.String()),
				zap.String("order_id", orderID),
			)
			result, err := exec.Run()

			// 保存执行结果
			resultJSON, _ := json.Marshal(result)
			if err != nil {
				failCount++
				global.Logger.Error("Task execution failed",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Error(err),
				)
				_ = service.InsightServiceApp.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressFailed, resultJSON)
				_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
					Username: username,
					OrderID:  order.OrderID,
					Msg:      "任务执行失败: " + err.Error(),
				})
			} else {
				successCount++
				global.Logger.Info("Task execution completed",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Int64("affected_rows", result.AffectedRows),
				)
				_ = service.InsightServiceApp.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressCompleted, resultJSON)
				_ = service.InsightServiceApp.CreateOpLog(ctx, &insight.OrderOpLog{
					Username: username,
					OrderID:  order.OrderID,
					Msg:      "任务执行成功，影响行数: " + strconv.FormatInt(result.AffectedRows, 10),
				})
			}
		}

		// 根据执行结果确定类型
		if executedCount == 0 {
			typeResult = "warning"
		} else if successCount == executedCount {
			typeResult = "success"
		} else if failCount == executedCount {
			typeResult = "error"
		} else {
			typeResult = "warning"
		}

		// 更新工单执行结果
		_ = service.InsightServiceApp.UpdateOrderExecuteResult(ctx, orderID, typeResult)

		// 检查所有任务是否完成，如果完成则更新工单状态
		h.updateOrderStatusToFinish(ctx, orderID)

		global.Logger.Info("Batch task execution completed",
			zap.String("order_id", orderID),
			zap.Int("executed_count", executedCount),
			zap.Int("success_count", successCount),
			zap.Int("fail_count", failCount),
			zap.String("result", typeResult),
		)
	}()
}

// ControlGhostRequest gh-ost 控制请求
type ControlGhostRequest struct {
	OrderID string `json:"order_id" binding:"required"` // 工单ID
	Action  string `json:"action" binding:"required"`   // 操作类型：throttle(暂停), unthrottle(恢复), panic(取消), chunk-size(调节速度)
	Value   *int   `json:"value,omitempty"`             // 操作值（仅用于 chunk-size，表示新的 chunk-size 值）
}

// ControlGhost 控制 gh-ost 执行
// @Summary 控制 gh-ost 执行（暂停/取消/速度调节）
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body ControlGhostRequest true "控制请求"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/ghost/control [post]
func (h *OrderHandler) ControlGhost(c *gin.Context) {
	var req ControlGhostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 验证操作类型
	validActions := map[string]bool{
		"throttle":   true, // 暂停
		"unthrottle": true, // 恢复
		"panic":      true, // 取消
		"chunk-size": true, // 调节速度
	}
	if !validActions[req.Action] {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "不支持的操作类型，支持的操作：throttle(暂停), unthrottle(恢复), panic(取消), chunk-size(调节速度)")
		return
	}

	// chunk-size 操作需要提供 value
	if req.Action == "chunk-size" {
		if req.Value == nil || *req.Value <= 0 {
			api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "chunk-size 操作需要提供有效的 value 值（大于 0）")
			return
		}
	}

	// 获取 socket 路径
	// 首先尝试从 Redis 获取
	socketPath, err := utils.GetGhostSocketPathFromOrderID(req.OrderID, "", "")
	if err != nil {
		// Redis 中没有，尝试从工单信息推断
		global.Logger.Warn("Failed to get ghost socket path from Redis, trying to infer from order",
			zap.String("order_id", req.OrderID),
			zap.Error(err),
		)

		// 获取工单信息
		order, err := service.InsightServiceApp.GetOrderByID(c.Request.Context(), req.OrderID)
		if err != nil {
			global.Logger.Error("Failed to get order info",
				zap.String("order_id", req.OrderID),
				zap.Error(err),
			)
			api.HandleError(c, http.StatusOK, fmt.Errorf("任务未找到或者未执行"), nil)
			return
		}

		// 从 SQL 内容中提取表名（如果是 ALTER TABLE 语句）
		var tableName string
		if order.SQLType == insight.SQLTypeDDL && order.Content != "" {
			// 尝试从 SQL 内容中提取表名
			extractedTable, err := parser.GetTableNameFromAlterStatement(order.Content)
			if err == nil {
				// 如果提取成功，处理可能的 schema.table 格式
				if strings.Contains(extractedTable, ".") {
					parts := strings.SplitN(extractedTable, ".", 2)
					tableName = strings.Trim(parts[1], "`")
				} else {
					tableName = strings.Trim(extractedTable, "`")
				}
			}
		}

		// 使用工单的 schema 和提取的表名推断 socket 路径
		if order.Schema != "" && tableName != "" {
			socketPath, err = utils.GetGhostSocketPathFromOrderID(req.OrderID, order.Schema, tableName)
			if err != nil {
				global.Logger.Error("Failed to get ghost socket path (inferred)",
					zap.String("order_id", req.OrderID),
					zap.String("schema", order.Schema),
					zap.String("table", tableName),
					zap.Error(err),
				)
				api.HandleError(c, http.StatusOK, fmt.Errorf("任务未找到或者未执行"), nil)
				return
			}
		} else {
			global.Logger.Error("Cannot infer socket path: missing schema or table",
				zap.String("order_id", req.OrderID),
				zap.String("schema", order.Schema),
				zap.String("table", tableName),
			)
			api.HandleError(c, http.StatusOK, fmt.Errorf("任务未找到或者未执行"), nil)
			return
		}
	}

	// 构建命令
	var command string
	switch req.Action {
	case "throttle":
		command = "throttle"
	case "unthrottle":
		command = "unthrottle"
	case "panic":
		command = "panic"
	case "chunk-size":
		command = fmt.Sprintf("chunk-size=%d", *req.Value)
	default:
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "不支持的操作类型")
		return
	}

	// 发送命令给 gh-ost
	if err := utils.GhostControl(socketPath, command); err != nil {
		global.Logger.Error("Failed to control gh-ost",
			zap.String("order_id", req.OrderID),
			zap.String("socket_path", socketPath),
			zap.String("command", command),
			zap.Error(err),
		)
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 推送消息到 WebSocket
	message := fmt.Sprintf("gh-ost 控制命令已发送：%s", command)
	if req.Action == "chunk-size" {
		message = fmt.Sprintf("gh-ost 速度已调节：chunk-size=%d", *req.Value)
	}
	if err := utils.PublishMessageToChannel(req.OrderID, message, "ghost"); err != nil {
		global.Logger.Warn("Failed to publish ghost control message",
			zap.String("order_id", req.OrderID),
			zap.Error(err),
		)
	}

	api.HandleSuccess(c, gin.H{
		"message": message,
	})
}

// GetOrderTables 获取工单场景的表列表（不检查 DAS 查询权限）
// @Summary 获取工单表列表
// @Tags 工单管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Param schema path string true "数据库名"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/orders/tables/{instance_id}/{schema} [get]
func (h *OrderHandler) GetOrderTables(c *gin.Context) {
	instanceID := c.Param("instance_id")
	schema := c.Param("schema")

	// 验证参数
	if instanceID == "" || schema == "" || instanceID == "undefined" || schema == "undefined" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取数据库配置
	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	db := &dao.MySQLDB{
		User:     config.UserName,
		Password: config.Password,
		Host:     config.Hostname,
		Port:     config.Port,
		Ctx:      ctx,
		Params:   map[string]string{"group_concat_max_len": "4194304"},
	}

	// 直接获取所有表，不检查用户的 DAS 查询权限
	// 工单场景下，用户需要能看到表结构来编写 SQL，权限检查在工单审批/执行时进行
	tables, err := db.GetTables(schema)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, tables)
}
