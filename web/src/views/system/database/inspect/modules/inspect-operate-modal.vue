<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, NSpace } from 'naive-ui';
import { jsonClone } from '@sa/utils';
import { fetchCreateInspectParam, fetchUpdateInspectParam, fetchGetDefaultInspectParams } from '@/service/api/inspect';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'InspectOperateModal'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Inspect.InspectParam | null;
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
    add: '添加审核参数',
    edit: '编辑审核参数'
  };
  return titles[props.operateType];
});

type Model = {
  params: string; // JSON 字符串
  remark: string;
};

const model = ref<Model>(createDefaultModel());
const loading = ref(false);

function createDefaultModel(): Model {
  return {
    params: '{}',
    remark: ''
  };
}

type RuleKey = Extract<keyof Model, 'params' | 'remark'>;

const rules: Record<RuleKey, App.Global.FormRule | App.Global.FormRule[]> = {
  params: [
    defaultRequiredRule,
    {
      validator: (_rule, value: string) => {
        if (!value || value.trim() === '') {
          return new Error('参数配置不能为空');
        }
        try {
          JSON.parse(value);
          return true;
        } catch (e) {
          return new Error('参数配置必须是有效的 JSON 格式');
        }
      },
      trigger: ['input', 'blur']
    }
  ],
  remark: defaultRequiredRule
};

// 加载默认参数
async function loadDefaultParams() {
  if (props.operateType === 'add' && !model.value.params || model.value.params === '{}') {
    try {
      loading.value = true;
      const res = await fetchGetDefaultInspectParams();
      const defaultParams = (res as any)?.data || res;
      if (defaultParams && typeof defaultParams === 'object') {
        model.value.params = JSON.stringify(defaultParams, null, 2);
      }
    } catch (error) {
      console.error('加载默认参数失败:', error);
    } finally {
      loading.value = false;
    }
  }
}

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    const rowData = jsonClone(props.rowData);
    // 处理 params：GORM 的 datatypes.JSON 会自动反序列化为对象
    if (rowData.params) {
      if (typeof rowData.params === 'object') {
        model.value.params = JSON.stringify(rowData.params, null, 2);
      } else if (typeof rowData.params === 'string') {
        try {
          // 如果是字符串，先解析再格式化
          const parsed = JSON.parse(rowData.params);
          model.value.params = JSON.stringify(parsed, null, 2);
        } catch (e) {
          model.value.params = rowData.params;
        }
      } else {
        model.value.params = '{}';
      }
    } else {
      model.value.params = '{}';
    }
    model.value.remark = rowData.remark || '';
  } else {
    // 新增时加载默认参数
    loadDefaultParams();
  }
}

function closeModal() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    // 解析 JSON 字符串为对象
    let paramsObj: Record<string, any>;
    try {
      paramsObj = JSON.parse(model.value.params);
    } catch (e) {
      window.$message?.error('参数配置格式错误，请检查 JSON 格式');
      return;
    }

    if (props.operateType === 'edit' && props.rowData) {
      const id = (props.rowData as any).ID || props.rowData.id;
      await fetchUpdateInspectParam(id, {
        params: paramsObj,
        remark: model.value.remark
      });
      window.$message?.success($t('common.updateSuccess'));
    } else {
      await fetchCreateInspectParam({
        params: paramsObj,
        remark: model.value.remark
      });
      window.$message?.success($t('common.addSuccess'));
    }
    closeModal();
    emit('submitted');
  } catch (error: any) {
    window.$message?.error(error?.message || $t('common.operationFailed') || '操作失败');
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
  <NModal
    v-model:show="visible"
    :title="title"
    preset="card"
    :style="{ width: '600px' }"
    :mask-closable="false"
  >
    <NForm
      ref="formRef"
      :model="model"
      :rules="rules"
      label-placement="left"
      :label-width="100"
    >
      <NFormItem label="备注" path="remark">
        <NInput
          v-model:value="model.remark"
          placeholder="请输入备注（如：表名的长度）"
          clearable
        />
      </NFormItem>
      <NFormItem label="参数配置" path="params">
        <NInput
          v-model:value="model.params"
          type="textarea"
          placeholder='请输入 JSON 格式的参数配置，例如：{"MAX_TABLE_NAME_LENGTH": 32}'
          :autosize="{ minRows: 12, maxRows: 20 }"
          clearable
          :loading="loading"
        />
        <template #feedback>
          <div class="text-gray-400 text-xs mt-4px">
            提示：参数配置必须是有效的 JSON 格式。可以使用默认参数作为模板。
          </div>
        </template>
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace :size="16">
        <NButton @click="closeModal">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" @click="handleSubmit" :loading="loading">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
