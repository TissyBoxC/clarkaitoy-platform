import { readFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

const appVersion = readBuildVersion()

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
  ],
  define: {
    'import.meta.env.VITE_APP_VERSION': JSON.stringify(appVersion),
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})

function readBuildVersion(): string {
  try {
    return readFileSync(
      fileURLToPath(new URL('../../VERSION', import.meta.url)),
      'utf8',
    ).trim()
  } catch {
    // 独立复制管理端目录时仍可构建，发布流程会提供根目录 VERSION。
    return '0.0.0'
  }
}
