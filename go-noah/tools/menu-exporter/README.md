# 菜单数据导出工具

## 功能说明

此工具用于从 MySQL 数据库中读取菜单数据，并生成菜单初始化器的 Go 代码。

## 使用方法

### 基本用法

```bash
cd go-noah
go run tools/menu-exporter/main.go \
  -host localhost \
  -port 3306 \
  -user root \
  -password qynfqepwq \
  -database nunu_test \
  -output internal/server/initializer/menu.go
```

### 参数说明

- `-host`: 数据库主机地址（默认: localhost）
- `-port`: 数据库端口（默认: 3306）
- `-user`: 数据库用户名（默认: root）
- `-password`: 数据库密码（**必填**）
- `-database`: 数据库名（默认: nunu_test）
- `-output`: 输出文件路径（默认: internal/server/initializer/menu.go）

### 示例

```bash
# 从本地数据库导出
go run tools/menu-exporter/main.go \
  -password qynfqepwq

# 指定完整参数
go run tools/menu-exporter/main.go \
  -host 192.168.1.100 \
  -port 3306 \
  -user root \
  -password qynfqepwq \
  -database nunu_test \
  -output internal/server/initializer/menu.go
```

## 工作原理

1. **连接数据库**：使用 GORM 连接 MySQL 数据库
2. **读取数据**：从 `menu` 表读取所有菜单数据，按 `parent_id` 和 `weight` 排序
3. **分组处理**：
   - 父菜单：`parent_id = 0` 的菜单
   - 子菜单：`parent_id != 0` 的菜单
4. **生成代码**：
   - 生成父菜单数组
   - 生成菜单映射（通过 `name` 字段）
   - 生成子菜单数组（使用映射的 ParentID）
5. **写入文件**：将生成的代码写入 `menu.go` 文件

## 注意事项

1. **Name 字段**：确保菜单的 `name` 字段唯一且不为空，用于建立父子菜单映射
2. **数据完整性**：确保数据库中的菜单数据完整、正确
3. **备份**：生成前建议备份原 `menu.go` 文件
4. **代码检查**：生成后检查代码格式和逻辑是否正确

## 字段映射

| 数据库字段 | Go 结构体字段 | 说明 |
|-----------|--------------|------|
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
| `order` | `Order` | 排序 |

## 故障排除

### 连接数据库失败

检查：
- 数据库服务是否运行
- 用户名密码是否正确
- 网络连接是否正常
- 数据库是否存在

### 生成代码格式错误

检查：
- 菜单数据是否完整
- `name` 字段是否唯一
- 是否有特殊字符需要转义

### 子菜单映射失败

确保：
- 父菜单的 `name` 字段不为空
- 子菜单的 `parent_id` 对应正确的父菜单ID
