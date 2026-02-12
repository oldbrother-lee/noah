# 菜单数据初始化维护方案

## 问题描述

当前菜单初始化器中的菜单数据是硬编码的，但实际开发中菜单数据存储在 MySQL 数据库（`nunu_test.menu` 表）中。需要将数据库中的完整菜单数据同步到菜单初始化器中。

## 方案对比

### 方案一：创建数据导出工具（推荐）

**优点：**
- 自动化程度高，一键生成代码
- 可重复使用，后续更新方便
- 支持增量更新
- 类型安全，直接生成 Go 代码

**缺点：**
- 需要开发一个小工具
- 需要配置数据库连接

**实现步骤：**
1. 创建一个命令行工具 `tools/menu-exporter/main.go`
2. 连接 MySQL 数据库，读取 `menu` 表数据
3. 按照父子关系组织菜单数据
4. 生成 Go 代码，更新 `internal/server/initializer/menu.go`

**使用方式：**
```bash
go run tools/menu-exporter/main.go \
  -host localhost \
  -port 3306 \
  -user root \
  -password qynfqepwq \
  -database nunu_test \
  -output internal/server/initializer/menu.go
```

---

### 方案二：SQL 导出 + 手动转换

**优点：**
- 实现简单，不需要开发工具
- 可以手动筛选和调整数据

**缺点：**
- 手动操作，容易出错
- 每次更新都需要重新操作
- 效率较低

**实现步骤：**
1. 从数据库导出菜单数据：
   ```sql
   SELECT * FROM menu ORDER BY parent_id, weight;
   ```
2. 导出为 JSON 或 CSV 格式
3. 手动转换为 Go 代码结构
4. 更新 `menu.go` 文件

---

### 方案三：从数据库直接读取（不推荐）

**优点：**
- 不需要维护代码中的菜单数据
- 数据始终与数据库同步

**缺点：**
- 初始化依赖数据库，不利于新环境部署
- 无法版本控制菜单数据
- 与 gin-vue-admin 的设计理念不符

---

## 推荐方案：方案一（数据导出工具）

### 工具设计

#### 1. 工具结构
```
tools/menu-exporter/
├── main.go          # 主程序
├── generator.go     # 代码生成逻辑
└── README.md        # 使用说明
```

#### 2. 功能特性
- 连接 MySQL 数据库
- 读取 `menu` 表所有数据
- 按 `parent_id` 组织菜单层级
- 生成符合 `gin-vue-admin` 风格的 Go 代码
- 自动更新 `menu.go` 文件
- 支持字段映射（数据库字段 -> Go 结构体字段）

#### 3. 生成的代码格式

```go
// 父级菜单
parentMenus := []model.Menu{
    {
        ParentID:  0,
        Path:      "/dashboard",
        Title:     "仪表盘",
        Name:      "Dashboard",
        Component: "RouteView",
        Redirect:  "/dashboard/analysis",
        Icon:      "DashboardOutlined",
        Locale:    "menu.dashboard",
        Weight:    1,
    },
    // ... 更多父菜单
}

// 创建父菜单后建立映射
menuNameMap := make(map[string]uint)
for _, menu := range parentMenus {
    menuNameMap[menu.Name] = menu.ID
}

// 子菜单
childMenus := []model.Menu{
    {
        ParentID:  menuNameMap["Dashboard"],
        Path:      "/dashboard/analysis",
        Title:     "分析页",
        Name:      "DashboardAnalysis",
        Component: "/dashboard/analysis",
        Icon:      "DashboardOutlined",
        Locale:    "menu.dashboard.analysis",
        Weight:    1,
        KeepAlive: true,
    },
    // ... 更多子菜单
}
```

#### 4. 字段映射规则

| 数据库字段 | Go 结构体字段 | 说明 |
|-----------|--------------|------|
| `id` | 不生成（自动生成） | 数据库自动生成 |
| `parent_id` | `ParentID` | 父菜单ID，0表示根菜单 |
| `path` | `Path` | 菜单路径 |
| `title` | `Title` | 菜单标题 |
| `name` | `Name` | 菜单名称（用于映射） |
| `component` | `Component` | 组件路径 |
| `locale` | `Locale` | 国际化key |
| `icon` | `Icon` | 图标 |
| `redirect` | `Redirect` | 重定向路径 |
| `url` | `URL` | iframe URL |
| `keep_alive` | `KeepAlive` | 是否保活 |
| `hide_in_menu` | `HideInMenu` | 是否隐藏 |
| `weight` | `Weight` | 排序权重 |
| `order` | `Order` | 排序（如果存在） |

#### 5. 处理逻辑

1. **读取数据**：从数据库读取所有菜单，按 `parent_id` 和 `weight` 排序
2. **分组处理**：
   - 父菜单：`parent_id = 0` 的菜单
   - 子菜单：`parent_id != 0` 的菜单
3. **建立映射**：使用 `name` 字段作为映射 key（如果 `name` 为空，使用 `path`）
4. **生成代码**：
   - 生成父菜单数组
   - 生成映射代码
   - 生成子菜单数组（使用映射的 ParentID）

#### 6. 特殊处理

- **空值处理**：空字符串字段不生成（使用 Go 的零值）
- **布尔值**：`keep_alive`、`hide_in_menu` 等布尔字段
- **多级菜单**：如果有多级菜单（三级及以上），需要递归处理
- **ID 处理**：不生成 `ID` 字段，让数据库自动生成

---

## 实现细节

### 工具代码结构

```go
// main.go
package main

import (
    "flag"
    "fmt"
    "go-noah/tools/menu-exporter/generator"
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
    
    gen := generator.NewGenerator(*host, *port, *user, *password, *database)
    if err := gen.Generate(*output); err != nil {
        fmt.Printf("生成失败: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("菜单代码生成成功！")
}
```

### 生成器核心逻辑

```go
// generator.go
type Generator struct {
    db *gorm.DB
}

func (g *Generator) Generate(outputPath string) error {
    // 1. 读取所有菜单
    var menus []model.Menu
    if err := g.db.Order("parent_id, weight").Find(&menus).Error; err != nil {
        return err
    }
    
    // 2. 分组：父菜单和子菜单
    parentMenus := []model.Menu{}
    childMenus := []model.Menu{}
    for _, menu := range menus {
        if menu.ParentID == 0 {
            parentMenus = append(parentMenus, menu)
        } else {
            childMenus = append(childMenus, menu)
        }
    }
    
    // 3. 生成 Go 代码
    code := g.generateCode(parentMenus, childMenus)
    
    // 4. 写入文件
    return os.WriteFile(outputPath, []byte(code), 0644)
}
```

---

## 使用流程

### 第一次使用

1. 开发工具代码
2. 运行工具导出菜单数据
3. 检查生成的代码
4. 提交到代码库

### 后续更新

1. 在数据库中修改菜单数据
2. 运行工具重新生成代码
3. 检查差异
4. 提交更新

---

## 注意事项

1. **数据一致性**：确保数据库中的菜单数据是完整的、正确的
2. **字段完整性**：检查所有必要字段是否都有值
3. **Name 唯一性**：确保 `name` 字段唯一，用于建立映射
4. **多级菜单**：如果有多级菜单，需要递归处理
5. **代码格式**：生成的代码需要符合 Go 代码规范
6. **备份**：生成前备份原文件

---

## 替代方案：简化版本

如果不想开发完整工具，可以创建一个简单的 SQL 查询脚本：

```sql
-- 导出菜单数据为 INSERT 语句格式
SELECT CONCAT(
    'model.Menu{',
    'ParentID: ', IFNULL(parent_id, 0), ', ',
    'Path: "', IFNULL(path, ''), '", ',
    'Title: "', IFNULL(title, ''), '", ',
    'Name: "', IFNULL(name, ''), '", ',
    'Component: "', IFNULL(component, ''), '", ',
    'Locale: "', IFNULL(locale, ''), '", ',
    'Icon: "', IFNULL(icon, ''), '", ',
    'Weight: ', IFNULL(weight, 0),
    '},'
) AS menu_code
FROM menu
ORDER BY parent_id, weight;
```

然后手动整理成 Go 代码。

---

## 建议

**推荐使用方案一**，因为：
1. 自动化程度高，减少人工错误
2. 可重复使用，长期维护成本低
3. 符合 gin-vue-admin 的设计理念
4. 可以扩展支持其他表的初始化数据导出

如果时间紧迫，可以先使用方案二（SQL 导出 + 手动转换），后续再开发工具。
