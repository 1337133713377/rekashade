<script setup lang="ts">
// Draws a crosshair at real pixel size, centred in its tile.
import { onMounted, ref, watch } from 'vue'
import { renderCrosshair } from '@/lib/crosshair'
import type { CrosshairCfg } from '@/lib/types'

const props = defineProps<{ cfg: CrosshairCfg }>()
const cv = ref<HTMLCanvasElement>()
const size = ref({ w: 0, h: 0 })
function draw() {
  const r = renderCrosshair(props.cfg)
  const c = cv.value
  if (!c) return
  c.width = r.canvas.width
  c.height = r.canvas.height
  c.getContext('2d')!.drawImage(r.canvas, 0, 0)
  const dpr = window.devicePixelRatio || 1
  size.value = { w: r.canvas.width / dpr, h: r.canvas.height / dpr }
}
watch(() => JSON.stringify(props.cfg), draw)
onMounted(draw)
</script>

<template>
  <canvas ref="cv" class="ct" :style="{ width: size.w + 'px', height: size.h + 'px' }" />
</template>

<style scoped>
.ct { image-rendering: pixelated; max-width: 100%; max-height: 100%; }
</style>
