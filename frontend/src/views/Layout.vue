<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSettingsStore, useNodeStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const settingsStore = useSettingsStore()
const nodeStore = useNodeStore()

const collapsed = ref(false)

onMounted(async () => {
  await settingsStore.fetchProxyMode()
})

/**
 * 菜单项
 */
const menuItems = [
  {
    path: '/dashboard',
    name: 'Dashboard',
    title: '仪表盘',
    icon: 'dashboard',
  },
  {
    path: '/subscriptions',
    name: 'Subscriptions',
    title: '订阅管理',
    icon: 'subscription',
  },
  {
    path: '/nodes',
    name: 'Nodes',
    title: '节点列表',
    icon: 'nodes',
  },
  {
    path: '/cluster',
    name: 'Cluster',
    title: '集群管理',
    icon: 'cluster',
  },
  {
    path: '/rules',
    name: 'Rules',
    title: '规则配置',
    icon: 'rules',
  },
  {
    path: '/settings',
    name: 'Settings',
    title: '系统设置',
    icon: 'settings',
  },
  {
    path: '/logs',
    name: 'Logs',
    title: '流量日志',
    icon: 'logs',
  },
]

/**
 * 当前激活的菜单
 */
const activeMenu = computed(() => route.name as string)

/**
 * 切换侧边栏折叠状态
 */
function toggleCollapse() {
  collapsed.value = !collapsed.value
}

/**
 * 导航到指定路由
 */
async function navigateTo(path: string) {
  if (route.path === path) {
    return
  }
  
  try {
    await router.push(path)
  } catch (error) {
    console.error('导航失败:', error)
  }
}

/**
 * 格式化速度
 */
function formatSpeed(bytesPerSecond: number): string {
  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond} B/s`
  } else if (bytesPerSecond < 1024 * 1024) {
    return `${(bytesPerSecond / 1024).toFixed(2)} KB/s`
  } else {
    return `${(bytesPerSecond / 1024 / 1024).toFixed(2)} MB/s`
  }
}
</script>

<template>
  <div class="layout">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed }">
      <!-- Logo -->
      <div class="sidebar-header">
        <div class="logo">
          <svg class="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2L2 7L12 12L22 7L12 2Z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              d="M2 17L12 22L22 17"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              d="M2 12L12 17L22 12"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span v-if="!collapsed" class="logo-text">Proxy Manager</span>
        </div>
      </div>

      <!-- 菜单 -->
      <nav class="sidebar-menu">
        <div
          v-for="item in menuItems"
          :key="item.path"
          class="menu-item"
          :class="{ active: activeMenu === item.name }"
          @click="navigateTo(item.path)"
        >
          <svg class="menu-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="2" />
            <path d="M9 9H15" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            <path d="M9 12H15" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            <path d="M9 15H12" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
          <span v-if="!collapsed" class="menu-text">{{ item.title }}</span>
        </div>
      </nav>

      <!-- 折叠按钮 -->
      <div class="sidebar-footer">
        <button class="collapse-btn" @click="toggleCollapse">
          <svg
            class="collapse-icon"
            :class="{ rotated: collapsed }"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M15 18L9 12L15 6"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="main-container">
      <!-- 头部 -->
      <header class="header">
        <div class="header-left">
          <h2 class="page-title">{{ route.meta.title }}</h2>
        </div>
        <div class="header-right">
          <!-- 连接状态 -->
          <div class="connection-status">
            <div class="status-indicator" :class="{ connected: settingsStore.isConnected }"></div>
            <span class="status-text">
              {{ settingsStore.isConnected ? '已连接' : '未连接' }}
            </span>
          </div>

          <!-- 当前节点 -->
          <div v-if="nodeStore.currentNode" class="current-node">
            <span class="node-name">{{ nodeStore.currentNode.name }}</span>
            <span class="node-latency">{{ nodeStore.currentNode.latency }}ms</span>
          </div>

          <!-- 速度显示 -->
          <div class="speed-display">
            <div class="speed-item">
              <svg class="speed-icon upload" viewBox="0 0 24 24" fill="none">
                <path
                  d="M12 19V5M12 5L5 12M12 5L19 12"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              <span>{{ formatSpeed(settingsStore.connectionStatus.uploadSpeed) }}</span>
            </div>
            <div class="speed-item">
              <svg class="speed-icon download" viewBox="0 0 24 24" fill="none">
                <path
                  d="M12 5V19M12 19L5 12M12 19L19 12"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              <span>{{ formatSpeed(settingsStore.connectionStatus.downloadSpeed) }}</span>
            </div>
          </div>

          <!-- 代理模式切换 -->
          <div class="proxy-mode">
            <button
              v-for="mode in ['rule', 'global', 'direct']"
              :key="mode"
              class="mode-btn"
              :class="{ active: settingsStore.proxyMode === mode }"
              @click="settingsStore.toggleProxyMode(mode as any)"
            >
              {{ mode === 'rule' ? '规则' : mode === 'global' ? '全局' : '直连' }}
            </button>
          </div>
        </div>
      </header>

      <!-- 内容区 -->
      <main class="content">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="route.fullPath" />
        </router-view>
      </main>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.layout {
  display: flex;
  width: 100%;
  height: 100%;
  background-color: $bg-color-dark;
}

// 侧边栏
.sidebar {
  display: flex;
  flex-direction: column;
  width: $sidebar-width;
  height: 100%;
  background-color: $sidebar-bg;
  border-right: 1px solid $border-color;
  transition: width $transition-duration $transition-timing;

  &.collapsed {
    width: $sidebar-collapsed-width;

    .logo-text,
    .menu-text {
      display: none;
    }
  }
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: center;
  height: $header-height;
  border-bottom: 1px solid $border-color;
}

.logo {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  color: $primary-color;
  font-size: $font-size-lg;
  font-weight: 600;
}

.logo-icon {
  width: 32px;
  height: 32px;
}

.sidebar-menu {
  flex: 1;
  padding: $spacing-md 0;
  overflow-y: auto;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-md $spacing-lg;
  color: $text-color-secondary;
  cursor: pointer;
  transition: all $transition-duration $transition-timing;

  &:hover {
    color: $text-color-primary;
    background-color: $bg-color-light;
  }

  &.active {
    color: $primary-color;
    background-color: rgba($primary-color, 0.1);
    border-right: 3px solid $primary-color;
  }
}

.menu-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.sidebar-footer {
  padding: $spacing-md;
  border-top: 1px solid $border-color;
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: $spacing-sm;
  background-color: transparent;
  color: $text-color-secondary;
  border-radius: $border-radius-base;

  &:hover {
    background-color: $bg-color-light;
    color: $text-color-primary;
  }
}

.collapse-icon {
  width: 20px;
  height: 20px;
  transition: transform $transition-duration $transition-timing;

  &.rotated {
    transform: rotate(180deg);
  }
}

// 主容器
.main-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

// 头部
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: $header-height;
  padding: 0 $spacing-lg;
  background-color: $header-bg;
  border-bottom: 1px solid $border-color;
}

.header-left {
  display: flex;
  align-items: center;
}

.page-title {
  font-size: $font-size-lg;
  font-weight: 500;
  color: $text-color-primary;
}

.header-right {
  display: flex;
  align-items: center;
  gap: $spacing-lg;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: $error-color;

  &.connected {
    background-color: $success-color;
  }
}

.status-text {
  font-size: $font-size-sm;
  color: $text-color-secondary;
}

.current-node {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.node-name {
  font-size: $font-size-sm;
  color: $text-color-primary;
  font-weight: 500;
}

.node-latency {
  font-size: $font-size-xs;
  color: $text-color-secondary;
}

.speed-display {
  display: flex;
  gap: $spacing-md;
}

.speed-item {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  font-size: $font-size-sm;
  color: $text-color-secondary;
}

.speed-icon {
  width: 16px;
  height: 16px;

  &.upload {
    color: $success-color;
  }

  &.download {
    color: $primary-color;
  }
}

.proxy-mode {
  display: flex;
  gap: $spacing-xs;
  padding: 2px;
  background-color: $bg-color-light;
  border-radius: $border-radius-base;
}

.mode-btn {
  padding: $spacing-xs $spacing-md;
  font-size: $font-size-sm;
  color: $text-color-secondary;
  background-color: transparent;
  border-radius: $border-radius-sm;

  &:hover {
    color: $text-color-primary;
  }

  &.active {
    color: $text-color-primary;
    background-color: $primary-color;
  }
}

// 内容区
.content {
  flex: 1;
  padding: $spacing-lg;
  overflow-y: auto;
  background-color: $bg-color-dark;
}

// 响应式
@media (max-width: $breakpoint-md) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    z-index: 1000;
    transform: translateX(-100%);

    &.collapsed {
      transform: translateX(0);
      width: $sidebar-width;
    }
  }

  .header-right {
    gap: $spacing-md;
  }

  .speed-display,
  .current-node {
    display: none;
  }
}
</style>
