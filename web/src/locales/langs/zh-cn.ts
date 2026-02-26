const local: App.I18n.Schema = {
  system: {
    title: 'Noah 管理系统',
    updateTitle: '系统版本更新通知',
    updateContent: '检测到系统有新版本发布，是否立即刷新页面？',
    updateConfirm: '立即刷新',
    updateCancel: '稍后再说'
  },
  common: {
    action: '操作',
    add: '新增',
    addSuccess: '添加成功',
    backToHome: '返回首页',
    batchDelete: '批量删除',
    cancel: '取消',
    close: '关闭',
    check: '勾选',
    expandColumn: '展开列',
    columnSetting: '列设置',
    exportResult: '导出',
    exportCsv: '导出 CSV',
    exportSql: '导出 SQL',
    exportTxt: '导出 TXT',
    config: '配置',
    confirm: '确认',
    delete: '删除',
    deleteSuccess: '删除成功',
    confirmDelete: '确认删除吗？',
    edit: '编辑',
    warning: '警告',
    error: '错误',
    index: '序号',
    keywordSearch: '请输入关键词搜索',
    logout: '退出登录',
    logoutConfirm: '确认退出登录吗？',
    lookForward: '敬请期待',
    modify: '修改',
    modifySuccess: '修改成功',
    noData: '无数据',
    operate: '操作',
    pleaseCheckValue: '请检查输入的值是否合法',
    refresh: '刷新',
    reset: '重置',
    search: '搜索',
    switch: '切换',
    tip: '提示',
    trigger: '触发',
    update: '更新',
    updateSuccess: '更新成功',
    userCenter: '个人中心',
    yesOrNo: {
      yes: '是',
      no: '否'
    }
  },
  request: {
    logout: '请求失败后登出用户',
    logoutMsg: '用户状态失效，请重新登录',
    logoutWithModal: '请求失败后弹出模态框再登出用户',
    logoutWithModalMsg: '用户状态失效，请重新登录',
    refreshToken: '请求的token已过期，刷新token',
    tokenExpired: 'token已过期'
  },
  theme: {
    themeSchema: {
      title: '主题模式',
      light: '亮色模式',
      dark: '暗黑模式',
      auto: '跟随系统'
    },
    grayscale: '灰色模式',
    colourWeakness: '色弱模式',
    layoutMode: {
      title: '布局模式',
      vertical: '左侧菜单模式',
      'vertical-mix': '左侧菜单混合模式',
      horizontal: '顶部菜单模式',
      'horizontal-mix': '顶部菜单混合模式',
      reverseHorizontalMix: '一级菜单与子级菜单位置反转'
    },
    recommendColor: '应用推荐算法的颜色',
    recommendColorDesc: '推荐颜色的算法参照',
    themeColor: {
      title: '主题颜色',
      primary: '主色',
      info: '信息色',
      success: '成功色',
      warning: '警告色',
      error: '错误色',
      followPrimary: '跟随主色'
    },
    scrollMode: {
      title: '滚动模式',
      wrapper: '外层滚动',
      content: '主体滚动'
    },
    page: {
      animate: '页面切换动画',
      mode: {
        title: '页面切换动画类型',
        'fade-slide': '滑动',
        fade: '淡入淡出',
        'fade-bottom': '底部消退',
        'fade-scale': '缩放消退',
        'zoom-fade': '渐变',
        'zoom-out': '闪现',
        none: '无'
      }
    },
    fixedHeaderAndTab: '固定头部和标签栏',
    header: {
      height: '头部高度',
      breadcrumb: {
        visible: '显示面包屑',
        showIcon: '显示面包屑图标'
      },
      multilingual: {
        visible: '显示多语言按钮'
      },
      globalSearch: {
        visible: '显示全局搜索按钮'
      }
    },
    tab: {
      visible: '显示标签栏',
      cache: '标签栏信息缓存',
      height: '标签栏高度',
      mode: {
        title: '标签栏风格',
        chrome: '谷歌风格',
        button: '按钮风格'
      }
    },
    sider: {
      inverted: '深色侧边栏',
      width: '侧边栏宽度',
      collapsedWidth: '侧边栏折叠宽度',
      mixWidth: '混合布局侧边栏宽度',
      mixCollapsedWidth: '混合布局侧边栏折叠宽度',
      mixChildMenuWidth: '混合布局子菜单宽度'
    },
    footer: {
      visible: '显示底部',
      fixed: '固定底部',
      height: '底部高度',
      right: '底部局右'
    },
    watermark: {
      visible: '显示全屏水印',
      text: '水印文本',
      enableUserName: '启用用户名水印'
    },
    themeDrawerTitle: '主题配置',
    pageFunTitle: '页面功能',
    resetCacheStrategy: {
      title: '重置缓存策略',
      close: '关闭页面',
      refresh: '刷新页面'
    },
    configOperation: {
      copyConfig: '复制配置',
      copySuccessMsg: '复制成功，请替换 src/theme/settings.ts 中的变量 themeSettings',
      resetConfig: '重置配置',
      resetSuccessMsg: '重置成功'
    }
  },
  route: {
    login: '登录',
    403: '无权限',
    404: '页面不存在',
    500: '服务器错误',
    'iframe-page': '外链页面',
    home: '首页',
    das: '数据库服务',
    das_modules_edit: 'SQL查询',
    das_modules_favorite: '收藏SQL',
    das_modules_history: '历史查询',
    sql: 'SQL查询',
    das_favorite: '收藏SQL',
    das_history: '历史查询',
    das_edit: 'SQL查询',
    'das_orders-list': '工单列表',
    'das_orders-detail': '工单详情',
    das_orders_list: '工单列表',
    das_orders: '提交工单',
    das_orders_ddl: 'DDL工单',
    das_orders_dml: 'DML工单',
    das_orders_export: '导出工单',
    das_orders_commit: '提交工单',
    das_orders_commit_ddl: 'DDL工单',
    das_orders_commit_dml: 'DML工单',
    das_orders_commit_export: '导出工单',
    system: '系统管理',
    system_menu: '菜单管理',
    system_user: '用户管理',
    system_permission: '权限管理',
    system_database: '数据库管理',
    system_database_config: '数据库配置',
    system_database_environment: '环境管理',
    system_database_inspect: '审核参数配置',
    system_database_permission: '权限管理',
    system_database_permission_assign: '权限分配',
    system_database_permission_template: '权限模板'
  },
  page: {
    manage: {
      common: {
        status: {
          enable: '启用',
          disable: '禁用'
        }
      },
      menu: {
        home: '首页',
        title: '菜单列表',
        id: 'ID',
        parentId: '父级菜单ID',
        menuType: '菜单类型',
        menuName: '菜单名称',
        routeName: '路由名称',
        routePath: '路由路径',
        pathParam: '路径参数',
        layout: '布局',
        page: '页面组件',
        i18nKey: '国际化key',
        icon: '图标',
        localIcon: '本地图标',
        iconTypeTitle: '图标类型',
        order: '排序',
        constant: '常量路由',
        keepAlive: '缓存路由',
        href: '外链',
        hideInMenu: '隐藏菜单',
        activeMenu: '高亮的菜单',
        multiTab: '支持多页签',
        fixedIndexInTab: '固定在页签中的序号',
        query: '路由参数',
        button: '按钮',
        buttonCode: '按钮编码',
        buttonDesc: '按钮描述',
        menuStatus: '菜单状态',
        addMenu: '新增菜单',
        addChildMenu: '新增子菜单',
        editMenu: '编辑菜单',
        type: {
          directory: '目录',
          menu: '菜单'
        },
        iconType: {
          iconify: 'Iconify图标',
          local: '本地图标'
        },
        form: {
          home: '请选择首页',
          menuType: '请选择菜单类型',
          menuName: '请输入菜单名称',
          routeName: '请输入路由名称',
          routePath: '请输入路由路径',
          pathParam: '请输入路径参数',
          page: '请选择页面组件',
          layout: '请选择布局组件',
          i18nKey: '请输入国际化key',
          icon: '请输入图标',
          localIcon: '请选择本地图标',
          order: '请输入排序',
          keepAlive: '请选择是否缓存路由',
          href: '请输入外链',
          hideInMenu: '请选择是否隐藏菜单',
          activeMenu: '请选择高亮的菜单的路由名称',
          multiTab: '请选择是否支持多标签',
          fixedInTab: '请选择是否固定在页签中',
          fixedIndexInTab: '请输入固定在页签中的序号',
          queryKey: '请输入路由参数Key',
          queryValue: '请输入路由参数Value',
          buttonCode: '请输入按钮编码',
          buttonDesc: '请输入按钮描述'
        }
      },
      user: {
        title: '用户列表',
        username: '用户名',
        nickname: '昵称',
        email: '邮箱',
        phone: '手机号',
        password: '密码',
        roles: '角色',
        createdAt: '创建时间',
        addUser: '添加用户',
        editUser: '编辑用户',
        form: {
          username: '请输入用户名',
          nickname: '请输入昵称',
          password: '请输入密码',
          passwordPlaceholder: '留空则不修改密码',
          passwordRequired: '密码不能为空',
          email: '请输入邮箱',
          phone: '请输入手机号',
          roles: '请选择角色'
        }
      },
      role: {
        title: '角色管理',
        roleName: '角色名称',
        roleCode: '角色编码',
        roleDesc: '角色描述',
        roleStatus: '角色状态',
        createdAt: '创建时间',
        addRole: '新增角色',
        editRole: '编辑角色',
        assignPermission: '分配权限',
        menuPermission: '菜单权限',
        apiPermission: 'API权限',
        buttonAuth: '按钮权限',
        menuAuth: '菜单权限',
        form: {
          roleName: '请输入角色名称',
          roleCode: '请输入角色编码',
          roleDesc: '请输入角色描述',
          roleStatus: '请选择角色状态'
        }
      },
      api: {
        title: 'API管理',
        group: '分组',
        name: '名称',
        path: '路径',
        method: '请求方法',
        createdAt: '创建时间',
        addApi: '新增API',
        editApi: '编辑API',
        syncRoute: '同步路由',
        confirmSync: '确认同步',
        freshCasbin: '刷新权限',
        batchDelete: '批量删除',
        syncApiTitle: '同步路由',
        syncApiTip: '对比当前代码路由与数据库，可新增或删除 API；不需要参与鉴权的接口可点击「忽略」。',
        newApis: '新增路由',
        newApisTip: '存在于当前路由中，但不存在于 API 表',
        deleteApis: '待删除路由',
        deleteApisTip: '存在于 API 表，但当前路由中已不存在',
        ignoreApis: '已忽略',
        ignoreApisTip: '不参与同步与鉴权',
        singleAdd: '单条新增',
        singleDelete: '单条删除',
        groupPlaceholder: '请选择或新增',
        nameRequired: '请填写API名称，或使用「AI 自动填充」',
        aiAutoFill: 'AI 自动填充',
        aiAutoFillComing: 'AI 自动填充功能敬请期待',
        aiAutoFillSuccess: '已根据路由自动填充名称与分组',
        noNewApisToFill: '暂无待填充的新路由',
        newGroup: '新建分组',
        newGroupPlaceholder: '请输入新分组名称',
        form: {
          group: '请输入分组',
          name: '请输入名称',
          path: '请输入路径',
          method: '请选择请求方法'
        }
      },
      database: {
        title: '数据库管理',
        environment: {
          title: '环境管理',
          name: '环境名称',
          createdAt: '创建时间',
          updatedAt: '更新时间',
          addEnvironment: '新增环境',
          editEnvironment: '编辑环境',
          form: {
            name: '请输入环境名称',
            nameRequired: '环境名称不能为空'
          }
        },
        config: {
          title: '数据库配置',
          instanceId: '实例ID',
          hostname: '主机名',
          port: '端口',
          userName: '用户名',
          password: '密码',
          useType: '用途',
          dbType: '数据库类型',
          environment: '环境',
          organizationKey: '组织架构',
          remark: '备注',
          createdAt: '创建时间',
          updatedAt: '更新时间',
          addConfig: '新增数据库配置',
          editConfig: '编辑数据库配置',
          form: {
            hostname: '请输入主机名',
            port: '请输入端口',
            userName: '请输入用户名',
            password: '请输入密码',
            useType: '请选择用途',
            dbType: '请选择数据库类型',
            environment: '请选择环境',
            organizationKey: '请输入组织架构',
            remark: '请输入备注'
          }
        },
        permission: {
          title: '数据库权限',
          username: '用户名',
          instanceId: '实例ID',
          schema: '库名',
          createdAt: '创建时间',
          updatedAt: '更新时间',
          addPermission: '新增权限',
          selectUser: '选择用户',
          selectUserFirst: '请先选择用户',
          pleaseSelectUser: '请选择用户',
          form: {
            instanceId: '请选择实例',
            schema: '请选择库名'
          }
        },
        permissionTemplate: {
          title: '权限模板',
          name: '模板名称',
          description: '模板描述',
          permissions: '权限数量',
          createdAt: '创建时间',
          updatedAt: '更新时间',
          addTemplate: '新增模板',
          editTemplate: '编辑模板',
          form: {
            name: '请输入模板名称',
            description: '请输入模板描述'
          }
        },
        permissionAssign: {
          title: '权限分配',
          roleTab: '角色权限',
          userTab: '用户权限'
        },
        rolePermission: {
          title: '角色权限',
          role: '角色',
          permissionType: '权限类型',
          permissionId: '权限ID',
          instanceId: '实例ID',
          schema: '库名',
          table: '表名',
          createdAt: '创建时间',
          form: {
            role: '请选择角色',
            permissionType: '请选择权限类型',
            permissionId: '请选择权限模板',
            instanceId: '请选择实例',
            schema: '请选择库名'
          }
        }
      }
    },
    login: {
      common: {
        loginOrRegister: '登录 / 注册',
        userNamePlaceholder: '请输入用户名',
        phonePlaceholder: '请输入手机号',
        codePlaceholder: '请输入验证码',
        passwordPlaceholder: '请输入密码',
        confirmPasswordPlaceholder: '请再次输入密码',
        codeLogin: '验证码登录',
        confirm: '确定',
        back: '返回',
        validateSuccess: '验证成功',
        loginSuccess: '登录成功',
        welcomeBack: '欢迎回来，{userName} ！'
      },
      pwdLogin: {
        title: '密码登录',
        rememberMe: '记住我',
        forgetPassword: '忘记密码？',
        register: '注册账号',
        otherAccountLogin: '其他账号登录',
        otherLoginMode: '其他登录方式',
        superAdmin: '超级管理员',
        admin: '管理员',
        user: '普通用户'
      },
      codeLogin: {
        title: '验证码登录',
        getCode: '获取验证码',
        reGetCode: '{time}秒后重新获取',
        sendCodeSuccess: '验证码发送成功',
        imageCodePlaceholder: '请输入图片验证码'
      },
      register: {
        title: '注册账号',
        agreement: '我已经仔细阅读并接受',
        protocol: '《用户协议》',
        policy: '《隐私权政策》'
      },
      resetPwd: {
        title: '重置密码'
      },
      bindWeChat: {
        title: '绑定微信'
      }
    },
    home: {
      branchDesc:
        '为了方便大家开发和更新合并，我们对main分支的代码进行了精简，只保留了首页菜单，其余内容已移至example分支进行维护。预览地址显示的内容即为example分支的内容。',
      greeting: '早安，{userName}, 今天又是充满活力的一天!',
      weatherDesc: '今日多云转晴，20℃ - 25℃!',
      projectCount: '项目数',
      todo: '待办',
      message: '消息',
      downloadCount: '下载量',
      registerCount: '注册量',
      schedule: '作息安排',
      study: '学习',
      work: '工作',
      rest: '休息',
      entertainment: '娱乐',
      visitCount: '访问量',
      turnover: '成交额',
      dealCount: '成交量',
      projectNews: {
        title: '项目动态',
        moreNews: '更多动态',
        desc1: 'Soybean 在2021年5月28日创建了开源项目 soybean-admin!',
        desc2: 'Yanbowe 向 soybean-admin 提交了一个bug，多标签栏不会自适应。',
        desc3: 'Soybean 准备为 soybean-admin 的发布做充分的准备工作!',
        desc4: 'Soybean 正在忙于为soybean-admin写项目说明文档！',
        desc5: 'Soybean 刚才把工作台页面随便写了一些，凑合能看了！'
      },
      creativity: '创意'
    }
  },
  form: {
    required: '不能为空',
    userName: {
      required: '请输入用户名',
      invalid: '用户名格式不正确'
    },
    phone: {
      required: '请输入手机号',
      invalid: '手机号格式不正确'
    },
    pwd: {
      required: '请输入密码',
      invalid: '密码格式不正确，6-18位字符，包含字母、数字、下划线'
    },
    confirmPwd: {
      required: '请输入确认密码',
      invalid: '两次输入密码不一致'
    },
    code: {
      required: '请输入验证码',
      invalid: '验证码格式不正确'
    },
    email: {
      required: '请输入邮箱',
      invalid: '邮箱格式不正确'
    }
  },
  dropdown: {
    closeCurrent: '关闭',
    closeOther: '关闭其它',
    closeLeft: '关闭左侧',
    closeRight: '关闭右侧',
    closeAll: '关闭所有'
  },
  icon: {
    themeConfig: '主题配置',
    themeSchema: '主题模式',
    lang: '切换语言',
    fullscreen: '全屏',
    fullscreenExit: '退出全屏',
    reload: '刷新页面',
    collapse: '折叠菜单',
    expand: '展开菜单',
    pin: '固定',
    unpin: '取消固定'
  },
  datatable: {
    itemCount: '共 {total} 条'
  }
};

export default local;
