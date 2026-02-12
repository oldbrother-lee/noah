-- 手动修复流程定义（适用于数据库中完全没有流程定义的情况）
-- 注意：执行前请先备份数据库！

-- 1. 创建 order_ddl 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
VALUES ('order_ddl', 'DDL工单审批流程', 'order_ddl', '用于DDL类型SQL的审批流程', 1, 1, NOW(), NOW());

-- 获取刚创建的 order_ddl 流程定义 ID
SET @ddl_flow_id = LAST_INSERT_ID();

-- 为 order_ddl 创建节点
INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
VALUES 
    (@ddl_flow_id, 'start', '开始', 'start', 1, '', '', '', '', 0, '', 'dba_approval', NOW(), NOW()),
    (@ddl_flow_id, 'dba_approval', 'DBA审批', 'approval', 2, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'dba_execute', NOW(), NOW()),
    (@ddl_flow_id, 'dba_execute', 'DBA执行', 'approval', 3, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'end', NOW(), NOW()),
    (@ddl_flow_id, 'end', '结束', 'end', 4, '', '', '', '', 0, '', '', NOW(), NOW());

-- 2. 创建 order_dml 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
VALUES ('order_dml', 'DML工单审批流程', 'order_dml', '用于DML类型SQL的审批流程', 1, 1, NOW(), NOW());

-- 获取刚创建的 order_dml 流程定义 ID
SET @dml_flow_id = LAST_INSERT_ID();

-- 为 order_dml 创建节点
INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
VALUES 
    (@dml_flow_id, 'start', '开始', 'start', 1, '', '', '', '', 0, '', 'dba_approval', NOW(), NOW()),
    (@dml_flow_id, 'dba_approval', 'DBA审批', 'approval', 2, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'dba_execute', NOW(), NOW()),
    (@dml_flow_id, 'dba_execute', 'DBA执行', 'approval', 3, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'end', NOW(), NOW()),
    (@dml_flow_id, 'end', '结束', 'end', 4, '', '', '', '', 0, '', '', NOW(), NOW());

-- 3. 创建 order_export 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
VALUES ('order_export', '数据导出审批流程', 'order_export', '用于数据导出的审批流程', 1, 1, NOW(), NOW());

-- 获取刚创建的 order_export 流程定义 ID
SET @export_flow_id = LAST_INSERT_ID();

-- 为 order_export 创建节点
INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
VALUES 
    (@export_flow_id, 'start', '开始', 'start', 1, '', '', '', '', 0, '', 'dba_approval', NOW(), NOW()),
    (@export_flow_id, 'dba_approval', 'DBA审批', 'approval', 2, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'dba_execute', NOW(), NOW()),
    (@export_flow_id, 'dba_execute', 'DBA执行', 'approval', 3, 'role', 'dba', 'any', 'to_start', 24, 'notify', 'end', NOW(), NOW()),
    (@export_flow_id, 'end', '结束', 'end', 4, '', '', '', '', 0, '', '', NOW(), NOW());

-- 4. 验证修复结果
SELECT 
    fd.id,
    fd.code,
    fd.name,
    fd.type,
    fd.status,
    COUNT(fn.id) as node_count
FROM flow_definitions fd
LEFT JOIN flow_nodes fn ON fd.id = fn.flow_def_id
WHERE fd.type IN ('order_ddl', 'order_dml', 'order_export')
GROUP BY fd.id, fd.code, fd.name, fd.type, fd.status
ORDER BY fd.type;
