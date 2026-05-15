<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import api from "../api/client";
import { listTemplates } from "../api/templates";
import type { Template } from "../api/templates";

interface Props {
  onDeploy?: (template: Template) => void;
}

const props = defineProps<Props>();
const authStore = useAuthStore();

const templates = ref<Template[]>([]);
const loading = ref(true);
const selectedCategory = ref<string>("all");

const categories = [
  { id: "all", label: "All Games", icon: "🎮" },
  { id: "survival", label: "Survival", icon: "🏕️" },
  { id: "fps", label: "FPS", icon: "🔫" },
  { id: "sandbox", label: "Sandbox", icon: "🏗️" },
  { id: "strategy", label: "Strategy", icon: "♟️" },
];

const gameIcons: Record<string, string> = {
  minecraft: "⛏️",
  "minecraft-paper": "⛏️",
  cs2: "🔫",
  factorio: "🏗️",
  valheim: "🪓",
  ark: "🦖",
  garrysmod: "🎭",
};

const gameCategories: Record<string, string[]> = {
  survival: ["minecraft", "minecraft-paper", "ark", "valheim"],
  fps: ["cs2", "garrysmod"],
  sandbox: ["factorio", "garrysmod"],
  strategy: [],
};

const getIcon = (id: string) => gameIcons[id] || "🎮";

const isFavorite = (id: string) => {
  const favs = JSON.parse(localStorage.getItem("favoriteTemplates") || "[]");
  return favs.includes(id);
};

const toggleFavorite = (id: string) => {
  const favs = JSON.parse(localStorage.getItem("favoriteTemplates") || "[]");
  const idx = favs.indexOf(id);
  if (idx > -1) {
    favs.splice(idx, 1);
  } else {
    favs.push(id);
  }
  localStorage.setItem("favoriteTemplates", JSON.stringify(favs));
};

const filteredTemplates = computed(() => {
  if (selectedCategory.value === "all") return templates.value;
  const cat = gameCategories[selectedCategory.value] || [];
  return templates.value.filter((t) => cat.includes(t.id));
});

const deploy = async (template: Template) => {
  if (props.onDeploy) {
    props.onDeploy(template);
    return;
  }

  const name = `${template.id}-${Math.floor(Math.random() * 1000)}`;

  // Open WebSocket for logging (fire-and-forget, not needed for deploy to succeed)
  const ws = new WebSocket(
    `ws://localhost:8080/deploy/stream?id=${name}&token=${authStore.token}`
  );

  ws.onmessage = (e) => {
    console.log("Deploy:", e.data);
  };

  try {
    await api.post(`/templates/${template.id}/deploy?name=${encodeURIComponent(name)}`);
    console.log("✅ Deployed:", name);
  } catch (err) {
    console.error("Deploy failed:", err);
  }
};

onMounted(async () => {
  try {
    templates.value = await listTemplates();
  } catch (err) {
    console.error("Failed to load templates:", err);
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="bg-bg text-text transition-colors duration-300">
    <div class="container mx-auto py-8 px-4">
      <div class="mb-8">
        <h1 class="text-3xl font-bold flex items-center gap-3">
          <span class="text-primary">🚀</span> Quick Launch
        </h1>
        <p class="text-sm opacity-50 mt-1">One-click deployment for popular game servers</p>
      </div>

      <!-- Category Filter -->
      <div class="flex gap-2 mb-6 flex-wrap">
        <button
          v-for="cat in categories"
          :key="cat.id"
          @click="selectedCategory = cat.id"
          :class="[
            'px-4 py-2 rounded-lg font-bold text-sm transition-all active:scale-95',
            selectedCategory === cat.id
              ? 'bg-primary text-navbar-text shadow-md'
              : 'bg-white/5 hover:bg-white/10 border border-current/10'
          ]"
        >
          {{ cat.icon }} {{ cat.label }}
        </button>
      </div>

      <!-- Game Grid -->
      <div v-if="loading" class="text-center opacity-50 py-16">Loading games...</div>
      <div v-else-if="filteredTemplates.length === 0" class="text-center opacity-40 py-16 border border-dashed border-current/20 rounded-xl">
        <p class="text-lg">No games in this category yet.</p>
        <p class="text-sm mt-1">Ask an admin to add templates.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="t in filteredTemplates"
          :key="t.id"
          class="border border-current/10 p-5 rounded-xl bg-white/5 hover:bg-white/10 transition-all"
        >
          <div class="flex items-start justify-between mb-3">
            <div class="flex items-center gap-3">
              <span class="text-3xl">{{ getIcon(t.id) }}</span>
              <div>
                <div class="font-bold text-lg text-primary">{{ t.name }}</div>
                <div class="text-xs opacity-50">{{ t.description }}</div>
              </div>
            </div>
            <button
              @click="toggleFavorite(t.id)"
              :class="isFavorite(t.id) ? 'text-yellow-500' : 'opacity-30 hover:opacity-100 text-white'"
              class="text-lg transition"
              title="Toggle favorite"
            >
              {{ isFavorite(t.id) ? '⭐' : '☆' }}
            </button>
          </div>

          <div class="flex items-center gap-2 text-xs opacity-50 font-mono mb-3">
            <span class="bg-white/5 px-2 py-0.5 rounded">{{ t.image }}</span>
            <span v-if="t.ports?.length">{{ t.ports.length }} port(s)</span>
            <span v-if="t.volumes?.length">{{ t.volumes.length }} volume(s)</span>
          </div>

          <div class="flex gap-2">
            <button
              @click="deploy(t)"
              class="flex-1 bg-secondary hover:bg-secondary-hover text-white px-4 py-2 rounded-lg font-bold text-sm transition-all active:scale-95 shadow-md"
            >
              🚀 Quick Deploy
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>