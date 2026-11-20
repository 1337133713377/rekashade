<script setup lang="ts">
import { computed } from 'vue'
import { Download, Upload, ShieldCheck, RotateCcw } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import { api } from '@/lib/api'
import { locale, t } from '@/lib/i18n'
import { confirmAsk, factoryReset, fail, importConfig, state, toast } from '@/lib/store'

const GRADS = ['#a78bfa,#7c3aed', '#f472b6,#be185d', '#22d3ee,#0e7490', '#4ade80,#15803d', '#fbbf24,#c2410c', '#f87171,#9f1239']
const initials = computed(() => (state.cfg.profile.name || 'Reka').trim().split(/\s+/).map((w) => w[0]).join('').slice(0, 2).toUpperCase())
const since = computed(() => {
  const d = new Date(state.cfg.profile.since)
  return isNaN(d.getTime()) ? '' : d.toLocaleDateString(locale.value, { day: 'numeric', month: 'long', year: 'numeric' })
})
const stats = computed(() => [
  { k: 'statPresets', v: state.cfg.presets.filter((p) => !p.builtin).length },
  { k: 'statCrosshairs', v: state.cfg.crosshair.saved.filter(Boolean).length },
  { k: 'statBinds', v: state.cfg.binds.filter((b) => b.key).length },
  { k: 'statRules', v: state.cfg.rules.length },
])

async function exportIt() {
  try { const p = await api.exportSettings(); if (p) toast('success', t('acc.exported', { path: p })) } catch (e) { fail(e) }
}
async function importIt() {
  try {
    const cfg = await api.importSettings()
    if (cfg) { importConfig(cfg); toast('success', t('acc.imported')) }
  } catch (e) { fail(e) }
}
async function resetAll() {
  if (!(await confirmAsk({ title: t('acc.resetConfirm'), body: t('acc.resetHint'), confirmLabel: t('acc.reset'), danger: true }))) return
  factoryReset()
  toast('success', t('acc.resetDone'))
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('acc.title')" :subtitle="t('acc.subtitle')" />
    <div class="grid">
      <section class="glass card profile">
        <div class="avatar" :style="{ background: `linear-gradient(135deg, ${GRADS[state.cfg.profile.avatar % GRADS.length]})` }">{{ initials }}</div>
        <div class="col grow">
          <label class="label-caps" for="pname">{{ t('acc.name') }}</label>
          <input id="pname" v-model="state.cfg.profile.name" class="input" maxlength="24" placeholder="Reka" spellcheck="false" />
          <div class="row">
            <span class="label-caps">{{ t('acc.avatar') }}</span>
            <button v-for="(g, i) in GRADS" :key="g" class="grad" :class="{ active: state.cfg.profile.avatar === i }" :style="{ background: `linear-gradient(135deg, ${g})` }" :aria-label="'avatar ' + (i + 1)" @click="state.cfg.profile.avatar = i" />
          </div>
          <p v-if="since" class="dim small">{{ t('acc.since', { date: since }) }}</p>
        </div>
      </section>
      <section class="glass card">
        <h2 class="card-title sm">{{ t('acc.stats') }}</h2>
        <div class="stats">
          <div v-for="s in stats" :key="s.k" class="stat"><span class="num mono">{{ s.v }}</span><span class="dim small">{{ t('acc.' + s.k) }}</span></div>
        </div>
      </section>
      <section class="glass card">
        <h2 class="card-title sm">{{ t('acc.backup') }}</h2>
        <p class="dim small">{{ t('acc.backupHint') }}</p>
        <div class="row">
          <button class="btn" @click="exportIt"><Upload :size="16" /> {{ t('acc.export') }}</button>
          <button class="btn" @click="importIt"><Download :size="16" /> {{ t('acc.import') }}</button>
        </div>
      </section>
      <section class="glass card">
        <div class="row"><ShieldCheck :size="20" class="acc" /><h2 class="card-title sm">{{ t('acc.privacy') }}</h2></div>
        <p class="muted small">{{ t('acc.privacyText') }}</p>
        <div class="spacer" />
        <div class="row">
          <button class="btn danger" @click="resetAll"><RotateCcw :size="16" /> {{ t('acc.reset') }}</button>
          <span class="dim tiny">{{ t('acc.resetHint') }}</span>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.card { padding: 18px 22px; display: flex; flex-direction: column; gap: 12px; }
.card-title.sm { font-size: 17px; }
.profile { flex-direction: row; align-items: flex-start; gap: 20px; }
.avatar { width: 92px; height: 92px; flex: none; border-radius: 28px; display: grid; place-items: center; font-size: 32px; font-weight: 800; color: #fff; box-shadow: 0 16px 34px -14px var(--a-glow), inset 0 1px 0 rgba(255, 255, 255, 0.35); }
.grad { width: 24px; height: 24px; border-radius: 50%; border: 0; cursor: pointer; }
.grad.active { box-shadow: 0 0 0 2px var(--bg), 0 0 0 4px #fff; }
.stats { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.stat { display: flex; flex-direction: column; gap: 2px; padding: 12px 14px; border-radius: 14px; background: rgba(0, 0, 0, 0.2); border: 1px solid var(--field-border); }
.num { font-size: 26px; font-weight: 800; }
.acc { color: var(--a1); }
</style>
