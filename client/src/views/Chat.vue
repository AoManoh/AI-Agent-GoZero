<template>
  <div class="page chat-page">
    <AuroraBackground />

    <header class="site-header">
      <button class="nav-button" type="button" @click="goHome">&larr; 结束面试</button>
      <h2 style="font-weight: 500">AI 模拟面试</h2>
      <nav class="main-nav">
        <ThemeToggle :theme="theme" @toggle="toggleTheme" />
      </nav>
    </header>

    <div class="chat-container">
      <div class="chat-log" ref="messagesContainer">
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['chat-message', msg.isUser ? 'user' : 'ai']"
        >
          <div class="avatar-block">
            <div class="avatar">
              <img
                v-if="!msg.isUser"
                src="http://pic.aomanoh.com/note/20250828030959474.jpeg"
                alt="AI"
              />
              <img
                v-else
                src="http://pic.aomanoh.com/note/20250903035824302.jpg"
                alt="User"
              />
            </div>
            <div class="timestamp">
              <div class="timestamp-date">{{ formatDate(msg.time) }}</div>
              <div class="timestamp-time">{{ formatTimeOnly(msg.time) }}</div>
            </div>
          </div>
          <div class="message-bubble">
            <p v-if="msg.isUser">{{ msg.content }}</p>
            <MarkdownMessage v-else :content="processMessageContent(msg.content || '')" />
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <div class="input-wrapper" :class="{ disabled: isConnecting }">
          <div class="chat-actions">
            <el-upload
              v-model:file-list="file"
              :limit="1"
              :auto-upload="false"
              :show-file-list="false"
              :on-exceed="handleExceed"
              :disabled="isConnecting"
            >
              <button
                class="action-btn"
                type="button"
                title="上传文件"
                :disabled="isConnecting"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="m18.375 12.739-7.693 7.693a4.5 4.5 0 0 1-6.364-6.364l10.94-10.94A3.375 3.375 0 1 1 18.374 7.5l-1.497 1.498a1.5 1.5 0 0 1-2.121-2.121l1.497-1.498 2.121-2.121a3.375 3.375 0 0 1 4.773 4.773l-10.94 10.94a4.5 4.5 0 1 1-6.364-6.364l7.693-7.693a.75.75 0 0 1 1.06 1.06z"
                  />
                </svg>
              </button>
            </el-upload>

            <el-upload
              v-model:file-list="knowledgeFile"
              :limit="1"
              accept=".pdf"
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleKnowledgeFileChange"
              :on-exceed="handleExceedKnowledge"
              :disabled="isConnecting || uploadingKnowledge"
            >
              <button
                class="action-btn"
                type="button"
                title="上传到知识库"
                :disabled="isConnecting || uploadingKnowledge"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6-2.292m0 0V3.75m0 12.558a8.966 8.966 0 0 1-6-2.292V3.75a8.966 8.966 0 0 1 6 2.292m0 12.558a8.966 8.966 0 0 0 6-2.292V3.75a8.966 8.966 0 0 0-6 2.292"
                  />
                </svg>
              </button>
            </el-upload>
          </div>

          <textarea
            ref="textareaRef"
            v-model="inputMessage"
            class="chat-textarea"
            placeholder="输入消息..."
            :disabled="isConnecting"
            @input="autoResizeTextarea"
            @keydown="handleKeydown"
          ></textarea>

          <button
            class="action-btn send-btn"
            type="button"
            title="发送"
            :disabled="isConnecting"
            @click="handleSendMessage"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M3.478 2.405a.75.75 0 0 0-.926.94l2.432 7.905H13.5a.75.75 0 0 1 0 1.5H4.984l-2.432 7.905a.75.75 0 0 0 .926.94 60.519 60.519 0 0 0 18.445-8.986.75.75 0 0 0 0-1.218A60.517 60.517 0 0 0 3.478 2.405Z"
              />
            </svg>
          </button>
        </div>

        <div v-if="fileName || uploadingKnowledge" class="input-hints">
          <span v-if="fileName" class="file-tag">已附加：{{ fileName }}</span>
          <span v-if="uploadingKnowledge" class="hint">知识库文件上传中...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import AuroraBackground from "../components/AuroraBackground.vue";
import ThemeToggle from "../components/ThemeToggle.vue";
import MarkdownMessage from "../components/MarkdownMessage.vue";
import { useChatStream } from "../composables/useChatStream";
import { useTheme } from "../composables/useTheme";
import { uploadKnowledge } from "../api/index.js";

const router = useRouter();
const { theme, toggleTheme } = useTheme();

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
    content: "你好，我是你的AI面试官。我们开始吧。记得请上传你的简历，然后告诉我你想开展怎样的面试奥~~~",
    time: Date.now(),
    isUser: false,
  },
];

const { messages, isStreaming, startStream, pushUserMessage, dispose } =
  useChatStream(initialMessages);
const chatId = ref(loadStoredChatId() || createChatId());

// File upload states
const file = ref([]);
const knowledgeFile = ref([]);
const inputMessage = ref("");
const messagesContainer = ref(null);
const textareaRef = ref(null);
const uploadingKnowledge = ref(false);

watch(
  () => chatId.value,
  (value) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("chatId", value);
    }
  }
);

const isConnecting = computed(() => isStreaming.value);
const fileName = computed(() => (file.value[0]?.name ? file.value[0].name : ""));

// Auto scroll to bottom when messages change
const scrollToBottom = () => {
  nextTick(() => {
    const container = messagesContainer.value;
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  });
};

watch(messages, () => {
  scrollToBottom();
}, { deep: true });

onMounted(() => {
  scrollToBottom();
});

// Process message content to handle escape sequences
const processMessageContent = (content) => {
  if (!content) return "";

  let processedContent = content;
  processedContent = processedContent.replace(/\\\\\\/g, "\\TEMP_BACKSLASH\\");
  processedContent = processedContent
    .replace(/\\n/g, "\n")
    .replace(/\\t/g, "\t")
    .replace(/\\r/g, "\r")
    .replace(/\\"/g, '"')
    .replace(/\\'/g, "'")
    .replace(/\\\//g, "/")
    .replace(/\\b/g, "\b")
    .replace(/\\f/g, "\f")
    .replace(/\\v/g, "\v")
    .replace(/\\\\/g, "\\")
    .replace(/\\TEMP_BACKSLASH\\/g, "\\");

  processedContent = processedContent.replace(/\\u([0-9a-fA-F]{4})/g, (match, hex) => {
    return String.fromCharCode(parseInt(hex, 16));
  });

  processedContent = processedContent
    .replace(/\\`/g, "`")
    .replace(/\\~/g, "~")
    .replace(/\\#/g, "#")
    .replace(/\\>/g, ">")
    .replace(/\\</g, "<")
    .replace(/\\&/g, "&")
    .replace(/\\\*/g, "*")
    .replace(/\\-/g, "-")
    .replace(/\\\+/g, "+")
    .replace(/\\=/g, "=")
    .replace(/\\\|/g, "|")
    .replace(/\\\[/g, "[")
    .replace(/\\\]/g, "]")
    .replace(/\\\{/g, "{")
    .replace(/\\\}/g, "}")
    .replace(/\\\(/g, "(")
    .replace(/\\\)/g, ")");

  return processedContent.trim();
};

// Handle send message
const sendMessage = () => {
  if (isConnecting.value) {
    ElMessage.warning("AI 正在回复，请稍候");
    return;
  }

  const content = inputMessage.value.trim();
  if (!content) {
    ElMessage.warning("请输入内容");
    return;
  }

  const formData = new FormData();
  formData.append("message", content);

  if (file.value.length > 0 && file.value[0]?.raw) {
    formData.append("file", file.value[0].raw);
  }

  pushUserMessage(content);
  startStream(formData, chatId.value);

  file.value = [];
  inputMessage.value = "";
  
  // Reset textarea height to default based on screen size
  if (textareaRef.value) {
    const isMobile = window.innerWidth <= 768;
    const defaultHeight = isMobile ? 44 : 48;
    textareaRef.value.style.height = `${defaultHeight}px`;
    textareaRef.value.style.overflowY = 'hidden';
  }
};

function handleSendMessage() {
  sendMessage();
}

const handleKeydown = (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    sendMessage();
  }
};

// Format date as MM/DD
const formatDate = (timestamp) => {
  if (!timestamp) return "";
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return "";
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${month}/${day}`;
};

// Format time as HH:MM
const formatTimeOnly = (timestamp) => {
  if (!timestamp) return "";
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return "";
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  return `${hours}:${minutes}`;
};

// Auto resize textarea based on content
const autoResizeTextarea = () => {
  const textarea = textareaRef.value;
  if (!textarea) return;
  
  // Determine min/max height based on screen size
  const isMobile = window.innerWidth <= 768;
  const minHeight = isMobile ? 44 : 48;
  const maxHeight = isMobile ? 150 : 200;
  
  // Reset height to default to recalculate scrollHeight
  textarea.style.height = `${minHeight}px`;
  
  // Check if content exceeds default height
  if (textarea.scrollHeight > minHeight) {
    // Calculate new height based on scrollHeight
    const newHeight = Math.min(textarea.scrollHeight, maxHeight);
    textarea.style.height = `${newHeight}px`;
    
    // Show scrollbar if content exceeds max height
    textarea.style.overflowY = textarea.scrollHeight > maxHeight ? 'auto' : 'hidden';
  } else {
    // Keep default height and hide scrollbar
    textarea.style.height = `${minHeight}px`;
    textarea.style.overflowY = 'hidden';
  }
};

// File upload handlers
const handleExceed = (files) => {
  file.value = [files[0]];
};

const handleExceedKnowledge = (files) => {
  knowledgeFile.value = [files[0]];
};

const handleKnowledgeFileChange = async (uploadFile) => {
  if (uploadFile.status !== "ready" || !uploadFile.raw) {
    return;
  }

  if (uploadFile.raw.type !== "application/pdf") {
    ElMessage.error("只支持PDF格式的文件");
    knowledgeFile.value = [];
    return;
  }

  if (uploadFile.raw.size > 500 * 1024) {
    ElMessage.error("文件大小不能超过500KB");
    knowledgeFile.value = [];
    return;
  }

  uploadingKnowledge.value = true;
  const loadingMessage = ElMessage({
    message: "正在上传知识库文件...",
    type: "info",
    duration: 0,
  });

  try {
    const formData = new FormData();
    formData.append("file", uploadFile.raw);
    await uploadKnowledge(formData);
    loadingMessage.close();
    ElMessage.success("知识库文件上传成功");
    knowledgeFile.value = [];
  } catch (error) {
    loadingMessage.close();
    ElMessage.error(error.response?.data?.message || "知识库文件上传失败，请重试");
    knowledgeFile.value = [];
  } finally {
    uploadingKnowledge.value = false;
  }
};

function goHome() {
  router.push({ name: "Home" });
}

onBeforeUnmount(() => {
  dispose();
});
</script>

<style scoped>
.chat-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.chat-page > *:not(.aurora-background) {
  position: relative;
  z-index: 1;
}

/* ===== Site Header ===== */
.site-header {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.site-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--color-text-primary);
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nav-button {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: 10px 18px;
  border-radius: 10px;
  font-weight: 500;
  transition: background-color 0.2s ease;
  cursor: pointer;
}

body.light-mode .nav-button {
  background: rgba(0, 0, 0, 0.05);
}

.nav-button:hover {
  background: rgba(255, 255, 255, 0.2);
}

body.light-mode .nav-button:hover {
  background: rgba(0, 0, 0, 0.1);
}

/* ===== Chat Container ===== */
.chat-container {
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  padding: 0 20px 20px;
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  height: calc(100vh - 93px - 85px);
}

/* ===== Chat Log ===== */
.chat-log {
  flex-grow: 1;
  overflow-y: auto;
  padding-right: 10px;
}

.chat-log::-webkit-scrollbar {
  width: 6px;
}

.chat-log::-webkit-scrollbar-track {
  background: transparent;
}

.chat-log::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

body.light-mode .chat-log::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
}

/* ===== Chat Message ===== */
.chat-message {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  max-width: 90%;
}

.chat-message.user {
  margin-left: auto;
  flex-direction: row-reverse;
}

/* ===== Avatar Block ===== */
.avatar-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* ===== Avatar ===== */
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  overflow: hidden;
}

.avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.avatar span {
  color: var(--color-text-primary);
  font-size: 1rem;
}

/* ===== Timestamp ===== */
.timestamp {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  line-height: 1.2;
  min-width: 40px;
}

.timestamp-date {
  font-weight: 500;
}

.timestamp-time {
  font-weight: 400;
  opacity: 0.9;
}

/* ===== Message Bubble ===== */
.message-bubble {
  background: var(--color-input-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
  line-height: 1.6;
  min-width: 0;
  word-wrap: break-word;
}

.chat-message.user .message-bubble {
  background: var(--color-user-bubble-bg);
  border-color: var(--color-user-bubble-border);
}

.message-bubble p {
  margin: 0;
  color: var(--color-text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

/* Deep selectors for dynamically rendered content */
.message-bubble :deep(p) {
  margin: 0 0 12px 0;
}

.message-bubble :deep(p:last-child) {
  margin-bottom: 0;
}

.message-bubble :deep(pre) {
  background-color: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  padding: 16px;
  margin-top: 12px;
  position: relative;
  overflow-x: auto;
}

body.light-mode .message-bubble :deep(pre) {
  background-color: rgba(0, 0, 0, 0.05);
}

.message-bubble :deep(code) {
  font-family: "SF Mono", "Fira Code", "Consolas", monospace;
  font-size: 0.9rem;
  line-height: 1.7;
}

.message-bubble :deep(.copy-code-btn) {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  border-radius: 4px;
  padding: 4px 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.message-bubble :deep(.copy-code-btn:hover) {
  background: rgba(255, 255, 255, 0.2);
}

body.light-mode .message-bubble :deep(.copy-code-btn) {
  background: rgba(0, 0, 0, 0.1);
  border-color: rgba(0, 0, 0, 0.15);
}

body.light-mode .message-bubble :deep(.copy-code-btn:hover) {
  background: rgba(0, 0, 0, 0.2);
}

/* ===== Chat Input Area ===== */
.chat-input-area {
  padding-top: 20px;
  flex-shrink: 0;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  background: var(--color-input-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 8px;
  backdrop-filter: blur(10px);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.input-wrapper:focus-within {
  border-color: var(--color-glow-1);
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.5);
}

.input-wrapper.disabled {
  opacity: 0.6;
  pointer-events: none;
}

/* ===== Chat Actions ===== */
.chat-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chat-actions :deep(.el-upload) {
  display: inline-flex;
}

.action-btn {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 8px;
  transition: color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn:hover {
  color: var(--color-text-primary);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn svg {
  width: 24px;
  height: 24px;
}

.send-btn {
  background: var(--color-glow-1);
  border-radius: 8px;
  color: #fff;
}

.send-btn:hover {
  background: #4f8fff;
}

.send-btn svg {
  color: #fff;
}

/* ===== Chat Textarea ===== */
.chat-textarea {
  flex-grow: 1;
  background: transparent;
  border: none;
  color: var(--color-text-primary);
  font-size: 1rem;
  resize: none;
  padding: 12px;
  height: 48px;
  max-height: 200px;
  font-family: inherit;
  line-height: 1.5;
  overflow-y: hidden;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.chat-textarea:focus {
  outline: none;
}

.chat-textarea::placeholder {
  color: var(--color-text-secondary);
}

.chat-textarea::-webkit-scrollbar {
  width: 4px;
}

.chat-textarea::-webkit-scrollbar-track {
  background: transparent;
}

.chat-textarea::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

body.light-mode .chat-textarea::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
}

/* ===== Input Hints ===== */
.input-hints {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--color-text-secondary);
  font-size: 0.85rem;
}

.file-tag {
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--color-border);
  background: var(--color-card-bg);
}

.hint {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* ===== Responsive Design ===== */
@media (max-width: 768px) {
  .site-header {
    padding: 24px 16px;
  }

  .site-header h2 {
    font-size: 1.3rem;
  }

  .nav-button {
    padding: 8px 14px;
  }

  .chat-container {
    padding: 0 16px 16px;
    height: calc(100vh - 80px - 20px);
  }

  .chat-message {
    max-width: 100%;
  }

  .avatar-block {
    gap: 6px;
  }

  .avatar {
    width: 36px;
    height: 36px;
  }

  .timestamp {
    font-size: 0.65rem;
    min-width: 36px;
    gap: 1px;
  }

  .input-wrapper {
    align-items: flex-end;
    padding: 10px;
    gap: 10px;
  }

  .chat-actions {
    align-items: flex-end;
    gap: 6px;
    padding-bottom: 2px;
  }
  
  .chat-actions :deep(.el-upload) {
    display: inline-flex;
    align-items: center;
  }

  .chat-textarea {
    height: 44px;
    min-height: 44px;
    max-height: 150px;
    padding: 11px 12px;
    font-size: 0.95rem;
    line-height: 1.4;
  }

  .action-btn {
    flex-shrink: 0;
    min-width: 40px;
    min-height: 40px;
    height: 40px;
    padding: 8px;
    margin-bottom: 2px;
  }

  .action-btn svg {
    width: 20px;
    height: 20px;
  }
  
  .send-btn {
    min-width: 44px;
    min-height: 44px;
    height: 44px;
    margin-bottom: 0;
  }
  
  .input-hints {
    margin-top: 8px;
    font-size: 0.8rem;
  }
}
</style>
