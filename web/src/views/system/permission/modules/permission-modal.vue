<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  fetchGetMenuList,
  fetchGetApis,
  fetchGetRolePermissions,
  fetchUpdateRolePermission
} from '@/service/api/admin';
import { $t } from '@/locales';

defineOptions({
  name: 'PermissionModal'
});

interface Props {
  /** the role data */
  role?: Api.Admin.Role | null;
}

const props = defineProps<Props>();

const visible = defineModel<boolean>('visible', {
  default: false
});

function closeModal() {
  visible.value = false;
}

const title = computed(() => `${$t('page.manage.role.assignPermission')} - ${props.role?.name || ''}`);

const loading = ref(false);

// 菜单权限
const menuTree = ref<any[]>([]);
const menuCheckedKeys = ref<string[]>([]);

// API权限
const apiTree = ref<any[]>([]);
const apiCheckedKeys = ref<string[]>([]);

// 加载菜单列表
async function loadMenus() {
  try {
    const { data } = await fetchGetMenuList();
    if (data && data.records) {
      menuTree.value = buildMenuTree(data.records);
    }
  } catch (error) {
    console.error('Failed to load menus:', error);
  }
}

// 构建菜单树
function buildMenuTree(menus: any[]): any[] {
  const map = new Map<number, any>();
  const tree: any[] = [];

  menus.forEach(menu => {
    map.set(menu.id, {
      key: `menu:${menu.routePath || menu.path},read`,
      label: menu.menuName || menu.title,
      children: []
    });
  });

  menus.forEach(menu => {
    const node = map.get(menu.id);
    if (menu.parentId && map.has(menu.parentId)) {
      map.get(menu.parentId).children.push(node);
    } else {
      tree.push(node);
    }
  });

  // 移除空的children
  const removeEmptyChildren = (nodes: any[]) => {
    nodes.forEach(node => {
      if (node.children && node.children.length === 0) {
        delete node.children;
      } else if (node.children) {
        removeEmptyChildren(node.children);
      }
    });
  };
  removeEmptyChildren(tree);

  return tree;
}

// 加载API列表
async function loadApis() {
  try {
    const { data } = await fetchGetApis({ page: 1, pageSize: 1000 });
    if (data && data.list) {
      apiTree.value = buildApiTree(data.list);
    }
  } catch (error) {
    console.error('Failed to load apis:', error);
  }
}

// 按分组构建API树
function buildApiTree(apis: any[]): any[] {
  const groupMap = new Map<string, any[]>();

  apis.forEach(api => {
    const group = api.group || '其他';
    if (!groupMap.has(group)) {
      groupMap.set(group, []);
    }
    groupMap.get(group)!.push({
      key: `api:${api.path},${api.method}`,
      label: `${api.name} [${api.method}] ${api.path}`
    });
  });

  const tree: any[] = [];
  groupMap.forEach((children, group) => {
    tree.push({
      key: `group:${group}`,
      label: group,
      children
    });
  });

  return tree;
}

// 加载角色权限
async function loadRolePermissions() {
  if (!props.role?.sid) return;

  try {
    const { data } = await fetchGetRolePermissions(props.role.sid);
    if (data && data.list) {
      // 分离菜单权限和API权限
      const menuPerms: string[] = [];
      const apiPerms: string[] = [];

      data.list.forEach((perm: string) => {
        if (perm.startsWith('menu:')) {
          menuPerms.push(perm);
        } else if (perm.startsWith('api:')) {
          apiPerms.push(perm);
        }
      });

      menuCheckedKeys.value = menuPerms;
      apiCheckedKeys.value = apiPerms;
    }
  } catch (error) {
    console.error('Failed to load role permissions:', error);
  }
}

async function handleSubmit() {
  if (!props.role?.sid) return;

  loading.value = true;
  try {
    const permissions = [...menuCheckedKeys.value, ...apiCheckedKeys.value];
    await fetchUpdateRolePermission({
      role: props.role.sid,
      list: permissions
    });
    window.$message?.success($t('common.updateSuccess'));
    closeModal();
  } catch (error) {
    window.$message?.error($t('common.operationFailed') || '操作失败');
  } finally {
    loading.value = false;
  }
}

async function init() {
  loading.value = true;
  await Promise.all([loadMenus(), loadApis(), loadRolePermissions()]);
  loading.value = false;
}

watch(visible, val => {
  if (val) {
    init();
  } else {
    menuCheckedKeys.value = [];
    apiCheckedKeys.value = [];
  }
});
</script>

<template>
  <NModal v-model:show="visible" :title="title" preset="card" class="w-800px">
    <NSpin :show="loading">
      <NTabs type="line" animated>
        <NTabPane name="menu" :tab="$t('page.manage.role.menuPermission')">
          <div class="h-400px overflow-auto">
            <NTree
              v-model:checked-keys="menuCheckedKeys"
              :data="menuTree"
              checkable
              cascade
              expand-on-click
              selectable
              block-line
            />
          </div>
        </NTabPane>
        <NTabPane name="api" :tab="$t('page.manage.role.apiPermission')">
          <div class="h-400px overflow-auto">
            <NTree
              v-model:checked-keys="apiCheckedKeys"
              :data="apiTree"
              checkable
              cascade
              expand-on-click
              selectable
              block-line
            />
          </div>
        </NTabPane>
      </NTabs>
    </NSpin>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="closeModal">{{ $t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped></style>
