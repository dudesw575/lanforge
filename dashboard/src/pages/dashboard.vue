<template>
  <div class="container mx-auto flex justify-center py-6">
    <div class="w-full max-w-3xl bg-gray-900 dark:bg-gray-800 text-green-400 text-xs p-4 rounded-md font-mono">
      <div class="mb-2"><strong>Auth Debug Panel</strong></div>
      <div>User: {{ authStore.user?.preferred_username || authStore.user?.sub }}</div>
      <div>Roles: {{ userRoles.join(", ") || "none" }}</div>
      <div>Authenticated: {{ authStore.authenticated }}</div>
      <div>Token expires in: {{ tokenExpiresIn }}s</div>
      <div v-if="tokenExpiresIn < 60" class="text-yellow-400">⚠️ Token expiring soon</div>
      <div v-if="tokenExpiresIn <= 0" class="text-red-400">❌ Token expired</div>
      <div>Token: {{ shortToken }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useAuthStore } from "../stores/auth";

const authStore = useAuthStore();
const userRoles = computed(
  () => authStore.user?.resource_access?.["lan-control-plane"]?.roles ?? []
);

const tokenExpiresIn = ref(0);
const shortToken = computed(() => authStore.token?.substring(0, 20) + "..." || "none");

let tokenTimer: ReturnType<typeof setInterval> | null = null;
onMounted(() => {
  tokenTimer = setInterval(() => {
    if (!authStore.expiresAt) return;
    tokenExpiresIn.value = authStore.expiresAt - Math.floor(Date.now() / 1000);
  }, 1000);
});
onUnmounted(() => {
  if (tokenTimer) clearInterval(tokenTimer);
});
</script>
