declare namespace Api {
  /**
   * namespace Das
   *
   * backend api module: "das" (Data Access Service)
   */
  namespace Das {
    interface Environment {
      id: number;
      name: string;
      description?: string;
    }

    interface Schema {
      id: number;
      name: string;
      instanceId: number;
      instanceName: string;
    }

    interface Table {
      id: number;
      name: string;
      schemaId: number;
      schemaName: string;
      comment?: string;
    }

    interface QueryRequest {
      sql: string;
      instanceId: number;
      schemaName: string;
      limit?: number;
    }

    interface QueryResult {
      columns: string[];
      rows: any[][];
      affectedRows?: number;
      executionTime: number;
      queryId?: string;
    }

    interface UserGrant {
      id: number;
      userId: number;
      resourceType: string;
      resourceId: number;
      permissions: string[];
    }

    interface DBDict {
      tables: TableInfo[];
    }

    interface TableInfo {
      tableName: string;
      tableComment?: string;
      columns: ColumnInfo[];
      indexes: IndexInfo[];
    }

    interface ColumnInfo {
      columnName: string;
      dataType: string;
      isNullable: boolean;
      columnDefault?: string;
      columnComment?: string;
      isPrimaryKey: boolean;
    }

    interface IndexInfo {
      indexName: string;
      columnNames: string[];
      isUnique: boolean;
    }

    interface History {
      ID: number;
      CreatedAt: string;
      UpdatedAt: string;
      DeletedAt: string | null;
      username: string;
      instance_id: string;
      schema: string;
      sql: string;
      duration: number;
      row_count: number;
      error: string;
      sqltext?: string;
      created_at?: string;
      return_rows?: number;
      status?: string;
      tables?: string;
    }

    interface HistoryListResponse {
      list: History[];
      total: number;
    }

    interface Favorite {
      ID: number;
      CreatedAt: string;
      UpdatedAt: string;
      DeletedAt: string | null;
      username: string;
      title: string;
      sql: string;
      id?: number;
      sqltext?: string;
      created_at?: string;
    }

    interface CreateFavoriteRequest {
      title: string;
      sql: string;
    }

    interface UpdateFavoriteRequest {
      id: number;
      title: string;
      sql: string;
    }

    interface SchemaPermission {
      id: number;
      username: string;
      schema: string;
      instance_id: string;
      created_at: string;
      updated_at: string;
    }

    interface TablePermission {
      id: number;
      username: string;
      schema: string;
      table: string;
      instance_id: string;
      rule: 'allow' | 'deny';
      created_at: string;
      updated_at: string;
    }

    interface UserPermissionsResponse {
      schema_permissions: SchemaPermission[];
      table_permissions: TablePermission[];
    }

    interface GrantSchemaPermissionRequest {
      username: string;
      instance_id: string;
      schema: string;
    }

    interface GrantTablePermissionRequest {
      username: string;
      instance_id: string;
      schema: string;
      table: string;
      rule?: 'allow' | 'deny';
    }

    // 保留旧的类型定义以兼容现有代码
    interface SchemaGrant {
      id: number;
      userId: number;
      userName: string;
      schemaId: number;
      schemaName: string;
      permissions: string[];
    }

    interface CreateSchemaGrantRequest {
      userId: number;
      schemaId: number;
      permissions: string[];
    }

    interface TableGrant {
      id: number;
      userId: number;
      userName: string;
      tableId: number;
      tableName: string;
      permissions: string[];
    }

    interface CreateTableGrantRequest {
      userId: number;
      tableId: number;
      permissions: string[];
    }

    interface Instance {
      id: number;
      name: string;
      host: string;
      port: number;
      dbType: string;
      environment: string;
    }

    // ==================== 权限模板 ====================
    interface PermissionObject {
      instance_id: string;
      schema: string;
      table?: string;
    }

    interface PermissionTemplate {
      id: number;
      name: string;
      description: string;
      permissions: PermissionObject[];
      created_at: string;
      updated_at: string;
    }

    interface PermissionTemplateCreateRequest {
      name: string;
      description?: string;
      permissions: PermissionObject[];
    }

    interface PermissionTemplateUpdateRequest {
      name: string;
      description?: string;
      permissions: PermissionObject[];
    }

    // ==================== 权限组 ====================
    interface DatabaseObject {
      instance_id: string;
      schema: string;
    }

    // ==================== 角色权限 ====================
    interface RolePermission {
      id: number;
      role: string;
      permission_type: 'object' | 'template';
      permission_id: number;
      instance_id?: string;
      schema?: string;
      table?: string;
      template_name?: string; // 权限模板名称（当 permission_type 为 template 时）
      created_at: string;
      updated_at: string;
    }

    interface RolePermissionCreateRequest {
      role: string;
      permission_type: 'object' | 'template';
      permission_id: number;
      instance_id?: string;
      schema?: string;
      table?: string;
    }

    // ==================== 用户权限（与角色同构：object/template，无 rule）====================
    interface UserPermission {
      id: number;
      username: string;
      permission_type: 'object' | 'template';
      permission_id: number;
      instance_id?: string;
      schema?: string;
      table?: string;
      created_at: string;
      updated_at: string;
    }

    interface UserPermissionCreateRequest {
      username: string;
      permission_type: 'object' | 'template';
      permission_id: number;
      instance_id?: string;
      schema?: string;
      table?: string;
    }
  }
}
