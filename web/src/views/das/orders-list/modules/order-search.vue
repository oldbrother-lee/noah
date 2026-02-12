<script setup lang="ts">
import { onMounted, ref, toRaw } from 'vue';
import { jsonClone } from '@sa/utils';
import { fetchOrdersEnvironments } from '@/service/api/orders';
import { useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'OrderSearch'
});

interface Emits {
  (e: 'search'): void;
  (e: 'reset'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

const model = defineModel<Record<string, any>>('model', { required: true });

// 环境选项
const environmentOptions = ref<{ label: string; value: string }[]>([]);

async function getEnvironments() {
  const { data: envs } = await fetchOrdersEnvironments();
  if (envs) {
    environmentOptions.value = envs.map(item => ({
      label: item.name,
      value: item.name
    }));
  }
}

// 状态选项
const statusOptions = [
  { label: '待审批', value: '待审批' },
  { label: '已驳回', value: '已驳回' },
  { label: '待执行', value: '待执行' },
  { label: '执行中', value: '执行中' },
  { label: '已完成', value: '已完成' },
  { label: '已失败', value: '已失败' }
];

const defaultModel = jsonClone(toRaw(model.value));

function resetModel() {
  Object.assign(model.value, defaultModel);
}

async function reset() {
  await restoreValidation();
  resetModel();
  emit('reset');
}

async function search() {
  await validate();
  emit('search');
}

onMounted(() => {
  getEnvironments();
});
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="order-search">
        <NForm ref="formRef" :model="model" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi span="24 s:12 m:6" label="环境" path="environment" class="pr-24px">
              <NSelect
                v-model:value="model.environment"
                placeholder="请选择环境"
                :options="environmentOptions"
                clearable
              />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" label="状态" path="status" class="pr-24px">
              <NSelect v-model:value="model.status" placeholder="请选择状态" :options="statusOptions" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" label="搜索" path="search" class="pr-24px">
              <NInput v-model:value="model.search" placeholder="搜索工单标题、申请人..." clearable />
            </NFormItemGi>
            <NFormItemGi span="24 m:6" class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="reset">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  {{ $t('common.reset') }}
                </NButton>
                <NButton type="primary" ghost @click="search">
                  <template #icon>
                    <icon-ic-round-search class="text-icon" />
                  </template>
                  {{ $t('common.search') }}
                </NButton>
              </NSpace>
            </NFormItemGi>
          </NGrid>
        </NForm>
      </NCollapseItem>
    </NCollapse>
  </NCard>
</template>

<style scoped></style>
