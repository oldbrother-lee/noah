package insight

import (
	"context"
	"go-noah/internal/model/insight"
	"strings"
)

// ============ DAS 用户权限管理 ============

// GetUserAuthorizedSchemas 获取用户授权的所有 schemas（关联查询）
// 现在基于角色权限，而不是直接权限
func (r *InsightRepository) GetUserAuthorizedSchemas(ctx context.Context, username string) ([]insight.UserAuthorizedSchema, error) {
	// 1. 获取用户实际生效的权限（基于角色）
	effectivePerms, err := r.GetUserEffectivePermissions(ctx, username)
	if err != nil {
		return nil, err
	}

	// 2. 如果没有权限，返回空列表
	if len(effectivePerms) == 0 {
		return []insight.UserAuthorizedSchema{}, nil
	}

	// 3. 收集所有唯一的 instance_id + schema 组合
	// 包括库权限和表权限（只要有表权限，也应该返回对应的库）
	permMap := make(map[string]bool)
	for _, perm := range effectivePerms {
		// 处理所有权限（包括库权限和表权限）
		// 只要有权限（无论是库权限还是表权限），都应该返回对应的库
		key := perm.InstanceID + ":" + perm.Schema
		permMap[key] = true
	}

	// 4. 构建查询条件
	if len(permMap) == 0 {
		return []insight.UserAuthorizedSchema{}, nil
	}

	// 5. 查询数据库配置和 schema 信息
	var results []insight.UserAuthorizedSchema
	query := r.DB(ctx).Table("db_configs b").
		Select("b.instance_id, b.db_type, c.`schema`, b.hostname, b.port, c.is_deleted, b.remark").
		Joins("JOIN db_schemas c ON b.instance_id = c.instance_id").
		Where("c.is_deleted = ?", false)

	// 6. 构建 WHERE 条件：匹配权限中的 instance_id 和 schema
	var conditions []string
	var args []interface{}
	for key := range permMap {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			instanceID := parts[0]
			schema := parts[1]
			// instance_id 在数据库中存储为 UUID 类型，需要转换为字符串进行比较
			// 使用 BINARY 确保精确匹配（区分大小写）
			conditions = append(conditions, "(BINARY CAST(b.instance_id AS CHAR(36)) = ? AND c.`schema` = ?)")
			args = append(args, instanceID, schema)
		}
	}

	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	err = query.Group("b.instance_id, b.db_type, c.`schema`, b.hostname, b.port, c.is_deleted, b.remark").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUserSchemaPermissions 获取用户的库权限
func (r *InsightRepository) GetUserSchemaPermissions(ctx context.Context, username string) ([]insight.DASUserSchemaPermission, error) {
	var perms []insight.DASUserSchemaPermission
	if err := r.DB(ctx).Where("username = ?", username).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// GetSchemaPermissionsByInstance 获取实例下的所有库权限
func (r *InsightRepository) GetSchemaPermissionsByInstance(ctx context.Context, instanceID string) ([]insight.DASUserSchemaPermission, error) {
	var perms []insight.DASUserSchemaPermission
	if err := r.DB(ctx).Where("instance_id = ?", instanceID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// CreateSchemaPermission 创建库权限
func (r *InsightRepository) CreateSchemaPermission(ctx context.Context, perm *insight.DASUserSchemaPermission) error {
	return r.DB(ctx).Create(perm).Error
}

// DeleteSchemaPermission 删除库权限
func (r *InsightRepository) DeleteSchemaPermission(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DASUserSchemaPermission{}, id).Error
}

// GetUserTablePermissions 获取用户的表权限
func (r *InsightRepository) GetUserTablePermissions(ctx context.Context, username string) ([]insight.DASUserTablePermission, error) {
	var perms []insight.DASUserTablePermission
	if err := r.DB(ctx).Where("username = ?", username).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// GetTablePermissionsByInstance 获取实例下的所有表权限
func (r *InsightRepository) GetTablePermissionsByInstance(ctx context.Context, instanceID string) ([]insight.DASUserTablePermission, error) {
	var perms []insight.DASUserTablePermission
	if err := r.DB(ctx).Where("instance_id = ?", instanceID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// CreateTablePermission 创建表权限
func (r *InsightRepository) CreateTablePermission(ctx context.Context, perm *insight.DASUserTablePermission) error {
	return r.DB(ctx).Create(perm).Error
}

// DeleteTablePermission 删除表权限
func (r *InsightRepository) DeleteTablePermission(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DASUserTablePermission{}, id).Error
}

// ============ DAS 允许的操作管理 ============

// GetAllowedOperations 获取所有允许的操作
func (r *InsightRepository) GetAllowedOperations(ctx context.Context) ([]insight.DASAllowedOperation, error) {
	var ops []insight.DASAllowedOperation
	if err := r.DB(ctx).Find(&ops).Error; err != nil {
		return nil, err
	}
	return ops, nil
}

// GetEnabledOperations 获取已启用的操作
func (r *InsightRepository) GetEnabledOperations(ctx context.Context) ([]insight.DASAllowedOperation, error) {
	var ops []insight.DASAllowedOperation
	if err := r.DB(ctx).Where("is_enable = ?", true).Find(&ops).Error; err != nil {
		return nil, err
	}
	return ops, nil
}

// UpdateAllowedOperation 更新操作状态
func (r *InsightRepository) UpdateAllowedOperation(ctx context.Context, op *insight.DASAllowedOperation) error {
	return r.DB(ctx).Save(op).Error
}

// ============ DAS 执行记录 ============

// CreateDASRecord 创建执行记录
func (r *InsightRepository) CreateDASRecord(ctx context.Context, record *insight.DASRecord) error {
	return r.DB(ctx).Create(record).Error
}

// GetDASRecords 获取执行记录
func (r *InsightRepository) GetDASRecords(ctx context.Context, username string, page, pageSize int) ([]insight.DASRecord, int64, error) {
	var records []insight.DASRecord
	var total int64

	query := r.DB(ctx).Model(&insight.DASRecord{})
	if username != "" {
		query = query.Where("username = ?", username)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// ============ DAS 收藏夹 ============

// GetFavorites 获取收藏夹列表
func (r *InsightRepository) GetFavorites(ctx context.Context, username string) ([]insight.DASFavorite, error) {
	var favorites []insight.DASFavorite
	if err := r.DB(ctx).Where("username = ?", username).Order("created_at DESC").Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

// CreateFavorite 创建收藏
func (r *InsightRepository) CreateFavorite(ctx context.Context, fav *insight.DASFavorite) error {
	return r.DB(ctx).Create(fav).Error
}

// UpdateFavorite 更新收藏
func (r *InsightRepository) UpdateFavorite(ctx context.Context, fav *insight.DASFavorite) error {
	return r.DB(ctx).Save(fav).Error
}

// DeleteFavorite 删除收藏
func (r *InsightRepository) DeleteFavorite(ctx context.Context, id uint, username string) error {
	return r.DB(ctx).Where("id = ? AND username = ?", id, username).Delete(&insight.DASFavorite{}).Error
}
