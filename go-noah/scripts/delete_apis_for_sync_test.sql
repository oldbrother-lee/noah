-- 删除部分 API 记录，用于测试「同步路由」功能
-- 执行后：这些 path+method 会出现在「同步路由」弹窗的「新增路由」中
-- 使用前请确认：仅用于测试环境

-- 示例 1：按 path 前缀删除（会删除所有匹配的 path，不限 method）
-- DELETE FROM `api` WHERE path LIKE '/v1/admin/department%';

-- 示例 2：按指定 path + method 删除几条（取消注释并修改为你要测的接口）
-- DELETE FROM `api` WHERE (path, method) IN (
--   ('/v1/admin/apis', 'GET'),
--   ('/v1/admin/api', 'POST'),
--   ('/v1/admin/api/sync', 'GET')
-- );

-- 示例 3：按 ID 删除（先查 id：SELECT id, path, method FROM api LIMIT 20;）
-- DELETE FROM `api` WHERE id IN (1, 2, 3);

-- 示例 4：按 id 范围删除（先查：SELECT id, path, method FROM api ORDER BY id LIMIT 20;）
-- DELETE FROM `api` WHERE id BETWEEN 100 AND 110;
