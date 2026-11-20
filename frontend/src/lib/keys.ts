// Windows virtual-key codes. In WebView2 (Chromium on Windows) KeyboardEvent
// .keyCode already is the layout-aware VK code, which is exactly what the
// Go side polls with GetAsyncKeyState.
import type { Bind } from './types'

const NAMES: Record<number, string> = {
  0x01: 'Mouse 1', 0x02: 'Mouse 2', 0x04: 'Mouse 3', 0x05: 'Mouse 4', 0x06: 'Mouse 5',
  0x08: 'Backspace', 0x09: 'Tab', 0x0d: 'Enter', 0x13: 'Pause', 0x14: 'Caps Lock', 0x1b: 'Esc', 0x20: 'Space',
  0x21: 'Page Up', 0x22: 'Page Down', 0x23: 'End', 0x24: 'Home', 0x25: '←', 0x26: '↑', 0x27: '→', 0x28: '↓',
  0x2c: 'Print Screen', 0x2d: 'Insert', 0x2e: 'Delete', 0x90: 'Num Lock', 0x91: 'Scroll Lock',
  0x6a: 'Num *', 0x6b: 'Num +', 0x6d: 'Num −', 0x6e: 'Num .', 0x6f: 'Num /',
  0xba: ';', 0xbb: '=', 0xbc: ',', 0xbd: '-', 0xbe: '.', 0xbf: '/', 0xc0: '`', 0xdb: '[', 0xdc: '\\', 0xdd: ']', 0xde: "'", 0xe2: '<',
}

const CODE_FALLBACK: Record<string, number> = {
  Backspace: 0x08, Tab: 0x09, Enter: 0x0d, Pause: 0x13, CapsLock: 0x14, Escape: 0x1b, Space: 0x20,
  PageUp: 0x21, PageDown: 0x22, End: 0x23, Home: 0x24, ArrowLeft: 0x25, ArrowUp: 0x26, ArrowRight: 0x27, ArrowDown: 0x28,
  PrintScreen: 0x2c, Insert: 0x2d, Delete: 0x2e, NumLock: 0x90, ScrollLock: 0x91,
  NumpadMultiply: 0x6a, NumpadAdd: 0x6b, NumpadSubtract: 0x6d, NumpadDecimal: 0x6e, NumpadDivide: 0x6f,
  Semicolon: 0xba, Equal: 0xbb, Comma: 0xbc, Minus: 0xbd, Period: 0xbe, Slash: 0xbf, Backquote: 0xc0,
  BracketLeft: 0xdb, Backslash: 0xdc, BracketRight: 0xdd, Quote: 0xde, IntlBackslash: 0xe2,
}

const MODIFIER_VKS = new Set([0x10, 0x11, 0x12, 0x5b, 0x5c, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5])

export function vkName(vk: number): string {
  if (vk >= 0x41 && vk <= 0x5a) return String.fromCharCode(vk)
  if (vk >= 0x30 && vk <= 0x39) return String.fromCharCode(vk)
  if (vk >= 0x70 && vk <= 0x87) return 'F' + (vk - 0x6f)
  if (vk >= 0x60 && vk <= 0x69) return 'Num ' + (vk - 0x60)
  return NAMES[vk] ?? `Key ${vk}`
}

/** Returns the VK code of a key event, or 0 for a lone modifier. */
export function vkFromEvent(e: KeyboardEvent): number {
  let vk = e.keyCode || 0
  if (!vk || vk === 229) {
    const c = e.code
    if (/^Key[A-Z]$/.test(c)) vk = c.charCodeAt(3)
    else if (/^Digit[0-9]$/.test(c)) vk = c.charCodeAt(5)
    else if (/^F([1-9]|1[0-9]|2[0-4])$/.test(c)) vk = 0x6f + Number(c.slice(1))
    else if (/^Numpad[0-9]$/.test(c)) vk = 0x60 + Number(c.slice(6))
    else vk = CODE_FALLBACK[c] ?? 0
  }
  return MODIFIER_VKS.has(vk) ? 0 : vk
}

export function vkFromMouse(button: number): number {
  return [0x01, 0x04, 0x02, 0x05, 0x06][button] ?? 0
}

export function bindLabel(b: Pick<Bind, 'key' | 'ctrl' | 'alt' | 'shift'>): string {
  if (!b.key) return ''
  const parts: string[] = []
  if (b.ctrl) parts.push('Ctrl')
  if (b.alt) parts.push('Alt')
  if (b.shift) parts.push('Shift')
  parts.push(vkName(b.key))
  return parts.join(' + ')
}
