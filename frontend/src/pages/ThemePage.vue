<script setup lang="ts">
import { Check } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import Segmented from '@/components/Segmented.vue'
import SettingRow from '@/components/SettingRow.vue'
import { t } from '@/lib/i18n'
import { state } from '@/lib/store'
import type { Accent, Backdrop } from '@/lib/types'

const ACCENTS: { id: Accent; c: [string, string] }[] = [
  { id: 'violet', c: ['#c4a8ff', '#7c3aed'] }, { id: 'magenta', c: ['#f5b6ff', '#b81fd1'] }, { id: 'cyan', c: ['#a5f3ff', '#0e8fc4'] },
  { id: 'emerald', c: ['#a7f3d0', '#059669'] }, { id: 'amber', c: ['#fde68a', '#d97706'] }, { id: 'crimson', c: ['#fecdd3', '#e11d48'] },
  { id: 'ice', c: ['#e0f2fe', '#6366f1'] },
]
const BACKDROPS: { id: Backdrop; bg: string }[] = [
  { id: 'aurora', bg: 'radial-gradient(80% 90% at 55% 0%, #5b2bb5, #110d1f 70%)' },
  { id: 'nebula', bg: 'radial-gradient(70% 80% at 80% 100%, #9d174d, transparent 70%), radial-gradient(70% 80% at 0% 80%, #1e3a8a, #0d0716 75%)' },
  { id: 'midnight', bg: 'radial-gradient(80% 90% at 50% 0%, #1e2a5a, #070912 70%)' },
  { id: 'oled', bg: '#000' },
]
</script>

<template>
  <div class="page">
    <PageHeader :title="t('theme.title')" :subtitle="t('theme.subtitle')" />
    <section class="glass card">
      <h2 class="card-title sm">{{ t('theme.accent') }}</h2>
      <div class="accents">
        <button v-for="a in ACCENTS" :key="a.id" class="acc" :class="{ active: state.cfg.theme.accent === a.id }" @click="state.cfg.theme.accent = a.id">
          <span class="dot swatch" :style="{ background: `linear-gradient(135deg, ${a.c[0]}, ${a.c[1]})` }"><Check v-if="state.cfg.theme.accent === a.id" :size="18" :stroke-width="3" /></span>
          <span>{{ t('theme.accents.' + a.id) }}</span>
        </button>
      </div>
    </section>
    <section class="glass card">
      <h2 class="card-title sm">{{ t('theme.backdrop') }}</h2>
      <div class="bds">
        <button v-for="b in BACKDROPS" :key="b.id" class="bd" :class="{ active: state.cfg.theme.backdrop === b.id }" @click="state.cfg.theme.backdrop = b.id">
          <span class="bd-prev swatch" :style="{ background: b.bg }"><span class="bd-card" /><span class="bd-card s" /></span>
          <span class="bd-name">{{ t('theme.backdrops.' + b.id) }}</span>
        </button>
      </div>
    </section>
    <section class="glass card rows">
      <SettingRow :title="t('theme.glass')" :hint="t('theme.glassHint')"><RSwitch v-model="state.cfg.theme.glass" :label="t('theme.glass')" /></SettingRow>
      <SettingRow :title="t('theme.motion')" :hint="t('theme.motionHint')"><RSwitch v-model="state.cfg.theme.motion" :label="t('theme.motion')" /></SettingRow>
      <SettingRow :title="t('theme.scale')">
        <Segmented :model-value="String(state.cfg.theme.scale)" :options="[{ value: '90', label: '90%' }, { value: '100', label: '100%' }, { value: '110', label: '110%' }]" @update:model-value="(v: string) => (state.cfg.theme.scale = Number(v))" />
      </SettingRow>
    </section>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.card { padding: 18px 22px; display: flex; flex-direction: column; gap: 14px; }
.card.rows { gap: 0; padding-top: 6px; padding-bottom: 6px; }
.card-title.sm { font-size: 17px; }
.accents { display: flex; gap: 18px; flex-wrap: wrap; }
.acc { display: flex; flex-direction: column; align-items: center; gap: 8px; border: 0; background: none; cursor: pointer; color: var(--text-2); font-weight: 650; font-size: 12.5px; }
.acc .dot { width: 54px; height: 54px; border-radius: 50%; display: grid; place-items: center; color: #fff; box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4); transition: transform var(--t-fast); }
.acc:hover .dot { transform: scale(1.05); }
.acc.active { color: var(--text); }
.acc.active .dot { box-shadow: 0 0 0 3px var(--bg), 0 0 0 5px rgba(255, 255, 255, 0.85); }
.bds { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.bd { display: flex; flex-direction: column; gap: 8px; padding: 6px 6px 10px; border-radius: 16px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; }
.bd.active { border-color: var(--a2); box-shadow: 0 0 0 1px var(--a2); }
.bd-prev { position: relative; height: 92px; border-radius: 11px; overflow: hidden; display: flex; gap: 6px; padding: 12px; }
.bd-card { flex: 1; border-radius: 8px; background: rgba(255, 255, 255, 0.08); border: 1px solid rgba(255, 255, 255, 0.08); }
.bd-card.s { flex: 0.6; }
.bd-name { font-weight: 700; font-size: 13px; }
</style>
