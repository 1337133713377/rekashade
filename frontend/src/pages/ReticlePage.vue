<script setup lang="ts">
import { computed, ref } from 'vue'
import { ImageUp, Crosshair as CrossIcon, Eye, EyeOff, Keyboard } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSlider from '@/components/RSlider.vue'
import RSelect from '@/components/RSelect.vue'
import SettingRow from '@/components/SettingRow.vue'
import OverlayPreview from '@/components/OverlayPreview.vue'
import { RETICLES, SWATCHES, importImage, reticleSvg } from '@/lib/crosshair'
import { t } from '@/lib/i18n'
import { bindLabel } from '@/lib/keys'
import { fail, state } from '@/lib/store'

const r = computed(() => state.cfg.reticle)
const ov = computed(() => state.cfg.overlay)
const enabled = computed({
  get: () => ov.value.enabled && ov.value.mode === 'reticle',
  set: (v: boolean) => {
    if (v) state.cfg.overlay.mode = 'reticle'
    state.cfg.overlay.enabled = v
  },
})
const svgUrl = (id: string) => 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(reticleSvg(id, r.value.color))
function pick(id: string) {
  state.cfg.reticle.source = 'library'
  state.cfg.reticle.libraryId = id
}
const file = ref<HTMLInputElement>()
async function onFile(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (!f) return
  try {
    state.cfg.reticle.custom = await importImage(f)
    state.cfg.reticle.source = 'custom'
  } catch (err) { fail(err) } finally { (e.target as HTMLInputElement).value = '' }
}

const screenOptions = computed(() => [
  { value: '__primary', label: t('common.default') + ' — ' + (state.displays.find((d) => d.primary) ? t('common.display', { n: state.displays.find((d) => d.primary)!.number }) : '') },
  ...state.displays.map((d) => ({ value: d.id, label: `${t('common.display', { n: d.number })} · ${d.width}×${d.height}` })),
])
const screen = computed({ get: () => ov.value.display || '__primary', set: (v: string) => { state.cfg.overlay.display = v === '__primary' ? '' : v } })
const hold = computed(() => state.cfg.binds.find((b) => b.action === 'holdHide'))
const reasonText = computed(() => {
  const s = state.status.overlay
  if (s.visible) return t('ret.visible')
  return { off: t('ret.reasonOff'), hold: t('ret.reasonHold'), notfullscreen: t('ret.reasonFs'), noimage: t('ret.reasonImg'), '': '' }[s.reason] ?? ''
})
</script>

<template>
  <div class="page">
    <PageHeader :title="t('ret.title')" :subtitle="t('ret.subtitle')">
      <span class="chip" :class="state.status.overlay.visible ? 'ok' : ''">
        <component :is="state.status.overlay.visible ? Eye : EyeOff" :size="14" /> {{ reasonText }}
      </span>
      <RSwitch v-model="enabled" size="lg" :label="t('ret.title')" />
    </PageHeader>

    <div class="grid">
      <section class="glass card">
        <h2 class="card-title">{{ t('ret.library') }}</h2>
        <div class="lib">
          <button v-for="x in RETICLES" :key="x.id" class="tile" :class="{ active: r.source === 'library' && r.libraryId === x.id }" @click="pick(x.id)">
            <img :src="svgUrl(x.id)" alt="" draggable="false" />
            <span>{{ t('ret.names.' + x.id) }}</span>
          </button>
          <button class="tile upload" :class="{ active: r.source === 'custom' }" @click="r.custom ? (state.cfg.reticle.source = 'custom') : file?.click()">
            <img v-if="r.custom" :src="r.custom" alt="" draggable="false" />
            <ImageUp v-else :size="26" />
            <span>{{ r.custom ? t('ret.yours') : t('ret.upload') }}</span>
          </button>
        </div>
        <div class="row">
          <button class="btn sm" @click="file?.click()"><ImageUp :size="15" /> {{ t('ret.upload') }}</button>
          <span class="dim tiny">{{ t('ret.uploadHint') }}</span>
          <input ref="file" type="file" accept="image/png,image/svg+xml,image/webp,image/gif" class="hidden" @change="onFile" />
        </div>
        <OverlayPreview :reticle="r" :label="t('xh.actualSize')" scene="neon" />
        <div class="swatches" :class="{ off: r.source === 'custom' }">
          <span class="label-caps">{{ t('ret.color') }}</span>
          <button v-for="c in SWATCHES" :key="c" class="sw" :class="{ active: r.color.toUpperCase() === c }" :style="{ background: c }" :aria-label="c" @click="state.cfg.reticle.color = c" />
        </div>
        <RSlider v-model="state.cfg.reticle.size" :label="t('ret.size')" :min="12" :max="300" :def="56" :format="(v: number) => `${v} px`" />
        <div class="row two">
          <RSlider v-model="state.cfg.reticle.opacity" class="grow" :label="t('ret.opacity')" :min="10" :max="100" :def="100" :format="(v: number) => `${v}%`" />
          <RSlider v-model="state.cfg.reticle.rotation" class="grow" :label="t('ret.rotation')" :min="-180" :max="180" :def="0" :format="(v: number) => `${v}°`" />
        </div>
        <label class="row small toggle-line"><RSwitch v-model="state.cfg.reticle.glow" size="sm" :label="t('ret.glow')" /> {{ t('ret.glow') }}</label>
      </section>

      <section class="glass card">
        <h2 class="card-title">{{ t('ret.placement') }}</h2>
        <div class="field">
          <span class="label-caps">{{ t('ret.screen') }}</span>
          <RSelect v-model="screen" :options="screenOptions" />
        </div>
        <div class="row between"><span class="label-caps">{{ t('ret.offset') }}</span><button class="btn sm ghost" @click="state.cfg.overlay.offsetX = 0; state.cfg.overlay.offsetY = 0"><CrossIcon :size="14" /> {{ t('ret.center') }}</button></div>
        <RSlider v-model="state.cfg.overlay.offsetX" :label="t('ret.offsetX')" :min="-200" :max="200" :def="0" :format="(v: number) => `${v > 0 ? '+' : ''}${v} px`" />
        <RSlider v-model="state.cfg.overlay.offsetY" :label="t('ret.offsetY')" :min="-200" :max="200" :def="0" :format="(v: number) => `${v > 0 ? '+' : ''}${v} px`" />
        <div class="label-caps sp">{{ t('ret.behavior') }}</div>
        <div class="rows">
          <SettingRow :title="t('ret.follow')" :hint="t('ret.followHint')"><RSwitch v-model="state.cfg.overlay.followWindow" :label="t('ret.follow')" /></SettingRow>
          <SettingRow :title="t('ret.onlyFs')" :hint="t('ret.onlyFsHint')"><RSwitch v-model="state.cfg.overlay.onlyFullscreen" :label="t('ret.onlyFs')" /></SettingRow>
          <SettingRow :title="t('ret.capture')" :hint="t('ret.captureHint')"><RSwitch v-model="state.cfg.overlay.hideFromCapture" :label="t('ret.capture')" /></SettingRow>
          <SettingRow :title="t('ret.hold')" :hint="hold && hold.key ? '' : t('ret.holdNone')">
            <button class="btn sm" @click="state.page = 'binds'"><Keyboard :size="15" /> {{ hold && hold.key ? bindLabel(hold) : t('binds.change') }}</button>
          </SettingRow>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; min-height: 100%; }
.grid { flex: 1; display: grid; grid-template-columns: 1.15fr 1fr; gap: 14px; }
.card { padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; }
.lib { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
.tile { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 5px; aspect-ratio: 1; padding: 6px; border-radius: 14px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.25); cursor: pointer; color: var(--text-2); font-size: 10.5px; font-weight: 700; overflow: hidden; }
.tile img { width: 34px; height: 34px; object-fit: contain; }
.tile span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%; }
.tile:hover { border-color: rgba(255, 255, 255, 0.2); }
.tile.active { border-color: var(--a1); box-shadow: 0 0 0 1px var(--a2), 0 0 24px -8px var(--a-glow); color: var(--text); }
.upload { border-style: dashed; }
.hidden { position: absolute; width: 0; height: 0; opacity: 0; pointer-events: none; }
.swatches { display: flex; align-items: center; gap: 9px; flex-wrap: wrap; }
.swatches.off { opacity: 0.4; pointer-events: none; }
.sw { width: 26px; height: 26px; border-radius: 50%; border: 2px solid rgba(255, 255, 255, 0.12); cursor: pointer; padding: 0; }
.sw.active { box-shadow: 0 0 0 2px var(--bg), 0 0 0 4px rgba(255, 255, 255, 0.9); }
.two { gap: 16px; align-items: flex-start; }
.toggle-line { gap: 12px; cursor: pointer; color: var(--text-2); }
.field { display: flex; flex-direction: column; gap: 8px; }
.field :deep(.sel) { width: 100%; }
.between { justify-content: space-between; }
.sp { margin-top: 4px; }
.rows { display: flex; flex-direction: column; }
</style>
