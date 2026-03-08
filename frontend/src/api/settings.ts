import { request } from './request'
import type {
  SystemSettings,
  UpdateSettingsRequest,
  ConnectionStatus,
  SystemInfo,
} from '@/types'

/**
 * 系统设置API
 */

/**
 * 获取系统设置
 */
export function getSettings(): Promise<SystemSettings> {
  return request.get('/settings')
}

/**
 * 更新系统设置
 */
export function updateSettings(data: UpdateSettingsRequest): Promise<SystemSettings> {
  return request.put('/settings', data)
}

/**
 * 获取连接状态
 */
export function getConnectionStatus(): Promise<ConnectionStatus> {
  return request.get('/connection/status')
}

/**
 * 获取系统信息
 */
export function getSystemInfo(): Promise<SystemInfo> {
  return request.get('/system/info')
}

/**
 * 获取代理模式
 */
export function getProxyMode(): Promise<{ proxyMode: string }> {
  return request.get('/proxy/mode')
}

/**
 * 设置代理模式
 */
export function setProxyMode(mode: string): Promise<{ proxyMode: string }> {
  return request.put('/proxy/mode', { mode })
}

/**
 * 启用系统代理
 */
export function enableSystemProxy(): Promise<{ enabled: boolean }> {
  return request.post('/system-proxy/enable')
}

/**
 * 禁用系统代理
 */
export function disableSystemProxy(): Promise<{ enabled: boolean }> {
  return request.post('/system-proxy/disable')
}

/**
 * 获取系统代理状态
 */
export function getSystemProxyStatus(): Promise<{
  enabled: boolean
  server: string
  port: number
  bypass: string
}> {
  return request.get('/system-proxy/status')
}

/**
 * 重启服务
 */
export function restartService(): Promise<void> {
  return request.post('/system/restart')
}

/**
 * 导出配置
 */
export function exportConfig(): Promise<string> {
  return request.get('/config/export')
}

/**
 * 导入配置
 */
export function importConfig(config: string): Promise<void> {
  return request.post('/config/import', { config })
}

/**
 * 清除缓存
 */
export function clearCache(): Promise<void> {
  return request.post('/system/clear-cache')
}
