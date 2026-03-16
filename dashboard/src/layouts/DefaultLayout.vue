<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useAuthStore } from "../stores/auth";
import { login as oidcLogin, logout as oidcLogout } from "../auth/oidc";

const authStore = useAuthStore();
const theme = ref<"light" | "dark">("light");

const isAdmin = computed(
  () =>
    authStore.user?.resource_access?.["lan-control-plane"]?.roles?.includes(
      "admin",
    ) ?? false,
);

const login = async () => await oidcLogin();
const logout = async () => await oidcLogout();

const toggleTheme = () => {
  theme.value = theme.value === "dark" ? "light" : "dark";
  document.documentElement.classList.toggle("dark", theme.value === "dark");
};

onMounted(() => {
  const saved = localStorage.getItem("theme") as "light" | "dark" | null;
  if (saved) {
    theme.value = saved;
    document.documentElement.classList.toggle("dark", saved === "dark");
  }
});

watch(theme, (val) => localStorage.setItem("theme", val));

const themeClass = computed(() => (theme.value === "dark" ? "dark" : "light"));
</script>

<template>
  <div :class="themeClass" class="min-h-screen flex flex-col">
    <header
      class="fixed top-0 left-0 w-full h-16 flex justify-between items-center px-6 z-50 shadow-md"
      :style="{
        backgroundColor: 'var(--color-navbar-bg)',
        color: 'var(--color-navbar-text)',
      }"
    >
      <router-link to="/" class="text-xl font-bold hover:opacity-80">
        LANForge
      </router-link>
      <nav class="flex gap-4 items-center">
        <template v-if="authStore.authenticated">
          <span
            >Welcome,
            {{
              authStore.user?.preferred_username || authStore.user?.sub
            }}</span
          >
          <span v-if="isAdmin" class="bg-red-600 px-2 py-1 rounded text-xs"
            >Admin</span
          >
          <button
            @click="logout"
            class="px-3 py-1 rounded"
            :style="{
              backgroundColor: 'var(--color-secondary)',
              color: 'var(--color-navbar-text)',
            }"
          >
            Logout
          </button>
        </template>
        <template v-else>
          <button
            @click="login"
            class="px-3 py-1 rounded"
            :style="{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-navbar-text)',
            }"
          >
            Login
          </button>
        </template>
        <button @click="toggleTheme" class="px-3 py-1 rounded border ml-2">
          {{ theme === "dark" ? "☀️ Light" : "🌙 Dark" }}
        </button>
      </nav>
    </header>

    <div class="flex-1 flex flex-col mt-18">
      <main class="flex-1 container mx-auto p-4 space-y-6 overflow-auto">
        <router-view />
      </main>
    </div>
  </div>
</template>
