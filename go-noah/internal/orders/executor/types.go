package executor

// DBConfig 数据库配置
type DBConfig struct {
	Hostname           string // 主机名
	Port               int    // 端口
	Charset            string // 字符集
	UserName           string // 用户名
	Password           string // 密码
	Schema             string // 数据库
	DBType             string // 数据库类型（MySQL/TiDB/ClickHouse）
	SQLType            string // SQL类型（DDL/DML/EXPORT）
	SQL                string // SQL语句
	OrderID            string // 工单ID
	TaskID             string // 任务ID
	ExportFileFormat   string // 导出文件格式
	GhostOkToDropTable bool   // gh-ost执行成功后自动删除旧表
	GenerateRollback   bool   // DML 是否生成回滚语句，仅 DML 有效
}

// ExportFile 导出文件信息
type ExportFile struct {
	FileName      string `json:"file_name"`      // 文件名
	FileSize      int64  `json:"file_size"`      // 文件大小
	FilePath      string `json:"file_path"`      // 文件路径
	ContentType   string `json:"content_type"`   // 内容类型
	EncryptionKey string `json:"encryption_key"` // 加密密钥
	ExportRows    int64  `json:"export_rows"`    // 导出行数
	DownloadUrl   string `json:"download_url"`   // 下载地址
}

// ReturnData 执行结果
type ReturnData struct {
	RollbackSQL     string `json:"rollback_sql"`      // 回滚SQL
	AffectedRows    int64  `json:"affected_rows"`     // 影响行数
	ExecuteCostTime string `json:"execute_cost_time"` // 执行耗时
	BackupCostTime  string `json:"backup_cost_time"`  // 备份耗时
	ExecuteLog      string `json:"execute_log"`       // 执行日志
	ExportFile             // 导出文件信息
	Error           string `json:"error"` // 错误信息
}

// Executor 执行器接口
type Executor interface {
	Run() (ReturnData, error)
}
