
<!-- <script setup lang="ts">
import { ref } from 'vue'

const roomName = ref<string>('general')
const elementUrl = 'https://chat.home.arpa/'
</script>

<template>
  <div class="chat-container h-screen flex flex-col">
    <main class="flex-1 bg-gray-900 overflow-hidden">
      <iframe
  :src="`${elementUrl}/#/room/#${roomName}?auto_login=true`"
  allow="microphone; camera; autoplay; display-capture; encrypted-media"
  class="w-full h-full border-none"
></iframe>
    </main>
  </div>
</template> -->

<script setup lang="ts">
import { ref, computed} from 'vue'

interface Props {
  // Pass the target room name or room ID from a parent route/state
  roomAlias?: string
}

const props = withDefaults(defineProps<Props>(), {
  roomAlias: 'general' // Default room fallback
})

const isFrameLoading = ref<boolean>(true)
const elementHost = 'https://chat.home.arpa'
const homeserverDomain = 'home.arpa' // Used to construct internal room targets

/**
 * Computes the optimized embedding URL.
 * - embed=true: Strips sidebars, settings buttons, and header bars.
 * - auto_login=true: Tells Element to immediately use the OIDC flow without prompting.
 */
const iframeUrl = computed(() => {
  // Standard Matrix room syntax is #room_name:domain.com
  const cleanRoomTarget = `#${props.roomAlias}:${homeserverDomain}`
  
  const baseUrl = `${elementHost}/#/room/${encodeURIComponent(cleanRoomTarget)}`
  const queryParams = new URLSearchParams({
    embed: 'true',
    auto_login: 'true'
  })

  return `${baseUrl}?${queryParams.toString()}`
})

const handleFrameLoad = () => {
  isFrameLoading.value = false
}
</script>

<template>
  <div class="chat-wrapper w-full h-full flex flex-col bg-gray-950 relative">
    <!-- Optional Loading State Spinner Overlay -->
    <div 
      v-if="isFrameLoading" 
      class="absolute inset-0 flex flex-col items-center justify-center bg-gray-900 text-gray-400 z-10"
    >
      <svg class="animate-spin h-10 w-10 text-emerald-500 mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p class="text-sm tracking-wide font-medium">Connecting to secure chat network...</p>
    </div>

    <!-- The Dedicated Secure App Frame -->
    <main class="flex-1 w-full h-full overflow-hidden relative">
      <iframe
        ref="chatFrame"
        :src="iframeUrl"
        @load="handleFrameLoad"
        class="w-full h-full border-0 rounded-none shadow-inner"
        scrolling="no"
        sandbox="allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-same-origin allow-scripts allow-downloads"
        allow="microphone; camera; autoplay; display-capture; encrypted-media; clipboard-write;"
        title="Embedded Communication Panel"
      ></iframe>
    </main>
  </div>
</template>

<style scoped>
.chat-wrapper {
  /* Ensures component occupies complete parent context sizing */
  height: 100%;
  width: 100%;
}
</style>