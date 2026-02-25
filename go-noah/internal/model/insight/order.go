package insight

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SQLType SQL类型
type SQLType string

const (
	SQLTypeDML    SQLType = "DML"
	SQLTypeDDL    SQLType = "DDL"
	SQLTypeExport SQLType = "EXPORT"
)

// DDLExecutionMode DDL 执行方式：gh-ost 或直接 ALTER，提交时用户必选，单工单仅一种
type DDLExecutionMode string

const (
	DDLExecutionModeGhost  DDLExecutionMode = "ghost"  // gh-ost 在线变更，表需有主键或唯一键
	DDLExecutionModeDirect DDLExecutionMode = "direct" // 直接 ALTER，可能锁表
)

// Progress 工单进度
type Progress string

const (
	ProgressPending   Progress = "待审核"
	ProgressRejected  Progress = "已驳回"
	ProgressApproved  Progress = "已批准"
	ProgressExecuting Progress = "执行中"
	ProgressClosed    Progress = "已关闭"
	ProgressCompleted Progress = "已完成"
	ProgressReviewed  Progress = "已复核"
)

// TaskProgress 任务进度
type TaskProgress string

const (
	TaskProgressPending   TaskProgress = "未执行"
	TaskProgressExecuting TaskProgress = "执行中"
	TaskProgressCompleted TaskProgress = "已完成"
	TaskProgressFailed    TaskProgress = "已失败"
	TaskProgressPaused    TaskProgress = "已暂停"
)

// ExportFileFormat 导出文件格式
type ExportFileFormat string

const (
	ExportFormatXLSX ExportFileFormat = "XLSX"
	ExportFormatCSV  ExportFileFormat = "CSV"
)

// OrderRecord 工单记录
type OrderRecord struct {
	gorm.Model
	Title               string           `gorm:"type:varchar(128);not null;default:'';comment:工单标题;index:idx_title" json:"title"`
	OrderID             uuid.UUID        `gorm:"type:char(36);comment:工单ID;uniqueIndex:uniq_order_id" json:"order_id"`
	HookOrderID         uuid.UUID        `gorm:"type:char(36);comment:HOOK源工单ID;index:idx_hook_order_id" json:"hook_order_id"`
	Remark              string           `gorm:"type:varchar(1024);not null;default:'';comment:工单备注" json:"remark"`
	IsRestrictAccess    bool             `gorm:"type:tinyint(1);not null;default:0;comment:是否限制访问" json:"is_restrict_access"`
	DBType              DbType           `gorm:"type:varchar(20);default:'MySQL';comment:DB类型" json:"db_type"`
	SQLType             SQLType          `gorm:"type:varchar(20);default:'DML';comment:SQL类型" json:"sql_type"`
	Environment         int              `gorm:"type:int;null;default:null;comment:环境;index" json:"-"` // 不返回环境ID，使用environment_name代替
	Applicant           string           `gorm:"type:varchar(32);not null;default:'';comment:申请人;index" json:"applicant"`
	Organization        string           `gorm:"type:varchar(256);not null;default:'';index;comment:组织" json:"organization"`
	Approver            datatypes.JSON   `gorm:"type:json;null;default:null;comment:工单审核人" json:"approver"`
	Executor            datatypes.JSON   `gorm:"type:json;null;default:null;comment:工单执行人" json:"executor"`
	Reviewer            datatypes.JSON   `gorm:"type:json;null;default:null;comment:工单复核人" json:"reviewer"`
	CC                  datatypes.JSON   `gorm:"type:json;null;default:null;comment:工单抄送人" json:"cc"`
	InstanceID          uuid.UUID        `gorm:"type:char(36);comment:关联db_configs的instance_id;index" json:"instance_id"`
	Schema              string           `gorm:"type:varchar(128);not null;default:'';comment:库名" json:"schema"`
	Progress            Progress         `gorm:"type:varchar(20);default:'待审核';comment:工单进度" json:"progress"`
	ExecuteResult       string           `gorm:"type:varchar(32);default:'';comment:执行结果(success,error,warning)" json:"execute_result"`
	ScheduleTime        *time.Time       `gorm:"type:datetime;null;default:null;comment:计划执行时间" json:"schedule_time"`
	FixVersion          string           `gorm:"type:varchar(128);not null;default:'';comment:上线版本;index" json:"fix_version"`
	Content             string           `gorm:"type:text;null;comment:工单内容" json:"content"`
	ExportFileFormat    ExportFileFormat `gorm:"type:varchar(10);default:'XLSX';comment:导出文件格式" json:"export_file_format"`
	FlowInstanceID      uint             `gorm:"index;comment:'关联流程实例ID'" json:"flow_instance_id"`
	GhostOkToDropTable  bool             `gorm:"type:tinyint(1);not null;default:0;comment:gh-ost执行成功后自动删除旧表" json:"ghost_ok_to_drop_table"`
	SchedulerRegistered bool             `gorm:"type:tinyint(1);not null;default:0;comment:定时任务是否已注册到调度器;index" json:"scheduler_registered"`
	GenerateRollback    *bool            `gorm:"type:tinyint(1);default:1;comment:DML工单是否生成回滚语句(仅DML有效)" json:"generate_rollback"`       // 用指针避免 GORM 忽略 false
	DDLExecutionMode    DDLExecutionMode `gorm:"type:varchar(20);default:'';comment:DDL执行方式(ghost/direct)，仅DDL有效" json:"ddl_execution_mode"` // ghost=gh-ost, direct=直接ALTER
}

func (OrderRecord) TableName() string {
	return "order_records"
}

func (o *OrderRecord) BeforeCreate(tx *gorm.DB) (err error) {
	o.OrderID, _ = uuid.NewUUID()
	return
}

// OrderOpLog 工单操作日志
type OrderOpLog struct {
	gorm.Model
	Username string    `gorm:"type:varchar(32);not null;index:idx_username;comment:操作用户" json:"username"`
	OrderID  uuid.UUID `gorm:"type:char(36);comment:工单ID;index:idx_order_id" json:"order_id"`
	Msg      string    `gorm:"type:varchar(1024);null;comment:操作信息" json:"msg"`
}

func (OrderOpLog) TableName() string {
	return "order_op_logs"
}

// OrderTask 工单任务
type OrderTask struct {
	gorm.Model
	OrderID  uuid.UUID      `gorm:"type:char(36);comment:关联order_records的order_id;index" json:"order_id"`
	TaskID   uuid.UUID      `gorm:"type:char(36);comment:任务ID;index" json:"task_id"`
	DBType   DbType         `gorm:"type:varchar(20);default:'MySQL';comment:DB类型" json:"db_type"`
	SQLType  SQLType        `gorm:"type:varchar(20);default:'DML';comment:SQL类型" json:"sql_type"`
	Executor string         `gorm:"type:varchar(128);null;default:null;comment:任务执行人" json:"executor"`
	SQL      string         `gorm:"type:text;null;comment:SQL语句" json:"sql"`
	Progress TaskProgress   `gorm:"type:varchar(20);default:'未执行';comment:进度" json:"progress"`
	Result   datatypes.JSON `gorm:"type:json;null;default:null;comment:执行结果" json:"result"`
}

func (OrderTask) TableName() string {
	return "order_tasks"
}

func (o *OrderTask) BeforeCreate(tx *gorm.DB) (err error) {
	o.TaskID, _ = uuid.NewUUID()
	return
}

// OrderMessage 消息推送记录
type OrderMessage struct {
	gorm.Model
	OrderID  uuid.UUID      `gorm:"type:char(36);comment:关联order_records的order_id;index" json:"order_id"`
	Receiver datatypes.JSON `gorm:"type:json;null;default:null;comment:接收消息的用户" json:"receiver"`
	Response string         `gorm:"type:text;null;comment:第三方返回的响应" json:"response"`
}

func (OrderMessage) TableName() string {
	return "order_messages"
}
