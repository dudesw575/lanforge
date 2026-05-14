<script setup lang="ts">
const props = defineProps({
  show: Boolean,
});

defineEmits(['close', 'opened']);
</script>

<template>
  <Transition name="modal" @after-enter="$emit('opened')">
    <div v-if="show" 
         class="fixed inset-0 z-[9999] bg-black/70 backdrop-blur-sm flex justify-center items-center p-4"
         @click.self="$emit('close')">
      
      <div class="modal-container flex flex-col h-5/6 w-11/12 max-w-6xl bg-bg border border-current/10 rounded-xl shadow-2xl overflow-hidden transition-transform duration-300">
        
        <header class="px-6 py-4 border-b border-current/10 bg-navbar-bg text-navbar-text flex justify-between items-center">
          <slot name="header">
            <span class="font-bold text-lg">Default Header</span>
          </slot>
          <button @click="$emit('close')" class="hover:opacity-70 text-2xl leading-none">&times;</button>
        </header>

        <div class="modal-body flex-1 overflow-auto p-6">
          <slot name="body"></slot>
        </div>

        <footer class="px-6 py-4 border-t border-current/10 bg-current/5 flex justify-between items-center">
          <div class="flex gap-3">
            <slot name="extra-actions"></slot>
          </div>

          <button 
            class="bg-primary hover:bg-primary-hover text-navbar-text px-6 py-2 rounded-lg font-bold shadow-md transition-colors active:scale-95" 
            @click="$emit('close')"
          >
            Close
          </button>
        </footer>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95);
  opacity: 0;
}
</style>