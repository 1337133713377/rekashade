<script setup lang="ts" generic="T extends string">
import { ToggleGroupItem, ToggleGroupRoot } from 'reka-ui'

defineProps<{ modelValue: T; options: { value: T; label: string }[]; size?: 'sm' | 'md' }>()
const emit = defineEmits<{ 'update:modelValue': [v: T] }>()
function onUpdate(v: unknown) {
  if (typeof v === 'string' && v) emit('update:modelValue', v as T)
}
</script>

<template>
  <ToggleGroupRoot type="single" class="seg" :class="size" :model-value="modelValue" @update:model-value="onUpdate">
    <ToggleGroupItem v-for="o in options" :key="o.value" :value="o.value" class="seg-item">{{ o.label }}</ToggleGroupItem>
  </ToggleGroupRoot>
</template>

<style scoped>
.seg { display: inline-flex; gap: 2px; padding: 3px; border-radius: 999px; background: rgba(0, 0, 0, 0.25); border: 1px solid var(--field-border); }
.seg-item {
  height: 32px; padding: 0 15px; border: 0; border-radius: 999px; background: transparent; color: var(--text-3);
  font-weight: 700; font-size: 13.5px; cursor: pointer; white-space: nowrap; transition: background var(--t-fast), color var(--t-fast);
}
.seg.sm .seg-item { height: 28px; padding: 0 12px; font-size: 12.5px; }
.seg-item:hover { color: var(--text); }
.seg-item[data-state='on'] { background: rgba(255, 255, 255, 0.12); color: var(--text); box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08); }
</style>
