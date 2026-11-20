<script setup lang="ts">
import { computed, ref } from 'vue'
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from 'reka-ui'
import { api } from '@/lib/api'
import { t } from '@/lib/i18n'
import { fail, state } from '@/lib/store'

const pending = computed(() => state.status.stretch.pending)
const open = computed(() => pending.value > 0)
const busy = ref(false)
const pct = computed(() => (pending.value / 15) * 100)

async function keep() { busy.value = true; try { await api.confirmResolution() } catch (e) { fail(e) } finally { busy.value = false } }
async function revert() { busy.value = true; try { await api.revertResolution() } catch (e) { fail(e) } finally { busy.value = false } }
</script>

<template>
  <DialogRoot :open="open">
    <DialogPortal>
      <DialogOverlay class="dlg-ov" />
      <DialogContent class="dlg rc" @escape-key-down="revert" @pointer-down-outside.prevent>
        <div class="ring" :style="{ '--p': pct + '%' }"><span class="mono">{{ pending }}</span></div>
        <DialogTitle class="dlg-title">{{ t('res.keepTitle') }}</DialogTitle>
        <DialogDescription class="dlg-body">
          {{ state.status.stretch.width }}×{{ state.status.stretch.height }} — {{ t('res.keepBody', { s: pending }) }}
        </DialogDescription>
        <div class="dlg-actions">
          <button class="btn" :disabled="busy" @click="revert">{{ t('res.revertNow') }}</button>
          <button class="btn primary" :disabled="busy" @click="keep">{{ t('res.keep') }}</button>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style scoped>
.rc { align-items: center; text-align: center; }
.rc :deep(.dlg-actions) { justify-content: center; }
.ring { width: 76px; height: 76px; border-radius: 50%; display: grid; place-items: center; background: conic-gradient(var(--a2) var(--p), rgba(255, 255, 255, 0.08) 0); position: relative; }
.ring::after { content: ''; position: absolute; inset: 6px; border-radius: 50%; background: #1a1430; }
.ring span { position: relative; z-index: 1; font-size: 24px; font-weight: 800; }
</style>
