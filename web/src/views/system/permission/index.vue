<script setup lang="tsx">
import { computed, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NSpace, NTag, NTabPane, NTabs } from 'naive-ui';
import {
  fetchGetRoles,
  fetchDeleteRole,
  fetchGetApis,
  fetchDeleteApi,
  fetchDeleteApisByIds,
  fetchFreshCasbin
} from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import RoleOperateDrawer from './modules/role-operate-drawer.vue';
import ApiOperateDrawer from './modules/api-operate-drawer.vue';
import PermissionModal from './modules/permission-modal.vue';
import ApiSyncModal from './modules/api-sync-modal.vue';

const appStore = useAppStore();

// ==================== 角色管理 ====================
const roleSearchParams = reactive({
  page: 1,
  pageSize: 10,
  sid: '',
  name: ''
});

const {
  columns: roleColumns,
  columnChecks: roleColumnChecks,
  data: roleData,
  loading: roleLoading,
  pagination: rolePagination,
  getData: getRoleData
} = useTable({
  apiFn: () => fetchGetRoles(roleSearchParams),
  transformer: res => {
    // res 可能是 { data: { list, total }, error } 或直接是 { list, total }
    const responseData = (res as any)?.data || res;
    if (responseData && responseData.list) {
      const { list = [], total = 0 } = responseData;
      const current = roleSearchParams.page;
      const size = roleSearchParams.pageSize;
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
    return { data: [], pageNum: 1, pageSize: 10, total: 0 };
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
      key: 'name',
      title: $t('page.manage.role.roleName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'sid',
      title: $t('page.manage.role.roleCode'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'createdAt',
      title: $t('page.manage.role.createdAt'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 260,
      fixed: 'right',
      render: (row: Api.Admin.Role) => (
        <div class="flex-center gap-8px">
          <NButton type="warning" ghost size="small" onClick={() => handlePermission(row)}>
            {$t('page.manage.role.assignPermission')}
          </NButton>
          <NButton type="primary" ghost size="small" onClick={() => handleEditRole(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDeleteRole(row.id)}>
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
  pagination: { pageSize: 10 }
});

const roleCheckedRowKeys = ref<(string | number)[]>([]);
const roleDrawerVisible = ref(false);
const roleOperateType = ref<NaiveUI.TableOperateType>('add');
const roleEditingData = ref<Api.Admin.Role | null>(null);

// 自定义分页，支持手动更新搜索参数
const roleTablePagination = computed(() => ({
  ...rolePagination,
  onUpdatePage: (page: number) => {
    roleSearchParams.page = page;
    getRoleData();
  },
  onUpdatePageSize: (pageSize: number) => {
    roleSearchParams.page = 1;
    roleSearchParams.pageSize = pageSize;
    getRoleData();
  }
}));

function handleAddRole() {
  roleOperateType.value = 'add';
  roleEditingData.value = null;
  roleDrawerVisible.value = true;
}

function handleEditRole(row: Api.Admin.Role) {
  roleOperateType.value = 'edit';
  roleEditingData.value = { ...row };
  roleDrawerVisible.value = true;
}

async function handleDeleteRole(id: number) {
  try {
    await fetchDeleteRole(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getRoleData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

function handleRoleSubmitted() {
  roleDrawerVisible.value = false;
  getRoleData();
}

// 权限分配
const permissionModalVisible = ref(false);
const currentRole = ref<Api.Admin.Role | null>(null);

function handlePermission(row: Api.Admin.Role) {
  currentRole.value = row;
  permissionModalVisible.value = true;
}

// ==================== API管理 ====================
const apiSearchParams = reactive({
  page: 1,
  pageSize: 10,
  group: '',
  name: '',
  path: '',
  method: ''
});

const {
  columns: apiColumns,
  columnChecks: apiColumnChecks,
  data: apiData,
  loading: apiLoading,
  pagination: apiPagination,
  getData: getApiData
} = useTable({
  apiFn: () => fetchGetApis(apiSearchParams),
  transformer: res => {
    // res 可能是 { data: { list, total }, error } 或直接是 { list, total }
    const responseData = (res as any)?.data || res;
    if (responseData && responseData.list) {
      const { list = [], total = 0 } = responseData;
      const current = apiSearchParams.page;
      const size = apiSearchParams.pageSize;
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
    return { data: [], pageNum: 1, pageSize: 10, total: 0 };
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
      key: 'group',
      title: $t('page.manage.api.group'),
      align: 'center',
      minWidth: 100
    },
    {
      key: 'name',
      title: $t('page.manage.api.name'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'path',
      title: $t('page.manage.api.path'),
      align: 'center',
      minWidth: 200
    },
    {
      key: 'method',
      title: $t('page.manage.api.method'),
      align: 'center',
      width: 80,
      render: (row: Api.Admin.Api) => {
        const methodColors: Record<string, NaiveUI.ThemeColor> = {
          GET: 'success',
          POST: 'info',
          PUT: 'warning',
          DELETE: 'error'
        };
        return <NTag type={methodColors[row.method] || 'default'}>{row.method}</NTag>;
      }
    },
    {
      key: 'createdAt',
      title: $t('page.manage.api.createdAt'),
      align: 'center',
      width: 160
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: Api.Admin.Api) => (
        <div class="flex-center gap-8px">
          <NButton type="primary" ghost size="small" onClick={() => handleEditApi(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDeleteApi(row.id)}>
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
  pagination: { pageSize: 10 }
});

const apiCheckedRowKeys = ref<(string | number)[]>([]);
const apiDrawerVisible = ref(false);
const apiOperateType = ref<NaiveUI.TableOperateType>('add');
const apiEditingData = ref<Api.Admin.Api | null>(null);

// 自定义分页，支持手动更新搜索参数
const apiTablePagination = computed(() => ({
  ...apiPagination,
  onUpdatePage: (page: number) => {
    apiSearchParams.page = page;
    getApiData();
  },
  onUpdatePageSize: (pageSize: number) => {
    apiSearchParams.page = 1;
    apiSearchParams.pageSize = pageSize;
    getApiData();
  }
}));

function handleAddApi() {
  apiOperateType.value = 'add';
  apiEditingData.value = null;
  apiDrawerVisible.value = true;
}

function handleEditApi(row: Api.Admin.Api) {
  apiOperateType.value = 'edit';
  apiEditingData.value = { ...row };
  apiDrawerVisible.value = true;
}

async function handleDeleteApi(id: number) {
  try {
    await fetchDeleteApi(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getApiData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

async function handleBatchDeleteApis() {
  const ids = apiCheckedRowKeys.value as number[];
  if (ids.length === 0) return;
  try {
    await fetchDeleteApisByIds(ids);
    window.$message?.success($t('common.deleteSuccess'));
    apiCheckedRowKeys.value = [];
    await getApiData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

const apiSyncModalVisible = ref(false);

function handleApiSynced() {
  getApiData();
}

async function handleFreshCasbin() {
  try {
    await fetchFreshCasbin();
    window.$message?.success('权限已刷新');
  } catch (error) {
    window.$message?.error('刷新失败');
  }
}

function handleApiSubmitted() {
  apiDrawerVisible.value = false;
  getApiData();
}
</script>

<template>
  <div class="flex-col-stretch gap-16px lt-sm:overflow-auto">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NTabs type="line" animated>
        <!-- 角色管理 Tab -->
        <NTabPane name="role" :tab="$t('page.manage.role.title')">
          <div class="flex flex-col gap-16px">
            <!-- 角色搜索 -->
            <NCard :bordered="false" size="small">
              <NForm :model="roleSearchParams" label-placement="left" :label-width="80">
                <NGrid responsive="screen" item-responsive>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.role.roleName')"
                    path="name"
                    class="pr-24px"
                  >
                    <NInput
                      v-model:value="roleSearchParams.name"
                      :placeholder="$t('page.manage.role.form.roleName')"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.role.roleCode')"
                    path="sid"
                    class="pr-24px"
                  >
                    <NInput
                      v-model:value="roleSearchParams.sid"
                      :placeholder="$t('page.manage.role.form.roleCode')"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi span="24 s:12 m:6" class="pr-24px">
                    <NSpace class="w-full" justify="end">
                      <NButton
                        @click="
                          () => {
                            roleSearchParams.name = '';
                            roleSearchParams.sid = '';
                            roleSearchParams.page = 1;
                            getRoleData();
                          }
                        "
                      >
                        <template #icon>
                          <icon-ic-round-refresh class="text-icon" />
                        </template>
                        {{ $t('common.reset') }}
                      </NButton>
                      <NButton type="primary" ghost @click="() => { roleSearchParams.page = 1; getRoleData(); }">
                        <template #icon>
                          <icon-ic-round-search class="text-icon" />
                        </template>
                        {{ $t('common.search') }}
                      </NButton>
                    </NSpace>
                  </NFormItemGi>
                </NGrid>
              </NForm>
            </NCard>

            <!-- 角色列表 -->
            <NCard :bordered="false" size="small">
              <template #header>
                <TableHeaderOperation
                  v-model:columns="roleColumnChecks"
                  :disabled-delete="roleCheckedRowKeys.length === 0"
                  :loading="roleLoading"
                  @add="handleAddRole"
                  @refresh="getRoleData"
                />
              </template>
              <NDataTable
                v-model:checked-row-keys="roleCheckedRowKeys"
                :columns="roleColumns"
                :data="roleData"
                size="small"
                :scroll-x="900"
                :loading="roleLoading"
                remote
                :row-key="row => row.id"
                :pagination="roleTablePagination"
              />
            </NCard>
          </div>
        </NTabPane>

        <!-- API管理 Tab -->
        <NTabPane name="api" :tab="$t('page.manage.api.title')">
          <div class="flex flex-col gap-16px">
            <!-- API搜索 -->
            <NCard :bordered="false" size="small">
              <NForm :model="apiSearchParams" label-placement="left" :label-width="80">
                <NGrid responsive="screen" item-responsive>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.api.group')"
                    path="group"
                    class="pr-24px"
                  >
                    <NInput
                      v-model:value="apiSearchParams.group"
                      :placeholder="$t('page.manage.api.form.group')"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.api.name')"
                    path="name"
                    class="pr-24px"
                  >
                    <NInput
                      v-model:value="apiSearchParams.name"
                      :placeholder="$t('page.manage.api.form.name')"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.api.path')"
                    path="path"
                    class="pr-24px"
                  >
                    <NInput
                      v-model:value="apiSearchParams.path"
                      :placeholder="$t('page.manage.api.form.path')"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi
                    span="24 s:12 m:6"
                    :label="$t('page.manage.api.method')"
                    path="method"
                    class="pr-24px"
                  >
                    <NSelect
                      v-model:value="apiSearchParams.method"
                      :placeholder="$t('page.manage.api.form.method')"
                      :options="[
                        { label: 'GET', value: 'GET' },
                        { label: 'POST', value: 'POST' },
                        { label: 'PUT', value: 'PUT' },
                        { label: 'DELETE', value: 'DELETE' }
                      ]"
                      clearable
                    />
                  </NFormItemGi>
                  <NFormItemGi span="24 m:12" class="pr-24px">
                    <NSpace class="w-full" justify="end">
                      <NButton
                        @click="
                          () => {
                            apiSearchParams.group = '';
                            apiSearchParams.name = '';
                            apiSearchParams.path = '';
                            apiSearchParams.method = '';
                            apiSearchParams.page = 1;
                            getApiData();
                          }
                        "
                      >
                        <template #icon>
                          <icon-ic-round-refresh class="text-icon" />
                        </template>
                        {{ $t('common.reset') }}
                      </NButton>
                      <NButton type="primary" ghost @click="() => { apiSearchParams.page = 1; getApiData(); }">
                        <template #icon>
                          <icon-ic-round-search class="text-icon" />
                        </template>
                        {{ $t('common.search') }}
                      </NButton>
                    </NSpace>
                  </NFormItemGi>
                </NGrid>
              </NForm>
            </NCard>

            <!-- API列表 -->
            <NCard :bordered="false" size="small">
              <template #header>
                <div class="flex-y-center justify-between">
                  <TableHeaderOperation
                    v-model:columns="apiColumnChecks"
                    :disabled-delete="apiCheckedRowKeys.length === 0"
                    :loading="apiLoading"
                    @add="handleAddApi"
                    @refresh="getApiData"
                    @delete="handleBatchDeleteApis"
                  />
                  <NSpace>
                    <NButton type="primary" secondary @click="apiSyncModalVisible = true">
                      {{ $t('page.manage.api.syncRoute') }}
                    </NButton>
                    <NButton secondary @click="handleFreshCasbin">
                      {{ $t('page.manage.api.freshCasbin') }}
                    </NButton>
                  </NSpace>
                </div>
              </template>
              <NDataTable
                v-model:checked-row-keys="apiCheckedRowKeys"
                :columns="apiColumns"
                :data="apiData"
                size="small"
                :scroll-x="900"
                :loading="apiLoading"
                remote
                :row-key="row => row.id"
                :pagination="apiTablePagination"
              />
            </NCard>
          </div>
        </NTabPane>
      </NTabs>
    </NCard>

    <!-- 角色操作抽屉 -->
    <RoleOperateDrawer
      v-model:visible="roleDrawerVisible"
      :operate-type="roleOperateType"
      :row-data="roleEditingData"
      @submitted="handleRoleSubmitted"
    />

    <!-- API操作抽屉 -->
    <ApiOperateDrawer
      v-model:visible="apiDrawerVisible"
      :operate-type="apiOperateType"
      :row-data="apiEditingData"
      @submitted="handleApiSubmitted"
    />

    <!-- 权限分配模态框 -->
    <PermissionModal v-model:visible="permissionModalVisible" :role="currentRole" />

    <!-- 同步路由弹窗 -->
    <ApiSyncModal v-model:visible="apiSyncModalVisible" @synced="handleApiSynced" />
  </div>
</template>

<style scoped></style>
