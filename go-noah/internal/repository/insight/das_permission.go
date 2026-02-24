package insight

import (
	"context"
	"go-noah/internal/model/insight"
	"github.com/duke-git/lancet/v2/convertor"
	"gorm.io/gorm"
)

// ============ 权限模板管理 ============

// GetPermissionTemplates 获取权限模板列表
func (r *InsightRepository) GetPermissionTemplates(ctx context.Context) ([]insight.DASPermissionTemplate, error) {
	var templates []insight.DASPermissionTemplate
	if err := r.DB(ctx).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetPermissionTemplate 获取权限模板详情
func (r *InsightRepository) GetPermissionTemplate(ctx context.Context, id uint) (*insight.DASPermissionTemplate, error) {
	var template insight.DASPermissionTemplate
	if err := r.DB(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// CreatePermissionTemplate 创建权限模板
func (r *InsightRepository) CreatePermissionTemplate(ctx context.Context, template *insight.DASPermissionTemplate) error {
	return r.DB(ctx).Create(template).Error
}

// UpdatePermissionTemplate 更新权限模板
func (r *InsightRepository) UpdatePermissionTemplate(ctx context.Context, template *insight.DASPermissionTemplate) error {
	return r.DB(ctx).Save(template).Error
}

// DeletePermissionTemplate 删除权限模板（软删除）
func (r *InsightRepository) DeletePermissionTemplate(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DASPermissionTemplate{}, id).Error
}

// ============ 角色权限管理 ============

// GetRolePermissions 获取角色权限列表
func (r *InsightRepository) GetRolePermissions(ctx context.Context, role string) ([]insight.DASRolePermission, error) {
	var perms []insight.DASRolePermission
	if err := r.DB(ctx).Where("role = ?", role).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// CreateRolePermission 创建角色权限
func (r *InsightRepository) CreateRolePermission(ctx context.Context, perm *insight.DASRolePermission) error {
	return r.DB(ctx).Create(perm).Error
}

// DeleteRolePermission 删除角色权限（软删除）
func (r *InsightRepository) DeleteRolePermission(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DASRolePermission{}, id).Error
}

// BatchCreateRolePermissions 批量创建角色权限
func (r *InsightRepository) BatchCreateRolePermissions(ctx context.Context, perms []insight.DASRolePermission) error {
	if len(perms) == 0 {
		return nil
	}
	return r.DB(ctx).Create(&perms).Error
}

// ============ 用户权限管理（与角色权限同构：object/template，无 rule）============

// GetUserPermissions 获取用户权限列表（新表 das_user_permissions）
func (r *InsightRepository) GetUserPermissions(ctx context.Context, username string) ([]insight.DASUserPermission, error) {
	var perms []insight.DASUserPermission
	if err := r.DB(ctx).Where("username = ?", username).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// CreateUserPermission 创建用户权限
func (r *InsightRepository) CreateUserPermission(ctx context.Context, perm *insight.DASUserPermission) error {
	return r.DB(ctx).Create(perm).Error
}

// DeleteUserPermission 删除用户权限（软删除）
func (r *InsightRepository) DeleteUserPermission(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DASUserPermission{}, id).Error
}

// ExpandUserPermissions 展开用户权限（与角色权限相同逻辑：object 即一条，template 展开）
func (r *InsightRepository) ExpandUserPermissions(ctx context.Context, username string) ([]insight.PermissionObject, error) {
	perms, err := r.GetUserPermissions(ctx, username)
	if err != nil {
		return nil, err
	}
	var result []insight.PermissionObject
	for _, perm := range perms {
		switch perm.PermissionType {
		case insight.PermissionTypeObject:
			result = append(result, insight.PermissionObject{
				InstanceID: perm.InstanceID,
				Schema:     perm.Schema,
				Table:      perm.Table,
			})
		case insight.PermissionTypeTemplate:
			template, err := r.GetPermissionTemplate(ctx, perm.PermissionID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				return nil, err
			}
			result = append(result, template.Permissions...)
		}
	}
	return result, nil
}

// ============ 权限展开和查询 ============

// ExpandRolePermissions 展开角色权限（将模板/组展开为具体权限对象）
func (r *InsightRepository) ExpandRolePermissions(ctx context.Context, role string) ([]insight.PermissionObject, error) {
	// 1. 获取角色权限
	rolePerms, err := r.GetRolePermissions(ctx, role)
	if err != nil {
		return nil, err
	}

	var result []insight.PermissionObject

	// 2. 展开权限
	for _, perm := range rolePerms {
		switch perm.PermissionType {
		case insight.PermissionTypeObject:
			// 直接权限对象
			result = append(result, insight.PermissionObject{
				InstanceID: perm.InstanceID,
				Schema:     perm.Schema,
				Table:      perm.Table,
			})

		case insight.PermissionTypeTemplate:
			// 展开权限模板
			template, err := r.GetPermissionTemplate(ctx, perm.PermissionID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue // 模板不存在，跳过
				}
				return nil, err
			}
			result = append(result, template.Permissions...)
		}
	}

	return result, nil
}

// GetUserEffectivePermissions 获取用户实际生效的权限（合并角色权限和直接权限）
func (r *InsightRepository) GetUserEffectivePermissions(ctx context.Context, username string) ([]insight.PermissionObject, error) {
	permissionMap := make(map[string]insight.PermissionObject) // 使用 map 去重

	// 1. 获取角色权限（通过角色展开）
	// 注意：Casbin 使用用户ID（字符串）作为 key，而不是用户名
	// 必须先通过用户名获取用户ID，然后使用用户ID查询角色
	enforcer := r.Enforcer()
	
	// 查询用户ID
	var userID uint
	err := r.DB(ctx).Table("admin_users").
		Select("id").
		Where("username = ?", username).
		Scan(&userID).Error
	if err != nil {
		// 用户不存在，返回空权限列表
		return []insight.PermissionObject{}, nil
	}
	
	if userID == 0 {
		// 用户ID无效，返回空权限列表
		return []insight.PermissionObject{}, nil
	}
	
	// 使用用户ID（字符串）查询角色
	userIDStr := convertor.ToString(userID)
	roles, err := enforcer.GetRolesForUser(userIDStr)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		perms, err := r.ExpandRolePermissions(ctx, role)
		if err != nil {
			return nil, err
		}
		for _, perm := range perms {
			key := perm.InstanceID + ":" + perm.Schema + ":" + perm.Table
			permissionMap[key] = perm
		}
	}

	// 2. 用户权限（与角色同构：object/template，无 rule）
	userPerms, err := r.ExpandUserPermissions(ctx, username)
	if err != nil {
		return nil, err
	}
	for _, perm := range userPerms {
		key := perm.InstanceID + ":" + perm.Schema + ":" + perm.Table
		permissionMap[key] = perm
	}

	// 3. 兼容旧数据：用户直接库权限（das_user_schema_permissions）
	schemaPerms, err := r.GetUserSchemaPermissions(ctx, username)
	if err != nil {
		return nil, err
	}

	// 将直接权限转换为 PermissionObject
	for _, perm := range schemaPerms {
		key := perm.InstanceID.String() + ":" + perm.Schema + ":"
		permissionMap[key] = insight.PermissionObject{
			InstanceID: perm.InstanceID.String(),
			Schema:     perm.Schema,
			Table:      "", // 库权限，表为空
		}
	}

	// 4. 转换为数组
	result := make([]insight.PermissionObject, 0, len(permissionMap))
	for _, perm := range permissionMap {
		result = append(result, perm)
	}

	return result, nil
}

