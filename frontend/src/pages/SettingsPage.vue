<script setup lang="ts">
import { computed, ref } from 'vue'
import { RefreshCw, FolderOpen, ClipboardCopy, Power, ShieldCheck, Loader2 } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSelect from '@/components/RSelect.vue'
import SettingRow from '@/components/SettingRow.vue'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { copyText, fail, state, toast } from '@/lib/store'

const lang = computed({ get: () => state.cfg.language, set: (v: string) => { state.cfg.language = v as 'auto' | 'en' | 'fr' } })
const langOptions = computed(() => [
  { value: 'auto', label: t('set.langAuto') }, { value: 'en', label: 'English' }, { value: 'fr', label: 'Français' },
])
const gpus = computed(() => (state.status.display.gpu ?? []).map((g) => g.toUpperCase()).join(', '))
const checking = ref(false)

async function check() {
  checking.value = true
  try {
    const r = await api.checkForUpdates()
    toast(r.available ? 'info' : 'success', r.available ? t('about.available', { v: r.latest?.version }) : t('about.upToDate'))
  } catch (e) { toast('error', t('about.checkFailed', { error: e instanceof Error ? e.message : String(e) })) } finally { checking.value = false }
}
async function diag() {
  try { await copyText(await api.diagnostics(), t('set.copied')) } catch (e) { fail(e) }
}
async function panic() {
  try { await api.panic(); toast('success', t('toast.panic')) } catch (e) { fail(e) }
}
async function extGamma() {
  try { await api.enableExtendedGamma() } catch (e) { fail(e) }
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('set.title')" :subtitle="t('set.subtitle')" />
    <div class="grid">
      <div class="col-stack">
        <section class="glass card">
          <h2 class="card-title sm">{{ t('set.general') }}</h2>
          <SettingRow :title="t('set.language')"><RSelect v-model="lang" :options="langOptions" /></SettingRow>
          <SettingRow :title="t('set.startWithWindows')"><RSwitch v-model="state.cfg.settings.startWithWindows" :label="t('set.startWithWindows')" /></SettingRow>
          <SettingRow :title="t('set.startMinimized')"><RSwitch v-model="state.cfg.settings.startMinimized" :disabled="!state.cfg.settings.startWithWindows" :label="t('set.startMinimized')" /></SettingRow>
          <SettingRow :title="t('set.closeToTray')"><RSwitch v-model="state.cfg.settings.closeToTray" :label="t('set.closeToTray')" /></SettingRow>
        </section>
        <section class="glass card">
          <h2 class="card-title sm">{{ t('set.updates') }}</h2>
          <SettingRow :title="t('set.autoUpdate')"><RSwitch v-model="state.cfg.settings.autoUpdate" :label="t('set.autoUpdate')" /></SettingRow>
          <SettingRow :title="t('set.autoInstall')"><RSwitch v-model="state.cfg.settings.autoInstall" :disabled="!state.cfg.settings.autoUpdate" :label="t('set.autoInstall')" /></SettingRow>
          <SettingRow :title="t('about.version', { v: state.version })" :hint="state.status.update?.available ? t('about.available', { v: state.status.update.latest?.version }) : ''">
            <button class="btn sm" :disabled="checking" @click="check"><Loader2 v-if="checking" :size="15" class="spin" /><RefreshCw v-else :size="15" /> {{ t('set.checkNow') }}</button>
          </SettingRow>
        </section>
      </div>
      <div class="col-stack">
        <section class="glass card">
          <h2 class="card-title sm">{{ t('set.display') }}</h2>
          <SettingRow :title="t('set.applyOnStartup')"><RSwitch v-model="state.cfg.settings.applyOnStartup" :label="t('set.applyOnStartup')" /></SettingRow>
          <SettingRow :title="t('set.restoreOnExit')"><RSwitch v-model="state.cfg.settings.restoreOnExit" :label="t('set.restoreOnExit')" /></SettingRow>
          <SettingRow :title="t('set.gpu')" :hint="state.status.gpuProbeFailed ? t('set.gpuFailed') : gpus ? t('set.gpuDetected', { list: gpus }) : t('set.gpuNone')">
            <RSwitch v-model="state.cfg.settings.gpuFeatures" :label="t('set.gpu')" />
          </SettingRow>
          <SettingRow :title="t('set.extGamma')" :hint="t('set.extGammaHint')">
            <span v-if="state.status.extendedGamma" class="chip ok"><ShieldCheck :size="13" /> {{ t('set.extGammaOn') }}</span>
            <button v-else class="btn sm" @click="extGamma">{{ t('set.enable') }}</button>
          </SettingRow>
        </section>
        <section class="glass card">
          <h2 class="card-title sm">{{ t('set.data') }}</h2>
          <div class="row wrap actions">
            <button class="btn sm" @click="api.openDataFolder().catch(fail)"><FolderOpen :size="15" /> {{ t('set.openFolder') }}</button>
            <button class="btn sm" @click="diag"><ClipboardCopy :size="15" /> {{ t('set.diagnostics') }}</button>
          </div>
          <SettingRow :title="t('set.panic')" :hint="t('set.panicHint')">
            <button class="btn sm danger" @click="panic"><Power :size="15" /> {{ t('set.panic') }}</button>
          </SettingRow>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; align-items: start; }
.col-stack { display: flex; flex-direction: column; gap: 14px; }
.card { padding: 16px 22px 8px; display: flex; flex-direction: column; }
.card-title.sm { font-size: 17px; margin-bottom: 2px; }
.actions { padding: 12px 0 4px; gap: 8px; }
.card :deep(.sel) { min-width: 160px; }
</style>
