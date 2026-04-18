<template>
  <div class="page chat-page" id="chat">
    <div class="workbench">
      <ChatSidebar @new-chat="handleNewChat" />

      <!-- Main Chat -->
      <div class="main-chat">
        <div class="mc-header">
          <div class="mc-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
            AI 模拟面试
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
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRouter } from "vue-router";
import ChatSidebar from "../components/ChatSidebar.vue";
import ChatInputDock from "../components/ChatInputDock.vue";
import ChatMessageList from "../components/ChatMessageList.vue";
import { useChatStream } from "../composables/useChatStream";

const router = useRouter();

const createChatId = () => {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  const randomPart = Math.random().toString(16).slice(2);
  return `${Date.now()}-${randomPart}`;
};

const loadStoredChatId = () => {
  if (typeof window !== "undefined") {
    return window.localStorage.getItem("chatId");
  }
  return null;
};

const initialMessages = [
  {
    role: "ai",
    type: "text",
    content: "你好，我是你的AI面试官。请问有什么需要考察的？你可以附带你的简历并说明你想要面试的技术方向与岗位。",
    time: Date.now(),
    isUser: false,
  },
];

const { messages, isStreaming, startStream, pushUserMessage, dispose } =
  useChatStream(initialMessages);
const chatId = ref(loadStoredChatId() || createChatId());

const isPending = computed(() => {
  if (!isStreaming.value) return false;
  const lastMsg = messages.value[messages.value.length - 1];
  if (lastMsg && lastMsg.isUser) return true;
  return false;
});

watch(
  () => chatId.value,
  (value) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("chatId", value);
    }
  }
);

const isConnecting = computed(() => isStreaming.value);

const handleSendMessage = (payload) => {
  const formData = new FormData();
  formData.append("message", payload.message);

  if (payload.file) {
    formData.append("file", payload.file);
  }

  pushUserMessage(payload.message);
  startStream(formData, chatId.value);
};

const handleNewChat = () => {
  chatId.value = createChatId();
  // Clear messages or handle session switch...
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
  border-radius: 6px;
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
