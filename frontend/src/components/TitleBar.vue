<script setup lang="ts">
import { computed } from 'vue'
import { Minus, X, Sparkles } from '@lucide/vue'
import wordmark from '@/assets/brand/reka-wordmark.svg'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { state } from '@/lib/store'

const active = computed(() => state.cfg.tuning.enabled || state.status.overlay.visible || state.status.stretch.active)
const upd = computed(() => (state.status.update?.available ? state.status.update.latest?.version : ''))
const parts = computed(() => {
  const p: string[] = []
  if (state.cfg.tuning.enabled) p.push(t('nav.display'))
  if (state.status.overlay.visible) p.push(t('nav.crosshair'))
  if (state.status.stretch.active) p.push(t('nav.resolution'))
  return p.join(' · ')
})
</script>

<template>
  <header class="tb" style="--wails-draggable: drag" @dblclick.prevent>
    <div class="brand">
      <img :src="wordmark" class="word" alt="Reka" draggable="false" />
      <span class="shade">SHADE</span>
      <span class="ver mono">v{{ state.version }}</span>
    </div>
    <div class="spacer" />
    <button v-if="upd" class="upd" style="--wails-draggable: no-drag" @click="state.page = 'about'">
      <Sparkles :size="14" /> {{ t('nav.update', { v: upd }) }}
    </button>
    <button class="status" style="--wails-draggable: no-drag" :title="parts" @click="state.page = 'display'">
      <span class="dot" :class="{ on: active }" />
      {{ active ? t('nav.statusActive') : t('nav.statusIdle') }}
    </button>
    <button class="wbtn" style="--wails-draggable: no-drag" :aria-label="t('common.minimize')" @click="api.windowMinimize()"><Minus :size="18" /></button>
    <button class="wbtn close" style="--wails-draggable: no-drag" :aria-label="t('common.closeWindow')" @click="api.windowClose()"><X :size="18" /></button>
  </header>
</template>

<style scoped>
.tb { position: relative; z-index: 5; display: flex; align-items: center; gap: 10px; height: 58px; padding: 0 12px 0 22px; border-bottom: 1px solid rgba(255, 255, 255, 0.05); background: rgba(8, 6, 16, 0.35); }
.brand { display: flex; align-items: center; gap: 10px; }
.word { height: 17px; display: block; }
.shade { font-size: 11px; font-weight: 800; letter-spacing: 0.32em; color: var(--text-3); margin-left: -2px; }
.ver { height: 24px; padding: 0 10px; display: inline-flex; align-items: center; border-radius: 999px; font-size: 12px; font-weight: 700; color: var(--text-2); background: rgba(255, 255, 255, 0.06); border: 1px solid var(--field-border); }
.status, .upd {
  display: inline-flex; align-items: center; gap: 9px; height: 36px; padding: 0 16px; border-radius: 999px; cursor: pointer;
  background: rgba(255, 255, 255, 0.06); border: 1px solid var(--field-border); font-weight: 650; font-size: 13.5px; color: var(--text-2);
}
.status:hover, .upd:hover { background: rgba(255, 255, 255, 0.1); color: var(--text); }
.upd { color: var(--a-text); background: var(--a-soft); border-color: transparent; }
.dot { width: 9px; height: 9px; border-radius: 50%; background: var(--text-4); }
.dot.on { background: var(--a2); box-shadow: 0 0 10px var(--a2); animation: pulse-dot 2.4s ease-in-out infinite; }
.wbtn { width: 42px; height: 36px; display: grid; place-items: center; border-radius: 10px; border: 0; background: transparent; color: var(--text-2); cursor: pointer; }
.wbtn:hover { background: rgba(255, 255, 255, 0.08); color: var(--text); }
.wbtn.close:hover { background: #e5484d; color: #fff; }
</style>
