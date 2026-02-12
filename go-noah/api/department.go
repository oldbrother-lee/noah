package api

// ============= 部门管理 API =============

// DepartmentItem 部门项
type DepartmentItem struct {
	ID       uint             `json:"id"`
	ParentID uint             `json:"parentId"`
	Name     string           `json:"name"`
	Code     string           `json:"code"`
	Path     string           `json:"path"`
	Level    int              `json:"level"`
	Leader   string           `json:"leader"`
	LeaderID uint             `json:"leaderId"`
	Sort     int              `json:"sort"`
	Status   int8             `json:"status"`
	Children []DepartmentItem `json:"children,omitempty"`
}

// DepartmentTreeData 部门树数据
type DepartmentTreeData struct {
	List []DepartmentItem `json:"list"`
}

// DepartmentListData 部门列表数据
type DepartmentListData struct {
	List []DepartmentItem `json:"list"`
}

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Leader   string `json:"leader"`
	LeaderID uint   `json:"leaderId"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Leader   string `json:"leader"`
	LeaderID uint   `json:"leaderId"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// DeleteDepartmentRequest 删除部门请求
type DeleteDepartmentRequest struct {
	ID uint `form:"id" binding:"required"`
}

// GetDepartmentRequest 获取部门详情请求
type GetDepartmentRequest struct {
	ID uint `form:"id" binding:"required"`
}

// GetDepartmentUsersRequest 获取部门用户请求
type GetDepartmentUsersRequest struct {
	DeptID uint `form:"deptId" binding:"required"`
}

// DepartmentUsersData 部门用户数据
type DepartmentUsersData struct {
	List []AdminUserItem `json:"list"`
}

// AdminUserItem 管理员用户项（简化）
type AdminUserItem struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

