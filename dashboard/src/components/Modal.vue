<script setup>
const props = defineProps({
  show: Boolean
})
</script>

<template>
  <Transition name="modal" @after-enter="$emit('opened')">
    <div v-if="show" class="modal-mask">
      <div class="modal-container flex flex-col h-5/6 w-11/12 max-w-6xl">
        <div class="modal-header">
          <slot name="header">default header</slot>
        </div>

        <div class="modal-body flex-1 overflow-hidden">
          <slot name="body"></slot>
        </div>

        <div class="modal-footer">
          <slot name="footer">
            <button class="modal-default-button" @click="$emit('close')">Close</button>
          </slot>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background-color: rgba(0,0,0,0.7);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal-container {
  background-color: #1e293b;
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  padding: 0.5rem 1rem;
  border-bottom: 1px solid #334155;
  color: #f1f5f9;
  font-weight: bold;
  font-size: 1.2rem;
}

.modal-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 1rem;
}

.modal-footer {
  padding: 0.5rem 1rem;
  border-top: 1px solid #334155;
  display: flex;
  justify-content: flex-end;
}

.modal-default-button {
  padding: 0.4rem 1rem;
  background-color: #3b82f6;
  color: #fff;
  border-radius: 0.25rem;
  cursor: pointer;
  border: none;
}

.modal-default-button:hover {
  background-color: #2563eb;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(1.05);
}
</style>