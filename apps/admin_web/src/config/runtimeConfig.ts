/// Runtime configuration for the operations console.
export interface RuntimeConfig {
  apiBaseUrl: string
  requestTimeoutMs: number
}

/// Reads Vite environment values with safe local defaults.
export function getRuntimeConfig(): RuntimeConfig {
  return {
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? 'https://api.example.invalid',
    requestTimeoutMs: Number(import.meta.env.VITE_API_TIMEOUT_MS ?? 15000),
  }
}
