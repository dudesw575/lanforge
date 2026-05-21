<script setup lang="ts">
import { ref, computed, onMounted} from "vue"
// import { useAuthStore } from "../stores/auth";
import { useWebSocketStream } from "../composables/useWebSocketStream";

const props = defineProps({
  containerID: String
})
// const authStore = useAuthStore();

const cpu = ref(0)
const memory = ref(0)
const memMax = ref(0)

const { connect } = useWebSocketStream();
// let ws : WebSocket | null = null;

const memPercent = computed(() => {
  if (!memMax.value || memMax.value === 0) return 0;
  return (memory.value / memMax.value) * 100;
});

// onMounted(() => {
//   const protocol = window.location.protocol === "https:" ? "wss" : "ws";
//   const wsUrl = `${protocol}://localhost:8080/containers/${props.containerID}/stats?token=${authStore.token}`;
//   ws = new WebSocket(wsUrl);

//   ws.onmessage = (e) => {
//     try {
//       const data = JSON.parse(e.data)
//       cpu.value = data.cpu
//       memory.value = data.memory
//       memMax.value = data.memMax
//     } catch (err) {
//       console.error("Failed to parse stats:", err)
//     }
//   }
// })

// onBeforeUnmount(() => {
//   if (ws) ws.close()
// })

onMounted(() => {
  connect(`/containers/${props.containerID}/stats`, {
    onMessage: (e) => {
      try {
        const data = JSON.parse(e.data);
        cpu.value = data.cpu;
        memory.value = data.memory;
        memMax.value = data.memMax;
      } catch (err) {
        console.error("Failed to parse stats:", err);
      }
    }
  });
});
</script>

<template>
  <div class="flex flex-wrap gap-8 p-4 bg-current/5 border border-current/10 rounded-xl font-mono text-sm items-center shadow-inner">
    
    <div class="flex items-center gap-3">
      <div class="flex flex-col">
        <span class="opacity-50 font-medium uppercase text-[10px] tracking-wider leading-none mb-1">CPU Load</span>
        <span class="text-secondary font-bold tabular-nums text-base leading-none">
          {{ cpu.toFixed(1) }}%
        </span>
      </div>
      <div class="w-20 h-2 bg-current/10 rounded-full overflow-hidden hidden sm:block">
        <div 
          class="h-full bg-secondary transition-all duration-700 ease-out" 
          :style="{ width: `${Math.min(cpu, 100)}%` }"
        ></div>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <div class="flex flex-col">
        <span class="opacity-50 font-medium uppercase text-[10px] tracking-wider leading-none mb-1">Memory Usage</span>
        <div class="flex items-baseline gap-1">
          <span class="text-primary font-bold tabular-nums text-base leading-none">
            {{ (memory / 1024 / 1024).toFixed(0) }}
          </span>
          <span class="text-[10px] opacity-40">/ {{ (memMax / 1024 / 1024).toFixed(0) }} MB</span>
        </div>
      </div>
      
      <div class="w-24 h-2 bg-current/10 rounded-full overflow-hidden hidden md:block">
        <div 
          class="h-full bg-primary transition-all duration-700 ease-out" 
          :style="{ width: `${Math.min(memPercent, 100)}%` }"
        ></div>
      </div>
    </div>

    <div class="ml-auto hidden lg:flex items-center gap-2 px-3 py-1 rounded-full bg-current/5 border border-current/5">
      <div :class="cpu > 80 || memPercent > 90 ? 'bg-red-500 animate-pulse' : 'bg-secondary'" class="w-2 h-2 rounded-full"></div>
      <span class="text-[10px] font-bold uppercase tracking-tighter opacity-70">
        {{ cpu > 80 || memPercent > 90 ? 'High Load' : 'Healthy' }}
      </span>
    </div>
  </div>
</template>