import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

const pinia = createPinia()

// 使用持久化插件
pinia.use(piniaPluginPersistedstate)

export default pinia

// 导出所有stores
export * from './subscription'
export * from './node'
export * from './rule'
export * from './settings'
