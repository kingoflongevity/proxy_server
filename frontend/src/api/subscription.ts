import { request } from './request'
import type {
  Subscription,
  CreateSubscriptionRequest,
  UpdateSubscriptionRequest,
} from '@/types'

/**
 * 订阅管理API
 */

/**
 * 获取所有订阅
 */
export function getSubscriptions(): Promise<Subscription[]> {
  return request.get('/subscriptions')
}

/**
 * 获取单个订阅
 */
export function getSubscription(id: string): Promise<Subscription> {
  return request.get(`/subscriptions/${id}`)
}

/**
 * 创建订阅
 */
export function createSubscription(data: CreateSubscriptionRequest): Promise<Subscription> {
  return request.post('/subscriptions', data)
}

/**
 * 更新订阅
 */
export function updateSubscription(
  id: string,
  data: UpdateSubscriptionRequest
): Promise<Subscription> {
  return request.put(`/subscriptions/${id}`, data)
}

/**
 * 删除订阅
 */
export function deleteSubscription(id: string): Promise<void> {
  return request.delete(`/subscriptions/${id}`)
}

/**
 * 更新订阅节点
 */
export function updateSubscriptionNodes(id: string): Promise<{ count: number }> {
  return request.post(`/subscriptions/${id}/update`)
}

/**
 * 测试订阅连接
 */
export function testSubscription(id: string): Promise<{ valid: boolean; message: string }> {
  return request.post(`/subscriptions/${id}/test`)
}
