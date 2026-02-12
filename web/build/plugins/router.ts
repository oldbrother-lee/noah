import type { RouteMeta } from 'vue-router';
import ElegantVueRouter from '@elegant-router/vue/vite';
import type { RouteKey } from '@elegant-router/types';

export function setupElegantRouter() {
  return ElegantVueRouter({
    layouts: {
      base: 'src/layouts/base-layout/index.vue',
      blank: 'src/layouts/blank-layout/index.vue'
    },
    routePathTransformer(routeName, routePath) {
      const key = routeName as RouteKey;

      if (key === 'login') {
        const modules: UnionKey.LoginModule[] = ['pwd-login', 'code-login', 'register', 'reset-pwd', 'bind-wechat'];

        const moduleReg = modules.join('|');

        return `/login/:module(${moduleReg})?`;
      }

      if (key === 'das_orders-detail') {
        return '/das/orders-detail/:id';
      }

      return routePath;
    },
    onRouteMetaGen(routeName) {
      const key = routeName as string;

      const constantRoutes = ['login', '403', '404', '500'];

      const meta: Partial<RouteMeta> = {
        title: key,
        i18nKey: `route.${key}` as App.I18n.I18nKey
      };

      if (constantRoutes.includes(key)) {
        meta.constant = true;
      }

      if (key === 'das_orders-detail') {
        meta.hideInMenu = true;
        meta.activeMenu = 'das_orders-list';
      }

      // Add icon and order for specific routes
      if (key === 'home') {
        meta.icon = 'mdi:monitor-dashboard';
        meta.order = 1;
      }

      if (key === 'das') {
        // 顶层“数据库服务”显示为菜单，并设置图标与顺序
        meta.icon = 'mdi:database-search';
        meta.order = 2;
      }

      if (key === 'manage') {
        meta.hideInMenu = true;
      }

      // 隐藏不需要展示的 DAS 子菜单
      if (key === 'das_favorite' || key === 'das_history' || key === 'das_orders_commit') {
        meta.hideInMenu = true;
      }

      // 为工单类型设置排序，确保菜单顺序为 DDL/DML/导出
      if (key === 'das_orders') {
        meta.order = 3;
      }
      // 旧的 commit_* 排序保留无影响，但新增直接子路由排序
      if (key === 'das_orders_ddl') {
        meta.order = 1;
      }
      if (key === 'das_orders_dml') {
        meta.order = 2;
      }
      if (key === 'das_orders_export') {
        meta.order = 3;
      }

      // 数据库管理菜单配置
      if (key === 'system_database') {
        meta.icon = 'mdi:database-cog';
        meta.order = 1;
      }
      if (key === 'system_database_environment') {
        meta.icon = 'mdi:server-network';
        meta.order = 1;
      }
      if (key === 'system_database_config') {
        meta.icon = 'mdi:database-settings';
        meta.order = 2;
      }
      // 权限管理相关菜单配置
      if (key === 'system_database_permission') {
        meta.icon = 'mdi:shield-account';
        meta.order = 3;
      }
      // 隐藏权限相关的子菜单，因为它们会作为权限管理的子菜单显示
      if (
        key === 'system_database_permission-group' ||
        key === 'system_database_permission-template' ||
        key === 'system_database_role-permission'
      ) {
        meta.hideInMenu = true;
      }

      return meta;
    }
  });
}
