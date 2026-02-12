package initializer

import (
	"context"
	"fmt"
	"go-noah/internal/model"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MenuInitializer struct {
	logger *log.Logger
}

func NewMenuInitializer(logger *log.Logger) *MenuInitializer {
	return &MenuInitializer{logger: logger}
}

func (m *MenuInitializer) Name() string {
	return "menu"
}

func (m *MenuInitializer) Order() int {
	return InitOrderMenu
}

func (m *MenuInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.AutoMigrate(&model.Menu{})
}

func (m *MenuInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.Migrator().HasTable(&model.Menu{})
}

func (m *MenuInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {
	var count int64
	db.Model(&model.Menu{}).Count(&count)
	return count > 0
}

func (m *MenuInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {
	if m.IsDataInitialized(ctx, db) {
		m.logger.Debug("菜单数据已存在，跳过初始化")
		return nil
	}

	// 定义顶级菜单（ParentID = 0）
	level0Menus := []model.Menu{
		model.Menu{
			ParentID:  0,
			Icon:      "mdi:monitor-dashboard",
			MenuName:  "首页",
			RouteName: "home",
			RoutePath: "/home",
			I18nKey:   "route.home",
		},
		model.Menu{
			ParentID:  0,
			Component: "layout.base",
			Icon:      "carbon:cloud-service-management",
			Order:     20,
			MenuType:  "1",
			MenuName:  "系统管理",
			RouteName: "system",
			RoutePath: "/system",
			I18nKey:   "route.system",
		},
		model.Menu{
			ParentID:  0,
			Component: "layout.base",
			Icon:      "mdi:database",
			Order:     5,
			MenuType:  "1",
			MenuName:  "数据库服务",
			RouteName: "das",
			RoutePath: "/das",
			I18nKey:   "route.das",
		},
	}

	// 先创建顶级菜单
	if err := db.Create(&level0Menus).Error; err != nil {
		m.logger.Error("创建顶级菜单失败", zap.Error(err))
		return err
	}

	// 建立菜单映射 - 通过唯一标识符查找已创建的菜单及其ID
	// 优先级：Name > Path > MenuName > RouteName > ID
	menuNameMap := make(map[string]uint)
	for _, menu := range level0Menus {
		var key string
		if menu.Name != "" {
			key = menu.Name
		} else if menu.Path != "" {
			key = menu.Path
		} else if menu.MenuName != "" {
			key = menu.MenuName
		} else if menu.RouteName != "" {
			key = menu.RouteName
		} else {
			// 使用 ID 作为最后的备选（需要先创建才能获取ID）
			key = fmt.Sprintf("menu_%d", menu.ID)
		}
		if key != "" {
			menuNameMap[key] = menu.ID
		}
	}

	// 定义二级菜单，并设置正确的 ParentID
	level1Menus := []model.Menu{
		// 系统管理 子菜单
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base$view.system_menu",
			Icon:      "material-symbols:route",
			Order:     1,
			MenuName:  "菜单管理",
			RouteName: "system_menu",
			RoutePath: "/system/menu",
			I18nKey:   "route.system_menu",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base$view.system_permission",
			Icon:      "carbon:user-role",
			Order:     3,
			MenuName:  "权限管理",
			RouteName: "system_permission",
			RoutePath: "/system/permission",
			I18nKey:   "route.system_permission",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base$view.system_user",
			Icon:      "ic:round-manage-accounts",
			Order:     2,
			MenuName:  "用户管理",
			RouteName: "system_user",
			RoutePath: "/system/user",
			I18nKey:   "route.system_user",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base",
			Icon:      "mdi:database-cog",
			Order:     4,
			MenuType:  "1",
			MenuName:  "系统管理",
			RouteName: "system_database",
			RoutePath: "/system/database",
			I18nKey:   "route.system_database",
		},
		// 数据库服务 子菜单
		model.Menu{
			ParentID:  menuNameMap["数据库服务"],
			Component: "view.das_edit",
			Icon:      "mdi:database-search",
			Order:     1,
			MenuName:  "sql查询",
			RouteName: "das_edit",
			RoutePath: "/das/edit",
			I18nKey:   "route.das_edit",
		},
		model.Menu{
			ParentID:  menuNameMap["数据库服务"],
			Icon:      "mdi:format-list-bulleted",
			Order:     2,
			MenuName:  "工单列表",
			RouteName: "das_orders-list",
			RoutePath: "/das/orders-list",
			I18nKey:   "route.das_orders-list",
		},
		model.Menu{
			ParentID:  menuNameMap["数据库服务"],
			Icon:      "mdi:file-document-edit",
			Order:     3,
			MenuType:  "1",
			MenuName:  "提交工单",
			RouteName: "das_orders_commit",
			RoutePath: "/das/orders/commit",
			I18nKey:   "route.das_orders_commit",
		},
		model.Menu{
			ParentID:   menuNameMap["数据库服务"],
			HideInMenu: true,
			MenuName:   "工单详情",
			RouteName:  "das_orders-detail",
			RoutePath:  "/das/orders-detail/::id",
			I18nKey:    "route.das_orders-detail",
		},
	}

	// 创建二级菜单
	if err := db.Create(&level1Menus).Error; err != nil {
		m.logger.Error("创建二级菜单失败", zap.Error(err))
		return err
	}

	// 更新映射，添加二级菜单
	for _, menu := range level1Menus {
		var key string
		if menu.Name != "" {
			key = menu.Name
		} else if menu.Path != "" {
			key = menu.Path
		} else if menu.MenuName != "" {
			key = menu.MenuName
		} else if menu.RouteName != "" {
			key = menu.RouteName
		} else {
			key = fmt.Sprintf("menu_%d", menu.ID)
		}
		if key != "" {
			menuNameMap[key] = menu.ID
		}
	}

	// 定义三级菜单，并设置正确的 ParentID
	level2Menus := []model.Menu{
		// 系统管理 子菜单
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base$view.system_database_environment",
			Icon:      "mdi:server-network",
			Order:     1,
			MenuName:  "环境管理",
			RouteName: "system_database_environment",
			RoutePath: "/system/database/environment",
			I18nKey:   "route.system_database_environment",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Component: "layout.base",
			Icon:      "mdi:database-settings",
			Order:     2,
			MenuName:  "实例配置",
			RouteName: "system_database_config",
			RoutePath: "/system/database/config",
			I18nKey:   "route.system_database_config",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Icon:      "mdi:shield-account",
			Order:     3,
			MenuName:  "权限管理",
			RouteName: "system_database_permission",
			RoutePath: "/system/database/permission",
			I18nKey:   "route.system_database_permission",
		},
		model.Menu{
			ParentID:  menuNameMap["系统管理"],
			Icon:      "mdi:shield-account",
			MenuName:  "审核参数",
			RouteName: "system_database_inspect",
			RoutePath: "/system/database/inspect",
			I18nKey:   "route.system_database_inspect",
		},
		// 提交工单 子菜单
		model.Menu{
			ParentID:  menuNameMap["提交工单"],
			Component: "view.das_orders_ddl",
			Icon:      "mdi:database-cog",
			Order:     1,
			MenuName:  "ddl工单",
			RouteName: "das_orders_ddl",
			RoutePath: "/das/orders/ddl",
			I18nKey:   "route.das_orders_ddl",
		},
		model.Menu{
			ParentID:  menuNameMap["提交工单"],
			Component: "view.das_orders_dml",
			Icon:      "mdi:database-edit",
			Order:     2,
			MenuName:  "dml工单",
			RouteName: "das_orders_dml",
			RoutePath: "/das/orders/dml",
			I18nKey:   "route.das_orders_dml",
		},
		model.Menu{
			ParentID:  menuNameMap["提交工单"],
			Component: "view.das_orders_export",
			Icon:      "mdi:database-export",
			Order:     3,
			MenuName:  "导出工单",
			RouteName: "das_orders_export",
			RoutePath: "/das/orders/export",
			I18nKey:   "route.das_orders_export",
		},
	}

	// 创建三级菜单
	if err := db.Create(&level2Menus).Error; err != nil {
		m.logger.Error("创建三级菜单失败", zap.Error(err))
		return err
	}

	m.logger.Info("菜单初始化完成",
		zap.Int("level0_count", 3),
		zap.Int("level1_count", 29),
		zap.Int("level2_count", 7))

	return nil
}
