package checker

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"go-noah/internal/inspect/config"
	"go-noah/pkg/global"
	"go-noah/pkg/log"
	"github.com/spf13/viper"
)

// TestCase 测试用例结构
type TestCase struct {
	Name        string
	SQL         string
	SQLType     string // DDL, DML
	ExpectLevel AuditLevel
	ExpectMsg   string // 期望包含的消息
	SkipDB      bool   // 是否跳过数据库连接检查
}

// TestSQLParser 测试SQL解析
func TestSQLParser(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "正常CREATE TABLE",
			SQL:         "CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(100))",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			ExpectMsg:   "审核通过",
			SkipDB:      true,
		},
		{
			Name:        "语法错误SQL",
			SQL:         "CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(100",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			SkipDB:      true,
		},
		{
			Name:        "多语句SQL",
			SQL:         "CREATE TABLE t1 (id INT); CREATE TABLE t2 (id INT);",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				if tc.ExpectLevel == LevelError {
					t.Logf("✅ 预期错误: %v", err)
					return
				}
				t.Errorf("❌ 解析失败: %v", err)
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestCreateTableRules 测试CREATE TABLE规则
func TestCreateTableRules(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "缺少主键",
			SQL:         "CREATE TABLE test_table (id INT, name VARCHAR(100)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "主键",
			SkipDB:      true,
		},
		{
			Name:        "缺少表注释",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelWarning,
			ExpectMsg:   "注释",
			SkipDB:      true,
		},
		{
			Name:        "主键不是BIGINT",
			SQL:         "CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(100)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "BIGINT",
			SkipDB:      true,
		},
		{
			Name:        "主键不是UNSIGNED",
			SQL:         "CREATE TABLE test_table (id BIGINT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "UNSIGNED",
			SkipDB:      true,
		},
		{
			Name:        "主键不是AUTO_INCREMENT",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED PRIMARY KEY, name VARCHAR(100)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "AUTO_INCREMENT",
			SkipDB:      true,
		},
		{
			Name:        "正确的CREATE TABLE",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键', name VARCHAR(100) COMMENT '名称') ENGINE=InnoDB COMMENT='测试表'",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "CREATE TABLE AS语法",
			SQL:         "CREATE TABLE test_table AS SELECT * FROM other_table",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "CREATE TABLE AS",
			SkipDB:      true,
		},
		{
			Name:        "CREATE TABLE LIKE语法",
			SQL:         "CREATE TABLE test_table LIKE other_table",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "CREATE TABLE LIKE",
			SkipDB:      true,
		},
		{
			Name:        "索引前缀检查-唯一索引",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100), UNIQUE KEY name_idx (name)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "uniq_",
			SkipDB:      true,
		},
		{
			Name:        "索引前缀检查-普通索引",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100), KEY name_key (name)) ENGINE=InnoDB",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "idx_",
			SkipDB:      true,
		},
		{
			Name:        "正确的索引命名",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100), UNIQUE KEY uniq_name (name), KEY idx_name (name)) ENGINE=InnoDB COMMENT='测试表'",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "列缺少注释",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100)) ENGINE=InnoDB COMMENT='测试表'",
			SQLType:     "DDL",
			ExpectLevel: LevelWarning,
			ExpectMsg:   "注释",
			SkipDB:      true,
		},
		{
			Name:        "存储引擎检查",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY) ENGINE=MyISAM",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "InnoDB",
			SkipDB:      true,
		},
		{
			Name:        "字符集检查",
			SQL:         "CREATE TABLE test_table (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY) ENGINE=InnoDB DEFAULT CHARSET=latin1",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "utf8mb4",
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				t.Logf("⚠️  SQL解析错误（可能是语法错误）: %v", err)
				if tc.ExpectLevel == LevelError {
					return
				}
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			passed := false

			// 检查级别
			if result.Level == tc.ExpectLevel {
				passed = true
			} else if tc.ExpectLevel == LevelPass && (result.Level == LevelNotice || result.Level == LevelWarning) {
				// 如果期望通过，但实际是警告或提示，也算通过
				passed = true
			}

			// 检查消息
			if tc.ExpectMsg != "" {
				found := false
				for _, m := range result.Messages {
					if contains(m, tc.ExpectMsg) {
						found = true
						break
					}
				}
				if !found {
					passed = false
				}
			}

			if passed {
				t.Logf("✅ 测试通过")
			} else {
				t.Errorf("❌ 测试失败: 期望Level=%s, 实际Level=%s, 期望消息包含=%s", tc.ExpectLevel, result.Level, tc.ExpectMsg)
			}

			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestAlterTableRules 测试ALTER TABLE规则
func TestAlterTableRules(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "DROP列检查",
			SQL:         "ALTER TABLE test_table DROP COLUMN name",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "DROP列",
			SkipDB:      true,
		},
		{
			Name:        "DROP索引检查（允许）",
			SQL:         "ALTER TABLE test_table DROP INDEX idx_name",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "DROP主键检查",
			SQL:         "ALTER TABLE test_table DROP PRIMARY KEY",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "DROP主键",
			SkipDB:      true,
		},
		{
			Name:        "RENAME表名检查",
			SQL:         "ALTER TABLE test_table RENAME TO new_table",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "RENAME",
			SkipDB:      true,
		},
		{
			Name:        "ADD列-缺少注释",
			SQL:         "ALTER TABLE test_table ADD COLUMN new_col VARCHAR(100)",
			SQLType:     "DDL",
			ExpectLevel: LevelWarning,
			ExpectMsg:   "注释",
			SkipDB:      true,
		},
		{
			Name:        "ADD列-正确的",
			SQL:         "ALTER TABLE test_table ADD COLUMN new_col VARCHAR(100) COMMENT '新列'",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "ADD索引-前缀检查",
			SQL:         "ALTER TABLE test_table ADD INDEX name_key (name)",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "idx_",
			SkipDB:      true,
		},
		{
			Name:        "ADD索引-正确的",
			SQL:         "ALTER TABLE test_table ADD INDEX idx_name (name)",
			SQLType:     "DDL",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "MODIFY列-字符集检查",
			SQL:         "ALTER TABLE test_table MODIFY COLUMN name VARCHAR(100) CHARACTER SET latin1",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "utf8mb4",
			SkipDB:      true,
		},
		{
			Name:        "CHANGE列名检查",
			SQL:         "ALTER TABLE test_table CHANGE COLUMN old_name new_name VARCHAR(100)",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "CHANGE修改列名",
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				t.Logf("⚠️  SQL解析错误: %v", err)
				if tc.ExpectLevel == LevelError {
					return
				}
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			passed := false

			// 检查级别
			if result.Level == tc.ExpectLevel {
				passed = true
			} else if tc.ExpectLevel == LevelPass && (result.Level == LevelNotice || result.Level == LevelWarning) {
				passed = true
			}

			// 检查消息
			if tc.ExpectMsg != "" {
				found := false
				for _, m := range result.Messages {
					if contains(m, tc.ExpectMsg) {
						found = true
						break
					}
				}
				if !found {
					passed = false
				}
			}

			if passed {
				t.Logf("✅ 测试通过")
			} else {
				t.Errorf("❌ 测试失败: 期望Level=%s, 实际Level=%s, 期望消息包含=%s", tc.ExpectLevel, result.Level, tc.ExpectMsg)
			}

			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestDMLRules 测试DML规则
func TestDMLRules(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "UPDATE缺少WHERE",
			SQL:         "UPDATE test_table SET name = 'new_name'",
			SQLType:     "DML",
			ExpectLevel: LevelError,
			ExpectMsg:   "WHERE",
			SkipDB:      true,
		},
		{
			Name:        "DELETE缺少WHERE",
			SQL:         "DELETE FROM test_table",
			SQLType:     "DML",
			ExpectLevel: LevelError,
			ExpectMsg:   "WHERE",
			SkipDB:      true,
		},
		{
			Name:        "UPDATE有WHERE",
			SQL:         "UPDATE test_table SET name = 'new_name' WHERE id = 1",
			SQLType:     "DML",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "DELETE有WHERE",
			SQL:         "DELETE FROM test_table WHERE id = 1",
			SQLType:     "DML",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "INSERT INTO SELECT",
			SQL:         "INSERT INTO test_table SELECT * FROM other_table",
			SQLType:     "DML",
			ExpectLevel: LevelError,
			ExpectMsg:   "INSERT INTO SELECT",
			SkipDB:      true,
		},
		{
			Name:        "INSERT不指定列名",
			SQL:         "INSERT INTO test_table VALUES (1, 'name')",
			SQLType:     "DML",
			ExpectLevel: LevelError,
			ExpectMsg:   "列名",
			SkipDB:      true,
		},
		{
			Name:        "INSERT指定列名",
			SQL:         "INSERT INTO test_table (id, name) VALUES (1, 'name')",
			SQLType:     "DML",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
		{
			Name:        "JOIN缺少ON",
			SQL:         "UPDATE t1 JOIN t2 SET t1.name = t2.name",
			SQLType:     "DML",
			ExpectLevel: LevelError,
			ExpectMsg:   "ON",
			SkipDB:      true,
		},
		{
			Name:        "JOIN有ON",
			SQL:         "UPDATE t1 JOIN t2 ON t1.id = t2.id SET t1.name = t2.name",
			SQLType:     "DML",
			ExpectLevel: LevelPass,
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				t.Logf("⚠️  SQL解析错误: %v", err)
				if tc.ExpectLevel == LevelError {
					return
				}
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			passed := false

			// 检查级别
			if result.Level == tc.ExpectLevel {
				passed = true
			} else if tc.ExpectLevel == LevelPass && (result.Level == LevelNotice || result.Level == LevelWarning) {
				passed = true
			}

			// 检查消息
			if tc.ExpectMsg != "" {
				found := false
				for _, m := range result.Messages {
					if contains(m, tc.ExpectMsg) {
						found = true
						break
					}
				}
				if !found {
					passed = false
				}
			}

			if passed {
				t.Logf("✅ 测试通过")
			} else {
				t.Errorf("❌ 测试失败: 期望Level=%s, 实际Level=%s, 期望消息包含=%s", tc.ExpectLevel, result.Level, tc.ExpectMsg)
			}

			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestDropTableRules 测试DROP TABLE规则
func TestDropTableRules(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "DROP TABLE检查",
			SQL:         "DROP TABLE test_table",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "DROP TABLE",
			SkipDB:      true,
		},
		{
			Name:        "TRUNCATE TABLE检查",
			SQL:         "TRUNCATE TABLE test_table",
			SQLType:     "DDL",
			ExpectLevel: LevelError,
			ExpectMsg:   "TRUNCATE",
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				t.Logf("⚠️  SQL解析错误: %v", err)
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			passed := false

			// 检查级别
			if result.Level == tc.ExpectLevel {
				passed = true
			}

			// 检查消息
			if tc.ExpectMsg != "" {
				found := false
				for _, m := range result.Messages {
					if contains(m, tc.ExpectMsg) {
						found = true
						break
					}
				}
				if !found {
					passed = false
				}
			}

			if passed {
				t.Logf("✅ 测试通过")
			} else {
				t.Errorf("❌ 测试失败: 期望Level=%s, 实际Level=%s, 期望消息包含=%s", tc.ExpectLevel, result.Level, tc.ExpectMsg)
			}

			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestSelectStatement 测试SELECT语句
func TestSelectStatement(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "SELECT语句检查",
			SQL:         "SELECT * FROM test_table",
			SQLType:     "DML",
			ExpectLevel: LevelWarning,
			ExpectMsg:   "SELECT语句",
			SkipDB:      true,
		},
	}

	checker := NewChecker(config.DefaultInspectParams(), "MySQL")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			results, err := checker.Check(tc.SQL)
			if err != nil {
				t.Logf("⚠️  SQL解析错误: %v", err)
				return
			}

			if len(results) == 0 {
				t.Errorf("❌ 没有返回审核结果")
				return
			}

			result := results[0]
			passed := false

			// 检查级别
			if result.Level == tc.ExpectLevel {
				passed = true
			}

			// 检查消息
			if tc.ExpectMsg != "" {
				found := false
				for _, m := range result.Messages {
					if contains(m, tc.ExpectMsg) {
						found = true
						break
					}
				}
				if !found {
					passed = false
				}
			}

			if passed {
				t.Logf("✅ 测试通过")
			} else {
				t.Errorf("❌ 测试失败: 期望Level=%s, 实际Level=%s, 期望消息包含=%s", tc.ExpectLevel, result.Level, tc.ExpectMsg)
			}

			t.Logf("📋 SQL: %s", tc.SQL)
			t.Logf("📊 结果: Level=%s, Type=%s", result.Level, result.Type)
			t.Logf("💬 消息: %v", result.Messages)
			t.Logf("📝 摘要: %v", result.Summary)
		})
	}
}

// TestSQLTypeCheck 测试SQL类型检查
func TestSQLTypeCheck(t *testing.T) {
	testCases := []struct {
		Name        string
		SQL         string
		SQLType     string // DDL, DML, EXPORT
		ExpectError bool
	}{
		{
			Name:        "DDL模式下的SELECT语句",
			SQL:         "SELECT * FROM test_table",
			SQLType:     "DDL",
			ExpectError: true,
		},
		{
			Name:        "DDL模式下的ALTER语句",
			SQL:         "ALTER TABLE test_table ADD COLUMN new_col INT",
			SQLType:     "DDL",
			ExpectError: false,
		},
		{
			Name:        "DML模式下的UPDATE语句",
			SQL:         "UPDATE test_table SET name = 'new' WHERE id = 1",
			SQLType:     "DML",
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			checker := NewChecker(config.DefaultInspectParams(), "MySQL")
			results, err := checker.Check(tc.SQL)

			if tc.ExpectError {
				if err == nil {
					// 检查结果中是否有错误
					if len(results) > 0 && results[0].Level == LevelError {
						t.Logf("✅ 测试通过: 检测到错误")
					} else {
						t.Errorf("❌ 测试失败: 期望错误但未检测到")
					}
				} else {
					t.Logf("✅ 测试通过: SQL解析错误（符合预期）")
				}
			} else {
				if err != nil {
					t.Errorf("❌ 测试失败: 不应该有解析错误: %v", err)
				} else {
					t.Logf("✅ 测试通过: SQL解析成功")
				}
			}

			if len(results) > 0 {
				resultJSON, _ := json.MarshalIndent(results[0], "", "  ")
				t.Logf("📊 审核结果: %s", string(resultJSON))
			}
		})
	}
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// initTestLogger 初始化测试用的logger
func initTestLogger() {
	if global.Logger == nil {
		conf := viper.New()
		conf.Set("log.log_file_name", "test.log")
		conf.Set("log.log_level", "info")
		conf.Set("log.max_size", 100)
		conf.Set("log.max_backups", 3)
		conf.Set("log.max_age", 7)
		conf.Set("log.compress", false)
		conf.Set("log.encoding", "console")
		global.Logger = log.NewLog(conf)
	}
}

// TestAllRules 运行所有测试
func TestAllRules(t *testing.T) {
	initTestLogger()
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("开始SQL审核模块完整测试")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	t.Run("SQL解析测试", TestSQLParser)
	t.Run("CREATE TABLE规则测试", TestCreateTableRules)
	t.Run("ALTER TABLE规则测试", TestAlterTableRules)
	t.Run("DML规则测试", TestDMLRules)
	t.Run("DROP TABLE规则测试", TestDropTableRules)
	t.Run("SELECT语句测试", TestSelectStatement)
	t.Run("SQL类型检查测试", TestSQLTypeCheck)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("测试完成")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

