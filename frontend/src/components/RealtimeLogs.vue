<template>
  <div class="realtime-logs">
    <div class="logs-header">
      <h3>实时日志</h3>
      <div class="logs-controls">
        <select v-model="logLevel" class="level-select">
          <option value="">全部级别</option>
          <option value="info">Info</option>
          <option value="warn">Warning</option>
          <option value="error">Error</option>
          <option value="debug">Debug</option>
        </select>
        <button class="btn btn-sm" @click="toggleAutoScroll">
          {{ autoScroll ? '停止滚动' : '自动滚动' }}
        </button>
        <button class="btn btn-sm btn-danger" @click="clearLogs">清空日志</button>
      </div>
    </div>

    <div class="logs-container" ref="logsContainer">
      <div v-if="logs.length === 0" class="empty-state">
        <p>暂无日志数据</p>
        <p class="hint">连接后将实时显示日志</p>
      </div>
      <div v-else class="logs-list">
        <div
          v-for="(log, index) in filteredLogs"
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
      <span class="log-count">日志数: {{ logs.length }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useWebSocket, type LogEntry } from '@/services/websocket'

const logs = ref<LogEntry[]>([])
const logLevel = ref('')
const autoScroll = ref(true)
const logsContainer = ref<HTMLElement | null>(null)
const maxLogs = 1000

const { connected: wsConnected, onMessage } = useWebSocket()

const filteredLogs = computed(() => {
  if (!logLevel.value) {
    return logs.value
  }
  return logs.value.filter(log => log.level.toLowerCase() === logLevel.value.toLowerCase())
})

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

function addLog(log: LogEntry) {
  logs.value.push(log)
  
  if (logs.value.length > maxLogs) {
    logs.value.shift()
  }
  
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
}

function scrollToBottom() {
  if (logsContainer.value) {
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
}

function toggleAutoScroll() {
  autoScroll.value = !autoScroll.value
  if (autoScroll.value) {
    scrollToBottom()
  }
}

function clearLogs() {
  logs.value = []
}

onMounted(() => {
  onMessage('log', (data: LogEntry) => {
    addLog(data)
  })
  
  onMessage('connected', () => {
    addLog({
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'WebSocket已连接',
      source: 'system'
    })
  })
})
</script>

<style lang="scss" scoped>
.realtime-logs {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);

  h3 {
    font-size: 16px;
    font-weight: 500;
    color: var(--text-primary);
    margin: 0;
  }
}

.logs-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.level-select {
  padding: 4px 8px;
  background-color: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
}

.btn {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background-color: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background-color: var(--bg-secondary);
    border-color: var(--primary-color);
  }

  &.btn-danger {
    color: var(--error-color);
    border-color: var(--error-color);

    &:hover {
      background-color: rgba(255, 71, 87, 0.1);
    }
  }

  &.btn-sm {
    padding: 4px 8px;
    font-size: 12px;
  }
}

.logs-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  background-color: var(--bg-primary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);

  p {
    margin: 4px 0;
  }

  .hint {
    font-size: 12px;
    color: var(--text-tertiary);
  }
}

.logs-list {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.log-entry {
  display: flex;
  padding: 2px 8px;
  border-radius: 2px;
  margin-bottom: 2px;
  
  &:hover {
    background-color: var(--bg-tertiary);
  }

  &.info {
    color: var(--text-primary);
  }

  &.warn {
    color: var(--warning-color);
    background-color: rgba(255, 184, 0, 0.05);
  }

  &.error {
    color: var(--error-color);
    background-color: rgba(255, 71, 87, 0.05);
  }

  &.debug {
    color: var(--text-tertiary);
  }
}

.log-time {
  color: var(--text-tertiary);
  margin-right: 8px;
  min-width: 70px;
}

.log-level {
  font-weight: 600;
  margin-right: 8px;
  min-width: 40px;
  text-transform: uppercase;
  font-size: 10px;

  &.info {
    color: var(--primary-color);
  }

  &.warn {
    color: var(--warning-color);
  }

  &.error {
    color: var(--error-color);
  }

  &.debug {
    color: var(--text-tertiary);
  }
}

.log-source {
  color: var(--accent-color);
  margin-right: 8px;
  font-size: 11px;
}

.log-message {
  flex: 1;
  word-break: break-all;
}

.logs-footer {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px;
  background-color: var(--bg-tertiary);
  border-top: 1px solid var(--border-color);
  font-size: 12px;
  color: var(--text-secondary);
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
</style>
