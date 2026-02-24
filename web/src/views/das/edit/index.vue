<script setup lang="ts">
import { ref, computed, onMounted, nextTick, onUnmounted, h, watch, markRaw, toRaw, getCurrentInstance } from 'vue';
import { useResizeObserver } from '@vueuse/core';
import { useI18n } from 'vue-i18n';
import { format } from 'sql-formatter';
import { fetchExecuteMySQLQuery, fetchExecuteClickHouseQuery, fetchSchemas, fetchTables, fetchUserGrants, fetchDBDict, fetchTableInfo } from '@/service/api/das';
import TableColumnSetting from '@/components/advanced/table-column-setting.vue';
import TableHeaderOperation from '@/components/advanced/table-header-operation.vue';
import History from '../history/index.vue';
import Favorite from '../favorite/index.vue';
import SvgIcon from '@/components/custom/svg-icon.vue';
import { useAppStore } from '@/store/modules/app';
// CodeMirror 6 imports
import { EditorState, Compartment, RangeSet, RangeSetBuilder } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, gutter, GutterMarker, ViewPlugin, ViewUpdate } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap, indentWithTab, insertNewlineAndIndent } from '@codemirror/commands';
import { sql } from '@codemirror/lang-sql';
import { foldGutter, foldKeymap, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language';
import { oneDark } from '@codemirror/theme-one-dark';
import { autocompletion, completionKeymap } from '@codemirror/autocomplete';
import { useRouter } from 'vue-router';
// 静态导入 vxe-table，随 DAS 编辑页 chunk 一起加载，避免分包后动态 import 未完成导致表格不渲染
// 使用 named 导入表格与列组件，在模板中按 PascalCase 使用，避免全局注册时机导致 resolve 失败
import VxeTableLib, { VxeTable, VxeColumn } from 'vxe-table';
import VxePcUI from 'vxe-pc-ui';
import 'vxe-table/lib/style.css';
import 'vxe-pc-ui/lib/style.css';
// 使用普通textarea

const { t } = useI18n();
const appStore = useAppStore();

// 响应式数据
const showLeftPanel = ref(!appStore.isMobile); // 移动端默认隐藏左侧面板
const selectedSchema = ref<any>({});
const activeKey = ref(localStorage.getItem('dms-active-key') || '1');
const newTabIndex = ref(Number(localStorage.getItem('dms-new-tab-index')) || 2);
const tabCompletion = ref<any>({});
// 改用网格布局管理左右区域占比，移动端自适应
const rightSpan = computed(() => {
  if (appStore.isMobile) return 24; // 移动端全宽
  return showLeftPanel.value ? 17 : 24;
});
const leftSpan = computed(() => {
  if (appStore.isMobile) return 0; // 移动端不显示左侧面板
  return showLeftPanel.value ? 7 : 0;
});

// 左侧数据
const schemas = ref<any[]>([]);
const bindTitle = ref('');
const showSearch = ref(false);
const treeLoading = ref(false);
const treeData = ref<any[]>([]);
const searchTreeData = ref<any[]>([]);
const refreshLoading = ref(false);
const tableInfoVisible = ref(false);
const selectedKeys = ref<any>({});
const leftTableSearch = ref('');
// 新增：追踪树展开的键集合 & 列分组展开集合
const expandedKeys = ref<any[]>([]);
const columnsGroupExpanded = ref<Set<string | number>>(new Set());
/** 表格就绪：VXE 在 onMounted 同步注册后为 true */
const vxeReady = ref(false);
/** VXE 注册失败时改用 NDataTable 兜底 */
const useFallbackTable = ref(false);

let vxeInstalled = false;

// 路由跳转：收藏SQL、历史查询
const router = useRouter();
const gotoFavorite = () => router.push({ name: 'das_favorite' });
const gotoHistory = () => router.push({ name: 'das_history' });

// 过滤后的树数据
const filteredTreeData = computed(() => {
  const kw = leftTableSearch.value.trim().toLowerCase();
  if (!kw || searchTreeData.value.length === 0) return treeData.value;
  return treeData.value.filter((node: any) => {
    const title = (node.label || node.title || '').toLowerCase();
    return title.includes(kw);
  });
});

// 自定义左侧 NTree 节点渲染：左侧字段名，右侧类型/列数
const renderTreeLabel = ({ option }: { option: any }) => {
  const label = option.label || '';

  // 叶子节点：列
  if (option.isLeaf) {
    const name = option.label;
    const type = option.colType || '';

    return h(
      'span',
      { class: 'das-tree-item' },
      [
        h('span', { class: 'das-tree-item-left' }, [
          h(SvgIcon, { icon: 'carbon:list', class: 'mr-6px text-14px text-info' }),
          h('span', { class: 'das-tree-item-name' }, name)
        ]),
        h('span', { class: 'das-tree-item-type' }, type)
      ]
    );
  }

  // 表节点：仅在展开时显示第二行“列(数量)”并可开关
  const count = Array.isArray(option.children) ? option.children.length : 0;
  const isNodeExpanded = expandedKeys.value?.includes?.(option.key);
  const groupOpened = columnsGroupExpanded.value.has(option.key);
  return h(
    'span',
    { class: 'das-tree-item das-tree-item-table' },
    [
      // 第一行：图标 + 表名
      h('span', { class: 'das-tree-item-left' }, [
        h(SvgIcon, { icon: 'mdi:table', class: 'text-14px text-info' }),
        h('span', { class: 'das-tree-item-name' }, label)
      ]),
      // 第二行：仅在展开时显示“列(数量)”且可独立展开/收起列清单
      isNodeExpanded && count ? h('span', { class: 'das-tree-item-meta-row' }, [
        h(
          'span',
          {
            class: 'das-tree-item-count das-tree-item-count-toggle',
            onClick: (e: MouseEvent) => { e.stopPropagation(); toggleColumnsGroup(option.key); }
          },
          [
            h(SvgIcon, { icon: groupOpened ? 'carbon:chevron-down' : 'carbon:chevron-right', class: 'mr-2px text-14px' }),
            `列(${count})`
          ]
        )
      ]) : null
    ]
  );
};

// 自定义展开/折叠图标为更常见的箭头（右/下），并禁用默认旋转
const renderSwitcherIcon = ({ expanded }: { expanded: boolean }) => {
  return h(
    SvgIcon,
    { icon: expanded ? 'carbon:chevron-down' : 'carbon:chevron-right', class: 'das-tree-switcher-icon', style: 'transform: none' }
  );
};

// 新增：切换表节点下“列”分组展开/收起
function toggleColumnsGroup(key: string | number) {
  if (columnsGroupExpanded.value.has(key)) {
    columnsGroupExpanded.value.delete(key);
  } else {
    columnsGroupExpanded.value.add(key);
  }
}

// 新增：按开关过滤子节点返回（仅当分组展开时返回列）
function getNodeChildren(option: any) {
  if (!option) return [];
  // 叶子节点原样返回（通常无 children）
  if (option.isLeaf) return option.children || [];
  // 表节点：只有当该表的“列”分组被打开时才返回列清单
  const key = option.key;
  if (columnsGroupExpanded.value.has(key)) {
    return option.children || [];
  }
  return [];
}

// 标签页数据
interface EditorPane {
  title: string;
  key: string;
  closable: boolean;
  sql: string;
  sessionVars: string;
  characterSet: string;
  theme: string;
  result?: any;
  loading?: boolean;
  responseMsg?: string;
  pagination?: {
    currentPage: number;
    pageSize: number;
    total: number;
  };
  bottomActiveTab?: string;
  editorHeight?: number;
  isEditing?: boolean;
  editingTitle?: string;
}

const panes = ref<EditorPane[]>([]);

// 监听 panes 变化并持久化标签页列表
watch(
  () => panes.value.map(p => p.key),
  (keys) => {
    localStorage.setItem('dms-panes-keys', JSON.stringify(keys));
  },
  { deep: true }
);

const defaultPageSize = 20;
const pageSizes = [10, 20, 50, 100];

const tabIndex = computed(() => {
  return panes.value.findIndex((v) => v.key === activeKey.value);
});

const currentPane = computed<EditorPane>(() => panes.value[tabIndex.value] || panes.value[0]);
const currentSchemaLabel = computed(() => {
  if (!selectedSchema.value?.schema) return '未选择库';
  const type = selectedSchema.value?.db_type || 'mysql';
  return `${selectedSchema.value.schema} · ${type}`;
});

const schemaError = ref<string | null>(null);

// OS Detection for Tooltip
const isMac = navigator.userAgent.includes('Mac');
const executeTooltip = computed(() => isMac ? '执行 Cmd+Enter' : '执行 Control+Enter');

// CodeMirror: per-pane editor实例及可重配置主题
const editorViews = ref<Record<string, EditorView | null>>({});
const themeCompartments = ref<Record<string, Compartment>>({});
const languageCompartments = ref<Record<string, Compartment>>({});

// 检查是否是黑暗模式
function isDarkMode(): boolean {
  return document.documentElement.classList.contains('dark');
}

// 黑暗模式光标主题 - 使用 baseTheme 确保优先级
const darkCursorTheme = EditorView.theme({
  '.cm-cursor': {
    borderLeft: '1.2px solid #ffffff !important'
  },
  '.cm-cursor-primary': {
    borderLeft: '1.2px solid #ffffff !important'
  },
  '&.cm-focused > .cm-scroller > .cm-cursorLayer .cm-cursor': {
    borderLeft: '1.2px solid #ffffff !important'
  },
  '.cm-content': {
    caretColor: '#ffffff !important'
  }
}, { dark: true });

// 浅色模式光标主题
const lightCursorTheme = EditorView.theme({
  '.cm-cursor': {
    borderLeft: '1.2px solid #000000'
  },
  '.cm-cursor-primary': {
    borderLeft: '1.2px solid #000000'
  },
  '.cm-content': {
    caretColor: '#000000'
  }
});

function getThemeExtension(theme: string) {
  const isDark = isDarkMode();
  // 仅在暗色主题时启用 oneDark，其余使用默认浅色
  if (theme === 'vs-dark' || theme === 'hc-black' || isDark) {
    // 将光标主题放在最后，确保覆盖 oneDark 的设置
    return [oneDark, darkCursorTheme];
  }
  // 浅色主题下启用默认的语法高亮样式
  return [syntaxHighlighting(defaultHighlightStyle, { fallback: true }), lightCursorTheme];
}

// 自定义自动补全主题
const customAutocompleteTheme = EditorView.theme({
  '.cm-tooltip.cm-tooltip-autocomplete': {
    border: 'none',
    borderRadius: '4px',
    boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
    backgroundColor: 'rgb(var(--container-bg-color))', // 使用容器背景色变量
    color: 'rgb(var(--base-text-color))', // 使用基础文本颜色变量
    minWidth: '250px'
  },
  '.cm-tooltip-autocomplete > ul > li': {
    display: 'flex',
    alignItems: 'center',
    padding: '4px 8px',
    lineHeight: '1.5',
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    fontSize: '13px',
    cursor: 'pointer'
  },
  '.cm-tooltip-autocomplete > ul > li[aria-selected]': {
    backgroundColor: 'rgb(var(--primary-color))', // 使用主题色变量
    color: '#fff'
  },
  '.cm-tooltip-autocomplete > ul > li[aria-selected] .cm-completionDetail': {
    color: 'rgba(255, 255, 255, 0.85)'
  },
  '.cm-completionLabel': {
    fontWeight: '500',
    marginRight: '8px'
  },
  '.cm-completionDetail': {
    marginLeft: 'auto',
    color: 'rgb(var(--base-text-color))', // 使用文本颜色
    opacity: '0.6',
    fontSize: '12px',
    fontStyle: 'normal'
  },
  '.cm-completionMatchedText': {
    textDecoration: 'none',
    fontWeight: 'bold',
    color: 'rgb(var(--primary-color))' // 使用主题色变量
  },
  '.cm-tooltip-autocomplete > ul > li[aria-selected] .cm-completionMatchedText': {
    color: '#fff'
  },
  // 图标通用样式
  '.cm-completionIcon': {
    display: 'inline-block',
    width: '16px',
    height: '16px',
    marginRight: '8px',
    backgroundRepeat: 'no-repeat',
    backgroundSize: 'contain',
    backgroundPosition: 'center',
    verticalAlign: 'middle',
    opacity: '0.8'
  },
  // 表格图标 (绿色网格)
  '.cm-completionIcon-table': {
    backgroundImage: `url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjMTBiOTgxIiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PHJlY3QgeD0iMyIgeT0iMyIgd2lkdGg9IjE4IiBoZWlnaHQ9IjE4IiByeD0iMiIgcnk9IjIiPjwvcmVjdD48bGluZSB4MT0iMyIgeTE9IjkiIHgyPSIyMSIgeTI9IjkiPjwvbGluZT48bGluZSB4MT0iMyIgeTE9IjE1IiB4Mj0iMjEiIHkyPSIxNSI+PC9saW5lPjxsaW5lIHgxPSI5IiB5MT0iMyIgeDI9IjkiIHkyPSIyMSI+PC9saW5lPjwvc3ZnPg==')`
  },
  // 字段图标 (蓝色属性)
  '.cm-completionIcon-column': {
    backgroundImage: `url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjM2I4MmY2IiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PHBhdGggZD0iTTIwLjI0IDEyLjI0YTYgNiAwIDAgMC04LjQ5LTguNDlMNSAxMC41VjE5aDguNXoiPjwvcGF0aD48bGluZSB4MT0iMTYiIHkxPSI4IiB4Mj0iMiIgeTI9IjIyIj48L2xpbmU+PGxpbmUgeDE9IjE3LjUiIHkxPSIxNSIgeDI9IjkiIHkyPSIxNSI+PC9saW5lPjwvc3ZnPg==')`
  },
  // 关键字图标 (橙色 Key)
  '.cm-completionIcon-keyword': {
    // 使用 Lucide Key 图标
    backgroundImage: `url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZjU5ZTBiIiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PGNpcmNsZSBjeD0iNy41IiBjeT0iMTUuNSIgcj0iNS41Ii8+PHBhdGggZD0ibTIxIDItOSA5Ii8+PHBhdGggZD0ibTIxIDJ2NmgtNnYiLz48L3N2Zz4=')`
  },
  // 函数图标 (紫色 Function)
  '.cm-completionIcon-function': {
    // 使用 Lucide Function Square (f in box)
    backgroundImage: `url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjYTg1NWY3IiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PHJlY3Qgd2lkdGg9IjE4IiBoZWlnaHQ9IjE4IiB4PSIzIiB5PSIzIiByeD0iMiIgcnk9IjIiLz48cGF0aCBkPSJNOSAxN2MyIDAgMi0yIDItMiIvPjxwYXRoIGQ9Ik05IDEyaDZtLTEgMHYtMy41YzAtMSAuNS0xLjUgMS41LTEuNSIvPjwvc3ZnPg==')`
  },
  // 常量/值图标 (青色)
  '.cm-completionIcon-constant': {
    backgroundImage: `url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjMDZMQjU2IiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PGNpcmNsZSBjeD0iMTIiIGN5PSIxMiIgcj0iMTAiLz48cGF0aCBkPSJtOSAxMiAyIDIgNC00Ii8+PC9zdmc+')` // Check Circle
  },
  // 隐藏默认的内容（如果有冲突）
  '.cm-completionIcon::after': { display: 'none' }
});

function schemaForCompletion(): Record<string, string[]> {
  const tablesMap = (tabCompletion.value?.tables || {}) as Record<string, string[]>;
  return tablesMap;
}

// 自定义SQL补全逻辑：支持表名（带注释）和字段名（带类型）
function customSQLCompletion(context: any) {
  const metadata = tabCompletion.value?.metadata || { tables: {} };
  
  // 1. 尝试匹配 "Table.Column" 模式 (检测点号)
  // 匹配：前面是单词，紧接一个点，然后是可选的当前输入
  const dotMatch = context.matchBefore(/(\w+)\.(\w*)$/);
  
  if (dotMatch) {
    const tableName = dotMatch.text.split('.')[0];
    const tableInfo = metadata.tables[tableName];
    
    if (tableInfo && tableInfo.columns) {
      const options = tableInfo.columns.map((col: any) => ({
        label: col.name,
        type: 'column',
        detail: col.type, // 显示字段类型
        boost: 10 // 提高优先级
      }));
      
      return {
        from: dotMatch.from + tableName.length + 1, //补全起点在点号之后
        options,
        validFor: /^\w*$/
      };
    }
  }

  // 2. 默认模式：提示表名 + 上下文相关字段
  // 将 \w+ 改为 \w* 以支持光标在空格后立即触发提示（此时匹配为空字符串）
  const wordMatch = context.matchBefore(/\w*$/);
  
  if (wordMatch) {
    // 只有当明确处于单词输入中或为空匹配（如空格后）时才继续
    // 如果需要更严格的空格触发，可以检查 context.explicit 或上一个字符是否为空格
    
    const tableNames = Object.keys(metadata.tables);
    if (tableNames.length === 0) return null;
    
    const options: any[] = [];
    
    // 扫描当前文档中出现过的表名（上下文感知）
    const docText = context.state.doc.toString();
    
    // 解析当前 SQL 语句中实际使用的表名（FROM 和 JOIN 后面的表）
    // 匹配 FROM tablename 或 JOIN tablename 格式
    const usedTablesInSQL = new Set<string>();
    const fromJoinPattern = /\b(?:FROM|JOIN)\s+[`"']?(\w+)[`"']?/gi;
    let tableMatch;
    while ((tableMatch = fromJoinPattern.exec(docText)) !== null) {
      const tblName = tableMatch[1];
      // 只有当表名在 metadata 中存在时才添加
      if (metadata.tables[tblName]) {
        usedTablesInSQL.add(tblName);
      }
    }
    // 也匹配 UPDATE tablename 格式
    const updatePattern = /\bUPDATE\s+[`"']?(\w+)[`"']?/gi;
    while ((tableMatch = updatePattern.exec(docText)) !== null) {
      const tblName = tableMatch[1];
      if (metadata.tables[tblName]) {
        usedTablesInSQL.add(tblName);
      }
    }
    
    // 检测当前上下文是否处于 FROM 或 JOIN 之后
    // 使用关键字检测代替简单的正则，以避免 "FROM table WHERE" 被误判
    const textBefore = context.state.sliceDoc(0, context.pos);
    const lastKeywordMatch = textBefore.match(/\b(SELECT|FROM|JOIN|WHERE|GROUP\s+BY|ORDER\s+BY|LIMIT|SET|UPDATE|DELETE|INSERT|HAVING|ON|AND|OR)\b/gi);
    const lastKeyword = lastKeywordMatch ? lastKeywordMatch[lastKeywordMatch.length - 1].toUpperCase().replace(/\s+/g, ' ') : '';
    
    const isTableContext = ['FROM', 'JOIN', 'UPDATE', 'INTO'].includes(lastKeyword);
    const isConditionContext = ['WHERE', 'HAVING', 'ON', 'AND', 'OR'].includes(lastKeyword);
    const isSelectContext = lastKeyword === 'SELECT';

    // 获取当前正在输入的单词之前的上一个有效 token (用于判断是否刚输入完字段)
    const textBeforeCurrentWord = context.state.sliceDoc(0, wordMatch.from);

    // 用户反馈：= 号后面不应该弹出提示（通常是在输入值）
    // 仅在自动触发时屏蔽，如果用户按快捷键强制触发则允许
    if (!context.explicit && /[=<>!]+\s*$/.test(textBeforeCurrentWord)) {
      return null;
    }

    const prevTokenMatch = textBeforeCurrentWord.match(/([`"']?[\w.]+\b[`"']?)\s*$/);
    const prevToken = prevTokenMatch ? prevTokenMatch[1].replace(/[`"']/g, '') : '';
    // 处理 Table.Field 格式，只取 Field 部分
    const prevTokenField = prevToken.includes('.') ? prevToken.split('.').pop() : prevToken;

    // 收集所有已知的字段名，用于判断 prevToken 是否为字段
    const allColumnNames = new Set<string>();
    Object.values(metadata.tables).forEach((t: any) => {
      if (t.columns) {
        t.columns.forEach((c: any) => allColumnNames.add(c.name));
      }
    });

    // H. 针对 "o" 开头的特殊高优先级提示（定义提前，以便后续逻辑使用）
    const isStartWithO = wordMatch.text.toLowerCase().startsWith('o');

    // C. 判断上一个词是否是字段名（用于后续调整优先级）
    const isPrevTokenColumn = allColumnNames.has(prevTokenField);
    
    // 如果处于条件上下文且上一个词是字段名，提示运算符（最高优先级）
    if (isConditionContext && isPrevTokenColumn) {
       const operators = [
         { label: '=', type: 'keyword', detail: 'Operator', boost: 100 }, // 最高优先级
         { label: '!=', type: 'keyword', detail: 'Operator', boost: 95 },
         { label: '<>', type: 'keyword', detail: 'Operator', boost: 95 },
         { label: '>', type: 'keyword', detail: 'Operator', boost: 90 },
         { label: '>=', type: 'keyword', detail: 'Operator', boost: 90 },
         { label: '<', type: 'keyword', detail: 'Operator', boost: 90 },
         { label: '<=', type: 'keyword', detail: 'Operator', boost: 90 },
         { label: 'IN', type: 'keyword', detail: 'Operator', boost: 85 },
         { label: 'LIKE', type: 'keyword', detail: 'Operator', boost: 85 },
         { label: 'NOT IN', type: 'keyword', detail: 'Operator', boost: 85 },
         { label: 'IS NULL', type: 'keyword', detail: 'Operator', boost: 85 },
         { label: 'IS NOT NULL', type: 'keyword', detail: 'Operator', boost: 85 },
         { label: 'BETWEEN', type: 'keyword', detail: 'Operator', boost: 80 },
         { label: 'NOT LIKE', type: 'keyword', detail: 'Operator', boost: 80 },
         { label: 'REGEXP', type: 'keyword', detail: 'Operator', boost: 75 }
       ];
       options.push(...operators);
    }

    // D. 如果处于 SELECT 上下文，提示聚合函数和常用常量
    if (isSelectContext) {
       // 检测是否刚刚输入了 *
       const isAfterStar = /\*\s*$/.test(textBeforeCurrentWord);
       
       if (isAfterStar) {
          // 如果在 * 后面，优先提示 FROM
          const afterStarOptions = [
            { label: 'FROM', type: 'keyword', detail: 'Keyword', boost: 60 },
            { label: 'FALSE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'TRUE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'UNKNOWN', type: 'constant', detail: 'Value', boost: 40 },
            { label: 'NULL', type: 'constant', detail: 'Value', boost: 40 },
            { label: 'ALTER', type: 'keyword', detail: 'Keyword', boost: 30 },
            { label: 'AND', type: 'keyword', detail: 'Keyword', boost: 30 },
            { label: 'BY', type: 'keyword', detail: 'Keyword', boost: 30 },
            { label: 'CREATE', type: 'keyword', detail: 'Keyword', boost: 30 },
            { label: 'COLUMN', type: 'keyword', detail: 'Keyword', boost: 30 },
            { label: 'COUNT()', type: 'function', detail: 'Function', boost: 30 },
            { label: 'BETWEEN', type: 'keyword', detail: 'Keyword', boost: 30 }
          ];
          options.push(...afterStarOptions);
       } else {
          // 正常的 SELECT 上下文提示
          const selectOptions = [
            { label: '*', type: 'keyword', detail: 'All Columns', boost: 50 },
            { label: 'DISTINCT()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'COUNT()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'MAX()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'MIN()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'SUM()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'FALSE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'TRUE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'UNKNOWN', type: 'constant', detail: 'Value', boost: 40 },
            { label: 'NULL', type: 'constant', detail: 'Value', boost: 40 }
          ];
          options.push(...selectOptions);
       }
    }

    // E. 如果上一个词是表名，提示连接查询关键字和 WHERE
    // 检查 metadata.tables 中是否存在该表名
    if (metadata.tables[prevToken]) {
       const afterTableOptions = [
        { label: 'WHERE', type: 'keyword', detail: 'Keyword', boost: 60 },
        // 如果已通过特殊逻辑添加了 oKeywords (isStartWithO)，这里就不要重复添加
        ...(isStartWithO ? [] : [
          { label: 'ORDER BY', type: 'keyword', detail: 'Keyword', boost: 58 },
        ]),
        { label: 'GROUP BY', type: 'keyword', detail: 'Keyword', boost: 57 },
         { label: 'LIMIT', type: 'keyword', detail: 'Keyword', boost: 56 },
         { label: ',', type: 'keyword', detail: 'Separator', boost: 55 },
         { label: 'INNER', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'OUTER', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'LEFT JOIN', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'RIGHT JOIN', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'CROSS', type: 'keyword', detail: 'Keyword', boost: 45 },
         { label: 'JOIN', type: 'keyword', detail: 'Keyword', boost: 45 },
         { label: 'STRAIGHT_JOIN', type: 'keyword', detail: 'Keyword', boost: 45 },
         { label: 'HAVING', type: 'keyword', detail: 'Keyword', boost: 40 }
       ];
       options.push(...afterTableOptions);
    }

    // F. 检测 ORDER BY / GROUP BY 上下文，提示字段和排序方向
    const isOrderByContext = lastKeyword === 'ORDER BY';
    const isGroupByContext = lastKeyword === 'GROUP BY';
    
    if (isOrderByContext || isGroupByContext) {
      // 提示排序方向（仅 ORDER BY）
      if (isOrderByContext) {
        options.push(
          { label: 'ASC', type: 'keyword', detail: 'Ascending', boost: 50 },
          { label: 'DESC', type: 'keyword', detail: 'Descending', boost: 50 }
        );
      }
      // 提示 LIMIT（ORDER BY 后常用）
      options.push(
        { label: 'LIMIT', type: 'keyword', detail: 'Keyword', boost: 45 }
      );
    }

    if (isStartWithO) {
      const oKeywords = [
        { label: 'ORDER BY', type: 'keyword', detail: 'Keyword', boost: 2000 },
        { label: 'OUTER JOIN', type: 'keyword', detail: 'Keyword', boost: 1900 },
        { label: 'ON', type: 'keyword', detail: 'Keyword', boost: 1800 },
        { label: 'OR', type: 'keyword', detail: 'Keyword', boost: 1700 },
        { label: 'ORDER', type: 'keyword', detail: 'Keyword', boost: 1600 },
        { label: 'OF', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'OLD', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'ONLY', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'OPEN', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'OPTION', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'ORDINALITY', type: 'keyword', detail: 'Keyword', boost: 1500 },
        { label: 'OUT', type: 'keyword', detail: 'Keyword', boost: 1500 }
      ];
      options.push(...oKeywords);
    }

    // G. 通用 SQL 关键字提示（始终可用，但优先级较低）
    const commonKeywords = [
      { label: 'SELECT', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'FROM', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'WHERE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'AND', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'OR', type: 'keyword', detail: 'Keyword', boost: -10 },
      // 如果已通过特殊逻辑添加了 oKeywords，这里就不要重复添加 ORDER BY 等
      ...(isStartWithO ? [] : [
        { label: 'ORDER BY', type: 'keyword', detail: 'Keyword', boost: 5 },
        { label: 'OR', type: 'keyword', detail: 'Keyword', boost: -10 },
      ]),
      { label: 'GROUP BY', type: 'keyword', detail: 'Keyword', boost: 5 },
      { label: 'HAVING', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LIMIT', type: 'keyword', detail: 'Keyword', boost: 5 },
      { label: 'OFFSET', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LEFT JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'RIGHT JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'INNER JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'OUTER JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'ON', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'AS', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'DISTINCT', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'UNION', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'UNION ALL', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'INSERT INTO', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'UPDATE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'DELETE FROM', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'SET', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'VALUES', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'ASC', type: 'keyword', detail: 'Ascending', boost: -10 },
      { label: 'DESC', type: 'keyword', detail: 'Descending', boost: -10 },
      { label: 'BETWEEN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'NOT', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LIKE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IS NULL', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IS NOT NULL', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'EXISTS', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'CASE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'WHEN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'THEN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'ELSE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'END', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'NULL', type: 'constant', detail: 'Value', boost: -10 },
      { label: 'TRUE', type: 'constant', detail: 'Boolean', boost: -10 },
      { label: 'FALSE', type: 'constant', detail: 'Boolean', boost: -10 }
    ];
    options.push(...commonKeywords);

    // 首先添加当前 SQL 中实际使用的表的字段
    usedTablesInSQL.forEach((tableName) => {
      const info = metadata.tables[tableName];
      if (info && info.columns) {
        info.columns.forEach((col: any) => {
          // 根据上下文调整字段优先级：
          // - 如果上一个词是字段名，说明需要运算符，字段优先级降低
          // - 如果在 FROM/JOIN 后，字段优先级降低
          // - 否则字段优先级较高
          let colBoost = 50;
          if (isPrevTokenColumn && isConditionContext) {
            colBoost = 20; // 上一个词是字段名，降低优先级让运算符优先
          } else if (isTableContext) {
            colBoost = -5; // FROM/JOIN 后应该是表名
          }
          
          options.push({
            label: col.name,
            type: 'column',
            detail: `${col.type} · ${tableName}`, // 优化显示格式：类型 · 表名
            boost: colBoost 
          });
        });
      }
    });

    tableNames.forEach((name) => {
      const info = metadata.tables[name];
      
      // A. 添加所有表名
      // 如果处于 FROM/JOIN 后，大幅提升表名的优先级，使其排在最前
      const tableBoost = isTableContext ? 20 : -1;
      
      options.push({
        label: name,
        type: 'table', // 图标通常为类/表
        detail: info?.comment || '表', // 优先显示注释，否则显示中文'表'
        boost: tableBoost 
      });

      // B. 只有当表不在当前 SQL 使用的表中时，才以较低优先级添加其字段
      // 避免重复添加已经在上面添加过的当前表字段
      if (!usedTablesInSQL.has(name) && info.columns) {
        // 不在当前 SQL 中使用的表的字段，优先级很低
        info.columns.forEach((col: any) => {
          const colBoost = isTableContext ? -10 : -5;
          
          options.push({
            label: col.name,
            type: 'column',
            detail: `${col.type} · ${name}`, // 优化显示格式：类型 · 表名
            boost: colBoost 
          });
        });
      }
    });
    
    return {
      from: wordMatch.from,
      options,
      validFor: /^\w*$/
    };
  }

  return null;
}

// Custom gutter markers
class LineNumberMarker extends GutterMarker {
  constructor(public number: number) { super(); }
  eq(other: GutterMarker) {
    return other instanceof LineNumberMarker && this.number === other.number;
  }
  toDOM() {
    return document.createTextNode(this.number.toString());
  }
}

class ExecuteMarker extends GutterMarker {
  constructor(private execute: () => void) { super(); }
  eq(other: GutterMarker) {
    return other instanceof ExecuteMarker;
  }
  toDOM() {
    const div = document.createElement('div');
    div.style.cursor = 'pointer';
    div.style.color = '#18a058'; // Use green color similar to Naive UI primary
    div.style.display = 'flex';
    div.style.justifyContent = 'center';
    div.style.alignItems = 'center';
    div.style.width = '100%';
    div.style.height = '100%';
    div.style.pointerEvents = 'auto';
    div.title = executeTooltip.value; // Add tooltip
    
    // Use a filled SVG icon for better visibility
    div.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>`;
    
    // Add hover effect via JS since we are in a component
    div.onmouseenter = () => { div.style.color = '#36ad6a'; div.style.transform = 'scale(1.1)'; div.style.transition = 'all 0.2s'; };
    div.onmouseleave = () => { div.style.color = '#18a058'; div.style.transform = 'scale(1)'; };

    // Use mousedown to prevent editor selection interference and ensure event capture
    div.onmousedown = (e) => {
      e.preventDefault();
      e.stopPropagation();
      this.execute();
    };
    return div;
  }
}

const createExecuteGutter = (execute: () => void) => {
  const executeLinePlugin = ViewPlugin.fromClass(class {
    markers: RangeSet<GutterMarker>;
    constructor(view: EditorView) {
      this.markers = this.buildMarkers(view);
    }
    update(update: ViewUpdate) {
      if (update.docChanged || update.selectionSet || update.viewportChanged) {
        this.markers = this.buildMarkers(update.view);
      }
    }
    buildMarkers(view: EditorView) {
      const builder = new RangeSetBuilder<GutterMarker>();
      const lines = view.state.doc.lines;
      
      const executeButtonLines = new Set<number>();
      for (const range of view.state.selection.ranges) {
        if (range.empty) continue;
        const startLine = view.state.doc.lineAt(range.from).number;
        executeButtonLines.add(startLine);
      }

      for (let i = 1; i <= lines; i++) {
        const line = view.state.doc.line(i);
        if (executeButtonLines.has(i)) {
          builder.add(line.from, line.from, new ExecuteMarker(execute));
        } else {
          builder.add(line.from, line.from, new LineNumberMarker(i));
        }
      }
      return builder.finish();
    }
  });

  return [
    executeLinePlugin,
    gutter({
      class: "cm-lineNumbers",
      markers: v => v.plugin(executeLinePlugin)?.markers || RangeSet.empty,
      initialSpacer: (view) => new LineNumberMarker(view.state.doc.lines)
    })
  ];
};

function createEditor(pane: EditorPane, el: HTMLElement) {
  // 若该 pane 已创建过编辑器，则避免重复创建以防渲染循环
  if (editorViews.value[pane.key]) return;
  const state = EditorState.create({
    doc: pane.sql || '',
    extensions: [
      createExecuteGutter(() => executeSQL(pane)),
      foldGutter(),
      // 语言通过 Compartment 动态配置（便于后续更新 schema）
      (languageCompartments.value[pane.key] = new Compartment()).of(sql({ schema: schemaForCompletion(), upperCaseKeywords: true })),
      // 主题通过 Compartment 动态配置
      (themeCompartments.value[pane.key] = new Compartment()).of(getThemeExtension(pane.theme)),
      // 注入自定义补全主题样式
      customAutocompleteTheme,
      history(),
      keymap.of([
        {
          key: 'Mod-Enter',
          run: () => {
            executeSQL(pane);
            return true;
          }
        },
        {
          key: 'Ctrl-Enter',
          run: () => {
            executeSQL(pane);
            return true;
          }
        },
        {
          key: 'Cmd-Enter',
          run: () => {
            executeSQL(pane);
            return true;
          }
        },
        {
          key: 'Enter',
          run: insertNewlineAndIndent
        },
        ...defaultKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...completionKeymap,
        indentWithTab
      ]),
      // 使用默认内置补全来源，并启用输入时触发
      autocompletion({ activateOnTyping: true, maxRenderedOptions: 50 }),
      // 以语言数据形式注入额外的自定义补全源
      EditorState.languageData.of(() => [{ autocomplete: customSQLCompletion }]),
      EditorView.updateListener.of((v) => {
        if (v.docChanged) {
          const text = v.state.doc.toString();
          pane.sql = text;
          saveCodeToCache(pane);
        }
      })
    ]
  });
  const view = new EditorView({ state, parent: el });
  editorViews.value[pane.key] = markRaw(view);
}

const onResizeStart = (event: MouseEvent, pane: EditorPane) => {
  const startY = event.clientY;
  const startHeight = pane.editorHeight || 300;
  
  const onMouseMove = (e: MouseEvent) => {
    e.preventDefault();
    const deltaY = e.clientY - startY;
    const newHeight = Math.max(100, startHeight + deltaY);
    pane.editorHeight = newHeight;
  };
  
  const onMouseUp = () => {
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
    document.body.style.cursor = '';
  };
  
  document.body.style.cursor = 'row-resize';
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
};

const showDictModal = ref(false);
const dictLoading = ref(false);
const dictHtmlContent = ref('');

function setEditorRef(pane: EditorPane, el: HTMLElement | null) {
  if (!el) return;
  const view = editorViews.value[pane.key];
  // 若已存在视图但被卸载，需要重新挂载到当前容器
  if (view) {
    if (view.dom.parentElement !== el) {
      el.innerHTML = '';
      el.appendChild(view.dom);
    }
    return;
  }
  createEditor(pane, el);
}

// keep editor in sync when sql changes externally (e.g., dblclick fill)
watch(
  () => panes.value.map(p => ({ key: p.key, sql: p.sql })),
  (list) => {
    list.forEach(({ key, sql }) => {
      const view = editorViews.value[key];
      if (view) {
        const cur = view.state.doc.toString();
        if (cur !== (sql || '')) {
          view.dispatch({
            changes: { from: 0, to: view.state.doc.length, insert: sql || '' }
          });
        }
      }
    });
  }
);

// 当主题改变时动态重新配置主题扩展
watch(
  () => panes.value.map(p => ({ key: p.key, theme: p.theme })),
  (list) => {
    list.forEach(({ key, theme }) => {
      const view = editorViews.value[key];
      const compartment = themeCompartments.value[key];
      if (view && compartment) {
        view.dispatch({ effects: compartment.reconfigure(getThemeExtension(theme)) });
      }
    });
  }
);

// 当表/列信息更新时，动态刷新语言扩展中的 schema，使内置补全即时生效
watch(
  () => tabCompletion.value,
  () => {
    Object.entries(editorViews.value).forEach(([key, view]) => {
      if (!view) return;
      const langComp = languageCompartments.value[key];
      if (!langComp) return;
      view.dispatch({ effects: langComp.reconfigure(sql({ schema: schemaForCompletion(), upperCaseKeywords: true })) });
    });
  },
  { deep: true }
);

// 监听dark模式切换，更新所有编辑器的主题
const darkModeObserver = new MutationObserver(() => {
  Object.entries(editorViews.value).forEach(([key, view]) => {
    if (!view) return;
    const compartment = themeCompartments.value[key];
    const pane = panes.value.find(p => p.key === key);
    if (compartment && pane) {
      view.dispatch({ effects: compartment.reconfigure(getThemeExtension(pane.theme)) });
    }
  });
});

// 编辑器实例引用
const editorRefs = ref<Record<string, any>>({});

// 左侧方法
const refreshSchemas = async () => {
  refreshLoading.value = true;
  try {
    await getSchemas();
    window.$message?.info('库列表刷新成功，请展开下拉列表查看');
  } finally {
    refreshLoading.value = false;
  }
};

const onSearch = (value: string) => {
  if (!value) {
    treeData.value = searchTreeData.value;
    return;
  }
  const searchResult = treeData.value.filter((item: any) => {
    const title = item.title || item.label || '';
    return title.indexOf(value) > -1;
  });
  treeData.value = searchResult;
};

const getGrants = async (params: any) => {
  try {
    const { data } = await fetchUserGrants(params);
    return data;
  } catch (error) {
    // 获取权限失败，返回空权限列表
    return { tables: [] };
  }
};

const getSchemas = async () => {
  try {
    const { data } = await fetchSchemas();
    schemas.value = data || [];
  } catch (error) {
    window.$message?.error('加载库列表失败');
  }
};

const getTables = async (value: string) => {
  searchTreeData.value = [];
  showSearch.value = true;
  treeLoading.value = true;
  leftTableSearch.value = '';
  schemaError.value = null;
  
  const vals = value.split('#');
  selectedSchema.value = {
    instance_id: vals[0],
    schema: vals[1],
    db_type: vals[2]
  };
  
  const params = {
    instance_id: vals[0],
    schema: vals[1]
  };
  
  try {
    const { data, error } = await fetchTables(params);
    if (error) {
      const msg = (error as any).response?.data?.message || (error as any).message || '加载失败';
      schemaError.value = msg;
      return;
    }
    if (data) {
      const grants = await getGrants(selectedSchema.value);
      renderTree(grants, data);
    }
  } catch (error: any) {
    const msg = error?.message || '加载失败';
    schemaError.value = msg;
    window.$message?.error(msg);
  } finally {
    treeLoading.value = false;
  }
};

// 刷新当前库的表列表（不改变已选择的库）
const refreshTables = async () => {
  if (!selectedSchema.value?.instance_id || !selectedSchema.value?.schema) {
    window.$message?.warning('请先选择左侧的库');
    return;
  }
  treeLoading.value = true;
  schemaError.value = null;
  try {
    const { data, error } = await fetchTables({
      instance_id: selectedSchema.value.instance_id,
      schema: selectedSchema.value.schema
    });
    if (error) {
      const msg = (error as any).response?.data?.message || (error as any).message || '刷新失败';
      schemaError.value = msg;
      return;
    }
    if (data) {
      const grants = await getGrants(selectedSchema.value);
      renderTree(grants, data);
      window.$message?.success('表列表已刷新');
    }
  } catch (error: any) {
    const msg = error?.message || '刷新失败';
    schemaError.value = msg;
    window.$message?.error(msg);
  } finally {
    treeLoading.value = false;
  }
};

const checkTableRule = (grants: any, table: string) => {
  if (grants?.tables?.length === 1 && grants.tables === '*') {
    return true;
  }
  if (!grants?.tables || !Array.isArray(grants.tables)) {
    return true;
  }
  
  let hasAllow = false;
  if (grants.tables[0]?.['rule'] === 'allow') {
    hasAllow = true;
  }
  
  if (hasAllow) {
    for (const v of grants.tables) {
      if (v['rule'] === 'allow' && v['table'] === table) {
        return true;
      }
    }
    return false;
  } else {
    for (const v of grants.tables) {
      if (v['rule'] === 'deny' && v['table'] === table) {
        return false;
      }
    }
    return true;
  }
};

const renderTree = (grants: any, data: any[]) => {
  const tmpTreeData: any[] = [];
  const tmpTabCompletion: any = { tables: {} };
  
  data.forEach((row: any) => {
    const tmpColumnsData: any[] = [];
    const columnsCompletion: string[] = [];
    
    // 解析列信息
    const columnsStr = row.columns || '';
    const columns = columnsStr.split('@@');
    
    columns.forEach((v: string) => {
      if (!v) return;
      const parts = v.split('$$');
      const colName = parts[0];
      const colType = parts.length > 1 ? parts[1] : '';

      tmpColumnsData.push({
        title: colName,
        label: colName,
        colType,
        key: `${row['table_schema']}#${row['table_name']}#${colName}`,
        isLeaf: true
      });
      columnsCompletion.push(colName);
    });
    
    // 检查表权限
    const rule = checkTableRule(grants, row.table_name) ? 'allow' : 'deny';
    const remark = row.table_comment ? ` (${row.table_comment})` : '';
    
    tmpTreeData.push({
      title: `${row.table_name}${remark}`,
      label: `${row.table_name}${remark}`,
      key: `${row['table_schema']}#${row['table_name']}`,
      rule,
      children: tmpColumnsData
    });
    
    tmpTabCompletion['tables'][row['table_name']] = columnsCompletion;
    
    // 存储额外的元数据供自定义补全使用
    if (!tmpTabCompletion['metadata']) tmpTabCompletion['metadata'] = { tables: {}, columns: {} };
    tmpTabCompletion['metadata'].tables[row['table_name']] = {
      comment: row.table_comment || '',
      columns: tmpColumnsData.map(c => ({ name: c.label, type: c.colType }))
    };
  });
  
  treeData.value = tmpTreeData;
  searchTreeData.value = [...tmpTreeData];
  tabCompletion.value = tmpTabCompletion;
  // 默认将每个表的“列”分组设为展开
  columnsGroupExpanded.value = new Set(tmpTreeData.map((n: any) => n.key));
};

// 右侧编辑器方法
const foldLeft = () => {
  showLeftPanel.value = !showLeftPanel.value;
};

const onEdit = (targetKey: string, action: 'add' | 'remove') => {
  if (action === 'add') {
    add();
  } else {
    remove(targetKey);
  }
};

const add = () => {
  const activeKeyValue = newTabIndex.value++;
  localStorage.setItem('dms-new-tab-index', newTabIndex.value.toString());
  const newPane: EditorPane = {
    title: `SQLConsole ${activeKeyValue}`,
    key: activeKeyValue.toString(),
    closable: true,
    sql: '',
    sessionVars: '',
    characterSet: 'utf8',
    theme: 'default',
    result: null,
    loading: false,
    responseMsg: '',
    pagination: { currentPage: 1, pageSize: defaultPageSize, total: 0 },
    bottomActiveTab: 'result',
    editorHeight: 300,
    isEditing: false,
    editingTitle: ''
  };
  
  // 从缓存加载状态
  loadPaneFromCache(newPane);
  
  panes.value.push(newPane);
  activeKey.value = activeKeyValue.toString();
};

const remove = (targetKey: string) => {
  // 销毁对应的编辑器实例，避免残留引用导致重新挂载失败
  const view = editorViews.value[targetKey];
  if (view) {
    view.destroy();
    delete editorViews.value[targetKey];
    delete themeCompartments.value[targetKey];
    delete languageCompartments.value[targetKey];
  }

  let activeKeyValue = activeKey.value;
  let lastIndex = -1;
  
  panes.value.forEach((pane, i) => {
    if (pane.key === targetKey) {
      lastIndex = i - 1;
    }
  });
  
  const newPanes = panes.value.filter((pane) => pane.key !== targetKey);
  
  if (newPanes.length && activeKeyValue === targetKey) {
    if (lastIndex >= 0) {
      activeKeyValue = newPanes[lastIndex].key;
    } else {
      activeKeyValue = newPanes[0].key;
    }
  }
  
  panes.value = newPanes;
  activeKey.value = activeKeyValue;
  localStorage.setItem('dms-active-key', activeKeyValue);
};

const changeTab = () => {
  localStorage.setItem('dms-active-key', activeKey.value);
};

// SQL执行
const parseSessionVars = (pane: EditorPane) => {
  const sessionVars: any = {};
  if (pane.sessionVars && pane.sessionVars.length > 0) {
    pane.sessionVars.split(';').forEach((v: string) => {
      const sessionVar = v.split('=');
      if (sessionVar.length === 2) {
        sessionVars[sessionVar[0].trim()] = sessionVar[1].trim();
      }
    });
  }
  return sessionVars;
};

const executeMySQLQuery = async (pane: EditorPane, data: any) => {
  const characterSet = {
    character_set_client: pane.characterSet,
    character_set_connection: pane.characterSet,
    character_set_results: pane.characterSet
  };
  
  data['params'] = { ...characterSet, ...parseSessionVars(pane) };
  
  const resMsgs: string[] = [];
  pane.loading = true;
  
  try {
    const response = await fetchExecuteMySQLQuery(data);

    if (response.error) {
      throw response.error;
    }

    const respData = response.data as any;
    
    resMsgs.push('结果: 执行成功');
    resMsgs.push(`耗时: ${respData?.duration || '-'}`);
    resMsgs.push(`SQL: ${respData?.sqltext || data.sqltext}`);
    resMsgs.push(`请求ID: ${(response as any).request_id || '-'}`);
    
    if (respData) {
      // 如果返回的data中有data字段（嵌套结构），提取出来
      if (respData.data && Array.isArray(respData.data)) {
        pane.result = {
          columns: respData.columns || [],
          rows: respData.data,
          data: respData.data,
          duration: respData.duration,
          sqltext: respData.sqltext,
          affected_rows: respData.affected_rows,
          affectedRows: respData.affectedRows
        };
      } else {
        pane.result = respData;
      }
    }
    
    pane.responseMsg = resMsgs.join('<br>');
    
    // 清除缓存，因为数据已更新
    tableDataCache.delete(pane);
    tableColumnsCache.delete(pane);
    
    // 更新表格列设置
    updateTableColumnChecks(pane);
    initPagination(pane);
    
    // Switch to result tab
    pane.bottomActiveTab = 'result';
    
    window.$message?.success('执行成功');
  } catch (error: any) {
    resMsgs.push('结果: 执行失败');
    const backendMsg = error?.response?.data?.message || error?.response?.data?.msg || error?.message || '未知错误';
    resMsgs.push(`错误: ${backendMsg}`);
    const requestId = error?.response?.headers?.['x-request-id'] || error?.response?.data?.request_id;
    if (requestId) {
       resMsgs.push(`请求ID: ${requestId}`);
    }
    pane.responseMsg = resMsgs.join('<br>');
    pane.result = null;
    
    // 清除缓存
    tableDataCache.delete(pane);
    tableColumnsCache.delete(pane);
    
    if (error?.message?.includes('sessionid')) {
      window.$message?.error('执行失败，认证过期，请刷新页面后重新执行');
    } else {
      window.$message?.error('执行失败');
    }
  } finally {
    pane.loading = false;
  }
};

const executeClickHouseQuery = async (pane: EditorPane, data: any) => {
  data['params'] = { ...parseSessionVars(pane) };
  
  const resMsgs: string[] = [];
  pane.loading = true;
  
  try {
    const response = await fetchExecuteClickHouseQuery(data);

    if (response.error) {
      throw response.error;
    }

    const respData = response.data as any;
    
    resMsgs.push('结果: 执行成功');
    resMsgs.push(`耗时: ${respData?.duration || '-'}`);
    resMsgs.push(`SQL: ${respData?.sqltext || data.sqltext}`);
    resMsgs.push(`请求ID: ${(response as any).request_id || '-'}`);
    
    if (respData) {
      // 如果返回的data中有data字段（嵌套结构），提取出来
      if (respData.data && Array.isArray(respData.data)) {
        pane.result = {
          columns: respData.columns || [],
          rows: respData.data,
          data: respData.data,
          duration: respData.duration,
          sqltext: respData.sqltext,
          affected_rows: respData.affected_rows,
          affectedRows: respData.affectedRows
        };
      } else {
        pane.result = respData;
      }
    }
    
    pane.responseMsg = resMsgs.join('<br>');
    
    // 清除缓存，因为数据已更新
    tableDataCache.delete(pane);
    tableColumnsCache.delete(pane);
    
    // 更新表格列设置
    updateTableColumnChecks(pane);
    initPagination(pane);
    
    // Switch to result tab
    pane.bottomActiveTab = 'result';
    
    window.$message?.success('执行成功');
  } catch (error: any) {
    resMsgs.push('结果: 执行失败');
    const backendMsg = error?.response?.data?.message || error?.response?.data?.msg || error?.message || '未知错误';
    resMsgs.push(`错误: ${backendMsg}`);
    const requestId = error?.response?.headers?.['x-request-id'] || error?.response?.data?.request_id;
    if (requestId) {
       resMsgs.push(`请求ID: ${requestId}`);
    }
    pane.responseMsg = resMsgs.join('<br>');
    pane.result = null;
    
    // 清除缓存
    tableDataCache.delete(pane);
    tableColumnsCache.delete(pane);
    
    if (error?.message?.includes('sessionid')) {
      window.$message?.error('执行失败，认证过期，请刷新页面后重新执行');
    } else {
      window.$message?.error('执行失败');
    }
  } finally {
    pane.loading = false;
  }
};

// 获取选中的 SQL，如果没有选区则返回全文
function getSqlToExecute(p: EditorPane): string {
  const view = toRaw(editorViews.value[p.key]);
  if (view) {
    const ranges = view.state.selection.ranges;
    for (const r of ranges) {
      if (!r.empty) {
        const sel = view.state.sliceDoc(r.from, r.to).trim();
        if (sel.length > 0) return sel;
      }
    }
  }
  return (p.sql || '').trim();
}

const executeSQL = (pane?: EditorPane) => {
  const p = pane || currentPane.value;
  
  // Ensure editor is focused so selection remains visible/active
  const view = toRaw(editorViews.value[p.key]);
  view?.focus();

  saveCodeToCache(p);

  if (Object.keys(selectedSchema.value).length === 0) {
    window.$message?.warning('请先选择左侧的库');
    return;
  }

  const sqltext = getSqlToExecute(p);
  if (!sqltext || sqltext.length === 0) {
    window.$message?.warning('请输入或选择要执行的SQL');
    return;
  }

  const data = {
    ...selectedSchema.value,
    sqltext
  };

  const dbType = selectedSchema.value['db_type']?.toLowerCase() || 'mysql';
  if (dbType === 'tidb' || dbType === 'mysql') {
    executeMySQLQuery(p, data);
  } else if (dbType === 'clickhouse') {
    executeClickHouseQuery(p, data);
  }
};

const handleReuseSQL = (pane: EditorPane, sql: string) => {
  if (!sql) return;
  const view = toRaw(editorViews.value[pane.key]);
  if (view) {
    const current = view.state.doc.toString();
    const separator = current ? '\n\n' : '';
    const insertPos = current.length;
    const insertText = `${separator}${sql}`;

    const tr = view.state.update({
      changes: { from: insertPos, insert: insertText },
      selection: { anchor: insertPos + separator.length, head: insertPos + insertText.length },
      scrollIntoView: true
    });
    view.dispatch(tr);

    // 立即执行选中的 SQL
    executeSQL(pane);
  } else {
    pane.sql = pane.sql ? `${pane.sql}\n\n${sql}` : sql;
  }
};

const formatSQL = (pane?: EditorPane, mode: 'format' | 'minify' = 'format') => {
  const p = pane || currentPane.value;
  const view = toRaw(editorViews.value[p.key]);

  // Helper functions
  const doMinify = (text: string) => text.replace(/\s+/g, ' ').trim();
  const doFormat = (text: string) => format(text, { language: 'mysql', keywordCase: 'upper' });
  const processText = (text: string) => mode === 'minify' ? doMinify(text) : doFormat(text);
  const successMsg = mode === 'minify' ? '压缩成功' : '格式化成功';
  const failMsg = mode === 'minify' ? '压缩失败' : '格式化失败，请检查SQL语法';

  if (!view) {
    try {
      p.sql = processText(p.sql);
      window.$message?.success(successMsg);
      saveCodeToCache(p);
    } catch (error) {
      window.$message?.warning(failMsg);
    }
    return;
  }

  const ranges = view.state.selection.ranges;
  const hasSelection = ranges.some((r) => !r.empty);

  try {
    if (hasSelection) {
      const changes = ranges
        .filter((r) => !r.empty)
        .map((r) => {
          const text = view.state.sliceDoc(r.from, r.to);
          return {
            from: r.from,
            to: r.to,
            insert: processText(text)
          };
        });

      view.dispatch({ changes });
      window.$message?.success(mode === 'minify' ? '已压缩选中内容' : '已格式化选中内容');
    } else {
      const text = view.state.doc.toString();
      const processed = processText(text);

      view.dispatch({
        changes: { from: 0, to: text.length, insert: processed }
      });
      window.$message?.success(successMsg);
    }
    
    // 保存到缓存（等待 updateListener 更新 p.sql 后）
    nextTick(() => saveCodeToCache(p));
  } catch (error) {
    window.$message?.warning(failMsg);
  }
};

const loadDBDictData = async () => {
  if (Object.keys(selectedSchema.value).length === 0) {
    window.$message?.warning('请先选择左侧的库');
    return;
  }
  
  // 验证参数，避免传递 undefined
  if (!selectedSchema.value.instance_id || !selectedSchema.value.schema) {
    window.$message?.warning('实例ID和库名不能为空');
    return;
  }
  
  dictLoading.value = true;
  try {
    const { data } = await fetchDBDict({
      instance_id: selectedSchema.value.instance_id,
      schema: selectedSchema.value.schema
    });
    
    if (!data) {
       window.$message?.warning('未获取到数据字典数据');
       return;
    }

    // 生成HTML
    const html = generateDBDictHtml(data, selectedSchema.value.schema);
    
    dictHtmlContent.value = html;
    showDictModal.value = true;
  } catch (error: any) {
    window.$message?.error(error?.message || '加载数据字典失败');
  } finally {
    dictLoading.value = false;
  }
};

const generateDBDictHtml = (data: any, schemaName: string) => {
  let tables: any[] = [];

  if (Array.isArray(data)) {
    tables = data.map((row: any) => {
      const columns: any[] = [];
      
      // 解析列信息：用 @@ 分隔每列，用 $$ 分隔列名、类型、注释
      if (row.columns) {
        const columnStrs = row.columns.split('@@').filter((s: string) => s.trim());
        columnStrs.forEach((colStr: string) => {
          const parts = colStr.split('$$');
          if (parts.length >= 2) {
            columns.push({
              columnName: parts[0] || '',
              dataType: parts[1] || '',
              columnComment: parts[2] || '',
              isNullable: true,
              columnDefault: '',
              characterSet: '',
              collation: '',
              isPrimaryKey: false
            });
          }
        });
      }
      
      // 判断主键：列名包含 'id' 且是第一个列，可能是主键
      if (columns.length > 0 && columns[0].columnName.toLowerCase().includes('id')) {
        columns[0].isPrimaryKey = true;
      }

      return {
        tableName: row.table_name || row.TABLE_NAME || '',
        tableComment: '',
        createTime: '',
        columns,
        indexes: []
      };
    });
  } else if (data && data.tables) {
    tables = data.tables;
  }
  
  let html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>数据字典 - ${schemaName}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; padding: 20px; color: #333; background-color: #f4f6f8; }
    .container { max-width: 1400px; margin: 0 auto; background: #fff; padding: 30px; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.05); }
    h1 { text-align: center; margin-bottom: 30px; color: #2c3e50; border-bottom: 2px solid #eaeaea; padding-bottom: 15px; }
    
    .toc { background: #f8f9fa; padding: 20px; border-radius: 6px; margin-bottom: 40px; border: 1px solid #e9ecef; }
    .toc h2 { margin-top: 0; font-size: 18px; border-bottom: 1px solid #ddd; padding-bottom: 10px; margin-bottom: 15px; color: #495057; }
    .toc ul { list-style: none; padding: 0; display: flex; flex-wrap: wrap; gap: 10px; }
    .toc li { margin: 0; }
    .toc a { display: block; padding: 6px 12px; background: #fff; border: 1px solid #ddd; border-radius: 4px; text-decoration: none; color: #007bff; font-size: 14px; transition: all 0.2s; }
    .toc a:hover { background: #e9ecef; border-color: #adb5bd; color: #0056b3; transform: translateY(-1px); box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
    
    .table-section { margin-bottom: 40px; border: 1px solid #e9ecef; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.03); overflow: hidden; background: #fff; }
    .table-header { background-color: #f8f9fa; padding: 15px 20px; border-bottom: 1px solid #e9ecef; display: flex; align-items: center; justify-content: space-between; }
    .table-title { display: flex; align-items: baseline; flex-wrap: wrap; gap: 10px; }
    .table-title h3 { margin: 0; margin-right: 15px; color: #2c3e50; font-size: 18px; }
    .table-title .meta { color: #6c757d; font-size: 14px; margin-right: 15px; }
    .back-to-top { font-size: 12px; color: #007bff; text-decoration: none; }
    
    .content { padding: 20px; }
    h4 { margin-top: 0; margin-bottom: 15px; color: #495057; border-left: 4px solid #007bff; padding-left: 10px; font-size: 16px; }
    
    table { width: 100%; border-collapse: collapse; margin-bottom: 25px; font-size: 14px; }
    th, td { border: 1px solid #dee2e6; padding: 10px 12px; text-align: left; }
    th { background-color: #f1f3f5; font-weight: 600; color: #495057; white-space: nowrap; }
    tr:nth-child(even) { background-color: #f8f9fa; }
    tr:hover { background-color: #f1f3f5; }
    
    .primary-key { color: #d63384; font-weight: bold; background: #fff0f6; padding: 2px 6px; border-radius: 3px; font-size: 12px; border: 1px solid #ffadd2; white-space: nowrap; }
    .nullable-no { color: #dc3545; font-weight: bold; }
    .nullable-yes { color: #28a745; }
    .data-type { color: #0d6efd; font-family: Consolas, Monaco, 'Andale Mono', monospace; }
    .meta-info { font-size: 12px; color: #868e96; }
  </style>
</head>
<body>
  <div class="container">
    <h1 id="top">数据字典: ${schemaName}</h1>
    
    <div class="toc">
      <h2>表目录 (${tables.length})</h2>
      <ul>
        ${tables.map(t => `<li><a href="javascript:void(0)" onclick="document.getElementById('${t.tableName}').scrollIntoView({behavior: 'smooth'}); return false;">${t.tableName}</a></li>`).join('')}
      </ul>
    </div>

    ${tables.map(table => `
    <div id="${table.tableName}" class="table-section">
      <div class="table-header">
        <div class="table-title">
          <h3>${table.tableName}</h3>
          <span class="meta">注释: ${table.tableComment || '暂无'}</span>
          ${table.createTime ? `<span class="meta">创建时间: ${table.createTime}</span>` : ''}
        </div>
        <a href="javascript:void(0)" onclick="document.getElementById('top').scrollIntoView({behavior: 'smooth'}); return false;" class="back-to-top">↑ 返回顶部</a>
      </div>
      
      <div class="content">
        <h4>列信息</h4>
        <table>
          <thead>
            <tr>
              <th>列名</th>
              <th>类型</th>
              <th>注释</th>
              <th>主键</th>
            </tr>
          </thead>
          <tbody>
            ${table.columns.map((col: any) => `
            <tr>
              <td style="font-weight: 500">${col.columnName || ''}</td>
              <td class="data-type">${col.dataType || ''}</td>
              <td>${col.columnComment || ''}</td>
              <td style="text-align: center;">${col.isPrimaryKey ? '<span class="primary-key">PK</span>' : ''}</td>
            </tr>
            `).join('')}
          </tbody>
        </table>

        ${table.indexes && table.indexes.length > 0 ? `
        <h4>索引信息</h4>
        <table>
          <thead>
            <tr>
              <th style="width: 30%">索引名</th>
              <th style="width: 50%">包含列</th>
              <th style="width: 20%">唯一</th>
            </tr>
          </thead>
          <tbody>
            ${table.indexes.map((idx: any) => `
            <tr>
              <td>${idx.indexName}</td>
              <td>${idx.columnNames.join(', ')}</td>
              <td style="text-align: center;">${idx.isUnique ? '是' : '否'}</td>
            </tr>
            `).join('')}
          </tbody>
        </table>
        ` : '<p class="meta-info">暂无索引信息</p>'}
      </div>
    </div>
    `).join('')}
  </div>
</body>
</html>`;
  
  return html;
};



// 缓存管理
const saveCodeToCache = (pane: EditorPane) => {
  localStorage.setItem(`dms-codemirror-${pane.key}`, pane.sql);
  localStorage.setItem(`dms-character-${pane.key}`, pane.characterSet);
  localStorage.setItem(`dms-theme-${pane.key}`, pane.theme);
  localStorage.setItem(`dms-sessionvars-${pane.key}`, pane.sessionVars);
  localStorage.setItem(`dms-title-${pane.key}`, pane.title);
};

const loadCodeFromCache = (key: string): string => {
  return localStorage.getItem(key) || '';
};

const loadPaneFromCache = (pane: EditorPane) => {
  pane.sql = loadCodeFromCache(`dms-codemirror-${pane.key}`);
  pane.characterSet = localStorage.getItem(`dms-character-${pane.key}`) || 'utf8';
  pane.theme = localStorage.getItem(`dms-theme-${pane.key}`) || 'default';
  pane.sessionVars = localStorage.getItem(`dms-sessionvars-${pane.key}`) || '';
  pane.title = localStorage.getItem(`dms-title-${pane.key}`) || pane.title;
};

// 标签重命名逻辑
const handleRenameTab = (pane: EditorPane) => {
  pane.isEditing = true;
  pane.editingTitle = pane.title;
};

const saveTabName = (pane: EditorPane) => {
  if (!pane.editingTitle || !pane.editingTitle.trim()) {
    window.$message?.warning('标签名称不能为空');
    return;
  }
  
  // 校验：支持中文、英文、数字和下划线组合，长度限制 1-64
  const nameRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]{1,64}$/;
  if (!nameRegex.test(pane.editingTitle)) {
    window.$message?.warning('标签名称仅支持中文、英文、数字和下划线，且长度不超过64个字符');
    return;
  }

  pane.title = pane.editingTitle.trim();
  pane.isEditing = false;
  saveCodeToCache(pane);
};

const cancelRenameTab = (pane: EditorPane) => {
  pane.isEditing = false;
  pane.editingTitle = pane.title;
};

// 表格数据 - 使用项目中的高级表格组件（已优化，使用缓存版本）

// 缓存转换后的数据，避免重复转换
const tableDataCache = new WeakMap<EditorPane, any[]>();
// 缓存列配置，避免重复计算
const tableColumnsCache = new WeakMap<EditorPane, any[]>();

// 获取表格列配置（使用缓存优化）
const getTableColumns = (pane: EditorPane) => {
  // 检查缓存
  if (tableColumnsCache.has(pane)) {
    return tableColumnsCache.get(pane)!;
  }

  const result = pane.result;
  if (!result) {
    tableColumnsCache.set(pane, []);
    return [];
  }
  
  const columns = result.columns || [];
  if (!columns.length) {
    tableColumnsCache.set(pane, []);
    return [];
  }
  
  const tableColumns = columns.map((name: string) => ({ 
    title: name, 
    key: name, 
    ellipsis: { tooltip: true },
    minWidth: 120
  }));
  
  // 缓存结果
  tableColumnsCache.set(pane, tableColumns);
  return tableColumns;
};

const getTableData = (pane: EditorPane) => {
  // 检查缓存
  if (tableDataCache.has(pane)) {
    return tableDataCache.get(pane)!;
  }

  const result = pane.result;
  if (!result) {
    tableDataCache.set(pane, []);
    return [];
  }
  
  const cols = result.columns || [];
  const rows = result.rows || result.data || [];
  
  if (!cols.length || !rows.length) {
    tableDataCache.set(pane, []);
    return [];
  }
  
  let data: any[] = [];
  
  // 如果rows中的元素已经是对象（键值对），直接返回
  if (rows.length > 0 && typeof rows[0] === 'object' && !Array.isArray(rows[0])) {
    data = rows;
  } else if (Array.isArray(rows) && rows.length > 0 && Array.isArray(rows[0])) {
    // 如果rows是二维数组，转换为对象数组
    // 使用 Object.create(null) 创建更快的对象
    data = rows.map((row: any[]) => {
      const obj: Record<string, any> = Object.create(null);
      const colCount = Math.min(cols.length, row.length);
      for (let idx = 0; idx < colCount; idx++) {
        obj[cols[idx]] = row[idx];
      }
      return obj;
    });
  }
  
  // 缓存结果
  tableDataCache.set(pane, data);
  return data;
};

// 初始化/更新分页统计
const initPagination = (pane: EditorPane) => {
  const total = getTableData(pane).length;
  if (!pane.pagination) {
    pane.pagination = { currentPage: 1, pageSize: defaultPageSize, total };
  } else {
    pane.pagination.total = total;
    const maxPage = Math.max(1, Math.ceil(total / (pane.pagination.pageSize || defaultPageSize)));
    if (pane.pagination.currentPage > maxPage) {
      pane.pagination.currentPage = maxPage;
    }
  }
};

// 按分页切片后的数据（优化：使用缓存避免重复计算）
const getPagedTableData = (pane: EditorPane) => {
  const full = getTableData(pane);
  if (!full.length) return [];
  
  const currentPage = pane.pagination?.currentPage ?? 1;
  const pageSize = pane.pagination?.pageSize ?? defaultPageSize;
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  
  // 如果数据量很大，使用更高效的方式
  if (full.length > 10000) {
    // 对于大数据量，只返回当前页的数据
    return full.slice(start, end);
  }
  
  return full.slice(start, end);
};

// 分页变更事件（优化：避免重复计算总数）
const onPageChange = (pane: EditorPane, currentPage: number, pageSize: number) => {
  const total = getTableData(pane).length;
  if (!pane.pagination) {
    pane.pagination = { currentPage, pageSize, total };
  } else {
    pane.pagination.currentPage = currentPage;
    pane.pagination.pageSize = pageSize;
    pane.pagination.total = total;
  }
};

// 表格列设置
const tableColumnChecks = ref<NaiveUI.TableColumnCheck[]>([]);
// 派生勾选映射，便于快速判断列是否显示
const tableColumnCheckMap = computed<Record<string, boolean>>(() => {
  const map: Record<string, boolean> = {};
  tableColumnChecks.value.forEach((c) => {
    map[c.key] = !!c.checked;
  });   
  return map;
});

// 根据勾选状态返回可见列
const getVisibleColumns = (pane: EditorPane) => {
  const columns = getTableColumns(pane);
  return columns.filter((c: any) => tableColumnCheckMap.value[c.key] !== false);
};

// 更新表格列设置（保留已有勾选状态）
const updateTableColumnChecks = (pane: EditorPane) => {
  const columns = getTableColumns(pane);
  const oldMap: Record<string, boolean> = {};
  tableColumnChecks.value.forEach((c) => {
    oldMap[c.key] = !!c.checked;
  });
  tableColumnChecks.value = columns.map((col: any) => ({
    key: col.key as string,
    title: col.title as string,
    checked: Object.prototype.hasOwnProperty.call(oldMap, col.key) ? oldMap[col.key] : true
  }));
};

// 编辑器事件处理
const onEditorChange = (pane: EditorPane, value: string) => {
  pane.sql = value;
  saveCodeToCache(pane);
};

// 已移除拖拽分隔逻辑，使用组件自身布局属性实现

// 树节点点击
const handleNodeClick = (keys: string[], e: any) => {
  // 节点点击处理逻辑
};

// 树节点双击填充SQL（追加模式）
const handleNodeDblClick = (key: string) => {
  if (!key) return;
  const parts = key.split('#');
  if (parts.length !== 2) return; // 只对表节点生效
  
  const [schema, table] = parts;
  const p = currentPane.value;
  const dbType = selectedSchema.value['db_type']?.toLowerCase() || 'mysql';
  
  const query = dbType === 'clickhouse'
    ? `SELECT * FROM "${schema}"."${table}" LIMIT 100;`
    : `SELECT * FROM \`${schema}\`.\`${table}\` LIMIT 100;`;
  
  // 追加到现有内容；若为空则直接赋值
  p.sql = p.sql && p.sql.length > 0 ? `${p.sql}\n\n${query}` : query;
};

// 左右高度同步逻辑
const rightContainerRef = ref<HTMLElement | null>(null);
const leftContainerStyle = ref({ height: 'auto', overflowY: 'auto' });

useResizeObserver(rightContainerRef, (entries) => {
  const entry = entries[0];
  const { height } = entry.contentRect;
  if (height > 0) {
    leftContainerStyle.value = {
      height: `${height}px`,
      overflowY: 'auto'
    };
  }
});

// Context Menu Logic
const showContextMenu = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuPane = ref<EditorPane | null>(null);
const hasSelection = ref(false);

const contextMenuOptions = computed(() => [
  {
    label: '执行选中 SQL',
    key: 'execute',
    icon: () => h(SvgIcon, { icon: 'carbon:flash' }),
    disabled: !hasSelection.value
  },
  {
    type: 'divider',
    key: 'd1'
  },
  {
    label: '复制',
    key: 'copy',
    icon: () => h(SvgIcon, { icon: 'carbon:copy' }),
    disabled: !hasSelection.value
  },
  {
    label: '剪切',
    key: 'cut',
    icon: () => h(SvgIcon, { icon: 'carbon:cut' }),
    disabled: !hasSelection.value
  },
  {
    type: 'divider',
    key: 'd2'
  },
  {
    label: '格式化 SQL',
    key: 'format',
    icon: () => h(SvgIcon, { icon: 'carbon:code' })
  }
]);

const handleContextMenu = (e: MouseEvent, pane: EditorPane) => {
  e.preventDefault();
  
  // Check if there is a selection
  const view = toRaw(editorViews.value[pane.key]);
  if (!view) return;
  
  const ranges = view.state.selection.ranges;
  hasSelection.value = ranges.some(r => !r.empty);
  
  showContextMenu.value = false;
  nextTick(() => {
    showContextMenu.value = true;
    contextMenuX.value = e.clientX;
    contextMenuY.value = e.clientY;
    contextMenuPane.value = pane;
  });
};

const handleContextSelect = async (key: string) => {
  showContextMenu.value = false;
  const pane = contextMenuPane.value;
  if (!pane) return;

  if (key === 'execute') {
    executeSQL(pane);
  } else if (key === 'format') {
    formatSQL(pane, 'format');
  } else if (key === 'copy' || key === 'cut') {
    const view = toRaw(editorViews.value[pane.key]);
    if (!view) return;
    
    const ranges = view.state.selection.ranges;
    const text = ranges
      .filter(r => !r.empty)
      .map(r => view.state.sliceDoc(r.from, r.to))
      .join('\n');
      
    if (!text) return;

    try {
      await navigator.clipboard.writeText(text);
      if (key === 'cut') {
        view.dispatch(view.state.replaceSelection(''));
        window.$message?.success('已剪切');
      } else {
        window.$message?.success('已复制');
      }
    } catch (err) {
      window.$message?.error('操作失败，请检查浏览器权限');
    }
  }
};

const onClickOutside = () => {
  showContextMenu.value = false;
};

// 监听标签页属性变化实现自动保存
watch(
  () => panes.value.map(p => ({ 
    key: p.key, 
    characterSet: p.characterSet, 
    theme: p.theme, 
    sessionVars: p.sessionVars 
  })),
  (list, oldList) => {
    list.forEach((item, index) => {
      const oldItem = oldList ? oldList[index] : null;
      if (oldItem && (
        item.characterSet !== oldItem.characterSet || 
        item.theme !== oldItem.theme || 
        item.sessionVars !== oldItem.sessionVars
      )) {
        const pane = panes.value.find(p => p.key === item.key);
        if (pane) {
          saveCodeToCache(pane);
        }
      }
    });
  },
  { deep: true }
);

onMounted(async () => {
  const instance = getCurrentInstance();
  const app = instance?.appContext?.app;
  if (app && !vxeInstalled) {
    try {
      app.use(VxePcUI);
      app.use(VxeTableLib);
      vxeInstalled = true;
    } catch (e) {
      console.error('VXE 表格注册失败，使用备用表格', e);
      useFallbackTable.value = true;
    }
  }
  vxeReady.value = true;

  await getSchemas();
  
  // 恢复标签页列表
  const savedKeys = localStorage.getItem('dms-panes-keys');
  if (savedKeys) {
    try {
      const keys = JSON.parse(savedKeys);
      if (Array.isArray(keys) && keys.length > 0) {
        keys.forEach(key => {
          const newPane: EditorPane = {
            title: `SQLConsole ${key}`, // 临时标题，loadPaneFromCache 会恢复真实标题
            key: key.toString(),
            closable: key !== '1',
            sql: '',
            sessionVars: '',
            characterSet: 'utf8',
            theme: 'default',
            result: null,
            loading: false,
            responseMsg: '',
            pagination: { currentPage: 1, pageSize: defaultPageSize, total: 0 },
            bottomActiveTab: 'result',
            editorHeight: 300,
            isEditing: false,
            editingTitle: ''
          };
          loadPaneFromCache(newPane);
          panes.value.push(newPane);
        });
      }
    } catch (e) {
      // 解析保存的 panes keys 失败，忽略错误
    }
  }

  if (panes.value.length === 0) {
    const defaultPane: EditorPane = {
      title: 'SQLConsole 1',
      key: '1',
      closable: true,
      sql: '',
      sessionVars: '',
      characterSet: 'utf8',
      theme: 'default',
      result: null,
      loading: false,
      responseMsg: '',
      pagination: { currentPage: 1, pageSize: defaultPageSize, total: 0 },
      bottomActiveTab: 'result',
      editorHeight: 300,
      isEditing: false,
      editingTitle: ''
    };
    loadPaneFromCache(defaultPane);
    panes.value.push(defaultPane);
  }
  
  // 确保所有标签的状态都被正确恢复
  panes.value.forEach(pane => {
    loadPaneFromCache(pane);
  });
  
  loadPaneFromCache(currentPane.value);
  // 监听HTML元素的class变化，以便在dark模式切换时更新编辑器主题
  darkModeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  });
});

onUnmounted(() => {
  darkModeObserver.disconnect();
});
</script>

<template>
  <NCard size="small" class="das-page" :content-style="{ padding: appStore.isMobile ? '8px' : '12px' }" style="height: auto; min-height: calc(100vh - 120px)">
    <NGrid cols="24" :x-gap="appStore.isMobile ? 8 : 12" :y-gap="appStore.isMobile ? 8 : 12" responsive="screen" style="height: 100%">
      <!-- 桌面端左侧面板 -->
      <NGi v-if="showLeftPanel && !appStore.isMobile" :span="leftSpan">
        <NCard size="small" title="数据库选择" :segmented="{ content: true }" class="das-left-card" :style="leftContainerStyle" :content-style="{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }">
          <template #header-extra>
            <NSpace :size="6">
              <NTooltip trigger="hover" placement="top" :show-arrow="false">
                <template #trigger>
                  <NButton quaternary circle size="small" :loading="refreshLoading" @click="refreshSchemas">
                    <template #icon><SvgIcon icon="carbon:renew" /></template>
                  </NButton>
                </template>
                刷新数据库列表
              </NTooltip>
              <NTooltip trigger="hover" placement="top" :show-arrow="false">
                <template #trigger>
                  <NButton quaternary circle size="small" @click="foldLeft">
                    <template #icon><SvgIcon icon="line-md:menu-fold-left" /></template>
                  </NButton>
                </template>
                折叠左侧面板
              </NTooltip>
            </NSpace>
          </template>
          <div style="flex-shrink: 0; padding-bottom: 10px;">
            <NSpace vertical :size="appStore.isMobile ? 8 : 10">
              <NSelect
                v-model:value="bindTitle"
                :options="schemas.map((s: any) => ({
                  label: `${s.remark || s.instanceName || s.hostname}:${s.schema}`,
                  value: `${s.instance_id}#${s.schema}#${s.db_type}`
                }))"
                filterable
                clearable
                :size="appStore.isMobile ? 'small' : 'medium'"
                placeholder="请选择库名..."
                @update:value="getTables"
              />
              <NInput
                v-if="showSearch"
                v-model:value="leftTableSearch"
                clearable
                :size="appStore.isMobile ? 'small' : 'medium'"
                placeholder="输入要搜索的表名..."
                @keyup.enter="onSearch(leftTableSearch)"
              />
              <NText v-if="!appStore.isMobile" depth="3" class="das-hint">搜索不到需要的表？试试刷新按钮。</NText>
            </NSpace>
          </div>
          <div style="flex: 1; min-height: 0; overflow: hidden;">
            <NScrollbar style="height: 100%">
              <NSpin :show="treeLoading">
                <NTree
                  :data="filteredTreeData"
                  block-line
                  show-line
                  :virtual-scroll="true"
                  :render-label="renderTreeLabel"
                  :render-switcher-icon="renderSwitcherIcon"
                  :get-children="getNodeChildren"
                  v-model:expanded-keys="expandedKeys"
                  :node-props="(info: any) => ({ onDblclick: () => handleNodeDblClick(info.option.key) })"
                  @update:selected-keys="handleNodeClick"
                />
              </NSpin>
            </NScrollbar>
          </div>
        </NCard>
      </NGi>
      <NGi :span="rightSpan">
        <div ref="rightContainerRef" class="das-right-container" style="display: flex; flex-direction: column; height: 100%; gap: 12px">
          <NCard v-if="!showLeftPanel && !appStore.isMobile" size="small" class="das-ghost-card" :bordered="false">
            <NButton quaternary size="small" @click="foldLeft">
              <template #icon><SvgIcon icon="line-md:menu-fold-right" /></template>
              展开数据库面板
            </NButton>
          </NCard>
          <!-- 移动端显示左侧面板按钮 -->
          <NCard v-if="appStore.isMobile && !showLeftPanel" size="small" class="das-mobile-header" :bordered="false" :content-style="{ padding: '8px' }">
            <NSpace justify="space-between" align="center">
              <NButton quaternary size="small" @click="foldLeft">
                <template #icon><SvgIcon icon="line-md:menu-fold-right" /></template>
                数据库
              </NButton>
              <NSpace :size="8">
                <NButton quaternary size="small" @click="gotoFavorite">
                  <template #icon><SvgIcon icon="carbon:star" /></template>
                </NButton>
                <NButton quaternary size="small" @click="gotoHistory">
                  <template #icon><SvgIcon icon="carbon:time" /></template>
                </NButton>
              </NSpace>
            </NSpace>
          </NCard>
          <NCard size="small" class="das-editor-shell" :segmented="{ content: true }" style="flex: 1; min-height: 0; display: flex; flex-direction: column;" :content-style="{ flex: 1, overflow: 'auto' }">
            <template #header>
              <NSpace justify="space-between" align="center" class="das-shell-header" :wrap="appStore.isMobile">
                <div class="das-title">
                  <span class="das-title-text">SQL 工作台</span>
                  <NTag size="small" :type="schemaError ? 'error' : 'success'" class="das-title-tag">{{ currentSchemaLabel }}</NTag>
                  <NText v-if="schemaError" type="error" class="das-title-error">{{ schemaError }}</NText>
                </div>
                <!-- <NSpace :size="8" wrap>
                  <NButton quaternary size="small" @click="gotoFavorite">
                    <template #icon><SvgIcon icon="carbon:star" /></template>
                    收藏 SQL
                  </NButton>
                  <NButton quaternary size="small" @click="gotoHistory">
                    <template #icon><SvgIcon icon="carbon:time" /></template>
                    历史查询
                  </NButton>
                  <NButton quaternary size="small" @click="loadDBDictData">
                    <template #icon><SvgIcon icon="carbon:document" /></template>
                    数据字典
                  </NButton>
                  <NButton quaternary size="small" @click="refreshTables">
                    <template #icon><SvgIcon icon="carbon:renew" /></template>
                    刷新表
                  </NButton>
                </NSpace> -->
              </NSpace>
            </template>
            <NTabs
              v-model:value="activeKey"
              type="card"
              size="small"
              addable
              @add="add"
              @close="remove"
              @update:value="changeTab"
            >
              <NTabPane
                v-for="pane in panes"
                :key="pane.key"
                :name="pane.key"
                :closable="pane.closable"
              >
                <template #tab>
                  <div class="tab-title-container" @dblclick="handleRenameTab(pane)" @contextmenu="handleContextMenu($event, pane)">
                    <NInput
                      v-if="pane.isEditing"
                      v-model:value="pane.editingTitle"
                      size="tiny"
                      style="width: 120px"
                      @blur="saveTabName(pane)"
                      @keyup.enter="saveTabName(pane)"
                      @keyup.esc="cancelRenameTab(pane)"
                      @click.stop
                      autofocus
                    />
                    <template v-else>
                      <span>{{ pane.title }}</span>
                      <NButton
                        quaternary
                        circle
                        size="tiny"
                        class="edit-icon-btn"
                        @click.stop="handleRenameTab(pane)"
                      >
                        <template #icon>
                          <SvgIcon icon="carbon:edit" />
                        </template>
                      </NButton>
                    </template>
                  </div>
                </template>
                <NSpace vertical :size="0">
                  <div class="code-editor-wrapper">
                    <div class="code-editor-container" :style="{ height: (appStore.isMobile ? (pane.editorHeight || 200) : (pane.editorHeight || 300)) + 'px' }" :ref="(el) => setEditorRef(pane, el as unknown as HTMLElement)" @contextmenu="handleContextMenu($event, pane)" />
                    <div v-if="!appStore.isMobile" class="resize-handle" @mousedown.prevent="onResizeStart($event, pane)">
                      <div class="resize-handle-bar"></div>
                    </div>
                  </div>
                  <NTabs
                    v-model:value="pane.bottomActiveTab"
                    type="line"
                    size="small"
                    class="das-result-tabs"
                  >
                    <template #suffix>
                      <NSpace :size="appStore.isMobile ? 4 : 6" wrap>
                        <NTooltip v-if="!appStore.isMobile" trigger="hover" :show-arrow="false">
                          <template #trigger>
                            <NButton size="tiny" type="primary" :loading="pane.loading" @click="executeSQL(pane)">
                              <template #icon><SvgIcon icon="carbon:flash" /></template>
                              执行 SQL
                            </NButton>
                          </template>
                          {{ executeTooltip }}
                        </NTooltip>
                        <NButton v-else size="tiny" type="primary" :loading="pane.loading" @click="executeSQL(pane)">
                          <template #icon><SvgIcon icon="carbon:flash" /></template>
                        </NButton>
                        <NTooltip v-if="!appStore.isMobile" trigger="hover" :show-arrow="false">
                          <template #trigger>
                            <NButton size="tiny" @click="(e) => formatSQL(pane, e.shiftKey ? 'minify' : 'format')">
                              <template #icon><SvgIcon icon="carbon:code" /></template>
                              格式化
                            </NButton>
                          </template>
                          点击格式化，按住 Shift 点击压缩
                        </NTooltip>
                        <NButton v-else size="tiny" @click="(e) => formatSQL(pane, e.shiftKey ? 'minify' : 'format')">
                          <template #icon><SvgIcon icon="carbon:code" /></template>
                        </NButton>
                        <NButton size="tiny" :loading="dictLoading" @click="loadDBDictData">
                          <template #icon><SvgIcon icon="carbon:document" /></template>
                          <span v-if="!appStore.isMobile">数据字典</span>
                        </NButton>
                        <!-- <NButton size="tiny" @click="gotoFavorite">
                          <template #icon><SvgIcon icon="carbon:star" /></template>
                          收藏 SQL
                        </NButton>
                        <NButton size="tiny" @click="gotoHistory">
                          <template #icon><SvgIcon icon="carbon:time" /></template>
                          历史查询
                        </NButton> -->
                      </NSpace>
                    </template>
                    <NTabPane name="my_sql" tab="我的 SQL">
                      <div style="padding: 12px 0;">
                        <NTabs type="line" size="small" class="das-mysql-tabs">
                          <NTabPane name="history" tab="历史查询">
                            <History :embedded="true" @reuse="(sql) => handleReuseSQL(pane, sql)" />
                          </NTabPane>
                          <NTabPane name="favorite" tab="收藏 SQL">
                            <Favorite :embedded="true" @reuse="(sql) => handleReuseSQL(pane, sql)" />
                          </NTabPane>
                        </NTabs>
                      </div>
                    </NTabPane>
                    <NTabPane name="result" tab="执行结果">
                      <div class="das-result-pane">
                        <div class="das-result-toolbar" style="margin-bottom: 8px;">
                          <NSpace justify="end">
                            <TableColumnSetting v-model:columns="tableColumnChecks" />
                          </NSpace>
                        </div>
                        <div v-if="pane.result">
                          <div v-if="getTableColumns(pane).length > 0">
                            <template v-if="vxeReady">
                              <VxeTable
                                v-if="!useFallbackTable"
                                :data="getPagedTableData(pane)"
                                border
                                stripe
                                :height="appStore.isMobile ? 300 : 400"
                                :column-config="{ resizable: true }"
                                :resizable-config="{ showDragTip: false }"
                                :scroll-y="{ enabled: true, mode: 'wheel' }"
                                :scroll-x="{ enabled: true }"
                                :optimization="{ animat: false }"
                                show-overflow
                                empty-text="暂无数据"
                              >
                                <VxeColumn
                                  v-for="col in getVisibleColumns(pane)"
                                  :key="col.key"
                                  :field="col.key"
                                  :title="col.title"
                                  :min-width="col.minWidth || 120"
                                />
                              </VxeTable>
                              <NDataTable
                                v-else
                                :columns="getVisibleColumns(pane).map(c => ({ key: c.key, title: c.title, width: c.minWidth || 120, ellipsis: { tooltip: true } }))"
                                :data="getPagedTableData(pane)"
                                :max-height="appStore.isMobile ? 300 : 400"
                                size="small"
                                bordered
                                striped
                                class="das-result-fallback-table"
                              />
                            <div class="das-result-meta">
                              <div class="das-result-stat">
                                <SvgIcon icon="carbon:checkmark" class="text-16px text-#18a058" />
                                <NText type="success">执行成功</NText>
                                <NText>当前返回 [{{ getTableData(pane).length }}] 行</NText>
                                <SvgIcon icon="carbon:time" class="ml-8px text-16px text-#2080f0" />
                                <NText type="info">耗时 [{{ pane.result?.duration ?? '-' }}]</NText>
                              </div>
                              <NPagination
                                v-model:page="pane.pagination!.currentPage"
                                v-model:page-size="pane.pagination!.pageSize"
                                :item-count="pane.pagination?.total ?? getTableData(pane).length"
                                :page-sizes="pageSizes"
                                show-size-picker
                                size="small"
                                :page-slot="9"
                                @update:page="(p) => onPageChange(pane, p, pane.pagination!.pageSize)"
                                @update:page-size="(s) => onPageChange(pane, pane.pagination!.currentPage, s)"
                              />
                            </div>
                            </template>
                            <div v-else class="flex-y-center gap-8px py-24px">
                              <NSpin size="small" />
                              <span class="text-14px text-gray-500">加载表格组件中...</span>
                            </div>
                          </div>
                          <div v-else-if="pane.responseMsg && pane.responseMsg.includes('执行失败')" class="das-result-error">
                            <div v-html="pane.responseMsg" style="padding: 16px; line-height: 1.8; color: #333;"></div>
                          </div>
                          <div v-else class="das-empty-holder">
                            <NEmpty description="暂无查询结果" />
                          </div>
                        </div>
                        <div v-else-if="pane.responseMsg && pane.responseMsg.includes('执行失败')" class="das-result-error">
                          <div v-html="pane.responseMsg" style="padding: 16px; line-height: 1.8; color: #333;"></div>
                        </div>
                        <div v-else class="das-empty-holder">
                          <NEmpty description="暂无查询结果" />
                        </div>
                      </div>
                    </NTabPane>
                  </NTabs>
                </NSpace>
              </NTabPane>
            </NTabs>
          </NCard>
        </div>
      </NGi>
    </NGrid>
    
    <!-- 移动端左侧面板（全屏覆盖） -->
    <NCard
      v-if="showLeftPanel && appStore.isMobile"
      size="small"
      title="数据库选择"
      :segmented="{ content: true }"
      class="das-left-card das-mobile-left-panel"
      :style="leftContainerStyle"
      :content-style="{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }"
    >
      <template #header-extra>
        <NSpace :size="4">
          <NButton quaternary circle size="small" :loading="refreshLoading" @click="refreshSchemas">
            <template #icon><SvgIcon icon="carbon:renew" /></template>
          </NButton>
          <NButton quaternary circle size="small" @click="foldLeft">
            <template #icon><SvgIcon icon="line-md:menu-fold-left" /></template>
          </NButton>
        </NSpace>
      </template>
      <div style="flex-shrink: 0; padding-bottom: 10px;">
        <NSpace vertical :size="8">
          <NSelect
            v-model:value="bindTitle"
            :options="schemas.map((s: any) => ({
              label: `${s.remark || s.instanceName || s.hostname}:${s.schema}`,
              value: `${s.instance_id}#${s.schema}#${s.db_type}`
            }))"
            filterable
            clearable
            size="small"
            placeholder="请选择库名..."
            @update:value="getTables"
          />
          <NInput
            v-if="showSearch"
            v-model:value="leftTableSearch"
            clearable
            size="small"
            placeholder="输入要搜索的表名..."
            @keyup.enter="onSearch(leftTableSearch)"
          />
        </NSpace>
      </div>
      <div style="flex: 1; min-height: 0; overflow: hidden;">
        <NScrollbar style="height: 100%">
          <NSpin :show="treeLoading">
            <NTree
              :data="filteredTreeData"
              block-line
              show-line
              :virtual-scroll="true"
              :render-label="renderTreeLabel"
              :render-switcher-icon="renderSwitcherIcon"
              :get-children="getNodeChildren"
              v-model:expanded-keys="expandedKeys"
              :node-props="(info: any) => ({ onDblclick: () => handleNodeDblClick(info.option.key) })"
              @update:selected-keys="handleNodeClick"
            />
          </NSpin>
        </NScrollbar>
      </div>
    </NCard>
    
    <NModal
      v-model:show="showDictModal"
      preset="card"
      :style="{ width: appStore.isMobile ? '95%' : '90%', height: appStore.isMobile ? '85vh' : '90vh', maxWidth: appStore.isMobile ? '100%' : '1600px' }"
      :title="`数据字典: ${selectedSchema.schema || ''}`"
      :bordered="false"
      size="huge"
    >
      <div style="width: 100%; height: 100%; overflow: hidden; border-radius: 4px; border: 1px solid var(--n-border-color);">
        <iframe
          :srcdoc="dictHtmlContent"
          style="width: 100%; height: 100%; border: none;"
          sandbox="allow-scripts"
        ></iframe>
      </div>
    </NModal>

    <NDropdown
      placement="bottom-start"
      trigger="manual"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuOptions"
      :show="showContextMenu"
      :on-clickoutside="onClickOutside"
      @select="handleContextSelect"
    />
  </NCard>
</template>

<style scoped>
.das-page {
  height: calc(100vh - 120px);
}

/* 移动端适配 */
@media (max-width: 640px) {
  .das-page {
    height: calc(100vh - 80px);
  }
  
  .das-right-container {
    gap: 8px !important;
  }
  
  .das-mobile-header {
    margin-bottom: 4px;
  }
  
  .das-shell-header {
    flex-wrap: wrap;
    gap: 8px !important;
  }
  
  .das-title {
    flex-wrap: wrap;
    gap: 4px !important;
    font-size: 14px;
  }
  
  .das-title-text {
    font-size: 14px;
  }
  
  .das-title-tag {
    font-size: 11px;
    padding: 2px 6px;
  }
  
  .das-title-error {
    font-size: 11px !important;
    display: block;
    width: 100%;
    margin-top: 4px;
  }
  
  .code-editor-container {
    font-size: 12px !important;
  }
  
  .code-editor-container :deep(.cm-editor) {
    font-size: 12px !important;
  }
  
  .das-editor-actions {
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .das-result-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .das-result-stat {
    font-size: 11px;
    gap: 8px;
  }
  
  .tab-title-container {
    max-width: 100px;
  }
  
  .das-mobile-left-panel {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 1000;
    width: 100%;
    height: 100vh;
    border-radius: 0;
    background-color: var(--n-color);
  }
  
  .das-left-card {
    height: 100%;
  }
}
.tab-title-container {
  display: flex;
  align-items: center;
  gap: 4px;
}
.tab-title-container .edit-icon-btn {
  opacity: 0.6;
  transition: opacity 0.2s;
  margin-left: 2px;
  padding: 0;
  width: 18px;
  height: 18px;
}
.tab-title-container:hover .edit-icon-btn {
  opacity: 0.8;
}
.tab-title-container .edit-icon-btn:hover {
  opacity: 1 !important;
}
.das-left-card {
  height: 100%;
}
.das-shell-header .das-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}
.das-subtitle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
}
.das-editor-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: 8px;
}
.das-result-card {
  padding-bottom: 4px;
}
.das-result-meta {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.das-result-stat {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}
.das-empty-holder {
  padding: 32px 0;
}
.das-hint {
  font-size: 12px;
}
/* 美化 SQL 编辑框：容器边框、圆角与内边距 */
.code-editor-wrapper {
  display: flex;
  flex-direction: column;
  margin-bottom: 8px;
}
.code-editor-container {
  border: 1px solid var(--n-border-color);
  border-bottom: none;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  background-color: var(--n-color);
  resize: none;
  overflow: hidden;
}
.resize-handle {
  height: 12px;
  cursor: row-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--n-color);
  border: 1px solid var(--n-border-color);
  border-top: 1px solid var(--n-border-color);
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
  transition: background-color 0.2s;
}
.resize-handle:hover {
  background-color: var(--n-action-color);
}
.resize-handle-bar {
  width: 32px;
  height: 4px;
  border-radius: 2px;
  background-color: var(--n-border-color);
}
.code-editor-container :deep(.cm-editor) {
  background-color: transparent;
  font-family: 'JetBrains Mono', 'Fira Code', Menlo, Monaco, 'Courier New', monospace;
  font-size: 13px;
  height: 100%;
}
.code-editor-container :deep(.cm-scroller) {
  height: 100%;
  padding: 4px;
  overflow: auto;
}
.code-editor-container :deep(.cm-gutters) {
  background-color: var(--n-color);
  border-right: 1px solid var(--n-border-color);
}
.code-editor-container :deep(.cm-activeLine) {
  background-color: rgba(0, 0, 0, 0.03);
}
.code-editor-container :deep(.cm-editor.cm-focused) {
  outline: none;
}
/* 左侧 NTree 自定义节点样式 */
:deep(.n-tree .n-tree-node) {
  --das-tree-type-color: var(--primary-color);
}
:deep(.n-tree) {
  /* 紧凑化节点前缀与内容之间的间距 */
  --n-node-gap: 0px;
}

:deep(.das-tree-item) {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
/* 表节点：两行布局 */
:deep(.das-tree-item-table) {
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}
:deep(.das-tree-item-meta-row) {
  display: inline-flex;
  width: 100%;
  justify-content: flex-start;
  gap: 8px;
}
:deep(.das-tree-item-left) {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  min-width: 0;
  flex: 1 1 auto;
}
:deep(.das-tree-item-left > .iconify),
:deep(.das-tree-item-left > svg) {
  display: inline-block;
  flex: 0 0 auto;
  vertical-align: middle;
}
:deep(.das-tree-item-name) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
:deep(.das-tree-item-type),
:deep(.das-tree-item-count) {
  color: var(--das-tree-type-color);
  font-size: 12px;
  flex: 0 0 auto;
  white-space: nowrap;
  margin-right: 12px;
}
/* 覆盖：表节点的数量允许换行（正常渲染） */
:deep(.das-tree-item-table .das-tree-item-count) {
  white-space: normal;
  margin-right: 0;
}
/* 新增：可点击的“列(数量)”视觉样式 */
:deep(.das-tree-item-count-toggle) {
  cursor: pointer;
  user-select: none;
  display: inline-flex;
  align-items: center;
  gap: 0px;
  white-space: nowrap;
}
:deep(.das-tree-item-count-toggle > .iconify),
:deep(.das-tree-item-count-toggle > svg) {
  display: inline-block;
  flex: 0 0 auto;
  vertical-align: middle;
}
:deep(.n-tree-node:hover) .das-tree-item-type,
:deep(.n-tree-node:hover) .das-tree-item-count {
  filter: saturate(1.2);
}
/* 新增：加/减号开关图标样式 */
:deep(.das-tree-switcher-icon) {
  font-size: 14px;
  line-height: 1;
  vertical-align: middle;
}
/* 收紧展开图标与节点内容之间的间距 */
:deep(.n-tree-node-switcher) {
  margin-right: 0px !important;
  transform: none !important;
}
/* 进一步收紧节点内容间距 */
:deep(.n-tree-node-content) {
  gap: 0px !important;
  padding-left: 0px !important;
  margin-left: 0px !important;
}
:deep(.n-tree-node-content__prefix) {
  margin-right: 0px !important;
}
</style>

<style>
/* vxe-table Dark Mode Support - Global Overrides */
html.dark .vxe-table--render-default {
  color: rgba(255, 255, 255, 0.82) !important;
  background-color: transparent !important;
}

html.dark .vxe-table--header-wrapper,
html.dark .vxe-table--body-wrapper,
html.dark .vxe-table--footer-wrapper {
  background-color: #18181c !important;
}

html.dark .vxe-header--row .vxe-header--column,
html.dark .vxe-body--row .vxe-body--column,
html.dark .vxe-footer--row .vxe-footer--column {
  background-color: #18181c !important;
  background-image: none !important;
  color: rgba(255, 255, 255, 0.82) !important;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09) !important;
  border-right: 1px solid rgba(255, 255, 255, 0.09) !important;
}

/* Hover effect */
html.dark .vxe-body--row.row--hover,
html.dark .vxe-body--row.row--hover .vxe-body--column {
  background-color: rgba(255, 255, 255, 0.08) !important;
}

/* Stripe effect */
html.dark .vxe-body--row.row--stripe,
html.dark .vxe-body--row.row--stripe .vxe-body--column {
  background-color: rgba(255, 255, 255, 0.04) !important;
}

/* Borders */
html.dark .vxe-table--border-line {
  border-color: rgba(255, 255, 255, 0.09) !important;
}

/* Scrollbar */
html.dark .vxe-table--body-wrapper::-webkit-scrollbar {
  width: 8px;
  height: 8px;
  background-color: #18181c;
}
html.dark .vxe-table--body-wrapper::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}
html.dark .vxe-table--body-wrapper::-webkit-scrollbar-track {
  background-color: #18181c;
}

/* CodeMirror Cursor Color - Dark Mode - 使用更强的选择器 */
html.dark .cm-editor .cm-cursorLayer .cm-cursor,
html.dark .cm-editor .cm-cursorLayer .cm-cursor-primary,
html.dark .cm-editor.cm-focused .cm-cursorLayer .cm-cursor,
html.dark .cm-editor.cm-focused .cm-cursorLayer .cm-cursor-primary {
  border-left: 1.2px solid #fff !important;
}
html.dark .cm-cursor,
html.dark .cm-cursor-primary {
  border-left: 1.2px solid #fff !important;
}
html.dark .cm-dropCursor {
  border-left: 1.2px solid #fff !important;
}
html.dark .cm-content,
html.dark .cm-line {
  caret-color: #fff !important;
}

/* CodeMirror Cursor Color - Light Mode */
html:not(.dark) .cm-editor .cm-cursorLayer .cm-cursor,
html:not(.dark) .cm-editor .cm-cursorLayer .cm-cursor-primary,
html:not(.dark) .cm-editor.cm-focused .cm-cursorLayer .cm-cursor {
  border-left: 1.2px solid #000 !important;
}
html:not(.dark) .cm-cursor,
html:not(.dark) .cm-cursor-primary {
  border-left: 1.2px solid #000 !important;
}
html:not(.dark) .cm-dropCursor {
  border-left: 1.2px solid #000 !important;
}
html:not(.dark) .cm-content {
  caret-color: #000 !important;
}
</style>