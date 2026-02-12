package repository

import (
	"context"
	"go-noah/internal/model"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) *UserRepository {
	return &UserRepository{
		Repository: repository,
	}
}

func (r *UserRepository) GetUser(ctx context.Context, uid uint64) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Where("uid = ?", uid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.DB(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUsers(ctx context.Context, page, pageSize int, search, organizationKey string, roleID uint64) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * pageSize
	query := r.DB(ctx).Model(&model.User{})

	// 搜索条件
	if search != "" {
		query = query.Where("username LIKE ? OR nick_name LIKE ? OR email LIKE ? OR mobile LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 角色过滤
	if roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.DB(ctx).Create(user).Error
}

func (r *UserRepository) UpdateUser(ctx context.Context, uid uint64, updates map[string]interface{}) error {
	return r.DB(ctx).Model(&model.User{}).Where("uid = ?", uid).Updates(updates).Error
}

func (r *UserRepository) DeleteUser(ctx context.Context, uid uint64) error {
	return r.DB(ctx).Where("uid = ?", uid).Delete(&model.User{}).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, uid uint64, password string) error {
	return r.DB(ctx).Model(&model.User{}).Where("uid = ?", uid).Update("password", password).Error
}
