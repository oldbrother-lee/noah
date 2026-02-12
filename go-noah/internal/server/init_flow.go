package server

import (
	"go-noah/internal/model"
	"go-noah/pkg/global"

	"go.uber.org/zap"
)

// InitDefaultFlowDefinitions 初始化默认流程定义
func InitDefaultFlowDefinitions() error {
	// DDL工单流程
	if err := initFlowDefinition("order_ddl", "order_ddl_default", "DDL工单审批流程", "DDL工单默认审批流程"); err != nil {
		return err
	}

	// DML工单流程
	if err := initFlowDefinition("order_dml", "order_dml_default", "DML工单审批流程", "DML工单默认审批流程"); err != nil {
		return err
	}

	// EXPORT工单流程
	if err := initFlowDefinition("order_export", "order_export_default", "EXPORT工单审批流程", "EXPORT工单默认审批流程"); err != nil {
		return err
	}

	return nil
}

// initFlowDefinition 初始化单个流程定义
func initFlowDefinition(flowType, code, name, description string) error {
	// 检查是否已存在
	var existing model.FlowDefinition
	if err := global.DB.Where("type = ? AND status = 1", flowType).First(&existing).Error; err == nil {
		// 已存在，跳过
		global.Logger.Info("流程定义已存在，跳过初始化", zap.String("type", flowType))
		return nil
	}

	// 创建流程定义
	flow := &model.FlowDefinition{
		Code:        code,
		Name:        name,
		Type:        flowType,
		Description: description,
		Version:     1,
		Status:      1,
	}

	if err := global.DB.Create(flow).Error; err != nil {
		global.Logger.Error("创建流程定义失败", zap.Error(err), zap.String("type", flowType))
		return err
	}

	// 创建流程节点
	nodes := []model.FlowNode{
		{
			FlowDefID:    flow.ID,
			NodeCode:     "start",
			NodeName:     "开始",
			NodeType:     model.NodeTypeStart,
			Sort:         1,
			NextNodeCode: "approval_1",
		},
		{
			FlowDefID:     flow.ID,
			NodeCode:      "approval_1",
			NodeName:      "审批节点",
			NodeType:      model.NodeTypeApproval,
			Sort:          2,
			ApproverType:  model.ApproverTypeSelfSelect, // 发起人自选
			MultiMode:     model.MultiModeAll,           // 会签（所有人通过）
			RejectAction:  model.RejectActionToStart,     // 驳回到发起人
			NextNodeCode:  "end",
		},
		{
			FlowDefID:    flow.ID,
			NodeCode:     "end",
			NodeName:     "结束",
			NodeType:     model.NodeTypeEnd,
			Sort:         3,
		},
	}

	if err := global.DB.Create(&nodes).Error; err != nil {
		global.Logger.Error("创建流程节点失败", zap.Error(err), zap.String("type", flowType))
		// 删除已创建的流程定义
		global.DB.Delete(&model.FlowDefinition{}, flow.ID)
		return err
	}

	global.Logger.Info("初始化流程定义成功", zap.String("type", flowType), zap.String("code", code))
	return nil
}
