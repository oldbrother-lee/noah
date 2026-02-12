package executor

import "fmt"

// ExecuteSQL 执行SQL的统一入口
type ExecuteSQL struct {
	Config   *DBConfig
	Executor Executor
}

// NewExecuteSQL 创建执行器
func NewExecuteSQL(config *DBConfig) (*ExecuteSQL, error) {
	var executor Executor

	switch config.DBType {
	case "MySQL", "TiDB":
		executor = NewMySQLExecutor(config)
	case "ClickHouse":
		// TODO: ClickHouse执行器
		return nil, fmt.Errorf("ClickHouse执行器暂未实现")
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.DBType)
	}

	return &ExecuteSQL{
		Config:   config,
		Executor: executor,
	}, nil
}

// Run 执行SQL
func (e *ExecuteSQL) Run() (ReturnData, error) {
	return e.Executor.Run()
}

