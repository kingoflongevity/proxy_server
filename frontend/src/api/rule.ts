import { request } from './request'
import type { ProxyRule, CreateRuleRequest, UpdateRuleRequest, RuleSet } from '@/types'

/**
 * 规则管理API
 */

/**
 * 获取所有规则
 */
export function getRules(): Promise<ProxyRule[]> {
  return request.get('/rules')
}

/**
 * 获取单个规则
 */
export function getRule(id: string): Promise<ProxyRule> {
  return request.get(`/rules/${id}`)
}

/**
 * 创建规则
 */
export function createRule(data: CreateRuleRequest): Promise<ProxyRule> {
  return request.post('/rules', data)
}

/**
 * 更新规则
 */
export function updateRule(id: string, data: UpdateRuleRequest): Promise<ProxyRule> {
  return request.put(`/rules/${id}`, data)
}

/**
 * 删除规则
 */
export function deleteRule(id: string): Promise<void> {
  return request.delete(`/rules/${id}`)
}

/**
 * 批量更新规则优先级
 */
export function updateRulesPriority(
  rules: Array<{ id: string; priority: number }>
): Promise<void> {
  return request.put('/rules/priority', { rules })
}

/**
 * 启用/禁用规则
 */
export function toggleRule(id: string, enabled: boolean): Promise<void> {
  return request.patch(`/rules/${id}`, { enabled })
}

/**
 * 获取规则集列表
 */
export function getRuleSets(): Promise<RuleSet[]> {
  return request.get('/rule-sets')
}

/**
 * 更新规则集
 */
export function updateRuleSet(id: string): Promise<{ count: number }> {
  return request.post(`/rule-sets/${id}/update`)
}
