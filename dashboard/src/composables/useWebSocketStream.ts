import { ref, onBeforeUnmount } from "vue";
import { buildWsUrl } from "../utils/websocket";

interface WebSocketOptions {
  onMessage?: (event: MessageEvent) => void;
  onError?: (event: Event) => void;
  onOpen?: (event: Event) => void;
}

export function useWebSocketStream() {
  const ws = ref<WebSocket | null>(null);

  const connect = (path: string, options: WebSocketOptions = {}) => {
    // Prevent duplicate connections if recalled
    if (ws.value) ws.value.close();

    const url = buildWsUrl(path);
    if (!url) return null;

    const socket = new WebSocket(url);

    socket.onopen = (e) => options.onOpen?.(e);
    socket.onerror = (e) => options.onError?.(e);
    socket.onmessage = (e) => options.onMessage?.(e);
    
    socket.onclose = () => {
      ws.value = null;
    };

    ws.value = socket;
    return socket;
  };

  const close = () => {
    if (ws.value) {
      ws.value.close();
      ws.value = null;
    }
  };

  // Safe structural auto-cleanup when component unmounts
  onBeforeUnmount(() => {
    close();
  });

  return {
    ws,
    connect,
    close,
  };
}