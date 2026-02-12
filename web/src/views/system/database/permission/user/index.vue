<script setup lang="tsx">
import { ref, onMounted } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable, NTag, NSpace, NSelect, NInput } from 'naive-ui';
import { fetchGetUserPermissions, fetchRevokeSchemaPermission } from '@/service/api/das';
import { fetchGetAdminUsers } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import PermissionOperateModal from '../modules/permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'SystemDatabasePermissionUser'
});

const appStore = useAppStore();

const searchUsername = ref<string>('');
const selectedUsername = ref<string>('');

const users = ref<Array<{ label: string; value: string }>>([]);
const showModal = ref(false);
const checkedRowKeys = ref<(string | number)[]>([]);

const { columns, columnChecks, data, loading, pagination, getData, getDataByPage } = useTable({
  apiFn: async () => {
    if (!selectedUsername.value) {
      return { data: [], pageNum: 1, pageSize: 10, total: 0 };
    }
    const res = await fetchGetUserPermissions(selectedUsername.value);
    const responseData = (res as any)?.data || res;
    const schemaPerms = responseData?.schema_permissions || [];
    
    return {
      data: schemaPerms.map((item: any, index: number) => ({
        ...item,
        index: index + 1
      })),
      pageNum: 1,
      pageSize: schemaPerms.length,
      total: schemaPerms.length
    };
  },
  columns: () => [
    { type: 'selection', align: 'center', width: 48 },
    { key: 'index', title: $t('common.index'), align: 'center', width: 80, render: (_: any, index: number) => index + 1 },
    { key: 'instance_id', title: $t('page.manage.database.permission.instanceId'), align: 'center', minWidth: 200 },
    { key: 'schema', title: $t('page.manage.database.permission.schema'), align: 'center', minWidth: 150 },
    { key: 'created_at', title: $t('page.manage.database.permission.createdAt'), align: 'center', width: 180 },
    { key: 'updated_at', title: $t('page.manage.database.permission.updatedAt'), align: 'center', width: 180 },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: any) => (
        <div class="flex-center gap-8px">
          <NPopconfirm onPositiveClick={() => handleDelete(row.id || row)}>
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

function handleUserChange(value: string) {
  selectedUsername.value = value;
  if (value) {
    getData();
  } else {
    data.value = [];
  }
}

function handleAdd() {
  if (!selectedUsername.value) {
    window.$message?.warning($t('page.manage.database.permission.selectUserFirst'));
    return;
  }
  showModal.value = true;
}

async function handleDelete(row: any) {
  if (!selectedUsername.value) {
    return;
  }
  try {
    const id = typeof row === 'object' ? row.id : row;
    await fetchRevokeSchemaPermission(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error: any) {
    window.$message?.error(error?.message || $t('common.deleteFailed') || '删除失败');
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
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :loading="loading"
            @refresh="getData"
          />
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
        remote
        :row-key="(row: any) => row.id || `${row.instance_id}_${row.schema}`"
        :pagination="pagination"
        class="sm:h-full"
      />
      <PermissionOperateModal
        v-model:visible="showModal"
        :username="selectedUsername"
        @submitted="handleModalSuccess"
      />
    </NCard>
  </div>
</template>

<style scoped></style>

