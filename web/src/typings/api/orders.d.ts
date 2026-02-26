declare namespace Api {
  /**
   * namespace Orders
   *
   * backend api module: "orders"
   */
  namespace Orders {
    interface Environment {
      id: number;
      name: string;
      description?: string;
    }

    interface Instance {
      id: number;
      name: string;
      host: string;
      port: number;
      dbType: string;
    }

    interface Schema {
      id: number;
      name: string;
      instanceId: number;
      instanceName: string;
    }

    interface User {
      id: number;
      username: string;
      realName: string;
      email: string;
      role: string;
    }

    interface SyntaxCheckRequest {
      sql: string;
      instanceId: number;
      dbType: string;
    }

    interface SyntaxCheckResult {
      data: {
        summary: string[] | null;
        level: string;
        affected_rows: number;
        type: string;
        finger_id: string;
        query: string;
      }[];
      /** 0: 通过, 1: 失败 */
      status: number;
    }

    interface CreateOrderRequest {
      title: string;
      description?: string;
      sql: string;
      instanceId: number;
      schemaName: string;
      orderType: string;
      reviewers: number[];
      auditors: number[];
      ccUsers?: number[];
      executeTime?: string;
    }

    interface Order {
      order_id: string;
      order_title: string;
      description?: string;
      sql: string;
      progress: string;
      sql_type: string;
      created_at: string;
      applicant: string;
      instance: string;
      schema: string;
      environment: string;
      execution_mode?: string;
      schedule_time?: string;
    }

    type CommonSearchParams = Pick<Api.Common.PaginatingCommonParams, 'current' | 'size'>;

    interface OrderSearchParams extends CommonSearchParams {
      environment?: number | null;
      progress?: string | null;
      search?: string | null;
      only_my_orders?: number;
    }

    interface OrdersList extends Api.Common.PaginatingQueryRecord<Order> {}

    interface OrderDetail extends Order {
      reviewers: User[];
      auditors: User[];
      ccUsers: User[];
      tasks: Task[];
      opLogs: OpLog[];
    }

    interface OpLog {
      id: number;
      orderId: number;
      action: string;
      operator: string;
      operateTime: string;
      comment?: string;
    }

    interface ApproveOrderRequest {
      order_id: string;
      status: 'pass' | 'reject';
      msg?: string;
    }

    interface FeedbackOrderRequest {
      orderId: number;
      comment: string;
    }

    interface ReviewOrderRequest {
      orderId: number;
      approved: boolean;
      comment?: string;
    }

    interface CloseOrderRequest {
      orderId: number;
      reason: string;
    }

    interface HookOrderRequest {
      orderId: number;
      hookType: string;
      hookUrl: string;
    }

    interface GenerateTasksRequest {
      orderId: number;
    }

    interface Task {
      id: number;
      orderId: number;
      sql: string;
      status: string;
      executeTime?: string;
      duration?: number;
      affectedRows?: number;
      errorMessage?: string;
    }

    interface TaskPreview {
      tasks: Task[];
      totalCount: number;
    }

    interface ExecuteTaskRequest {
      task_id: string;
      /** DDL 执行时：gh-ost 成功后是否自动删除旧表 */
      ghost_ok_to_drop_table?: boolean;
    }

    interface ExecuteAllTasksRequest {
      order_id: string;
      /** DDL 执行时：gh-ost 成功后是否自动删除旧表 */
      ghost_ok_to_drop_table?: boolean;
    }

    interface TaskResult {
      success: boolean;
      message?: string;
      affectedRows?: number;
      executionTime?: number;
    }
  }
}
