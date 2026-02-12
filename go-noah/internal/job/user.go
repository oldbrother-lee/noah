package job

import (
	"context"
	"go-noah/internal/repository"
)

// UserJob 用户任务层（简化版：直接使用结构体，不定义接口）
type UserJob struct {
	userRepo *repository.UserRepository
	*Job
}

func NewUserJob(
	job *Job,
	userRepo *repository.UserRepository,
) *UserJob {
	return &UserJob{
		userRepo: userRepo,
		Job:      job,
	}
}

func (t *UserJob) KafkaConsumer(ctx context.Context) error {
	// do something
	return nil
}
