declare namespace Api {
  /**
   * namespace Users
   *
   * backend api module: "users"
   */
  namespace Users {
    interface User {
      id: number;
      username: string;
      realName: string;
      email: string;
      phone?: string;
      status: Api.Common.EnableStatus;
      roles: Role[];
      organizations: Organization[];
      createTime: string;
    }

    interface UsersList extends Api.Common.PaginatingQueryRecord<User> {}

    interface CreateUserRequest {
      username: string;
      realName: string;
      email: string;
      phone?: string;
      password: string;
      roleIds: number[];
      organizationIds: number[];
    }

    interface UpdateUserRequest {
      uid: number;
      username: string;
      realName: string;
      email: string;
      phone?: string;
      status: Api.Common.EnableStatus;
      roleIds: number[];
      organizationIds: number[];
    }

    interface ChangePasswordRequest {
      uid: number;
      newPassword: string;
    }

    interface Role {
      id: number;
      name: string;
      description?: string;
      permissions: string[];
      status: Api.Common.EnableStatus;
      createTime: string;
    }

    interface CreateRoleRequest {
      name: string;
      description?: string;
      permissions: string[];
    }

    interface UpdateRoleRequest {
      id: number;
      name: string;
      description?: string;
      permissions: string[];
      status: Api.Common.EnableStatus;
    }

    interface Organization {
      id: number;
      name: string;
      description?: string;
      parentId?: number;
      level: number;
      status: Api.Common.EnableStatus;
      createTime: string;
      children?: Organization[];
    }

    interface CreateOrganizationRequest {
      name: string;
      description?: string;
      permissions: string[];
    }

    interface CreateChildOrganizationRequest {
      name: string;
      description?: string;
      parentId: number;
    }

    interface UpdateOrganizationRequest {
      id: number;
      name: string;
      description?: string;
      status: Api.Common.EnableStatus;
    }

    interface OrganizationUser {
      id: number;
      organizationId: number;
      organizationName: string;
      userId: number;
      username: string;
      realName: string;
    }

    interface BindOrganizationUsersRequest {
      organizationId: number;
      userIds: number[];
    }

    interface DeleteOrganizationUsersRequest {
      organizationId: number;
      userIds: number[];
    }

    interface DBConfig {
      id: number;
      name: string;
      host: string;
      port: number;
      username: string;
      dbType: string;
      environment: string;
      status: Api.Common.EnableStatus;
      createTime: string;
    }

    interface CreateDBConfigRequest {
      name: string;
      host: string;
      port: number;
      username: string;
      password: string;
      dbType: string;
      environment: string;
    }

    interface UpdateDBConfigRequest {
      id: number;
      name: string;
      host: string;
      port: number;
      username: string;
      password?: string;
      dbType: string;
      environment: string;
      status: Api.Common.EnableStatus;
    }
  }
}
