import { request } from '../request';

/** DAS (Data Access Service) API - 新路由: /api/v1/insight/das/... */

/**
 * Get environments
 */
export function fetchEnvironments(params?: Record<string, any>) {
  return request<Api.Das.Environment[]>({
    url: '/api/v1/insight/environments',
    method: 'get',
    params
  });
}

/**
 * Get authorized schemas (用户授权的所有schemas)
 */
export function fetchSchemas(params?: Record<string, any>) {
  return request<Api.Das.Schema[]>({
    url: '/api/v1/insight/das/schemas',
    method: 'get',
    params
  });
}

/**
 * Get authorized tables for schema
 * @param params - { instance_id, schema }
 */
export function fetchTables(params: { instance_id: string; schema: string }) {
  return request<Api.Das.Table[]>({
    url: `/api/v1/insight/das/tables/${params.instance_id}/${params.schema}`,
    method: 'get'
  });
}

/**
 * Get table columns
 * @param params - { instance_id, schema, table }
 */
export function fetchTableColumns(params: { instance_id: string; schema: string; table: string }) {
  // 验证参数
  if (!params?.instance_id || !params?.schema || !params?.table) {
    return Promise.reject(new Error('instance_id、schema 和 table 参数不能为空'));
  }
  if (params.instance_id === 'undefined' || params.schema === 'undefined' || params.table === 'undefined') {
    return Promise.reject(new Error('参数不能为 undefined'));
  }
  return request<Api.Das.ColumnInfo[]>({
    url: `/api/v1/insight/das/columns/${params.instance_id}/${params.schema}/${params.table}`,
    method: 'get'
  });
}

/**
 * Execute SQL query
 */
export function fetchExecuteQuery(data: Api.Das.QueryRequest) {
  return request<Api.Das.QueryResult>({
    url: '/api/v1/insight/das/query',
    method: 'post',
    data,
    skipErrorHandler: true,
    timeout: 30 * 60  * 1000 
  } as any);
}

/**
 * Execute MySQL/TiDB query (alias for fetchExecuteQuery)
 */
export function fetchExecuteMySQLQuery(data: Api.Das.QueryRequest) {
  return fetchExecuteQuery(data);
}

/**
 * Execute ClickHouse query (alias for fetchExecuteQuery)
 */
export function fetchExecuteClickHouseQuery(data: Api.Das.QueryRequest) {
  return fetchExecuteQuery(data);
}

/**
 * Get user grants/permissions
 */
export function fetchUserGrants(params?: Record<string, any>) {
  return request<Api.Das.UserGrant>({
    url: '/api/v1/insight/das/permissions',
    method: 'get',
    params
  });
}

/**
 * Get user permissions (schema and table)
 */
export function fetchGetUserPermissions(username: string) {
  return request<Api.Das.UserPermissionsResponse>({
    url: '/api/v1/insight/das/permissions',
    method: 'get',
    params: { username }
  });
}

/**
 * Grant schema permission
 */
export function fetchGrantSchemaPermission(data: Api.Das.GrantSchemaPermissionRequest) {
  return request<Api.Das.SchemaPermission>({
    url: '/api/v1/insight/das/permissions/schema',
    method: 'post',
    data
  });
}

/**
 * Revoke schema permission
 */
export function fetchRevokeSchemaPermission(id: number) {
  return request({
    url: `/api/v1/insight/das/permissions/schema/${id}`,
    method: 'delete'
  });
}

/**
 * Grant table permission
 */
export function fetchGrantTablePermission(data: Api.Das.GrantTablePermissionRequest) {
  return request<Api.Das.TablePermission>({
    url: '/api/v1/insight/das/permissions/table',
    method: 'post',
    data
  });
}

/**
 * Revoke table permission
 */
export function fetchRevokeTablePermission(id: number) {
  return request({
    url: `/api/v1/insight/das/permissions/table/${id}`,
    method: 'delete'
  });
}

/**
 * Get database dictionary (table info with columns)
 */
export function fetchDBDict(params?: Record<string, any>) {
  if (!params?.instance_id || !params?.schema) {
    return Promise.reject(new Error('instance_id 和 schema 参数不能为空'));
  }
  return request<Api.Das.DBDict>({
    url: `/api/v1/insight/das/tables/${params.instance_id}/${params.schema}`,
    method: 'get'
  });
}

/**
 * Get query history/records
 */
export function fetchHistory(params?: Record<string, any>) {
  return request<Api.Das.HistoryListResponse>({
    url: '/api/v1/insight/das/records',
    method: 'get',
    params
  });
}

/**
 * Get favorites
 */
export function fetchFavorites(params?: Record<string, any>) {
  return request<Api.Das.Favorite[]>({
    url: '/api/v1/insight/das/favorites',
    method: 'get',
    params
  });
}

/**
 * Create favorite
 */
export function fetchCreateFavorite(data: Api.Das.CreateFavoriteRequest) {
  return request<Api.Das.Favorite>({
    url: '/api/v1/insight/das/favorites',
    method: 'post',
    data
  });
}

/**
 * Update favorite
 */
export function fetchUpdateFavorite(data: Api.Das.UpdateFavoriteRequest) {
  return request<Api.Das.Favorite>({
    url: `/api/v1/insight/das/favorites/${data.id}`,
    method: 'put',
    data
  });
}

/**
 * Delete favorite
 */
export function fetchDeleteFavorite(id: string | number) {
  return request({
    url: `/api/v1/insight/das/favorites/${id}`,
    method: 'delete'
  });
}

/**
 * Get table info
 */
export function fetchTableInfo(params?: Record<string, any>) {
  // 验证参数
  if (!params?.instance_id || !params?.schema || !params?.table) {
    return Promise.reject(new Error('instance_id、schema 和 table 参数不能为空'));
  }
  if (params.instance_id === 'undefined' || params.schema === 'undefined' || params.table === 'undefined') {
    return Promise.reject(new Error('参数不能为 undefined'));
  }
  return request<Api.Das.TableInfo>({
    url: `/api/v1/insight/das/columns/${params.instance_id}/${params.schema}/${params.table}`,
    method: 'get'
  });
}

// Admin APIs for DAS management

/**
 * Admin: Get schemas list grant (user schema permissions)
 */
export function fetchAdminSchemasListGrant(params?: Record<string, any>) {
  return request<Api.Das.SchemaGrant[]>({
    url: '/api/v1/insight/das/permissions',
    method: 'get',
    params
  });
}

/**
 * Admin: Create schemas grant
 */
export function fetchAdminCreateSchemasGrant(data: Api.Das.CreateSchemaGrantRequest) {
  return request({
    url: '/api/v1/insight/das/permissions/schema',
    method: 'post',
    data
  });
}

/**
 * Admin: Delete schemas grant
 */
export function fetchAdminDeleteSchemasGrant(id: string | number) {
  return request({
    url: `/api/v1/insight/das/permissions/schema/${id}`,
    method: 'delete'
  });
}

/**
 * Admin: Get tables grant
 */
export function fetchAdminTablesGrant(params?: Record<string, any>) {
  return request<Api.Das.TableGrant[]>({
    url: '/api/v1/insight/das/permissions',
    method: 'get',
    params
  });
}

/**
 * Admin: Create tables grant
 */
export function fetchAdminCreateTablesGrant(data: Api.Das.CreateTableGrantRequest) {
  return request({
    url: '/api/v1/insight/das/permissions/schema',
    method: 'post',
    data
  });
}

/**
 * Admin: Delete tables grant
 */
export function fetchAdminDeleteTablesGrant(id: string | number) {
  return request({
    url: `/api/v1/insight/das/permissions/schema/${id}`,
    method: 'delete'
  });
}

/**
 * Admin: Get instances list (dbconfigs)
 */
export function fetchAdminInstancesList(params?: Record<string, any>) {
  return request<Api.Das.Instance[]>({
    url: '/api/v1/insight/dbconfigs',
    method: 'get',
    params
  });
}

/**
 * Admin: Get schemas list for instance
 */
export function fetchAdminSchemasList(params?: Record<string, any>) {
  return request<Api.Das.Schema[]>({
    url: `/api/v1/insight/dbconfigs/${params?.instance_id}/schemas`,
    method: 'get'
  });
}

/**
 * Admin: Get tables list
 */
export function fetchAdminTablesList(params?: Record<string, any>) {
  return request<Api.Das.Table[]>({
    url: `/api/v1/insight/das/tables/${params?.instance_id}/${params?.schema}`,
    method: 'get'
  });
}

// ==================== 权限模板管理 ====================

/**
 * Get permission templates
 */
export function fetchGetPermissionTemplates() {
  return request<Api.Das.PermissionTemplate[]>({
    url: '/api/v1/insight/das/permissions/templates',
    method: 'get'
  });
}

/**
 * Get permission template by ID
 */
export function fetchGetPermissionTemplate(id: number) {
  return request<Api.Das.PermissionTemplate>({
    url: `/api/v1/insight/das/permissions/templates/${id}`,
    method: 'get'
  });
}

/**
 * Create permission template
 */
export function fetchCreatePermissionTemplate(data: Api.Das.PermissionTemplateCreateRequest) {
  return request<Api.Das.PermissionTemplate>({
    url: '/api/v1/insight/das/permissions/templates',
    method: 'post',
    data
  });
}

/**
 * Update permission template
 */
export function fetchUpdatePermissionTemplate(id: number, data: Api.Das.PermissionTemplateUpdateRequest) {
  return request<Api.Das.PermissionTemplate>({
    url: `/api/v1/insight/das/permissions/templates/${id}`,
    method: 'put',
    data
  });
}

/**
 * Delete permission template
 */
export function fetchDeletePermissionTemplate(id: number) {
  return request({
    url: `/api/v1/insight/das/permissions/templates/${id}`,
    method: 'delete'
  });
}

// ==================== 角色权限管理 ====================

/**
 * Get role permissions
 */
export function fetchGetRolePermissions(role: string) {
  return request<Api.Das.RolePermission[]>({
    url: `/api/v1/insight/das/permissions/roles/${role}`,
    method: 'get'
  });
}

/**
 * Create role permission
 */
export function fetchCreateRolePermission(data: Api.Das.RolePermissionCreateRequest) {
  return request<Api.Das.RolePermission>({
    url: '/api/v1/insight/das/permissions/roles',
    method: 'post',
    data
  });
}

/**
 * Delete role permission
 */
export function fetchDeleteRolePermission(id: number) {
  return request({
    url: `/api/v1/insight/das/permissions/roles/${id}`,
    method: 'delete'
  });
}

/**
 * Get user effective permissions
 */
export function fetchGetUserEffectivePermissions(username: string) {
  return request<Api.Das.PermissionObject[]>({
    url: '/api/v1/insight/das/permissions/users',
    method: 'get',
    params: { username }
  });
}
