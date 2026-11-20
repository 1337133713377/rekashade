<script setup lang="ts">
import { TriangleAlert, Copyright, ShieldCheck, Scale } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'

const LIBS = [
  ['Wails', 'MIT', 'https://github.com/wailsapp/wails'],
  ['Go x/sys', 'BSD-3-Clause', 'https://pkg.go.dev/golang.org/x/sys'],
  ['Vue.js', 'MIT', 'https://github.com/vuejs/core'],
  ['Reka UI', 'MIT', 'https://github.com/unovue/reka-ui'],
  ['Lucide', 'ISC', 'https://github.com/lucide-icons/lucide'],
  ['Plus Jakarta Sans', 'OFL-1.1', 'https://github.com/tokotype/PlusJakartaSans'],
  ['Unbounded (logo)', 'OFL-1.1', 'https://github.com/w3f/unbounded'],
]
</script>

<template>
  <div class="page">
    <PageHeader :title="t('legal.title')" />
    <div class="grid">
      <section class="glass card">
        <div class="row"><TriangleAlert :size="19" class="acc" /><h2 class="card-title sm">{{ t('legal.disclaimerTitle') }}</h2></div>
        <p class="muted small">{{ t('legal.disclaimer') }}</p>
      </section>
      <section class="glass card">
        <div class="row"><ShieldCheck :size="19" class="acc" /><h2 class="card-title sm">{{ t('legal.privacyTitle') }}</h2></div>
        <p class="muted small">{{ t('legal.privacy') }}</p>
      </section>
      <section class="glass card">
        <div class="row"><Copyright :size="19" class="acc" /><h2 class="card-title sm">{{ t('legal.trademarksTitle') }}</h2></div>
        <p class="muted small">{{ t('legal.trademarks') }}</p>
      </section>
      <section class="glass card">
        <div class="row"><Scale :size="19" class="acc" /><h2 class="card-title sm">{{ t('legal.thirdParty') }}</h2></div>
        <p class="dim small">{{ t('legal.thirdPartyHint') }}</p>
        <div class="libs">
          <button v-for="[n, l, u] in LIBS" :key="n" class="lib" @click="api.openURL(u)"><span class="grow">{{ n }}</span><span class="chip">{{ l }}</span></button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; align-items: start; }
.card { padding: 18px 22px; display: flex; flex-direction: column; gap: 10px; }
.card-title.sm { font-size: 17px; }
.acc { color: var(--a1); }
.libs { display: flex; flex-direction: column; gap: 6px; }
.lib { display: flex; align-items: center; gap: 10px; height: 38px; padding: 0 12px; border-radius: 12px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; font-weight: 650; font-size: 13px; text-align: left; }
.lib:hover { border-color: rgba(255, 255, 255, 0.2); }
</style>
