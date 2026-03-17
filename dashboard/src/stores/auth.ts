import { defineStore } from "pinia"

export interface UserProfile {
  sub: string
  preferred_username?: string
  name?: string
  email?: string
  resource_access?: Record<string, { roles: string[] }>
  roles?: string[]
}

interface AuthState {
  authenticated: boolean
  user: UserProfile | null
  token: string | null
  refreshToken: string | null
  expiresAt: number | null
  ready: boolean
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState & { ready: boolean } => ({
    authenticated: false,
    user: null,
    token: null,
    refreshToken: null,
    expiresAt: null,
    ready: false,
  }),

  getters: {
    isAuthenticated(state): boolean {
      if (!state.authenticated || !state.token || !state.expiresAt) return false
      const now = Math.floor(Date.now() / 1000)
      return now < state.expiresAt
    },
  },

  actions: {
    setUser(
      user: UserProfile,
      token: string,
      refreshToken?: string,
      expiresAt?: number,
    ) {
      this.user = user
      this.token = token
      this.refreshToken = refreshToken ?? null
      this.expiresAt = expiresAt ?? null

      const now = Math.floor(Date.now() / 1000)
      this.authenticated = !!token && expiresAt ? now < expiresAt : false
    },

    logout() {
      this.user = null
      this.token = null
      this.refreshToken = null
      this.expiresAt = null
      this.authenticated = false
    },
  },

  persist: true,
})