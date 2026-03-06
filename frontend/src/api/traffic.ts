import { request } from './request'

export interface TrafficLog {
  id: string
  timestamp: string
  client_ip: string
  method: string
  url: string
  path: string
  query_string?: string
  headers?: Record<string, string>
  body?: string
  body_size: number
  user_agent?: string
  protocol: string
  status_code: number
  response_time_ms: number
  response_size: number
  error?: string
  user_id?: string
}

export interface LogStats {
  total_requests: number
  total_traffic: number
  upload_bytes: number
  download_bytes: number
  avg_response_time_ms: number
  status_counts: Record<string, number>
  top_domains: Array<{
    domain: string
    request_count: number
    upload_bytes: number
    download_bytes: number
  }>
  top_clients: Array<{
    client_ip: string
    request_count: number
    upload_bytes: number
    download_bytes: number
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
