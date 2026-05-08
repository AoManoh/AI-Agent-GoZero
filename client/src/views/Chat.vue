<template>
  <div class="page chat-page" id="chat">
    <div class="workbench">
      <ChatSidebar
        :sessions="sessions"
        :active-session-id="activeSessionId"
        :username="username"
        @new-chat="handleNewChat"
        @select-session="handleSelectSession"
      />

      <!-- Main Chat -->
      <div class="main-chat">
        <div class="mc-header">
          <div class="mc-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
            {{ activeSession?.title || 'AI 模拟面试' }}
          </div>
          <button class="btn-ghost" @click="goHome">&larr; 结束面试</button>
        </div>

        <ChatMessageList
          :messages="messages"
          :isConnecting="isConnecting"
          :isPending="isPending"
        />

        <ChatInputDock
          :isConnecting="isConnecting"
          @send-message="handleSendMessage"
          @stop-stream="handleStopStream"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, watch } from "vue";
import { useRouter } from "vue-router";
import ChatSidebar from "../components/ChatSidebar.vue";
import ChatInputDock from "../components/ChatInputDock.vue";
import ChatMessageList from "../components/ChatMessageList.vue";
import { useChatStream } from "../composables/useChatStream";
import { useInterviewSessions } from "../composables/useInterviewSessions";
import { authStorage } from "../api/core.js";

const router = useRouter();

const initialMessages = [
  {
    role: "ai",
    type: "text",
    content: "你好，我是你的AI面试官。请问有什么需要考察的？你可以附带你的简历并说明你想要面试的技术方向与岗位。",
    time: Date.now(),
    isUser: false,
  },
];

const {
  sessions,
  activeSessionId,
  activeSession,
  activateSession,
  createSession,
  updateActiveSession,
} = useInterviewSessions(initialMessages);

const username = authStorage.getSession().username || "";

const {
  messages,
  isStreaming,
  startStream,
  stopStream,
  pushUserMessage,
  setMessages,
  dispose,
} = useChatStream(activeSession.value?.messages || initialMessages);

// 用户主动点击"停止生成"按钮：中断当前 SSE 流，AI 消息会自动追加"已被用户中断"标记。
const handleStopStream = () => {
  stopStream();
};

const isPending = computed(() => {
  if (!isStreaming.value) return false;
  const lastMsg = messages.value[messages.value.length - 1];
  return Boolean(lastMsg && lastMsg.isUser);
});

const isConnecting = computed(() => isStreaming.value);

const deriveTitleFromMessage = (text) => {
  if (!text) return "";
  const trimmed = String(text).trim().replace(/\s+/g, " ");
  return trimmed.length > 24 ? `${trimmed.slice(0, 24)}…` : trimmed;
};

watch(
  () => activeSessionId.value,
  () => {
    setMessages(activeSession.value?.messages || initialMessages);
  }
);

watch(
  messages,
  (next) => {
    if (!activeSession.value) return;
    const updates = { messages: next };
    if (
      activeSession.value.title === "未命名会话" ||
      !activeSession.value.title
    ) {
      const firstUser = next.find((m) => m.isUser && m.content);
      if (firstUser) {
        updates.title = deriveTitleFromMessage(firstUser.content) || "未命名会话";
      }
    }
    updateActiveSession(updates);
  },
  { deep: true }
);

const handleSendMessage = (payload) => {
  const formData = new FormData();
  formData.append("message", payload.message);

  if (payload.file) {
    formData.append("file", payload.file);
  }

  pushUserMessage(payload.message);
  startStream(formData, activeSessionId.value);
};

const handleNewChat = () => {
  createSession({ messages: initialMessages });
};

const handleSelectSession = (sessionId) => {
  activateSession(sessionId);
};

function goHome() {
  router.push({ name: "Home" });
}

onBeforeUnmount(() => {
  dispose();
});
</script>

<style scoped>
/* V8 Productivity Layout Styles */
.chat-page {
  position: relative;
  z-index: 1;
  background: transparent;
  padding: 0;
  margin: 0;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.workbench {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

/* Main Chat Area Layout */
.main-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
}

.mc-header {
  height: 65px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  flex-shrink: 0;
  background: rgba(4, 4, 6, 0.4);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

.mc-title {
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-ghost {
  background: transparent;
  border: none;
  color: var(--t2);
  cursor: pointer;
  font-family: var(--sans);
  font-size: 13px;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  transition: background 0.2s;
}

.btn-ghost:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--t);
}

@media (max-width: 768px) {
  .mc-header {
    padding: 0 16px;
  }
}
</style>
