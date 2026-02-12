-- =====================================================
-- goInsight 数据迁移脚本
-- 用于将 goInsight 数据库中的数据迁移到 go-noah
-- =====================================================

-- 注意事项:
-- 1. 执行前请先备份数据库
-- 2. 根据实际表结构调整字段映射
-- 3. 迁移用户数据时注意密码兼容性（两边都用 bcrypt）

-- =====================================================
-- 1. 环境数据迁移
-- =====================================================
INSERT INTO db_environments (id, name, created_at, updated_at)
SELECT id, name, created_at, updated_at
FROM insight_db_environments
ON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = VALUES(updated_at);

-- =====================================================
-- 2. 数据库配置迁移
-- =====================================================
INSERT INTO db_configs (
    id, instance_id, hostname, port, user_name, password,
    use_type, db_type, environment, inspect_params,
    organization_key, organization_path, remark,
    created_at, updated_at
)
SELECT
    id, instance_id, hostname, port, user_name, password,
    use_type, db_type, environment, inspect_params,
    organization_key, organization_path, remark,
    created_at, updated_at
FROM insight_db_config
ON DUPLICATE KEY UPDATE
    hostname = VALUES(hostname),
    port = VALUES(port),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 3. Schema 信息迁移
-- =====================================================
INSERT INTO db_schemas (id, instance_id, `schema`, is_deleted, created_at, updated_at)
SELECT id, instance_id, `schema`, is_deleted, created_at, updated_at
FROM insight_db_schemas
ON DUPLICATE KEY UPDATE
    is_deleted = VALUES(is_deleted),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 4. 组织架构迁移
-- =====================================================
INSERT INTO organizations (id, name, parent_id, `key`, level, path, creator, updater, created_at, updated_at)
SELECT id, name, parent_id, `key`, level, path, creator, updater, created_at, updated_at
FROM insight_organizations
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    `key` = VALUES(`key`),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 5. 组织用户关联迁移
-- =====================================================
INSERT INTO organization_users (id, uid, organization_key, created_at, updated_at)
SELECT id, uid, organization_key, created_at, updated_at
FROM insight_organizations_users
ON DUPLICATE KEY UPDATE
    organization_key = VALUES(organization_key),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 6. DAS 库权限迁移
-- =====================================================
INSERT INTO das_user_schema_permissions (id, username, `schema`, instance_id, created_at, updated_at)
SELECT id, username, `schema`, instance_id, created_at, updated_at
FROM insight_das_user_schema_permissions
ON DUPLICATE KEY UPDATE
    updated_at = VALUES(updated_at);

-- =====================================================
-- 7. DAS 表权限迁移
-- =====================================================
INSERT INTO das_user_table_permissions (id, username, `schema`, `table`, instance_id, rule, created_at, updated_at)
SELECT id, username, `schema`, `table`, instance_id, rule, created_at, updated_at
FROM insight_das_user_table_permissions
ON DUPLICATE KEY UPDATE
    rule = VALUES(rule),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 8. DAS 允许的操作迁移
-- =====================================================
INSERT INTO das_allowed_operations (id, name, is_enable, remark, created_at, updated_at)
SELECT id, name, is_enable, remark, created_at, updated_at
FROM insight_das_allowed_operations
ON DUPLICATE KEY UPDATE
    is_enable = VALUES(is_enable),
    remark = VALUES(remark),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 9. DAS 执行记录迁移（如果需要历史数据）
-- =====================================================
-- 注意: 此表数据量可能很大，根据需要决定是否迁移
-- INSERT INTO das_records (id, username, instance_id, `schema`, `sql`, duration, row_count, error, created_at, updated_at)
-- SELECT id, username, instance_id, `schema`, `sql`, duration, row_count, error, created_at, updated_at
-- FROM insight_das_records;

-- =====================================================
-- 10. DAS 收藏夹迁移
-- =====================================================
INSERT INTO das_favorites (id, username, title, `sql`, created_at, updated_at)
SELECT id, username, title, `sql`, created_at, updated_at
FROM insight_das_favorites
ON DUPLICATE KEY UPDATE
    title = VALUES(title),
    `sql` = VALUES(`sql`),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 11. 审核参数迁移
-- =====================================================
INSERT INTO inspect_params (id, params, remark, created_at, updated_at)
SELECT id, params, remark, created_at, updated_at
FROM insight_inspect_params
ON DUPLICATE KEY UPDATE
    params = VALUES(params),
    remark = VALUES(remark),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 12. 工单记录迁移
-- =====================================================
INSERT INTO order_records (
    id, title, order_id, hook_order_id, remark, is_restrict_access,
    db_type, sql_type, environment, applicant, organization,
    approver, executor, reviewer, cc, instance_id, `schema`,
    progress, execute_result, schedule_time, fix_version, content,
    export_file_format, created_at, updated_at
)
SELECT
    id, title, order_id, hook_order_id, remark, is_restrict_access,
    db_type, sql_type, environment, applicant, organization,
    approver, executor, reviewer, cc, instance_id, `schema`,
    progress, execute_result, schedule_time, fix_version, content,
    export_file_format, created_at, updated_at
FROM insight_order_records
ON DUPLICATE KEY UPDATE
    progress = VALUES(progress),
    execute_result = VALUES(execute_result),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 13. 工单任务迁移
-- =====================================================
INSERT INTO order_tasks (
    id, order_id, task_id, db_type, sql_type, executor,
    `sql`, progress, result, created_at, updated_at
)
SELECT
    id, order_id, task_id, db_type, sql_type, executor,
    `sql`, progress, result, created_at, updated_at
FROM insight_order_tasks
ON DUPLICATE KEY UPDATE
    progress = VALUES(progress),
    result = VALUES(result),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 14. 工单操作日志迁移
-- =====================================================
INSERT INTO order_op_logs (id, username, order_id, msg, created_at, updated_at)
SELECT id, username, order_id, msg, created_at, updated_at
FROM insight_order_oplogs
ON DUPLICATE KEY UPDATE
    msg = VALUES(msg),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 15. 工单消息记录迁移
-- =====================================================
INSERT INTO order_messages (id, order_id, receiver, response, created_at, updated_at)
SELECT id, order_id, receiver, response, created_at, updated_at
FROM insight_order_messages
ON DUPLICATE KEY UPDATE
    response = VALUES(response),
    updated_at = VALUES(updated_at);

-- =====================================================
-- 16. 用户数据迁移（如果需要合并用户）
-- =====================================================
-- 注意: go-noah 使用 admin_users 表，goInsight 使用 insight_users 表
-- 需要根据实际情况决定合并策略

-- 方案A: 将 insight_users 合并到 admin_users（保留 go-noah 现有用户）
-- INSERT INTO admin_users (username, nickname, password, email, phone, created_at, updated_at)
-- SELECT username, nick_name, password, email, mobile, date_joined, updated_at
-- FROM insight_users
-- WHERE username NOT IN (SELECT username FROM admin_users)
-- ON DUPLICATE KEY UPDATE
--     nickname = VALUES(nickname),
--     email = VALUES(email),
--     phone = VALUES(phone),
--     updated_at = VALUES(updated_at);

-- 方案B: 创建新的用户扩展表保存 goInsight 特有字段
-- CREATE TABLE IF NOT EXISTS admin_user_ext (
--     id BIGINT UNSIGNED PRIMARY KEY,
--     admin_user_id INT UNSIGNED NOT NULL,
--     avatar_file VARCHAR(254),
--     is_superuser BOOLEAN DEFAULT FALSE,
--     is_active BOOLEAN DEFAULT TRUE,
--     is_staff BOOLEAN DEFAULT FALSE,
--     is_two_fa BOOLEAN DEFAULT FALSE,
--     otp_secret VARCHAR(128),
--     last_login DATETIME,
--     FOREIGN KEY (admin_user_id) REFERENCES admin_users(id)
-- );

-- =====================================================
-- 验证迁移结果
-- =====================================================
SELECT 'db_environments' as table_name, COUNT(*) as row_count FROM db_environments
UNION ALL
SELECT 'db_configs', COUNT(*) FROM db_configs
UNION ALL
SELECT 'db_schemas', COUNT(*) FROM db_schemas
UNION ALL
SELECT 'organizations', COUNT(*) FROM organizations
UNION ALL
SELECT 'organization_users', COUNT(*) FROM organization_users
UNION ALL
SELECT 'das_user_schema_permissions', COUNT(*) FROM das_user_schema_permissions
UNION ALL
SELECT 'das_user_table_permissions', COUNT(*) FROM das_user_table_permissions
UNION ALL
SELECT 'das_allowed_operations', COUNT(*) FROM das_allowed_operations
UNION ALL
SELECT 'das_favorites', COUNT(*) FROM das_favorites
UNION ALL
SELECT 'inspect_params', COUNT(*) FROM inspect_params
UNION ALL
SELECT 'order_records', COUNT(*) FROM order_records
UNION ALL
SELECT 'order_tasks', COUNT(*) FROM order_tasks
UNION ALL
SELECT 'order_op_logs', COUNT(*) FROM order_op_logs
UNION ALL
SELECT 'order_messages', COUNT(*) FROM order_messages;

