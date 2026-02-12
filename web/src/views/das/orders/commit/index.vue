<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  NText,
  useMessage
} from 'naive-ui';
import type { PaginationProps } from 'naive-ui';
import { format } from 'sql-formatter';
// CodeMirror 6 imports (align with SQL 查询页)
import { EditorState, Compartment } from '@codemirror/state';
import { EditorView, keymap, lineNumbers } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
import { sql } from '@codemirror/lang-sql';
import { defaultHighlightStyle, foldGutter, foldKeymap, syntaxHighlighting } from '@codemirror/language';
import { autocompletion, completionKeymap } from '@codemirror/autocomplete';
import {
  fetchCreateOrder,
  fetchOrdersEnvironments,
  fetchOrdersInstances,
  fetchOrdersSchemas,
  fetchOrdersUsers,
  fetchOrderTables,
  fetchSyntaxCheck
} from '@/service/api/orders';
import { useAppStore } from '@/store/modules/app';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const appStore = useAppStore();

// 页面标题与工单类型
const sqlType = ref<string>('DDL');
const pageTitle = computed(() => `提交${sqlType.value}工单`);
const isExportOrder = computed(() => sqlType.value.toLowerCase() === 'export');

// 表单模型
interface FormModel {
  title: string;
  remark?: string;
  isRestrictAccess: boolean;
  dbType: 'MySQL' | 'TiDB';
  environment?: number | null;
  instanceId?: string | null;  // UUID 字符串
  schema?: string | null;
  exportFileFormat?: 'XLSX' | 'CSV';
  approver: string[]; // username list
  executor: string[]; // username list
  reviewer: string[]; // username list
  cc: string[]; // username list
  content: string;
  scheduleTime?: string | null;
}

const formModel = reactive<FormModel>({
  title: '',
  remark: '',
  isRestrictAccess: true,
  dbType: 'MySQL',
  environment: null,
  instanceId: null,
  schema: null,
  exportFileFormat: 'XLSX',
  approver: [],
  executor: [],
  reviewer: [],
  cc: [],
  content: '',
  scheduleTime: null
});

// 下拉数据源
const environments = ref<any[]>([]);
const instances = ref<any[]>([]);
const schemas = ref<any[]>([]);
const users = ref<any[]>([]);

// 加载器状态
const loading = ref(false);
const checking = ref(false);

function inferSqlTypeFromPath() {
  const seg = route.path.split('/').pop()?.toUpperCase();
  if (seg && ['DDL', 'DML', 'EXPORT'].includes(seg)) {
    sqlType.value = seg as any;
  } else {
    sqlType.value = 'DDL';
  }
}

async function loadEnvironments() {
  const res = await fetchOrdersEnvironments({ is_page: false } as any);
  environments.value = (res as any)?.data ?? [];
}

async function loadInstances(envId: number | null) {
  if (!envId) {
    instances.value = [];
    return;
  }
  const res = await fetchOrdersInstances({ id: envId, db_type: formModel.dbType, is_page: false } as any);
  instances.value = (res as any)?.data ?? [];
}

async function loadSchemas(instanceId: number | null) {
  if (!instanceId) {
    schemas.value = [];
    return;
  }
  const res = await fetchOrdersSchemas({ instance_id: instanceId, is_page: false } as any);
  schemas.value = (res as any)?.data ?? [];
  // 如果已选择 schema，加载表结构用于补全
  if (formModel.schema && formModel.instanceId) {
    await loadTablesForCompletion();
  }
}

// 加载表结构用于 SQL 补全
async function loadTablesForCompletion() {
  if (!formModel.instanceId || !formModel.schema) {
    tabCompletion.value = { tables: {}, metadata: { tables: {} } };
    updateEditorSchema();
    return;
  }

  try {
    // 使用工单专用接口，不检查 DAS 查询权限
    const { data, error } = await fetchOrderTables({
      instance_id: formModel.instanceId!,
      schema: formModel.schema!
    });
    
    if (error || !data) {
      tabCompletion.value = { tables: {}, metadata: { tables: {} } };
      updateEditorSchema();
      return;
    }

    // 构建补全数据
    const tmpTabCompletion: any = { tables: {}, metadata: { tables: {} } };
    
    data.forEach((row: any) => {
      // 列分隔符是 @@，字段内部用 $$ 分隔 (name$$type$$comment)
      const columns = (row.columns || '').split('@@').filter((v: string) => v);
      const columnsCompletion: string[] = [];
      const tmpColumnsData: any[] = [];

      columns.forEach((v: string) => {
        if (!v) return;
        const parts = v.split('$$');
        const colName = parts[0];
        const colType = parts.length > 1 ? parts[1] : '';

        tmpColumnsData.push({
          name: colName,
          type: colType
        });
        columnsCompletion.push(colName);
      });

      tmpTabCompletion.tables[row.table_name] = columnsCompletion;
      tmpTabCompletion.metadata.tables[row.table_name] = {
        comment: row.table_comment || '',
        columns: tmpColumnsData
      };
    });

    tabCompletion.value = tmpTabCompletion;
    updateEditorSchema();
  } catch (error: any) {
    console.error('加载表结构失败:', error);
    tabCompletion.value = { tables: {}, metadata: { tables: {} } };
    updateEditorSchema();
  }
}

// 更新编辑器 schema
function updateEditorSchema() {
  if (!editorView.value || !languageCompartment.value) return;
  editorView.value.dispatch({
    effects: languageCompartment.value.reconfigure(
      sql({ schema: schemaForCompletion(), upperCaseKeywords: true })
    )
  });
}

// 获取补全 schema
function schemaForCompletion(): Record<string, string[]> {
  const tablesMap = (tabCompletion.value?.tables || {}) as Record<string, string[]>;
  return tablesMap;
}

// 自定义 SQL 补全逻辑：支持表名（带注释）和字段名（带类型），包含上下文感知
function customSQLCompletion(context: any) {
  const metadata = tabCompletion.value?.metadata || { tables: {} };
  
  // 1. 尝试匹配 "Table.Column" 模式 (检测点号)
  const dotMatch = context.matchBefore(/(\w+)\.(\w*)$/);
  
  if (dotMatch) {
    const tableName = dotMatch.text.split('.')[0];
    const tableInfo = metadata.tables[tableName];
    
    if (tableInfo && tableInfo.columns) {
      const options = tableInfo.columns.map((col: any) => ({
        label: col.name,
        type: 'column',
        detail: col.type,
        boost: 10
      }));
      
      return {
        from: dotMatch.from + tableName.length + 1,
        options,
        validFor: /^\w*$/
      };
    }
  }

  // 2. 默认模式：提示表名 + 上下文相关字段
  // 将 \w+ 改为 \w* 以支持光标在空格后立即触发提示（此时匹配为空字符串）
  const wordMatch = context.matchBefore(/\w*$/);
  
  if (wordMatch) {
    const tableNames = Object.keys(metadata.tables);
    // 即使没有表数据，也提供关键字补全
    
    const options: any[] = [];
    
    // 扫描当前文档中出现过的表名（上下文感知）
    const docText = context.state.doc.toString();
    
    // 解析当前 SQL 语句中实际使用的表名（FROM 和 JOIN 后面的表）
    const usedTablesInSQL = new Set<string>();
    const fromJoinPattern = /\b(?:FROM|JOIN)\s+[`"']?(\w+)[`"']?/gi;
    let tableMatch;
    while ((tableMatch = fromJoinPattern.exec(docText)) !== null) {
      const tblName = tableMatch[1];
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
    const textBefore = context.state.sliceDoc(0, context.pos);
    const lastKeywordMatch = textBefore.match(/\b(SELECT|FROM|JOIN|WHERE|GROUP\s+BY|ORDER\s+BY|LIMIT|SET|UPDATE|DELETE|INSERT|HAVING|ON|AND|OR)\b/gi);
    const lastKeyword = lastKeywordMatch ? lastKeywordMatch[lastKeywordMatch.length - 1].toUpperCase().replace(/\s+/g, ' ') : '';
    
    const isTableContext = ['FROM', 'JOIN', 'UPDATE', 'INTO'].includes(lastKeyword);
    const isConditionContext = ['WHERE', 'HAVING', 'ON', 'AND', 'OR'].includes(lastKeyword);
    const isSelectContext = lastKeyword === 'SELECT';
    
    // 获取当前正在输入的单词之前的上一个有效 token
    const textBeforeCurrentWord = context.state.sliceDoc(0, wordMatch.from);
    
    // = 号后面不应该弹出提示（通常是在输入值）
    if (!context.explicit && /[=<>!]+\s*$/.test(textBeforeCurrentWord)) {
      return null;
    }

    const prevTokenMatch = textBeforeCurrentWord.match(/([`"']?[\w.]+\b[`"']?)\s*$/);
    const prevToken = prevTokenMatch ? prevTokenMatch[1].replace(/[`"']/g, '') : '';
    const prevTokenField = prevToken.includes('.') ? prevToken.split('.').pop() : prevToken;

    // 收集所有已知的字段名
    const allColumnNames = new Set<string>();
    Object.values(metadata.tables).forEach((t: any) => {
      if (t.columns) {
        t.columns.forEach((c: any) => allColumnNames.add(c.name));
      }
    });

    const isStartWithO = wordMatch.text.toLowerCase().startsWith('o');
    const isPrevTokenColumn = allColumnNames.has(prevTokenField);
    
    // 如果处于条件上下文且上一个词是字段名，提示运算符
    if (isConditionContext && isPrevTokenColumn) {
       const operators = [
         { label: '=', type: 'keyword', detail: 'Operator', boost: 100 },
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
         { label: 'NOT LIKE', type: 'keyword', detail: 'Operator', boost: 80 }
       ];
       options.push(...operators);
    }

    // 如果处于 SELECT 上下文，提示聚合函数和常用常量
    if (isSelectContext) {
       const isAfterStar = /\*\s*$/.test(textBeforeCurrentWord);
       
       if (isAfterStar) {
          const afterStarOptions = [
            { label: 'FROM', type: 'keyword', detail: 'Keyword', boost: 60 },
            { label: 'FALSE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'TRUE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'NULL', type: 'constant', detail: 'Value', boost: 40 }
          ];
          options.push(...afterStarOptions);
       } else {
          const selectOptions = [
            { label: '*', type: 'keyword', detail: 'All Columns', boost: 50 },
            { label: 'DISTINCT()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'COUNT()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'MAX()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'MIN()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'SUM()', type: 'function', detail: 'Function', boost: 45 },
            { label: 'FALSE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'TRUE', type: 'constant', detail: 'Boolean', boost: 40 },
            { label: 'NULL', type: 'constant', detail: 'Value', boost: 40 }
          ];
          options.push(...selectOptions);
       }
    }

    // 如果上一个词是表名，提示连接查询关键字和 WHERE
    if (metadata.tables[prevToken]) {
       const afterTableOptions = [
        { label: 'WHERE', type: 'keyword', detail: 'Keyword', boost: 60 },
        ...(isStartWithO ? [] : [
          { label: 'ORDER BY', type: 'keyword', detail: 'Keyword', boost: 58 },
        ]),
        { label: 'GROUP BY', type: 'keyword', detail: 'Keyword', boost: 57 },
         { label: 'LIMIT', type: 'keyword', detail: 'Keyword', boost: 56 },
         { label: ',', type: 'keyword', detail: 'Separator', boost: 55 },
         { label: 'INNER', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'LEFT JOIN', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'RIGHT JOIN', type: 'keyword', detail: 'Keyword', boost: 50 },
         { label: 'JOIN', type: 'keyword', detail: 'Keyword', boost: 45 },
         { label: 'HAVING', type: 'keyword', detail: 'Keyword', boost: 40 }
       ];
       options.push(...afterTableOptions);
    }

    // 检测 ORDER BY / GROUP BY 上下文
    const isOrderByContext = lastKeyword === 'ORDER BY';
    const isGroupByContext = lastKeyword === 'GROUP BY';
    
    if (isOrderByContext || isGroupByContext) {
      if (isOrderByContext) {
        options.push(
          { label: 'ASC', type: 'keyword', detail: 'Ascending', boost: 50 },
          { label: 'DESC', type: 'keyword', detail: 'Descending', boost: 50 }
        );
      }
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
        { label: 'ORDER', type: 'keyword', detail: 'Keyword', boost: 1600 }
      ];
      options.push(...oKeywords);
    }

    // 通用 SQL 关键字提示
    const commonKeywords = [
      { label: 'SELECT', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'FROM', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'WHERE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'AND', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'OR', type: 'keyword', detail: 'Keyword', boost: -10 },
      ...(isStartWithO ? [] : [
        { label: 'ORDER BY', type: 'keyword', detail: 'Keyword', boost: 5 },
      ]),
      { label: 'GROUP BY', type: 'keyword', detail: 'Keyword', boost: 5 },
      { label: 'HAVING', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LIMIT', type: 'keyword', detail: 'Keyword', boost: 5 },
      { label: 'JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LEFT JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'RIGHT JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'INNER JOIN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'ON', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'AS', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'INSERT INTO', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'UPDATE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'DELETE FROM', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'SET', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'BETWEEN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IN', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'LIKE', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IS NULL', type: 'keyword', detail: 'Keyword', boost: -10 },
      { label: 'IS NOT NULL', type: 'keyword', detail: 'Keyword', boost: -10 },
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
          let colBoost = 50;
          if (isPrevTokenColumn && isConditionContext) {
            colBoost = 20;
          } else if (isTableContext) {
            colBoost = -5;
          }
          
          options.push({
            label: col.name,
            type: 'column',
            detail: `${col.type} · ${tableName}`,
            boost: colBoost 
          });
        });
      }
    });

    // 添加所有表名
    tableNames.forEach((name) => {
      const info = metadata.tables[name];
      const tableBoost = isTableContext ? 20 : -1;
      
      options.push({
        label: name,
        type: 'table',
        detail: info?.comment || '表',
        boost: tableBoost 
      });

      // 只有当表不在当前 SQL 使用的表中时，才添加其字段
      if (!usedTablesInSQL.has(name) && info.columns) {
        info.columns.forEach((col: any) => {
          const colBoost = isTableContext ? -10 : -5;
          
          options.push({
            label: col.name,
            type: 'column',
            detail: `${col.type} · ${name}`,
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

async function loadUsers() {
  const res = await fetchOrdersUsers();
  // admin/users 返回分页格式 { list: [...], total: ... }
  users.value = (res as any)?.data?.list ?? (res as any)?.data ?? [];
}

// 编辑器设置：右侧改为与 SQL 查询页一致的 CodeMirror
const editorRoot = ref<HTMLElement | null>(null);
const editorView = ref<EditorView | null>(null);
const languageCompartment = ref<Compartment | null>(null);
const tabCompletion = ref<any>({ tables: {}, metadata: { tables: {} } });

function initEditor() {
  if (editorView.value || !editorRoot.value) return;
  
  // 创建 Compartment 用于动态更新 schema
  languageCompartment.value = new Compartment();
  
  const state = EditorState.create({
    doc: formModel.content || '',
    extensions: [
      lineNumbers(),
      foldGutter(),
      languageCompartment.value.of(sql({ schema: schemaForCompletion(), upperCaseKeywords: true })),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      history(),
      keymap.of([...defaultKeymap, ...historyKeymap, ...foldKeymap, ...completionKeymap, indentWithTab]),
      autocompletion({ activateOnTyping: true, maxRenderedOptions: 50 }),
      EditorState.languageData.of(() => [{ autocomplete: customSQLCompletion }]),
      EditorView.updateListener.of(v => {
        if (v.docChanged) {
          formModel.content = v.state.doc.toString();
        }
      })
    ]
  });
  editorView.value = new EditorView({ state, parent: editorRoot.value });
}

// 外部更新（如格式化）时同步到编辑器
watch(
  () => formModel.content,
  val => {
    syntaxStatus.value = null;
    syntaxRows.value = [];

    const view = editorView.value;
    if (!view) return;
    const cur = view.state.doc.toString();
    if (cur !== (val || '')) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: val || '' } });
    }
  }
);

// 监听 tabCompletion 变化，更新编辑器 schema
watch(
  () => tabCompletion.value,
  () => {
    updateEditorSchema();
  },
  { deep: true }
);

// 监听 instanceId 和 schema 变化，加载表结构用于补全
watch(
  () => formModel.instanceId,
  async (newVal) => {
    if (newVal && formModel.schema) {
      await loadTablesForCompletion();
    }
  }
);

watch(
  () => formModel.schema,
  async (newVal) => {
    if (newVal && formModel.instanceId) {
      await loadTablesForCompletion();
    }
  }
);

const leftContentRef = ref<HTMLElement | null>(null);
const leftContentHeight = ref<number>(0);
let leftResizeObserver: ResizeObserver | null = null;

onMounted(async () => {
  inferSqlTypeFromPath();
  await loadEnvironments();
  await loadUsers();
  initEditor();
  
  // 检查是否有回滚工单参数（从执行结果页面跳转过来）
  const query = route.query;
  if (query.rollback_sql) {
    // 如果是回滚工单，设置为 DML 类型
    sqlType.value = 'DML';
    
    // 预填充基本信息
    if (query.title) {
      formModel.title = decodeURIComponent(query.title as string);
    }
    if (query.db_type) {
      formModel.dbType = query.db_type as 'MySQL' | 'TiDB';
    }
    
    // 不预填充 instance_id 和 schema，让用户自己选择
    
    // 预填充回滚 SQL（最后设置，确保编辑器已初始化）
    const rollbackSQL = decodeURIComponent(query.rollback_sql as string);
    formModel.content = rollbackSQL;
    
    // 更新编辑器内容
    if (editorView.value) {
      const transaction = editorView.value.state.update({
        changes: { from: 0, to: editorView.value.state.doc.length, insert: rollbackSQL }
      });
      editorView.value.dispatch(transaction);
    }
  }
  
  // 观察左侧卡片内容高度变化，右侧编辑器按此高度限制（仅桌面端）
  if (leftContentRef.value && !appStore.isMobile) {
    leftResizeObserver = new ResizeObserver(entries => {
      const entry = entries[0];
      if (entry) {
        leftContentHeight.value = Math.round(entry.contentRect.height);
      }
    });
    leftResizeObserver.observe(leftContentRef.value);
  } else if (appStore.isMobile) {
    // 移动端设置固定高度
    leftContentHeight.value = 400;
  }
});

onUnmounted(() => {
  if (editorView.value) {
    editorView.value.destroy();
    editorView.value = null;
  }
  if (leftResizeObserver) {
    leftResizeObserver.disconnect();
    leftResizeObserver = null;
  }
});

function onDBTypeChange() {
  formModel.environment = null;
  formModel.instanceId = null;
  formModel.schema = null;
  instances.value = [];
  schemas.value = [];
}

async function onEnvironmentChange(val: number) {
  formModel.instanceId = null;
  formModel.schema = null;
  await loadInstances(val ?? null);
}

async function onInstanceChange(val: number) {
  formModel.schema = null;
  await loadSchemas(val ?? null);
}

function formatSQL() {
  try {
    formModel.content = format(formModel.content || '', { language: 'mysql' });
    message.success('格式化完成');
  } catch (e) {
    message.error('格式化失败');
  }
}

const syntaxRows = ref<any[]>([]);
const syntaxStatus = ref<number | null>(null);
const showFingerprint = ref(false);

// 计算提交按钮是否应该禁用
const isSubmitDisabled = computed(() => {
  // 语法检查未通过（status !== 0）或未检查（status === null）时禁用
  return syntaxStatus.value !== 0;
});
const visibleSyntaxColumns = computed(() =>
  syntaxColumns.filter((col: any) => col.key !== 'finger_id' || showFingerprint.value)
);
function isPass(row: any) {
  return row?.level === 'INFO' && (!row?.summary || row.summary.length === 0);
}
const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  itemCount: 0,
  pageSizes: [10, 20, 50, 100],
  onUpdatePage: (page: number) => {
    pagination.page = page;
  },
  onUpdatePageSize: (size: number) => {
    pagination.pageSize = size;
    pagination.page = 1;
  }
});
watch(syntaxRows, rows => {
  pagination.itemCount = rows.length;
  pagination.page = 1;
});
const syntaxColumns = [
  { title: '错误级别', key: 'level', width: 80 },
  { title: '影响行数', key: 'affected_rows', width: 90 },
  { title: '类型', key: 'type', width: 90 },
  { title: '指纹', key: 'finger_id', width: 120 },
  {
    title: '信息提示',
    key: 'summary',
    width: 300,
    ellipsis: { tooltip: true },
    render: (row: any) => (row.summary && row.summary.length ? row.summary.join('；') : '—')
  },
  { title: 'SQL', key: 'query', width: 500, ellipsis: { tooltip: true } },
  {
    title: '检测结果',
    key: 'result',
    width: 100,
    render: (row: any) =>
      h(NTag, { type: isPass(row) ? 'success' : 'error' }, { default: () => (isPass(row) ? '通过' : '失败') })
  }
];
async function syntaxCheck() {
  if (!formModel.content) {
    message.warning('输入内容不能为空');
    return;
  }
  if (!formModel.environment) {
    message.warning('请选择环境');
    return;
  }
  if (!formModel.instanceId) {
    message.warning('请选择实例');
    return;
  }
  if (!formModel.schema) {
    message.warning('请选择库名');
    return;
  }
  checking.value = true;
  syntaxStatus.value = null;
  syntaxRows.value = [];
  try {
    const data = {
      db_type: formModel.dbType,
      sql_type: sqlType.value,
      instance_id: formModel.instanceId,
      schema: formModel.schema,
      content: formModel.content
    };
    const resp: any = await fetchSyntaxCheck(data as any);
    console.log('语法检查响应:', resp);

    // createFlatRequest 返回 { data, error, response } 格式
    // 错误时: data=null, error=AxiosError
    // 成功时: data=响应数据, error=null
    
    // 检查是否有错误（error 不为 null 或 data 为 null）
    if (resp.error || resp.data === null || resp.data === undefined) {
      // 请求失败
      syntaxStatus.value = 1;
      // 尝试从错误响应中获取详细数据用于展示
      const errorData = resp.error?.response?.data?.data ?? resp.response?.data?.data ?? [];
      syntaxRows.value = Array.isArray(errorData) ? errorData : [];
      // 错误消息已由全局拦截器显示，不再重复
      return;
    }

    // 请求成功，data 格式：{status: 0/1, data: [...]}（与老服务一致）
    const resultData = resp.data?.data ?? [];
    syntaxRows.value = Array.isArray(resultData) ? resultData : [];
    
    // 检查 status 字段（与老服务一致）
    // status: 0表示语法检查通过，1表示语法检查不通过
    const status = resp.data?.status ?? 1; // 默认不通过
    syntaxStatus.value = status;
    
    if (status === 0) {
      message.success('语法检查通过，您可以提交工单了，O(∩_∩)O');
    } else {
      message.warning('语法检查未通过，请修复问题后重新检查');
    }
  } catch (e: any) {
    console.error('语法检查失败:', e);
    message.error(e?.message || '语法检查失败');
    syntaxStatus.value = null;
  } finally {
    checking.value = false;
  }
}

// 审核人验证已移除，由流程引擎自动分配

async function submitOrder() {
  loading.value = true;
  try {
    if (!formModel.title || formModel.title.length < 5) {
      message.error('请填写标题(不少于5个字符)');
      return;
    }
    if (!formModel.environment || !formModel.instanceId || !formModel.schema) {
      message.error('请完善环境/实例/库名');
      return;
    }
    // 审核人由流程引擎自动分配，无需验证
    if (!formModel.content) {
      message.error('提交的SQL内容不能为空');
      return;
    }

    if (syntaxStatus.value !== 0) {
      message.error('语法检测未通过不允许提交工单');
      return;
    }

    const payload = {
      title: formModel.title,
      remark: formModel.remark,
      is_restrict_access: formModel.isRestrictAccess,
      db_type: formModel.dbType,
      environment: formModel.environment,
      instance_id: formModel.instanceId,
      schema: formModel.schema,
      export_file_format: formModel.exportFileFormat,
      // 审核人、执行人、复核人、抄送人由流程引擎自动分配，无需传递
      approver: [],
      executor: [],
      reviewer: [],
      cc: [],
      sql_type: sqlType.value,
      content: formModel.content,
      schedule_time: formModel.scheduleTime // 后端已支持 "yyyy-MM-dd HH:mm:ss" 格式
    };

    const res: any = await fetchCreateOrder(payload as any);
    
    // createFlatRequest 返回 { data, error, response } 格式
    // 错误时: data=null, error=AxiosError
    // 成功时: data=后端响应的data字段（工单对象）, error=null, response=完整响应
    // 后端成功响应格式: { code: 0, message: "ok", data: {...} }
    // 注意：res.data 是后端响应的 data 字段，不是整个响应对象
    // 响应码在 res.response.data.code 中
    
    // 检查是否有错误
    if (res.error) {
      // 请求失败，错误消息已由全局拦截器显示
      message.warning('工单提交失败');
      return;
    }
    
    // 检查响应码（成功时为 0）
    // 响应码在 res.response.data.code 中，而不是 res.data.code
    const successCode = import.meta.env.VITE_SERVICE_SUCCESS_CODE || '0';
    const responseCode = String(res.response?.data?.code ?? '');
    
    if (responseCode === successCode) {
      message.success('工单提交成功');
      router.push('/das/orders-list');
    } else {
      // 如果响应码不匹配，显示后端返回的消息
      const errorMessage = res.response?.data?.message || '工单提交失败';
      message.warning(errorMessage);
    }
  } catch (e: any) {
    message.error(e?.message || '工单提交失败');
  } finally {
    loading.value = false;
  }
}

watch(
  () => route.path,
  () => inferSqlTypeFromPath()
);
</script>

<template>
  <div class="order-commit-page">
    <NCard :title="pageTitle" :content-style="{ padding: appStore.isMobile ? '8px' : '16px' }">
      <NGrid :x-gap="appStore.isMobile ? 8 : 16" :y-gap="appStore.isMobile ? 8 : 16" responsive="screen" style="align-items: stretch">
        <!-- 左侧表单 -->
        <NGi :span="appStore.isMobile ? 24 : 8">
          <NCard style="height: 100%">
            <div ref="leftContentRef">
              <NForm :label-placement="appStore.isMobile ? 'top' : 'left'" :label-width="appStore.isMobile ? 'auto' : 96">
                <NFormItem label="标题">
                  <NInput v-model:value="formModel.title" :size="appStore.isMobile ? 'small' : 'medium'" placeholder="请输入工单标题" />
                </NFormItem>
                <NFormItem label="备注">
                  <NInput
                    v-model:value="formModel.remark"
                    type="textarea"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :autosize="{ minRows: 2, maxRows: 6 }"
                    placeholder="请输入工单需求或备注"
                  />
                </NFormItem>
                <NFormItem label="限制访问">
                  <NSwitch v-model:value="formModel.isRestrictAccess" />
                </NFormItem>
                <NFormItem label="DB类型">
                  <NSelect
                    v-model:value="formModel.dbType"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :options="[
                      { label: 'MySQL', value: 'MySQL' },
                      { label: 'TiDB', value: 'TiDB' }
                    ]"
                    @update:value="onDBTypeChange"
                  />
                </NFormItem>
                <NFormItem label="环境">
                  <NSelect
                    v-model:value="formModel.environment"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :options="environments.map((e: any) => ({ label: e.name, value: e.ID }))"
                    filterable
                    clearable
                    placeholder="请选择工单环境"
                    @update:value="onEnvironmentChange"
                  />
                </NFormItem>
                <NFormItem label="实例">
                  <NSelect
                    v-model:value="formModel.instanceId"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :options="instances.map((i: any) => ({ label: i.remark, value: i.instance_id }))"
                    filterable
                    clearable
                    placeholder="请选择数据库实例"
                    @update:value="async (val) => {
                      await onInstanceChange(val);
                      // 实例变化时，如果已选择 schema，重新加载表结构
                      if (formModel.schema) {
                        await loadTablesForCompletion();
                      }
                    }"
                  />
                </NFormItem>
                <NFormItem label="库名">
                  <NSelect
                    v-model:value="formModel.schema"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :options="schemas.map((s: any) => ({ label: s.schema, value: s.schema }))"
                    filterable
                    clearable
                    placeholder="请选择数据库"
                    @update:value="async (val) => {
                      formModel.schema = val;
                      await loadTablesForCompletion();
                    }"
                  />
                </NFormItem>
                <NFormItem v-if="isExportOrder" label="文件格式">
                  <NSelect
                    v-model:value="formModel.exportFileFormat"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    :options="[
                      { label: 'XLSX', value: 'XLSX' },
                      { label: 'CSV', value: 'CSV' }
                    ]"
                  />
                </NFormItem>
                <NFormItem label="定时执行">
                  <NDatePicker
                    v-model:formatted-value="formModel.scheduleTime"
                    type="datetime"
                    :size="appStore.isMobile ? 'small' : 'medium'"
                    clearable
                    value-format="yyyy-MM-dd HH:mm:ss"
                    placeholder="请选择计划执行时间(可选)"
                    style="width: 100%"
                  />
                </NFormItem>
                <!-- 审核人、执行人、复核人、抄送人已通过流程引擎配置，无需手动选择 -->
                <NFormItem>
                  <NButton 
                    type="primary" 
                    :size="appStore.isMobile ? 'small' : 'medium'" 
                    :loading="loading" 
                    :disabled="isSubmitDisabled"
                    :block="appStore.isMobile" 
                    @click="submitOrder"
                  >
                    提交
                  </NButton>
                </NFormItem>
              </NForm>
            </div>
          </NCard>
        </NGi>
        <!-- 右侧编辑区域 -->
        <NGi :span="appStore.isMobile ? 24 : 16">
          <NCard class="editor-card" :style="{ height: appStore.isMobile ? 'auto' : '100%' }">
            <div class="editor-inner" :style="{ height: appStore.isMobile ? '400px' : (leftContentHeight > 0 ? leftContentHeight + 'px' : '500px') }">
              <NAlert type="info" title="说明" :closable="!appStore.isMobile" :size="appStore.isMobile ? 'small' : 'medium'">支持多条SQL语句，每条SQL须以 ; 结尾</NAlert>
              <div style="margin: 8px 0">
                <NSpace :size="appStore.isMobile ? 6 : 12" :wrap="appStore.isMobile" align="center">
                  <NButton :size="appStore.isMobile ? 'tiny' : 'small'" type="primary" secondary @click="formatSQL">
                    <template v-if="appStore.isMobile" #icon>
                      <div class="i-ant-design:format-painter-outlined" />
                    </template>
                    <span v-if="!appStore.isMobile">格式化</span>
                  </NButton>
                  <NButton :size="appStore.isMobile ? 'tiny' : 'small'" type="primary" secondary :loading="checking" @click="syntaxCheck">
                    <template v-if="appStore.isMobile" #icon>
                      <div class="i-ant-design:check-circle-outlined" />
                    </template>
                    <span v-if="!appStore.isMobile">语法检查</span>
                  </NButton>
                </NSpace>
              </div>
              <!-- 替换 textarea 为与 SQL 查询一致的 CodeMirror 编辑器 -->
              <div ref="editorRoot" class="code-editor-container" />
            </div>
          </NCard>
        </NGi>
      </NGrid>
    </NCard>
    <NCard v-if="syntaxRows.length" title="语法检查结果" :style="{ marginTop: appStore.isMobile ? '8px' : '12px' }" :content-style="{ padding: appStore.isMobile ? '8px' : '16px' }">
      <NDataTable
        :columns="visibleSyntaxColumns"
        :data="syntaxRows"
        :pagination="pagination"
        size="small"
        single-line
        table-layout="fixed"
        :scroll-x="appStore.isMobile ? 800 : 1200"
      />
    </NCard>
  </div>
</template>

<style scoped>
.order-commit-page {
  padding: 0;
}

:deep(.n-card .n-card__content) {
  padding: 12px;
}

/* 参考 SQL 查询页的编辑器样式 */
.editor-card :deep(.n-card__content) {
  /* 右侧卡片内容作为外层容器，不再直接拉伸 */
  display: flex;
  flex-direction: column;
}
.editor-inner {
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 限制整体高度，内部滚动 */
}
.code-editor-container {
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background-color: var(--n-color);
  margin-bottom: 8px;
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0; /* 允许内部滚动 */
}
.code-editor-container :deep(.cm-editor) {
  background-color: transparent;
  font-family: 'JetBrains Mono', 'Fira Code', Menlo, Monaco, 'Courier New', monospace;
  font-size: 13px;
  height: 100%; /* 填满容器以启用滚动 */
}
.code-editor-container :deep(.cm-scroller) {
  height: 100%;
  padding: 4px;
  overflow: auto; /* 内容超出时滚动 */
}
.code-editor-container :deep(.cm-gutters) {
  background-color: var(--n-color);
  border-right: 1px solid var(--n-border-color);
}
.code-editor-container :deep(.cm-activeLine) {
  background-color: rgba(0, 0, 0, 0.03);
}

/* 移动端适配 */
@media (max-width: 640px) {
  .order-commit-page {
    padding: 0;
  }

  :deep(.n-card .n-card__content) {
    padding: 8px;
  }

  .editor-inner {
    min-height: 300px;
  }

  .code-editor-container {
    min-height: 300px;
    font-size: 12px;
  }

  .code-editor-container :deep(.cm-editor) {
    font-size: 12px !important;
  }

  /* 表单优化 */
  :deep(.n-form-item) {
    margin-bottom: 16px;
  }

  :deep(.n-form-item-label) {
    font-size: 13px;
    margin-bottom: 4px;
  }

  /* 按钮优化 */
  :deep(.n-button) {
    font-size: 13px;
  }

  /* 选择框优化 */
  :deep(.n-select) {
    font-size: 13px;
  }

  /* 输入框优化 */
  :deep(.n-input) {
    font-size: 13px;
  }

  /* 表格优化 */
  :deep(.n-data-table) {
    font-size: 12px;
  }

  /* 卡片标题优化 */
  :deep(.n-card-header) {
    padding: 12px;
    font-size: 16px;
  }

  /* 警告框优化 */
  :deep(.n-alert) {
    font-size: 12px;
    padding: 8px;
  }
}
</style>
