<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from "vue"; // ⭐ added onUnmounted
import { useAuthStore } from "../stores/auth";
import { login as oidcLogin} from "../auth/oidc";
import api from "../api/client";
import ContainerLogs from "../components/ContainerLogs.vue";
import ContainerStats from "../components/ContainerStats.vue";
import Modal from "../components/Modal.vue";

interface Container {
  Id: string;
  Names: string[];
  State: string;
}

const authStore = useAuthStore();
const containers = ref<Container[]>([]);

const selectedContainer = ref<Container | null>(null);
const logsVisible = ref(false);
let logsComp = ref(null);

const confirmVisible = ref(false);
const containerToRemove = ref<Container | null>(null);

// ⭐ polling interval reference
let pollTimer: ReturnType<typeof setInterval> | null = null;

// Roles
const userRoles = computed(() => authStore.user?.resource_access?.["lan-control-plane"]?.roles ?? []);
const canStartStop = computed(() => userRoles.value.includes("operator") || userRoles.value.includes("admin"));
const isAdmin = computed(() => userRoles.value.includes("admin"));

// ⭐ Token countdown debug
const tokenExpiresIn = ref(0)

setInterval(() => {
  if (!authStore.expiresAt) return
  tokenExpiresIn.value = authStore.expiresAt - Math.floor(Date.now() / 1000)
}, 1000)


// Axios interceptor for auth token
api.interceptors.request.use(config => {
  if (authStore.token) config.headers.Authorization = `Bearer ${authStore.token}`;
  return config;
});

// Fetch containers
const fetchContainers = async () => {
  try {
    const res = await api.get<Container[]>("/containers");
    containers.value = res.data;
  } catch (err) {
    console.error("Failed to fetch containers:", err);
  }
};

// Container actions
const start = async (id: string) => { await api.post(`/containers/${id}/start`); await fetchContainers(); };
const stop = async (id: string) => { await api.post(`/containers/${id}/stop`); await fetchContainers(); };
const remove = async (id: string) => { await api.delete(`/containers/${id}`); await fetchContainers(); };

const confirmRemove = (container: Container) => { containerToRemove.value = container; confirmVisible.value = true; };
const cancelRemove = () => { containerToRemove.value = null; confirmVisible.value = false; };

const removeConfirmed = async () => { 
  if (!containerToRemove.value) return;
  await remove(containerToRemove.value.Id); 
  cancelRemove();
};

const showLogs = (container: Container) => { selectedContainer.value = container; logsVisible.value = true; };
const closeLogs = () => { logsVisible.value = false; selectedContainer.value = null; };

// Login/Logout
// const login = async () => { await oidcLogin(); };

// On mount
onMounted(async () => {
  if (authStore.authenticated) {
    await fetchContainers();
    pollTimer = setInterval(fetchContainers, 5000); // ⭐ saved reference
  }
});

// ⭐ cleanup polling
onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
});
</script>


<template>
  <div class="container mx-auto flex justify-center py-6">

    <div v-if="!authStore.authenticated">
      <p>You are not logged in.</p>
      <!-- redirect to index.vue -->
      <!-- <button @click="login">Login</button> -->
    </div>

    <div v-else>
      <h1 class="text-xl mb-4">Containers</h1>

      <div v-for="c in containers" :key="c.Id" class="border p-2 mb-2 flex justify-between items-center">
        <div>{{ c.Names[0].replace("/", "") }} - {{ c.State }}</div>
        <div class="flex gap-2">
          <button v-if="canStartStop" @click="start(c.Id)" class="px-2 py-1 bg-green-500 text-white rounded">Start</button>
          <button v-if="canStartStop" @click="stop(c.Id)" class="px-2 py-1 bg-red-500 text-white rounded">Stop</button>
          <button @click="showLogs(c)" class="px-2 py-1 bg-blue-500 text-white rounded">Logs</button>
          <button v-if="isAdmin" @click="confirmRemove(c)" class="px-2 py-1 bg-red-700 text-white rounded">Remove</button>
        </div>
      </div>

      <Teleport to="body">
        <Modal :show="logsVisible" @close="closeLogs" @opened="logsComp?.fitTerminal()">
          <template #header>
            <span class="truncate text-white">Container: {{ selectedContainer?.Names[0].replace('/', '') }}</span>
          </template>
          <template #body>
            <div class="flex flex-col h-full overflow-hidden">
              <ContainerStats :containerID="selectedContainer?.Id" />
              <div class="flex-1 min-h-0 mt-4 bg-[#0f172a] rounded-sm">
                <ContainerLogs ref="logsComp" v-if="logsVisible" :containerID="selectedContainer?.Id" />
              </div>
            </div>
          </template>
        </Modal>
      </Teleport>

      <Teleport to="body">
        <Modal :show="confirmVisible" @close="cancelRemove">
          <template #header>
            <span class="text-white text-center">Confirm Remove</span>
          </template>
          <template #body>
            <h3 class="text-center text-lg">
              Are you sure you want to remove container:
              <strong>{{ containerToRemove?.Names[0].replace("/", "") }}</strong>?
            </h3>
            <div class="mt-4 flex justify-end gap-2">
              <button @click="cancelRemove" class="px-3 py-1 bg-gray-500 text-white rounded hover:scale-105 transition-transform">No</button>
              <button @click="removeConfirmed" class="px-3 py-1 bg-red-700 text-white rounded hover:scale-105 transition-transform">Yes</button>
            </div>
          </template>
        </Modal>
      </Teleport>
    </div>

  </div>
</template>

<style scoped>
h1 { margin-bottom: 1rem; }
button { margin-left: 0.5rem; }
.bg-red-700 { background-color: #b91c1c !important; }
</style>