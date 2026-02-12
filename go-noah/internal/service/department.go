package service

import (
	"context"
	"fmt"
	"go-noah/api"
	"go-noah/internal/model"
	"go-noah/internal/repository"
	"go-noah/pkg/global"
)

// DepartmentServiceApp 全局 Service 实例
var DepartmentServiceApp = new(DepartmentService)

type DepartmentService struct{}

func (s *DepartmentService) getDeptRepo() *repository.DepartmentRepository {
	return repository.NewDepartmentRepository(repository.NewRepository(global.Logger, global.DB, global.Enforcer))
}

// GetDepartmentTree 获取部门树
func (s *DepartmentService) GetDepartmentTree(ctx context.Context) (*api.DepartmentTreeData, error) {
	repo := s.getDeptRepo()
	list, err := repo.GetDepartmentTree(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树结构
	tree := s.buildDepartmentTree(list, 0)
	return &api.DepartmentTreeData{
		List: tree,
	}, nil
}

// buildDepartmentTree 构建部门树
func (s *DepartmentService) buildDepartmentTree(list []model.Department, parentID uint) []api.DepartmentItem {
	var result []api.DepartmentItem
	for _, dept := range list {
		if dept.ParentID == parentID {
			item := api.DepartmentItem{
				ID:       dept.ID,
				ParentID: dept.ParentID,
				Name:     dept.Name,
				Code:     dept.Code,
				Path:     dept.Path,
				Level:    dept.Level,
				Leader:   dept.Leader,
				LeaderID: dept.LeaderID,
				Sort:     dept.Sort,
				Status:   dept.Status,
				Children: s.buildDepartmentTree(list, dept.ID),
			}
			result = append(result, item)
		}
	}
	return result
}

// GetDepartmentList 获取部门列表（扁平）
func (s *DepartmentService) GetDepartmentList(ctx context.Context) (*api.DepartmentListData, error) {
	repo := s.getDeptRepo()
	list, err := repo.GetDepartmentList(ctx)
	if err != nil {
		return nil, err
	}

	var items []api.DepartmentItem
	for _, dept := range list {
		items = append(items, api.DepartmentItem{
			ID:       dept.ID,
			ParentID: dept.ParentID,
			Name:     dept.Name,
			Code:     dept.Code,
			Path:     dept.Path,
			Level:    dept.Level,
			Leader:   dept.Leader,
			LeaderID: dept.LeaderID,
			Sort:     dept.Sort,
			Status:   dept.Status,
		})
	}

	return &api.DepartmentListData{
		List: items,
	}, nil
}

// GetDepartment 获取部门详情
func (s *DepartmentService) GetDepartment(ctx context.Context, id uint) (*api.DepartmentItem, error) {
	repo := s.getDeptRepo()
	dept, err := repo.GetDepartment(ctx, id)
	if err != nil {
		return nil, err
	}

	return &api.DepartmentItem{
		ID:       dept.ID,
		ParentID: dept.ParentID,
		Name:     dept.Name,
		Code:     dept.Code,
		Path:     dept.Path,
		Level:    dept.Level,
		Leader:   dept.Leader,
		LeaderID: dept.LeaderID,
		Sort:     dept.Sort,
		Status:   dept.Status,
	}, nil
}

// CreateDepartment 创建部门
func (s *DepartmentService) CreateDepartment(ctx context.Context, req *api.CreateDepartmentRequest) error {
	repo := s.getDeptRepo()
	// 检查编码是否重复
	existing, _ := repo.GetDepartmentByCode(ctx, req.Code)
	if existing != nil {
		return fmt.Errorf("部门编码 %s 已存在", req.Code)
	}

	// 计算 path 和 level
	var path string
	var level int
	if req.ParentID > 0 {
		parent, err := repo.GetDepartment(ctx, req.ParentID)
		if err != nil {
			return fmt.Errorf("父部门不存在")
		}
		level = parent.Level + 1
		// path 在创建后更新
	} else {
		level = 1
	}

	dept := &model.Department{
		ParentID: req.ParentID,
		Name:     req.Name,
		Code:     req.Code,
		Level:    level,
		Leader:   req.Leader,
		LeaderID: req.LeaderID,
		Sort:     req.Sort,
		Status:   req.Status,
	}

	if err := repo.CreateDepartment(ctx, dept); err != nil {
		return err
	}

	// 更新 path
	if req.ParentID > 0 {
		parent, _ := repo.GetDepartment(ctx, req.ParentID)
		path = fmt.Sprintf("%s%d/", parent.Path, dept.ID)
	} else {
		path = fmt.Sprintf("/%d/", dept.ID)
	}
	dept.Path = path
	return repo.UpdateDepartment(ctx, dept)
}

// UpdateDepartment 更新部门
func (s *DepartmentService) UpdateDepartment(ctx context.Context, req *api.UpdateDepartmentRequest) error {
	repo := s.getDeptRepo()
	dept, err := repo.GetDepartment(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("部门不存在")
	}

	// 检查编码是否重复（排除自己）
	if req.Code != dept.Code {
		existing, _ := repo.GetDepartmentByCode(ctx, req.Code)
		if existing != nil && existing.ID != req.ID {
			return fmt.Errorf("部门编码 %s 已存在", req.Code)
		}
	}

	dept.Name = req.Name
	dept.Code = req.Code
	dept.Leader = req.Leader
	dept.LeaderID = req.LeaderID
	dept.Sort = req.Sort
	dept.Status = req.Status

	return repo.UpdateDepartment(ctx, dept)
}

// DeleteDepartment 删除部门
func (s *DepartmentService) DeleteDepartment(ctx context.Context, id uint) error {
	repo := s.getDeptRepo()
	return repo.DeleteDepartment(ctx, id)
}

// GetDepartmentUsers 获取部门用户
func (s *DepartmentService) GetDepartmentUsers(ctx context.Context, deptID uint) (*api.DepartmentUsersData, error) {
	repo := s.getDeptRepo()
	users, err := repo.GetDepartmentUsers(ctx, deptID)
	if err != nil {
		return nil, err
	}

	var items []api.AdminUserItem
	for _, u := range users {
		items = append(items, api.AdminUserItem{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Phone:    u.Phone,
		})
	}

	return &api.DepartmentUsersData{
		List: items,
	}, nil
}

