<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<template>
  <section class="ph glass">
    <div class="ph-text">
      <h1>{{ title }}</h1>
      <div class="ph-sub">
        <slot name="sub">{{ subtitle }}</slot>
      </div>
    </div>
    <div class="ph-actions"><slot /></div>
  </section>
</template>

<style scoped>
.ph { display: flex; align-items: center; gap: 18px; min-height: 88px; padding: 16px 24px; flex: none; }
.ph-text { flex: 1; min-width: 0; }
h1 { font-size: 29px; font-weight: 800; letter-spacing: -0.02em; line-height: 1.15; }
.ph-sub { margin-top: 6px; display: flex; align-items: center; gap: 10px; color: var(--text-2); font-size: 14px; }
.ph-actions { display: flex; align-items: center; gap: 10px; flex: none; }
</style>
