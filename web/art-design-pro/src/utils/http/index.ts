/**
 * HTTP 请求封装模块
 * 基于 Axios 封装的 HTTP 请求工具，提供统一的请求/响应处理
 *
 * ## 主要功能
 *
 * - 请求/响应拦截器（自动添加 Token、统一错误处理）
 * - 401 未授权自动登出（带防抖机制）
 * - 请求失败自动重试（可配置）
 * - 统一的成功/错误消息提示
 * - 支持 GET/POST/PUT/DELETE 等常用方法
 *
 * @module utils/http
 * @author Art Design Pro Team
 */

import axios, { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '@/store/modules/user'
import { ApiStatus } from './status'
import { HttpError, handleError, showError, showSuccess } from './error'
import { $t } from '@/locales'
import { BaseResponse } from '@/types'

/** 请求配置常量 */
const REQUEST_TIMEOUT = 15000
const LOGOUT_DELAY = 500
const MAX_RETRIES = 0
const RETRY_DELAY = 1000
const UNAUTHORIZED_DEBOUNCE_TIME = 3000

/** 401防抖状态 */
let isUnauthorizedErrorShown = false
let unauthorizedTimer: NodeJS.Timeout | null = null

/** 扩展 AxiosRequestConfig */
interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  showErrorMessage?: boolean
  showSuccessMessage?: boolean
  _retry?: boolean
}

const { VITE_API_URL, VITE_WITH_CREDENTIALS } = import.meta.env

/** 正在刷新 Token 的 Promise 锁 */
let refreshTokenPromise: Promise<string> | null = null

/** 检查 Refresh Token 是否有效 */
function isRefreshTokenValid(): boolean {
  const userStore = useUserStore()
  if (!userStore.refreshToken) return false
  if (userStore.refreshTokenExpiresAt && Date.now() >= userStore.refreshTokenExpiresAt) {
    return false
  }
  return true
}

/** 检查 Access Token 是否已过期 */
function isAccessTokenExpired(): boolean {
  const userStore = useUserStore()
  if (!userStore.accessToken) return false
  if (userStore.tokenExpiresAt && Date.now() >= userStore.tokenExpiresAt) {
    return true
  }
  return false
}

/** 执行 Token 刷新（单例锁机制） */
async function doRefreshToken(): Promise<string> {
  if (refreshTokenPromise) {
    return refreshTokenPromise
  }

  refreshTokenPromise = (async () => {
    const userStore = useUserStore()
    const currentRefreshToken = userStore.refreshToken

    if (!currentRefreshToken) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await axios.post<BaseResponse<Api.Auth.LoginResponse>>(
        `${VITE_API_URL || ''}/admin/refresh_token`,
        { refreshToken: currentRefreshToken },
        { headers: { 'Content-Type': 'application/json' } }
      )

      if (response.data && response.data.code === ApiStatus.success && response.data.data) {
        const { token, refreshToken, expiresAt, refreshExpiresAt } = response.data.data
        userStore.setToken(token, refreshToken, expiresAt, refreshExpiresAt)
        return token
      } else {
        throw new Error(response.data?.msg || 'Refresh token failed')
      }
    } catch (err) {
      logOut()
      throw err
    } finally {
      refreshTokenPromise = null
    }
  })()

  return refreshTokenPromise
}

/** Axios实例 */
const axiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: VITE_API_URL,
  withCredentials: VITE_WITH_CREDENTIALS === 'true',
  validateStatus: (status) => status >= 200 && status < 300,
  transformResponse: [
    (data, headers) => {
      const contentType = String(headers['content-type'] || '')
      if (contentType.includes('application/json')) {
        try {
          return JSON.parse(data)
        } catch {
          return data
        }
      }
      return data
    }
  ]
})

/** 请求拦截器 */
axiosInstance.interceptors.request.use(
  async (request: InternalAxiosRequestConfig) => {
    const userStore = useUserStore()
    const isAuthEndpoint =
      request.url?.includes('/admin/login') ||
      request.url?.includes('/admin/refresh_token') ||
      request.url?.includes('/admin/logout')

    // 如果 Access Token 已过期但 Refresh Token 有效，提前无感刷新 Token
    if (!isAuthEndpoint && isAccessTokenExpired() && isRefreshTokenValid()) {
      try {
        const newToken = await doRefreshToken()
        request.headers.set('Authorization', newToken)
      } catch {
        // doRefreshToken 失败会自动触发 logOut
      }
    } else if (userStore.accessToken) {
      request.headers.set('Authorization', userStore.accessToken)
    }

    if (
      request.data &&
      typeof request.data === 'object' &&
      !(request.data instanceof FormData) &&
      !request.headers['Content-Type']
    ) {
      request.headers.set('Content-Type', 'application/json')
      request.data = JSON.stringify(request.data)
    }

    return request
  },
  (error) => {
    showError(createHttpError($t('httpMsg.requestConfigError'), ApiStatus.error))
    return Promise.reject(error)
  }
)

/** 响应拦截器 */
axiosInstance.interceptors.response.use(
  async (response: AxiosResponse<BaseResponse>) => {
    const { code, msg } = response.data
    if (code === ApiStatus.success) return response

    const config = response.config as InternalAxiosRequestConfig & ExtendedAxiosRequestConfig
    const isAuthEndpoint =
      config.url?.includes('/admin/login') ||
      config.url?.includes('/admin/refresh_token') ||
      config.url?.includes('/admin/logout')

    // 401 响应处理：重试刷新 Token 并重新发起原请求
    if (code === ApiStatus.unauthorized && !config._retry && !isAuthEndpoint) {
      if (isRefreshTokenValid()) {
        config._retry = true
        try {
          const newToken = await doRefreshToken()
          config.headers.set('Authorization', newToken)
          return axiosInstance(config)
        } catch {
          handleUnauthorizedError(msg)
        }
      } else {
        handleUnauthorizedError(msg)
      }
    }

    if (code === ApiStatus.unauthorized) handleUnauthorizedError(msg)
    throw createHttpError(msg || $t('httpMsg.requestFailed'), code)
  },
  async (error) => {
    const config = error.config as
      (InternalAxiosRequestConfig & ExtendedAxiosRequestConfig) | undefined
    const isAuthEndpoint =
      config?.url?.includes('/admin/login') ||
      config?.url?.includes('/admin/refresh_token') ||
      config?.url?.includes('/admin/logout')

    if (
      error.response?.status === ApiStatus.unauthorized &&
      config &&
      !config._retry &&
      !isAuthEndpoint
    ) {
      if (isRefreshTokenValid()) {
        config._retry = true
        try {
          const newToken = await doRefreshToken()
          config.headers.set('Authorization', newToken)
          return axiosInstance(config)
        } catch {
          handleUnauthorizedError()
        }
      } else {
        handleUnauthorizedError()
      }
    }

    if (error.response?.status === ApiStatus.unauthorized) handleUnauthorizedError()
    return Promise.reject(handleError(error))
  }
)

/** 统一创建HttpError */
function createHttpError(message: string, code: number) {
  return new HttpError(message, code)
}

/** 处理401错误（带防抖） */
function handleUnauthorizedError(message?: string): never {
  const error = createHttpError(message || $t('httpMsg.unauthorized'), ApiStatus.unauthorized)

  if (!isUnauthorizedErrorShown) {
    isUnauthorizedErrorShown = true
    logOut()

    unauthorizedTimer = setTimeout(resetUnauthorizedError, UNAUTHORIZED_DEBOUNCE_TIME)

    showError(error, true)
    throw error
  }

  throw error
}

/** 重置401防抖状态 */
function resetUnauthorizedError() {
  isUnauthorizedErrorShown = false
  if (unauthorizedTimer) clearTimeout(unauthorizedTimer)
  unauthorizedTimer = null
}

/** 退出登录函数 */
function logOut() {
  setTimeout(() => {
    useUserStore().logOut()
  }, LOGOUT_DELAY)
}

/** 是否需要重试 */
function shouldRetry(statusCode: number) {
  return [
    ApiStatus.requestTimeout,
    ApiStatus.internalServerError,
    ApiStatus.badGateway,
    ApiStatus.serviceUnavailable,
    ApiStatus.gatewayTimeout
  ].includes(statusCode)
}

/** 请求重试逻辑 */
async function retryRequest<T>(
  config: ExtendedAxiosRequestConfig,
  retries: number = MAX_RETRIES
): Promise<T> {
  try {
    return await request<T>(config)
  } catch (error) {
    if (retries > 0 && error instanceof HttpError && shouldRetry(error.code)) {
      await delay(RETRY_DELAY)
      return retryRequest<T>(config, retries - 1)
    }
    throw error
  }
}

/** 延迟函数 */
function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** 请求函数 */
async function request<T = any>(config: ExtendedAxiosRequestConfig): Promise<T> {
  // POST | PUT 参数自动填充
  if (
    ['POST', 'PUT'].includes(config.method?.toUpperCase() || '') &&
    config.params &&
    !config.data
  ) {
    config.data = config.params
    config.params = undefined
  }

  try {
    const res = await axiosInstance.request<BaseResponse<T>>(config)

    // 显示成功消息
    if (config.showSuccessMessage && res.data.msg) {
      showSuccess(res.data.msg)
    }

    return res.data.data as T
  } catch (error) {
    if (error instanceof HttpError && error.code !== ApiStatus.unauthorized) {
      const showMsg = config.showErrorMessage !== false
      showError(error, showMsg)
    }
    return Promise.reject(error)
  }
}

/** API方法集合 */
const api = {
  get<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'GET' })
  },
  post<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'POST' })
  },
  put<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'PUT' })
  },
  del<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'DELETE' })
  },
  request<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>(config)
  }
}

export default api
