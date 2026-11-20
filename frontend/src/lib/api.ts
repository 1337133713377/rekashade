// Thin typed layer over the Go bindings (window.go.main.App.*). When the UI
// runs in a normal browser (development, screenshots) a mock backend is used.
import type {
  AppStatus, Bootstrap, CleanResult, Config, DisplayInfo, ResolutionRequest, StretchStatus, TweakStatus, UpdateCheck,
} from './types'
import { createMock } from './mock'

export interface Backend {
  bootstrap(): Promise<Bootstrap>
  saveConfig(cfg: Config): Promise<void>
  setOverlayImage(dataUrl: string, anchorX: number, anchorY: number): Promise<void>
  getDisplays(): Promise<DisplayInfo[]>
  refreshDisplays(): Promise<DisplayInfo[]>
  getStatus(): Promise<AppStatus>
  testResolution(req: ResolutionRequest): Promise<StretchStatus>
  confirmResolution(): Promise<void>
  revertResolution(): Promise<void>
  optimizerList(): Promise<TweakStatus[]>
  optimizerSet(id: string, on: boolean): Promise<TweakStatus>
  optimizerAll(on: boolean): Promise<TweakStatus[]>
  cleanTemp(): Promise<CleanResult>
  checkForUpdates(): Promise<UpdateCheck>
  installUpdate(): Promise<void>
  windowMinimize(): Promise<void>
  windowClose(): Promise<void>
  quitApp(): Promise<void>
  openURL(url: string): Promise<void>
  pauseBinds(paused: boolean): Promise<void>
  exportSettings(): Promise<string>
  importSettings(): Promise<Config | null>
  enableExtendedGamma(): Promise<void>
  runningApps(): Promise<string[]>
  panic(): Promise<void>
  openDataFolder(): Promise<void>
  diagnostics(): Promise<string>
  on(event: string, cb: (data: unknown) => void): () => void
}

function wails(): Backend {
  const call = <T>(name: string, ...args: unknown[]): Promise<T> => {
    const fn = window.go?.main?.App?.[name]
    if (!fn) return Promise.reject(new Error(`binding ${name} missing`))
    return fn(...args) as Promise<T>
  }
  return {
    bootstrap: () => call('Bootstrap'),
    saveConfig: (cfg) => call('SaveConfig', cfg),
    setOverlayImage: (d, x, y) => call('SetOverlayImage', d, x, y),
    getDisplays: () => call('GetDisplays'),
    refreshDisplays: () => call('RefreshDisplays'),
    getStatus: () => call('GetStatus'),
    testResolution: (req) => call('TestResolution', req),
    confirmResolution: () => call('ConfirmResolution'),
    revertResolution: () => call('RevertResolution'),
    optimizerList: () => call('OptimizerList'),
    optimizerSet: (id, on) => call('OptimizerSet', id, on),
    optimizerAll: (on) => call('OptimizerAll', on),
    cleanTemp: () => call('CleanTemp'),
    checkForUpdates: () => call('CheckForUpdates'),
    installUpdate: () => call('InstallUpdate'),
    windowMinimize: () => call('WindowMinimize'),
    windowClose: () => call('WindowClose'),
    quitApp: () => call('QuitApp'),
    openURL: (u) => call('OpenURL', u),
    pauseBinds: (p) => call('PauseBinds', p),
    exportSettings: () => call('ExportSettings'),
    importSettings: () => call('ImportSettings'),
    enableExtendedGamma: () => call('EnableExtendedGamma'),
    runningApps: () => call('RunningApps'),
    panic: () => call('Panic'),
    openDataFolder: () => call('OpenDataFolder'),
    diagnostics: () => call('Diagnostics'),
    on: (event, cb) => window.runtime?.EventsOn(event, (...d: unknown[]) => cb(d[0])) ?? (() => {}),
  }
}

export const isNative = typeof window !== 'undefined' && !!window.go?.main?.App
export const api: Backend = isNative ? wails() : createMock()

export function errText(e: unknown): string {
  if (e instanceof Error) return e.message
  return typeof e === 'string' ? e : JSON.stringify(e)
}
