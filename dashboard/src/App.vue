<template>
  <DefaultLayout />
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { initOidc } from "./auth/oidc"
import DefaultLayout from "./layouts/DefaultLayout.vue"

const ready = ref(false)

onMounted(async () => {
  await initOidc()
  ready.value = true
})
</script>
