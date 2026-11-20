<script setup lang="ts">
defineProps<{ title: string; hint?: string; tag?: string }>()
</script>

<template>
  <div class="srow">
    <div class="srow-text">
      <div class="srow-title">{{ title }} <span v-if="tag" class="chip">{{ tag }}</span></div>
      <div v-if="hint" class="srow-hint">{{ hint }}</div>
      <slot name="below" />
    </div>
    <div class="srow-ctrl"><slot /></div>
  </div>
</template>

<style scoped>
.srow { display: flex; align-items: center; gap: 18px; padding: 14px 0; border-bottom: 1px solid var(--panel-border); }
.srow:last-child { border-bottom: 0; }
.srow-text { flex: 1; min-width: 0; }
.srow-title { display: flex; align-items: center; gap: 8px; font-weight: 680; font-size: 14.5px; }
.srow-hint { margin-top: 3px; font-size: 12.5px; color: var(--text-3); line-height: 1.45; }
.srow-ctrl { flex: none; display: flex; align-items: center; gap: 10px; }
</style>
