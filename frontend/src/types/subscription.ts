/**
 * 订阅类型定义
 */
export interface Subscription {
  id: string
  name: string
  url: string
  type: SubscriptionType
  parseFormat: ParseFormat
  autoUpdate: boolean
  updateInterval: number // 更新间隔（小时）
  lastUpdate: string
  nodeCount: number
  status: SubscriptionStatus
  createdAt: string
  updatedAt: string
}

export type SubscriptionType = 'ss' | 'ssr' | 'vmess' | 'vless' | 'trojan' | 'hysteria' | 'mixed'

export type SubscriptionStatus = 'active' | 'inactive' | 'updating' | 'error'

/**
 * 解析格式类型
 */
export type ParseFormat = 'auto' | 'base64' | 'clash' | 'surge' | 'quantumult' | 'ssd'

/**
 * 解析格式选项
 */
export const PARSE_FORMAT_OPTIONS: { value: ParseFormat; label: string; description: string }[] = [
  { value: 'auto', label: '自动检测', description: '自动识别订阅格式，推荐使用' },
  { value: 'base64', label: 'Base64', description: 'Base64编码的节点链接列表（v2rayN格式）' },
  { value: 'clash', label: 'Clash', description: 'Clash/Mihomo配置文件格式（YAML）' },
  { value: 'surge', label: 'Surge', description: 'Surge配置文件格式' },
  { value: 'quantumult', label: 'Quantumult', description: 'Quantumult/Quantumult X格式' },
  { value: 'ssd', label: 'SSD', description: 'SSD订阅格式（ShadowsocksD）' },
]

/**
 * 创建订阅请求
 */
export interface CreateSubscriptionRequest {
  name: string
  url: string
  type: SubscriptionType
  parseFormat?: ParseFormat
  autoUpdate?: boolean
  updateInterval?: number
}

/**
 * 更新订阅请求
 */
export interface UpdateSubscriptionRequest {
  name?: string
  url?: string
  type?: SubscriptionType
  parseFormat?: ParseFormat
  autoUpdate?: boolean
  updateInterval?: number
}
