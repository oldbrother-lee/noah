package insight

import (
	"context"
	"go-noah/internal/model/insight"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============ 工单管理 ============

// OrderWithInstance 工单记录（包含实例名称和环境名称）
type OrderWithInstance struct {
	insight.OrderRecord
	InstanceName      string `gorm:"column:instance_name" json:"instance_name"`
	EnvironmentName   string `gorm:"column:environment_name" json:"environment"`          // 使用 environment 作为 JSON 字段名，覆盖 OrderRecord 中的 environment ID
	ApplicantNickname string `gorm:"column:applicant_nickname" json:"applicant_nickname"` // 申请人昵称（通过 JOIN 查询获取）
}

// GetOrders 获取工单列表
func (r *InsightRepository) GetOrders(ctx context.Context, params *OrderQueryParams) ([]OrderWithInstance, int64, error) {
	var ordersWithInstance []OrderWithInstance
	var total int64

	// 使用 JOIN 查询，关联 db_configs 表获取实例名称，关联 db_environments 表获取环境名称，关联 admin_users 表获取申请人昵称
	query := r.DB(ctx).Table("order_records a").
		Select("a.*, CONCAT(COALESCE(b.hostname, ''), ':', COALESCE(b.port, 0)) as instance_name, COALESCE(c.name, '') as environment_name, COALESCE(d.nickname, '') as applicant_nickname").
		Joins("LEFT JOIN db_configs b ON a.instance_id = b.instance_id").
		Joins("LEFT JOIN db_environments c ON a.environment = c.id").
		Joins("LEFT JOIN admin_users d ON a.applicant = d.username")

	// 应用过滤条件
	if params.Applicant != "" {
		query = query.Where("a.applicant = ?", params.Applicant)
	}
	if params.Progress != "" {
		query = query.Where("a.progress = ?", params.Progress)
	}
	if params.Environment > 0 {
		query = query.Where("a.environment = ?", params.Environment)
	}
	if params.SQLType != "" {
		query = query.Where("a.sql_type = ?", params.SQLType)
	}
	if params.DBType != "" {
		query = query.Where("a.db_type = ?", params.DBType)
	}
	if params.Title != "" {
		query = query.Where("a.title LIKE ?", "%"+params.Title+"%")
	}

	// 计算总数（需要单独查询，因为 JOIN 会影响计数）
	countQuery := r.DB(ctx).Model(&insight.OrderRecord{})
	if params.Applicant != "" {
		countQuery = countQuery.Where("applicant = ?", params.Applicant)
	}
	if params.Progress != "" {
		countQuery = countQuery.Where("progress = ?", params.Progress)
	}
	if params.Environment > 0 {
		countQuery = countQuery.Where("environment = ?", params.Environment)
	}
	if params.SQLType != "" {
		countQuery = countQuery.Where("sql_type = ?", params.SQLType)
	}
	if params.DBType != "" {
		countQuery = countQuery.Where("db_type = ?", params.DBType)
	}
	if params.Title != "" {
		countQuery = countQuery.Where("title LIKE ?", "%"+params.Title+"%")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("a.created_at DESC").Offset(offset).Limit(params.PageSize).Scan(&ordersWithInstance).Error; err != nil {
		return nil, 0, err
	}

	return ordersWithInstance, total, nil
}

// OrderQueryParams 工单查询参数
type OrderQueryParams struct {
	Page        int
	PageSize    int
	Applicant   string
	Progress    string
	Environment int
	SQLType     string
	DBType      string
	Title       string
}

// GetOrderByID 根据OrderID获取工单（包含实例名称和环境名称）
func (r *InsightRepository) GetOrderByID(ctx context.Context, orderID string) (*OrderWithInstance, error) {
	var orderWithInstance OrderWithInstance
	// 使用 Model 和 Preload 或者直接查询 order_records 表，然后手动设置关联字段
	// 先查询工单记录
	var orderRecord insight.OrderRecord
	if err := r.DB(ctx).Where("order_id = ?", orderID).First(&orderRecord).Error; err != nil {
		return nil, err
	}

	// 设置工单记录到 OrderWithInstance
	orderWithInstance.OrderRecord = orderRecord

	// 查询实例名称和环境名称
	var instanceName, environmentName string
	r.DB(ctx).Table("db_configs").
		Select("CONCAT(COALESCE(hostname, ''), ':', COALESCE(port, 0))").
		Where("instance_id = ?", orderRecord.InstanceID).
		Scan(&instanceName)

	r.DB(ctx).Table("db_environments").
		Select("COALESCE(name, '')").
		Where("id = ?", orderRecord.Environment).
		Scan(&environmentName)

	orderWithInstance.InstanceName = instanceName
	orderWithInstance.EnvironmentName = environmentName

	return &orderWithInstance, nil
}

// CreateOrder 创建工单
func (r *InsightRepository) CreateOrder(ctx context.Context, order *insight.OrderRecord) error {
	return r.DB(ctx).Create(order).Error
}

// UpdateOrder 更新工单
func (r *InsightRepository) UpdateOrder(ctx context.Context, order *insight.OrderRecord) error {
	return r.DB(ctx).Save(order).Error
}

// UpdateOrderFields 更新工单的指定字段
func (r *InsightRepository) UpdateOrderFields(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.DB(ctx).Model(&insight.OrderRecord{}).Where("order_id = ?", orderID).Updates(updates).Error
}

// UpdateOrderProgress 更新工单进度
func (r *InsightRepository) UpdateOrderProgress(ctx context.Context, orderID string, progress insight.Progress) error {
	return r.DB(ctx).Model(&insight.OrderRecord{}).
		Where("order_id = ?", orderID).
		Update("progress", progress).Error
}

// ============ 工单任务管理 ============

// GetOrderTasks 获取工单的任务列表
func (r *InsightRepository) GetOrderTasks(ctx context.Context, orderID string) ([]insight.OrderTask, error) {
	var tasks []insight.OrderTask
	if err := r.DB(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID 根据TaskID获取任务
func (r *InsightRepository) GetTaskByID(ctx context.Context, taskID string) (*insight.OrderTask, error) {
	var task insight.OrderTask
	if err := r.DB(ctx).Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateOrderTasks 批量创建任务
func (r *InsightRepository) CreateOrderTasks(ctx context.Context, tasks []insight.OrderTask) error {
	return r.DB(ctx).CreateInBatches(tasks, 100).Error
}

// UpdateTaskProgress 更新任务进度
func (r *InsightRepository) UpdateTaskProgress(ctx context.Context, taskID string, progress insight.TaskProgress, result []byte) error {
	updates := map[string]interface{}{
		"progress": progress,
	}
	if result != nil {
		updates["result"] = result
	}
	return r.DB(ctx).Model(&insight.OrderTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

// CheckTasksProgressIsDoing 检查工单是否有任务正在执行中
func (r *InsightRepository) CheckTasksProgressIsDoing(ctx context.Context, orderID string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&insight.OrderTask{}).
		Where("order_id = ? AND progress = ?", orderID, insight.TaskProgressExecuting).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil // 返回 true 表示没有执行中的任务
}

// CheckTasksProgressIsPause 检查工单是否有已暂停的任务
func (r *InsightRepository) CheckTasksProgressIsPause(ctx context.Context, orderID string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&insight.OrderTask{}).
		Where("order_id = ? AND progress = ?", orderID, insight.TaskProgressPaused).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil // 返回 true 表示没有已暂停的任务
}

// UpdateOrderExecuteResult 更新工单执行结果
func (r *InsightRepository) UpdateOrderExecuteResult(ctx context.Context, orderID string, result string) error {
	return r.DB(ctx).Model(&insight.OrderRecord{}).
		Where("order_id = ?", orderID).
		Update("execute_result", result).Error
}

// CheckAllTasksCompleted 检查所有任务是否都已完成
func (r *InsightRepository) CheckAllTasksCompleted(ctx context.Context, orderID string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&insight.OrderTask{}).
		Where("order_id = ? AND progress != ?", orderID, insight.TaskProgressCompleted).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil // 返回 true 表示所有任务都已完成
}

// UpdateTaskAndOrderProgress 使用事务同时更新任务和工单状态
func (r *InsightRepository) UpdateTaskAndOrderProgress(ctx context.Context, taskID string, orderID string, taskProgress insight.TaskProgress, orderProgress insight.Progress) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新任务状态
		if err := tx.Model(&insight.OrderTask{}).
			Where("task_id = ?", taskID).
			Update("progress", taskProgress).Error; err != nil {
			return err
		}
		// 更新工单状态
		if err := tx.Model(&insight.OrderRecord{}).
			Where("order_id = ?", orderID).
			Update("progress", orderProgress).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteOrder 删除工单
func (r *InsightRepository) DeleteOrder(ctx context.Context, orderID string) error {
	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}
	return r.DB(ctx).Where("order_id = ?", orderUUID).Delete(&insight.OrderRecord{}).Error
}

// DeleteOrderTasks 删除工单的所有任务
func (r *InsightRepository) DeleteOrderTasks(ctx context.Context, orderID string) error {
	return r.DB(ctx).Where("order_id = ?", orderID).Delete(&insight.OrderTask{}).Error
}

// ============ 操作日志管理 ============

// CreateOpLog 创建操作日志
func (r *InsightRepository) CreateOpLog(ctx context.Context, log *insight.OrderOpLog) error {
	return r.DB(ctx).Create(log).Error
}

// GetOpLogs 获取工单的操作日志
func (r *InsightRepository) GetOpLogs(ctx context.Context, orderID string) ([]insight.OrderOpLog, error) {
	var logs []insight.OrderOpLog
	if err := r.DB(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ============ 消息记录管理 ============

// CreateOrderMessage 创建消息记录
func (r *InsightRepository) CreateOrderMessage(ctx context.Context, msg *insight.OrderMessage) error {
	return r.DB(ctx).Create(msg).Error
}

// GetOrderMessages 获取工单的消息记录
func (r *InsightRepository) GetOrderMessages(ctx context.Context, orderID uuid.UUID) ([]insight.OrderMessage, error) {
	var messages []insight.OrderMessage
	if err := r.DB(ctx).Where("order_id = ?", orderID).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetOrdersByStatusAndScheduleTime 获取指定状态且有定时执行时间的工单
func (r *InsightRepository) GetOrdersByStatusAndScheduleTime(ctx context.Context, progress insight.Progress) ([]insight.OrderRecord, error) {
	var orders []insight.OrderRecord
	if err := r.DB(ctx).
		Where("progress = ?", progress).
		Where("schedule_time IS NOT NULL").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrdersPendingSchedulerRegistration 获取需要注册到调度器的定时工单
// 条件：已批准 + 有定时执行时间 + 未注册到调度器
// 优化：限制查询数量，只查询必要的字段，添加时间范围限制
func (r *InsightRepository) GetOrdersPendingSchedulerRegistration(ctx context.Context, limit int) ([]insight.OrderRecord, error) {
	if limit <= 0 {
		limit = 100 // 默认每次最多处理100个工单
	}

	var orders []insight.OrderRecord
	now := time.Now()
	// 只查询必要的字段，减少数据传输量
	// 限制查询范围：查询过去24小时到未来30天内的定时工单
	// 注意：已过期的工单也会被查询到，但在处理时会跳过执行，只标记为已注册
	// 使用复合索引：progress + scheduler_registered + schedule_time
	if err := r.DB(ctx).
		Select("id, order_id, progress, schedule_time, applicant, executor").
		Where("progress = ?", insight.ProgressApproved).
		Where("scheduler_registered = ?", false).
		Where("schedule_time IS NOT NULL").
		Where("schedule_time >= ?", now.Add(-24*time.Hour)).   // 查询过去24小时内的（处理服务器重启的情况，已过期的会被标记但不执行）
		Where("schedule_time <= ?", now.Add(30*24*time.Hour)). // 最多查询未来30天内的
		Order("schedule_time ASC").                            // 按时间排序，优先处理即将执行的
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// MarkSchedulerRegistered 标记工单已注册到调度器
func (r *InsightRepository) MarkSchedulerRegistered(ctx context.Context, orderID string) error {
	return r.DB(ctx).Model(&insight.OrderRecord{}).
		Where("order_id = ?", orderID).
		Update("scheduler_registered", true).Error
}

// MarkSchedulerUnregistered 标记工单未注册到调度器（用于清理）
func (r *InsightRepository) MarkSchedulerUnregistered(ctx context.Context, orderID string) error {
	return r.DB(ctx).Model(&insight.OrderRecord{}).
		Where("order_id = ?", orderID).
		Update("scheduler_registered", false).Error
}

// ============ 事务操作 ============

// WithTransaction 事务包装
func (r *InsightRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB(ctx).Transaction(fn)
}
