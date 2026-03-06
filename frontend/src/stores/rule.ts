import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProxyRule, CreateRuleRequest, UpdateRuleRequest, RuleSet } from '@/types'
import * as ruleApi from '@/api/rule'

/**
 * 规则管理Store
 */
export const useRuleStore = defineStore('rule', () => {
  // 状态
  const rules = ref<ProxyRule[]>([])
  const ruleSets = ref<RuleSet[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const enabledRules = computed(() => rules.value.filter((rule) => rule.enabled))

  const ruleCount = computed(() => rules.value.length)

  const enabledRuleCount = computed(() => enabledRules.value.length)

  const rulesByType = computed(() => {
    const grouped: Record<string, ProxyRule[]> = {}
    rules.value.forEach((rule) => {
      if (!grouped[rule.type]) {
        grouped[rule.type] = []
      }
      grouped[rule.type].push(rule)
    })
    return grouped
  })

  // Actions

  /**
   * 获取所有规则
   */
  async function fetchRules() {
    loading.value = true
    error.value = null
    try {
      rules.value = await ruleApi.getRules()
    } catch (e: any) {
      error.value = e.message || '获取规则列表失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取单个规则
   */
  async function fetchRule(id: string) {
    loading.value = true
    error.value = null
    try {
      return await ruleApi.getRule(id)
    } catch (e: any) {
      error.value = e.message || '获取规则详情失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建规则
   */
  async function createRule(data: CreateRuleRequest) {
    loading.value = true
    error.value = null
    try {
      const rule = await ruleApi.createRule(data)
      rules.value.push(rule)
      // 按优先级排序
      rules.value.sort((a, b) => b.priority - a.priority)
      return rule
    } catch (e: any) {
      error.value = e.message || '创建规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新规则
   */
  async function updateRule(id: string, data: UpdateRuleRequest) {
    loading.value = true
    error.value = null
    try {
      const rule = await ruleApi.updateRule(id, data)
      const index = rules.value.findIndex((r) => r.id === id)
      if (index !== -1) {
        rules.value[index] = rule
      }
      // 按优先级排序
      rules.value.sort((a, b) => b.priority - a.priority)
      return rule
    } catch (e: any) {
      error.value = e.message || '更新规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除规则
   */
  async function deleteRule(id: string) {
    loading.value = true
    error.value = null
    try {
      await ruleApi.deleteRule(id)
      rules.value = rules.value.filter((r) => r.id !== id)
    } catch (e: any) {
      error.value = e.message || '删除规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量更新规则优先级
   */
  async function updateRulesPriority(rulePriorities: Array<{ id: string; priority: number }>) {
    loading.value = true
    error.value = null
    try {
      await ruleApi.updateRulesPriority(rulePriorities)
      // 更新本地状态
      rulePriorities.forEach(({ id, priority }) => {
        const rule = rules.value.find((r) => r.id === id)
        if (rule) {
          rule.priority = priority
        }
      })
      // 按优先级排序
      rules.value.sort((a, b) => b.priority - a.priority)
    } catch (e: any) {
      error.value = e.message || '更新规则优先级失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 启用/禁用规则
   */
  async function toggleRule(id: string, enabled: boolean) {
    loading.value = true
    error.value = null
    try {
      await ruleApi.toggleRule(id, enabled)
      const rule = rules.value.find((r) => r.id === id)
      if (rule) {
        rule.enabled = enabled
      }
    } catch (e: any) {
      error.value = e.message || '切换规则状态失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取规则集列表
   */
  async function fetchRuleSets() {
    loading.value = true
    error.value = null
    try {
      ruleSets.value = await ruleApi.getRuleSets()
    } catch (e: any) {
      error.value = e.message || '获取规则集列表失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新规则集
   */
  async function updateRuleSet(id: string) {
    loading.value = true
    error.value = null
    try {
      const result = await ruleApi.updateRuleSet(id)
      // 更新规则集信息
      const ruleSet = ruleSets.value.find((rs) => rs.id === id)
      if (ruleSet) {
        ruleSet.lastUpdate = new Date().toISOString()
        ruleSet.ruleCount = result.count
      }
      return result
    } catch (e: any) {
      error.value = e.message || '更新规则集失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 清除错误
   */
  function clearError() {
    error.value = null
  }

  return {
    // 状态
    rules,
    ruleSets,
    loading,
    error,
    // 计算属性
    enabledRules,
    ruleCount,
    enabledRuleCount,
    rulesByType,
    // Actions
    fetchRules,
    fetchRule,
    createRule,
    updateRule,
    deleteRule,
    updateRulesPriority,
    toggleRule,
    fetchRuleSets,
    updateRuleSet,
    clearError,
  }
})
