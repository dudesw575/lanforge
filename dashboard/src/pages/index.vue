

<script setup lang="ts">
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { login as oidcLogin } from "../auth/oidc";

const authStore = useAuthStore();
const router = useRouter();

const login = async () => {
  await oidcLogin();
};

const goDashboard = () => router.push("/dashboard");
</script>

<template>
  <div class="container mx-auto text-center py-20">
    <h2 class="text-3xl mb-6">Welcome to LAN Party Control</h2>
    <p class="mb-8">Manage your containers and monitor your LAN Party environment.</p>

    <button
      v-if="!authStore.authenticated"
      @click="login"
      class="px-6 py-3 rounded text-white"
      :style="{ backgroundColor: 'var(--color-primary)' }"
    >
      Login
    </button>

    <button
      v-else
      @click="goDashboard"
      class="px-6 py-3 rounded text-white"
      :style="{ backgroundColor: 'var(--color-secondary)' }"
    >
      Go to Dashboard
    </button>
  </div>
</template>


