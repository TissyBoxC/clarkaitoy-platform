import { defineStore } from 'pinia'
import { ref } from 'vue'

import { createHttpClient } from '@/api/httpClient'
import { mapApiError, type ApiError } from '@/api/apiError'

/// Platform-safe AI account projection for operations staff.
export interface AdminAiAccount {
  providerAccountId: string
  status: string
  balanceUsd: number
  concurrencyLimit: number
  /** Platform-approved pool the guardian may choose from. */
  availableModels: string[]
  /** Guardian's explicit selection; empty means all available models. */
  selectedModels: string[]
  /** Effective allowlist currently applied to the AI service. */
  allowedModels: string[]
  credentialReady: boolean
  updatedAt: string
}

/// Loads and updates AI account policy through device_platform.
export const useAiAccountStore = defineStore('admin-ai-accounts', () => {
  const httpClient = createHttpClient()
  const accounts = ref<AdminAiAccount[]>([])
  const isLoading = ref(false)
  const error = ref<ApiError | null>(null)

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const response = await httpClient.get('/api/v1/admin/ai-accounts')
      accounts.value = (response.data.data.accounts ?? []).map(toAccount)
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function update(account: AdminAiAccount): Promise<boolean> {
    error.value = null
    try {
      await httpClient.put(
        `/api/v1/admin/ai-accounts/${account.providerAccountId}`,
        {
          status: account.status,
          balance_usd: account.balanceUsd,
          concurrency_limit: account.concurrencyLimit,
          available_models: account.availableModels,
          reason: 'managed from admin console',
        },
      )
      await load()
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    }
  }

  return { accounts, error, isLoading, load, update }
})

function toAccount(value: Record<string, unknown>): AdminAiAccount {
  return {
    providerAccountId: String(value.provider_account_id ?? ''),
    status: String(value.status ?? ''),
    balanceUsd: Number(value.balance_usd ?? 0),
    concurrencyLimit: Number(value.concurrency_limit ?? 1),
    availableModels: modelList(value.available_models),
    selectedModels: modelList(value.selected_models),
    allowedModels: Array.isArray(value.allowed_models)
      ? value.allowed_models.map(String)
      : [],
    credentialReady: value.credential_ready === true,
    updatedAt: String(value.updated_at ?? ''),
  }
}

function modelList(value: unknown): string[] {
  return Array.isArray(value) ? value.map(String) : []
}
