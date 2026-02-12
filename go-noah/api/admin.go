package api

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}
type LoginResponseData struct {
	AccessToken string `json:"accessToken"`
}
type LoginResponse struct {
	Response
	Data LoginResponseData
}

type AdminUserDataItem struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username" binding:"required" example:"张三"`
	Nickname  string   `json:"nickname" binding:"required" example:"小Baby"`
	Password  string   `json:"password" binding:"required" example:"123456"`
	Email     string   `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Phone     string   `json:"phone" example:"1858888888"`
	Roles     []string `json:"roles" example:""`
	UpdatedAt string   `json:"updatedAt"`
	CreatedAt string   `json:"createdAt"`
}
type GetAdminUsersRequest struct {
	Page     int    `form:"page" binding:"required" example:"1"`
	PageSize int    `form:"pageSize" binding:"required" example:"10"`
	Username string `form:"username" binding:"" example:"张三"`
	Nickname string `form:"nickname" binding:"" example:"小Baby"`
	Phone    string `form:"phone" binding:"" example:"1858888888"`
	Email    string `form:"email" binding:"" example:"1234@gmail.com"`
}
type GetAdminUserResponseData struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username" example:"张三"`
	Nickname  string   `json:"nickname" example:"小Baby"`
	Password  string   `json:"password" example:"123456"`
	Email     string   `json:"email" example:"1234@gmail.com"`
	Phone     string   `form:"phone" json:"phone" example:"1858888888"`
	Roles     []string `json:"roles" example:""`
	UpdatedAt string   `json:"updatedAt"`
	CreatedAt string   `json:"createdAt"`
}
type GetAdminUserResponse struct {
	Response
	Data GetAdminUserResponseData
}
type GetAdminUsersResponseData struct {
	List  []AdminUserDataItem `json:"list"`
	Total int64               `json:"total"`
}
type GetAdminUsersResponse struct {
	Response
	Data GetAdminUsersResponseData
}
type AdminUserCreateRequest struct {
	Username string   `json:"username" binding:"required" example:"张三"`
	Nickname string   `json:"nickname" binding:"" example:"小Baby"`
	Password string   `json:"password" binding:"required" example:"123456"`
	Email    string   `json:"email" binding:"" example:"1234@gmail.com"`
	Phone    string   `form:"phone" binding:"" example:"1858888888"`
	Roles    []string `json:"roles" example:""`
}
type AdminUserUpdateRequest struct {
	ID       uint     `json:"id"`
	Username string   `json:"username" binding:"required" example:"张三"`
	Nickname string   `json:"nickname" binding:"" example:"小Baby"`
	Password string   `json:"password" binding:"" example:"123456"`
	Email    string   `json:"email" binding:"" example:"1234@gmail.com"`
	Phone    string   `form:"phone" binding:"" example:"1858888888"`
	Roles    []string `json:"roles" example:""`
}
type AdminUserDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}

type MenuDataItem struct {
	ID         uint   `json:"id,omitempty"`         // 唯一id，使用整数表示
	ParentID   uint   `json:"parentId,omitempty"`   // 父级菜单的id，使用整数表示
	Weight     int    `json:"weight"`               // 排序权重
	Path       string `json:"path"`                 // 地址
	Title      string `json:"title"`                // 展示名称
	Name       string `json:"name,omitempty"`       // 同路由中的name，唯一标识
	Component  string `json:"component,omitempty"`  // 绑定的组件
	Locale     string `json:"locale,omitempty"`     // 本地化标识
	Icon       string `json:"icon,omitempty"`       // 图标，使用字符串表示
	Redirect   string `json:"redirect,omitempty"`   // 重定向地址
	KeepAlive  bool   `json:"keepAlive,omitempty"`  // 是否保活
	HideInMenu bool   `json:"hideInMenu,omitempty"` // 是否保活
	URL        string `json:"url,omitempty"`        // iframe模式下的跳转url，不能与path重复
	UpdatedAt  string `json:"updatedAt,omitempty"`  // 是否保活
}
type GetMenuResponseData struct {
	List []MenuDataItem `json:"list"`
}

type GetMenuResponse struct {
	Response
	Data GetMenuResponseData
}

// Soybean-admin 格式的菜单响应
type SoybeanMenuDataItem struct {
	ID         uint                   `json:"id,omitempty"`
	CreateBy   string                 `json:"createBy,omitempty"`
	CreateTime string                 `json:"createTime,omitempty"`
	UpdateBy   string                 `json:"updateBy,omitempty"`
	UpdateTime string                 `json:"updateTime,omitempty"`
	Status     string                 `json:"status,omitempty"`
	ParentID   uint                   `json:"parentId,omitempty"`
	MenuType   string                 `json:"menuType,omitempty"`
	MenuName   string                 `json:"menuName,omitempty"`
	RouteName  string                 `json:"routeName,omitempty"`
	RoutePath  string                 `json:"routePath,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Order      int                    `json:"order,omitempty"`
	I18nKey    string                 `json:"i18nKey,omitempty"`
	Icon       string                 `json:"icon,omitempty"`
	IconType   string                 `json:"iconType,omitempty"`
	MultiTab   bool                   `json:"multiTab,omitempty"`
	HideInMenu bool                   `json:"hideInMenu,omitempty"`
	ActiveMenu string                 `json:"activeMenu,omitempty"`
	KeepAlive  bool                   `json:"keepAlive,omitempty"`
	Constant   bool                   `json:"constant,omitempty"`
	Href       string                 `json:"href,omitempty"`
	Query      []map[string]string    `json:"query,omitempty"`
	Buttons    []map[string]string    `json:"buttons,omitempty"`
	Children   []*SoybeanMenuDataItem `json:"children,omitempty"`
}

type GetSoybeanMenuResponseData struct {
	Records []*SoybeanMenuDataItem `json:"records"`
	Current int                    `json:"current"`
	Size    int                    `json:"size"`
	Total   int                    `json:"total"`
}

type GetSoybeanMenuResponse struct {
	Response
	Data GetSoybeanMenuResponseData
}

type MenuCreateRequest struct {
	ParentID   uint   `json:"parentId,omitempty"`   // 父级菜单的id，使用整数表示
	Weight     int    `json:"weight"`               // 排序权重
	Path       string `json:"path"`                 // 地址
	Title      string `json:"title"`                // 展示名称
	Name       string `json:"name,omitempty"`       // 同路由中的name，唯一标识
	Component  string `json:"component,omitempty"`  // 绑定的组件
	Locale     string `json:"locale,omitempty"`     // 本地化标识
	Icon       string `json:"icon,omitempty"`       // 图标，使用字符串表示
	Redirect   string `json:"redirect,omitempty"`   // 重定向地址
	KeepAlive  bool   `json:"keepAlive,omitempty"`  // 是否保活
	HideInMenu bool   `json:"hideInMenu,omitempty"` // 是否保活
	URL        string `json:"url,omitempty"`        // iframe模式下的跳转url，不能与path重复

	// Soybean-admin 格式字段
	MenuType   string `json:"menuType,omitempty"`   // 菜单类型:1-目录,2-菜单
	MenuName   string `json:"menuName,omitempty"`   // 菜单名称
	RouteName  string `json:"routeName,omitempty"`  // 路由名称
	RoutePath  string `json:"routePath,omitempty"`  // 路由路径
	I18nKey    string `json:"i18nKey,omitempty"`    // 国际化key
	IconType   string `json:"iconType,omitempty"`   // 图标类型:1-iconify,2-local
	Order      int    `json:"order,omitempty"`      // 排序
	Status     string `json:"status,omitempty"`     // 状态:1-启用,2-禁用
	MultiTab   bool   `json:"multiTab,omitempty"`   // 是否多标签
	ActiveMenu string `json:"activeMenu,omitempty"` // 激活菜单
	Constant   bool   `json:"constant,omitempty"`   // 是否常量
	Href       string `json:"href,omitempty"`       // 外部链接
}
type MenuUpdateRequest struct {
	ID         uint   `json:"id,omitempty"`         // 唯一id，使用整数表示
	ParentID   uint   `json:"parentId,omitempty"`   // 父级菜单的id，使用整数表示
	Weight     int    `json:"weight"`               // 排序权重
	Path       string `json:"path"`                 // 地址
	Title      string `json:"title"`                // 展示名称
	Name       string `json:"name,omitempty"`       // 同路由中的name，唯一标识
	Component  string `json:"component,omitempty"`  // 绑定的组件
	Locale     string `json:"locale,omitempty"`     // 本地化标识
	Icon       string `json:"icon,omitempty"`       // 图标，使用字符串表示
	Redirect   string `json:"redirect,omitempty"`   // 重定向地址
	KeepAlive  bool   `json:"keepAlive,omitempty"`  // 是否保活
	HideInMenu bool   `json:"hideInMenu,omitempty"` // 是否保活
	URL        string `json:"url,omitempty"`        // iframe模式下的跳转url，不能与path重复
	UpdatedAt  string `json:"updatedAt"`

	// Soybean-admin 格式字段
	MenuType   string `json:"menuType,omitempty"`   // 菜单类型:1-目录,2-菜单
	MenuName   string `json:"menuName,omitempty"`   // 菜单名称
	RouteName  string `json:"routeName,omitempty"`  // 路由名称
	RoutePath  string `json:"routePath,omitempty"`  // 路由路径
	I18nKey    string `json:"i18nKey,omitempty"`    // 国际化key
	IconType   string `json:"iconType,omitempty"`   // 图标类型:1-iconify,2-local
	Order      int    `json:"order,omitempty"`      // 排序
	Status     string `json:"status,omitempty"`     // 状态:1-启用,2-禁用
	MultiTab   bool   `json:"multiTab,omitempty"`   // 是否多标签
	ActiveMenu string `json:"activeMenu,omitempty"` // 激活菜单
	Constant   bool   `json:"constant,omitempty"`   // 是否常量
	Href       string `json:"href,omitempty"`       // 外部链接
}
type MenuDeleteRequest struct {
	ID uint `form:"id"` // 唯一id，使用整数表示
}
type GetRoleListRequest struct {
	Page     int    `form:"page" binding:"required" example:"1"`
	PageSize int    `form:"pageSize" binding:"required" example:"10"`
	Sid      string `form:"sid" binding:"" example:"1"`
	Name     string `form:"name" binding:"" example:"Admin"`
}
type RoleDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Sid       string `json:"sid"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}
type GetRolesResponseData struct {
	List  []RoleDataItem `json:"list"`
	Total int64          `json:"total"`
}
type GetRolesResponse struct {
	Response
	Data GetRolesResponseData
}
type RoleCreateRequest struct {
	Sid  string `form:"sid" binding:"required" example:"1"`
	Name string `form:"name" binding:"required" example:"Admin"`
}
type RoleUpdateRequest struct {
	ID   uint   `form:"id" binding:"required" example:"1"`
	Sid  string `form:"sid" binding:"required" example:"1"`
	Name string `form:"name" binding:"required" example:"Admin"`
}
type RoleDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}
type PermissionCreateRequest struct {
	Sid  string `form:"sid" binding:"required" example:"1"`
	Name string `form:"name" binding:"required" example:"Admin"`
}
type GetApisRequest struct {
	Page     int    `form:"page" binding:"required" example:"1"`
	PageSize int    `form:"pageSize" binding:"required" example:"10"`
	Group    string `form:"group" binding:"" example:"权限管理"`
	Name     string `form:"name" binding:"" example:"菜单列表"`
	Path     string `form:"path" binding:"" example:"/v1/test"`
	Method   string `form:"method" binding:"" example:"GET"`
}
type ApiDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Group     string `json:"group"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}
type GetApisResponseData struct {
	List   []ApiDataItem `json:"list"`
	Total  int64         `json:"total"`
	Groups []string      `json:"groups"`
}
type GetApisResponse struct {
	Response
	Data GetApisResponseData
}
type ApiCreateRequest struct {
	Group  string `form:"group" binding:"" example:"权限管理"`
	Name   string `form:"name" binding:"" example:"菜单列表"`
	Path   string `form:"path" binding:"" example:"/v1/test"`
	Method string `form:"method" binding:"" example:"GET"`
}
type ApiUpdateRequest struct {
	ID     uint   `form:"id" binding:"required" example:"1"`
	Group  string `form:"group" binding:"" example:"权限管理"`
	Name   string `form:"name" binding:"" example:"菜单列表"`
	Path   string `form:"path" binding:"" example:"/v1/test"`
	Method string `form:"method" binding:"" example:"GET"`
}
type ApiDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}

// SyncApiResponse 同步 API 对比结果（代码路由 vs 数据库）
type SyncApiResponse struct {
	NewApis    []SyncApiItem `json:"newApis"`    // 代码中有、数据库中无
	DeleteApis []ApiDataItem `json:"deleteApis"` // 数据库中有、代码中无
	IgnoreApis []SyncApiItem `json:"ignoreApis"` // 已忽略的路由（不参与鉴权）
}
type SyncApiItem struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Group  string `json:"group,omitempty"`
	Name   string `json:"name,omitempty"`
}

// EnterSyncApiRequest 确认同步：将对比结果写入数据库
type EnterSyncApiRequest struct {
	NewApis    []ApiCreateRequest `json:"newApis"`    // 要新增的 API（可带 group/name）
	DeleteApis []SyncApiItem      `json:"deleteApis"` // 要删除的 API（path+method）
}

// IgnoreApiRequest 忽略/取消忽略 API（不参与同步与鉴权）
type IgnoreApiRequest struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
	Flag   bool   `json:"flag"` // true=加入忽略列表，false=从忽略列表移除
}

// GetApiByIdRequest 按 ID 获取单条 API
type GetApiByIdRequest struct {
	ID uint `form:"id" binding:"required"`
}

// DeleteApisByIdsRequest 批量删除 API
type DeleteApisByIdsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// ApiAiFillRequest AI 自动填充请求：传入待填充的 path+method 列表
type ApiAiFillRequest struct {
	Items []SyncApiItem `json:"items" binding:"required"` // path + method
}

// ApiAiFillItem AI 自动填充单条结果（与 SyncApiItem 一致，带 group/name）
type ApiAiFillItem struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Group  string `json:"group"`
	Name   string `json:"name"`
}

type GetUserPermissionsData struct {
	List []string `json:"list"`
}
type GetRolePermissionsRequest struct {
	Role string `form:"role" binding:"required" example:"admin"`
}
type GetRolePermissionsData struct {
	List []string `json:"list"`
}
type UpdateRolePermissionRequest struct {
	Role string   `form:"role" binding:"required" example:"admin"`
	List []string `form:"list" binding:"required" example:""`
}

// ==================== 动态路由 ====================

// ElegantRouteMeta 路由元信息（soybean-admin格式）
type ElegantRouteMeta struct {
	Title           string `json:"title,omitempty"`
	I18nKey         string `json:"i18nKey,omitempty"`
	Icon            string `json:"icon,omitempty"`
	LocalIcon       string `json:"localIcon,omitempty"` // 本地图标名称（当 iconType = "2" 时使用）
	Order           int    `json:"order,omitempty"`
	KeepAlive       bool   `json:"keepAlive,omitempty"`
	Constant        bool   `json:"constant,omitempty"`
	HideInMenu      bool   `json:"hideInMenu,omitempty"`
	ActiveMenu      string `json:"activeMenu,omitempty"`
	MultiTab        bool   `json:"multiTab,omitempty"`
	FixedIndexInTab *int   `json:"fixedIndexInTab,omitempty"`
	Href            string `json:"href,omitempty"`
}

// ElegantRoute 路由项（soybean-admin格式）
type ElegantRoute struct {
	Name      string           `json:"name"`
	Path      string           `json:"path"`
	Component string           `json:"component,omitempty"`
	Redirect  string           `json:"redirect,omitempty"`
	Meta      ElegantRouteMeta `json:"meta"`
	Children  []ElegantRoute   `json:"children,omitempty"`
}

// UserRouteData 用户路由数据
type UserRouteData struct {
	Routes []ElegantRoute `json:"routes"`
	Home   string         `json:"home"`
}

// GetUserRoutesResponse 获取用户路由响应
type GetUserRoutesResponse struct {
	Response
	Data UserRouteData `json:"data"`
}

// RouteInfo 路由信息（用于同步路由到数据库）
type RouteInfo struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}
