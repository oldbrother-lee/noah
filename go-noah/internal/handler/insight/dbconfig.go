package insight

import (
	"encoding/json"
	"go-noah/api"
	"go-noah/internal/model/insight"
	"go-noah/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// DBConfigHandlerApp 全局 Handler 实例
var DBConfigHandlerApp = new(DBConfigHandler)

// DBConfigHandler 数据库配置Handler
type DBConfigHandler struct{}

// GetDBConfigsRequest 获取数据库配置列表请求
type GetDBConfigsRequest struct {
	UseType     string `form:"use_type"`     // 用途：查询/工单
	Environment int    `form:"environment"` // 环境ID
	ID          int    `form:"id"`           // 环境ID（兼容前端传递的 id 参数）
	DbType      string `form:"db_type"`      // 数据库类型：MySQL/TiDB
}

// GetDBConfigs 获取数据库配置列表
// @Summary 获取数据库配置列表
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param use_type query string false "用途：查询/工单"
// @Param environment query int false "环境ID"
// @Param id query int false "环境ID（兼容参数）"
// @Param db_type query string false "数据库类型：MySQL/TiDB"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs [get]
func (h *DBConfigHandler) GetDBConfigs(c *gin.Context) {
	var req GetDBConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 兼容前端传递的 id 参数，如果 environment 为空但 id 有值，使用 id
	environment := req.Environment
	if environment == 0 && req.ID > 0 {
		environment = req.ID
	}

	// 获取配置列表
	configs, err := service.InsightServiceApp.GetDBConfigs(c.Request.Context(), insight.UseType(req.UseType), environment)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 如果指定了 db_type，进一步过滤
	if req.DbType != "" {
		filteredConfigs := make([]insight.DBConfig, 0)
		for _, config := range configs {
			if string(config.DbType) == req.DbType {
				filteredConfigs = append(filteredConfigs, config)
			}
		}
		configs = filteredConfigs
	}

	// 隐藏密码
	for i := range configs {
		configs[i].Password = "******"
	}

	api.HandleSuccess(c, configs)
}

// GetDBConfig 获取单个数据库配置
// @Summary 获取单个数据库配置
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs/{instance_id} [get]
func (h *DBConfigHandler) GetDBConfig(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	config, err := service.InsightServiceApp.GetDBConfigByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 隐藏密码
	config.Password = "******"

	api.HandleSuccess(c, config)
}

// CreateDBConfigRequest 创建数据库配置请求
type CreateDBConfigRequest struct {
	Hostname        string                 `json:"hostname" binding:"required"`
	Port            int                    `json:"port" binding:"required"`
	UserName        string                 `json:"user_name" binding:"required"`
	Password        string                 `json:"password" binding:"required"`
	UseType         string                 `json:"use_type" binding:"required"` // 查询/工单
	DbType          string                 `json:"db_type" binding:"required"`  // MySQL/TiDB/ClickHouse
	Environment     int                    `json:"environment"`
	OrganizationKey string                 `json:"organization_key"`
	Remark          string                 `json:"remark"`
	InspectParams   map[string]interface{} `json:"inspect_params,omitempty"` // 审核参数（JSON对象）
}

// CreateDBConfig 创建数据库配置
// @Summary 创建数据库配置
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body CreateDBConfigRequest true "数据库配置信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs [post]
func (h *DBConfigHandler) CreateDBConfig(c *gin.Context) {
	var req CreateDBConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 显式生成 instance_id，与老系统保持一致
	instanceID, err := uuid.NewUUID()
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	config := &insight.DBConfig{
		InstanceID:      instanceID,
		Hostname:        req.Hostname,
		Port:            req.Port,
		UserName:        req.UserName,
		Password:        req.Password,
		UseType:         insight.UseType(req.UseType),
		DbType:          insight.DbType(req.DbType),
		Environment:     req.Environment,
		OrganizationKey: req.OrganizationKey,
		Remark:          req.Remark,
	}
	if req.InspectParams != nil {
		if bs, err := json.Marshal(req.InspectParams); err == nil {
			config.InspectParams = datatypes.JSON(bs)
		} else {
			api.HandleError(c, http.StatusBadRequest, err, nil)
			return
		}
	}

	if err := service.InsightServiceApp.CreateDBConfig(c.Request.Context(), config); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 隐藏密码
	config.Password = "******"
	api.HandleSuccess(c, config)
}

// UpdateDBConfigRequest 更新数据库配置请求
type UpdateDBConfigRequest struct {
	Hostname        *string                `json:"hostname,omitempty"`
	Port            *int                   `json:"port,omitempty"`
	UserName        *string                `json:"user_name,omitempty"`
	Password        *string                `json:"password,omitempty"`
	UseType         *string                `json:"use_type,omitempty"`
	DbType          *string                `json:"db_type,omitempty"`
	Environment     *int                   `json:"environment,omitempty"`
	OrganizationKey *string                `json:"organization_key,omitempty"`
	Remark          *string                `json:"remark,omitempty"`
	InspectParams   *map[string]interface{} `json:"inspect_params,omitempty"` // 审核参数（JSON对象）
}

// UpdateDBConfig 更新数据库配置
// @Summary 更新数据库配置
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body UpdateDBConfigRequest true "数据库配置信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs/{id} [put]
func (h *DBConfigHandler) UpdateDBConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	var req UpdateDBConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	updates := make(map[string]interface{})
	if req.Hostname != nil {
		updates["hostname"] = *req.Hostname
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.UserName != nil {
		updates["user_name"] = *req.UserName
	}
	if req.Password != nil {
		// 编辑时允许不更新密码：空字符串则跳过
		if *req.Password != "" {
			updates["password"] = *req.Password
		}
	}
	if req.UseType != nil {
		updates["use_type"] = *req.UseType
	}
	if req.DbType != nil {
		updates["db_type"] = *req.DbType
	}
	if req.Environment != nil {
		updates["environment"] = *req.Environment
	}
	if req.OrganizationKey != nil {
		updates["organization_key"] = *req.OrganizationKey
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if req.InspectParams != nil {
		bs, err := json.Marshal(*req.InspectParams)
		if err != nil {
			api.HandleError(c, http.StatusBadRequest, err, nil)
			return
		}
		updates["inspect_params"] = datatypes.JSON(bs)
	}

	if len(updates) == 0 {
		api.HandleSuccess(c, nil)
		return
	}

	if err := service.InsightServiceApp.UpdateDBConfigFields(c.Request.Context(), uint(id), updates); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, nil)
}

// DeleteDBConfig 删除数据库配置
// @Summary 删除数据库配置
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs/{id} [delete]
func (h *DBConfigHandler) DeleteDBConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.DeleteDBConfig(c.Request.Context(), uint(id)); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, nil)
}

// GetSchemas 获取实例下的Schema列表
// @Summary 获取Schema列表
// @Tags 数据库配置
// @Security Bearer
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/dbconfigs/{instance_id}/schemas [get]
func (h *DBConfigHandler) GetSchemas(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	schemas, err := service.InsightServiceApp.GetSchemasByInstanceID(c.Request.Context(), instanceID)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, schemas)
}

