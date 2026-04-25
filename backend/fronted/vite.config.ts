import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'

export default defineConfig(({ mode }) => ({
  plugins: [vue(), vueJsx()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '^/(message|base|user|post|shop|video|jwt|upload)': {
        target: 'http://127.0.0.1:22001',
        changeOrigin: true
      },
      '^/(images|videos)': {
        target: 'http://127.0.0.1:22001',
        changeOrigin: true
      }
    }
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify(mode === 'production' ? 'production' : 'development')
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
}))
