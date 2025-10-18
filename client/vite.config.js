import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  base: '/',
  server: {
    port: 3000,
    cors: true,
    host: '0.0.0.0',
    // 添加新域名到 allowedHosts 数组
    allowedHosts: ['794jx56302cy.vicp.fun', 'ai.dayu.club'],
    proxy: {
      '/api': {
        // target: 'http://bear-ai1.natapp1.cc',
        target: 'http://101.42.249.106:8123/api',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
