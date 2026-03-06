<script setup lang="ts">
import { onMounted } from 'vue'
import { useSettingsStore } from '@/stores'

const settingsStore = useSettingsStore()

/**
 * 初始化应用
 */
onMounted(async () => {
  // 加载系统设置
  try {
    await settingsStore.fetchSettings()
  } catch (error) {
    console.error('加载设置失败:', error)
  }
})
</script>

<template>
  <router-view />
</template>

<style lang="scss">
#app {
  width: 100%;
  height: 100%;
}
</style>
