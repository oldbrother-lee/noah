<script setup lang="ts">
import { computed } from 'vue';
import type { PopoverPlacement } from 'naive-ui';
import { $t } from '@/locales';

defineOptions({ name: 'ThemeSchemaSwitch' });

interface Props {
  /** Theme schema */
  themeSchema: UnionKey.ThemeScheme;
  /** Show tooltip */
  showTooltip?: boolean;
  /** Tooltip placement */
  tooltipPlacement?: PopoverPlacement;
}

const props = withDefaults(defineProps<Props>(), {
  showTooltip: true,
  tooltipPlacement: 'bottom'
});

interface Emits {
  (e: 'switch'): void;
}

const emit = defineEmits<Emits>();

function handleSwitch() {
  emit('switch');
}

const localIcons: Record<UnionKey.ThemeScheme, string> = {
  light: 'sun',
  dark: 'moon',
  auto: 'brightness-auto'
};

const localIcon = computed(() => localIcons[props.themeSchema]);

const tooltipContent = computed(() => {
  if (!props.showTooltip) return '';

  return $t('icon.themeSchema');
});
</script>

<template>
  <ButtonIcon
    :local-icon="localIcon"
    :tooltip-content="tooltipContent"
    :tooltip-placement="tooltipPlacement"
    @click="handleSwitch"
  />
</template>

<style scoped></style>
