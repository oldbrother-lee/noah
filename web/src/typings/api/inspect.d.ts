declare namespace Api {
  namespace Inspect {
    /**
     * 审核参数
     */
    interface InspectParam {
      id: number;
      params: Record<string, any>; // JSON 格式的审核参数
      remark: string; // 备注
      createdAt: string;
      updatedAt: string;
    }

    /**
     * 审核参数列表响应
     */
    interface InspectParamListResponse {
      list: InspectParam[];
      total: number;
    }

    /**
     * 创建审核参数请求
     */
    interface InspectParamCreateRequest {
      params: Record<string, any>;
      remark: string;
    }

    /**
     * 更新审核参数请求
     */
    interface InspectParamUpdateRequest {
      params: Record<string, any>;
      remark?: string;
    }

    /**
     * 默认审核参数响应
     */
    interface DefaultInspectParams {
      [key: string]: any;
    }
  }
}
