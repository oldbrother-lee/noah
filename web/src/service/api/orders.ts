import { request, requestRaw } from '../request';

/** Orders API - 新路由: /api/v1/insight/... */

/**
 * Get environments for orders
 */
export function fetchOrdersEnvironments(params?: Record<string, any>) {
  return request<Api.Orders.Environment[]>({
    url: '/api/v1/insight/environments',
    method: 'get',
    params
  });
}

/**
 * Get instances (dbconfigs) for specified environment
 */
export function fetchOrdersInstances(params?: Record<string, any>) {
  // 提交工单时，只返回"工单"类型的实例
  const queryParams = {
    ...params,
    use_type: '工单' // 固定为工单类型
  };
  return request<Api.Orders.Instance[]>({
    url: '/api/v1/insight/dbconfigs',
    method: 'get',
    params: queryParams
  });
}

/**
 * Get schemas for specified instance
 */
export function fetchOrdersSchemas(params?: Record<string, any>) {
  return request<Api.Orders.Schema[]>({
    url: `/api/v1/insight/dbconfigs/${params?.instance_id}/schemas`,
    method: 'get'
  });
}

/**
 * Get tables for order scenario (without DAS permission check)
 * @param params - { instance_id, schema }
 */
export function fetchOrderTables(params: { instance_id: string; schema: string }) {
  return request<any[]>({
    url: `/api/v1/insight/orders/tables/${params.instance_id}/${params.schema}`,
    method: 'get'
  });
}

/**
 * Get users for review/audit/cc
 */
export function fetchOrdersUsers(params?: Record<string, any>) {
  return request<Api.Orders.User[]>({
    url: '/api/v1/admin/users',
    method: 'get',
    params: {
      page: 1,
      pageSize: 1000, // 获取足够多的用户供选择
      ...params
    }
  });
}

/**
 * Syntax check for SQL (SQL审核)
 */
export function fetchSyntaxCheck(data: Api.Orders.SyntaxCheckRequest) {
  return request<Api.Orders.SyntaxCheckResult>({
    url: '/api/v1/insight/inspect/sql',
    method: 'post',
    data,
    timeout: 30 * 60 * 1000
  });
}

/**
 * DDL 预检：检查 ALTER TABLE 涉及的表是否有主键或唯一键（语法检查时调用，执行方式为 gh-ost 时）
 */
export function fetchCheckDDL(data: { instance_id: string; schema: string; content: string }) {
  return request<{ tables_without_pk_or_uk: string[] }>({
    url: '/api/v1/insight/orders/check-ddl',
    method: 'post',
    data
  });
}

/**
 * Create order
 */
export function fetchCreateOrder(data: Api.Orders.CreateOrderRequest) {
  return request<Api.Orders.Order>({
    url: '/api/v1/insight/orders',
    method: 'post',
    data
  });
}

/**
 * Get orders list
 */
export function fetchOrdersList(params?: Record<string, any>) {
  return request<Api.Orders.OrdersList>({
    url: '/api/v1/insight/orders',
    method: 'get',
    params
  });
}

/**
 * Get my orders list
 */
export function fetchMyOrdersList(params?: Record<string, any>) {
  return request<Api.Orders.OrdersList>({
    url: '/api/v1/insight/orders/my',
    method: 'get',
    params
  });
}

/**
 * Get order detail
 */
export function fetchOrderDetail(id: string) {
  return request<Api.Orders.OrderDetail>({
    url: `/api/v1/insight/orders/${id}`,
    method: 'get',
    timeout: 120 * 1000
  });
}

/**
 * Get operation logs
 */
export function fetchOpLogs(params?: Record<string, any>) {
  return request<Api.Orders.OpLog[]>({
    url: `/api/v1/insight/orders/${params?.order_id}/logs`,
    method: 'get'
  });
}

/**
 * Get gh-ost progress from Redis cache
 */
export function fetchGhostProgress(orderId: string) {
  return request<{
    percent?: number;
    current?: number;
    total?: number;
    eta?: string;
    operation?: string;
  } | null>({
    url: `/api/v1/insight/orders/${orderId}/ghost-progress`,
    method: 'get'
  });
}

/**
 * 更新工单进度 API
 */
export function fetchUpdateOrderProgress(data: { order_id: string; progress: string; remark?: string }) {
  return request({
    url: '/api/v1/insight/orders/progress',
    method: 'put',
    timeout: 120 * 1000,
    data
  });
}

/**
 * Approve order 审核 API
 */
export function fetchApproveOrder(data: {
  order_id: string;
  status: 'pass' | 'reject';
  msg?: string;
  ghost_ok_to_drop_table?: boolean;
}) {
  return request({
    url: '/api/v1/insight/orders/approve',
    method: 'post',
    data
  });
}

/**
 * Feedback order 反馈 API
 */
export function fetchFeedbackOrder(data: Api.Orders.FeedbackOrderRequest) {
  return fetchUpdateOrderProgress({
    order_id: data.order_id,
    progress: '已驳回',
    remark: data.remark
  });
}

/**
 * Review order 复核 API
 */
export function fetchReviewOrder(data: Api.Orders.ReviewOrderRequest) {
  return fetchUpdateOrderProgress({
    order_id: data.order_id,
    progress: '已复核',
    remark: data.remark
  });
}

/**
 * Close order 关闭 API
 */
export function fetchCloseOrder(data: Api.Orders.CloseOrderRequest) {
  return fetchUpdateOrderProgress({
    order_id: data.order_id,
    progress: '已关闭',
    remark: data.remark
  });
}

/**
 * Update order schedule time
 */
export function fetchUpdateOrderSchedule(data: { order_id: string; schedule_time: string }) {
  return request({
    url: '/api/v1/insight/orders/progress',
    method: 'put',
    data
  });
}

/**
 * Hook order (创建关联工单)
 */
export function fetchHookOrder(data: Api.Orders.HookOrderRequest) {
  return request({
    url: '/api/v1/insight/orders',
    method: 'post',
    data
  });
}

/**
 * Generate tasks 生成任务 API
 */
export function fetchGenerateTasks(data: Api.Orders.GenerateTasksRequest) {
  return request<Api.Orders.Task[]>({
    url: '/api/v1/insight/orders',
    method: 'post',
    data
  });
}

/**
 * 获取任务列表 API
 */
export function fetchTasks(params: { order_id: string }) {
  return request<Api.Orders.Task[]>({
    url: `/api/v1/insight/orders/${params.order_id}/tasks`,
    method: 'get'
  });
}

/**
 * 获取任务回滚SQL API（按需加载）
 */
export function fetchTaskRollbackSQL(params: { order_id: string; task_id: string }) {
  return request<{ rollback_sql: string }>({
    url: `/api/v1/insight/orders/${params.order_id}/tasks/${params.task_id}/rollback-sql`,
    method: 'get'
  });
}

/**
 *  预览任务 API
 */
export function fetchPreviewTasks(params?: Record<string, any>) {
  return request<Api.Orders.TaskPreview>({
    url: `/api/v1/insight/orders/${params?.order_id}/tasks`,
    method: 'get'
  });
}

/**
 * 更新任务进度 API
 */
export function fetchUpdateTaskProgress(data: { task_id: string; progress: string }) {
  return request({
    url: '/api/v1/insight/orders/tasks/progress',
    method: 'put',
    data
  });
}

/**
 * 执行单个任务 API
 */
export function fetchExecuteSingleTask(data: Api.Orders.ExecuteTaskRequest) {
  return request<Api.Orders.TaskResult>({
    url: '/api/v1/insight/orders/tasks/execute',
    method: 'post',
    data
  });
}

/**
 * 执行所有任务 API
 */
export function fetchExecuteAllTasks(data: Api.Orders.ExecuteAllTasksRequest) {
  return requestRaw<any>({
    url: '/api/v1/insight/orders/tasks/execute',
    method: 'post',
    data,
    timeout: 24 * 60 * 60 * 1000
  });
}

/**
 * 下载导出文件 API
 */
export function fetchDownloadExportFile(taskId: string | number) {
  return request<Blob>({
    url: `/api/v1/insight/orders/download/${taskId}`,
    method: 'get',
    responseType: 'blob'
  });
}

/**
 * 控制 gh-ost 执行 API
 */
export function fetchControlGhost(data: {
  order_id: string;
  action: 'throttle' | 'unthrottle' | 'panic' | 'chunk-size';
  value?: number;
}) {
  return request<{ message: string }>({
    url: '/api/v1/insight/orders/ghost/control',
    method: 'post',
    data
  });
}
