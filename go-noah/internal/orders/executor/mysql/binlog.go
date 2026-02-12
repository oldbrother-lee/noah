package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"go-noah/pkg/global"
	"go-noah/pkg/utils"
	"strings"
	"time"

	mysqlpkg "github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"go.uber.org/zap"
)

// BinlogConfig binlog配置
type BinlogConfig struct {
	Hostname string
	Port     int
	UserName string
	Password string
	Schema   string
}

// Binlog binlog解析器
type Binlog struct {
	Config        *BinlogConfig
	ConnectionID  int64
	StartFile     string
	StartPosition int64
	EndFile       string
	EndPosition   int64
}

// parserTableStmt 解析表结构
func (b *Binlog) parserTableStmt(table string) (*ast.CreateTableStmt, error) {
	// 连接数据库
	db, err := b.connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 查询表结构
	var tableName, createTableSQL string
	err = db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", table)).Scan(&tableName, &createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	// 解析表结构
	stmt, err := parser.New().ParseOneStmt(createTableSQL, "", "")
	if err != nil {
		return nil, fmt.Errorf("解析表结构失败: %w", err)
	}

	switch s := stmt.(type) {
	case *ast.CreateTableStmt:
		return s, nil
	default:
		return nil, fmt.Errorf("不是CREATE TABLE语句")
	}
}

// extractPK 提取主键
func (b *Binlog) extractPK(stmt *ast.CreateTableStmt) (bool, []string) {
	var keys []string
	// 从列定义中提取主键
	for _, col := range stmt.Cols {
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionPrimaryKey {
				keys = append(keys, col.Name.Name.O)
			}
		}
	}
	// 从约束中提取主键
	for _, cons := range stmt.Constraints {
		if cons.Tp == ast.ConstraintPrimaryKey {
			for _, col := range cons.Keys {
				if !utils.IsContain(keys, col.Column.Name.O) {
					keys = append(keys, col.Column.Name.O)
				}
			}
		}
	}
	return len(keys) > 0, keys
}

// getNonGeneratedCols 获取非计算列的索引映射
// 返回: binlog列索引 -> stmt.Cols索引的映射
func getNonGeneratedCols(stmt *ast.CreateTableStmt) []int {
	var colIndices []int
	for i, col := range stmt.Cols {
		if !isGenerated(col.Options) {
			colIndices = append(colIndices, i)
		}
	}
	return colIndices
}

// Run 解析binlog生成回滚SQL
func (b *Binlog) Run() (string, error) {
	cfg := replication.BinlogSyncerConfig{
		ServerID:   20231108 + uint32(uint32(time.Now().Unix())%10000),
		Flavor:     "mysql",
		Host:       b.Config.Hostname,
		Port:       uint16(b.Config.Port),
		User:       b.Config.UserName,
		Password:   b.Config.Password,
		UseDecimal: true,
	}
	syncer := replication.NewBinlogSyncer(cfg)
	defer syncer.Close()

	// 定义开始结束的pos
	startPosition := mysqlpkg.Position{Name: b.StartFile, Pos: uint32(b.StartPosition)}
	stopPosition := mysqlpkg.Position{Name: b.EndFile, Pos: uint32(b.EndPosition)}

	// 开启同步
	streamer, err := syncer.StartSync(startPosition)
	if err != nil {
		return "", fmt.Errorf("启动binlog同步失败: %w", err)
	}

	// 声明当前的pos
	currentPosition := startPosition
	// 获取当前事件的thread id，用来和执行SQL的thread id进行比较
	var currentThreadID uint32
	// 回滚SQL
	var rbsqls []string

	// 循环解析binlog事件
	for {
		e, err := streamer.GetEvent(context.Background())
		if err != nil {
			return "", fmt.Errorf("获取binlog事件失败: %w", err)
		}

		if e.Header.LogPos > 0 {
			currentPosition.Pos = e.Header.LogPos
		}

		if e.Header.EventType == replication.ROTATE_EVENT {
			if event, ok := e.Event.(*replication.RotateEvent); ok {
				currentPosition = mysqlpkg.Position{
					Name: string(event.NextLogName),
					Pos:  uint32(event.Position),
				}
			}
		}

		if currentPosition.Compare(startPosition) == -1 {
			continue
		}
		// 如果当前pos大于停止的pos，退出
		if currentPosition.Compare(stopPosition) > -1 {
			break
		}

		// 事件类型判断
		switch e.Header.EventType {
		case replication.QUERY_EVENT:
			if event, ok := e.Event.(*replication.QueryEvent); ok {
				currentThreadID = event.SlaveProxyID
			}
		case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
			if event, ok := e.Event.(*replication.RowsEvent); ok {
				// 获取表的stmt
				tableName := fmt.Sprintf("`%s`.`%s`", event.Table.Schema, event.Table.Table)
				stmt, err := b.parserTableStmt(tableName)
				if err != nil {
					return "", err
				}
				// 解析回滚语句（INSERT -> DELETE）
				if b.ConnectionID == int64(currentThreadID) {
					sql, err := b.generateDeleteSql(event, stmt)
					if err != nil {
						return "", err
					}
					if sql != "" {
						rbsqls = append(rbsqls, sql)
					}
				}
			}
		case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
			if event, ok := e.Event.(*replication.RowsEvent); ok {
				// 获取表的stmt
				tableName := fmt.Sprintf("`%s`.`%s`", event.Table.Schema, event.Table.Table)
				stmt, err := b.parserTableStmt(tableName)
				if err != nil {
					return "", err
				}
				// 解析回滚语句（DELETE -> INSERT）
				if b.ConnectionID == int64(currentThreadID) {
					sql, err := b.generateInsertSql(event, stmt)
					if err != nil {
						return "", err
					}
					if sql != "" {
						rbsqls = append(rbsqls, sql)
					}
				}
			}
		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
			if event, ok := e.Event.(*replication.RowsEvent); ok {
				// 获取表的stmt
				tableName := fmt.Sprintf("`%s`.`%s`", event.Table.Schema, event.Table.Table)
				stmt, err := b.parserTableStmt(tableName)
				if err != nil {
					return "", err
				}
				// 解析回滚语句（UPDATE -> 反向UPDATE）
				if b.ConnectionID == int64(currentThreadID) {
					sql, err := b.generateUpdateSql(event, stmt)
					if err != nil {
						return "", err
					}
					if sql != "" {
						rbsqls = append(rbsqls, sql)
					}
				}
			}
		}
	}

	return strings.Join(rbsqls, ";\r\n"), nil
}

// generateUpdateSql 生成UPDATE回滚SQL
func (b *Binlog) generateUpdateSql(e *replication.RowsEvent, stmt *ast.CreateTableStmt) (string, error) {
	template := "UPDATE `%s`.`%s` SET %s WHERE"
	hasPrimaryKey, PrimaryKeys := b.extractPK(stmt)
	colIndices := getNonGeneratedCols(stmt)

	// 验证和日志：检查 binlog rows 长度与表结构列数
	if len(e.Rows) > 0 {
		firstRowLen := len(e.Rows[0])
		expectedColCount := len(colIndices)

		global.Logger.Info("Binlog列数验证(UPDATE)",
			zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
			zap.Int("binlog_rows_length", firstRowLen),
			zap.Int("table_non_generated_cols_count", expectedColCount),
			zap.Int("table_total_cols_count", len(stmt.Cols)),
			zap.Bool("has_primary_key", hasPrimaryKey),
			zap.Int("rows_count", len(e.Rows)),
		)

		if firstRowLen != expectedColCount {
			global.Logger.Warn("Binlog rows长度与表结构列数不匹配，将只使用前N个列(UPDATE)",
				zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
				zap.Int("binlog_rows_length", firstRowLen),
				zap.Int("expected_cols_count", expectedColCount),
			)
		}
	}

	var (
		oldValues []driver.Value
		newValues []driver.Value
		newSql    string
	)

	var rbsqls []string
	var sets []string
	var sql string

	// e.Rows:  [[old values] [new values] [old values] [new values] ...]
	for i, rows := range e.Rows {
		var columns []string
		if i%2 == 0 {
			// old values - 生成SET部分
			for binlogIdx, d := range rows {
				if binlogIdx >= len(colIndices) {
					break
				}
				stmtIdx := colIndices[binlogIdx]
				if stmtIdx >= len(stmt.Cols) {
					break
				}
				col := stmt.Cols[stmtIdx]
				if isUnsigned(col.Tp.GetFlag()) {
					d = processValue(d, col.Tp.GetType())
				}
				sets = append(sets, fmt.Sprintf(" `%s`=?", col.Name.Name.O))
				newValues = append(newValues, d)
			}
			sql = fmt.Sprintf(template, e.Table.Schema, e.Table.Table, strings.Join(sets, ","))
		} else {
			// new values - 生成WHERE部分
			for binlogIdx, d := range rows {
				if binlogIdx >= len(colIndices) {
					break
				}
				stmtIdx := colIndices[binlogIdx]
				if stmtIdx >= len(stmt.Cols) {
					break
				}
				col := stmt.Cols[stmtIdx]
				if hasPrimaryKey {
					if utils.IsContain(PrimaryKeys, col.Name.Name.O) {
						if isUnsigned(col.Tp.GetFlag()) {
							d = processValue(d, col.Tp.GetType())
						}
						oldValues = append(oldValues, d)
						if d == nil {
							columns = append(columns, fmt.Sprintf(" `%s` IS ?", col.Name.Name.O))
						} else {
							columns = append(columns, fmt.Sprintf(" `%s`=?", col.Name.Name.O))
						}
					}
				} else {
					if isUnsigned(col.Tp.GetFlag()) {
						d = processValue(d, col.Tp.GetType())
					}
					oldValues = append(oldValues, d)
					if d == nil {
						columns = append(columns, fmt.Sprintf(" `%s` IS ?", col.Name.Name.O))
					} else {
						columns = append(columns, fmt.Sprintf(" `%s`=?", col.Name.Name.O))
					}
				}
				// 重置
				sets = []string{}
			}
			// 如果没有生成任何 WHERE 条件，跳过这条记录
			if len(columns) == 0 {
				oldValues = nil
				newValues = nil
				continue
			}
			// 确保 oldValues 和 columns 长度一致
			if len(oldValues) != len(columns) {
				return "", fmt.Errorf("无法生成UPDATE回滚SQL：WHERE条件参数数量(%d)与列数量(%d)不匹配", len(oldValues), len(columns))
			}
			newSql = strings.Join([]string{sql, strings.Join(columns, " AND")}, "")
			newValues = append(newValues, oldValues...)
			r, err := interpolateParams(newSql, newValues, true)
			if err != nil {
				return "", fmt.Errorf("生成UPDATE回滚SQL失败: %w", err)
			}
			rbsqls = append(rbsqls, string(r))
			oldValues = nil
			newValues = nil
		}
	}
	return strings.Join(rbsqls, ";\r\n"), nil
}

// generateInsertSql 生成INSERT回滚SQL（DELETE）
func (b *Binlog) generateInsertSql(e *replication.RowsEvent, stmt *ast.CreateTableStmt) (string, error) {
	var columns []string
	template := "INSERT INTO `%s`.`%s`(%s) VALUES(%s)"
	colIndices := getNonGeneratedCols(stmt)
	for _, stmtIdx := range colIndices {
		col := stmt.Cols[stmtIdx]
		columns = append(columns, fmt.Sprintf("`%s`", col.Name.Name.O))
	}
	paramValues := strings.TrimRight(strings.Repeat("?,", len(columns)), ",")
	sql := fmt.Sprintf(template, e.Table.Schema, e.Table.Table,
		strings.Join(columns, ","), paramValues)

	// 验证和日志：检查 binlog rows 长度与表结构列数
	if len(e.Rows) > 0 {
		firstRowLen := len(e.Rows[0])
		expectedColCount := len(colIndices)

		global.Logger.Info("Binlog列数验证(INSERT)",
			zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
			zap.Int("binlog_rows_length", firstRowLen),
			zap.Int("table_non_generated_cols_count", expectedColCount),
			zap.Int("table_total_cols_count", len(stmt.Cols)),
		)

		if firstRowLen != expectedColCount {
			global.Logger.Warn("Binlog rows长度与表结构列数不匹配，将只使用前N个列(INSERT)",
				zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
				zap.Int("binlog_rows_length", firstRowLen),
				zap.Int("expected_cols_count", expectedColCount),
			)
		}
	}

	var rbsqls []string
	for rowIdx, rows := range e.Rows {
		var vv []driver.Value
		for binlogIdx, d := range rows {
			if binlogIdx >= len(colIndices) {
				// 如果 binlog 列数多于表结构列数，记录并跳过多余的列
				if rowIdx == 0 && binlogIdx == len(colIndices) {
					global.Logger.Warn("Binlog包含额外列，已跳过(INSERT)",
						zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
						zap.Int("skipped_from_index", binlogIdx),
						zap.Int("total_binlog_cols", len(rows)),
						zap.Int("used_cols", len(colIndices)),
					)
				}
				break
			}
			stmtIdx := colIndices[binlogIdx]
			if stmtIdx >= len(stmt.Cols) {
				break
			}
			col := stmt.Cols[stmtIdx]
			if isUnsigned(col.Tp.GetFlag()) {
				d = processValue(d, col.Tp.GetType())
			}
			vv = append(vv, d)
		}

		// 验证 vv 长度与 columns 长度是否匹配
		if len(vv) != len(columns) {
			global.Logger.Error("INSERT回滚SQL生成失败：参数数量与列数量不匹配",
				zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
				zap.Int("vv_length", len(vv)),
				zap.Int("columns_length", len(columns)),
				zap.Int("binlog_rows_length", len(rows)),
				zap.Int("expected_cols_count", len(colIndices)),
			)
			return "", fmt.Errorf("生成INSERT回滚SQL失败：参数数量(%d)与列数量(%d)不匹配，binlog rows长度(%d)与表结构列数(%d)不匹配", len(vv), len(columns), len(rows), len(colIndices))
		}

		r, err := interpolateParams(sql, vv, false)
		if err != nil {
			return "", fmt.Errorf("生成回滚SQL失败: %w", err)
		}
		rbsqls = append(rbsqls, string(r))
	}
	return strings.Join(rbsqls, ";\r\n"), nil
}

// generateDeleteSql 生成DELETE回滚SQL（INSERT）
func (b *Binlog) generateDeleteSql(e *replication.RowsEvent, stmt *ast.CreateTableStmt) (string, error) {
	template := "DELETE FROM `%s`.`%s` WHERE "
	sql := fmt.Sprintf(template, e.Table.Schema, e.Table.Table)
	hasPrimaryKey, PrimaryKeys := b.extractPK(stmt)
	colIndices := getNonGeneratedCols(stmt)

	// 验证和日志：检查 binlog rows 长度与表结构列数
	if len(e.Rows) > 0 {
		firstRowLen := len(e.Rows[0])
		expectedColCount := len(colIndices)

		global.Logger.Info("Binlog列数验证",
			zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
			zap.Int("binlog_rows_length", firstRowLen),
			zap.Int("table_non_generated_cols_count", expectedColCount),
			zap.Int("table_total_cols_count", len(stmt.Cols)),
			zap.Bool("has_primary_key", hasPrimaryKey),
			zap.Int("rows_count", len(e.Rows)),
		)

		// 记录表结构列信息
		var colNames []string
		for _, stmtIdx := range colIndices {
			if stmtIdx < len(stmt.Cols) {
				colNames = append(colNames, stmt.Cols[stmtIdx].Name.Name.O)
			}
		}
		global.Logger.Info("表结构非计算列列表",
			zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
			zap.Strings("columns", colNames),
		)

		// 如果长度不匹配，记录警告但继续处理（只使用前 N 个列）
		if firstRowLen != expectedColCount {
			global.Logger.Warn("Binlog rows长度与表结构列数不匹配，将只使用前N个列",
				zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
				zap.Int("binlog_rows_length", firstRowLen),
				zap.Int("expected_cols_count", expectedColCount),
				zap.String("reason", "可能原因：binlog中包含隐藏列、虚拟列或表结构已变更"),
			)
		}
	}

	var rbsqls []string
	// e.Rows为insert into xx values(),(),();
	for rowIdx, rows := range e.Rows {
		var vv []driver.Value
		var columns []string
		// 判断是否有主键并提取主键
		if hasPrimaryKey {
			// 有主键：只使用主键列生成 WHERE 条件
			for binlogIdx, d := range rows {
				if binlogIdx >= len(colIndices) {
					break
				}
				stmtIdx := colIndices[binlogIdx]
				if stmtIdx >= len(stmt.Cols) {
					break
				}
				col := stmt.Cols[stmtIdx]
				if utils.IsContain(PrimaryKeys, col.Name.Name.O) {
					vv = append(vv, d)
					columns = append(columns, fmt.Sprintf("`%s`=?", col.Name.Name.O))
				}
			}
		} else {
			// 没有主键：使用所有非计算列生成 WHERE 条件
			for binlogIdx, d := range rows {
				if binlogIdx >= len(colIndices) {
					// 如果 binlog 列数多于表结构列数，记录并跳过多余的列
					if rowIdx == 0 && binlogIdx == len(colIndices) {
						global.Logger.Warn("Binlog包含额外列，已跳过",
							zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
							zap.Int("skipped_from_index", binlogIdx),
							zap.Int("total_binlog_cols", len(rows)),
							zap.Int("used_cols", len(colIndices)),
						)
					}
					break
				}
				stmtIdx := colIndices[binlogIdx]
				if stmtIdx >= len(stmt.Cols) {
					break
				}
				col := stmt.Cols[stmtIdx]
				vv = append(vv, d)
				if d == nil {
					columns = append(columns, fmt.Sprintf("`%s` IS ?", col.Name.Name.O))
				} else {
					columns = append(columns, fmt.Sprintf("`%s`=?", col.Name.Name.O))
				}
			}
		}

		// 记录生成的 WHERE 条件信息
		if rowIdx == 0 {
			global.Logger.Info("生成的WHERE条件",
				zap.String("table", fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Table)),
				zap.Int("where_conditions_count", len(columns)),
				zap.Strings("columns", columns),
			)
		}

		// 如果没有生成任何 WHERE 条件，说明表结构有问题或 binlog 数据不完整
		if len(columns) == 0 {
			return "", fmt.Errorf("无法生成回滚SQL：表 %s.%s 没有可用的列来生成WHERE条件（可能没有主键且binlog数据为空）", e.Table.Schema, e.Table.Table)
		}

		// 确保 vv 和 columns 长度一致
		if len(vv) != len(columns) {
			return "", fmt.Errorf("无法生成回滚SQL：参数数量(%d)与列数量(%d)不匹配", len(vv), len(columns))
		}

		newSql := strings.Join([]string{sql, strings.Join(columns, " AND ")}, "")
		r, err := interpolateParams(newSql, vv, false)
		if err != nil {
			return "", fmt.Errorf("生成回滚SQL失败: %w", err)
		}
		rbsqls = append(rbsqls, string(r))
	}
	return strings.Join(rbsqls, ";\r\n"), nil
}

// connect 连接数据库
func (b *Binlog) connect() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		b.Config.UserName,
		b.Config.Password,
		b.Config.Hostname,
		b.Config.Port,
		b.Config.Schema,
	)
	return sql.Open("mysql", dsn)
}
