/// Runtime configuration for the operations console.
export interface RuntimeConfig {
  apiBaseUrl: string
  requestTimeoutMs: number
  adminBuildVersion: string
}

/// Reads Vite environment values with a same-origin default.
///
/// The production console is served behind an Nginx /api proxy. Keeping the
/// default relative avoids CORS and makes the image portable across hosts.
export function getRuntimeConfig(): RuntimeConfig {
  return {
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '',
    requestTimeoutMs: Number(import.meta.env.VITE_API_TIMEOUT_MS ?? 15000),
    adminBuildVersion: import.meta.env.VITE_APP_VERSION ?? '0.0.0',
  }
}
