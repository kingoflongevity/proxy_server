import axios, { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import type { ApiError } from '@/types'

const instance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

instance.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) {
        return data.data as any
      }
      const error: ApiError = {
        code: data.code,
        message: data.message || '请求失败',
        details: data.message,
      }
      return Promise.reject(error)
    }
    return data as any
  },
  (error) => {
    const apiError: ApiError = {
      code: error.response?.status || 500,
      message: error.response?.data?.message || error.message || '网络请求失败',
      details: error.response?.data?.details,
    }
    return Promise.reject(apiError)
  }
)

export const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.get(url, config) as unknown as Promise<T>
  },

  post<T = any>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    if (data instanceof FormData) {
      const formDataConfig = {
        ...config,
        headers: {
          ...config?.headers,
          'Content-Type': 'multipart/form-data',
        },
      }
      return instance.post(url, data, formDataConfig) as unknown as Promise<T>
    }
    return instance.post(url, data, config) as unknown as Promise<T>
  },

  put<T = any>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    return instance.put(url, data, config) as unknown as Promise<T>
  },

  patch<T = any>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    return instance.patch(url, data, config) as unknown as Promise<T>
  },

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.delete(url, config) as unknown as Promise<T>
  },
}

export default instance
