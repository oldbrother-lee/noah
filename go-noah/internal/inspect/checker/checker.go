package checker

import (
	"fmt"
	"strings"

	"go-noah/internal/inspect/config"
	"go-noah/internal/inspect/dao"
	"go-noah/internal/inspect/parser"
	"go-noah/pkg/global"
	"go-noah/pkg/kv"
	"go-noah/pkg/query"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"go.uber.org/zap"
)

// AuditLevel 审核级别
type AuditLevel string

const (
	LevelPass    AuditLevel = "PASS"    // 通过
	LevelInfo    AuditLevel = "INFO"    // 信息（审核通过）
	LevelNotice  AuditLevel = "NOTICE"  // 提示
	LevelWarning AuditLevel = "WARNING" // 警告
	LevelError   AuditLevel = "ERROR"   // 错误
)

// AuditResult 单条SQL审核结果
type AuditResult struct {
	SQL           string     `json:"query"`          // 原始SQL（使用query以兼容老服务）
	Type          string     `json:"type"`           // SQL类型
	Level         AuditLevel `json:"level"`          // 审核级别
	AffectedRows  int64      `json:"affected_rows"`  // 影响行数
	Messages      []string   `json:"messages"`       // 审核信息
	Summary       []string   `json:"summary"`        // 摘要（兼容原接口）
	FixSuggestion string     `json:"fix_suggestion"` // 修复建议
}

// Checker SQL审核器
type Checker struct {
	Params     *config.InspectParams // 审核参数
	DBType     string                // 数据库类型
	Results    []*AuditResult        // 审核结果
	DBHost     string                // 数据库主机
	DBPort     int                   // 数据库端口
	DBUser     string                // 数据库用户
	DBPassword string                // 数据库密码
	DBSchema   string                // 数据库名
}

// NewChecker 创建审核器
func NewChecker(params *config.InspectParams, dbType string) *Checker {
	if params == nil {
		params = config.DefaultInspectParams()
	}
	return &Checker{
		Params:  params,
		DBType:  dbType,
		Results: make([]*AuditResult, 0),
	}
}

// SetDBInfo 设置数据库连接信息
func (c *Checker) SetDBInfo(host string, port int, user, password, schema string) {
	c.DBHost = host
	c.DBPort = port
	c.DBUser = user
	c.DBPassword = password
	c.DBSchema = schema
}

// Check 执行审核（使用规则引擎架构）
func (c *Checker) Check(sqlText string) ([]*AuditResult, error) {
	// 初始化数据库连接
	db := &dao.DB{
		User:     c.DBUser,
		Password: c.DBPassword,
		Host:     c.DBHost,
		Port:     c.DBPort,
		Database: c.DBSchema,
	}

	// 创建 KVCache（使用简单ID，因为没有 gin.Context）
	requestID := "inspect_" + query.Id(query.Fingerprint(sqlText))
	kvCache := kv.NewKVCache(requestID)
	defer kvCache.Delete(requestID)

	// 获取数据库变量
	dbVars, err := dao.GetDBVars(db)
	if err != nil {
		if global.Logger != nil {
			global.Logger.Warn("获取DB变量失败，使用默认值",
				zap.Error(err),
				zap.String("host", c.DBHost),
				zap.Int("port", c.DBPort),
			)
		}
		dbVars = map[string]string{
			"dbVersion":              "",
			"dbCharset":              "utf8mb4",
			"largePrefix":            "OFF",
			"innodbDefaultRowFormat": "dynamic",
		}
	}
	for k, v := range dbVars {
		kvCache.Put(k, v)
	}

	charset := dbVars["dbCharset"]
	if charset == "" {
		charset = "utf8mb4"
	}

	// 解析SQL
	audit, _, err := parser.NewParse(sqlText, charset, "")
	if err != nil {
		return nil, fmt.Errorf("SQL解析错误: %s", err.Error())
	}

	// 创建 SyntaxInspectService
	service := &SyntaxInspectService{
		DB:            db,
		Audit:         audit,
		InspectParams: c.Params,
		Charset:       charset,
		Collation:     "",
	}

	// 创建 Stmt
	stmt := &Stmt{service}

	var returnDataList []ReturnData
	var mergeAlters []string

	// 遍历每条语句进行审核
	for _, stmtNode := range audit.TiStmt {
		// 移除SQL尾部的分号
		sqlTrim := strings.TrimSuffix(stmtNode.Text(), ";")
		// 生成指纹ID
		fingerId := query.Id(query.Fingerprint(sqlTrim))
		// 存储指纹ID
		kvCache.Put(fingerId, true)

		switch stmtNode.(type) {
		case *ast.SelectStmt:
			// select语句不允许审核
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmtNode.Text(), Type: "DML", Level: "WARN"}
			data.Summary = append(data.Summary, "发现SELECT语句，请删除SELECT语句后重新审核")
			returnDataList = append(returnDataList, data)
		case *ast.CreateTableStmt:
			returnDataList = append(returnDataList, stmt.CreateTableStmt(stmtNode, kvCache, fingerId))
		case *ast.CreateViewStmt:
			returnDataList = append(returnDataList, stmt.CreateViewStmt(stmtNode, kvCache, fingerId))
		case *ast.AlterTableStmt:
			data, mergeAlter := stmt.AlterTableStmt(stmtNode, kvCache, fingerId)
			// 只添加非空的 mergeAlter，用于后续合并检查
			if len(mergeAlter) > 0 {
				mergeAlters = append(mergeAlters, mergeAlter)
			}
			returnDataList = append(returnDataList, data)
		case *ast.DropTableStmt, *ast.TruncateTableStmt:
			returnDataList = append(returnDataList, stmt.DropTableStmt(stmtNode, kvCache, fingerId))
		case *ast.DeleteStmt, *ast.InsertStmt, *ast.UpdateStmt:
			returnDataList = append(returnDataList, stmt.DMLStmt(stmtNode, kvCache, fingerId))
		case *ast.RenameTableStmt:
			returnDataList = append(returnDataList, stmt.RenameTableStmt(stmtNode, kvCache, fingerId))
		case *ast.AnalyzeTableStmt:
			returnDataList = append(returnDataList, stmt.AnalyzeTableStmt(stmtNode, kvCache, fingerId))
		case *ast.CreateDatabaseStmt:
			returnDataList = append(returnDataList, stmt.CreateDatabaseStmt(stmtNode, kvCache, fingerId))
		default:
			// 不允许的其他语句
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmtNode.Text(), Type: "", Level: "WARN"}
			data.Summary = append(data.Summary, "未识别或禁止的审核语句，请联系数据库管理员")
			returnDataList = append(returnDataList, data)
		}
	}

	// 判断多条alter语句是否需要合并
	if len(mergeAlters) > 1 {
		mergeData := service.mergeAlters(kvCache, mergeAlters)
		if len(mergeData.Summary) > 0 {
			returnDataList = append(returnDataList, mergeData)
		}
	}

	// 转换为 AuditResult
	var results []*AuditResult
	for _, rd := range returnDataList {
		results = append(results, rd.ToAuditResult())
	}

	return results, nil
}
