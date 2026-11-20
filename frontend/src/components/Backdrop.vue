<template>
  <div class="bd" aria-hidden="true">
    <div class="glow g1" />
    <div class="glow g2" />
    <div class="glow g3" />
    <div class="streak" />
    <div class="streak s2" />
    <div class="stars" />
    <div class="vignette" />
  </div>
</template>

<style scoped>
.bd { position: fixed; inset: 0; z-index: 0; overflow: hidden; background: radial-gradient(120% 90% at 50% 0%, var(--bg-2), var(--bg) 60%); pointer-events: none; }
.glow { position: absolute; border-radius: 50%; filter: blur(90px); opacity: 0.55; }
.g1 { width: 620px; height: 520px; left: 34%; top: -180px; background: radial-gradient(circle, var(--a3), transparent 65%); }
.g2 { width: 520px; height: 520px; right: -160px; bottom: -200px; background: radial-gradient(circle, var(--a2), transparent 65%); opacity: 0.32; }
.g3 { width: 460px; height: 380px; left: -140px; bottom: 10%; background: radial-gradient(circle, #3b1d7a, transparent 70%); opacity: 0.45; }
.streak {
  position: absolute; left: 48%; top: -12%; width: 190px; height: 130%; transform: rotate(24deg);
  background: linear-gradient(180deg, transparent, color-mix(in srgb, var(--a2) 55%, transparent) 28%, color-mix(in srgb, var(--a1) 40%, transparent) 46%, transparent 78%);
  filter: blur(46px); opacity: 0.55;
}
.streak.s2 { left: 56%; width: 60px; filter: blur(18px); opacity: 0.35; }
.stars {
  position: absolute; inset: 0; opacity: 0.55;
  background-image:
    radial-gradient(1px 1px at 12% 22%, #fff8, transparent), radial-gradient(1px 1px at 72% 12%, #fff6, transparent),
    radial-gradient(1px 1px at 86% 58%, #fff5, transparent), radial-gradient(1.5px 1.5px at 34% 76%, #fff6, transparent),
    radial-gradient(1px 1px at 58% 44%, #fff4, transparent), radial-gradient(1px 1px at 22% 92%, #fff5, transparent),
    radial-gradient(1px 1px at 94% 86%, #fff4, transparent), radial-gradient(1.5px 1.5px at 64% 88%, #fff5, transparent);
}
.vignette { position: absolute; inset: 0; background: radial-gradient(120% 100% at 50% 40%, transparent 55%, rgba(0, 0, 0, 0.45)); }
/* No :global() here: in a scoped block, ":global(X) .glow" compiles to just
   "X", which put these rules on <html> itself (Midnight dimmed the whole app
   to 25%, OLED hid it). A plain ancestor selector stays scoped to .glow etc. */
:root[data-backdrop='midnight'] .glow { opacity: 0.32; }
:root[data-backdrop='midnight'] .streak { opacity: 0.25; }
:root[data-backdrop='oled'] .glow,
:root[data-backdrop='oled'] .streak,
:root[data-backdrop='oled'] .stars { display: none; }
:root[data-backdrop='nebula'] .g2 { opacity: 0.5; background: radial-gradient(circle, #db2777, transparent 65%); }
:root[data-backdrop='nebula'] .g3 { opacity: 0.6; background: radial-gradient(circle, #1d4ed8, transparent 70%); }
</style>
