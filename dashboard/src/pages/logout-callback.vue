<template>
  <div>Logging out...</div>
</template>

<script setup lang="ts">
import { onMounted } from "vue"
import { useRouter } from "vue-router"
import { userManager } from "../auth/oidc"
import { useAuthStore } from "../stores/auth"

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  try {
    await userManager.signoutRedirectCallback()
  } catch (err) {
    console.warn("Logout callback error:", err)
  }

  authStore.logout()

  console.log("✅ User logged out")

  router.replace("/")
})
</script>

