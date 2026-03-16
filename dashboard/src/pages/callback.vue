<template>
  <div>Logging in...</div>
</template>

<script setup lang="ts">
import { onMounted } from "vue"
import { useRouter } from "vue-router"
import { userManager } from "../auth/oidc"

const router = useRouter()

onMounted(async () => {
  await userManager.signinRedirectCallback()
  router.replace("/dashboard")
})
</script>
