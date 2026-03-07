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
  authType: string
  osType: string
  status: string
  proxyPort?: number
  proxyType?: string
}

interface ScanResult {
  ip: string
  osType: string
  port: number
}

interface TopologyGroup {
  id: string
  name: string
  nodes: { id: string; name: string; status: string }[]
}

interface TopologyConnection {
  from: string
  to: string
  type: string
}

const servers = ref<ClusterServer[]>([])
const groups = ref<TopologyGroup[]>([])
const connections = ref<TopologyConnection[]>([])

const showScanDialog = ref(false)
const showAddServerDialog = ref(false)
const showDeployDialog = ref(false)

const scanning = ref(false)
const deploying = ref(false)
const scanProgress = ref(0)
const deployProgress = ref(0)
const scanResults = ref<ScanResult[]>([])

const scanForm = ref({
  cidr: '',
  ports: [22],
})

const serverForm = ref({
  name: '',
  ip: '',
  port: 22,
  username: '',
  password: '',
  privateKey: '',
  authType: 'password',
  osType: 'ubuntu',
})

const deployForm = ref({
  proxyPort: 10808,
  proxyType: 'mixed',
})

const selectedServer = ref<ClusterServer | null>(null)

onMounted(async () => {
  await fetchServers()
  await fetchTopology()
  await getCurrentNetworkSegment()
})

async function fetchServers() {
  try {
    const data = await request.get<ClusterServer[]>('/cluster/servers')
    servers.value = data as any
  } catch (error) {
    console.error('获取服务器列表失败:', error)
  }
}

async function fetchTopology() {
  try {
    const data = await request.get<any>('/cluster/topology')
    const topology = data as any
    groups.value = topology.groups || []
    const conns = topology.connections || []
    connections.value = conns.map((c: any) => ({
      from: c.source || c.from,
      to: c.target || c.to,
      type: c.type,
    }))
  } catch (error) {
    console.error('获取拓扑失败:', error)
  }
}

async function getCurrentNetworkSegment() {
  try {
    const data = await request.get<any>('/cluster/network-segment')
    const segment = data as any
    if (segment && segment.cidr) {
      scanForm.value.cidr = segment.cidr
    }
  } catch (error) {
    console.error('获取当前网段失败:', error)
    // 设置默认值
    scanForm.value.cidr = '192.168.1.0/24'
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
    const data = await request.post<any>('/cluster/scan', scanForm.value)
    const result = data as any
    
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
      const data = await request.get<any>(`/cluster/scan/${taskId}`)
      const task = data as any
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

function addScannedServer(result: ScanResult) {
  serverForm.value.ip = result.ip
  serverForm.value.port = result.port
  serverForm.value.osType = result.osType
  showScanDialog.value = false
  showAddServerDialog.value = true
}

async function addServer() {
  try {
    await request.post('/cluster/servers', serverForm.value)
    showAddServerDialog.value = false
    await fetchServers()
  } catch (error) {
    console.error('创建服务器失败:', error)
  }
}

async function testConnection(id: string) {
  try {
    await request.post(`/cluster/servers/${id}/test`)
    alert('连接成功')
  } catch (error) {
    console.error('连接失败:', error)
    alert('连接失败')
  }
}

async function deployProxy(server: ClusterServer) {
  deploying.value = true
  deployProgress.value = 0
  selectedServer.value = server
  showDeployDialog.value = true
  
  const progressInterval = setInterval(() => {
    if (deployProgress.value < 90) {
      deployProgress.value += Math.random() * 5
    }
  }, 500)

  try {
    const data = await request.post<any>('/cluster/deploy', deployForm.value)
    const result = data as any
    
    clearInterval(progressInterval)
    deployProgress.value = 100
    
    if (result && result.taskId) {
      await pollDeployResult(result.taskId)
    }
  } catch (error) {
    console.error('部署失败:', error)
  } finally {
    clearInterval(progressInterval)
    deploying.value = false
  }
}

async function pollDeployResult(taskId: string) {
  let attempts = 0
  const maxAttempts = 120
  
  while (attempts < maxAttempts) {
    try {
      const data = await request.get<any>(`/cluster/deploy/${taskId}`)
      const task = data as any
      deployProgress.value = task.progress || 0
      
      if (task.status === 'completed') {
        await fetchServers()
        return
      } else if (task.status === 'failed') {
        throw new Error('部署任务失败')
      }
    } catch (error) {
      console.error('获取部署结果失败:', error)
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000))
    attempts++
  }
}

async function handleDeleteServer(id: string) {
  if (confirm('确定要删除此服务器吗？')) {
    try {
      await request.delete(`/cluster/servers/${id}`)
      await fetchServers()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }
}

async function handleBackup(id: string) {
  try {
    const data = await request.get<Blob>(`/cluster/servers/${id}/backup`, {
      responseType: 'blob',
    })
    const response = data as any
    const url = window.URL.createObjectURL(new Blob([response]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `backup-${id}.zip`)
    document.body.appendChild(link)
    link.click()
    link.remove()
  } catch (error) {
    console.error('备份失败:', error)
  }
}
</script>

<template>
  <div class="cluster">
    <div class="header-actions">
      <h2 class="page-title">集群管理</h2>
      <div class="action-buttons">
        <button type="button" class="btn" @click="showScanDialog = true">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C15.3019 3 18.1885 4.77814 19.7545 7.42909"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
            <path
              d="M16 3L20 7L16 11"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <span>扫描网络</span>
        </button>
        <button type="button" class="btn" @click="showAddServerDialog = true">
          <svg viewBox="0 0 24 24" fill="none">
            <path
              d="M12 5V19M5 12H19"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
          </svg>
          <span>添加服务器</span>
        </button>
      </div>
    </div>

    <div class="cluster-content">
      <div class="topology-section">
        <h3>拓扑图</h3>
        <TopologyGraph :groups="groups" :connections="connections" />
      </div>

      <div class="servers-section">
        <h3>服务器列表</h3>
        <div class="server-list">
          <div v-for="server in servers" :key="server.id" class="server-card">
            <div class="server-header">
              <h4>{{ server.name }}</h4>
              <span class="status-badge" :class="server.status">{{ server.status }}</span>
            </div>
            <div class="server-info">
              <div class="info-row">
                <span class="label">IP:</span>
                <span class="value">{{ server.ip }}:{{ server.port }}</span>
              </div>
              <div class="info-row">
                <span class="label">系统:</span>
                <span class="value">{{ server.osType }}</span>
              </div>
              <div class="info-row">
                <span class="label">认证:</span>
                <span class="value">{{ server.authType }}</span>
              </div>
              <div v-if="server.proxyPort" class="info-row">
                <span class="label">代理:</span>
                <span class="value">{{ server.proxyType }}://{{ server.ip }}:{{ server.proxyPort }}</span>
              </div>
            </div>
            <div class="server-actions">
              <button class="btn btn-sm" @click="testConnection(server.id)" title="测试连接">
                <svg viewBox="0 0 24 24" fill="none">
                  <path
                    d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
              <button class="btn btn-sm" @click="deployProxy(server)" title="部署代理">
                <svg viewBox="0 0 24 24" fill="none">
                  <path
                    d="M4 16V17C4 18.6569 5.34315 20 7 20H17C18.6569 20 20 18.6569 20 17V16M12 4V16M12 4L8 8M12 4L16 8"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
              <button class="btn btn-sm" @click="handleBackup(server.id)" title="备份">
                <svg viewBox="0 0 24 24" fill="none">
                  <path
                    d="M8 7V3M16 7V3M7 11H17M5 21H19C20.1046 21 21 20.1046 21 19V7C21 5.89543 20.1046 5 19 5H5C3.89543 5 3 5.89543 3 7V19C3 20.1046 3.89543 21 5 21Z"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
              <button class="btn btn-sm btn-danger" @click="handleDeleteServer(server.id)" title="删除">
                <svg viewBox="0 0 24 24" fill="none">
                  <path
                    d="M3 6H5H21M8 6V4C8 3.46957 8.21071 2.96086 8.58579 2.58579C8.96086 2.21071 9.46957 2 10 2H14C14.5304 2 15.0391 2.21071 15.4142 2.58579C15.7893 2.96086 16 3.46957 16 4V6M19 6V20C19 20.5304 18.7893 21.0391 18.4142 21.4142C18.0391 21.7893 17.5304 22 17 22H7C6.46957 22 5.96086 21.7893 5.58579 21.4142C5.21071 21.0391 5 20.5304 5 20V6H19Z"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showScanDialog" class="dialog-overlay" @click.self="showScanDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>扫描网络</h3>
          <button type="button" class="close-btn" @click="showScanDialog = false">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M18 6L6 18M6 6L18 18"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>CIDR 网段</label>
            <input
              v-model="scanForm.cidr"
              type="text"
              placeholder="例如: 192.168.1.0/24"
              :disabled="scanning"
            />
          </div>
          <div v-if="scanning" class="progress-section">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: scanProgress + '%' }"></div>
            </div>
            <div class="progress-text">扫描进度: {{ Math.round(scanProgress) }}%</div>
          </div>
          <div v-if="scanResults.length > 0" class="scan-results">
            <h4>扫描结果</h4>
            <div class="result-list">
              <div v-for="result in scanResults" :key="result.ip" class="result-item">
                <span>{{ result.ip }} ({{ result.osType }})</span>
                <button class="btn btn-sm" @click="addScannedServer(result)">添加</button>
              </div>
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
          <button type="button" class="close-btn" @click="showAddServerDialog = false">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M18 6L6 18M6 6L18 18"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </div>
        <form class="dialog-body" @submit.prevent="addServer">
          <div class="form-group">
            <label>服务器名称</label>
            <input v-model="serverForm.name" type="text" required placeholder="输入服务器名称" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>IP 地址</label>
              <input v-model="serverForm.ip" type="text" required placeholder="输入IP地址" />
            </div>
            <div class="form-group">
              <label>端口</label>
              <input v-model.number="serverForm.port" type="number" required />
            </div>
          </div>
          <div class="form-group">
            <label>用户名</label>
            <input v-model="serverForm.username" type="text" required placeholder="输入用户名" />
          </div>
          <div class="form-group">
            <label>认证方式</label>
            <select v-model="serverForm.authType">
              <option value="password">密码</option>
              <option value="key">密钥</option>
            </select>
          </div>
          <div v-if="serverForm.authType === 'password'" class="form-group">
            <label>密码</label>
            <input v-model="serverForm.password" type="password" placeholder="输入密码" />
          </div>
          <div v-else class="form-group">
            <label>私钥</label>
            <textarea v-model="serverForm.privateKey" placeholder="输入私钥内容" rows="4"></textarea>
          </div>
          <div class="form-group">
            <label>操作系统</label>
            <select v-model="serverForm.osType">
              <option value="ubuntu">Ubuntu</option>
              <option value="centos">CentOS</option>
              <option value="debian">Debian</option>
              <option value="other">其他</option>
            </select>
          </div>
          <div class="dialog-footer">
            <button type="button" class="btn" @click="showAddServerDialog = false">取消</button>
            <button type="submit" class="btn btn-primary">确定</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showDeployDialog" class="dialog-overlay" @click.self="showDeployDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>部署代理 - {{ selectedServer?.name }}</h3>
          <button type="button" class="close-btn" @click="showDeployDialog = false">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M18 6L6 18M6 6L18 18"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>代理端口</label>
            <input v-model.number="deployForm.proxyPort" type="number" required />
          </div>
          <div class="form-group">
            <label>代理类型</label>
            <select v-model="deployForm.proxyType">
              <option value="mixed">混合 (HTTP+SOCKS5)</option>
              <option value="http">HTTP</option>
              <option value="socks">SOCKS5</option>
            </select>
          </div>
          <div v-if="deploying" class="progress-section">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: deployProgress + '%' }"></div>
            </div>
            <div class="progress-text">部署进度: {{ Math.round(deployProgress) }}%</div>
          </div>
        </div>
        <div class="dialog-footer">
          <button type="button" class="btn" @click="showDeployDialog = false">取消</button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="deploying"
            @click="deployProxy(selectedServer!)"
          >
            {{ deploying ? '部署中...' : '开始部署' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.cluster {
  padding: 24px;
  position: relative;
  min-height: 100%;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;

  svg {
    width: 16px;
    height: 16px;
  }

  &:hover {
    background-color: var(--bg-secondary);
    border-color: var(--primary-color);
  }

  &.btn-primary {
    background-color: var(--primary-color);
    border-color: var(--primary-color);
    color: #fff;

    &:hover {
      background-color: #3d7eff;
    }
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

.cluster-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.topology-section,
.servers-section {
  background: var(--bg-tertiary);
  border-radius: 12px;
  padding: 24px;
  border: 1px solid var(--border-color);

  h3 {
    margin-bottom: 16px;
    font-size: 16px;
    font-weight: 500;
    color: var(--text-primary);
  }
}

.topology-section {
  max-height: 400px;
  overflow: hidden;
}

.servers-section {
  flex: 1;
}

.server-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
  max-height: 600px;
  overflow-y: auto;
}

.server-card {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 16px;
  border: 1px solid var(--border-color);
  transition: all 0.3s;

  &:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.15);
  }
}

.server-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;

  h4 {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
  }
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;

  &.active {
    background: rgba(0, 200, 83, 0.15);
    color: var(--success-color);
  }

  &.inactive {
    background: rgba(139, 149, 165, 0.15);
    color: var(--text-secondary);
  }

  &.error {
    background: rgba(255, 71, 87, 0.15);
    color: var(--error-color);
  }
}

.server-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;

  .label {
    color: var(--text-secondary);
  }

  .value {
    color: var(--text-primary);
  }
}

.server-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
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
  z-index: 1000;
}

.dialog {
  width: 100%;
  max-width: 500px;
  background-color: var(--bg-secondary);
  border-radius: 12px;
  border: 1px solid var(--border-color);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);

  h3 {
    font-size: 16px;
    font-weight: 500;
    color: var(--text-primary);
  }
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border-radius: 6px;
  color: var(--text-secondary);

  svg {
    width: 20px;
    height: 20px;
  }

  &:hover {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
  }
}

.dialog-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;

  label {
    display: block;
    margin-bottom: 8px;
    font-size: 14px;
    color: var(--text-secondary);
  }

  input,
  select,
  textarea {
    width: 100%;
    padding: 10px 12px;
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 14px;

    &:focus {
      outline: none;
      border-color: var(--primary-color);
    }

    &::placeholder {
      color: var(--text-tertiary);
    }
  }
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.progress-section {
  margin-top: 16px;
}

.progress-bar {
  height: 8px;
  background-color: var(--bg-tertiary);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color), var(--accent-color));
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  text-align: center;
}

.scan-results {
  margin-top: 16px;

  h4 {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 12px;
  }
}

.result-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: var(--bg-tertiary);
  border-radius: 6px;
  border: 1px solid var(--border-color);

  span {
    font-size: 14px;
    color: var(--text-primary);
  }
}
</style>
