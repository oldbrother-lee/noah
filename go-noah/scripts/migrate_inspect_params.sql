-- =====================================================
-- 审核参数配置迁移脚本
-- 用于从 goInsight 数据库迁移 inspect_params 数据到 go-noah
-- =====================================================

-- 注意事项:
-- 1. 执行前请先备份数据库
-- 2. 老系统表名: insight_inspect_params（带 insight_ 前缀）
-- 3. 新系统表名: inspect_params（不带前缀）
-- 4. 如果新系统表为空，会自动初始化默认参数（通过代码）
-- 5. 如果老系统有自定义参数，需要手动迁移

-- =====================================================
-- 方案1: 从老数据库直接迁移（推荐用于生产环境）
-- =====================================================

-- 前提条件:
-- 1. 老数据库和新数据库在同一台服务器，或者可以跨库访问
-- 2. 如果不在同一服务器，需要先导出数据再导入

-- 步骤1: 检查老数据库数据
-- SELECT COUNT(*) FROM insight_inspect_params;  -- 应该 >= 80

-- 步骤2: 检查新数据库数据
-- SELECT COUNT(*) FROM inspect_params;  -- 如果为 0，需要迁移

-- 步骤3: 执行迁移（使用 INSERT IGNORE 避免重复）
INSERT IGNORE INTO inspect_params (params, remark, created_at, updated_at)
SELECT params, remark, created_at, updated_at
FROM insight_inspect_params
WHERE remark IS NOT NULL AND remark != '';

-- 步骤4: 验证迁移结果
SELECT COUNT(*) as total_count FROM inspect_params;
SELECT remark FROM inspect_params ORDER BY id LIMIT 10;

-- =====================================================
-- 方案2: 如果老数据库和新数据库不在同一服务器
-- =====================================================

-- 步骤1: 从老数据库导出数据
-- mysqldump -u username -p database_name insight_inspect_params > inspect_params.sql
-- 或者使用 SELECT 导出为 CSV/JSON

-- 步骤2: 在新数据库执行 INSERT 语句
-- 根据导出的数据格式调整 INSERT 语句

-- =====================================================
-- 方案3: 如果老数据库表名不同或结构不同
-- =====================================================

-- 如果老数据库表名不是 insight_inspect_params，需要调整 FROM 子句
-- 例如：FROM goinsight_db.insight_inspect_params

-- =====================================================
-- 方案4: 只迁移自定义参数（保留新系统的默认参数）
-- =====================================================

-- 如果新系统已经初始化了默认参数，只想迁移老系统的自定义参数：
-- 1. 先查看老系统有哪些参数是新系统没有的
-- SELECT remark FROM insight_inspect_params 
-- WHERE remark NOT IN (SELECT remark FROM inspect_params);

-- 2. 只迁移这些差异参数
-- INSERT IGNORE INTO inspect_params (params, remark, created_at, updated_at)
-- SELECT params, remark, created_at, updated_at
-- FROM insight_inspect_params
-- WHERE remark NOT IN (SELECT remark FROM inspect_params);

-- =====================================================
-- 方案5: 强制覆盖（谨慎使用）
-- =====================================================

-- 如果需要用老系统的数据完全覆盖新系统（会丢失新系统的默认值）:
-- DELETE FROM inspect_params;
-- INSERT INTO inspect_params (params, remark, created_at, updated_at)
-- SELECT params, remark, created_at, updated_at
-- FROM insight_inspect_params;

-- =====================================================
-- 验证和检查
-- =====================================================

-- 1. 检查数据总数（应该 >= 80）
SELECT COUNT(*) as total_count FROM inspect_params;

-- 2. 检查是否有重复的 remark（不应该有，因为有 uniqueIndex）
SELECT remark, COUNT(*) as count 
FROM inspect_params 
GROUP BY remark 
HAVING count > 1;

-- 3. 检查关键参数是否存在
SELECT remark FROM inspect_params 
WHERE remark IN (
    '表名的长度',
    '检查表是否有注释',
    '是否检查表的字符集和排序规则',
    'DML语句必须有where条件',
    '最大影响行数，默认100'
);

-- 4. 检查参数格式是否正确（JSON 格式）
SELECT id, remark, 
       JSON_VALID(params) as is_valid_json,
       JSON_TYPE(params) as json_type
FROM inspect_params
WHERE JSON_VALID(params) = 0;  -- 如果有结果，说明有无效的 JSON

-- =====================================================
-- 故障排查
-- =====================================================

-- 问题1: 迁移后数据为空
-- 解决: 检查表名是否正确，检查 WHERE 条件是否过滤掉了所有数据

-- 问题2: 迁移后数据重复
-- 解决: 使用 INSERT IGNORE 或先删除再插入

-- 问题3: JSON 格式错误
-- 解决: 检查老数据库的 params 字段格式，可能需要转换

-- 问题4: 迁移后参数不生效
-- 解决: 
-- 1. 检查服务是否重启（新系统启动时会自动初始化，如果表为空）
-- 2. 检查代码中的 InitializeInspectParamsIfNeeded 是否被调用
-- 3. 检查日志是否有初始化错误

-- =====================================================
-- 回滚方案（如果需要）
-- =====================================================

-- 如果迁移失败，可以删除迁移的数据：
-- DELETE FROM inspect_params WHERE created_at > '迁移时间';
-- 然后重启服务，让系统自动初始化默认参数
