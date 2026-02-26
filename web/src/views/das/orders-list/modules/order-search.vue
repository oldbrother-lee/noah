<script setup lang="ts">
import { onMounted, ref, toRaw, watch } from 'vue';
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

// 环境选项（value 为环境 ID，与后端 environment 参数一致）
const environmentOptions = ref<{ label: string; value: number }[]>([]);

async function getEnvironments() {
  const { data: envs } = await fetchOrdersEnvironments();
  if (envs && Array.isArray(envs)) {
    environmentOptions.value = envs.map((item: any) => {
      const id = item.id ?? item.ID;
      const name = item.name ?? item.Name ?? '';
      return { label: name, value: Number(id) };
    }).filter(opt => !Number.isNaN(opt.value));
  }
}

// 环境加载后：若当前值是字符串（旧数据）或无效数字，归一为 null 或合法 ID，避免 NSelect 显示“全选”或选择无反应
watch(environmentOptions, opts => {
  if (opts.length === 0) return;
  const v = model.value.environment;
  if (v === undefined || v === null) return;
  if (typeof v === 'string') {
    const byName = opts.find(o => o.label === v);
    model.value.environment = byName ? byName.value : null;
  } else if (typeof v === 'number' && !opts.some(o => o.value === v)) {
    model.value.environment = null;
  }
}, { immediate: true });

// 进度选项（与列表展示的 progress 一致）
const progressOptions = [
  { label: '待审核', value: '待审核' },
  { label: '已批准', value: '已批准' },
  { label: '已驳回', value: '已驳回' },
  { label: '执行中', value: '执行中' },
  { label: '已完成', value: '已完成' },
  { label: '已关闭', value: '已关闭' }
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
            <NFormItemGi span="24 s:12 m:6" label="进度" path="progress" class="pr-24px">
              <NSelect v-model:value="model.progress" placeholder="请选择进度" :options="progressOptions" clearable />
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
