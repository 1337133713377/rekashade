// Crosshair model + pixel-exact renderer. The same canvas output is shown in
// the preview and sent to the native overlay, so they always match.
import type { CrosshairCfg, ReticleCfg } from './types'

export type CrosshairShape = Pick<CrosshairCfg, 'ring' | 'arm' | 'gap' | 'thickness' | 'dot' | 'outline' | 'rotate' | 'tStyle'>

export interface Design {
  id: string
  shape: CrosshairShape
}

const s = (ring: number, arm: number, gap: number, thickness: number, dot: number, outline = 1, rotate = false, tStyle = false): CrosshairShape => ({
  ring, arm, gap, thickness, dot, outline, rotate, tStyle,
})

export const DESIGNS: Design[] = [
  { id: 'scope', shape: s(13, 21, 0, 3, 0) },
  { id: 'ring', shape: s(9, 0, 0, 2, 0) },
  { id: 'ring-ticks', shape: s(10, 4, 3, 2, 0) },
  { id: 'ring-dot', shape: s(9, 0, 0, 2, 3) },
  { id: 'plus', shape: s(0, 7, 0, 2, 0) },
  { id: 'dot', shape: s(0, 0, 0, 2, 4) },
  { id: 'gap-cross', shape: s(0, 6, 4, 2, 0) },
  { id: 'bold-dot', shape: s(9, 0, 0, 3, 5) },
  { id: 'ring-cross', shape: s(9, 9, 0, 2, 0) },
  { id: 'outer-ticks', shape: s(9, 6, 11, 2, 0) },
  { id: 'x-cross', shape: s(0, 7, 3, 2, 0, 1, true) },
  { id: 't-cross', shape: s(0, 7, 3, 2, 2, 1, false, true) },
]

export const SWATCHES = ['#000000', '#FFFFFF', '#FF2BD6', '#2F6BFF', '#22E3FF', '#3BFF5B', '#FFE81F', '#FF2D2D']

export const LIMITS = {
  ring: [0, 40], arm: [0, 40], gap: [0, 20], thickness: [1, 8], dot: [0, 12], outline: [0, 4], opacity: [10, 100],
} as const

export function defaultCrosshair(): CrosshairCfg {
  return {
    design: 'scope', color: '#FFFFFF', customColors: [],
    ...DESIGNS[0].shape, outlineColor: '#000000', opacity: 100,
    saved: [null, null, null, null, null, null, null, null],
  }
}

export interface Rendered {
  canvas: HTMLCanvasElement
  anchorX: number
  anchorY: number
}

function clampInt(v: number, lo: number, hi: number) {
  return Math.max(lo, Math.min(hi, Math.round(v)))
}

/** Renders a crosshair at real pixel size. Odd-sized, centre pixel = anchor. */
export function renderCrosshair(c: CrosshairCfg): Rendered {
  const ring = clampInt(c.ring, 0, 40)
  const arm = clampInt(c.arm, 0, 40)
  const gap = clampInt(c.gap, 0, 20)
  const t = clampInt(c.thickness, 1, 8)
  const dot = clampInt(c.dot, 0, 12)
  const o = clampInt(c.outline, 0, 4)
  const h0 = Math.floor((t - 1) / 2)
  const reach = Math.max(ring + t, t + gap + arm, Math.ceil(dot / 2) + 1, 2)
  const ext = Math.ceil((c.rotate ? reach * 1.42 : reach) + o + 2)
  const size = ext * 2 + 1
  const cx = ext

  const shape = document.createElement('canvas')
  shape.width = shape.height = size
  const g = shape.getContext('2d')!

  const rects: [number, number, number, number][] = []
  const lo = cx - h0 // first pixel of the line thickness
  if (arm > 0) {
    rects.push([lo + t + gap, lo, arm, t]) // right
    rects.push([lo - gap - arm, lo, arm, t]) // left
    rects.push([lo, lo + t + gap, t, arm]) // bottom
    if (!c.tStyle) rects.push([lo, lo - gap - arm, t, arm]) // top
    if (gap === 0) rects.push([lo, lo, t, t]) // join
  }
  const dotRect = dot > 0 ? [cx - Math.floor((dot - 1) / 2), cx - Math.floor((dot - 1) / 2), dot, dot] : null
  const roundDot = dot >= 4

  const drawAll = (pad: number, color: string) => {
    g.fillStyle = color
    g.strokeStyle = color
    g.save()
    if (c.rotate) {
      g.translate(cx + 0.5, cx + 0.5)
      g.rotate(Math.PI / 4)
      g.translate(-(cx + 0.5), -(cx + 0.5))
    }
    for (const [x, y, w, h] of rects) g.fillRect(x - pad, y - pad, w + pad * 2, h + pad * 2)
    g.restore()
    if (ring > 0) {
      g.beginPath()
      g.lineWidth = t + pad * 2
      g.arc(cx + 0.5, cx + 0.5, ring + t / 2, 0, Math.PI * 2)
      g.stroke()
    }
    if (dotRect) {
      if (roundDot) {
        g.beginPath()
        g.arc(cx + 0.5, cx + 0.5, dot / 2 + pad, 0, Math.PI * 2)
        g.fill()
      } else {
        g.fillRect(dotRect[0] - pad, dotRect[1] - pad, dotRect[2] + pad * 2, dotRect[3] + pad * 2)
      }
    }
  }

  if (o > 0) drawAll(o, c.outlineColor || '#000000')
  if (o > 0) {
    // Punch the inner shape out of the outline so semi-transparent colours stay clean.
    g.globalCompositeOperation = 'destination-out'
    drawAll(0, '#000')
    g.globalCompositeOperation = 'source-over'
  }
  drawAll(0, c.color)

  const out = document.createElement('canvas')
  out.width = out.height = size
  const og = out.getContext('2d')!
  og.globalAlpha = clampInt(c.opacity, 10, 100) / 100
  og.drawImage(shape, 0, 0)
  return { canvas: out, anchorX: cx, anchorY: cx }
}

// ---------------------------------------------------------------------------
// Image reticles
// ---------------------------------------------------------------------------

export interface LibraryReticle {
  id: string
  svg: string // uses currentColor
}

const svg = (body: string) =>
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">${body}</svg>`

export const RETICLES: LibraryReticle[] = [
  { id: 'holo', svg: svg('<circle cx="50" cy="50" r="30" stroke-width="5"/><circle cx="50" cy="50" r="4.5" fill="currentColor" stroke="none"/><path d="M50 8v9M50 83v9M8 50h9M83 50h9" stroke-width="5"/>') },
  { id: 'chevron', svg: svg('<path d="M28 66 50 44l22 22" stroke-width="7"/><circle cx="50" cy="32" r="3.5" fill="currentColor" stroke="none"/>') },
  { id: 'mil-dot', svg: svg('<path d="M50 6v88M6 50h88" stroke-width="2.5"/><g fill="currentColor" stroke="none"><circle cx="50" cy="22" r="3"/><circle cx="50" cy="36" r="3"/><circle cx="50" cy="64" r="3"/><circle cx="50" cy="78" r="3"/><circle cx="22" cy="50" r="3"/><circle cx="36" cy="50" r="3"/><circle cx="64" cy="50" r="3"/><circle cx="78" cy="50" r="3"/></g>') },
  { id: 'triangle', svg: svg('<path d="M50 40 62 62H38Z" stroke-width="5"/><circle cx="50" cy="50" r="2.5" fill="currentColor" stroke="none"/>') },
  { id: 'bracket', svg: svg('<path d="M30 22H22v14M70 22h8v14M30 78H22V64M70 78h8V64" stroke-width="6"/><circle cx="50" cy="50" r="4" fill="currentColor" stroke="none"/>') },
  { id: 'diamond', svg: svg('<path d="M50 30 70 50 50 70 30 50Z" stroke-width="5"/><circle cx="50" cy="50" r="3" fill="currentColor" stroke="none"/>') },
  { id: 'tactical-x', svg: svg('<path d="M30 30l12 12M70 30 58 42M30 70l12-12M70 70 58 58" stroke-width="6"/><circle cx="50" cy="50" r="3" fill="currentColor" stroke="none"/>') },
  { id: 'circle-split', svg: svg('<path d="M50 18a32 32 0 0 1 32 32M50 82a32 32 0 0 1-32-32" stroke-width="6"/><circle cx="50" cy="50" r="4" fill="currentColor" stroke="none"/>') },
]

export function defaultReticle(): ReticleCfg {
  return { source: 'library', libraryId: 'holo', custom: null, size: 56, color: '#FF2BD6', opacity: 100, rotation: 0, glow: true }
}

export function reticleSvg(id: string, color: string): string {
  const r = RETICLES.find((x) => x.id === id) ?? RETICLES[0]
  return r.svg.replace('<svg ', `<svg style="color:${color}" `).replace(/currentColor/g, color)
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = reject
    img.src = src
  })
}

export async function renderReticle(r: ReticleCfg): Promise<Rendered> {
  const size = clampInt(r.size, 12, 300)
  const glow = r.glow ? Math.ceil(size * 0.12) + 2 : 0
  const half = Math.ceil((size * 1.42) / 2) + glow + 2
  const dim = half * 2 + 1
  const canvas = document.createElement('canvas')
  canvas.width = canvas.height = dim
  const g = canvas.getContext('2d')!
  const src = r.source === 'custom' && r.custom
    ? r.custom
    : 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(reticleSvg(r.libraryId, r.color))
  try {
    const img = await loadImage(src)
    const scale = size / Math.max(img.naturalWidth || size, img.naturalHeight || size)
    const w = (img.naturalWidth || size) * scale
    const h = (img.naturalHeight || size) * scale
    g.save()
    g.globalAlpha = clampInt(r.opacity, 10, 100) / 100
    g.translate(half + 0.5, half + 0.5)
    g.rotate(((r.rotation || 0) * Math.PI) / 180)
    if (glow) {
      g.shadowColor = r.source === 'custom' ? 'rgba(255,255,255,0.65)' : r.color
      g.shadowBlur = glow
    }
    g.drawImage(img, -w / 2, -h / 2, w, h)
    g.restore()
  } catch {
    /* unreadable image: leave the canvas empty */
  }
  return { canvas, anchorX: half, anchorY: half }
}

/** Shrinks a user-picked image so settings stay small (max 256 px). */
export async function importImage(file: File): Promise<string> {
  const url = URL.createObjectURL(file)
  try {
    const img = await loadImage(url)
    const max = 256
    const scale = Math.min(1, max / Math.max(img.naturalWidth, img.naturalHeight))
    const c = document.createElement('canvas')
    c.width = Math.max(1, Math.round(img.naturalWidth * scale))
    c.height = Math.max(1, Math.round(img.naturalHeight * scale))
    c.getContext('2d')!.drawImage(img, 0, 0, c.width, c.height)
    return c.toDataURL('image/png')
  } finally {
    URL.revokeObjectURL(url)
  }
}
