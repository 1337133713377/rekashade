<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Maximize2, RotateCcw, Plus, X, TriangleAlert, Check, Keyboard } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSelect from '@/components/RSelect.vue'
import Segmented from '@/components/Segmented.vue'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { bindLabel } from '@/lib/keys'
import { SCENES } from '@/lib/scenes'
import { fail, state } from '@/lib/store'
import type { DisplayInfo, Scaling } from '@/lib/types'

const PRESETS = [
  { w: 1440, h: 1080, r: '4:3' }, { w: 1280, h: 960, r: '4:3' }, { w: 1024, h: 768, r: '4:3' },
  { w: 1280, h: 1024, r: '5:4' }, { w: 1680, h: 1050, r: '16:10' }, { w: 1600, h: 900, r: '16:9' },
]

const res = computed(() => state.cfg.resolution)
const disp = computed<DisplayInfo | undefined>(() =>
  state.displays.find((d) => d.id.toLowerCase() === res.value.display.toLowerCase()) ?? state.displays.find((d) => d.primary) ?? state.displays[0])
const displayOptions = computed(() => state.displays.map((d) => ({ value: d.id, label: `${t('common.display', { n: d.number })} · ${d.monitor || d.adapter}` })))
const displayId = computed({ get: () => disp.value?.id ?? '', set: (v: string) => { state.cfg.resolution.display = v } })

const isCustom = ref(false)
const cw = ref(res.value.width)
const ch = ref(res.value.height)
const sel = computed(() => (isCustom.value ? { w: cw.value, h: ch.value } : { w: res.value.width, h: res.value.height }))

function choose(w: number, h: number) {
  isCustom.value = false
  state.cfg.resolution.width = w
  state.cfg.resolution.height = h
  state.cfg.resolution.refresh = 0
}
function useCustom() {
  isCustom.value = true
  cw.value = res.value.width
  ch.value = res.value.height
}
watch([cw, ch], ([w, h]) => {
  if (!isCustom.value) return
  if (w >= 640 && h >= 480 && w <= 16384 && h <= 16384) { state.cfg.resolution.width = Math.round(w); state.cfg.resolution.height = Math.round(h) }
})
onMounted(() => { isCustom.value = !PRESETS.some((p) => p.w === res.value.width && p.h === res.value.height) })

const mode = computed(() => disp.value?.modes.find((m) => m.width === sel.value.w && m.height === sel.value.h))
const listed = computed(() => !!mode.value)
const refreshOptions = computed(() => [{ value: '0', label: t('res.auto') }, ...(mode.value?.rates ?? []).map((r) => ({ value: String(r), label: `${r} Hz` }))])
const refresh = computed({ get: () => String(res.value.refresh || 0), set: (v: string) => { state.cfg.resolution.refresh = Number(v) } })
const scaling = computed({ get: () => res.value.scaling, set: (v: Scaling) => { state.cfg.resolution.scaling = v } })

const st = computed(() => state.status.stretch)
const nativeW = computed(() => disp.value?.nativeWidth || disp.value?.width || 1920)
const nativeH = computed(() => disp.value?.nativeHeight || disp.value?.height || 1080)
const busy = ref(false)

async function test() {
  if (!disp.value) return
  busy.value = true
  try {
    await api.testResolution({ display: disp.value.id, width: sel.value.w, height: sel.value.h, refresh: res.value.refresh, scaling: res.value.scaling })
  } catch (e) { fail(e) } finally { busy.value = false }
}
async function revert() {
  busy.value = true
  try { await api.revertResolution() } catch (e) { fail(e) } finally { busy.value = false }
}

// preview geometry (percentages inside a box with the panel's aspect ratio)
const frame = computed(() => {
  const pw = nativeW.value, ph = nativeH.value, gw = sel.value.w, gh = sel.value.h
  if (res.value.scaling === 'stretch') return { w: 100, h: 100, sx: (pw / ph) / (gw / gh) }
  if (res.value.scaling === 'center') return { w: Math.min(100, (gw / pw) * 100), h: Math.min(100, (gh / ph) * 100), sx: 1 }
  const s = Math.min(pw / gw, ph / gh)
  return { w: ((gw * s) / pw) * 100, h: ((gh * s) / ph) * 100, sx: 1 }
})

// auto stretch
const running = ref<string[]>([])
const typed = ref('')
onMounted(async () => { try { running.value = await api.runningApps() } catch { running.value = [] } })
const runningOptions = computed(() => running.value.filter((e) => !res.value.autoGames.includes(e)).map((e) => ({ value: e, label: e })))
const pick = ref('')
function addGame(exe: string) {
  let e = exe.trim().toLowerCase().split(/[\\/]/).pop() ?? ''
  if (e && !e.endsWith('.exe')) e += '.exe'
  if (e && !res.value.autoGames.includes(e)) state.cfg.resolution.autoGames.push(e)
  typed.value = ''
  pick.value = ''
}
const hotkey = computed(() => {
  const b = state.cfg.binds.find((x) => x.action === 'toggleStretch')
  return b && b.key ? bindLabel(b) : ''
})
</script>

<template>
  <div class="page">
    <PageHeader :title="t('res.title')" :subtitle="t('res.subtitle')">
      <span v-if="st.active" class="chip accent big"><Maximize2 :size="14" /> {{ t('res.stretched', { w: st.width, h: st.height }) }}</span>
      <span v-else class="chip big">{{ t('res.native', { w: disp?.width ?? 0, h: disp?.height ?? 0 }) }}</span>
      <button v-if="st.active" class="btn" :disabled="busy" @click="revert"><RotateCcw :size="16" /> {{ t('res.revert') }}</button>
    </PageHeader>

    <div class="grid">
      <section class="glass card">
        <h2 class="card-title">{{ t('res.choose') }}</h2>
        <div v-if="state.displays.length > 1" class="field">
          <span class="label-caps">{{ t('res.display') }}</span>
          <RSelect v-model="displayId" :options="displayOptions" />
        </div>
        <div class="field">
          <span class="label-caps">{{ t('res.popular') }}</span>
          <div class="sizes">
            <button v-for="p in PRESETS" :key="p.w + 'x' + p.h" class="size" :class="{ active: !isCustom && res.width === p.w && res.height === p.h }" @click="choose(p.w, p.h)">
              <span class="mono strong">{{ p.w }}×{{ p.h }}</span>
              <span class="ratio">{{ p.r }}</span>
              <span v-if="disp && !disp.modes.some((m) => m.width === p.w && m.height === p.h)" class="nl" :title="t('res.notListed')"><TriangleAlert :size="12" /></span>
            </button>
            <button class="size" :class="{ active: isCustom }" @click="useCustom">
              <span class="strong">{{ t('res.custom') }}</span><span class="ratio">W×H</span>
            </button>
          </div>
        </div>
        <div v-if="isCustom" class="row">
          <label class="mini"><span class="dim tiny">{{ t('res.width') }}</span><input v-model.number="cw" class="input mono" type="number" min="640" max="16384" /></label>
          <span class="x">×</span>
          <label class="mini"><span class="dim tiny">{{ t('res.height') }}</span><input v-model.number="ch" class="input mono" type="number" min="480" max="16384" /></label>
        </div>
        <div class="pair">
          <span class="label-caps">{{ t('res.refresh') }}</span>
          <RSelect v-model="refresh" :options="refreshOptions" />
          <span class="label-caps">{{ t('res.scaling') }}</span>
          <Segmented v-model="scaling" size="sm" :options="[{ value: 'stretch', label: t('res.stretch') }, { value: 'aspect', label: t('res.aspect') }, { value: 'center', label: t('res.center') }]" />
        </div>
        <p v-if="!listed" class="warnline small"><TriangleAlert :size="15" /> {{ t('res.notListed') }}</p>
        <div class="spacer" />
        <button class="btn primary lg block" :disabled="busy || !disp" @click="test"><Maximize2 :size="17" /> {{ t('res.test') }}</button>
        <p class="dim small center">{{ t('res.inGame') }}</p>
      </section>

      <div class="rightcol">
        <section class="glass card">
          <div class="row"><h2 class="card-title">{{ t('res.preview') }}</h2><div class="spacer" /><span v-if="res.confirmed" class="chip ok"><Check :size="13" /> {{ t('res.confirmed') }}</span></div>
          <div class="panel" :style="{ aspectRatio: `${nativeW} / ${nativeH}` }">
            <div class="game" :style="{ width: frame.w + '%', height: frame.h + '%' }">
              <div class="game-inner" :style="{ aspectRatio: `${sel.w} / ${sel.h}`, transform: `scaleX(${frame.sx})` }">
                <img :src="SCENES.canyon" alt="" draggable="false" />
              </div>
            </div>
            <span class="cap tl">{{ t('res.game', { w: sel.w, h: sel.h }) }}</span>
            <span class="cap br">{{ t('res.panel', { w: nativeW, h: nativeH }) }}</span>
          </div>
          <p v-if="hotkey" class="dim small row"><Keyboard :size="15" /> {{ t('res.hotkey', { key: hotkey }) }}</p>
        </section>

        <section class="glass card">
          <div class="row">
            <h2 class="card-title sm">{{ t('res.autoTitle') }}</h2><div class="spacer" />
            <RSwitch v-model="state.cfg.resolution.autoApply" :disabled="!res.confirmed" :label="t('res.autoTitle')" />
          </div>
          <p class="dim small">{{ res.confirmed ? t('res.autoHint') : t('res.needConfirm') }}</p>
          <div class="games">
            <span v-for="(g, i) in res.autoGames" :key="g" class="gchip mono">{{ g }}<button :aria-label="t('common.remove')" @click="state.cfg.resolution.autoGames.splice(i, 1)"><X :size="13" /></button></span>
            <span v-if="!res.autoGames.length" class="dim small">{{ t('res.noGames') }}</span>
          </div>
          <div class="row">
            <RSelect v-model="pick" class="grow" :options="runningOptions" :placeholder="t('presets.pickRunning')" @update:model-value="addGame" />
            <form class="row grow" @submit.prevent="addGame(typed)">
              <input v-model="typed" class="input" :placeholder="t('presets.typeExe')" spellcheck="false" />
              <button class="btn icon" type="submit" :aria-label="t('common.add')"><Plus :size="17" /></button>
            </form>
          </div>
          <label class="row small toggle-line">
            <RSwitch v-model="state.cfg.resolution.revertOnExit" size="sm" :label="t('res.revertOnExit')" />
            {{ t('res.revertOnExit') }}
          </label>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; min-height: 100%; }
.chip.big { height: 32px; padding: 0 14px; font-size: 12.5px; }
.grid { flex: 1; display: grid; grid-template-columns: 1.12fr 1fr; gap: 14px; }
.card { padding: 16px 20px; display: flex; flex-direction: column; gap: 12px; }
.card-title.sm { font-size: 17px; }
.field { display: flex; flex-direction: column; gap: 8px; }
.field :deep(.sel) { width: 100%; }
.sizes { display: grid; grid-template-columns: repeat(4, 1fr); gap: 7px; }
.size { position: relative; min-width: 0; display: flex; flex-direction: column; align-items: flex-start; gap: 1px; padding: 9px 10px; border-radius: 14px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.2); cursor: pointer; text-align: left; transition: all var(--t-fast); }
.size:hover { border-color: rgba(255, 255, 255, 0.18); }
.size.active { border-color: var(--a2); background: var(--a-soft); box-shadow: 0 0 0 1px var(--a2); }
.strong { font-weight: 750; font-size: 13px; }
.ratio { font-size: 11.5px; font-weight: 700; color: var(--text-3); }
.nl { position: absolute; top: 8px; right: 8px; color: var(--warn); }
.mini { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.x { padding-top: 18px; color: var(--text-3); font-weight: 700; }
.pair { display: grid; grid-template-columns: auto 1fr; align-items: center; gap: 10px 14px; }
.pair :deep(.sel) { width: 100%; }
.pair .seg { justify-self: start; }
.warnline { display: flex; gap: 8px; color: #fde68a; line-height: 1.45; }
.warnline :deep(svg) { flex: none; margin-top: 2px; }
.center { text-align: center; }
.rightcol { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
.panel { position: relative; width: 100%; border-radius: 14px; background: #000; display: grid; place-items: center; overflow: hidden; box-shadow: 0 0 0 6px rgba(255, 255, 255, 0.05), inset 0 0 0 1px rgba(255, 255, 255, 0.06); }
.game { display: grid; place-items: center; overflow: hidden; transition: width var(--t-med) var(--ease), height var(--t-med) var(--ease); }
.game-inner { height: 100%; transform-origin: center; transition: transform var(--t-med) var(--ease); }
.game-inner img { width: 100%; height: 100%; object-fit: cover; display: block; }
.cap { position: absolute; height: 24px; padding: 0 10px; display: inline-flex; align-items: center; border-radius: 999px; background: rgba(10, 8, 20, 0.78); font-size: 11.5px; font-weight: 700; }
.tl { left: 10px; top: 10px; } .br { right: 10px; bottom: 10px; color: var(--text-2); }
.games { display: flex; flex-wrap: wrap; gap: 8px; }
.gchip { display: inline-flex; align-items: center; gap: 6px; height: 30px; padding: 0 6px 0 12px; border-radius: 999px; background: var(--a-soft); color: var(--a-text); font-size: 12px; font-weight: 700; }
.gchip button { width: 20px; height: 20px; display: grid; place-items: center; border: 0; border-radius: 50%; background: rgba(0, 0, 0, 0.25); color: inherit; cursor: pointer; }
.grow :deep(.sel) { width: 100%; min-width: 0; }
.toggle-line { gap: 12px; cursor: pointer; color: var(--text-2); }
</style>
