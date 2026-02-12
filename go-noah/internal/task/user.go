package task

import (
	"context"
	"go-noah/internal/repository"
)

// UserTask 用户任务层（简化版：直接使用结构体，不定义接口）
type UserTask struct {
	userRepo *repository.UserRepository
	*Task
}

func NewUserTask(
	task *Task,
	userRepo *repository.UserRepository,
) *UserTask {
	return &UserTask{
		userRepo: userRepo,
		Task:     task,
	}
}

func (t *UserTask) CheckUser(ctx context.Context) error {
	// do something
	t.logger.Info("CheckUser")
	return nil
}
