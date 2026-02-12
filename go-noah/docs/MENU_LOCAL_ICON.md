# SoybeanAdmin 动态菜单使用本地图标配置指南

## 概述

SoybeanAdmin 支持两种图标类型：
- **Iconify 图标**（`iconType = "1"`）：使用 Iconify 图标库，如 `mdi:monitor-dashboard`
- **本地图标**（`iconType = "2"`）：使用项目本地的 SVG 图标文件

## 配置步骤

### 1. 数据库配置

在 `menu` 表中设置以下字段：

```sql
UPDATE menu SET 
  icon_type = '2',  -- 设置为本地图标类型
  icon = 'icon-name'  -- 本地图标文件名（不含路径和扩展名）
WHERE id = ?;
```

**字段说明：**
- `icon_type`: 
  - `"1"` = Iconify 图标（默认）
  - `"2"` = 本地图标
- `icon`: 
  - 当 `icon_type = "1"` 时：Iconify 图标名称，如 `mdi:monitor-dashboard`
  - 当 `icon_type = "2"` 时：本地图标文件名（不含路径和 `.svg` 扩展名），如 `dashboard`（对应 `src/assets/icons/dashboard.svg`）

### 2. 本地图标文件位置

将 SVG 图标文件放在前端项目的以下目录：

```
web/src/assets/icons/
```

**示例：**
- 图标文件名：`dashboard.svg`
- 数据库 `icon` 字段值：`dashboard`
- 完整路径：`web/src/assets/icons/dashboard.svg`

### 3. 代码中的处理

后端代码已经支持 `iconType` 字段：

```go
// internal/service/admin.go
iconType := menu.IconType
if iconType == "" {
    iconType = "1" // 默认 iconify
}

item := &api.SoybeanMenuDataItem{
    Icon:     menu.Icon,
    IconType: iconType,
    // ...
}
```

前端会根据 `iconType` 自动处理：
- `iconType = "1"`：使用 Iconify 图标
- `iconType = "2"`：使用本地 SVG 图标

## 示例

### 示例 1：使用本地图标

**数据库配置：**
```sql
UPDATE menu SET 
  icon_type = '2',
  icon = 'dashboard'
WHERE menu_name = '首页';
```

**文件位置：**
```
web/src/assets/icons/dashboard.svg
```

### 示例 2：使用 Iconify 图标（默认）

**数据库配置：**
```sql
UPDATE menu SET 
  icon_type = '1',  -- 或留空（默认为 "1"）
  icon = 'mdi:monitor-dashboard'
WHERE menu_name = '首页';
```

## 批量更新

如果需要批量将菜单图标改为本地图标：

```sql
-- 将所有菜单的图标类型改为本地图标
UPDATE menu SET icon_type = '2';

-- 或者只更新特定菜单
UPDATE menu SET icon_type = '2' WHERE menu_name IN ('首页', '系统管理', '数据库服务');
```

## 图标资源下载

### 推荐的免费 SVG 图标网站

#### 1. **Iconify** (https://iconify.design/)
- **特点**：提供超过 200,000 个免费图标
- **优势**：
  - 支持按图标集、分类搜索
  - 可以直接下载 SVG 格式
  - 包含 Material Design、Font Awesome、Heroicons 等多个图标库
- **使用方法**：
  1. 访问 https://iconify.design/
  2. 搜索需要的图标（如 "dashboard"、"user"、"settings"）
  3. 点击图标，选择 "Download" → "SVG"
  4. 保存到 `web/src/assets/icons/` 目录

#### 2. **Heroicons** (https://heroicons.com/)
- **特点**：由 Tailwind CSS 团队维护
- **优势**：
  - 提供简洁的 SVG 图标
  - 支持 Outline 和 Solid 两种风格
  - 完全免费，MIT 许可证
- **使用方法**：
  1. 访问 https://heroicons.com/
  2. 选择 Outline 或 Solid 风格
  3. 点击图标，复制 SVG 代码或下载文件

#### 3. **Feather Icons** (https://feathericons.com/)
- **特点**：简洁美观的线性图标
- **优势**：
  - 提供 SVG 格式下载
  - 完全免费，MIT 许可证
  - 图标风格统一
- **使用方法**：
  1. 访问 https://feathericons.com/
  2. 点击图标，复制 SVG 代码或下载文件

#### 4. **Material Design Icons** (https://materialdesignicons.com/)
- **特点**：Google Material Design 风格的图标
- **优势**：
  - 提供 SVG 格式下载
  - 完全免费
  - 图标数量庞大
- **使用方法**：
  1. 访问 https://materialdesignicons.com/
  2. 搜索需要的图标
  3. 点击图标，下载 SVG 文件

#### 5. **Remix Icon** (https://remixicon.com/)
- **特点**：超过 2,500 个图标
- **优势**：
  - 提供 SVG 格式下载
  - 完全免费，Apache 2.0 许可证
  - 图标风格现代
- **使用方法**：
  1. 访问 https://remixicon.com/
  2. 搜索或浏览图标
  3. 点击图标，下载 SVG 文件

#### 6. **Tabler Icons** (https://tabler.io/icons)
- **特点**：超过 4,000 个免费 SVG 图标
- **优势**：
  - 提供 SVG 格式下载
  - 完全免费，MIT 许可证
  - 图标风格简洁
- **使用方法**：
  1. 访问 https://tabler.io/icons
  2. 搜索或浏览图标
  3. 点击图标，下载 SVG 文件

#### 7. **Icon Park** (https://iconpark.bytedance.com/)
- **特点**：字节跳动开源图标库
- **优势**：
  - 提供数千个免费 SVG 图标
  - 支持自定义颜色和大小
  - 可直接下载 SVG 格式
- **使用方法**：
  1. 访问 https://iconpark.bytedance.com/
  2. 搜索需要的图标
  3. 自定义颜色和大小（可选）
  4. 下载 SVG 文件

### 下载步骤示例（以 Iconify 为例）

1. **访问 Iconify 网站**
   ```
   https://iconify.design/
   ```

2. **搜索图标**
   - 在搜索框输入关键词，如 "dashboard"、"user"、"settings"
   - 浏览搜索结果，选择喜欢的图标

3. **下载 SVG**
   - 点击选中的图标
   - 在弹出窗口中点击 "Download" 按钮
   - 选择 "SVG" 格式
   - 保存文件

4. **放置文件**
   - 将下载的 SVG 文件重命名为合适的名称（如 `dashboard.svg`）
   - 放到 `web/src/assets/icons/` 目录
   - 确保文件名与数据库中的 `icon` 字段值匹配

### 批量下载工具

如果需要批量下载图标，可以使用以下工具：

1. **Iconify CLI**
   ```bash
   npm install -g @iconify/cli
   iconify download mdi:dashboard --output web/src/assets/icons/
   ```

2. **Iconify API**
   - 可以通过 API 批量下载图标
   - 文档：https://iconify.design/docs/api/

### 图标命名建议

- 使用小写字母和连字符：`dashboard.svg`、`user-management.svg`
- 避免使用空格和特殊字符
- 保持命名简洁明了
- 与数据库中的 `icon` 字段值保持一致

### 常用菜单图标推荐

| 菜单名称 | 推荐图标 | 来源 | 搜索关键词 |
|---------|---------|------|-----------|
| 首页/仪表盘 | `dashboard` | Heroicons / Material Design | dashboard, home |
| 用户管理 | `user` / `users` | Heroicons / Feather | user, users, account |
| 角色管理 | `shield` / `user-group` | Heroicons / Feather | shield, role, permission |
| 菜单管理 | `menu` / `list` | Heroicons / Feather | menu, list, navigation |
| 系统设置 | `settings` / `cog` | Heroicons / Feather | settings, cog, gear |
| 数据库 | `database` | Heroicons / Feather | database, db |
| 工单 | `file-text` / `document` | Heroicons / Feather | file, document, order |
| 审核 | `check-circle` / `shield-check` | Heroicons / Feather | check, audit, verify |
| SQL查询 | `search` / `code` | Heroicons / Feather | search, query, code |
| 环境管理 | `server` / `cloud` | Heroicons / Feather | server, cloud, environment |

## 注意事项

1. **图标文件名**：数据库中的 `icon` 字段值必须与 SVG 文件名（不含扩展名）完全匹配
2. **文件格式**：本地图标必须是 SVG 格式
3. **文件路径**：图标文件必须放在 `web/src/assets/icons/` 目录下
4. **大小写敏感**：文件名大小写必须匹配
5. **默认值**：如果 `icon_type` 为空，默认为 `"1"`（Iconify）
6. **图标尺寸**：建议使用 24x24 或 20x20 的图标，确保显示效果一致
7. **图标颜色**：SVG 图标中的 `fill` 或 `stroke` 颜色会被前端主题色覆盖，建议使用 `currentColor`

## 验证

配置完成后，重启前端服务，检查菜单图标是否正确显示。如果图标未显示，请检查：

1. 数据库中的 `icon_type` 是否为 `"2"`
2. `icon` 字段值是否与 SVG 文件名匹配
3. SVG 文件是否存在于 `web/src/assets/icons/` 目录
4. 浏览器控制台是否有错误信息
5. SVG 文件格式是否正确（可以用文本编辑器打开检查）
