<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, NSpace } from 'naive-ui';
import { jsonClone } from '@sa/utils';
import { fetchCreateEnvironment, fetchUpdateEnvironment } from '@/service/api/admin';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'EnvironmentOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Admin.Environment | null;
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
    add: $t('page.manage.database.environment.addEnvironment'),
    edit: $t('page.manage.database.environment.editEnvironment')
  };
  return titles[props.operateType];
});

type Model = Api.Admin.EnvironmentCreateRequest;

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: ''
  };
}

type RuleKey = Extract<keyof Model, 'name'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule
};

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, jsonClone(props.rowData));
  }
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (props.operateType === 'edit' && props.rowData) {
      await fetchUpdateEnvironment(props.rowData.id, model.value);
      window.$message?.success($t('common.updateSuccess'));
    } else {
      await fetchCreateEnvironment(model.value);
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
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" :style="{ width: '480px' }" :mask-closable="false">
    <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="100">
      <NFormItem :label="$t('page.manage.database.environment.name')" path="name">
        <NInput
          v-model:value="model.name"
          :placeholder="$t('page.manage.database.environment.form.name')"
          clearable
        />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace :size="16">
        <NButton @click="closeModal">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>

