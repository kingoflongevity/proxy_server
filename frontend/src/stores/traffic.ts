import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  getTrafficLogs,
  getLogStats,
  clearLogs,
  getTrafficSummary,
  type TrafficLog,
  type LogStats
} from '@/api/traffic'

export const useLogStore = defineStore('logs', () => {
  const logs = ref<TrafficLog[]>([])
  const stats = ref<LogStats | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchLogs(params?: {
    start_time?: string
    end_time?: string
    client_ip?: string
    method?: string
    status_code?: number
    url?: string
    keyword?: string
    limit?: number
    offset?: number
  }) {
    loading.value = true
    error.value = null
    try {
      const result = await getTrafficLogs(params)
      logs.value = result.logs
      total.value = result.total
    } catch (e: any) {
      error.value = e.message || '获取日志失败'
      console.error('获取日志失败:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchStats(params?: {
    start_time?: string
    end_time?: string
  }) {
    try {
      const result = await getLogStats(params)
      stats.value = result
    } catch (e: any) {
      console.error('获取统计失败:', e)
    }
  }

  async function fetchSummary(params?: {
    start_time?: string
    end_time?: string
  }) {
    try {
      const result = await getTrafficSummary(params)
      stats.value = result
    } catch (e: any) {
      console.error('获取摘要失败:', e)
    }
  }

  async function clearOldLogs(before?: string) {
    loading.value = true
    try {
      await clearLogs(before)
      await fetchLogs()
    } catch (e: any) {
      error.value = e.message || '清理日志失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    logs,
    stats,
    total,
    loading,
    error,
    fetchLogs,
    fetchStats,
    fetchSummary,
    clearOldLogs
  }
})
