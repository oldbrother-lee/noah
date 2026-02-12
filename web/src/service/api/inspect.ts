import { request } from '../request';

/** Inspect (SQL审核) API */

/**
 * Get inspect params list
 */
export function fetchGetInspectParams(params?: { page?: number; page_size?: number; remark?: string }) {
  return request<Api.Inspect.InspectParamListResponse>({
    url: '/api/v1/insight/inspect/params',
    method: 'get',
    params
  });
}

/**
 * Get inspect param by id
 */
export function fetchGetInspectParam(id: number) {
  return request<Api.Inspect.InspectParam>({
    url: `/api/v1/insight/inspect/params/${id}`,
    method: 'get'
  });
}

/**
 * Create inspect param
 */
export function fetchCreateInspectParam(data: Api.Inspect.InspectParamCreateRequest) {
  return request<Api.Inspect.InspectParam>({
    url: '/api/v1/insight/inspect/params',
    method: 'post',
    data
  });
}

/**
 * Update inspect param
 */
export function fetchUpdateInspectParam(id: number, data: Api.Inspect.InspectParamUpdateRequest) {
  return request<void>({
    url: `/api/v1/insight/inspect/params/${id}`,
    method: 'put',
    data
  });
}

/**
 * Delete inspect param
 */
export function fetchDeleteInspectParam(id: number) {
  return request<void>({
    url: `/api/v1/insight/inspect/params/${id}`,
    method: 'delete'
  });
}

/**
 * Get default inspect params
 */
export function fetchGetDefaultInspectParams() {
  return request<Api.Inspect.DefaultInspectParams>({
    url: '/api/v1/insight/inspect/params/default',
    method: 'get'
  });
}
