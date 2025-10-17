<template>
  <div class="interview-master-container">
    <div class="header">
      <div class="back-button" @click="goBack">返回</div>
      <h1 class="title">AI模拟面试官</h1>
      <div class="chat-id">会话ID: {{ chatId }}</div>
    </div>

    <!-- 注意事项区域 -->
    <div class="notice-section" v-if="showNotice">
      <div class="notice-container">
        <div class="notice-header">
          <span class="notice-icon">⚠️</span>
          <h3 class="notice-title">注意事项</h3>
          <button class="notice-close-btn" @click="closeNotice" title="关闭提示">
            <span>×</span>
          </button>
        </div>
        <div class="notice-content">
          <p>使用前建议请先上传你的简历，如果没有简历，请按照默认要求进行对话，这将保证你对话的准确性。</p>
        </div>
      </div>
    </div>

    <div class="content-wrapper">
      <div class="chat-area">
        <ChatRoom
          :messages="messages"
          :connection-status="connectionStatus"
          ai-type="interview"
          @send-message="sendMessage"
        />
      </div>
    </div>

    <!-- <div class="footer-container">
      <AppFooter />
    </div> -->
    <footer>
      <div class="footer-bottom">
        <div class="footer-bottom-content">
          <div class="copyright">
            <span class="copyright-symbol">©</span>
            {{ currentYear }} AI 面试官 - 让每一次面试都成为成长的机会
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useHead } from "@vueuse/head";
import ChatRoom from "../components/ChatRoom.vue";
import { chatWithLoveApp } from "../api";

// 设置页面标题和元数据
useHead({
  title: "AI面试 - aomanoh",
  meta: [
    {
      name: "description",
      content:
        "AI面试助手是专业面试助手，帮你解答各种面试问题，提供面试问题回答",
    },
    {
      name: "keywords",
      content: "AI面试助手,模拟面试,问题回答",
    },
  ],
});

const router = useRouter();
const messages = ref([]);
const chatId = ref("");
const connectionStatus = ref("disconnected");
const showNotice = ref(true); // 控制注意事项显示状态
let eventSource = null;

// 获取当前年份
const currentYear = computed(() => new Date().getFullYear());

// 生成随机会话ID
const generateChatId = () => {
  return "interview_" + Math.random().toString(36).substring(2, 10);
};

// 添加消息到列表
const addMessage = (content, isUser) => {
  messages.value.push({
    content,
    isUser,
    time: new Date().getTime(),
  });
};

// 发送消息
const sendMessage = (formData) => {
  // 从 FormData 中提取消息内容用于显示
  const message = formData.get("message") || "";

  addMessage(message, true);

  // 连接SSE
  if (eventSource) {
    eventSource.close();
  }

  // 创建一个空的AI回复消息
  addMessage("", false);
  const aiMessageIndex = messages.value.length - 1; // 修正：获取刚创建的AI消息的索引

  // 添加 chatId 到 FormData
  formData.append("chatId", chatId.value);

  connectionStatus.value = "connecting";

  // 定义消息处理回调
  const handleMessage = (data) => {
    if (data && data !== "[DONE]") {
      // 更新最新的AI消息内容，而不是创建新消息
      if (aiMessageIndex < messages.value.length) {
        messages.value[aiMessageIndex].content += data;
      }
    }

    if (data === "[DONE]") {
      connectionStatus.value = "disconnected";
      eventSource.close();
    }
  };

  // 定义错误处理回调
  const handleError = (error) => {
    connectionStatus.value = "error";
    eventSource.close();
  };

  // 使用新的 SSE 方法处理 FormData
  eventSource = chatWithLoveApp(formData, chatId.value);

  if (eventSource.onmessage !== undefined) {
    eventSource.onmessage = handleMessage;
    eventSource.onerror = handleError;
  } else {
    eventSource.onmessage = (event) => {
      handleMessage(event.data);
    };
    eventSource.onerror = handleError;
  }
};

// 返回主页
const goBack = () => {
  router.push("/");
};

// 关闭注意事项（仅当次会话有效）
const closeNotice = () => {
  showNotice.value = false;
  // 不再保存到localStorage，页面刷新后会恢复显示
};

// 页面加载时添加欢迎消息
onMounted(() => {
  // 注意事项始终显示，不检查localStorage状态（用于提示作用）
  showNotice.value = true;

  // 生成聊天ID
  chatId.value = generateChatId();

  // 添加欢迎消息
  addMessage(
    "我是你的 AI 面试官，请你点击下方《知识库PDF》上传你的简历（如果有），我会结合你的情况和需求帮你模拟面试和辅导。\n" +
      "比如：我叫 XXX（（简历里你的姓名-如果上传了简历）），是一名5年经验的Go后端开发工程师，目前比较欠缺分布式微服务、以及三高一海等相关经验，请你结合市场上的需求，帮我做模拟面试和辅导，给我相关的面试题和答案。",
    false
  );
});

// 组件销毁前关闭SSE连接
onBeforeUnmount(() => {
  if (eventSource) {
    eventSource.close();
  }
});
</script>

<style scoped>
.interview-master-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: #e3f2fd;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: #2196f3;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-button {
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: opacity 0.2s;
}

.back-button:hover {
  opacity: 0.8;
}

.back-button:before {
  content: "←";
  margin-right: 8px;
}

.title {
  font-size: 20px;
  font-weight: bold;
  margin: 0;
}

.chat-id {
  font-size: 14px;
  opacity: 0.8;
}

/* 注意事项样式 */
.notice-section {
  background: linear-gradient(135deg, #fff3cd 0%, #ffeaa7 100%);
  border-bottom: 1px solid #e6cc00;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.notice-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px;
}

.notice-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  position: relative;
}

.notice-close-btn {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  background: linear-gradient(135deg, #f59e0b, #d97706);
  border: 2px solid #ffffff;
  border-radius: 50%;
  font-size: 16px;
  color: #ffffff;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-weight: bold;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
  transition: all 0.3s ease;
}

.notice-close-btn:hover {
  background: linear-gradient(135deg, #d97706, #b45309);
  transform: translateY(-50%) scale(1.1);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.6);
}

.notice-close-btn:active {
  transform: translateY(-50%) scale(0.95);
  box-shadow: 0 2px 6px rgba(245, 158, 11, 0.5);
}

.notice-icon {
  font-size: 18px;
}

.notice-title {
  font-size: 16px;
  font-weight: 600;
  color: #856404;
  margin: 0;
}

.notice-content {
  margin-left: 26px;
}

.notice-content p {
  margin: 0;
  color: #664d03;
  font-size: 14px;
  line-height: 1.5;
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.chat-area {
  flex: 1;
  padding: 16px;
  overflow: hidden;
  position: relative;
  /* 设置最小高度确保内容显示正常 */
  min-height: calc(100vh - 56px - 180px); /* 100vh减去头部高度和页脚高度 */
  margin-bottom: 16px; /* 为页脚留出空间 */
}

.footer-container {
  margin-top: auto;
}

/* 响应式样式 */
@media (max-width: 768px) {
  .header {
    padding: 12px 16px;
  }

  .title {
    font-size: 18px;
  }

  .chat-id {
    font-size: 12px;
  }

  .chat-area {
    padding: 12px;
    min-height: calc(100vh - 48px - 160px); /* 调整计算值 */
    margin-bottom: 12px;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 10px 12px;
  }

  .back-button {
    font-size: 14px;
  }

  .title {
    font-size: 16px;
  }

  .chat-id {
    display: none;
  }

  .chat-area {
    padding: 8px;
    min-height: calc(100vh - 42px - 150px); /* 再次调整计算值 */
    margin-bottom: 8px;
  }
}

.footer-bottom {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(10px);
}

.footer-bottom-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
}

.copyright {
  font-family: "Inter", sans-serif;
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.copyright-symbol {
  font-weight: 600;
  color: #6366f1;
}

.tech-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(
    135deg,
    rgba(99, 102, 241, 0.2),
    rgba(16, 185, 129, 0.2)
  );
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 20px;
  padding: 6px 12px;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.8rem;
  color: #6366f1;
  font-weight: 500;
}

.badge-icon {
  font-size: 0.9rem;
  animation: rocket 2s ease-in-out infinite;
}

@keyframes rocket {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .footer-content {
    grid-template-columns: 1fr 1fr;
    gap: 30px;
    padding: 50px 20px 30px;
  }

  .brand-section {
    grid-column: 1 / -1;
  }
}

@media (max-width: 768px) {
  .footer-content {
    grid-template-columns: 1fr;
    gap: 30px;
    padding: 40px 20px 30px;
  }

  .footer-logo {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }

  .brand-description {
    text-align: center;
  }

  .footer-section h4 {
    justify-content: center;
  }

  .footer-links {
    align-items: center;
  }

  .footer-bottom-content {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }
}

@media (max-width: 480px) {
  .footer-content {
    padding: 30px 15px 20px;
    gap: 25px;
  }

  .logo-icon {
    width: 40px;
    height: 40px;
  }

  .code-brackets {
    font-size: 1.2rem;
  }

  .logo-text h3 {
    font-size: 1.2rem;
  }

  .footer-section h4 {
    font-size: 1rem;
    margin-bottom: 15px;
  }

  .footer-link {
    font-size: 0.85rem;
    justify-content: center;
  }

  .copyright {
    font-size: 0.8rem;
    text-align: center;
  }

  .tech-badge {
    font-size: 0.75rem;
    padding: 5px 10px;
  }
}

/* 暗色主题增强 */
@media (prefers-color-scheme: dark) {
  .app-footer {
    background: linear-gradient(135deg, #0a0f1c 0%, #1a1f2e 100%);
  }
}

/* 打印样式 */
@media print {
  .app-footer {
    background: white !important;
    color: black !important;
  }

  .footer-link,
  .copyright,
  .tech-badge {
    color: black !important;
  }
}
</style>
