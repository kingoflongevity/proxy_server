import { request } from './request'

export interface TrafficLog {
  id: string
  timestamp: string
  clientIp: string
  serverIp: string
  domain: string
  method: string
  path: string
  uploadBytes: number
  downloadBytes: number
  durationMs: number
  statusCode: number
  protocol: string
  userId?: string
}

export interface LogStats {
  totalRequests: number
  totalTraffic: number
  uploadBytes: number
  downloadBytes: number
  avgResponseTimeMs: number
  statusCounts: Record<string, number>
  topDomains: Array<{
    domain: string
    requestCount: number
    uploadBytes: number
    downloadBytes: number
  }>
  topClients: Array<{
    clientIp: string
    requestCount: number
    uploadBytes: number
    downloadBytes: number
  }>
}

export function getTrafficLogs(params?: {
  start_time?: string
  end_time?: string
  client_ip?: string
  method?: string
  status_code?: number
  url?: string
  keyword?: string
  user_id?: string
  limit?: number
  offset?: number
}): Promise<{ logs: TrafficLog[]; total: number }> {
  return request.get('/traffic-logs', { params })
}

export function getLogStats(params?: {
  start_time?: string
  end_time?: string
}): Promise<LogStats> {
  return request.get('/traffic-logs/stats', { params })
}

export function clearLogs(before?: string): Promise<void> {
  return request.delete('/traffic-logs', { params: { before } })
}

export function getTrafficSummary(params?: {
  start_time?: string
  end_time?: string
}): Promise<LogStats> {
  return request.get('/traffic-stats/summary', { params })
}
