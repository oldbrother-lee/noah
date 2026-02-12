package initializer

import (
	"context"
	"go-noah/internal/model"
	"go-noah/pkg/log"
	"net/http"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RBACInitializer struct {
	logger   *log.Logger
	enforcer *casbin.SyncedEnforcer
}

func NewRBACInitializer(logger *log.Logger, enforcer *casbin.SyncedEnforcer) *RBACInitializer {
	return &RBACInitializer{
		logger:   logger,
		enforcer: enforcer,
	}
}

func (r *RBACInitializer) Name() string {
	return "rbac"
}

func (r *RBACInitializer) Order() int {
	return InitOrderRBAC
}

func (r *RBACInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	// RBAC 使用 Casbin，不需要创建表
	return nil
}

func (r *RBACInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	// Casbin 表由 enforcer 管理
	return true
}

func (r *RBACInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	// 检查是否有 admin 用户的角色绑定
	roles, err := r.enforcer.GetRolesForUser(model.AdminUserID)
	if err != nil || len(roles) == 0 {
		return false
	}
	// 检查是否有 admin 角色
	for _, role := range roles {
		if role == model.AdminRole {
			return true
		}
	}
	return false
}

func (r *RBACInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	if r.IsDataInitialized(ctx, db) {
		r.logger.Debug("RBAC 数据已存在，跳过初始化")
		return nil
	}

	// 先创建额外的角色（与旧代码 initialRBAC 保持一致）
	// 注意：旧代码中这些角色没有显式设置 Status，使用数据库默认值 1
	extraRoles := []model.Role{
		{Sid: "1000", Name: "运营人员", Description: "运营人员，有限的管理权限", DataScope: model.DataScopeDept},
		{Sid: "1001", Name: "访客", Description: "只读权限", DataScope: model.DataScopeSelf},
	}
	for _, role := range extraRoles {
		var existingRole model.Role
		if err := db.Where("sid = ? OR name = ?", role.Sid, role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					r.logger.Warn("创建角色失败", zap.String("sid", role.Sid), zap.Error(err))
				}
			}
		}
	}

	// 清空现有策略
	r.enforcer.ClearPolicy()
	if err := r.enforcer.SavePolicy(); err != nil {
		r.logger.Error("清空策略失败", zap.Error(err))
		return err
	}

	// 为 admin 用户添加 admin 角色
	_, err := r.enforcer.AddRoleForUser(model.AdminUserID, model.AdminRole)
	if err != nil {
		r.logger.Error("为 admin 用户添加 admin 角色失败", zap.Error(err))
		return err
	}

	// 从数据库查询实际存在的菜单，为 admin 角色添加菜单权限
	// 注意：应该只分配数据库中实际存在的菜单权限，而不是从 menuData 变量分配
	var menus []model.Menu
	if err := db.Find(&menus).Error; err != nil {
		r.logger.Error("查询菜单数据失败", zap.Error(err))
	} else {
		for _, menu := range menus {
			if menu.Path != "" {
				r.addPermissionForRole(model.AdminRole, model.MenuResourcePrefix+menu.Path, "read")
			}
		}
		r.logger.Info("为 admin 角色添加菜单权限", zap.Int("menu_count", len(menus)))
	}

	// 为 admin 角色添加所有 API 权限
	apiList := make([]model.Api, 0)
	if err := db.Find(&apiList).Error; err == nil {
		for _, api := range apiList {
			r.addPermissionForRole(model.AdminRole, model.ApiResourcePrefix+api.Path, api.Method)
		}
	}

	// 为 user 用户（ID=2）添加运营人员角色（与旧代码保持一致）
	_, err = r.enforcer.AddRoleForUser("2", "1000")
	if err != nil {
		r.logger.Error("为 user 用户添加运营人员角色失败", zap.Error(err))
		return err
	}

	// 添加运营人员权限（与旧代码保持一致）
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile/basic", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile/advanced", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/workplace", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/analysis", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/account/settings", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/account/center", "read")
	r.addPermissionForRole("1000", model.MenuResourcePrefix+"/account", "read")
	r.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/menus", http.MethodGet)
	r.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/admin/user", http.MethodGet)

	r.logger.Info("RBAC 初始化成功")
	return nil
}

func (r *RBACInitializer) addPermissionForRole(role, resource, action string) {
	_, err := r.enforcer.AddPermissionForUser(role, resource, action)
	if err != nil {
		r.logger.Debug("为角色添加权限失败", zap.String("role", role), zap.String("resource", resource), zap.String("action", action), zap.Error(err))
		return
	}
	r.logger.Debug("为角色添加权限", zap.String("role", role), zap.String("resource", resource), zap.String("action", action))
}
