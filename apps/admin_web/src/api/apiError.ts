import axios from 'axios'

import type { APIResponseEnvelope } from '../../../../packages/contracts/generated/typescript/envelope'

/// Categories let pages choose a safe next action without reading raw errors.
export type ApiErrorKind =
  | 'unauthenticated'
  | 'insufficient_permission'
  | 'not_found'
  | 'validation'
  | 'rate_limited'
  | 'network'
  | 'service_unavailable'
  | 'unexpected'

/// Mapped error safe for logs and UI state.
export interface ApiError {
  kind: ApiErrorKind
  message: string
  retryable: boolean
}

/// Maps transport and contract errors to stable user-facing categories.
export function mapApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status
    const envelope = error.response?.data as APIResponseEnvelope | undefined
    const code = envelope?.error?.code

    if (status === 401) {
      return { kind: 'unauthenticated', message: '登录已过期，请重新登录', retryable: false }
    }
    if (status === 403) {
      return { kind: 'insufficient_permission', message: '你没有权限访问此页面', retryable: false }
    }
    if (status === 404) {
      return { kind: 'not_found', message: '没有找到这条内容', retryable: false }
    }
    if (status === 422 || code === 'invalid_request') {
      return { kind: 'validation', message: '请检查填写的内容', retryable: false }
    }
    if (status === 429) {
      return { kind: 'rate_limited', message: '操作太频繁，请稍后重试', retryable: true }
    }
    if (!error.response) {
      return { kind: 'network', message: '网络连接不稳定，请检查后重试', retryable: true }
    }
    if (status !== undefined && status >= 500) {
      return { kind: 'service_unavailable', message: '服务暂时不可用，请稍后重试', retryable: true }
    }
  }

  return { kind: 'unexpected', message: '操作没有完成，请稍后重试', retryable: true }
}
