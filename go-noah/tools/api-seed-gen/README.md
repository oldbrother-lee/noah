# API 种子生成器（GVA 同款：源码维护）

从**测试环境**的 `api` 表读取已分类好的 API 数据，生成 **Go 源码** `internal/server/source/api_seed.go`，供「初始化数据库」时使用。**仅需从测试环境导出一次，之后在源码中手动维护该文件。**

## 使用方式

在项目根目录执行：

```bash
# 使用 config/local.yml 中的 data.db.user（测试环境 DSN）连接并生成
go run ./tools/api-seed-gen -conf config/local.yml

# 指定配置与输出路径
go run ./tools/api-seed-gen -conf config/local.yml -out internal/server/source/api_seed.go
```

## 配置说明

- `-conf`：配置文件路径，其中 **`data.db.user.driver`** 与 **`data.db.user.dsn`** 为测试环境数据库连接（例如 `config/local.yml` 第 19–20 行）。
- `-out`：输出的种子 **Go 文件**路径，默认 `internal/server/source/api_seed.go`。

## 与初始化数据库的关系

1. 在测试环境维护好 `api` 表（分组、名称、路径、方法）。
2. **运行本工具一次**，从测试环境生成 `api_seed.go`（内含 `var ApiSeedData = []ApiSeedItem{ ... }`）。
3. 之后在源码中**手动维护** `internal/server/source/api_seed.go` 即可。
4. 新环境首次启动或执行「初始化数据库」时，若 **api 表为空**，会从 `ApiSeedData` 插入种子数据；若表已有数据则跳过。

与 GVA 一致：种子数据放在 Go 源码中，便于版本管理与人工维护。
