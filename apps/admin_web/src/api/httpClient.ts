import axios, { type AxiosInstance } from 'axios'

import { getRuntimeConfig } from '@/config/runtimeConfig'
import { installAuthInterceptors, type RefreshSession } from '@/api/authRefresh'
import {
  SessionStorageAuthSessionStore,
  type AuthSessionStore,
} from '@/api/authSession'

/// Creates the shared HTTP client used by feature API modules.
export function createHttpClient(options?: {
  sessionStore?: AuthSessionStore
  refreshSession?: RefreshSession
}): AxiosInstance {
  const runtimeConfig = getRuntimeConfig()

  const httpClient = axios.create({
    baseURL: runtimeConfig.apiBaseUrl,
    timeout: runtimeConfig.requestTimeoutMs,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  const sessionStore =
    options?.sessionStore ?? new SessionStorageAuthSessionStore()
  const refreshSession =
    options?.refreshSession ??
    (async (refreshToken: string) => {
      const response = await axios.post(
        `${runtimeConfig.apiBaseUrl}/api/v1/auth/refresh`,
        { refresh_token: refreshToken },
        {
          timeout: runtimeConfig.requestTimeoutMs,
          headers: { 'Content-Type': 'application/json' },
        },
      )
      const data = response.data?.data
      return {
        accessToken: String(data?.access_token ?? ''),
        refreshToken: String(data?.refresh_token ?? ''),
      }
    })

  installAuthInterceptors(httpClient, sessionStore, refreshSession)
  return httpClient
}
