<template>
  <div class="logs-page">
    <div class="page-header">
      <h1 class="page-title">日志管理</h1>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="handleRefresh">
          <svg class="icon" viewBox="0 0 24 24" fill="none">
            <path d="M4 12a8 8 0 0 1 8-8 8 8 0 0 1 8 8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M20 12a8 8 0 0 1-8 8 8 8 0 0 1-8-8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          刷新
        </button>
        <button class="btn btn-secondary" @click="handleExport">
          <svg class="icon" viewBox="0 0 24 24" fill="none">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          导出日志
        </button>
        <button class="btn btn-danger" @click="handleClearLogs">
          <svg class="icon" viewBox="0 0 24 24" fill="none">
            <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m3 0v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          清理日志
        </button>
      </div>
    </div>

    <!-- 主标签页 -->
    <div class="main-tabs">
      <button class="main-tab" :class="{ active: mainTab === 'system' }" @click="mainTab = 'system'">
        <svg class="tab-icon" viewBox="0 0 24 24" fill="none">
          <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        系统日志
      </button>
      <button class="main-tab" :class="{ active: mainTab === 'traffic' }" @click="mainTab = 'traffic'">
        <svg class="tab-icon" viewBox="0 0 24 24" fill="none">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        流量日志
      </button>
    </div>

    <!-- 系统日志 -->
    <div v-show="mainTab === 'system'" class="tab-content">
      <div class="sub-tabs">
        <button class="sub-tab" :class="{ active: systemTab === 'realtime' }" @click="systemTab = 'realtime'">
          实时日志
          <span class="realtime-indicator" v-if="wsConnected"></span>
        </button>
        <button class="sub-tab" :class="{ active: systemTab === 'history' }" @click="systemTab = 'history'">
          历史日志
        </button>
      </div>

      <!-- 实时系统日志 -->
      <div v-show="systemTab === 'realtime'" class="realtime-container">
        <div class="realtime-controls">
          <select v-model="systemLogLevel" class="level-select">
            <option value="">全部级别</option>
            <option value="INFO">Info</option>
            <option value="WARN">Warning</option>
            <option value="ERROR">Error</option>
            <option value="DEBUG">Debug</option>
          </select>
          <button class="btn btn-sm" @click="toggleAutoScroll">
            {{ autoScroll ? '停止滚动' : '自动滚动' }}
          </button>
          <button class="btn btn-sm btn-danger" @click="clearSystemLogs">清空日志</button>
        </div>

        <div class="logs-container" ref="systemLogsContainer">
          <div v-if="systemLogs.length === 0" class="empty-state">
            <p>暂无系统日志数据</p>
            <p class="hint">系统运行日志将实时显示在这里</p>
          </div>
          <div v-else class="logs-list">
            <div
              v-for="(log, index) in filteredSystemLogs"
              :key="index"
              class="log-entry"
              :class="log.level.toLowerCase()"
            >
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span class="log-level" :class="log.level.toLowerCase()">{{ log.level }}</span>
              <span class="log-source" v-if="log.source">[{{ log.source }}]</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
        </div>

        <div class="logs-footer">
          <span class="connection-status" :class="{ connected: wsConnected }">
            {{ wsConnected ? '已连接' : '未连接' }}
          </span>
          <span class="log-count">日志数: {{ systemLogs.length }}</span>
        </div>
      </div>

      <!-- 历史系统日志 -->
      <div v-show="systemTab === 'history'" class="history-container">
        <div class="filter-bar">
          <div class="filter-item">
            <label>日志级别：</label>
            <select v-model="systemFilters.level">
              <option value="">全部</option>
              <option value="INFO">Info</option>
              <option value="WARN">Warning</option>
              <option value="ERROR">Error</option>
              <option value="DEBUG">Debug</option>
            </select>
          </div>
          <div class="filter-item">
            <label>时间范围：</label>
            <select v-model="systemFilters.timeRange">
              <option value="">全部</option>
              <option value="today">今天</option>
              <option value="week">本周</option>
              <option value="month">本月</option>
            </select>
          </div>
          <div class="filter-item">
            <label>关键词：</label>
            <input type="text" v-model="systemFilters.keyword" placeholder="搜索日志内容" />
          </div>
          <button class="btn btn-sm btn-primary" @click="searchSystemLogs">搜索</button>
        </div>

        <div class="logs-table-container">
          <table class="logs-table">
            <thead>
              <tr>
                <th>时间</th>
                <th>级别</th>
                <th>来源</th>
                <th>消息</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="historySystemLogs.length === 0">
                <td colspan="4" class="empty-cell">暂无历史日志</td>
              </tr>
              <tr v-for="(log, index) in historySystemLogs" :key="index">
                <td>{{ formatDateTime(log.timestamp) }}</td>
                <td>
                  <span class="log-level-badge" :class="log.level.toLowerCase()">{{ log.level }}</span>
                </td>
                <td>{{ log.source || '-' }}</td>
                <td class="message-cell">{{ log.message }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 流量日志 -->
    <div v-show="mainTab === 'traffic'" class="tab-content">
      <div class="sub-tabs">
        <button class="sub-tab" :class="{ active: trafficTab === 'realtime' }" @click="trafficTab = 'realtime'">
          实时流量
          <span class="realtime-indicator" v-if="wsConnected"></span>
        </button>
        <button class="sub-tab" :class="{ active: trafficTab === 'history' }" @click="trafficTab = 'history'">
          历史流量
        </button>
      </div>

      <!-- 实时流量日志 -->
      <div v-show="trafficTab === 'realtime'" class="realtime-container">
        <div class="realtime-controls">
          <select v-model="trafficLogLevel" class="level-select">
            <option value="">全部方法</option>
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="DELETE">DELETE</option>
          </select>
          <button class="btn btn-sm" @click="toggleTrafficAutoScroll">
            {{ trafficAutoScroll ? '停止滚动' : '自动滚动' }}
          </button>
          <button class="btn btn-sm btn-danger" @click="clearTrafficLogs">清空日志</button>
        </div>

        <div class="logs-container" ref="trafficLogsContainer">
          <div v-if="trafficLogs.length === 0" class="empty-state">
            <p>暂无流量日志数据</p>
            <p class="hint">代理流量将实时显示在这里</p>
          </div>
          <div v-else class="logs-list">
            <div
              v-for="(log, index) in filteredTrafficLogs"
              :key="index"
              class="log-entry traffic"
            >
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span class="method-badge" :class="log.method?.toLowerCase()">{{ log.method }}</span>
              <span class="log-url">{{ log.url }}</span>
              <span class="status-badge" :class="getStatusClass(log.statusCode)">{{ log.statusCode }}</span>
              <span class="log-duration">{{ log.responseTime }}ms</span>
            </div>
          </div>
        </div>

        <div class="logs-footer">
          <span class="connection-status" :class="{ connected: wsConnected }">
            {{ wsConnected ? '已连接' : '未连接' }}
          </span>
          <span class="log-count">日志数: {{ trafficLogs.length }}</span>
        </div>
      </div>

      <!-- 历史流量日志 -->
      <div v-show="trafficTab === 'history'" class="history-container">
        <!-- 统计卡片 -->
        <div class="stats-cards" v-if="logStore.stats">
          <div class="stat-card">
            <div class="stat-label">总请求数</div>
            <div class="stat-value">{{ logStore.stats.totalRequests }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">总流量</div>
            <div class="stat-value">{{ formatBytes(logStore.stats.totalTraffic) }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">上传流量</div>
            <div class="stat-value">{{ formatBytes(logStore.stats.uploadBytes) }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">下载流量</div>
            <div class="stat-value">{{ formatBytes(logStore.stats.downloadBytes) }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">平均响应时间</div>
            <div class="stat-value">{{ logStore.stats.avgResponseTimeMs }}ms</div>
          </div>
        </div>

        <!-- 筛选 -->
        <div class="filter-bar">
          <div class="filter-item">
            <label>时间范围：</label>
            <select v-model="trafficFilters.timeRange" @change="handleTrafficFilterChange">
              <option value="">全部</option>
              <option value="today">今天</option>
              <option value="week">本周</option>
              <option value="month">本月</option>
            </select>
          </div>
          <div class="filter-item">
            <label>方法：</label>
            <select v-model="trafficFilters.method" @change="handleTrafficFilterChange">
              <option value="">全部</option>
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>
          <div class="filter-item">
            <label>状态码：</label>
            <input type="number" v-model.number="trafficFilters.statusCode" @change="handleTrafficFilterChange" placeholder="如：200" />
          </div>
          <div class="filter-item">
            <label>关键词：</label>
            <input type="text" v-model="trafficFilters.keyword" @input="handleTrafficFilterChange" placeholder="URL或错误信息" />
          </div>
        </div>

        <!-- 日志列表 -->
        <div class="logs-table-container">
          <table class="logs-table">
            <thead>
              <tr>
                <th>时间</th>
                <th>客户端IP</th>
                <th>方法</th>
                <th>URL</th>
                <th>状态码</th>
                <th>响应时间</th>
                <th>响应大小</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="logStore.loading">
                <td colspan="8" class="loading-cell">加载中...</td>
              </tr>
              <tr v-else-if="logStore.logs.length === 0">
                <td colspan="8" class="empty-cell">暂无流量日志</td>
              </tr>
              <tr v-for="log in logStore.logs" :key="log.id">
                <td>{{ formatDateTime(log.timestamp) }}</td>
                <td>{{ log.clientIp }}</td>
                <td>
                  <span class="method-badge" :class="log.method.toLowerCase()">{{ log.method }}</span>
                </td>
                <td class="url-cell" :title="log.domain + log.path">{{ log.domain }}{{ log.path }}</td>
                <td>
                  <span class="status-badge" :class="getStatusClass(log.statusCode)">
                    {{ log.statusCode }}
                  </span>
                </td>
                <td>{{ log.durationMs }}ms</td>
                <td>{{ formatBytes(log.uploadBytes + log.downloadBytes) }}</td>
                <td>
                  <button class="btn-link" @click="showLogDetail(log)">详情</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
        <div class="pagination" v-if="logStore.total > pageSize">
          <button class="btn" :disabled="page <= 1" @click="changePage(-1)">上一页</button>
          <span class="page-info">{{ page }} / {{ Math.ceil(logStore.total / pageSize) }}</span>
          <button class="btn" :disabled="page >= Math.ceil(logStore.total / pageSize)" @click="changePage(1)">下一页</button>
        </div>
      </div>
    </div>

    <!-- 日志详情弹窗 -->
    <div class="modal" v-if="selectedLog" @click.self="selectedLog = null">
      <div class="modal-content">
        <div class="modal-header">
          <h3>请求详情</h3>
          <button class="close-btn" @click="selectedLog = null">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-row">
            <span class="detail-label">时间：</span>
            <span>{{ formatDateTime(selectedLog.timestamp) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">客户端IP：</span>
            <span>{{ selectedLog.clientIp }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">服务端IP：</span>
            <span>{{ selectedLog.serverIp }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">域名：</span>
            <span>{{ selectedLog.domain }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">方法：</span>
            <span class="method-badge" :class="selectedLog.method.toLowerCase()">{{ selectedLog.method }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">路径：</span>
            <span>{{ selectedLog.path }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">协议：</span>
            <span>{{ selectedLog.protocol }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">状态码：</span>
            <span class="status-badge" :class="getStatusClass(selectedLog.statusCode)">{{ selectedLog.statusCode }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">上传流量：</span>
            <span>{{ formatBytes(selectedLog.uploadBytes) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">下载流量：</span>
            <span>{{ formatBytes(selectedLog.downloadBytes) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">响应时间：</span>
            <span>{{ selectedLog.durationMs }}ms</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { useLogStore } from '@/stores/traffic'
import { useWebSocket, type LogEntry } from '@/services/websocket'
import type { TrafficLog } from '@/api/traffic'

const logStore = useLogStore()
const { connected: wsConnected, onMessage } = useWebSocket()

// 主标签页
const mainTab = ref('system')
const systemTab = ref('realtime')
const trafficTab = ref('realtime')

// 系统日志
const systemLogs = ref<LogEntry[]>([])
const systemLogLevel = ref('')
const autoScroll = ref(true)
const systemLogsContainer = ref<HTMLElement | null>(null)
const maxSystemLogs = 500

// 流量日志
const trafficLogs = ref<any[]>([])
const trafficLogLevel = ref('')
const trafficAutoScroll = ref(true)
const trafficLogsContainer = ref<HTMLElement | null>(null)
const maxTrafficLogs = 500

// 历史日志
const historySystemLogs = ref<any[]>([])
const page = ref(1)
const pageSize = 20
const selectedLog = ref<TrafficLog | null>(null)

// 筛选条件
const systemFilters = reactive({
  level: '',
  timeRange: '',
  keyword: ''
})

const trafficFilters = reactive({
  timeRange: '',
  method: '',
  statusCode: null as number | null,
  keyword: ''
})

// 计算属性
const filteredSystemLogs = computed(() => {
  if (!systemLogLevel.value) {
    return systemLogs.value
  }
  return systemLogs.value.filter(log => log.level.toLowerCase() === systemLogLevel.value.toLowerCase())
})

const filteredTrafficLogs = computed(() => {
  if (!trafficLogLevel.value) {
    return trafficLogs.value
  }
  return trafficLogs.value.filter(log => log.method === trafficLogLevel.value)
})

// 方法
function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

function formatDateTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString('zh-CN')
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function getStatusClass(status: number): string {
  if (status >= 200 && status < 300) return 'success'
  if (status >= 300 && status < 400) return 'redirect'
  if (status >= 400 && status < 500) return 'client-error'
  if (status >= 500) return 'server-error'
  return ''
}

function addSystemLog(log: LogEntry) {
  systemLogs.value.push(log)
  
  if (systemLogs.value.length > maxSystemLogs) {
    systemLogs.value.shift()
  }
  
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom(systemLogsContainer.value)
    })
  }
}

function addTrafficLog(data: any) {
  trafficLogs.value.push({
    timestamp: new Date().toISOString(),
    method: data.method || 'GET',
    url: data.url || data.path || '/',
    statusCode: data.statusCode || data.status_code || 200,
    responseTime: data.responseTime || data.response_time || 0
  })
  
  if (trafficLogs.value.length > maxTrafficLogs) {
    trafficLogs.value.shift()
  }
  
  if (trafficAutoScroll.value) {
    nextTick(() => {
      scrollToBottom(trafficLogsContainer.value)
    })
  }
}

function scrollToBottom(container: HTMLElement | null) {
  if (container) {
    container.scrollTop = container.scrollHeight
  }
}

function toggleAutoScroll() {
  autoScroll.value = !autoScroll.value
  if (autoScroll.value) {
    scrollToBottom(systemLogsContainer.value)
  }
}

function toggleTrafficAutoScroll() {
  trafficAutoScroll.value = !trafficAutoScroll.value
  if (trafficAutoScroll.value) {
    scrollToBottom(trafficLogsContainer.value)
  }
}

function clearSystemLogs() {
  systemLogs.value = []
}

function clearTrafficLogs() {
  trafficLogs.value = []
}

function handleRefresh() {
  if (mainTab.value === 'system') {
    searchSystemLogs()
  } else {
    loadTrafficLogs()
    loadTrafficStats()
  }
}

function handleExport() {
  let logs: any[]
  let filename: string
  
  if (mainTab.value === 'system') {
    logs = systemTab.value === 'realtime' ? systemLogs.value : historySystemLogs.value
    filename = `system-logs-${new Date().toISOString().split('T')[0]}.json`
  } else {
    logs = trafficTab.value === 'realtime' ? trafficLogs.value : logStore.logs
    filename = `traffic-logs-${new Date().toISOString().split('T')[0]}.json`
  }
  
  const blob = new Blob([JSON.stringify(logs, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function handleClearLogs() {
  if (confirm('确定要清理所有日志吗？')) {
    if (mainTab.value === 'system') {
      clearSystemLogs()
    } else {
      clearTrafficLogs()
      logStore.clearOldLogs().catch(() => {})
    }
  }
}

function searchSystemLogs() {
  // 这里应该调用后端API搜索系统日志
  console.log('搜索系统日志', systemFilters)
}

function loadTrafficLogs() {
  const params: any = {
    limit: pageSize,
    offset: (page.value - 1) * pageSize
  }

  if (trafficFilters.method) {
    params.method = trafficFilters.method
  }
  if (trafficFilters.statusCode) {
    params.status_code = trafficFilters.statusCode
  }
  if (trafficFilters.keyword) {
    params.keyword = trafficFilters.keyword
  }

  const now = new Date()
  if (trafficFilters.timeRange === 'today') {
    params.start_time = new Date(now.setHours(0, 0, 0, 0)).toISOString()
  } else if (trafficFilters.timeRange === 'week') {
    params.start_time = new Date(now.setDate(now.getDate() - 7)).toISOString()
  } else if (trafficFilters.timeRange === 'month') {
    params.start_time = new Date(now.setMonth(now.getMonth() - 1)).toISOString()
  }

  logStore.fetchLogs(params)
}

function loadTrafficStats() {
  logStore.fetchStats()
}

function handleTrafficFilterChange() {
  page.value = 1
  loadTrafficLogs()
  loadTrafficStats()
}

function changePage(delta: number) {
  page.value += delta
  loadTrafficLogs()
}

function showLogDetail(log: TrafficLog) {
  selectedLog.value = log
}

onMounted(() => {
  // 监听WebSocket消息
  onMessage('log', (data: LogEntry) => {
    addSystemLog(data)
  })
  
  onMessage('traffic', (data: any) => {
    addTrafficLog(data)
  })
  
  onMessage('connected', () => {
    addSystemLog({
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'WebSocket已连接',
      source: 'system'
    })
  })
  
  // 加载流量日志
  loadTrafficLogs()
  loadTrafficStats()
})

// 监听标签页切换
watch(mainTab, (newVal) => {
  if (newVal === 'traffic') {
    loadTrafficLogs()
    loadTrafficStats()
  }
})
</script>

<style lang="scss" scoped>
.logs-page {
  padding: 24px;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 48px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-shrink: 0;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-actions {
  display: flex;
  gap: 12px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;

  &.btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    &:hover { 
      background: var(--bg-secondary);
      border-color: var(--primary-color);
    }
  }

  &.btn-primary {
    background: var(--primary-color);
    color: white;
    &:hover { background: #3d7eff; }
  }

  &.btn-danger {
    background: var(--error-color);
    color: #fff;
    &:hover { background: #b91c1c; }
  }

  &.btn-link {
    background: none;
    color: var(--primary-color);
    padding: 4px 8px;
    &:hover { text-decoration: underline; }
  }

  &.btn-sm {
    padding: 4px 12px;
    font-size: 12px;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.icon {
  width: 16px;
  height: 16px;
}

// 主标签页
.main-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 0;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.main-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    color: var(--text-primary);
  }

  &.active {
    color: var(--primary-color);
    border-bottom-color: var(--primary-color);
  }
}

.tab-icon {
  width: 18px;
  height: 18px;
}

.tab-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

// 子标签页
.sub-tabs {
  display: flex;
  gap: 0;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.sub-tab {
  padding: 8px 16px;
  background: none;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;

  &:hover {
    color: var(--text-primary);
    background: var(--bg-tertiary);
  }

  &.active {
    color: var(--primary-color);
    background: rgba(22, 93, 255, 0.1);
  }
}

.realtime-indicator {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--success-color);
  margin-left: 8px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

// 实时日志容器
.realtime-container,
.history-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-top: 16px;
}

.realtime-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.level-select {
  padding: 6px 12px;
  background-color: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
}

.logs-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  background-color: var(--bg-primary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);

  p { margin: 4px 0; }
  .hint { font-size: 12px; color: var(--text-tertiary); }
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.log-entry {
  display: flex;
  padding: 4px 8px;
  border-radius: 2px;
  gap: 8px;
  
  &:hover {
    background-color: var(--bg-tertiary);
  }

  &.info { color: var(--text-primary); }
  &.warn { color: var(--warning-color); background-color: rgba(255, 184, 0, 0.05); }
  &.error { color: var(--error-color); background-color: rgba(255, 71, 87, 0.05); }
  &.debug { color: var(--text-tertiary); }
  &.traffic { color: var(--text-primary); }
}

.log-time {
  color: var(--text-tertiary);
  min-width: 70px;
}

.log-level {
  font-weight: 600;
  min-width: 50px;
  text-transform: uppercase;
  font-size: 10px;

  &.info { color: var(--primary-color); }
  &.warn { color: var(--warning-color); }
  &.error { color: var(--error-color); }
  &.debug { color: var(--text-tertiary); }
}

.log-source {
  color: var(--accent-color);
  font-size: 11px;
}

.log-message {
  flex: 1;
  word-break: break-all;
}

.log-url {
  flex: 1;
  word-break: break-all;
  color: var(--text-primary);
}

.log-duration {
  color: var(--text-secondary);
  min-width: 60px;
  text-align: right;
}

.logs-footer {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.connection-status {
  display: flex;
  align-items: center;
  
  &::before {
    content: '';
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: var(--error-color);
    margin-right: 6px;
  }

  &.connected::before {
    background-color: var(--success-color);
  }
}

.log-count {
  color: var(--text-tertiary);
}

// 历史日志
.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;

  label {
    color: var(--text-secondary);
    font-size: 13px;
  }

  select, input {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 6px 12px;
    color: var(--text-primary);
    font-size: 13px;

    &:focus {
      outline: none;
      border-color: var(--primary-color);
    }

    &::placeholder {
      color: var(--text-tertiary);
    }
  }
}

// 统计卡片
.stats-cards {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  flex-shrink: 0;
}

.stat-card {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 16px;
  text-align: center;
  border: 1px solid var(--border-color);
  transition: all 0.3s;

  &:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.15);
  }
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

// 日志表格
.logs-table-container {
  flex: 1;
  overflow: auto;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.logs-table {
  width: 100%;
  border-collapse: collapse;

  th, td {
    padding: 12px 16px;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
  }

  th {
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    font-weight: 500;
    font-size: 12px;
    text-transform: uppercase;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  td {
    color: var(--text-primary);
    font-size: 13px;
  }

  tbody tr:hover {
    background: var(--bg-tertiary);
  }
}

.loading-cell, .empty-cell {
  text-align: center;
  color: var(--text-secondary);
  padding: 40px !important;
}

.url-cell {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-cell {
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.method-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;

  &.get { background: rgba(0, 200, 83, 0.15); color: var(--success-color); }
  &.post { background: rgba(22, 93, 255, 0.15); color: var(--primary-color); }
  &.put { background: rgba(255, 184, 0, 0.15); color: var(--warning-color); }
  &.delete { background: rgba(255, 71, 87, 0.15); color: var(--error-color); }
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;

  &.success { background: rgba(0, 200, 83, 0.15); color: var(--success-color); }
  &.redirect { background: rgba(22, 93, 255, 0.15); color: var(--primary-color); }
  &.client-error { background: rgba(255, 184, 0, 0.15); color: var(--warning-color); }
  &.server-error { background: rgba(255, 71, 87, 0.15); color: var(--error-color); }
}

.log-level-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;

  &.info { background: rgba(22, 93, 255, 0.15); color: var(--primary-color); }
  &.warn { background: rgba(255, 184, 0, 0.15); color: var(--warning-color); }
  &.error { background: rgba(255, 71, 87, 0.15); color: var(--error-color); }
  &.debug { background: rgba(139, 149, 165, 0.15); color: var(--text-secondary); }
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  flex-shrink: 0;
}

.page-info {
  color: var(--text-secondary);
}

// 弹窗
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-secondary);
  border-radius: 12px;
  width: 90%;
  max-width: 800px;
  max-height: 80vh;
  overflow: auto;
  border: 1px solid var(--border-color);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);

  h3 {
    color: var(--text-primary);
    margin: 0;
  }
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 24px;
  cursor: pointer;
  &:hover { color: var(--text-primary); }
}

.modal-body {
  padding: 24px;
}

.detail-row {
  margin-bottom: 16px;

  .detail-label {
    color: var(--text-secondary);
    margin-right: 8px;
  }
}

@media (max-width: 1200px) {
  .stats-cards {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .filter-bar {
    flex-direction: column;
  }
}
</style>
