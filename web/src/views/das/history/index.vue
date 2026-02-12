<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue';
import { NButton, NTag } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { fetchCreateFavorite, fetchHistory } from '@/service/api/das';

const props = defineProps<{
  embedded?: boolean;
}>();

const emit = defineEmits<{
  (e: 'reuse', sql: string): void;
}>();

const { t } = useI18n();

// 响应式数据
const loading = ref(false);
const historyList = ref<any[]>([]);
const searchKeyword = ref('');

const pagination = reactive({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  itemCount: 0, // 总记录数
  onChange: (page: number) => {
    pagination.page = page;
    loadHistory();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
    loadHistory();
  }
});

watch(searchKeyword, () => {
  pagination.page = 1;
});

// 计算属性
const filteredHistory = computed(() => {
  if (!searchKeyword.value) {
    return historyList.value;
  }
  const keyword = searchKeyword.value.toLowerCase();
  return historyList.value.filter(
    item =>
      (item.sqltext || '').toLowerCase().includes(keyword) ||
      (item.schema || '').toLowerCase().includes(keyword) ||
      (item.tables || '').toLowerCase().includes(keyword)
  );
});

// 方法
const loadHistory = async () => {
  loading.value = true;
  try {
    const { data, error } = await fetchHistory({
      page: pagination.page,
      page_size: pagination.pageSize
    });
    if (!error && data) {
      const responseData = data as any;
      let rawList: any[] = [];
      if (Array.isArray(responseData)) {
        rawList = responseData;
      } else if (responseData.list && Array.isArray(responseData.list)) {
        rawList = responseData.list;
        // 更新分页总数
        if (responseData.total !== undefined) {
          pagination.itemCount = responseData.total;
        }
      }
      
      // 字段映射：后端字段 -> 前端字段
      historyList.value = rawList.map((item: any) => {
        // 格式化时间
        const formatDateTime = (dateStr: string | null | undefined): string => {
          if (!dateStr) return '';
          try {
            const date = new Date(dateStr);
            return date.toLocaleString('zh-CN', {
              year: 'numeric',
              month: '2-digit',
              day: '2-digit',
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit'
            });
          } catch (e) {
            return dateStr;
          }
        };
        
        return {
          ...item,
          // 字段映射
          sqltext: item.sql || item.sqltext || '', // sql -> sqltext
          created_at: formatDateTime(item.CreatedAt), // CreatedAt -> created_at (格式化)
          return_rows: item.row_count || 0, // row_count -> return_rows
          status: item.error && item.error.trim() ? 'error' : 'success', // 根据 error 判断状态
          tables: item.tables || item.schema || '', // 如果没有 tables，使用 schema
          error_msg: item.error || '' // error -> error_msg
        };
      });
    } else {
      window.$message?.error('加载历史查询失败');
    }
  } catch (error) {
    console.error('Failed to load history:', error);
    window.$message?.error('加载历史查询失败');
  } finally {
    loading.value = false;
  }
};

const reuseQuery = (history: any) => {
  if (props.embedded) {
    emit('reuse', history.sqltext);
    return;
  }
  console.log('Reuse query:', history.sqltext);
  window.$message?.success('SQL已复制到剪贴板');
  navigator.clipboard.writeText(history.sqltext);
};

const addToFavorites = async (history: any) => {
  try {
    const { error } = await fetchCreateFavorite({
      sql: history.sqltext || history.sql || '',
      title: `From History: ${history.schema || ''}`
    });
    if (!error) {
      window.$message?.success('已添加到收藏');
    } else {
      window.$message?.error('添加收藏失败');
    }
  } catch (error) {
    console.error('Failed to add to favorites:', error);
    window.$message?.error('添加收藏失败');
  }
};

const getStatusType = (status: string) => {
  switch (status) {
    case 'success':
      return 'success';
    case 'error':
      return 'error';
    case 'warning':
      return 'warning';
    default:
      return 'default';
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'success':
      return '成功';
    case 'error':
      return '失败';
    case 'warning':
      return '警告';
    default:
      return '未知';
  }
};

onMounted(() => {
  loadHistory();
});

const columns: any[] = [
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row: any) {
      // 使用映射后的 status 字段，如果没有则根据 error 判断
      const status = row.status || (row.error && row.error.trim() ? 'error' : 'success');
      const type = getStatusType(status);
      const text = getStatusText(status);
      return h(NTag, { type: type as any, size: 'small' }, { default: () => text });
    }
  },
  {
    title: '库',
    key: 'schema',
    width: 200,
    ellipsis: {
      tooltip: true
    },
    render(row: any) {
      return row.schema || '';
    }
  },
  {
    title: 'SQL',
    key: 'sqltext',
    width: 400,
    ellipsis: {
      tooltip: true
    }
  },
  { title: '耗时(ms)', key: 'duration', width: 120, ellipsis: { tooltip: true } },
  { title: '影响行数', key: 'return_rows', width: 100, ellipsis: { tooltip: true } },
  { title: '执行时间', key: 'created_at', width: 180, ellipsis: { tooltip: true } },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    fixed: 'right',
    render(row: any) {
      return h('div', { style: 'display:flex; gap:8px;' }, [
        h(
          NButton,
          {
            type: 'primary',
            size: 'tiny',
            ghost: true,
            onClick: () => reuseQuery(row)
          },
          { default: () => '执行' }
        ),
        h(
          NButton,
          {
            size: 'tiny',
            quaternary: true,
            onClick: () => addToFavorites(row)
          },
          { default: () => '收藏' }
        )
      ]);
    }
  }
];
</script>

<template>
  <div class="history-container">
    <NCard :bordered="!embedded" size="small" :content-style="{ padding: embedded ? '0' : '' }">
      <template #header>
        <NSpace justify="space-between" align="center">
          <span v-if="!embedded">历史查询</span>
          <NSpace size="small">
            <NInput v-model:value="searchKeyword" placeholder="搜索" clearable size="tiny" style="width: 180px">
              <template #prefix>
                <SvgIcon icon="carbon:search" />
              </template>
            </NInput>
            <NButton type="primary" size="tiny" ghost @click="loadHistory">
              <template #icon>
                <SvgIcon icon="carbon:renew" />
              </template>
              刷新
            </NButton>
          </NSpace>
        </NSpace>
      </template>

      <NSpin :show="loading">
        <div v-if="filteredHistory.length === 0" class="empty-state">
          <NEmpty description="暂无历史查询记录" />
        </div>
        <div v-else>
          <NDataTable
            :columns="columns"
            :data="filteredHistory"
            size="small"
            :pagination="pagination"
            :scroll-x="1200"
          />
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.history-container {
  height: 100%;
}

.history-list,
.history-item,
.sql-content,
.execution-info {
  display: none;
}
</style>
