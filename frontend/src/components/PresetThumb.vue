<script setup lang="ts">
import LookFilter from './LookFilter.vue'
import { SCENES } from '@/lib/scenes'
import type { ColorParams } from '@/lib/types'

const props = withDefaults(defineProps<{ params: ColorParams; uid: string; scene?: string }>(), { scene: 'canyon' })
const fid = `pt-${props.uid}`
</script>

<template>
  <div class="pt">
    <LookFilter :id="fid" :params="params" :samples="24" />
    <img :src="SCENES[scene] ?? SCENES.canyon" alt="" draggable="false" :style="{ filter: `url(#${fid})` }" />
  </div>
</template>

<style scoped>
.pt { position: relative; width: 100%; aspect-ratio: 16 / 9; border-radius: 12px; overflow: hidden; background: #000; }
img { width: 100%; height: 100%; object-fit: cover; display: block; }
</style>
