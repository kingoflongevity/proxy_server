<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useNodeStore } from '@/stores'
import type { ProxyType, NodeStatus } from '@/types'

const nodeStore = useNodeStore()

const searchKeyword = ref('')
const selectedType = ref<ProxyType[]>([])
const selectedStatus = ref<NodeStatus[]>([])
const sortBy = ref('latency')
const sortOrder = ref<'asc' | 'desc'>('asc')

/**
 * 初始化
 */
onMounted(async () => {
  await nodeStore.fetchNodes()
  await nodeStore.fetchCurrentNode()
})

/**
 * 过滤后的节点列表
 */
const filteredNodes = computed(() => {
  let result = [...nodeStore.nodes]

  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(
      (node) =>
        node.name.toLowerCase().includes(keyword) ||
        node.server.toLowerCase().includes(keyword)
    )
  }

  // 类型过滤
  if (selectedType.value.length > 0) {
    result = result.filter((node) => selectedType.value.includes(node.type))
  }

  // 状态过滤
  if (selectedStatus.value.length > 0) {
    result = result.filter((node) => selectedStatus.value.includes(node.status))
  }

  // 排序
  result.sort((a, b) => {
    let valueA: number
    let valueB: number

    if (sortBy.value === 'latency') {
      valueA = a.latency
      valueB = b.latency
    } else if (sortBy.value === 'name') {
      return sortOrder.value === 'asc'
        ? a.name.localeCompare(b.name)
        : b.name.localeCompare(a.name)
    } else {
      valueA = 0
      valueB = 0
    }

    return sortOrder.value === 'asc' ? valueA - valueB : valueB - valueA
  })

  return result
})

/**
 * 测试单个节点
 */
async function testNode(id: string) {
  try {
    await nodeStore.testNodeLatency(id)
  } catch (error) {
    console.error('测试节点失败:', error)
  }
}

/**
 * 测试所有节点
 */
async function testAllNodes() {
  try {
    await nodeStore.testAllNodes()
  } catch (error) {
    console.error('测试所有节点失败:', error)
  }
}

/**
 * 选择节点
 */
async function selectNode(id: string) {
  try {
    await nodeStore.selectNode(id)
  } catch (error) {
    console.error('选择节点失败:', error)
  }
}

/**
 * 获取状态颜色
 */
function getStatusColor(status: NodeStatus): string {
  const colorMap: Record<NodeStatus, string> = {
    available: 'success',
    unavailable: 'error',
    testing: 'warning',
    unknown: 'secondary',
  }
  return colorMap[status]
}

/**
 * 获取状态文本
 */
function getStatusText(status: NodeStatus): string {
  const textMap: Record<NodeStatus, string> = {
    available: '可用',
    unavailable: '不可用',
    testing: '测试中',
    unknown: '未知',
  }
  return textMap[status]
}

/**
 * 切换排序
 */
function toggleSort(field: string) {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortOrder.value = 'asc'
  }
}

/**
 * 重置过滤器
 */
function resetFilters() {
  searchKeyword.value = ''
  selectedType.value = []
  selectedStatus.value = []
}
</script>

<template>
  <div class="nodes">
    <!-- 头部操作栏 -->
    <div class="header-actions">
      <h2 class="page-title">节点列表</h2>
      <button class="btn btn-primary" @click="testAllNodes" :disabled="nodeStore.testing">
        <svg viewBox="0 0 24 24" fill="none">
          <path
            d="M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C15.3019 3 18.1885 4.77814 19.7545 7.42909"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
        <span>{{ nodeStore.testing ? '测试中...' : '测试所有节点' }}</span>
      </button>
    </div>

    <!-- 过滤器 -->
    <div class="filters">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none">
          <circle cx="11" cy="11" r="8" stroke="currentColor" stroke-width="2" />
          <path d="M21 21L16.65 16.65" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
        <input
          v-model="searchKeyword"
          type="text"
          placeholder="搜索节点名称或地址..."
        />
      </div>

      <div class="filter-group">
        <label>类型:</label>
        <select v-model="selectedType" multiple>
          <option value="ss">SS</option>
          <option value="ssr">SSR</option>
          <option value="vmess">VMess</option>
          <option value="vless">VLESS</option>
          <option value="trojan">Trojan</option>
          <option value="hysteria">Hysteria</option>
        </select>
      </div>

      <div class="filter-group">
        <label>状态:</label>
        <select v-model="selectedStatus" multiple>
          <option value="available">可用</option>
          <option value="unavailable">不可用</option>
          <option value="testing">测试中</option>
          <option value="unknown">未知</option>
        </select>
      </div>

      <button class="btn" @click="resetFilters">重置</button>
    </div>

    <!-- 节点表格 -->
    <div class="nodes-table">
      <table>
        <thead>
          <tr>
            <th class="sortable" @click="toggleSort('name')">
              节点名称
              <span v-if="sortBy === 'name'" class="sort-icon">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </th>
            <th>类型</th>
            <th>地址</th>
            <th class="sortable" @click="toggleSort('latency')">
              延迟
              <span v-if="sortBy === 'latency'" class="sort-icon">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="node in filteredNodes"
            :key="node.id"
            :class="{ selected: nodeStore.currentNode?.id === node.id }"
          >
            <td class="node-name">{{ node.name }}</td>
            <td>
              <span class="type-badge">{{ node.type.toUpperCase() }}</span>
            </td>
            <td class="node-address">{{ node.server }}:{{ node.port }}</td>
            <td>
              <span class="latency" :class="{
                'latency-good': node.latency < 100,
                'latency-medium': node.latency >= 100 && node.latency < 300,
                'latency-bad': node.latency >= 300
              }">
                {{ node.latency }}ms
              </span>
            </td>
            <td>
              <span class="status-badge" :class="getStatusColor(node.status)">
                {{ getStatusText(node.status) }}
              </span>
            </td>
            <td>
              <div class="actions">
                <button
                  class="btn btn-sm"
                  @click="selectNode(node.id)"
                  :disabled="nodeStore.currentNode?.id === node.id"
                  title="选择节点"
                >
                  <svg viewBox="0 0 24 24" fill="none">
                    <path
                      d="M20 6L9 17L4 12"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                </button>
                <button
                  class="btn btn-sm"
                  @click="testNode(node.id)"
                  :disabled="node.status === 'testing'"
                  title="测试延迟"
                >
                  <svg viewBox="0 0 24 24" fill="none">
                    <path
                      d="M12 2L2 7L12 12L22 7L12 2Z"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 空状态 -->
      <div v-if="filteredNodes.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2" />
          <path
            d="M12 2V6M12 18V22M2 12H6M18 12H22"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
        <p>暂无节点</p>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.nodes {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.header-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-md;
  background-color: $bg-color-light;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  color: $text-color-primary;
  font-size: $font-size-base;
  cursor: pointer;
  transition: all $transition-duration $transition-timing;

  svg {
    width: 16px;
    height: 16px;
  }

  &:hover:not(:disabled) {
    background-color: $bg-color-lighter;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.btn-primary {
    background-color: $primary-color;
    border-color: $primary-color;
    color: white;

    &:hover:not(:disabled) {
      background-color: lighten($primary-color, 10%);
    }
  }

  &.btn-sm {
    padding: $spacing-xs $spacing-sm;
    font-size: $font-size-sm;
  }
}

.filters {
  display: flex;
  gap: $spacing-md;
  padding: $spacing-md;
  background-color: $bg-color-light;
  border-radius: $border-radius-base;
  border: 1px solid $border-color;
}

.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-md;
  background-color: $bg-color-darker;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;

  svg {
    width: 16px;
    height: 16px;
    color: $text-color-secondary;
  }

  input {
    flex: 1;
    background: none;
    border: none;
    color: $text-color-primary;
    font-size: $font-size-base;

    &::placeholder {
      color: $text-color-secondary;
    }

    &:focus {
      outline: none;
    }
  }
}

.filter-group {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  label {
    font-size: $font-size-sm;
    color: $text-color-secondary;
  }

  select {
    padding: $spacing-sm $spacing-md;
    background-color: $bg-color-darker;
    border: 1px solid $border-color;
    border-radius: $border-radius-base;
    color: $text-color-primary;
    font-size: $font-size-base;

    &:focus {
      border-color: $primary-color;
      outline: none;
    }
  }
}

.nodes-table {
  background-color: $bg-color-light;
  border-radius: $border-radius-lg;
  border: 1px solid $border-color;
  overflow: hidden;

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background-color: $bg-color-darker;
  }

  th {
    padding: $spacing-md;
    text-align: left;
    font-size: $font-size-sm;
    font-weight: 500;
    color: $text-color-secondary;
    border-bottom: 1px solid $border-color;

    &.sortable {
      cursor: pointer;
      user-select: none;

      &:hover {
        color: $text-color-primary;
      }
    }
  }

  td {
    padding: $spacing-md;
    font-size: $font-size-sm;
    color: $text-color-primary;
    border-bottom: 1px solid $border-color;
  }

  tr {
    &:hover {
      background-color: $bg-color-darker;
    }

    &.selected {
      background-color: rgba($primary-color, 0.1);
    }
  }

  tbody tr:last-child td {
    border-bottom: none;
  }
}

.node-name {
  font-weight: 500;
}

.node-address {
  font-family: monospace;
  color: $text-color-secondary;
}

.type-badge {
  display: inline-block;
  padding: 2px $spacing-sm;
  background-color: rgba($primary-color, 0.1);
  color: $primary-color;
  border-radius: $border-radius-sm;
  font-size: $font-size-xs;
  font-weight: 500;
}

.latency {
  font-weight: 500;

  &.latency-good {
    color: $success-color;
  }

  &.latency-medium {
    color: $warning-color;
  }

  &.latency-bad {
    color: $error-color;
  }
}

.status-badge {
  display: inline-block;
  padding: 2px $spacing-sm;
  border-radius: $border-radius-sm;
  font-size: $font-size-xs;
  font-weight: 500;

  &.success {
    background-color: rgba($success-color, 0.1);
    color: $success-color;
  }

  &.error {
    background-color: rgba($error-color, 0.1);
    color: $error-color;
  }

  &.warning {
    background-color: rgba($warning-color, 0.1);
    color: $warning-color;
  }

  &.secondary {
    background-color: rgba($text-color-secondary, 0.1);
    color: $text-color-secondary;
  }
}

.actions {
  display: flex;
  gap: $spacing-xs;
}

.sort-icon {
  margin-left: $spacing-xs;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-xl * 2;
  color: $text-color-secondary;

  svg {
    width: 64px;
    height: 64px;
    margin-bottom: $spacing-md;
  }
}

// 响应式
@media (max-width: $breakpoint-md) {
  .filters {
    flex-direction: column;
  }

  .nodes-table {
    overflow-x: auto;

    table {
      min-width: 800px;
    }
  }
}
</style>
