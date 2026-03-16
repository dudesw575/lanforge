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

// Expose a method to parent to refit terminal
const fitTerminal = () => {
  if (fitAddon && terminalEl.value) fitAddon.fit();
};
defineExpose({ fitTerminal });

onMounted(async () => {
  term = new Terminal({
    convertEol: true,
    cursorBlink: true,
    fontSize: 13,
    fontFamily: "Menlo, Monaco, 'Courier New', monospace",
    scrollback: 5000,
    theme: { background: "#0f172a", foreground: "#e5e7eb", cursor: "#22c55e" },
  });

  fitAddon = new FitAddon();
  term.loadAddon(fitAddon);
  term.open(terminalEl.value!);

  await nextTick();
  fitTerminal();

  resizeObserver = new ResizeObserver(() => fitTerminal());
  resizeObserver.observe(terminalEl.value!);

  connect();
});

function connect() {
  if (!authStore.token) {
    term.writeln("\x1b[31m[Cannot connect: no auth token]\x1b[0m");
    return;
  }

  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = `${protocol}://localhost:8080/containers/${props.containerID}/logs?token=${authStore.token}`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => term.writeln("\x1b[32m[Connected to container logs]\x1b[0m");
  ws.onmessage = (e) => {
    term.write(e.data);
    term.scrollToBottom();
  };
  ws.onclose = () => term.writeln("\x1b[33m[Log stream closed]\x1b[0m");
  ws.onerror = () => term.writeln("\x1b[31m[WebSocket error]\x1b[0m");
}

onBeforeUnmount(() => {
  ws?.close();
  resizeObserver?.disconnect();
  term?.dispose();
});
</script>

<template>
  <div ref="terminalEl" class="w-full h-full"></div>
</template>