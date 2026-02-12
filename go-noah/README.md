# Noah - 数据库管理平台

Noah 是一个企业级数据库管理平台，提供 SQL 审核、工单管理、数据访问服务（DAS）等核心功能，帮助企业规范数据库操作流程，提升数据库安全性和运维效率。

## ✨ 核心功能

### 🔍 SQL 审核引擎
- **智能 SQL 解析**：基于 TiDB Parser 进行 SQL 语法解析和验证
- **多维度审核规则**：支持 DDL、DML、DROP 等各类 SQL 语句审核
- **审核级别**：PASS、INFO、NOTICE、WARNING、ERROR 五级审核
- **修复建议**：自动提供 SQL 优化和修复建议
- **可配置审核参数**：支持自定义审核规则和阈值

### 📋 工单管理系统
- **全流程工单管理**：从创建、审核、执行到复核的完整流程
- **多类型工单支持**：DML、DDL、EXPORT 等不同类型工单
- **审批流程**：支持多级审批和自定义审批流程
- **定时执行**：支持 cron 表达式定时执行工单
- **在线 DDL**：集成 gh-ost 支持在线表结构变更
- **实时进度推送**：WebSocket 实时推送工单执行进度
- **回滚支持**：自动生成回滚 SQL，支持快速回滚

### 🔐 数据访问服务（DAS）
- **安全 SQL 查询**：提供安全的数据库查询接口
- **细粒度权限控制**：支持 Schema 和 Table 级别的权限管理
- **查询限制**：自动添加 LIMIT 限制，防止大查询影响数据库
- **执行时间控制**：支持最大执行时间限制
- **查询记录**：完整的查询历史记录和审计
- **收藏功能**：常用 SQL 收藏和管理
- **权限模板**：支持权限模板快速授权

### 👥 权限管理系统
- **RBAC 权限模型**：基于 Casbin 实现角色权限控制
- **多级权限控制**：支持 API 和菜单级别的权限控制
- **LDAP 集成**：支持 LDAP 统一认证
- **部门管理**：组织架构和部门管理
- **用户管理**：完整的用户、角色、权限管理

### 🏢 组织架构管理
- **树形组织结构**：支持多级组织架构
- **环境管理**：开发、测试、生产等多环境管理
- **数据库配置管理**：统一管理数据库连接配置
- **用户组织绑定**：灵活的用户与组织关系管理

## 🚀 技术栈

### 后端技术
- **语言**：Go 1.24+
- **Web 框架**：Gin
- **ORM**：GORM（支持 MySQL、PostgreSQL、SQLite）
- **权限管理**：Casbin（RBAC）
- **认证**：JWT
- **SQL 解析**：TiDB Parser
- **缓存**：Redis
- **日志**：Zap + Lumberjack
- **定时任务**：gocron
- **WebSocket**：Gorilla WebSocket

### 前端技术
- **框架**：Vue 3 + TypeScript
- **UI 组件**：SoybeanAdmin（基于 NaiveUI）
- **构建工具**：Vite 7
- **状态管理**：Pinia
- **样式方案**：UnoCSS
- **SQL 编辑器**：CodeMirror / Monaco Editor

## 📋 系统要求

- **Go**：1.24 或更高版本
- **Node.js**：18 或更高版本
- **数据库**：MySQL 5.7+ / PostgreSQL 10+ / SQLite 3
- **Redis**：5.0+（可选，用于缓存和 WebSocket 消息推送）

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd go-noah
```

### 2. 配置数据库

编辑 `config/local.yml` 文件，配置数据库连接信息：

```yaml
data:
  db:
    user:
      driver: mysql
      dsn: root:password@tcp(127.0.0.1:3306)/noah?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    db: 0
```

### 3. 初始化数据库

执行数据库迁移，初始化表结构和基础数据：

```bash
go run cmd/migration/main.go
```

### 4. 启动后端服务

```bash
go run cmd/server/main.go
```

后端服务默认运行在 `http://localhost:8000`

### 5. 启动定时任务服务（可选）

如果需要定时任务功能（如定时执行工单），需要单独启动任务服务：

```bash
go run cmd/task/main.go
```

### 6. 启动前端服务

```bash
cd web
pnpm install
pnpm run dev
```

前端服务默认运行在 `http://localhost:6678`

### 7. 访问系统

- **前端地址**：http://localhost:6678
- **后端 API**：http://localhost:8000
- **API 文档**：http://localhost:8000/swagger/index.html

**默认账号**：
- 超管账号：`admin`
- 超管密码：`123456`

## 📦 项目结构

```
go-noah/
├── cmd/                    # 入口程序
│   ├── server/            # HTTP 服务器
│   ├── task/              # 定时任务服务
│   └── migration/         # 数据库迁移工具
├── internal/              # 内部业务代码
│   ├── handler/          # HTTP 请求处理器
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # 中间件（JWT、权限、日志等）
│   ├── router/           # 路由配置
│   ├── inspect/          # SQL 审核模块
│   ├── orders/           # 工单执行器
│   ├── das/              # DAS 数据访问服务
│   └── task/             # 定时任务
├── pkg/                  # 公共包
│   ├── jwt/             # JWT 认证
│   ├── log/             # 日志组件
│   ├── config/          # 配置管理
│   ├── noah/            # 应用初始化
│   └── ...
├── api/                  # API 定义和响应处理
├── config/               # 配置文件
│   ├── local.yml        # 本地开发配置
│   └── prod.yml         # 生产环境配置
├── deploy/               # 部署相关
│   ├── build/           # Docker 构建文件
│   └── docker-compose/   # Docker Compose 配置
├── scripts/              # SQL 脚本
└── web/                  # 前端项目
```

## 🔧 配置说明

主要配置项位于 `config/local.yml`：

```yaml
# HTTP 服务配置
http:
  host: 127.0.0.1
  port: 8000

# 数据库配置
data:
  db:
    user:
      driver: mysql
      dsn: root:password@tcp(127.0.0.1:3306)/noah?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    db: 0

# LDAP 认证配置（可选）
ldap:
  enable: true
  host: ldap.example.com
  port: 389
  base_dn: "dc=example,dc=com"

# gh-ost 配置（在线 DDL）
ghost:
  path: "/usr/local/bin/gh-ost"
  args:
    - "--allow-on-master"
    - "--assume-rbr"

# DAS 查询限制
das:
  max_execution_time: 600000    # 最大执行时间（毫秒）
  default_return_rows: 1000     # 默认返回行数
  max_return_rows: 10000        # 最大返回行数

# 定时任务配置
crontab:
  sync_db_metas: "*/5 * * * *"              # 同步数据库元数据
  scan_scheduled_orders: "*/30 * * * * *"  # 扫描定时工单

# 消息通知配置
notify:
  dingtalk:
    enable: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"

# LLM 配置（可选，用于「权限管理 - 同步路由」中的「AI 自动填充」API 名称与分组）
# 兼容 OpenAI / 国内大模型（DeepSeek、通义、月之暗面等 OpenAI 兼容接口）
llm:
  enable: false
  base_url: "https://api.openai.com/v1"   # 或国内代理/本地模型地址
  api_key: ""                              # 也可不填，改用环境变量 LLM_API_KEY
  model: "gpt-3.5-turbo"
```

## 📦 打包部署

### 开发环境打包

```bash
# 打包当前平台
make build

# 打包所有平台
make build-all
```

### 生产环境部署

#### 方式一：单二进制部署（推荐）

```bash
# 1. 构建前端
cd web
pnpm run build

# 2. 构建后端（包含前端静态资源）
cd ../go-noah
go build -o server cmd/server/main.go

# 3. 运行
./server
```

#### 方式二：Docker 部署

```bash
# 构建镜像
docker build -f deploy/build/Dockerfile -t noah:latest .

# 使用 Docker Compose
cd deploy/docker-compose
docker compose up -d
```

#### 方式三：前后端分离部署

- **后端**：部署 Go 二进制文件或 Docker 容器
- **前端**：使用 Nginx 部署静态资源，反向代理到后端 API

## 🔑 权限管理

系统采用 RBAC（基于角色的访问控制）模型，使用 Casbin 进行权限管理。

### 权限策略示例

**API 接口权限**：
```
p, admin, api:/api/v1/insight/orders, GET
p, admin, api:/api/v1/insight/orders, POST
p, user, api:/api/v1/insight/orders/my, GET
```

**菜单权限**：
```
p, admin, menu:/insight/orders, read
p, user, menu:/insight/orders/my, read
```

### 权限操作流程

1. **添加 API 接口**：权限模块 → 接口管理 → 添加 API
2. **添加前端菜单**：权限模块 → 菜单管理 → 添加菜单
3. **分配权限**：权限模块 → 角色管理 → 添加角色/分配权限

## 📚 核心功能使用

### SQL 审核

1. 进入 SQL 审核页面
2. 输入或粘贴 SQL 语句
3. 选择审核参数（或使用默认参数）
4. 点击审核，查看审核结果和修复建议

### 创建工单

1. 进入工单管理 → 创建工单
2. 填写工单信息（标题、SQL、环境、数据库等）
3. 选择审批流程
4. 提交工单，等待审核

### DAS 查询

1. 进入 DAS 查询页面
2. 选择数据库实例和 Schema
3. 编写 SQL 查询语句
4. 执行查询，查看结果

## 🛠️ 开发指南

### 添加新的 API 接口

1. 在 `internal/handler` 中创建或修改处理器
2. 在 `internal/router` 中注册路由
3. 在 `api` 包中定义 API 结构
4. 在权限管理中添加 API 接口配置

### 添加新的审核规则

1. 在 `internal/inspect/rules` 中添加规则文件
2. 在 `internal/inspect/checker` 中注册规则
3. 在审核参数配置中添加规则开关

## 📝 API 文档

启动服务后，访问 `http://localhost:8000/swagger/index.html` 查看完整的 API 文档。

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

## 📜 许可证

本项目基于 **MIT License** 开源。

## 📞 联系方式

如有问题或建议，请提交 Issue 或联系项目维护者。