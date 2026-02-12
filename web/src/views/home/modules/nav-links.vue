<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/store/modules/app';

defineOptions({
  name: 'NavLinks'
});

const appStore = useAppStore();

const gap = computed(() => (appStore.isMobile ? 0 : 16));

interface LinkItem {
  id: number;
  label: string;
  desc?: string;
  url: string;
  icon?: string;
  color?: string;
}

interface Category {
  id: number;
  title: string;
  icon?: string;
  items: LinkItem[];
}

const categories: Category[] = [
  {
    id: 1,
    title: '内部系统',
    icon: 'carbon:cloud-service-management',
    items: [
      {
        id: 1,
        label: '企享云业务支撑系统',
        desc: '业务运营与支撑平台',
        url: 'https://ebc.qixiangyun.com/#/root/simplelogin?appkey=10009001',
        icon: 'carbon:application-web',
        color: '#1890ff'
      },
      {
        id: 2,
        label: 'HTQ任务调度',
        desc: '分布式任务调度平台',
        url: 'https://htq.qixiangyun.com/',
        icon: 'carbon:task',
        color: '#52c41a'
      },
      {
        id: 3,
        label: 'jira研发任务管理',
        desc: '项目与任务追踪管理',
        url: 'http://jira.qixiangyun.com',
        icon: 'carbon:chart-bar-floating',
        color: '#2f54eb'
      },
      {
        id: 4,
        label: 'wiki文档协同管理',
        desc: '团队知识库与文档协作',
        url: 'http://wiki.qixiangyun.com',
        icon: 'carbon:document',
        color: '#faad14'
      },
      {
        id: 5,
        label: '企享云产品对接文档',
        desc: '新版产品接口对接文档',
        url: 'http://openapi.qixiangyun.com/',
        icon: 'carbon:api',
        color: '#722ed1'
      },
      {
        id: 6,
        label: '对接文档（V1）',
        desc: '旧版产品接口文档',
        url: 'https://opendoc.qixiangyun.com',
        icon: 'carbon:document-blank',
        color: '#eb2f96'
      },
      {
        id: 7,
        label: 'grafana监控',
        desc: '系统性能与指标监控',
        url: 'http://grafana.qixiangyun.com',
        icon: 'carbon:dashboard',
        color: '#f5222d'
      },
      {
        id: 8,
        label: '阿里云RAM',
        desc: '阿里云访问控制登录',
        url: 'https://signin.aliyun.com/login.htm?callback=https%3A%2F%2Fwww.aliyun.com%2F#/main',
        icon: 'carbon:user-avatar-filled',
        color: '#ff85c0'
      },
      {
        id: 9,
        label: '公司官网',
        desc: '企享云官方网站',
        url: 'https://www.qixiangyun.com',
        icon: 'carbon:building',
        color: '#13c2c2'
      },
      {
        id: 10,
        label: 'ebc生产环境',
        desc: '生产环境业务系统',
        url: 'https://ebc.qixiangyun.com',
        icon: 'carbon:cloud',
        color: '#52c41a'
      },
      {
        id: 11,
        label: 'ebc预生产环境',
        desc: '预发布环境业务系统',
        url: 'http://release.qixiangyun.com',
        icon: 'carbon:cloud-upload',
        color: '#fa8c16'
      },
      {
        id: 12,
        label: 'ebc测试环境',
        desc: '测试环境业务系统',
        url: 'http://test.qixiangyun.com',
        icon: 'carbon:cloud-satellite',
        color: '#fadb14'
      },
      {
        id: 13,
        label: '业务运营系统',
        desc: 'BOSS后台管理系统',
        url: 'https://boss.qixiangyun.com',
        icon: 'carbon:operations-record',
        color: '#2f54eb'
      },
      {
        id: 14,
        label: '部署运营管理',
        desc: 'Tkubord容器管理平台',
        url: 'http://tkubord.qixiangyun.com',
        icon: 'carbon:deploy',
        color: '#722ed1'
      },
      {
        id: 15,
        label: '构建部署管理',
        desc: 'Jenkins持续集成平台',
        url: 'https://devops.qixiangyun.com',
        icon: 'carbon:tool-box',
        color: '#1890ff'
      },
      {
        id: 16,
        label: '代码库',
        desc: 'Gitlab代码托管服务',
        url: 'http://git.qixiangyun.com',
        icon: 'carbon:code',
        color: '#fa541c'
      },
      {
        id: 17,
        label: '镜像库',
        desc: 'Harbor容器镜像仓库',
        url: 'http://harbor.qixiangyun.com',
        icon: 'carbon:container-software',
        color: '#13c2c2'
      },
      {
        id: 18,
        label: 'jira手机版',
        desc: '移动端任务管理入口',
        url: 'http://wenti.qixiangyun.com/',
        icon: 'carbon:mobile',
        color: '#eb2f96'
      },
      {
        id: 19,
        label: '企享云运维平台',
        desc: '一站式运维管理平台',
        url: 'http://devops.qixiangyun.com/',
        icon: 'carbon:settings',
        color: '#52c41a'
      }
    ]
  },
  {
    id: 2,
    title: '常用工具',
    icon: 'carbon:tool-box',
    items: [
      {
        id: 101,
        label: 'JSON格式化',
        desc: '在线JSON校验格式化工具',
        url: 'https://www.json.cn/',
        icon: 'carbon:json',
        color: '#faad14'
      },
      {
        id: 102,
        label: '时间戳转换',
        desc: 'Unix时间戳转换工具',
        url: 'https://tool.lu/timestamp/',
        icon: 'carbon:time',
        color: '#1890ff'
      },
      {
        id: 103,
        label: 'Base64编解码',
        desc: 'Base64在线编码解码',
        url: 'https://base64.us/',
        icon: 'carbon:data-blob',
        color: '#52c41a'
      }
    ]
  }
];

function handleLinkClick(url: string) {
  window.open(url, '_blank');
}
</script>

<template>
  <NSpace vertical :size="24">
    <div v-for="category in categories" :key="category.id">
      <div class="mb-16px flex items-center gap-8px">
        <SvgIcon :icon="category.icon || 'carbon:catalog'" class="text-20px text-gray-500" />
        <h3 class="text-16px text-gray-700 font-bold dark:text-gray-200">{{ category.title }}</h3>
      </div>
      <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
        <NGi v-for="item in category.items" :key="item.id" span="24 s:12 m:8 l:6">
          <NCard
            hoverable
            size="small"
            class="group relative h-full cursor-pointer overflow-hidden transition-all hover:shadow-lg hover:-translate-y-2px"
            @click="handleLinkClick(item.url)"
          >
            <div
              class="absolute inset-0 transition-opacity duration-300"
              :style="{
                background: `linear-gradient(to right, transparent, ${item.color}25)`,
                maskImage: 'linear-gradient(to bottom, black 0%, transparent 50%)',
                WebkitMaskImage: 'linear-gradient(to bottom, black 0%, transparent 50%)'
              }"
            ></div>
            <div class="relative h-full flex flex-col gap-12px">
              <div class="flex items-start gap-12px">
                <div
                  class="h-40px w-40px flex shrink-0 items-center justify-center rounded-8px transition-colors"
                  :style="{ backgroundColor: `${item.color}1a` }"
                >
                  <SvgIcon :icon="item.icon || 'carbon:link'" class="text-24px" :style="{ color: item.color }" />
                </div>
                <div class="min-w-0 flex flex-col flex-1 gap-2px">
                  <div class="flex items-center justify-between">
                    <span class="truncate text-14px text-gray-800 font-bold dark:text-gray-200">{{ item.label }}</span>
                    <span class="shrink-0 text-12px text-gray-400">[{{ category.title }}]</span>
                  </div>
                </div>
              </div>

              <div class="flex-1">
                <p class="line-clamp-2 text-12px text-gray-500 dark:text-gray-400" :title="item.desc">
                  {{ item.desc || '暂无描述' }}
                </p>
              </div>

              <div class="flex items-center justify-between pt-12px">
                <span class="max-w-[70%] truncate text-12px text-gray-400" :title="item.url">{{ item.url }}</span>
                <span class="text-12px text-primary">进入</span>
              </div>
            </div>
          </NCard>
        </NGi>
      </NGrid>
    </div>
  </NSpace>
</template>

<style scoped></style>
