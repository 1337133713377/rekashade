<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { MoreHorizontal, Plus, Download, Copy, CopyPlus, Pencil, Trash2, Pin, PinOff, Gamepad2, X, Check } from '@lucide/vue'
import { DropdownMenuContent, DropdownMenuItem, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuSeparator, DropdownMenuTrigger } from 'reka-ui'
import PageHeader from '@/components/PageHeader.vue'
import PresetThumb from '@/components/PresetThumb.vue'
import RSwitch from '@/components/RSwitch.vue'
import RSelect from '@/components/RSelect.vue'
import { api } from '@/lib/api'
import { decodePreset, presetCode } from '@/lib/codes'
import { t } from '@/lib/i18n'
import { activePreset, applyPreset, ask, confirmAsk, copyText, currentDisplay, currentParams, editableTuning, newId, readClipboard, state, toast } from '@/lib/store'
import type { Preset } from '@/lib/types'

const builtins = computed(() => state.cfg.presets.filter((p) => p.builtin))
const mine = computed(() => state.cfg.presets.filter((p) => !p.builtin))
const activeId = computed(() => (currentDisplay.value ? activePreset(currentDisplay.value.id)?.id : undefined))
const dispName = computed(() => (currentDisplay.value ? t('common.display', { n: currentDisplay.value.number }) : ''))

function use(p: Preset) {
  applyPreset(p)
  state.cfg.tuning.enabled = true
}

async function saveCurrent() {
  const n = mine.value.length + 1
  const name = await ask({ title: t('presets.namePrompt'), value: t('presets.newName', { n }), confirmLabel: t('common.save') })
  if (!name?.trim()) return
  const p: Preset = { id: newId('my'), name: name.trim().slice(0, 40), tags: ['custom'], params: { ...currentParams.value } }
  state.cfg.presets.push(p)
  if (currentDisplay.value) editableTuning(currentDisplay.value.id).presetId = p.id
  toast('success', t('presets.saved', { name: p.name }))
}

async function importCode() {
  const clip = await readClipboard()
  const code = await ask({ title: t('presets.importCode'), placeholder: t('presets.pasteCode'), value: clip.startsWith('RKP1-') ? clip : '', confirmLabel: t('common.import') })
  if (!code) return
  const d = decodePreset(code)
  if (!d) { toast('error', t('presets.invalidCode')); return }
  state.cfg.presets.push({ id: newId('shared'), name: d.name, tags: ['shared'], params: d.params })
  toast('success', t('presets.imported', { name: d.name }))
}

function duplicate(p: Preset) {
  const copy: Preset = { id: newId('my'), name: `${p.name} 2`.slice(0, 40), tags: ['custom'], params: { ...p.params } }
  const i = state.cfg.presets.findIndex((x) => x.id === p.id)
  state.cfg.presets.splice(i + 1, 0, copy)
}
async function rename(p: Preset) {
  const name = await ask({ title: t('presets.rename'), value: p.name, confirmLabel: t('common.save') })
  if (name?.trim()) p.name = name.trim().slice(0, 40)
}
async function remove(p: Preset) {
  if (!(await confirmAsk({ title: t('presets.deleteConfirm', { name: p.name }), confirmLabel: t('common.delete'), danger: true }))) return
  state.cfg.presets = state.cfg.presets.filter((x) => x.id !== p.id)
  state.cfg.ui.pinned = state.cfg.ui.pinned.filter((x) => x !== p.id)
  state.cfg.rules = state.cfg.rules.filter((r) => r.presetId !== p.id)
}
function togglePin(p: Preset) {
  const pins = state.cfg.ui.pinned
  if (pins.includes(p.id)) state.cfg.ui.pinned = pins.filter((x) => x !== p.id)
  else state.cfg.ui.pinned = [p.id, ...pins].slice(0, 4)
}

// ---- game profiles
const running = ref<string[]>([])
const pickExe = ref('')
const typed = ref('')
const presetOptions = computed(() => state.cfg.presets.map((p) => ({ value: p.id, label: p.name })))
const runningOptions = computed(() => running.value.filter((e) => !state.cfg.rules.some((r) => r.exe === e)).map((e) => ({ value: e, label: e })))
async function loadRunning() { try { running.value = await api.runningApps() } catch { running.value = [] } }
onMounted(loadRunning)

function normalize(s: string) {
  let v = s.trim().toLowerCase().split(/[\\/]/).pop() ?? ''
  if (v && !v.endsWith('.exe')) v += '.exe'
  return v
}
function addRule(exe: string) {
  const e = normalize(exe)
  if (!e || state.cfg.rules.some((r) => r.exe === e)) return
  state.cfg.rules.push({ exe: e, presetId: activeId.value ?? 'ultra-vivid', enabled: true })
  pickExe.value = ''
  typed.value = ''
}
function tagLabel(tag: string) {
  const v = t('presets.tags.' + tag)
  return v.startsWith('presets.') ? tag : v
}
</script>

<template>
  <div class="page">
    <PageHeader :title="t('presets.title')" :subtitle="t('presets.subtitle')">
      <button class="btn" @click="importCode"><Download :size="16" /> {{ t('presets.importCode') }}</button>
      <button class="btn primary" @click="saveCurrent"><Plus :size="16" /> {{ t('presets.saveCurrent') }}</button>
    </PageHeader>

    <div class="layout">
      <div class="lists">
        <section v-for="group in [{ key: 'mine', list: mine }, { key: 'builtin', list: builtins }]" :key="group.key" class="glass block">
          <div class="row"><h2 class="card-title">{{ t('presets.' + group.key) }}</h2><span class="chip">{{ group.list.length }}</span></div>
          <p v-if="!group.list.length" class="dim small">{{ t('presets.emptyMine') }}</p>
          <div class="cards">
            <article v-for="p in group.list" :key="p.id" class="pc" :class="{ active: state.cfg.tuning.enabled && activeId === p.id }">
              <button class="pc-thumb" @click="use(p)"><PresetThumb :params="p.params" :uid="'pp-' + p.id" :scene="state.cfg.ui.scene" /></button>
              <div class="pc-body">
                <div class="row">
                  <span class="pc-name">{{ p.name }}</span>
                  <Pin v-if="state.cfg.ui.pinned.includes(p.id)" :size="13" class="dim" />
                  <div class="spacer" />
                  <DropdownMenuRoot>
                    <DropdownMenuTrigger class="btn ghost icon sm" :aria-label="p.name"><MoreHorizontal :size="17" /></DropdownMenuTrigger>
                    <DropdownMenuPortal>
                      <DropdownMenuContent class="menu" align="end" :side-offset="4">
                        <DropdownMenuItem class="menu-item" @select="copyText(presetCode(p))"><Copy :size="15" /> {{ t('presets.copyCode') }}</DropdownMenuItem>
                        <DropdownMenuItem class="menu-item" @select="duplicate(p)"><CopyPlus :size="15" /> {{ t('presets.duplicate') }}</DropdownMenuItem>
                        <DropdownMenuItem class="menu-item" @select="togglePin(p)">
                          <component :is="state.cfg.ui.pinned.includes(p.id) ? PinOff : Pin" :size="15" />
                          {{ state.cfg.ui.pinned.includes(p.id) ? t('presets.unpin') : t('presets.pin') }}
                        </DropdownMenuItem>
                        <template v-if="!p.builtin">
                          <DropdownMenuSeparator class="menu-sep" />
                          <DropdownMenuItem class="menu-item" @select="rename(p)"><Pencil :size="15" /> {{ t('presets.rename') }}</DropdownMenuItem>
                          <DropdownMenuItem class="menu-item danger" @select="remove(p)"><Trash2 :size="15" /> {{ t('presets.delete') }}</DropdownMenuItem>
                        </template>
                      </DropdownMenuContent>
                    </DropdownMenuPortal>
                  </DropdownMenuRoot>
                </div>
                <div class="row tags">
                  <span v-for="tag in p.tags ?? []" :key="tag" class="chip">{{ tagLabel(tag) }}</span>
                </div>
                <button class="btn sm block" :class="state.cfg.tuning.enabled && activeId === p.id ? 'soft' : ''" @click="use(p)">
                  <Check v-if="state.cfg.tuning.enabled && activeId === p.id" :size="15" />
                  {{ state.cfg.tuning.enabled && activeId === p.id ? t('presets.active') : t('presets.applyTo', { display: dispName }) }}
                </button>
              </div>
            </article>
          </div>
        </section>
      </div>

      <aside class="glass block rules">
        <div class="row"><Gamepad2 :size="20" class="acc" /><h2 class="card-title">{{ t('presets.rules') }}</h2></div>
        <p class="dim small">{{ t('presets.rulesHint') }}</p>
        <p v-if="!state.cfg.tuning.enabled" class="chip warn">{{ t('presets.rulesNeedTuning') }}</p>
        <div v-if="!state.cfg.rules.length" class="empty dim small">{{ t('presets.noRules') }}</div>
        <div v-for="(r, i) in state.cfg.rules" :key="r.exe" class="rule">
          <div class="row">
            <span class="exe mono" :title="r.exe">{{ r.exe }}</span>
            <div class="spacer" />
            <RSwitch v-model="r.enabled" size="sm" :label="r.exe" />
            <button class="btn ghost icon sm" :aria-label="t('common.remove')" @click="state.cfg.rules.splice(i, 1)"><X :size="16" /></button>
          </div>
          <RSelect v-model="r.presetId" :options="presetOptions" />
        </div>
        <div class="divider" />
        <div class="label-caps">{{ t('presets.addRule') }}</div>
        <div class="row">
          <RSelect v-model="pickExe" class="grow" :options="runningOptions" :placeholder="t('presets.pickRunning')" @update:model-value="addRule" />
        </div>
        <form class="row" @submit.prevent="addRule(typed)">
          <input v-model="typed" class="input" :placeholder="t('presets.typeExe')" spellcheck="false" />
          <button class="btn icon" type="submit" :aria-label="t('common.add')"><Plus :size="17" /></button>
        </form>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.layout { display: grid; grid-template-columns: 1fr 300px; gap: 14px; align-items: start; }
.lists { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
.block { padding: 18px 20px 20px; display: flex; flex-direction: column; gap: 12px; }
.cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(168px, 1fr)); gap: 12px; }
.pc { display: flex; flex-direction: column; border-radius: 18px; border: 1px solid var(--field-border); background: rgba(0, 0, 0, 0.18); overflow: hidden; transition: border-color var(--t-fast), box-shadow var(--t-fast); }
.pc:hover { border-color: rgba(255, 255, 255, 0.16); }
.pc.active { border-color: var(--a2); box-shadow: 0 0 0 1px var(--a2), 0 12px 28px -14px var(--a-glow); }
.pc-thumb { padding: 6px 6px 0; border: 0; background: none; cursor: pointer; }
.pc-body { display: flex; flex-direction: column; gap: 8px; padding: 8px 10px 10px; }
.pc-name { font-weight: 750; font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tags { gap: 5px; flex-wrap: wrap; min-height: 24px; }
.tags .chip { height: 20px; font-size: 10.5px; padding: 0 8px; }
.rules { position: sticky; top: 0; gap: 12px; }
.acc { color: var(--a1); }
.empty { padding: 14px; border-radius: 12px; border: 1px dashed var(--field-border); text-align: center; }
.rule { display: flex; flex-direction: column; gap: 8px; padding: 10px; border-radius: 14px; background: rgba(0, 0, 0, 0.2); border: 1px solid var(--field-border); }
.rule :deep(.sel) { width: 100%; }
.exe { font-size: 12.5px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 150px; }
.grow :deep(.sel), .row > :deep(.sel) { width: 100%; }
</style>

<style>
.menu { z-index: 80; min-width: 200px; padding: 5px; border-radius: 14px; background: #1b1532; border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: 0 20px 50px -12px rgba(0, 0, 0, 0.75); font-family: var(--font); color: var(--text); }
.menu-item { display: flex; align-items: center; gap: 10px; height: 36px; padding: 0 10px; border-radius: 9px; font-size: 13.5px; font-weight: 600; cursor: pointer; outline: none; }
.menu-item[data-highlighted] { background: rgba(255, 255, 255, 0.08); }
.menu-item.danger { color: #fca5a5; }
.menu-sep { height: 1px; margin: 4px 6px; background: rgba(255, 255, 255, 0.08); }
</style>
