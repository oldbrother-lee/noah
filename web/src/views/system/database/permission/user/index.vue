<script setup lang="tsx">
import { ref, onMounted } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable, NTag, NSpace, NSelect } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchGetUserPermissionList, fetchDeleteUserPermission } from '@/service/api/das';
import { fetchGetAdminUsers } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import UserPermissionOperateModal from '../modules/user-permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'SystemDatabasePermissionUser'
});

const appStore = useAppStore();
const selectedUsername = ref<string>('');
const users = ref<Array<{ label: string; value: string }>>([]);
const showModal = ref(false);
const checkedRowKeys = ref<(string | number)[]>([]);
const loading = ref(false);
const data = ref<Api.Das.UserPermission[]>([]);

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
    title: $t('page.manage.database.permission.instanceId'),
    align: 'center',
    minWidth: 200,
    render: row => (row.permission_type === 'object' ? <span>{row.instance_id || '-'}</span> : <span>-</span>)
  },
  {
    key: 'schema',
    title: $t('page.manage.database.permission.schema'),
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
    title: $t('page.manage.database.permission.createdAt'),
    align: 'center',
    width: 180
  },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center',
    width: 100,
    render: (row: any) => (
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
    return;
  }
  loading.value = true;
  try {
    const res = await fetchGetUserPermissionList(selectedUsername.value);
    const list = (res as any)?.data ?? res;
    data.value = Array.isArray(list) ? list : [];
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function handleUserChange(value: string) {
  selectedUsername.value = value;
  if (value) getData();
  else data.value = [];
}

function handleAdd() {
  if (!selectedUsername.value) {
    window.$message?.warning($t('page.manage.database.permission.selectUserFirst'));
    return;
  }
  showModal.value = true;
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
    window.$message?.error(e?.message || $t('common.deleteFailed') || '删除失败');
  }
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
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard
      :title="$t('page.manage.database.permission.title')"
      :bordered="false"
      size="small"
      class="card-wrapper sm:flex-1-hidden"
    >
      <template #header-extra>
        <NSpace>
          <NSelect
            v-model:value="selectedUsername"
            :options="users"
            filterable
            clearable
            :placeholder="$t('page.manage.database.permission.selectUser')"
            class="w-200px"
            @update:value="handleUserChange"
          />
          <NButton type="primary" @click="handleAdd">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            {{ $t('page.manage.database.permission.addPermission') }}
          </NButton>
          <TableHeaderOperation :loading="loading" @refresh="getData" />
        </NSpace>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1200"
        :loading="loading"
        :row-key="(row: any) => row.id ?? row.ID"
        class="sm:h-full"
      />
      <UserPermissionOperateModal
        v-model:visible="showModal"
        :username="selectedUsername || ''"
        @submitted="handleModalSuccess"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
