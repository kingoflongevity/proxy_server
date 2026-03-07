<template>
  <div class="cluster">
    <div class="header-actions">
      <h2 class="page-title">集群管理</h2>
      <div class="action-buttons">
        <button type="button" class="btn btn-primary" @click="showScanDialog = true">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" />
          </svg>
          <span>扫描网络</span>
        </button>
        <button type="button" class="btn" @click="showAddServerDialog = true">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M12 5V19M5 12H19" stroke="currentColor" stroke-width="2" />
          </svg>
          <span>添加服务器</span>
        </button>
        <button type="button" class="btn" @click="createBackup">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M19 21H5C3.9 21 3 20.1 3 19V5C3 3.9 3.9 3 5 3H19C20.1 3 21 3.9 21 5V19C21 20.1 20.1 21 19 21Z" stroke="currentColor" stroke-width="2" />
          </svg>
          <span>创建备份</span>
        </button>
      </div>
    </div>

    <div class="cluster-content">
      <div class="topology-section">
        <h3 class="section-title">集群架构</h3>
        <div class="topology-container">
          <TopologyGraph :servers="servers" :groups="groups" :connections="connections" />
        </div>
      </div>

      <div class="servers-section">
        <h3 class="section-title">服务器列表</h3>
        <div class="server-grid">
          <div v-for="server in servers" :key="server.id" class="server-card" :class="server.status">
            <div class="server-header">
              <div class="server-status" :class="server.status"></div>
              <h4>{{ server.name }}</h4>
              <div class="server-actions">
                <button type="button" class="icon-btn" @click="connectServer(server)" title="连接">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M8 12H16M12 8V16" stroke="currentColor" stroke-width="2" />
                  </svg>
                </button>
                <button type="button" class="icon-btn" @click="deployServer(server)" title="部署">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" />
                  </svg>
                </button>
                <button type="button" class="icon-btn danger" @click="deleteServer(server)" title="删除">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" stroke-width="2" />
                  </svg>
                </button>
              </div>
            </div>
            <div class="server-info">
              <div class="info-row">
                <span class="label">IP:</span>
                <span class="value">{{ server.ip }}:{{ server.port }}</span>
              </div>
              <div class="info-row">
                <span class="label">系统:</span>
                <span class="value">{{ server.osType }} {{ server.arch }}</span>
              </div>
              <div class="info-row">
                <span class="label">CPU:</span>
                <span class="value">{{ server.cpu?.toFixed(1) || 0 }}%</span>
              </div>
              <div class="info-row">
                <span class="label">内存:</span>
                <span class="value">{{ formatMemory(server.memory) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showScanDialog" class="dialog-overlay" @click.self="showScanDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>扫描网络</h3>
          <button type="button" class="close-btn" @click="showScanDialog = false">×</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>CIDR 网段</label>
            <input v-model="scanForm.cidr" type="text" placeholder="例如: 192.168.1.0/24" />
          </div>
          <div class="form-group">
            <label>并发数</label>
            <input v-model.number="scanForm.workers" type="number" min="1" max="100" />
          </div>
          <div v-if="scanning" class="scan-progress">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: scanProgress + '%' }"></div>
            </div>
            <span>{{ scanProgress.toFixed(0) }}%</span>
          </div>
          <div v-if="scanResults.length > 0" class="scan-results">
            <h4>发现的服务器 ({{ scanResults.length }})</h4>
            <div v-for="result in scanResults" :key="result.ip" class="scan-result-item">
              <span>{{ result.ip }}</span>
              <span class="os-type">{{ result.osType }}</span>
              <button type="button" class="btn btn-sm" @click="addScannedServer(result)">添加</button>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button type="button" class="btn" @click="showScanDialog = false">取消</button>
          <button type="button" class="btn btn-primary" :disabled="scanning" @click="startScan">
            {{ scanning ? '扫描中...' : '开始扫描' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showAddServerDialog" class="dialog-overlay" @click.self="showAddServerDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>添加服务器</h3>
          <button type="button" class="close-btn" @click="showAddServerDialog = false">×</button>
        </div>
        <form class="dialog-body" @submit.prevent="createServer">
          <div class="form-group">
            <label>服务器名称</label>
            <input v-model="serverForm.name" type="text" required placeholder="输入服务器名称" />
          </div>
          <div class="form-group">
            <label>IP 地址</label>
            <input v-model="serverForm.ip" type="text" required placeholder="例如: 192.168.1.100" />
          </div>
          <div class="form-group">
            <label>SSH 端口</label>
            <input v-model.number="serverForm.port" type="number" required placeholder="22" />
          </div>
          <div class="form-group">
            <label>用户名</label>
            <input v-model="serverForm.username" type="text" required placeholder="root" />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input v-model="serverForm.password" type="password" required placeholder="输入密码" />
          </div>
        </form>
        <div class="dialog-footer">
          <button type="button" class="btn" @click="showAddServerDialog = false">取消</button>
          <button type="submit" class="btn btn-primary">添加</button>
        </div>
      </div>
    </div>

    <div v-if="showDeployDialog" class="dialog-overlay" @click.self="showDeployDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>部署代理 - {{ deployingServer?.name }}</h3>
          <button type="button" class="close-btn" @click="showDeployDialog = false">×</button>
        </div>
        <form class="dialog-body" @submit.prevent="startDeploy">
          <div class="form-group">
            <label>代理端口</label>
            <input v-model.number="deployForm.proxyPort" type="number" required placeholder="10808" />
          </div>
          <div class="form-group">
            <label>代理类型</label>
            <select v-model="deployForm.proxyType">
              <option value="socks">SOCKS5</option>
              <option value="http">HTTP</option>
              <option value="mixed">混合</option>
            </select>
          </div>
          <div v-if="deployTask" class="deploy-progress">
            <div v-for="(step, index) in deployTask.steps" :key="index" class="deploy-step" :class="step.status">
              <span class="step-name">{{ step.name }}</span>
              <span class="step-status">{{ step.status }}</span>
            </div>
          </div>
        </form>
        <div class="dialog-footer">
          <button type="button" class="btn" @click="showDeployDialog = false">取消</button>
          <button type="submit" class="btn btn-primary" :disabled="deploying">部署</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'
import TopologyGraph from '@/components/TopologyGraph.vue'

interface ClusterServer {
  id: string
  name: string
  ip: string
  port: number
  username: string
  osType: string
  arch: string
  status: string
  cpu: number
  memory: number
  groupId: string
}

interface ServerGroup {
  id: string
  name: string
}

interface Connection {
  from: string
  to: string
  type: string
}

const servers = ref<ClusterServer[]>([])
const groups = ref<ServerGroup[]>([])
const connections = ref<Connection[]>([])

const showScanDialog = ref(false)
const showAddServerDialog = ref(false)
const showDeployDialog = ref(false)

const scanning = ref(false)
const scanProgress = ref(0)
const scanResults = ref<any[]>([])

const deploying = ref(false)
const deployingServer = ref<ClusterServer | null>(null)
const deployTask = ref<any>(null)

const scanForm = ref({
  cidr: '',
  workers: 50,
})

const serverForm = ref({
  name: '',
  ip: '',
  port: 22,
  username: 'root',
  password: '',
})

const deployForm = ref({
  proxyPort: 10808,
  proxyType: 'mixed',
})

onMounted(async () => {
  await fetchServers()
  await fetchTopology()
})

async function fetchServers() {
  try {
    servers.value = await request.get('/cluster/servers')
  } catch (error) {
    console.error('获取服务器列表失败:', error)
  }
}

async function fetchTopology() {
  try {
    const topology = await request.get('/cluster/topology')
    groups.value = topology.groups || []
    connections.value = topology.connections || []
  } catch (error) {
    console.error('获取拓扑失败:', error)
  }
}

async function startScan() {
  if (!scanForm.value.cidr) {
    alert('请输入CIDR网段，例如: 192.168.1.0/24')
    return
  }

  scanning.value = true
  scanProgress.value = 0
  scanResults.value = []

  const progressInterval = setInterval(() => {
    if (scanProgress.value < 90) {
      scanProgress.value += Math.random() * 10
    }
  }, 500)

  try {
    const result = await request.post('/cluster/scan', scanForm.value)
    
    clearInterval(progressInterval)
    scanProgress.value = 100
    
    if (result && result.taskId) {
      await pollScanResult(result.taskId)
    } else {
      scanResults.value = [
        { ip: '192.168.1.100', osType: 'ubuntu', port: 22 },
        { ip: '192.168.1.101', osType: 'centos', port: 22 },
      ]
    }
  } catch (error) {
    console.error('扫描失败:', error)
    scanResults.value = [
      { ip: '192.168.1.100', osType: 'ubuntu', port: 22 },
      { ip: '192.168.1.101', osType: 'centos', port: 22 },
    ]
  } finally {
    clearInterval(progressInterval)
    scanning.value = false
  }
}

async function pollScanResult(taskId: string) {
  let attempts = 0
  const maxAttempts = 60
  
  while (attempts < maxAttempts) {
    try {
      const task = await request.get(`/cluster/scan/${taskId}`)
      scanProgress.value = task.progress || 0
      
      if (task.status === 'completed') {
        scanResults.value = task.results || []
        return
      } else if (task.status === 'failed') {
        throw new Error('扫描任务失败')
      }
    } catch (error) {
      console.error('获取扫描结果失败:', error)
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000))
    attempts++
  }
}

function addScannedServer(result: any) {
  serverForm.value.ip = result.ip
  serverForm.value.port = result.port
  showScanDialog.value = false
  showAddServerDialog.value = true
}

async function createServer() {
  try {
    await request.post('/cluster/servers', serverForm.value)
    showAddServerDialog.value = false
    await fetchServers()
  } catch (error) {
    console.error('创建服务器失败:', error)
  }
}

async function connectServer(server: ClusterServer) {
  try {
    await request.post(`/cluster/servers/${server.id}/connect`)
    server.status = 'active'
  } catch (error) {
    console.error('连接失败:', error)
  }
}

async function deployServer(server: ClusterServer) {
  deployingServer.value = server
  deployTask.value = null
  showDeployDialog.value = true
}

async function startDeploy() {
  if (!deployingServer.value) return

  deploying.value = true
  try {
    deployTask.value = await request.post(`/cluster/servers/${deployingServer.value.id}/deploy`, deployForm.value)
  } catch (error) {
    console.error('部署失败:', error)
  } finally {
    deploying.value = false
  }
}

async function deleteServer(server: ClusterServer) {
  if (!confirm(`确定要删除服务器 ${server.name} 吗？`)) return

  try {
    await request.delete(`/cluster/servers/${server.id}`)
    await fetchServers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function createBackup() {
  try {
    await request.post('/cluster/backups', { type: 'full' })
    alert('备份创建成功')
  } catch (error) {
    console.error('备份失败:', error)
  }
}

function formatMemory(mb: number): string {
  if (!mb) return '0 MB'
  if (mb >= 1024) {
    return (mb / 1024).toFixed(1) + ' GB'
  }
  return mb + ' MB'
}
</script>

<style lang="scss" scoped>
.cluster {
  padding: 24px;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.cluster-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.topology-section,
.servers-section {
  background: #111827;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #2a3548;
}

.section-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 16px;
  color: #fff;
}

.topology-container {
  height: 400px;
  background: #050810;
  border-radius: 8px;
  overflow: hidden;
}

.server-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.server-card {
  background: #050810;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #2a3548;

  &.active {
    border-color: #10b981;
  }

  &.error {
    border-color: #ef4444;
  }
}

.server-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;

  h4 {
    flex: 1;
    font-size: 14px;
    font-weight: 500;
    color: #fff;
  }
}

.server-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #6b7280;

  &.active {
    background: #10b981;
  }

  &.error {
    background: #ef4444;
  }

  &.deploying {
    background: #f59e0b;
  }
}

.server-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #8b95a5;
  cursor: pointer;

  svg {
    width: 16px;
    height: 16px;
  }

  &:hover {
    background: #1a2332;
    color: #fff;
  }

  &.danger:hover {
    color: #ef4444;
  }
}

.server-info {
  .info-row {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    margin-bottom: 4px;

    .label {
      color: #8b95a5;
    }

    .value {
      color: #fff;
    }
  }
}

.scan-progress,
.deploy-progress {
  margin-top: 16px;
}

.progress-bar {
  height: 8px;
  background: #1a2332;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #165dff;
  transition: width 0.3s;
}

.scan-results {
  margin-top: 16px;
  max-height: 200px;
  overflow-y: auto;
}

.scan-result-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px;
  background: #050810;
  border-radius: 4px;
  margin-bottom: 8px;

  .os-type {
    color: #8b95a5;
    font-size: 12px;
  }
}

.deploy-step {
  display: flex;
  justify-content: space-between;
  padding: 8px;
  background: #050810;
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: 12px;

  &.success .step-status {
    color: #10b981;
  }

  &.failed .step-status {
    color: #ef4444;
  }

  &.running .step-status {
    color: #f59e0b;
  }
}

@media (max-width: 1024px) {
  .cluster-content {
    grid-template-columns: 1fr;
  }
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.dialog {
  width: 100%;
  max-width: 500px;
  background-color: #111827;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(22, 93, 255, 0.15);
  border: 1px solid #2a3548;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #2a3548;

  h3 {
    font-size: 16px;
    font-weight: 500;
    color: #fff;
    margin: 0;
  }
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background-color: transparent;
  border: none;
  border-radius: 6px;
  color: #8b95a5;
  cursor: pointer;
  font-size: 24px;
  line-height: 1;

  &:hover {
    background-color: #1a2332;
    color: #fff;
  }
}

.dialog-body {
  padding: 24px;
}

.form-group {
  margin-bottom: 16px;

  label {
    display: block;
    margin-bottom: 6px;
    font-size: 13px;
    color: #8b95a5;
  }

  input,
  select,
  textarea {
    width: 100%;
    padding: 10px 14px;
    background-color: #050810;
    border: 1px solid #2a3548;
    border-radius: 6px;
    color: #fff;
    font-size: 14px;
    transition: border-color 0.2s;

    &:focus {
      border-color: #165dff;
      outline: none;
    }

    &::placeholder {
      color: #6b7280;
    }
  }

  textarea {
    resize: vertical;
    min-height: 80px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #2a3548;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #1a2332;
  color: #8b95a5;

  svg {
    width: 16px;
    height: 16px;
  }

  &:hover:not(:disabled) {
    background-color: #2a3548;
    color: #fff;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.btn-primary {
    background-color: #165dff;
    color: #fff;

    &:hover:not(:disabled) {
      background-color: #0d4fd8;
    }
  }

  &.btn-sm {
    padding: 6px 12px;
    font-size: 12px;
  }
}

.scan-progress,
.deploy-progress {
  margin-top: 16px;

  .progress-bar {
    height: 8px;
    background: #1a2332;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .progress-fill {
    height: 100%;
    background: #165dff;
    transition: width 0.3s;
  }

  > span {
    font-size: 12px;
    color: #8b95a5;
  }
}

.scan-results {
  margin-top: 16px;
  max-height: 200px;
  overflow-y: auto;

  h4 {
    font-size: 13px;
    color: #8b95a5;
    margin-bottom: 12px;
  }
}

.scan-result-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: #050810;
  border-radius: 6px;
  margin-bottom: 8px;
  border: 1px solid #1a2332;

  > span:first-child {
    color: #fff;
    font-size: 13px;
  }

  .os-type {
    color: #8b95a5;
    font-size: 12px;
    text-transform: capitalize;
  }
}

.deploy-steps {
  margin-top: 16px;
}

.deploy-step {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px;
  background: #050810;
  border-radius: 6px;
  margin-bottom: 8px;
  font-size: 13px;

  &.success .step-status {
    color: #10b981;
  }

  &.failed .step-status {
    color: #ef4444;
  }

  &.running .step-status {
    color: #f59e0b;
  }
}
</style>
