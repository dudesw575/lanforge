<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";

const props = defineProps<{ isOpen: boolean }>();
const emit = defineEmits(["close"]);

const isCollapsed = ref(false);
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-50 bg-navbar-bg text-navbar-text border-r border-current/10 transition-all duration-300 lg:relative lg:translate-x-0"
    :class="[
      isOpen ? 'translate-x-0' : '-translate-x-full',
      isCollapsed ? 'w-20' : 'w-64',
    ]"
  >
    <div class="h-full flex flex-col">
      <div
        class="h-16 flex items-center px-4 border-b border-current/10"
        :class="isCollapsed ? 'justify-center' : 'justify-between px-6'"
      >
        <button
          @click="emit('close')"
          class="lg:hidden p-2 hover:bg-current/10 rounded-lg transition-colors"
        >
          ✕
        </button>
        <span
          v-if="!isCollapsed"
          class="text-xl font-black tracking-tighter uppercase truncate"
        >
          <router-link to="/" class="flex items-center gap-1">
            LAN<span class="text-primary">Forge</span>
          </router-link>
        </span>

        <button
          @click="isCollapsed = !isCollapsed"
          class="hidden lg:flex p-2 hover:bg-current/10 rounded-lg transition-colors"
        >
          <span class="text-xs">{{ isCollapsed ? "▶" : "◀" }}</span>
        </button>
      </div>

      <nav class="flex-1 p-3 space-y-1">
        <router-link
          to="/dashboard"
          class="nav-item"
          :title="isCollapsed ? 'Dashboard' : ''"
        >
          <span class="nav-icon">📦</span>
          <span v-if="!isCollapsed" class="font-semibold whitespace-nowrap"
            >Dashboard</span
          >
        </router-link>

        <router-link
          to="/templates"
          class="nav-item"
          :title="isCollapsed ? 'Templates' : ''"
        >
          <span class="nav-icon">📜</span>
          <span v-if="!isCollapsed" class="font-semibold whitespace-nowrap"
            >Templates</span
          >
        </router-link>

        <router-link
          to="/settings"
          class="nav-item"
          :title="isCollapsed ? 'Settings' : ''"
        >
          <span class="nav-icon">⚙️</span>
          <span v-if="!isCollapsed" class="font-semibold whitespace-nowrap"
            >Settings</span
          >
        </router-link>
      </nav>
    </div>
  </aside>
</template>
