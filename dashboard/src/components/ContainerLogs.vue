<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { useAuthStore } from "../stores/auth";

const props = defineProps({ containerID: String });
const terminalEl = ref<HTMLDivElement | null>(null);
let term: Terminal, fitAddon: FitAddon, ws: WebSocket, resizeObserver: ResizeObserver;

const authStore = useAuthStore();

const fitTerminal = () => {
  if (fitAddon && terminalEl.value && terminalEl.value.offsetHeight > 0) {
    fitAddon.fit();
  }
};

const downloadLogs = () => {
  if (!term) return;
  const buffer = term.buffer.active;
  let logText = "";
  for (let i = 0; i < buffer.length; i++) {
    const line = buffer.getLine(i);
    if (line) logText += line.translateToString() + "\n";
  }
  const blob = new Blob([logText], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `container-${props.containerID}-logs.txt`;
  link.click();
  URL.revokeObjectURL(url);
};

defineExpose({ fitTerminal, downloadLogs });

onMounted(async () => {
  term = new Terminal({
    convertEol: true,
    cursorBlink: true,
    fontSize: 13,
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
    scrollback: 10000,
    theme: {
      background: "#0f172a", 
      foreground: "#f8fafc", 
      cursor: "#3b82f6",    
      selectionBackground: "rgba(59, 130, 246, 0.3)",
    },
    allowTransparency: true,
  });

  fitAddon = new FitAddon();
  term.loadAddon(fitAddon);
  term.open(terminalEl.value!);

  await nextTick();
  setTimeout(fitTerminal, 100);

  resizeObserver = new ResizeObserver(() => fitTerminal());
  resizeObserver.observe(terminalEl.value!);
  
  window.addEventListener('resize', fitTerminal);
  connect();
});

function connect() {
  if (!authStore.token) return;
  
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${protocol}://localhost:8080/containers/${props.containerID}/logs?token=${authStore.token}`;
  
  console.log("Connecting to logs:", wsUrl);
  
  ws = new WebSocket(wsUrl);
  
  ws.onmessage = (e) => {
    term.write(e.data);
  };

  ws.onerror = (err) => {
    term.write("\r\n\x1b[31m[Error] Failed to connect to log stream.\x1b[0m\r\n");
    console.error("WebSocket Error:", err);
  };
}

onBeforeUnmount(() => {
  ws?.close();
  resizeObserver?.disconnect();
  window.removeEventListener('resize', fitTerminal);
  term?.dispose();
});
</script>

<template>
  <div class="w-full h-full min-h-[200px] bg-[#0f172a] rounded-lg overflow-hidden">
    <div ref="terminalEl" class="w-full h-full"></div>
  </div>
</template>

<style scoped>
:deep(.xterm) {
  padding: 12px;
  height: 100%;
}
:deep(.xterm-viewport) {
  background-color: transparent !important;
}
</style>