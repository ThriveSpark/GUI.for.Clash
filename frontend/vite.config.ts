import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import Components from 'unplugin-vue-components/vite'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    port: 9245
  },
  plugins: [
    vue(),
    Components({
      types: [],
      dts: 'src/components/components.d.ts',
      globs: ['src/components/*/index.vue']
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@wails': fileURLToPath(new URL('./bindings', import.meta.url))
    }
  },
  build: {
    chunkSizeWarningLimit: 2048 // 2MB
  }
})
