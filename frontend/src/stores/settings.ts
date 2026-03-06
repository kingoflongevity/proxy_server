import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type {
  SystemSettings,
  UpdateSettingsRequest,
  ConnectionStatus,
  SystemInfo,
  Theme,
  Language,
} from '@/types'
import * as settingsApi from '@/api/settings'

/**
 * 系统设置Store
 */
export const useSettingsStore = defineStore('settings', () => {
  // 状态
  const settings = ref<SystemSettings>({
    theme: 'dark',
    language: 'zh-CN',
    autoStart: false,
    silentStart: false,
    allowLan: true,
    bindAddress: '0.0.0.0',
    port: 7890,
    socksPort: 10808,
    httpPort: 10809,
    mixedPort: 10810,
    logLevel: 'info',
    connectionStats: true,
    proxyMode: 'rule',
    // 高级设置默认值
    dnsServers: ['https://dns.google/dns-query', '1.1.1.1', '8.8.8.8'],
    enableMux: false,
    enableIpv6: false,
    domainStrategy: 'IPIfNonMatch',
    tunMode: false,
  })

  const connectionStatus = ref<ConnectionStatus>({
    connected: false,
    currentMode: 'rule',
    uploadSpeed: 0,
    downloadSpeed: 0,
    uploadTotal: 0,
    downloadTotal: 0,
    connectionCount: 0,
  })

  const systemInfo = ref<SystemInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const theme = computed(() => settings.value.theme)

  const language = computed(() => settings.value.language)

  const proxyMode = computed(() => settings.value.proxyMode)

  const isConnected = computed(() => connectionStatus.value.connected)

  // 监听主题变化
  watch(
    () => settings.value.theme,
    (newTheme) => {
      applyTheme(newTheme)
    }
  )

  // 监听语言变化
  watch(
    () => settings.value.language,
    (newLanguage) => {
      applyLanguage(newLanguage)
    }
  )

  // Actions

  /**
   * 获取系统设置
   */
  async function fetchSettings() {
    loading.value = true
    error.value = null
    try {
      const result = await settingsApi.getSettings()
      // 合并后端返回的数据和默认值，确保所有字段都有值
      settings.value = {
        theme: result.theme || 'dark',
        language: result.language || 'zh-CN',
        autoStart: result.autoStart ?? false,
        silentStart: result.silentStart ?? false,
        allowLan: result.allowLan ?? true,
        bindAddress: result.bindAddress || '0.0.0.0',
        port: result.port || 7890,
        socksPort: result.socksPort || 10808,
        httpPort: result.httpPort || 10809,
        mixedPort: result.mixedPort || 10810,
        logLevel: result.logLevel || 'info',
        connectionStats: result.connectionStats ?? true,
        proxyMode: result.proxyMode || 'rule',
        // 高级设置
        dnsServers: result.dnsServers || ['https://dns.google/dns-query', '1.1.1.1', '8.8.8.8'],
        enableMux: result.enableMux ?? false,
        enableIpv6: result.enableIpv6 ?? false,
        domainStrategy: result.domainStrategy || 'IPIfNonMatch',
        tunMode: result.tunMode ?? false,
      }
      // 应用主题和语言
      applyTheme(settings.value.theme)
      applyLanguage(settings.value.language)
    } catch (e: any) {
      error.value = e.message || '获取系统设置失败'
      // 即使API失败，也应用本地默认值
      applyTheme(settings.value.theme)
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新系统设置
   */
  async function updateSettings(data: UpdateSettingsRequest) {
    loading.value = true
    error.value = null
    try {
      // 先立即应用主题变化到本地状态，提供即时反馈
      if (data.theme) {
        settings.value.theme = data.theme
        applyTheme(data.theme)
      }
      if (data.language) {
        settings.value.language = data.language
        applyLanguage(data.language)
      }
      if (data.proxyMode) {
        settings.value.proxyMode = data.proxyMode
      }
      
      // 然后发送API请求
      const result = await settingsApi.updateSettings(data)
      // 更新本地状态
      settings.value = { ...settings.value, ...result }
      return result
    } catch (e: any) {
      error.value = e.message || '更新系统设置失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取连接状态
   */
  async function fetchConnectionStatus() {
    try {
      const result = await settingsApi.getConnectionStatus()
      connectionStatus.value = result
      return result
    } catch (e: any) {
      console.error('获取连接状态失败:', e)
      throw e
    }
  }

  /**
   * 获取系统信息
   */
  async function fetchSystemInfo() {
    loading.value = true
    error.value = null
    try {
      systemInfo.value = await settingsApi.getSystemInfo()
      return systemInfo.value
    } catch (e: any) {
      error.value = e.message || '获取系统信息失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取当前代理模式（实时）
   */
  async function fetchProxyMode() {
    try {
      const result = await settingsApi.getProxyMode()
      if (result && result.proxyMode) {
        settings.value.proxyMode = result.proxyMode as any
      }
      return result
    } catch (e: any) {
      console.error('获取代理模式失败:', e)
    }
  }

  /**
   * 切换主题
   */
  async function toggleTheme(theme: Theme) {
    return updateSettings({ theme })
  }

  /**
   * 切换语言
   */
  async function toggleLanguage(language: Language) {
    return updateSettings({ language })
  }

  /**
   * 切换代理模式
   */
  async function toggleProxyMode(mode: SystemSettings['proxyMode']) {
    return updateSettings({ proxyMode: mode })
  }

  /**
   * 重启服务
   */
  async function restartService() {
    loading.value = true
    error.value = null
    try {
      await settingsApi.restartService()
    } catch (e: any) {
      error.value = e.message || '重启服务失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 导出配置
   */
  async function exportConfig() {
    loading.value = true
    error.value = null
    try {
      return await settingsApi.exportConfig()
    } catch (e: any) {
      error.value = e.message || '导出配置失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 导入配置
   */
  async function importConfig(config: string) {
    loading.value = true
    error.value = null
    try {
      await settingsApi.importConfig(config)
      // 重新加载设置
      await fetchSettings()
    } catch (e: any) {
      error.value = e.message || '导入配置失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 清除缓存
   */
  async function clearCache() {
    loading.value = true
    error.value = null
    try {
      await settingsApi.clearCache()
    } catch (e: any) {
      error.value = e.message || '清除缓存失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 应用主题
   */
  function applyTheme(theme: Theme) {
    const root = document.documentElement
    if (theme === 'auto') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
    } else {
      root.setAttribute('data-theme', theme)
    }
  }

  /**
   * 应用语言
   */
  function applyLanguage(language: Language) {
    // 这里可以集成i18n
    document.documentElement.setAttribute('lang', language)
  }

  /**
   * 清除错误
   */
  function clearError() {
    error.value = null
  }

  return {
    // 状态
    settings,
    connectionStatus,
    systemInfo,
    loading,
    error,
    // 计算属性
    theme,
    language,
    proxyMode,
    isConnected,
    // Actions
    fetchSettings,
    updateSettings,
    fetchConnectionStatus,
    fetchSystemInfo,
    fetchProxyMode,
    toggleTheme,
    toggleLanguage,
    toggleProxyMode,
    restartService,
    exportConfig,
    importConfig,
    clearCache,
    clearError,
  }
})
