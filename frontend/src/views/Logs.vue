<template>
  <div class="logs-page">
    <div class="page-header">
      <h1 class="page-title">流量日志</h1>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="handleRefresh">
          <svg class="icon" viewBox="0 0 24 24" fill="none">
            <path d="M4 12a8 8 0 0 1 8-8 8 8 0 0 1 8 8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M20 12a8 8 0 0 1-8 8 8 8 0 0 1-8-8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          刷新
        </button>
        <button class="btn btn-danger" @click="handleClearLogs">
          <svg class="icon" viewBox="0 0 24 24" fill="none">
            <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m3 0v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          清理日志
        </button>
      </div>
    </div>

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
        <select v-model="filters.timeRange" @change="handleFilterChange">
          <option value="">全部</option>
          <option value="today">今天</option>
          <option value="week">本周</option>
          <option value="month">本月</option>
        </select>
      </div>
      <div class="filter-item">
        <label>方法：</label>
        <select v-model="filters.method" @change="handleFilterChange">
          <option value="">全部</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
        </select>
      </div>
      <div class="filter-item">
        <label>状态码：</label>
        <input type="number" v-model.number="filters.statusCode" @change="handleFilterChange" placeholder="如：200" />
      </div>
      <div class="filter-item">
        <label>关键词：</label>
        <input type="text" v-model="filters.keyword" @input="handleFilterChange" placeholder="URL或错误信息" />
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
            <td colspan="8" class="empty-cell">暂无日志</td>
          </tr>
          <tr v-for="log in logStore.logs" :key="log.id">
            <td>{{ formatTime(log.timestamp) }}</td>
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
            <span>{{ formatTime(selectedLog.timestamp) }}</span>
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
import { ref, reactive, onMounted } from 'vue'
import { useLogStore } from '@/stores/traffic'
import type { TrafficLog } from '@/api/traffic'

const logStore = useLogStore()

const page = ref(1)
const pageSize = 20
const selectedLog = ref<TrafficLog | null>(null)

const filters = reactive({
  timeRange: '',
  method: '',
  statusCode: null as number | null,
  keyword: ''
})

onMounted(() => {
  loadLogs()
  loadStats()
})

function loadLogs() {
  const params: any = {
    limit: pageSize,
    offset: (page.value - 1) * pageSize
  }

  if (filters.method) {
    params.method = filters.method
  }
  if (filters.statusCode) {
    params.status_code = filters.statusCode
  }
  if (filters.keyword) {
    params.keyword = filters.keyword
  }

  const now = new Date()
  if (filters.timeRange === 'today') {
    params.start_time = new Date(now.setHours(0, 0, 0, 0)).toISOString()
  } else if (filters.timeRange === 'week') {
    params.start_time = new Date(now.setDate(now.getDate() - 7)).toISOString()
  } else if (filters.timeRange === 'month') {
    params.start_time = new Date(now.setMonth(now.getMonth() - 1)).toISOString()
  }

  logStore.fetchLogs(params)
}

function loadStats() {
  logStore.fetchStats()
}

function handleFilterChange() {
  page.value = 1
  loadLogs()
  loadStats()
}

function handleRefresh() {
  loadLogs()
  loadStats()
}

function changePage(delta: number) {
  page.value += delta
  loadLogs()
}

function showLogDetail(log: TrafficLog) {
  selectedLog.value = log
}

function handleClearLogs() {
  if (confirm('确定要清理所有日志吗？')) {
    logStore.clearOldLogs().then(() => {
      alert('日志清理成功')
    }).catch(() => {
      alert('日志清理失败')
    })
  }
}

function formatTime(timestamp: string): string {
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
</script>

<style lang="scss" scoped>
.logs-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
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
    background: #374151;
    color: #fff;
    &:hover { background: #4b5563; }
  }

  &.btn-danger {
    background: #dc2626;
    color: #fff;
    &:hover { background: #b91c1c; }
  }

  &.btn-link {
    background: none;
    color: #3b82f6;
    padding: 4px 8px;
    &:hover { text-decoration: underline; }
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

.stats-cards {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: #1f2937;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.stat-label {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;

  label {
    color: #9ca3af;
    font-size: 14px;
  }

  select, input {
    background: #1f2937;
    border: 1px solid #374151;
    border-radius: 6px;
    padding: 6px 12px;
    color: #fff;
    font-size: 14px;

    &:focus {
      outline: none;
      border-color: #3b82f6;
    }
  }
}

.logs-table-container {
  background: #1f2937;
  border-radius: 8px;
  overflow: hidden;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;

  th, td {
    padding: 12px 16px;
    text-align: left;
    border-bottom: 1px solid #374151;
  }

  th {
    background: #111827;
    color: #9ca3af;
    font-weight: 500;
    font-size: 12px;
    text-transform: uppercase;
  }

  td {
    color: #d1d5db;
    font-size: 14px;
  }

  tbody tr:hover {
    background: #374151;
  }
}

.loading-cell, .empty-cell {
  text-align: center;
  color: #9ca3af;
  padding: 40px !important;
}

.url-cell {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.method-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;

  &.get { background: #065f46; color: #34d399; }
  &.post { background: #1e40af; color: #60a5fa; }
  &.put { background: #92400e; color: #fbbf24; }
  &.delete { background: #991b1b; color: #f87171; }
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;

  &.success { background: #065f46; color: #34d399; }
  &.redirect { background: #1e40af; color: #60a5fa; }
  &.client-error { background: #92400e; color: #fbbf24; }
  &.server-error { background: #991b1b; color: #f87171; }
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
}

.page-info {
  color: #9ca3af;
}

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
  background: #1f2937;
  border-radius: 12px;
  width: 90%;
  max-width: 800px;
  max-height: 80vh;
  overflow: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #374151;

  h3 {
    color: #fff;
    margin: 0;
  }
}

.close-btn {
  background: none;
  border: none;
  color: #9ca3af;
  font-size: 24px;
  cursor: pointer;
  &:hover { color: #fff; }
}

.modal-body {
  padding: 24px;
}

.detail-row {
  margin-bottom: 16px;

  .detail-label {
    color: #9ca3af;
    margin-right: 8px;
  }

  .error-text {
    color: #f87171;
  }
}

.code-block {
  background: #111827;
  padding: 12px;
  border-radius: 6px;
  color: #d1d5db;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
