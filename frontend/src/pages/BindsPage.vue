<script setup lang="ts">
import { computed } from 'vue'
import { X, RotateCcw, ShieldCheck, Lightbulb } from '@lucide/vue'
import defaults from '@defaults'
import PageHeader from '@/components/PageHeader.vue'
import KeyCapture from '@/components/KeyCapture.vue'
import { t } from '@/lib/i18n'
import { state } from '@/lib/store'
import type { Bind } from '@/lib/types'

const ORDER = ['toggleTuning', 'toggleCrosshair', 'nextPreset', 'prevPreset', 'toggleStretch', 'showApp', 'panic', 'holdHide']
const binds = computed(() => [...state.cfg.binds].sort((a, b) => ORDER.indexOf(a.action) - ORDER.indexOf(b.action)))

const sig = (b: Pick<Bind, 'key' | 'ctrl' | 'alt' | 'shift'>) => `${b.key}|${b.ctrl}|${b.alt}|${b.shift}`
function conflict(b: Bind): string {
  if (!b.key) return ''
  const other = state.cfg.binds.find((x) => x !== b && x.key && sig(x) === sig(b))
  return other ? t('binds.conflict', { action: t('binds.' + other.action) }) : ''
}
function set(b: Bind, v: Pick<Bind, 'key' | 'ctrl' | 'alt' | 'shift'>) { Object.assign(b, v) }
function clear(b: Bind) { Object.assign(b, { key: 0, ctrl: false, alt: false, shift: false }) }
function restore() { state.cfg.binds = (defaults.binds as Bind[]).map((b) => ({ ...b })) }
</script>

<template>
  <div class="page">
    <PageHeader :title="t('binds.title')" :subtitle="t('binds.subtitle')">
      <button class="btn" @click="restore"><RotateCcw :size="16" /> {{ t('binds.reset') }}</button>
    </PageHeader>
    <div class="grid">
      <section class="glass card">
        <div v-for="b in binds" :key="b.action" class="bind">
          <div class="grow">
            <div class="bname">{{ t('binds.' + b.action) }} <span class="chip" :class="b.action === 'holdHide' ? 'accent' : ''">{{ b.action === 'holdHide' ? t('binds.hold') : t('binds.toggle') }}</span></div>
            <div v-if="conflict(b)" class="conf small">{{ conflict(b) }}</div>
          </div>
          <KeyCapture :bind="b" :hold="b.action === 'holdHide'" @set="(v) => set(b, v)" />
          <button class="btn ghost icon" :disabled="!b.key" :aria-label="t('binds.clear')" :title="t('binds.clear')" @click="clear(b)"><X :size="17" /></button>
        </div>
        <p class="dim small">{{ t('binds.mouseHint') }}</p>
      </section>
      <aside class="side">
        <section class="glass card">
          <div class="row"><ShieldCheck :size="20" class="acc" /><h2 class="card-title sm">{{ t('binds.how') }}</h2></div>
          <p class="muted small">{{ t('binds.howText') }}</p>
        </section>
        <section class="glass card">
          <div class="row"><Lightbulb :size="20" class="acc" /><h2 class="card-title sm">{{ t('binds.tip') }}</h2></div>
          <p class="muted small">{{ t('binds.holdHint') }}</p>
        </section>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.grid { display: grid; grid-template-columns: 1fr 300px; gap: 14px; align-items: start; }
.card { padding: 16px 20px; display: flex; flex-direction: column; gap: 4px; }
.side { display: flex; flex-direction: column; gap: 14px; }
.side .card { gap: 10px; }
.card-title.sm { font-size: 17px; }
.acc { color: var(--a1); }
.bind { display: flex; align-items: center; gap: 12px; padding: 11px 0; border-bottom: 1px solid var(--panel-border); }
.bind:last-of-type { border-bottom: 0; }
.bname { display: flex; align-items: center; gap: 10px; font-weight: 680; font-size: 14.5px; }
.bname .chip { height: 19px; font-size: 10px; padding: 0 7px; }
.conf { color: #fca5a5; margin-top: 3px; }
</style>
