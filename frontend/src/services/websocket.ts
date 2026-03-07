import { ref, onMounted, onUnmounted } from 'vue'

export interface WebSocketMessage {
  type: string
  data: any
}

export interface LogEntry {
  timestamp: string
  level: string
  message: string
  source: string
  connectionId?: string
}

export interface ConnectionInfo {
  id: string
  protocol: string
  sourceIp: string
  destHost: string
  destPort: number
  uploadBytes: number
  downloadBytes: number
  startTime: string
}

export interface NodeUpdate {
  nodeId: string
  status: string
  latency: number
  timestamp: string
}

export interface SubscriptionUpdate {
  subscriptionId: string
  nodeCount: number
  status: string
  timestamp: string
}

class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectTimer: number | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 3000
  private listeners: Map<string, Set<(data: any) => void>> = new Map()
  
  connected = ref(false)
  connecting = ref(false)
  error = ref<string | null>(null)

  /**
   * 连接WebSocket
   */
  connect(url?: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.connecting.value = true
    this.error.value = null

    const wsUrl = url || this.getWebSocketUrl()
    
    try {
      this.ws = new WebSocket(wsUrl)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.connected.value = true
        this.connecting.value = false
        this.reconnectAttempts = 0
        this.emit('connected', { status: 'connected' })
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.emit(message.type, message.data)
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        this.error.value = 'WebSocket连接错误'
        this.connecting.value = false
      }

      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.connected.value = false
        this.connecting.value = false
        this.scheduleReconnect()
      }
    } catch (e) {
      console.error('Failed to create WebSocket:', e)
      this.error.value = '创建WebSocket连接失败'
      this.connecting.value = false
      this.scheduleReconnect()
    }
  }

  /**
   * 断开WebSocket连接
   */
  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    
    this.connected.value = false
    this.connecting.value = false
  }

  /**
   * 重连机制
   */
  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached')
      return
    }

    if (this.reconnectTimer) {
      return
    }

    this.reconnectAttempts++
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`)
    
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, this.reconnectDelay)
  }

  /**
   * 获取WebSocket URL
   */
  private getWebSocketUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/ws`
  }

  /**
   * 订阅事件
   */
  on(event: string, callback: (data: any) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  /**
   * 取消订阅事件
   */
  off(event: string, callback: (data: any) => void) {
    const callbacks = this.listeners.get(event)
    if (callbacks) {
      callbacks.delete(callback)
    }
  }

  /**
   * 触发事件
   */
  private emit(event: string, data: any) {
    const callbacks = this.listeners.get(event)
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(data)
        } catch (e) {
          console.error(`Error in WebSocket listener for ${event}:`, e)
        }
      })
    }
  }

  /**
   * 发送消息
   */
  send(type: string, data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }))
    }
  }
}

export const wsService = new WebSocketService()

/**
 * WebSocket连接组合式函数
 */
export function useWebSocket() {
  const connected = wsService.connected
  const connecting = wsService.connecting
  const error = wsService.error

  const onMessage = (type: string, callback: (data: any) => void) => {
    wsService.on(type, callback)
    
    onUnmounted(() => {
      wsService.off(type, callback)
    })
  }

  onMounted(() => {
    if (!connected.value && !connecting.value) {
      wsService.connect()
    }
  })

  onUnmounted(() => {
    // 注意：不要在这里disconnect，因为其他组件可能还在使用
  })

  return {
    connected,
    connecting,
    error,
    connect: () => wsService.connect(),
    disconnect: () => wsService.disconnect(),
    send: wsService.send.bind(wsService),
    onMessage,
  }
}
