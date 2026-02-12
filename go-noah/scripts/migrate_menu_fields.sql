-- 为 menu 表添加 Soybean-admin 格式的新字段
-- 执行此 SQL 脚本以更新数据库表结构

ALTER TABLE `menu` 
ADD COLUMN `menu_type` varchar(10) DEFAULT '2' COMMENT '菜单类型:1-目录,2-菜单' AFTER `weight`,
ADD COLUMN `menu_name` varchar(100) DEFAULT '' COMMENT '菜单名称' AFTER `menu_type`,
ADD COLUMN `route_name` varchar(100) DEFAULT '' COMMENT '路由名称' AFTER `menu_name`,
ADD COLUMN `route_path` varchar(255) DEFAULT '' COMMENT '路由路径' AFTER `route_name`,
ADD COLUMN `i18n_key` varchar(100) DEFAULT '' COMMENT '国际化key' AFTER `route_path`,
ADD COLUMN `icon_type` varchar(10) DEFAULT '1' COMMENT '图标类型:1-iconify,2-local' AFTER `i18n_key`,
ADD COLUMN `order` int DEFAULT 0 COMMENT '排序' AFTER `icon_type`,
ADD COLUMN `status` varchar(10) DEFAULT '1' COMMENT '状态:1-启用,2-禁用' AFTER `order`,
ADD COLUMN `multi_tab` tinyint(1) DEFAULT 0 COMMENT '是否多标签' AFTER `status`,
ADD COLUMN `active_menu` varchar(100) DEFAULT '' COMMENT '激活菜单' AFTER `multi_tab`,
ADD COLUMN `constant` tinyint(1) DEFAULT 0 COMMENT '是否常量' AFTER `active_menu`,
ADD COLUMN `href` varchar(255) DEFAULT '' COMMENT '外部链接' AFTER `constant`;

-- 更新现有数据：将 title 映射到 menu_name，path 映射到 route_path，name 映射到 route_name
-- 先更新基本字段
UPDATE `menu` SET 
  `menu_name` = COALESCE(NULLIF(`title`, ''), ''),
  `route_name` = COALESCE(NULLIF(`name`, ''), ''),
  `route_path` = COALESCE(NULLIF(`path`, ''), ''),
  `i18n_key` = COALESCE(NULLIF(`locale`, ''), ''),
  `order` = `weight`;

-- 更新 menu_type：如果有子菜单则为目录(1)，否则为菜单(2)
UPDATE `menu` m1
SET `menu_type` = CASE 
  WHEN EXISTS (SELECT 1 FROM `menu` m2 WHERE m2.parent_id = m1.id) THEN '1'
  ELSE '2'
END;

