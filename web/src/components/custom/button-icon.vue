<script setup lang="ts">
import type { PopoverPlacement } from 'naive-ui';
import { twMerge } from 'tailwind-merge';

defineOptions({
  name: 'ButtonIcon',
  inheritAttrs: false
});

interface Props {
  /** Button class */
  class?: string;
  /** Iconify icon name（会请求在线接口，首屏可能较慢） */
  icon?: string;
  /** 本地 SVG 图标名（优先使用，不请求网络） */
  localIcon?: string;
  /** Tooltip content */
  tooltipContent?: string;
  /** Tooltip placement */
  tooltipPlacement?: PopoverPlacement;
  zIndex?: number;
}

const props = withDefaults(defineProps<Props>(), {
  class: '',
  icon: '',
  localIcon: '',
  tooltipContent: '',
  tooltipPlacement: 'bottom',
  zIndex: 98
});

const DEFAULT_CLASS = 'h-[36px] text-icon';
</script>

<template>
  <NTooltip :placement="tooltipPlacement" :z-index="zIndex" :disabled="!tooltipContent">
    <template #trigger>
      <NButton quaternary :class="twMerge(DEFAULT_CLASS, props.class)" v-bind="$attrs">
        <div class="flex-center gap-8px">
          <slot>
            <SvgIcon :icon="localIcon ? undefined : icon" :local-icon="localIcon" />
          </slot>
        </div>
      </NButton>
    </template>
    {{ tooltipContent }}
  </NTooltip>
</template>

<style scoped></style>
