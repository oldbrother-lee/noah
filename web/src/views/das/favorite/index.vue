<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { NButton } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { fetchDeleteFavorite, fetchFavorites } from '@/service/api/das';

const props = defineProps<{
  embedded?: boolean;
}>();

const emit = defineEmits<{
  (e: 'reuse', sql: string): void;
}>();

const { t } = useI18n();

// 响应式数据
const loading = ref(false);
const favorites = ref<any[]>([]);

// 方法
const loadFavorites = async () => {
  loading.value = true;
  try {
    const { data, error } = await fetchFavorites();
    if (!error && data) {
      // 字段映射：后端字段 -> 前端字段
      const rawList = Array.isArray(data) ? data : [];
      favorites.value = rawList.map((item: any) => {
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
          id: item.ID || item.id, // ID -> id
          sqltext: item.sql || item.sqltext || '', // sql -> sqltext
          created_at: formatDateTime(item.CreatedAt) // CreatedAt -> created_at (格式化)
        };
      });
    } else {
      window.$message?.error('加载收藏SQL失败');
    }
  } catch (error) {
    console.error('Failed to load favorites:', error);
    window.$message?.error('加载收藏SQL失败');
  } finally {
    loading.value = false;
  }
};

const useFavorite = (favorite: any) => {
  if (props.embedded) {
    emit('reuse', favorite.sqltext);
    return;
  }
  console.log('Use favorite SQL:', favorite.sqltext);
  window.$message?.success('SQL已复制到剪贴板');
  navigator.clipboard.writeText(favorite.sqltext);
};

const deleteFavorite = async (id: number) => {
  try {
    const { error } = await fetchDeleteFavorite(id);
    if (!error) {
      favorites.value = favorites.value.filter(item => item.id !== id);
      window.$message?.success('删除成功');
    } else {
      window.$message?.error('删除失败');
    }
  } catch (error) {
    console.error('Failed to delete favorite:', error);
    window.$message?.error('删除失败');
  }
};

onMounted(() => {
  loadFavorites();
});

const columns: any[] = [
  { title: 'SQL', key: 'sqltext', width: 400, ellipsis: { tooltip: true } },
  { title: '描述', key: 'title', width: 200, ellipsis: { tooltip: true } },
  { title: '创建时间', key: 'created_at', width: 180, ellipsis: { tooltip: true } },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    fixed: 'right',
    render(row: any) {
      return h('div', { style: 'display:flex; gap:8px;' }, [
        h(
          NButton,
          {
            type: 'primary',
            size: 'tiny',
            ghost: true,
            onClick: () => useFavorite(row)
          },
          { default: () => '执行' }
        ),
        h(
          NButton,
          {
            type: 'error',
            size: 'tiny',
            quaternary: true,
            onClick: () => deleteFavorite(row.id)
          },
          { default: () => '删除' }
        )
      ]);
    }
  }
];
</script>

<template>
  <div class="favorite-container">
    <NCard :bordered="!embedded" size="small" :content-style="{ padding: embedded ? '0' : '' }">
      <template v-if="!embedded" #header>
        <span>收藏SQL</span>
      </template>

      <NSpin :show="loading">
        <div v-if="favorites.length === 0" class="empty-state">
          <NEmpty description="暂无收藏SQL" />
        </div>
        <div v-else>
          <NDataTable
            :columns="columns"
            :data="favorites"
            size="small"
            :scroll-x="1000"
            :style="{ height: '300px' }"
            flex-height
          />
        </div>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.favorite-container {
  height: 100%;
}

.favorites-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.favorite-item {
  border: 1px solid var(--border-color);
}

.favorite-desc {
  font-weight: 500;
}

.favorite-time {
  font-size: 12px;
  color: var(--text-color-3);
}

.sql-content {
  margin: 8px 0;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
}
</style>
