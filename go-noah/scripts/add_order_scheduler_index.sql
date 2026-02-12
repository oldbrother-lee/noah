-- =====================================================
-- 定时工单调度器性能优化索引
-- 用于优化定时工单扫描查询性能
-- =====================================================

-- 1. 添加 scheduler_registered 字段（如果不存在）
-- 注意：GORM AutoMigrate 会自动添加字段，但如果表已存在且字段不存在，需要手动添加
ALTER TABLE `order_records` 
ADD COLUMN IF NOT EXISTS `scheduler_registered` tinyint(1) NOT NULL DEFAULT 0 COMMENT '定时任务是否已注册到调度器' AFTER `ghost_ok_to_drop_table`;

-- 2. 创建复合索引，优化定时工单扫描查询
-- 索引顺序：progress -> scheduler_registered -> schedule_time
-- 这样可以高效过滤：已批准 + 未注册 + 有定时时间的工单
CREATE INDEX IF NOT EXISTS `idx_order_scheduler_scan` 
ON `order_records` (`progress`, `scheduler_registered`, `schedule_time`);

-- 3. 验证索引是否创建成功
-- SELECT * FROM information_schema.STATISTICS 
-- WHERE TABLE_SCHEMA = DATABASE() 
-- AND TABLE_NAME = 'order_records' 
-- AND INDEX_NAME = 'idx_order_scheduler_scan';
