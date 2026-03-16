import { UserManager, WebStorageStateStore, User, Log } from "oidc-client-ts"
import { useAuthStore, type UserProfile } from "../stores/auth"

export interface OidcProfile {
  sub: string
  name?: string
  email?: string
  preferred_username?: string
  resource_access?: Record<string, { roles: string[] }>
}

// ⭐ Enable verbose oidc-client logging
Log.setLogger(console)
Log.setLevel(Log.DEBUG)

// ⭐ shared settings
const oidcSettings = {
  authority: "http://localhost:8081/realms/LanParty",
  client_id: "lan-control-plane",

  redirect_uri: window.location.origin + "/callback",
  silent_redirect_uri: window.location.origin + "/silent-renew",

  response_type: "code",
  scope: "openid profile email offline_access",

  post_logout_redirect_uri: window.location.origin,

  automaticSilentRenew: true,
  accessTokenExpiringNotificationTimeInSeconds: 60,
  monitorSession: true,

  userStore: new WebStorageStateStore({
    store: window.localStorage,
  }),
}

// ⭐ single instance used everywhere
export const userManager = new UserManager(oidcSettings)

// ⭐ prevent double initialization
let initialized = false

export async function initOidc() {
  if (initialized) return
  initialized = true

  const authStore = useAuthStore()

  // ⭐ fires on login AND token refresh
  userManager.events.addUserLoaded((user: User) => {
    const expiresIn = user.expires_in ?? 0
    const expTime = new Date(Date.now() + expiresIn * 1000)

    console.log("✅ OIDC USER LOADED / REFRESHED")
    console.log("Access token expires in:", expiresIn, "seconds")
    console.log("Expires at:", expTime.toLocaleTimeString())

    const profile = user.profile as unknown as OidcProfile

    const roles =
      profile.resource_access?.["lan-control-plane"]?.roles ?? []

    const expiresAt =
      Math.floor(Date.now() / 1000) + (user.expires_in ?? 3600)

    authStore.setUser(
      {
        sub: profile.sub,
        name: profile.name,
        email: profile.email,
        preferred_username: profile.preferred_username,
        resource_access: profile.resource_access,
        roles,
      },
      user.access_token,
      user.refresh_token,
      expiresAt,
    )
  })

  // ⭐ when token about to expire
  userManager.events.addAccessTokenExpiring(() => {
    console.log("⚠️ Access token expiring soon — silent renew should start")
  })

  // ⭐ when token fully expired
  userManager.events.addAccessTokenExpired(() => {
    console.log("❌ Access token expired")
  })

  // ⭐ silent renew failed
  userManager.events.addSilentRenewError((err) => {
    console.error("🚨 Silent renew error:", err)
  })

  // ⭐ user removed from storage
  userManager.events.addUserUnloaded(() => {
    console.log("👋 User unloaded from session")
    authStore.logout()
  })

  // ⭐ identity provider logout detected
  userManager.events.addUserSignedOut(() => {
    console.log("🔐 User signed out from IdP")
    authStore.logout()
  })

  // ⭐ load stored user on startup
  const user = await userManager.getUser()

  if (user) {
    console.log("🔁 Restoring existing user session")

    authStore.setUser(
      user.profile as unknown as UserProfile,
      user.access_token,
      user.refresh_token,
      Math.floor(Date.now() / 1000) + (user.expires_in ?? 3600),
    )
  } else {
    console.log("ℹ️ No existing OIDC session found")
  }
}

export async function login() {
  await userManager.removeUser()
  console.log("➡️ Redirecting to login")
  await userManager.signinRedirect()
}

export async function logout() {
  try {
    console.log("➡️ Redirecting to logout")
    await userManager.signoutRedirect({
      post_logout_redirect_uri: window.location.origin + "/logout-callback"
    })
  } catch (err) {
    console.error("Logout redirect failed", err)
    useAuthStore().logout()
  }
}

export async function getToken(): Promise<string | null> {
  const user = await userManager.getUser()

  if (!user) {
    console.log("⚠️ No user session when requesting token")
    return null
  }

  const expiresIn = user.expires_in ?? 0
  console.log("🔑 Token requested, expires in", expiresIn, "seconds")

  return user.access_token ?? null
}
