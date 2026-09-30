import axios, { type AxiosInstance } from 'axios'

import { getRuntimeConfig } from '@/config/runtimeConfig'

/// Creates the shared HTTP client used by feature API modules.
export function createHttpClient(): AxiosInstance {
  const runtimeConfig = getRuntimeConfig()

  return axios.create({
    baseURL: runtimeConfig.apiBaseUrl,
    timeout: runtimeConfig.requestTimeoutMs,
    headers: {
      'Content-Type': 'application/json',
    },
  })
}
