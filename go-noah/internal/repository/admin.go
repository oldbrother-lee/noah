package repository

import (
	"context"
	"fmt"
	"go-noah/api"
	"go-noah/internal/model"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
)

// AdminRepository 管理员数据访问层（简化版：直接使用结构体，不定义接口）
type AdminRepository struct {
	*Repository
}

func NewAdminRepository(repository *Repository) *AdminRepository {
	return &AdminRepository{
		Repository: repository,
	}
}

func (r *AdminRepository) CasbinRoleDelete(ctx context.Context, role string) error {
	_, err := r.e.DeleteRole(role)
	return err
}

func (r *AdminRepository) GetRole(ctx context.Context, id uint) (model.Role, error) {
	m := model.Role{}
	return m, r.DB(ctx).Where("id = ?", id).First(&m).Error
}
func (r *AdminRepository) GetRoleBySid(ctx context.Context, sid string) (model.Role, error) {
	m := model.Role{}
	return m, r.DB(ctx).Where("sid = ?", sid).First(&m).Error
}

func (r *AdminRepository) DeleteUserRoles(ctx context.Context, uid uint) error {
	_, err := r.e.DeleteRolesForUser(convertor.ToString(uid))
	return err
}
func (r *AdminRepository) UpdateUserRoles(ctx context.Context, uid uint, roles []string) error {
	if len(roles) == 0 {
		_, err := r.e.DeleteRolesForUser(convertor.ToString(uid))
		return err
	}
	old, err := r.e.GetRolesForUser(convertor.ToString(uid))
	if err != nil {
		return err
	}
	oldMap := make(map[string]struct{})
	newMap := make(map[string]struct{})
	for _, v := range old {
		oldMap[v] = struct{}{}
	}
	for _, v := range roles {
		newMap[v] = struct{}{}
	}
	addRoles := make([]string, 0)
	delRoles := make([]string, 0)

	for key, _ := range oldMap {
		if _, exists := newMap[key]; !exists {
			delRoles = append(delRoles, key)
		}
	}
	for key, _ := range newMap {
		if _, exists := oldMap[key]; !exists {
			addRoles = append(addRoles, key)
		}
	}
	if len(addRoles) == 0 && len(delRoles) == 0 {
		return nil
	}
	for _, role := range delRoles {
		if _, err := r.e.DeleteRoleForUser(convertor.ToString(uid), role); err != nil {
			r.logger.WithContext(ctx).Error("DeleteRoleForUser error", zap.Error(err))
			return err
		}
	}

	if len(addRoles) > 0 {
		_, err = r.e.AddRolesForUser(convertor.ToString(uid), addRoles)
		return err
	}
	return nil
}

func (r *AdminRepository) GetAdminUserByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	m := model.AdminUser{}
	return m, r.DB(ctx).Where("username = ?", username).First(&m).Error
}

func (r *AdminRepository) GetAdminUsers(ctx context.Context, req *api.GetAdminUsersRequest) ([]model.AdminUser, int64, error) {
	var list []model.AdminUser
	var total int64
	scope := r.DB(ctx).Model(&model.AdminUser{})
	if req.Username != "" {
		scope = scope.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Nickname != "" {
		scope = scope.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Email != "" {
		scope = scope.Where("email LIKE ?", "%"+req.Email+"%")
	}
	if req.Phone != "" {
		scope = scope.Where("phone LIKE ?", "%"+req.Phone+"%")
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func (r *AdminRepository) GetAdminUser(ctx context.Context, uid uint) (model.AdminUser, error) {
	m := model.AdminUser{}
	return m, r.DB(ctx).Where("id = ?", uid).First(&m).Error
}

func (r *AdminRepository) AdminUserUpdate(ctx context.Context, m *model.AdminUser) error {
	return r.DB(ctx).Where("id = ?", m.ID).Updates(m).Error
}

func (r *AdminRepository) AdminUserCreate(ctx context.Context, m *model.AdminUser) error {
	return r.DB(ctx).Create(m).Error
}

func (r *AdminRepository) AdminUserDelete(ctx context.Context, id uint) error {
	return r.DB(ctx).Where("id = ?", id).Delete(&model.AdminUser{}).Error
}

func (r *AdminRepository) UpdateRolePermission(ctx context.Context, role string, newPermSet map[string]struct{}) error {
	if len(newPermSet) == 0 {
		return nil
	}
	// 获取当前角色的所有权限
	oldPermissions, err := r.e.GetPermissionsForUser(role)
	if err != nil {
		return err
	}

	// 将旧权限转换为 map 方便查找
	oldPermSet := make(map[string]struct{})
	for _, perm := range oldPermissions {
		if len(perm) == 3 {
			oldPermSet[strings.Join([]string{perm[1], perm[2]}, model.PermSep)] = struct{}{}
		}
	}

	// 找出需要删除的权限
	var removePermissions [][]string
	for key, _ := range oldPermSet {
		if _, exists := newPermSet[key]; !exists {
			removePermissions = append(removePermissions, strings.Split(key, model.PermSep))
		}
	}

	// 找出需要添加的权限
	var addPermissions [][]string
	for key, _ := range newPermSet {
		if _, exists := oldPermSet[key]; !exists {
			addPermissions = append(addPermissions, strings.Split(key, model.PermSep))
		}

	}

	// 先移除多余的权限（使用 DeletePermissionForUser 逐条删除）
	for _, perm := range removePermissions {
		_, err := r.e.DeletePermissionForUser(role, perm...)
		if err != nil {
			return fmt.Errorf("移除权限失败: %v", err)
		}
	}

	// 再添加新的权限
	if len(addPermissions) > 0 {
		_, err = r.e.AddPermissionsForUser(role, addPermissions...)
		if err != nil {
			return fmt.Errorf("添加新权限失败: %v", err)
		}
	}

	return nil
}

func (r *AdminRepository) GetApiGroups(ctx context.Context) ([]string, error) {
	res := make([]string, 0)
	if err := r.DB(ctx).Model(&model.Api{}).Group("`group`").Pluck("`group`", &res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *AdminRepository) GetApis(ctx context.Context, req *api.GetApisRequest) ([]model.Api, int64, error) {
	var list []model.Api
	var total int64
	scope := r.DB(ctx).Model(&model.Api{})
	if req.Name != "" {
		scope = scope.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Group != "" {
		scope = scope.Where("`group` LIKE ?", "%"+req.Group+"%")
	}
	if req.Path != "" {
		scope = scope.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Method != "" {
		scope = scope.Where("method = ?", req.Method)
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order("`group` ASC").Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func (r *AdminRepository) ApiUpdate(ctx context.Context, m *model.Api) error {
	return r.DB(ctx).Where("id = ?", m.ID).Updates(m).Error
}

func (r *AdminRepository) ApiCreate(ctx context.Context, m *model.Api) error {
	return r.DB(ctx).Create(m).Error
}

func (r *AdminRepository) ApiDelete(ctx context.Context, id uint) error {
	return r.DB(ctx).Where("id = ?", id).Delete(&model.Api{}).Error
}

// GetApiByID 按 ID 获取单条 API
func (r *AdminRepository) GetApiByID(ctx context.Context, id uint) (*model.Api, error) {
	var m model.Api
	err := r.DB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ApiDeleteByPathMethod 按 path+method 删除 API（用于确认同步时删除）
func (r *AdminRepository) ApiDeleteByPathMethod(ctx context.Context, path, method string) error {
	return r.DB(ctx).Where("path = ? AND method = ?", path, method).Delete(&model.Api{}).Error
}

// ApisCreateBatch 批量创建 API
func (r *AdminRepository) ApisCreateBatch(ctx context.Context, list []*model.Api) error {
	if len(list) == 0 {
		return nil
	}
	return r.DB(ctx).Create(&list).Error
}

// GetIgnoreApis 获取所有忽略的 API 列表
func (r *AdminRepository) GetIgnoreApis(ctx context.Context) ([]model.ApiIgnore, error) {
	var list []model.ApiIgnore
	err := r.DB(ctx).Find(&list).Error
	return list, err
}

// IgnoreApiCreate 加入忽略列表
func (r *AdminRepository) IgnoreApiCreate(ctx context.Context, path, method string) error {
	return r.DB(ctx).Create(&model.ApiIgnore{Path: path, Method: method}).Error
}

// IgnoreApiDelete 从忽略列表移除
func (r *AdminRepository) IgnoreApiDelete(ctx context.Context, path, method string) error {
	return r.DB(ctx).Where("path = ? AND method = ?", path, method).Delete(&model.ApiIgnore{}).Error
}

// ClearCasbinForApi 清除该 API 在 Casbin 中的所有策略（删除 API 时调用）
func (r *AdminRepository) ClearCasbinForApi(path, method string) error {
	obj := model.ApiResourcePrefix + path
	_, err := r.e.RemoveFilteredPolicy(1, obj, method)
	return err
}

// CheckApiExists 检查 API 是否已存在（基于 path + method）
func (r *AdminRepository) CheckApiExists(ctx context.Context, path, method string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Api{}).Where("path = ? AND method = ?", path, method).Count(&count).Error
	return count > 0, err
}

func (r *AdminRepository) GetUserPermissions(ctx context.Context, uid uint) ([][]string, error) {
	return r.e.GetImplicitPermissionsForUser(convertor.ToString(uid))

}
func (r *AdminRepository) GetRolePermissions(ctx context.Context, role string) ([][]string, error) {
	return r.e.GetPermissionsForUser(role)
}
func (r *AdminRepository) GetUserRoles(ctx context.Context, uid uint) ([]string, error) {
	return r.e.GetRolesForUser(convertor.ToString(uid))
}
func (r *AdminRepository) MenuUpdate(ctx context.Context, m *model.Menu) error {
	return r.DB(ctx).Where("id = ?", m.ID).Updates(m).Error
}

func (r *AdminRepository) MenuCreate(ctx context.Context, m *model.Menu) error {
	return r.DB(ctx).Create(m).Error
}

func (r *AdminRepository) MenuDelete(ctx context.Context, id uint) error {
	return r.DB(ctx).Where("id = ?", id).Delete(&model.Menu{}).Error
}

func (r *AdminRepository) GetMenuList(ctx context.Context) ([]model.Menu, error) {
	var menuList []model.Menu
	if err := r.DB(ctx).Order("weight DESC").Find(&menuList).Error; err != nil {
		return nil, err
	}
	return menuList, nil
}

func (r *AdminRepository) RoleUpdate(ctx context.Context, m *model.Role) error {
	return r.DB(ctx).Model(&model.Role{}).Where("id = ?", m.ID).UpdateColumn("name", m.Name).Error
}

func (r *AdminRepository) RoleCreate(ctx context.Context, m *model.Role) error {
	return r.DB(ctx).Create(m).Error
}

func (r *AdminRepository) RoleDelete(ctx context.Context, id uint) error {
	return r.DB(ctx).Where("id = ?", id).Delete(&model.Role{}).Error
}

func (r *AdminRepository) GetRoles(ctx context.Context, req *api.GetRoleListRequest) ([]model.Role, int64, error) {
	var list []model.Role
	var total int64
	scope := r.DB(ctx).Model(&model.Role{})
	if req.Name != "" {
		scope = scope.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Sid != "" {
		scope = scope.Where("sid = ?", req.Sid)
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}
