package api

import "time"

// ============= 流程定义 API =============

// FlowDefinitionItem 流程定义项
type FlowDefinitionItem struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Version     int    `json:"version"`
	Status      int8   `json:"status"`
}

// FlowDefinitionListData 流程定义列表数据
type FlowDefinitionListData struct {
	List []FlowDefinitionItem `json:"list"`
}

// FlowNodeItem 流程节点项
type FlowNodeItem struct {
	ID            uint   `json:"id"`
	NodeCode      string `json:"nodeCode"`
	NodeName      string `json:"nodeName"`
	NodeType      string `json:"nodeType"`
	Sort          int    `json:"sort"`
	ApproverType  string `json:"approverType"`
	ApproverIDs   string `json:"approverIds"`
	MultiMode     string `json:"multiMode"`
	RejectAction  string `json:"rejectAction"`
	TimeoutHours  int    `json:"timeoutHours"`
	TimeoutAction string `json:"timeoutAction"`
	NextNodeCode  string `json:"nextNodeCode"`
}

// FlowDefinitionDetail 流程定义详情
type FlowDefinitionDetail struct {
	ID          uint           `json:"id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Version     int            `json:"version"`
	Status      int8           `json:"status"`
	Nodes       []FlowNodeItem `json:"nodes"`
}

// CreateFlowDefinitionRequest 创建流程定义请求
type CreateFlowDefinitionRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

// UpdateFlowDefinitionRequest 更新流程定义请求
type UpdateFlowDefinitionRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

// DeleteFlowDefinitionRequest 删除流程定义请求
type DeleteFlowDefinitionRequest struct {
	ID uint `form:"id" binding:"required"`
}

// GetFlowDefinitionRequest 获取流程定义请求
type GetFlowDefinitionRequest struct {
	ID uint `form:"id" binding:"required"`
}

// SaveFlowNodesRequest 保存流程节点请求
type SaveFlowNodesRequest struct {
	FlowDefID uint                `json:"flowDefId" binding:"required"`
	Nodes     []FlowNodeItemInput `json:"nodes"`
}

// FlowNodeItemInput 流程节点输入
type FlowNodeItemInput struct {
	NodeCode      string `json:"nodeCode" binding:"required"`
	NodeName      string `json:"nodeName" binding:"required"`
	NodeType      string `json:"nodeType" binding:"required"`
	Sort          int    `json:"sort"`
	ApproverType  string `json:"approverType"`
	ApproverIDs   string `json:"approverIds"`
	MultiMode     string `json:"multiMode"`
	RejectAction  string `json:"rejectAction"`
	TimeoutHours  int    `json:"timeoutHours"`
	TimeoutAction string `json:"timeoutAction"`
	NextNodeCode  string `json:"nextNodeCode"`
}

// ============= 流程实例 API =============

// StartFlowRequest 发起流程请求
type StartFlowRequest struct {
	BusinessType string `json:"businessType" binding:"required"` // 业务类型：order_ddl, order_dml, order_export
	BusinessID   string `json:"businessId" binding:"required"`   // 业务ID
	Title        string `json:"title" binding:"required"`        // 流程标题
	InitiatorID  uint   `json:"initiatorId"`                     // 发起人ID
	Initiator    string `json:"initiator"`                       // 发起人用户名
}

// StartFlowResponse 发起流程响应
type StartFlowResponse struct {
	FlowInstanceID uint `json:"flowInstanceId"`
}

// FlowInstanceDetail 流程实例详情
type FlowInstanceDetail struct {
	ID              uint           `json:"id"`
	FlowCode        string         `json:"flowCode"`
	BusinessType    string         `json:"businessType"`
	BusinessID      string         `json:"businessId"`
	Title           string         `json:"title"`
	Initiator       string         `json:"initiator"`
	Status          string         `json:"status"`
	CurrentNodeCode string         `json:"currentNodeCode"`
	StartTime       time.Time      `json:"startTime"`
	EndTime         *time.Time     `json:"endTime"`
	Tasks           []FlowTaskItem `json:"tasks"`
	Logs            []FlowLogItem  `json:"logs"`
	Nodes           []FlowNodeItem `json:"nodes"` // 流程节点信息（用于显示流程步骤）
}

// GetFlowInstanceRequest 获取流程实例请求
type GetFlowInstanceRequest struct {
	ID uint `form:"id" binding:"required"`
}

// ============= 审批任务 API =============

// FlowTaskItem 审批任务项
type FlowTaskItem struct {
	ID               uint       `json:"id"`
	FlowInstID       uint       `json:"flowInstId"`
	NodeCode         string     `json:"nodeCode"`
	NodeName         string     `json:"nodeName"`
	Assignee         string     `json:"assignee"`
	AssigneeNickname string     `json:"assignee_nickname"`
	Status           string     `json:"status"`
	Action           string     `json:"action"`
	Comment          string     `json:"comment"`
	ActionTime       *time.Time `json:"actionTime"`
	DueTime          *time.Time `json:"dueTime"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// FlowTaskListData 审批任务列表数据
type FlowTaskListData struct {
	List  []FlowTaskItem `json:"list"`
	Total int64          `json:"total"`
}

// GetMyPendingTasksRequest 获取我的待办任务请求
type GetMyPendingTasksRequest struct {
	Page     int `form:"page" binding:"required"`
	PageSize int `form:"pageSize" binding:"required"`
}

// ApproveTaskRequest 审批通过请求
type ApproveTaskRequest struct {
	TaskID     uint   `json:"taskId" binding:"required"`
	Comment    string `json:"comment"`
	OperatorID uint   `json:"operatorId"`
	Operator   string `json:"operator"`
}

// RejectTaskRequest 审批驳回请求
type RejectTaskRequest struct {
	TaskID     uint   `json:"taskId" binding:"required"`
	Comment    string `json:"comment" binding:"required"`
	OperatorID uint   `json:"operatorId"`
	Operator   string `json:"operator"`
}

// ============= 流程日志 API =============

// FlowLogItem 流程日志项
type FlowLogItem struct {
	ID        uint      `json:"id"`
	NodeCode  string    `json:"nodeCode"`
	NodeName  string    `json:"nodeName"`
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}
