/// Stores admin tokens for the lifetime of one browser tab.
///
/// sessionStorage reduces accidental persistence compared with localStorage
/// while still surviving route changes and reloads in the active console tab.
export interface AuthSessionStore {
  readAccessToken(): string | null
  readRefreshToken(): string | null
  saveSession(accessToken: string, refreshToken: string): void
  clearSession(): void
}

const accessTokenKey = 'sprout.admin.accessToken'
const refreshTokenKey = 'sprout.admin.refreshToken'

/// Browser session storage used by the production admin console.
export class SessionStorageAuthSessionStore implements AuthSessionStore {
  readAccessToken(): string | null {
    return readValue(accessTokenKey)
  }

  readRefreshToken(): string | null {
    return readValue(refreshTokenKey)
  }

  saveSession(accessToken: string, refreshToken: string): void {
    if (typeof window === 'undefined') {
      return
    }
    window.sessionStorage.setItem(accessTokenKey, accessToken)
    window.sessionStorage.setItem(refreshTokenKey, refreshToken)
  }

  clearSession(): void {
    if (typeof window === 'undefined') {
      return
    }
    window.sessionStorage.removeItem(accessTokenKey)
    window.sessionStorage.removeItem(refreshTokenKey)
  }
}

/// Keeps tokens in memory for tests and non-browser runtimes.
export class MemoryAuthSessionStore implements AuthSessionStore {
  private accessToken: string | null = null
  private refreshToken: string | null = null

  readAccessToken(): string | null {
    return this.accessToken
  }

  readRefreshToken(): string | null {
    return this.refreshToken
  }

  saveSession(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken
    this.refreshToken = refreshToken
  }

  clearSession(): void {
    this.accessToken = null
    this.refreshToken = null
  }
}

function readValue(key: string): string | null {
  if (typeof window === 'undefined') {
    return null
  }
  return window.sessionStorage.getItem(key)
}
