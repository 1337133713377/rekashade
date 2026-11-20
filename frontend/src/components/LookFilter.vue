<script setup lang="ts">
// SVG filter reproducing the native pipeline: Windows colour matrix (hue +
// saturation) → GPU vibrance → hardware gamma ramp (tone curves).
import { computed } from 'vue'
import { buildRamp, colorMatrix, saturate, svgMatrix } from '@/lib/color'
import type { ColorParams } from '@/lib/types'

const props = withDefaults(defineProps<{ id: string; params: ColorParams; samples?: number }>(), { samples: 48 })
const f = computed(() => {
  const ramp = buildRamp(props.params, props.samples)
  return {
    matrix: svgMatrix(colorMatrix(props.params.hue, props.params.saturation)),
    vib: svgMatrix(saturate(props.params.vibrance / 100)),
    tables: ramp.map((ch) => ch.map((v) => v.toFixed(4)).join(' ')),
  }
})
</script>

<template>
  <svg class="lf" width="0" height="0" aria-hidden="true" focusable="false">
    <filter :id="id" color-interpolation-filters="sRGB" x="0" y="0" width="100%" height="100%">
      <feColorMatrix type="matrix" :values="f.matrix" />
      <feColorMatrix type="matrix" :values="f.vib" />
      <feComponentTransfer>
        <feFuncR type="table" :tableValues="f.tables[0]" />
        <feFuncG type="table" :tableValues="f.tables[1]" />
        <feFuncB type="table" :tableValues="f.tables[2]" />
      </feComponentTransfer>
    </filter>
  </svg>
</template>

<style scoped>
.lf { position: absolute; width: 0; height: 0; overflow: hidden; pointer-events: none; }
</style>
