declare namespace Api {
  /**
   * namespace Auth
   *
   * backend api module: "auth"
   */
  namespace Auth {
    interface LoginToken {
      accessToken: string;
      token?: string; // 兼容字段，如果后端返回 accessToken，前端会映射到 token
    }

    interface UserInfo {
      userId: string;
      userName: string;
      /** 登录用户名，与工单执行人列表中的值一致，用于权限判断 */
      username?: string;
      roles: string[];
      buttons: string[];
    }
  }
}
