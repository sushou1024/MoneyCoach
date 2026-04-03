import { Platform } from 'react-native'

import { getCurrentLocale } from '../utils/i18n'

type RefreshTokenHandler = () => Promise<string | null>

let refreshTokenHandler: RefreshTokenHandler | null = null
let refreshInFlight: Promise<string | null> | null = null

export function setRefreshTokenHandler(handler: RefreshTokenHandler | null) {
  refreshTokenHandler = handler
}

async function refreshAccessTokenOnce() {
  if (!refreshTokenHandler) return null
  if (!refreshInFlight) {
    refreshInFlight = refreshTokenHandler()
      .catch(() => null)
      .finally(() => {
        refreshInFlight = null
      })
  }
  return refreshInFlight
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, any>
}

export interface ApiResponse<T> {
  status: number
  data?: T
  error?: ApiError
}

const DEFAULT_IOS_BASE = 'http://localhost:8080'
const DEFAULT_ANDROID_BASE = 'http://10.0.2.2:8080'

export function getApiBaseUrl() {
  const base = process.env.EXPO_PUBLIC_API_BASE_URL
  const androidBase = process.env.EXPO_PUBLIC_API_BASE_URL_ANDROID
  if (Platform.OS === 'android' && androidBase) {
    return androidBase
  }
  if (base) {
    return base
  }
  return Platform.OS === 'android' ? DEFAULT_ANDROID_BASE : DEFAULT_IOS_BASE
}

function randomHex(size: number) {
  let result = ''
  for (let i = 0; i < size; i += 1) {
    result += Math.floor(Math.random() * 16).toString(16)
  }
  return result
}

function createIdempotencyKey() {
  return `${randomHex(8)}-${randomHex(4)}-${randomHex(4)}-${randomHex(4)}-${randomHex(12)}`
}

export async function apiRequest<T>(
  path: string,
  options: {
    method?: string
    body?: any
    token?: string
    headers?: Record<string, string>
    idempotencyKey?: string
  } = {}
): Promise<ApiResponse<T>> {
  const url = `${getApiBaseUrl()}${path}`
  const method = options.method ?? 'GET'
  const baseHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    'Accept-Language': getCurrentLocale(),
    ...options.headers,
  }
  if (method !== 'GET' && method !== 'HEAD' && !baseHeaders['Idempotency-Key']) {
    baseHeaders['Idempotency-Key'] = options.idempotencyKey ?? createIdempotencyKey()
  }
  const body = options.body ? JSON.stringify(options.body) : undefined

  const doRequest = async (token?: string): Promise<ApiResponse<T>> => {
    const headers: Record<string, string> = { ...baseHeaders }
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    let response: Response
    try {
      response = await fetch(url, {
        method,
        headers,
        body,
      })
    } catch (err) {
      return {
        status: 0,
        error: {
          code: 'NETWORK_ERROR',
          message: err instanceof Error ? err.message : 'Network request failed',
        },
      }
    }

    const contentType = response.headers.get('content-type') ?? ''
    const isJSON = contentType.includes('application/json')
    const data = isJSON ? await response.json() : undefined

    if (data && typeof data === 'object' && 'error' in data) {
      const payload = data as { error?: ApiError }
      if (payload.error) {
        return { status: response.status, error: payload.error }
      }
    }

    if (!response.ok) {
      const error = (data && (data as { error?: ApiError }).error) ?? {
        code: 'UNKNOWN_ERROR',
        message: 'Request failed',
      }
      return { status: response.status, error }
    }

    return { status: response.status, data }
  }

  const first = await doRequest(options.token)
  if (first.status === 401 && options.token && refreshTokenHandler) {
    const refreshed = await refreshAccessTokenOnce()
    if (refreshed && refreshed !== options.token) {
      return doRequest(refreshed)
    }
  }
  return first
}
