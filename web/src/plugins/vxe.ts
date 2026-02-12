import type { App } from 'vue';

let installed = false;

/**
 * 按需加载 VxeTable + VxePC-UI，仅在进入 DAS 编辑页等使用表格时加载，减轻首屏体积
 */
export async function ensureVxe(app: App) {
  if (installed) return;

  const [VxeUIModule, VXETableModule] = await Promise.all([
    import('vxe-pc-ui'),
    import('vxe-table')
  ]);

  await Promise.all([
    import('vxe-table/lib/style.css'),
    import('vxe-pc-ui/lib/style.css')
  ]);

  app.use(VxeUIModule.default);
  app.use(VXETableModule.default);
  installed = true;
}
