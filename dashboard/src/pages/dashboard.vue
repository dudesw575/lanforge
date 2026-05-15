<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
import api from "../api/client";
import Modal from "../components/Modal.vue";
import ContainerLogs from "../components/ContainerLogs.vue";
import ContainerStats from "../components/ContainerStats.vue";
import GameLauncher from "../components/GameLauncher.vue";
import { listTemplates, type Template } from "../api/templates";

interface Container {
  Id: string;
  Names: string[];
  State: string;
}

const router = useRouter();
const logViewer = ref<InstanceType<typeof ContainerLogs> | null>(null);
const authStore = useAuthStore();
const containers = ref<Container[]>([]);
const templates = ref<Template[]>([]);
const selectedContainer = ref<Container | null>(null);
const logsVisible = ref(false);
let pollTimer: any = null;

const confirmVisible = ref(false);
const containerToRemove = ref<Container | null>(null);
const removeVolumes = ref(false);

const userRoles = computed(
  () => authStore.user?.resource_access?.["lan-control-plane"]?.roles ?? [],
);

const canStartStop = computed(
  () =>
    userRoles.value.includes("operator") || userRoles.value.includes("admin"),
);

const isAdmin = computed(() => userRoles.value.includes("admin"));


onMounted(async () => {
  if (!authStore.isAuthenticated) {
    return router.push("/");
  }

  try {
    await fetchContainers();
    await fetchTemplates();
    pollTimer = setInterval(fetchContainers, 5000);
  } catch (err) {
    console.error("Failed to initialize dashboard data:", err);
  }
});

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});

const fetchContainers = async () => {
  try {
    const res = await api.get<Container[]>("/containers");
    containers.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

const fetchTemplates = async () => {
  try {
    templates.value = await listTemplates();
  } catch (err) {
    console.error(err);
  }
};

const start = async (id: string) => {
  await api.post(`/containers/${id}/start`);
  fetchContainers();
};

const stop = async (id: string) => {
  await api.post(`/containers/${id}/stop`);
  fetchContainers();
};

const showLogs = (c: Container) => {
  selectedContainer.value = c;
  logsVisible.value = true;
};

const closeLogs = () => {
  logsVisible.value = false;
  selectedContainer.value = null;
};

const confirmRemove = (container: Container) => {
  containerToRemove.value = container;
  removeVolumes.value = false;
  confirmVisible.value = true;
};

const removeConfirmed = async () => {
  if (containerToRemove.value) {
    await api.delete(`/containers/${containerToRemove.value.Id}?volumes=${removeVolumes.value}`);
    confirmVisible.value = false;
    containerToRemove.value = null;
    fetchContainers();
  }
};
</script>

<template>
  <div class="bg-bg text-text transition-colors duration-300">
    <div class="container mx-auto py-8 px-4">
    <GameLauncher class="mb-12" />
      
      <!-- Templates Summary
      <section class="mb-10">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold flex items-center gap-2">
            <span class="text-primary">📜</span> Templates
          </h2>
          <router-link
            to="/templates"
            class="text-primary hover:underline font-bold text-sm"
          >
            View All ({{ templates.length }}) →
          </router-link>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="t in templates.slice(0, 6)"
            :key="t.id"
            class="border border-current/10 p-4 rounded-xl bg-white/5 flex items-center justify-between hover:border-primary/30 transition-all"
          >
            <div>
              <div class="font-bold text-primary">{{ t.name }}</div>
              <div class="text-xs opacity-50 mt-0.5">{{ t.image }}</div>
            </div>
            <span class="text-[10px] font-mono opacity-30 bg-white/5 px-2 py-0.5 rounded">{{ t.id }}</span>
          </div>
          <div
            v-if="templates.length === 0"
            class="col-span-full text-center opacity-40 py-8 border border-dashed border-current/20 rounded-xl"
          >
            No templates yet.
          </div>
        </div>
      </section> -->

      <!-- Containers -->
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold flex items-center gap-2">
            <span class="text-secondary">⦿</span> Containers
          </h2>
          <span class="text-sm opacity-50">Polling every 5s</span>
        </div>

        <div
          v-for="c in containers"
          :key="c.Id"
          class="border border-current/10 p-4 mb-4 rounded-xl flex flex-wrap justify-between items-center bg-white/5 hover:bg-white/10 transition-all shadow-sm"
        >
          <div class="flex items-center gap-4">
            <div class="relative flex h-3 w-3">
              <span v-if="c.State === 'running'" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-secondary opacity-75"></span>
              <span 
                :class="c.State === 'running' ? 'bg-secondary' : 'bg-red-500'"
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
          </div>

          <div class="flex gap-2 mt-4 sm:mt-0">
            <button
              v-if="canStartStop"
              @click="start(c.Id)"
              :disabled="c.State === 'running'"
              class="px-4 py-1.5 bg-secondary hover:bg-secondary-hover text-navbar-text rounded-md disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Start
            </button>

            <button
              v-if="canStartStop"
              @click="stop(c.Id)"
              :disabled="c.State !== 'running'"
              class="px-4 py-1.5 bg-orange-500 hover:bg-orange-600 text-white rounded-md disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Stop
            </button>

            <button
              @click="showLogs(c)"
              class="px-4 py-1.5 bg-primary hover:bg-primary-hover text-navbar-text rounded-md"
            >
              Logs & Stats
            </button>

            <button
              v-if="isAdmin"
              @click="confirmRemove(c)"
              class="px-4 py-1.5 bg-red-600 hover:bg-red-700 text-white rounded-md"
            >
              Remove
            </button>
          </div>
        </div>

        <div v-if="containers.length === 0" class="text-center opacity-40 py-16 border border-dashed border-current/20 rounded-xl">
          <p class="text-lg">No containers running.</p>
          <p class="text-sm mt-1">Deploy a template from the <router-link to="/templates" class="text-primary hover:underline">Templates</router-link> page to get started.</p>
        </div>
      </section>
    </div>

    <Teleport to="body">
      <Modal :show="logsVisible" @close="closeLogs" @opened="logViewer?.fitTerminal()">
        <template #header>
          <span class="text-navbar-text font-bold text-lg">
            Inspecting: {{ selectedContainer?.Names?.[0]?.replace("/", "") }}
          </span>
        </template>

        <template #body>
          <div class="flex flex-col h-full w-full overflow-hidden min-h-0">
            <div class="flex-shrink-0 mb-6">
              <ContainerStats :containerID="selectedContainer?.Id" />
            </div>

            <h3 class="text-xs font-black opacity-40 uppercase mb-3 tracking-widest">
              Live Logs
            </h3>

            <div class="flex-1 bg-black/40 rounded-lg border border-current/10 overflow-hidden min-h-0">
              <ContainerLogs
                ref="logViewer"
                v-if="logsVisible"
                :containerID="selectedContainer?.Id"
              />
            </div>
          </div>
        </template>

        <template #extra-actions>
          <div class="flex justify-between w-full items-center">
            <button
              @click="logViewer?.downloadLogs()"
              class="bg-primary/10 hover:bg-primary/20 text-current px-4 py-2 rounded-lg font-bold transition flex items-center gap-2 border border-current/10"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a2 2 0 002 2h12a2 2 0 002-2v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download
            </button>
          </div>
        </template>
      </Modal>
    </Teleport>

    <Teleport to="body">
      <Modal :show="confirmVisible" @close="confirmVisible = false">
        <template #header><span class="text-white font-bold">Dangerous Action</span></template>
        <template #body>
          <div class="p-2 text-center">
            <p class="mb-4 text-lg">Destroy container <strong>{{ containerToRemove?.Names?.[0]?.replace("/", "") }}</strong>?</p>
            <div class="bg-red-900/20 p-4 rounded-lg border border-red-700/50 mb-4">
              <label class="flex items-center justify-center gap-3 cursor-pointer">
                <input type="checkbox" v-model="removeVolumes" class="w-5 h-5 accent-red-600" />
                <span class="text-red-200">Wipe all persistent data volumes?</span>
              </label>
            </div>
          </div>
        </template>
        <template #extra-actions>
          <div class="flex justify-end gap-3 w-full">
            <button @click="removeConfirmed" class="px-6 py-2 bg-red-600 hover:bg-red-500 rounded font-bold">Confirm Delete</button>
          </div>
        </template>
      </Modal>
    </Teleport>
  </div>
</template>