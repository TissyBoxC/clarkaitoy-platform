/// Stores the short-lived access token and renewable refresh token outside UI state.
export interface AuthSessionStore {
  readAccessToken(): string | null
  readRefreshToken(): string | null
  saveSession(accessToken: string, refreshToken: string): void
  clearSession(): void
}

/// Keeps tokens in memory until a platform-specific secure store is supplied.
///
/// This default is intentionally non-persistent: an admin must sign in again
/// after a reload rather than leaving a credential in browser storage.
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
