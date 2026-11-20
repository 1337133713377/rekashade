<script setup lang="ts">
import { computed, type Component } from 'vue'
import { Contrast, Layers, Maximize2, Crosshair, Focus, Keyboard, Gauge, Palette, Users, User, Settings, Info, FileText } from '@lucide/vue'
import { t } from '@/lib/i18n'
import { currentDisplay, lookName, state } from '@/lib/store'

const menu: { id: string; icon: Component }[] = [
  { id: 'display', icon: Contrast }, { id: 'presets', icon: Layers }, { id: 'resolution', icon: Maximize2 },
  { id: 'crosshair', icon: Crosshair }, { id: 'reticle', icon: Focus }, { id: 'binds', icon: Keyboard },
  { id: 'optimizer', icon: Gauge }, { id: 'theme', icon: Palette }, { id: 'community', icon: Users },
]
const general: { id: string; icon: Component }[] = [
  { id: 'account', icon: User }, { id: 'settings', icon: Settings }, { id: 'about', icon: Info }, { id: 'legal', icon: FileText },
]
const on = computed(() => state.cfg.tuning.enabled)
const look = computed(() => (currentDisplay.value ? lookName(currentDisplay.value.id) : '—'))
const disp = computed(() => (currentDisplay.value ? t('common.display', { n: currentDisplay.value.number }) : ''))
</script>

<template>
  <nav class="sb">
    <div class="sec">{{ t('nav.menu') }}</div>
    <button v-for="it in menu" :key="it.id" class="item" :class="{ active: state.page === it.id }" @click="state.page = it.id">
      <component :is="it.icon" :size="19" :stroke-width="1.9" />
      <span>{{ t('nav.' + it.id) }}</span>
    </button>
    <div class="sec">{{ t('nav.general') }}</div>
    <button v-for="it in general" :key="it.id" class="item" :class="{ active: state.page === it.id }" @click="state.page = it.id">
      <component :is="it.icon" :size="19" :stroke-width="1.9" />
      <span>{{ t('nav.' + it.id) }}</span>
    </button>
    <div class="spacer" />
    <div class="tcard glass">
      <div class="row">
        <span class="label-caps">{{ t('nav.tuning') }}</span>
        <div class="spacer" />
        <button class="pill-on" :class="{ off: !on }" @click="state.cfg.tuning.enabled = !on">{{ on ? t('common.on') : t('common.off') }}</button>
      </div>
      <div class="tname">{{ on ? look : t('nav.tuningOff') }}</div>
      <div class="tsub">{{ state.status.rule ? state.status.rule : disp }}</div>
    </div>
  </nav>
</template>

<style scoped>
.sb { position: relative; z-index: 2; display: flex; flex-direction: column; gap: 2px; width: 236px; flex: none; padding: 14px 10px 14px 10px; border-right: 1px solid rgba(255, 255, 255, 0.05); background: rgba(8, 6, 16, 0.28); }
.sec { padding: 12px 14px 6px; font-size: 11.5px; font-weight: 800; letter-spacing: 0.1em; color: var(--text-4); }
.sec:first-child { padding-top: 4px; }
.item {
  display: flex; align-items: center; gap: 13px; height: 38px; padding: 0 14px; border-radius: 13px; border: 1px solid transparent;
  background: transparent; color: var(--text-2); font-weight: 650; font-size: 14.5px; cursor: pointer; text-align: left;
  transition: background var(--t-fast), color var(--t-fast), border-color var(--t-fast);
}
.item:hover { background: rgba(255, 255, 255, 0.045); color: var(--text); }
.item.active {
  color: var(--text); border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}
.item.active :deep(svg) { color: var(--a1); }
.tcard { padding: 14px 16px 14px; border-radius: 18px; margin-top: 8px; }
.tcard .pill-on { border: 0; cursor: pointer; height: 24px; }
.tname { margin-top: 10px; font-size: 16px; font-weight: 780; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tsub { font-size: 12.5px; color: var(--text-3); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>
