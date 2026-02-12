<script setup lang="tsx">
import { ref, reactive } from 'vue';
import { NButton, NPopconfirm, NCard, NDataTable, NTag, NInput, NSpace } from 'naive-ui';
import { fetchGetInspectParams, fetchDeleteInspectParam } from '@/service/api/inspect';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import InspectOperateModal from './modules/inspect-operate-modal.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';

defineOptions({
  name: 'SystemDatabaseInspect'
});

const appStore = useAppStore();

const searchParams = reactive({
  current: 1,
  size: 10,
  remark: ''
});

const remarkKeyword = ref('');

const { columns, columnChecks, data, loading, pagination, getData, getDataByPage } = useTable({
  apiFn: (params?: any) => {
    // useTable 会将 apiParams 作为参数传递，参数格式是 { current, size }
    const page = params?.current || searchParams.current;
    const pageSize = params?.size || searchParams.size;
    const remark = searchParams.remark || '';
    return fetchGetInspectParams({ page, page_size: pageSize, remark }) as any;
  },
  apiParams: searchParams as any,
  transformer: res => {
    const responseData = (res as any)?.data || res;
    const list: Api.Inspect.InspectParam[] = responseData.list || [];
    const total = responseData.total || 0;

    const current = searchParams.current;
    const size = searchParams.size;
    const pageSize = size <= 0 ? 10 : size;

    return {
      data: list.map((item: any, index: number) => {
        // 统一处理 ID 字段：转换为 id（小写）
        const normalizedItem = { ...item };
        if (normalizedItem.ID && !normalizedItem.id) {
          normalizedItem.id = normalizedItem.ID;
        }
        // 处理 params 字段：如果是字符串，尝试解析为 JSON；如果已经是对象，保持不变
        if (typeof normalizedItem.params === 'string') {
          try {
            normalizedItem.params = JSON.parse(normalizedItem.params);
          } catch (e) {
            normalizedItem.params = {};
          }
        } else if (!normalizedItem.params || typeof normalizedItem.params !== 'object') {
          normalizedItem.params = {};
        }
        return {
          ...normalizedItem,
          index: (current - 1) * pageSize + index + 1
        };
      }),
      pageNum: current,
      pageSize: pageSize,
      total: total
    };
  },
  columns: (): any => [
    {
      type: 'selection',
      align: 'center',
      width: 48
    },
    {
      key: 'index',
      title: $t('common.index'),
      align: 'center',
      width: 80,
      render: (_: any, index: number) => index + 1
    },
    {
      key: 'remark',
      title: '备注',
      align: 'center',
      minWidth: 200
    },
    {
      key: 'params',
      title: '参数配置',
      align: 'center',
      minWidth: 300,
      render: (row: Api.Inspect.InspectParam) => {
        const params = row.params || {};
        const keys = Object.keys(params);
        if (keys.length === 0) {
          return <span class="text-gray-400">无参数</span>;
        }
        return (
          <div class="flex flex-wrap gap-4px">
            {keys.slice(0, 3).map(key => (
              <NTag size="small" type="info">
                {key}: {String(params[key])}
              </NTag>
            ))}
            {keys.length > 3 && (
              <NTag size="small" type="default">
                +{keys.length - 3}
              </NTag>
            )}
          </div>
        );
      }
    },
    {
      key: 'createdAt',
      title: '创建时间',
      align: 'center',
      width: 180
    },
    {
      key: 'updatedAt',
      title: '更新时间',
      align: 'center',
      width: 180
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 130,
      render: (row: Api.Inspect.InspectParam) => (
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
const modalVisible = ref(false);
const operateType = ref<NaiveUI.TableOperateType>('add');
const editingData = ref<Api.Inspect.InspectParam | null>(null);

function handleAdd() {
  operateType.value = 'add';
  editingData.value = null;
  modalVisible.value = true;
}

function handleEdit(row: Api.Inspect.InspectParam) {
  operateType.value = 'edit';
  // 确保 id 字段存在（后端可能返回 ID 大写）
  const rowData: any = { ...row };
  if (!rowData.id && (row as any).ID) {
    rowData.id = (row as any).ID;
  }
  editingData.value = rowData;
  modalVisible.value = true;
}

async function handleDelete(idOrRow: number | Api.Inspect.InspectParam) {
  try {
    // 兼容传入 row 对象或 id 数字
    let id: number;
    if (typeof idOrRow === 'number') {
      id = idOrRow;
    } else {
      id = (idOrRow as any).ID || idOrRow.id;
    }
    if (!id) {
      window.$message?.error('无法获取参数ID');
      return;
    }
    await fetchDeleteInspectParam(id);
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
  searchParams.remark = remarkKeyword.value.trim();
  searchParams.current = 1;
  getDataByPage(1);
}

function handleReset() {
  remarkKeyword.value = '';
  searchParams.remark = '';
  searchParams.current = 1;
  getDataByPage(1);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard title="审核参数配置" :bordered="false" size="small" class="card-wrapper">
      <NSpace :vertical="true" :size="16">
        <NSpace :size="12" align="center">
          <span class="text-14px">备注搜索：</span>
          <NInput
            v-model:value="remarkKeyword"
            placeholder="请输入备注关键词"
            clearable
            style="width: 300px"
            @keyup.enter="handleSearch"
          />
          <NButton type="primary" ghost @click="handleSearch">搜索</NButton>
          <NButton @click="handleReset">重置</NButton>
        </NSpace>
      </NSpace>
    </NCard>
    <NCard
      title="审核参数列表"
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
        :scroll-x="1200"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="pagination"
        class="sm:h-full"
      />
      <InspectOperateModal
        v-model:visible="modalVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleSubmitted"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
