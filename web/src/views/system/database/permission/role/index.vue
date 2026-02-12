<script setup lang="tsx">
import { ref, onMounted } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable, NSelect, NTag } from 'naive-ui';
import {
  fetchGetRolePermissions,
  fetchDeleteRolePermission
} from '@/service/api/das';
import { fetchGetRoles } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import RolePermissionOperateModal from './modules/role-permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'SystemDatabaseRolePermission'
});

const appStore = useAppStore();

const selectedRole = ref<string>('');
const roleOptions = ref<{ label: string; value: string }[]>([]);

const { columns, columnChecks, data, loading, pagination, getData } = useTable({
  apiFn: async () => {
    if (!selectedRole.value) {
      return { data: [], pageNum: 1, pageSize: 10, total: 0 };
    }
    const res = await fetchGetRolePermissions(selectedRole.value);
    const responseData = (res as any)?.data || res;
    const list = Array.isArray(responseData) ? responseData : [];

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
    { type: 'selection', align: 'center', width: 48 },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 80,
      render: (_: any, index: number) => index + 1
    },
    {
      key: 'permission_type',
      title: '权限类型',
      align: 'center',
      width: 120,
      render: (row: Api.Das.RolePermission) => {
        const typeMap: Record<string, { label: string; type: string }> = {
          object: { label: '直接权限', type: 'info' },
          template: { label: '权限模板', type: 'success' },
          group: { label: '权限组', type: 'warning' }
        };
        const typeInfo = typeMap[row.permission_type] || { label: row.permission_type, type: 'default' };
        return <NTag type={typeInfo.type as any}>{typeInfo.label}</NTag>;
      }
    },
    {
      key: 'permission_id',
      title: '权限ID',
      align: 'center',
      width: 100
    },
    {
      key: 'instance_id',
      title: '实例ID',
      align: 'center',
      minWidth: 200,
      render: (row: Api.Das.RolePermission) => row.instance_id || '-'
    },
    {
      key: 'schema',
      title: '库名',
      align: 'center',
      minWidth: 150,
      render: (row: Api.Das.RolePermission) => row.schema || '-'
    },
    {
      key: 'table',
      title: '表名',
      align: 'center',
      minWidth: 150,
      render: (row: Api.Das.RolePermission) => row.table || '-'
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
      render: (row: Api.Das.RolePermission) => (
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
      )
    }
  ],
  pagination: { pageSize: 10, pageSizes: [10, 20, 50, 100], showQuickJumper: true }
});

const checkedRowKeys = ref<(string | number)[]>([]);
const modalVisible = ref(false);

async function loadRoles() {
  try {
    const res = await fetchGetRoles({ page: 1, pageSize: 1000 });
    const responseData = (res as any)?.data || res;
    if (responseData && responseData.list) {
      roleOptions.value = responseData.list.map((role: any) => ({
        label: `${role.name} (${role.sid})`,
        value: role.sid
      }));
    }
  } catch (error) {
    console.error('Failed to load roles:', error);
  }
}

function handleAdd() {
  if (!selectedRole.value) {
    window.$message?.warning('请先选择角色');
    return;
  }
  modalVisible.value = true;
}

async function handleDelete(id: number) {
  try {
    await fetchDeleteRolePermission(id);
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

function handleRoleChange() {
  getData();
}

onMounted(() => {
  loadRoles();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard
      :title="$t('page.manage.database.rolePermission.title')"
      :bordered="false"
      size="small"
      class="card-wrapper sm:flex-1-hidden"
    >
      <template #header-extra>
        <div class="flex items-center gap-16px">
          <NSelect
            v-model:value="selectedRole"
            :options="roleOptions"
            placeholder="选择角色"
            clearable
            class="w-200px"
            @update:value="handleRoleChange"
          />
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :disabled-delete="checkedRowKeys.length === 0"
            :loading="loading"
            :disabled-add="!selectedRole"
            @add="handleAdd"
            @delete="handleBatchDelete"
            @refresh="getData"
          />
        </div>
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1500"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="pagination"
        class="sm:h-full"
      />
      <RolePermissionOperateModal
        v-model:visible="modalVisible"
        :role="selectedRole"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>

