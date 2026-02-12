<script setup lang="ts">
import { computed, toRaw } from 'vue';
import { jsonClone } from '@sa/utils';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'UserSearch'
});

interface Emits {
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

interface SearchParams {
  page: number;
  pageSize: number;
  username?: string;
  nickname?: string;
  phone?: string;
  email?: string;
}

const model = defineModel<SearchParams>('model', { required: true });

type RuleKey = Extract<keyof SearchParams, 'email' | 'phone'>;

const rules = computed<Record<RuleKey, App.Global.FormRule>>(() => {
  const { patternRules } = useFormRules();
  return {
    email: patternRules.email,
    phone: patternRules.phone
  };
});

const defaultModel = jsonClone(toRaw(model.value));

function resetModel() {
  Object.assign(model.value, defaultModel);
  model.value.page = 1;
  model.value.pageSize = 10;
}

async function reset() {
  await restoreValidation();
  resetModel();
}

async function search() {
  await validate();
  model.value.page = 1;
  emit('search');
}
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="user-search">
        <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.manage.user.username')" path="username" class="pr-24px">
              <NInput v-model:value="model.username" :placeholder="$t('page.manage.user.form.username')" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.manage.user.nickname')" path="nickname" class="pr-24px">
              <NInput v-model:value="model.nickname" :placeholder="$t('page.manage.user.form.nickname')" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.manage.user.phone')" path="phone" class="pr-24px">
              <NInput v-model:value="model.phone" :placeholder="$t('page.manage.user.form.phone')" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.manage.user.email')" path="email" class="pr-24px">
              <NInput v-model:value="model.email" :placeholder="$t('page.manage.user.form.email')" clearable />
            </NFormItemGi>
            <NFormItemGi span="24 m:12" class="pr-24px">
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

