import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@defaults': fileURLToPath(new URL('../internal/config/defaults.json', import.meta.url)),
    },
  },
  server: { fs: { allow: ['..'] } },
  build: {
    target: 'chrome110',
    chunkSizeWarningLimit: 900,
    assetsInlineLimit: 8192,
  },
})
