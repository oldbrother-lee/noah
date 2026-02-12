import { request } from '../request';

/** Users Management API */

/**
 * Get users list
 */
export function fetchUsers(params?: Record<string, any>) {
  return request<Api.Users.UsersList>({
    url: '/api/v1/admin/users',
    method: 'get',
    params
  });
}

/**
 * Create user
 */
export function fetchCreateUser(data: Api.Users.CreateUserRequest) {
  return request<Api.Users.User>({
    url: '/api/v1/admin/users',
    method: 'post',
    data
  });
}

/**
 * Update user
 */
export function fetchUpdateUser(data: Api.Users.UpdateUserRequest) {
  return request<Api.Users.User>({
    url: `/api/v1/admin/users/${data.uid}`,
    method: 'put',
    data
  });
}

/**
 * Delete user
 */
export function fetchDeleteUser(uid: string | number) {
  return request({
    url: `/api/v1/admin/users/${uid}`,
    method: 'delete'
  });
}

/**
 * Change user password
 */
export function fetchChangeUserPassword(data: Api.Users.ChangePasswordRequest) {
  return request({
    url: '/api/v1/admin/users/password',
    method: 'put',
    data
  });
}

/**
 * Get roles list
 */
export function fetchRoles(params?: Record<string, any>) {
  return request<Api.Users.Role[]>({
    url: '/api/v1/admin/roles',
    method: 'get',
    params
  });
}

/**
 * Create role
 */
export function fetchCreateRole(data: Api.Users.CreateRoleRequest) {
  return request<Api.Users.Role>({
    url: '/api/v1/admin/roles',
    method: 'post',
    data
  });
}

/**
 * Update role
 */
export function fetchUpdateRole(data: Api.Users.UpdateRoleRequest) {
  return request<Api.Users.Role>({
    url: `/api/v1/admin/roles/${data.id}`,
    method: 'put',
    data
  });
}

/**
 * Delete role
 */
export function fetchDeleteRole(id: string | number) {
  return request({
    url: `/api/v1/admin/roles/${id}`,
    method: 'delete'
  });
}

/**
 * Get organizations list
 */
export function fetchOrganizations(params?: Record<string, any>) {
  return request<Api.Users.Organization[]>({
    url: '/api/v1/admin/organizations',
    method: 'get',
    params
  });
}

/**
 * Create root organization
 */
export function fetchCreateRootOrganization(data: Api.Users.CreateOrganizationRequest) {
  return request<Api.Users.Organization>({
    url: '/api/v1/admin/organizations/root',
    method: 'post',
    data
  });
}

/**
 * Create child organization
 */
export function fetchCreateChildOrganization(data: Api.Users.CreateChildOrganizationRequest) {
  return request<Api.Users.Organization>({
    url: '/api/v1/admin/organizations/child',
    method: 'post',
    data
  });
}

/**
 * Update organization
 */
export function fetchUpdateOrganization(data: Api.Users.UpdateOrganizationRequest) {
  return request<Api.Users.Organization>({
    url: `/api/v1/admin/organizations/${data.id}`,
    method: 'put',
    data
  });
}

/**
 * Delete organization
 */
export function fetchDeleteOrganization(id: string | number) {
  return request({
    url: `/api/v1/admin/organizations/${id}`,
    method: 'delete'
  });
}

/**
 * Get organization users
 */
export function fetchOrganizationUsers(params?: Record<string, any>) {
  return request<Api.Users.OrganizationUser[]>({
    url: '/api/v1/admin/organizations/users',
    method: 'get',
    params
  });
}

/**
 * Bind organization users
 */
export function fetchBindOrganizationUsers(data: Api.Users.BindOrganizationUsersRequest) {
  return request({
    url: '/api/v1/admin/organizations/users/bind',
    method: 'post',
    data
  });
}

/**
 * Delete organization users
 */
export function fetchDeleteOrganizationUsers(data: Api.Users.DeleteOrganizationUsersRequest) {
  return request({
    url: '/api/v1/admin/organizations/users',
    method: 'delete',
    data
  });
}

/**
 * Admin: Get DB config
 */
export function fetchAdminDBConfig(params?: Record<string, any>) {
  return request<Api.Users.DBConfig[]>({
    url: '/api/v1/admin/dbconfig',
    method: 'get',
    params
  });
}

/**
 * Admin: Create DB config
 */
export function fetchAdminCreateDBConfig(data: Api.Users.CreateDBConfigRequest) {
  return request<Api.Users.DBConfig>({
    url: '/api/v1/admin/dbconfig',
    method: 'post',
    data
  });
}

/**
 * Admin: Update DB config
 */
export function fetchAdminUpdateDBConfig(data: Api.Users.UpdateDBConfigRequest) {
  return request<Api.Users.DBConfig>({
    url: `/api/v1/admin/dbconfig/${data.id}`,
    method: 'put',
    data
  });
}

/**
 * Admin: Delete DB config
 */
export function fetchAdminDeleteDBConfig(id: string | number) {
  return request({
    url: `/api/v1/admin/dbconfig/${id}`,
    method: 'delete'
  });
}
