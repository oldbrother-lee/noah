<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue';
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NSelect,
  NDataTable
} from 'naive-ui';
import { fetchCreateUserPermission } from '@/service/api/das';
import { fetchGetPermissionTemplates } from '@/service/api/das';
import { fetchGetDBConfigs } from '@/service/api/admin';
import { fetchOrdersSchemas, fetchOrderTables } from '@/service/api/orders';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'UserPermissionOperateModal'
});

interface Props {
  username: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{ (e: 'submitted'): void }>();

const visible = defineModel<boolean>('visible', { default: false });

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => '新增用户权限');

type Model = {
  username: string;
  permission_type: 'object' | 'template';
  permission_id: number | null;
  instance_id: string;
  schema: string;
  table: string;
};

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    username: '',
    permission_type: 'object',
    permission_id: null,
    instance_id: '',
    schema: '',
    table: ''
  };
}

const rules = computed(() => {
  const base: Record<string, App.Global.FormRule> = {
    username: defaultRequiredRule,
    permission_type: defaultRequiredRule
  };
  if (model.value.permission_type === 'template') {
    base.permission_id = {
      required: true,
      message: '请选择权限模板',
      trigger: ['blur', 'change'],
      validator: (_r: any, v: any) => (v == null || v === '' || Number(v) <= 0 ? new Error('请选择权限模板') : true)
    };
  } else {
    base.instance_id = defaultRequiredRule;
    base.schema = defaultRequiredRule;
  }
  return base;
});

const permissionTypeOptions = [
  { label: '直接权限', value: 'object' },
  { label: '权限模板', value: 'template' }
];

const templateOptions = ref<{ label: string; value: number }[]>([]);
const dbConfigOptions = ref<{ label: string; value: string }[]>([]);
const schemaOptions = ref<{ label: string; value: string }[]>([]);
const showTableModal = ref(false);
const tableList = ref<any[]>([]);
const selectedTableKeys = ref<string[]>([]);
const tableSearchText = ref('');
const tableLoading = ref(false);
const tablePagination = ref({ page: 1, pageSize: 30, itemCount: 0 });

async function loadTemplates() {
  try {
    const res = await fetchGetPermissionTemplates();
    const data = (res as any)?.data ?? res;
    const list = Array.isArray(data) ? data : [];
    templateOptions.value = list.map((t: any) => ({
      label: `${t.name} (${t.permissions?.length ?? 0}项)`,
      value: Number(t.id ?? t.ID)
    }));
  } catch (e) {
    console.error(e);
  }
}

async function loadDBConfigs() {
  try {
    const res = await fetchGetDBConfigs({ useType: '查询' } as any);
    const data = (res as any)?.data ?? res;
    const list = Array.isArray(data) ? data : [];
    dbConfigOptions.value = list.map((c: any) => ({
      label: `${c.hostname}:${c.port}${c.instance_id ? ` (${c.instance_id})` : ''}`,
      value: c.instance_id
    }));
  } catch (e) {
    console.error(e);
  }
}

async function loadSchemas(instanceId: string) {
  if (!instanceId) {
    schemaOptions.value = [];
    return;
  }
  try {
    const res = await fetchOrdersSchemas({ instance_id: instanceId });
    const data = (res as any)?.data ?? res;
    const list = Array.isArray(data) ? data : [];
    schemaOptions.value = list.map((s: any) => ({
      label: s.schema ?? s.name,
      value: s.schema ?? s.name
    }));
  } catch (e) {
    schemaOptions.value = [];
  }
}

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

async function loadTables() {
  if (!model.value.instance_id || !model.value.schema) return;
  tableLoading.value = true;
  try {
    const res = await fetchOrderTables({ instance_id: model.value.instance_id, schema: model.value.schema });
    const data = (res as any)?.data ?? res;
    let list = Array.isArray(data) ? data : [];
    if (tableSearchText.value) {
      const kw = tableSearchText.value.toLowerCase();
      list = list.filter((t: any) => (t.table_name ?? t.name ?? t.TableName ?? '').toLowerCase().includes(kw));
    }
    tablePagination.value.itemCount = list.length;
    const start = (tablePagination.value.page - 1) * tablePagination.value.pageSize;
    tableList.value = list.slice(start, start + tablePagination.value.pageSize);
  } catch (e) {
    tableList.value = [];
  } finally {
    tableLoading.value = false;
  }
}

function confirmTableSelection() {
  if (selectedTableKeys.value.length === 0) {
    window.$message?.warning('请至少选择一个表');
    return;
  }
  const [, tableName] = selectedTableKeys.value[0].split('#');
  model.value.table = tableName;
  (model.value as any).selectedTables = selectedTableKeys.value.map((k: string) => k.split('#')[1]);
  showTableModal.value = false;
}

function handleAddSchemaPermission() {
  if (!model.value.instance_id || !model.value.schema) return;
  model.value.table = '';
  (model.value as any).selectedTables = [];
}

const tableColumns = [
  { type: 'selection', width: 48 },
  { key: 'table_schema', title: '库名称', width: 200, render: () => model.value.schema || '-' },
  { key: 'table_name', title: '表名称', width: 200, render: (row: any) => row.table_name ?? row.name ?? row.TableName ?? '' },
  { key: 'table_comment', title: '描述', width: 200, render: (row: any) => row.table_comment ?? row.comment ?? row.TableComment ?? '-' }
];

function handleInitModel() {
  model.value = createDefaultModel();
  model.value.username = props.username;
  (model.value as any).selectedTables = [];
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  try {
    const username = model.value.username;
    if (model.value.permission_type === 'template') {
      await fetchCreateUserPermission({
        username,
        permission_type: 'template',
        permission_id: model.value.permission_id!,
        instance_id: '',
        schema: '',
        table: ''
      });
      window.$message?.success($t('common.addSuccess'));
    } else {
      const selectedTables = (model.value as any).selectedTables || [];
      if (selectedTables.length > 0) {
        await Promise.all(
          selectedTables.map((tableName: string) =>
            fetchCreateUserPermission({
              username,
              permission_type: 'object',
              permission_id: 0,
              instance_id: model.value.instance_id,
              schema: model.value.schema,
              table: tableName
            })
          )
        );
        window.$message?.success(`成功添加 ${selectedTables.length} 个表权限`);
      } else {
        await fetchCreateUserPermission({
          username,
          permission_type: 'object',
          permission_id: 0,
          instance_id: model.value.instance_id,
          schema: model.value.schema,
          table: ''
        });
        window.$message?.success($t('common.addSuccess'));
      }
    }
    closeModal();
    emit('submitted');
  } catch (e) {
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
  () => {
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
  (val) => {
    if (val) loadSchemas(val);
    else {
      schemaOptions.value = [];
      model.value.schema = '';
      model.value.table = '';
      (model.value as any).selectedTables = [];
    }
  }
);

watch(
  () => model.value.schema,
  (val) => {
    if (!val) {
      model.value.table = '';
      (model.value as any).selectedTables = [];
    }
  }
);

onMounted(() => {
  loadTemplates();
  loadDBConfigs();
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" :style="{ width: '700px' }" :mask-closable="false">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="120">
      <NFormItem label="用户名" path="username">
        <NInput v-model:value="model.username" disabled clearable />
      </NFormItem>
      <NFormItem label="权限类型" path="permission_type">
        <NSelect v-model:value="model.permission_type" :options="permissionTypeOptions" placeholder="请选择权限类型" />
      </NFormItem>
      <NFormItem v-if="model.permission_type === 'template'" label="权限模板" path="permission_id">
        <NSelect
          v-model:value="model.permission_id"
          :options="templateOptions"
          placeholder="请选择权限模板"
          clearable
          @update:value="() => formRef?.validate(undefined, () => {})"
        />
      </NFormItem>
      <template v-if="model.permission_type === 'object'">
        <NFormItem label="实例ID" path="instance_id">
          <NSelect v-model:value="model.instance_id" :options="dbConfigOptions" placeholder="请选择实例" clearable />
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
            <NButton type="primary" :disabled="!model.instance_id || !model.schema" @click="openTableModal">选择表</NButton>
            <NButton :disabled="!model.instance_id || !model.schema" @click="handleAddSchemaPermission">添加整个库</NButton>
          </NSpace>
          <div v-if="model.table || (model as any).selectedTables?.length" class="mt-8px text-12px text-gray-500">
            {{ (model as any).selectedTables?.length ? `已选择 ${(model as any).selectedTables.length} 个表` : model.table ? `当前表: ${model.table}` : '整个库' }}
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
  <NModal v-model:show="showTableModal" title="选择表" preset="card" :style="{ width: '800px' }" :mask-closable="false">
    <div class="flex flex-col gap-16px">
      <div class="flex gap-8px items-center">
        <span class="w-80px">表名称:</span>
        <NInput v-model:value="tableSearchText" placeholder="请输入表名" clearable style="flex: 1" @keyup.enter="loadTables" />
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
        @update:page="(p: number) => { tablePagination.page = p; loadTables(); }"
        @update:page-size="(ps: number) => { tablePagination.pageSize = ps; tablePagination.page = 1; loadTables(); }"
      />
      <div class="text-12px text-gray-500">已选择 {{ selectedTableKeys.length }} 个表</div>
    </div>
    <template #footer>
      <NSpace :size="16">
        <NButton @click="showTableModal = false">取消</NButton>
        <NButton type="primary" @click="confirmTableSelection">确定</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
