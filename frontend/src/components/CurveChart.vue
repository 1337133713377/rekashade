<script setup lang="ts">
import { computed } from 'vue'
import { curve } from '@/lib/color'
import type { ColorParams } from '@/lib/types'

const props = defineProps<{ params: ColorParams }>()
const N = 64
const paths = computed(() => [0, 1, 2].map((c) => {
  let d = ''
  for (let i = 0; i <= N; i++) {
    const x = i / N
    const y = curve(props.params, c, x)
    d += `${i ? 'L' : 'M'}${(x * 256).toFixed(1)} ${(256 - y * 256).toFixed(1)}`
  }
  return d
}))
</script>

<template>
  <div class="cc">
    <svg viewBox="-8 -8 272 272" preserveAspectRatio="none">
      <g stroke="rgba(255,255,255,.07)" stroke-width="1">
        <path v-for="i in 3" :key="'v' + i" :d="`M${i * 64} 0V256`" />
        <path v-for="i in 3" :key="'h' + i" :d="`M0 ${i * 64}H256`" />
      </g>
      <rect x="0" y="0" width="256" height="256" fill="none" stroke="rgba(255,255,255,.12)" />
      <path d="M0 256L256 0" stroke="rgba(255,255,255,.28)" stroke-dasharray="5 6" fill="none" />
      <path :d="paths[0]" stroke="#ff5a7a" stroke-width="3" fill="none" style="mix-blend-mode: screen" />
      <path :d="paths[1]" stroke="#4ade80" stroke-width="3" fill="none" style="mix-blend-mode: screen" />
      <path :d="paths[2]" stroke="#60a5fa" stroke-width="3" fill="none" style="mix-blend-mode: screen" />
    </svg>
    <div class="legend"><span class="r">R</span><span class="g">G</span><span class="b">B</span></div>
  </div>
</template>

<style scoped>
.cc { position: relative; width: 100%; aspect-ratio: 16 / 9; border-radius: 16px; overflow: hidden; background: radial-gradient(120% 120% at 30% 0%, rgba(155, 109, 255, 0.12), rgba(5, 4, 12, 0.7)); box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06); }
svg { position: absolute; inset: 14px 14px 14px 14px; width: calc(100% - 28px); height: calc(100% - 28px); }
.legend { position: absolute; right: 14px; top: 12px; display: flex; gap: 6px; font-size: 11px; font-weight: 800; }
.legend span { width: 22px; height: 22px; display: grid; place-items: center; border-radius: 7px; background: rgba(0, 0, 0, 0.4); }
.r { color: #ff7a93; } .g { color: #6ee7a0; } .b { color: #7db6ff; }
</style>
