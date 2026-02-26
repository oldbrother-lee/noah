<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { NPagination } from 'naive-ui';
// CodeMirror 6 imports
import { EditorState } from '@codemirror/state';
import { EditorView, lineNumbers } from '@codemirror/view';
import { sql } from '@codemirror/lang-sql';
import { defaultHighlightStyle, foldGutter, syntaxHighlighting } from '@codemirror/language';
import { oneDark } from '@codemirror/theme-one-dark';

interface Props {
  /** SQL内容，支持多条SQL语句 */
  sqlContent?: string;
  /** 是否显示分页 */
  showPagination?: boolean;
  /** 每页显示的SQL语句数量 */
  pageSize?: number;
  /** 编辑器主题 */
  theme?: 'light' | 'dark';
  /** 编辑器高度 (传入 'auto' 则自适应内容) */
  height?: string;
  /** 自适应模式下的最大高度 */
  maxHeight?: string;
  /** 自适应模式下的最小高度 */
  minHeight?: string;
}

const props = withDefaults(defineProps<Props>(), {
  sqlContent: '',
  showPagination: true,
  pageSize: 10,
  theme: 'light',
  height: 'auto',
  maxHeight: '500px',
  minHeight: '100px'
});

// 计算实际高度
const computedHeight = computed(() => {
  if (props.height !== 'auto') {
    return props.height;
  }
  // 自适应模式：根据当前页内容行数计算高度
  const lines = displayContent.value.split('\n').length;
  const lineHeight = 22; // 大约每行高度
  const padding = 24; // 上下 padding
  const calculatedHeight = lines * lineHeight + padding;
  
  // 限制在 min 和 max 之间
  const minH = parseInt(props.minHeight) || 100;
  const maxH = parseInt(props.maxHeight) || 500;
  const finalHeight = Math.max(minH, Math.min(calculatedHeight, maxH));
  
  return `${finalHeight}px`;
});

const emit = defineEmits<{
  pageChange: [page: number];
  pageSizeChange: [pageSize: number];
}>();

// 编辑器相关
const editorRoot = ref<HTMLElement | null>(null);
const editorView = ref<EditorView | null>(null);

// 分页相关
const currentPage = ref(1);
const pageSize = ref(props.pageSize);

// 解析SQL语句（按分号分割）
const sqlStatements = computed(() => {
  if (!props.sqlContent) return [];

  // 按分号分割SQL语句，过滤空语句
  const statements = props.sqlContent
    .split(';')
    .map(stmt => stmt.trim())
    .filter(stmt => stmt.length > 0);

  return statements;
});

// 总页数
const totalPages = computed(() => {
  if (!props.showPagination) return 1;
  return Math.ceil(sqlStatements.value.length / pageSize.value);
});

// 当前页显示的SQL语句
const currentPageStatements = computed(() => {
  if (!props.showPagination) {
    return sqlStatements.value;
  }

  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return sqlStatements.value.slice(start, end);
});

// 当前页显示的SQL内容
const displayContent = computed(() => {
  return currentPageStatements.value.join(';\n') + (currentPageStatements.value.length > 0 ? ';' : '');
});

// 获取主题扩展
function getThemeExtension(theme: string) {
  if (theme === 'dark') {
    return oneDark;
  }
  return [syntaxHighlighting(defaultHighlightStyle, { fallback: true })];
}

// 初始化编辑器
const initEditor = () => {
  if (editorView.value || !editorRoot.value) return;

  const state = EditorState.create({
    doc: displayContent.value,
    extensions: [
      lineNumbers(),
      foldGutter(),
      sql({ upperCaseKeywords: true }),
      getThemeExtension(props.theme),
      // 设置为只读
      EditorView.editable.of(false),
      EditorState.readOnly.of(true),
      // 自定义样式
      EditorView.theme({
        '&': {
          display: 'flex',
          flexDirection: 'column'
        },
        '.cm-content': {
          padding: '12px',
          flex: '1 1 auto',
          minHeight: '0' // 允许内容收缩
        },
        '.cm-focused': {
          outline: 'none'
        },
        '.cm-editor': {
          borderRadius: '4px',
          border: '1px solid var(--n-border-color)',
          minHeight: '100%', // 至少占满容器，内容多时可撑高，由外层容器滚动
          display: 'flex',
          flexDirection: 'column',
          backgroundColor: 'var(--n-color)',
          color: 'var(--n-text-color)'
        },
        '.cm-scroller': {
          fontFamily: 'v-mono, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
          fontSize: '14px',
          lineHeight: '1.6',
          overflowY: 'auto', // 内容超出时垂直滚动
          overflowX: 'auto', // 长行可横向滚动
          flex: '1 1 auto',
          minHeight: '0'
        },
        // 行号区域样式
        '.cm-gutters': {
          backgroundColor: 'var(--n-color)', // 跟随主题背景色
          color: 'var(--n-text-color-3)', // 弱文本颜色
          borderRight: '1px solid var(--n-border-color)' // 右侧边框
        },
        '.cm-activeLineGutter': {
          backgroundColor: 'transparent',
          color: 'var(--n-text-color-2)' // 激活行号颜色加深
        },
        // 自定义滚动条样式
        '.cm-scroller::-webkit-scrollbar': {
          width: '8px',
          height: '8px'
        },
        '.cm-scroller::-webkit-scrollbar-track': {
          background: 'var(--n-scrollbar-track-color, #f1f1f1)',
          borderRadius: '4px'
        },
        '.cm-scroller::-webkit-scrollbar-thumb': {
          background: 'var(--n-scrollbar-color, #c1c1c1)',
          borderRadius: '4px'
        },
        '.cm-scroller::-webkit-scrollbar-thumb:hover': {
          background: 'var(--n-scrollbar-color-hover, #a8a8a8)'
        }
      })
    ]
  });

  editorView.value = new EditorView({
    state,
    parent: editorRoot.value
  });
};

// 更新编辑器内容
function updateEditorContent() {
  const view = editorView.value;
  if (!view) return;

  const newContent = displayContent.value;
  const currentContent = view.state.doc.toString();

  if (currentContent !== newContent) {
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: newContent
      }
    });
  }
}

// 处理分页变化
function handlePageChange(page: number) {
  currentPage.value = page;
  emit('pageChange', page);
  nextTick(() => {
    updateEditorContent();
  });
}

// 处理页大小变化
function handlePageSizeChange(size: number) {
  pageSize.value = size;
  currentPage.value = 1; // 重置到第一页
  emit('pageSizeChange', size);
  nextTick(() => {
    updateEditorContent();
  });
}

// 监听内容变化
watch(
  () => displayContent.value,
  () => {
    updateEditorContent();
  }
);

// 监听主题变化
watch(
  () => props.theme,
  newTheme => {
    const view = editorView.value;
    if (view) {
      // 重新创建编辑器以应用新主题
      view.destroy();
      editorView.value = null;
      nextTick(() => {
        initEditor();
      });
    }
  }
);

onMounted(() => {
  nextTick(() => {
    initEditor();
  });
});

onUnmounted(() => {
  if (editorView.value) {
    editorView.value.destroy();
    editorView.value = null;
  }
});

// 暴露方法给父组件
defineExpose({
  /** 跳转到指定页 */
  goToPage: (page: number) => {
    if (page >= 1 && page <= totalPages.value) {
      handlePageChange(page);
    }
  },
  /** 获取当前页码 */
  getCurrentPage: () => currentPage.value,
  /** 获取总页数 */
  getTotalPages: () => totalPages.value,
  /** 获取当前页的SQL语句 */
  getCurrentStatements: () => currentPageStatements.value
});
</script>

<template>
  <div class="readonly-sql-editor">
    <!-- SQL编辑器容器 -->
    <div ref="editorRoot" class="editor-container" :style="{ height: computedHeight }" />

    <!-- 分页控制器 -->
    <div v-if="showPagination && totalPages > 1" class="pagination-container">
      <NPagination
        v-model:page="currentPage"
        :page-count="totalPages"
        :page-size="pageSize"
        :show-size-picker="true"
        :page-sizes="[10, 20, 50, 100]"
        show-quick-jumper
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </div>
  </div>
</template>

<style scoped>
.readonly-sql-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.editor-container {
  position: relative;
  width: 100%;
  max-width: 100%;
  max-height: 100%;
  overflow: auto; /* 内容超出时可滚动查看 */
  transition: height 0.2s ease;
}

.pagination-container {
  display: flex;
  justify-content: center;
  padding: 8px 0;
  flex-shrink: 0; /* 防止分页器被压缩 */
  border-top: 1px solid var(--n-border-color, #e0e0e6);
}
</style>
