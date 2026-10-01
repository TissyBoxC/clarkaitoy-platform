import axios, { type AxiosInstance } from 'axios'

import { getRuntimeConfig } from '@/config/runtimeConfig'
import { installAuthInterceptors, type RefreshSession } from '@/api/authRefresh'
import { MemoryAuthSessionStore, type AuthSessionStore } from '@/api/authSession'

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

  const sessionStore = options?.sessionStore ?? new MemoryAuthSessionStore()
  const refreshSession =
    options?.refreshSession ??
    (async () => {
      // The refresh endpoint is wired when auth API contracts are implemented.
      throw new Error('session refresh is not configured')
    })

  installAuthInterceptors(httpClient, sessionStore, refreshSession)
  return httpClient
}
