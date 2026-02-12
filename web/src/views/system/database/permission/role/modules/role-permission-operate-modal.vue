<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue';
import type { FormRules } from 'naive-ui';
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSelect,
  NInputNumber,
  NDataTable
} from 'naive-ui';
import { fetchCreateRolePermission } from '@/service/api/das';
import { fetchGetPermissionTemplates, fetchTables } from '@/service/api/das';
import { fetchGetDBConfigs } from '@/service/api/admin';
import { fetchOrdersSchemas } from '@/service/api/orders';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'RolePermissionOperateModal'
});

interface Props {
  role: string;
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
  return '新增角色权限';
});

type Model = {
  role: string;
  permission_type: 'object' | 'template';
  permission_id: number | null;
  instance_id: string;
  schema: string;
  table: string;
};

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    role: '',
    permission_type: 'object',
    permission_id: null,
    instance_id: '',
    schema: '',
    table: ''
  };
}

const rules = computed(() => {
  const baseRules: Record<string, App.Global.FormRule> = {
    role: defaultRequiredRule,
    permission_type: defaultRequiredRule
  };

  if (model.value.permission_type === 'template') {
    baseRules.permission_id = {
      required: true,
      message: '请选择权限模板',
      trigger: ['blur', 'change'],
      validator: (_rule, value) => {
        // 检查值是否为有效数字且大于0
        if (value === null || value === undefined || value === '') {
          return new Error('请选择权限模板');
        }
        const numValue = Number(value);
        if (isNaN(numValue) || numValue <= 0) {
          return new Error('请选择权限模板');
        }
        return true;
      }
    };
  } else {
    baseRules.instance_id = defaultRequiredRule;
    baseRules.schema = defaultRequiredRule;
  }

  return baseRules;
});

const permissionTypeOptions = [
  { label: '直接权限', value: 'object' },
  { label: '权限模板', value: 'template' }
];

// 权限模板选项
const templateOptions = ref<{ label: string; value: number }[]>([]);

// 直接权限相关
const dbConfigOptions = ref<{ label: string; value: string }[]>([]);
const schemaOptions = ref<{ label: string; value: string }[]>([]);

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

async function loadTemplates() {
  try {
    const res = await fetchGetPermissionTemplates();
    const responseData = (res as any)?.data || res;
    const templates = Array.isArray(responseData) ? responseData : [];
    templateOptions.value = templates.map((t: any) => {
      // 统一处理 ID 字段：转换为 id（小写）
      const templateId = t.id || t.ID;
      return {
        label: `${t.name} (${t.permissions?.length || 0}项)`,
        value: Number(templateId) // 确保是数字类型
      };
    });
  } catch (error) {
    console.error('Failed to load templates:', error);
  }
}


async function loadDBConfigs() {
  try {
    const res = await fetchGetDBConfigs({ useType: '查询' } as any);
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
  model.value.role = props.role;
  (model.value as any).selectedTables = [];
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (model.value.permission_type === 'template') {
      // 模板权限：直接提交
      const submitData: Api.Das.RolePermissionCreateRequest = {
        role: model.value.role,
        permission_type: model.value.permission_type,
        permission_id: model.value.permission_id || 0,
        instance_id: '',
        schema: '',
        table: ''
      };
      await fetchCreateRolePermission(submitData);
      window.$message?.success($t('common.addSuccess'));
      closeModal();
      emit('submitted');
    } else {
      // 直接权限：支持批量创建表权限
      const selectedTables = (model.value as any).selectedTables || [];
      
      if (selectedTables.length > 0) {
        // 批量创建表权限
        const promises = selectedTables.map((tableName: string) => {
          const submitData: Api.Das.RolePermissionCreateRequest = {
            role: model.value.role,
            permission_type: model.value.permission_type,
            permission_id: 0, // 由后端生成
            instance_id: model.value.instance_id,
            schema: model.value.schema,
            table: tableName
          };
          return fetchCreateRolePermission(submitData);
        });
        
        await Promise.all(promises);
        window.$message?.success(`成功添加 ${selectedTables.length} 个表权限`);
      } else {
        // 添加整个库的权限
        const submitData: Api.Das.RolePermissionCreateRequest = {
          role: model.value.role,
          permission_type: model.value.permission_type,
          permission_id: 0, // 由后端生成
          instance_id: model.value.instance_id,
          schema: model.value.schema,
          table: '' // 空表示整个库
        };
        await fetchCreateRolePermission(submitData);
        window.$message?.success($t('common.addSuccess'));
      }
      
      closeModal();
      emit('submitted');
    }
  } catch (error) {
    window.$message?.error($t('common.operationFailed') || '操作失败');
  }
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    loadTemplates();
    loadDBConfigs();
  }
});

watch(
  () => model.value.permission_type,
  newVal => {
    // 切换权限类型时，清空相关字段
    model.value.permission_id = null;
    model.value.instance_id = '';
    model.value.schema = '';
    model.value.table = '';
    (model.value as any).selectedTables = [];
    schemaOptions.value = [];
  }
);

watch(
  () => model.value.instance_id,
  newVal => {
    if (newVal) {
      loadSchemas(newVal);
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
  newVal => {
    if (!newVal) {
      model.value.table = '';
      (model.value as any).selectedTables = [];
    }
  }
);

watch(
  () => model.value.permission_id,
  (newVal) => {
    // 当 permission_id 改变时，手动触发验证以清除错误
    if (newVal && newVal > 0) {
      // 延迟一下，确保值已经更新
      setTimeout(() => {
        formRef.value?.validate(undefined, () => {});
      }, 0);
    }
  }
);

onMounted(() => {
  loadTemplates();
  loadDBConfigs();
});
</script>

<template>
  <NModal
    v-model:show="visible"
    :title="title"
    preset="card"
    :style="{ width: '700px' }"
    :mask-closable="false"
  >
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
      <NFormItem label="角色" path="role">
        <NInput v-model:value="model.role" :disabled="true" clearable />
      </NFormItem>
      <NFormItem label="权限类型" path="permission_type">
        <NSelect
          v-model:value="model.permission_type"
          :options="permissionTypeOptions"
          placeholder="请选择权限类型"
        />
      </NFormItem>
      <NFormItem
        v-if="model.permission_type === 'template'"
        label="权限模板"
        path="permission_id"
      >
        <NSelect
          v-model:value="model.permission_id"
          :options="templateOptions"
          placeholder="请选择权限模板"
          clearable
          @update:value="(val) => {
            model.permission_id = val as number | null;
            // 值改变后，延迟验证以清除错误
            setTimeout(() => {
              formRef?.validate(undefined, () => {});
            }, 100);
          }"
        />
      </NFormItem>
      <template v-if="model.permission_type === 'object'">
        <NFormItem label="实例ID" path="instance_id">
          <NSelect
            v-model:value="model.instance_id"
            :options="dbConfigOptions"
            placeholder="请选择实例"
            clearable
          />
        </NFormItem>
        <NFormItem label="库名" path="schema">
          <NSelect
            v-model:value="model.schema"
            :options="schemaOptions"
            placeholder="请选择库名"
            :disabled="!model.instance_id"
            clearable
          />
        </NFormItem>
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

