package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-noah/api"
	"go-noah/internal/inspect/parser"
	"go-noah/internal/model/insight"
	"go-noah/internal/orders/executor"
	"go-noah/internal/repository"
	insightRepo "go-noah/internal/repository/insight"
	"go-noah/pkg/global"
	"go-noah/pkg/notifier"
	"strings"

	"go.uber.org/zap"
)

// InsightServiceApp 全局 Service 实例
var InsightServiceApp = new(InsightService)

// InsightService goInsight 功能的业务逻辑层
type InsightService struct{}

func (s *InsightService) getRepo() *insightRepo.InsightRepository {
	return insightRepo.NewInsightRepository(
		repository.NewRepository(global.Logger, global.DB, global.Enforcer),
		global.Logger,
		global.Enforcer,
	)
}

// ============ 环境管理 ============

func (s *InsightService) GetEnvironments(ctx context.Context) ([]insight.DBEnvironment, error) {
	return s.getRepo().GetEnvironments(ctx)
}

func (s *InsightService) CreateEnvironment(ctx context.Context, env *insight.DBEnvironment) error {
	return s.getRepo().CreateEnvironment(ctx, env)
}

func (s *InsightService) UpdateEnvironment(ctx context.Context, id uint, name string) error {
	env := &insight.DBEnvironment{}
	env.ID = id
	env.Name = name
	return s.getRepo().UpdateEnvironment(ctx, env)
}

func (s *InsightService) DeleteEnvironment(ctx context.Context, id uint) error {
	return s.getRepo().DeleteEnvironment(ctx, id)
}

// ============ 数据库配置管理 ============

func (s *InsightService) GetDBConfigs(ctx context.Context, useType insight.UseType, environment int) ([]insight.DBConfig, error) {
	return s.getRepo().GetDBConfigs(ctx, useType, environment)
}

func (s *InsightService) GetDBConfigByInstanceID(ctx context.Context, instanceID string) (*insight.DBConfig, error) {
	return s.getRepo().GetDBConfigByInstanceID(ctx, instanceID)
}

func (s *InsightService) CreateDBConfig(ctx context.Context, config *insight.DBConfig) error {
	// TODO: 密码加密存储
	return s.getRepo().CreateDBConfig(ctx, config)
}

func (s *InsightService) UpdateDBConfig(ctx context.Context, config *insight.DBConfig) error {
	// TODO: 密码加密存储
	return s.getRepo().UpdateDBConfig(ctx, config)
}

func (s *InsightService) UpdateDBConfigFields(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.getRepo().UpdateDBConfigFields(ctx, id, updates)
}

func (s *InsightService) DeleteDBConfig(ctx context.Context, id uint) error {
	return s.getRepo().DeleteDBConfig(ctx, id)
}

// ============ Schema 管理 ============

func (s *InsightService) GetSchemasByInstanceID(ctx context.Context, instanceID string) ([]insight.DBSchema, error) {
	return s.getRepo().GetSchemasByInstanceID(ctx, instanceID)
}

// ============ 组织管理 ============

func (s *InsightService) GetOrganizations(ctx context.Context) ([]insight.Organization, error) {
	return s.getRepo().GetOrganizations(ctx)
}

func (s *InsightService) CreateOrganization(ctx context.Context, org *insight.Organization) error {
	return s.getRepo().CreateOrganization(ctx, org)
}

func (s *InsightService) UpdateOrganization(ctx context.Context, org *insight.Organization) error {
	return s.getRepo().UpdateOrganization(ctx, org)
}

func (s *InsightService) DeleteOrganization(ctx context.Context, id uint64) error {
	return s.getRepo().DeleteOrganization(ctx, id)
}

func (s *InsightService) GetOrganizationByID(ctx context.Context, id uint64) (*insight.Organization, error) {
	return s.getRepo().GetOrganizationByID(ctx, id)
}

func (s *InsightService) GetOrganizationUsers(ctx context.Context, orgKey string) ([]insight.OrganizationUser, error) {
	return s.getRepo().GetOrganizationUsers(ctx, orgKey)
}

func (s *InsightService) BindOrganizationUser(ctx context.Context, ou *insight.OrganizationUser) error {
	return s.getRepo().BindOrganizationUser(ctx, ou)
}

func (s *InsightService) UnbindOrganizationUser(ctx context.Context, uid uint64) error {
	return s.getRepo().UnbindOrganizationUser(ctx, uid)
}

// ============ DAS 权限管理 ============

// GetUserAuthorizedSchemas 获取用户授权的所有 schemas
func (s *InsightService) GetUserAuthorizedSchemas(ctx context.Context, username string) ([]insight.UserAuthorizedSchema, error) {
	return s.getRepo().GetUserAuthorizedSchemas(ctx, username)
}

func (s *InsightService) GetUserSchemaPermissions(ctx context.Context, username string) ([]insight.DASUserSchemaPermission, error) {
	return s.getRepo().GetUserSchemaPermissions(ctx, username)
}

func (s *InsightService) CreateSchemaPermission(ctx context.Context, perm *insight.DASUserSchemaPermission) error {
	return s.getRepo().CreateSchemaPermission(ctx, perm)
}

func (s *InsightService) DeleteSchemaPermission(ctx context.Context, id uint) error {
	return s.getRepo().DeleteSchemaPermission(ctx, id)
}

func (s *InsightService) GetUserTablePermissions(ctx context.Context, username string) ([]insight.DASUserTablePermission, error) {
	return s.getRepo().GetUserTablePermissions(ctx, username)
}

// ============ 权限模板管理 ============

func (s *InsightService) GetPermissionTemplates(ctx context.Context) ([]insight.DASPermissionTemplate, error) {
	return s.getRepo().GetPermissionTemplates(ctx)
}

func (s *InsightService) GetPermissionTemplate(ctx context.Context, id uint) (*insight.DASPermissionTemplate, error) {
	return s.getRepo().GetPermissionTemplate(ctx, id)
}

func (s *InsightService) CreatePermissionTemplate(ctx context.Context, template *insight.DASPermissionTemplate) error {
	return s.getRepo().CreatePermissionTemplate(ctx, template)
}

func (s *InsightService) UpdatePermissionTemplate(ctx context.Context, template *insight.DASPermissionTemplate) error {
	return s.getRepo().UpdatePermissionTemplate(ctx, template)
}

func (s *InsightService) DeletePermissionTemplate(ctx context.Context, id uint) error {
	return s.getRepo().DeletePermissionTemplate(ctx, id)
}

// ============ 角色权限管理 ============

func (s *InsightService) GetRolePermissions(ctx context.Context, role string) ([]insight.DASRolePermission, error) {
	return s.getRepo().GetRolePermissions(ctx, role)
}

func (s *InsightService) CreateRolePermission(ctx context.Context, perm *insight.DASRolePermission) error {
	return s.getRepo().CreateRolePermission(ctx, perm)
}

func (s *InsightService) DeleteRolePermission(ctx context.Context, id uint) error {
	return s.getRepo().DeleteRolePermission(ctx, id)
}

func (s *InsightService) BatchCreateRolePermissions(ctx context.Context, perms []insight.DASRolePermission) error {
	return s.getRepo().BatchCreateRolePermissions(ctx, perms)
}

// ============ 用户权限管理（与角色同构：object/template，无 rule）============

func (s *InsightService) GetUserPermissions(ctx context.Context, username string) ([]insight.DASUserPermission, error) {
	return s.getRepo().GetUserPermissions(ctx, username)
}

func (s *InsightService) CreateUserPermission(ctx context.Context, perm *insight.DASUserPermission) error {
	return s.getRepo().CreateUserPermission(ctx, perm)
}

func (s *InsightService) DeleteUserPermission(ctx context.Context, id uint) error {
	return s.getRepo().DeleteUserPermission(ctx, id)
}

// ============ 权限查询 ============

func (s *InsightService) GetUserEffectivePermissions(ctx context.Context, username string) ([]insight.PermissionObject, error) {
	return s.getRepo().GetUserEffectivePermissions(ctx, username)
}

func (s *InsightService) ExpandRolePermissions(ctx context.Context, role string) ([]insight.PermissionObject, error) {
	return s.getRepo().ExpandRolePermissions(ctx, role)
}

func (s *InsightService) CreateTablePermission(ctx context.Context, perm *insight.DASUserTablePermission) error {
	return s.getRepo().CreateTablePermission(ctx, perm)
}

func (s *InsightService) DeleteTablePermission(ctx context.Context, id uint) error {
	return s.getRepo().DeleteTablePermission(ctx, id)
}

// ============ DAS 执行记录 ============

func (s *InsightService) CreateDASRecord(ctx context.Context, record *insight.DASRecord) error {
	return s.getRepo().CreateDASRecord(ctx, record)
}

func (s *InsightService) GetDASRecords(ctx context.Context, username string, page, pageSize int) ([]insight.DASRecord, int64, error) {
	return s.getRepo().GetDASRecords(ctx, username, page, pageSize)
}

// ============ DAS 收藏夹 ============

func (s *InsightService) GetFavorites(ctx context.Context, username string) ([]insight.DASFavorite, error) {
	return s.getRepo().GetFavorites(ctx, username)
}

func (s *InsightService) CreateFavorite(ctx context.Context, fav *insight.DASFavorite) error {
	return s.getRepo().CreateFavorite(ctx, fav)
}

func (s *InsightService) UpdateFavorite(ctx context.Context, fav *insight.DASFavorite) error {
	return s.getRepo().UpdateFavorite(ctx, fav)
}

func (s *InsightService) DeleteFavorite(ctx context.Context, id uint, username string) error {
	return s.getRepo().DeleteFavorite(ctx, id, username)
}

// ============ 工单管理 ============

func (s *InsightService) GetOrders(ctx context.Context, params *insightRepo.OrderQueryParams) ([]insightRepo.OrderWithInstance, int64, error) {
	return s.getRepo().GetOrders(ctx, params)
}

func (s *InsightService) GetOrderByID(ctx context.Context, orderID string) (*insightRepo.OrderWithInstance, error) {
	return s.getRepo().GetOrderByID(ctx, orderID)
}

func (s *InsightService) CreateOrder(ctx context.Context, order *insight.OrderRecord) error {
	return s.getRepo().CreateOrder(ctx, order)
}

func (s *InsightService) UpdateOrder(ctx context.Context, order *insight.OrderRecord) error {
	return s.getRepo().UpdateOrder(ctx, order)
}

func (s *InsightService) UpdateOrderFields(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return s.getRepo().UpdateOrderFields(ctx, orderID, updates)
}

func (s *InsightService) UpdateOrderProgress(ctx context.Context, orderID string, progress insight.Progress) error {
	// 先更新进度
	if err := s.getRepo().UpdateOrderProgress(ctx, orderID, progress); err != nil {
		return err
	}

	// 如果进度变为"已批准"，自动生成任务
	if progress == insight.ProgressApproved {
		// 检查任务是否已存在
		tasks, err := s.getRepo().GetOrderTasks(ctx, orderID)
		if err == nil && len(tasks) == 0 {
			// 任务不存在，自动生成
			_ = s.GenerateOrderTasks(ctx, orderID)
		}

		// 注意：定时任务注册由 TaskServer 定期扫描数据库来完成
		// HTTP 服务器只负责更新工单状态，不直接注册定时任务
		global.Logger.Info("工单审核通过，等待 TaskServer 注册定时任务",
			zap.String("order_id", orderID),
		)
	}

	// 如果进度变为"执行中"，清理调度器注册标记（由 TaskServer 负责移除定时任务）
	if progress == insight.ProgressExecuting {
		// 清理注册标记，TaskServer 扫描时会发现并移除定时任务
		_ = s.getRepo().MarkSchedulerUnregistered(ctx, orderID)
		global.Logger.Info("工单开始执行，清理调度器注册标记",
			zap.String("order_id", orderID),
		)
	}

	return nil
}

// GenerateOrderTasks 为工单生成任务
func (s *InsightService) GenerateOrderTasks(ctx context.Context, orderID string) error {
	// 获取工单信息
	order, err := s.getRepo().GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 检查任务是否已存在
	existingTasks, err := s.getRepo().GetOrderTasks(ctx, orderID)
	if err == nil && len(existingTasks) > 0 {
		// 任务已存在，跳过
		return nil
	}

	// 拆分 SQL
	sqls, err := s.splitSQLText(order.Content)
	if err != nil {
		return err
	}

	// 创建任务
	var tasks []insight.OrderTask
	for _, sql := range sqls {
		task := insight.OrderTask{
			OrderID:  order.OrderID,
			DBType:   order.DBType,
			SQLType:  order.SQLType,
			SQL:      sql,
			Progress: insight.TaskProgressPending,
		}
		tasks = append(tasks, task)
	}

	// 批量创建任务
	if len(tasks) > 0 {
		return s.getRepo().CreateOrderTasks(ctx, tasks)
	}

	return nil
}

// splitSQLText 拆分SQL文本（内部辅助方法）
func (s *InsightService) splitSQLText(sqltext string) ([]string, error) {
	// 使用 inspect parser 拆分 SQL
	audit, warns, err := parser.ParseSQL(sqltext)
	if err != nil {
		return nil, err
	}
	if len(warns) > 0 {
		// 有警告但不影响拆分
	}

	var sqls []string
	for _, stmt := range audit.TiStmt {
		sqls = append(sqls, stmt.Text())
	}
	return sqls, nil
}

// ============ 工单任务管理 ============

func (s *InsightService) GetOrderTasks(ctx context.Context, orderID string) ([]insight.OrderTask, error) {
	tasks, err := s.getRepo().GetOrderTasks(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	// 优化：移除 result 中的 rollback_sql，减少响应大小
	// 只保留 has_rollback_sql 标志位
	for i := range tasks {
		if len(tasks[i].Result) > 0 {
			var resultMap map[string]interface{}
			if err := json.Unmarshal(tasks[i].Result, &resultMap); err == nil {
				// 检查是否有 rollback_sql
				hasRollbackSQL := false
				if rollbackSQL, ok := resultMap["rollback_sql"].(string); ok && rollbackSQL != "" {
					hasRollbackSQL = true
				}
				// 移除 rollback_sql，添加 has_rollback_sql 标志
				delete(resultMap, "rollback_sql")
				resultMap["has_rollback_sql"] = hasRollbackSQL
				// 重新序列化
				if newResult, err := json.Marshal(resultMap); err == nil {
					tasks[i].Result = newResult
				}
			}
		}
	}
	
	return tasks, nil
}

func (s *InsightService) GetTaskByID(ctx context.Context, taskID string) (*insight.OrderTask, error) {
	return s.getRepo().GetTaskByID(ctx, taskID)
}

// GetTaskRollbackSQL 获取任务回滚SQL（按需加载）
func (s *InsightService) GetTaskRollbackSQL(ctx context.Context, taskID string) (string, error) {
	task, err := s.getRepo().GetTaskByID(ctx, taskID)
	if err != nil {
		return "", err
	}
	
	if len(task.Result) == 0 {
		return "", nil
	}
	
	var resultMap map[string]interface{}
	if err := json.Unmarshal(task.Result, &resultMap); err != nil {
		return "", err
	}
	
	if rollbackSQL, ok := resultMap["rollback_sql"].(string); ok {
		return rollbackSQL, nil
	}
	
	return "", nil
}

func (s *InsightService) CreateOrderTasks(ctx context.Context, tasks []insight.OrderTask) error {
	return s.getRepo().CreateOrderTasks(ctx, tasks)
}

func (s *InsightService) UpdateTaskProgress(ctx context.Context, taskID string, progress insight.TaskProgress, result []byte) error {
	return s.getRepo().UpdateTaskProgress(ctx, taskID, progress, result)
}

// CheckTasksProgressIsDoing 检查工单是否有任务正在执行中
func (s *InsightService) CheckTasksProgressIsDoing(ctx context.Context, orderID string) (bool, error) {
	return s.getRepo().CheckTasksProgressIsDoing(ctx, orderID)
}

// CheckTasksProgressIsPause 检查工单是否有已暂停的任务
func (s *InsightService) CheckTasksProgressIsPause(ctx context.Context, orderID string) (bool, error) {
	return s.getRepo().CheckTasksProgressIsPause(ctx, orderID)
}

// UpdateOrderExecuteResult 更新工单执行结果
func (s *InsightService) UpdateOrderExecuteResult(ctx context.Context, orderID string, result string) error {
	return s.getRepo().UpdateOrderExecuteResult(ctx, orderID, result)
}

// CheckAllTasksCompleted 检查所有任务是否都已完成
func (s *InsightService) CheckAllTasksCompleted(ctx context.Context, orderID string) (bool, error) {
	return s.getRepo().CheckAllTasksCompleted(ctx, orderID)
}

// UpdateTaskAndOrderProgress 使用事务同时更新任务和工单状态
func (s *InsightService) UpdateTaskAndOrderProgress(ctx context.Context, taskID string, orderID string, taskProgress insight.TaskProgress, orderProgress insight.Progress) error {
	return s.getRepo().UpdateTaskAndOrderProgress(ctx, taskID, orderID, taskProgress, orderProgress)
}

// ExecuteOrder 执行工单的所有任务（用于定时任务调度器）
func (s *InsightService) ExecuteOrder(ctx context.Context, orderID string, username string) error {
	global.Logger.Info("ExecuteOrder 被调用",
		zap.String("order_id", orderID),
		zap.String("username", username),
	)

	// 获取工单信息
	order, err := s.getRepo().GetOrderByID(ctx, orderID)
	if err != nil {
		global.Logger.Error("获取工单信息失败",
			zap.String("order_id", orderID),
			zap.Error(err),
		)
		return fmt.Errorf("获取工单信息失败: %w", err)
	}

	global.Logger.Info("获取工单信息成功",
		zap.String("order_id", orderID),
		zap.String("progress", string(order.Progress)),
	)

	// 检查工单状态（允许"已批准"或"执行中"状态，因为可能已经手动执行过）
	if order.Progress != insight.ProgressApproved && order.Progress != insight.ProgressExecuting {
		global.Logger.Warn("工单状态不允许执行",
			zap.String("order_id", orderID),
			zap.String("progress", string(order.Progress)),
		)
		return fmt.Errorf("工单状态不允许执行，当前状态: %s", order.Progress)
	}

	// 如果已经是执行中状态，检查是否有未完成的任务
	if order.Progress == insight.ProgressExecuting {
		tasks, err := s.getRepo().GetOrderTasks(ctx, orderID)
		if err == nil {
			hasPendingTasks := false
			for _, task := range tasks {
				if task.Progress != insight.TaskProgressCompleted {
					hasPendingTasks = true
					break
				}
			}
			if !hasPendingTasks {
				global.Logger.Info("工单已在执行中且所有任务已完成",
					zap.String("order_id", orderID),
				)
				return nil
			}
		}
	}

	// 检查是否有任务正在执行中
	noExecutingTasks, err := s.CheckTasksProgressIsDoing(ctx, orderID)
	if err != nil {
		return fmt.Errorf("检查任务状态失败: %w", err)
	}
	if !noExecutingTasks {
		return fmt.Errorf("当前有任务正在执行中，请先等待执行完成")
	}

	// 获取工单的所有任务
	tasks, err := s.getRepo().GetOrderTasks(ctx, orderID)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有需要执行的任务")
	}

	// 更新工单状态为执行中
	if err := s.UpdateOrderProgress(ctx, orderID, insight.ProgressExecuting); err != nil {
		return fmt.Errorf("更新工单状态失败: %w", err)
	}

	// 记录操作日志
	_ = s.CreateOpLog(ctx, &insight.OrderOpLog{
		Username: username,
		OrderID:  order.OrderID,
		Msg:      "定时任务自动执行",
	})

	// 获取数据库配置
	dbConfig, err := s.GetDBConfigByInstanceID(ctx, order.InstanceID.String())
	if err != nil {
		return fmt.Errorf("获取数据库配置失败: %w", err)
	}

	// 在后台 goroutine 中异步执行所有任务
	go func() {
		ctx := context.Background()
		var executedCount, successCount, failCount int

		// 逐个执行任务
		for _, task := range tasks {
			// 跳过已完成的任务
			if task.Progress == insight.TaskProgressCompleted {
				continue
			}

			executedCount++

			// 更新任务状态为执行中
			_ = s.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressExecuting, nil)

			// 创建执行器配置
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
			}

			// 创建执行器
			exec, err := executor.NewExecuteSQL(execConfig)
			if err != nil {
				failCount++
				global.Logger.Error("创建执行器失败",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Error(err),
				)
				_ = s.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressFailed, nil)
				continue
			}

			// 执行SQL
			result, err := exec.Run()
			resultJSON, _ := json.Marshal(result)

			if err != nil {
				failCount++
				global.Logger.Error("任务执行失败",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Error(err),
				)
				_ = s.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressFailed, resultJSON)
			} else {
				successCount++
				global.Logger.Info("任务执行成功",
					zap.String("task_id", task.TaskID.String()),
					zap.String("order_id", orderID),
					zap.Int64("affected_rows", result.AffectedRows),
				)
				_ = s.UpdateTaskProgress(ctx, task.TaskID.String(), insight.TaskProgressCompleted, resultJSON)
			}
		}

		// 检查所有任务是否完成
		allCompleted, _ := s.CheckAllTasksCompleted(ctx, orderID)
		if allCompleted {
			_ = s.UpdateOrderProgress(ctx, orderID, insight.ProgressCompleted)
			_ = s.CreateOpLog(ctx, &insight.OrderOpLog{
				Username: username,
				OrderID:  order.OrderID,
				Msg:      fmt.Sprintf("定时任务执行完成，成功: %d, 失败: %d", successCount, failCount),
			})

			// 同步流程引擎：自动完成执行节点任务，推进到结束节点
			go func() {
				order, err := s.GetOrderByID(context.Background(), orderID)
				if err != nil || order == nil {
					return
				}

				// 如果工单有关联的流程实例，自动完成执行节点任务
				hasFlowInstance := false
				if order.FlowInstanceID > 0 {
					// 获取流程实例
					flowInstance, err := FlowServiceApp.GetFlowInstanceDetail(context.Background(), order.FlowInstanceID)
					if err != nil || flowInstance == nil {
						return
					}

					// 查找执行节点的待处理任务
					for _, task := range flowInstance.Tasks {
						// 查找执行节点（nodeCode 包含 "execute" 或 nodeName 包含 "执行"）
						if (strings.Contains(task.NodeCode, "execute") || strings.Contains(task.NodeName, "执行")) &&
							task.Status == "pending" {
							// 自动完成执行节点任务
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
							_ = FlowServiceApp.ApproveTask(context.Background(), &api.ApproveTaskRequest{
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
	}()

	return nil
}

// ============ 操作日志管理 ============

func (s *InsightService) CreateOpLog(ctx context.Context, log *insight.OrderOpLog) error {
	return s.getRepo().CreateOpLog(ctx, log)
}

func (s *InsightService) GetOpLogs(ctx context.Context, orderID string) ([]insight.OrderOpLog, error) {
	return s.getRepo().GetOpLogs(ctx, orderID)
}
