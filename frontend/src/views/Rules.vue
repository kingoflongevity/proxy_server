<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRuleStore, useNodeStore } from '@/stores'
import type { CreateRuleRequest, ProxyRule, RuleType, RuleAction } from '@/types'

const ruleStore = useRuleStore()
const nodeStore = useNodeStore()

const showAddDialog = ref(false)
const editingRule = ref<ProxyRule | null>(null)

const formData = ref<CreateRuleRequest>({
  type: 'DOMAIN',
  pattern: '',
  action: 'PROXY',
  enabled: true,
  priority: 100,
  description: '',
})

/**
 * 初始化
 */
onMounted(async () => {
  await ruleStore.fetchRules()
  await nodeStore.fetchNodes()
})

/**
 * 打开添加对话框
 */
function openAddDialog() {
  editingRule.value = null
  formData.value = {
    type: 'DOMAIN',
    pattern: '',
    action: 'PROXY',
    enabled: true,
    priority: 100,
    description: '',
  }
  showAddDialog.value = true
}

/**
 * 打开编辑对话框
 */
function openEditDialog(rule: ProxyRule) {
  editingRule.value = rule
  formData.value = {
    type: rule.type,
    pattern: rule.pattern,
    action: rule.action,
    target: rule.target,
    enabled: rule.enabled,
    priority: rule.priority,
    description: rule.description,
  }
  showAddDialog.value = true
}

/**
 * 提交表单
 */
async function handleSubmit() {
  try {
    if (editingRule.value) {
      await ruleStore.updateRule(editingRule.value.id, formData.value)
    } else {
      await ruleStore.createRule(formData.value)
    }
    showAddDialog.value = false
  } catch (error) {
    console.error('保存规则失败:', error)
  }
}

/**
 * 删除规则
 */
async function handleDelete(id: string) {
  if (confirm('确定要删除这条规则吗？')) {
    try {
      await ruleStore.deleteRule(id)
    } catch (error) {
      console.error('删除规则失败:', error)
    }
  }
}

/**
 * 切换规则状态
 */
async function toggleRule(id: string, enabled: boolean) {
  try {
    await ruleStore.toggleRule(id, enabled)
  } catch (error) {
    console.error('切换规则状态失败:', error)
  }
}

/**
 * 获取规则类型文本
 */
function getRuleTypeText(type: RuleType): string {
  const typeMap: Record<RuleType, string> = {
    DOMAIN: '域名匹配',
    'DOMAIN-SUFFIX': '域名后缀',
    'DOMAIN-KEYWORD': '域名关键词',
    'IP-CIDR': 'IP段',
    'SRC-IP-CIDR': '源IP段',
    GEOIP: '地理位置',
    'DST-PORT': '目标端口',
    'SRC-PORT': '源端口',
    'PROCESS-NAME': '进程名',
    'RULE-SET': '规则集',
    MATCH: '匹配所有',
  }
  return typeMap[type] || type
}

/**
 * 获取动作文本
 */
function getActionText(action: RuleAction): string {
  const actionMap: Record<RuleAction, string> = {
    DIRECT: '直连',
    REJECT: '拒绝',
    PROXY: '代理',
    node: '指定节点',
  }
  return actionMap[action] || action
}

/**
 * 获取动作颜色
 */
function getActionColor(action: RuleAction): string {
  const colorMap: Record<RuleAction, string> = {
    DIRECT: 'success',
    REJECT: 'error',
    PROXY: 'primary',
    node: 'warning',
  }
  return colorMap[action] || 'secondary'
}
</script>

<template>
  <div class="rules">
    <!-- 头部操作栏 -->
    <div class="header-actions">
      <h2 class="page-title">规则配置</h2>
      <button class="btn btn-primary" @click="openAddDialog">
        <svg viewBox="0 0 24 24" fill="none">
          <path d="M12 5V19M5 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
        <span>添加规则</span>
      </button>
    </div>

    <!-- 规则列表 -->
    <div class="rules-list">
      <div
        v-for="rule in ruleStore.rules"
        :key="rule.id"
        class="rule-card"
        :class="{ disabled: !rule.enabled }"
      >
        <div class="rule-header">
          <div class="rule-info">
            <span class="rule-type">{{ getRuleTypeText(rule.type) }}</span>
            <span class="rule-pattern">{{ rule.pattern }}</span>
          </div>
          <div class="rule-actions">
            <button class="btn btn-sm" @click="toggleRule(rule.id, !rule.enabled)" :title="rule.enabled ? '禁用' : '启用'">
              <svg v-if="rule.enabled" viewBox="0 0 24 24" fill="none">
                <path
                  d="M1 12S5 4 12 4S23 12 23 12S19 20 12 20S1 12 1 12Z"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2" />
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none">
                <path
                  d="M17.94 17.94A10.07 10.07 0 0 1 12 20C5 20 1 12 1 12A18.45 18.45 0 0 1 5.06 6.06M9.9 4.24A9.12 9.12 0 0 1 12 4C19 4 23 12 23 12A18.5 18.5 0 0 1 19.18 17.18"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M1 1L23 23"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                />
              </svg>
            </button>
            <button class="btn btn-sm" @click="openEditDialog(rule)" title="编辑">
              <svg viewBox="0 0 24 24" fill="none">
                <path
                  d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
            <button class="btn btn-sm btn-danger" @click="handleDelete(rule.id)" title="删除">
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

        <div class="rule-body">
          <div class="rule-meta">
            <span class="action-badge" :class="getActionColor(rule.action)">
              {{ getActionText(rule.action) }}
            </span>
            <span v-if="rule.target" class="rule-target">→ {{ rule.target }}</span>
            <span class="rule-priority">优先级: {{ rule.priority }}</span>
          </div>
          <div v-if="rule.description" class="rule-description">
            {{ rule.description }}
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="ruleStore.rules.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none">
          <path
            d="M12 2L2 7L12 12L22 7L12 2Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <p>暂无规则</p>
        <button class="btn btn-primary" @click="openAddDialog">添加第一条规则</button>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <div v-if="showAddDialog" class="dialog-overlay" @click.self="showAddDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ editingRule ? '编辑规则' : '添加规则' }}</h3>
          <button class="close-btn" @click="showAddDialog = false">
            <svg viewBox="0 0 24 24" fill="none">
              <path d="M18 6L6 18M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            </svg>
          </button>
        </div>

        <form class="dialog-body" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label>规则类型</label>
            <select v-model="formData.type" required>
              <option value="DOMAIN">域名匹配</option>
              <option value="DOMAIN-SUFFIX">域名后缀</option>
              <option value="DOMAIN-KEYWORD">域名关键词</option>
              <option value="IP-CIDR">IP段</option>
              <option value="SRC-IP-CIDR">源IP段</option>
              <option value="GEOIP">地理位置</option>
              <option value="DST-PORT">目标端口</option>
              <option value="SRC-PORT">源端口</option>
              <option value="PROCESS-NAME">进程名</option>
              <option value="RULE-SET">规则集</option>
              <option value="MATCH">匹配所有</option>
            </select>
          </div>

          <div class="form-group">
            <label>匹配模式</label>
            <input v-model="formData.pattern" type="text" required placeholder="输入匹配模式" />
          </div>

          <div class="form-group">
            <label>动作</label>
            <select v-model="formData.action" required>
              <option value="DIRECT">直连</option>
              <option value="REJECT">拒绝</option>
              <option value="PROXY">代理</option>
              <option value="node">指定节点</option>
            </select>
          </div>

          <div v-if="formData.action === 'node'" class="form-group">
            <label>目标节点</label>
            <select v-model="formData.target">
              <option v-for="node in nodeStore.nodes" :key="node.id" :value="node.id">
                {{ node.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>优先级</label>
            <input v-model.number="formData.priority" type="number" min="1" max="10000" />
          </div>

          <div class="form-group">
            <label>描述</label>
            <input v-model="formData.description" type="text" placeholder="规则描述（可选）" />
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="formData.enabled" type="checkbox" />
              <span>启用规则</span>
            </label>
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
.rules {
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

.rules-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.rule-card {
  background-color: $bg-color-light;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  padding: $spacing-md;
  transition: all $transition-duration $transition-timing;

  &.disabled {
    opacity: 0.5;
  }
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-sm;
}

.rule-info {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.rule-type {
  display: inline-block;
  padding: 2px $spacing-sm;
  background-color: rgba($primary-color, 0.1);
  color: $primary-color;
  border-radius: $border-radius-sm;
  font-size: $font-size-xs;
  font-weight: 500;
}

.rule-pattern {
  font-family: monospace;
  font-size: $font-size-sm;
  color: $text-color-primary;
}

.rule-actions {
  display: flex;
  gap: $spacing-xs;
}

.rule-body {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.rule-meta {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  font-size: $font-size-sm;
}

.action-badge {
  display: inline-block;
  padding: 2px $spacing-sm;
  border-radius: $border-radius-sm;
  font-size: $font-size-xs;
  font-weight: 500;

  &.primary {
    background-color: rgba($primary-color, 0.1);
    color: $primary-color;
  }

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
}

.rule-target {
  color: $text-color-secondary;
}

.rule-priority {
  color: $text-color-secondary;
  font-size: $font-size-xs;
}

.rule-description {
  font-size: $font-size-sm;
  color: $text-color-secondary;
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
  z-index: 2000;
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
  .header-actions {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-md;
  }
}
</style>
