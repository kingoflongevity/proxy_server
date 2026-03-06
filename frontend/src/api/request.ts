import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import type { ApiResponse, ApiError } from '@/types'

/**
 * 创建Axios实例
 */
const instance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

/**
 * 请求拦截器
 */
instance.interceptors.request.use(
  (config) => {
    // 可以在这里添加token等认证信息
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

/**
 * 响应拦截器
 */
instance.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    // 处理后端响应格式 {code: 0, message: "success", data: ...}
    if (response.data && typeof response.data === 'object') {
      if ('code' in response.data) {
        if (response.data.code === 0) {
          return response.data.data
        }
        // 业务错误
        const error: ApiError = {
          code: response.data.code,
          message: response.data.message || '请求失败',
          details: response.data.message,
        }
        return Promise.reject(error)
      }
    }
    return response.data
  },
  (error) => {
    // HTTP错误
    const apiError: ApiError = {
      code: error.response?.status || 500,
      message: error.response?.data?.message || error.message || '网络请求失败',
      details: error.response?.data?.details,
    }
    return Promise.reject(apiError)
  }
)

/**
 * 封装请求方法
 */
export const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.get(url, config)
  },

  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    if (data instanceof FormData) {
      const formDataConfig = {
        ...config,
        headers: {
          ...config?.headers,
          'Content-Type': 'multipart/form-data',
        },
      }
      return instance.post(url, data, formDataConfig)
    }
    return instance.post(url, data, config)
  },

  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return instance.put(url, data, config)
  },

  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return instance.patch(url, data, config)
  },

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return instance.delete(url, config)
  },
}

export default instance
