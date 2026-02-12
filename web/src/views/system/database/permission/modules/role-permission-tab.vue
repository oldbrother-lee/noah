<script setup lang="tsx">
import { ref, onMounted, reactive } from 'vue';
import { NButton, NPopconfirm, NTag, NSelect, NForm, NGrid, NFormItemGi, NCard, NDataTable } from 'naive-ui';
import type { DataTableColumns, PaginationProps } from 'naive-ui';
import {
  fetchGetRolePermissions,
  fetchDeleteRolePermission
} from '@/service/api/das';
import { fetchGetRoles } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import RolePermissionOperateModal from '../role/modules/role-permission-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'RolePermissionTab'
});

const appStore = useAppStore();

const selectedRole = ref<string>('');
const roleOptions = ref<{ label: string; value: string }[]>([]);
const loading = ref(false);
const data = ref<Api.Das.RolePermission[]>([]);
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

const columns: DataTableColumns<Api.Das.RolePermission> = [
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
      const typeMap: Record<string, { label: string; type: NaiveUI.ThemeColor }> = {
        object: { label: '直接权限', type: 'info' },
        template: { label: '权限模板', type: 'success' }
      };
      const typeInfo = typeMap[row.permission_type] || { label: row.permission_type, type: 'default' };
      return <NTag type={typeInfo.type}>{typeInfo.label}</NTag>;
    }
  },
  {
    key: 'template_name',
    title: '权限模板',
    align: 'center',
    width: 150,
    render: row => {
      if (row.permission_type === 'template') {
        return <span>{(row as any).template_name || `模板ID: ${row.permission_id}`}</span>;
      }
      return <span>-</span>;
    }
  },
  {
    key: 'instance_id',
    title: '实例ID',
    align: 'center',
    minWidth: 200,
    render: row => {
      if (row.permission_type === 'template') {
        return <span>-</span>;
      }
      return <span>{row.instance_id || '-'}</span>;
    }
  },
  {
    key: 'schema',
    title: '库名',
    align: 'center',
    minWidth: 120,
    render: row => {
      if (row.permission_type === 'template') {
        return <span>-</span>;
      }
      return <span>{row.schema || '-'}</span>;
    }
  },
  {
    key: 'table',
    title: '表名',
    align: 'center',
    minWidth: 120,
    render: row => {
      if (row.permission_type === 'template') {
        return <span>-</span>;
      }
      return <span>{row.table || '-'}</span>;
    }
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

const columnChecks = ref(
  columns
    .filter(col => 'key' in col && col.key)
    .map(col => ({
      key: (col as any).key as string,
      title: (col as any).title as string,
      checked: true
    }))
);

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

async function getData() {
  if (!selectedRole.value) {
    data.value = [];
    pagination.itemCount = 0;
    return;
  }

  loading.value = true;
  try {
    const res = await fetchGetRolePermissions(selectedRole.value);
    const responseData = (res as any)?.data || res;
    const list = Array.isArray(responseData) ? responseData : [];
    // 统一处理 ID 字段：转换为 id（小写）
    data.value = list.map((item: any) => {
      if (item.ID && !item.id) {
        item.id = item.ID;
      }
      return item;
    });
    pagination.itemCount = data.value.length;
  } catch (error) {
    console.error('Failed to load role permissions:', error);
  } finally {
    loading.value = false;
  }
}

function handleAdd() {
  if (!selectedRole.value) {
    window.$message?.warning('请先选择角色');
    return;
  }
  modalVisible.value = true;
}

async function handleDelete(row: any) {
  const id = row.id || row.ID;
  if (!id) {
    window.$message?.error('无法获取权限ID');
    return;
  }
  try {
    await fetchDeleteRolePermission(id);
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
  <div class="h-full flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NForm label-placement="left" :label-width="80">
        <NGrid responsive="screen" item-responsive>
          <NFormItemGi span="24 s:12 m:6" label="选择角色" class="pr-24px">
            <NSelect
              v-model:value="selectedRole"
              :options="roleOptions"
              placeholder="请选择角色"
              filterable
              clearable
              class="w-full"
              @update:value="handleRoleChange"
            />
          </NFormItemGi>
        </NGrid>
      </NForm>
    </NCard>
    <NCard
      title="角色权限列表"
      :bordered="false"
      size="small"
      class="card-wrapper sm:flex-1-hidden"
    >
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :disabled-add="!selectedRole"
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
        :row-key="(row: any) => row.id || row.ID || `${row.role}_${row.permission_id}`"
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
