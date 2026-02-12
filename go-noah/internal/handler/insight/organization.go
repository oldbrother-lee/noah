package insight

import (
	"encoding/json"
	"go-noah/api"
	"go-noah/internal/handler"
	"go-noah/internal/model/insight"
	"go-noah/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// OrganizationHandlerApp 全局 Handler 实例
var OrganizationHandlerApp = new(OrganizationHandler)

// OrganizationHandler 组织管理 Handler
type OrganizationHandler struct{}

// OrganizationTreeNode 组织树节点
type OrganizationTreeNode struct {
	ID        uint64                  `json:"id"`
	Name      string                  `json:"name"`
	ParentID  uint64                  `json:"parent_id"`
	Key       string                  `json:"key"`
	Level     uint64                  `json:"level"`
	Path      []string                `json:"path"`
	Children  []*OrganizationTreeNode `json:"children"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
}

// GetOrganizations 获取组织列表
// @Summary 获取组织列表
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations [get]
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	orgs, err := service.InsightServiceApp.GetOrganizations(c.Request.Context())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}
	api.HandleSuccess(c, orgs)
}

// GetOrganizationTree 获取组织树
// @Summary 获取组织树
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations/tree [get]
func (h *OrganizationHandler) GetOrganizationTree(c *gin.Context) {
	orgs, err := service.InsightServiceApp.GetOrganizations(c.Request.Context())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// 构建树形结构
	tree := h.buildTree(orgs)
	api.HandleSuccess(c, tree)
}

// buildTree 构建组织树
func (h *OrganizationHandler) buildTree(orgs []insight.Organization) []*OrganizationTreeNode {
	nodeMap := make(map[uint64]*OrganizationTreeNode)
	var roots []*OrganizationTreeNode

	// 创建节点
	for _, org := range orgs {
		var path []string
		if org.Path != nil {
			_ = json.Unmarshal(org.Path, &path)
		}

		node := &OrganizationTreeNode{
			ID:        org.ID,
			Name:      org.Name,
			ParentID:  org.ParentID,
			Key:       org.Key,
			Level:     org.Level,
			Path:      path,
			Children:  []*OrganizationTreeNode{},
			CreatedAt: org.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: org.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		nodeMap[org.ID] = node
	}

	// 构建树
	for _, node := range nodeMap {
		if node.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint64 `json:"parent_id"`
}

// CreateOrganization 创建组织
// @Summary 创建组织
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body CreateOrganizationRequest true "组织信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 计算 key 和 level
	var key string
	var level uint64 = 1
	var path []string

	if req.ParentID > 0 {
		parent, err := service.InsightServiceApp.GetOrganizationByID(c.Request.Context(), req.ParentID)
		if err != nil {
			api.HandleError(c, http.StatusBadRequest, err, nil)
			return
		}
		level = parent.Level + 1

		// 构建 path
		if parent.Path != nil {
			_ = json.Unmarshal(parent.Path, &path)
		}
		path = append(path, parent.Name)

		// 构建 key
		if parent.Key != "" {
			key = parent.Key + "/" + req.Name
		} else {
			key = parent.Name + "/" + req.Name
		}
	} else {
		key = req.Name
		path = []string{}
	}

	pathJSON, _ := json.Marshal(path)

	org := &insight.Organization{
		Name:      req.Name,
		ParentID:  req.ParentID,
		Key:       key,
		Level:     level,
		Path:      pathJSON,
		Creator:   username,
		Updater:   username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := service.InsightServiceApp.CreateOrganization(c.Request.Context(), org); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, org)
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	ID   uint64 `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// UpdateOrganization 更新组织
// @Summary 更新组织
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body UpdateOrganizationRequest true "组织信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 获取当前用户
	userId := handler.GetUserIdFromCtx(c)
	username := ""
	if userId > 0 {
		user, err := service.AdminServiceApp.GetAdminUser(c, userId)
		if err == nil {
			username = user.Username
		}
	}

	// 获取现有组织
	org, err := service.InsightServiceApp.GetOrganizationByID(c.Request.Context(), req.ID)
	if err != nil {
		api.HandleError(c, http.StatusNotFound, err, nil)
		return
	}

	// 更新 key（如果名称变了）
	if org.Name != req.Name {
		newKey := strings.Replace(org.Key, org.Name, req.Name, -1)
		org.Key = newKey
	}

	org.Name = req.Name
	org.Updater = username
	org.UpdatedAt = time.Now()

	if err := service.InsightServiceApp.UpdateOrganization(c.Request.Context(), org); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, org)
}

// DeleteOrganization 删除组织
// @Summary 删除组织
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "组织ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	// 检查是否有子节点
	orgs, err := service.InsightServiceApp.GetOrganizations(c.Request.Context())
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	for _, org := range orgs {
		if org.ParentID == id {
			api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, "该组织下有子节点，不能删除")
			return
		}
	}

	if err := service.InsightServiceApp.DeleteOrganization(c.Request.Context(), id); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

// GetOrganizationUsers 获取组织下的用户
// @Summary 获取组织下的用户
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param key query string true "组织Key"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations/users [get]
func (h *OrganizationHandler) GetOrganizationUsers(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	users, err := service.InsightServiceApp.GetOrganizationUsers(c.Request.Context(), key)
	if err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, users)
}

// BindUserRequest 绑定用户请求
type BindUserRequest struct {
	UID             uint64 `json:"uid" binding:"required"`
	OrganizationKey string `json:"organization_key" binding:"required"`
}

// BindUser 绑定用户到组织
// @Summary 绑定用户到组织
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body BindUserRequest true "绑定信息"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations/users [post]
func (h *OrganizationHandler) BindUser(c *gin.Context) {
	var req BindUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleError(c, http.StatusBadRequest, err, nil)
		return
	}

	// 先解绑原有组织
	_ = service.InsightServiceApp.UnbindOrganizationUser(c.Request.Context(), req.UID)

	// 绑定到新组织
	ou := &insight.OrganizationUser{
		UID:             req.UID,
		OrganizationKey: req.OrganizationKey,
	}

	if err := service.InsightServiceApp.BindOrganizationUser(c.Request.Context(), ou); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, ou)
}

// UnbindUser 解绑用户
// @Summary 解绑用户
// @Tags 组织管理
// @Security Bearer
// @Accept json
// @Produce json
// @Param uid path int true "用户ID"
// @Success 200 {object} api.Response
// @Router /api/v1/insight/organizations/users/{uid} [delete]
func (h *OrganizationHandler) UnbindUser(c *gin.Context) {
	uidStr := c.Param("uid")
	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		api.HandleError(c, http.StatusBadRequest, api.ErrBadRequest, nil)
		return
	}

	if err := service.InsightServiceApp.UnbindOrganizationUser(c.Request.Context(), uid); err != nil {
		api.HandleError(c, http.StatusInternalServerError, err, nil)
		return
	}

	api.HandleSuccess(c, nil)
}

