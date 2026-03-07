import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProxyNode, NodeFilter, NodeSort } from '@/types'
import * as nodeApi from '@/api/node'

/**
 * 节点管理Store
 */
export const useNodeStore = defineStore('node', () => {
  // 状态
  const nodes = ref<ProxyNode[]>([])
  const currentNode = ref<ProxyNode | null>(null)
  const loading = ref(false)
  const testing = ref(false)
  const error = ref<string | null>(null)
  const filter = ref<NodeFilter>({})
  const sort = ref<NodeSort>({
    field: 'latency',
    order: 'asc',
  })

  // 计算属性
  const filteredNodes = computed(() => {
    let result = [...nodes.value]

    // 关键词筛选
    if (filter.value.keyword) {
      const keyword = filter.value.keyword.toLowerCase()
      result = result.filter(
        (node) =>
          node.name.toLowerCase().includes(keyword) ||
          node.server.toLowerCase().includes(keyword)
      )
    }

    // 类型筛选
    if (filter.value.type && filter.value.type.length > 0) {
      result = result.filter((node) => filter.value.type!.includes(node.type))
    }

    // 状态筛选
    if (filter.value.status && filter.value.status.length > 0) {
      result = result.filter((node) => filter.value.status!.includes(node.status))
    }

    // 国家筛选
    if (filter.value.country && filter.value.country.length > 0) {
      result = result.filter((node) => filter.value.country!.includes(node.country || ''))
    }

    // 延迟范围筛选
    if (filter.value.latencyRange) {
      const { min, max } = filter.value.latencyRange
      result = result.filter((node) => node.latency >= min && node.latency <= max)
    }

    // 排序
    result.sort((a, b) => {
      const field = sort.value.field
      const order = sort.value.order === 'asc' ? 1 : -1

      if (field === 'name') {
        return a.name.localeCompare(b.name) * order
      } else if (field === 'latency') {
        return (a.latency - b.latency) * order
      } else if (field === 'uploadSpeed') {
        return (a.uploadSpeed - b.uploadSpeed) * order
      } else if (field === 'downloadSpeed') {
        return (a.downloadSpeed - b.downloadSpeed) * order
      } else if (field === 'lastTest') {
        return (
          (new Date(a.lastTest).getTime() - new Date(b.lastTest).getTime()) * order
        )
      }
      return 0
    })

    return result
  })

  const availableNodes = computed(() =>
    filteredNodes.value.filter((node) => node.status === 'available')
  )

  const nodeCount = computed(() => nodes.value.length)

  const availableCount = computed(() =>
    nodes.value.filter((node) => node.status === 'available').length
  )

  // Actions

  /**
   * 获取节点列表
   */
  async function fetchNodes() {
    loading.value = true
    error.value = null
    try {
      const result = await nodeApi.getNodes()
      nodes.value = result.items
    } catch (e: any) {
      error.value = e.message || '获取节点列表失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取单个节点
   */
  async function fetchNode(id: string) {
    loading.value = true
    error.value = null
    try {
      return await nodeApi.getNode(id)
    } catch (e: any) {
      error.value = e.message || '获取节点详情失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 测试节点延迟
   */
  async function testNodeLatency(id: string) {
    testing.value = true
    error.value = null
    try {
      const result = await nodeApi.testNodeLatency(id)
      // 更新节点信息
      const index = nodes.value.findIndex((n) => n.id === id)
      if (index !== -1) {
        nodes.value[index].latency = result.latency
        nodes.value[index].status = result.status
        nodes.value[index].lastTest = result.testTime
      }
      return result
    } catch (e: any) {
      error.value = e.message || '测试节点延迟失败'
      throw e
    } finally {
      testing.value = false
    }
  }

  /**
   * 批量测试节点延迟
   */
  async function testNodesLatency(ids: string[]) {
    testing.value = true
    error.value = null
    try {
      const results = await nodeApi.testNodesLatency(ids)
      // 批量更新节点信息
      results.forEach((result) => {
        const index = nodes.value.findIndex((n) => n.id === result.nodeId)
        if (index !== -1) {
          nodes.value[index].latency = result.latency
          nodes.value[index].status = result.status
          nodes.value[index].lastTest = result.testTime
        }
      })
      return results
    } catch (e: any) {
      error.value = e.message || '批量测试节点延迟失败'
      throw e
    } finally {
      testing.value = false
    }
  }

  /**
   * 测试所有节点
   */
  async function testAllNodes() {
    testing.value = true
    error.value = null
    try {
      const ids = nodes.value.map((n) => n.id)
      return await testNodesLatency(ids)
    } catch (e: any) {
      error.value = e.message || '测试所有节点失败'
      throw e
    } finally {
      testing.value = false
    }
  }

  /**
   * 选择节点
   */
  async function selectNode(id: string) {
    loading.value = true
    error.value = null
    try {
      await nodeApi.selectNode(id)
      const node = nodes.value.find((n) => n.id === id)
      if (node) {
        currentNode.value = node
      }
    } catch (e: any) {
      error.value = e.message || '选择节点失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取当前选中节点
   */
  async function fetchCurrentNode() {
    loading.value = true
    error.value = null
    try {
      currentNode.value = await nodeApi.getCurrentNode()
      return currentNode.value
    } catch (e: any) {
      error.value = e.message || '获取当前节点失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * 设置筛选条件
   */
  function setFilter(newFilter: Partial<NodeFilter>) {
    filter.value = { ...filter.value, ...newFilter }
  }

  /**
   * 设置排序
   */
  function setSort(newSort: Partial<NodeSort>) {
    sort.value = { ...sort.value, ...newSort }
  }

  /**
   * 重置筛选和排序
   */
  function resetFilterAndSort() {
    filter.value = {}
    sort.value = { field: 'latency', order: 'asc' }
  }

  /**
   * 清除错误
   */
  function clearError() {
    error.value = null
  }

  return {
    // 状态
    nodes,
    currentNode,
    loading,
    testing,
    error,
    filter,
    sort,
    // 计算属性
    filteredNodes,
    availableNodes,
    nodeCount,
    availableCount,
    // Actions
    fetchNodes,
    fetchNode,
    testNodeLatency,
    testNodesLatency,
    testAllNodes,
    selectNode,
    fetchCurrentNode,
    setFilter,
    setSort,
    resetFilterAndSort,
    clearError,
  }
})
