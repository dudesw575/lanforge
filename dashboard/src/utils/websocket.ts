import { useAuthStore } from "../stores/auth";

// Targeted by Docker runtime entrypoint replace
const WS_URL_PLACEHOLDER = "VITE_APP_WS_URL_PLACEHOLDER";

/**
 * Resolves the base WebSocket URL matching your environment config
 */
export function getBaseWsUrl(): string {
  // Use replaced placeholder, fallback to vite env, fallback to localhost
  const resolvedHost = WS_URL_PLACEHOLDER.startsWith("VITE_APP_") 
    ? (import.meta.env.VITE_APP_WS_URL || "localhost:8080") 
    : WS_URL_PLACEHOLDER;

  // Clean protocol matching window context
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  
  // Strip protocols if they were included by accident in the env string
  const cleanHost = resolvedHost.replace(/^wss?:\/\//, '');

  return `${protocol}://${cleanHost}`;
}

/**
 * Creates a complete WebSocket URL with nested paths and auth tokens
 */
export function buildWsUrl(path: string): string | null {
  const authStore = useAuthStore();
  if (!authStore.token) {
    console.warn("⚠️ WebSocket attempt blocked: No authentication token found.");
    return null;
  }

  const base = getBaseWsUrl();
  // Ensure paths marry cleanly with or without leading slashes
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  const separator = path.includes('?') ? '&' : '?';

  return `${base}${cleanPath}${separator}token=${authStore.token}`;
}