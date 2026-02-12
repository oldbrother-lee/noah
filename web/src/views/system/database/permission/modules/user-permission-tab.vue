<script setup lang="tsx">
import { ref, onMounted, reactive, computed } from 'vue';
import { NButton, NPopconfirm, NSelect, NForm, NGrid, NFormItemGi, NCard, NDataTable, NTabs, NTabPane, NTag } from 'naive-ui';
import type { DataTableColumns, PaginationProps } from 'naive-ui';
import { fetchGetUserPermissions, fetchRevokeSchemaPermission, fetchRevokeTablePermission } from '@/service/api/das';
import { fetchGetAdminUsers } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import PermissionOperateModal from './permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'UserPermissionTab'
});

const appStore = useAppStore();

const selectedUsername = ref<string>('');
const users = ref<Array<{ label: string; value: string }>>([]);
const loading = ref(false);
const data = ref<any[]>([]);
const tableData = ref<any[]>([]);
const activeTab = ref<'schema' | 'table'>('schema');
const checkedRowKeys = ref<(string | number)[]>([]);
const showModal = ref(false);

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

const schemaColumns: DataTableColumns<any> = [
  { type: 'selection', width: 48 },
  {
    key: 'index',
    title: $t('common.index'),
    align: 'center',
    width: 64,
    render: (_, index) => index + 1
  },
  {
    key: 'instance_id',
    title: $t('page.manage.database.permission.instanceId'),
    align: 'center',
    minWidth: 200
  },
  {
    key: 'schema',
    title: $t('page.manage.database.permission.schema'),
    align: 'center',
    minWidth: 120
  },
  {
    key: 'created_at',
    title: $t('page.manage.database.permission.createdAt'),
    align: 'center',
    width: 180
  },
  {
    key: 'updated_at',
    title: $t('page.manage.database.permission.updatedAt'),
    align: 'center',
    width: 180
  },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center',
    width: 100,
    render: row => (
      <NPopconfirm onPositiveClick={() => handleDeleteSchema(row)}>
        {{
          default: () => $t('common.confirmDelete'),
          trigger: () => (
            <NButton type="error" ghost size="small">
              {$t('common.delete')}
            </NButton>
          )
        }}
      </NPopconfirm>
    )
  }
];

const tableColumns: DataTableColumns<any> = [
  { type: 'selection', width: 48 },
  {
    key: 'index',
    title: $t('common.index'),
    align: 'center',
    width: 64,
    render: (_, index) => index + 1
  },
  {
    key: 'instance_id',
    title: $t('page.manage.database.permission.instanceId'),
    align: 'center',
    minWidth: 200
  },
  {
    key: 'schema',
    title: $t('page.manage.database.permission.schema'),
    align: 'center',
    minWidth: 120
  },
  {
    key: 'table',
    title: '表名',
    align: 'center',
    minWidth: 120
  },
  {
    key: 'rule',
    title: '规则',
    align: 'center',
    width: 100,
    render: row => (
      <NTag type={row.rule === 'allow' ? 'success' : 'error'} size="small">
        {row.rule === 'allow' ? '允许' : '拒绝'}
      </NTag>
    )
  },
  {
    key: 'created_at',
    title: $t('page.manage.database.permission.createdAt'),
    align: 'center',
    width: 180
  },
  {
    key: 'updated_at',
    title: $t('page.manage.database.permission.updatedAt'),
    align: 'center',
    width: 180
  },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center',
    width: 100,
    render: row => (
      <NPopconfirm onPositiveClick={() => handleDeleteTable(row)}>
        {{
          default: () => $t('common.confirmDelete'),
          trigger: () => (
            <NButton type="error" ghost size="small">
              {$t('common.delete')}
            </NButton>
          )
        }}
      </NPopconfirm>
    )
  }
];

const schemaColumnChecks = ref(
  schemaColumns
    .filter(col => 'key' in col && col.key)
    .map(col => ({
      key: (col as any).key as string,
      title: (col as any).title as string,
      checked: true
    }))
);

const tableColumnChecks = ref(
  tableColumns
    .filter(col => 'key' in col && col.key)
    .map(col => ({
      key: (col as any).key as string,
      title: (col as any).title as string,
      checked: true
    }))
);

const currentColumnChecks = computed({
  get: () => activeTab.value === 'schema' ? schemaColumnChecks.value : tableColumnChecks.value,
  set: (value) => {
    if (activeTab.value === 'schema') {
      schemaColumnChecks.value = value;
    } else {
      tableColumnChecks.value = value;
    }
  }
});

async function loadUsers() {
  try {
    const res = await fetchGetAdminUsers({ page: 1, pageSize: 1000 });
    const responseData = (res as any)?.data || res;
    const userList = responseData?.list || [];
    users.value = userList.map((u: any) => ({
      label: `${u.username}${u.nickname ? ` (${u.nickname})` : ''}`,
      value: u.username
    }));
  } catch (error) {
    console.error('Failed to load users:', error);
  }
}

async function getData() {
  if (!selectedUsername.value) {
    data.value = [];
    tableData.value = [];
    pagination.itemCount = 0;
    return;
  }

  loading.value = true;
  try {
    const res = await fetchGetUserPermissions(selectedUsername.value);
    const responseData = (res as any)?.data || res;
    const schemaPerms = responseData?.schema_permissions || [];
    const tablePerms = responseData?.table_permissions || [];
    
    // 统一处理 ID 字段：转换为 id（小写）
    data.value = schemaPerms.map((item: any) => {
      if (item.ID && !item.id) {
        item.id = item.ID;
      }
      return item;
    });
    
    tableData.value = tablePerms.map((item: any) => {
      if (item.ID && !item.id) {
        item.id = item.ID;
      }
      return item;
    });
    
    pagination.itemCount = activeTab.value === 'schema' ? data.value.length : tableData.value.length;
  } catch (error) {
    console.error('Failed to load user permissions:', error);
  } finally {
    loading.value = false;
  }
}

function handleUserChange(value: string) {
  selectedUsername.value = value;
  if (value) {
    getData();
  } else {
    data.value = [];
    pagination.itemCount = 0;
  }
}

function handleAdd() {
  if (!selectedUsername.value) {
    window.$message?.warning($t('page.manage.database.permission.selectUserFirst'));
    return;
  }
  showModal.value = true;
}

async function handleDeleteSchema(row: any) {
  if (!selectedUsername.value) {
    return;
  }
  try {
    const id = row.id || row.ID;
    if (!id) {
      window.$message?.error('无法获取权限ID');
      return;
    }
    await fetchRevokeSchemaPermission(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error: any) {
    window.$message?.error(error?.message || '删除失败');
  }
}

async function handleDeleteTable(row: any) {
  if (!selectedUsername.value) {
    return;
  }
  try {
    const id = row.id || row.ID;
    if (!id) {
      window.$message?.error('无法获取权限ID');
      return;
    }
    await fetchRevokeTablePermission(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error: any) {
    window.$message?.error(error?.message || '删除失败');
  }
}

async function handleBatchDelete() {
  console.log(checkedRowKeys.value);
  window.$message?.info('批量删除功能待实现');
}

async function handleModalSuccess() {
  showModal.value = false;
  await getData();
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="h-full flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NForm label-placement="left" :label-width="80">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:6" label="选择用户" class="pr-24px">
            <NSelect
              v-model:value="selectedUsername"
              :options="users"
              filterable
              clearable
              :placeholder="$t('page.manage.database.permission.selectUser')"
              class="w-full"
              @update:value="handleUserChange"
            />
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NCard>
    <NCard
      title="用户权限列表"
      :bordered="false"
      size="small"
      class="card-wrapper sm:flex-1-hidden"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="currentColumnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :disabled-add="!selectedUsername"
          :loading="loading"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        />
      </template>
      <NTabs v-model:value="activeTab" type="line" @update:value="() => { pagination.itemCount = activeTab === 'schema' ? data.length : tableData.length; }">
        <NTabPane name="schema" tab="库权限">
          <NDataTable
            v-model:checked-row-keys="checkedRowKeys"
            :columns="schemaColumns"
            :data="data"
            size="small"
            :flex-height="!appStore.isMobile"
            :scroll-x="962"
            :loading="loading"
            remote
            :row-key="(row: any) => {
              const id = row.id || row.ID;
              if (id) return id;
              return `${row.instance_id}_${row.schema}`;
            }"
            :pagination="pagination"
            class="sm:h-full"
          />
        </NTabPane>
        <NTabPane name="table" tab="表权限">
          <NDataTable
            v-model:checked-row-keys="checkedRowKeys"
            :columns="tableColumns"
            :data="tableData"
            size="small"
            :flex-height="!appStore.isMobile"
            :scroll-x="1100"
            :loading="loading"
            remote
            :row-key="(row: any) => {
              const id = row.id || row.ID;
              if (id) return id;
              return `${row.instance_id}_${row.schema}_${row.table}`;
            }"
            :pagination="pagination"
            class="sm:h-full"
          />
        </NTabPane>
      </NTabs>
      <PermissionOperateModal
        v-model:visible="showModal"
        :username="selectedUsername"
        @submitted="handleModalSuccess"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
