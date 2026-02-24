# DAS 权限接口越权改造方案

## 背景

- `GET /api/v1/insight/das/permissions`：已改造为**仅当前登录用户**，不再接受 `username`，防越权。
- `GET /api/v1/insight/das/permissions/users?username=xxx`：仍接受任意 `username`，任意登录用户可查他人「生效权限」，存在越权风险。

## 改造目标

- 消除「按他人 username 查权限」的能力（除非是管理员专用接口）。
- 与 `GetUserPermissions` 策略一致：普通接口只服务「当前用户」。

---

## 方案 A：仅当前用户（已实现）

**思路**：`/das/permissions/users` 不再接受 `username`，只返回当前登录用户的生效权限。

| 项目 | 内容 |
|------|------|
| 接口 | `GET /api/v1/insight/das/permissions/users` |
| 变更 | 删除 query 参数 `username`；从 context 取当前 userId → username，再查该用户的生效权限 |
| 鉴权 | 未登录 401；登录后仅能查自己 |
| 前端 | `fetchGetUserEffectivePermissions()` 不再传参；若有调用处需改为「当前用户」用途 |

**优点**：实现简单，与 `GetUserPermissions` 一致，彻底杜绝越权。  
**缺点**：若后续需要「管理员查看某用户的生效权限」，需单独增加管理员接口（见方案 B 的 admin 扩展）。

---

## 方案 B：按身份分流（可选扩展）

若业务需要**管理员查看任意用户的生效权限**，可二选一：

### B1. 保留原 path，按身份分流

- 不传 `username` 或 `username` 与当前用户一致 → 返回当前用户生效权限。
- 传 `username` 且与当前用户不一致 → 校验当前用户是否为管理员（如 userID==1 或 Casbin 具备 admin 角色），是则返回该用户的生效权限，否则 403。

### B2. 新增管理员专用接口（推荐）

- 保持 `GET /api/v1/insight/das/permissions/users` 仅当前用户（方案 A）。
- 新增 `GET /api/v1/admin/das/user-effective-permissions?username=xxx`，仅注册在 admin 路由下，由 Casbin 限制为管理员可访问，返回指定用户的生效权限（PermissionObject[]）。

当前实现采用**方案 A**；若需要管理员按用户查生效权限，再按 **B2** 增加 admin 接口即可。

---

## 涉及文件

| 类型 | 文件 | 变更摘要 |
|------|------|----------|
| 后端 | `internal/handler/insight/das.go` | `GetUserEffectivePermissions` 去掉 query username，改为仅当前用户 |
| 前端 | `web/src/service/api/das.ts` | `fetchGetUserEffectivePermissions()` 不再传 `username` |

---

## 已完成的改造汇总

1. **GetUserPermissions**（`GET /api/v1/insight/das/permissions`）  
   - 仅当前用户，不传 `username`。  
   - 前端：`fetchGetUserPermissions(params?)` 仅支持 `instance_id` / `schema` 过滤。

2. **管理员按用户查 DAS 权限**  
   - 新增 `GET /api/v1/admin/das/user-permissions?username=xxx`（admin 路由，Casbin 控制）。  
   - 返回与 DAS 同格式的 `schema_permissions`、`table_permissions`。  
   - 前端：`fetchGetDASUserPermissionsForAdmin(username)`，用于「系统管理-数据库权限-用户权限」页。

3. **GetUserEffectivePermissions**（本方案）  
   - `GET /api/v1/insight/das/permissions/users` 改为仅当前用户，不再接受 `username`。  
   - 前端：`fetchGetUserEffectivePermissions()` 无参。
