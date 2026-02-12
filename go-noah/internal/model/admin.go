package model

import "gorm.io/gorm"

const (
	AdminRole          = "admin"
	AdminUserID        = "1"
	MenuResourcePrefix = "menu:"
	ApiResourcePrefix  = "api:"
	PermSep            = ","

	// 数据范围常量
	DataScopeAll      = 1 // 全部数据
	DataScopeCustom   = 2 // 自定义数据
	DataScopeDept     = 3 // 本部门数据
	DataScopeDeptTree = 4 // 本部门及以下数据
	DataScopeSelf     = 5 // 仅本人数据

	// 业务角色
	RoleDBA       = "dba"
	RoleDeveloper = "developer"
)

type AdminUser struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);not null;uniqueIndex;comment:'用户名'" json:"username"`
	Nickname string `gorm:"type:varchar(50);not null;comment:'昵称'" json:"nickname"`
	Password string `gorm:"type:varchar(255);not null;comment:'密码'" json:"-"`
	Email    string `gorm:"type:varchar(100);not null;comment:'电子邮件'" json:"email"`
	Phone    string `gorm:"type:varchar(20);not null;comment:'手机号'" json:"phone"`
	DeptID   uint   `gorm:"index;default:0;comment:'部门ID'" json:"dept_id"`
	Status   int8   `gorm:"default:1;comment:'状态:1启用,0禁用'" json:"status"`
}

func (m *AdminUser) TableName() string {
	return "admin_users"
}

type Role struct {
	gorm.Model
	Name        string `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;comment:角色名"`
	Sid         string `json:"sid" gorm:"column:sid;type:varchar(100);uniqueIndex;comment:角色标识"`
	Description string `json:"description" gorm:"column:description;type:varchar(255);comment:角色描述"`
	DataScope   int8   `json:"data_scope" gorm:"column:data_scope;default:1;comment:数据范围:1全部,2自定义,3本部门,4本部门及以下,5仅本人"`
	Sort        int    `json:"sort" gorm:"column:sort;default:0;comment:排序"`
	Status      int8   `json:"status" gorm:"column:status;default:1;comment:状态:1启用,0禁用"`
}

func (m *Role) TableName() string {
	return "roles"
}

type Api struct {
	gorm.Model
	Group  string `gorm:"type:varchar(100);not null;comment:'API分组'"`
	Name   string `gorm:"type:varchar(100);not null;comment:'API名称'"`
	Path   string `gorm:"type:varchar(255);not null;comment:'API路径'"`
	Method string `gorm:"type:varchar(20);not null;comment:'HTTP方法'"`
}

func (m *Api) TableName() string {
	return "api"
}

// ApiIgnore 忽略的 API（不参与同步与鉴权）
type ApiIgnore struct {
	gorm.Model
	Path   string `gorm:"type:varchar(255);not null;uniqueIndex:idx_api_ignore_path_method;comment:'API路径'" json:"path"`
	Method string `gorm:"type:varchar(20);not null;uniqueIndex:idx_api_ignore_path_method;comment:'HTTP方法'" json:"method"`
}

func (m *ApiIgnore) TableName() string {
	return "api_ignore"
}
