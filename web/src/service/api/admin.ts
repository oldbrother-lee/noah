import { request } from '../request';

/** Admin API */

/**
 * Login
 */
export function fetchAdminLogin(data: Api.Admin.LoginRequest) {
  return request<Api.Admin.LoginResponse>({
    url: '/api/v1/login',
    method: 'post',
    data
  });
}

/**
 * Get current admin user info
 */
export function fetchGetAdminUser() {
  return request<Api.Admin.AdminUser>({
    url: '/api/v1/admin/user',
    method: 'get'
  });
}

/**
 * Get admin user menus
 */
export function fetchGetMenus() {
  return request<Api.Admin.MenuListResponse>({
    url: '/api/v1/menus',
    method: 'get'
  });
}

// ==================== Menu Management ====================

/**
 * Get admin menus (for management) - legacy
 */
export function fetchGetAdminMenus() {
  return request<Api.Admin.MenuListResponse>({
    url: '/api/v1/admin/menus',
    method: 'get'
  });
}

/**
 * Get menu list (soybean-admin format)
 */
export function fetchGetMenuList() {
  return request<Api.SystemManage.MenuList>({
    url: '/api/v1/admin/menus',
    method: 'get'
  });
}

/**
 * Get all pages
 */
export function fetchGetAllPages() {
  return request<string[]>({
    url: '/api/v1/admin/menu/pages',
    method: 'get'
  });
}

/**
 * Create menu
 */
export function fetchCreateMenu(data: Api.Admin.MenuCreateRequest) {
  return request({
    url: '/api/v1/admin/menu',
    method: 'post',
    data
  });
}

/**
 * Update menu
 */
export function fetchUpdateMenu(data: Api.Admin.MenuUpdateRequest) {
  return request({
    url: '/api/v1/admin/menu',
    method: 'put',
    data
  });
}

/**
 * Delete menu
 */
export function fetchDeleteMenu(id: number) {
  return request({
    url: '/api/v1/admin/menu',
    method: 'delete',
    params: { id }
  });
}

// ==================== User Management ====================

/**
 * Get admin users list
 */
export function fetchGetAdminUsers(params?: {
  page: number;
  pageSize: number;
  username?: string;
  nickname?: string;
  phone?: string;
  email?: string;
}) {
  return request<Api.Admin.AdminUserListResponse>({
    url: '/api/v1/admin/users',
    method: 'get',
    params
  });
}

/**
 * Create admin user
 */
export function fetchCreateAdminUser(data: Api.Admin.AdminUserCreateRequest) {
  return request({
    url: '/api/v1/admin/user',
    method: 'post',
    data
  });
}

/**
 * Update admin user
 */
export function fetchUpdateAdminUser(data: Api.Admin.AdminUserUpdateRequest) {
  return request({
    url: '/api/v1/admin/user',
    method: 'put',
    data
  });
}

/**
 * Delete admin user
 */
export function fetchDeleteAdminUser(id: number) {
  return request({
    url: '/api/v1/admin/user',
    method: 'delete',
    params: { id }
  });
}

// ==================== Role Management ====================

/**
 * Get roles list
 */
export function fetchGetRoles(params?: {
  page: number;
  pageSize: number;
  sid?: string;
  name?: string;
}) {
  return request<Api.Admin.RoleListResponse>({
    url: '/api/v1/admin/roles',
    method: 'get',
    params
  });
}

/**
 * Create role
 */
export function fetchCreateRole(data: Api.Admin.RoleCreateRequest) {
  return request({
    url: '/api/v1/admin/role',
    method: 'post',
    data
  });
}

/**
 * Update role
 */
export function fetchUpdateRole(data: Api.Admin.RoleUpdateRequest) {
  return request({
    url: '/api/v1/admin/role',
    method: 'put',
    data
  });
}

/**
 * Delete role
 */
export function fetchDeleteRole(id: number) {
  return request({
    url: '/api/v1/admin/role',
    method: 'delete',
    params: { id }
  });
}

// ==================== API Management ====================

/**
 * Get APIs list
 */
export function fetchGetApis(params?: {
  page: number;
  pageSize: number;
  group?: string;
  name?: string;
  path?: string;
  method?: string;
}) {
  return request<Api.Admin.ApiListResponse>({
    url: '/api/v1/admin/apis',
    method: 'get',
    params
  });
}

/**
 * Create API
 */
export function fetchCreateApi(data: Api.Admin.ApiCreateRequest) {
  return request({
    url: '/api/v1/admin/api',
    method: 'post',
    data
  });
}

/**
 * Update API
 */
export function fetchUpdateApi(data: Api.Admin.ApiUpdateRequest) {
  return request({
    url: '/api/v1/admin/api',
    method: 'put',
    data
  });
}

/**
 * Delete API
 */
export function fetchDeleteApi(id: number) {
  return request({
    url: '/api/v1/admin/api',
    method: 'delete',
    params: { id }
  });
}

/**
 * Sync API - 对比代码路由与数据库，返回 newApis / deleteApis / ignoreApis
 */
export function fetchSyncApi() {
  return request<Api.Admin.SyncApiResponse>({
    url: '/api/v1/admin/api/sync',
    method: 'get'
  });
}

/**
 * Enter Sync API - 确认同步（写入新增、删除待删并清理 Casbin）
 */
export function fetchEnterSyncApi(data: Api.Admin.EnterSyncApiRequest) {
  return request({
    url: '/api/v1/admin/api/sync',
    method: 'post',
    data
  });
}

/**
 * Ignore API - 忽略/取消忽略（不参与同步与鉴权）
 */
export function fetchIgnoreApi(data: Api.Admin.IgnoreApiRequest) {
  return request({
    url: '/api/v1/admin/api/ignore',
    method: 'post',
    data
  });
}

/**
 * Get API by ID
 */
export function fetchGetApiById(id: number) {
  return request<Api.Admin.Api>({
    url: '/api/v1/admin/api/detail',
    method: 'get',
    params: { id }
  });
}

/**
 * Batch delete APIs
 */
export function fetchDeleteApisByIds(ids: number[]) {
  return request({
    url: '/api/v1/admin/api/batch',
    method: 'delete',
    data: { ids }
  });
}

/**
 * Fresh Casbin - 从数据库重新加载 Casbin 策略
 */
export function fetchFreshCasbin() {
  return request({
    url: '/api/v1/admin/casbin/fresh',
    method: 'get'
  });
}

/**
 * AI 自动填充 - 根据 path+method 列表返回建议的 group/name
 */
export function fetchApiAiFill(data: Api.Admin.ApiAiFillRequest) {
  return request<Api.Admin.ApiAiFillItem[]>({
    url: '/api/v1/admin/api/ai-fill',
    method: 'post',
    data
  });
}

// ==================== Permission Management ====================

/**
 * Get user permissions
 */
export function fetchGetUserPermissions() {
  return request<Api.Admin.UserPermissionsResponse>({
    url: '/api/v1/admin/user/permissions',
    method: 'get'
  });
}

/**
 * Get role permissions
 */
export function fetchGetRolePermissions(role: string) {
  return request<Api.Admin.RolePermissionsResponse>({
    url: '/api/v1/admin/role/permissions',
    method: 'get',
    params: { role }
  });
}

/**
 * Update role permissions
 */
export function fetchUpdateRolePermission(data: Api.Admin.UpdateRolePermissionRequest) {
  return request({
    url: '/api/v1/admin/role/permission',
    method: 'put',
    data
  });
}

// ==================== Environment Management ====================

/**
 * Get environments list
 */
export function fetchGetEnvironments() {
  return request<Api.Admin.Environment[]>({
    url: '/api/v1/insight/environments',
    method: 'get'
  });
}

/**
 * Create environment
 */
export function fetchCreateEnvironment(data: Api.Admin.EnvironmentCreateRequest) {
  return request<Api.Admin.Environment>({
    url: '/api/v1/insight/environments',
    method: 'post',
    data
  });
}

/**
 * Update environment
 */
export function fetchUpdateEnvironment(id: number, data: Api.Admin.EnvironmentUpdateRequest) {
  return request<Api.Admin.Environment>({
    url: `/api/v1/insight/environments/${id}`,
    method: 'put',
    data
  });
}

/**
 * Delete environment
 */
export function fetchDeleteEnvironment(id: number) {
  return request({
    url: `/api/v1/insight/environments/${id}`,
    method: 'delete'
  });
}

// ==================== Database Config Management ====================

/**
 * Get database configs list
 */
export function fetchGetDBConfigs(params?: { useType?: string; environment?: number }) {
  // 将驼峰命名转换为下划线命名，以匹配后端参数
  const queryParams: any = {};
  if (params?.useType) {
    queryParams.use_type = params.useType;
  }
  if (params?.environment !== undefined) {
    queryParams.environment = params.environment;
  }
  return request<Api.Admin.DBConfig[]>({
    url: '/api/v1/insight/dbconfigs',
    method: 'get',
    params: queryParams
  });
}

/**
 * Get database config by instance ID
 */
export function fetchGetDBConfig(instanceId: string) {
  return request<Api.Admin.DBConfig>({
    url: `/api/v1/insight/dbconfigs/${instanceId}`,
    method: 'get'
  });
}

/**
 * Create database config
 */
export function fetchCreateDBConfig(data: Api.Admin.DBConfigCreateRequest) {
  return request<Api.Admin.DBConfig>({
    url: '/api/v1/insight/dbconfigs',
    method: 'post',
    data
  });
}

/**
 * Update database config
 */
export function fetchUpdateDBConfig(id: number, data: Api.Admin.DBConfigUpdateRequest) {
  return request<Api.Admin.DBConfig>({
    url: `/api/v1/insight/dbconfigs/${id}`,
    method: 'put',
    data
  });
}

/**
 * Delete database config
 */
export function fetchDeleteDBConfig(id: number) {
  return request({
    url: `/api/v1/insight/dbconfigs/${id}`,
    method: 'delete'
  });
}

