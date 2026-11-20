<script setup lang="ts">
import { computed, ref } from 'vue'
import { RefreshCw, Download, Loader2, ShieldCheck, Ban, Info, ExternalLink, CircleCheck } from '@lucide/vue'
import lockup from '@/assets/brand/reka-wordmark.svg'
import PageHeader from '@/components/PageHeader.vue'
import GithubIcon from '@/components/GithubIcon.vue'
import { api, errText } from '@/lib/api'
import { locale, t } from '@/lib/i18n'
import { state, toast } from '@/lib/store'

const REPO = 'https://github.com/1337133713377/rekashade'
const checking = ref(false)
const installing = ref(false)
const error = ref('')
const upd = computed(() => state.status.update)
const pct = computed(() => {
  const p = state.updateProgress
  return p && p.total > 0 ? Math.min(100, Math.round((p.done / p.total) * 100)) : 0
})
const checkedAt = computed(() => {
  const d = upd.value?.checkedAt ? new Date(upd.value.checkedAt) : null
  return d && !isNaN(d.getTime()) ? d.toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' }) : ''
})

async function check() {
  checking.value = true
  error.value = ''
  try { await api.checkForUpdates() } catch (e) { error.value = t('about.checkFailed', { error: errText(e) }) } finally { checking.value = false }
}
async function install() {
  installing.value = true
  error.value = ''
  try { await api.installUpdate() } catch (e) { error.value = errText(e); toast('error', t('toast.updateFailed', { error: errText(e) })) } finally { installing.value = false }
}
const safe = ['safe1', 'safe2', 'safe3', 'safe4', 'safe5']
const credits = [
  ['Wails', 'https://wails.io'], ['Vue', 'https://vuejs.org'], ['Reka UI', 'https://reka-ui.com'],
  ['Lucide', 'https://lucide.dev'], ['Plus Jakarta Sans', 'https://fonts.google.com/specimen/Plus+Jakarta+Sans'], ['Unbounded', 'https://fonts.google.com/specimen/Unbounded'],
]
</script>

<template>
  <div class="page">
    <PageHeader :title="t('about.title')" :subtitle="t('about.tagline')" />
    <div class="grid">
      <section class="glass card hero">
        <img :src="lockup" class="lockup" alt="Reka" />
        <div class="label-caps shade">SHADE</div>
        <div class="row ver"><span class="chip accent mono">{{ t('about.version', { v: state.version }) }}</span></div>
        <div class="upd">
          <template v-if="upd?.available">
            <div class="row"><Download :size="18" class="acc" /><b>{{ t('about.available', { v: upd.latest?.version }) }}</b></div>
            <div v-if="installing || state.status.updating" class="bar"><span :style="{ width: pct + '%' }" /></div>
            <p v-if="installing || state.status.updating" class="dim small">{{ t('about.downloading', { pct }) }}</p>
            <button v-else class="btn primary block" @click="install"><Download :size="16" /> {{ t('about.install') }}</button>
            <details v-if="upd.latest?.notes" class="notes"><summary>{{ t('about.notes') }}</summary><pre>{{ upd.latest.notes }}</pre></details>
          </template>
          <template v-else-if="upd">
            <div class="row"><CircleCheck :size="18" class="ok" /><b>{{ t('about.upToDate') }}</b></div>
          </template>
          <p v-if="checkedAt" class="dim tiny">{{ t('about.lastChecked', { time: checkedAt }) }}</p>
          <p v-if="error" class="err small">{{ error }}</p>
          <div class="row wrap">
            <button class="btn sm" :disabled="checking" @click="check"><Loader2 v-if="checking" :size="16" class="spin" /><RefreshCw v-else :size="16" /> {{ checking ? t('about.checking') : t('about.check') }}</button>
            <button class="btn sm ghost" @click="api.openURL(REPO + '/releases')"><GithubIcon /> GitHub</button>
          </div>
        </div>
      </section>
      <div class="side">
        <section class="glass card">
          <div class="row"><ShieldCheck :size="20" class="acc" /><h2 class="card-title sm">{{ t('about.safeTitle') }}</h2></div>
          <ul class="safe">
            <li v-for="k in safe" :key="k"><CircleCheck :size="16" class="ok" /><span>{{ t('about.' + k) }}</span></li>
          </ul>
          <p class="never small"><Ban :size="16" /><span>{{ t('about.never') }}</span></p>
        </section>
        <section class="glass card">
          <div class="row"><Info :size="20" class="acc" /><h2 class="card-title sm">{{ t('about.limitsTitle') }}</h2></div>
          <p class="muted small">{{ t('about.limits') }}</p>
          <div class="label-caps">{{ t('about.credits') }}</div>
          <div class="row wrap credits">
            <button v-for="[n, u] in credits" :key="n" class="chip" @click="api.openURL(u)">{{ n }} <ExternalLink :size="11" /></button>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.grid { display: grid; grid-template-columns: 330px 1fr; gap: 14px; align-items: start; }
.side { display: flex; flex-direction: column; gap: 14px; }
.card { padding: 18px 22px; display: flex; flex-direction: column; gap: 12px; }
.card-title.sm { font-size: 17px; }
.hero { align-items: stretch; }
.lockup { width: 190px; margin: 10px auto 0; display: block; filter: drop-shadow(0 12px 30px var(--a-glow)); }
.shade { text-align: center; letter-spacing: 0.62em; margin-top: -4px; padding-left: 0.62em; }
.ver { justify-content: center; }
.upd { display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 16px; background: rgba(0, 0, 0, 0.22); border: 1px solid var(--field-border); }
.bar { height: 8px; border-radius: 8px; background: var(--track); overflow: hidden; }
.bar span { display: block; height: 100%; background: var(--accent-grad); transition: width 0.2s; }
.notes summary { cursor: pointer; font-weight: 700; font-size: 13px; }
.notes pre { white-space: pre-wrap; font-family: var(--font); font-size: 12.5px; color: var(--text-2); max-height: 160px; overflow: auto; margin: 8px 0 0; }
.acc { color: var(--a1); flex: none; }
.ok { color: var(--ok); flex: none; }
.err { color: #fca5a5; }
.safe { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 9px; }
.safe li { display: flex; gap: 10px; font-size: 13px; color: var(--text-2); line-height: 1.45; }
.safe li :deep(svg) { margin-top: 2px; }
.never { display: flex; gap: 10px; padding: 10px 12px; border-radius: 12px; background: rgba(248, 113, 113, 0.08); color: #fecaca; line-height: 1.45; }
.never :deep(svg) { flex: none; margin-top: 2px; }
.credits { gap: 6px; }
.credits .chip { cursor: pointer; }
</style>
