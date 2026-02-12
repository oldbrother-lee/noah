package api

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=2,max=32" example:"zhangsan"`
	Password    string `json:"password" binding:"required,min=7,max=128" example:"1234567"`
	Email       string `json:"email" binding:"required,min=3,max=254,email" example:"zhangsan@example.com"`
	NickName    string `json:"nick_name" binding:"required,min=1,max=32" example:"张三"`
	Mobile      string `json:"mobile" example:"13800138000"`
	RoleID      uint64 `json:"role_id" example:"1"`
	IsTwoFA     bool   `json:"is_two_fa" example:"false"`
	IsSuperuser bool   `json:"is_superuser" example:"false"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username    string `json:"username" binding:"required,min=2,max=32" example:"zhangsan"`
	Email       string `json:"email" binding:"required,min=3,max=254,email" example:"zhangsan@example.com"`
	NickName    string `json:"nick_name" binding:"required,min=1,max=32" example:"张三"`
	Mobile      string `json:"mobile" example:"13800138000"`
	RoleID      uint64 `json:"role_id" example:"1"`
	IsTwoFA     bool   `json:"is_two_fa" example:"false"`
	IsSuperuser bool   `json:"is_superuser" example:"false"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// GetUsersRequest 获取用户列表请求
type GetUsersRequest struct {
	Page            int    `form:"page" binding:"required" example:"1"`
	PageSize        int    `form:"page_size" binding:"required" example:"10"`
	Search          string `form:"search" example:"张三"`
	OrganizationKey string `form:"organization_key" example:""`
	RoleID          uint64 `form:"role_id" example:"0"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	UID            uint64 `json:"uid" binding:"required" example:"1"`
	Password       string `json:"password" binding:"required,min=7,max=32" example:"1234567"`
	VerifyPassword string `json:"verify_password" binding:"required,min=7,max=32" example:"1234567"`
}

// UserData 用户数据
type UserData struct {
	Uid         uint64 `json:"uid"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	NickName    string `json:"nick_name"`
	Mobile      string `json:"mobile"`
	AvatarFile  string `json:"avatar_file"`
	RoleID      uint64 `json:"role_id"`
	IsSuperuser bool   `json:"is_superuser"`
	IsActive    bool   `json:"is_active"`
	IsStaff     bool   `json:"is_staff"`
	IsTwoFA     bool   `json:"is_two_fa"`
	LastLogin   string `json:"last_login,omitempty"`
	DateJoined  string `json:"date_joined,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// GetUsersResponseData 获取用户列表响应数据
type GetUsersResponseData struct {
	List  []UserData `json:"list"`
	Total int64      `json:"total"`
}

// GetUsersResponse 获取用户列表响应
type GetUsersResponse struct {
	Response
	Data GetUsersResponseData
}

// GetUserResponseData 获取用户详情响应数据
type GetUserResponseData struct {
	UserData
}

// GetUserResponse 获取用户详情响应
type GetUserResponse struct {
	Response
	Data GetUserResponseData
}

