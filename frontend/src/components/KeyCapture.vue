<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { bindLabel, vkFromEvent, vkFromMouse } from '@/lib/keys'
import type { Bind } from '@/lib/types'

const props = defineProps<{ bind: Bind; hold?: boolean }>()
const emit = defineEmits<{ set: [b: Pick<Bind, 'key' | 'ctrl' | 'alt' | 'shift'>] }>()
const capturing = ref(false)

function stop() {
  capturing.value = false
  window.removeEventListener('keydown', onKey, true)
  window.removeEventListener('mousedown', onMouse, true)
  window.removeEventListener('contextmenu', block, true)
  api.pauseBinds(false).catch(() => {})
}
function block(e: Event) { e.preventDefault() }
function onKey(e: KeyboardEvent) {
  e.preventDefault()
  e.stopPropagation()
  if (e.key === 'Escape') { stop(); return }
  const vk = vkFromEvent(e)
  if (!vk) return // wait for a non-modifier key
  emit('set', props.hold ? { key: vk, ctrl: false, alt: false, shift: false } : { key: vk, ctrl: e.ctrlKey, alt: e.altKey, shift: e.shiftKey })
  stop()
}
function onMouse(e: MouseEvent) {
  if (e.button === 0) return // left click stays a click
  e.preventDefault()
  e.stopPropagation()
  const vk = vkFromMouse(e.button)
  if (vk) emit('set', { key: vk, ctrl: props.hold ? false : e.ctrlKey, alt: props.hold ? false : e.altKey, shift: props.hold ? false : e.shiftKey })
  stop()
}
function start() {
  if (capturing.value) { stop(); return }
  capturing.value = true
  api.pauseBinds(true).catch(() => {})
  setTimeout(() => {
    window.addEventListener('keydown', onKey, true)
    window.addEventListener('mousedown', onMouse, true)
    window.addEventListener('contextmenu', block, true)
  }, 0)
}
onBeforeUnmount(() => { if (capturing.value) stop() })
</script>

<template>
  <button class="kc" :class="{ capturing, empty: !bind.key }" @click="start">
    <template v-if="capturing">{{ t('binds.press') }}</template>
    <template v-else-if="bind.key">
      <kbd v-for="(p, i) in bindLabel(bind).split(' + ')" :key="i">{{ p }}</kbd>
    </template>
    <template v-else>{{ t('binds.unbound') }}</template>
  </button>
</template>

<style scoped>
.kc { display: inline-flex; align-items: center; gap: 5px; min-width: 150px; height: 38px; padding: 0 12px; border-radius: 12px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.22); cursor: pointer; font-weight: 650; font-size: 13px; color: var(--text-2); }
.kc:hover { border-color: rgba(255, 255, 255, 0.18); }
.kc.empty { color: var(--text-4); }
.kc.capturing { border-color: var(--a2); color: var(--a-text); box-shadow: 0 0 0 4px var(--a-soft); animation: pulse-dot 1.4s ease-in-out infinite; }
kbd { display: inline-flex; align-items: center; height: 24px; padding: 0 8px; border-radius: 7px; font-family: var(--font); font-size: 12px; font-weight: 750; color: var(--text); background: rgba(255, 255, 255, 0.09); border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: inset 0 -2px 0 rgba(0, 0, 0, 0.3); }
</style>
