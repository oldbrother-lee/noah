-- 修复流程定义（如果不存在或未启用）
-- 注意：执行前请先备份数据库！

-- 1. 检查并创建 order_ddl 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
SELECT 
    'order_ddl',
    'DDL工单审批流程',
    'order_ddl',
    '用于DDL类型SQL的审批流程',
    1,
    1,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM flow_definitions WHERE type = 'order_ddl' AND status = 1
);

-- 2. 检查并创建 order_dml 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
SELECT 
    'order_dml',
    'DML工单审批流程',
    'order_dml',
    '用于DML类型SQL的审批流程',
    1,
    1,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM flow_definitions WHERE type = 'order_dml' AND status = 1
);

-- 3. 检查并创建 order_export 流程定义
INSERT INTO flow_definitions (code, name, type, description, version, status, created_at, updated_at)
SELECT 
    'order_export',
    '数据导出审批流程',
    'order_export',
    '用于数据导出的审批流程',
    1,
    1,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM flow_definitions WHERE type = 'order_export' AND status = 1
);

-- 4. 如果流程定义存在但未启用，启用它们
UPDATE flow_definitions 
SET status = 1, updated_at = NOW()
WHERE type IN ('order_ddl', 'order_dml', 'order_export') AND status != 1;

-- 5. 为 order_ddl 创建默认节点（如果不存在）
SET @ddl_flow_id = (SELECT id FROM flow_definitions WHERE type = 'order_ddl' AND status = 1 LIMIT 1);

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @ddl_flow_id,
    'start',
    '开始',
    'start',
    1,
    '',
    '',
    '',
    '',
    0,
    '',
    'dba_approval',
    NOW(),
    NOW()
WHERE @ddl_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @ddl_flow_id AND node_code = 'start');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @ddl_flow_id,
    'dba_approval',
    'DBA审批',
    'approval',
    2,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'dba_execute',
    NOW(),
    NOW()
WHERE @ddl_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @ddl_flow_id AND node_code = 'dba_approval');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @ddl_flow_id,
    'dba_execute',
    'DBA执行',
    'approval',
    3,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'end',
    NOW(),
    NOW()
WHERE @ddl_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @ddl_flow_id AND node_code = 'dba_execute');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @ddl_flow_id,
    'end',
    '结束',
    'end',
    4,
    '',
    '',
    '',
    '',
    0,
    '',
    '',
    NOW(),
    NOW()
WHERE @ddl_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @ddl_flow_id AND node_code = 'end');

-- 6. 为 order_dml 创建默认节点（如果不存在）
SET @dml_flow_id = (SELECT id FROM flow_definitions WHERE type = 'order_dml' AND status = 1 LIMIT 1);

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @dml_flow_id,
    'start',
    '开始',
    'start',
    1,
    '',
    '',
    '',
    '',
    0,
    '',
    'dba_approval',
    NOW(),
    NOW()
WHERE @dml_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @dml_flow_id AND node_code = 'start');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @dml_flow_id,
    'dba_approval',
    'DBA审批',
    'approval',
    2,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'dba_execute',
    NOW(),
    NOW()
WHERE @dml_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @dml_flow_id AND node_code = 'dba_approval');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @dml_flow_id,
    'dba_execute',
    'DBA执行',
    'approval',
    3,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'end',
    NOW(),
    NOW()
WHERE @dml_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @dml_flow_id AND node_code = 'dba_execute');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @dml_flow_id,
    'end',
    '结束',
    'end',
    4,
    '',
    '',
    '',
    '',
    0,
    '',
    '',
    NOW(),
    NOW()
WHERE @dml_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @dml_flow_id AND node_code = 'end');

-- 7. 为 order_export 创建默认节点（如果不存在）
SET @export_flow_id = (SELECT id FROM flow_definitions WHERE type = 'order_export' AND status = 1 LIMIT 1);

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @export_flow_id,
    'start',
    '开始',
    'start',
    1,
    '',
    '',
    '',
    '',
    0,
    '',
    'dba_approval',
    NOW(),
    NOW()
WHERE @export_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @export_flow_id AND node_code = 'start');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @export_flow_id,
    'dba_approval',
    'DBA审批',
    'approval',
    2,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'dba_execute',
    NOW(),
    NOW()
WHERE @export_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @export_flow_id AND node_code = 'dba_approval');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @export_flow_id,
    'dba_execute',
    'DBA执行',
    'approval',
    3,
    'role',
    'dba',
    'any',
    'to_start',
    24,
    'notify',
    'end',
    NOW(),
    NOW()
WHERE @export_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @export_flow_id AND node_code = 'dba_execute');

INSERT INTO flow_nodes (flow_def_id, node_code, node_name, node_type, sort, approver_type, approver_ids, multi_mode, reject_action, timeout_hours, timeout_action, next_node_code, created_at, updated_at)
SELECT 
    @export_flow_id,
    'end',
    '结束',
    'end',
    4,
    '',
    '',
    '',
    '',
    0,
    '',
    '',
    NOW(),
    NOW()
WHERE @export_flow_id IS NOT NULL 
  AND NOT EXISTS (SELECT 1 FROM flow_nodes WHERE flow_def_id = @export_flow_id AND node_code = 'end');

-- 8. 验证修复结果
SELECT 
    fd.type,
    fd.status,
    COUNT(fn.id) as node_count
FROM flow_definitions fd
LEFT JOIN flow_nodes fn ON fd.id = fn.flow_def_id
WHERE fd.type IN ('order_ddl', 'order_dml', 'order_export')
GROUP BY fd.type, fd.status
ORDER BY fd.type;
