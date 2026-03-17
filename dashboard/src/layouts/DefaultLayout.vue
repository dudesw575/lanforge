<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router"; // Import useRoute
import { useAuthStore } from "../stores/auth";
import { login as oidcLogin, logout as oidcLogout } from "../auth/oidc";
import Sidebar from "../components/Sidebar.vue";

const authStore = useAuthStore();
const route = useRoute(); // Access current route context
const theme = ref<"light" | "dark">("light");
const isSidebarOpen = ref(true);
const searchQuery = ref(""); 

const showDashboardFurniture = computed(() => {
  return authStore.isAuthenticated && route.path !== "/";
});

const isAdmin = computed(
  () => authStore.user?.resource_access?.["lan-control-plane"]?.roles?.includes("admin") ?? false,
);

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
  if (window.innerWidth < 1024) isSidebarOpen.value = false;
});

watch(theme, (val) => localStorage.setItem("theme", val));
const themeClass = computed(() => (theme.value === "dark" ? "dark" : "light"));
</script>

<template>
  <div :class="themeClass" class="min-h-screen flex bg-bg text-text transition-colors duration-300">
    
    <Sidebar 
      v-if="showDashboardFurniture" 
      :is-open="isSidebarOpen" 
      @close="isSidebarOpen = false" 
    />

    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
      
      <header 
        v-if="showDashboardFurniture" 
        class="h-16 flex justify-between items-center px-6 z-40 border-b border-current/10 bg-bg/80 backdrop-blur-md shrink-0"
      >
        <div class="flex items-center gap-6 flex-1">

          <div class="hidden md:flex relative w-full max-w-md items-center group">
            <span class="absolute left-3 opacity-40 group-focus-within:opacity-100 transition-opacity">🔍</span>
            <input 
              v-model="searchQuery"
              type="text"
              placeholder="Search containers..."
              class="w-full bg-current/5 border border-current/10 rounded-xl py-2 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary/40 transition-all outline-none"
            />
          </div>
          
          <span class="font-bold text-lg lg:hidden shrink-0">LANForge</span>
        </div>
        
        <nav class="flex gap-4 items-center ml-4">
          <template v-if="authStore.isAuthenticated">
            <div class="hidden sm:flex flex-col items-end">
              <span class="text-[10px] opacity-50 font-black uppercase tracking-widest leading-none mb-1">User</span>
              <span class="text-sm font-bold leading-none">{{ authStore.user?.preferred_username }}</span>
            </div>
            
            <span v-if="isAdmin" class="bg-red-600/10 text-red-500 border border-red-500/30 px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-tighter">
              Admin
            </span>
          </template>
          
          <div class="h-6 w-px bg-current/10 mx-1 hidden sm:block"></div>

          <button @click="toggleTheme" class="p-2 rounded-lg border border-current/10 hover:bg-current/5 transition-colors" title="Toggle Theme">
            {{ theme === 'dark' ? "☀️" : "🌙" }}
          </button>

          <button 
            v-if="authStore.isAuthenticated" 
            @click="oidcLogout" 
            class="bg-secondary/10 text-secondary border border-secondary/20 px-4 py-1.5 rounded-lg text-sm font-bold hover:bg-secondary hover:text-white transition-all active:scale-95"
          >
            Logout
          </button>
          
          <button 
            v-else 
            @click="oidcLogin" 
            class="bg-primary hover:bg-primary-hover text-white px-6 py-1.5 rounded-lg text-sm font-bold shadow-lg shadow-primary/20 transition-all active:scale-95"
          >
            Login
          </button>
        </nav>
      </header>

      <main 
        class="flex-1 overflow-y-auto"
        :class="showDashboardFurniture ? 'p-4 md:p-8' : 'p-0'"
      >
        <div :class="showDashboardFurniture ? 'max-w-7xl mx-auto' : 'w-full h-full'">
          <router-view :search-query="searchQuery" />
        </div>
      </main>
    </div>

    <Transition enter-active-class="transition-opacity duration-300" enter-from-class="opacity-0" leave-active-class="transition-opacity duration-300" leave-to-class="opacity-0">
      <div v-if="showDashboardFurniture && isSidebarOpen" @click="isSidebarOpen = false" class="fixed inset-0 bg-black/50 z-40 lg:hidden backdrop-blur-sm"></div>
    </Transition>
  </div>
</template>