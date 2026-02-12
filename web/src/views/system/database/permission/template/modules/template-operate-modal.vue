<script setup lang="tsx">
import { computed, ref, watch } from 'vue';
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NDataTable,
  NTag,
  NSelect,
  NPopconfirm
} from 'naive-ui';
import { jsonClone } from '@sa/utils';
import {
  fetchCreatePermissionTemplate,
  fetchUpdatePermissionTemplate,
  fetchTables
} from '@/service/api/das';
import { fetchGetDBConfigs, fetchGetEnvironments } from '@/service/api/admin';
import { fetchOrdersSchemas } from '@/service/api/orders';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'TemplateOperateModal'
});

interface Props {
  operateType: NaiveUI.TableOperateType;
  rowData?: Api.Das.PermissionTemplate | null;
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
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.manage.database.permissionTemplate.addTemplate'),
    edit: $t('page.manage.database.permissionTemplate.editTemplate')
  };
  return titles[props.operateType];
});

type Model = Api.Das.PermissionTemplateCreateRequest;

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    description: '',
    permissions: []
  };
}

type RuleKey = Extract<keyof Model, 'name' | 'permissions'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  permissions: defaultRequiredRule
};

// 权限列表
const permissions = ref<Api.Das.PermissionObject[]>([]);

// 环境选项
const environmentOptions = ref<{ label: string; value: number }[]>([]);
const selectedEnvironment = ref<number | null>(null);

// 所有数据库配置（未过滤）
const allDBConfigs = ref<any[]>([]);
// 数据库配置选项（根据环境过滤后）
const dbConfigOptions = ref<{ label: string; value: string }[]>([]);
const schemaOptions = ref<{ label: string; value: string }[]>([]);

// 当前选择的实例ID
const selectedInstanceId = ref<string>('');
const selectedSchema = ref<string>('');
const selectedTable = ref<string>('');

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

async function loadEnvironments() {
  try {
    const res = await fetchGetEnvironments();
    const responseData = (res as any)?.data || res;
    const environments = Array.isArray(responseData) ? responseData : [];
    // 使用 Map 去重，确保每个环境名称唯一
    const uniqueEnvs = new Map();
    environments.forEach((env: any) => {
      const id = Number(env.id || env.ID);
      if (!uniqueEnvs.has(id)) {
        uniqueEnvs.set(id, {
          label: env.name || `环境${id}`,
          value: id
        });
      }
    });
    environmentOptions.value = Array.from(uniqueEnvs.values());
  } catch (error) {
    console.error('Failed to load environments:', error);
  }
}

async function loadDBConfigs() {
  try {
    // 只获取"查询"类型的数据库配置，防止越权
    const res = await fetchGetDBConfigs({ useType: '查询' } as any);
    const responseData = (res as any)?.data || res;
    allDBConfigs.value = Array.isArray(responseData) ? responseData : [];
    filterDBConfigsByEnvironment();
  } catch (error) {
    console.error('Failed to load DB configs:', error);
  }
}

function filterDBConfigsByEnvironment() {
  let configs = allDBConfigs.value;
  
  // 如果选择了环境，则过滤
  if (selectedEnvironment.value !== null && selectedEnvironment.value !== undefined) {
    const envId = Number(selectedEnvironment.value);
    configs = configs.filter((config: any) => {
      const configEnv = config.environment;
      // 处理 null、undefined 和类型转换
      if (configEnv === null || configEnv === undefined) {
        return false;
      }
      const configEnvNum = Number(configEnv);
      return configEnvNum === envId;
    });
    
    // 调试日志（开发时使用）
    if (import.meta.env.DEV) {
      console.log('环境过滤:', {
        selectedEnv: envId,
        totalConfigs: allDBConfigs.value.length,
        filteredConfigs: configs.length,
        sampleConfig: configs[0]
      });
    }
  }
  
  // 使用 Map 去重，确保每个实例唯一
  const uniqueConfigs = new Map();
  configs.forEach((config: any) => {
    const instanceId = config.instance_id;
    if (instanceId && !uniqueConfigs.has(instanceId)) {
      // 确保 label 格式正确，显示 hostname:port (remark)
      const hostname = config.hostname || config.Hostname || '';
      const port = config.port || config.Port || '';
      const remark = config.remark || config.Remark || '';
      const label = hostname && port 
        ? `${hostname}:${port}${remark ? ` (${remark})` : ''}`
        : remark || instanceId;
      uniqueConfigs.set(instanceId, {
        label,
        value: instanceId
      });
    }
  });
  dbConfigOptions.value = Array.from(uniqueConfigs.values());
  
  // 如果当前选择的实例不在过滤后的列表中，清空选择
  if (selectedInstanceId.value && !dbConfigOptions.value.some(opt => opt.value === selectedInstanceId.value)) {
    selectedInstanceId.value = '';
    selectedSchema.value = '';
    selectedTable.value = '';
    schemaOptions.value = [];
  }
}

async function loadSchemas(instanceId: string) {
  if (!instanceId) {
    schemaOptions.value = [];
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
  if (!selectedInstanceId.value || !selectedSchema.value) {
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
  if (!selectedInstanceId.value || !selectedSchema.value) {
    return;
  }
  
  tableLoading.value = true;
  try {
    const res = await fetchTables({
      instance_id: selectedInstanceId.value,
      schema: selectedSchema.value
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
  
  // 为每个选中的表添加权限
  selectedTableKeys.value.forEach((tableKey: string) => {
    const [, tableName] = tableKey.split('#');
    const perm: Api.Das.PermissionObject = {
      instance_id: selectedInstanceId.value,
      schema: selectedSchema.value,
      table: tableName || undefined
    };
    
    // 检查是否已存在
    const exists = permissions.value.some(
      p =>
        p.instance_id === perm.instance_id &&
        p.schema === perm.schema &&
        p.table === perm.table
    );
    
    if (!exists) {
      permissions.value.push(perm);
    }
  });
  
  model.value.permissions = permissions.value;
  showTableModal.value = false;
  selectedTableKeys.value = [];
}

// 添加整个库的权限（不选择表）
function handleAddSchemaPermission() {
  if (!selectedInstanceId.value || !selectedSchema.value) {
    window.$message?.warning('请选择实例和库名');
    return;
  }

  const perm: Api.Das.PermissionObject = {
    instance_id: selectedInstanceId.value,
    schema: selectedSchema.value,
    table: undefined // 整个库的权限
  };

  // 检查是否已存在
  const exists = permissions.value.some(
    p =>
      p.instance_id === perm.instance_id &&
      p.schema === perm.schema &&
      !p.table // 整个库的权限
  );

  if (exists) {
    window.$message?.warning('该库权限已存在');
    return;
  }

  permissions.value.push(perm);
  model.value.permissions = permissions.value;

  // 清空选择
  selectedInstanceId.value = '';
  selectedSchema.value = '';
  schemaOptions.value = [];
}

function handleDeletePermission(index: number) {
  permissions.value.splice(index, 1);
  model.value.permissions = permissions.value;
}

const permissionColumns = [
  {
    key: 'instance_id',
    title: '实例ID',
    width: 200
  },
  {
    key: 'schema',
    title: '库名',
    width: 150
  },
  {
    key: 'table',
    title: '表名',
    width: 150,
    render: (row: Api.Das.PermissionObject) => row.table || '整个库'
  },
  {
    key: 'operate',
    title: '操作',
    width: 100,
    render: (_: any, index: number) => (
      <NButton
        type="error"
        size="small"
        onClick={() => handleDeletePermission(index)}
      >
        删除
      </NButton>
    )
  }
];

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
    render: () => selectedSchema.value || '-'
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
  permissions.value = [];
  selectedEnvironment.value = null;
  selectedInstanceId.value = '';
  selectedSchema.value = '';
  selectedTable.value = '';
  schemaOptions.value = [];

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, jsonClone(props.rowData));
    permissions.value = props.rowData.permissions || [];
    model.value.permissions = permissions.value;
    // 确保 id 字段存在（后端可能返回 ID 大写）
    if (!model.value.id && (props.rowData as any).ID) {
      (model.value as any).id = (props.rowData as any).ID;
    }
  }
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  if (permissions.value.length === 0) {
    window.$message?.warning('请至少添加一个权限');
    return;
  }

  try {
    if (props.operateType === 'edit' && props.rowData) {
      // 获取 ID
      const templateId = (props.rowData as any).ID || props.rowData.id || (model.value as any).id;
      if (!templateId) {
        window.$message?.error('无法获取模板ID');
        return;
      }
      await fetchUpdatePermissionTemplate(templateId, model.value);
      window.$message?.success($t('common.updateSuccess'));
    } else {
      await fetchCreatePermissionTemplate(model.value);
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
    loadEnvironments();
    loadDBConfigs();
  }
});

watch(
  () => selectedEnvironment.value,
  () => {
    filterDBConfigsByEnvironment();
  }
);

watch(
  () => selectedInstanceId.value,
  newVal => {
    if (newVal) {
      loadSchemas(newVal);
    } else {
      schemaOptions.value = [];
      selectedSchema.value = '';
    }
  }
);
</script>

<template>
  <NModal
    v-model:show="visible"
    :title="title"
    preset="card"
    :style="{ width: '900px' }"
    :mask-closable="false"
  >
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
      <NFormItem :label="$t('page.manage.database.permissionTemplate.name')" path="name">
        <NInput
          v-model:value="model.name"
          :placeholder="$t('page.manage.database.permissionTemplate.form.name')"
          clearable
        />
      </NFormItem>
      <NFormItem :label="$t('page.manage.database.permissionTemplate.description')" path="description">
        <NInput
          v-model:value="model.description"
          :placeholder="$t('page.manage.database.permissionTemplate.form.description')"
          type="textarea"
          clearable
        />
      </NFormItem>
      <NFormItem label="权限配置" path="permissions">
        <div class="w-full">
          <div class="flex flex-col gap-8px mb-16px">
            <NSelect
              v-model:value="selectedEnvironment"
              :options="environmentOptions"
              placeholder="选择环境"
              clearable
            />
            <NSelect
              v-model:value="selectedInstanceId"
              :options="dbConfigOptions"
              placeholder="选择实例"
              clearable
            />
            <NSelect
              v-model:value="selectedSchema"
              :options="schemaOptions"
              placeholder="选择库名"
              :disabled="!selectedInstanceId"
              clearable
            />
            <NSpace>
              <NButton type="primary" @click="openTableModal" :disabled="!selectedInstanceId || !selectedSchema">
                选择表
              </NButton>
              <NButton @click="handleAddSchemaPermission" :disabled="!selectedInstanceId || !selectedSchema">
                添加整个库
              </NButton>
            </NSpace>
          </div>
          <NDataTable
            :columns="permissionColumns"
            :data="permissions"
            size="small"
            :bordered="true"
            max-height="300px"
          />
        </div>
      </NFormItem>
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
        :row-key="(row: any) => `${selectedSchema.value}#${row.table_name || row.name || row.TableName}`"
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

