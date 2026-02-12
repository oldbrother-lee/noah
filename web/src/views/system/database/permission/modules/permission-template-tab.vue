<script setup lang="tsx">
import { reactive, ref, onMounted } from 'vue';
import { NButton, NPopconfirm, NTag, NCard, NDataTable } from 'naive-ui';
import type { DataTableColumns, PaginationProps } from 'naive-ui';
import {
  fetchGetPermissionTemplates,
  fetchDeletePermissionTemplate
} from '@/service/api/das';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import TemplateOperateModal from '../template/modules/template-operate-modal.vue';
import PermissionTemplateSearch from './permission-template-search.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'PermissionTemplateTab'
});

const appStore = useAppStore();

const searchParams = reactive({
  name: '',
  description: ''
});

const loading = ref(false);
const data = ref<Api.Das.PermissionTemplate[]>([]);
const checkedRowKeys = ref<(string | number)[]>([]);
const modalVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Das.PermissionTemplate | null>(null);

const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  itemCount: 0,
  onUpdatePage: (page: number) => {
    pagination.page = page;
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
  }
});

const columns: DataTableColumns<Api.Das.PermissionTemplate> = [
  { type: 'selection', width: 48 },
  {
    key: 'index',
    title: $t('common.index'),
    align: 'center',
    width: 64,
    render: (_, index) => index + 1
  },
  {
    key: 'name',
    title: $t('page.manage.database.permissionTemplate.name'),
    align: 'center',
    minWidth: 100
  },
  {
    key: 'description',
    title: $t('page.manage.database.permissionTemplate.description'),
    align: 'center',
    minWidth: 150
  },
  {
    key: 'permissions',
    title: $t('page.manage.database.permissionTemplate.permissions'),
    align: 'center',
    width: 100,
    render: row => {
      const count = row.permissions?.length || 0;
      return <NTag type="info">{count} 项</NTag>;
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
    render: row => (
      <div class="flex-center gap-8px">
        <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
          {$t('common.edit')}
        </NButton>
        <NPopconfirm onPositiveClick={() => handleDelete(row)}>
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
];

const columnChecks = ref(
  columns
    .filter(col => 'key' in col && col.key)
    .map(col => ({
      key: (col as any).key as string,
      title: (col as any).title as string,
      checked: true
    }))
);

async function getData() {
  loading.value = true;
  try {
    const res = await fetchGetPermissionTemplates();
    const responseData = (res as any)?.data || res;
    let list: Api.Das.PermissionTemplate[] = [];
    if (Array.isArray(responseData)) {
      list = responseData;
    } else if (responseData && Array.isArray(responseData.list)) {
      list = responseData.list;
    }

    // 应用搜索过滤
    if (searchParams.name) {
      list = list.filter(item =>
        item.name?.toLowerCase().includes(searchParams.name.toLowerCase())
      );
    }
    if (searchParams.description) {
      list = list.filter(item =>
        item.description?.toLowerCase().includes(searchParams.description.toLowerCase())
      );
    }

    // 处理 ID 字段
    data.value = list.map((item: any) => {
      if (item.ID && !item.id) {
        item.id = item.ID;
      }
      return item;
    });
    pagination.itemCount = data.value.length;
  } catch (error) {
    console.error('Failed to load templates:', error);
  } finally {
    loading.value = false;
  }
}

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  modalVisible.value = true;
}

function handleEdit(row: Api.Das.PermissionTemplate) {
  operateType.value = 'edit';
  const rowData: any = { ...row };
  if (!rowData.id && (row as any).ID) {
    rowData.id = (row as any).ID;
  }
  editingData.value = rowData;
  modalVisible.value = true;
}

async function handleDelete(row: Api.Das.PermissionTemplate) {
  try {
    const id = (row as any).ID || row.id;
    if (!id) {
      window.$message?.error('无法获取模板ID');
      return;
    }
    await fetchDeletePermissionTemplate(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error) {
    window.$message?.error('删除失败');
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

function handleSearch() {
  getData();
}

onMounted(() => {
  getData();
});
</script>

<template>
  <div class="h-full flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <PermissionTemplateSearch v-model:model="searchParams" @search="handleSearch" />
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
        :scroll-x="962"
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
