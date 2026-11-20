<script setup lang="ts">
// Shows the crosshair/reticle at its true on-screen pixel size over a scene.
import { onMounted, ref, watch } from 'vue'
import { SCENES } from '@/lib/scenes'
import { renderCrosshair, renderReticle } from '@/lib/crosshair'
import type { CrosshairCfg, ReticleCfg } from '@/lib/types'

const props = withDefaults(defineProps<{ crosshair?: CrosshairCfg; reticle?: ReticleCfg; scene?: string; label?: string }>(), { scene: 'canyon', label: '' })
const cv = ref<HTMLCanvasElement>()
const size = ref({ w: 0, h: 0 })
let seq = 0

async function draw() {
  const id = ++seq
  const r = props.reticle ? await renderReticle(props.reticle) : props.crosshair ? renderCrosshair(props.crosshair) : null
  if (!r || id !== seq || !cv.value) return
  const c = cv.value
  c.width = r.canvas.width
  c.height = r.canvas.height
  const g = c.getContext('2d')!
  g.clearRect(0, 0, c.width, c.height)
  g.drawImage(r.canvas, 0, 0)
  const dpr = window.devicePixelRatio || 1
  size.value = { w: r.canvas.width / dpr, h: r.canvas.height / dpr }
}

watch(() => [JSON.stringify(props.crosshair), JSON.stringify(props.reticle)], draw)
onMounted(draw)
</script>

<template>
  <div class="op">
    <img :src="SCENES[scene] ?? SCENES.canyon" alt="" draggable="false" />
    <canvas ref="cv" class="cv" :style="{ width: size.w + 'px', height: size.h + 'px' }" />
    <span v-if="label" class="tag">{{ label }}</span>
  </div>
</template>

<style scoped>
.op { position: relative; width: 100%; aspect-ratio: 16 / 6.4; border-radius: 16px; overflow: hidden; background: #000; box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06); }
img { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; }
.cv { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); image-rendering: pixelated; }
.tag { position: absolute; left: 10px; top: 10px; height: 25px; padding: 0 11px; display: inline-flex; align-items: center; border-radius: 999px; background: rgba(10, 8, 20, 0.75); font-size: 11.5px; font-weight: 800; letter-spacing: 0.06em; }
</style>
