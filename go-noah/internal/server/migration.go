package server

import (
	"context"
	"encoding/json"
	"fmt"
	"go-noah/api"
	"go-noah/internal/model"
	"go-noah/internal/model/insight"
	"go-noah/pkg/log"
	"go-noah/pkg/sid"
	"net/http"
	"os"
	"strings"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MigrateServer struct {
	db  *gorm.DB
	log *log.Logger
	sid *sid.Sid
	e   *casbin.SyncedEnforcer
}

func NewMigrateServer(
	db *gorm.DB,
	log *log.Logger,
	sid *sid.Sid,
	e *casbin.SyncedEnforcer,
) *MigrateServer {
	return &MigrateServer{
		e:   e,
		db:  db,
		log: log,
		sid: sid,
	}
}
func (m *MigrateServer) Start(ctx context.Context) error {
	// 只执行 AutoMigrate，不删除表（安全模式）
	// 如果需要删除表重建，请手动执行 DropTable
	// m.db.Migrator().DropTable(
	// 	&model.AdminUser{},
	// 	&model.Menu{},
	// 	&model.Role{},
	// 	&model.Api{},
	// )
	if err := m.db.AutoMigrate(
		&model.AdminUser{},
		&model.Menu{},
		&model.Role{},
		&model.Api{},
		&model.ApiIgnore{},
		// 新增: 部门管理
		&model.Department{},
		&model.RoleDepartment{},
		// 新增: 审批流程
		&model.FlowDefinition{},
		&model.FlowNode{},
		&model.FlowInstance{},
		&model.FlowTask{},
		&model.FlowLog{},
		&model.FlowCC{},
		// goInsight 功能表
		&insight.DBEnvironment{},
		&insight.DBConfig{},
		&insight.DBSchema{},
		&insight.Organization{},
		&insight.OrganizationUser{},
		&insight.DASUserSchemaPermission{},
		&insight.DASUserTablePermission{},
		&insight.DASAllowedOperation{},
		&insight.DASRecord{},
		&insight.DASFavorite{},
		// 新增: 权限管理
		&insight.DASPermissionTemplate{},
		&insight.DASRolePermission{},
		&insight.OrderRecord{},
		&insight.OrderTask{},
		&insight.OrderOpLog{},
		&insight.OrderMessage{},
		&insight.InspectParams{},
	); err != nil {
		m.log.Error("user migrate error", zap.Error(err))
		return err
	}

	// 更新现有菜单数据：将旧字段映射到新字段
	m.updateMenuData(ctx)
	err := m.initialAdminUser(ctx)
	if err != nil {
		m.log.Error("initialAdminUser error", zap.Error(err))
	}

	err = m.initialMenuData(ctx)
	if err != nil {
		m.log.Error("initialMenuData error", zap.Error(err))
	}

	err = m.initialApisData(ctx)
	if err != nil {
		m.log.Error("initialApisData error", zap.Error(err))
	}

	err = m.initialRBAC(ctx)
	if err != nil {
		m.log.Error("initialRBAC error", zap.Error(err))
	}

	// 新增: 初始化部门数据
	err = m.initialDepartments(ctx)
	if err != nil {
		m.log.Error("initialDepartments error", zap.Error(err))
	}

	// 新增: 初始化审批流程定义
	err = m.initialFlowDefinitions(ctx)
	if err != nil {
		m.log.Error("initialFlowDefinitions error", zap.Error(err))
	}

	// 新增: 初始化审核参数
	err = m.initialInspectParams(ctx)
	if err != nil {
		m.log.Error("initialInspectParams error", zap.Error(err))
	}

	m.log.Info("AutoMigrate success")
	os.Exit(0)
	return nil
}

// AutoMigrateTables 自动迁移数据库表（服务器启动时调用）
func AutoMigrateTables(db *gorm.DB, logger *log.Logger) error {
	if err := db.AutoMigrate(
		&model.AdminUser{},
		&model.Menu{},
		&model.Role{},
		&model.Api{},
		&model.ApiIgnore{},
		// 新增: 部门管理
		&model.Department{},
		&model.RoleDepartment{},
		// 新增: 审批流程
		&model.FlowDefinition{},
		&model.FlowNode{},
		&model.FlowInstance{},
		&model.FlowTask{},
		&model.FlowLog{},
		&model.FlowCC{},
		// goInsight 功能表
		&insight.DBEnvironment{},
		&insight.DBConfig{},
		&insight.DBSchema{},
		&insight.Organization{},
		&insight.OrganizationUser{},
		&insight.DASUserSchemaPermission{},
		&insight.DASUserTablePermission{},
		&insight.DASAllowedOperation{},
		&insight.DASRecord{},
		&insight.DASFavorite{},
		// 新增: 权限管理
		&insight.DASPermissionTemplate{},
		&insight.DASRolePermission{},
		&insight.OrderRecord{},
		&insight.OrderTask{},
		&insight.OrderOpLog{},
		&insight.OrderMessage{},
		&insight.InspectParams{},
	); err != nil {
		logger.Error("AutoMigrate tables error", zap.Error(err))
		return err
	}
	logger.Info("AutoMigrate tables success")

	// 创建定时工单调度器性能优化索引
	if err := createOrderSchedulerIndex(db, logger); err != nil {
		logger.Warn("创建定时工单调度器索引失败", zap.Error(err))
		// 不阻止服务启动，只记录警告
	}

	return nil
}

// createOrderSchedulerIndex 创建定时工单调度器性能优化索引
func createOrderSchedulerIndex(db *gorm.DB, logger *log.Logger) error {
	// 检查索引是否已存在
	hasIndex := db.Migrator().HasIndex(&insight.OrderRecord{}, "idx_order_scheduler_scan")
	if hasIndex {
		logger.Debug("定时工单调度器索引已存在，跳过创建")
		return nil
	}

	// 创建复合索引：progress + scheduler_registered + schedule_time
	// 用于优化定时工单扫描查询性能
	err := db.Exec(`
		CREATE INDEX idx_order_scheduler_scan 
		ON order_records (progress, scheduler_registered, schedule_time)
	`).Error

	if err != nil {
		// 如果索引已存在（可能在其他地方创建），忽略错误
		if contains(err.Error(), "Duplicate key name") || contains(err.Error(), "already exists") {
			logger.Debug("定时工单调度器索引已存在（可能在其他地方创建）")
			return nil
		}
		return fmt.Errorf("创建定时工单调度器索引失败: %w", err)
	}

	logger.Info("成功创建定时工单调度器性能优化索引", zap.String("index", "idx_order_scheduler_scan"))
	return nil
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// InitializeAdminUserIfNeeded 检查并初始化 admin 用户（如果不存在则初始化）
func InitializeAdminUserIfNeeded(db *gorm.DB, logger *log.Logger) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("1234.Com!"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 检查 admin 用户是否已存在，如果不存在则创建
	var adminUser model.AdminUser
	if err := db.Where("id = ?", 1).First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&model.AdminUser{
				Model:    gorm.Model{ID: 1},
				Username: "admin",
				Password: string(hashedPassword),
				Nickname: "Admin",
				Email:    "admin@example.com",
				Phone:    "",
				Status:   1,
			}).Error; err != nil {
				logger.Error("创建 admin 用户失败", zap.Error(err))
				return err
			}
			logger.Info("自动创建 admin 用户成功（默认密码: 123456）")
		} else {
			logger.Error("查询 admin 用户失败", zap.Error(err))
			return err
		}
	}

	// 检查 user 用户是否已存在，如果不存在则创建
	var userUser model.AdminUser
	if err := db.Where("id = ?", 2).First(&userUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&model.AdminUser{
				Model:    gorm.Model{ID: 2},
				Username: "user",
				Password: string(hashedPassword),
				Nickname: "运营人员",
				Email:    "user@example.com",
				Phone:    "",
				Status:   1,
			}).Error; err != nil {
				logger.Error("创建 user 用户失败", zap.Error(err))
				return err
			}
			logger.Info("自动创建 user 用户成功（默认密码: 123456）")
		} else {
			logger.Error("查询 user 用户失败", zap.Error(err))
			return err
		}
	}

	return nil
}

// InitializeRolesIfNeeded 检查并初始化系统角色（如果不存在则初始化）
func InitializeRolesIfNeeded(db *gorm.DB, logger *log.Logger) error {
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
					logger.Warn("创建角色失败", zap.String("sid", role.Sid), zap.Error(err))
				} else {
					createdCount++
					logger.Info("创建角色成功", zap.String("sid", role.Sid), zap.String("name", role.Name))
				}
			} else {
				logger.Warn("查询角色失败", zap.String("sid", role.Sid), zap.Error(err))
			}
		}
	}

	if createdCount > 0 {
		logger.Info("角色初始化完成", zap.Int("created_count", createdCount))
	} else {
		logger.Debug("角色已存在，跳过初始化")
	}

	return nil
}

// InitializeUserRolesIfNeeded 检查并初始化用户角色绑定
func InitializeUserRolesIfNeeded(db *gorm.DB, logger *log.Logger, e *casbin.SyncedEnforcer) error {
	// 为 admin 用户（ID=1）绑定 admin 角色
	roles, err := e.GetRolesForUser(model.AdminUserID)
	if err != nil {
		logger.Error("获取 admin 用户角色失败", zap.Error(err))
		return err
	}

	// 检查是否已有 admin 角色
	hasAdminRole := false
	for _, r := range roles {
		if r == model.AdminRole {
			hasAdminRole = true
			break
		}
	}

	if !hasAdminRole {
		_, err = e.AddRoleForUser(model.AdminUserID, model.AdminRole)
		if err != nil {
			logger.Error("为 admin 用户添加 admin 角色失败", zap.Error(err))
			return err
		}
		logger.Info("为 admin 用户添加 admin 角色成功")
	} else {
		logger.Debug("admin 用户已有 admin 角色，跳过绑定")
	}

	return nil
}

func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
func (m *MigrateServer) initialAdminUser(ctx context.Context) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 检查用户是否已存在，如果不存在则创建
	var adminUser model.AdminUser
	if err := m.db.Where("id = ?", 1).First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := m.db.Create(&model.AdminUser{
				Model:    gorm.Model{ID: 1},
				Username: "admin",
				Password: string(hashedPassword),
				Nickname: "Admin",
			}).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var userUser model.AdminUser
	if err := m.db.Where("id = ?", 2).First(&userUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := m.db.Create(&model.AdminUser{
				Model:    gorm.Model{ID: 2},
				Username: "user",
				Password: string(hashedPassword),
				Nickname: "运营人员",
			}).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
func (m *MigrateServer) initialRBAC(ctx context.Context) error {
	roles := []model.Role{
		{Sid: model.AdminRole, Name: "超级管理员", Description: "系统最高权限，可管理所有功能", DataScope: model.DataScopeAll},
		{Sid: model.RoleDBA, Name: "DBA", Description: "数据库管理员，可管理数据库和审批工单", DataScope: model.DataScopeAll},
		{Sid: model.RoleDeveloper, Name: "开发人员", Description: "普通开发人员，可提交工单和查询数据", DataScope: model.DataScopeDeptTree},
		{Sid: "1000", Name: "运营人员", Description: "运营人员，有限的管理权限", DataScope: model.DataScopeDept},
		{Sid: "1001", Name: "访客", Description: "只读权限", DataScope: model.DataScopeSelf},
	}

	// 只创建不存在的角色
	for _, role := range roles {
		var existingRole model.Role
		if err := m.db.Where("sid = ? OR name = ?", role.Sid, role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := m.db.Create(&role).Error; err != nil {
					m.log.Warn("create role error", zap.String("sid", role.Sid), zap.Error(err))
				}
			} else {
				m.log.Warn("check role error", zap.String("sid", role.Sid), zap.Error(err))
			}
		}
		// 如果角色已存在，跳过创建
	}
	m.e.ClearPolicy()
	err := m.e.SavePolicy()
	if err != nil {
		m.log.Error("m.e.SavePolicy error", zap.Error(err))
		return err
	}
	_, err = m.e.AddRoleForUser(model.AdminUserID, model.AdminRole)
	if err != nil {
		m.log.Error("m.e.AddRoleForUser error", zap.Error(err))
		return err
	}
	menuList := make([]api.MenuDataItem, 0)
	err = json.Unmarshal([]byte(menuData), &menuList)
	if err != nil {
		m.log.Error("json.Unmarshal error", zap.Error(err))
		return err
	}
	for _, item := range menuList {
		m.addPermissionForRole(model.AdminRole, model.MenuResourcePrefix+item.Path, "read")
	}
	apiList := make([]model.Api, 0)
	err = m.db.Find(&apiList).Error
	if err != nil {
		m.log.Error("m.db.Find(&apiList).Error error", zap.Error(err))
		return err
	}
	for _, api := range apiList {
		m.addPermissionForRole(model.AdminRole, model.ApiResourcePrefix+api.Path, api.Method)
	}

	// 添加运营人员权限
	_, err = m.e.AddRoleForUser("2", "1000")
	if err != nil {
		m.log.Error("m.e.AddRoleForUser error", zap.Error(err))
		return err
	}
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile/basic", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile/advanced", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/profile", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/workplace", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/analysis", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/account/settings", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/account/center", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/account", "read")
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/menus", http.MethodGet)
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/admin/user", http.MethodGet)

	return nil
}
func (m *MigrateServer) addPermissionForRole(role, resource, action string) {
	_, err := m.e.AddPermissionForUser(role, resource, action)
	if err != nil {
		m.log.Sugar().Info("为角色 %s 添加权限 %s:%s 失败: %v", role, resource, action, err)
		return
	}
	fmt.Printf("为角色 %s 添加权限: %s %s\n", role, resource, action)
}
func (m *MigrateServer) initialApisData(ctx context.Context) error {
	// API 数据现在由 HTTP 服务器启动时自动从 Gin 路由同步
	// 这里不再需要手动维护 API 列表
	// 参见 internal/server/http.go 中的 syncRoutesToDB 函数
	m.log.Info("API数据将由HTTP服务器启动时自动同步")
	return nil
}
func (m *MigrateServer) initialMenuData(ctx context.Context) error {
	menuList := make([]api.MenuDataItem, 0)
	err := json.Unmarshal([]byte(menuData), &menuList)
	if err != nil {
		m.log.Error("json.Unmarshal error", zap.Error(err))
		return err
	}

	// 只创建不存在的菜单
	for _, item := range menuList {
		var existingMenu model.Menu
		if err := m.db.Where("id = ?", item.ID).First(&existingMenu).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 菜单不存在，创建新菜单
				menu := model.Menu{
					Model: gorm.Model{
						ID: item.ID,
					},
					ParentID:   item.ParentID,
					Path:       item.Path,
					Title:      item.Title,
					Name:       item.Name,
					Component:  item.Component,
					Locale:     item.Locale,
					Weight:     item.Weight,
					Icon:       item.Icon,
					Redirect:   item.Redirect,
					URL:        item.URL,
					KeepAlive:  item.KeepAlive,
					HideInMenu: item.HideInMenu,
				}
				if err := m.db.Create(&menu).Error; err != nil {
					m.log.Warn("create menu error", zap.Uint("id", item.ID), zap.Error(err))
				}
			} else {
				m.log.Warn("check menu error", zap.Uint("id", item.ID), zap.Error(err))
			}
		}
		// 如果菜单已存在，跳过创建
	}

	return nil
}

// updateMenuData 更新现有菜单数据，将旧字段映射到新字段
func (m *MigrateServer) updateMenuData(ctx context.Context) error {
	var menus []model.Menu
	if err := m.db.Find(&menus).Error; err != nil {
		return err
	}

	for _, menu := range menus {
		update := make(map[string]interface{})

		// 如果新字段为空，则从旧字段映射
		if menu.MenuName == "" && menu.Title != "" {
			update["menu_name"] = menu.Title
		}
		if menu.RouteName == "" && menu.Name != "" {
			update["route_name"] = menu.Name
		}
		if menu.RoutePath == "" && menu.Path != "" {
			update["route_path"] = menu.Path
		}
		if menu.I18nKey == "" && menu.Locale != "" {
			update["i18n_key"] = menu.Locale
		}
		if menu.Order == 0 && menu.Weight != 0 {
			update["order"] = menu.Weight
		}
		if menu.MenuType == "" {
			// 检查是否有子菜单
			var childCount int64
			m.db.Model(&model.Menu{}).Where("parent_id = ?", menu.ID).Count(&childCount)
			if childCount > 0 {
				update["menu_type"] = "1" // 目录
			} else {
				update["menu_type"] = "2" // 菜单
			}
		}
		if menu.Status == "" {
			update["status"] = "1" // 默认启用
		}
		if menu.IconType == "" {
			update["icon_type"] = "1" // 默认 iconify
		}

		if len(update) > 0 {
			if err := m.db.Model(&model.Menu{}).Where("id = ?", menu.ID).Updates(update).Error; err != nil {
				m.log.Warn("update menu data error", zap.Uint("id", menu.ID), zap.Error(err))
			}
		}
	}

	return nil
}

// GetMenuData 获取菜单数据（供初始化器使用）
func GetMenuData() string {
	return menuData
}

var menuData = `[
 {
    "id": 18,
    "parentId": 15,
    "path": "/access/admin",
    "title": "管理员账号",
    "name": "accessAdmin",
    "component": "/access/admin",
    "locale": "menu.access.admin"
  },
  {
    "id": 2,
    "parentId": 0,
    "title": "分析页",
    "icon": "DashboardOutlined",
    "component": "/dashboard/analysis",
    "path": "/dashboard/analysis",
    "name": "DashboardAnalysis",
    "keepAlive": true,
    "locale": "menu.dashboard.analysis",
    "weight": 2
  },
  {
    "id": 1,
    "parentId": 0,
    "title": "仪表盘",
    "icon": "DashboardOutlined",
    "component": "RouteView",
    "redirect": "/dashboard/analysis",
    "path": "/dashboard",
    "name": "Dashboard",
    "locale": "menu.dashboard"
  },
  {
    "id": 3,
    "parentId": 0,
    "title": "表单页",
    "icon": "FormOutlined",
    "component": "RouteView",
    "redirect": "/form/basic",
    "path": "/form",
    "name": "Form",
    "locale": "menu.form"
  },
  {
    "id": 5,
    "parentId": 0,
    "title": "链接",
    "icon": "LinkOutlined",
    "component": "RouteView",
    "redirect": "/link/iframe",
    "path": "/link",
    "name": "Link",
    "locale": "menu.link"
  },
  {
    "id": 6,
    "parentId": 5,
    "title": "AntDesign",
    "url": "https://ant.design/",
    "component": "Iframe",
    "path": "/link/iframe",
    "name": "LinkIframe",
    "keepAlive": true,
    "locale": "menu.link.iframe"
  },
  {
    "id": 7,
    "parentId": 5,
    "title": "AntDesignVue",
    "url": "https://antdv.com/",
    "component": "Iframe",
    "path": "/link/antdv",
    "name": "LinkAntdv",
    "keepAlive": true,
    "locale": "menu.link.antdv"
  },
  {
    "id": 8,
    "parentId": 5,
    "path": "https://www.baidu.com",
    "name": "LinkExternal",
    "title": "跳转百度",
    "locale": "menu.link.external"
  },
  {
    "id": 9,
    "parentId": 0,
    "title": "菜单",
    "icon": "BarsOutlined",
    "component": "RouteView",
    "path": "/menu",
    "redirect": "/menu/menu1",
    "name": "Menu",
    "locale": "menu.menu"
  },
  {
    "id": 10,
    "parentId": 9,
    "title": "菜单1",
    "component": "/menu/menu1",
    "path": "/menu/menu1",
    "name": "MenuMenu11",
    "keepAlive": true,
    "locale": "menu.menu.menu1"
  },
  {
    "id": 11,
    "parentId": 9,
    "title": "菜单2",
    "component": "/menu/menu2",
    "path": "/menu/menu2",
    "keepAlive": true,
    "locale": "menu.menu.menu2"
  },
  {
    "id": 12,
    "parentId": 9,
    "path": "/menu/menu3",
    "redirect": "/menu/menu3/menu1",
    "title": "菜单1-1",
    "component": "RouteView",
    "locale": "menu.menu.menu3"
  },
  {
    "id": 13,
    "parentId": 12,
    "path": "/menu/menu3/menu1",
    "component": "/menu/menu-1-1/menu1",
    "title": "菜单1-1-1",
    "keepAlive": true,
    "locale": "menu.menu3.menu1"
  },
  {
    "id": 14,
    "parentId": 12,
    "path": "/menu/menu3/menu2",
    "component": "/menu/menu-1-1/menu2",
    "title": "菜单1-1-2",
    "keepAlive": true,
    "locale": "menu.menu3.menu2"
  },
  {
    "id": 15,
    "path": "/access",
    "component": "RouteView",
    "redirect": "/access/common",
    "title": "权限模块",
    "name": "Access",
    "parentId": 0,
    "icon": "ClusterOutlined",
    "locale": "menu.access",
    "weight": 1
  },
  {
    "id": 51,
    "parentId": 15,
    "path": "/access/role",	
    "title": "角色管理",
    "name": "AccessRoles",
    "component": "/access/role",
    "locale": "menu.access.roles"
  },
{
    "id": 52,
    "parentId": 15,
    "path": "/access/menu",	
    "title": "菜单管理",
    "name": "AccessMenu",
    "component": "/access/menu",
    "locale": "menu.access.menus"
  },
{
    "id": 53,
    "parentId": 15,
    "path": "/access/api",	
    "title": "API管理",
    "name": "AccessAPI",
    "component": "/access/api",
    "locale": "menu.access.api"
  },
  {
    "id": 19,
    "parentId": 0,
    "title": "异常页",
    "icon": "WarningOutlined",
    "component": "RouteView",
    "redirect": "/exception/403",
    "path": "/exception",
    "name": "Exception",
    "locale": "menu.exception"
  },
  {
    "id": 20,
    "parentId": 19,
    "path": "/exception/403",
    "title": "403",
    "name": "403",
    "component": "/exception/403",
    "locale": "menu.exception.not-permission"
  },
  {
    "id": 21,
    "parentId": 19,
    "path": "/exception/404",
    "title": "404",
    "name": "404",
    "component": "/exception/404",
    "locale": "menu.exception.not-find"
  },
  {
    "id": 22,
    "parentId": 19,
    "path": "/exception/500",
    "title": "500",
    "name": "500",
    "component": "/exception/500",
    "locale": "menu.exception.server-error"
  },
  {
    "id": 23,
    "parentId": 0,
    "title": "结果页",
    "icon": "CheckCircleOutlined",
    "component": "RouteView",
    "redirect": "/result/success",
    "path": "/result",
    "name": "Result",
    "locale": "menu.result"
  },
  {
    "id": 24,
    "parentId": 23,
    "path": "/result/success",
    "title": "成功页",
    "name": "ResultSuccess",
    "component": "/result/success",
    "locale": "menu.result.success"
  },
  {
    "id": 25,
    "parentId": 23,
    "path": "/result/fail",
    "title": "失败页",
    "name": "ResultFail",
    "component": "/result/fail",
    "locale": "menu.result.fail"
  },
  {
    "id": 26,
    "parentId": 0,
    "title": "列表页",
    "icon": "TableOutlined",
    "component": "RouteView",
    "redirect": "/list/card-list",
    "path": "/list",
    "name": "List",
    "locale": "menu.list"
  },
  {
    "id": 27,
    "parentId": 26,
    "path": "/list/card-list",
    "title": "卡片列表",
    "name": "ListCard",
    "component": "/list/card-list",
    "locale": "menu.list.card-list"
  },
  {
    "id": 28,
    "parentId": 0,
    "title": "详情页",
    "icon": "ProfileOutlined",
    "component": "RouteView",
    "redirect": "/profile/basic",
    "path": "/profile",
    "name": "Profile",
    "locale": "menu.profile"
  },
  {
    "id": 29,
    "parentId": 28,
    "path": "/profile/basic",
    "title": "基础详情页",
    "name": "ProfileBasic",
    "component": "/profile/basic/index",
    "locale": "menu.profile.basic"
  },
  {
    "id": 30,
    "parentId": 26,
    "path": "/list/search-list",
    "title": "搜索列表",
    "name": "SearchList",
    "component": "/list/search-list",
    "locale": "menu.list.search-list"
  },
  {
    "id": 31,
    "parentId": 30,
    "path": "/list/search-list/articles",
    "title": "搜索列表（文章）",
    "name": "SearchListArticles",
    "component": "/list/search-list/articles",
    "locale": "menu.list.search-list.articles"
  },
  {
    "id": 32,
    "parentId": 30,
    "path": "/list/search-list/projects",
    "title": "搜索列表（项目）",
    "name": "SearchListProjects",
    "component": "/list/search-list/projects",
    "locale": "menu.list.search-list.projects"
  },
  {
    "id": 33,
    "parentId": 30,
    "path": "/list/search-list/applications",
    "title": "搜索列表（应用）",
    "name": "SearchListApplications",
    "component": "/list/search-list/applications",
    "locale": "menu.list.search-list.applications"
  },
  {
    "id": 34,
    "parentId": 26,
    "path": "/list/basic-list",
    "title": "标准列表",
    "name": "BasicCard",
    "component": "/list/basic-list",
    "locale": "menu.list.basic-list"
  },
  {
    "id": 35,
    "parentId": 28,
    "path": "/profile/advanced",
    "title": "高级详细页",
    "name": "ProfileAdvanced",
    "component": "/profile/advanced/index",
    "locale": "menu.profile.advanced"
  },
  {
    "id": 4,
    "parentId": 3,
    "title": "基础表单",
    "component": "/form/basic-form/index",
    "path": "/form/basic-form",
    "name": "FormBasic",
    "keepAlive": false,
    "locale": "menu.form.basic-form"
  },
  {
    "id": 36,
    "parentId": 0,
    "title": "个人页",
    "icon": "UserOutlined",
    "component": "RouteView",
    "redirect": "/account/center",
    "path": "/account",
    "name": "Account",
    "locale": "menu.account"
  },
  {
    "id": 37,
    "parentId": 36,
    "path": "/account/center",
    "title": "个人中心",
    "name": "AccountCenter",
    "component": "/account/center",
    "locale": "menu.account.center"
  },
  {
    "id": 38,
    "parentId": 36,
    "path": "/account/settings",
    "title": "个人设置",
    "name": "AccountSettings",
    "component": "/account/settings",
    "locale": "menu.account.settings"
  },
  {
    "id": 39,
    "parentId": 3,
    "title": "分步表单",
    "component": "/form/step-form/index",
    "path": "/form/step-form",
    "name": "FormStep",
    "keepAlive": false,
    "locale": "menu.form.step-form"
  },
  {
    "id": 40,
    "parentId": 3,
    "title": "高级表单",
    "component": "/form/advanced-form/index",
    "path": "/form/advanced-form",
    "name": "FormAdvanced",
    "keepAlive": false,
    "locale": "menu.form.advanced-form"
  },
  {
    "id": 41,
    "parentId": 26,
    "path": "/list/table-list",
    "title": "查询表格",
    "name": "ConsultTable",
    "component": "/list/table-list",
    "locale": "menu.list.consult-table"
  },
  {
    "id": 42,
    "parentId": 1,
    "title": "监控页",
    "component": "/dashboard/monitor",
    "path": "/dashboard/monitor",
    "name": "DashboardMonitor",
    "keepAlive": true,
    "locale": "menu.dashboard.monitor"
  },
  {
    "id": 43,
    "parentId": 1,
    "title": "工作台",
    "component": "/dashboard/workplace",
    "path": "/dashboard/workplace",
    "name": "DashboardWorkplace",
    "keepAlive": true,
    "locale": "menu.dashboard.workplace"
  },
  {
    "id": 44,
    "parentId": 26,
    "path": "/list/crud-table",
    "title": "增删改查表格",
    "name": "CrudTable",
    "component": "/list/crud-table",
    "locale": "menu.list.crud-table"
  },
  {
    "id": 45,
    "parentId": 9,
    "path": "/menu/menu4",
    "redirect": "/menu/menu4/menu1",
    "title": "菜单2-1",
    "component": "RouteView",
    "locale": "menu.menu.menu4"
  },
  {
    "id": 46,
    "parentId": 45,
    "path": "/menu/menu4/menu1",
    "component": "/menu/menu-2-1/menu1",
    "title": "菜单2-1-1",
    "keepAlive": true,
    "locale": "menu.menu4.menu1"
  },
  {
    "id": 47,
    "parentId": 45,
    "path": "/menu/menu4/menu2",
    "component": "/menu/menu-2-1/menu2",
    "title": "菜单2-1-2",
    "keepAlive": true,
    "locale": "menu.menu4.menu2"
  }
]`

// initialDepartments 初始化部门数据
func (m *MigrateServer) initialDepartments(ctx context.Context) error {
	departments := []model.Department{
		{
			Model:    gorm.Model{ID: 1},
			ParentID: 0,
			Name:     "总公司",
			Code:     "HQ",
			Path:     "/1/",
			Level:    1,
			Leader:   "admin",
			LeaderID: 1,
			Sort:     1,
			Status:   1,
		},
		{
			Model:    gorm.Model{ID: 2},
			ParentID: 1,
			Name:     "技术部",
			Code:     "TECH",
			Path:     "/1/2/",
			Level:    2,
			Sort:     1,
			Status:   1,
		},
		{
			Model:    gorm.Model{ID: 3},
			ParentID: 1,
			Name:     "运营部",
			Code:     "OPS",
			Path:     "/1/3/",
			Level:    2,
			Sort:     2,
			Status:   1,
		},
		{
			Model:    gorm.Model{ID: 4},
			ParentID: 2,
			Name:     "DBA组",
			Code:     "DBA",
			Path:     "/1/2/4/",
			Level:    3,
			Sort:     1,
			Status:   1,
		},
		{
			Model:    gorm.Model{ID: 5},
			ParentID: 2,
			Name:     "开发组",
			Code:     "DEV",
			Path:     "/1/2/5/",
			Level:    3,
			Sort:     2,
			Status:   1,
		},
	}

	for _, dept := range departments {
		var existing model.Department
		if err := m.db.Where("id = ? OR code = ?", dept.ID, dept.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := m.db.Create(&dept).Error; err != nil {
					m.log.Warn("create department error", zap.String("code", dept.Code), zap.Error(err))
				}
			} else {
				m.log.Warn("check department error", zap.String("code", dept.Code), zap.Error(err))
			}
		}
	}

	m.log.Info("初始化部门数据完成")
	return nil
}

// initialFlowDefinitions 初始化审批流程定义
func (m *MigrateServer) initialFlowDefinitions(ctx context.Context) error {
	// DDL工单审批流程
	ddlFlow := model.FlowDefinition{
		Code:        "order_ddl",
		Name:        "DDL工单审批流程",
		Type:        "order_ddl",
		Description: "用于DDL类型SQL的审批流程",
		Version:     1,
		Status:      1,
	}

	// DML工单审批流程
	dmlFlow := model.FlowDefinition{
		Code:        "order_dml",
		Name:        "DML工单审批流程",
		Type:        "order_dml",
		Description: "用于DML类型SQL的审批流程",
		Version:     1,
		Status:      1,
	}

	// 数据导出审批流程
	exportFlow := model.FlowDefinition{
		Code:        "order_export",
		Name:        "数据导出审批流程",
		Type:        "order_export",
		Description: "用于数据导出的审批流程",
		Version:     1,
		Status:      1,
	}

	flows := []model.FlowDefinition{ddlFlow, dmlFlow, exportFlow}

	for _, flow := range flows {
		var existing model.FlowDefinition
		// 检查是否存在启用的流程定义（按 type 和 status = 1 查找）
		if err := m.db.Where("type = ? AND status = 1", flow.Type).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 检查是否存在但未启用的流程定义
				var disabledFlow model.FlowDefinition
				if err2 := m.db.Where("type = ?", flow.Type).First(&disabledFlow).Error; err2 == nil {
					// 存在但未启用，更新为启用状态
					disabledFlow.Status = 1
					if err := m.db.Save(&disabledFlow).Error; err != nil {
						m.log.Warn("update flow definition status error", zap.String("type", flow.Type), zap.Error(err))
					} else {
						// 检查是否有节点，如果没有则创建
						var nodeCount int64
						m.db.Model(&model.FlowNode{}).Where("flow_def_id = ?", disabledFlow.ID).Count(&nodeCount)
						if nodeCount == 0 {
							m.createDefaultFlowNodes(disabledFlow.ID, disabledFlow.Code)
						}
						m.log.Info("已启用流程定义", zap.String("type", flow.Type), zap.Uint("id", disabledFlow.ID))
					}
				} else {
					// 完全不存在，创建新的流程定义
					// 不指定 ID，让数据库自动生成
					newFlow := flow
					newFlow.Model = gorm.Model{} // 清空 ID，让数据库自动生成
					if err := m.db.Create(&newFlow).Error; err != nil {
						m.log.Warn("create flow definition error", zap.String("type", flow.Type), zap.Error(err))
					} else {
						// 为每个流程创建默认节点
						m.createDefaultFlowNodes(newFlow.ID, newFlow.Code)
						m.log.Info("已创建流程定义", zap.String("type", flow.Type), zap.Uint("id", newFlow.ID))
					}
				}
			} else {
				m.log.Warn("check flow definition error", zap.String("type", flow.Type), zap.Error(err))
			}
		} else {
			// 已存在启用的流程定义，检查是否有节点
			var nodeCount int64
			m.db.Model(&model.FlowNode{}).Where("flow_def_id = ?", existing.ID).Count(&nodeCount)
			if nodeCount == 0 {
				// 有流程定义但没有节点，创建默认节点
				m.createDefaultFlowNodes(existing.ID, existing.Code)
				m.log.Info("已为流程定义创建默认节点", zap.String("type", flow.Type), zap.Uint("id", existing.ID))
			}
		}
	}

	m.log.Info("初始化审批流程定义完成")
	return nil
}

// createDefaultFlowNodes 为流程创建默认节点
func (m *MigrateServer) createDefaultFlowNodes(flowDefID uint, flowCode string) {
	nodes := []model.FlowNode{
		{
			FlowDefID:    flowDefID,
			NodeCode:     "start",
			NodeName:     "开始",
			NodeType:     model.NodeTypeStart,
			Sort:         1,
			NextNodeCode: "dba_approval",
		},
		{
			FlowDefID:     flowDefID,
			NodeCode:      "dba_approval",
			NodeName:      "DBA审批",
			NodeType:      model.NodeTypeApproval,
			Sort:          2,
			ApproverType:  model.ApproverTypeRole,
			ApproverIDs:   model.RoleDBA,
			MultiMode:     model.MultiModeAny,
			RejectAction:  model.RejectActionToStart,
			TimeoutHours:  24,
			TimeoutAction: "notify",
			NextNodeCode:  "dba_execute",
		},
		{
			FlowDefID:     flowDefID,
			NodeCode:      "dba_execute",
			NodeName:      "DBA执行",
			NodeType:      model.NodeTypeApproval,
			Sort:          3,
			ApproverType:  model.ApproverTypeRole,
			ApproverIDs:   model.RoleDBA,
			MultiMode:     model.MultiModeAny,
			RejectAction:  model.RejectActionToStart,
			TimeoutHours:  24,
			TimeoutAction: "notify",
			NextNodeCode:  "end",
		},
		{
			FlowDefID:    flowDefID,
			NodeCode:     "end",
			NodeName:     "结束",
			NodeType:     model.NodeTypeEnd,
			Sort:         4,
			NextNodeCode: "",
		},
	}

	for _, node := range nodes {
		if err := m.db.Create(&node).Error; err != nil {
			m.log.Warn("create flow node error",
				zap.String("flowCode", flowCode),
				zap.String("nodeCode", node.NodeCode),
				zap.Error(err))
		}
	}
}

// initialInspectParams 初始化SQL审核参数（调用公共函数）
func (m *MigrateServer) initialInspectParams(ctx context.Context) error {
	return InitializeInspectParams(m.db, m.log)
}

// InitializeInspectParams 初始化SQL审核参数（公共函数，可在服务启动和迁移中使用）
func InitializeInspectParams(db *gorm.DB, logger *log.Logger) error {
	var params []map[string]interface{} = []map[string]interface{}{
		// TABLE
		{"params": map[string]int{"MAX_TABLE_NAME_LENGTH": 32}, "remark": "表名的长度"},
		{"params": map[string]bool{"CHECK_TABLE_COMMENT": true}, "remark": "检查表是否有注释"},
		{"params": map[string]int{"TABLE_COMMENT_LENGTH": 64}, "remark": "表注释的长度"},
		{"params": map[string]bool{"CHECK_IDENTIFIER": true}, "remark": "对象名必须使用字符串范围为正则[a-zA-Z0-9_]"},
		{"params": map[string]bool{"CHECK_IDENTIFER_KEYWORD": false}, "remark": "对象名是否可以使用关键字"},
		{"params": map[string]bool{"CHECK_TABLE_CHARSET": true}, "remark": "是否检查表的字符集和排序规则"},
		{"params": map[string][]map[string]string{"TABLE_SUPPORT_CHARSET": {
			{"charset": "utf8", "recommend": "utf8_general_ci"},
			{"charset": "utf8mb4", "recommend": "utf8mb4_general_ci"},
		}}, "remark": "表支持的字符集"},
		{"params": map[string]bool{"CHECK_TABLE_ENGINE": true}, "remark": "是否检查表的存储引擎"},
		{"params": map[string][]string{"TABLE_SUPPORT_ENGINE": {"InnoDB"}}, "remark": "表支持的存储引擎"},
		{"params": map[string]bool{"ENABLE_PARTITION_TABLE": false}, "remark": "是否启用分区表"},
		{"params": map[string]bool{"CHECK_TABLE_PRIMARY_KEY": true}, "remark": "检查表是否有主键"},
		{"params": map[string]bool{"TABLE_AT_LEAST_ONE_COLUMN": true}, "remark": "表至少要有一列，语法默认支持"},
		{"params": map[string]bool{"CHECK_TABLE_AUDIT_TYPE_COLUMNS": true}, "remark": "启用审计类型的字段(col1 datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP && col2 datetime DEFAULT CURRENT_TIMESTAMP)"},
		{"params": map[string]bool{"ENABLE_CREATE_TABLE_AS": false}, "remark": "是否允许create table as语法"},
		{"params": map[string]bool{"ENABLE_CREATE_TABLE_LIKE": false}, "remark": "是否允许create table like语法"},
		{"params": map[string]bool{"ENABLE_FOREIGN_KEY": false}, "remark": "是否启用外键"},
		{"params": map[string]bool{"CHECK_TABLE_AUTOINCREMENT_INIT_VALUE": true}, "remark": "检查建表是自增列初始值是否为1"},
		{"params": map[string]bool{"ENABLE_CREATE_VIEW": true}, "remark": "是否支持创建和使用视图"},
		{"params": map[string]interface{}{"INNODB_ROW_FORMAT": []string{"DYNAMIC"}}, "remark": "InnoDB表支持的行格式"},
		// COLUMN
		{"params": map[string]int{"MAX_COLUMN_NAME_LENGTH": 64}, "remark": "列名的长度"},
		{"params": map[string]bool{"CHECK_COLUMN_CHARSET": true}, "remark": "是否检查列的字符集"},
		{"params": map[string]bool{"CHECK_COLUMN_COMMENT": true}, "remark": "是否检查列的注释"},
		{"params": map[string]int{"COLUMN_MAX_CHAR_LENGTH": 64}, "remark": "char长度大于N的时候需要改为varchar"},
		{"params": map[string]int{"MAX_VARCHAR_LENGTH": 16383}, "remark": "最大允许定义的varchar长度"},
		{"params": map[string]bool{"ENABLE_COLUMN_BLOB_TYPE": true}, "remark": "是否允许列的类型为BLOB/TEXT"},
		{"params": map[string]bool{"ENABLE_COLUMN_JSON_TYPE": true}, "remark": "是否允许列的类型为JSON"},
		{"params": map[string]bool{"ENABLE_COLUMN_BIT_TYPE": true}, "remark": "是否允许列的类型为BIT"},
		{"params": map[string]bool{"ENABLE_COLUMN_TIMESTAMP_TYPE": false}, "remark": "是否允许列的类型为TIMESTAMP"},
		{"params": map[string]bool{"CHECK_PRIMARYKEY_USE_BIGINT": true}, "remark": "主键是否为bigint"},
		{"params": map[string]bool{"CHECK_PRIMARYKEY_USE_UNSIGNED": true}, "remark": "主键bigint是否为unsigned"},
		{"params": map[string]bool{"CHECK_PRIMARYKEY_USE_AUTO_INCREMENT": true}, "remark": "主键是否定义为自增"},
		{"params": map[string]bool{"ENABLE_COLUMN_NOT_NULL": true}, "remark": "是否允许列定义为NOT NULL"},
		{"params": map[string]bool{"ENABLE_COLUMN_TIME_NULL": true}, "remark": "是否允许时间类型设置为NULL"},
		{"params": map[string]bool{"CHECK_COLUMN_DEFAULT_VALUE": true}, "remark": "列必须要有默认值"},
		{"params": map[string]bool{"CHECK_COLUMN_FLOAT_DOUBLE": true}, "remark": "将float/double转成int/bigint/decimal等"},
		{"params": map[string]bool{"ENABLE_COLUMN_TYPE_CHANGE": false}, "remark": "是否允许变更列类型"},
		{"params": map[string]bool{"ENABLE_COLUMN_TYPE_CHANGE_COMPATIBLE": true}, "remark": "允许tinyint-> int、int->bigint、char->varchar等"},
		{"params": map[string]bool{"ENABLE_COLUMN_CHANGE_COLUMN_NAME": false}, "remark": "是否允许CHANGE修改列名操作"},
		// INDEX
		{"params": map[string]bool{"CHECK_UNIQ_INDEX_PREFIX": true}, "remark": "是否检查唯一索引前缀，如唯一索引必须以uniq_为前缀"},
		{"params": map[string]bool{"CHECK_SECONDARY_INDEX_PREFIX": true}, "remark": "是否检查二级索引前缀，如普通索引必须以idx_为前缀"},
		{"params": map[string]bool{"CHECK_FULLTEXT_INDEX_PREFIX": true}, "remark": "是否检查全文索引前缀，如全文索引必须以full_为前缀"},
		{"params": map[string]string{"UNQI_INDEX_PREFIX": "UNIQ_"}, "remark": "定义唯一索引前缀，不区分大小写"},
		{"params": map[string]string{"SECONDARY_INDEX_PREFIX": "IDX_"}, "remark": "定义二级索引前缀，不区分大小写"},
		{"params": map[string]string{"FULLTEXT_INDEX_PREFIX": "FULL_"}, "remark": "定义全文索引前缀，不区分大小写"},
		{"params": map[string]int{"SECONDARY_INDEX_MAX_KEY_PARTS": 8}, "remark": "组成二级索引的列数不能超过指定的个数,包括唯一索引"},
		{"params": map[string]int{"PRIMARYKEY_MAX_KEY_PARTS": 1}, "remark": "组成主键索引的列数不能超过指定的个数"},
		{"params": map[string]int{"MAX_INDEX_KEYS": 12}, "remark": "最多有N个索引，包括唯一索引/二级索引"},
		{"params": map[string]bool{"ENABLE_INDEX_RENAME": false}, "remark": "是否允许rename索引名"},
		{"params": map[string]bool{"ENABLE_REDUNDANT_INDEX": false}, "remark": "是否允许冗余索引"},
		// ALTER
		{"params": map[string]bool{"ENABLE_DROP_COLS": true}, "remark": "是否允许DROP列"},
		{"params": map[string]bool{"ENABLE_DROP_INDEXES": true}, "remark": "是否允许DROP索引"},
		{"params": map[string]bool{"ENABLE_DROP_PRIMARYKEY": false}, "remark": "是否允许DROP主键"},
		{"params": map[string]bool{"ENABLE_DROP_TABLE": true}, "remark": "是否允许DROP TABLE"},
		{"params": map[string]bool{"ENABLE_TRUNCATE_TABLE": true}, "remark": "是否允许TRUNCATE TABLE"},
		{"params": map[string]bool{"ENABLE_RENAME_TABLE_NAME": false}, "remark": "是否允许rename表名"},
		{"params": map[string]bool{"ENABLE_MYSQL_MERGE_ALTER_TABLE": true}, "remark": "MySQL同一个表的多个ALTER是否合并为单条语句"},
		{"params": map[string]bool{"ENABLE_TIDB_MERGE_ALTER_TABLE": false}, "remark": "TiDB同一个表的多个ALTER是否合并为单条语句"},
		// DML
		{"params": map[string]bool{"DML_MUST_HAVE_WHERE": true}, "remark": "DML语句必须有where条件"},
		{"params": map[string]bool{"DML_DISABLE_LIMIT": true}, "remark": "DML语句中不允许有LIMIT"},
		{"params": map[string]bool{"DML_DISABLE_ORDERBY": true}, "remark": "DML语句中不允许有orderby"},
		{"params": map[string]bool{"DML_DISABLE_SUBQUERY": true}, "remark": "DML语句不能有子查询"},
		{"params": map[string]bool{"CHECK_DML_JOIN_WITH_ON": true}, "remark": "DML的JOIN语句必须有ON语句"},
		{"params": map[string]string{"EXPLAIN_RULE": "first"}, "remark": "explain判断受影响行数时使用的规则('first', 'max')。 'first': 使用第一行的explain结果作为受影响行数, 'max': 使用explain结果中的最大值作为受影响行数"},
		{"params": map[string]int{"MAX_AFFECTED_ROWS": 100}, "remark": "最大影响行数，默认100"},
		{"params": map[string]int{"MAX_INSERT_ROWS": 100}, "remark": " 一次最多允许insert的行, eg: insert into tbl(col,...) values(row1), (row2)..."},
		{"params": map[string]bool{"DISABLE_REPLACE": true}, "remark": "是否禁用replace语句"},
		{"params": map[string]bool{"DISABLE_INSERT_INTO_SELECT": true}, "remark": "是否禁用insert/replace into select语法"},
		{"params": map[string]bool{"DISABLE_ON_DUPLICATE": true}, "remark": "是否禁止insert on duplicate语法"},
		// 禁止语法审核的表
		{"params": map[string]interface{}{"DISABLE_AUDIT_DML_TABLES": []map[string]interface{}{
			{"DB": "d1", "Tables": []string{"t1", "t2"}, "Reason": "研发禁止审核和提交"},
			{"DB": "d2", "Tables": []string{"t1", "t2"}, "Reason": "研发禁止审核和提交"},
		}}, "remark": "禁止指定的表的DML语句进行审核"},
		{"params": map[string]interface{}{"DISABLE_AUDIT_DDL_TABLES": []map[string]interface{}{
			{"DB": "d1", "Tables": []string{"t1", "t2"}, "Reason": "研发禁止审核和提交"},
			{"DB": "d2", "Tables": []string{"t1", "t2"}, "Reason": "研发禁止审核和提交"},
		}}, "remark": "禁止指定的表的DDL语句进行审核"},
	}

	for _, i := range params {
		var inspectParams insight.InspectParams
		jsonParams, err := json.Marshal(i["params"])
		if err != nil {
			logger.Error("marshal inspect params failed", zap.Error(err))
			return err
		}

		// 使用 Remark 作为唯一标识查找（因为模型中有 uniqueIndex:uniq_remark）
		result := db.Where("remark = ?", i["remark"].(string)).First(&inspectParams)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				if err := db.Create(&insight.InspectParams{
					Params: jsonParams,
					Remark: i["remark"].(string),
				}).Error; err != nil {
					logger.Error("create inspect params failed",
						zap.String("remark", i["remark"].(string)),
						zap.Error(err))
					return err
				}
			} else {
				return result.Error
			}
		} else {
			// 记录已存在，跳过（幂等性）
			logger.Debug("inspect params already exists",
				zap.String("remark", i["remark"].(string)))
		}
	}

	logger.Info("初始化审核参数完成")
	return nil
}

// InitializeFlowDefinitionsIfNeeded 检查并初始化流程定义（如果不存在则初始化）
func InitializeFlowDefinitionsIfNeeded(db *gorm.DB, logger *log.Logger) error {
	ctx := context.Background()
	migrateServer := &MigrateServer{
		db:  db,
		log: logger,
	}
	return migrateServer.initialFlowDefinitions(ctx)
}

// InitializeInspectParamsIfNeeded 检查并初始化SQL审核参数（如果不存在则初始化）
func InitializeInspectParamsIfNeeded(db *gorm.DB, logger *log.Logger) error {
	// 检查 inspect_params 表是否有数据
	var count int64
	if err := db.Model(&insight.InspectParams{}).Count(&count).Error; err != nil {
		logger.Error("检查审核参数数据失败", zap.Error(err))
		return err
	}

	// 如果表为空或数据不足（少于预期的最小记录数），则初始化
	if count == 0 {
		logger.Info("检测到审核参数表为空，开始初始化...")
		return InitializeInspectParams(db, logger)
	}

	// 如果已有数据，记录日志但不初始化（幂等性）
	logger.Debug("审核参数数据已存在，跳过初始化", zap.Int64("count", count))
	return nil
}
