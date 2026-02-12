package model

import "gorm.io/gorm"

// Department 部门表
type Department struct {
	gorm.Model
	ParentID uint   `gorm:"index;default:0;comment:'父部门ID'" json:"parent_id"`
	Name     string `gorm:"type:varchar(50);not null;comment:'部门名称'" json:"name"`
	Code     string `gorm:"type:varchar(50);uniqueIndex;comment:'部门编码'" json:"code"`
	Path     string `gorm:"type:varchar(255);index;comment:'部门路径,如/1/2/3/'" json:"path"`
	Level    int    `gorm:"default:1;comment:'层级'" json:"level"`
	Leader   string `gorm:"type:varchar(50);comment:'负责人用户名'" json:"leader"`
	LeaderID uint   `gorm:"index;default:0;comment:'负责人ID'" json:"leader_id"`
	Sort     int    `gorm:"default:0;comment:'排序'" json:"sort"`
	Status   int8   `gorm:"default:1;comment:'状态:1启用,0禁用'" json:"status"`
}

func (Department) TableName() string {
	return "departments"
}

// RoleDepartment 角色自定义数据范围关联表
type RoleDepartment struct {
	RoleID uint `gorm:"primaryKey;comment:'角色ID'" json:"role_id"`
	DeptID uint `gorm:"primaryKey;comment:'部门ID'" json:"dept_id"`
}

func (RoleDepartment) TableName() string {
	return "role_departments"
}

