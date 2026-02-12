-- 更新权限分配菜单配置
-- 将"用户权限"和"角色权限"合并为"权限分配"菜单

-- 1. 更新"权限管理"菜单的路由配置（如果需要）
-- 确保父菜单指向统一页面
UPDATE `menu` 
SET 
  `route_path` = '/system/database/permission',
  `route_name` = 'system_database_permission',
  `component` = 'layout.base'
WHERE `route_name` = 'system_database_permission' 
   OR `path` = '/system/database/permission';

-- 2. 删除"用户权限"子菜单（如果存在）
DELETE FROM `menu` 
WHERE `route_name` = 'system_database_permission_user'
   OR `path` = '/system/database/permission/user';

-- 3. 删除"角色权限"子菜单（如果存在）
DELETE FROM `menu` 
WHERE `route_name` = 'system_database_permission_role'
   OR `path` = '/system/database/permission/role';

-- 4. 确保"权限管理"菜单指向统一页面（如果不是目录类型）
-- 如果"权限管理"没有其他子菜单，将其改为直接指向权限分配页面
UPDATE `menu` m1
SET 
  `route_path` = '/system/database/permission',
  `route_name` = 'system_database_permission',
  `component` = 'view.system_database_permission',
  `menu_type` = '2'  -- 菜单类型
WHERE m1.`route_name` = 'system_database_permission'
  AND NOT EXISTS (
    SELECT 1 FROM `menu` m2 
    WHERE m2.`parent_id` = m1.`id` 
    AND m2.`route_name` NOT IN ('system_database_permission_user', 'system_database_permission_role')
  );

-- 注意：如果需要保留"权限管理"作为目录，并添加"权限分配"作为子菜单，可以使用以下SQL：
-- 但如果"权限模板"还在，那么"权限管理"应该保持为目录类型

-- 5. 检查"权限管理"是否还有其他子菜单（如"权限模板"）
-- 如果有其他子菜单，"权限管理"应该保持为目录类型
UPDATE `menu` m1
SET 
  `menu_type` = '1'  -- 目录类型
WHERE m1.`route_name` = 'system_database_permission'
  AND EXISTS (
    SELECT 1 FROM `menu` m2 
    WHERE m2.`parent_id` = m1.`id` 
    AND m2.`route_name` = 'system_database_permission_template'
  );
