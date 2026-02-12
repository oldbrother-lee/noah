package repository

import (
	"context"
	"go-noah/internal/model"
	"time"
)

// FlowRepository 审批流程数据访问层
type FlowRepository struct {
	*Repository
}

func NewFlowRepository(repository *Repository) *FlowRepository {
	return &FlowRepository{
		Repository: repository,
	}
}

// ============= 流程定义 =============

// GetFlowDefinitionList 获取流程定义列表
func (r *FlowRepository) GetFlowDefinitionList(ctx context.Context) ([]model.FlowDefinition, error) {
	var list []model.FlowDefinition
	err := r.DB(ctx).Order("id").Find(&list).Error
	return list, err
}

// GetFlowDefinition 根据ID获取流程定义
func (r *FlowRepository) GetFlowDefinition(ctx context.Context, id uint) (*model.FlowDefinition, error) {
	var flow model.FlowDefinition
	err := r.DB(ctx).Where("id = ?", id).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// GetFlowDefinitionByCode 根据编码获取流程定义
func (r *FlowRepository) GetFlowDefinitionByCode(ctx context.Context, code string) (*model.FlowDefinition, error) {
	var flow model.FlowDefinition
	err := r.DB(ctx).Where("code = ?", code).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// GetFlowDefinitionByType 根据业务类型获取启用的流程定义
func (r *FlowRepository) GetFlowDefinitionByType(ctx context.Context, flowType string) (*model.FlowDefinition, error) {
	var flow model.FlowDefinition
	err := r.DB(ctx).Where("type = ? AND status = 1", flowType).Order("version DESC").First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// CreateFlowDefinition 创建流程定义
func (r *FlowRepository) CreateFlowDefinition(ctx context.Context, flow *model.FlowDefinition) error {
	return r.DB(ctx).Create(flow).Error
}

// UpdateFlowDefinition 更新流程定义
func (r *FlowRepository) UpdateFlowDefinition(ctx context.Context, flow *model.FlowDefinition) error {
	return r.DB(ctx).Save(flow).Error
}

// DeleteFlowDefinition 删除流程定义
func (r *FlowRepository) DeleteFlowDefinition(ctx context.Context, id uint) error {
	// 同时删除流程节点
	r.DB(ctx).Where("flow_def_id = ?", id).Delete(&model.FlowNode{})
	return r.DB(ctx).Delete(&model.FlowDefinition{}, id).Error
}

// ============= 流程节点 =============

// GetFlowNodes 获取流程的所有节点
func (r *FlowRepository) GetFlowNodes(ctx context.Context, flowDefID uint) ([]model.FlowNode, error) {
	var nodes []model.FlowNode
	err := r.DB(ctx).Where("flow_def_id = ?", flowDefID).Order("sort").Find(&nodes).Error
	return nodes, err
}

// GetFlowNode 根据ID获取节点
func (r *FlowRepository) GetFlowNode(ctx context.Context, id uint) (*model.FlowNode, error) {
	var node model.FlowNode
	err := r.DB(ctx).Where("id = ?", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetFlowNodeByCode 根据编码获取节点
func (r *FlowRepository) GetFlowNodeByCode(ctx context.Context, flowDefID uint, nodeCode string) (*model.FlowNode, error) {
	var node model.FlowNode
	err := r.DB(ctx).Where("flow_def_id = ? AND node_code = ?", flowDefID, nodeCode).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// CreateFlowNode 创建流程节点
func (r *FlowRepository) CreateFlowNode(ctx context.Context, node *model.FlowNode) error {
	return r.DB(ctx).Create(node).Error
}

// UpdateFlowNode 更新流程节点
func (r *FlowRepository) UpdateFlowNode(ctx context.Context, node *model.FlowNode) error {
	return r.DB(ctx).Save(node).Error
}

// DeleteFlowNode 删除流程节点
func (r *FlowRepository) DeleteFlowNode(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.FlowNode{}, id).Error
}

// BatchCreateFlowNodes 批量创建流程节点
func (r *FlowRepository) BatchCreateFlowNodes(ctx context.Context, nodes []model.FlowNode) error {
	return r.DB(ctx).Create(&nodes).Error
}

// DeleteFlowNodesByFlowDefID 删除流程的所有节点
func (r *FlowRepository) DeleteFlowNodesByFlowDefID(ctx context.Context, flowDefID uint) error {
	return r.DB(ctx).Where("flow_def_id = ?", flowDefID).Delete(&model.FlowNode{}).Error
}

// ============= 流程实例 =============

// CreateFlowInstance 创建流程实例
func (r *FlowRepository) CreateFlowInstance(ctx context.Context, instance *model.FlowInstance) error {
	return r.DB(ctx).Create(instance).Error
}

// GetFlowInstance 获取流程实例
func (r *FlowRepository) GetFlowInstance(ctx context.Context, id uint) (*model.FlowInstance, error) {
	var instance model.FlowInstance
	err := r.DB(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetFlowInstanceByBusiness 根据业务获取流程实例
func (r *FlowRepository) GetFlowInstanceByBusiness(ctx context.Context, businessType, businessID string) (*model.FlowInstance, error) {
	var instance model.FlowInstance
	err := r.DB(ctx).Where("business_type = ? AND business_id = ?", businessType, businessID).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// UpdateFlowInstance 更新流程实例
func (r *FlowRepository) UpdateFlowInstance(ctx context.Context, instance *model.FlowInstance) error {
	return r.DB(ctx).Save(instance).Error
}

// UpdateFlowInstanceStatus 更新流程实例状态
func (r *FlowRepository) UpdateFlowInstanceStatus(ctx context.Context, id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == model.FlowStatusApproved || status == model.FlowStatusRejected || status == model.FlowStatusCancelled {
		now := time.Now()
		updates["end_time"] = &now
	}
	return r.DB(ctx).Model(&model.FlowInstance{}).Where("id = ?", id).Updates(updates).Error
}

// GetMyInitiatedFlows 获取我发起的流程
func (r *FlowRepository) GetMyInitiatedFlows(ctx context.Context, userID uint, page, pageSize int) ([]model.FlowInstance, int64, error) {
	var list []model.FlowInstance
	var total int64

	db := r.DB(ctx).Model(&model.FlowInstance{}).Where("initiator_id = ?", userID)
	db.Count(&total)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ============= 审批任务 =============

// CreateFlowTask 创建审批任务
func (r *FlowRepository) CreateFlowTask(ctx context.Context, task *model.FlowTask) error {
	return r.DB(ctx).Create(task).Error
}

// GetFlowTask 获取审批任务
func (r *FlowRepository) GetFlowTask(ctx context.Context, id uint) (*model.FlowTask, error) {
	var task model.FlowTask
	err := r.DB(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetFlowTasksByInstance 获取流程实例的所有任务
func (r *FlowRepository) GetFlowTasksByInstance(ctx context.Context, flowInstID uint) ([]model.FlowTask, error) {
	var tasks []model.FlowTask
	err := r.DB(ctx).Where("flow_inst_id = ?", flowInstID).Order("created_at").Find(&tasks).Error
	return tasks, err
}

// GetPendingTasksByInstance 获取流程实例的待处理任务
func (r *FlowRepository) GetPendingTasksByInstance(ctx context.Context, flowInstID uint) ([]model.FlowTask, error) {
	var tasks []model.FlowTask
	err := r.DB(ctx).Where("flow_inst_id = ? AND status = ?", flowInstID, model.TaskStatusPending).Find(&tasks).Error
	return tasks, err
}

// UpdateFlowTask 更新审批任务
func (r *FlowRepository) UpdateFlowTask(ctx context.Context, task *model.FlowTask) error {
	return r.DB(ctx).Save(task).Error
}

// GetMyPendingTasks 获取我的待办任务
func (r *FlowRepository) GetMyPendingTasks(ctx context.Context, userID uint, page, pageSize int) ([]model.FlowTask, int64, error) {
	var list []model.FlowTask
	var total int64

	db := r.DB(ctx).Model(&model.FlowTask{}).Where("assignee_id = ? AND status = ?", userID, model.TaskStatusPending)
	db.Count(&total)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetMyProcessedTasks 获取我已处理的任务
func (r *FlowRepository) GetMyProcessedTasks(ctx context.Context, userID uint, page, pageSize int) ([]model.FlowTask, int64, error) {
	var list []model.FlowTask
	var total int64

	db := r.DB(ctx).Model(&model.FlowTask{}).Where("assignee_id = ? AND status != ?", userID, model.TaskStatusPending)
	db.Count(&total)
	err := db.Order("action_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ============= 流程日志 =============

// CreateFlowLog 创建流程日志
func (r *FlowRepository) CreateFlowLog(ctx context.Context, log *model.FlowLog) error {
	return r.DB(ctx).Create(log).Error
}

// GetFlowLogs 获取流程日志
func (r *FlowRepository) GetFlowLogs(ctx context.Context, flowInstID uint) ([]model.FlowLog, error) {
	var logs []model.FlowLog
	err := r.DB(ctx).Where("flow_inst_id = ?", flowInstID).Order("created_at").Find(&logs).Error
	return logs, err
}

// ============= 抄送 =============

// CreateFlowCC 创建抄送记录
func (r *FlowRepository) CreateFlowCC(ctx context.Context, cc *model.FlowCC) error {
	return r.DB(ctx).Create(cc).Error
}

// GetMyCCList 获取抄送给我的流程
func (r *FlowRepository) GetMyCCList(ctx context.Context, userID uint, page, pageSize int) ([]model.FlowCC, int64, error) {
	var list []model.FlowCC
	var total int64

	db := r.DB(ctx).Model(&model.FlowCC{}).Where("user_id = ?", userID)
	db.Count(&total)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// MarkCCAsRead 标记抄送为已读
func (r *FlowRepository) MarkCCAsRead(ctx context.Context, id uint) error {
	now := time.Now()
	return r.DB(ctx).Model(&model.FlowCC{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read":   true,
		"read_time": &now,
	}).Error
}

