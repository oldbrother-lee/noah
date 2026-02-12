// Shim type declarations for third-party modules without bundled types in this project
declare module 'vxe-table' {
  import type { App } from 'vue';
  const VXETable: { install(app: App): void };
  export default VXETable;
  export const VxeTable: any;
  export const VxeColumn: any;
  export const VxePager: any;
}

declare module 'xe-utils';
