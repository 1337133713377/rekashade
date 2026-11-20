// Hand-made vector scenes used to preview looks (no game screenshots needed).
// Rendered as <img> data URLs so their gradient ids never clash.

function rng(seed: number) {
  let s = seed >>> 0
  return () => {
    s = (s * 1664525 + 1013904223) >>> 0
    return s / 4294967296
  }
}

const canyon = (() => {
  const r = rng(7)
  let flowers = ''
  for (let i = 0; i < 26; i++) {
    const x = (i * 37 + r() * 20) % 480
    const y = 212 + r() * 50
    const c = ['#ef4444', '#f59e0b', '#ec4899', '#f97316'][i % 4]
    flowers += `<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${(2 + r() * 2.2).toFixed(1)}" fill="${c}"/>`
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 270">
<defs>
<linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#1d4ed8"/><stop offset=".5" stop-color="#5aa5f5"/><stop offset="1" stop-color="#fcd9a8"/></linearGradient>
<radialGradient id="sun" cx=".79" cy=".25" r=".42"><stop offset="0" stop-color="#fff7d6"/><stop offset=".22" stop-color="#fde68a" stop-opacity=".85"/><stop offset="1" stop-color="#fde68a" stop-opacity="0"/></radialGradient>
<linearGradient id="m1" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fb923c"/><stop offset="1" stop-color="#c2410c"/></linearGradient>
<linearGradient id="m2" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#f97316"/><stop offset="1" stop-color="#9a3412"/></linearGradient>
<linearGradient id="sand" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fcd34d"/><stop offset="1" stop-color="#d97706"/></linearGradient>
<linearGradient id="road" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#b8b2ad"/><stop offset="1" stop-color="#78716c"/></linearGradient>
<linearGradient id="wall" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fef3c7"/><stop offset="1" stop-color="#fcd9a0"/></linearGradient>
</defs>
<rect width="480" height="270" fill="url(#sky)"/>
<rect width="480" height="270" fill="url(#sun)"/>
<circle cx="378" cy="66" r="20" fill="#fffbeb"/>
<g fill="#fff" opacity=".9"><ellipse cx="96" cy="52" rx="34" ry="10"/><ellipse cx="120" cy="45" rx="22" ry="11"/><ellipse cx="250" cy="34" rx="28" ry="8"/><ellipse cx="268" cy="29" rx="16" ry="8"/></g>
<path d="M0 168 L36 152 L70 152 L82 136 L128 136 L142 152 L188 152 L200 162 L258 162 L270 142 L330 142 L346 160 L400 160 L412 147 L456 147 L470 160 L480 160 V195 H0Z" fill="url(#m2)" opacity=".82"/>
<path d="M296 196 L308 122 L338 113 L378 113 L398 126 L430 120 L470 128 L480 131 V205 H296Z" fill="url(#m1)"/>
<path d="M312 142 H474 M306 160 H478 M302 178 H480" stroke="#fed7aa" stroke-opacity=".45" stroke-width="2"/>
<path d="M0 186 Q120 174 240 186 T480 183 V270 H0Z" fill="url(#sand)"/>
<path d="M206 270 L234 188 L252 188 L306 270Z" fill="url(#road)"/>
<path d="M243 196 V206 M245 218 V232 M248 246 V266" stroke="#fef9c3" stroke-width="3"/>
<g>
<path d="M58 152 L120 118 L190 152Z" fill="#b45309"/>
<path d="M58 152 L120 118 L126 121 L66 154Z" fill="#92400e"/>
<rect x="68" y="150" width="112" height="50" fill="url(#wall)"/>
<path d="M68 160 H180 M68 170 H180 M68 180 H180 M68 190 H180" stroke="#e7c58b" stroke-width="1.2"/>
<rect x="111" y="166" width="20" height="34" fill="#dc2626"/><rect x="111" y="166" width="20" height="34" fill="none" stroke="#fff" stroke-width="2"/>
<rect x="80" y="162" width="20" height="18" fill="#1e3a8a"/><rect x="80" y="162" width="20" height="18" fill="none" stroke="#fff" stroke-width="2"/>
<rect x="146" y="162" width="20" height="18" fill="#1e3a8a"/><rect x="146" y="162" width="20" height="18" fill="none" stroke="#fff" stroke-width="2"/>
<path d="M60 200 H194" stroke="#a16207" stroke-width="3"/>
<rect x="186" y="150" width="4" height="52" fill="#fff"/><rect x="58" y="150" width="4" height="52" fill="#fff"/>
</g>
<g stroke="#92400e" stroke-width="2.4"><path d="M330 204 V226 M352 204 V226 M374 204 V226 M396 204 V226 M418 204 V226 M440 204 V226 M462 204 V226"/><path d="M322 210 H476 M322 220 H476" stroke-width="1.8"/></g>
<g fill="#16a34a"><rect x="24" y="200" width="9" height="46" rx="4.5"/><rect x="13" y="214" width="8" height="18" rx="4"/><rect x="16" y="222" width="12" height="7" rx="3.5"/><rect x="35" y="208" width="8" height="16" rx="4"/><rect x="30" y="218" width="12" height="7" rx="3.5"/>
<rect x="430" y="232" width="8" height="34" rx="4"/><rect x="420" y="242" width="7" height="14" rx="3.5"/><rect x="423" y="250" width="10" height="6" rx="3"/></g>
<g fill="#15803d" opacity=".9"><ellipse cx="190" cy="232" rx="18" ry="7"/><ellipse cx="290" cy="250" rx="22" ry="8"/><ellipse cx="90" cy="248" rx="26" ry="9"/></g>
${flowers}
</svg>`
})()

const neon = (() => {
  const r = rng(3)
  let stars = ''
  for (let i = 0; i < 50; i++) stars += `<circle cx="${(r() * 480).toFixed(1)}" cy="${(r() * 120).toFixed(1)}" r="${(0.5 + r() * 1.1).toFixed(2)}" fill="#fff" opacity="${(0.3 + r() * 0.7).toFixed(2)}"/>`
  const colors = ['#22d3ee', '#f472b6', '#fde047', '#a78bfa', '#34d399']
  let buildings = ''
  let x = 0
  let bi = 0
  while (x < 480) {
    const w = 34 + Math.floor(r() * 40)
    const h = 70 + Math.floor(r() * 110)
    const y = 200 - h
    const fill = ['#1e1b4b', '#271a54', '#312e81', '#1c1638'][bi % 4]
    buildings += `<rect x="${x}" y="${y}" width="${w}" height="${h + 6}" fill="${fill}"/>`
    for (let wy = y + 8; wy < 194; wy += 11) {
      for (let wx = x + 5; wx < x + w - 6; wx += 9) {
        if (r() < 0.46) buildings += `<rect x="${wx}" y="${wy}" width="5" height="6" fill="${colors[Math.floor(r() * colors.length)]}" opacity="${(0.55 + r() * 0.45).toFixed(2)}"/>`
      }
    }
    x += w + 2
    bi++
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 270">
<defs>
<linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#090c2c"/><stop offset=".55" stop-color="#3b0764"/><stop offset="1" stop-color="#db2777"/></linearGradient>
<radialGradient id="moon" cx=".18" cy=".2" r=".2"><stop offset="0" stop-color="#fef3c7" stop-opacity=".9"/><stop offset="1" stop-color="#fef3c7" stop-opacity="0"/></radialGradient>
<linearGradient id="street" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#1e1036"/><stop offset="1" stop-color="#05040d"/></linearGradient>
<filter id="glow" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur stdDeviation="3.2" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
</defs>
<rect width="480" height="270" fill="url(#sky)"/>
${stars}
<rect width="480" height="270" fill="url(#moon)"/>
<circle cx="86" cy="52" r="17" fill="#fef9e7"/>
${buildings}
<g filter="url(#glow)" fill="none" stroke-width="3.2">
<rect x="70" y="120" width="64" height="22" rx="6" stroke="#22d3ee"/>
<rect x="262" y="96" width="78" height="24" rx="7" stroke="#f472b6"/>
<path d="M392 132 h46" stroke="#fde047"/><path d="M180 150 v30" stroke="#a78bfa"/>
</g>
<rect y="200" width="480" height="70" fill="url(#street)"/>
<g opacity=".55"><rect x="70" y="204" width="64" height="40" fill="#22d3ee" opacity=".22"/><rect x="262" y="204" width="78" height="52" fill="#f472b6" opacity=".2"/><rect x="392" y="204" width="46" height="34" fill="#fde047" opacity=".16"/></g>
<path d="M0 238 H480" stroke="#f472b6" stroke-opacity=".35" stroke-dasharray="18 14" stroke-width="2"/>
<g><rect x="30" y="160" width="3" height="44" fill="#4b5563"/><circle cx="31.5" cy="158" r="6" fill="#fde68a" filter="url(#glow)"/>
<rect x="210" y="160" width="3" height="44" fill="#4b5563"/><circle cx="211.5" cy="158" r="6" fill="#fde68a" filter="url(#glow)"/>
<rect x="440" y="160" width="3" height="44" fill="#4b5563"/><circle cx="441.5" cy="158" r="6" fill="#fde68a" filter="url(#glow)"/></g>
</svg>`
})()

const forest = (() => {
  const r = rng(11)
  let trees = ''
  for (let i = 0; i < 18; i++) {
    const x = (i * 29 + r() * 18) % 480
    const h = 34 + r() * 30
    const y = 186 - r() * 14
    const c = ['#14532d', '#166534', '#15803d'][i % 3]
    trees += `<path d="M${x.toFixed(1)} ${(y - h).toFixed(1)} L${(x + h * 0.32).toFixed(1)} ${y.toFixed(1)} L${(x - h * 0.32).toFixed(1)} ${y.toFixed(1)}Z" fill="${c}"/><rect x="${(x - 1.5).toFixed(1)}" y="${y.toFixed(1)}" width="3" height="6" fill="#422006"/>`
  }
  let flowers = ''
  for (let i = 0; i < 34; i++) {
    const c = ['#facc15', '#f43f5e', '#a855f7', '#fb923c', '#f8fafc'][i % 5]
    flowers += `<circle cx="${(r() * 480).toFixed(1)}" cy="${(236 + r() * 32).toFixed(1)}" r="${(1.6 + r() * 2).toFixed(1)}" fill="${c}"/>`
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 270">
<defs>
<linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#0284c7"/><stop offset=".6" stop-color="#7dd3fc"/><stop offset="1" stop-color="#e0f2fe"/></linearGradient>
<linearGradient id="mt" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#64748b"/><stop offset="1" stop-color="#334155"/></linearGradient>
<linearGradient id="hill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#4ade80"/><stop offset="1" stop-color="#15803d"/></linearGradient>
<linearGradient id="lake" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#38bdf8"/><stop offset="1" stop-color="#0369a1"/></linearGradient>
<radialGradient id="sun" cx=".2" cy=".18" r=".35"><stop offset="0" stop-color="#fffbeb"/><stop offset=".3" stop-color="#fef3c7" stop-opacity=".7"/><stop offset="1" stop-color="#fef3c7" stop-opacity="0"/></radialGradient>
</defs>
<rect width="480" height="270" fill="url(#sky)"/>
<rect width="480" height="270" fill="url(#sun)"/>
<path d="M0 170 L60 108 L96 132 L150 70 L204 128 L240 104 L300 156 L340 96 L398 140 L440 112 L480 140 V200 H0Z" fill="url(#mt)"/>
<path d="M150 70 L168 89 L158 92 L150 84 L141 94 L133 88Z M340 96 L356 113 L346 114 L340 108 L333 116 L326 111Z M60 108 L73 121 L64 122 L59 118 L52 124Z" fill="#f8fafc"/>
<path d="M0 196 Q90 160 180 186 T360 178 T480 186 V270 H0Z" fill="url(#hill)"/>
${trees}
<path d="M150 214 Q250 198 360 214 Q300 236 210 232 Q160 228 150 214Z" fill="url(#lake)"/>
<path d="M190 218 H240 M262 222 H320 M210 226 H260" stroke="#e0f2fe" stroke-opacity=".7" stroke-width="1.6"/>
<g><rect x="372" y="186" width="44" height="28" fill="#dc2626"/><path d="M366 188 L394 166 L422 188Z" fill="#7f1d1d"/><rect x="389" y="198" width="10" height="16" fill="#fef3c7"/><rect x="377" y="192" width="8" height="8" fill="#fde68a"/></g>
<path d="M0 232 Q120 222 240 236 T480 230 V270 H0Z" fill="#16a34a"/>
${flowers}
</svg>`
})()

const toUrl = (svg: string) => 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg)

export const SCENES: Record<string, string> = {
  canyon: toUrl(canyon),
  neon: toUrl(neon),
  forest: toUrl(forest),
}

export const SCENE_IDS = ['canyon', 'neon', 'forest'] as const
