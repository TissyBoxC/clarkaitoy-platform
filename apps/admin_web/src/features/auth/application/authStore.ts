import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { createHttpClient } from '@/api/httpClient'
import { SessionStorageAuthSessionStore } from '@/api/authSession'
import { mapApiError, type ApiError } from '@/api/apiError'

/// One administrator session returned by the device platform.
export interface AdminAccount {
  id: string
  email: string
  displayName: string
  role: string
}

/// Owns administrator login, session restoration, and logout.
export const useAuthStore = defineStore('admin-auth', () => {
  const sessionStore = new SessionStorageAuthSessionStore()
  const httpClient = createHttpClient({ sessionStore })
  const account = ref<AdminAccount | null>(null)
  const isRestoring = ref(true)
  const error = ref<ApiError | null>(null)

  const isSignedIn = computed(() => account.value !== null)

  async function restoreSession(): Promise<void> {
    isRestoring.value = true
    error.value = null
    try {
      if (sessionStore.readRefreshToken() === null) {
        account.value = null
        return
      }
      const response = await httpClient.get('/api/v1/auth/me')
      account.value = toAccount(response.data.data.account)
    } catch (caught: unknown) {
      sessionStore.clearSession()
      account.value = null
    } finally {
      isRestoring.value = false
    }
  }

  async function login(email: string, password: string): Promise<boolean> {
    error.value = null
    try {
      const response = await httpClient.post('/api/v1/auth/login', {
        email,
        password,
      })
      const data = response.data.data
      sessionStore.saveSession(data.access_token, data.refresh_token)
      account.value = toAccount(data.account)
      if (account.value.role !== 'admin') {
        sessionStore.clearSession()
        account.value = null
        error.value = {
          kind: 'insufficient_permission',
          message: '你没有权限访问此页面',
          retryable: false,
        }
        return false
      }
      return true
    } catch (caught: unknown) {
      error.value = mapApiError(caught)
      return false
    }
  }

  async function logout(): Promise<void> {
    const refreshToken = sessionStore.readRefreshToken()
    try {
      if (refreshToken !== null) {
        await httpClient.post('/api/v1/auth/logout', {
          refresh_token: refreshToken,
        })
      }
    } catch {
      // Local sign-out still succeeds when the network is unavailable.
    } finally {
      sessionStore.clearSession()
      account.value = null
    }
  }

  return {
    account,
    error,
    isRestoring,
    isSignedIn,
    login,
    logout,
    restoreSession,
  }
})

function toAccount(value: {
  id: string
  email: string
  display_name: string
  role: string
}): AdminAccount {
  return {
    id: value.id,
    email: value.email,
    displayName: value.display_name,
    role: value.role,
  }
}
