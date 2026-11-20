<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Trash2, Loader2 } from '@lucide/vue'
import PageHeader from '@/components/PageHeader.vue'
import RSwitch from '@/components/RSwitch.vue'
import { api } from '@/lib/api'
import { fmtBytes, t, tweakText } from '@/lib/i18n'
import { fail, toast } from '@/lib/store'
import type { TweakStatus } from '@/lib/types'

const list = ref<TweakStatus[]>([])
const loading = ref(true)
const busy = ref<string | null>(null)
const cleaning = ref(false)
const cleaned = ref('')

onMounted(async () => {
  try { list.value = await api.optimizerList() } catch (e) { fail(e) } finally { loading.value = false }
})
const visible = computed(() => list.value.filter((x) => x.supported))

async function toggle(tw: TweakStatus, on: boolean) {
  busy.value = tw.id
  try {
    const s = await api.optimizerSet(tw.id, on)
    Object.assign(tw, s)
  } catch (e) {
    toast('error', t('opt.failed', { name: tweakText(tw.id)[0], error: e instanceof Error ? e.message : String(e) }))
    try { list.value = await api.optimizerList() } catch { /* keep */ }
  } finally { busy.value = null }
}
async function all(on: boolean) {
  busy.value = 'all'
  try {
    list.value = await api.optimizerAll(on)
    toast('success', on ? t('opt.allDone') : t('opt.allUndone'))
  } catch (e) { fail(e) } finally { busy.value = null }
}
async function clean() {
  cleaning.value = true
  try {
    const r = await api.cleanTemp()
    cleaned.value = t('opt.cleaned', { files: r.files, size: fmtBytes(r.bytes), skipped: r.skipped })
  } catch (e) { fail(e) } finally { cleaning.value = false }
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('opt.title')" :subtitle="t('opt.subtitle')">
      <button class="btn lg" :disabled="!!busy" @click="all(false)">{{ t('opt.undoAll') }}</button>
      <button class="btn primary lg" :disabled="!!busy" @click="all(true)">{{ t('opt.optimizeAll') }}</button>
    </PageHeader>

    <div v-if="loading" class="loading dim"><Loader2 :size="20" class="spin" /> {{ t('common.loading') }}</div>
    <div v-else class="grid">
      <article v-for="tw in visible" :key="tw.id" class="glass tw">
        <div class="grow">
          <div class="tw-name">{{ tweakText(tw.id)[0] }}</div>
          <div class="tw-desc">{{ tweakText(tw.id)[1] }}</div>
          <span v-if="tw.on && !tw.canUndo" class="chip ok tw-chip" :title="t('opt.alreadyOnHint')">{{ t('opt.alreadyOn') }}</span>
          <span v-if="tw.error" class="chip bad tw-chip" :title="tw.error">{{ t('common.error') }}</span>
        </div>
        <RSwitch :model-value="tw.on" :disabled="busy !== null || (tw.on && !tw.canUndo)" :label="tweakText(tw.id)[0]" @update:model-value="(v: boolean) => toggle(tw, v)" />
      </article>
      <article class="glass tw wide">
        <div class="grow">
          <div class="tw-name">{{ t('opt.clean') }}</div>
          <div class="tw-desc">{{ cleaned || t('opt.cleanDesc') }}</div>
        </div>
        <button class="btn lg" :disabled="cleaning" @click="clean">
          <Loader2 v-if="cleaning" :size="16" class="spin" /><Trash2 v-else :size="16" /> {{ t('opt.cleanBtn') }}
        </button>
      </article>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.loading { display: flex; align-items: center; gap: 10px; padding: 30px; justify-content: center; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.tw { display: flex; align-items: center; gap: 16px; padding: 16px 20px; min-height: 84px; border-radius: 20px; }
.tw.wide { grid-column: 1 / -1; }
.tw-name { font-weight: 760; font-size: 15px; }
.tw-desc { margin-top: 3px; font-size: 12.5px; color: var(--text-3); line-height: 1.4; }
.tw-chip { margin-top: 7px; height: 20px; font-size: 10.5px; }
</style>
