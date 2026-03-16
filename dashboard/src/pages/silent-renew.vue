<template>
  <div></div>
</template>

<script setup lang="ts">
import { onMounted } from "vue"
import { userManager } from "../auth/oidc"

onMounted(async () => {
  await userManager.signinSilentCallback()
})
</script>

