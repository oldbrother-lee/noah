package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// 流程状态常量
const (
	FlowStatusDraft     = "draft"     // 草稿
	FlowStatusPending   = "pending"   // 待审批
	FlowStatusRunning   = "running"   // 审批中
	FlowStatusApproved  = "approved"  // 已通过
	FlowStatusRejected  = "rejected"  // 已驳回
	FlowStatusCancelled = "cancelled" // 已撤销
	FlowStatusSuspended = "suspended" // 已挂起
)

// 节点类型常量
const (
	NodeTypeStart     = "start"     // 开始节点
	NodeTypeEnd       = "end"       // 结束节点
	NodeTypeApproval  = "approval"  // 审批节点
	NodeTypeCC        = "cc"        // 抄送节点
	NodeTypeCondition = "condition" // 条件节点
	NodeTypeParallel  = "parallel"  // 并行节点
	NodeTypeSerial    = "serial"    // 串行节点
)

// 审批人类型常量
const (
	ApproverTypeUser       = "user"        // 指定用户
	ApproverTypeRole       = "role"        // 指定角色
	ApproverTypeDeptLeader = "dept_leader" // 部门负责人
	ApproverTypeSuperior   = "superior"    // 上级主管
	ApproverTypeSelfSelect = "self_select" // 发起人自选
	ApproverTypeAuto       = "auto"        // 自动通过
)

// 多人审批模式
const (
	MultiModeAny        = "any"        // 或签（任一人通过即可）
	MultiModeAll        = "all"        // 会签（所有人通过）
	MultiModeSequential = "sequential" // 依次审批
)

// 驳回动作
const (
	RejectActionToStart = "to_start" // 驳回到发起人
	RejectActionToPrev  = "to_prev"  // 驳回到上一节点
	RejectActionToEnd   = "to_end"   // 直接结束
)

// 任务状态
const (
	TaskStatusPending     = "pending"     // 待处理
	TaskStatusApproved    = "approved"    // 已通过
	TaskStatusRejected    = "rejected"    // 已驳回
	TaskStatusTransferred = "transferred" // 已转办
	TaskStatusDelegated   = "delegated"   // 已委托
)

// FlowDefinition 流程定义
type FlowDefinition struct {
	gorm.Model
	Code        string         `gorm:"type:varchar(50);uniqueIndex;comment:'流程编码'" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null;comment:'流程名称'" json:"name"`
	Type        string         `gorm:"type:varchar(50);index;comment:'业务类型:order_ddl,order_dml,order_export'" json:"type"`
	Description string         `gorm:"type:varchar(500);comment:'流程描述'" json:"description"`
	Version     int            `gorm:"default:1;comment:'版本号'" json:"version"`
	Status      int8           `gorm:"default:1;comment:'状态:1启用,0禁用'" json:"status"`
	Config      datatypes.JSON `gorm:"type:json;comment:'流程配置JSON'" json:"config"`
}

func (FlowDefinition) TableName() string {
	return "flow_definitions"
}

// FlowNode 流程节点
type FlowNode struct {
	gorm.Model
	FlowDefID     uint           `gorm:"index;comment:'流程定义ID'" json:"flow_def_id"`
	NodeCode      string         `gorm:"type:varchar(50);comment:'节点编码'" json:"node_code"`
	NodeName      string         `gorm:"type:varchar(100);comment:'节点名称'" json:"node_name"`
	NodeType      string         `gorm:"type:varchar(20);comment:'节点类型'" json:"node_type"`
	Sort          int            `gorm:"default:0;comment:'顺序'" json:"sort"`
	ApproverType  string         `gorm:"type:varchar(20);comment:'审批人类型'" json:"approver_type"`
	ApproverIDs   string         `gorm:"type:varchar(500);comment:'审批人ID列表,逗号分隔'" json:"approver_ids"`
	MultiMode     string         `gorm:"type:varchar(20);default:'any';comment:'多人审批模式'" json:"multi_mode"`
	RejectAction  string         `gorm:"type:varchar(20);default:'to_start';comment:'驳回动作'" json:"reject_action"`
	TimeoutHours  int            `gorm:"default:0;comment:'超时时间(小时),0表示不限'" json:"timeout_hours"`
	TimeoutAction string         `gorm:"type:varchar(20);comment:'超时动作:auto_pass,auto_reject,notify'" json:"timeout_action"`
	NextNodeCode  string         `gorm:"type:varchar(50);comment:'下一节点编码'" json:"next_node_code"`
	Conditions    datatypes.JSON `gorm:"type:json;comment:'条件配置JSON'" json:"conditions"`
}

func (FlowNode) TableName() string {
	return "flow_nodes"
}

// FlowInstance 流程实例
type FlowInstance struct {
	gorm.Model
	FlowDefID       uint           `gorm:"index;comment:'流程定义ID'" json:"flow_def_id"`
	FlowCode        string         `gorm:"type:varchar(50);index;comment:'流程编码'" json:"flow_code"`
	BusinessType    string         `gorm:"type:varchar(50);index;comment:'业务类型'" json:"business_type"`
	BusinessID      string         `gorm:"type:varchar(50);index;comment:'业务ID'" json:"business_id"`
	Title           string         `gorm:"type:varchar(200);comment:'流程标题'" json:"title"`
	InitiatorID     uint           `gorm:"index;comment:'发起人ID'" json:"initiator_id"`
	Initiator       string         `gorm:"type:varchar(50);comment:'发起人用户名'" json:"initiator"`
	Status          string         `gorm:"type:varchar(20);index;default:'pending';comment:'状态'" json:"status"`
	CurrentNodeID   uint           `gorm:"index;comment:'当前节点ID'" json:"current_node_id"`
	CurrentNodeCode string         `gorm:"type:varchar(50);comment:'当前节点编码'" json:"current_node_code"`
	StartTime       time.Time      `gorm:"comment:'开始时间'" json:"start_time"`
	EndTime         *time.Time     `gorm:"comment:'结束时间'" json:"end_time"`
	Variables       datatypes.JSON `gorm:"type:json;comment:'流程变量'" json:"variables"`
}

func (FlowInstance) TableName() string {
	return "flow_instances"
}

// FlowTask 审批任务
type FlowTask struct {
	gorm.Model
	FlowInstID       uint       `gorm:"index;comment:'流程实例ID'" json:"flow_inst_id"`
	FlowNodeID       uint       `gorm:"index;comment:'流程节点ID'" json:"flow_node_id"`
	NodeCode         string     `gorm:"type:varchar(50);comment:'节点编码'" json:"node_code"`
	NodeName         string     `gorm:"type:varchar(100);comment:'节点名称'" json:"node_name"`
	AssigneeID       uint       `gorm:"index;comment:'审批人ID'" json:"assignee_id"`
	Assignee         string     `gorm:"type:varchar(50);comment:'审批人用户名'" json:"assignee"`
	AssigneeNickname string     `gorm:"type:varchar(50);comment:'审批人昵称'" json:"assignee_nickname"`
	Status           string     `gorm:"type:varchar(20);index;default:'pending';comment:'状态'" json:"status"`
	Action           string     `gorm:"type:varchar(20);comment:'动作:approve,reject,transfer,delegate'" json:"action"`
	Comment          string     `gorm:"type:varchar(1000);comment:'审批意见'" json:"comment"`
	ActionTime       *time.Time `gorm:"comment:'处理时间'" json:"action_time"`
	DueTime          *time.Time `gorm:"comment:'截止时间'" json:"due_time"`
	TransferToID     uint       `gorm:"comment:'转办给用户ID'" json:"transfer_to_id"`
	TransferTo       string     `gorm:"type:varchar(50);comment:'转办给用户名'" json:"transfer_to"`
	DelegateToID     uint       `gorm:"comment:'委托给用户ID'" json:"delegate_to_id"`
	DelegateTo       string     `gorm:"type:varchar(50);comment:'委托给用户名'" json:"delegate_to"`
}

func (FlowTask) TableName() string {
	return "flow_tasks"
}

// FlowLog 流程日志
type FlowLog struct {
	gorm.Model
	FlowInstID uint   `gorm:"index;comment:'流程实例ID'" json:"flow_inst_id"`
	FlowNodeID uint   `gorm:"comment:'流程节点ID'" json:"flow_node_id"`
	NodeCode   string `gorm:"type:varchar(50);comment:'节点编码'" json:"node_code"`
	NodeName   string `gorm:"type:varchar(100);comment:'节点名称'" json:"node_name"`
	OperatorID uint   `gorm:"comment:'操作人ID'" json:"operator_id"`
	Operator   string `gorm:"type:varchar(50);comment:'操作人用户名'" json:"operator"`
	Action     string `gorm:"type:varchar(50);comment:'动作描述'" json:"action"`
	Comment    string `gorm:"type:varchar(1000);comment:'备注'" json:"comment"`
	Duration   int64  `gorm:"comment:'耗时(秒)'" json:"duration"`
}

func (FlowLog) TableName() string {
	return "flow_logs"
}

// FlowCC 抄送记录
type FlowCC struct {
	gorm.Model
	FlowInstID uint       `gorm:"index;comment:'流程实例ID'" json:"flow_inst_id"`
	UserID     uint       `gorm:"index;comment:'用户ID'" json:"user_id"`
	Username   string     `gorm:"type:varchar(50);comment:'用户名'" json:"username"`
	IsRead     bool       `gorm:"default:false;comment:'是否已读'" json:"is_read"`
	ReadTime   *time.Time `gorm:"comment:'阅读时间'" json:"read_time"`
}

func (FlowCC) TableName() string {
	return "flow_cc"
}
