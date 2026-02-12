<script setup lang="tsx">
import { ref } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable, NTag } from 'naive-ui';
import {
  fetchGetPermissionTemplates,
  fetchDeletePermissionTemplate
} from '@/service/api/das';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import TemplateOperateModal from './modules/template-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'SystemDatabasePermissionTemplate'
});

const appStore = useAppStore();

const { columns, columnChecks, data, loading, pagination, getData } = useTable({
  apiFn: () => fetchGetPermissionTemplates(),
  transformer: res => {
    const responseData = (res as any)?.data || res;
    let list: Api.Das.PermissionTemplate[] = [];
    if (Array.isArray(responseData)) {
      list = responseData;
    } else if (responseData && Array.isArray(responseData.list)) {
      list = responseData.list;
    }

    return {
      data: list.map((item: any, index: number) => {
        // 统一处理 ID 字段：转换为 id（小写）
        const normalizedItem = { ...item };
        if (normalizedItem.ID && !normalizedItem.id) {
          normalizedItem.id = normalizedItem.ID;
        }
        return {
          ...normalizedItem,
          index: index + 1
        };
      }),
      pageNum: 1,
      pageSize: list.length,
      total: list.length
    };
  },
  columns: () => [
    { type: 'selection', align: 'center', width: 48 },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 80,
      render: (_: any, index: number) => index + 1
    },
    {
      key: 'name',
      title: $t('page.manage.database.permissionTemplate.name'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'description',
      title: $t('page.manage.database.permissionTemplate.description'),
      align: 'center',
      minWidth: 200
    },
    {
      key: 'permissions',
      title: $t('page.manage.database.permissionTemplate.permissions'),
      align: 'center',
      minWidth: 200,
      render: (row: Api.Das.PermissionTemplate) => {
        const count = row.permissions?.length || 0;
        return <span>{count} {count === 1 ? '项' : '项'}</span>;
      }
    },
    {
      key: 'created_at',
      title: $t('page.manage.database.permissionTemplate.createdAt'),
      align: 'center',
      width: 180
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: Api.Das.PermissionTemplate) => (
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
  pagination: { pageSize: 10, pageSizes: [10, 20, 50, 100], showQuickJumper: true }
});

const checkedRowKeys = ref<(string | number)[]>([]);
const modalVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Das.PermissionTemplate | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  modalVisible.value = true;
}

function handleEdit(row: Api.Das.PermissionTemplate) {
  operateType.value = 'edit';
  // 确保 id 字段存在（后端可能返回 ID 大写）
  const rowData: any = { ...row };
  if (!rowData.id && (row as any).ID) {
    rowData.id = (row as any).ID;
  }
  editingData.value = rowData;
  modalVisible.value = true;
}

async function handleDelete(idOrRow: number | Api.Das.PermissionTemplate) {
  try {
    // 兼容传入 row 对象或 id 数字
    let id: number;
    if (typeof idOrRow === 'number') {
      id = idOrRow;
    } else {
      id = (idOrRow as any).ID || idOrRow.id;
    }
    if (!id) {
      window.$message?.error('无法获取模板ID');
      return;
    }
    await fetchDeletePermissionTemplate(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

async function handleBatchDelete() {
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
      :title="$t('page.manage.database.permissionTemplate.title')"
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
        :scroll-x="1200"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="pagination"
        class="sm:h-full"
      />
      <TemplateOperateModal
        v-model:visible="modalVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>

