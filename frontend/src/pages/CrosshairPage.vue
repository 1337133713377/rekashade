<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, Copy, ClipboardPaste, Save } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSlider from '@/components/RSlider.vue'
import CrosshairTile from '@/components/CrosshairTile.vue'
import OverlayPreview from '@/components/OverlayPreview.vue'
import { DESIGNS, SWATCHES } from '@/lib/crosshair'
import { decodeCrosshair, encodeCrosshair } from '@/lib/codes'
import { t } from '@/lib/i18n'
import { ask, copyText, readClipboard, state, toast } from '@/lib/store'
import type { CrosshairCfg } from '@/lib/types'

const xh = computed(() => state.cfg.crosshair)
const enabled = computed({
  get: () => state.cfg.overlay.enabled && state.cfg.overlay.mode === 'crosshair',
  set: (v: boolean) => {
    if (v) state.cfg.overlay.mode = 'crosshair'
    state.cfg.overlay.enabled = v
  },
})
const reticleActive = computed(() => state.cfg.overlay.enabled && state.cfg.overlay.mode === 'reticle')

function pickDesign(id: string) {
  const d = DESIGNS.find((x) => x.id === id)
  if (!d) return
  Object.assign(state.cfg.crosshair, d.shape, { design: id })
}
const tileCfg = (id: string): CrosshairCfg => {
  const d = DESIGNS.find((x) => x.id === id)!
  return { ...xh.value, ...d.shape, opacity: 100, saved: [] }
}
const palette = computed(() => [...SWATCHES, ...xh.value.customColors])
const colorInput = ref<HTMLInputElement>()
function addColor(e: Event) {
  const v = (e.target as HTMLInputElement).value.toUpperCase()
  if (!palette.value.includes(v)) state.cfg.crosshair.customColors = [...xh.value.customColors, v].slice(-4)
  state.cfg.crosshair.color = v
}
type Key = 'ring' | 'arm' | 'gap' | 'thickness' | 'dot' | 'outline' | 'opacity'
const sliders: { key: Key; min: number; max: number; def: number; unit: string }[] = [
  { key: 'ring', min: 0, max: 40, def: 13, unit: 'px' }, { key: 'arm', min: 0, max: 40, def: 21, unit: 'px' },
  { key: 'gap', min: 0, max: 20, def: 0, unit: 'px' }, { key: 'thickness', min: 1, max: 8, def: 3, unit: 'px' },
  { key: 'dot', min: 0, max: 12, def: 0, unit: 'px' }, { key: 'outline', min: 0, max: 4, def: 1, unit: 'px' },
  { key: 'opacity', min: 10, max: 100, def: 100, unit: '%' },
]
function setKey(k: Key, v: number) {
  state.cfg.crosshair[k] = v
  if (k !== 'opacity') state.cfg.crosshair.design = 'custom'
}

function saveCurrent() {
  const slots = state.cfg.crosshair.saved
  const i = slots.findIndex((s) => !s)
  if (i < 0) { toast('error', t('xh.slotsFull')); return }
  slots[i] = { name: `#${i + 1}`, code: encodeCrosshair(xh.value) }
  toast('success', t('xh.savedSlot', { n: i + 1 }))
}
function loadSlot(i: number) {
  const s = state.cfg.crosshair.saved[i]
  if (!s) return
  const d = decodeCrosshair(s.code)
  if (d) { Object.assign(state.cfg.crosshair, d); toast('success', t('xh.loaded')) }
}
function clearSlot(i: number) { state.cfg.crosshair.saved[i] = null }
const slotCfg = (code: string): CrosshairCfg => ({ ...xh.value, ...(decodeCrosshair(code) ?? {}), saved: [] })

async function paste() {
  const clip = await readClipboard()
  const code = await ask({ title: t('xh.pasteCode'), placeholder: t('xh.pastePrompt'), value: clip.startsWith('RKX1-') ? clip : '', confirmLabel: t('common.import') })
  if (!code) return
  const d = decodeCrosshair(code)
  if (!d) { toast('error', t('xh.invalid')); return }
  Object.assign(state.cfg.crosshair, d)
  toast('success', t('xh.loaded'))
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('xh.title')" :subtitle="reticleActive ? t('xh.reticleActive') : t('xh.subtitle')">
      <RSwitch v-model="enabled" size="lg" :label="t('xh.title')" />
    </PageHeader>

    <div class="grid">
      <section class="glass card">
        <h2 class="card-title">{{ t('xh.designs') }}</h2>
        <div class="designs">
          <button v-for="d in DESIGNS" :key="d.id" class="tile" :class="{ active: xh.design === d.id }" :title="d.id" @click="pickDesign(d.id)">
            <CrosshairTile :cfg="tileCfg(d.id)" />
          </button>
        </div>

        <div class="label-caps">{{ t('xh.color') }}</div>
        <div class="swatches">
          <button v-for="c in palette" :key="c" class="sw" :class="{ active: xh.color.toUpperCase() === c.toUpperCase() }" :style="{ background: c }" :aria-label="c" @click="state.cfg.crosshair.color = c" />
          <button class="sw add" :aria-label="t('xh.pickColor')" @click="colorInput?.click()"><Plus :size="18" /></button>
          <input ref="colorInput" type="color" class="hidden" :value="xh.color" @change="addColor" />
          <div class="spacer" />
          <span class="label-caps">{{ t('xh.outlineColor') }}</span>
          <button v-for="c in ['#000000', '#FFFFFF']" :key="'o' + c" class="sw small" :class="{ active: xh.outlineColor.toUpperCase() === c }" :style="{ background: c }" :aria-label="c" @click="state.cfg.crosshair.outlineColor = c" />
        </div>

        <div class="label-caps">{{ t('xh.my') }}</div>
        <div class="row">
          <button class="btn primary" @click="saveCurrent"><Save :size="16" /> {{ t('xh.saveCurrent') }}</button>
          <button class="btn" @click="copyText(encodeCrosshair(xh))"><Copy :size="16" /> {{ t('xh.copyCode') }}</button>
          <button class="btn" @click="paste"><ClipboardPaste :size="16" /> {{ t('xh.pasteCode') }}</button>
        </div>
        <div class="slots">
          <button v-for="(s, i) in xh.saved" :key="i" class="slot" :class="{ filled: !!s }" :title="s ? t('xh.slotHint') : t('xh.empty')" @click="loadSlot(i)" @contextmenu.prevent="clearSlot(i)">
            <CrosshairTile v-if="s" :cfg="slotCfg(s.code)" />
          </button>
        </div>
        <p class="dim small">{{ t('xh.codesHint') }}</p>
      </section>

      <section class="glass card">
        <h2 class="card-title">{{ t('xh.customize') }}</h2>
        <OverlayPreview :crosshair="xh" :label="t('xh.actualSize')" :scene="state.cfg.ui.scene" />
        <div class="sliders">
          <RSlider v-for="s in sliders" :key="s.key" :label="t('xh.' + s.key)" :model-value="xh[s.key]" :min="s.min" :max="s.max" :def="s.def"
            :format="(v: number) => `${v} ${s.unit}`" @update:model-value="(v: number) => setKey(s.key, v)" />
        </div>
        <div class="row toggles">
          <label class="row small"><RSwitch v-model="state.cfg.crosshair.rotate" size="sm" :label="t('xh.rotate')" /> {{ t('xh.rotate') }}</label>
          <label class="row small"><RSwitch v-model="state.cfg.crosshair.tStyle" size="sm" :label="t('xh.tStyle')" /> {{ t('xh.tStyle') }}</label>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; min-height: 100%; }
.grid { flex: 1; display: grid; grid-template-columns: 1.35fr 1fr; gap: 14px; }
.card { padding: 16px 20px; display: flex; flex-direction: column; gap: 11px; }
.designs { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.tile { aspect-ratio: 1; display: grid; place-items: center; border-radius: 16px; border: 1px solid var(--field-border); background: linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(0, 0, 0, 0.25)); cursor: pointer; overflow: hidden; transition: all var(--t-fast); }
.tile:hover { border-color: rgba(255, 255, 255, 0.2); }
.tile.active { border-color: var(--a1); box-shadow: 0 0 0 1px var(--a2), 0 0 24px -6px var(--a-glow); }
.swatches { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.sw { width: 34px; height: 34px; border-radius: 50%; border: 2px solid rgba(255, 255, 255, 0.12); cursor: pointer; padding: 0; transition: transform var(--t-fast); }
.sw:hover { transform: scale(1.06); }
.sw.active { box-shadow: 0 0 0 3px var(--bg), 0 0 0 5px rgba(255, 255, 255, 0.9); }
.sw.small { width: 24px; height: 24px; }
.sw.add { display: grid; place-items: center; background: transparent; color: var(--text-2); border: 2px dashed rgba(255, 255, 255, 0.25); }
.hidden { position: absolute; width: 0; height: 0; opacity: 0; pointer-events: none; }
.slots { display: grid; grid-template-columns: repeat(8, 1fr); gap: 8px; }
.slot { aspect-ratio: 1; border-radius: 12px; border: 1.5px dashed rgba(255, 255, 255, 0.14); background: rgba(0, 0, 0, 0.15); cursor: pointer; display: grid; place-items: center; overflow: hidden; }
.slot.filled { border-style: solid; border-color: var(--field-border); background: rgba(0, 0, 0, 0.3); }
.slot.filled:hover { border-color: var(--a2); }
.sliders { display: flex; flex-direction: column; gap: 4px; }
.toggles { gap: 22px; }
.toggles label { gap: 10px; cursor: pointer; color: var(--text-2); }
</style>
