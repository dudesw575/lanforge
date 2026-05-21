<script setup lang="ts">
import { ref, onMounted, Teleport } from "vue";
import Modal from "../components/Modal.vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
import ContainerStats from "../components/ContainerStats.vue";


interface Container {
  Id: string;
  Names: string[];
  State: string;
}

const authStore = useAuthStore();
const router = useRouter();

const appInfo = ref({ version: "", mode: import.meta.env.MODE });
const deployStatus = ref({ status: "", lastUpdated: "" });
const containers = ref<Container[]>([]);
const loading = ref(false);
const error = ref("");
const showDeleteModal = ref(false);
const containerToDelete = ref<string | null>(null);

async function fetchData() {
  error.value = "";
  loading.value = true;
  const headers: HeadersInit = {};
  if (authStore.token) {
    headers["Authorization"] = `Bearer ${authStore.token}`;
  }
  try {
    const [verRes, depRes, contRes] = await Promise.all([
      fetch("/api/version", { headers }),
      fetch("/api/deploy/status", { headers }),
      fetch("/api/containers", { headers }),
    ]);
    if (verRes.ok) {
      const v = await verRes.json();
      appInfo.value.version = v.version;
    }
    if (depRes.ok) {
      const d = await depRes.json();
      deployStatus.value = d;
    }
    if (contRes.ok) {
      const c = await contRes.json();
      containers.value = c;
    }
  } catch (e) {
    console.error("Failed to fetch settings data", e);
    error.value = "Failed to load data. Please try again.";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchData();
});
function logout() {
  authStore.token = "";
  router.push({ name: "login" });
}

// function openDeleteModal(id: string) {
//   containerToDelete.value = id;
//   showDeleteModal.value = true;
// }

async function confirmDelete() {
  if (!containerToDelete.value) return;
  const headers: HeadersInit = {};
  if (authStore.token) {
    headers["Authorization"] = `Bearer ${authStore.token}`;
  }
  try {
    const res = await fetch(`/api/containers/${containerToDelete.value}`, {
      method: "DELETE",
      headers,
    });
    if (!res.ok) throw new Error("Delete failed");
    // Refresh list
    await fetchData();
  } catch (e) {
    console.error("Delete error", e);
    error.value = "Failed to delete container.";
  } finally {
    showDeleteModal.value = false;
    containerToDelete.value = null;
  }
}
</script>

<template>
  <div class="bg-bg text-text transition-colors duration-300">
    <div class="container mx-auto py-8 px-4">

      <!-- Page Header -->
      <div class="flex items-center justify-between mb-8">
        <h1 class="text-4xl font-black tracking-tight flex items-center gap-3">
          <span class="text-primary">⚙️</span>
          Settings
        </h1>

        <div class="text-sm opacity-50">
          System & Deployment Overview
        </div>
      </div>

      <!-- Error -->
      <div
        v-if="error"
        class="mb-6 border border-red-500/30 bg-red-500/10 text-red-300 rounded-xl p-4"
      >
        {{ error }}
      </div>

      <!-- App Info -->
      <section class="mb-10">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold flex items-center gap-2">
            <span class="text-primary">🖥️</span>
            App Info
          </h2>
        </div>

        <div
          class="border border-current/10 rounded-xl bg-white/5 p-6 hover:border-primary/20 transition-all"
        >
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <div class="text-xs uppercase tracking-[0.2em] opacity-40 font-black mb-1">
                Version
              </div>

              <div class="font-mono text-lg font-bold text-primary">
                {{ appInfo.version || "Loading..." }}
              </div>
            </div>

            <div>
              <div class="text-xs uppercase tracking-[0.2em] opacity-40 font-black mb-1">
                Mode
              </div>

              <div class="font-mono text-lg font-bold">
                {{ appInfo.mode }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Deployment -->
      <section class="mb-10">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold flex items-center gap-2">
            <span class="text-secondary">🚀</span>
            Deployment
          </h2>

          <span class="text-sm opacity-50">
            Live deployment state
          </span>
        </div>

        <div
          class="border border-current/10 rounded-xl bg-white/5 p-6 hover:border-primary/20 transition-all shadow-sm"
        >
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div>
              <div class="text-xs uppercase tracking-[0.2em] opacity-40 font-black mb-1">
                Status
              </div>

              <div class="flex items-center gap-3">
                <div class="relative flex h-3 w-3">
                  <span
                    v-if="deployStatus.status === 'running'"
                    class="animate-ping absolute inline-flex h-full w-full rounded-full bg-secondary opacity-75"
                  ></span>

                  <span
                    :class="deployStatus.status === 'running'
                      ? 'bg-secondary'
                      : 'bg-red-500'"
                    class="relative inline-flex rounded-full h-3 w-3"
                  ></span>
                </div>

                <span class="font-bold text-lg">
                  {{ deployStatus.status || "Loading..." }}
                </span>
              </div>
            </div>

            <div v-if="deployStatus.lastUpdated">
              <div class="text-xs uppercase tracking-[0.2em] opacity-40 font-black mb-1">
                Last Updated
              </div>

              <div class="font-mono text-sm opacity-70">
                {{ deployStatus.lastUpdated }}
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <button
              @click="fetchData"
              :disabled="loading"
              class="px-5 py-2 bg-primary hover:bg-primary-hover text-navbar-text rounded-lg font-bold transition disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Refresh
            </button>

            <button
              @click="logout"
              class="px-5 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-bold transition"
            >
              Logout
            </button>
          </div>
        </div>
      </section>

      <!-- Containers -->
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold flex items-center gap-2">
            <span class="text-secondary">⦿</span>
            Containers
          </h2>

          <span class="text-sm opacity-50">
            Live status
          </span>
        </div>

        <!-- Loading -->
        <div
          v-if="loading"
          class="text-center opacity-50 py-10 border border-dashed border-current/20 rounded-xl"
        >
          Loading containers...
        </div>

        <!-- Empty -->
        <div
          v-else-if="containers.length === 0"
          class="text-center opacity-40 py-16 border border-dashed border-current/20 rounded-xl"
        >
          <p class="text-lg">No containers found.</p>
        </div>

        <!-- Container Rows -->
        <div
          v-else
          v-for="c in containers"
          :key="c.Id"
          class="border border-current/10 p-4 mb-4 rounded-xl flex flex-wrap justify-between items-center bg-white/5 hover:bg-white/10 transition-all shadow-sm"
        >
          <div class="flex items-center gap-4">
            <div class="relative flex h-3 w-3">
              <span
                v-if="c.State === 'running'"
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-secondary opacity-75"
              ></span>

              <span
                :class="c.State === 'running'
                  ? 'bg-secondary'
                  : 'bg-red-500'"
                class="relative inline-flex rounded-full h-3 w-3"
              ></span>
            </div>

            <div>
              <div class="font-mono font-bold text-lg tracking-tight">
                {{ c.Names?.[0]?.replace("/", "") }}
              </div>

              <div class="text-[10px] opacity-50 uppercase tracking-widest font-black">
                {{ c.State }}
              </div>
            </div>
            <div class="flex-shrink-0">
              <ContainerStats :containerID="c?.Id" />
            </div>
          </div>

        </div>
      </section>
    </div>

    <!-- Delete Modal -->
    <Teleport to="body">
      <Modal
        :show="showDeleteModal"
        @close="showDeleteModal = false"
      >
        <template #header>
          <span class="text-white font-bold text-lg">
            Dangerous Action
          </span>
        </template>

        <template #body>
          <div class="p-2 text-center">
            <p class="mb-4 text-lg">
              Are you sure you want to delete this container?
            </p>

            <div
              class="bg-red-900/20 p-4 rounded-lg border border-red-700/50"
            >
              <p class="text-red-200 text-sm">
                This action cannot be undone.
              </p>
            </div>
          </div>
        </template>

        <template #extra-actions>
          <div class="flex justify-end gap-3 w-full">
            <button
              @click="showDeleteModal = false"
              class="px-5 py-2 rounded-lg border border-current/10 hover:bg-white/5 transition"
            >
              Cancel
            </button>

            <button
              @click="confirmDelete"
              class="px-5 py-2 bg-red-600 hover:bg-red-500 rounded-lg font-bold text-white transition"
            >
              Confirm Delete
            </button>
          </div>
        </template>
      </Modal>
    </Teleport>
  </div>
</template>