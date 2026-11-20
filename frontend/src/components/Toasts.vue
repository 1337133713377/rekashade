<script setup lang="ts">
import { CircleCheck, CircleAlert, Info } from '@lucide/vue'
import { state } from '@/lib/store'
</script>

<template>
  <div class="toasts" aria-live="polite">
    <TransitionGroup name="toast">
      <div v-for="m in state.toasts" :key="m.id" class="toast" :class="m.kind">
        <CircleCheck v-if="m.kind === 'success'" :size="18" />
        <CircleAlert v-else-if="m.kind === 'error'" :size="18" />
        <Info v-else :size="18" />
        <span>{{ m.text }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toasts { position: fixed; right: 22px; bottom: 22px; z-index: 90; display: flex; flex-direction: column; gap: 10px; align-items: flex-end; pointer-events: none; }
.toast {
  display: flex; align-items: center; gap: 10px; max-width: 420px; padding: 12px 16px; border-radius: 14px; font-weight: 600; font-size: 13.5px;
  background: rgba(26, 20, 46, 0.96); border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: 0 18px 40px -12px rgba(0, 0, 0, 0.7); pointer-events: auto;
}
.toast.success :deep(svg) { color: var(--ok); }
.toast.error { border-color: rgba(248, 113, 113, 0.35); }
.toast.error :deep(svg) { color: var(--bad); }
.toast.info :deep(svg) { color: var(--a1); }
.toast-enter-active, .toast-leave-active { transition: all 0.28s var(--ease); }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(10px) scale(0.98); }
</style>
