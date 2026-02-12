package repository

import (
	"context"
	"fmt"
	"go-noah/internal/model"
)

// DepartmentRepository 部门数据访问层
type DepartmentRepository struct {
	*Repository
}

func NewDepartmentRepository(repository *Repository) *DepartmentRepository {
	return &DepartmentRepository{
		Repository: repository,
	}
}

// GetDepartmentTree 获取部门树
func (r *DepartmentRepository) GetDepartmentTree(ctx context.Context) ([]model.Department, error) {
	var departments []model.Department
	err := r.DB(ctx).Where("status = ?", 1).Order("sort, id").Find(&departments).Error
	return departments, err
}

// GetDepartmentList 获取部门列表
func (r *DepartmentRepository) GetDepartmentList(ctx context.Context) ([]model.Department, error) {
	var departments []model.Department
	err := r.DB(ctx).Order("sort, id").Find(&departments).Error
	return departments, err
}

// GetDepartment 根据ID获取部门
func (r *DepartmentRepository) GetDepartment(ctx context.Context, id uint) (*model.Department, error) {
	var dept model.Department
	err := r.DB(ctx).Where("id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// GetDepartmentByCode 根据编码获取部门
func (r *DepartmentRepository) GetDepartmentByCode(ctx context.Context, code string) (*model.Department, error) {
	var dept model.Department
	err := r.DB(ctx).Where("code = ?", code).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// CreateDepartment 创建部门
func (r *DepartmentRepository) CreateDepartment(ctx context.Context, dept *model.Department) error {
	return r.DB(ctx).Create(dept).Error
}

// UpdateDepartment 更新部门
func (r *DepartmentRepository) UpdateDepartment(ctx context.Context, dept *model.Department) error {
	return r.DB(ctx).Save(dept).Error
}

// DeleteDepartment 删除部门
func (r *DepartmentRepository) DeleteDepartment(ctx context.Context, id uint) error {
	// 检查是否有子部门
	var count int64
	r.DB(ctx).Model(&model.Department{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("该部门下存在子部门，无法删除")
	}

	// 检查是否有用户
	r.DB(ctx).Model(&model.AdminUser{}).Where("dept_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("该部门下存在用户，无法删除")
	}

	return r.DB(ctx).Delete(&model.Department{}, id).Error
}

// GetDepartmentsByIDs 根据ID列表获取部门
func (r *DepartmentRepository) GetDepartmentsByIDs(ctx context.Context, ids []uint) ([]model.Department, error) {
	var departments []model.Department
	if len(ids) == 0 {
		return departments, nil
	}
	err := r.DB(ctx).Where("id IN ?", ids).Find(&departments).Error
	return departments, err
}

// GetChildDepartments 获取子部门（递归）
func (r *DepartmentRepository) GetChildDepartments(ctx context.Context, parentID uint) ([]model.Department, error) {
	var departments []model.Department
	// 使用 path 字段来查找所有子部门
	var parent model.Department
	if err := r.DB(ctx).Where("id = ?", parentID).First(&parent).Error; err != nil {
		return nil, err
	}

	// 查找所有 path 以父部门 path 开头的部门
	err := r.DB(ctx).Where("path LIKE ?", parent.Path+"%").Find(&departments).Error
	return departments, err
}

// GetDepartmentUsers 获取部门用户
func (r *DepartmentRepository) GetDepartmentUsers(ctx context.Context, deptID uint) ([]model.AdminUser, error) {
	var users []model.AdminUser
	err := r.DB(ctx).Where("dept_id = ?", deptID).Find(&users).Error
	return users, err
}

// UpdateDepartmentPath 更新部门路径（用于移动部门时）
func (r *DepartmentRepository) UpdateDepartmentPath(ctx context.Context, id uint, newPath string, newLevel int) error {
	return r.DB(ctx).Model(&model.Department{}).Where("id = ?", id).Updates(map[string]interface{}{
		"path":  newPath,
		"level": newLevel,
	}).Error
}

