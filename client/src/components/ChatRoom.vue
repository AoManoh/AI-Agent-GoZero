<template>
  <div class="chat-room">
    <div class="chat-container">
      <div class="chat-log" ref="messagesContainer">
        <div class="chat-messages">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            :class="['chat-message', msg.isUser ? 'user' : 'ai']"
          >
            <div class="avatar-block">
              <div class="avatar">
                <img
                  v-if="msg.isUser"
                  src="http://pic.aomanoh.com/note/20250903035824302.jpg"
                  alt="用户头像"
                  class="avatar-image"
                />
                <AiAvatarFallback v-else :type="aiType" />
              </div>
              <div class="timestamp">{{ formatTime(msg.time) }}</div>
            </div>
            <div class="message-bubble" :class="{ user: msg.isUser }">
              <div class="message-content" v-if="msg.isUser">
                {{ msg.content }}
              </div>
              <MarkdownMessage
                v-else
                :content="processMessageContent(msg.content || '')"
                class="message-content"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <div class="input-wrapper" :class="{ disabled: isConnecting }">
          <div class="chat-actions">
            <el-upload
              class="action-upload"
              v-model:file-list="file"
              :limit="1"
              :auto-upload="false"
              :show-file-list="false"
              :before-remove="beforeRemove"
              :on-exceed="handleExceed"
              :disabled="isConnecting"
            >
              <button class="action-btn" type="button" title="上传文件" :disabled="isConnecting">
                <el-icon><Paperclip /></el-icon>
              </button>
            </el-upload>

            <el-upload
              class="action-upload"
              v-model:file-list="knowledgeFile"
              :limit="1"
              accept=".pdf"
              :auto-upload="false"
              :show-file-list="false"
              :before-remove="beforeRemoveKnowledge"
              :on-exceed="handleExceedKnowledge"
              :on-change="handleKnowledgeFileChange"
              :disabled="isConnecting || uploadingKnowledge"
            >
              <button
                class="action-btn"
                type="button"
                title="上传到知识库"
                :disabled="isConnecting || uploadingKnowledge"
              >
                <el-icon><Document /></el-icon>
              </button>
            </el-upload>
          </div>

          <textarea
            v-model="inputMessage"
            class="chat-textarea"
            placeholder="输入消息..."
            :disabled="isConnecting"
            @keydown="handleKeydown"
          ></textarea>

          <button
            class="action-btn send-btn"
            type="button"
            title="发送"
            :disabled="isConnecting"
            @click="handleSendMessage"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M3.478 2.405a.75.75 0 0 0-.926.94l2.432 7.905H13.5a.75.75 0 0 1 0 1.5H4.984l-2.432 7.905a.75.75 0 0 0 .926.94 60.519 60.519 0 0 0 18.445-8.986.75.75 0 0 0 0-1.218A60.517 60.517 0 0 0 3.478 2.405Z"/></svg>
          </button>
        </div>

        <div class="input-hints">
          <span v-if="fileName" class="file-tag">已附加：{{ fileName }}</span>
          <span v-if="uploadingKnowledge" class="hint">知识库文件上传中...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { Paperclip, Document } from "@element-plus/icons-vue";
import MarkdownMessage from "./MarkdownMessage.vue";
import AiAvatarFallback from "./AiAvatarFallback.vue";
import { uploadKnowledge } from "../api/index.js";

const props = defineProps({
  messages: {
    type: Array,
    default: () => [],
  },
  connectionStatus: {
    type: String,
    default: "disconnected",
  },
  aiType: {
    type: String,
    default: "default", // 'interview' 或 'super'
  },
});

const emit = defineEmits(["send-message"]);

const file = ref([]);
const knowledgeFile = ref([]);
const inputMessage = ref("");
const messagesContainer = ref(null);
const uploadingKnowledge = ref(false);

const isConnecting = computed(() => props.connectionStatus === "connecting");
const fileName = computed(() => (file.value[0]?.name ? file.value[0].name : ""));

const scrollToBottom = () => {
  nextTick(() => {
    const container = messagesContainer.value;
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  });
};

watch(
  () => props.messages,
  () => {
    scrollToBottom();
  },
  { deep: true }
);

onMounted(() => {
  scrollToBottom();
});

// 处理消息内容中的转义符
const processMessageContent = (content) => {
  if (!content) return "";

  // 处理各种转义符，按照正确的顺序处理避免冲突
  let processedContent = content;

  // 首先处理双重转义（避免 \\n 被错误处理）
  processedContent = processedContent.replace(/\\\\\\/g, "\\TEMP_BACKSLASH\\");

  // 处理常见的转义符
  processedContent = processedContent
    .replace(/\\n/g, "\n") // 将 \n 转换为真正的换行符
    .replace(/\\t/g, "\t") // 将 \t 转换为制表符
    .replace(/\\r/g, "\r") // 将 \r 转换为回车符
    .replace(/\\"/g, '"') // 将 \" 转换为双引号
    .replace(/\\'/g, "'") // 将 \' 转换为单引号
    .replace(/\\\//g, "/") // 将 \/ 转换为斜杠
    .replace(/\\b/g, "\b") // 将 \b 转换为退格符
    .replace(/\\f/g, "\f") // 将 \f 转换为换页符
    .replace(/\\v/g, "\v") // 将 \v 转换为垂直制表符
    .replace(/\\\\/g, "\\") // 将 \\ 转换为反斜杠
    .replace(/\\TEMP_BACKSLASH\\/g, "\\"); // 恢复临时标记的反斜杠

  // 处理可能的 Unicode 转义序列
  processedContent = processedContent.replace(
    /\\u([0-9a-fA-F]{4})/g,
    (match, hex) => {
      return String.fromCharCode(parseInt(hex, 16));
    }
  );

  // 处理特殊的 markdown 字符转义
  processedContent = processedContent
    .replace(/\\`/g, "`") // 将 \` 转换为反引号
    .replace(/\\~/g, "~") // 将 \~ 转换为波浪号
    .replace(/\\#/g, "#") // 将 \# 转换为井号
    .replace(/\\>/g, ">") // 将 \> 转换为大于号
    .replace(/\\</g, "<") // 将 \< 转换为小于号
    .replace(/\\&/g, "&") // 将 \& 转换为和号
    .replace(/\\\*/g, "*") // 将 \* 转换为星号
    .replace(/\\-/g, "-") // 将 \- 转换为横线
    .replace(/\\\+/g, "+") // 将 \+ 转换为加号
    .replace(/\\=/g, "=") // 将 \= 转换为等号
    .replace(/\\\|/g, "|") // 将 \| 转换为竖线
    .replace(/\\\[/g, "[") // 将 \[ 转换为左方括号
    .replace(/\\\]/g, "]") // 将 \] 转换为右方括号
    .replace(/\\\{/g, "{") // 将 \{ 转换为左大括号
    .replace(/\\\}/g, "}") // 将 \} 转换为右大括号
    .replace(/\\\(/g, "(") // 将 \( 转换为左小括号
    .replace(/\\\)/g, ")"); // 将 \) 转换为右小括号

  // 处理列表格式，确保列表项前后有适当的换行
  processedContent = processedContent.replace(/^(\s*)([\-\*\+])\s+/gm, "$1$2 ");
  processedContent = processedContent.replace(/^(\s*)(\d+\.)\s+/gm, "$1$2 ");

  // 确保列表项之间有适当的空行处理
  processedContent = processedContent.replace(
    /(\n\s*[\-\*\+\d\.]\s+)/g,
    "\n$1"
  );

  // 去除前后空格并返回
  return processedContent.trim();
};

// 处理发送消息
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

  emit("send-message", formData);
  file.value = [];
  inputMessage.value = "";
};

const handleSendMessage = () => {
  sendMessage();
};

const handleKeydown = (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    sendMessage();
  }
};

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return "";
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return "";
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  return `${month}/${day} ${hours}:${minutes}`;
};

// 处理文件上传相关函数
const beforeRemove = () => true;

const handleExceed = (files) => {
  file.value = [files[0]];
};

const beforeRemoveKnowledge = () => true;

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

  if (uploadFile.raw.size > 50 * 1024 * 1024) {
    ElMessage.error("文件大小不能超过50MB");
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
    ElMessage.error(
      error.response?.data?.message || "知识库文件上传失败，请重试"
    );
    knowledgeFile.value = [];
  } finally {
    uploadingKnowledge.value = false;
  }
};
</script>

<style scoped>
.chat-room {
  width: 100%;
  display: flex;
  justify-content: center;
  padding-bottom: 24px;
}

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

.chat-log {
  flex: 1;
  overflow-y: auto;
  padding-right: 10px;
  display: flex;
  flex-direction: column;
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

.chat-message {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  max-width: 90%;
}

.chat-message.ai {
  margin-right: auto;
}

.chat-message.user {
  margin-left: auto;
  flex-direction: row-reverse;
}

.avatar-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

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

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.message-bubble {
  background: var(--color-input-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
  line-height: 1.6;
  color: var(--color-text-primary);
  min-width: 0;
}

.chat-message.user .message-bubble {
  background: var(--color-user-bubble-bg);
  border-color: var(--color-user-bubble-border);
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.message-content :deep(p) {
  margin: 0;
}

.message-content :deep(pre) {
  margin-top: 12px;
  padding: 16px;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.3);
}

body.light-mode .message-content :deep(pre) {
  background: rgba(0, 0, 0, 0.06);
}

.timestamp {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.chat-input-area {
  padding-top: 20px;
  flex-shrink: 0;
}

.input-wrapper {
  display: flex;
  align-items: center;
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
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
}

.input-wrapper.disabled {
  opacity: 0.6;
}

.chat-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-upload {
  display: inline-flex;
}

.action-upload :deep(.el-upload) {
  display: inline-flex;
}

.action-btn {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s ease;
}

.action-btn:hover {
  color: var(--color-text-primary);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn :deep(svg) {
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

.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.send-btn svg {
  width: 24px;
  height: 24px;
  fill: currentColor;
}

.chat-textarea {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--color-text-primary);
  font-size: 1rem;
  line-height: 1.6;
  resize: none;
  padding: 12px;
  min-height: 48px;
}

.chat-textarea:focus {
  outline: none;
}

.chat-textarea::placeholder {
  color: var(--color-text-secondary);
}

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

@media (max-width: 768px) {
  .chat-container {
    padding: 0 16px 16px;
    min-height: calc(100vh - 120px);
  }

  .chat-message {
    max-width: 100%;
  }

  .chat-textarea {
    min-height: 60px;
  }
}
</style>
