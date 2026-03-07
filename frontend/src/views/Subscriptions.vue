<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSubscriptionStore } from '@/stores'
import type { CreateSubscriptionRequest, Subscription } from '@/types'

const subscriptionStore = useSubscriptionStore()

const showAddDialog = ref(false)
const editingSubscription = ref<Subscription | null>(null)

const formData = ref<CreateSubscriptionRequest>({
  name: '',
  url: '',
  type: 'mixed',
  autoUpdate: true,
  updateInterval: 24,
})

onMounted(async () => {
  await subscriptionStore.fetchSubscriptions()
})

/**
 * 打开添加对话框
 */
function openAddDialog() {
  console.log('打开添加对话框')
  editingSubscription.value = null
  formData.value = {
    name: '',
    url: '',
    type: 'mixed',
    autoUpdate: true,
    updateInterval: 24,
  }
  showAddDialog.value = true
  console.log('showAddDialog:', showAddDialog.value)
}

/**
 * 打开编辑对话框
 */
function openEditDialog(subscription: Subscription) {
  editingSubscription.value = subscription
  formData.value = {
    name: subscription.name,
    url: subscription.url,
    type: subscription.type,
    autoUpdate: subscription.autoUpdate,
    updateInterval: subscription.updateInterval,
  }
  showAddDialog.value = true
}

/**
 * 提交表单
 */
async function handleSubmit() {
  try {
    if (editingSubscription.value) {
      await subscriptionStore.updateSubscription(editingSubscription.value.id, formData.value)
    } else {
      await subscriptionStore.createSubscription(formData.value)
    }
    showAddDialog.value = false
    editingSubscription.value = null
  } catch (error) {
    console.error('保存订阅失败:', error)
  }
}

/**
 * 删除订阅
 */
async function handleDelete(id: string) {
  if (confirm('确定要删除这个订阅吗？')) {
    try {
      await subscriptionStore.deleteSubscription(id)
    } catch (error) {
      console.error('删除订阅失败:', error)
    }
  }
}

/**
 * 更新订阅节点
 */
async function handleUpdate(id: string) {
  try {
    await subscriptionStore.updateSubscriptionNodes(id)
    alert('更新成功')
  } catch (error) {
    console.error('更新订阅节点失败:', error)
    alert('更新失败')
  }
}

/**
 * 格式化日期
 */
function formatDate(dateString: string): string {
  if (!dateString || dateString === '0001-01-01T00:00:00Z') {
    return '从未更新'
  }
  const date = new Date(dateString)
  if (isNaN(date.getTime())) {
    return '从未更新'
  }
  return date.toLocaleString('zh-CN')
}

/**
 * 获取状态文本
 */
function getStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    active: '正常',
    inactive: '未激活',
    updating: '更新中',
    error: '错误',
  }
  return statusMap[status] || status
}

/**
 * 获取状态颜色
 */
function getStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    active: 'success',
    inactive: 'secondary',
    updating: 'warning',
    error: 'error',
  }
  return colorMap[status] || 'secondary'
}
</script>

<template>
  <div class="subscriptions">
    <!-- 头部操作栏 -->
    <div class="header-actions">
      <h2 class="page-title">订阅管理</h2>
      <button type="button" class="btn btn-primary" @click="() => openAddDialog()">
        <svg viewBox="0 0 24 24" fill="none">
          <path
            d="M12 5V19M5 12H19"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
        <span>添加订阅</span>
      </button>
    </div>

    <!-- 订阅列表 -->
    <div class="subscription-list">
      <div
        v-for="subscription in subscriptionStore.subscriptions"
        :key="subscription.id"
        class="subscription-card"
      >
        <div class="card-header">
          <div class="card-title">
            <h3>{{ subscription.name }}</h3>
            <span class="status-badge" :class="getStatusColor(subscription.status)">
              {{ getStatusText(subscription.status) }}
            </span>
          </div>
          <div class="card-actions">
            <button class="btn btn-sm" @click="handleUpdate(subscription.id)" title="更新节点">
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
            </button>
            <button class="btn btn-sm" @click="openEditDialog(subscription)" title="编辑">
              <svg viewBox="0 0 24 24" fill="none">
                <path
                  d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M18.5 2.50001C18.8978 2.10219 19.4374 1.87869 20 1.87869C20.5626 1.87869 21.1022 2.10219 21.5 2.50001C21.8978 2.89784 22.1213 3.4374 22.1213 4.00001C22.1213 4.56262 21.8978 5.10219 21.5 5.50001L12 15L8 16L9 12L18.5 2.50001Z"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
            <button class="btn btn-sm btn-danger" @click="handleDelete(subscription.id)" title="删除">
              <svg viewBox="0 0 24 24" fill="none">
                <path
                  d="M3 6H5H21"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M8 6V4C8 3.46957 8.21071 2.96086 8.58579 2.58579C8.96086 2.21071 9.46957 2 10 2H14C14.5304 2 15.0391 2.21071 15.4142 2.58579C15.7893 2.96086 16 3.46957 16 4V6M19 6V20C19 20.5304 18.7893 21.0391 18.4142 21.4142C18.0391 21.7893 17.5304 22 17 22H7C6.46957 22 5.96086 21.7893 5.58579 21.4142C5.21071 21.0391 5 20.5304 5 20V6H19Z"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
          </div>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="label">类型:</span>
            <span class="value">{{ (subscription.type || 'unknown').toUpperCase() }}</span>
          </div>
          <div class="info-row">
            <span class="label">节点数:</span>
            <span class="value">{{ subscription.nodeCount }}</span>
          </div>
          <div class="info-row">
            <span class="label">自动更新:</span>
            <span class="value">{{ subscription.autoUpdate ? '是' : '否' }}</span>
          </div>
          <div class="info-row">
            <span class="label">更新间隔:</span>
            <span class="value">{{ subscription.updateInterval }}小时</span>
          </div>
          <div class="info-row">
            <span class="label">最后更新:</span>
            <span class="value">{{ formatDate(subscription.lastUpdate) }}</span>
          </div>
        </div>

        <div class="card-footer">
          <div class="url-display">{{ subscription.url }}</div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="subscriptionStore.subscriptions.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none">
          <path
            d="M12 2L2 7L12 12L22 7L12 2Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d="M2 17L12 22L22 17"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <p>暂无订阅</p>
        <button type="button" class="btn btn-primary" @click="() => openAddDialog()">添加第一个订阅</button>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <div v-if="showAddDialog" class="dialog-overlay" @click.self="showAddDialog = false">
      <div class="dialog" @click.stop>
        <div class="dialog-header">
          <h3>{{ editingSubscription ? '编辑订阅' : '添加订阅' }}</h3>
          <button type="button" class="close-btn" @click="showAddDialog = false">
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

        <form class="dialog-body" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label>订阅名称</label>
            <input v-model="formData.name" type="text" required placeholder="输入订阅名称" />
          </div>

          <div class="form-group">
            <label>订阅地址</label>
            <input v-model="formData.url" type="url" required placeholder="输入订阅地址" />
          </div>

          <div class="form-group">
            <label>订阅类型</label>
            <select v-model="formData.type">
              <option value="mixed">混合</option>
              <option value="ss">Shadowsocks</option>
              <option value="ssr">ShadowsocksR</option>
              <option value="vmess">VMess</option>
              <option value="vless">VLESS</option>
              <option value="trojan">Trojan</option>
              <option value="hysteria">Hysteria</option>
            </select>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="formData.autoUpdate" type="checkbox" />
              <span>自动更新</span>
            </label>
          </div>

          <div v-if="formData.autoUpdate" class="form-group">
            <label>更新间隔（小时）</label>
            <input v-model.number="formData.updateInterval" type="number" min="1" max="168" />
          </div>

          <div class="dialog-footer">
            <button type="button" class="btn" @click="showAddDialog = false">取消</button>
            <button type="submit" class="btn btn-primary">确定</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.subscriptions {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
  position: relative;
  min-height: 100%;
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

  &:hover {
    background-color: $bg-color-lighter;
  }

  &.btn-primary {
    background-color: $primary-color;
    border-color: $primary-color;
    color: white;

    &:hover {
      background-color: lighten($primary-color, 10%);
    }
  }

  &.btn-danger {
    color: $error-color;
    border-color: $error-color;

    &:hover {
      background-color: rgba($error-color, 0.1);
    }
  }

  &.btn-sm {
    padding: $spacing-xs $spacing-sm;
    font-size: $font-size-sm;
  }
}

.subscription-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: $spacing-lg;
}

.subscription-card {
  background-color: $bg-color-light;
  border: 1px solid $border-color;
  border-radius: $border-radius-lg;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-md $spacing-lg;
  border-bottom: 1px solid $border-color;
}

.card-title {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  h3 {
    font-size: $font-size-base;
    font-weight: 500;
    color: $text-color-primary;
  }
}

.status-badge {
  padding: 2px $spacing-sm;
  border-radius: $border-radius-sm;
  font-size: $font-size-xs;
  font-weight: 500;

  &.success {
    background-color: rgba($success-color, 0.1);
    color: $success-color;
  }

  &.warning {
    background-color: rgba($warning-color, 0.1);
    color: $warning-color;
  }

  &.error {
    background-color: rgba($error-color, 0.1);
    color: $error-color;
  }

  &.secondary {
    background-color: rgba($text-color-secondary, 0.1);
    color: $text-color-secondary;
  }
}

.card-actions {
  display: flex;
  gap: $spacing-xs;
}

.card-body {
  padding: $spacing-md $spacing-lg;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: $spacing-xs 0;

  .label {
    color: $text-color-secondary;
    font-size: $font-size-sm;
  }

  .value {
    color: $text-color-primary;
    font-size: $font-size-sm;
    font-weight: 500;
  }
}

.card-footer {
  padding: $spacing-sm $spacing-lg;
  background-color: $bg-color-darker;
}

.url-display {
  font-size: $font-size-xs;
  color: $text-color-secondary;
  word-break: break-all;
}

.empty-state {
  grid-column: 1 / -1;
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

  p {
    margin-bottom: $spacing-lg;
  }
}

// 对话框
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  width: 100%;
  max-width: 500px;
  background-color: $bg-color-light;
  border-radius: $border-radius-lg;
  box-shadow: $box-shadow-lg;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-lg;
  border-bottom: 1px solid $border-color;

  h3 {
    font-size: $font-size-lg;
    font-weight: 500;
    color: $text-color-primary;
  }
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background-color: transparent;
  border-radius: $border-radius-base;
  color: $text-color-secondary;

  svg {
    width: 20px;
    height: 20px;
  }

  &:hover {
    background-color: $bg-color-lighter;
    color: $text-color-primary;
  }
}

.dialog-body {
  padding: $spacing-lg;
}

.form-group {
  margin-bottom: $spacing-md;

  label {
    display: block;
    margin-bottom: $spacing-xs;
    font-size: $font-size-sm;
    color: $text-color-secondary;
  }

  input,
  select {
    width: 100%;
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

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    cursor: pointer;

    input[type='checkbox'] {
      width: auto;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-sm;
  margin-top: $spacing-lg;
}

// 响应式
@media (max-width: $breakpoint-md) {
  .subscription-list {
    grid-template-columns: 1fr;
  }

  .header-actions {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-md;
  }
}
</style>
