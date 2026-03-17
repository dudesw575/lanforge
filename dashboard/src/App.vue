<template>
  <div v-if="!authStore.ready" class="flex items-center justify-center min-h-screen bg-bg text-primary">
     <div class="flex flex-col items-center gap-4">
        <div class="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
        <p class="font-mono text-sm tracking-widest uppercase opacity-50">Initializing Forge...</p>
     </div>
  </div>

  <DefaultLayout v-else />
</template>

<script setup lang="ts">
import { useAuthStore } from "./stores/auth"
import DefaultLayout from "./layouts/DefaultLayout.vue"

const authStore = useAuthStore()
</script>