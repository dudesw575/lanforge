<template>
  <div class="flex flex-col min-h-screen">
    
    <header class="relative bg-primary text-navbar-text py-24 md:py-32 overflow-hidden flex items-center justify-center">
      <div class="container mx-auto text-center relative z-10 px-4">
        <h1 class="text-6xl md:text-7xl font-extrabold mb-6 drop-shadow-lg">LANForge</h1>
        <p class="text-xl md:text-2xl max-w-3xl mx-auto mb-10 drop-shadow-md opacity-90">
          The ultimate platform to deploy, manage, and monitor your game servers and containers.
          Lightning-fast, secure, and intuitive.
        </p>

        <div v-if="!authStore.isAuthenticated" class="flex flex-col items-center gap-8 max-w-2xl mx-auto">
          <!-- 🛠️ 5th Grader SSL Setup Banner -->
          <div class="w-full bg-white/10 backdrop-blur-md border border-white/20 p-6 rounded-2xl text-left shadow-xl">
            <div class="flex items-start gap-4">
              <span class="text-3xl">🔒</span>
              <div>
                <h3 class="text-xl font-bold mb-1 text-secondary">First Time Setup Required</h3>
                <p class="text-sm opacity-90 mb-4">
                  Because LANForge runs locally on your network, you need to trust its local security certificate before logging in.
                </p>
                <div class="flex flex-wrap gap-4 items-center">
                  <a href="https://home.arpa/cert" 
                     download="lanforge-root.crt"
                     class="px-5 py-2.5 bg-white text-primary hover:bg-slate-100 font-bold rounded-xl text-sm transition-all shadow active:scale-95 flex items-center gap-2">
                    📥 Download Certificate
                  </a>
                  <button @click="showInstructions = !showInstructions" 
                          class="text-sm underline hover:text-secondary transition font-medium">
                    {{ showInstructions ? 'Hide Instructions' : 'How do I install this?' }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Expandable Instructions Drawer -->
            <transition name="slide-fade">
              <div v-if="showInstructions" class="mt-6 pt-4 border-t border-white/10 text-sm grid md:grid-cols-2 gap-4 opacity-95">
                <div class="bg-black/20 p-4 rounded-xl">
                  <p class="font-bold text-secondary mb-1">🪟 Windows Setup</p>
                  <ol class="list-decimal list-inside space-y-1 opacity-90 text-xs">
                    <li>Double-click the downloaded file.</li>
                    <li>Click <strong>Install Certificate...</strong></li>
                    <li>Select <strong>Local Machine</strong>, then Next.</li>
                    <li>Choose <strong>Place all certificates in the following store</strong>.</li>
                    <li>Browse and choose <strong>Trusted Root Certification Authorities</strong>.</li>
                  </ol>
                </div>
                <div class="bg-black/20 p-4 rounded-xl">
                  <p class="font-bold text-secondary mb-1">🍏 Mac Setup</p>
                  <ol class="list-decimal list-inside space-y-1 opacity-90 text-xs">
                    <li>Double-click the downloaded file to open Keychain Access.</li>
                    <li>Locate <strong>caddy-root</strong> or <strong>LANForge</strong> under certificates.</li>
                    <li>Double-click it, expand the <strong>Trust</strong> section.</li>
                    <li>Change "When using this certificate" to <strong>Always Trust</strong>.</li>
                  </ol>
                </div>
              </div>
            </transition>
          </div>

          <!-- Main Login Action -->
          <div class="flex justify-center w-full">
            <button @click="login"
                    class="w-full sm:w-auto px-12 py-4 bg-secondary hover:bg-secondary-hover text-navbar-text rounded-xl font-bold text-xl transition-all shadow-lg active:scale-95 transform hover:-translate-y-0.5">
              Proceed to Login &rarr;
            </button>
          </div>
        </div>
      </div>

      <div class="absolute top-0 left-0 w-full h-full pointer-events-none">
        <div class="absolute bg-white/10 w-72 h-72 rounded-full -top-16 -left-16 animate-blob"></div>
        <div class="absolute bg-white/20 w-96 h-96 rounded-full -bottom-32 -right-32 animate-blob animation-delay-2000"></div>
        <div class="absolute bg-white/10 w-48 h-48 rounded-full -bottom-16 -left-20 animate-blob animation-delay-4000"></div>
      </div>
    </header>

    <section v-if="!authStore.isAuthenticated" class="py-24 bg-slate-50 dark:bg-slate-900/20">
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

    <footer class="bg-navbar-bg text-navbar-text py-8 text-center mt-auto border-t border-current/5">
      <p>© 2026 LANForge. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "../stores/auth";
import { ref, watchEffect } from "vue";
import { login as loginOidc } from "../auth/oidc";
import { useRouter } from "vue-router";

const authStore = useAuthStore();
const router = useRouter();

// Toggle states for instructions drawer
const showInstructions = ref(false);

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

/* Transition Animations for Drawer */
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}
.slide-fade-leave-active {
  transition: all 0.2s ease-in;
}
.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translateY(-10px);
  opacity: 0;
}
</style>