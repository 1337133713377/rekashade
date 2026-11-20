// Mirror of internal/colorx/colorx.go — keep both in sync.
// The preview uses these formulas so it shows what Windows will apply.
import type { ColorParams, DisplayParams } from './types'

export type Mat3 = [[number, number, number], [number, number, number], [number, number, number]]

const clamp01 = (x: number) => (x < 0 ? 0 : x > 1 ? 1 : x)

export function neutralDisplay(): DisplayParams {
  return { vibrance: 100, brightness: 100, contrast: 100, gamma: 1, temperature: 6500, red: 100, green: 100, blue: 100 }
}

export function neutralParams(): ColorParams {
  return { hue: 0, saturation: 100, ...neutralDisplay() }
}

function toneBase(t: DisplayParams, x: number): number {
  let g = t.gamma
  if (!(g >= 0.1)) g = 1
  let y = Math.pow(x, 1 / g)
  y = (y - 0.5) * (t.contrast / 100) + 0.5
  const b = (t.brightness - 100) / 100
  y = y * (1 + b * 0.6) + Math.max(b, 0) * 0.12 * (1 - y)
  return y
}

export function kelvinRGB(k: number): [number, number, number] {
  const t = k / 100
  let r: number, g: number, b: number
  if (t <= 66) {
    r = 255
    g = 99.4708025861 * Math.log(t) - 161.1195681661
  } else {
    r = 329.698727446 * Math.pow(t - 60, -0.1332047592)
    g = 288.1221695283 * Math.pow(t - 60, -0.0755148492)
  }
  if (t >= 66) b = 255
  else if (t <= 19) b = 0
  else b = 138.5177312231 * Math.log(t - 10) - 305.0447927307
  const c = (v: number) => Math.min(255, Math.max(0, v)) / 255
  return [c(r), c(g), c(b)]
}

export function temperatureGains(k: number): [number, number, number] {
  if (k <= 0 || Math.abs(k - 6500) < 1) return [1, 1, 1]
  const [r, g, b] = kelvinRGB(k)
  const [r0, g0, b0] = kelvinRGB(6500)
  const gains: [number, number, number] = [r / r0, g / g0, b / b0]
  const m = Math.max(...gains)
  return [gains[0] / m, gains[1] / m, gains[2] / m]
}

export function channelGains(t: DisplayParams): [number, number, number] {
  const tg = temperatureGains(t.temperature)
  return [(t.red / 100) * tg[0], (t.green / 100) * tg[1], (t.blue / 100) * tg[2]]
}

/** Tone curve of channel c (0=R,1=G,2=B) at x in 0..1. */
export function curve(t: DisplayParams, c: number, x: number): number {
  return clamp01(toneBase(t, x) * channelGains(t)[c])
}

/** Ramp as normalised floats [channel][0..n-1], monotonic like the Go version. */
export function buildRamp(t: DisplayParams, n = 256): number[][] {
  const gains = channelGains(t)
  const out = [new Array<number>(n), new Array<number>(n), new Array<number>(n)]
  for (let i = 0; i < n; i++) {
    const base = toneBase(t, i / (n - 1))
    for (let c = 0; c < 3; c++) out[c][i] = Math.round(clamp01(base * gains[c]) * 65535) / 65535
  }
  for (let c = 0; c < 3; c++) for (let i = 1; i < n; i++) if (out[c][i] < out[c][i - 1]) out[c][i] = out[c][i - 1]
  return out
}

export function identity3(): Mat3 {
  return [[1, 0, 0], [0, 1, 0], [0, 0, 1]]
}

export function saturate(s: number): Mat3 {
  return [
    [0.213 + 0.787 * s, 0.715 - 0.715 * s, 0.072 - 0.072 * s],
    [0.213 - 0.213 * s, 0.715 + 0.285 * s, 0.072 - 0.072 * s],
    [0.213 - 0.213 * s, 0.715 - 0.715 * s, 0.072 + 0.928 * s],
  ]
}

export function hueRotate(deg: number): Mat3 {
  const a = (deg * Math.PI) / 180
  const c = Math.cos(a)
  const s = Math.sin(a)
  return [
    [0.213 + c * 0.787 - s * 0.213, 0.715 - c * 0.715 - s * 0.715, 0.072 - c * 0.072 + s * 0.928],
    [0.213 - c * 0.213 + s * 0.143, 0.715 + c * 0.285 + s * 0.14, 0.072 - c * 0.072 - s * 0.283],
    [0.213 - c * 0.213 - s * 0.787, 0.715 - c * 0.715 + s * 0.715, 0.072 + c * 0.928 + s * 0.072],
  ]
}

export function mul(a: Mat3, b: Mat3): Mat3 {
  const m = [[0, 0, 0], [0, 0, 0], [0, 0, 0]] as Mat3
  for (let i = 0; i < 3; i++) for (let j = 0; j < 3; j++) for (let k = 0; k < 3; k++) m[i][j] += a[i][k] * b[k][j]
  return m
}

export function colorMatrix(hueDeg: number, saturationPct: number): Mat3 {
  return mul(hueRotate(hueDeg), saturate(saturationPct / 100))
}

/** SVG feColorMatrix "values" (4x5, column-vector form). */
export function svgMatrix(m: Mat3): string {
  const r = (v: number) => v.toFixed(5)
  return [
    `${r(m[0][0])} ${r(m[0][1])} ${r(m[0][2])} 0 0`,
    `${r(m[1][0])} ${r(m[1][1])} ${r(m[1][2])} 0 0`,
    `${r(m[2][0])} ${r(m[2][1])} ${r(m[2][2])} 0 0`,
    '0 0 0 1 0',
  ].join(' ')
}

/** Everything an SVG filter needs to preview a look. */
export function previewFilter(p: ColorParams, samples = 64) {
  const ramp = buildRamp(p, samples)
  return {
    matrix: svgMatrix(colorMatrix(p.hue, p.saturation)),
    vibrance: (p.vibrance / 100).toFixed(4),
    tables: ramp.map((ch) => ch.map((v) => v.toFixed(4)).join(' ')) as [string, string, string],
  }
}

export function isNeutral(p: ColorParams): boolean {
  const n = neutralParams()
  return (Object.keys(n) as (keyof ColorParams)[]).every((k) => Math.abs(p[k] - n[k]) < 1e-6)
}

export function sameParams(a: ColorParams, b: ColorParams): boolean {
  return (Object.keys(neutralParams()) as (keyof ColorParams)[]).every((k) => Math.abs(a[k] - b[k]) < 0.001)
}
