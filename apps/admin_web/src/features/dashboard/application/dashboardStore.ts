import { defineStore } from 'pinia'
import { ref } from 'vue'

import { mapApiError, type ApiError } from '@/api/apiError'
import { createHttpClient } from '@/api/httpClient'

export interface RecentRelease {
  version: string
  channel: string
  kind: string
  platform: string
  status: string
  publishedAt: string
}

export interface DashboardOverview {
  parentAccountCount: number
  activeDeviceCount: number
  onlineDeviceCount: number
  todayConversationCount: number
  todaySpentUsd: number
  totalBalanceUsd: number
  pendingReleaseCount: number
  recentReleases: RecentRelease[]
}

// 首页只读取聚合结果，避免浏览器分别请求多个统计接口造成数字不一致。
export const useDashboardStore = defineStore('admin-dashboard', () => {
  const httpClient = createHttpClient()
  const overview = ref<DashboardOverview | null>(null)
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const response = await httpClient.get('/api/v1/admin/overview')
      overview.value = toOverview(response.data.data ?? {})
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  return { error, isLoading, overview, load }
})

function toOverview(value: Record<string, unknown>): DashboardOverview {
  const releases = Array.isArray(value.recent_releases)
    ? value.recent_releases
    : []

  return {
    parentAccountCount: numberValue(value.parent_account_count),
    activeDeviceCount: numberValue(value.active_device_count),
    onlineDeviceCount: numberValue(value.online_device_count),
    todayConversationCount: numberValue(value.today_conversation_count),
    todaySpentUsd: numberValue(value.today_spent_usd),
    totalBalanceUsd: numberValue(value.total_balance_usd),
    pendingReleaseCount: numberValue(value.pending_release_count),
    recentReleases: releases.map(toRecentRelease),
  }
}

function toRecentRelease(value: unknown): RecentRelease {
  const record = isRecord(value) ? value : {}
  return {
    version: stringValue(record.version),
    channel: stringValue(record.channel, 'stable'),
    kind: stringValue(record.kind, 'resource'),
    platform: stringValue(record.platform, 'all'),
    status: stringValue(record.status, 'draft'),
    publishedAt: stringValue(record.published_at),
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function numberValue(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}
