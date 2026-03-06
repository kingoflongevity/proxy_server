<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores'
import type { Theme, Language, LogLevel, ProxyMode } from '@/types'

const settingsStore = useSettingsStore()

const activeTab = ref<'general' | 'network' | 'advanced'>('general')

/**
 * 初始化
 */
onMounted(async () => {
  await settingsStore.fetchSettings()
  await settingsStore.fetchSystemInfo()
})

/**
 * 切换主题
 */
async function handleThemeChange(theme: Theme) {
  try {
    await settingsStore.toggleTheme(theme)
  } catch (error) {
    console.error('切换主题失败:', error)
  }
}

/**
 * 切换语言
 */
async function handleLanguageChange(language: Language) {
  try {
    await settingsStore.toggleLanguage(language)
  } catch (error) {
    console.error('切换语言失败:', error)
  }
}

/**
 * 切换代理模式
 */
async function handleProxyModeChange(mode: ProxyMode) {
  try {
    await settingsStore.toggleProxyMode(mode)
  } catch (error) {
    console.error('切换代理模式失败:', error)
  }
}

/**
 * 更新设置
 */
async function updateSetting(key: string, value: any) {
  try {
    await settingsStore.updateSettings({ [key]: value })
  } catch (error) {
    console.error('更新设置失败:', error)
  }
}

/**
 * 重启服务
 */
async function handleRestart() {
  if (confirm('确定要重启服务吗？')) {
    try {
      await settingsStore.restartService()
      alert('服务重启中...')
    } catch (error) {
      console.error('重启服务失败:', error)
    }
  }
}

/**
 * 导出配置
 */
async function handleExportConfig() {
  try {
    const config = await settingsStore.exportConfig()
    const blob = new Blob([config], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'proxy-config.json'
    a.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('导出配置失败:', error)
  }
}

/**
 * 导入配置
 */
async function handleImportConfig(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const config = await file.text()
    await settingsStore.importConfig(config)
    alert('配置导入成功')
  } catch (error) {
    console.error('导入配置失败:', error)
    alert('配置导入失败')
  }
}

/**
 * 清除缓存
 */
async function handleClearCache() {
  if (confirm('确定要清除缓存吗？')) {
    try {
      await settingsStore.clearCache()
      alert('缓存清除成功')
    } catch (error) {
      console.error('清除缓存失败:', error)
    }
  }
}

/**
 * 格式化运行时间
 */
function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  let result = ''
  if (days > 0) result += `${days}天 `
  if (hours > 0) result += `${hours}小时 `
  result += `${minutes}分钟`
  return result
}

/**
 * 格式化内存
 */
function formatMemory(bytes: number): string {
  const mb = bytes / 1024 / 1024
  return `${mb.toFixed(2)} MB`
}
</script>

<template>
  <div class="settings">
    <h2 class="page-title">系统设置</h2>

    <!-- 标签页 -->
    <div class="tabs">
      <button
        class="tab"
        :class="{ active: activeTab === 'general' }"
        @click="activeTab = 'general'"
      >
        常规设置
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'network' }"
        @click="activeTab = 'network'"
      >
        网络设置
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'advanced' }"
        @click="activeTab = 'advanced'"
      >
        高级设置
      </button>
    </div>

    <!-- 常规设置 -->
    <div v-if="activeTab === 'general'" class="settings-section">
      <div class="setting-item">
        <div class="setting-label">
          <h4>主题</h4>
          <p>选择应用的主题外观</p>
        </div>
        <div class="setting-control">
          <select
            :value="settingsStore.settings.theme"
            @change="handleThemeChange(($event.target as HTMLSelectElement).value as Theme)"
          >
            <option value="dark">深色</option>
            <option value="light">浅色</option>
            <option value="auto">跟随系统</option>
          </select>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>语言</h4>
          <p>选择应用的语言</p>
        </div>
        <div class="setting-control">
          <select
            :value="settingsStore.settings.language"
            @change="handleLanguageChange(($event.target as HTMLSelectElement).value as Language)"
          >
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
          </select>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>代理模式</h4>
          <p>选择默认的代理模式</p>
        </div>
        <div class="setting-control">
          <div class="radio-group">
            <label class="radio-label">
              <input
                type="radio"
                name="proxyMode"
                value="rule"
                :checked="settingsStore.settings.proxyMode === 'rule'"
                @change="handleProxyModeChange('rule')"
              />
              <span>规则模式</span>
            </label>
            <label class="radio-label">
              <input
                type="radio"
                name="proxyMode"
                value="global"
                :checked="settingsStore.settings.proxyMode === 'global'"
                @change="handleProxyModeChange('global')"
              />
              <span>全局模式</span>
            </label>
            <label class="radio-label">
              <input
                type="radio"
                name="proxyMode"
                value="direct"
                :checked="settingsStore.settings.proxyMode === 'direct'"
                @change="handleProxyModeChange('direct')"
              />
              <span>直连模式</span>
            </label>
          </div>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>开机自启</h4>
          <p>系统启动时自动运行</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.autoStart"
              @change="updateSetting('autoStart', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <!-- 高级网络设置 -->
      <h3 class="section-title">高级网络设置</h3>

      <div class="setting-item">
        <div class="setting-label">
          <h4>启用 Mux</h4>
          <p>多路复用，提升并发性能</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.enableMux"
              @change="updateSetting('enableMux', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>IPv6 支持</h4>
          <p>启用 IPv6 代理</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.enableIpv6"
              @change="updateSetting('enableIpv6', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>TUN 模式</h4>
          <p>启用 TUN 虚拟网卡（需要管理员权限）</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.tunMode"
              @change="updateSetting('tunMode', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>域名策略</h4>
          <p>DNS 解析策略</p>
        </div>
        <div class="setting-control">
          <select
            :value="settingsStore.settings.domainStrategy"
            @change="updateSetting('domainStrategy', ($event.target as HTMLSelectElement).value)"
          >
            <option value="IPIfNonMatch">IPIfNonMatch</option>
            <option value="IPOnDemand">IPOnDemand</option>
            <option value="always">always</option>
            <option value="false">false</option>
          </select>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>DNS 服务器</h4>
          <p>自定义 DNS 服务器（每行一个）</p>
        </div>
        <div class="setting-control">
          <textarea
            class="dns-input"
            :value="settingsStore.settings.dnsServers?.join('\n')"
            @blur="(e) => updateSetting('dnsServers', ($event.target as HTMLTextAreaElement).value.split('\n').filter(s => s.trim()))"
            placeholder="https://dns.google/dns-query&#10;1.1.1.1&#10;8.8.8.8"
            rows="3"
          ></textarea>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>静默启动</h4>
          <p>启动时不显示主窗口</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.silentStart"
              @change="updateSetting('silentStart', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>
    </div>

    <!-- 网络设置 -->
    <div v-if="activeTab === 'network'" class="settings-section">
      <div class="setting-item">
        <div class="setting-label">
          <h4>允许局域网连接</h4>
          <p>允许局域网内的设备连接代理</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.allowLan"
              @change="updateSetting('allowLan', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>绑定地址</h4>
          <p>代理服务绑定的IP地址</p>
        </div>
        <div class="setting-control">
          <input
            type="text"
            :value="settingsStore.settings.bindAddress"
            @blur="updateSetting('bindAddress', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>HTTP端口</h4>
          <p>HTTP代理端口</p>
        </div>
        <div class="setting-control">
          <input
            type="number"
            :value="settingsStore.settings.httpPort"
            @blur="updateSetting('httpPort', parseInt(($event.target as HTMLInputElement).value))"
          />
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>SOCKS5端口</h4>
          <p>SOCKS5代理端口</p>
        </div>
        <div class="setting-control">
          <input
            type="number"
            :value="settingsStore.settings.socksPort"
            @blur="updateSetting('socksPort', parseInt(($event.target as HTMLInputElement).value))"
          />
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>混合端口</h4>
          <p>混合代理端口（HTTP+SOCKS5）</p>
        </div>
        <div class="setting-control">
          <input
            type="number"
            :value="settingsStore.settings.mixedPort"
            @blur="updateSetting('mixedPort', parseInt(($event.target as HTMLInputElement).value))"
          />
        </div>
      </div>
    </div>

    <!-- 高级设置 -->
    <div v-if="activeTab === 'advanced'" class="settings-section">
      <div class="setting-item">
        <div class="setting-label">
          <h4>日志级别</h4>
          <p>设置日志输出级别</p>
        </div>
        <div class="setting-control">
          <select
            :value="settingsStore.settings.logLevel"
            @change="updateSetting('logLevel', ($event.target as HTMLSelectElement).value)"
          >
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warning">Warning</option>
            <option value="error">Error</option>
            <option value="silent">Silent</option>
          </select>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>连接统计</h4>
          <p>启用连接统计功能</p>
        </div>
        <div class="setting-control">
          <label class="switch">
            <input
              type="checkbox"
              :checked="settingsStore.settings.connectionStats"
              @change="updateSetting('connectionStats', ($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>系统信息</h4>
        </div>
        <div class="setting-control">
          <div v-if="settingsStore.systemInfo" class="system-info">
            <div class="info-row">
              <span class="label">版本:</span>
              <span class="value">{{ settingsStore.systemInfo.version }}</span>
            </div>
            <div class="info-row">
              <span class="label">运行时间:</span>
              <span class="value">{{ formatUptime(settingsStore.systemInfo.uptime) }}</span>
            </div>
            <div class="info-row">
              <span class="label">操作系统:</span>
              <span class="value">{{ settingsStore.systemInfo.os }}</span>
            </div>
            <div class="info-row">
              <span class="label">架构:</span>
              <span class="value">{{ settingsStore.systemInfo.arch }}</span>
            </div>
            <div class="info-row">
              <span class="label">内存使用:</span>
              <span class="value">
                {{ formatMemory(settingsStore.systemInfo.memory.used) }} /
                {{ formatMemory(settingsStore.systemInfo.memory.total) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-label">
          <h4>配置管理</h4>
        </div>
        <div class="setting-control">
          <div class="button-group">
            <button class="btn" @click="handleExportConfig">导出配置</button>
            <label class="btn">
              导入配置
              <input type="file" accept=".json" @change="handleImportConfig" hidden />
            </label>
            <button class="btn btn-danger" @click="handleClearCache">清除缓存</button>
            <button class="btn btn-danger" @click="handleRestart">重启服务</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.settings {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.tabs {
  display: flex;
  gap: $spacing-xs;
  border-bottom: 1px solid $border-color;
}

.tab {
  padding: $spacing-md $spacing-lg;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: $text-color-secondary;
  font-size: $font-size-base;
  cursor: pointer;
  transition: all $transition-duration $transition-timing;

  &:hover {
    color: $text-color-primary;
  }

  &.active {
    color: $primary-color;
    border-bottom-color: $primary-color;
  }
}

.settings-section {
  background-color: $bg-color-light;
  border-radius: $border-radius-lg;
  border: 1px solid $border-color;
  overflow: hidden;
}

.setting-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: $spacing-lg;
  border-bottom: 1px solid $border-color;

  &:last-child {
    border-bottom: none;
  }
}

.setting-label {
  flex: 1;

  h4 {
    font-size: $font-size-base;
    font-weight: 500;
    color: $text-color-primary;
    margin-bottom: $spacing-xs;
  }

  p {
    font-size: $font-size-sm;
    color: $text-color-secondary;
  }
}

.setting-control {
  flex-shrink: 0;
  margin-left: $spacing-lg;
}

input[type='text'],
input[type='number'],
select {
  padding: $spacing-sm $spacing-md;
  background-color: $bg-color-darker;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  color: $text-color-primary;
  font-size: $font-size-base;
  min-width: 200px;

  &:focus {
    border-color: $primary-color;
    outline: none;
  }
}

.radio-group {
  display: flex;
  gap: $spacing-lg;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  cursor: pointer;
  font-size: $font-size-sm;
  color: $text-color-primary;

  input[type='radio'] {
    width: auto;
  }
}

.switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 24px;

  input {
    opacity: 0;
    width: 0;
    height: 0;

    &:checked + .slider {
      background-color: $primary-color;

      &:before {
        transform: translateX(24px);
      }
    }
  }

  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: $bg-color-lighter;
    transition: $transition-duration;
    border-radius: 24px;

    &:before {
      position: absolute;
      content: '';
      height: 18px;
      width: 18px;
      left: 3px;
      bottom: 3px;
      background-color: white;
      transition: $transition-duration;
      border-radius: 50%;
    }
  }
}

.system-info {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  min-width: 300px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  font-size: $font-size-sm;

  .label {
    color: $text-color-secondary;
  }

  .value {
    color: $text-color-primary;
    font-weight: 500;
  }
}

.button-group {
  display: flex;
  gap: $spacing-sm;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: $spacing-sm $spacing-md;
  background-color: $bg-color-darker;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  color: $text-color-primary;
  font-size: $font-size-sm;
  cursor: pointer;
  transition: all $transition-duration $transition-timing;

  &:hover {
    background-color: $bg-color-lighter;
  }

  &.btn-danger {
    color: $error-color;
    border-color: $error-color;

    &:hover {
      background-color: rgba($error-color, 0.1);
    }
  }
}

// 响应式
@media (max-width: $breakpoint-md) {
  .setting-item {
    flex-direction: column;
    gap: $spacing-md;
  }

  .setting-control {
    margin-left: 0;
    width: 100%;
  }

  input[type='text'],
  input[type='number'],
  select {
    width: 100%;
  }

  .radio-group {
    flex-direction: column;
    gap: $spacing-sm;
  }

  .button-group {
    flex-direction: column;
  }
}
</style>
