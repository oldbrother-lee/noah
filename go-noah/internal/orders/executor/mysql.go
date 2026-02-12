package executor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-noah/internal/inspect/parser"
	mysqlpkg "go-noah/internal/orders/executor/mysql"
	"go-noah/pkg/global"
	"go-noah/pkg/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// MySQLExecutor MySQL执行器
type MySQLExecutor struct {
	Config *DBConfig
}

// NewMySQLExecutor 创建MySQL执行器
func NewMySQLExecutor(config *DBConfig) *MySQLExecutor {
	return &MySQLExecutor{Config: config}
}

// Run 执行SQL
func (e *MySQLExecutor) Run() (ReturnData, error) {
	switch e.Config.SQLType {
	case "DDL":
		return e.ExecuteDDL()
	case "DML":
		return e.ExecuteDML()
	case "EXPORT":
		return e.ExecuteExport()
	default:
		return ReturnData{Error: fmt.Sprintf("不支持的SQL类型: %s", e.Config.SQLType)}, fmt.Errorf("不支持的SQL类型: %s", e.Config.SQLType)
	}
}

// Connect 连接数据库
func (e *MySQLExecutor) Connect() (*sql.DB, error) {
	config := mysql.Config{
		User:                 e.Config.UserName,
		Passwd:               e.Config.Password,
		Addr:                 fmt.Sprintf("%s:%d", e.Config.Hostname, e.Config.Port),
		Net:                  "tcp",
		DBName:               e.Config.Schema,
		AllowNativePasswords: true,
		Timeout:              10 * time.Second,
		ReadTimeout:          300 * time.Second,
		WriteTimeout:         300 * time.Second,
	}

	if e.Config.Charset != "" {
		config.Params = map[string]string{"charset": e.Config.Charset}
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// ExecuteDDL 执行DDL语句
func (e *MySQLExecutor) ExecuteDDL() (ReturnData, error) {
	// 解析SQL类型，判断是否需要使用 gh-ost
	sqlType, err := parser.GetSqlStatement(e.Config.SQL)
	if err != nil {
		return ReturnData{Error: err.Error()}, err
	}

	switch sqlType {
	case "AlterTable":
		// ALTER TABLE 使用 gh-ost 在线执行
		return e.ExecuteDDLWithGhost()
	case "CreateDatabase", "CreateTable", "CreateView":
		// CREATE 语句直接执行
		return e.ExecuteOnlineDDL()
	case "DropTable", "DropIndex":
		// DROP 语句直接执行
		return e.ExecuteOnlineDDL()
	case "TruncateTable":
		// TRUNCATE 语句直接执行
		return e.ExecuteOnlineDDL()
	case "RenameTable":
		// RENAME TABLE 不支持，建议使用 ALTER TABLE ... RENAME
		return ReturnData{Error: "请更正为alter table ... rename语法"}, errors.New("请更正为alter table ... rename语法")
	case "CreateIndex":
		// CREATE INDEX 不支持，建议使用 ALTER TABLE ... ADD
		return ReturnData{Error: "请更正为alter table ... add语法"}, errors.New("请更正为alter table ... add语法")
	case "DropDatabase":
		// DROP DATABASE 禁止执行
		return ReturnData{Error: "【风险】禁止执行drop database操作"}, errors.New("【风险】禁止执行drop database操作")
	default:
		return ReturnData{Error: fmt.Sprintf("当前SQL未匹配到规则，执行失败，SQL类型为：%s", sqlType)}, fmt.Errorf("当前SQL未匹配到规则，执行失败，SQL类型为：%s", sqlType)
	}
}

// ExecuteOnlineDDL 执行Online DDL语句（直接执行）
func (e *MySQLExecutor) ExecuteOnlineDDL() (ReturnData, error) {
	var data ReturnData
	var executeLog []string

	logMessage := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)
		executeLog = append(executeLog, logMsg)
		// 发布消息到 Redis（用于 WebSocket 推送）
		if e.Config.OrderID != "" {
			_ = utils.PublishMessageToChannel(e.Config.OrderID, logMsg, "")
		}
	}

	// 连接数据库
	logMessage(fmt.Sprintf("连接数据库 %s:%d...", e.Config.Hostname, e.Config.Port))
	db, err := e.Connect()
	if err != nil {
		logMessage(fmt.Sprintf("连接失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	defer db.Close()
	logMessage("连接成功")

	// 获取连接ID
	var connectionID int64
	if err := db.QueryRow("SELECT CONNECTION_ID()").Scan(&connectionID); err != nil {
		logMessage(fmt.Sprintf("获取Connection ID失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	logMessage(fmt.Sprintf("Connection ID: %d", connectionID))

	// 启动 PROCESSLIST 监控（在单独的 goroutine 中）
	var ch1 chan int64
	if e.Config.OrderID != "" {
		ch1 = make(chan int64)
		go e.GetProcesslist(e.Config.OrderID, connectionID, ch1)
		// 确保在所有情况下都关闭 channel（函数返回时）
		defer func() {
			if ch1 != nil {
				close(ch1)
			}
		}()
	}

	// 执行SQL
	logMessage(fmt.Sprintf("执行SQL: %s", truncateSQL(e.Config.SQL, 200)))
	startTime := time.Now()

	// 发送开始信号（通知监控可以开始）
	if ch1 != nil {
		select {
		case ch1 <- 1:
		default:
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	result, err := db.ExecContext(ctx, e.Config.SQL)

	if err != nil {
		logMessage(fmt.Sprintf("执行失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}

	affectedRows, _ := result.RowsAffected()
	executeCostTime := time.Since(startTime).String()

	logMessage(fmt.Sprintf("执行成功，影响行数: %d，耗时: %s", affectedRows, executeCostTime))

	data.AffectedRows = affectedRows
	data.ExecuteCostTime = executeCostTime
	data.ExecuteLog = strings.Join(executeLog, "\n")
	return data, nil
}

// ExecuteDDLWithGhost 使用 gh-ost 执行 ALTER TABLE 语句
func (e *MySQLExecutor) ExecuteDDLWithGhost() (ReturnData, error) {
	var data ReturnData
	var executeLog []string

	logMessage := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logMsg := fmt.Sprintf("[%s] %s\n", timestamp, msg)
		executeLog = append(executeLog, logMsg)
		// 发布消息到 Redis（用于 WebSocket 推送，类型为 "ghost"）
		if e.Config.OrderID != "" {
			if err := utils.PublishMessageToChannel(e.Config.OrderID, logMsg, "ghost"); err != nil {
				global.Logger.Error("Failed to publish ghost message to Redis", zap.String("order_id", e.Config.OrderID), zap.Error(err))
			}
		} else {
			global.Logger.Warn("OrderID is empty, cannot publish ghost message to Redis", zap.String("task_id", e.Config.TaskID))
		}
	}

	logErrorAndReturn := func(err error, errMsg string) (ReturnData, error) {
		logMessage(errMsg + err.Error())
		data.ExecuteLog = strings.Join(executeLog, "")
		data.Error = err.Error()
		return data, err
	}

	// 检查 gh-ost 配置
	if global.Conf == nil {
		return logErrorAndReturn(errors.New("配置未初始化"), "gh-ost 配置未初始化，错误：")
	}

	ghostPath := global.Conf.GetString("ghost.path")
	if ghostPath == "" {
		return logErrorAndReturn(errors.New("gh-ost 路径未配置"), "gh-ost 路径未配置，错误：")
	}

	ghostArgs := global.Conf.GetStringSlice("ghost.args")
	if ghostArgs == nil {
		ghostArgs = []string{} // 使用默认参数
	}

	// 移除SQL语句前后的所有空白字符
	newSQL := strings.TrimSpace(e.Config.SQL)
	logMessage("移除SQL语句前后的所有空白字符，包括空格、制表符、换行符等")

	// 获取表名
	fullTableName, err := parser.GetTableNameFromAlterStatement(e.Config.SQL)
	if err != nil {
		return logErrorAndReturn(err, "解析SQL提取表名失败，错误：")
	}
	logMessage("从SQL语句中提取表名成功")

	// 处理表名：如果包含 schema.table 格式，需要分离
	var databaseName, tableName string
	if strings.Contains(fullTableName, ".") {
		parts := strings.SplitN(fullTableName, ".", 2)
		databaseName = strings.Trim(parts[0], "`")
		tableName = strings.Trim(parts[1], "`")
	} else {
		databaseName = e.Config.Schema
		tableName = strings.Trim(fullTableName, "`")
	}

	// 正则匹配提取 ALTER 子句
	syntax := `(?i)^ALTER(\s+)TABLE(\s+)([\S]*)(\s+)(ADD|CHANGE|RENAME|MODIFY|DROP|ENGINE|CONVERT)(\s*)([\S\s]*)`
	re, err := regexp.Compile(syntax)
	if err != nil {
		return logErrorAndReturn(err, "正则匹配SQL语句失败，错误：")
	}
	match := re.FindStringSubmatch(newSQL)
	if len(match) < 5 {
		return logErrorAndReturn(errors.New("正则匹配失败"), "正则匹配SQL语句失败")
	}
	logMessage("正则匹配SQL语句成功")

	// 将反引号处理为空，将双引号处理成单引号
	alterClause := strings.Join(match[5:], "")
	alterClause = strings.ReplaceAll(alterClause, "`", "")
	alterClause = strings.ReplaceAll(alterClause, "\"", "'")
	logMessage("将反引号处理为空，将双引号处理成单引号")

	// 生成 gh-ost 命令
	logMessage("生成gh-ost执行命令")

	// 生成 socket 路径并保存到 Redis（用于后续控制）
	socketPath := utils.GetGhostSocketPath(databaseName, tableName)
	if e.Config.OrderID != "" {
		if err := utils.SetGhostSocketPathToOrderID(e.Config.OrderID, socketPath); err != nil {
			global.Logger.Warn("Failed to save ghost socket path to Redis, control features may not work",
				zap.String("order_id", e.Config.OrderID),
				zap.String("socket_path", socketPath),
				zap.Error(err),
			)
		}
	}

	ghostCMDParts := []string{
		ghostPath,
		strings.Join(ghostArgs, " "),
		fmt.Sprintf("--user=\"%s\" --password=\"%s\"", e.Config.UserName, e.Config.Password),
		fmt.Sprintf("--host=\"%s\" --port=%d", e.Config.Hostname, e.Config.Port),
		fmt.Sprintf("--database=%s --table=%s", databaseName, tableName),
		fmt.Sprintf("--alter=\"%s\" --execute", alterClause),
		fmt.Sprintf("--serve-socket-file=%s", socketPath),
	}

	// 如果是阿里云 RDS，添加特殊参数
	if strings.Contains(e.Config.Hostname, "rds.aliyuncs.com") {
		ghostCMDParts = append(ghostCMDParts, "--aliyun-rds=true")
		ghostCMDParts = append(ghostCMDParts, fmt.Sprintf("--assume-master-host=\"%s\"", e.Config.Hostname))
	}

	// 如果选择了自动删除旧表，添加 -ok-to-drop-table 参数
	global.Logger.Info("检查 gh-ost 参数",
		zap.String("order_id", e.Config.OrderID),
		zap.Bool("ghost_ok_to_drop_table", e.Config.GhostOkToDropTable),
	)
	if e.Config.GhostOkToDropTable {
		ghostCMDParts = append(ghostCMDParts, "-ok-to-drop-table")
		logMessage("已添加 -ok-to-drop-table 参数：操作成功后自动删除旧表")
		global.Logger.Info("已添加 -ok-to-drop-table 参数到 gh-ost 命令",
			zap.String("order_id", e.Config.OrderID),
		)
	} else {
		global.Logger.Info("未添加 -ok-to-drop-table 参数",
			zap.String("order_id", e.Config.OrderID),
			zap.Bool("ghost_ok_to_drop_table", e.Config.GhostOkToDropTable),
		)
	}

	ghostCMD := strings.Join(ghostCMDParts, " ")

	startTime := time.Now()

	// 打印命令（掩码 password）
	re = regexp.MustCompile(`--password="([^"]*)"`)
	printGhostCMD := re.ReplaceAllString(ghostCMD, `--password="..."`)
	logMessage(fmt.Sprintf("执行gh-ost命令：%s", printGhostCMD))

	// 执行 gh-ost 命令
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan string, 100) // 使用缓冲通道，避免阻塞
	var commandOutput []string
	var commandError error
	done := make(chan error, 1) // 使用 error channel

	// 读取输出（在单独的 goroutine 中）
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				commandOutput = append(commandOutput, msg)
				// 实时推送日志（去除末尾换行符）
				trimmedMsg := strings.TrimRight(msg, "\n")
				if trimmedMsg != "" {
					logMessage(trimmedMsg)
					// 解析进度信息并推送
					e.parseAndPublishGhostProgress(trimmedMsg)
				}
			}
		}
	}()

	// 执行命令（在单独的 goroutine 中）
	go func() {
		done <- utils.Command(ctx, ch, ghostCMD)
	}()

	// 等待命令执行完成
	commandError = <-done
	// 命令执行完成，关闭 channel 让读取 goroutine 退出
	close(ch)
	// 等待读取 goroutine 处理完剩余输出
	time.Sleep(200 * time.Millisecond)

	if commandError != nil {
		executeLog = append(executeLog, commandOutput...)
		return logErrorAndReturn(commandError, "执行失败，错误：")
	}

	executeLog = append(executeLog, commandOutput...)
	executeCostTime := time.Since(startTime).String()

	logMessage("gh-ost命令执行成功")

	// 返回数据
	data.ExecuteLog = strings.Join(executeLog, "")
	data.ExecuteCostTime = executeCostTime
	return data, nil
}

// ExecuteDML 执行DML语句
func (e *MySQLExecutor) ExecuteDML() (ReturnData, error) {
	var data ReturnData
	var executeLog []string

	logMessage := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)
		executeLog = append(executeLog, logMsg)
		// 发布消息到 Redis（用于 WebSocket 推送）
		if e.Config.OrderID != "" {
			_ = utils.PublishMessageToChannel(e.Config.OrderID, logMsg, "")
		}
	}

	// 连接数据库
	logMessage(fmt.Sprintf("连接数据库 %s:%d...", e.Config.Hostname, e.Config.Port))
	db, err := e.Connect()
	if err != nil {
		logMessage(fmt.Sprintf("连接失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	defer db.Close()
	logMessage("连接成功")

	// 获取连接ID
	connectionID, err := mysqlpkg.GetConnectionID(db)
	if err != nil {
		logMessage(fmt.Sprintf("获取Connection ID失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	logMessage(fmt.Sprintf("Connection ID: %d", connectionID))

	// 获取执行开始前的binlog position
	var startFile string
	var startPosition int64
	startFile, startPosition, err = mysqlpkg.GetBinlogPos(db)
	if err != nil {
		logMessage(fmt.Sprintf("获取Start Binlog Position失败: %s（可能未开启binlog）", err.Error()))
		// 获取binlog position失败时继续执行
		startFile = ""
		startPosition = 0
	} else {
		logMessage(fmt.Sprintf("Start Binlog File: %s, Position: %d", startFile, startPosition))
	}

	// 启动 PROCESSLIST 监控（在单独的 goroutine 中）
	var ch1 chan int64
	if e.Config.OrderID != "" {
		ch1 = make(chan int64)
		go e.GetProcesslist(e.Config.OrderID, connectionID, ch1)
		// 确保在所有情况下都关闭 channel（函数返回时）
		defer func() {
			if ch1 != nil {
				close(ch1)
			}
		}()
	}

	// 开启事务
	logMessage("开启事务...")
	tx, err := db.Begin()
	if err != nil {
		logMessage(fmt.Sprintf("开启事务失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}

	// 执行SQL
	logMessage(fmt.Sprintf("执行SQL: %s", truncateSQL(e.Config.SQL, 200)))
	startTime := time.Now()

	// 发送开始信号（通知监控可以开始）
	if ch1 != nil {
		select {
		case ch1 <- 1:
		default:
		}
	}

	result, err := tx.Exec(e.Config.SQL)
	if err != nil {
		tx.Rollback()
		logMessage(fmt.Sprintf("执行失败，已回滚: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		logMessage(fmt.Sprintf("提交事务失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}

	affectedRows, _ := result.RowsAffected()
	executeCostTime := time.Since(startTime).String()

	logMessage(fmt.Sprintf("执行成功，影响行数: %d，耗时: %s", affectedRows, executeCostTime))

	data.AffectedRows = affectedRows
	data.ExecuteCostTime = executeCostTime

	// 如果影响行数大于0，且成功获取了binlog position，生成回滚SQL
	var rollbackSQL, backupCostTime string
	if affectedRows > 0 && startFile != "" {
		// 获取执行后的binlog position
		endFile, endPosition, err := mysqlpkg.GetBinlogPos(db)
		if err != nil {
			logMessage(fmt.Sprintf("获取End Binlog Position失败: %s", err.Error()))
		} else {
			logMessage(fmt.Sprintf("End Binlog File: %s, Position: %d", endFile, endPosition))
			logMessage("开始解析Binlog生成回滚SQL...")
			backupStartTime := time.Now()

			binlog := mysqlpkg.Binlog{
				Config: &mysqlpkg.BinlogConfig{
					Hostname: e.Config.Hostname,
					Port:     e.Config.Port,
					UserName: e.Config.UserName,
					Password: e.Config.Password,
					Schema:   e.Config.Schema,
				},
				ConnectionID:  connectionID,
				StartFile:     startFile,
				StartPosition: startPosition,
				EndFile:       endFile,
				EndPosition:   endPosition,
			}
			rollbackSQL, err = binlog.Run()
			if err != nil {
				logMessage(fmt.Sprintf("生成回滚SQL失败: %s", err.Error()))
			} else {
				backupCostTime = time.Since(backupStartTime).String()
				logMessage(fmt.Sprintf("生成回滚SQL成功，耗时: %s", backupCostTime))
			}
		}
	}

	data.RollbackSQL = rollbackSQL
	data.BackupCostTime = backupCostTime
	data.ExecuteLog = strings.Join(executeLog, "\n")
	return data, nil
}

// ExecuteExport 执行数据导出
func (e *MySQLExecutor) ExecuteExport() (ReturnData, error) {
	var data ReturnData
	var executeLog []string

	logMessage := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)
		executeLog = append(executeLog, logMsg)
		// 发布消息到 Redis（用于 WebSocket 推送）
		if e.Config.OrderID != "" {
			_ = utils.PublishMessageToChannel(e.Config.OrderID, logMsg, "")
		}
	}

	// 连接数据库
	logMessage(fmt.Sprintf("连接数据库 %s:%d...", e.Config.Hostname, e.Config.Port))
	db, err := e.Connect()
	if err != nil {
		logMessage(fmt.Sprintf("连接失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	defer db.Close()
	logMessage("连接成功")

	// 执行查询
	logMessage(fmt.Sprintf("执行查询: %s", truncateSQL(e.Config.SQL, 200)))
	startTime := time.Now()

	rows, err := db.Query(e.Config.SQL)
	if err != nil {
		logMessage(fmt.Sprintf("查询失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		logMessage(fmt.Sprintf("获取列信息失败: %s", err.Error()))
		data.ExecuteLog = strings.Join(executeLog, "\n")
		data.Error = err.Error()
		return data, err
	}

	// 统计行数
	var rowCount int64
	for rows.Next() {
		rowCount++
	}

	executeCostTime := time.Since(startTime).String()
	logMessage(fmt.Sprintf("查询成功，列数: %d，行数: %d，耗时: %s", len(columns), rowCount, executeCostTime))

	// TODO: 实际导出文件逻辑
	data.ExportRows = rowCount
	data.ExecuteCostTime = executeCostTime
	data.ExecuteLog = strings.Join(executeLog, "\n")
	return data, nil
}

// GetProcesslist 获取MySQL类数据库的Processlist（监控SQL执行状态）
// dbconfig: 数据库配置（用于创建新连接）
// orderID: 工单ID（用于WebSocket推送）
// connectionID: 连接ID（需要监控的连接）
// ch: 控制通道（当channel关闭时，停止监控）
func (e *MySQLExecutor) GetProcesslist(orderID string, connectionID int64, ch <-chan int64) {
	// 创建新的数据库连接
	monitorDB, err := e.Connect()
	if err != nil {
		global.Logger.Error("Failed to create monitor connection", zap.Error(err), zap.String("order_id", orderID))
		return
	}
	defer monitorDB.Close()

	// 构造查询SQL
	querySQL := fmt.Sprintf("SELECT * FROM INFORMATION_SCHEMA.PROCESSLIST WHERE ID=%d", connectionID)

	// 循环监控
	for {
		exitFlag := false
		select {
		case _, ok := <-ch:
			// 收到channel关闭信号，停止监控
			if !ok {
				exitFlag = true
			}
		case <-time.After(500 * time.Millisecond):
			// 每500ms查询一次
		}

		if exitFlag {
			break
		}

		// 执行查询
		rows, err := monitorDB.Query(querySQL)
		if err != nil {
			global.Logger.Error("Failed to get processlist", zap.Error(err), zap.String("order_id", orderID), zap.Int64("connection_id", connectionID))
			break
		}

		// 获取列信息
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			global.Logger.Error("Failed to get columns", zap.Error(err), zap.String("order_id", orderID))
			break
		}

		// 准备扫描数据
		vals := make([]interface{}, len(columns))
		for i := range columns {
			vals[i] = new(sql.RawBytes)
		}

		// 读取数据
		var rowData map[string]interface{}
		if rows.Next() {
			if err := rows.Scan(vals...); err != nil {
				rows.Close()
				global.Logger.Error("Failed to scan processlist row", zap.Error(err), zap.String("order_id", orderID))
				break
			}

			// 转换为 map
			rowData = make(map[string]interface{}, len(columns))
			for i, c := range vals {
				switch v := c.(type) {
				case *sql.RawBytes:
					if *v == nil {
						rowData[columns[i]] = nil
					} else {
						rowData[columns[i]] = string(*v)
					}
				}
			}
		}
		rows.Close()

		// 如果没有数据，停止监控
		if len(rowData) == 0 {
			global.Logger.Debug("Processlist row not found, stopping monitor", zap.String("order_id", orderID), zap.Int64("connection_id", connectionID))
			break
		}

		// 发布进程信息到 Redis（WebSocket推送）
		if orderID != "" {
			if err := utils.PublishMessageToChannel(orderID, rowData, "processlist"); err != nil {
				global.Logger.Error("Failed to publish processlist message", zap.Error(err), zap.String("order_id", orderID))
			}
		}
	}

	global.Logger.Debug("Processlist monitor stopped", zap.String("order_id", orderID), zap.Int64("connection_id", connectionID))
}

// parseAndPublishGhostProgress 解析 gh-ost 输出中的进度信息并推送
// gh-ost 输出示例：
// - "Copy: 12345/1000000 1.23%"
// - "Progress: 1.23% ETA: 2h30m"
// - "Rows: 12345/1000000 (1.23%)"
func (e *MySQLExecutor) parseAndPublishGhostProgress(line string) {
	if e.Config.OrderID == "" {
		return
	}

	// 定义多种进度匹配模式
	patterns := []struct {
		regex   *regexp.Regexp
		getData func([]string) map[string]interface{}
	}{
		// 匹配 "Copy: 12345/1000000 1.23%" 格式
		{
			regex: regexp.MustCompile(`(?i)Copy:\s*(\d+)/(\d+)\s+([\d.]+)%`),
			getData: func(matches []string) map[string]interface{} {
				current, _ := strconv.ParseInt(matches[1], 10, 64)
				total, _ := strconv.ParseInt(matches[2], 10, 64)
				percent, _ := strconv.ParseFloat(matches[3], 64)
				return map[string]interface{}{
					"current":   current,
					"total":     total,
					"percent":   percent,
					"operation": "copy",
				}
			},
		},
		// 匹配 "Progress: 1.23% ETA: 2h30m" 格式
		{
			regex: regexp.MustCompile(`(?i)Progress:\s*([\d.]+)%\s+ETA:\s*([\w]+)`),
			getData: func(matches []string) map[string]interface{} {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				return map[string]interface{}{
					"percent":   percent,
					"eta":       matches[2],
					"operation": "progress",
				}
			},
		},
		// 匹配 "Rows: 12345/1000000 (1.23%)" 格式
		{
			regex: regexp.MustCompile(`(?i)Rows:\s*(\d+)/(\d+)\s+\(([\d.]+)%\)`),
			getData: func(matches []string) map[string]interface{} {
				current, _ := strconv.ParseInt(matches[1], 10, 64)
				total, _ := strconv.ParseInt(matches[2], 10, 64)
				percent, _ := strconv.ParseFloat(matches[3], 64)
				return map[string]interface{}{
					"current":   current,
					"total":     total,
					"percent":   percent,
					"operation": "rows",
				}
			},
		},
		// 匹配包含百分比的通用格式，如 "1.23%" 或 "23.45%"
		{
			regex: regexp.MustCompile(`(?i)(\d+\.?\d*)\s*%`),
			getData: func(matches []string) map[string]interface{} {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				// 只处理 0-100 之间的百分比
				if percent >= 0 && percent <= 100 {
					return map[string]interface{}{
						"percent":   percent,
						"operation": "general",
					}
				}
				return nil
			},
		},
	}

	// 尝试匹配各种格式
	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(line)
		if len(matches) > 1 {
			progressData := pattern.getData(matches)
			if progressData != nil {
				// 推送进度信息到 WebSocket（类型为 "ghost-progress"）
				if err := utils.PublishMessageToChannel(e.Config.OrderID, progressData, "ghost-progress"); err != nil {
					global.Logger.Error("Failed to publish ghost progress", zap.String("order_id", e.Config.OrderID), zap.Error(err))
				}

				// 保存进度到 Redis 缓存
				if err := utils.SaveGhostProgressToRedis(e.Config.OrderID, progressData); err != nil {
					global.Logger.Warn("Failed to save ghost progress to Redis cache",
						zap.String("order_id", e.Config.OrderID),
						zap.Error(err),
					)
				}

				// 匹配第一个成功的模式后退出
				break
			}
		}
	}
}

// truncateSQL 截断SQL用于日志显示
func truncateSQL(sql string, maxLen int) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", " ")
	if len(sql) > maxLen {
		return sql[:maxLen] + "..."
	}
	return sql
}
