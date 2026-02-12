-- 检查流程定义是否存在
-- 执行方法：在数据库中执行以下SQL语句

-- 1. 查看所有流程定义
SELECT 
    id,
    code,
    name,
    type,
    description,
    version,
    status,
    created_at,
    updated_at
FROM flow_definitions
ORDER BY type, id;

-- 2. 检查是否存在启用的流程定义（status = 1）
SELECT 
    type,
    COUNT(*) as count,
    GROUP_CONCAT(id) as ids
FROM flow_definitions
WHERE status = 1
GROUP BY type;

-- 3. 检查每个流程定义是否有节点
SELECT 
    fd.id,
    fd.code,
    fd.name,
    fd.type,
    fd.status,
    COUNT(fn.id) as node_count
FROM flow_definitions fd
LEFT JOIN flow_nodes fn ON fd.id = fn.flow_def_id
GROUP BY fd.id, fd.code, fd.name, fd.type, fd.status
ORDER BY fd.type, fd.id;

-- 4. 检查 order_ddl 流程定义（应该存在且 status = 1）
SELECT 
    fd.*,
    COUNT(fn.id) as node_count
FROM flow_definitions fd
LEFT JOIN flow_nodes fn ON fd.id = fn.flow_def_id
WHERE fd.type = 'order_ddl'
GROUP BY fd.id;

-- 5. 如果不存在或未启用，可以手动创建（参考下面的SQL）
