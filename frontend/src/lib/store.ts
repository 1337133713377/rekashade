import { computed, reactive, watch } from 'vue'
import defaults from '@defaults'
import { api, errText } from './api'
import { lang, t } from './i18n'
import { neutralParams, sameParams } from './color'
import { defaultCrosshair, defaultReticle, renderCrosshair, renderReticle } from './crosshair'
import type {
  AppStatus, ColorParams, Config, CrosshairCfg, DisplayInfo, DisplayParams, DisplayTuning, Preset, ProfileCfg, ReticleCfg, ThemeCfg, ToastMsg, UiCfg,
} from './types'

interface DialogState {
  kind: 'prompt' | 'confirm'
  title: string
  body?: string
  placeholder?: string
  value: string
  confirmLabel?: string
  danger?: boolean
  resolve: (v: string | null) => void
}

interface State {
  ready: boolean
  version: string
  cfg: Config
  displays: DisplayInfo[]
  status: AppStatus
  selected: string
  firstRun: boolean
  updated: boolean
  mock: boolean
  systemLanguage: 'en' | 'fr'
  windowsBuild: number
  dataDir: string
  toasts: ToastMsg[]
  updateProgress: { done: number; total: number } | null
  dialog: DialogState | null
  page: string
}

const emptyStatus = (): AppStatus => ({
  display: { active: false, matrix: 'off', gamma: {}, vibrance: {}, gpu: [] },
  overlay: { visible: false, reason: 'off', hasImage: false, holdHidden: false },
  stretch: { active: false, auto: false, device: '', width: 0, height: 0, refresh: 0, pending: 0 },
  rule: '', update: null, updating: false, extendedGamma: false, startup: false, gpuProbeFailed: false,
})

export const state = reactive<State>({
  ready: false, version: '', cfg: {} as Config, displays: [], status: emptyStatus(), selected: '',
  firstRun: false, updated: false, mock: false, systemLanguage: 'en', windowsBuild: 0, dataDir: '',
  toasts: [], updateProgress: null, dialog: null, page: 'display',
})

// ---------------------------------------------------------------------------
// UI-owned defaults (the Go side keeps these sections verbatim)
// ---------------------------------------------------------------------------

function ensureUi(cfg: Config) {
  cfg.crosshair = { ...defaultCrosshair(), ...((cfg.crosshair || {}) as Partial<CrosshairCfg>) }
  const saved = Array.isArray(cfg.crosshair.saved) ? cfg.crosshair.saved.slice(0, 8) : []
  while (saved.length < 8) saved.push(null)
  cfg.crosshair.saved = saved
  if (!Array.isArray(cfg.crosshair.customColors)) cfg.crosshair.customColors = []
  cfg.reticle = { ...defaultReticle(), ...((cfg.reticle || {}) as Partial<ReticleCfg>) }
  cfg.theme = { accent: 'violet', backdrop: 'aurora', glass: true, motion: true, scale: 100, ...((cfg.theme || {}) as Partial<ThemeCfg>) }
  cfg.profile = { name: '', avatar: 0, since: new Date().toISOString(), ...((cfg.profile || {}) as Partial<ProfileCfg>) }
  cfg.ui = { page: 'display', displayTab: 'color', scene: 'canyon', previewMode: 'demo', pinned: ['stock', 'vivid-pop', 'golden-hour', 'ultra-vivid'], ...((cfg.ui || {}) as Partial<UiCfg>) }
  if (!Array.isArray(cfg.ui.pinned)) cfg.ui.pinned = []
  return cfg
}

// ---------------------------------------------------------------------------
// Toasts & dialogs
// ---------------------------------------------------------------------------

let toastId = 0
export function toast(kind: ToastMsg['kind'], text: string) {
  const id = ++toastId
  state.toasts.push({ id, kind, text })
  if (state.toasts.length > 4) state.toasts.shift()
  setTimeout(() => {
    const i = state.toasts.findIndex((x) => x.id === id)
    if (i >= 0) state.toasts.splice(i, 1)
  }, kind === 'error' ? 6000 : 3600)
}

export function fail(e: unknown) {
  toast('error', errText(e))
}

export function ask(opts: { title: string; placeholder?: string; value?: string; confirmLabel?: string }): Promise<string | null> {
  return new Promise((resolve) => {
    state.dialog = { kind: 'prompt', title: opts.title, placeholder: opts.placeholder, value: opts.value ?? '', confirmLabel: opts.confirmLabel, resolve }
  })
}

export function confirmAsk(opts: { title: string; body?: string; confirmLabel?: string; danger?: boolean }): Promise<boolean> {
  return new Promise((resolve) => {
    state.dialog = { kind: 'confirm', title: opts.title, body: opts.body, value: '', confirmLabel: opts.confirmLabel, danger: opts.danger, resolve: (v) => resolve(v !== null) }
  })
}

export async function copyText(text: string, message = t('toast.copied')) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    ta.remove()
  }
  toast('success', message)
}

export async function readClipboard(): Promise<string> {
  try {
    return (await navigator.clipboard.readText()) || ''
  } catch {
    return ''
  }
}

// ---------------------------------------------------------------------------
// Derived state
// ---------------------------------------------------------------------------

export const currentDisplay = computed<DisplayInfo | undefined>(
  () => state.displays.find((d) => d.id === state.selected) ?? state.displays.find((d) => d.primary) ?? state.displays[0],
)

export function displayLabel(id: string): string {
  const d = state.displays.find((x) => x.id.toLowerCase() === id.toLowerCase())
  return d ? t('common.display', { n: d.number }) : id
}

export function tuningFor(id: string): DisplayTuning {
  const key = Object.keys(state.cfg.tuning.displays).find((k) => k.toLowerCase() === id.toLowerCase())
  if (key) return state.cfg.tuning.displays[key]
  const { hue: _h, saturation: _s, ...rest } = neutralParams()
  return { presetId: 'stock', ...rest }
}

/** Ensures the selected display has an entry and returns it (reactive). */
export function editableTuning(id: string): DisplayTuning {
  const key = Object.keys(state.cfg.tuning.displays).find((k) => k.toLowerCase() === id.toLowerCase())
  if (key) return state.cfg.tuning.displays[key]
  state.cfg.tuning.displays[id] = { ...tuningFor(id) }
  return state.cfg.tuning.displays[id]
}

export function paramsFor(id: string): ColorParams {
  const d = tuningFor(id)
  return {
    hue: state.cfg.tuning.hue, saturation: state.cfg.tuning.saturation, vibrance: d.vibrance, brightness: d.brightness,
    contrast: d.contrast, gamma: d.gamma, temperature: d.temperature, red: d.red, green: d.green, blue: d.blue,
  }
}

export const currentParams = computed(() => {
  if (currentDisplay.value) return paramsFor(currentDisplay.value.id)
  // No display known (yet): hue and saturation are global, keep them live.
  return { ...neutralParams(), hue: state.cfg.tuning.hue, saturation: state.cfg.tuning.saturation }
})

export function presetById(id: string): Preset | undefined {
  return state.cfg.presets?.find((p) => p.id === id)
}

/** The preset a display is showing, or undefined when the look was tweaked. */
export function activePreset(id: string): Preset | undefined {
  const p = presetById(tuningFor(id).presetId)
  return p && sameParams(p.params, paramsFor(id)) ? p : undefined
}

export function lookName(id: string): string {
  return activePreset(id)?.name ?? t('nav.customLook')
}

// ---------------------------------------------------------------------------
// Mutations
// ---------------------------------------------------------------------------

export function setParam(key: keyof ColorParams, value: number) {
  if (key === 'hue' || key === 'saturation') {
    state.cfg.tuning[key] = value
    return
  }
  const d = currentDisplay.value
  if (!d) return
  editableTuning(d.id)[key as keyof DisplayParams] = value
}

export function applyPreset(p: Preset, displayId = currentDisplay.value?.id) {
  if (!displayId) return
  const { hue, saturation, ...rest } = p.params
  state.cfg.tuning.hue = hue
  state.cfg.tuning.saturation = saturation
  const tune = editableTuning(displayId)
  Object.assign(tune, rest, { presetId: p.id })
}

/** Replaces the settings with an imported file (UI defaults filled in). */
export function importConfig(cfg: Config) {
  state.cfg = ensureUi({ ...state.cfg, ...cfg })
}

/** Factory settings, keeping the profile, language and app settings. */
export function factoryReset() {
  const keep = { profile: state.cfg.profile, language: state.cfg.language, settings: state.cfg.settings, version: state.cfg.version }
  const cfg = {
    ...keep,
    tuning: { enabled: false, hue: 0, saturation: 100, displays: {} },
    presets: (defaults.presets as Preset[]).map((p) => ({ ...p, params: { ...p.params } })),
    binds: (defaults.binds as Config['binds']).map((b) => ({ ...b })),
    rules: [],
    overlay: { enabled: false, mode: 'crosshair', display: '', offsetX: 0, offsetY: 0, followWindow: false, onlyFullscreen: false, hideFromCapture: false },
    resolution: { display: '', width: 1440, height: 1080, refresh: 0, scaling: 'stretch', confirmed: false, revertOnExit: true, autoApply: false, autoGames: [] },
  } as unknown as Config
  state.cfg = ensureUi(cfg)
  state.page = 'account'
}

export function newId(prefix: string) {
  return `${prefix}-${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`
}

// ---------------------------------------------------------------------------
// Overlay bitmap (rendered here, shown natively)
// ---------------------------------------------------------------------------

let overlayTimer = 0
let overlayBusy = false
let overlayAgain = false

async function pushOverlay() {
  if (overlayBusy) { overlayAgain = true; return }
  overlayBusy = true
  try {
    const cfg = state.cfg
    const r = cfg.overlay.mode === 'reticle' ? await renderReticle(cfg.reticle) : renderCrosshair(cfg.crosshair)
    await api.setOverlayImage(r.canvas.toDataURL('image/png'), r.anchorX, r.anchorY)
  } catch (e) {
    console.warn('overlay render failed', e)
  } finally {
    overlayBusy = false
    if (overlayAgain) { overlayAgain = false; scheduleOverlay() }
  }
}

function scheduleOverlay() {
  clearTimeout(overlayTimer)
  overlayTimer = window.setTimeout(pushOverlay, 40)
}

// ---------------------------------------------------------------------------
// Theme & language
// ---------------------------------------------------------------------------

function applyTheme() {
  const th = state.cfg.theme
  const root = document.documentElement
  root.dataset.accent = th.accent
  root.dataset.backdrop = th.backdrop
  root.dataset.glass = th.glass ? 'on' : 'off'
  root.dataset.motion = th.motion ? 'on' : 'off'
  root.style.setProperty('--ui-zoom', String((th.scale || 100) / 100))
}

function applyLanguage() {
  const l = state.cfg.language
  lang.value = l === 'fr' || l === 'en' ? l : state.systemLanguage
  document.documentElement.lang = lang.value
}

// ---------------------------------------------------------------------------
// Boot
// ---------------------------------------------------------------------------

let lastSent = ''
let saveTimer = 0

function flushSave() {
  const json = JSON.stringify(state.cfg)
  if (json === lastSent) return
  lastSent = json
  api.saveConfig(JSON.parse(json)).catch(fail)
}

/** Asks the backend again for monitors if the first answer was empty. */
async function retryDisplays() {
  for (const delay of [500, 1500, 3000, 6000]) {
    await new Promise((r) => setTimeout(r, delay))
    if (state.displays.length) return
    try {
      const list = await api.refreshDisplays()
      if (list?.length) {
        state.displays = list
        if (!state.displays.some((d) => d.id === state.selected)) state.selected = (state.displays.find((d) => d.primary) ?? state.displays[0]).id
        return
      }
    } catch { /* keep trying */ }
  }
}

export async function boot() {
  const b = await api.bootstrap()
  state.version = b.version
  state.cfg = ensureUi(b.config)
  state.displays = b.displays || []
  state.status = b.status || emptyStatus()
  state.firstRun = b.firstRun
  state.updated = b.updated
  state.mock = b.mock
  state.systemLanguage = b.systemLanguage
  state.windowsBuild = b.windowsBuild
  state.dataDir = b.dataDir
  state.selected = (state.displays.find((d) => d.primary) ?? state.displays[0])?.id ?? ''
  state.page = state.cfg.ui.page || 'display'
  lastSent = JSON.stringify(state.cfg)
  applyLanguage()
  applyTheme()

  api.on('rs:config', (c) => {
    const cfg = ensureUi(c as Config)
    const json = JSON.stringify(cfg)
    if (json === lastSent) return
    lastSent = json
    state.cfg = cfg
  })
  api.on('rs:status', (s) => { state.status = s as AppStatus })
  api.on('rs:displays', (d) => {
    state.displays = (d as DisplayInfo[]) || []
    if (!state.displays.some((x) => x.id === state.selected)) state.selected = state.displays[0]?.id ?? ''
  })
  api.on('rs:toast', (m) => {
    const msg = m as { kind: ToastMsg['kind']; key: string; params?: Record<string, unknown> }
    toast(msg.kind, t(msg.key, msg.params))
  })
  api.on('rs:update-progress', (p) => { state.updateProgress = p as { done: number; total: number } })

  watch(() => state.cfg, () => {
    clearTimeout(saveTimer)
    saveTimer = window.setTimeout(flushSave, 30)
  }, { deep: true })
  watch(() => [state.cfg.language, state.systemLanguage], applyLanguage)
  watch(() => state.cfg.theme, applyTheme, { deep: true })
  watch(() => [state.cfg.overlay.mode, JSON.stringify(state.cfg.crosshair), JSON.stringify(state.cfg.reticle)], scheduleOverlay)
  watch(() => state.page, (p) => { state.cfg.ui.page = p })

  scheduleOverlay()
  state.ready = true
  if (!state.displays.length) void retryDisplays()
  if (state.updated) toast('success', t('about.updatedTo', { v: state.version }))
}
