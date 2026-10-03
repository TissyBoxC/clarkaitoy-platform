import { defineStore } from 'pinia'
import { ref } from 'vue'

import { mapApiError, type ApiError } from '@/api/apiError'
import { createHttpClient } from '@/api/httpClient'

export interface AdminFamilyAiAccount {
  providerAccountId: string
  status: string
  balanceUsd: number
  concurrencyLimit: number
  availableModels: string[]
  selectedModels: string[]
  allowedModels: string[]
  credentialReady: boolean
  updatedAt: string
}

export interface AdminFamily {
  parentAccountId: string
  phone: string
  parentDisplayName: string
  guardianFamilyName: string
  childNickname: string
  childBirthday: string
  status: string
  createdAt: string
  aiAccount: AdminFamilyAiAccount | null
}

export interface CreateAdminFamilyInput {
  phone: string
  password: string
  guardianFamilyName: string
  childNickname: string
  childBirthday: string
}

// 管理端以家长账号为主，AI 额度与模型只作为该账号的附属信息更新。
export const useFamilyAccountStore = defineStore('admin-families', () => {
  const httpClient = createHttpClient()
  const families = ref<AdminFamily[]>([])
  const isLoading = ref(false)
  const isSubmitting = ref(false)
  const error = ref<ApiError | null>(null)
  const lastMessage = ref('')

  async function load(): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      const response = await httpClient.get('/api/v1/admin/families')
      const payload = response.data.data ?? {}
      const records = payload.accounts ?? payload.families ?? payload.parent_accounts ?? []
      families.value = Array.isArray(records) ? records.map(toFamily) : []
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
    } finally {
      isLoading.value = false
    }
  }

  async function create(input: CreateAdminFamilyInput): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      const response = await httpClient.post('/api/v1/admin/families', {
        phone: input.phone.trim(),
        password: input.password,
        guardian_family_name: input.guardianFamilyName.trim(),
        child_nickname: input.childNickname.trim(),
        child_birthday: input.childBirthday,
      })
      await load()
      const aiAccount = response.data?.data?.ai_account
      lastMessage.value =
        aiAccount == null
          ? '家长账号已创建，AI 服务仍在准备中。可以在列表中重试开通。'
          : '家长账号已创建，家长 AI 服务已开通。'
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  async function retryAiAccount(parentAccountId: string): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.post(
        `/api/v1/admin/families/${encodeURIComponent(parentAccountId)}/ai-account`,
      )
      await load()
      lastMessage.value = '家长 AI 服务已准备完成。'
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  async function updateAiAccount(account: AdminFamilyAiAccount): Promise<boolean> {
    isSubmitting.value = true
    error.value = null
    lastMessage.value = ''
    try {
      await httpClient.put(
        `/api/v1/admin/ai-accounts/${encodeURIComponent(account.providerAccountId)}`,
        {
          status: account.status,
          balance_usd: account.balanceUsd,
          concurrency_limit: account.concurrencyLimit,
          available_models: account.availableModels,
          reason: '由管理端调整家长 AI 账号',
        },
      )
      await load()
      lastMessage.value = '家长账号设置已保存。'
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    error,
    families,
    isLoading,
    isSubmitting,
    lastMessage,
    create,
    load,
    retryAiAccount,
    updateAiAccount,
  }
})

function toFamily(value: Record<string, unknown>): AdminFamily {
  const aiAccountValue =
    value.ai_account ?? value.parent_ai_account ?? value.aiAccount ?? value

  return {
    parentAccountId: stringValue(value.parent_account_id ?? value.id),
    phone: stringValue(value.phone ?? value.phone_number),
    parentDisplayName: stringValue(
      value.display_name ?? value.parent_display_name,
    ),
    guardianFamilyName: stringValue(
      value.guardian_family_name ?? value.parent_family_name,
    ),
    childNickname: stringValue(value.child_nickname),
    childBirthday: stringValue(value.child_birthday),
    status: stringValue(value.status, 'active'),
    createdAt: stringValue(value.created_at),
    aiAccount: toAiAccount(aiAccountValue),
  }
}

function toAiAccount(value: unknown): AdminFamilyAiAccount | null {
  if (!isRecord(value)) {
    return null
  }
  const providerAccountId = stringValue(
    value.provider_account_id ?? value.providerAccountId,
  )
  if (!providerAccountId) {
    return null
  }

  return {
    providerAccountId,
    status: stringValue(value.status, 'active'),
    balanceUsd: numberValue(value.balance_usd ?? value.balanceUsd),
    concurrencyLimit: numberValue(
      value.concurrency_limit ?? value.concurrencyLimit,
      1,
    ),
    availableModels: modelList(value.available_models ?? value.availableModels),
    selectedModels: modelList(value.selected_models ?? value.selectedModels),
    allowedModels: modelList(value.allowed_models ?? value.allowedModels),
    credentialReady:
      value.credential_ready === true || value.credentialReady === true,
    updatedAt: stringValue(value.updated_at ?? value.updatedAt),
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function numberValue(value: unknown, fallback = 0): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function modelList(value: unknown): string[] {
  return Array.isArray(value) ? value.map(String) : []
}
