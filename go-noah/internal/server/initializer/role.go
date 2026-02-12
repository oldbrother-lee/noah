package initializer

import (
	"context"
	"go-noah/internal/model"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RoleInitializer struct {
	logger *log.Logger
}

func NewRoleInitializer(logger *log.Logger) *RoleInitializer {
	return &RoleInitializer{logger: logger}
}

func (r *RoleInitializer) Name() string {
	return "role"
}

func (r *RoleInitializer) Order() int {
	return InitOrderRole
}

func (r *RoleInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.AutoMigrate(&model.Role{})
}

func (r *RoleInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.Migrator().HasTable(&model.Role{})
}

func (r *RoleInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	var count int64
	db.Model(&model.Role{}).Where("sid = ?", model.AdminRole).Count(&count)
	return count > 0
}

func (r *RoleInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	if r.IsDataInitialized(ctx, db) {
		r.logger.Debug("角色数据已存在，跳过初始化")
		return nil
	}

	// 注意：这里只初始化基础角色，额外的角色（运营人员、访客）在 RBAC 初始化器中创建
	// 因为它们在 initialRBAC 中创建，保持与旧代码一致
	roles := []model.Role{
		{Sid: model.AdminRole, Name: "超级管理员", Description: "系统最高权限，可管理所有功能", DataScope: model.DataScopeAll, Status: 1},
		{Sid: model.RoleDBA, Name: "DBA", Description: "数据库管理员，可管理数据库和审批工单", DataScope: model.DataScopeAll, Status: 1},
		{Sid: model.RoleDeveloper, Name: "开发人员", Description: "普通开发人员，可提交工单和查询数据", DataScope: model.DataScopeDeptTree, Status: 1},
	}

	createdCount := 0
	for _, role := range roles {
		var existingRole model.Role
		if err := db.Where("sid = ?", role.Sid).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					r.logger.Warn("创建角色失败", zap.String("sid", role.Sid), zap.Error(err))
					continue
				}
				createdCount++
				r.logger.Info("创建角色成功", zap.String("sid", role.Sid), zap.String("name", role.Name))
			} else {
				r.logger.Warn("查询角色失败", zap.String("sid", role.Sid), zap.Error(err))
			}
		}
	}

	if createdCount > 0 {
		r.logger.Info("角色初始化完成", zap.Int("created_count", createdCount))
	} else {
		r.logger.Debug("角色已存在，跳过初始化")
	}

	// 将角色数据存入 context，供后续初始化器使用
	ctx = context.WithValue(ctx, "roles", roles)
	return nil
}
