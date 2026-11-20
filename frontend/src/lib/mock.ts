// Browser-only stand-in for the Go side, so the UI can be developed and
// screenshotted without Windows. Nothing here runs in the real app.
import defaults from '@defaults'
import type { Backend } from './api'
import type { AppStatus, Bind, Config, DisplayInfo, Preset, TweakStatus, UpdateCheck } from './types'

const KEY = 'rekashade-mock-config'

function goDefaults(): Config {
  return {
    version: 1,
    language: 'auto',
    tuning: { enabled: false, hue: 0, saturation: 100, displays: {} },
    presets: (defaults.presets as Preset[]).map((p) => ({ ...p, params: { ...p.params } })),
    overlay: { enabled: false, mode: 'crosshair', display: '', offsetX: 0, offsetY: 0, followWindow: false, onlyFullscreen: false, hideFromCapture: false },
    binds: (defaults.binds as Bind[]).map((b) => ({ ...b })),
    resolution: { display: '', width: 1440, height: 1080, refresh: 0, scaling: 'stretch', confirmed: false, revertOnExit: true, autoApply: false, autoGames: [] },
    rules: [],
    settings: { startWithWindows: false, startMinimized: true, closeToTray: true, autoUpdate: true, autoInstall: false, applyOnStartup: true, restoreOnExit: true, gpuFeatures: true },
  } as unknown as Config
}

const modes = (list: [number, number, number[]][]) => list.map(([width, height, rates]) => ({ width, height, rates }))

const DISPLAYS: DisplayInfo[] = [
  {
    id: '\\\\.\\DISPLAY1', number: 1, label: 'Display 1', monitor: 'LG ULTRAGEAR', adapter: 'NVIDIA GeForce RTX 4070', vendor: 'nvidia',
    primary: true, x: 0, y: 0, width: 2560, height: 1440, refresh: 165, nativeWidth: 2560, nativeHeight: 1440, scaling: 'preferred', vibrance: 'nvidia',
    modes: modes([[2560, 1440, [165, 144, 120, 60]], [1920, 1080, [165, 144, 120, 60]], [1680, 1050, [144, 60]], [1600, 900, [144, 60]],
      [1440, 1080, [165, 144, 60]], [1280, 1024, [144, 75, 60]], [1280, 960, [144, 60]], [1024, 768, [144, 75, 60]], [800, 600, [60]]]),
  },
  {
    id: '\\\\.\\DISPLAY2', number: 2, label: 'Display 2', monitor: 'DELL S2421H', adapter: 'NVIDIA GeForce RTX 4070', vendor: 'nvidia',
    primary: false, x: 2560, y: 0, width: 1920, height: 1080, refresh: 75, nativeWidth: 1920, nativeHeight: 1080, scaling: 'preferred', vibrance: 'nvidia',
    modes: modes([[1920, 1080, [75, 60]], [1680, 1050, [60]], [1600, 900, [60]], [1280, 1024, [75, 60]], [1024, 768, [75, 60]]]),
  },
]

export function createMock(): Backend {
  const listeners = new Map<string, Set<(d: unknown) => void>>()
  const emit = (name: string, data: unknown) => listeners.get(name)?.forEach((cb) => cb(data))
  let cfg: Config
  try {
    cfg = JSON.parse(localStorage.getItem(KEY) || '') as Config
  } catch {
    cfg = goDefaults()
  }
  if (!cfg || !cfg.presets) cfg = goDefaults()
  let stretch = { active: false, auto: false, device: '', width: 0, height: 0, refresh: 0, pending: 0 }
  let countdown: number | undefined
  const tweaks: TweakStatus[] = ['powerPlan', 'gameMode', 'gameDVR', 'windowedBoost', 'mouseAccel', 'stickyKeys', 'notifications', 'backgroundApps', 'transparency', 'animations']
    .map((id) => ({ id, on: id === 'gameMode' || id === 'stickyKeys' || id === 'animations', supported: true, canUndo: id !== 'gameMode' }))
  let update: UpdateCheck | null = null
  let overlayImage = false

  const status = (): AppStatus => ({
    display: {
      active: cfg.tuning.enabled,
      matrix: cfg.tuning.enabled && (cfg.tuning.hue !== 0 || cfg.tuning.saturation !== 100) ? 'on' : 'off',
      gamma: Object.fromEntries(DISPLAYS.map((d) => [d.id, { ok: true, limited: false, strength: 1 }])),
      vibrance: Object.fromEntries(DISPLAYS.map((d) => [d.id, cfg.tuning.enabled ? 'nvidia' : 'off'])),
      gpu: ['nvidia'],
    },
    overlay: {
      visible: cfg.overlay.enabled && overlayImage, hasImage: overlayImage, holdHidden: false,
      reason: !cfg.overlay.enabled ? 'off' : overlayImage ? '' : 'noimage',
    },
    stretch: { ...stretch },
    rule: '',
    update,
    updating: false,
    extendedGamma: false,
    startup: cfg.settings.startWithWindows,
    gpuProbeFailed: false,
  })
  const pushStatus = () => emit('rs:status', status())
  const save = () => {
    try { localStorage.setItem(KEY, JSON.stringify(cfg)) } catch { /* quota */ }
  }
  const wait = (ms = 120) => new Promise((r) => setTimeout(r, ms))

  return {
    async bootstrap() {
      return {
        version: '1.0.0', config: JSON.parse(JSON.stringify(cfg)),
        // Test switch: simulate a backend that answered before it listed the monitors.
        displays: localStorage.getItem('rekashade-mock-no-displays') ? [] : DISPLAYS,
        status: status(), firstRun: !localStorage.getItem(KEY),
        updated: false, systemLanguage: navigator.language.startsWith('fr') ? 'fr' : 'en', windowsBuild: 22631, dataDir: 'C:\\Users\\you\\AppData\\Roaming\\RekaShade', mock: true,
      }
    },
    async saveConfig(next) {
      cfg = JSON.parse(JSON.stringify(next))
      save()
      pushStatus()
    },
    async setOverlayImage() { overlayImage = true; pushStatus() },
    async getDisplays() { return DISPLAYS },
    async refreshDisplays() { return DISPLAYS },
    async getStatus() { return status() },
    async testResolution(req) {
      await wait(300)
      stretch = { active: true, auto: false, device: req.display || DISPLAYS[0].id, width: req.width, height: req.height, refresh: req.refresh, pending: 15 }
      cfg.resolution = { ...cfg.resolution, display: stretch.device, width: req.width, height: req.height, refresh: req.refresh, scaling: req.scaling, confirmed: false }
      emit('rs:config', JSON.parse(JSON.stringify(cfg)))
      clearInterval(countdown)
      countdown = window.setInterval(() => {
        stretch.pending--
        if (stretch.pending <= 0) {
          clearInterval(countdown)
          stretch = { active: false, auto: false, device: '', width: 0, height: 0, refresh: 0, pending: 0 }
          emit('rs:toast', { kind: 'info', key: 'toast.resReverted' })
        }
        pushStatus()
      }, 1000)
      pushStatus()
      return { ...stretch }
    },
    async confirmResolution() {
      clearInterval(countdown)
      stretch.pending = 0
      cfg.resolution.confirmed = true
      emit('rs:config', JSON.parse(JSON.stringify(cfg)))
      pushStatus()
    },
    async revertResolution() {
      clearInterval(countdown)
      stretch = { active: false, auto: false, device: '', width: 0, height: 0, refresh: 0, pending: 0 }
      pushStatus()
    },
    async optimizerList() { await wait(250); return tweaks.map((t) => ({ ...t })) },
    async optimizerSet(id, on) {
      await wait(150)
      const t = tweaks.find((x) => x.id === id)!
      t.on = on
      t.canUndo = on
      return { ...t }
    },
    async optimizerAll(on) {
      await wait(400)
      tweaks.forEach((t) => { if (on) { if (!t.on) { t.on = true; t.canUndo = true } } else if (t.canUndo) { t.on = false; t.canUndo = false } })
      return tweaks.map((t) => ({ ...t }))
    },
    async cleanTemp() { await wait(700); return { files: 1284, bytes: 734_003_200, skipped: 17 } },
    async checkForUpdates() {
      await wait(600)
      update = {
        current: '1.0.0', available: true, checkedAt: new Date().toISOString(),
        latest: { version: '1.1.0', name: 'Reka Shade 1.1.0', notes: '- Faster preset switching\n- New reticles\n- Bug fixes', publishedAt: new Date().toISOString(), pageUrl: 'https://github.com/1337133713377/rekashade/releases', assetSize: 11_200_000 },
      }
      pushStatus()
      return update
    },
    async installUpdate() {
      for (let i = 0; i <= 20; i++) { await wait(80); emit('rs:update-progress', { done: i * 560_000, total: 11_200_000 }) }
    },
    async windowMinimize() {},
    async windowClose() {},
    async quitApp() {},
    async openURL(u) { window.open(u, '_blank') },
    async pauseBinds() {},
    async exportSettings() { return 'C:\\Users\\you\\Documents\\RekaShade-settings.json' },
    async importSettings() { return null },
    async enableExtendedGamma() {},
    async runningApps() { return ['valorant-win64-shipping.exe', 'cs2.exe', 'fortniteclient-win64-shipping.exe', 'discord.exe', 'chrome.exe'] },
    async panic() {
      cfg.tuning.enabled = false
      cfg.overlay.enabled = false
      emit('rs:config', JSON.parse(JSON.stringify(cfg)))
      pushStatus()
    },
    async openDataFolder() {},
    async diagnostics() { return JSON.stringify({ version: '1.0.0', mock: true, displays: DISPLAYS.length }, null, 2) },
    on(event, cb) {
      if (!listeners.has(event)) listeners.set(event, new Set())
      listeners.get(event)!.add(cb)
      return () => listeners.get(event)?.delete(cb)
    },
  }
}
