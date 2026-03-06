/**
 * 代理规则类型定义
 */
export interface ProxyRule {
  id: string
  type: RuleType
  pattern: string
  action: RuleAction
  target?: string // 代理节点ID或代理组名称
  enabled: boolean
  priority: number
  description?: string
  createdAt: string
  updatedAt: string
}

export type RuleType =
  | 'DOMAIN' // 域名完全匹配
  | 'DOMAIN-SUFFIX' // 域名后缀匹配
  | 'DOMAIN-KEYWORD' // 域名关键词匹配
  | 'IP-CIDR' // IP段匹配
  | 'SRC-IP-CIDR' // 源IP段匹配
  | 'GEOIP' // 地理位置IP匹配
  | 'DST-PORT' // 目标端口匹配
  | 'SRC-PORT' // 源端口匹配
  | 'PROCESS-NAME' // 进程名匹配
  | 'RULE-SET' // 规则集
  | 'MATCH' // 匹配所有

export type RuleAction = 'DIRECT' | 'REJECT' | 'PROXY' | 'node'

/**
 * 创建规则请求
 */
export interface CreateRuleRequest {
  type: RuleType
  pattern: string
  action: RuleAction
  target?: string
  enabled?: boolean
  priority?: number
  description?: string
}

/**
 * 更新规则请求
 */
export interface UpdateRuleRequest {
  type?: RuleType
  pattern?: string
  action?: RuleAction
  target?: string
  enabled?: boolean
  priority?: number
  description?: string
}

/**
 * 规则集
 */
export interface RuleSet {
  id: string
  name: string
  type: 'remote' | 'local'
  url?: string
  path?: string
  ruleCount: number
  lastUpdate: string
  autoUpdate: boolean
  updateInterval: number
}
