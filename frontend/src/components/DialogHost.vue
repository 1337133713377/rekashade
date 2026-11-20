<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from 'reka-ui'
import { t } from '@/lib/i18n'
import { state } from '@/lib/store'

const open = computed(() => !!state.dialog)
const input = ref<HTMLInputElement>()
watch(open, async (o) => { if (o && state.dialog?.kind === 'prompt') { await nextTick(); input.value?.focus(); input.value?.select() } })

function done(ok: boolean) {
  const d = state.dialog
  if (!d) return
  state.dialog = null
  d.resolve(ok ? (d.kind === 'prompt' ? d.value : 'ok') : null)
}
</script>

<template>
  <DialogRoot :open="open" @update:open="(o) => { if (!o) done(false) }">
    <DialogPortal>
      <DialogOverlay class="dlg-ov" />
      <DialogContent v-if="state.dialog" class="dlg">
        <DialogTitle class="dlg-title">{{ state.dialog.title }}</DialogTitle>
        <DialogDescription v-if="state.dialog.body" class="dlg-body">{{ state.dialog.body }}</DialogDescription>
        <form v-if="state.dialog.kind === 'prompt'" @submit.prevent="done(true)">
          <input ref="input" v-model="state.dialog.value" class="input" :placeholder="state.dialog.placeholder" spellcheck="false" maxlength="2000" />
        </form>
        <div class="dlg-actions">
          <button class="btn" @click="done(false)">{{ t('common.cancel') }}</button>
          <button class="btn" :class="state.dialog.danger ? 'danger' : 'primary'" @click="done(true)">{{ state.dialog.confirmLabel || t('common.confirm') }}</button>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style>
.dlg-ov { position: fixed; inset: 0; z-index: 70; background: rgba(4, 3, 10, 0.6); backdrop-filter: blur(4px); }
.dlg {
  position: fixed; left: 50%; top: 50%; z-index: 71; width: min(460px, 92vw); transform: translate(-50%, -50%); padding: 24px;
  border-radius: 22px; background: #1a1430; border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: 0 30px 80px -20px rgba(0, 0, 0, 0.8);
  font-family: var(--font); color: var(--text); display: flex; flex-direction: column; gap: 14px;
}
.dlg-title { font-size: 19px; font-weight: 780; }
.dlg-body { color: var(--text-2); font-size: 14px; line-height: 1.5; }
.dlg-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 4px; }
</style>
