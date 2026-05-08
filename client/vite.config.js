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
      // user 服务：注册 / 登录 / 会话 / 简历上传等。本地 user-api 默认 :8124，prefix /api/users。
      '/api/users': {
        target: 'http://localhost:8124',
        changeOrigin: true,
      },
      // chat 服务：SSE 流式对话 / 知识库上传等。本地 chat 默认 :8123，prefix /api/ai。
      '/api/ai': {
        target: 'http://localhost:8123',
        changeOrigin: true,
      },
      // 其余 /api/* 兜底走外部生产环境（兼容旧路径，可按需删除）。
      // '/api': {
      //   target: 'http://101.42.249.106:8123/api',
      //   changeOrigin: true,
      //   rewrite: (path) => path.replace(/^\/api/, '')
      // },
    }
  }
})
