package insight

import (
	"context"
	"go-noah/internal/model/insight"
	"go-noah/internal/repository"
	"go-noah/pkg/log"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

// InsightRepository goInsight 功能的数据访问层
type InsightRepository struct {
	repo     *repository.Repository
	logger   *log.Logger
	enforcer *casbin.SyncedEnforcer
}

func NewInsightRepository(r *repository.Repository, logger *log.Logger, e *casbin.SyncedEnforcer) *InsightRepository {
	return &InsightRepository{
		repo:     r,
		logger:   logger,
		enforcer: e,
	}
}

func (r *InsightRepository) DB(ctx context.Context) *gorm.DB {
	return r.repo.DB(ctx)
}

func (r *InsightRepository) Logger() *log.Logger {
	return r.logger
}

func (r *InsightRepository) Enforcer() *casbin.SyncedEnforcer {
	return r.enforcer
}

// ============ 环境管理 ============

// GetEnvironments 获取环境列表
func (r *InsightRepository) GetEnvironments(ctx context.Context) ([]insight.DBEnvironment, error) {
	var environments []insight.DBEnvironment
	if err := r.DB(ctx).Find(&environments).Error; err != nil {
		return nil, err
	}
	return environments, nil
}

// CreateEnvironment 创建环境
func (r *InsightRepository) CreateEnvironment(ctx context.Context, env *insight.DBEnvironment) error {
	return r.DB(ctx).Create(env).Error
}

// UpdateEnvironment 更新环境
func (r *InsightRepository) UpdateEnvironment(ctx context.Context, env *insight.DBEnvironment) error {
	return r.DB(ctx).Save(env).Error
}

// DeleteEnvironment 删除环境
func (r *InsightRepository) DeleteEnvironment(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DBEnvironment{}, id).Error
}

// ============ 数据库配置管理 ============

// GetDBConfigs 获取数据库配置列表
func (r *InsightRepository) GetDBConfigs(ctx context.Context, useType insight.UseType, environment int) ([]insight.DBConfig, error) {
	var configs []insight.DBConfig
	query := r.DB(ctx)
	if useType != "" {
		query = query.Where("use_type = ?", useType)
	}
	if environment > 0 {
		query = query.Where("environment = ?", environment)
	}
	if err := query.Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetDBConfigByInstanceID 根据实例ID获取配置
func (r *InsightRepository) GetDBConfigByInstanceID(ctx context.Context, instanceID string) (*insight.DBConfig, error) {
	var config insight.DBConfig
	if err := r.DB(ctx).Where("instance_id = ?", instanceID).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateDBConfig 创建数据库配置
func (r *InsightRepository) CreateDBConfig(ctx context.Context, config *insight.DBConfig) error {
	return r.DB(ctx).Create(config).Error
}

// UpdateDBConfig 更新数据库配置
func (r *InsightRepository) UpdateDBConfig(ctx context.Context, config *insight.DBConfig) error {
	return r.DB(ctx).Save(config).Error
}

func (r *InsightRepository) UpdateDBConfigFields(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.DB(ctx).Model(&insight.DBConfig{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDBConfig 删除数据库配置
func (r *InsightRepository) DeleteDBConfig(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&insight.DBConfig{}, id).Error
}

// ============ Schema 管理 ============

// GetSchemasByInstanceID 获取实例下的所有Schema
func (r *InsightRepository) GetSchemasByInstanceID(ctx context.Context, instanceID string) ([]insight.DBSchema, error) {
	var schemas []insight.DBSchema
	if err := r.DB(ctx).Where("instance_id = ? AND is_deleted = ?", instanceID, false).Find(&schemas).Error; err != nil {
		return nil, err
	}
	return schemas, nil
}

// SyncSchemas 同步Schema信息
func (r *InsightRepository) SyncSchemas(ctx context.Context, instanceID string, schemas []string) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 标记所有现有schema为已删除
		if err := tx.Model(&insight.DBSchema{}).
			Where("instance_id = ?", instanceID).
			Update("is_deleted", true).Error; err != nil {
			return err
		}

		// 更新或创建schema
		for _, schema := range schemas {
			var existing insight.DBSchema
			err := tx.Where("instance_id = ? AND `schema` = ?", instanceID, schema).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&insight.DBSchema{
					InstanceID: existing.InstanceID,
					Schema:     schema,
					IsDeleted:  false,
				}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// 更新现有记录
				if err := tx.Model(&existing).Update("is_deleted", false).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ============ 组织管理 ============

// GetOrganizations 获取组织树
func (r *InsightRepository) GetOrganizations(ctx context.Context) ([]insight.Organization, error) {
	var orgs []insight.Organization
	if err := r.DB(ctx).Order("level ASC, id ASC").Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrganizationByID 根据ID获取组织
func (r *InsightRepository) GetOrganizationByID(ctx context.Context, id uint64) (*insight.Organization, error) {
	var org insight.Organization
	if err := r.DB(ctx).First(&org, id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// CreateOrganization 创建组织
func (r *InsightRepository) CreateOrganization(ctx context.Context, org *insight.Organization) error {
	return r.DB(ctx).Create(org).Error
}

// UpdateOrganization 更新组织
func (r *InsightRepository) UpdateOrganization(ctx context.Context, org *insight.Organization) error {
	return r.DB(ctx).Save(org).Error
}

// DeleteOrganization 删除组织
func (r *InsightRepository) DeleteOrganization(ctx context.Context, id uint64) error {
	return r.DB(ctx).Delete(&insight.Organization{}, id).Error
}

// GetOrganizationUsers 获取组织下的用户
func (r *InsightRepository) GetOrganizationUsers(ctx context.Context, orgKey string) ([]insight.OrganizationUser, error) {
	var users []insight.OrganizationUser
	if err := r.DB(ctx).Where("organization_key LIKE ?", orgKey+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// BindOrganizationUser 绑定用户到组织
func (r *InsightRepository) BindOrganizationUser(ctx context.Context, ou *insight.OrganizationUser) error {
	return r.DB(ctx).Create(ou).Error
}

// UnbindOrganizationUser 解绑用户
func (r *InsightRepository) UnbindOrganizationUser(ctx context.Context, uid uint64) error {
	return r.DB(ctx).Where("uid = ?", uid).Delete(&insight.OrganizationUser{}).Error
}

