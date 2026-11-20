<script setup lang="ts">
import { ref } from 'vue'
import { Download, Copy, ExternalLink, Bug } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import CrosshairTile from '@/components/CrosshairTile.vue'
import PresetThumb from '@/components/PresetThumb.vue'
import GithubIcon from '@/components/GithubIcon.vue'
import { api } from '@/lib/api'
import { decodeAny, encodeCrosshair, encodePreset } from '@/lib/codes'
import { t } from '@/lib/i18n'
import { applyPreset, copyText, currentParams, lookName, currentDisplay, newId, state, toast } from '@/lib/store'
import type { ColorParams, CrosshairCfg } from '@/lib/types'

const REPO = 'https://github.com/1337133713377/rekashade'
const code = ref('')

const XH_PICKS: { name: string; cfg: Partial<CrosshairCfg> }[] = [
  { name: 'Tac Dot', cfg: { ring: 0, arm: 0, gap: 0, thickness: 2, dot: 3, outline: 1, color: '#22E3FF', rotate: false, tStyle: false } },
  { name: 'Classic Gap', cfg: { ring: 0, arm: 5, gap: 3, thickness: 2, dot: 0, outline: 1, color: '#3BFF5B', rotate: false, tStyle: false } },
  { name: 'Neon Ring', cfg: { ring: 8, arm: 0, gap: 0, thickness: 2, dot: 2, outline: 0, color: '#FF2BD6', rotate: false, tStyle: false } },
  { name: 'Sniper Ring', cfg: { ring: 16, arm: 8, gap: 18, thickness: 1, dot: 2, outline: 1, color: '#FFFFFF', rotate: false, tStyle: false } },
  { name: 'Micro T', cfg: { ring: 0, arm: 4, gap: 2, thickness: 1, dot: 1, outline: 1, color: '#FFE81F', rotate: false, tStyle: true } },
  { name: 'Bold X', cfg: { ring: 0, arm: 6, gap: 3, thickness: 3, dot: 0, outline: 1, color: '#FF2D2D', rotate: true, tStyle: false } },
]
const P = (hue: number, saturation: number, vibrance: number, brightness: number, contrast: number, gamma: number, temperature: number, red = 100, green = 100, blue = 100): ColorParams =>
  ({ hue, saturation, vibrance, brightness, contrast, gamma, temperature, red, green, blue })
const PRESET_PICKS: { name: string; params: ColorParams }[] = [
  { name: 'Night Raid', params: P(0, 120, 115, 114, 98, 1.42, 6800) },
  { name: 'Battle Royale Pop', params: P(4, 175, 160, 102, 106, 1.06, 6500, 102, 100, 98) },
  { name: 'Retro CRT', params: P(-6, 115, 105, 98, 118, 0.95, 5600, 104, 100, 94) },
  { name: 'Ice Clarity', params: P(0, 125, 120, 100, 108, 1.1, 7600, 96, 100, 104) },
]

function useCrosshair(c: Partial<CrosshairCfg>) {
  Object.assign(state.cfg.crosshair, c, { design: 'custom' })
  state.cfg.overlay.mode = 'crosshair'
  toast('success', t('com.importedCrosshair'))
}
function usePreset(name: string, params: ColorParams) {
  const p = { id: newId('shared'), name, tags: ['shared'], params: { ...params } }
  state.cfg.presets.push(p)
  applyPreset(p)
  toast('success', t('com.importedPreset', { name }))
}
function importCode() {
  const d = decodeAny(code.value)
  if (!d) { toast('error', t('com.invalid')); return }
  if (d.kind === 'crosshair') useCrosshair(d.data)
  else if (d.kind === 'preset') usePreset(d.data.name, d.data.params)
  else { Object.assign(state.cfg.reticle, d.data); toast('success', t('com.importedReticle')) }
  code.value = ''
}
const xhCfg = (c: Partial<CrosshairCfg>): CrosshairCfg => ({ ...state.cfg.crosshair, ...c, opacity: 100, saved: [] })
</script>

<template>
  <div class="page">
    <PageHeader :title="t('com.title')" :subtitle="t('com.subtitle')" />
    <section class="glass card">
      <h2 class="card-title sm">{{ t('com.paste') }}</h2>
      <form class="row" @submit.prevent="importCode">
        <input v-model="code" class="input mono" :placeholder="t('com.placeholder')" spellcheck="false" />
        <button class="btn primary" type="submit" :disabled="!code.trim()"><Download :size="16" /> {{ t('com.import') }}</button>
      </form>
    </section>
    <div class="grid">
      <section class="glass card">
        <h2 class="card-title sm">{{ t('com.picks') }}</h2>
        <div class="label-caps">{{ t('com.crosshairs') }}</div>
        <div class="xh">
          <button v-for="x in XH_PICKS" :key="x.name" class="pick" @click="useCrosshair(x.cfg)">
            <span class="pick-tile"><CrosshairTile :cfg="xhCfg(x.cfg)" /></span>
            <span class="pick-name">{{ x.name }}</span>
          </button>
        </div>
        <div class="label-caps">{{ t('com.presets') }}</div>
        <div class="pp">
          <button v-for="p in PRESET_PICKS" :key="p.name" class="pick" @click="usePreset(p.name, p.params)">
            <PresetThumb :params="p.params" :uid="'com-' + p.name.replace(/\W/g, '')" :scene="state.cfg.ui.scene" />
            <span class="pick-name">{{ p.name }}</span>
          </button>
        </div>
      </section>
      <aside class="side">
        <section class="glass card">
          <h2 class="card-title sm">{{ t('com.share') }}</h2>
          <p class="dim small">{{ t('com.shareHint') }}</p>
          <button class="btn block" @click="copyText(encodeCrosshair(state.cfg.crosshair))"><Copy :size="16" /> {{ t('com.shareCrosshair') }}</button>
          <button class="btn block" @click="copyText(encodePreset(currentDisplay ? lookName(currentDisplay.id) : 'Look', currentParams))"><Copy :size="16" /> {{ t('com.shareLook') }}</button>
        </section>
        <section class="glass card">
          <h2 class="card-title sm">{{ t('com.links') }}</h2>
          <button class="link" @click="api.openURL(REPO)"><GithubIcon /> <span class="grow">{{ t('com.github') }}</span><ExternalLink :size="15" class="dim" /></button>
          <button class="link" @click="api.openURL(REPO + '/issues')"><Bug :size="18" /> <span class="grow">{{ t('com.issues') }}</span><ExternalLink :size="15" class="dim" /></button>
        </section>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.card { padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; }
.card-title.sm { font-size: 17px; }
.grid { display: grid; grid-template-columns: 1fr 300px; gap: 14px; align-items: start; }
.side { display: flex; flex-direction: column; gap: 14px; }
.xh { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.pp { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
.pick { display: flex; flex-direction: column; gap: 7px; padding: 6px 6px 9px; border-radius: 16px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; transition: border-color var(--t-fast), transform var(--t-fast); }
.pick:hover { border-color: var(--a2); transform: translateY(-1px); }
.pick-tile { aspect-ratio: 1; display: grid; place-items: center; border-radius: 12px; background: linear-gradient(180deg, #2a3a5c, #11131f); overflow: hidden; }
.pick-name { font-size: 12px; font-weight: 700; color: var(--text-2); text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.link { display: flex; align-items: center; gap: 12px; height: 46px; padding: 0 14px; border-radius: 14px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; font-weight: 650; text-align: left; }
.link:hover { border-color: rgba(255, 255, 255, 0.2); }
</style>
