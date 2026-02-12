<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

interface Props {
  modelValue: string;
  language?: string;
  theme?: string;
  height?: string;
}

interface Emits {
  (e: 'update:modelValue', value: string): void;
  (e: 'change', value: string): void;
  (e: 'ready', editor: any): void;
}

const props = withDefaults(defineProps<Props>(), {
  language: 'sql',
  theme: 'vs-dark',
  height: '300px'
});

const emit = defineEmits<Emits>();

const editorRef = ref<HTMLElement>();
const editor = ref<any>();
const isEditorReady = ref(false);
const isInitializing = ref(false);

// 初始化编辑器
const initEditor = async () => {
  if (!editorRef.value || isInitializing.value) return;

  isInitializing.value = true;

  try {
    // 动态导入Monaco Editor
    const monaco = await import('monaco-editor');

    // 创建编辑器 - 使用最简单的配置
    editor.value = monaco.editor.create(editorRef.value, {
      value: props.modelValue || '',
      language: props.language,
      theme: props.theme,
      fontSize: 14,
      lineNumbers: 'on',
      wordWrap: 'on',
      minimap: { enabled: false },
      automaticLayout: true,
      scrollBeyondLastLine: false,
      readOnly: false,
      selectOnLineNumbers: true,
      roundedSelection: false,
      cursorStyle: 'line'
    });

    // 监听内容变化 - 添加防抖
    let changeTimeout: any = null;
    editor.value.onDidChangeModelContent(() => {
      if (changeTimeout) {
        clearTimeout(changeTimeout);
      }
      changeTimeout = setTimeout(() => {
        const value = editor.value.getValue();
        emit('update:modelValue', value);
        emit('change', value);
      }, 100);
    });

    isEditorReady.value = true;
    emit('ready', editor.value);
  } catch (error) {
    console.error('Failed to initialize Monaco Editor:', error);
  } finally {
    isInitializing.value = false;
  }
};

// 监听modelValue变化 - 添加防抖
let updateTimeout: any = null;
watch(
  () => props.modelValue,
  newValue => {
    if (editor.value && editor.value.getValue() !== newValue) {
      if (updateTimeout) {
        clearTimeout(updateTimeout);
      }
      updateTimeout = setTimeout(() => {
        if (editor.value) {
          editor.value.setValue(newValue || '');
        }
      }, 50);
    }
  }
);

// 监听主题变化
watch(
  () => props.theme,
  newTheme => {
    if (editor.value && newTheme) {
      editor.value.updateOptions({ theme: newTheme });
    }
  }
);

onMounted(() => {
  nextTick(() => {
    initEditor();
  });
});

onUnmounted(() => {
  if (editor.value) {
    try {
      editor.value.dispose();
    } catch (error) {
      console.error('Error disposing editor:', error);
    }
  }
});

// 暴露编辑器实例
defineExpose({
  editor: () => editor.value,
  isReady: () => isEditorReady.value
});
</script>

<template>
  <div ref="editorRef" class="monaco-editor-container" :style="{ height: height }" />
</template>

<style scoped>
.monaco-editor-container {
  width: 100%;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  overflow: hidden;
}
</style>
