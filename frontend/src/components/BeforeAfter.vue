<script setup lang="ts">
import { ref } from 'vue'
import { ChevronsLeftRight } from '@lucide/vue'
import LookFilter from './LookFilter.vue'
import { t } from '@/lib/i18n'
import type { ColorParams } from '@/lib/types'

const props = defineProps<{ src: string; params: ColorParams; uid: string; active?: boolean }>()
const pos = ref(36)
const el = ref<HTMLElement>()
let dragging = false

function setFrom(e: PointerEvent) {
  const r = el.value!.getBoundingClientRect()
  pos.value = Math.min(98, Math.max(2, ((e.clientX - r.left) / r.width) * 100))
}
function down(e: PointerEvent) {
  dragging = true
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
  setFrom(e)
}
function move(e: PointerEvent) { if (dragging) setFrom(e) }
function up() { dragging = false }
function key(e: KeyboardEvent) {
  if (e.key === 'ArrowLeft') pos.value = Math.max(2, pos.value - 4)
  if (e.key === 'ArrowRight') pos.value = Math.min(98, pos.value + 4)
}
const fid = `ba-${props.uid}`
</script>

<template>
  <div ref="el" class="ba" tabindex="0" role="slider" :aria-valuenow="Math.round(pos)" aria-valuemin="0" aria-valuemax="100"
    @pointerdown="down" @pointermove="move" @pointerup="up" @pointercancel="up" @keydown="key">
    <LookFilter :id="fid" :params="params" />
    <img :src="src" class="layer" alt="" draggable="false" />
    <img :src="src" class="layer after" alt="" draggable="false" :style="{ filter: `url(#${fid})`, clipPath: `inset(0 0 0 ${pos}%)` }" />
    <div class="line" :style="{ left: pos + '%' }">
      <div class="knob"><ChevronsLeftRight :size="18" :stroke-width="2.4" /></div>
    </div>
    <span class="tag tb">{{ t('display.before') }}</span>
    <span class="tag ta" :class="{ on: active }">{{ t('display.after') }}</span>
  </div>
</template>

<style scoped>
.ba { position: relative; width: 100%; aspect-ratio: 16 / 9; border-radius: 16px; overflow: hidden; cursor: ew-resize; touch-action: none; background: #000; box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06); }
.layer { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; display: block; }
.line { position: absolute; top: 0; bottom: 0; width: 2px; margin-left: -1px; background: rgba(255, 255, 255, 0.95); box-shadow: 0 0 12px rgba(0, 0, 0, 0.4); }
.knob { position: absolute; top: 50%; left: 50%; width: 38px; height: 38px; margin: -19px 0 0 -19px; display: grid; place-items: center; border-radius: 50%; background: #fff; color: #18122b; box-shadow: 0 6px 18px rgba(0, 0, 0, 0.45); }
.tag { position: absolute; top: 12px; height: 26px; padding: 0 12px; display: inline-flex; align-items: center; border-radius: 999px; font-size: 12px; font-weight: 800; letter-spacing: 0.06em; }
.tb { left: 12px; background: rgba(10, 8, 20, 0.72); color: #fff; }
.ta { right: 12px; background: rgba(10, 8, 20, 0.72); color: #fff; }
.ta.on { background: linear-gradient(135deg, var(--a2), var(--a3)); }
</style>
