<script setup lang="tsx">
import { computed, ref, watch } from 'vue';
import { NButton, NDataTable, NTag, NSpace, NAlert, NSpin, NAutoComplete, NInput } from 'naive-ui';
import {
  fetchSyncApi,
  fetchEnterSyncApi,
  fetchIgnoreApi,
  fetchGetApis,
  fetchCreateApi,
  fetchDeleteApi,
  fetchApiAiFill
} from '@/service/api/admin';
import { $t } from '@/locales';

defineOptions({
  name: 'ApiSyncModal'
});

const visible = defineModel<boolean>('visible', { default: false });

interface Emits {
  (e: 'synced'): void;
}

const emit = defineEmits<Emits>();

const loading = ref(false);
const syncing = ref(false);
const aiFillLoading = ref(false);
const syncData = ref<Api.Admin.SyncApiResponse | null>(null);
const errorMsg = ref('');
const apiGroups = ref<string[]>([]);

/** 新增路由的可编辑行（分组、名称可改） */
type NewApiRow = Api.Admin.SyncApiItem & { group: string; name: string };
const newApisRows = ref<NewApiRow[]>([]);

/** 分组选项：已有分组 + 当前表格里已输入的自定义分组（便于下拉选择且自定义输入不丢） */
const groupOptions = computed(() => {
  const fromApi = ['其他', ...apiGroups.value];
  const fromRows = newApisRows.value.map(r => r.group).filter(Boolean);
  return Array.from(new Set([...fromApi, ...fromRows])).map(g => ({ label: g, value: g }));
});

function methodTag(method: string) {
  const colors: Record<string, 'success' | 'info' | 'warning' | 'error'> = {
    GET: 'success',
    POST: 'info',
    PUT: 'warning',
    DELETE: 'error'
  };
  return <NTag type={colors[method] || 'default'} size="small">{method}</NTag>;
}

function updateNewApiRow(path: string, method: string, field: 'group' | 'name', value: string) {
  const r = newApisRows.value.find(x => x.path === path && x.method === method);
  if (r) r[field] = value;
}

async function loadGroups() {
  try {
    const res = await fetchGetApis({ page: 1, pageSize: 1 });
    const data = (res as any)?.data ?? res;
    const list = (data?.groups ?? []) as string[];
    apiGroups.value = Array.isArray(list) ? list : [];
  } catch {
    apiGroups.value = [];
  }
}

async function loadSync() {
  if (!visible.value) return;
  loading.value = true;
  errorMsg.value = '';
  syncData.value = null;
  newApisRows.value = [];
  try {
    await loadGroups();
    const res = await fetchSyncApi();
    const data = (res as any)?.data ?? res;
    syncData.value = data as Api.Admin.SyncApiResponse;
    newApisRows.value = (syncData.value?.newApis ?? []).map(a => ({
      ...a,
      group: a.group || '其他',
      name: a.name || ''
    }));
  } catch (e: any) {
    errorMsg.value = e?.message || (e?.msg as string) || '获取同步数据失败';
  } finally {
    loading.value = false;
  }
}

async function handleSingleAdd(row: NewApiRow) {
  const name = (row.name ?? '').trim();
  if (!name) {
    window.$message?.warning($t('page.manage.api.nameRequired'));
    return;
  }
  try {
    await fetchCreateApi({
      group: row.group || '其他',
      name,
      path: row.path,
      method: row.method
    });
    window.$message?.success($t('common.addSuccess'));
    await loadSync();
  } catch (e: any) {
    const msg = (e?.message ?? e?.msg) || '新增失败';
    window.$message?.error(msg);
  }
}

async function handleSingleDelete(row: Api.Admin.Api) {
  try {
    await fetchDeleteApi(row.id);
    window.$message?.success($t('common.deleteSuccess'));
    await loadSync();
  } catch {
    window.$message?.error('删除失败');
  }
}

async function handleIgnore(row: Api.Admin.SyncApiItem, flag: boolean) {
  try {
    await fetchIgnoreApi({ path: row.path, method: row.method, flag });
    window.$message?.success(flag ? '已忽略' : '已取消忽略');
    await loadSync();
  } catch {
    window.$message?.error('操作失败');
  }
}

async function handleConfirmSync() {
  if (!syncData.value) return;
  const emptyNameRows = newApisRows.value.filter(r => !(r.name ?? '').trim());
  if (emptyNameRows.length > 0) {
    window.$message?.warning($t('page.manage.api.nameRequired'));
    return;
  }
  syncing.value = true;
  try {
    await fetchEnterSyncApi({
      newApis: newApisRows.value.map(a => ({
        group: a.group || '其他',
        name: (a.name ?? '').trim(),
        path: a.path,
        method: a.method
      })),
      deleteApis: syncData.value.deleteApis.map(a => ({ path: a.path, method: a.method }))
    });
    window.$message?.success($t('common.updateSuccess'));
    visible.value = false;
    emit('synced');
  } catch (e: any) {
    const msg = (e?.message ?? e?.msg) || '同步失败';
    window.$message?.error(msg);
  } finally {
    syncing.value = false;
  }
}

function handleClose() {
  visible.value = false;
}

async function handleAiAutoFill() {
  if (!newApisRows.value.length) {
    window.$message?.info($t('page.manage.api.noNewApisToFill'));
    return;
  }
  aiFillLoading.value = true;
  try {
    const res = await fetchApiAiFill({
      items: newApisRows.value.map(r => ({ path: r.path, method: r.method }))
    });
    // 兼容：接口可能返回 data 包装或直接返回数组
    const arr = Array.isArray(res) ? res : (res as any)?.data ?? [];
    if (!Array.isArray(arr) || arr.length === 0) {
      window.$message?.warning('未获取到填充数据');
      return;
    }
    // 用新数组赋值以触发响应式更新，表格才会刷新
    newApisRows.value = newApisRows.value.map(row => {
      const item = arr.find((d: Api.Admin.ApiAiFillItem) => d.path === row.path && d.method === row.method);
      if (item) {
        return {
          ...row,
          group: item.group || row.group || '其他',
          name: item.name || row.name || ''
        };
      }
      return row;
    });
    window.$message?.success($t('page.manage.api.aiAutoFillSuccess'));
  } catch (e: any) {
    const msg = (e?.message ?? e?.msg) || 'AI 自动填充失败';
    window.$message?.error(msg);
  } finally {
    aiFillLoading.value = false;
  }
}

const newApisColumns = [
  { key: 'path', title: $t('page.manage.api.path'), width: 200, ellipsis: { tooltip: true } },
  {
    key: 'group',
    title: $t('page.manage.api.group'),
    width: 160,
    render: (row: NewApiRow) => (
      <NAutoComplete
        size="small"
        value={row.group}
        options={groupOptions.value}
        placeholder={$t('page.manage.api.groupPlaceholder')}
        clearable
        onUpdateValue={v => updateNewApiRow(row.path, row.method, 'group', v ?? '')}
      />
    )
  },
  {
    key: 'name',
    title: $t('page.manage.api.name'),
    width: 140,
    render: (row: NewApiRow) => (
      <NInput
        size="small"
        value={row.name}
        placeholder={$t('page.manage.api.form.name')}
        clearable
        onUpdateValue={v => updateNewApiRow(row.path, row.method, 'name', v ?? '')}
      />
    )
  },
  { key: 'method', title: '请求', width: 90, render: (row: NewApiRow) => methodTag(row.method) },
  {
    key: '_actions',
    title: $t('common.operate'),
    width: 180,
    fixed: 'right' as const,
    render: (row: NewApiRow) => (
      <NSpace size="small">
        <NButton type="primary" size="small" onClick={() => handleSingleAdd(row)}>
          + {$t('page.manage.api.singleAdd')}
        </NButton>
        <NButton size="small" tertiary onClick={() => handleIgnore(row, true)}>
          忽略
        </NButton>
      </NSpace>
    )
  }
];

const deleteApisColumns = [
  { key: 'path', title: $t('page.manage.api.path'), width: 200, ellipsis: { tooltip: true } },
  { key: 'method', title: $t('page.manage.api.method'), width: 80, render: (row: Api.Admin.Api) => methodTag(row.method) },
  { key: 'group', title: $t('page.manage.api.group'), width: 100 },
  { key: 'name', title: $t('page.manage.api.name'), width: 120 },
  {
    key: '_actions',
    title: $t('common.operate'),
    width: 120,
    fixed: 'right' as const,
    render: (row: Api.Admin.Api) => (
      <NButton size="small" tertiary type="error" onClick={() => handleSingleDelete(row)}>
        {$t('page.manage.api.singleDelete')}
      </NButton>
    )
  }
];

const ignoreApisColumns = [
  { key: 'path', title: $t('page.manage.api.path'), width: 220, ellipsis: { tooltip: true } },
  { key: 'method', title: $t('page.manage.api.method'), width: 80, render: (row: Api.Admin.SyncApiItem) => methodTag(row.method) },
  {
    key: '_unignore',
    title: '',
    width: 100,
    render: (row: Api.Admin.SyncApiItem) => (
      <NButton size="small" tertiary onClick={() => handleIgnore(row, false)}>
        取消忽略
      </NButton>
    )
  }
];

const hasAnyChange = computed(
  () =>
    (newApisRows.value.length > 0 || (syncData.value?.deleteApis?.length ?? 0) > 0) &&
    !loading.value
);

watch(visible, v => {
  if (v) loadSync();
});
</script>

<template>
  <NModal
    :show="visible"
    :title="$t('page.manage.api.syncApiTitle')"
    class="w-900px"
    preset="card"
    @update:show="(v: boolean) => (visible = v)"
  >
    <NAlert v-if="errorMsg" type="error" class="mb-16px">
      {{ errorMsg }}
    </NAlert>
    <p class="text-gray-500 text-sm mb-16px">
      {{ $t('page.manage.api.syncApiTip') }}
    </p>

    <NSpin :show="loading">
      <template v-if="syncData">
        <!-- 新增路由：名称必填，支持 AI 自动填充（占位） -->
        <div class="mb-20px">
          <h4 class="mb-8px flex items-center gap-2 flex-wrap">
            <span>{{ $t('page.manage.api.newApis') }}</span>
            <span class="text-xs text-gray-500 font-normal">{{ $t('page.manage.api.newApisTip') }}</span>
            <NButton
              type="primary"
              size="small"
              tertiary
              :loading="aiFillLoading"
              :disabled="!newApisRows.length"
              @click="handleAiAutoFill"
            >
              {{ $t('page.manage.api.aiAutoFill') }}
            </NButton>
          </h4>
          <NDataTable
            :columns="newApisColumns"
            :data="newApisRows"
            :bordered="false"
            size="small"
            max-height="280"
            :scroll-x="760"
          />
        </div>
        <!-- 待删除路由：单条删除 -->
        <div class="mb-20px">
          <h4 class="mb-8px">
            {{ $t('page.manage.api.deleteApis') }}
            <span class="text-xs text-gray-500 font-normal ml-8px">{{ $t('page.manage.api.deleteApisTip') }}</span>
          </h4>
          <NDataTable
            :columns="deleteApisColumns"
            :data="syncData.deleteApis"
            :bordered="false"
            size="small"
            max-height="200"
            :scroll-x="620"
          />
        </div>
        <!-- 已忽略 -->
        <div class="mb-20px">
          <h4 class="mb-8px">
            {{ $t('page.manage.api.ignoreApis') }}
            <span class="text-xs text-gray-500 font-normal ml-8px">{{ $t('page.manage.api.ignoreApisTip') }}</span>
          </h4>
          <NDataTable
            :columns="ignoreApisColumns"
            :data="syncData.ignoreApis"
            :bordered="false"
            size="small"
            max-height="160"
          />
        </div>
      </template>
    </NSpin>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">{{ $t('common.cancel') }}</NButton>
        <NButton
          type="primary"
          :loading="syncing"
          :disabled="!hasAnyChange"
          @click="handleConfirmSync"
        >
          {{ $t('page.manage.api.confirmSync') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
