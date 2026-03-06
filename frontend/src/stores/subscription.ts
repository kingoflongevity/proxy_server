import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Subscription, CreateSubscriptionRequest, UpdateSubscriptionRequest } from '@/types'
import * as subscriptionApi from '@/api/subscription'

/**
 * 订阅管理Store
 */
export const useSubscriptionStore = defineStore('subscription', () => {
  // 状态
  const subscriptions = ref<Subscription[]>([])
  const currentSubscription = ref<Subscription | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const activeSubscriptions = computed(() =>
    subscriptions.value.filter((sub) => sub.status === 'active')
  )

  const subscriptionCount = computed(() => subscriptions.value.length)

  const totalNodes = computed(() =>
    subscriptions.value.reduce((sum, sub) => sum + sub.nodeCount, 0)
  )

  // Actions

  /**
   * 获取所有订阅
   */
  async function fetchSubscriptions() {
    loading.value = true
    error.value = null
    try {
      subscriptions.value = await subscriptionApi.getSubscriptions() || []
    } catch (e: any) {
      error.value = e.message || '获取订阅列表失败'
      subscriptions.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取单个订阅
   */
  async function fetchSubscription(id: string) {
    loading.value = true
    error.value = null
    try {
      currentSubscription.value = await subscriptionApi.getSubscription(id)
      return currentSubscription.value
    } catch (e: any) {
      error.value = e.message || '获取订阅详情失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建订阅
   */
  async function createSubscription(data: CreateSubscriptionRequest) {
    loading.value = true
    error.value = null
    try {
      const subscription = await subscriptionApi.createSubscription(data)
      subscriptions.value.push(subscription)
      return subscription
    } catch (e: any) {
      error.value = e.message || '创建订阅失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新订阅
   */
  async function updateSubscription(id: string, data: UpdateSubscriptionRequest) {
    loading.value = true
    error.value = null
    try {
      const subscription = await subscriptionApi.updateSubscription(id, data)
      const index = subscriptions.value.findIndex((s) => s.id === id)
      if (index !== -1) {
        subscriptions.value[index] = subscription
      }
      if (currentSubscription.value?.id === id) {
        currentSubscription.value = subscription
      }
      return subscription
    } catch (e: any) {
      error.value = e.message || '更新订阅失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除订阅
   */
  async function deleteSubscription(id: string) {
    loading.value = true
    error.value = null
    try {
      await subscriptionApi.deleteSubscription(id)
      subscriptions.value = subscriptions.value.filter((s) => s.id !== id)
      if (currentSubscription.value?.id === id) {
        currentSubscription.value = null
      }
    } catch (e: any) {
      error.value = e.message || '删除订阅失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新订阅节点
   */
  async function updateSubscriptionNodes(id: string) {
    loading.value = true
    error.value = null
    try {
      const result = await subscriptionApi.updateSubscriptionNodes(id)
      // 更新订阅状态和节点数
      const index = subscriptions.value.findIndex((s) => s.id === id)
      if (index !== -1) {
        subscriptions.value[index].nodeCount = result.count
        subscriptions.value[index].lastUpdate = new Date().toISOString()
      }
      return result
    } catch (e: any) {
      error.value = e.message || '更新订阅节点失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 测试订阅连接
   */
  async function testSubscription(id: string) {
    loading.value = true
    error.value = null
    try {
      return await subscriptionApi.testSubscription(id)
    } catch (e: any) {
      error.value = e.message || '测试订阅连接失败'
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
    subscriptions,
    currentSubscription,
    loading,
    error,
    // 计算属性
    activeSubscriptions,
    subscriptionCount,
    totalNodes,
    // Actions
    fetchSubscriptions,
    fetchSubscription,
    createSubscription,
    updateSubscription,
    deleteSubscription,
    updateSubscriptionNodes,
    testSubscription,
    clearError,
  }
})
