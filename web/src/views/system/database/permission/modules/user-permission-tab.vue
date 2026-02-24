<script setup lang="tsx">
import { ref, onMounted, reactive } from 'vue';
import { NButton, NPopconfirm, NTag, NSelect, NCard, NDataTable } from 'naive-ui';
import type { DataTableColumns, PaginationProps } from 'naive-ui';
import { fetchGetUserPermissionList, fetchDeleteUserPermission } from '@/service/api/das';
import { fetchGetAdminUsers } from '@/service/api/admin';
import { $t } from '@/locales';
import UserPermissionOperateModal from './user-permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'UserPermissionTab'
});

const selectedUsername = ref<string>('');
const users = ref<Array<{ label: string; value: string }>>([]);
const loading = ref(false);
const data = ref<Api.Das.UserPermission[]>([]);
const checkedRowKeys = ref<(string | number)[]>([]);
const modalVisible = ref(false);

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

const typeMap: Record<string, { label: string; type: NaiveUI.ThemeColor }> = {
  object: { label: '直接权限', type: 'info' },
  template: { label: '权限模板', type: 'success' }
};

const columns: DataTableColumns<Api.Das.UserPermission> = [
  { type: 'selection', width: 48 },
  {
    key: 'index',
    title: $t('common.index'),
    align: 'center',
    width: 64,
    render: (_, index) => index + 1
  },
  {
    key: 'permission_type',
    title: '权限类型',
    align: 'center',
    width: 100,
    render: row => {
      const info = typeMap[row.permission_type] || { label: row.permission_type, type: 'default' };
      return <NTag type={info.type}>{info.label}</NTag>;
    }
  },
  {
    key: 'template_name',
    title: '权限模板',
    align: 'center',
    width: 150,
    render: row => (row.permission_type === 'template' ? <span>{(row as any).template_name || `模板ID: ${row.permission_id}`}</span> : <span>-</span>)
  },
  {
    key: 'instance_id',
    title: '实例ID',
    align: 'center',
    minWidth: 200,
    render: row => (row.permission_type === 'object' ? <span>{row.instance_id || '-'}</span> : <span>-</span>)
  },
  {
    key: 'schema',
    title: '库名',
    align: 'center',
    minWidth: 120,
    render: row => (row.permission_type === 'object' ? <span>{row.schema || '-'}</span> : <span>-</span>)
  },
  {
    key: 'table',
    title: '表名',
    align: 'center',
    minWidth: 120,
    render: row => (row.permission_type === 'object' ? <span>{row.table || '-'}</span> : <span>-</span>)
  },
  {
    key: 'created_at',
    title: '创建时间',
    align: 'center',
    width: 180
  },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center',
    width: 100,
    render: row => (
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
    )
  }
];

async function loadUsers() {
  try {
    const res = await fetchGetAdminUsers({ page: 1, pageSize: 1000 });
    const responseData = (res as any)?.data || res;
    const list = responseData?.list || [];
    users.value = list.map((u: any) => ({ label: u.username || u.userName, value: u.username || u.userName }));
  } catch (e) {
    console.error(e);
  }
}

async function getData() {
  if (!selectedUsername.value) {
    data.value = [];
    pagination.itemCount = 0;
    return;
  }
  loading.value = true;
  try {
    const res = await fetchGetUserPermissionList(selectedUsername.value);
    const list = (res as any)?.data ?? res;
    data.value = Array.isArray(list) ? list : [];
    pagination.itemCount = data.value.length;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function handleUserChange(value: string) {
  selectedUsername.value = value;
  if (value) getData();
  else {
    data.value = [];
    pagination.itemCount = 0;
  }
}

function handleAdd() {
  if (!selectedUsername.value) {
    window.$message?.warning($t('page.manage.database.permission.selectUserFirst'));
    return;
  }
  modalVisible.value = true;
}

async function handleDelete(row: any) {
  const id = row.id ?? row.ID;
  if (!id) {
    window.$message?.error('无法获取权限ID');
    return;
  }
  try {
    await fetchDeleteUserPermission(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (e: any) {
    window.$message?.error(e?.message || '删除失败');
  }
}

function handleSubmitted() {
  getData();
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="flex flex-col gap-16px">
    <NCard size="small">
      <div class="flex flex-wrap items-center gap-12px">
        <span class="font-medium">选择用户:</span>
        <NSelect
          v-model:value="selectedUsername"
          :options="users"
          filterable
          clearable
          placeholder="请选择用户"
          style="width: 280px"
          @update:value="handleUserChange"
        />
      </div>
    </NCard>
    <NCard size="small">
      <TableHeaderOperation :loading="loading" @refresh="getData">
        <template #default>
          <NButton type="primary" :disabled="!selectedUsername" @click="handleAdd">
            {{ $t('page.manage.database.permission.addPermission') }}
          </NButton>
        </template>
      </TableHeaderOperation>
      <NDataTable
        :columns="columns"
        :data="data"
        :loading="loading"
        :row-key="(row: any) => row.id ?? row.ID"
        :pagination="pagination"
        size="small"
      />
    </NCard>
    <UserPermissionOperateModal
      v-model:visible="modalVisible"
      :username="selectedUsername || ''"
      @submitted="handleSubmitted"
    />
  </div>
</template>
