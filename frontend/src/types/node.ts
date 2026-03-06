/**
 * 代理节点类型定义
 */
export interface ProxyNode {
  id: string
  name: string
  type: ProxyType
  server: string
  port: number
  country?: string
  region?: string
  latency: number // 延迟（毫秒）
  status: NodeStatus
  subscriptionId: string
  config: Record<string, any>
  lastTest: string
  uploadSpeed: number // 上传速度 (bytes/s)
  downloadSpeed: number // 下载速度 (bytes/s)
}

export type ProxyType = 'ss' | 'ssr' | 'vmess' | 'vless' | 'trojan' | 'hysteria' | 'http' | 'socks5'

export type NodeStatus = 'available' | 'unavailable' | 'testing' | 'unknown'

/**
 * 节点筛选条件
 */
export interface NodeFilter {
  keyword?: string
  type?: ProxyType[]
  status?: NodeStatus[]
  country?: string[]
  subscriptionId?: string
  latencyRange?: {
    min: number
    max: number
  }
}

/**
 * 节点排序
 */
export type NodeSortField = 'name' | 'latency' | 'uploadSpeed' | 'downloadSpeed' | 'lastTest'

export type NodeSortOrder = 'asc' | 'desc'

export interface NodeSort {
  field: NodeSortField
  order: NodeSortOrder
}

/**
 * 节点测试结果
 */
export interface NodeTestResult {
  nodeId: string
  latency: number
  status: NodeStatus
  testTime: string
  error?: string
}
