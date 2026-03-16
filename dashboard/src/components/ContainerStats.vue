<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue"
import { useAuthStore } from "../stores/auth";

const props = defineProps({
  containerID: String
})
const authStore = useAuthStore();

const cpu = ref(0)
const memory = ref(0)
const memMax = ref(0)

let ws

onMounted(() => {
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${protocol}://localhost:8080/containers/${props.containerID}/stats?token=${authStore.token}`;
  ws = new WebSocket(wsUrl);

  ws.onmessage = (e) => {
    const data = JSON.parse(e.data)
    cpu.value = data.cpu
    memory.value = data.memory
    memMax.value = data.memMax
  }
})

onBeforeUnmount(() => {
  if (ws) ws.close()
})
</script>

<template>
<div class="stats-bar">
  <div class="stat-item">
    <span class="stat-name">CPU:</span>
    <span class="stat-value text-green-400">{{ cpu.toFixed(2) }}%</span>
  </div>

  <div class="stat-item">
    <span class="stat-name">Memory:</span>
    <span class="stat-value text-blue-400">
      {{ (memory / 1024 / 1024).toFixed(0) }} MB
    </span>
  </div>

  <div class="stat-item">
    <span class="stat-name">Limit:</span>
    <span class="stat-value text-purple-400">
      {{ (memMax / 1024 / 1024).toFixed(0) }} MB
    </span>
  </div>
</div>
</template>

<style scoped>
.stats-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  padding: 0.75rem 1rem;
  background-color: #1e1e1e;
  border-radius: 8px;
  border: 1px solid #444;
  font-family: monospace;
  font-size: 0.95rem;
  justify-content: flex-start;
  align-items: center;
}

.stat-item {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.stat-name {
  color: #888;
  font-weight: 500;
}

.stat-value {
  font-weight: 600;
}
</style>