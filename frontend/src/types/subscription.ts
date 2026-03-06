/**
 * 订阅类型定义
 */
export interface Subscription {
  id: string
  name: string
  url: string
  type: SubscriptionType
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
 * 创建订阅请求
 */
export interface CreateSubscriptionRequest {
  name: string
  url: string
  type: SubscriptionType
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
  autoUpdate?: boolean
  updateInterval?: number
}
