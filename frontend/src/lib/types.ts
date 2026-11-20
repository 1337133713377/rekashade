// Mirrors internal/config/config.go and the Go bindings in app*.go.

export interface DisplayParams {
  vibrance: number
  brightness: number
  contrast: number
  gamma: number
  temperature: number
  red: number
  green: number
  blue: number
}

export interface ColorParams extends DisplayParams {
  hue: number
  saturation: number
}

export interface Preset {
  id: string
  name: string
  builtin?: boolean
  tags?: string[]
  params: ColorParams
}

export interface DisplayTuning extends DisplayParams {
  presetId: string
}

export interface Tuning {
  enabled: boolean
  hue: number
  saturation: number
  displays: Record<string, DisplayTuning>
}

export type OverlayMode = 'crosshair' | 'reticle'

export interface OverlayCfg {
  enabled: boolean
  mode: OverlayMode
  display: string
  offsetX: number
  offsetY: number
  followWindow: boolean
  onlyFullscreen: boolean
  hideFromCapture: boolean
}

export type BindAction =
  | 'toggleTuning'
  | 'toggleCrosshair'
  | 'nextPreset'
  | 'prevPreset'
  | 'toggleStretch'
  | 'showApp'
  | 'panic'
  | 'holdHide'

export interface Bind {
  action: BindAction
  key: number
  ctrl: boolean
  alt: boolean
  shift: boolean
}

export type Scaling = 'stretch' | 'aspect' | 'center'

export interface ResolutionCfg {
  display: string
  width: number
  height: number
  refresh: number
  scaling: Scaling
  confirmed: boolean
  revertOnExit: boolean
  autoApply: boolean
  autoGames: string[]
}

export interface GameRule {
  exe: string
  presetId: string
  enabled: boolean
}

export interface Settings {
  startWithWindows: boolean
  startMinimized: boolean
  closeToTray: boolean
  autoUpdate: boolean
  autoInstall: boolean
  applyOnStartup: boolean
  restoreOnExit: boolean
  gpuFeatures: boolean
}

export interface SavedCrosshair {
  name: string
  code: string
}

export interface CrosshairCfg {
  design: string
  color: string
  customColors: string[]
  ring: number
  arm: number
  gap: number
  thickness: number
  dot: number
  outline: number
  outlineColor: string
  opacity: number
  rotate: boolean
  tStyle: boolean
  saved: (SavedCrosshair | null)[]
}

export interface ReticleCfg {
  source: 'library' | 'custom'
  libraryId: string
  custom: string | null
  size: number
  color: string
  opacity: number
  rotation: number
  glow: boolean
}

export type Accent = 'violet' | 'magenta' | 'cyan' | 'emerald' | 'amber' | 'crimson' | 'ice'
export type Backdrop = 'aurora' | 'nebula' | 'midnight' | 'oled'

export interface ThemeCfg {
  accent: Accent
  backdrop: Backdrop
  glass: boolean
  motion: boolean
  scale: number
}

export interface ProfileCfg {
  name: string
  avatar: number
  since: string
}

export interface UiCfg {
  page: string
  displayTab: 'tone' | 'color'
  scene: string
  previewMode: 'demo' | 'curve'
  pinned: string[]
}

export interface Config {
  version: number
  language: 'auto' | 'en' | 'fr'
  tuning: Tuning
  presets: Preset[]
  overlay: OverlayCfg
  binds: Bind[]
  resolution: ResolutionCfg
  rules: GameRule[]
  settings: Settings
  crosshair: CrosshairCfg
  reticle: ReticleCfg
  theme: ThemeCfg
  profile: ProfileCfg
  ui: UiCfg
}

export interface DisplayMode {
  width: number
  height: number
  rates: number[]
}

export interface DisplayInfo {
  id: string
  number: number
  label: string
  monitor: string
  adapter: string
  vendor: 'nvidia' | 'amd' | 'intel' | 'other'
  primary: boolean
  x: number
  y: number
  width: number
  height: number
  refresh: number
  nativeWidth: number
  nativeHeight: number
  scaling: string
  modes: DisplayMode[]
  vibrance: '' | 'nvidia' | 'amd'
}

export interface GammaStatus {
  ok: boolean
  limited: boolean
  strength: number
  error?: string
}

export interface DisplayStatus {
  active: boolean
  matrix: 'off' | 'on' | 'error'
  matrixError?: string
  gamma: Record<string, GammaStatus>
  vibrance: Record<string, string>
  gpu: string[] | null
  gpuError?: string
}

export interface OverlayStatus {
  visible: boolean
  reason: '' | 'off' | 'noimage' | 'hold' | 'notfullscreen'
  hasImage: boolean
  holdHidden: boolean
}

export interface StretchStatus {
  active: boolean
  auto: boolean
  device: string
  width: number
  height: number
  refresh: number
  pending: number
  error?: string
}

export interface Release {
  version: string
  name: string
  notes: string
  publishedAt: string
  pageUrl: string
  assetSize: number
}

export interface UpdateCheck {
  current: string
  available: boolean
  latest?: Release
  checkedAt: string
}

export interface AppStatus {
  display: DisplayStatus
  overlay: OverlayStatus
  stretch: StretchStatus
  rule: string
  update?: UpdateCheck | null
  updating: boolean
  extendedGamma: boolean
  startup: boolean
  gpuProbeFailed: boolean
}

export interface Bootstrap {
  version: string
  config: Config
  displays: DisplayInfo[]
  status: AppStatus
  firstRun: boolean
  updated: boolean
  systemLanguage: 'en' | 'fr'
  windowsBuild: number
  dataDir: string
  mock: boolean
}

export interface TweakStatus {
  id: string
  on: boolean
  supported: boolean
  canUndo: boolean
  error?: string
}

export interface CleanResult {
  files: number
  bytes: number
  skipped: number
}

export interface ResolutionRequest {
  display: string
  width: number
  height: number
  refresh: number
  scaling: Scaling
}

export interface ToastMsg {
  id: number
  kind: 'info' | 'success' | 'error'
  text: string
}
