<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable } from 'naive-ui';
import { fetchGetEnvironments, fetchDeleteEnvironment } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import EnvironmentOperateModal from './modules/environment-operate-modal.vue';

defineOptions({
  name: 'SystemDatabaseEnvironment'
});

const appStore = useAppStore();

const { columns, columnChecks, data, loading, pagination, getData, getDataByPage } = useTable({
  apiFn: () => fetchGetEnvironments(),
  transformer: res => {
    // res 可能是 { data: [...], error } 或直接是 [...]
    const responseData = (res as any)?.data || res;
    let list: Api.Admin.Environment[] = [];
    if (Array.isArray(responseData)) {
      list = responseData;
    } else if (responseData && Array.isArray(responseData.list)) {
      list = responseData.list;
    }
    
    return {
      data: list.map((item: any, index: number) => ({
        ...item,
        index: index + 1
      })),
      pageNum: 1,
      pageSize: list.length,
      total: list.length
    };
  },
  columns: () => [
    {
      type: 'selection',
      align: 'center',
      width: 48
    },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 80,
      render: (_: any, index: number) => index + 1
    },
    {
      key: 'name',
      title: $t('page.manage.database.environment.name'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'createdAt',
      title: $t('page.manage.database.environment.createdAt'),
      align: 'center',
      width: 180
    },
    {
      key: 'updatedAt',
      title: $t('page.manage.database.environment.updatedAt'),
      align: 'center',
      width: 180
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: Api.Admin.Environment) => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDelete(row.id)}>
            {{
              default: () => $t('common.confirmDelete'),
              trigger: () => (
                <NButton type="error" ghost size="small">
                  {$t('common.delete')}
                </NButton>
              )
            }}
          </NPopconfirm>
        </div>
      )
    }
  ],
  pagination: {
    pageSize: 10,
    pageSizes: [10, 20, 50, 100],
    showQuickJumper: true
  }
});

const checkedRowKeys = ref<(string | number)[]>([]);
const modalVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Admin.Environment | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  modalVisible.value = true;
}

function handleEdit(row: Api.Admin.Environment) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  modalVisible.value = true;
}

async function handleDelete(id: number) {
  try {
    await fetchDeleteEnvironment(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

async function handleBatchDelete() {
  // TODO: 实现批量删除
  console.log(checkedRowKeys.value);
  window.$message?.info('批量删除功能待实现');
}

function handleSubmitted() {
  modalVisible.value = false;
  getData();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard
      :title="$t('page.manage.database.environment.title')"
      :bordered="false"
      size="small"
      class="card-wrapper sm:flex-1-hidden"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        />
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="800"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="pagination"
        class="sm:h-full"
      />
      <EnvironmentOperateModal
        v-model:visible="modalVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>

