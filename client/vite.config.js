import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const userProxyTarget = env.VITE_DEV_USER_PROXY_TARGET || 'http://127.0.0.1:8124'
  const chatProxyTarget = env.VITE_DEV_CHAT_PROXY_TARGET || 'http://127.0.0.1:8123'

  return {
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
          target: userProxyTarget,
          changeOrigin: true,
        },
        // chat 服务：SSE 流式对话 / 知识库上传等。本地 chat 默认 :8123，prefix /api/ai。
        '/api/ai': {
          target: chatProxyTarget,
          changeOrigin: true,
        },
      }
    }
  }
})
