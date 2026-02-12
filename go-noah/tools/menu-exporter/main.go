package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"go-noah/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	var (
		host     = flag.String("host", "localhost", "数据库主机")
		port     = flag.Int("port", 3306, "数据库端口")
		user     = flag.String("user", "root", "数据库用户")
		password = flag.String("password", "", "数据库密码")
		database = flag.String("database", "nunu_test", "数据库名")
		output   = flag.String("output", "internal/server/initializer/menu.go", "输出文件路径")
	)
	flag.Parse()

	if *password == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须提供数据库密码\n")
		flag.Usage()
		os.Exit(1)
	}

	gen, err := NewGenerator(*host, *port, *user, *password, *database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer gen.Close()

	if err := gen.Generate(*output); err != nil {
		fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 菜单代码生成成功！输出文件: %s\n", *output)
}

type Generator struct {
	db *gorm.DB
}

// NewGenerator 创建生成器实例
func NewGenerator(host string, port int, user, password, database string) (*Generator, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return &Generator{db: db}, nil
}

// Close 关闭数据库连接
func (g *Generator) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Generate 生成菜单初始化代码
func (g *Generator) Generate(outputPath string) error {
	// 1. 读取所有菜单数据
	var menus []model.Menu
	if err := g.db.Order("parent_id ASC, weight ASC, id ASC").Find(&menus).Error; err != nil {
		return fmt.Errorf("读取菜单数据失败: %w", err)
	}

	if len(menus) == 0 {
		return fmt.Errorf("数据库中没有菜单数据")
	}

	// 过滤掉测试数据：排除包含 "menu.menu" 或 "menu.list" 的菜单
	filteredMenus := []model.Menu{}
	for _, menu := range menus {
		// 跳过测试菜单（通过 I18nKey 或 Locale 判断）
		if menu.I18nKey != "" && (strings.Contains(menu.I18nKey, "menu.menu") || strings.Contains(menu.I18nKey, "menu.list")) {
			continue
		}
		if menu.Locale != "" && (strings.Contains(menu.Locale, "menu.menu") || strings.Contains(menu.Locale, "menu.list")) {
			continue
		}
		// 跳过路径包含测试数据的菜单
		if menu.Path != "" && (strings.Contains(menu.Path, "/menu/menu") || strings.Contains(menu.Path, "/list/search-list")) {
			continue
		}
		if menu.RoutePath != "" && (strings.Contains(menu.RoutePath, "/menu/menu") || strings.Contains(menu.RoutePath, "/list/search-list")) {
			continue
		}
		filteredMenus = append(filteredMenus, menu)
	}
	menus = filteredMenus

	if len(menus) == 0 {
		return fmt.Errorf("过滤后没有有效的菜单数据")
	}

	// 2. 按层级分组菜单
	// 构建菜单映射：ID -> Menu
	menuMap := make(map[uint]*model.Menu)
	for i := range menus {
		menuMap[menus[i].ID] = &menus[i]
	}

	// 按层级分组
	level0Menus := []model.Menu{} // 顶级菜单（parent_id = 0）
	level1Menus := []model.Menu{} // 二级菜单（parent_id = 某个顶级菜单的ID）
	level2Menus := []model.Menu{} // 三级菜单（parent_id = 某个二级菜单的ID）

	for _, menu := range menus {
		if menu.ParentID == 0 {
			level0Menus = append(level0Menus, menu)
		} else {
			// 检查父菜单是否存在
			if parentMenu, ok := menuMap[menu.ParentID]; ok {
				if parentMenu.ParentID == 0 {
					// 父菜单是顶级菜单，这是二级菜单
					level1Menus = append(level1Menus, menu)
				} else {
					// 父菜单不是顶级菜单，这是三级菜单
					level2Menus = append(level2Menus, menu)
				}
			} else {
				// 父菜单不存在，可能是数据问题，暂时当作二级菜单处理
				level1Menus = append(level1Menus, menu)
			}
		}
	}

	// 3. 按 weight 排序
	sort.Slice(level0Menus, func(i, j int) bool {
		if level0Menus[i].Weight != level0Menus[j].Weight {
			return level0Menus[i].Weight < level0Menus[j].Weight
		}
		return level0Menus[i].ID < level0Menus[j].ID
	})

	sort.Slice(level1Menus, func(i, j int) bool {
		if level1Menus[i].ParentID != level1Menus[j].ParentID {
			return level1Menus[i].ParentID < level1Menus[j].ParentID
		}
		if level1Menus[i].Weight != level1Menus[j].Weight {
			return level1Menus[i].Weight < level1Menus[j].Weight
		}
		return level1Menus[i].ID < level1Menus[j].ID
	})

	sort.Slice(level2Menus, func(i, j int) bool {
		if level2Menus[i].ParentID != level2Menus[j].ParentID {
			return level2Menus[i].ParentID < level2Menus[j].ParentID
		}
		if level2Menus[i].Weight != level2Menus[j].Weight {
			return level2Menus[i].Weight < level2Menus[j].Weight
		}
		return level2Menus[i].ID < level2Menus[j].ID
	})

	// 4. 生成 Go 代码
	code := g.generateCode(level0Menus, level1Menus, level2Menus)

	// 5. 写入文件
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// generateCode 生成 Go 代码
func (g *Generator) generateCode(level0Menus, level1Menus, level2Menus []model.Menu) string {
	var sb strings.Builder

	// 文件头
	sb.WriteString("package initializer\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"go-noah/internal/model\"\n")
	sb.WriteString("\t\"go-noah/pkg/log\"\n\n")
	sb.WriteString("\t\"go.uber.org/zap\"\n")
	sb.WriteString("\t\"gorm.io/gorm\"\n")
	sb.WriteString(")\n\n")

	// 结构体定义
	sb.WriteString("type MenuInitializer struct {\n")
	sb.WriteString("\tlogger *log.Logger\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func NewMenuInitializer(logger *log.Logger) *MenuInitializer {\n")
	sb.WriteString("\treturn &MenuInitializer{logger: logger}\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (m *MenuInitializer) Name() string {\n")
	sb.WriteString("\treturn \"menu\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (m *MenuInitializer) Order() int {\n")
	sb.WriteString("\treturn InitOrderMenu\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (m *MenuInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {\n")
	sb.WriteString("\treturn db.AutoMigrate(&model.Menu{})\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (m *MenuInitializer) IsTableCreated(ctx context.Context, db *gorm.DB) bool {\n")
	sb.WriteString("\treturn db.Migrator().HasTable(&model.Menu{})\n")
	sb.WriteString("}\n\n")

	sb.WriteString("func (m *MenuInitializer) IsDataInitialized(ctx context.Context, db *gorm.DB) bool {\n")
	sb.WriteString("\tvar count int64\n")
	sb.WriteString("\tdb.Model(&model.Menu{}).Count(&count)\n")
	sb.WriteString("\treturn count > 0\n")
	sb.WriteString("}\n\n")

	// InitializeData 方法
	sb.WriteString("func (m *MenuInitializer) InitializeData(ctx context.Context, db *gorm.DB) error {\n")
	sb.WriteString("\tif m.IsDataInitialized(ctx, db) {\n")
	sb.WriteString("\t\tm.logger.Debug(\"菜单数据已存在，跳过初始化\")\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n\n")

	// 生成顶级菜单数组
	sb.WriteString("\t// 定义顶级菜单（ParentID = 0）\n")
	sb.WriteString("\tlevel0Menus := []model.Menu{\n")
	for _, menu := range level0Menus {
		sb.WriteString("\t\t")
		sb.WriteString(g.generateMenuStruct(menu))
		sb.WriteString(",\n")
	}
	sb.WriteString("\t}\n\n")

	// 创建顶级菜单
	sb.WriteString("\t// 先创建顶级菜单\n")
	sb.WriteString("\tif err := db.Create(&level0Menus).Error; err != nil {\n")
	sb.WriteString("\t\tm.logger.Error(\"创建顶级菜单失败\", zap.Error(err))\n")
	sb.WriteString("\t\treturn err\n")
	sb.WriteString("\t}\n\n")

	// 建立顶级菜单映射
	sb.WriteString("\t// 建立菜单映射 - 通过唯一标识符查找已创建的菜单及其ID\n")
	sb.WriteString("\t// 优先级：Name > Path > MenuName > RouteName > ID\n")
	sb.WriteString("\tmenuNameMap := make(map[string]uint)\n")
	sb.WriteString("\tfor _, menu := range level0Menus {\n")
	sb.WriteString("\t\tvar key string\n")
	sb.WriteString("\t\tif menu.Name != \"\" {\n")
	sb.WriteString("\t\t\tkey = menu.Name\n")
	sb.WriteString("\t\t} else if menu.Path != \"\" {\n")
	sb.WriteString("\t\t\tkey = menu.Path\n")
	sb.WriteString("\t\t} else if menu.MenuName != \"\" {\n")
	sb.WriteString("\t\t\tkey = menu.MenuName\n")
	sb.WriteString("\t\t} else if menu.RouteName != \"\" {\n")
	sb.WriteString("\t\t\tkey = menu.RouteName\n")
	sb.WriteString("\t\t} else {\n")
	sb.WriteString("\t\t\t// 使用 ID 作为最后的备选（需要先创建才能获取ID）\n")
	sb.WriteString("\t\t\tkey = fmt.Sprintf(\"menu_%d\", menu.ID)\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\tif key != \"\" {\n")
	sb.WriteString("\t\t\tmenuNameMap[key] = menu.ID\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n\n")

	// 处理二级菜单
	if len(level1Menus) > 0 {
		// 建立 parentID 到 parentKey 的映射（用于二级菜单查找父菜单的 key）
		parentIDToKey := make(map[uint]string)
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
				key = fmt.Sprintf("menu_%d", menu.ID)
			}
			if key != "" {
				parentIDToKey[menu.ID] = key
			}
		}

		// 按父菜单分组
		level1MenusByParent := make(map[uint][]model.Menu)
		for _, menu := range level1Menus {
			level1MenusByParent[menu.ParentID] = append(level1MenusByParent[menu.ParentID], menu)
		}

		sb.WriteString("\t// 定义二级菜单，并设置正确的 ParentID\n")
		sb.WriteString("\tlevel1Menus := []model.Menu{\n")

		// 按父菜单分组输出二级菜单
		for _, parentMenu := range level0Menus {
			if children, ok := level1MenusByParent[parentMenu.ID]; ok {
				// 添加注释
				title := parentMenu.Title
				if title == "" {
					title = parentMenu.MenuName
				}
				if title == "" {
					title = parentMenu.RouteName
				}
				sb.WriteString(fmt.Sprintf("\t\t// %s 子菜单\n", title))
				for _, menu := range children {
					sb.WriteString("\t\t")
					sb.WriteString(g.generateChildMenuStruct(menu, parentIDToKey[menu.ParentID]))
					sb.WriteString(",\n")
				}
			}
		}
		sb.WriteString("\t}\n\n")

		// 创建二级菜单
		sb.WriteString("\t// 创建二级菜单\n")
		sb.WriteString("\tif err := db.Create(&level1Menus).Error; err != nil {\n")
		sb.WriteString("\t\tm.logger.Error(\"创建二级菜单失败\", zap.Error(err))\n")
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n\n")

		// 更新映射，添加二级菜单
		sb.WriteString("\t// 更新映射，添加二级菜单\n")
		sb.WriteString("\tfor _, menu := range level1Menus {\n")
		sb.WriteString("\t\tvar key string\n")
		sb.WriteString("\t\tif menu.Name != \"\" {\n")
		sb.WriteString("\t\t\tkey = menu.Name\n")
		sb.WriteString("\t\t} else if menu.Path != \"\" {\n")
		sb.WriteString("\t\t\tkey = menu.Path\n")
		sb.WriteString("\t\t} else if menu.MenuName != \"\" {\n")
		sb.WriteString("\t\t\tkey = menu.MenuName\n")
		sb.WriteString("\t\t} else if menu.RouteName != \"\" {\n")
		sb.WriteString("\t\t\tkey = menu.RouteName\n")
		sb.WriteString("\t\t} else {\n")
		sb.WriteString("\t\t\tkey = fmt.Sprintf(\"menu_%d\", menu.ID)\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tif key != \"\" {\n")
		sb.WriteString("\t\t\tmenuNameMap[key] = menu.ID\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n\n")
	}

	// 处理三级菜单
	if len(level2Menus) > 0 {
		// 建立 parentID 到 parentKey 的映射（用于三级菜单查找父菜单的 key）
		parentIDToKey := make(map[uint]string)
		// 合并 level0 和 level1 菜单
		allParentMenus := append(level0Menus, level1Menus...)
		for _, menu := range allParentMenus {
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
				parentIDToKey[menu.ID] = key
			}
		}

		// 按父菜单分组
		level2MenusByParent := make(map[uint][]model.Menu)
		for _, menu := range level2Menus {
			level2MenusByParent[menu.ParentID] = append(level2MenusByParent[menu.ParentID], menu)
		}

		sb.WriteString("\t// 定义三级菜单，并设置正确的 ParentID\n")
		sb.WriteString("\tlevel2Menus := []model.Menu{\n")

		// 按父菜单分组输出三级菜单
		for _, parentMenu := range allParentMenus {
			if children, ok := level2MenusByParent[parentMenu.ID]; ok {
				// 添加注释
				title := parentMenu.Title
				if title == "" {
					title = parentMenu.MenuName
				}
				if title == "" {
					title = parentMenu.RouteName
				}
				sb.WriteString(fmt.Sprintf("\t\t// %s 子菜单\n", title))
				for _, menu := range children {
					sb.WriteString("\t\t")
					sb.WriteString(g.generateChildMenuStruct(menu, parentIDToKey[menu.ParentID]))
					sb.WriteString(",\n")
				}
			}
		}
		sb.WriteString("\t}\n\n")

		// 创建三级菜单
		sb.WriteString("\t// 创建三级菜单\n")
		sb.WriteString("\tif err := db.Create(&level2Menus).Error; err != nil {\n")
		sb.WriteString("\t\tm.logger.Error(\"创建三级菜单失败\", zap.Error(err))\n")
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n\n")
	}

	// 日志输出
	sb.WriteString("\tm.logger.Info(\"菜单初始化完成\",\n")
	sb.WriteString(fmt.Sprintf("\t\tzap.Int(\"level0_count\", %d),\n", len(level0Menus)))
	sb.WriteString(fmt.Sprintf("\t\tzap.Int(\"level1_count\", %d),\n", len(level1Menus)))
	sb.WriteString(fmt.Sprintf("\t\tzap.Int(\"level2_count\", %d))\n\n", len(level2Menus)))

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateMenuStruct 生成父菜单的结构体代码
func (g *Generator) generateMenuStruct(menu model.Menu) string {
	var sb strings.Builder

	sb.WriteString("model.Menu{\n")
	sb.WriteString("\t\t\tParentID:  0,\n")

	// Path
	if menu.Path != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tPath:      %q,\n", menu.Path))
	}

	// Title
	if menu.Title != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTitle:     %q,\n", menu.Title))
	}

	// Name（重要：用于映射）
	if menu.Name != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tName:      %q,\n", menu.Name))
	}

	// Component
	if menu.Component != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tComponent: %q,\n", menu.Component))
	}

	// Locale
	if menu.Locale != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tLocale:    %q,\n", menu.Locale))
	}

	// Icon
	if menu.Icon != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tIcon:      %q,\n", menu.Icon))
	}

	// Redirect
	if menu.Redirect != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRedirect:  %q,\n", menu.Redirect))
	}

	// URL
	if menu.URL != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tURL:       %q,\n", menu.URL))
	}

	// KeepAlive
	if menu.KeepAlive {
		sb.WriteString("\t\t\tKeepAlive: true,\n")
	}

	// HideInMenu
	if menu.HideInMenu {
		sb.WriteString("\t\t\tHideInMenu: true,\n")
	}

	// Weight
	if menu.Weight != 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tWeight:    %d,\n", menu.Weight))
	}

	// Order
	if menu.Order != 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tOrder:     %d,\n", menu.Order))
	}

	// Status
	if menu.Status != "" && menu.Status != "1" {
		sb.WriteString(fmt.Sprintf("\t\t\tStatus:    %q,\n", menu.Status))
	}

	// Target
	if menu.Target != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTarget:    %q,\n", menu.Target))
	}

	// Href
	if menu.Href != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tHref:      %q,\n", menu.Href))
	}

	// MenuType
	if menu.MenuType != "" && menu.MenuType != "2" {
		sb.WriteString(fmt.Sprintf("\t\t\tMenuType:  %q,\n", menu.MenuType))
	}

	// MenuName
	if menu.MenuName != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tMenuName:  %q,\n", menu.MenuName))
	}

	// RouteName
	if menu.RouteName != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRouteName: %q,\n", menu.RouteName))
	}

	// RoutePath
	if menu.RoutePath != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRoutePath: %q,\n", menu.RoutePath))
	}

	// I18nKey
	if menu.I18nKey != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tI18nKey:   %q,\n", menu.I18nKey))
	}

	// IconType
	if menu.IconType != "" && menu.IconType != "1" {
		sb.WriteString(fmt.Sprintf("\t\t\tIconType:  %q,\n", menu.IconType))
	}

	// MultiTab
	if menu.MultiTab {
		sb.WriteString("\t\t\tMultiTab:  true,\n")
	}

	// ActiveMenu
	if menu.ActiveMenu != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tActiveMenu: %q,\n", menu.ActiveMenu))
	}

	// Constant
	if menu.Constant {
		sb.WriteString("\t\t\tConstant:  true,\n")
	}

	sb.WriteString("\t\t}")

	return sb.String()
}

// generateChildMenuStruct 生成子菜单的结构体代码
func (g *Generator) generateChildMenuStruct(menu model.Menu, parentKey string) string {
	var sb strings.Builder

	sb.WriteString("model.Menu{\n")

	// ParentID：使用父菜单的 key（Name 或 Path）来查找映射
	if parentKey != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tParentID:  menuNameMap[\"%s\"],\n", parentKey))
	} else {
		sb.WriteString(fmt.Sprintf("\t\t\tParentID:  %d, // 父菜单ID（注意：父菜单Name和Path都为空，请检查）\n", menu.ParentID))
	}

	// Path
	if menu.Path != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tPath:      %q,\n", menu.Path))
	}

	// Title
	if menu.Title != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTitle:     %q,\n", menu.Title))
	}

	// Name（重要：用于映射）
	if menu.Name != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tName:      %q,\n", menu.Name))
	}

	// Component
	if menu.Component != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tComponent: %q,\n", menu.Component))
	}

	// Locale
	if menu.Locale != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tLocale:    %q,\n", menu.Locale))
	}

	// Icon
	if menu.Icon != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tIcon:      %q,\n", menu.Icon))
	}

	// Redirect
	if menu.Redirect != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRedirect:  %q,\n", menu.Redirect))
	}

	// URL
	if menu.URL != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tURL:       %q,\n", menu.URL))
	}

	// KeepAlive
	if menu.KeepAlive {
		sb.WriteString("\t\t\tKeepAlive: true,\n")
	}

	// HideInMenu
	if menu.HideInMenu {
		sb.WriteString("\t\t\tHideInMenu: true,\n")
	}

	// Weight
	if menu.Weight != 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tWeight:    %d,\n", menu.Weight))
	}

	// Order
	if menu.Order != 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tOrder:     %d,\n", menu.Order))
	}

	// Status
	if menu.Status != "" && menu.Status != "1" {
		sb.WriteString(fmt.Sprintf("\t\t\tStatus:    %q,\n", menu.Status))
	}

	// Target
	if menu.Target != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTarget:    %q,\n", menu.Target))
	}

	// Href
	if menu.Href != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tHref:      %q,\n", menu.Href))
	}

	// MenuType
	if menu.MenuType != "" && menu.MenuType != "2" {
		sb.WriteString(fmt.Sprintf("\t\t\tMenuType:  %q,\n", menu.MenuType))
	}

	// MenuName
	if menu.MenuName != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tMenuName:  %q,\n", menu.MenuName))
	}

	// RouteName
	if menu.RouteName != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRouteName: %q,\n", menu.RouteName))
	}

	// RoutePath
	if menu.RoutePath != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tRoutePath: %q,\n", menu.RoutePath))
	}

	// I18nKey
	if menu.I18nKey != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tI18nKey:   %q,\n", menu.I18nKey))
	}

	// IconType
	if menu.IconType != "" && menu.IconType != "1" {
		sb.WriteString(fmt.Sprintf("\t\t\tIconType:  %q,\n", menu.IconType))
	}

	// MultiTab
	if menu.MultiTab {
		sb.WriteString("\t\t\tMultiTab:  true,\n")
	}

	// ActiveMenu
	if menu.ActiveMenu != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tActiveMenu: %q,\n", menu.ActiveMenu))
	}

	// Constant
	if menu.Constant {
		sb.WriteString("\t\t\tConstant:  true,\n")
	}

	sb.WriteString("\t\t}")

	return sb.String()
}
