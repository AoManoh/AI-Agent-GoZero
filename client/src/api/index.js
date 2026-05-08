import axios from 'axios'
import { ACCESS_TOKEN_KEY } from './core.js'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const readAccessToken = () => {
  if (typeof window === 'undefined') return ''
  return window.localStorage.getItem(ACCESS_TOKEN_KEY) || ''
}

// 创建axios实例
const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 60000
})

// 自动附加 JWT Bearer token（与 core.js 行为对齐，避免登录后调用 admin/private 接口 401）
request.interceptors.request.use((config) => {
  const token = readAccessToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 封装SSE连接 - 支持FormData
export const connectSSEWithFormData = (url, formData) => {
  // 构建基础URL
  const fullUrl = `${API_BASE_URL}${url}`

  let abortController = null
  let messageCallback = null
  let errorCallback = null

  // 创建EventSource，但需要先通过POST请求获取SSE流
  // 由于EventSource只支持GET请求，我们需要使用fetch来支持FormData
  const fetchSSE = async () => {
    try {
      abortController = new AbortController()

      const accessToken = readAccessToken()
      const headers = {
        // 不设置Content-Type，让浏览器自动设置multipart/form-data
      }
      if (accessToken) {
        headers.Authorization = `Bearer ${accessToken}`
      }

      const response = await fetch(fullUrl, {
        method: 'POST',
        body: formData,
        signal: abortController.signal,
        headers,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()

      const processStream = async () => {
        try {
          let buffer = ''

          while (true) {
            const { done, value } = await reader.read()

            if (done) {
              if (messageCallback) messageCallback('[DONE]')
              break
            }

            const chunk = decoder.decode(value, { stream: true })
            buffer += chunk

            // 按行分割，但保留不完整的行
            const lines = buffer.split('\n')
            buffer = lines.pop() || '' // 保留最后一个可能不完整的行

            for (const line of lines) {
              if (line.startsWith('data: ')) {
                // 关键：不要 .trim()，否则代码块里 token 之间的前后空格会被吃掉，
                // 导致 markdown 渲染出 "typegstruct{" 这种空格全丢的样子。
                // 仅在判定 [DONE] 标记时用 trim 比较，传给 callback 的内容必须保留原样。
                const data = line.slice(6).replace(/\r$/, '')
                const dataTrimmed = data.trim()
                if (dataTrimmed === '[DONE]') {
                  if (messageCallback) messageCallback('[DONE]')
                } else if (data) {
                  if (messageCallback) messageCallback(data)
                }
              } else if (line.trim() && !line.startsWith(':')) {
                // 处理没有 'data: ' 前缀的行（某些SSE实现可能直接发送数据）
                if (line.trim() === '[DONE]') {
                  if (messageCallback) messageCallback('[DONE]')
                } else if (line.trim()) {
                  if (messageCallback) messageCallback(line.trim())
                }
              }
            }
          }
        } catch (error) {
          if (errorCallback) errorCallback(error)
        }
      }

      processStream()
    } catch (error) {
      if (errorCallback) errorCallback(error)
    }
  }

  // 启动SSE流处理
  fetchSSE()

  // 返回一个模拟的EventSource对象，提供close方法和事件监听器
  const sseObject = {
    onmessage: null,
    onerror: null,
    close: () => {
      if (abortController) {
        abortController.abort()
      }
    }
  }

  // 重写 onmessage 和 onerror 的 setter，以便动态更新回调
  Object.defineProperty(sseObject, 'onmessage', {
    get: () => messageCallback,
    set: (callback) => {
      messageCallback = callback
    }
  })

  Object.defineProperty(sseObject, 'onerror', {
    get: () => errorCallback,
    set: (callback) => {
      errorCallback = callback
    }
  })

  return sseObject
}

// 封装SSE连接 - 原有版本（用于向后兼容）
export const connectSSE = (url, params, onMessage, onError) => {
  // 构建带参数的URL
  const queryString = Object.keys(params)
    .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
    .join('&')

  const fullUrl = `${API_BASE_URL}${url}?${queryString}`

  // 创建EventSource
  const eventSource = new EventSource(fullUrl)

  eventSource.onmessage = event => {
    let data = event.data

    // 检查是否是特殊标记
    if (data === '[DONE]') {
      if (onMessage) onMessage('[DONE]')
    } else {
      // 处理普通消息
      if (onMessage) onMessage(data)
    }
  }

  eventSource.onerror = error => {
    if (onError) onError(error)
    eventSource.close()
  }

  // 返回eventSource实例，以便后续可以关闭连接
  return eventSource
}

// AI面试助手聊天 - 支持FormData
export const chatWithLoveApp = (formData, chatId) => {
  // 如果传入的是FormData，使用新的SSE方法
  if (formData instanceof FormData) {
    return connectSSEWithFormData('/ai/interview_app/chat/sse', formData)
  }

  // 向后兼容：如果传入的是字符串消息，使用原有方法
  return connectSSE('/ai/interview_app/chat/sse', { message: formData, chatId })
}

// AI超级智能体聊天
export const chatWithManus = (message) => {
  return connectSSE('/ai/manus/chat', { message })
}

// 上传知识库文件
export const uploadKnowledge = (formData) => {
  return request.post('/ai/knowledge/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    timeout: 120000 // 2分钟超时，适用于文件上传
  })
}

export default {
  chatWithLoveApp,
  chatWithManus,
  uploadKnowledge
}