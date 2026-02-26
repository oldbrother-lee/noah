const local: App.I18n.Schema = {
  system: {
    title: 'Noah 管理系统',
    updateTitle: 'System Version Update Notification',
    updateContent: 'A new version of the system has been detected. Do you want to refresh the page immediately?',
    updateConfirm: 'Refresh immediately',
    updateCancel: 'Later'
  },
  common: {
    action: 'Action',
    add: 'Add',
    addSuccess: 'Add Success',
    backToHome: 'Back to home',
    batchDelete: 'Batch Delete',
    cancel: 'Cancel',
    close: 'Close',
    check: 'Check',
    expandColumn: 'Expand Column',
    columnSetting: 'Column Setting',
    exportResult: 'Export',
    exportCsv: 'Export CSV',
    exportSql: 'Export SQL',
    exportTxt: 'Export TXT',
    config: 'Config',
    confirm: 'Confirm',
    delete: 'Delete',
    deleteSuccess: 'Delete Success',
    confirmDelete: 'Are you sure you want to delete?',
    edit: 'Edit',
    warning: 'Warning',
    error: 'Error',
    index: 'Index',
    keywordSearch: 'Please enter keyword',
    logout: 'Logout',
    logoutConfirm: 'Are you sure you want to log out?',
    lookForward: 'Coming soon',
    modify: 'Modify',
    modifySuccess: 'Modify Success',
    noData: 'No Data',
    operate: 'Operate',
    pleaseCheckValue: 'Please check whether the value is valid',
    refresh: 'Refresh',
    reset: 'Reset',
    search: 'Search',
    switch: 'Switch',
    tip: 'Tip',
    trigger: 'Trigger',
    update: 'Update',
    updateSuccess: 'Update Success',
    userCenter: 'User Center',
    yesOrNo: {
      yes: 'Yes',
      no: 'No'
    }
  },
  request: {
    logout: 'Logout user after request failed',
    logoutMsg: 'User status is invalid, please log in again',
    logoutWithModal: 'Pop up modal after request failed and then log out user',
    logoutWithModalMsg: 'User status is invalid, please log in again',
    refreshToken: 'The requested token has expired, refresh the token',
    tokenExpired: 'The requested token has expired'
  },
  theme: {
    themeSchema: {
      title: 'Theme Schema',
      light: 'Light',
      dark: 'Dark',
      auto: 'Follow System'
    },
    grayscale: 'Grayscale',
    colourWeakness: 'Colour Weakness',
    layoutMode: {
      title: 'Layout Mode',
      vertical: 'Vertical Menu Mode',
      horizontal: 'Horizontal Menu Mode',
      'vertical-mix': 'Vertical Mix Menu Mode',
      'horizontal-mix': 'Horizontal Mix menu Mode',
      reverseHorizontalMix: 'Reverse first level menus and child level menus position'
    },
    recommendColor: 'Apply Recommended Color Algorithm',
    recommendColorDesc: 'The recommended color algorithm refers to',
    themeColor: {
      title: 'Theme Color',
      primary: 'Primary',
      info: 'Info',
      success: 'Success',
      warning: 'Warning',
      error: 'Error',
      followPrimary: 'Follow Primary'
    },
    scrollMode: {
      title: 'Scroll Mode',
      wrapper: 'Wrapper',
      content: 'Content'
    },
    page: {
      animate: 'Page Animate',
      mode: {
        title: 'Page Animate Mode',
        fade: 'Fade',
        'fade-slide': 'Slide',
        'fade-bottom': 'Fade Zoom',
        'fade-scale': 'Fade Scale',
        'zoom-fade': 'Zoom Fade',
        'zoom-out': 'Zoom Out',
        none: 'None'
      }
    },
    fixedHeaderAndTab: 'Fixed Header And Tab',
    header: {
      height: 'Header Height',
      breadcrumb: {
        visible: 'Breadcrumb Visible',
        showIcon: 'Breadcrumb Icon Visible'
      },
      multilingual: {
        visible: 'Display multilingual button'
      },
      globalSearch: {
        visible: 'Display GlobalSearch button'
      }
    },
    tab: {
      visible: 'Tab Visible',
      cache: 'Tag Bar Info Cache',
      height: 'Tab Height',
      mode: {
        title: 'Tab Mode',
        chrome: 'Chrome',
        button: 'Button'
      }
    },
    sider: {
      inverted: 'Dark Sider',
      width: 'Sider Width',
      collapsedWidth: 'Sider Collapsed Width',
      mixWidth: 'Mix Sider Width',
      mixCollapsedWidth: 'Mix Sider Collapse Width',
      mixChildMenuWidth: 'Mix Child Menu Width'
    },
    footer: {
      visible: 'Footer Visible',
      fixed: 'Fixed Footer',
      height: 'Footer Height',
      right: 'Right Footer'
    },
    watermark: {
      visible: 'Watermark Full Screen Visible',
      text: 'Watermark Text',
      enableUserName: 'Enable User Name Watermark'
    },
    themeDrawerTitle: 'Theme Configuration',
    pageFunTitle: 'Page Function',
    resetCacheStrategy: {
      title: 'Reset Cache Strategy',
      close: 'Close Page',
      refresh: 'Refresh Page'
    },
    configOperation: {
      copyConfig: 'Copy Config',
      copySuccessMsg: 'Copy Success, Please replace the variable "themeSettings" in "src/theme/settings.ts"',
      resetConfig: 'Reset Config',
      resetSuccessMsg: 'Reset Success'
    }
  },
  route: {
    login: 'Login',
    403: 'No Permission',
    404: 'Page Not Found',
    500: 'Server Error',
    'iframe-page': 'Iframe',
    home: 'Home',
    das: 'Data Access Service',
    das_modules: 'DAS Modules',
    das_modules_edit: 'SQL Editor',
    das_modules_favorite: 'Favorite SQL',
    das_modules_history: 'Query History',
    sql: 'SQL Query',
    das_favorite: 'Favorite SQL',
    das_history: 'Query History',
    das_edit: 'SQL Query',
    'das_orders-list': 'Order List',
    'das_orders-detail': 'Order Detail',
    das_orders: 'Orders',
    das_orders_ddl: 'DDL Order',
    das_orders_dml: 'DML Order',
    das_orders_export: 'Export Order',
    das_orders_commit: 'Commit Order',
    das_orders_commit_ddl: 'DDL Order',
    das_orders_commit_dml: 'DML Order',
    das_orders_commit_export: 'Export Order',
    system: 'System Management',
    system_menu: 'Menu Management',
    system_user: 'User Management',
    system_permission: 'Permission Management',
    system_database: 'Database Management',
    system_database_config: 'Database Configuration',
    system_database_environment: 'Environment Management',
    system_database_inspect: 'Inspect Parameters',
    system_database_permission: 'Permission Management',
    system_database_permission_assign: 'Permission Assignment',
    system_database_permission_template: 'Permission Template'
  },
  page: {
    login: {
      common: {
        loginOrRegister: 'Login / Register',
        userNamePlaceholder: 'Please enter user name',
        phonePlaceholder: 'Please enter phone number',
        codePlaceholder: 'Please enter verification code',
        passwordPlaceholder: 'Please enter password',
        confirmPasswordPlaceholder: 'Please enter password again',
        codeLogin: 'Verification code login',
        confirm: 'Confirm',
        back: 'Back',
        validateSuccess: 'Verification passed',
        loginSuccess: 'Login successfully',
        welcomeBack: 'Welcome back, {userName} !'
      },
      pwdLogin: {
        title: 'Password Login',
        rememberMe: 'Remember me',
        forgetPassword: 'Forget password?',
        register: 'Register',
        otherAccountLogin: 'Other Account Login',
        otherLoginMode: 'Other Login Mode',
        superAdmin: 'Super Admin',
        admin: 'Admin',
        user: 'User'
      },
      codeLogin: {
        title: 'Verification Code Login',
        getCode: 'Get verification code',
        reGetCode: 'Reacquire after {time}s',
        sendCodeSuccess: 'Verification code sent successfully',
        imageCodePlaceholder: 'Please enter image verification code'
      },
      register: {
        title: 'Register',
        agreement: 'I have read and agree to',
        protocol: '《User Agreement》',
        policy: '《Privacy Policy》'
      },
      resetPwd: {
        title: 'Reset Password'
      },
      bindWeChat: {
        title: 'Bind WeChat'
      }
    },
    home: {
      branchDesc:
        'For the convenience of everyone in developing and updating the merge, we have streamlined the code of the main branch, only retaining the homepage menu, and the rest of the content has been moved to the example branch for maintenance. The preview address displays the content of the example branch.',
      greeting: 'Good morning, {userName}, today is another day full of vitality!',
      weatherDesc: 'Today is cloudy to clear, 20℃ - 25℃!',
      projectCount: 'Project Count',
      todo: 'Todo',
      message: 'Message',
      downloadCount: 'Download Count',
      registerCount: 'Register Count',
      schedule: 'Work and rest Schedule',
      study: 'Study',
      work: 'Work',
      rest: 'Rest',
      entertainment: 'Entertainment',
      visitCount: 'Visit Count',
      turnover: 'Turnover',
      dealCount: 'Deal Count',
      projectNews: {
        title: 'Project News',
        moreNews: 'More News',
        desc1: 'Soybean created the open source project soybean-admin on May 28, 2021!',
        desc2: 'Yanbowe submitted a bug to soybean-admin, the multi-tab bar will not adapt.',
        desc3: 'Soybean is ready to do sufficient preparation for the release of soybean-admin!',
        desc4: 'Soybean is busy writing project documentation for soybean-admin!',
        desc5: 'Soybean just wrote some of the workbench pages casually, and it was enough to see!'
      },
      creativity: 'Creativity'
    },
    manage: {
      common: {
        status: {
          enable: 'Enable',
          disable: 'Disable'
        }
      },
      menu: {
        home: 'Home',
        title: 'Menu List',
        id: 'ID',
        parentId: 'Parent Menu ID',
        menuType: 'Menu Type',
        menuName: 'Menu Name',
        routeName: 'Route Name',
        routePath: 'Route Path',
        pathParam: 'Path Parameter',
        layout: 'Layout',
        page: 'Page Component',
        i18nKey: 'i18n Key',
        icon: 'Icon',
        localIcon: 'Local Icon',
        iconTypeTitle: 'Icon Type',
        order: 'Order',
        constant: 'Constant Route',
        keepAlive: 'Keep Alive',
        href: 'External Link',
        hideInMenu: 'Hide in Menu',
        activeMenu: 'Active Menu',
        multiTab: 'Multi Tab',
        fixedIndexInTab: 'Fixed Index in Tab',
        query: 'Route Query',
        button: 'Button',
        buttonCode: 'Button Code',
        buttonDesc: 'Button Description',
        menuStatus: 'Menu Status',
        addMenu: 'Add Menu',
        addChildMenu: 'Add Child Menu',
        editMenu: 'Edit Menu',
        type: {
          directory: 'Directory',
          menu: 'Menu'
        },
        iconType: {
          iconify: 'Iconify Icon',
          local: 'Local Icon'
        },
        form: {
          home: 'Please select home page',
          menuType: 'Please select menu type',
          menuName: 'Please enter menu name',
          routeName: 'Please enter route name',
          routePath: 'Please enter route path',
          pathParam: 'Please enter path parameter',
          page: 'Please select page component',
          layout: 'Please select layout component',
          i18nKey: 'Please enter i18n key',
          icon: 'Please enter icon',
          localIcon: 'Please select local icon',
          order: 'Please enter order',
          keepAlive: 'Please select whether to keep alive',
          href: 'Please enter external link',
          hideInMenu: 'Please select whether to hide in menu',
          activeMenu: 'Please select active menu route name',
          multiTab: 'Please select whether to support multi tab',
          fixedInTab: 'Please select whether to fix in tab',
          fixedIndexInTab: 'Please enter fixed index in tab',
          queryKey: 'Please enter route query key',
          queryValue: 'Please enter route query value',
          buttonCode: 'Please enter button code',
          buttonDesc: 'Please enter button description'
        }
      },
      user: {
        title: 'User List',
        username: 'Username',
        nickname: 'Nickname',
        email: 'Email',
        phone: 'Phone',
        password: 'Password',
        roles: 'Roles',
        createdAt: 'Created At',
        addUser: 'Add User',
        editUser: 'Edit User',
        form: {
          username: 'Please enter username',
          nickname: 'Please enter nickname',
          password: 'Please enter password',
          passwordPlaceholder: 'Leave blank to not change password',
          passwordRequired: 'Password cannot be empty',
          email: 'Please enter email',
          phone: 'Please enter phone',
          roles: 'Please select roles'
        }
      },
      role: {
        title: 'Role Management',
        roleName: 'Role Name',
        roleCode: 'Role Code',
        roleDesc: 'Role Description',
        roleStatus: 'Role Status',
        createdAt: 'Created At',
        addRole: 'Add Role',
        editRole: 'Edit Role',
        assignPermission: 'Assign Permission',
        menuPermission: 'Menu Permission',
        apiPermission: 'API Permission',
        buttonAuth: 'Button Permission',
        menuAuth: 'Menu Permission',
        form: {
          roleName: 'Please enter role name',
          roleCode: 'Please enter role code',
          roleDesc: 'Please enter role description',
          roleStatus: 'Please select role status'
        }
      },
      api: {
        title: 'API Management',
        group: 'Group',
        name: 'Name',
        path: 'Path',
        method: 'Method',
        createdAt: 'Created At',
        addApi: 'Add API',
        editApi: 'Edit API',
        syncRoute: 'Sync Route',
        confirmSync: 'Confirm Sync',
        freshCasbin: 'Refresh Casbin',
        batchDelete: 'Batch Delete',
        syncApiTitle: 'Sync Route',
        syncApiTip: 'Compare code routes with DB. Add or remove APIs; use Ignore for routes that do not need auth.',
        newApis: 'New Routes',
        newApisTip: 'In current routes but not in API table',
        deleteApis: 'Routes to Remove',
        deleteApisTip: 'In API table but no longer in current routes',
        ignoreApis: 'Ignored',
        ignoreApisTip: 'Not synced and not checked for auth',
        singleAdd: 'Add One',
        singleDelete: 'Remove One',
        groupPlaceholder: 'Select or add new',
        nameRequired: 'Please enter API name or use AI Auto Fill',
        aiAutoFill: 'AI Auto Fill',
        aiAutoFillComing: 'AI Auto Fill coming soon',
        aiAutoFillSuccess: 'Name and group filled from route suggestions',
        noNewApisToFill: 'No new routes to fill',
        newGroup: 'New Group',
        newGroupPlaceholder: 'Enter new group name',
        form: {
          group: 'Please enter group',
          name: 'Please enter name',
          path: 'Please enter path',
          method: 'Please select method'
        }
      },
      database: {
        title: 'Database Management',
        environment: {
          title: 'Environment Management',
          name: 'Environment Name',
          createdAt: 'Created At',
          updatedAt: 'Updated At',
          addEnvironment: 'Add Environment',
          editEnvironment: 'Edit Environment',
          form: {
            name: 'Please enter environment name',
            nameRequired: 'Environment name cannot be empty'
          }
        },
        config: {
          title: 'Database Configuration',
          instanceId: 'Instance ID',
          hostname: 'Hostname',
          port: 'Port',
          userName: 'Username',
          password: 'Password',
          useType: 'Use Type',
          dbType: 'Database Type',
          environment: 'Environment',
          organizationKey: 'Organization',
          remark: 'Remark',
          createdAt: 'Created At',
          updatedAt: 'Updated At',
          addConfig: 'Add Database Config',
          editConfig: 'Edit Database Config',
          form: {
            hostname: 'Please enter hostname',
            port: 'Please enter port',
            userName: 'Please enter username',
            password: 'Please enter password',
            useType: 'Please select use type',
            dbType: 'Please select database type',
            environment: 'Please select environment',
            organizationKey: 'Please enter organization',
            remark: 'Please enter remark'
          }
        },
        permission: {
          title: 'Database Permission',
          username: 'Username',
          instanceId: 'Instance ID',
          schema: 'Schema',
          createdAt: 'Created At',
          updatedAt: 'Updated At',
          addPermission: 'Add Permission',
          selectUser: 'Select User',
          selectUserFirst: 'Please select a user first',
          pleaseSelectUser: 'Please select a user',
          form: {
            instanceId: 'Please select instance',
            schema: 'Please select schema'
          }
        },
        permissionTemplate: {
          title: 'Permission Template',
          name: 'Template Name',
          description: 'Template Description',
          permissions: 'Permission Count',
          createdAt: 'Created At',
          updatedAt: 'Updated At',
          addTemplate: 'Add Template',
          editTemplate: 'Edit Template',
          form: {
            name: 'Please enter template name',
            description: 'Please enter template description'
          }
        },
        permissionAssign: {
          title: 'Permission Assignment',
          roleTab: 'Role Permission',
          userTab: 'User Permission'
        },
        rolePermission: {
          title: 'Role Permission',
          role: 'Role',
          permissionType: 'Permission Type',
          permissionId: 'Permission ID',
          instanceId: 'Instance ID',
          schema: 'Schema',
          table: 'Table',
          createdAt: 'Created At',
          form: {
            role: 'Please select role',
            permissionType: 'Please select permission type',
            permissionId: 'Please select template',
            instanceId: 'Please select instance',
            schema: 'Please select schema'
          }
        }
      }
    }
  },
  form: {
    required: 'Cannot be empty',
    userName: {
      required: 'Please enter user name',
      invalid: 'User name format is incorrect'
    },
    phone: {
      required: 'Please enter phone number',
      invalid: 'Phone number format is incorrect'
    },
    pwd: {
      required: 'Please enter password',
      invalid: '6-18 characters, including letters, numbers, and underscores'
    },
    confirmPwd: {
      required: 'Please enter password again',
      invalid: 'The two passwords are inconsistent'
    },
    code: {
      required: 'Please enter verification code',
      invalid: 'Verification code format is incorrect'
    },
    email: {
      required: 'Please enter email',
      invalid: 'Email format is incorrect'
    }
  },
  dropdown: {
    closeCurrent: 'Close Current',
    closeOther: 'Close Other',
    closeLeft: 'Close Left',
    closeRight: 'Close Right',
    closeAll: 'Close All'
  },
  icon: {
    themeConfig: 'Theme Configuration',
    themeSchema: 'Theme Schema',
    lang: 'Switch Language',
    fullscreen: 'Fullscreen',
    fullscreenExit: 'Exit Fullscreen',
    reload: 'Reload Page',
    collapse: 'Collapse Menu',
    expand: 'Expand Menu',
    pin: 'Pin',
    unpin: 'Unpin'
  },
  datatable: {
    itemCount: 'Total {total} items'
  }
};

export default local;
