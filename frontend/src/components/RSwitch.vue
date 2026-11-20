<script setup lang="ts">
import { SwitchRoot, SwitchThumb } from 'reka-ui'

withDefaults(defineProps<{ modelValue: boolean; size?: 'lg' | 'md' | 'sm'; disabled?: boolean; label?: string }>(), { size: 'md', label: '' })
const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()
</script>

<template>
  <SwitchRoot
    class="sw" :class="size" :model-value="modelValue" :disabled="disabled" :aria-label="label"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
  >
    <SwitchThumb class="sw-thumb" />
  </SwitchRoot>
</template>

<style scoped>
.sw {
  --w: 54px; --h: 30px; --k: 24px;
  position: relative; flex: none; width: var(--w); height: var(--h); padding: 0; border-radius: 999px; cursor: pointer;
  border: 1px solid var(--field-border); background: rgba(255, 255, 255, 0.08);
  transition: background var(--t-med) var(--ease), border-color var(--t-med), box-shadow var(--t-med);
}
.sw.lg { --w: 82px; --h: 44px; --k: 36px; }
.sw.sm { --w: 42px; --h: 24px; --k: 18px; }
.sw[data-state='checked'] {
  border-color: transparent; background: linear-gradient(135deg, var(--a2), var(--a3));
  box-shadow: 0 8px 22px -8px var(--a-glow), inset 0 1px 0 rgba(255, 255, 255, 0.22);
}
.sw:disabled { opacity: 0.45; cursor: default; }
.sw-thumb {
  position: absolute; top: 50%; left: 3px; width: var(--k); height: var(--k); border-radius: 50%; background: #fff;
  transform: translateY(-50%); box-shadow: 0 3px 10px rgba(0, 0, 0, 0.4);
  transition: left var(--t-med) var(--ease);
}
.sw-thumb[data-state='checked'] { left: calc(var(--w) - var(--k) - 5px); }
</style>
