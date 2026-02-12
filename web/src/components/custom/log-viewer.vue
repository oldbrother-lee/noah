<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { EditorState } from '@codemirror/state';
import { EditorView, lineNumbers } from '@codemirror/view';
import { sql } from '@codemirror/lang-sql';
import { defaultHighlightStyle, foldGutter, syntaxHighlighting } from '@codemirror/language';
import { oneDark } from '@codemirror/theme-one-dark';

interface Props {
  content?: string;
  theme?: 'light' | 'dark';
  height?: string;
}

const props = withDefaults(defineProps<Props>(), {
  content: '',
  theme: 'light',
  height: '400px'
});

const editorRoot = ref<HTMLElement | null>(null);
const editorView = ref<EditorView | null>(null);

const initEditor = () => {
  if (!editorRoot.value) return;

  const extensions = [
    lineNumbers(),
    foldGutter(),
    sql(), // Use SQL mode as in the original component
    syntaxHighlighting(defaultHighlightStyle),
    EditorView.editable.of(false),
    EditorView.lineWrapping,
    EditorView.theme({
      '&': { height: props.height },
      '.cm-scroller': {
        overflow: 'auto',
        fontFamily: 'v-mono, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
        lineHeight: '1.6',
        fontSize: '14px'
      },
      '.cm-editor': {
        backgroundColor: 'var(--n-color)',
        color: 'var(--n-text-color)'
      },
      '.cm-gutters': {
        backgroundColor: 'var(--n-color)',
        color: 'var(--n-text-color-3)',
        borderRight: '1px solid var(--n-border-color)'
      },
      '.cm-activeLineGutter': {
        backgroundColor: 'transparent',
        color: 'var(--n-text-color-2)'
      }
    })
  ];

  if (props.theme === 'dark') {
    extensions.push(oneDark);
  }

  const state = EditorState.create({
    doc: props.content,
    extensions
  });

  editorView.value = new EditorView({
    state,
    parent: editorRoot.value
  });
};

watch(
  () => props.content,
  newContent => {
    if (editorView.value) {
      const currentDoc = editorView.value.state.doc.toString();
      if (currentDoc !== newContent) {
        // Replace entire content for simplicity, or we could append if we managed state differently.
        // But since prop is the full content, we replace.
        // Optimization: if newContent starts with currentDoc, just append the difference.

        const transaction: any = {};

        if (newContent.startsWith(currentDoc)) {
          const appendText = newContent.slice(currentDoc.length);
          transaction.changes = { from: currentDoc.length, insert: appendText };
        } else {
          transaction.changes = { from: 0, to: currentDoc.length, insert: newContent };
        }

        editorView.value.dispatch(transaction);

        // Scroll to bottom
        setTimeout(() => {
          if (editorView.value) {
            editorView.value.dispatch({
              effects: EditorView.scrollIntoView(editorView.value.state.doc.length, { y: 'end' })
            });
          }
        }, 50);
      }
    }
  }
);

onMounted(() => {
  initEditor();
});

onUnmounted(() => {
  if (editorView.value) {
    editorView.value.destroy();
  }
});
</script>

<template>
  <div class="log-viewer">
    <div ref="editorRoot" class="editor-container"></div>
  </div>
</template>

<style scoped>
.log-viewer {
  width: 100%;
  border: 1px solid var(--n-border-color);
  border-radius: 4px;
  overflow: hidden;
  background-color: var(--n-color);
}
.editor-container {
  width: 100%;
}
</style>
