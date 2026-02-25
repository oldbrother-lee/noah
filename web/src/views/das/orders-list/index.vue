<script setup lang="tsx">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { NCard, NDataTable, NSwitch, NTag } from 'naive-ui';
import { fetchOrdersList } from '@/service/api/orders';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { useRouterPush } from '@/hooks/common/router';
import { $t } from '@/locales';
import OrderSearch from './modules/order-search.vue';

const appStore = useAppStore();
const { routerPushByKey } = useRouterPush();

/**
 * Order search parameters
 * 工单搜索参数
 */
const initialSearchParams: Api.Orders.OrderSearchParams = {
  current: 1,
  size: 10,
  environment: null,
  status: null,
  search: null,
  only_my_orders: 0
};

const onlyMyOrders = ref(false);

/**
 * Handle "Only My Orders" switch change
 * 处理"只看我的"开关变化
 * @param val boolean value
 */
function handleMyOrdersChange(val: boolean) {
  updateSearchParams({
    only_my_orders: val ? 1 : 0
  });
  getDataByPage();
}

/**
 * Get progress tag color
 * 获取进度标签颜色
 * @param progress Progress status string
 * @returns NaiveUI theme color
 */
function getProgressTagColor(progress: string): NaiveUI.ThemeColor {
  // 与工单详情页面的 progressTypeMap 保持一致
  switch (progress) {
    case '待审核':
      return 'warning';
    case '已批准':
      return 'info'; // 蓝色
    case '已驳回':
      return 'error';
    case '执行中':
      return 'info';
    case '已完成':
      return 'success';
    case '已关闭':
      return 'default';
    // 兼容旧的状态值（如果有的话）
    case '待审批':
    case '待执行':
      return 'warning';
    case '已失败':
      return 'error';
    default:
      return 'default';
  }
}

/**
 * Handle row click
 * 处理行点击
 * @param row Order data
 */
function handleRowClick(row: Api.Orders.Order) {
  routerPushByKey('das_orders-detail', { params: { id: row.order_id } });
}

const rowProps = (row: Api.Orders.Order) => {
  return {
    style: 'cursor: pointer;',
    onClick: () => handleRowClick(row)
  };
};

/**
 * Reset search parameters
 * 重置搜索参数
 */
function handleReset() {
  onlyMyOrders.value = false;
  updateSearchParams({
    only_my_orders: 0,
    environment: null,
    status: null,
    search: null
  });
  getDataByPage();
}

/**
 * Table configuration
 * 表格配置
 */
const { columns, data, getData, getDataByPage, loading, mobilePagination, searchParams, updateSearchParams } = useTable({
  apiFn: fetchOrdersList,
  apiParams: initialSearchParams,
  showTotal: true,
  pagination: {
    pageSize: 10,
    pageSizes: [10, 20, 50, 100],
    showQuickJumper: true
  },
  transformer: res => {
    // 新框架响应格式: { code, message, data: { list, total } }
    const responseData = (res as any)?.data || {};
    const records = responseData.list || [];
    const total = responseData.total || 0;
    const current = (searchParams as any).current || 1;
    const size = (searchParams as any).size || 10;

    // 格式化时间函数：将时间格式化为 "2026-02-02 10:10:10" 格式
    const formatDateTime = (dateStr: string | null | undefined): string => {
      if (!dateStr) return '';
      try {
        const date = new Date(dateStr);
        if (isNaN(date.getTime())) return dateStr; // 如果解析失败，返回原字符串
        
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
      } catch (e) {
        return dateStr;
      }
    };

    const recordsWithIndex = records.map((item: any, index: number) => {
      return {
        ...item,
        index: (current - 1) * size + index + 1,
        order_title: item.title || '',
        instance: item.instance_name || item.instance_id || '',
        environment: item.environment_name || item.environment || '',
        applicant: item.applicant || '',
        applicant_nickname: item.applicant_nickname || '', // 保留申请人昵称字段
        created_at: item.CreatedAt ? new Date(item.CreatedAt).toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit'
        }) : '',
        schedule_time: formatDateTime(item.schedule_time || item.ScheduleTime)
      };
    });

    return {
      data: recordsWithIndex,
      pageNum: current,
      pageSize: size,
      total
    };
  },
  columns: () => [
    {
      key: 'order_id',
      title: '工单编号',
      align: 'center',
      width: 320,
      ellipsis: { tooltip: true }
    },
    {
      key: 'progress',
      title: '进度',
      align: 'center',
      width: 100,
      render: row => {
        const color = getProgressTagColor(row.progress);
        return (
          <NTag type={color} size="small">
            {row.progress}
          </NTag>
        );
      }
    },
    {
      key: 'order_title',
      title: '工单标题',
      align: 'center',
      width: 180,
      ellipsis: {
        tooltip: true
      }
    },
    {
      key: 'execution_mode',
      title: '执行方式',
      align: 'center',
      width: 150,
      render: row => {
        if (row.schedule_time) {
          return (
            <div style="display: flex; flex-direction: column; align-items: center; font-size: 12px;">
              <NTag type="info" size="small" style="margin-bottom: 4px;">
                定时执行
              </NTag>
              <span style="color: #666;">{row.schedule_time}</span>
            </div>
          );
        }
        return (
          <NTag type="default" size="small">
            立即执行
          </NTag>
        );
      }
    },
    {
      key: 'applicant',
      title: '申请人',
      align: 'center',
      width: 100,
      render: row => {
        // 优先显示昵称，如果没有昵称则显示用户名
        const displayName = row.applicant_nickname || row.applicant || '';
        return <span>{displayName}</span>;
      }
    },
    {
      key: 'sql_type',
      title: 'SQL类型',
      align: 'center',
      width: 100
    },
    {
      key: 'environment',
      title: '环境',
      align: 'center',
      width: 100,
      render: row => {
        // 使用环境名称而不是 ID
        const envName = row.environment_name || row.environment || '';
        const tagMap: Record<string, NaiveUI.ThemeColor> = {
          test: 'primary',
          prod: 'error',
          dev: 'info'
        };
        const type = tagMap[envName.toLowerCase()] || 'default';
        return (
          <NTag type={type} size="small">
            {envName}
          </NTag>
        );
      }
    },
    {
      key: 'instance',
      title: '实例',
      align: 'center',
      width: 150,
      ellipsis: {
        tooltip: true
      }
    },
    {
      key: 'schema',
      title: '库名',
      align: 'center',
      width: 100,
      ellipsis: {
        tooltip: true
      }
    },
    {
      key: 'created_at',
      title: '创建时间',
      align: 'center',
      width: 180
    }
  ]
});

// Auto refresh timer
let refreshTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  // Auto refresh every 30 seconds
  refreshTimer = setInterval(() => {
    getData();
  }, 30000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <OrderSearch v-model:model="searchParams" @search="getDataByPage" @reset="handleReset" />
    <NCard title="工单列表" :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
      <template #header-extra>
        <div class="flex-y-center gap-12px">
          <span class="text-14px">只看我的</span>
          <NSwitch v-model:value="onlyMyOrders" size="small" @update:value="handleMyOrdersChange" />
        </div>
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        :flex-height="!appStore.isMobile"
        :scroll-x="962"
        :loading="loading"
        remote
        :row-key="row => row.order_id"
        :pagination="mobilePagination"
        :row-props="rowProps"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
