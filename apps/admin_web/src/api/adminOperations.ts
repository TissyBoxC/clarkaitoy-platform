import type { AxiosInstance } from 'axios'

import { createHttpClient } from '@/api/httpClient'

export interface AIModelOption {
  id: string
  latencyMs: number | null
}

export interface AIAccountDefaults {
  defaultBalanceUsd: number | null
  defaultConcurrencyLimit: number | null
  source: string
}

export interface ReleaseArtifact {
  version: string
  kind: string
  platform: string
  downloadUrl: string
  sha256: string
}

export interface AdminOperationsClient {
  loadAIAccountDefaults(): Promise<AIAccountDefaults | null>
  loadAIModels(): Promise<AIModelOption[]>
  loadReleaseArtifact(
    version: string,
    kind: string,
    platform: string,
  ): Promise<ReleaseArtifact | null>
}

/// Creates the client for backend-owned AI defaults, model telemetry, and
/// release artifact lookup. The UI never guesses provider credentials or files.
export function createAdminOperationsClient(
  httpClient: AxiosInstance = createHttpClient(),
): AdminOperationsClient {
  return {
    async loadAIAccountDefaults(): Promise<AIAccountDefaults | null> {
      const response = await httpClient.get('/api/v1/admin/ai-account-defaults')
      const value = recordValue(response.data?.data)
      if (Object.keys(value).length === 0) {
        return null
      }
      return {
        defaultBalanceUsd: nullableNumber(
          value.default_balance_usd ?? value.defaultBalanceUsd,
        ),
        defaultConcurrencyLimit: nullableNumber(
          value.default_concurrency ?? value.defaultConcurrencyLimit,
        ),
        source: stringValue(value.source, 'sub2api'),
      }
    },

    async loadAIModels(): Promise<AIModelOption[]> {
      const response = await httpClient.get('/api/v1/admin/ai-models')
      const records = response.data?.data?.models
      if (!Array.isArray(records)) {
        return []
      }
      return records
        .map(toAIModelOption)
        .filter((model): model is AIModelOption => model !== null)
    },

    async loadReleaseArtifact(
      version: string,
      kind: string,
      platform: string,
    ): Promise<ReleaseArtifact | null> {
      const response = await httpClient.get(
        `/api/v1/admin/release-artifacts/${encodeURIComponent(version.trim())}`,
        { params: { kind, platform } },
      )
      const value = recordValue(response.data?.data?.artifact)
      if (Object.keys(value).length === 0) {
        return null
      }
      return {
        version: stringValue(value.version, version.trim()),
        kind: stringValue(value.kind, kind),
        platform: stringValue(value.platform, platform),
        downloadUrl: stringValue(value.download_url ?? value.downloadUrl),
        sha256: stringValue(value.sha256),
      }
    },
  }
}

function toAIModelOption(value: unknown): AIModelOption | null {
  const record = recordValue(value)
  const id = stringValue(record.id ?? record.model ?? record.name)
  if (!id) {
    return null
  }
  return {
    id,
    latencyMs: nullableNumber(record.latency_ms ?? record.latencyMs),
  }
}

function recordValue(value: unknown): Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function nullableNumber(value: unknown): number | null {
  if (value === null || value === undefined || value === '') {
    return null
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : null
}
