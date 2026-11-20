<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { TooltipProvider } from 'reka-ui'
import Backdrop from '@/components/Backdrop.vue'
import TitleBar from '@/components/TitleBar.vue'
import Sidebar from '@/components/Sidebar.vue'
import Toasts from '@/components/Toasts.vue'
import DialogHost from '@/components/DialogHost.vue'
import ResolutionConfirm from '@/components/ResolutionConfirm.vue'
import DisplayPage from '@/pages/DisplayPage.vue'
import PresetsPage from '@/pages/PresetsPage.vue'
import ResolutionPage from '@/pages/ResolutionPage.vue'
import CrosshairPage from '@/pages/CrosshairPage.vue'
import ReticlePage from '@/pages/ReticlePage.vue'
import BindsPage from '@/pages/BindsPage.vue'
import OptimizerPage from '@/pages/OptimizerPage.vue'
import ThemePage from '@/pages/ThemePage.vue'
import CommunityPage from '@/pages/CommunityPage.vue'
import AccountPage from '@/pages/AccountPage.vue'
import SettingsPage from '@/pages/SettingsPage.vue'
import AboutPage from '@/pages/AboutPage.vue'
import LegalPage from '@/pages/LegalPage.vue'
import { boot, state } from '@/lib/store'
import { errText } from '@/lib/api'

const pages: Record<string, Component> = {
  display: DisplayPage, presets: PresetsPage, resolution: ResolutionPage, crosshair: CrosshairPage, reticle: ReticlePage,
  binds: BindsPage, optimizer: OptimizerPage, theme: ThemePage, community: CommunityPage, account: AccountPage,
  settings: SettingsPage, about: AboutPage, legal: LegalPage,
}
const page = computed(() => pages[state.page] ?? DisplayPage)
const bootError = ref('')

onMounted(async () => {
  try {
    await boot()
  } catch (e) {
    bootError.value = errText(e)
  }
})
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <div class="app">
      <Backdrop />
      <TitleBar v-if="state.ready" />
      <div v-if="state.ready" class="body">
        <Sidebar />
        <main class="main">
          <Transition name="fade" mode="out-in">
            <component :is="page" :key="state.page" />
          </Transition>
        </main>
      </div>
      <div v-else class="boot">
        <img src="@/assets/brand/reka-mark.svg" class="boot-mark" alt="" />
        <p v-if="bootError" class="boot-err">{{ bootError }}</p>
      </div>
      <Toasts />
      <DialogHost />
      <ResolutionConfirm v-if="state.ready" />
    </div>
  </TooltipProvider>
</template>

<style scoped>
.app { position: relative; height: 100%; display: flex; flex-direction: column; }
.body { position: relative; z-index: 1; flex: 1; min-height: 0; display: flex; }
.main { flex: 1; min-width: 0; overflow-y: auto; overflow-x: hidden; padding: 14px 18px 16px 16px; }
.boot { position: relative; z-index: 1; flex: 1; display: grid; place-items: center; align-content: center; gap: 14px; }
.boot-mark { width: 84px; height: 84px; animation: pulse-dot 1.6s ease-in-out infinite; filter: drop-shadow(0 10px 30px var(--a-glow)); }
.boot-err { color: var(--bad); max-width: 520px; text-align: center; }
</style>
