<script setup lang="ts">
import { ref } from 'vue';
import { NTabs, NTabPane } from 'naive-ui';
import { $t } from '@/locales';
import RolePermissionTab from './modules/role-permission-tab.vue';
import UserPermissionTab from './modules/user-permission-tab.vue';
import PermissionTemplateTab from './modules/permission-template-tab.vue';

defineOptions({
  name: 'DatabasePermission'
});

const activeTab = ref('role');
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NTabs v-model:value="activeTab" type="line" animated class="flex-1">
      <NTabPane name="role" :tab="$t('page.manage.database.permissionAssign.roleTab')">
        <RolePermissionTab />
      </NTabPane>
      <NTabPane name="user" :tab="$t('page.manage.database.permissionAssign.userTab')">
        <UserPermissionTab />
      </NTabPane>
      <NTabPane name="template" :tab="$t('page.manage.database.permissionTemplate.title')">
        <PermissionTemplateTab />
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
:deep(.n-tabs-pane-wrapper) {
  flex: 1;
  display: flex;
  flex-direction: column;
}

:deep(.n-tab-pane) {
  flex: 1;
  display: flex;
  flex-direction: column;
}
</style>
