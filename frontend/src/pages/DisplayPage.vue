<script setup lang="ts">
import { computed } from 'vue'
import { TriangleAlert, Gamepad2, Check } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSlider from '@/components/RSlider.vue'
import Segmented from '@/components/Segmented.vue'
import BeforeAfter from '@/components/BeforeAfter.vue'
import CurveChart from '@/components/CurveChart.vue'
import PresetThumb from '@/components/PresetThumb.vue'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { SCENES } from '@/lib/scenes'
import { applyPreset, activePreset, currentDisplay, currentParams, fail, lookName, presetById, setParam, state } from '@/lib/store'
import type { ColorParams } from '@/lib/types'

const on = computed({ get: () => state.cfg.tuning.enabled, set: (v: boolean) => { state.cfg.tuning.enabled = v } })
const disp = computed(() => currentDisplay.value)
const p = computed(() => currentParams.value)
const subtitle = computed(() => {
  if (!on.value) return state.firstRun ? t('display.welcomeText') : t('display.off')
  return disp.value ? t('display.onDisplay', { preset: lookName(disp.value.id), display: t('common.display', { n: disp.value.number }) }) : ''
})

const tab = computed({ get: () => state.cfg.ui.displayTab, set: (v: 'tone' | 'color') => { state.cfg.ui.displayTab = v } })
const scene = computed({ get: () => state.cfg.ui.scene, set: (v: string) => { state.cfg.ui.scene = v } })
const mode = computed({ get: () => state.cfg.ui.previewMode, set: (v: 'demo' | 'curve') => { state.cfg.ui.previewMode = v } })

const pct = (v: number) => `${Math.round(v)}%`
const deg = (v: number) => `${Math.round(v)}°`

const vibBackend = computed(() => {
  const id = disp.value?.id ?? ''
  const live = state.status.display.vibrance?.[id]
  if (live === 'nvidia' || live === 'amd') return live
  if (live === 'matrix') return ''
  return disp.value?.vibrance ?? ''
})
const vibChip = computed(() => (vibBackend.value ? t('display.gpu') : t('display.windowsEngine')))
const vibTitle = computed(() => (vibBackend.value ? t('display.vibGpu', { gpu: vibBackend.value.toUpperCase() }) : t('display.vibMatrix')))

type SliderDef = { key: keyof ColorParams; min: number; max: number; step: number; def: number; fmt: (v: number) => string; chip?: string; chipTitle?: string; chipKind?: 'accent' | '' }
const colorSliders = computed<SliderDef[]>(() => [
  { key: 'hue', min: -180, max: 180, step: 1, def: 0, fmt: deg, chip: t('display.allScreens'), chipTitle: t('display.noteSat') },
  { key: 'saturation', min: 0, max: 300, step: 1, def: 100, fmt: pct, chip: t('display.allScreens'), chipTitle: t('display.noteSat') },
  { key: 'vibrance', min: 0, max: 200, step: 1, def: 100, fmt: pct, chip: vibChip.value, chipTitle: vibTitle.value, chipKind: vibBackend.value ? 'accent' : '' },
  { key: 'red', min: 0, max: 200, step: 1, def: 100, fmt: pct },
  { key: 'green', min: 0, max: 200, step: 1, def: 100, fmt: pct },
  { key: 'blue', min: 0, max: 200, step: 1, def: 100, fmt: pct },
])
const toneSliders: SliderDef[] = [
  { key: 'brightness', min: 50, max: 150, step: 1, def: 100, fmt: pct },
  { key: 'contrast', min: 50, max: 150, step: 1, def: 100, fmt: pct },
  { key: 'gamma', min: 0.4, max: 2.5, step: 0.01, def: 1, fmt: (v) => v.toFixed(2) },
  { key: 'temperature', min: 3500, max: 10000, step: 50, def: 6500, fmt: (v) => `${Math.round(v)} K` },
]
const sliders = computed(() => (tab.value === 'color' ? colorSliders.value : toneSliders))

const pinned = computed(() => state.cfg.ui.pinned.map((id) => presetById(id)).filter((x) => !!x).slice(0, 4))
const activeId = computed(() => (disp.value ? activePreset(disp.value.id)?.id : undefined))

function resetNeutral() {
  const stock = presetById('stock')
  if (stock) applyPreset(stock)
}
function pick(id: string) {
  const preset = presetById(id)
  if (!preset) return
  applyPreset(preset)
  on.value = true
}

const startWithWindows = computed({
  get: () => state.cfg.settings.startWithWindows,
  set: (v: boolean) => { state.cfg.settings.startWithWindows = v },
})

const gammaInfo = computed(() => {
  const id = disp.value?.id ?? ''
  const g = state.status.display.gamma?.[id] ?? Object.entries(state.status.display.gamma || {}).find(([k]) => k.toLowerCase() === id.toLowerCase())?.[1]
  return g
})
async function unlock() {
  try { await api.enableExtendedGamma() } catch (e) { fail(e) }
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('display.title')">
      <template #sub>
        <span class="pill-on" :class="{ off: !on }">{{ on ? t('common.on') : t('common.off') }}</span>
        <span class="sub-text">{{ subtitle }}</span>
      </template>
      <div class="dtabs">
        <button v-for="d in state.displays" :key="d.id" class="dtab" :class="{ active: d.id === disp?.id }" :title="`${d.monitor} · ${d.width}×${d.height}`" @click="state.selected = d.id">
          {{ t('common.display', { n: d.number }) }}
        </button>
      </div>
      <RSwitch v-model="on" size="lg" :label="t('display.title')" />
    </PageHeader>

    <div class="grid">
      <section class="adjust glass">
        <div class="row head">
          <h2 class="card-title">{{ t('display.adjust') }}</h2>
          <div class="spacer" />
          <Segmented v-model="tab" :options="[{ value: 'tone', label: t('display.tone') }, { value: 'color', label: t('display.color') }]" />
        </div>
        <p class="dim small hint">{{ t('display.dblclick') }}</p>
        <div class="sliders">
          <RSlider
            v-for="s in sliders" :key="s.key" :label="t('display.' + s.key)" :model-value="p[s.key]" :min="s.min" :max="s.max" :step="s.step"
            :def="s.def" :format="s.fmt" :chip="s.chip" :chip-title="s.chipTitle" :chip-kind="s.chipKind ?? ''"
            @update:model-value="(v: number) => setParam(s.key, v)"
          />
        </div>
        <div class="spacer" />
        <label class="check">
          <input v-model="startWithWindows" type="checkbox" />
          <span class="box"><Check :size="14" :stroke-width="3" /></span>
          {{ t('display.startWithWindows') }}
        </label>
        <p class="dim small">{{ t('display.noteSat') }}</p>
        <button class="btn block reset" @click="resetNeutral">{{ t('display.resetNeutral') }}</button>
      </section>

      <div class="right">
        <section class="preview glass">
          <div class="row head">
            <h2 class="card-title">{{ t('display.preview') }}</h2>
            <Segmented v-model="scene" size="sm" :options="[{ value: 'canyon', label: t('display.canyon') }, { value: 'neon', label: t('display.neon') }, { value: 'forest', label: t('display.forest') }]" />
            <div class="spacer" />
            <Segmented v-model="mode" size="sm" :options="[{ value: 'demo', label: t('display.demo') }, { value: 'curve', label: t('display.curve') }]" />
          </div>
          <BeforeAfter v-if="mode === 'demo'" :src="SCENES[scene] ?? SCENES.canyon" :params="p" uid="main" :active="on" />
          <CurveChart v-else :params="p" />
          <div v-if="on && gammaInfo && gammaInfo.limited" class="status warn">
            <TriangleAlert :size="16" />
            <span class="grow">{{ t('display.limited', { pct: Math.round(gammaInfo.strength * 100), display: t('common.display', { n: disp?.number }) }) }}</span>
            <button v-if="!state.status.extendedGamma" class="btn sm soft" @click="unlock">{{ t('display.unlock') }}</button>
          </div>
          <div v-else-if="on && gammaInfo && !gammaInfo.ok" class="status warn">
            <TriangleAlert :size="16" /><span class="grow">{{ t('display.gammaError', { display: t('common.display', { n: disp?.number }) }) }}</span>
          </div>
          <div v-else-if="on && state.status.display.matrix === 'error'" class="status warn">
            <TriangleAlert :size="16" /><span class="grow">{{ t('display.matrixError', { error: state.status.display.matrixError }) }}</span>
          </div>
          <div v-else-if="state.status.rule" class="status">
            <Gamepad2 :size="16" /><span class="grow">{{ t('display.rule', { exe: state.status.rule }) }}</span>
          </div>
          <p v-else class="dim small">{{ mode === 'demo' ? t('display.dragHint') : t('display.curveHint') }}</p>
        </section>

        <section class="presets glass">
          <div class="row head">
            <h2 class="card-title sm">{{ t('display.colorPresets') }}</h2>
            <div class="spacer" />
            <button class="btn sm" @click="state.page = 'presets'">{{ t('display.allPresets') }}</button>
          </div>
          <div class="thumbs">
            <button v-for="pr in pinned" :key="pr!.id" class="thumb" :class="{ active: on && activeId === pr!.id }" @click="pick(pr!.id)">
              <PresetThumb :params="pr!.params" :uid="'pin-' + pr!.id" :scene="scene" />
              <span class="tname">{{ pr!.name }}</span>
            </button>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; height: 100%; min-height: 640px; }
.sub-text { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; font-size: 13.5px; line-height: 1.35; }
.dtabs { display: flex; gap: 8px; margin-right: 18px; }
.dtab { height: 40px; padding: 0 18px; border-radius: 999px; border: 1px solid var(--field-border); background: rgba(255, 255, 255, 0.05); font-weight: 700; font-size: 14px; cursor: pointer; color: var(--text-2); }
.dtab:hover { color: var(--text); background: rgba(255, 255, 255, 0.09); }
.dtab.active { color: #fff; border-color: transparent; background: linear-gradient(135deg, var(--a2), var(--a3)); box-shadow: 0 8px 20px -8px var(--a-glow); }

.status { display: flex; align-items: center; gap: 10px; min-height: 22px; font-size: 13px; color: var(--text-2); }
.status.warn > svg { color: var(--warn); flex: none; }
.status > svg { color: var(--a1); flex: none; }

.grid { flex: 1; display: grid; grid-template-columns: minmax(330px, 0.68fr) 1fr; gap: 14px; min-height: 0; }
.adjust { display: flex; flex-direction: column; gap: 10px; padding: 18px 22px 18px; }
.head { gap: 10px; }
.hint { margin-top: -6px; }
.sliders { display: flex; flex-direction: column; gap: 8px; margin-top: 2px; }
.reset { margin-top: 2px; height: 44px; }

.check { display: inline-flex; align-items: center; gap: 12px; font-weight: 650; cursor: pointer; }
.check input { position: absolute; opacity: 0; pointer-events: none; }
.check .box { width: 26px; height: 26px; border-radius: 8px; display: grid; place-items: center; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.25); color: transparent; transition: all var(--t-fast); }
.check input:checked + .box { background: linear-gradient(135deg, var(--a2), var(--a3)); border-color: transparent; color: #fff; }
.check input:focus-visible + .box { outline: 2px solid var(--a2); outline-offset: 2px; }

.right { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
.preview { display: flex; flex-direction: column; gap: 12px; padding: 18px 20px 16px; }
.presets { padding: 16px 20px 18px; display: flex; flex-direction: column; gap: 12px; }
.card-title.sm { font-size: 17px; }
.thumbs { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.thumb { display: flex; flex-direction: column; gap: 7px; padding: 5px 5px 8px; border-radius: 16px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; transition: border-color var(--t-fast), box-shadow var(--t-fast), transform var(--t-fast); }
.thumb:hover { border-color: rgba(255, 255, 255, 0.2); transform: translateY(-1px); }
.thumb.active { border-color: var(--a2); box-shadow: 0 0 0 1px var(--a2), 0 10px 24px -10px var(--a-glow); }
.tname { font-size: 12.5px; font-weight: 700; text-align: center; color: var(--text-2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.thumb.active .tname { color: var(--text); }
</style>
