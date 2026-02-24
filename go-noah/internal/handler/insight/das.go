package insight

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"go-noah/api"
	"go-noah/internal/das/dao"
	dasParser "go-noah/internal/das/parser"
	"go-noah/internal/handler"
	"go-noah/internal/inspect/extract"
	"go-noah/internal/inspect/parser"
	"go-noah/internal/model/insight"
	"go-noah/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"gorm.io/gorm"
)

// DASHandlerApp 全局 Handler 实例
var DASHandlerApp = new(DASHandler)

// DASHandler DAS Handler
type DASHandler struct{}

// validateInstanceForPermission 验证实例是否为"查询"类型，防止越权
func (h *DASHandler) validateInstanceForPermission(ctx context.Context, instanceID string) error {
	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("获取数据库配置失败: %w", err)
	}

	if config.UseType != insight.UseTypeQuery {
		return errors.New("只能绑定类型为'查询'的数据库实例，防止越权")
	}

	return nil
}

// ExecuteQueryRequest 执行查询请求
type ExecuteQueryRequest struct {
	InstanceID string            `json:"instance_id" binding:"required"`
	Schema     string            `json:"schema" binding:"required"`
	SQLText    string            `json:"sqltext" binding:"required"`
	Params     map[string]string `json:"params"`
}

// ExecuteQueryResponse 执行查询响应
type ExecuteQueryResponse struct {
	Columns    []string                 `json:"columns"`
	Data       []map[string]interface{} `json:"data"`
	Duration   int64                    `json:"duration"`
	RowCount   int                      `json:"row_count"`
	SQLText    string                   `json:"sqltext"`     // 原始 SQL
	RewriteSQL string                   `json:"rewrite_sql"` // 重写后的 SQL
}

// ExecuteQuery 执行SQL查询
// @Summary 执行SQL查询
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body ExecuteQueryRequest true "查询请求"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/query [post]
func (h *DASHandler) ExecuteQuery(c *gin.Context) {
	var req ExecuteQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	// 获取用户名
	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 获取数据库配置
	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), req.InstanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 检查用户权限（基于角色权限）
	effectivePerms, err := service.InsightServiceApp.GetUserEffectivePermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 检查是否有该库的权限
	hasSchemaPermission := false
	var allowedTables []string // 允许访问的表列表（如果为空，表示整个库都有权限）

	for _, perm := range effectivePerms {
		if perm.InstanceID == req.InstanceID && perm.Schema == req.Schema {
			if perm.Table != "" {
				// 有表权限限制，收集允许的表
				allowedTables = append(allowedTables, strings.ToLower(perm.Table))
			} else {
				// 整个库都有权限，清空表限制
				allowedTables = nil
				hasSchemaPermission = true
				break
			}
		}
	}

	// 如果没有库权限，拒绝访问
	if !hasSchemaPermission && len(allowedTables) == 0 {
		api.HandleError(c, http.StatusForbidden, api.ErrForbidden, nil)
		return
	}

	// 如果有表权限限制，需要检查 SQL 是否只访问允许的表
	if len(allowedTables) > 0 {
		// 解析 SQL 提取表名
		audit, warns, err := parser.ParseSQL(req.SQLText)
		if err != nil {
			// SQL 解析失败，可能是非标准 SQL（如 SHOW, DESCRIBE 等），允许执行
			hasSchemaPermission = true
		} else if len(warns) > 0 {
			// 有警告但不影响解析，继续检查
		}

		if audit != nil && len(audit.TiStmt) > 0 {
			// 提取 SQL 中涉及的表名
			traverse := &extract.TraverseStatement{}
			for _, stmt := range audit.TiStmt {
				stmt.Accept(traverse)
			}

			// 检查 SQL 中涉及的表是否都在允许列表中
			sqlTables := make(map[string]bool)
			for _, table := range traverse.Tables {
				sqlTables[strings.ToLower(table)] = true
			}

			// 如果 SQL 中没有涉及任何表（如 SHOW, DESCRIBE 等），允许执行
			if len(sqlTables) == 0 {
				hasSchemaPermission = true
			} else {
				// 检查所有 SQL 中的表是否都在允许列表中
				allTablesAllowed := true
				for table := range sqlTables {
					tableAllowed := false
					for _, allowedTable := range allowedTables {
						if table == allowedTable {
							tableAllowed = true
							break
						}
					}
					if !tableAllowed {
						allTablesAllowed = false
						break
					}
				}
				hasSchemaPermission = allTablesAllowed
			}
		} else {
			// 无法解析 SQL，可能是非标准 SQL，允许执行
			hasSchemaPermission = true
		}
	}

	if !hasSchemaPermission {
		api.HandleError(c, http.StatusForbidden, api.ErrForbidden, nil)
		return
	}

	// 生成 request ID
	requestID, err := uuid.NewUUID()
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	requestIDStr := requestID.String()

	// 重写 SQL（增加 hint 和重写 limit）
	rewriteSQL := req.SQLText
	audit, warns, err := parser.ParseSQL(req.SQLText)
	if err == nil && len(audit.TiStmt) > 0 {
		// 只处理 SELECT 语句和 UNION 语句
		stmtLoop:
		for _, stmt := range audit.TiStmt {
			switch stmt.(type) {
			case *ast.SelectStmt, *ast.SetOprStmt:
				rewrite := &dasParser.Rewrite{
					Stmt:      stmt,
					RequestID: requestIDStr,
					DbType:    string(config.DbType),
				}
				rewriteSQL = rewrite.Run()
				break stmtLoop // 只处理第一个 SELECT/UNION 语句
			}
		}
		// 如果有警告但不影响解析，继续执行
		_ = warns
	}

	// 创建数据库连接
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	db := &dao.MySQLDB{
		User:     config.UserName,
		Password: config.Password,
		Host:     config.Hostname,
		Port:     config.Port,
		Database: req.Schema,
		Params:   req.Params,
		Ctx:      ctx,
	}

	// 执行查询（使用重写后的 SQL）
	startTime := time.Now()
	columns, data, err := db.Query(rewriteSQL)
	duration := time.Since(startTime).Milliseconds()

	// 记录执行日志（保存原始 SQL 和重写后的 SQL）
	record := &insight.DASRecord{
		Username:   username,
		InstanceID: config.InstanceID,
		Schema:     req.Schema,
		SQL:        req.SQLText, // 保存原始 SQL
		Duration:   duration,
		RowCount:   int64(len(data)),
	}
	if err != nil {
		record.Error = err.Error()
	}
	_ = service.InsightServiceApp.CreateDASRecord(c.Request.Context(), record)

	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, ExecuteQueryResponse{
		Columns:    columns,
		Data:       data,
		Duration:   duration,
		RowCount:   len(data),
		SQLText:    req.SQLText, // 原始 SQL
		RewriteSQL: rewriteSQL,  // 重写后的 SQL
	})
}

// UserSchemaResult 用户授权的 schema 返回结构
type UserSchemaResult struct {
	InstanceID uuid.UUID `json:"instance_id"`
	Schema     string    `json:"schema"`
	DbType     string    `json:"db_type"`
	Hostname   string    `json:"hostname"`
	Port       int       `json:"port"`
	IsDeleted  bool      `json:"is_deleted"`
	Remark     string    `json:"remark"`
}

// GetUserSchemas 获取用户有权限的所有Schema（用于SQL查询页面）
// @Summary 获取用户授权的Schema列表
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/schemas [get]
func (h *DASHandler) GetUserSchemas(c *gin.Context) {
	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	// 获取用户名
	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 获取用户授权的 schemas
	results, err := service.InsightServiceApp.GetUserAuthorizedSchemas(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, results)
}

// GetSchemas 获取实例下的数据库列表
// @Summary 获取数据库列表
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/schemas/{instance_id} [get]
func (h *DASHandler) GetSchemas(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取数据库配置
	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	db := &dao.MySQLDB{
		User:     config.UserName,
		Password: config.Password,
		Host:     config.Hostname,
		Port:     config.Port,
		Ctx:      ctx,
	}

	schemas, err := db.GetSchemas()
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, schemas)
}

// GetTables 获取指定库的表列表（只返回用户有权限的表）
// @Summary 获取表列表
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Param schema path string true "数据库名"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/tables/{instance_id}/{schema} [get]
func (h *DASHandler) GetTables(c *gin.Context) {
	instanceID := c.Param("instance_id")
	schema := c.Param("schema")
	// 验证参数：不能为空，也不能是 "undefined"（前端可能传递字符串 "undefined"）
	if instanceID == "" || schema == "" || instanceID == "undefined" || schema == "undefined" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	// 获取数据库配置
	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	db := &dao.MySQLDB{
		User:     config.UserName,
		Password: config.Password,
		Host:     config.Hostname,
		Port:     config.Port,
		Ctx:      ctx,
		Params:   map[string]string{"group_concat_max_len": "4194304"},
	}

	// 获取用户名
	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 获取用户的有效权限（包括角色权限）
	effectivePerms, err := service.InsightServiceApp.GetUserEffectivePermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 检查用户对该库的权限
	hasSchemaPermission := false
	allowedTables := make(map[string]bool) // 允许访问的表列表（如果为空，表示整个库都有权限）

	for _, perm := range effectivePerms {
		if perm.InstanceID == instanceID && perm.Schema == schema {
			if perm.Table != "" {
				// 有表权限限制，收集允许的表
				allowedTables[strings.ToLower(perm.Table)] = true
			} else {
				// 整个库都有权限，清空表限制
				allowedTables = nil
				hasSchemaPermission = true
				break
			}
		}
	}

	// 如果没有权限，返回空列表（即使是管理员，如果没有权限也返回空）
	if !hasSchemaPermission && len(allowedTables) == 0 {
		api.HandleSuccess(c, []interface{}{})
		return
	}

	// 获取所有表
	allTables, err := db.GetTables(schema)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 如果有整个库的权限，返回所有表
	if hasSchemaPermission {
		api.HandleSuccess(c, allTables)
		return
	}

	// 如果只有表权限，过滤表列表
	var filteredTables []dao.TableMetaData
	for _, table := range allTables {
		tableName := strings.ToLower(table.TableName)
		if allowedTables[tableName] {
			filteredTables = append(filteredTables, table)
		}
	}

	api.HandleSuccess(c, filteredTables)
}

// GetTableColumns 获取表结构
// @Summary 获取表结构
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Param schema path string true "数据库名"
// @Param table path string true "表名"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/columns/{instance_id}/{schema}/{table} [get]
func (h *DASHandler) GetTableColumns(c *gin.Context) {
	instanceID := c.Param("instance_id")
	schema := c.Param("schema")
	table := c.Param("table")
	// 验证参数：不能为空，也不能是 "undefined"（前端可能传递字符串 "undefined"）
	if instanceID == "" || schema == "" || table == "" || table == "undefined" || schema == "undefined" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	db := &dao.MySQLDB{
		User:     config.UserName,
		Password: config.Password,
		Host:     config.Hostname,
		Port:     config.Port,
		Ctx:      ctx,
	}

	columns, err := db.GetTableColumns(schema, table)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, columns)
}

// GetRecords 获取执行记录
// @Summary 获取执行记录
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/records [get]
func (h *DASHandler) GetRecords(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := service.InsightServiceApp.GetDASRecords(c.Request.Context(), username, page, pageSize)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, gin.H{
		"list":  records,
		"total": total,
	})
}

// ============ 收藏夹 ============

// CreateFavoriteRequest 创建收藏请求
type CreateFavoriteRequest struct {
	Title string `json:"title" binding:"required"`
	SQL   string `json:"sql" binding:"required"`
}

// GetFavorites 获取收藏列表
// @Summary 获取收藏列表
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/favorites [get]
func (h *DASHandler) GetFavorites(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	favorites, err := service.InsightServiceApp.GetFavorites(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, favorites)
}

// CreateFavorite 创建收藏
// @Summary 创建收藏
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body CreateFavoriteRequest true "收藏信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/favorites [post]
func (h *DASHandler) CreateFavorite(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	var req CreateFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	fav := &insight.DASFavorite{
		Username: username,
		Title:    req.Title,
		SQL:      req.SQL,
	}

	if err := service.InsightServiceApp.CreateFavorite(c.Request.Context(), fav); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, fav)
}

// DeleteFavorite 删除收藏
// @Summary 删除收藏
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "收藏ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/favorites/{id} [delete]
func (h *DASHandler) DeleteFavorite(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}

	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteFavorite(c.Request.Context(), uint(id), username); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// ============ 权限管理 ============

// GrantSchemaPermissionRequest 授权Schema权限请求
type GrantSchemaPermissionRequest struct {
	Username   string `json:"username" binding:"required"`
	InstanceID string `json:"instance_id" binding:"required"`
	Schema     string `json:"schema" binding:"required"`
}

// GrantTablePermissionRequest 授权Table权限请求
type GrantTablePermissionRequest struct {
	Username   string `json:"username" binding:"required"`
	InstanceID string `json:"instance_id" binding:"required"`
	Schema     string `json:"schema" binding:"required"`
	Table      string `json:"table" binding:"required"`
	Rule       string `json:"rule"` // allow 或 deny，默认为 allow
}

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id query string false "实例ID（可选，用于过滤）"
// @Param schema query string false "库名（可选，用于过滤）"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/permissions [get]
func (h *DASHandler) GetUserPermissions(c *gin.Context) {
	// 仅使用当前登录用户，不允许传 username，避免越权查询他人权限
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}
	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 获取过滤参数
	instanceID := c.Query("instance_id")
	schema := c.Query("schema")

	// 如果提供了过滤参数，返回用户的有效权限（包括角色权限）
	if instanceID != "" || schema != "" {
		// 获取用户的有效权限（包括角色权限和直接权限）
		effectivePerms, err := service.InsightServiceApp.GetUserEffectivePermissions(c.Request.Context(), username)
		if err != nil {
			api.HandleError(c, http.StatusInternalServerError, err, nil)
			return
		}

		// 过滤权限
		var filteredPerms []insight.PermissionObject
		for _, perm := range effectivePerms {
			if instanceID != "" && perm.InstanceID != instanceID {
				continue
			}
			if schema != "" && perm.Schema != schema {
				continue
			}
			filteredPerms = append(filteredPerms, perm)
		}

		// 转换为 DAS 编辑页面需要的格式
		var schemaPerms []insight.DASUserSchemaPermission
		var tablePerms []insight.DASUserTablePermission
		tables := []gin.H{}

		hasSchemaPermission := false      // 是否有整个库的权限
		tableMap := make(map[string]bool) // 去重表权限

		for _, perm := range filteredPerms {
			if perm.Table == "" {
				// 整个库的权限
				instanceUUID, err := uuid.Parse(perm.InstanceID)
				if err == nil {
					schemaPerms = append(schemaPerms, insight.DASUserSchemaPermission{
						Username:   username,
						InstanceID: instanceUUID,
						Schema:     perm.Schema,
					})
					hasSchemaPermission = true
				}
			} else {
				// 表权限
				instanceUUID, err := uuid.Parse(perm.InstanceID)
				if err == nil {
					tableKey := perm.Table
					if !tableMap[tableKey] {
						tablePerms = append(tablePerms, insight.DASUserTablePermission{
							Username:   username,
							InstanceID: instanceUUID,
							Schema:     perm.Schema,
							Table:      perm.Table,
							Rule:       insight.RuleAllow,
						})
						tables = append(tables, gin.H{
							"table": perm.Table,
							"rule":  "allow",
						})
						tableMap[tableKey] = true
					}
				}
			}
		}

		// 如果有库权限但没有表权限，表示所有表都有权限
		if hasSchemaPermission && len(tables) == 0 {
			tables = []gin.H{{"table": "*", "rule": "allow"}}
		}

		api.HandleSuccess(c, gin.H{
			"schema_permissions": schemaPerms,
			"table_permissions":  tablePerms,
			"tables":             tables, // 兼容 DAS 编辑页面的格式
		})
		return
	}

	// 如果没有过滤参数，返回用户的直接权限（用于权限管理页面）
	schemaPerms, err := service.InsightServiceApp.GetUserSchemaPermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	tablePerms, err := service.InsightServiceApp.GetUserTablePermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, gin.H{
		"schema_permissions": schemaPerms,
		"table_permissions":  tablePerms,
	})
}

// ============ 权限模板管理 ============

// GetPermissionTemplates 获取权限模板列表
func (h *DASHandler) GetPermissionTemplates(c *gin.Context) {
	templates, err := service.InsightServiceApp.GetPermissionTemplates(c.Request.Context())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, templates)
}

// GetPermissionTemplate 获取权限模板详情
func (h *DASHandler) GetPermissionTemplate(c *gin.Context) {
	id := c.Param("id")
	var templateID uint
	if _, err := fmt.Sscanf(id, "%d", &templateID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	template, err := service.InsightServiceApp.GetPermissionTemplate(c.Request.Context(), templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.HandleError(c, http.StatusNotFound, err, nil)
			return
		}
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, template)
}

// CreatePermissionTemplate 创建权限模板
func (h *DASHandler) CreatePermissionTemplate(c *gin.Context) {
	var req struct {
		Name        string                     `json:"name" binding:"required"`
		Description string                     `json:"description"`
		Permissions []insight.PermissionObject `json:"permissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 验证所有权限中的实例ID是否为"查询"类型
	for _, perm := range req.Permissions {
		if perm.InstanceID != "" {
			if err := h.validateInstanceForPermission(c.Request.Context(), perm.InstanceID); err != nil {
				api.HandleError(c, http.StatusBadRequest, err, nil)
				return
			}
		}
	}

	template := &insight.DASPermissionTemplate{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	if err := service.InsightServiceApp.CreatePermissionTemplate(c.Request.Context(), template); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, template)
}

// UpdatePermissionTemplate 更新权限模板
func (h *DASHandler) UpdatePermissionTemplate(c *gin.Context) {
	id := c.Param("id")
	var templateID uint
	if _, err := fmt.Sscanf(id, "%d", &templateID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	var req struct {
		Name        string                     `json:"name" binding:"required"`
		Description string                     `json:"description"`
		Permissions []insight.PermissionObject `json:"permissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 验证所有权限中的实例ID是否为"查询"类型
	for _, perm := range req.Permissions {
		if perm.InstanceID != "" {
			if err := h.validateInstanceForPermission(c.Request.Context(), perm.InstanceID); err != nil {
				api.HandleError(c, http.StatusBadRequest, err, nil)
				return
			}
		}
	}

	template := &insight.DASPermissionTemplate{
		Model:       gorm.Model{ID: templateID},
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	if err := service.InsightServiceApp.UpdatePermissionTemplate(c.Request.Context(), template); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, template)
}

// DeletePermissionTemplate 删除权限模板
func (h *DASHandler) DeletePermissionTemplate(c *gin.Context) {
	id := c.Param("id")
	var templateID uint
	if _, err := fmt.Sscanf(id, "%d", &templateID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	if err := service.InsightServiceApp.DeletePermissionTemplate(c.Request.Context(), templateID); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// ============ 角色权限管理 ============

// GetRolePermissions 获取角色权限列表
func (h *DASHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		api.HandleError(c, http.StatusBadRequest, fmt.Errorf("role is required"), nil)
		return
	}

	perms, err := service.InsightServiceApp.GetRolePermissions(c.Request.Context(), role)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 为模板类型的权限关联查询模板名称
	type RolePermissionWithTemplate struct {
		insight.DASRolePermission
		TemplateName string `json:"template_name,omitempty"`
	}

	result := make([]RolePermissionWithTemplate, 0, len(perms))
	for _, perm := range perms {
		rp := RolePermissionWithTemplate{
			DASRolePermission: perm,
		}
		// 如果是模板类型，查询模板名称
		if perm.PermissionType == insight.PermissionTypeTemplate {
			template, err := service.InsightServiceApp.GetPermissionTemplate(c.Request.Context(), perm.PermissionID)
			if err == nil && template != nil {
				rp.TemplateName = template.Name
			}
		}
		result = append(result, rp)
	}

	api.HandleSuccess(c, result)
}

// CreateRolePermission 创建角色权限
func (h *DASHandler) CreateRolePermission(c *gin.Context) {
	var req struct {
		Role           string `json:"role" binding:"required"`
		PermissionType string `json:"permission_type" binding:"required,oneof=object template"`
		PermissionID   uint   `json:"permission_id"`
		InstanceID     string `json:"instance_id,omitempty"`
		Schema         string `json:"schema,omitempty"`
		Table          string `json:"table,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 根据权限类型进行验证
	if req.PermissionType == "template" {
		if req.PermissionID == 0 {
			api.HandleError(c, http.StatusBadRequest, errors.New("权限模板类型必须提供 permission_id"), nil)
			return
		}
	} else if req.PermissionType == "object" {
		if req.InstanceID == "" || req.Schema == "" {
			api.HandleError(c, http.StatusBadRequest, errors.New("直接权限类型必须提供 instance_id 和 schema"), nil)
			return
		}

		// 验证实例是否为"查询"类型，防止越权
		if err := h.validateInstanceForPermission(c.Request.Context(), req.InstanceID); err != nil {
			api.HandleError(c, http.StatusBadRequest, err, nil)
			return
		}

		// 对于直接权限，使用 instance_id + schema 生成一个唯一的 permission_id
		// 使用 MD5 哈希的前 8 个字节转换为 uint64，然后取模确保在合理范围内
		if req.PermissionID == 0 {
			hash := md5.Sum([]byte(req.InstanceID + ":" + req.Schema))
			// 将前 8 个字节转换为 uint64，然后取模到合理的范围（避免溢出）
			var hashValue uint64
			for i := 0; i < 8; i++ {
				hashValue = hashValue<<8 | uint64(hash[i])
			}
			// 取模到 1000000000 以内，确保不会溢出 uint 类型
			req.PermissionID = uint(hashValue % 1000000000)
			// 确保不为 0
			if req.PermissionID == 0 {
				req.PermissionID = 1
			}
		}
	}

	perm := &insight.DASRolePermission{
		Role:           req.Role,
		PermissionType: insight.PermissionType(req.PermissionType),
		PermissionID:   req.PermissionID,
		InstanceID:     req.InstanceID,
		Schema:         req.Schema,
		Table:          req.Table,
	}

	if err := service.InsightServiceApp.CreateRolePermission(c.Request.Context(), perm); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, perm)
}

// DeleteRolePermission 删除角色权限
func (h *DASHandler) DeleteRolePermission(c *gin.Context) {
	id := c.Param("id")
	var permID uint
	if _, err := fmt.Sscanf(id, "%d", &permID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteRolePermission(c.Request.Context(), permID); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// GetUserEffectivePermissions 获取当前用户实际生效的权限（不接收 username，防越权）
func (h *DASHandler) GetUserEffectivePermissions(c *gin.Context) {
	userId := handler.GetUserIdFromCtx(c)
	if userId == 0 {
		api.HandleError(c, http.StatusUnauthorized, api.ErrUnauthorized, nil)
		return
	}
	username, err := h.getUsernameByID(c, userId)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	perms, err := service.InsightServiceApp.GetUserEffectivePermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, perms)
}

// ============ 用户权限管理（与角色权限同构：object/template，无 rule）===========

// GetUserPermissionList 按用户名获取用户权限列表（与角色权限同结构，供管理端「用户权限」Tab 使用）
func (h *DASHandler) GetUserPermissionList(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		api.HandleError(c, http.StatusBadRequest, errors.New("username is required"), nil)
		return
	}
	list, err := service.InsightServiceApp.GetUserPermissions(c.Request.Context(), username)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, list)
}

// CreateUserPermission 创建用户权限（请求体与角色权限一致：permission_type object/template，无 rule）
func (h *DASHandler) CreateUserPermission(c *gin.Context) {
	var req struct {
		Username       string `json:"username" binding:"required"`
		PermissionType string `json:"permission_type" binding:"required,oneof=object template"`
		PermissionID   uint   `json:"permission_id"`
		InstanceID     string `json:"instance_id,omitempty"`
		Schema         string `json:"schema,omitempty"`
		Table          string `json:"table,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}
	if req.PermissionType == "template" {
		if req.PermissionID == 0 {
			api.HandleError(c, http.StatusBadRequest, errors.New("权限模板类型必须提供 permission_id"), nil)
			return
		}
	} else if req.PermissionType == "object" {
		if req.InstanceID == "" || req.Schema == "" {
			api.HandleError(c, http.StatusBadRequest, errors.New("直接权限类型必须提供 instance_id 和 schema"), nil)
			return
		}
		if err := h.validateInstanceForPermission(c.Request.Context(), req.InstanceID); err != nil {
			api.HandleError(c, http.StatusBadRequest, err, nil)
			return
		}
		if req.PermissionID == 0 {
			hash := md5.Sum([]byte(req.InstanceID + ":" + req.Schema))
			var hashValue uint64
			for i := 0; i < 8; i++ {
				hashValue = hashValue<<8 | uint64(hash[i])
			}
			req.PermissionID = uint(hashValue % 1000000000)
			if req.PermissionID == 0 {
				req.PermissionID = 1
			}
		}
	}
	perm := &insight.DASUserPermission{
		Username:       req.Username,
		PermissionType: insight.PermissionType(req.PermissionType),
		PermissionID:   req.PermissionID,
		InstanceID:     req.InstanceID,
		Schema:         req.Schema,
		Table:          req.Table,
	}
	if err := service.InsightServiceApp.CreateUserPermission(c.Request.Context(), perm); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, perm)
}

// DeleteUserPermission 删除用户权限
func (h *DASHandler) DeleteUserPermission(c *gin.Context) {
	id := c.Param("id")
	var permID uint
	if _, err := fmt.Sscanf(id, "%d", &permID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}
	if err := service.InsightServiceApp.DeleteUserPermission(c.Request.Context(), permID); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, nil)
}

// GrantSchemaPermission 授权Schema权限
// @Summary 授权Schema权限
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body GrantSchemaPermissionRequest true "授权信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/permissions/schema [post]
func (h *DASHandler) GrantSchemaPermission(c *gin.Context) {
	var req GrantSchemaPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 验证实例是否为"查询"类型，防止越权
	if err := h.validateInstanceForPermission(c.Request.Context(), req.InstanceID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	instanceUUID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	perm := &insight.DASUserSchemaPermission{
		Username:   req.Username,
		InstanceID: instanceUUID,
		Schema:     req.Schema,
	}

	if err := service.InsightServiceApp.CreateSchemaPermission(c.Request.Context(), perm); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, perm)
}

// RevokeSchemaPermission 撤销Schema权限
// @Summary 撤销Schema权限
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/permissions/schema/{id} [delete]
func (h *DASHandler) RevokeSchemaPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteSchemaPermission(c.Request.Context(), uint(id)); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// GrantTablePermission 授权Table权限
// @Summary 授权Table权限
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body GrantTablePermissionRequest true "授权信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/permissions/table [post]
func (h *DASHandler) GrantTablePermission(c *gin.Context) {
	var req GrantTablePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 验证实例是否为"查询"类型，防止越权
	if err := h.validateInstanceForPermission(c.Request.Context(), req.InstanceID); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	instanceUUID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 设置默认规则为 allow
	rule := insight.RuleAllow
	if req.Rule == "deny" {
		rule = insight.RuleDeny
	}

	perm := &insight.DASUserTablePermission{
		Username:   req.Username,
		InstanceID: instanceUUID,
		Schema:     req.Schema,
		Table:      req.Table,
		Rule:       rule,
	}

	if err := service.InsightServiceApp.CreateTablePermission(c.Request.Context(), perm); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, perm)
}

// RevokeTablePermission 撤销Table权限
// @Summary 撤销Table权限
// @Tags DAS
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/das/permissions/table/{id} [delete]
func (h *DASHandler) RevokeTablePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteTablePermission(c.Request.Context(), uint(id)); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// getUsernameByID 通过用户ID获取用户名
func (h *DASHandler) getUsernameByID(c *gin.Context, userId uint) (string, error) {
	user, err := service.AdminServiceApp.GetAdminUser(c, userId)
	if err != nil {
		return "", fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user.Username, nil
}
