<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { jsonClone } from '@sa/utils';
import { fetchCreateApi, fetchUpdateApi } from '@/service/api/admin';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'ApiOperateDrawer'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Admin.Api | null;
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
    add: $t('page.manage.api.addApi'),
    edit: $t('page.manage.api.editApi')
  };
  return titles[props.operateType];
});

type Model = {
  group: string;
  name: string;
  path: string;
  method: string;
};

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    group: '',
    name: '',
    path: '',
    method: 'GET'
  };
}

const rules = {
  name: defaultRequiredRule,
  path: defaultRequiredRule,
  method: defaultRequiredRule
};

const methodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' }
];

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, jsonClone(props.rowData));
  }
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (props.operateType === 'edit') {
      await fetchUpdateApi({
        id: props.rowData!.id,
        group: model.value.group,
        name: model.value.name,
        path: model.value.path,
        method: model.value.method
      });
      window.$message?.success($t('common.updateSuccess'));
    } else {
      await fetchCreateApi({
        group: model.value.group,
        name: model.value.name,
        path: model.value.path,
        method: model.value.method
      });
      window.$message?.success($t('common.addSuccess'));
    }
    closeDrawer();
    emit('submitted');
  } catch (error) {
    window.$message?.error($t('common.operationFailed') || '操作失败');
  }
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
  }
});
</script>

<template>
  <NDrawer v-model:show="visible" display-directive="show" :width="400">
    <NDrawerContent :title="title" :native-scrollbar="false" closable>
      <NForm ref="formRef" :model="model" :rules="rules">
        <NFormItem :label="$t('page.manage.api.group')" path="group">
          <NInput v-model:value="model.group" :placeholder="$t('page.manage.api.form.group')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.api.name')" path="name">
          <NInput v-model:value="model.name" :placeholder="$t('page.manage.api.form.name')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.api.path')" path="path">
          <NInput v-model:value="model.path" :placeholder="$t('page.manage.api.form.path')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.api.method')" path="method">
          <NSelect v-model:value="model.method" :options="methodOptions" :placeholder="$t('page.manage.api.form.method')" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace :size="16">
          <NButton @click="closeDrawer">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped></style>
