<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const activeTab = ref('edit');

const tabs = [
  {
    key: 'edit',
    label: t('route.das_modules_edit'),
    icon: 'i-carbon-edit'
  },
  {
    key: 'favorite',
    label: t('route.das_modules_favorite'),
    icon: 'i-carbon-favorite'
  },
  {
    key: 'history',
    label: t('route.das_modules_history'),
    icon: 'i-carbon-time'
  }
];

const EditComponent = defineAsyncComponent(() => import('./edit/index.vue'));
const FavoriteComponent = defineAsyncComponent(() => import('./favorite/index.vue'));
const HistoryComponent = defineAsyncComponent(() => import('./history/index.vue'));
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NTabs v-model:value="activeTab" type="line" size="small">
      <NTabPane v-for="tab in tabs" :key="tab.key" :name="tab.key" :tab="tab.label">
        <template #tab>
          <NSpace align="center" :size="4">
            <SvgIcon :icon="tab.icon" class="text-16px" />
            <span>{{ tab.label }}</span>
          </NSpace>
        </template>
        <div v-if="tab.key === 'edit'" class="tab-content">
          <EditComponent />
        </div>
        <div v-else-if="tab.key === 'favorite'" class="tab-content">
          <FavoriteComponent />
        </div>
        <div v-else-if="tab.key === 'history'" class="tab-content">
          <HistoryComponent />
        </div>
      </NTabPane>
    </NTabs>
  </NCard>
</template>

<style scoped>
.card-wrapper {
  height: calc(100vh - 120px);
}

.tab-content {
  height: calc(100vh - 200px);
  overflow-y: auto;
}
</style>
