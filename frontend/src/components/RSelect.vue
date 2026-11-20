<script setup lang="ts">
import { ChevronDown, Check } from '@lucide/vue'
import { SelectContent, SelectItem, SelectItemIndicator, SelectItemText, SelectPortal, SelectRoot, SelectTrigger, SelectValue, SelectViewport } from 'reka-ui'

defineProps<{ modelValue: string; options: { value: string; label: string }[]; placeholder?: string; disabled?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [v: string] }>()
</script>

<template>
  <SelectRoot :model-value="modelValue" :disabled="disabled" @update:model-value="(v) => emit('update:modelValue', String(v))">
    <SelectTrigger class="sel" :aria-label="placeholder">
      <SelectValue :placeholder="placeholder" />
      <ChevronDown :size="16" class="dim" />
    </SelectTrigger>
    <SelectPortal>
      <SelectContent class="sel-pop" position="popper" :side-offset="6">
        <SelectViewport class="sel-vp">
          <SelectItem v-for="o in options" :key="o.value" :value="o.value" class="sel-item">
            <SelectItemText>{{ o.label }}</SelectItemText>
            <SelectItemIndicator class="sel-ind"><Check :size="15" /></SelectItemIndicator>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>

<style scoped>
.sel {
  display: inline-flex; align-items: center; justify-content: space-between; gap: 10px; min-width: 170px; height: 40px; padding: 0 12px 0 14px;
  border-radius: 12px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.22); cursor: pointer; font-weight: 600; font-size: 13.5px;
}
.sel:hover { border-color: rgba(255, 255, 255, 0.16); }
</style>

<style>
.sel-pop {
  z-index: 80; min-width: var(--reka-select-trigger-width); max-height: 320px; overflow: hidden; border-radius: 14px; padding: 5px;
  background: #1b1532; border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: 0 20px 50px -12px rgba(0, 0, 0, 0.75);
  font-family: var(--font); color: var(--text);
}
.sel-vp { max-height: 300px; }
.sel-item {
  display: flex; align-items: center; justify-content: space-between; gap: 12px; height: 36px; padding: 0 10px; border-radius: 9px;
  font-size: 13.5px; font-weight: 600; cursor: pointer; outline: none;
}
.sel-item[data-highlighted] { background: rgba(255, 255, 255, 0.08); }
.sel-item[data-state='checked'] { color: var(--a-text); }
.sel-ind { display: inline-flex; }
</style>
