# 工单详情页优化方案

## 目标

- 首屏快速打开：不因 `order.content`、`tasks[].sql` 过大而卡顿或超时。
- 避免大 JSON 一次性进前端内存与 DOM，减少「卡出登录」风险。
- 保持现有功能：SQL 内容、任务列表、执行结果、日志等均可用。

**设计原则**：不采用「只返回前 N 字」的截断方式，避免用户误以为数据就是这么多；以**分页**和**按需拉取**为主，这样一旦出现大工单也能正常打开。

---

## 一、整体思路

| 层级     | 策略 |
|----------|------|
| **后端** | 详情主接口不返 `order.content`、不返任务列表（或只返任务总数）；任务列表改为分页接口，且每页不包含 `sql`；工单 content、单条任务 sql 单独按需接口。 |
| **前端** | 首屏只拉工单元数据 + logs；任务表用分页接口按页加载；切到「SQL 内容」再拉 content；任务表格「查看 SQL」再拉该条 sql。 |

---

## 二、后端改动

### 2.1 详情主接口瘦身（不截断、不返大字段）

**接口**：`GET /api/v1/insight/orders/:order_id`（现有 `GetOrder`）

- **不再返回**：
  - `order.content`（整段不返，不做「前 500 字」等截断，避免误导）。
- **任务列表**：
  - **不再**在主接口里返回 `tasks` 数组；或只返回 `task_total`（任务总数）供前端分页用。
  - 任务数据一律由分页接口 `GET /orders/:order_id/tasks` 提供。

实现建议：在 repository/service 层用「轻量 Order」结构体，查询时 `SELECT` 排除 `content` 列；主接口不再调用返回全量 tasks 的逻辑，只查 order 元数据 + logs + flowInstance + 任务总数。

权限、logs、flowInstance 逻辑不变。

### 2.2 任务列表分页接口（必做）

**接口**：`GET /api/v1/insight/orders/:order_id/tasks?page=1&size=20`

- **作用**：分页返回当前工单的任务列表，**不包含** `sql` 字段。
- **参数**：`page`（从 1 起）、`size`（每页条数，建议默认 20，最大可限制如 100）。
- **响应**：`{ "list": [...], "total": N }`，list 中每条为 task 元数据（task_id、order_id、progress、result、sql_type、executor 等），**无 sql**。
- **权限**：与主详情一致（相关人员可访问）。

说明：无论任务多少都走分页，这样大工单首屏只拉第一页，不会出现「任务数不多就不做」导致大详情打不开的情况。

现有 `GET /api/v1/insight/orders/:order_id/tasks` 若无分页，可改为上述分页语义；若有其他地方依赖「全量 tasks 一次返回」，可保留一个内部或可选参数 `?all=1` 仅兼容旧调用，新详情页一律用分页。

### 2.3 按需拉取工单 content

**接口**：`GET /api/v1/insight/orders/:order_id/content`

- **作用**：仅返回当前工单的完整 `content`（LONGTEXT），用于「SQL 内容」Tab；**不截断**，按需一次拉全量。
- **权限**：与主详情一致。
- **响应**：`{ "content": "..." }` 或 `text/plain`。

路由需注册在 `/orders/:order_id` 之前（例如与 download 并列）。

若后续要支持「超长 content 分片展示」（如按行/按段分页），可再增加 `?offset=&limit=` 等参数，首期可只做全量拉取。

### 2.4 按需拉取单条任务 SQL（必做）

**接口**：`GET /api/v1/insight/orders/:order_id/tasks/:task_id/sql`

- **作用**：返回指定任务的完整 `sql`，用于任务表格「查看 SQL」弹窗；**不截断**。
- **权限**：同工单详情。
- **响应**：`{ "sql": "..." }`。

任务列表分页里不带 sql，避免单页 20 条大 SQL 一起返回；用户点「查看」再拉该条，大详情也能稳定打开。

---

## 三、前端改动

### 3.1 请求拆分

| 时机           | 请求 | 用途 |
|----------------|------|------|
| 进入详情页     | `GET /orders/:id`（瘦身后，无 content、无 tasks 列表） | 工单元数据、task_total、logs、flowInstance |
| 任务表格       | `GET /orders/:id/tasks?page=1&size=20` | 分页加载任务列表（无 sql） |
| 切换到「SQL内容」Tab | `GET /orders/:id/content` | 拉取完整 content，填入 SQL 编辑器 |
| 点击某任务「查看 SQL」 | `GET /orders/:id/tasks/:task_id/sql` | 弹窗展示该条完整 SQL |

### 3.2 详情页数据流

- **orderDetail**：只存主接口返回的 order（无 content），不存任何「前 N 字」摘要，避免界面暗示数据就这么多。
- **任务列表**：用分页请求，`tasksList` 只存当前页的轻量 task（无 sql）；表格分页器用 `task_total` 与 `page/size` 请求下一页。
- **content**：进入页面不请求；当 `activeTab === 'sql-content'` 时若尚未拉过，再请求 `GET /orders/:id/content`，结果写入 `localSqlContent`（或 `orderContent`），并设已拉取标记。
- **displaySqlContent**：用 `localSqlContent`；未加载前该 Tab 内显示 loading 或「加载 SQL」，不显示截断预览。

### 3.3 任务表格

- **数据**：来自分页接口，每行无 `sql` 字段。
- **「SQL 语句」列**：不渲染完整 SQL，改为「查看」按钮；点击后请求 `GET /orders/:id/tasks/:task_id/sql`，在弹窗/抽屉中展示完整 `sql`。
- 避免在表格行内展示长文本或「前 200 字」，防止误导且减少 DOM/内存。

### 3.4 兼容与降级

- 若主接口暂时仍返回 `tasks` 全量：前端可优先用分页接口填表，无分页时再回退到主接口的 tasks（并建议尽快后端下线主接口里的 tasks）。
- 若 `GET /orders/:id/content` 尚未提供：前端可 404 时回退到主接口带 `?full=1` 的 content（若后端保留该可选参数），默认不传 `full`。

---

## 四、实施顺序建议

1. **后端**  
   - 主接口瘦身：不返 `order.content`，不返 `tasks`（或只返 `task_total`）。  
   - 实现 `GET /orders/:order_id/tasks?page=&size=` 分页，返回 list + total，list 中不含 sql。  
   - 实现 `GET /orders/:order_id/content` 与 `GET /orders/:order_id/tasks/:task_id/sql`。  

2. **前端**  
   - 详情首屏只请求瘦身后的主接口；任务表改为分页请求并渲染分页器。  
   - 「SQL 内容」Tab 首次进入时请求 content 并展示完整内容。  
   - 任务表「SQL 语句」列改为「查看」按钮，点击请求单任务 sql 接口并在弹窗展示。  

3. **不做的**  
   - 不在任何接口中返回「前 N 字」content/sql 摘要作为主数据，避免用户误以为数据就这些。  

---

## 五、验收预期

- 大工单（content 数 MB、任务数百条）下：  
  - 首屏只拉 order 元数据 + 第一页任务（无 sql），响应小、打开快。  
  - 任务翻页、查看某条 SQL、打开 SQL 内容 Tab 均按需请求，页面不卡死、不误以为数据被截断。  
- 「卡出登录」风险降低：首屏与分页请求短小，token 不易在长请求中过期。
