// Share codes: compact, copy-pasteable strings for crosshairs and presets.
//   crosshair → "RKX1-…"   preset → "RKP1-…"   reticle → "RKR1-…"
import type { ColorParams, CrosshairCfg, Preset, ReticleCfg } from './types'

const b64url = (s: string) =>
  btoa(unescape(encodeURIComponent(s))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')

const unb64url = (s: string) => {
  const pad = s.length % 4 === 0 ? '' : '='.repeat(4 - (s.length % 4))
  return decodeURIComponent(escape(atob(s.replace(/-/g, '+').replace(/_/g, '/') + pad)))
}

const hex = (c: string) => (/^#?[0-9a-fA-F]{6}$/.test(c) ? c.replace('#', '').toUpperCase() : 'FFFFFF')
const num = (v: unknown, lo: number, hi: number, def: number) => {
  const n = Number(v)
  return Number.isFinite(n) ? Math.max(lo, Math.min(hi, n)) : def
}

export function encodeCrosshair(c: CrosshairCfg): string {
  const a = [c.ring, c.arm, c.gap, c.thickness, c.dot, c.outline, c.opacity, c.rotate ? 1 : 0, c.tStyle ? 1 : 0, hex(c.color), hex(c.outlineColor), c.design]
  return 'RKX1-' + b64url(JSON.stringify(a))
}

export function decodeCrosshair(code: string): Partial<CrosshairCfg> | null {
  const m = /^RKX1-([A-Za-z0-9_-]+)$/.exec(code.trim())
  if (!m) return null
  try {
    const a = JSON.parse(unb64url(m[1]))
    if (!Array.isArray(a) || a.length < 11) return null
    return {
      ring: num(a[0], 0, 40, 0), arm: num(a[1], 0, 40, 0), gap: num(a[2], 0, 20, 0),
      thickness: num(a[3], 1, 8, 2), dot: num(a[4], 0, 12, 0), outline: num(a[5], 0, 4, 1),
      opacity: num(a[6], 10, 100, 100), rotate: a[7] === 1, tStyle: a[8] === 1,
      color: '#' + hex(String(a[9])), outlineColor: '#' + hex(String(a[10])),
      design: typeof a[11] === 'string' ? a[11].slice(0, 32) : 'custom',
    }
  } catch {
    return null
  }
}

const PARAM_KEYS: (keyof ColorParams)[] = ['hue', 'saturation', 'vibrance', 'brightness', 'contrast', 'gamma', 'temperature', 'red', 'green', 'blue']
const RANGES: Record<keyof ColorParams, [number, number, number]> = {
  hue: [-180, 180, 0], saturation: [0, 300, 100], vibrance: [0, 200, 100], brightness: [50, 150, 100],
  contrast: [50, 150, 100], gamma: [0.4, 2.5, 1], temperature: [3500, 10000, 6500], red: [0, 200, 100],
  green: [0, 200, 100], blue: [0, 200, 100],
}

export function encodePreset(name: string, p: ColorParams): string {
  return 'RKP1-' + b64url(JSON.stringify([name.slice(0, 40), ...PARAM_KEYS.map((k) => Math.round(p[k] * 100) / 100)]))
}

export function decodePreset(code: string): { name: string; params: ColorParams } | null {
  const m = /^RKP1-([A-Za-z0-9_-]+)$/.exec(code.trim())
  if (!m) return null
  try {
    const a = JSON.parse(unb64url(m[1]))
    if (!Array.isArray(a) || a.length < 11) return null
    const params = {} as ColorParams
    PARAM_KEYS.forEach((k, i) => {
      const [lo, hi, def] = RANGES[k]
      params[k] = num(a[i + 1], lo, hi, def)
    })
    return { name: String(a[0] || 'Shared preset').slice(0, 40), params }
  } catch {
    return null
  }
}

export function encodeReticle(r: ReticleCfg): string | null {
  if (r.source !== 'library') return null
  return 'RKR1-' + b64url(JSON.stringify([r.libraryId, r.size, hex(r.color), r.opacity, r.rotation, r.glow ? 1 : 0]))
}

export function decodeReticle(code: string): Partial<ReticleCfg> | null {
  const m = /^RKR1-([A-Za-z0-9_-]+)$/.exec(code.trim())
  if (!m) return null
  try {
    const a = JSON.parse(unb64url(m[1]))
    if (!Array.isArray(a) || a.length < 6) return null
    return {
      source: 'library', libraryId: String(a[0]).slice(0, 32), size: num(a[1], 12, 300, 56),
      color: '#' + hex(String(a[2])), opacity: num(a[3], 10, 100, 100), rotation: num(a[4], -180, 180, 0), glow: a[5] === 1,
    }
  } catch {
    return null
  }
}

export type DecodedCode =
  | { kind: 'crosshair'; data: Partial<CrosshairCfg> }
  | { kind: 'preset'; data: { name: string; params: ColorParams } }
  | { kind: 'reticle'; data: Partial<ReticleCfg> }

export function decodeAny(code: string): DecodedCode | null {
  const t = code.trim()
  if (t.startsWith('RKX1-')) { const d = decodeCrosshair(t); return d ? { kind: 'crosshair', data: d } : null }
  if (t.startsWith('RKP1-')) { const d = decodePreset(t); return d ? { kind: 'preset', data: d } : null }
  if (t.startsWith('RKR1-')) { const d = decodeReticle(t); return d ? { kind: 'reticle', data: d } : null }
  return null
}

export function presetCode(p: Preset): string {
  return encodePreset(p.name, p.params)
}
