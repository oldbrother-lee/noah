-- 更新权限管理相关菜单的路由名称
-- 将子菜单的完整路由名称改为只保留最后一部分，以便后端自动拼接

-- 更新权限组菜单
UPDATE `menu` 
SET `route_name` = 'group'
WHERE `route_name` = 'system_database_permission-group' 
   OR `route_name` = 'system_database_permission_group'
   OR `path` = '/system/database/permission-group';

-- 更新权限模板菜单
UPDATE `menu` 
SET `route_name` = 'template'
WHERE `route_name` = 'system_database_permission-template' 
   OR `route_name` = 'system_database_permission_template'
   OR `path` = '/system/database/permission-template';

-- 更新角色权限菜单
UPDATE `menu` 
SET `route_name` = 'role'
WHERE `route_name` = 'system_database_role-permission' 
   OR `route_name` = 'system_database_permission_role'
   OR `path` = '/system/database/role-permission';

-- 更新用户权限菜单
UPDATE `menu` 
SET `route_name` = 'user'
WHERE `route_name` = 'system_database_permission_user'
   OR `path` = '/system/database/permission/user';

-- 确保父菜单的路由名称正确
UPDATE `menu` 
SET `route_name` = 'system_database_permission'
WHERE `path` = '/system/database/permission' 
  AND `parent_id` != 0;

-- 验证更新结果
SELECT 
    id,
    parent_id,
    path,
    route_name,
    name,
    title
FROM `menu`
WHERE `path` LIKE '/system/database/permission%'
ORDER BY `parent_id`, `id`;

