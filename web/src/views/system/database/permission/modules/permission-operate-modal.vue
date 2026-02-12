<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, NSpace, NSelect, NDataTable } from 'naive-ui';
import { jsonClone } from '@sa/utils';
import { fetchGrantSchemaPermission, fetchGrantTablePermission, fetchTables } from '@/service/api/das';
import { fetchGetDBConfigs } from '@/service/api/admin';
import { fetchOrdersSchemas } from '@/service/api/orders';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'PermissionOperateModal'
});

interface Props {
  username: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => {
  return $t('page.manage.database.permission.addPermission');
});

type Model = Api.Das.GrantSchemaPermissionRequest & {
  table?: string;
  rule?: 'allow' | 'deny';
};

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    username: '',
    instance_id: '',
    schema: '',
    table: '',
    rule: 'allow'
  };
}

type RuleKey = Extract<keyof Model, 'username' | 'instance_id' | 'schema'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  username: defaultRequiredRule,
  instance_id: defaultRequiredRule,
  schema: defaultRequiredRule
};

const dbConfigOptions = ref<{ label: string; value: string }[]>([]);
const schemaOptions = ref<{ label: string; value: string }[]>([]);
const permissionType = ref<'schema' | 'table'>('schema');

// 表选择相关
const showTableModal = ref(false);
const tableList = ref<any[]>([]);
const selectedTableKeys = ref<string[]>([]);
const tableSearchText = ref('');
const tableLoading = ref(false);
const tablePagination = ref({
  page: 1,
  pageSize: 30,
  itemCount: 0
});

async function getDBConfigs() {
  try {
    const res = await fetchGetDBConfigs({ useType: '查询' });
    const responseData = (res as any)?.data || res;
    const configs = Array.isArray(responseData) ? responseData : [];
    dbConfigOptions.value = configs.map((config: any) => ({
      label: `${config.hostname}:${config.port}${config.instance_id ? ` (${config.instance_id})` : ''}`,
      value: config.instance_id
    }));
  } catch (error) {
    console.error('Failed to load DB configs:', error);
  }
}

async function getSchemas(instanceId: string) {
  if (!instanceId) {
    schemaOptions.value = [];
    tableOptions.value = [];
    return;
  }
  try {
    const res = await fetchOrdersSchemas({ instance_id: instanceId });
    const responseData = (res as any)?.data || res;
    const schemas = Array.isArray(responseData) ? responseData : [];
    schemaOptions.value = schemas.map((schema: any) => ({
      label: schema.schema || schema.name,
      value: schema.schema || schema.name
    }));
  } catch (error) {
    console.error('Failed to load schemas:', error);
    schemaOptions.value = [];
  }
}

// 打开表选择弹窗
async function openTableModal() {
  if (!model.value.instance_id || !model.value.schema) {
    window.$message?.warning('请先选择实例和库名');
    return;
  }
  
  showTableModal.value = true;
  selectedTableKeys.value = [];
  tableSearchText.value = '';
  tablePagination.value.page = 1;
  await loadTables();
}

// 加载表列表
async function loadTables() {
  if (!model.value.instance_id || !model.value.schema) {
    return;
  }
  
  tableLoading.value = true;
  try {
    const res = await fetchTables({
      instance_id: model.value.instance_id,
      schema: model.value.schema
    });
    const responseData = (res as any)?.data || res;
    const tables = Array.isArray(responseData) ? responseData : [];
    
    // 应用搜索过滤
    let filteredTables = tables;
    if (tableSearchText.value) {
      const searchLower = tableSearchText.value.toLowerCase();
      filteredTables = tables.filter((table: any) => {
        const tableName = (table.table_name || table.name || table.TableName || '').toLowerCase();
        return tableName.includes(searchLower);
      });
    }
    
    tablePagination.value.itemCount = filteredTables.length;
    
    // 分页
    const start = (tablePagination.value.page - 1) * tablePagination.value.pageSize;
    const end = start + tablePagination.value.pageSize;
    tableList.value = filteredTables.slice(start, end);
  } catch (error) {
    console.error('Failed to load tables:', error);
    window.$message?.error('加载表列表失败');
    tableList.value = [];
  } finally {
    tableLoading.value = false;
  }
}

// 确认选择表
function confirmTableSelection() {
  if (selectedTableKeys.value.length === 0) {
    window.$message?.warning('请至少选择一个表');
    return;
  }
  
  // 设置第一个选中的表（用于显示）
  const [schema, tableName] = selectedTableKeys.value[0].split('#');
  model.value.table = tableName;
  
  // 存储所有选中的表（用于批量提交）
  (model.value as any).selectedTables = selectedTableKeys.value.map((key: string) => {
    const [, name] = key.split('#');
    return name;
  });
  
  showTableModal.value = false;
}

// 添加整个库的权限（不选择表）
function handleAddSchemaPermission() {
  if (!model.value.instance_id || !model.value.schema) {
    window.$message?.warning('请选择实例和库名');
    return;
  }
  
  model.value.table = '';
  (model.value as any).selectedTables = [];
  permissionType.value = 'schema';
}

// 表选择表格的列
const tableColumns = [
  {
    type: 'selection',
    width: 48
  },
  {
    key: 'table_schema',
    title: '库名称',
    width: 200,
    render: () => model.value.schema || '-'
  },
  {
    key: 'table_name',
    title: '表名称',
    width: 200,
    render: (row: any) => row.table_name || row.name || row.TableName || ''
  },
  {
    key: 'table_comment',
    title: '描述',
    width: 200,
    render: (row: any) => row.table_comment || row.comment || row.TableComment || '-'
  }
];

function handleInitModel() {
  model.value = createDefaultModel();
  model.value.username = props.username;
  (model.value as any).selectedTables = [];
  permissionType.value = 'schema';
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (permissionType.value === 'table') {
      // 表权限：支持批量创建
      const selectedTables = (model.value as any).selectedTables || [];
      
      if (selectedTables.length > 0) {
        // 批量创建表权限
        const promises = selectedTables.map((tableName: string) => {
          return fetchGrantTablePermission({
            username: model.value.username,
            instance_id: model.value.instance_id,
            schema: model.value.schema,
            table: tableName,
            rule: model.value.rule || 'allow'
          });
        });
        
        await Promise.all(promises);
        window.$message?.success(`成功添加 ${selectedTables.length} 个表权限`);
      } else if (model.value.table) {
        // 单个表权限
        await fetchGrantTablePermission({
          username: model.value.username,
          instance_id: model.value.instance_id,
          schema: model.value.schema,
          table: model.value.table,
          rule: model.value.rule || 'allow'
        });
        window.$message?.success($t('common.addSuccess'));
      } else {
        window.$message?.warning('请选择表或添加整个库权限');
        return;
      }
    } else {
      // 库权限
      await fetchGrantSchemaPermission({
        username: model.value.username,
        instance_id: model.value.instance_id,
        schema: model.value.schema
      });
      window.$message?.success($t('common.addSuccess'));
    }
    closeModal();
    emit('submitted');
  } catch (error) {
    window.$message?.error($t('common.operationFailed') || '操作失败');
  }
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    getDBConfigs();
  }
});

watch(
  () => model.value.instance_id,
  (newVal) => {
    if (newVal) {
      getSchemas(newVal);
    } else {
      schemaOptions.value = [];
      model.value.schema = '';
      model.value.table = '';
      (model.value as any).selectedTables = [];
    }
  }
);

watch(
  () => model.value.schema,
  (newVal) => {
    if (!newVal) {
      model.value.table = '';
      (model.value as any).selectedTables = [];
    }
  }
);

watch(permissionType, () => {
  if (permissionType.value === 'schema') {
    model.value.table = '';
    (model.value as any).selectedTables = [];
  }
});

onMounted(() => {
  getDBConfigs();
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" :style="{ width: '600px' }" :mask-closable="false">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
      <NFormItem :label="$t('page.manage.database.permission.username')" path="username">
        <NInput v-model:value="model.username" :disabled="true" clearable />
      </NFormItem>
      <NFormItem :label="$t('page.manage.database.permission.instanceId')" path="instance_id">
        <NSelect
          v-model:value="model.instance_id"
          :options="dbConfigOptions"
          :placeholder="$t('page.manage.database.permission.form.instanceId')"
          clearable
        />
      </NFormItem>
      <NFormItem :label="$t('page.manage.database.permission.schema')" path="schema">
        <NSelect
          v-model:value="model.schema"
          :options="schemaOptions"
          :placeholder="$t('page.manage.database.permission.form.schema')"
          :disabled="!model.instance_id"
          clearable
        />
      </NFormItem>
      <NFormItem label="权限类型">
        <NSelect
          v-model:value="permissionType"
          :options="[
            { label: '库权限', value: 'schema' },
            { label: '表权限', value: 'table' }
          ]"
        />
      </NFormItem>
      <template v-if="permissionType === 'table'">
        <NFormItem label="表权限">
          <NSpace>
            <NButton type="primary" @click="openTableModal" :disabled="!model.instance_id || !model.schema">
              选择表
            </NButton>
            <NButton @click="handleAddSchemaPermission" :disabled="!model.instance_id || !model.schema">
              添加整个库
            </NButton>
          </NSpace>
          <div v-if="model.table || (model as any).selectedTables?.length > 0" class="mt-8px text-12px text-gray-500">
            {{
              (model as any).selectedTables?.length > 0
                ? `已选择 ${(model as any).selectedTables.length} 个表`
                : model.table
                  ? `当前表: ${model.table}`
                  : '整个库'
            }}
          </div>
        </NFormItem>
        <NFormItem label="规则">
          <NSelect
            v-model:value="model.rule"
            :options="[
              { label: '允许 (allow)', value: 'allow' },
              { label: '拒绝 (deny)', value: 'deny' }
            ]"
          />
        </NFormItem>
      </template>
    </NForm>
    <template #footer>
      <NSpace :size="16">
        <NButton @click="closeModal">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
  
  <!-- 表选择弹窗 -->
  <NModal
    v-model:show="showTableModal"
    title="选择表"
    preset="card"
    :style="{ width: '800px' }"
    :mask-closable="false"
  >
    <div class="flex flex-col gap-16px">
      <div class="flex gap-8px items-center">
        <span class="w-80px">表名称:</span>
        <NInput
          v-model:value="tableSearchText"
          placeholder="请输入表名"
          clearable
          style="flex: 1"
          @keyup.enter="loadTables"
        />
        <NButton type="primary" @click="loadTables">查询</NButton>
      </div>
      <NDataTable
        v-model:checked-row-keys="selectedTableKeys"
        :columns="tableColumns"
        :data="tableList"
        :loading="tableLoading"
        :pagination="tablePagination"
        :row-key="(row: any) => `${model.schema}#${row.table_name || row.name || row.TableName}`"
        size="small"
        @update:page="(page: number) => { tablePagination.page = page; loadTables(); }"
        @update:page-size="(pageSize: number) => { tablePagination.pageSize = pageSize; tablePagination.page = 1; loadTables(); }"
      />
      <div class="text-12px text-gray-500">
        已选择 {{ selectedTableKeys.length }} 个表
      </div>
    </div>
    <template #footer>
      <NSpace :size="16">
        <NButton @click="showTableModal = false">取消</NButton>
        <NButton type="primary" @click="confirmTableSelection">确定</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>

