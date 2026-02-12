package initializer

import (
	"context"
	"go-noah/internal/model"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserInitializer struct {
	logger *log.Logger
}

func NewUserInitializer(logger *log.Logger) *UserInitializer {
	return &UserInitializer{logger: logger}
}

func (u *UserInitializer) Name() string {
	return "user"
}

func (u *UserInitializer) Order() int {
	return InitOrderUser
}

func (u *UserInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.AutoMigrate(&model.AdminUser{})
}

func (u *UserInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.Migrator().HasTable(&model.AdminUser{})
}

func (u *UserInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	var count int64
	db.Model(&model.AdminUser{}).Where("id = ?", 1).Count(&count)
	return count > 0
}

func (u *UserInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	if u.IsDataInitialized(ctx, db) {
		u.logger.Debug("用户数据已存在，跳过初始化")
		return nil
	}

	// 使用与旧代码一致的密码 "1234.Com!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("1234.Com!"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []model.AdminUser{
		{
			Model:    gorm.Model{ID: 1},
			Username: "admin",
			Password: string(hashedPassword),
			Nickname: "Admin",
			Email:    "admin@example.com",
			Phone:    "",
			Status:   1,
		},
		{
			Model:    gorm.Model{ID: 2},
			Username: "user",
			Password: string(hashedPassword),
			Nickname: "运营人员",
			Email:    "user@example.com",
			Phone:    "",
			Status:   1,
		},
	}

	for _, user := range users {
		var existingUser model.AdminUser
		if err := db.Where("id = ?", user.ID).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&user).Error; err != nil {
					u.logger.Error("创建用户失败", zap.String("username", user.Username), zap.Error(err))
					return err
				}
				u.logger.Info("创建用户成功", zap.String("username", user.Username))
			} else {
				u.logger.Error("查询用户失败", zap.String("username", user.Username), zap.Error(err))
				return err
			}
		}
	}

	// 将用户数据存入 context，供后续初始化器使用
	ctx = context.WithValue(ctx, "users", users)
	u.logger.Info("用户初始化成功")
	return nil
}
