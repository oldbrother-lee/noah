<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { jsonClone } from '@sa/utils';
import { fetchGetRoles, fetchCreateAdminUser, fetchUpdateAdminUser } from '@/service/api/admin';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'UserOperateDrawer'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Admin.AdminUser | null;
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
const { defaultRequiredRule, patternRules } = useFormRules();

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.manage.user.addUser'),
    edit: $t('page.manage.user.editUser')
  };
  return titles[props.operateType];
});

type Model = Omit<Api.Admin.AdminUserCreateRequest, 'password'> & {
  password?: string;
};

const model = ref<Model>(createDefaultModel());

function createDefaultModel(): Model {
  return {
    username: '',
    nickname: '',
    email: '',
    phone: '',
    roles: []
  };
}

type RuleKey = Extract<keyof Model, 'username' | 'email'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  username: defaultRequiredRule,
  email: patternRules.email
};

/** the enabled role options */
const roleOptions = ref<CommonType.Option<string>[]>([]);

async function getRoleOptions() {
  try {
    const { data } = await fetchGetRoles({ page: 1, pageSize: 1000 });
    if (data && data.list) {
      roleOptions.value = data.list.map(item => ({
        label: item.name,
        value: item.sid
      }));
    }
  } catch (error) {
    console.error('Failed to get roles:', error);
  }
}

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    const { password, ...rest } = props.rowData;
    Object.assign(model.value, jsonClone(rest));
    // 编辑时不设置密码，密码字段留空
    model.value.password = undefined;
  }
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (props.operateType === 'edit') {
      const updateData: Api.Admin.AdminUserUpdateRequest = {
        id: props.rowData!.id,
        username: model.value.username,
        nickname: model.value.nickname,
        email: model.value.email,
        phone: model.value.phone,
        roles: model.value.roles
      };
      // 如果密码不为空，则更新密码
      if (model.value.password && model.value.password.trim()) {
        updateData.password = model.value.password;
      }
      await fetchUpdateAdminUser(updateData);
      window.$message?.success($t('common.updateSuccess'));
    } else {
      if (!model.value.password || !model.value.password.trim()) {
        window.$message?.error($t('page.manage.user.form.passwordRequired') || '密码不能为空');
        return;
      }
      await fetchCreateAdminUser({
        username: model.value.username,
        nickname: model.value.nickname,
        password: model.value.password,
        email: model.value.email,
        phone: model.value.phone,
        roles: model.value.roles
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
    getRoleOptions();
  }
});
</script>

<template>
  <NDrawer v-model:show="visible" display-directive="show" :width="360">
    <NDrawerContent :title="title" :native-scrollbar="false" closable>
      <NForm ref="formRef" :model="model" :rules="rules">
        <NFormItem :label="$t('page.manage.user.username')" path="username">
          <NInput v-model:value="model.username" :placeholder="$t('page.manage.user.form.username')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.nickname')" path="nickname">
          <NInput v-model:value="model.nickname" :placeholder="$t('page.manage.user.form.nickname')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.password')" path="password">
          <NInput
            v-model:value="model.password"
            type="password"
            show-password-on="click"
            :placeholder="
              operateType === 'edit'
                ? $t('page.manage.user.form.passwordPlaceholder')
                : $t('page.manage.user.form.password')
            "
          />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.email')" path="email">
          <NInput v-model:value="model.email" :placeholder="$t('page.manage.user.form.email')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.phone')" path="phone">
          <NInput v-model:value="model.phone" :placeholder="$t('page.manage.user.form.phone')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.roles')" path="roles">
          <NSelect
            v-model:value="model.roles"
            multiple
            :options="roleOptions"
            :placeholder="$t('page.manage.user.form.roles')"
          />
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

