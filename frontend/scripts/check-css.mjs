// Fails the build if a compiled rule hides or dims the whole page.
// Guards against Vue's scoped ":global(X) .child" pitfall, which compiles to
// plain "X" and once put `display: none` on <html> (the OLED backdrop).
import { readdirSync, readFileSync } from 'node:fs'

const dir = new URL('../dist/assets/', import.meta.url)
const page = /^(:root|html|body|#app)(\[[^\]]*\])*$/
const hiding = /(^|;)\s*(opacity|display|visibility|filter)\s*:/
const bad = []
for (const f of readdirSync(dir).filter((n) => n.endsWith('.css'))) {
  const css = readFileSync(new URL(f, dir), 'utf8')
  for (const [, sel, body] of css.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
    if (sel.split(',').some((s) => page.test(s.trim())) && hiding.test(body)) bad.push(`${f}: ${sel.trim()} {${body}}`)
  }
}
if (bad.length) {
  console.error('check-css: rules that would hide or dim the whole app:\n' + bad.join('\n'))
  process.exit(1)
}
console.log('check-css: ok')
