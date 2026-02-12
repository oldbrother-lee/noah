<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { enableStatusRecord } from '@/constants/business';
import { fetchGetAdminUsers, fetchDeleteAdminUser, fetchGetRoles } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import UserOperateDrawer from './modules/user-operate-drawer.vue';
import UserSearch from './modules/user-search.vue';

const appStore = useAppStore();

const searchParams = reactive({
  page: 1,
  pageSize: 10,
  username: '',
  nickname: '',
  phone: '',
  email: ''
});

const { columns, columnChecks, data, loading, pagination, getData, getDataByPage } = useTable({
  apiFn: () => fetchGetAdminUsers(searchParams),
  transformer: res => {
    // res 可能是 { data: { list, total }, error } 或直接是 { list, total }
    const responseData = (res as any)?.data || res;
    if (responseData && responseData.list) {
      const { list = [], total = 0 } = responseData;
      const current = searchParams.page;
      const size = searchParams.pageSize;
      const pageSize = size <= 0 ? 10 : size;
      const recordsWithIndex = list.map((item: any, index: number) => ({
        ...item,
        index: (current - 1) * pageSize + index + 1
      }));
      return {
        data: recordsWithIndex,
        pageNum: current,
        pageSize,
        total
      };
    }
    return {
      data: [],
      pageNum: 1,
      pageSize: 10,
      total: 0
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
      width: 64,
      render: (_: any, index: number) => index + 1
    },
    {
      key: 'username',
      title: $t('page.manage.user.username'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'nickname',
      title: $t('page.manage.user.nickname'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'email',
      title: $t('page.manage.user.email'),
      align: 'center',
      minWidth: 150
    },
    {
      key: 'phone',
      title: $t('page.manage.user.phone'),
      align: 'center',
      width: 120
    },
    {
      key: 'roles',
      title: $t('page.manage.user.roles'),
      align: 'center',
      minWidth: 150,
      render: (row: Api.Admin.AdminUser) => {
        if (!row.roles || row.roles.length === 0) {
          return <span>-</span>;
        }
        return (
          <div class="flex-center gap-4px">
            {row.roles.map((role: string) => (
              <NTag size="small" type="info">
                {role}
              </NTag>
            ))}
          </div>
        );
      }
    },
    {
      key: 'createdAt',
      title: $t('page.manage.user.createdAt'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: Api.Admin.AdminUser) => (
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
const drawerVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Admin.AdminUser | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  drawerVisible.value = true;
}

function handleEdit(row: Api.Admin.AdminUser) {
  operateType.value = 'edit';
  editingData.value = { ...row };
  drawerVisible.value = true;
}

async function handleDelete(id: number) {
  try {
    await fetchDeleteAdminUser(id);
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

function handleSearch() {
  searchParams.page = 1;
  getDataByPage(1);
}

function handleSubmitted() {
  drawerVisible.value = false;
  getDataByPage();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <UserSearch v-model:model="searchParams" @search="handleSearch" />
    <NCard :title="$t('page.manage.user.title')" :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
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
      <UserOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>

