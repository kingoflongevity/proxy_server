<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSubscriptionStore, useNodeStore, useSettingsStore } from '@/stores'
import request from '@/api/request'

const subscriptionStore = useSubscriptionStore()
const nodeStore = useNodeStore()
const settingsStore = useSettingsStore()

const coreVersion = ref<string>('')
const coreInstalled = ref<boolean>(false)

interface CoreInfo {
  version: string
  installPath: string
  downloadUrl: string
  installed: boolean
}

async function fetchCoreInfo() {
  try {
    const info = await request.get<CoreInfo>('/core/info')
    coreVersion.value = info.version || '未安装'
    coreInstalled.value = info.installed
  } catch (error) {
    console.error('获取内核信息失败:', error)
    coreVersion.value = '未安装'
  }
}

onMounted(async () => {
  try {
    await subscriptionStore.fetchSubscriptions()
    await nodeStore.fetchNodes()
    await settingsStore.fetchConnectionStatus()
    await nodeStore.fetchCurrentNode()
    await fetchCoreInfo()
  } catch (error) {
    console.error('加载数据失败:', error)
  }
})

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

/**
 * 格式化流量
 */
function formatTraffic(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`
  } else if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(2)} KB`
  } else if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  } else {
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
  }
}
</script>

<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon subscription">
          <svg viewBox="0 0 24 24" fill="none">
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
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">订阅数量</div>
          <div class="stat-value">{{ subscriptionStore.subscriptionCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon nodes">
          <svg viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2" />
            <path
              d="M12 2V6M12 18V22M2 12H6M18 12H22"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">节点数量</div>
          <div class="stat-value">{{ nodeStore.nodeCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon available">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M20 6L9 17L4 12"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">可用节点</div>
          <div class="stat-value">{{ nodeStore.availableCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon connection">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2Z"
              stroke="currentColor"
              stroke-width="2"
            />
            <path
              d="M12 6V12L16 14"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">连接状态</div>
          <div class="stat-value" :class="{ connected: settingsStore.isConnected }">
            {{ settingsStore.isConnected ? '已连接' : '未连接' }}
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon core">
          <svg viewBox="0 0 24 24" fill="none">
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
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">内核版本</div>
          <div class="stat-value" :class="{ installed: coreInstalled }">
            {{ coreVersion || '未安装' }}
          </div>
        </div>
      </div>
    </div>

    <!-- 连接信息 -->
    <div class="section">
      <h3 class="section-title">连接信息</h3>
      <div class="connection-info">
        <div class="info-item">
          <span class="info-label">当前节点</span>
          <span class="info-value">{{ nodeStore.currentNode?.name || '未选择' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">代理模式</span>
          <span class="info-value">
            {{
              settingsStore.proxyMode === 'rule'
                ? '规则模式'
                : settingsStore.proxyMode === 'global'
                  ? '全局模式'
                  : '直连模式'
            }}
          </span>
        </div>
        <div class="info-item">
          <span class="info-label">上传速度</span>
          <span class="info-value">{{ formatSpeed(settingsStore.connectionStatus.uploadSpeed) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">下载速度</span>
          <span class="info-value">{{ formatSpeed(settingsStore.connectionStatus.downloadSpeed) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">上传流量</span>
          <span class="info-value">{{ formatTraffic(settingsStore.connectionStatus.uploadTotal) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">下载流量</span>
          <span class="info-value">{{ formatTraffic(settingsStore.connectionStatus.downloadTotal) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">连接数</span>
          <span class="info-value">{{ settingsStore.connectionStatus.connectionCount }}</span>
        </div>
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="section">
      <h3 class="section-title">快速操作</h3>
      <div class="quick-actions">
        <button class="action-btn" @click="$router.push('/subscriptions')">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 5V19M12 5L5 12M12 5L19 12"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span>添加订阅</span>
        </button>
        <button class="action-btn" @click="$router.push('/nodes')">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 2L2 7L12 12L22 7L12 2Z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span>测试节点</span>
        </button>
        <button class="action-btn" @click="$router.push('/rules')">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 2L2 7L12 12L22 7L12 2Z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span>配置规则</span>
        </button>
        <button class="action-btn" @click="$router.push('/settings')">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 2L2 7L12 12L22 7L12 2Z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span>系统设置</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

// 统计卡片网格
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: $spacing-lg;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-lg;
  background-color: $bg-color-light;
  border-radius: $border-radius-lg;
  border: 1px solid $border-color;
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: $border-radius-base;
  font-size: 24px;

  svg {
    width: 24px;
    height: 24px;
  }

  &.subscription {
    background-color: rgba($primary-color, 0.1);
    color: $primary-color;
  }

  &.nodes {
    background-color: rgba($success-color, 0.1);
    color: $success-color;
  }

  &.available {
    background-color: rgba($warning-color, 0.1);
    color: $warning-color;
  }

  &.connection {
    background-color: rgba($error-color, 0.1);
    color: $error-color;
  }

  &.core {
    background-color: rgba(#9333ea, 0.1);
    color: #9333ea;
  }
}

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: $font-size-sm;
  color: $text-color-secondary;
  margin-bottom: $spacing-xs;
}

.stat-value {
  font-size: $font-size-xl;
  font-weight: 600;
  color: $text-color-primary;

  &.connected {
    color: $success-color;
  }
}

// 区块
.section {
  background-color: $bg-color-light;
  border-radius: $border-radius-lg;
  padding: $spacing-lg;
  border: 1px solid $border-color;
}

.section-title {
  font-size: $font-size-lg;
  font-weight: 500;
  color: $text-color-primary;
  margin-bottom: $spacing-lg;
}

// 连接信息
.connection-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: $spacing-md;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
  padding: $spacing-md;
  background-color: $bg-color-darker;
  border-radius: $border-radius-base;
}

.info-label {
  font-size: $font-size-sm;
  color: $text-color-secondary;
}

.info-value {
  font-size: $font-size-base;
  color: $text-color-primary;
  font-weight: 500;
}

// 快速操作
.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: $spacing-md;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-lg;
  background-color: $bg-color-darker;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  color: $text-color-primary;
  transition: all $transition-duration $transition-timing;

  svg {
    width: 24px;
    height: 24px;
  }

  &:hover {
    background-color: $bg-color-lighter;
    border-color: $primary-color;
    color: $primary-color;
  }
}

// 响应式
@media (max-width: $breakpoint-md) {
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .connection-info {
    grid-template-columns: 1fr;
  }

  .quick-actions {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
