import { request } from './request'
import type { ProxyNode, NodeFilter, NodeSort, NodeTestResult } from '@/types'

/**
 * 节点管理API
 */

/**
 * 获取节点列表
 */
export function getNodes(params?: {
  filter?: NodeFilter
  sort?: NodeSort
  page?: number
  pageSize?: number
}): Promise<{
  items: ProxyNode[]
  total: number
}> {
  return request.get('/nodes', { params })
}

/**
 * 获取单个节点
 */
export function getNode(id: string): Promise<ProxyNode> {
  return request.get(`/nodes/${id}`)
}

/**
 * 测试节点延迟
 */
export function testNodeLatency(id: string): Promise<NodeTestResult> {
  return request.post(`/nodes/${id}/test`)
}

/**
 * 批量测试节点延迟
 */
export function testNodesLatency(ids: string[]): Promise<NodeTestResult[]> {
  return request.post('/nodes/test', { ids })
}

/**
 * 测试所有节点
 */
export function testAllNodes(): Promise<{ taskId: string }> {
  return request.post('/nodes/test-all')
}

/**
 * 选择节点
 */
export function selectNode(id: string): Promise<void> {
  return request.post(`/nodes/${id}/select`)
}

/**
 * 获取当前选中节点
 */
export function getCurrentNode(): Promise<ProxyNode | null> {
  return request.get('/nodes/current')
}

/**
 * 获取节点统计信息
 */
export function getNodeStats(id: string): Promise<{
  uploadSpeed: number
  downloadSpeed: number
  uploadTotal: number
  downloadTotal: number
  connectionCount: number
}> {
  return request.get(`/nodes/${id}/stats`)
}
