import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv, type ConfigEnv } from 'vite'

export default defineConfig(({ mode }: ConfigEnv) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiTarget = env.VITE_API_BASE_URL || 'http://localhost:8080'
  const srcPath = fileURLToPath(new URL('./src', import.meta.url))

  return {
    plugins: [vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag: string) => tag.startsWith('cap-'),
        },
      },
    })],
    resolve: {
      alias: {
        '@': path.resolve(srcPath),
      },
    },
    build: {
      outDir: 'dist',
      emptyOutDir: true,
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
        },
        '/.well-known': {
          target: apiTarget,
          changeOrigin: true,
        },
        '/oauth': {
          target: apiTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
