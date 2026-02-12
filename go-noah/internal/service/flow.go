package service

import (
	"context"
	"errors"
	"fmt"
	"go-noah/api"
	"go-noah/internal/model"
	"go-noah/internal/model/insight"
	"go-noah/internal/repository"
	"go-noah/pkg/global"
	"go-noah/pkg/notifier"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FlowServiceApp 全局 Service 实例
var FlowServiceApp = new(FlowService)

type FlowService struct{}

func (s *FlowService) getFlowRepo() *repository.FlowRepository {
	return repository.NewFlowRepository(repository.NewRepository(global.Logger, global.DB, global.Enforcer))
}

func (s *FlowService) ensureOrderExecuteNode(ctx context.Context, flowDefID uint, businessType string) {
	if businessType != "order_ddl" && businessType != "order_dml" && businessType != "order_export" {
		return
	}

	repo := s.getFlowRepo()
	if _, err := repo.GetFlowNodeByCode(ctx, flowDefID, "dba_execute"); err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Warn("获取执行节点失败", zap.Error(err), zap.Uint("flow_def_id", flowDefID))
		return
	}

	approvalNode, err := repo.GetFlowNodeByCode(ctx, flowDefID, "dba_approval")
	if err != nil {
		global.Logger.Warn("获取审批节点失败", zap.Error(err), zap.Uint("flow_def_id", flowDefID))
		return
	}

	if approvalNode.NextNodeCode != "dba_execute" {
		approvalNode.NextNodeCode = "dba_execute"
		if err := repo.UpdateFlowNode(ctx, approvalNode); err != nil {
			global.Logger.Warn("更新审批节点失败", zap.Error(err), zap.Uint("flow_def_id", flowDefID))
			return
		}
	}

	executeNode := &model.FlowNode{
		FlowDefID:     flowDefID,
		NodeCode:      "dba_execute",
		NodeName:      "DBA执行",
		NodeType:      model.NodeTypeApproval,
		Sort:          approvalNode.Sort + 1,
		ApproverType:  model.ApproverTypeRole,
		ApproverIDs:   model.RoleDBA,
		MultiMode:     model.MultiModeAny,
		RejectAction:  model.RejectActionToStart,
		TimeoutHours:  24,
		TimeoutAction: "notify",
		NextNodeCode:  "end",
	}
	if err := repo.CreateFlowNode(ctx, executeNode); err != nil {
		global.Logger.Warn("创建执行节点失败", zap.Error(err), zap.Uint("flow_def_id", flowDefID))
		return
	}

	if endNode, err := repo.GetFlowNodeByCode(ctx, flowDefID, "end"); err == nil {
		if endNode.Sort <= executeNode.Sort {
			endNode.Sort = executeNode.Sort + 1
			if err := repo.UpdateFlowNode(ctx, endNode); err != nil {
				global.Logger.Warn("更新结束节点排序失败", zap.Error(err), zap.Uint("flow_def_id", flowDefID))
			}
		}
	}
}

// ============= 流程定义 =============

// GetFlowDefinitionList 获取流程定义列表
func (s *FlowService) GetFlowDefinitionList(ctx context.Context) (*api.FlowDefinitionListData, error) {
	repo := s.getFlowRepo()
	list, err := repo.GetFlowDefinitionList(ctx)
	if err != nil {
		return nil, err
	}

	var items []api.FlowDefinitionItem
	for _, f := range list {
		items = append(items, api.FlowDefinitionItem{
			ID:          f.ID,
			Code:        f.Code,
			Name:        f.Name,
			Type:        f.Type,
			Description: f.Description,
			Version:     f.Version,
			Status:      f.Status,
		})
	}

	return &api.FlowDefinitionListData{
		List: items,
	}, nil
}

// GetFlowDefinition 获取流程定义详情（包含节点）
func (s *FlowService) GetFlowDefinition(ctx context.Context, id uint) (*api.FlowDefinitionDetail, error) {
	repo := s.getFlowRepo()
	flow, err := repo.GetFlowDefinition(ctx, id)
	if err != nil {
		return nil, err
	}

	nodes, err := repo.GetFlowNodes(ctx, id)
	if err != nil {
		return nil, err
	}

	var nodeItems []api.FlowNodeItem
	for _, n := range nodes {
		nodeItems = append(nodeItems, api.FlowNodeItem{
			ID:            n.ID,
			NodeCode:      n.NodeCode,
			NodeName:      n.NodeName,
			NodeType:      n.NodeType,
			Sort:          n.Sort,
			ApproverType:  n.ApproverType,
			ApproverIDs:   n.ApproverIDs,
			MultiMode:     n.MultiMode,
			RejectAction:  n.RejectAction,
			TimeoutHours:  n.TimeoutHours,
			TimeoutAction: n.TimeoutAction,
			NextNodeCode:  n.NextNodeCode,
		})
	}

	return &api.FlowDefinitionDetail{
		ID:          flow.ID,
		Code:        flow.Code,
		Name:        flow.Name,
		Type:        flow.Type,
		Description: flow.Description,
		Version:     flow.Version,
		Status:      flow.Status,
		Nodes:       nodeItems,
	}, nil
}

// CreateFlowDefinition 创建流程定义
func (s *FlowService) CreateFlowDefinition(ctx context.Context, req *api.CreateFlowDefinitionRequest) error {
	repo := s.getFlowRepo()
	// 检查编码是否重复
	existing, _ := repo.GetFlowDefinitionByCode(ctx, req.Code)
	if existing != nil {
		return fmt.Errorf("流程编码 %s 已存在", req.Code)
	}

	flow := &model.FlowDefinition{
		Code:        req.Code,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Version:     1,
		Status:      req.Status,
	}

	return repo.CreateFlowDefinition(ctx, flow)
}

// UpdateFlowDefinition 更新流程定义
func (s *FlowService) UpdateFlowDefinition(ctx context.Context, req *api.UpdateFlowDefinitionRequest) error {
	repo := s.getFlowRepo()
	flow, err := repo.GetFlowDefinition(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("流程定义不存在")
	}

	flow.Name = req.Name
	flow.Description = req.Description
	flow.Status = req.Status

	return repo.UpdateFlowDefinition(ctx, flow)
}

// DeleteFlowDefinition 删除流程定义
func (s *FlowService) DeleteFlowDefinition(ctx context.Context, id uint) error {
	repo := s.getFlowRepo()
	return repo.DeleteFlowDefinition(ctx, id)
}

// SaveFlowNodes 保存流程节点（批量）
func (s *FlowService) SaveFlowNodes(ctx context.Context, req *api.SaveFlowNodesRequest) error {
	repo := s.getFlowRepo()
	// 先删除旧节点
	if err := repo.DeleteFlowNodesByFlowDefID(ctx, req.FlowDefID); err != nil {
		return err
	}

	// 创建新节点
	var nodes []model.FlowNode
	for _, n := range req.Nodes {
		nodes = append(nodes, model.FlowNode{
			FlowDefID:     req.FlowDefID,
			NodeCode:      n.NodeCode,
			NodeName:      n.NodeName,
			NodeType:      n.NodeType,
			Sort:          n.Sort,
			ApproverType:  n.ApproverType,
			ApproverIDs:   n.ApproverIDs,
			MultiMode:     n.MultiMode,
			RejectAction:  n.RejectAction,
			TimeoutHours:  n.TimeoutHours,
			TimeoutAction: n.TimeoutAction,
			NextNodeCode:  n.NextNodeCode,
		})
	}

	if len(nodes) > 0 {
		return repo.BatchCreateFlowNodes(ctx, nodes)
	}
	return nil
}

// ============= 流程实例 =============

// StartFlow 发起流程
func (s *FlowService) StartFlow(ctx context.Context, req *api.StartFlowRequest) (*api.StartFlowResponse, error) {
	repo := s.getFlowRepo()
	// 获取流程定义
	flowDef, err := repo.GetFlowDefinitionByType(ctx, req.BusinessType)
	if err != nil {
		return nil, fmt.Errorf("未找到业务类型 %s 对应的流程定义", req.BusinessType)
	}

	s.ensureOrderExecuteNode(ctx, flowDef.ID, req.BusinessType)

	// 获取开始节点
	nodes, err := repo.GetFlowNodes(ctx, flowDef.ID)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("流程节点配置错误")
	}

	var startNode *model.FlowNode
	for i := range nodes {
		if nodes[i].NodeType == model.NodeTypeStart {
			startNode = &nodes[i]
			break
		}
	}
	if startNode == nil {
		return nil, fmt.Errorf("未找到开始节点")
	}

	// 创建流程实例
	now := time.Now()
	instance := &model.FlowInstance{
		FlowDefID:       flowDef.ID,
		FlowCode:        flowDef.Code,
		BusinessType:    req.BusinessType,
		BusinessID:      req.BusinessID,
		Title:           req.Title,
		InitiatorID:     req.InitiatorID,
		Initiator:       req.Initiator,
		Status:          model.FlowStatusRunning,
		CurrentNodeCode: startNode.NextNodeCode,
		StartTime:       now,
	}

	if err := repo.CreateFlowInstance(ctx, instance); err != nil {
		return nil, err
	}

	// 记录日志
	repo.CreateFlowLog(ctx, &model.FlowLog{
		FlowInstID: instance.ID,
		NodeCode:   "start",
		NodeName:   "开始",
		OperatorID: req.InitiatorID,
		Operator:   req.Initiator,
		Action:     "发起流程",
	})

	// 创建第一个审批任务
	nextNode, err := repo.GetFlowNodeByCode(ctx, flowDef.ID, startNode.NextNodeCode)
	if err != nil {
		return nil, err
	}

	if err := s.createApprovalTasks(ctx, instance, nextNode); err != nil {
		return nil, err
	}

	return &api.StartFlowResponse{
		FlowInstanceID: instance.ID,
	}, nil
}

// createApprovalTasks 创建审批任务
func (s *FlowService) createApprovalTasks(ctx context.Context, instance *model.FlowInstance, node *model.FlowNode) error {
	repo := s.getFlowRepo()
	// 根据审批人类型确定审批人
	var assignees []struct {
		ID       uint
		Username string
		Nickname string
	}

	switch node.ApproverType {
	case model.ApproverTypeRole:
		// 获取角色下的所有用户
		roles := strings.Split(node.ApproverIDs, ",")
		for _, role := range roles {
			users, _ := global.Enforcer.GetUsersForRole(role)
			for _, uidStr := range users {
				uidInt, _ := convertor.ToInt(uidStr)
				uid := uint(uidInt)
				// 根据用户ID查询真实用户名和昵称
				username := uidStr // 默认使用ID作为fallback
				nickname := ""
				if adminUser, err := AdminServiceApp.GetAdminUser(ctx, uid); err == nil {
					username = adminUser.Username
					nickname = adminUser.Nickname
					global.Logger.Info("获取用户昵称", zap.String("username", username), zap.String("nickname", nickname))
				}
				assignees = append(assignees, struct {
					ID       uint
					Username string
					Nickname string
				}{ID: uid, Username: username, Nickname: nickname})
			}
		}
	case model.ApproverTypeUser:
		// 指定用户
		userIDs := strings.Split(node.ApproverIDs, ",")
		for _, uidStr := range userIDs {
			uidInt, _ := convertor.ToInt(uidStr)
			uid := uint(uidInt)
			// 根据用户ID查询真实用户名和昵称
			username := uidStr // 默认使用ID作为fallback
			nickname := ""
			if adminUser, err := AdminServiceApp.GetAdminUser(ctx, uid); err == nil {
				username = adminUser.Username
				nickname = adminUser.Nickname
			}
			assignees = append(assignees, struct {
				ID       uint
				Username string
				Nickname string
			}{ID: uid, Username: username, Nickname: nickname})
		}
	case model.ApproverTypeAuto:
		// 自动通过，直接推进到下一节点
		return s.moveToNextNode(ctx, instance, node)
	}

	// 创建任务
	for _, assignee := range assignees {
		task := &model.FlowTask{
			FlowInstID:       instance.ID,
			FlowNodeID:       node.ID,
			NodeCode:         node.NodeCode,
			NodeName:         node.NodeName,
			AssigneeID:       assignee.ID,
			Assignee:         assignee.Username,
			AssigneeNickname: assignee.Nickname,
			Status:           model.TaskStatusPending,
		}

		if node.TimeoutHours > 0 {
			dueTime := time.Now().Add(time.Duration(node.TimeoutHours) * time.Hour)
			task.DueTime = &dueTime
		}

		if err := repo.CreateFlowTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

// ApproveTask 审批通过
func (s *FlowService) ApproveTask(ctx context.Context, req *api.ApproveTaskRequest) error {
	repo := s.getFlowRepo()
	task, err := repo.GetFlowTask(ctx, req.TaskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if task.Status != model.TaskStatusPending {
		return fmt.Errorf("任务已处理")
	}

	// 更新任务状态
	now := time.Now()
	task.Status = model.TaskStatusApproved
	task.Action = "approve"
	task.Comment = req.Comment
	task.ActionTime = &now

	if err := repo.UpdateFlowTask(ctx, task); err != nil {
		return err
	}

	// 记录日志
	repo.CreateFlowLog(ctx, &model.FlowLog{
		FlowInstID: task.FlowInstID,
		FlowNodeID: task.FlowNodeID,
		NodeCode:   task.NodeCode,
		NodeName:   task.NodeName,
		OperatorID: req.OperatorID,
		Operator:   req.Operator,
		Action:     "审批通过",
		Comment:    req.Comment,
	})

	// 获取流程实例
	instance, err := repo.GetFlowInstance(ctx, task.FlowInstID)
	if err != nil {
		return err
	}

	// 获取当前节点
	node, err := repo.GetFlowNode(ctx, task.FlowNodeID)
	if err != nil {
		return err
	}

	// 判断是否所有任务都已完成（会签模式）
	if node.MultiMode == model.MultiModeAll {
		// 会签模式：需要所有任务都完成
		pendingTasks, _ := repo.GetPendingTasksByInstance(ctx, instance.ID)
		// 只检查当前节点的待处理任务
		currentNodePendingTasks := make([]model.FlowTask, 0)
		for _, t := range pendingTasks {
			if t.NodeCode == node.NodeCode {
				currentNodePendingTasks = append(currentNodePendingTasks, t)
			}
		}
		if len(currentNodePendingTasks) > 0 {
			return nil // 当前节点还有待处理任务
		}
	}
	// MultiModeAny（或签）模式：只要有一个任务通过，就可以推进到下一节点

	// 如果是执行节点通过，更新工单状态为"已完成"（DBA确认执行完成）
	if node.NodeCode == "dba_execute" || strings.Contains(node.NodeName, "执行") {
		s.syncOrderStatusOnFlowCompleted(ctx, instance)
		// 执行节点通过后，推进到下一节点（可能是结束节点）
		// 如果下一节点是结束节点，不发送"审批通过"通知，因为 syncOrderStatusOnFlowCompleted 已经发送了完成通知
		nextNodeCode := node.NextNodeCode
		if nextNodeCode == "" || nextNodeCode == "end" {
			// 下一节点是结束节点，直接推进，不发送审批通过通知
			return s.moveToNextNode(ctx, instance, node)
		}
		// 检查下一节点是否是结束节点
		flowDef, _ := repo.GetFlowDefinition(ctx, instance.FlowDefID)
		if flowDef != nil {
			nextNode, err := repo.GetFlowNodeByCode(ctx, flowDef.ID, nextNodeCode)
			if err == nil && nextNode != nil && nextNode.NodeType == model.NodeTypeEnd {
				// 下一节点是结束节点，直接推进，不发送审批通过通知
				return s.moveToNextNode(ctx, instance, node)
			}
		}
	}

	// 发送通知：审批通过，通知申请人（非结束节点的情况）
	if instance.BusinessType == "order_ddl" || instance.BusinessType == "order_dml" || instance.BusinessType == "order_export" {
		go func() {
			order, err := InsightServiceApp.GetOrderByID(context.Background(), instance.BusinessID)
			if err == nil && order != nil {
				msg := fmt.Sprintf("您好，%s审批通过了工单\n>工单标题：%s\n>附加消息：%s", req.Operator, order.Title, req.Comment)
				notifier.SendOrderNotification(order.OrderID.String(), order.Title, order.Applicant, []string{}, msg)
			}
		}()
	}

	// 推进到下一节点
	return s.moveToNextNode(ctx, instance, node)
}

// RejectTask 审批驳回
func (s *FlowService) RejectTask(ctx context.Context, req *api.RejectTaskRequest) error {
	repo := s.getFlowRepo()
	task, err := repo.GetFlowTask(ctx, req.TaskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	if task.Status != model.TaskStatusPending {
		return fmt.Errorf("任务已处理")
	}

	// 更新任务状态
	now := time.Now()
	task.Status = model.TaskStatusRejected
	task.Action = "reject"
	task.Comment = req.Comment
	task.ActionTime = &now

	if err := repo.UpdateFlowTask(ctx, task); err != nil {
		return err
	}

	// 记录日志
	repo.CreateFlowLog(ctx, &model.FlowLog{
		FlowInstID: task.FlowInstID,
		FlowNodeID: task.FlowNodeID,
		NodeCode:   task.NodeCode,
		NodeName:   task.NodeName,
		OperatorID: req.OperatorID,
		Operator:   req.Operator,
		Action:     "审批驳回",
		Comment:    req.Comment,
	})

	// 更新流程实例状态
	if err := repo.UpdateFlowInstanceStatus(ctx, task.FlowInstID, model.FlowStatusRejected); err != nil {
		return err
	}

	// 同步更新工单状态
	instance, _ := repo.GetFlowInstance(ctx, task.FlowInstID)
	if instance != nil && (instance.BusinessType == "order_ddl" || instance.BusinessType == "order_dml" || instance.BusinessType == "order_export") {
		_ = InsightServiceApp.UpdateOrderProgress(ctx, instance.BusinessID, insight.ProgressRejected)

		// 发送通知：审批驳回，通知申请人
		go func() {
			order, err := InsightServiceApp.GetOrderByID(context.Background(), instance.BusinessID)
			if err == nil && order != nil {
				msg := fmt.Sprintf("您好，%s驳回了工单\n>工单标题：%s\n>附加消息：%s", req.Operator, order.Title, req.Comment)
				notifier.SendOrderNotification(order.OrderID.String(), order.Title, order.Applicant, []string{}, msg)
			}
		}()
	}

	return nil
}

// moveToNextNode 推进到下一节点
func (s *FlowService) moveToNextNode(ctx context.Context, instance *model.FlowInstance, currentNode *model.FlowNode) error {
	repo := s.getFlowRepo()
	if currentNode.NextNodeCode == "" || currentNode.NextNodeCode == "end" {
		// 流程结束，保持当前节点不变（用于前端显示）
		// 更新流程状态为已批准
		if err := repo.UpdateFlowInstanceStatus(ctx, instance.ID, model.FlowStatusApproved); err != nil {
			return err
		}
		// 如果当前节点是执行节点，工单状态已经在 syncOrderStatusOnFlowCompleted 中更新为"已完成"
		// 这里不需要再调用 syncOrderStatusOnFlowApproved，避免将"已完成"改回"已批准"
		if currentNode.NodeCode == "dba_execute" || strings.Contains(currentNode.NodeName, "执行") {
			// 执行节点到达结束节点，工单状态已经是"已完成"，不需要更新
			return nil
		}
		// 同步更新工单状态（非执行节点的情况）
		s.syncOrderStatusOnFlowApproved(ctx, instance)
		return nil
	}

	// 获取下一节点
	flowDef, _ := repo.GetFlowDefinition(ctx, instance.FlowDefID)
	nextNode, err := repo.GetFlowNodeByCode(ctx, flowDef.ID, currentNode.NextNodeCode)
	if err != nil {
		return err
	}

	if nextNode.NodeType == model.NodeTypeEnd {
		// 流程结束，更新当前节点为当前节点（执行节点）（用于前端显示）
		// 不修改 CurrentNodeCode，保持为执行节点
		// 更新流程状态为已批准
		if err := repo.UpdateFlowInstanceStatus(ctx, instance.ID, model.FlowStatusApproved); err != nil {
			return err
		}
		// 同步更新工单状态（执行节点通过后，工单状态已经在 syncOrderStatusOnFlowCompleted 中更新为"已完成"）
		// 这里不需要再调用 syncOrderStatusOnFlowApproved
		return nil
	}

	// 更新当前节点
	instance.CurrentNodeCode = nextNode.NodeCode
	instance.CurrentNodeID = nextNode.ID
	if err := repo.UpdateFlowInstance(ctx, instance); err != nil {
		return err
	}

	// 如果是执行节点（从审批节点进入执行节点），同步更新工单状态为"执行中"
	if nextNode.NodeCode == "dba_execute" || strings.Contains(nextNode.NodeName, "执行") {
		s.syncOrderStatusOnFlowExecute(ctx, instance)
	}

	// 创建新的审批任务
	return s.createApprovalTasks(ctx, instance, nextNode)
}

// syncOrderStatusOnFlowApproved 流程审批通过后同步工单状态
func (s *FlowService) syncOrderStatusOnFlowApproved(ctx context.Context, instance *model.FlowInstance) {
	if instance.BusinessType != "order_ddl" && instance.BusinessType != "order_dml" && instance.BusinessType != "order_export" {
		return
	}

	// 更新工单状态为"已批准"
	if err := InsightServiceApp.UpdateOrderProgress(ctx, instance.BusinessID, insight.ProgressApproved); err != nil {
		global.Logger.Warn("同步工单状态失败", zap.Error(err), zap.String("business_id", instance.BusinessID))
		return
	}

	// UpdateOrderProgress 内部会自动生成任务，这里不需要再次调用
}

// syncOrderStatusOnFlowExecute 流程执行节点通过后同步工单状态（进入执行阶段）
func (s *FlowService) syncOrderStatusOnFlowExecute(ctx context.Context, instance *model.FlowInstance) {
	if instance.BusinessType != "order_ddl" && instance.BusinessType != "order_dml" && instance.BusinessType != "order_export" {
		return
	}

	// 更新工单状态为"已批准"（待执行），触发任务生成
	// 工单状态保持为"已批准"，等待真正开始执行任务时再更新为"执行中"
	if err := InsightServiceApp.UpdateOrderProgress(ctx, instance.BusinessID, insight.ProgressApproved); err != nil {
		global.Logger.Warn("同步工单状态为已批准失败", zap.Error(err), zap.String("business_id", instance.BusinessID))
		return
	}

	global.Logger.Info("工单审批通过，状态已更新为已批准（待执行）",
		zap.String("business_id", instance.BusinessID),
	)
}

// syncOrderStatusOnFlowCompleted 流程执行节点通过后同步工单状态（执行完成）
func (s *FlowService) syncOrderStatusOnFlowCompleted(ctx context.Context, instance *model.FlowInstance) {
	if instance.BusinessType != "order_ddl" && instance.BusinessType != "order_dml" && instance.BusinessType != "order_export" {
		return
	}

	// 更新工单状态为"已完成"（DBA确认执行完成）
	if err := InsightServiceApp.UpdateOrderProgress(ctx, instance.BusinessID, insight.ProgressCompleted); err != nil {
		global.Logger.Warn("同步工单状态失败", zap.Error(err), zap.String("business_id", instance.BusinessID))
		return
	}

	// 发送通知：工单完成，通知申请人
	go func() {
		order, err := InsightServiceApp.GetOrderByID(context.Background(), instance.BusinessID)
		if err == nil && order != nil {
			msg := fmt.Sprintf("您好，工单已经执行完成，请悉知\n>工单标题：%s", order.Title)
			notifier.SendOrderNotification(order.OrderID.String(), order.Title, order.Applicant, []string{}, msg)
		}
	}()
}

// GetMyPendingTasks 获取我的待办任务
func (s *FlowService) GetMyPendingTasks(ctx context.Context, userID uint, page, pageSize int) (*api.FlowTaskListData, error) {
	repo := s.getFlowRepo()
	list, total, err := repo.GetMyPendingTasks(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var items []api.FlowTaskItem
	for _, t := range list {
		items = append(items, api.FlowTaskItem{
			ID:         t.ID,
			FlowInstID: t.FlowInstID,
			NodeCode:   t.NodeCode,
			NodeName:   t.NodeName,
			Status:     t.Status,
			CreatedAt:  t.CreatedAt,
		})
	}

	return &api.FlowTaskListData{
		List:  items,
		Total: total,
	}, nil
}

// GetFlowInstanceDetail 获取流程实例详情
func (s *FlowService) GetFlowInstanceDetail(ctx context.Context, id uint) (*api.FlowInstanceDetail, error) {
	repo := s.getFlowRepo()
	instance, err := repo.GetFlowInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	s.ensureOrderExecuteNode(ctx, instance.FlowDefID, instance.BusinessType)

	// 获取流程节点信息
	nodes, _ := repo.GetFlowNodes(ctx, instance.FlowDefID)

	// 获取任务列表
	tasks, _ := repo.GetFlowTasksByInstance(ctx, id)
	var taskItems []api.FlowTaskItem
	for _, t := range tasks {
		assignee := t.Assignee
		assigneeNickname := t.AssigneeNickname
		// 如果 assignee 是纯数字（可能是用户ID），尝试查询真实用户名和昵称
		if assignee != "" {
			if uidInt, err := convertor.ToInt(assignee); err == nil {
				// 是数字，尝试查询用户名和昵称
				if adminUser, err := AdminServiceApp.GetAdminUser(ctx, uint(uidInt)); err == nil {
					assignee = adminUser.Username
					if assigneeNickname == "" {
						assigneeNickname = adminUser.Nickname
					}
				}
			}
		}

		taskItems = append(taskItems, api.FlowTaskItem{
			ID:               t.ID,
			FlowInstID:       t.FlowInstID,
			NodeCode:         t.NodeCode,
			NodeName:         t.NodeName,
			Assignee:         assignee,
			AssigneeNickname: assigneeNickname,
			Status:           t.Status,
			Action:           t.Action,
			Comment:          t.Comment,
			ActionTime:       t.ActionTime,
			CreatedAt:        t.CreatedAt,
		})
	}

	// 获取日志列表
	logs, _ := repo.GetFlowLogs(ctx, id)
	var logItems []api.FlowLogItem
	for _, l := range logs {
		logItems = append(logItems, api.FlowLogItem{
			ID:        l.ID,
			NodeCode:  l.NodeCode,
			NodeName:  l.NodeName,
			Operator:  l.Operator,
			Action:    l.Action,
			Comment:   l.Comment,
			CreatedAt: l.CreatedAt,
		})
	}

	// 构建节点信息（用于前端显示流程步骤）
	// 返回全部节点，保持与流程定义一致（包括开始/结束）
	// 节点已经按照 Sort 字段排序（GetFlowNodes 中已排序）
	var nodeItems []api.FlowNodeItem
	for _, n := range nodes {
		nodeItems = append(nodeItems, api.FlowNodeItem{
			ID:            n.ID,
			NodeCode:      n.NodeCode,
			NodeName:      n.NodeName,
			NodeType:      n.NodeType,
			Sort:          n.Sort,
			ApproverType:  n.ApproverType,
			ApproverIDs:   n.ApproverIDs,
			MultiMode:     n.MultiMode,
			RejectAction:  n.RejectAction,
			TimeoutHours:  n.TimeoutHours,
			TimeoutAction: n.TimeoutAction,
			NextNodeCode:  n.NextNodeCode,
		})
	}

	return &api.FlowInstanceDetail{
		ID:              instance.ID,
		FlowCode:        instance.FlowCode,
		BusinessType:    instance.BusinessType,
		BusinessID:      instance.BusinessID,
		Title:           instance.Title,
		Initiator:       instance.Initiator,
		Status:          instance.Status,
		CurrentNodeCode: instance.CurrentNodeCode,
		StartTime:       instance.StartTime,
		EndTime:         instance.EndTime,
		Tasks:           taskItems,
		Logs:            logItems,
		Nodes:           nodeItems,
	}, nil
}
