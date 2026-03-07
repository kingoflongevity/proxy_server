<template>
  <div class="topology-graph" ref="containerRef">
    <canvas ref="canvasRef" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp" @wheel="onWheel"></canvas>
    <div class="legend">
      <div class="legend-item">
        <span class="legend-dot active"></span>
        <span>运行中</span>
      </div>
      <div class="legend-item">
        <span class="legend-dot idle"></span>
        <span>空闲</span>
      </div>
      <div class="legend-item">
        <span class="legend-dot error"></span>
        <span>错误</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'

interface Server {
  id: string
  name: string
  ip: string
  status: string
  groupId?: string
}

interface Group {
  id: string
  name: string
}

interface Connection {
  from: string
  to: string
  type: string
}

const props = defineProps<{
  servers: Server[]
  groups: Group[]
  connections: Connection[]
}>()

const containerRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

const scale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const isDragging = ref(false)
const lastMouseX = ref(0)
const lastMouseY = ref(0)

let ctx: CanvasRenderingContext2D | null = null
let animationId: number | null = null

const nodePositions = computed(() => {
  const positions: Record<string, { x: number; y: number; type: string; name: string; status: string }> = {}

  const centerX = 400
  const centerY = 200
  const radius = 150

  props.groups.forEach((group, index) => {
    const angle = (index / props.groups.length) * Math.PI * 2 - Math.PI / 2
    positions[group.id] = {
      x: centerX + Math.cos(angle) * radius,
      y: centerY + Math.sin(angle) * radius,
      type: 'group',
      name: group.name,
      status: 'active',
    }
  })

  positions['master'] = {
    x: centerX,
    y: centerY,
    type: 'master',
    name: '主控节点',
    status: 'active',
  }

  props.servers.forEach((server, index) => {
    const angle = (index / props.servers.length) * Math.PI * 2
    const serverRadius = radius + 100
    positions[server.id] = {
      x: centerX + Math.cos(angle) * serverRadius,
      y: centerY + Math.sin(angle) * serverRadius,
      type: 'server',
      name: server.name,
      status: server.status,
    }
  })

  return positions
})

function draw() {
  if (!ctx || !canvasRef.value || !containerRef.value) return

  const width = containerRef.value.clientWidth
  const height = containerRef.value.clientHeight

  canvasRef.value.width = width * window.devicePixelRatio
  canvasRef.value.height = height * window.devicePixelRatio
  canvasRef.value.style.width = width + 'px'
  canvasRef.value.style.height = height + 'px'

  ctx.scale(window.devicePixelRatio, window.devicePixelRatio)

  ctx.fillStyle = '#050810'
  ctx.fillRect(0, 0, width, height)

  ctx.save()
  ctx.translate(offsetX.value, offsetY.value)
  ctx.scale(scale.value, scale.value)

  drawConnections()
  drawNodes()

  ctx.restore()

  animationId = requestAnimationFrame(draw)
}

function drawConnections() {
  if (!ctx) return

  props.connections.forEach((conn) => {
    const from = nodePositions.value[conn.from]
    const to = nodePositions.value[conn.to]

    if (from && to) {
      ctx!.beginPath()
      ctx!.moveTo(from.x, from.y)
      ctx!.lineTo(to.x, to.y)
      ctx!.strokeStyle = conn.type === 'proxy' ? '#165DFF' : '#2A3548'
      ctx!.lineWidth = 2
      ctx!.stroke()
    }
  })

  const master = nodePositions.value['master']
  if (master) {
    props.servers.forEach((server) => {
      const serverPos = nodePositions.value[server.id]
      if (serverPos) {
        ctx!.beginPath()
        ctx!.moveTo(master.x, master.y)
        ctx!.lineTo(serverPos.x, serverPos.y)
        ctx!.strokeStyle = server.status === 'active' ? '#10B981' : '#2A3548'
        ctx!.lineWidth = 2
        ctx!.setLineDash([5, 5])
        ctx!.stroke()
        ctx!.setLineDash([])
      }
    })
  }
}

function drawNodes() {
  if (!ctx) return

  Object.entries(nodePositions.value).forEach(([id, pos]) => {
    const size = pos.type === 'master' ? 50 : pos.type === 'group' ? 40 : 30

    ctx!.beginPath()
    if (pos.type === 'master') {
      ctx!.arc(pos.x, pos.y, size, 0, Math.PI * 2)
    } else if (pos.type === 'group') {
      ctx!.roundRect(pos.x - size, pos.y - size / 2, size * 2, size, 8)
    } else {
      ctx!.arc(pos.x, pos.y, size, 0, Math.PI * 2)
    }

    let color = '#2A3548'
    if (pos.status === 'active') {
      color = '#10B981'
    } else if (pos.status === 'error') {
      color = '#EF4444'
    } else if (pos.status === 'deploying') {
      color = '#F59E0B'
    }

    ctx!.fillStyle = color
    ctx!.globalAlpha = 0.2
    ctx!.fill()
    ctx!.globalAlpha = 1

    ctx!.strokeStyle = color
    ctx!.lineWidth = 2
    ctx!.stroke()

    ctx!.fillStyle = '#FFFFFF'
    ctx!.font = '12px sans-serif'
    ctx!.textAlign = 'center'
    ctx!.textBaseline = 'middle'
    ctx!.fillText(pos.name, pos.x, pos.y + size + 15)
  })
}

function onMouseDown(e: MouseEvent) {
  isDragging.value = true
  lastMouseX.value = e.clientX
  lastMouseY.value = e.clientY
}

function onMouseMove(e: MouseEvent) {
  if (isDragging.value) {
    offsetX.value += e.clientX - lastMouseX.value
    offsetY.value += e.clientY - lastMouseY.value
    lastMouseX.value = e.clientX
    lastMouseY.value = e.clientY
  }
}

function onMouseUp() {
  isDragging.value = false
}

function onWheel(e: WheelEvent) {
  e.preventDefault()
  const delta = e.deltaY > 0 ? 0.9 : 1.1
  scale.value = Math.max(0.5, Math.min(2, scale.value * delta))
}

onMounted(() => {
  if (canvasRef.value) {
    ctx = canvasRef.value.getContext('2d')
    draw()
  }
})

onUnmounted(() => {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
})

watch(
  () => [props.servers, props.groups, props.connections],
  () => {
    draw()
  },
  { deep: true }
)
</script>

<style lang="scss" scoped>
.topology-graph {
  width: 100%;
  height: 100%;
  position: relative;
  cursor: grab;

  &:active {
    cursor: grabbing;
  }

  canvas {
    width: 100%;
    height: 100%;
  }

  .legend {
    position: absolute;
    bottom: 16px;
    left: 16px;
    display: flex;
    gap: 16px;
    background: rgba(17, 24, 39, 0.8);
    padding: 8px 12px;
    border-radius: 6px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #8b95a5;
  }

  .legend-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;

    &.active {
      background: #10b981;
    }

    &.idle {
      background: #6b7280;
    }

    &.error {
      background: #ef4444;
    }
  }
}
</style>
