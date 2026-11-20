<script setup lang="ts">
import { computed } from 'vue'
import { SliderRange, SliderRoot, SliderThumb, SliderTrack } from 'reka-ui'

const props = withDefaults(defineProps<{
  label: string
  modelValue: number
  min: number
  max: number
  step?: number
  def?: number
  format?: (v: number) => string
  disabled?: boolean
  chip?: string
  chipKind?: 'accent' | 'ok' | 'warn' | ''
  chipTitle?: string
}>(), { step: 1, def: undefined, format: undefined, chip: '', chipKind: '', chipTitle: '' })

const emit = defineEmits<{ 'update:modelValue': [v: number] }>()

const model = computed({
  get: () => [props.modelValue],
  set: (v: number[] | undefined) => { if (v && v.length) emit('update:modelValue', v[0]) },
})
const shown = computed(() => (props.format ? props.format(props.modelValue) : String(props.modelValue)))
function reset() {
  if (props.def !== undefined && !props.disabled) emit('update:modelValue', props.def)
}
</script>

<template>
  <div class="rs" :class="{ disabled }" @dblclick="reset">
    <div class="rs-head">
      <span class="rs-label">
        {{ label }}
        <span v-if="chip" class="chip" :class="chipKind" :title="chipTitle">{{ chip }}</span>
      </span>
      <span class="rs-value mono">{{ shown }}</span>
    </div>
    <SliderRoot v-model="model" class="rs-root" :min="min" :max="max" :step="step" :disabled="disabled" thumb-alignment="contain">
      <SliderTrack class="rs-track">
        <SliderRange class="rs-range" />
      </SliderTrack>
      <SliderThumb class="rs-thumb" :aria-label="label" />
    </SliderRoot>
  </div>
</template>

<style scoped>
.rs { display: flex; flex-direction: column; gap: 5px; }
.rs.disabled { opacity: 0.45; }
.rs-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.rs-label { display: inline-flex; align-items: center; gap: 8px; font-weight: 680; font-size: 14.5px; }
.rs-label .chip { height: 19px; padding: 0 7px; font-size: 10px; }
.rs-value { font-size: 13px; font-weight: 700; color: var(--a-text); }
.rs-root { position: relative; display: flex; align-items: center; height: 22px; touch-action: none; user-select: none; cursor: pointer; }
.rs-track { position: relative; flex: 1; height: 6px; border-radius: 6px; background: var(--track); overflow: hidden; }
.rs-range { position: absolute; height: 100%; border-radius: 6px; background: linear-gradient(90deg, var(--a3), var(--a2) 70%, var(--a1)); }
.rs-thumb {
  display: block; width: 32px; height: 21px; border-radius: 11px; background: #f6f3ff; cursor: grab;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.45), 0 0 0 1px rgba(0, 0, 0, 0.08), inset 0 -2px 0 rgba(120, 100, 170, 0.18);
  transition: transform var(--t-fast) var(--ease), box-shadow var(--t-fast);
}
.rs-thumb:hover { transform: scale(1.05); }
.rs-thumb:active { cursor: grabbing; }
.rs-thumb:focus-visible { outline: none; box-shadow: 0 0 0 4px var(--a-soft), 0 4px 12px rgba(0, 0, 0, 0.45); }
/* Windows contrast themes repaint backgrounds, which erased the track and
   thumb: draw them with the theme's own colours instead. */
@media (forced-colors: active) {
  .rs-root { forced-color-adjust: none; }
  .rs-track { background: GrayText; }
  .rs-range { background: Highlight; }
  .rs-thumb { background: ButtonText; border: 2px solid Canvas; box-shadow: none; }
  .rs-thumb:focus-visible { outline: 2px solid Highlight; outline-offset: 1px; box-shadow: none; }
}
</style>
