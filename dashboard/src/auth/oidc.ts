import { UserManager, WebStorageStateStore, User, Log } from "oidc-client-ts"
import { useAuthStore, type UserProfile } from "../stores/auth"

export interface OidcProfile {
  sub: string
  name?: string
  email?: string
  preferred_username?: string
  resource_access?: Record<string, { roles: string[] }>
}

// Log.setLogger(console)
// Log.setLevel(Log.DEBUG)

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

  userStore: new WebStorageStateStore({ store: window.localStorage }),
}

export const userManager = new UserManager(oidcSettings)
let initialized = false

export async function initOidc() {
  if (initialized) return
  initialized = true

  const authStore = useAuthStore()

  userManager.events.addAccessTokenExpired(() => {
    console.log("❌ Token expired — logging out")
    authStore.logout()
  })

  userManager.events.addSilentRenewError((err) => {
    console.error("🚨 Silent renew error", err)
    authStore.logout()
  })

  userManager.events.addUserLoaded((user: User) => {
    const expiresAt = Math.floor(Date.now() / 1000) + (user.expires_in ?? 3600)
    const profile = user.profile as unknown as OidcProfile
    const roles = profile.resource_access?.["lan-control-plane"]?.roles ?? []

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
      expiresAt
    )
  })

  try {
    const user = await userManager.getUser()
    if (user && (user.expires_in ?? 0) > 0) {
      console.log("🔁 Restoring existing user session")
      const expiresAt = Math.floor(Date.now() / 1000) + (user.expires_in ?? 3600)
      authStore.setUser(
        user.profile as unknown as UserProfile,
        user.access_token,
        user.refresh_token,
        expiresAt
      )
    } else {
      console.log("ℹ️ No valid OIDC session")
      authStore.logout()
    }
  } finally {
    authStore.ready = true
    console.log("✅ Auth initialization complete")
  }
}

export async function login() {
  await userManager.removeUser()
  console.log("➡️ Redirecting to login")
  await userManager.signinRedirect()
}

export async function logout() {
  const authStore = useAuthStore()
  try {
    console.log("➡️ Redirecting to logout")
    await userManager.signoutRedirect({
      post_logout_redirect_uri: window.location.origin + "/logout-callback",
    })
  } catch (err) {
    console.error("Logout redirect failed", err)
    authStore.logout()
  }
}

export async function getToken(): Promise<string | null> {
  const authStore = useAuthStore()
  if (!authStore.ready) await initOidc()

  const user = await userManager.getUser()
  if (!user) return null
  if ((user.expires_in ?? 0) <= 0) return null

  return user.access_token ?? null
}