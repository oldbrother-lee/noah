<script setup lang="tsx">
import { ref } from 'vue';
import type { Ref } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { useBoolean } from '@sa/hooks';
import { yesOrNoRecord } from '@/constants/common';
import { enableStatusRecord, menuTypeRecord } from '@/constants/business';
import { fetchGetMenuList, fetchDeleteMenu } from '@/service/api/admin';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import SvgIcon from '@/components/custom/svg-icon.vue';
import MenuOperateModal, { type OperateType } from './modules/menu-operate-modal.vue';

const appStore = useAppStore();

const { bool: visible, setTrue: openModal } = useBoolean();

const wrapperRef = ref<HTMLElement | null>(null);

// 将扁平菜单数据转换为树形结构
function buildMenuTree(menus: any[]): (Api.SystemManage.Menu & { children?: Api.SystemManage.Menu[] })[] {
  const map = new Map<number, Api.SystemManage.Menu & { children?: Api.SystemManage.Menu[] }>();
  const tree: (Api.SystemManage.Menu & { children?: Api.SystemManage.Menu[] })[] = [];

  // 第一遍：创建所有节点的映射，统一处理 ID 字段
  menus.forEach((menu: any) => {
    const menuId = menu.ID || menu.id;
    const normalizedMenu = {
      ...menu,
      id: menuId,
      parentId: menu.parentId || menu.parent_id || 0
    };
    map.set(menuId, {
      ...normalizedMenu,
      children: []
    });
  });

  // 第二遍：构建树形结构
  menus.forEach((menu: any) => {
    const menuId = menu.ID || menu.id;
    const node = map.get(menuId)!;
    const parentId = menu.parentId || menu.parent_id || 0;
    
    if (parentId && map.has(parentId)) {
      const parent = map.get(parentId)!;
      if (!parent.children) {
        parent.children = [];
      }
      parent.children.push(node);
    } else {
      tree.push(node);
    }
  });

  // 按 order 排序
  const sortByOrder = (nodes: (Api.SystemManage.Menu & { children?: Api.SystemManage.Menu[] })[]) => {
    nodes.sort((a, b) => (a.order || 0) - (b.order || 0));
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        sortByOrder(node.children);
      }
    });
  };
  sortByOrder(tree);

  return tree;
}

// 使用useTable hook
const { columns, columnChecks, data, loading, pagination, getData, getDataByPage } = useTable({
  apiFn: fetchGetMenuList,
  transformer: res => {
    const { data: responseData } = res;
    if (responseData && responseData.records) {
      const { records = [] } = responseData;
      // 转换为树形结构
      const treeData = buildMenuTree(records);
      return {
        data: treeData,
        pageNum: 1,
        pageSize: treeData.length,
        total: treeData.length
      };
    }
    return {
      data: [],
      pageNum: 1,
      pageSize: 0,
      total: 0
    };
  },
  columns: () => [
    {
      type: 'selection',
      align: 'center',
      width: 48
    },
    {
      key: 'id',
      title: $t('page.manage.menu.id'),
      align: 'center',
      width: 100,
      render: (row: Api.SystemManage.Menu) => {
        const hasChildren = row.children && row.children.length > 0;
        // 这里不显示展开/折叠图标，因为 NDataTable 会自动处理
        return <span>{row.id}</span>;
      }
    },
    {
      key: 'menuType',
      title: $t('page.manage.menu.menuType'),
      align: 'center',
      width: 80,
      render: (row: Api.SystemManage.Menu) => {
        const tagMap: Record<Api.SystemManage.MenuType, NaiveUI.ThemeColor> = {
          '1': 'default',
          '2': 'primary'
        };
        const label = $t(menuTypeRecord[row.menuType]);
        return <NTag type={tagMap[row.menuType]}>{label}</NTag>;
      }
    },
    {
      key: 'menuName',
      title: $t('page.manage.menu.menuName'),
      align: 'center',
      minWidth: 120,
      render: (row: Api.SystemManage.Menu) => {
        const { i18nKey, menuName } = row;
        const label = i18nKey ? $t(i18nKey) : menuName;
        return <span>{label}</span>;
      }
    },
    {
      key: 'icon',
      title: $t('page.manage.menu.icon'),
      align: 'center',
      width: 60,
      render: (row: Api.SystemManage.Menu) => {
        const icon = row.iconType === '1' ? row.icon : undefined;
        const localIcon = row.iconType === '2' ? row.icon : undefined;
        return (
          <div class="flex-center">
            <SvgIcon icon={icon} localIcon={localIcon} class="text-icon" />
          </div>
        );
      }
    },
    {
      key: 'routeName',
      title: $t('page.manage.menu.routeName'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'routePath',
      title: $t('page.manage.menu.routePath'),
      align: 'center',
      minWidth: 120
    },
    {
      key: 'status',
      title: $t('page.manage.menu.menuStatus'),
      align: 'center',
      width: 80,
      render: (row: Api.SystemManage.Menu) => {
        if (row.status === null) {
          return null;
        }
        const tagMap: Record<Api.Common.EnableStatus, NaiveUI.ThemeColor> = {
          '1': 'success',
          '2': 'warning'
        };
        const label = $t(enableStatusRecord[row.status]);
        return <NTag type={tagMap[row.status]}>{label}</NTag>;
      }
    },
    {
      key: 'hideInMenu',
      title: $t('page.manage.menu.hideInMenu'),
      align: 'center',
      width: 80,
      render: (row: Api.SystemManage.Menu) => {
        const hide: CommonType.YesOrNo = row.hideInMenu ? 'Y' : 'N';
        const tagMap: Record<CommonType.YesOrNo, NaiveUI.ThemeColor> = {
          Y: 'error',
          N: 'default'
        };
        const label = $t(yesOrNoRecord[hide]);
        return <NTag type={tagMap[hide]}>{label}</NTag>;
      }
    },
    {
      key: 'parentId',
      title: $t('page.manage.menu.parentId'),
      width: 90,
      align: 'center'
    },
    {
      key: 'order',
      title: $t('page.manage.menu.order'),
      align: 'center',
      width: 60
    },
    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 230,
      render: (row: Api.SystemManage.Menu) => (
        <div class="flex-center justify-end gap-8px">
          {row.menuType === '1' && (
            <NButton type="primary" ghost size="small" onClick={() => handleAddChildMenu(row)}>
              {$t('page.manage.menu.addChildMenu')}
            </NButton>
          )}
          <NButton type="primary" ghost size="small" onClick={() => handleEdit(row)}>
            {$t('common.edit')}
          </NButton>
          <NPopconfirm onPositiveClick={() => handleDelete(row.id)}>
            {{
              default: () => $t('common.confirmDelete'),
              trigger: () => (
                <NButton type="error" ghost size="small">
                  {$t('common.delete')}
                </NButton>
              )
            }}
          </NPopconfirm>
        </div>
      )
    }
  ]
});

const checkedRowKeys = ref<(string | number)[]>([]);

async function onDeleted() {
  await getData();
}

async function onBatchDeleted() {
  await getData();
}

const operateType = ref<OperateType>('add');

function handleAdd() {
  operateType.value = 'add';
  openModal();
}

async function handleBatchDelete() {
  // TODO: 实现批量删除
  console.log(checkedRowKeys.value);
  onBatchDeleted();
}

async function handleDelete(id: number) {
  try {
    await fetchDeleteMenu(id);
    window.$message?.success($t('common.deleteSuccess'));
    await getData();
  } catch (error) {
    window.$message?.error($t('common.deleteFailed') || '删除失败');
  }
}

/** the edit menu data or the parent menu data when adding a child menu */
const editingData: Ref<Api.SystemManage.Menu | null> = ref(null);

function handleEdit(item: Api.SystemManage.Menu) {
  operateType.value = 'edit';
  editingData.value = { ...item };
  openModal();
}

function handleAddChildMenu(item: Api.SystemManage.Menu) {
  operateType.value = 'addChild';
  editingData.value = { ...item };
  openModal();
}

// 不再调用不存在的 pages 接口
const allPages = ref<string[]>([]);
</script>

<template>
  <div ref="wrapperRef" class="flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard :title="$t('page.manage.menu.title')" :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          @add="handleAdd"
          @delete="handleBatchDelete"
          @refresh="getData"
        />
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1088"
        :loading="loading"
        :row-key="row => row.id"
        :default-expand-all="false"
        :pagination="false"
        class="sm:h-full"
      />
      <MenuOperateModal
        v-model:visible="visible"
        :operate-type="operateType"
        :row-data="editingData"
        :all-pages="allPages"
        @submitted="getData"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
