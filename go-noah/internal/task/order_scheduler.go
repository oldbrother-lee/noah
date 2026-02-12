package task

import (
	"context"
	"encoding/json"
	"go-noah/internal/model/insight"
	insightRepo "go-noah/internal/repository/insight"
	"go-noah/internal/repository"
	"go-noah/pkg/global"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 全局调度器实例
var globalOrderScheduler *OrderScheduler
var schedulerOnce sync.Once

// OrderScheduler 工单定时任务调度器
type OrderScheduler struct {
	jobMap   map[string]*time.Timer
	mapMutex sync.Mutex
	repo     *insightRepo.InsightRepository
	logger   *zap.Logger
	executor func(ctx context.Context, orderID string, username string) error
}

// NewOrderScheduler 创建工单调度器
func NewOrderScheduler(repo *insightRepo.InsightRepository, logger *zap.Logger) *OrderScheduler {
	return &OrderScheduler{
		jobMap: make(map[string]*time.Timer),
		repo:   repo,
		logger: logger,
	}
}

// GetOrderScheduler 获取全局调度器实例（单例模式）
func GetOrderScheduler() *OrderScheduler {
	schedulerOnce.Do(func() {
		baseRepo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
		repo := insightRepo.NewInsightRepository(baseRepo, global.Logger, global.Enforcer)
		globalOrderScheduler = NewOrderScheduler(repo, global.Logger.Logger)
	})
	return globalOrderScheduler
}

// SetExecutor 设置执行器函数
func (s *OrderScheduler) SetExecutor(exec func(ctx context.Context, orderID string, username string) error) {
	s.executor = exec
}

// Start 启动调度器
func (s *OrderScheduler) Start(ctx context.Context) {
	s.logger.Info("工单定时任务调度器已启动")
	
	// 扫描并注册已批准的定时工单
	s.ScanAndRegister(ctx)
}

// Stop 停止调度器
func (s *OrderScheduler) Stop() {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	
	// 停止所有定时器
	for orderID, timer := range s.jobMap {
		timer.Stop()
		delete(s.jobMap, orderID)
	}
	
	s.logger.Info("工单定时任务调度器已停止")
}

// ScanAndRegister 扫描数据库中的已批准定时工单并注册
// 优化：分批处理，限制每次处理的工单数量，避免一次性加载过多数据
func (s *OrderScheduler) ScanAndRegister(ctx context.Context) {
	startTime := time.Now()
	batchSize := 50 // 每次处理50个工单

	// 分批处理，直到没有更多需要注册的工单
	totalCount := 0
	for {
		// 获取一批需要注册的定时工单
		orders, err := s.repo.GetOrdersPendingSchedulerRegistration(ctx, batchSize)
		if err != nil {
			s.logger.Error("扫描定时工单失败", zap.Error(err))
			return
		}

		// 如果没有更多工单，退出循环
		if len(orders) == 0 {
			break
		}

		// 处理这批工单
		batchCount := 0
		now := time.Now()
		for _, order := range orders {
			if order.ScheduleTime == nil {
				continue
			}

			// 检查定时时间是否已过期
			if order.ScheduleTime.Before(now) {
				// 定时时间已过期，不执行，只标记为已注册（跳过）
				if err := s.repo.MarkSchedulerRegistered(ctx, order.OrderID.String()); err != nil {
					s.logger.Warn("标记过期工单已注册失败",
						zap.String("order_id", order.OrderID.String()),
						zap.Error(err),
					)
				} else {
					s.logger.Info("定时工单已过期，跳过执行",
						zap.String("order_id", order.OrderID.String()),
						zap.Time("schedule_time", *order.ScheduleTime),
						zap.Duration("expired_by", now.Sub(*order.ScheduleTime)),
					)
				}
				continue
			}

			// 定时时间未过期，注册定时任务
			s.AddJob(ctx, &order)
			// 标记为已注册
			if err := s.repo.MarkSchedulerRegistered(ctx, order.OrderID.String()); err != nil {
				s.logger.Warn("标记工单已注册失败",
					zap.String("order_id", order.OrderID.String()),
					zap.Error(err),
				)
			} else {
				batchCount++
				s.logger.Debug("注册定时任务并标记",
					zap.String("order_id", order.OrderID.String()),
					zap.Time("schedule_time", *order.ScheduleTime),
				)
			}
		}

		totalCount += batchCount

		// 如果这批处理的工单数量少于批次大小，说明已经处理完了
		if len(orders) < batchSize {
			break
		}

		// 如果处理时间过长（超过5秒），记录警告并退出，避免阻塞
		if time.Since(startTime) > 5*time.Second {
			s.logger.Warn("扫描定时工单耗时过长，暂停本次扫描",
				zap.Duration("elapsed", time.Since(startTime)),
				zap.Int("processed", totalCount),
			)
			break
		}
	}

	if totalCount > 0 {
		s.logger.Info("扫描并注册定时工单完成",
			zap.Int("count", totalCount),
			zap.Duration("elapsed", time.Since(startTime)),
		)
	}
}

// AddJob 添加定时执行任务
func (s *OrderScheduler) AddJob(ctx context.Context, order *insight.OrderRecord) {
	if order.ScheduleTime == nil {
		return
	}

	orderID := order.OrderID.String()
	targetTime := *order.ScheduleTime
	now := time.Now()

	// 如果定时时间已过期，不执行（已在 ScanAndRegister 中处理，这里不应该到达）
	if targetTime.Before(now) {
		s.logger.Warn("定时工单时间已过期，不执行",
			zap.String("order_id", orderID),
			zap.Time("schedule_time", targetTime),
			zap.Duration("expired_by", now.Sub(targetTime)),
		)
		return
	}

	// 移除已存在的任务（如果存在）
	s.RemoveJob(orderID)

	// 计算延迟时间
	delay := targetTime.Sub(now)

	// 使用 time.AfterFunc 在指定时间执行一次
	timer := time.AfterFunc(delay, func() {
		s.logger.Info("定时任务回调函数被触发",
			zap.String("order_id", orderID),
			zap.Time("target_time", targetTime),
			zap.Time("current_time", time.Now()),
		)
		
		s.mapMutex.Lock()
		delete(s.jobMap, orderID)
		s.mapMutex.Unlock()

		s.executeOrder(context.Background(), orderID, order.Applicant)
	})

	s.mapMutex.Lock()
	s.jobMap[orderID] = timer
	s.mapMutex.Unlock()

	s.logger.Info("注册定时任务成功",
		zap.String("order_id", orderID),
		zap.Time("schedule_time", targetTime),
		zap.Duration("delay", delay),
	)
}

// RemoveJob 移除定时任务
func (s *OrderScheduler) RemoveJob(orderID string) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	if timer, ok := s.jobMap[orderID]; ok {
		timer.Stop()
		delete(s.jobMap, orderID)
		s.logger.Info("移除定时任务", zap.String("order_id", orderID))
	}

	// 清理注册标记
	ctx := context.Background()
	_ = s.repo.MarkSchedulerUnregistered(ctx, orderID)
}

// executeOrder 执行工单
func (s *OrderScheduler) executeOrder(ctx context.Context, orderID string, defaultUsername string) {
	// 重新查询工单信息，确保状态是最新的
	orderWithInstance, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		s.logger.Error("获取工单信息失败",
			zap.String("order_id", orderID),
			zap.Error(err),
		)
		return
	}

	// 清理注册标记（定时任务已执行，不再需要标记）
	_ = s.repo.MarkSchedulerUnregistered(ctx, orderID)

	// 再次检查工单状态
	if orderWithInstance.Progress != insight.ProgressApproved {
		s.logger.Warn("工单状态不是已批准，取消执行",
			zap.String("order_id", orderID),
			zap.String("progress", string(orderWithInstance.Progress)),
		)
		return
	}

	// 获取执行人
	username := defaultUsername
	if orderWithInstance.Executor != nil && len(orderWithInstance.Executor) > 0 {
		var executorList []string
		if err := json.Unmarshal(orderWithInstance.Executor, &executorList); err == nil && len(executorList) > 0 {
			username = executorList[0]
		}
	}

	s.logger.Info("开始执行定时工单",
		zap.String("order_id", orderID),
		zap.String("username", username),
		zap.Bool("executor_set", s.executor != nil),
	)

	if s.executor != nil {
		s.logger.Info("调用执行器函数",
			zap.String("order_id", orderID),
		)
		if err := s.executor(ctx, orderID, username); err != nil {
			s.logger.Error("定时工单执行失败",
				zap.String("order_id", orderID),
				zap.Error(err),
			)
		} else {
			s.logger.Info("定时工单执行完成",
				zap.String("order_id", orderID),
			)
		}
	} else {
		s.logger.Error("执行器未设置，无法执行定时工单",
			zap.String("order_id", orderID),
		)
	}
}
