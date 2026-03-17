<template>
  <div class="flex flex-col">
    
    <header class="relative bg-primary text-navbar-text py-32 overflow-hidden flex items-center justify-center">
      <div class="container mx-auto text-center relative z-10 px-4">
        <h1 class="text-6xl md:text-7xl font-extrabold mb-6 drop-shadow-lg">LANForge</h1>
        <p class="text-xl md:text-2xl max-w-3xl mx-auto mb-10 drop-shadow-md opacity-90">
          The ultimate platform to deploy, manage, and monitor your game servers and containers.
          Lightning-fast, secure, and intuitive.
        </p>
        <div v-if="!authStore.isAuthenticated" class="flex justify-center gap-6">
          <button @click="login"
                  class="px-8 py-4 bg-secondary hover:bg-secondary-hover text-navbar-text rounded-xl font-bold text-lg transition-all shadow-lg active:scale-95">
            Login to Get Started
          </button>
        </div>
      </div>

      <div class="absolute top-0 left-0 w-full h-full pointer-events-none">
        <div class="absolute bg-white/10 w-72 h-72 rounded-full -top-16 -left-16 animate-blob"></div>
        <div class="absolute bg-white/20 w-96 h-96 rounded-full -bottom-32 -right-32 animate-blob animation-delay-2000"></div>
        <div class="absolute bg-white/10 w-48 h-48 rounded-full -bottom-16 -left-20 animate-blob animation-delay-4000"></div>
      </div>
    </header>

    <section v-if="!authStore.isAuthenticated" class="py-24">
      <div class="container mx-auto grid md:grid-cols-3 gap-12 text-center px-4">
        
        <div class="p-8 bg-white dark:bg-slate-800/50 border border-transparent dark:border-slate-700 rounded-2xl shadow-xl hover:shadow-2xl transition-all">
          <div class="text-primary mb-4 text-5xl">🚀</div>
          <h3 class="text-2xl font-bold mb-2">Fast Deployments</h3>
          <p class="opacity-80">
            Launch containers and game servers in seconds using ready-to-use templates.
          </p>
        </div>

        <div class="p-8 bg-white dark:bg-slate-800/50 border border-transparent dark:border-slate-700 rounded-2xl shadow-xl hover:shadow-2xl transition-all">
          <div class="text-primary mb-4 text-5xl">📊</div>
          <h3 class="text-2xl font-bold mb-2">Live Monitoring</h3>
          <p class="opacity-80">
            View logs and resource stats in real-time to keep your servers running smoothly.
          </p>
        </div>

        <div class="p-8 bg-white dark:bg-slate-800/50 border border-transparent dark:border-slate-700 rounded-2xl shadow-xl hover:shadow-2xl transition-all">
          <div class="text-primary mb-4 text-5xl">🔒</div>
          <h3 class="text-2xl font-bold mb-2">Secure Access</h3>
          <p class="opacity-80">
            Built-in OIDC authentication with roles ensures your servers are always safe.
          </p>
        </div>
      </div>
    </section>

    <footer class="bg-navbar-bg text-navbar-text py-8 text-center mt-auto">
      <p>© 2026 LANForge. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "../stores/auth";
import { watchEffect } from "vue";
import { login as loginOidc } from "../auth/oidc";
import { useRouter } from "vue-router";

const authStore = useAuthStore();
const router = useRouter();

const login = async () => {
  await loginOidc();
};

watchEffect(() => {
  if (authStore.isAuthenticated) {
    router.push("/dashboard");
  }
});
</script>

<style scoped>
@keyframes blob {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -20px) scale(1.1); }
  66% { transform: translate(-20px, 20px) scale(0.9); }
}
.animate-blob {
  animation: blob 8s infinite ease-in-out;
}
.animation-delay-2000 { animation-delay: 2s; }
.animation-delay-4000 { animation-delay: 4s; }
</style>