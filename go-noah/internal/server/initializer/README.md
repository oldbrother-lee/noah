# 初始化器系统

## 概述

本目录实现了基于接口的初始化器系统，参考了 gin-vue-admin 的设计，提供了更好的代码组织和依赖管理。

## 架构设计

### 核心组件

1. **Initializer 接口** (`interface.go`)
   - 定义了所有初始化器必须实现的方法
   - 包括表迁移、数据初始化、状态检查等

2. **注册表** (`registry.go`)
   - 管理所有初始化器的注册
   - 支持按顺序排序
   - 防止名称冲突

3. **初始化器实现**
   - `role.go`: 角色初始化器
   - `user.go`: 用户初始化器
   - `rbac.go`: RBAC 权限初始化器

4. **管理器** (`manager.go`)
   - 统一管理初始化流程
   - 支持完整初始化和按需初始化

## 使用方法

### 1. 注册初始化器

```go
import "go-noah/internal/server/initializer"

// 注册所有初始化器
initializer.RegisterAll(logger, enforcer)
```

### 2. 执行初始化

```go
// 创建管理器
manager := initializer.NewManager(db, logger)

// 完整初始化（用于迁移工具）
ctx := context.Background()
err := manager.InitializeAll(ctx)

// 按需初始化（用于服务器启动）
err := manager.InitializeIfNeeded(ctx)
```

### 3. 添加新的初始化器

1. 创建新的初始化器文件，实现 `Initializer` 接口
2. 在 `register.go` 中注册
3. 在 `constants.go` 中定义初始化顺序

示例：

```go
// my_initializer.go
type MyInitializer struct {
    logger *log.Logger
}

func (m *MyInitializer) Name() string {
    return "my_initializer"
}

func (m *MyInitializer) Order() int {
    return InitOrderMyFeature
}

func (m *MyInitializer) MigrateTable(ctx context.Context, db *gorm.DB) error {
    return db.AutoMigrate(&model.MyModel{})
}

// ... 实现其他方法
```

## 初始化顺序

初始化顺序由 `Order()` 方法返回的数字决定，数字越小越先执行。

当前顺序：
- 100: Role（角色）
- 200: User（用户）
- 300: Menu（菜单）
- 400: API（接口）
- 500: Dept（部门）
- 600: UserRole（用户角色绑定）
- 700: RBAC（权限策略）
- 800: Flow（流程定义）
- 900: Inspect（审核参数）
- 1000: Insight（业务功能）

## 依赖关系

初始化器之间可以通过 `context.Context` 传递数据：

```go
// 在初始化器中存入数据
ctx = context.WithValue(ctx, "roles", roles)

// 在后续初始化器中获取数据
roles, _ := ctx.Value("roles").([]model.Role)
```

## 注意事项

1. **幂等性**: 所有初始化器都应该支持多次调用而不产生副作用
2. **错误处理**: 初始化失败不应该阻止服务启动（记录警告即可）
3. **向后兼容**: 保留旧的初始化函数，逐步迁移

## 迁移计划

### 已完成
- ✅ 接口定义和注册机制
- ✅ Role 初始化器
- ✅ User 初始化器
- ✅ RBAC 初始化器
- ✅ 管理器实现
- ✅ 集成到服务器启动流程

### 待完成
- ⏳ Menu 初始化器
- ⏳ API 初始化器（已由路由自动同步）
- ⏳ Flow 初始化器
- ⏳ Inspect 初始化器
- ⏳ 迁移工具集成

## 参考

- [gin-vue-admin 初始化机制](https://github.com/flipped-aurora/gin-vue-admin)
- [改进方案文档](../../../DATABASE_INITIALIZATION_COMPARISON.md)
