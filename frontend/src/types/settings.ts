/**
 * 系统设置类型定义
 */
export interface SystemSettings {
  theme: Theme
  language: Language
  autoStart: boolean
  silentStart: boolean
  allowLan: boolean
  bindAddress: string
  port: number
  socksPort: number
  httpPort: number
  mixedPort: number
  logLevel: LogLevel
  connectionStats: boolean
  proxyMode: ProxyMode
}

export type Theme = 'dark' | 'light' | 'auto'

export type Language = 'zh-CN' | 'en-US'

export type LogLevel = 'debug' | 'info' | 'warning' | 'error' | 'silent'

export type ProxyMode = 'rule' | 'global' | 'direct'

/**
 * 连接状态
 */
export interface ConnectionStatus {
  connected: boolean
  currentNode?: string
  currentMode: ProxyMode
  uploadSpeed: number
  downloadSpeed: number
  uploadTotal: number
  downloadTotal: number
  connectionCount: number
}

/**
 * 系统信息
 */
export interface SystemInfo {
  version: string
  uptime: number
  os: string
  arch: string
  goVersion: string
  memory: {
    used: number
    total: number
  }
}

/**
 * 更新设置请求
 */
export interface UpdateSettingsRequest {
  theme?: Theme
  language?: Language
  autoStart?: boolean
  silentStart?: boolean
  allowLan?: boolean
  bindAddress?: string
  port?: number
  socksPort?: number
  httpPort?: number
  mixedPort?: number
  logLevel?: LogLevel
  connectionStats?: boolean
  proxyMode?: ProxyMode
}
