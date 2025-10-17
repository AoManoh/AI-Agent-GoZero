<template>
  <div class="chat-container">
    <!-- 聊天记录区域 -->
    <div class="chat-messages" ref="messagesContainer">
      <div
        v-for="(msg, index) in messages"
        :key="index"
        class="message-wrapper"
      >
        <!-- AI消息 -->
        <div v-if="!msg.isUser" class="message ai-message" :class="[msg.type]">
          <div class="avatar ai-avatar">
            <AiAvatarFallback :type="aiType" />
          </div>
          <div class="message-bubble">
            <!-- 使用 XMarkdown 渲染 AI 消息 -->
            <XMarkdown
              :markdown="processMessageContent(msg.content || '')"
              :enableLatex="true"
              :enableBreaks="true"
              :allowHtml="false"
              class="message-content"
            />
            <div class="message-time">{{ formatTime(msg.time) }}</div>
          </div>
        </div>

        <!-- 用户消息 -->
        <div v-else class="message user-message" :class="[msg.type]">
          <div class="message-bubble">
            <div class="message-content">{{ msg.content }}</div>
            <div class="message-time">{{ formatTime(msg.time) }}</div>
          </div>
          <div class="avatar user-avatar">
            <img src="http://pic.aomanoh.com/note/20250903035824302.jpg" alt="用户头像" class="avatar-image" />
          </div>
        </div>
      </div>
    </div>

    <!-- 使用 Sender 组件作为输入区域 -->
    <div class="chat-input-container">
      <Sender
        v-model="inputMessage"
        variant="updown"
        :disabled="connectionStatus === 'connecting'"
        :loading="connectionStatus === 'connecting'"
        submit-type="enter"
        :auto-size="{ minRows: 3, maxRows: 4 }"
        clearable
        placeholder="请输入消息..."
        class="chat-sender"
        @submit="handleSendMessage"
      >
        <template #prefix>
          <div
            style="
              display: flex;
              align-items: center;
              gap: 8px;
              flex-wrap: wrap;
            "
          >
            <el-upload
              v-model:file-list="file"
              :limit="1"
              style="
                height: 24px;
                width: auto;
                padding: 0 8px;
                border: 1px solid #ccc;
                border-radius: 16px;
              "
              :before-remove="beforeRemove"
              :auto-upload="false"
              :on-exceed="handleExceed"
              :show-file-list="false"
            >
              <el-icon style="height: 24px"><Paperclip /></el-icon>
              <el-text
                v-if="file[0]?.name"
                style="margin-left: 4px"
                size="small"
                >{{ file[0]?.name }}</el-text
              >
            </el-upload>
            <el-upload
              v-model:file-list="knowledgeFile"
              :limit="1"
              accept=".pdf"
              style="
                height: 24px;
                width: auto;
                padding: 0 8px;
                border: 1px solid #ff6b35;
                border-radius: 16px;
                background-color: #fff5f2;
              "
              :before-remove="beforeRemoveKnowledge"
              :auto-upload="false"
              :on-exceed="handleExceedKnowledge"
              :show-file-list="false"
              :on-change="handleKnowledgeFileChange"
            >
              <el-icon style="height: 24px; color: #ff6b35"
                ><Document
              /></el-icon>
              <el-text
                v-if="knowledgeFile[0]?.name"
                style="margin-left: 4px; color: #ff6b35"
                size="small"
                >{{ knowledgeFile[0]?.name }}</el-text
              >
              <el-text
                v-else
                style="margin-left: 4px; color: #ff6b35"
                size="small"
                >知识库PDF</el-text
              >
            </el-upload>
            <div
              :class="{ isSelect }"
              style="
                display: flex;
                align-items: center;
                gap: 4px;
                padding: 2px 12px;
                border: 1px solid silver;
                border-radius: 15px;
                cursor: pointer;
                font-size: 12px;
              "
              @click="isSelect = !isSelect"
            >
              <el-icon><ElementPlus /></el-icon>
              <span>深度思考</span>
            </div>
          </div>
        </template>

        <!-- <template #action-list>
          <div style="display: flex; align-items: center; gap: 8px">
            <el-button round color="#626aef">
              <el-icon><Promotion /></el-icon>
            </el-button>
          </div>
        </template> -->
      </Sender>
    </div>
  </div>
</template>

<script setup>
import { ElementPlus, Paperclip, Document } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { ref } from "vue";
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

const file = ref([]);
const knowledgeFile = ref([]);
const isSelect = ref(false);
const emit = defineEmits(["send-message", "upload-knowledge"]);

const inputMessage = ref("");
const messagesContainer = ref(null);

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
const handleSendMessage = (message) => {
  if (!message.trim()) return;

  // 创建 FormData 格式
  const formData = new FormData();
  formData.append("message", message);

  // 只有当文件存在时才添加文件
  if (file.value && file.value.length > 0 && file.value[0]) {
    formData.append("file", file.value[0].raw);
  }

  emit("send-message", formData);
  file.value = [];
  inputMessage.value = "";
};

// 格式化时间
const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
};

// 处理文件上传相关函数
const beforeRemove = (uploadFile, uploadFiles) => {
  return true; // 允许删除
};

const handleExceed = (files, uploadFiles) => {
  file.value = [files[0]];
};

// 处理知识库PDF上传相关函数
const beforeRemoveKnowledge = (uploadFile, uploadFiles) => {
  return true; // 允许删除
};

const handleExceedKnowledge = (files, uploadFiles) => {
  knowledgeFile.value = [files[0]];
};

const handleKnowledgeFileChange = async (uploadFile, uploadFiles) => {
  if (uploadFile.status === "ready" && uploadFile.raw) {
    // 验证文件类型
    if (uploadFile.raw.type !== "application/pdf") {
      ElMessage.error("只支持PDF格式的文件");
      knowledgeFile.value = [];
      return;
    }

    // 验证文件大小（限制为500KB）
    if (uploadFile.raw.size > 500 * 1024) {
      ElMessage.error("文件大小不能超过500KB");
      knowledgeFile.value = [];
      return;
    }

    try {
      // 显示上传中提示
      const loadingMessage = ElMessage({
        message: "正在上传知识库文件...",
        type: "info",
        duration: 0, // 不自动关闭
      });

      // 创建FormData并上传
      const formData = new FormData();
      formData.append("file", uploadFile.raw);

      // 调用上传接口
      await uploadKnowledge(formData);

      // 关闭加载提示
      loadingMessage.close();

      // 显示上传成功提示
      ElMessage.success("知识库文件上传成功");
    } catch (error) {
      // 关闭加载提示（如果存在）
      ElMessage.closeAll();

      // 显示错误提示
      ElMessage.error(
        error.response?.data?.message || "知识库文件上传失败，请重试"
      );

      // 清空文件列表
      knowledgeFile.value = [];
    }
  }
};
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  background-color: #f5f5f5;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  padding-bottom: 100px; /* 为输入框留出更多空间 */
  display: flex;
  flex-direction: column;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 92px; /* 与输入框高度相匹配 */
}

.message-wrapper {
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  width: 100%;
}

.message {
  display: flex;
  align-items: flex-start;
  max-width: 85%;
  margin-bottom: 8px;
}

.user-message {
  margin-left: auto; /* 用户消息靠右 */
  flex-direction: row; /* 正常顺序，先气泡后头像 */
}

.ai-message {
  margin-right: auto; /* AI消息靠左 */
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar {
  margin-left: 8px; /* 用户头像在右侧，左边距 */
}

.ai-avatar {
  margin-right: 8px; /* AI头像在左侧，右边距 */
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #007bff;
  color: white;
  font-weight: bold;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.message-bubble {
  padding: 12px;
  border-radius: 18px;
  position: relative;
  word-wrap: break-word;
  min-width: 100px; /* 最小宽度 */
}

.user-message .message-bubble {
  background-color: #007bff;
  color: white;
  border-bottom-right-radius: 4px;
  text-align: left;
}

.ai-message .message-bubble {
  background-color: #e9e9eb;
  color: #333;
  border-bottom-left-radius: 4px;
  text-align: left;
}

.message-content {
  font-size: 16px;
  line-height: 1.5;
}

/* XMarkdown 组件专用样式 */
.ai-message .message-content :deep(div) {
  font-family: inherit;
  font-size: 16px;
  line-height: 1.5;
  color: #333;
  background: transparent;
}

/* 标题样式 */
.ai-message .message-content :deep(h1),
.ai-message .message-content :deep(h2),
.ai-message .message-content :deep(h3),
.ai-message .message-content :deep(h4),
.ai-message .message-content :deep(h5),
.ai-message .message-content :deep(h6) {
  margin-top: 0.5em;
  margin-bottom: 0.5em;
  color: #333;
  font-weight: bold;
}

/* 段落样式 */
.ai-message .message-content :deep(p) {
  margin-top: 0;
  margin-bottom: 0.8em;
  word-wrap: break-word;
}

/* 列表样式 - 关键修复 */
.ai-message .message-content :deep(ul),
.ai-message .message-content :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
  list-style-position: outside;
}

.ai-message .message-content :deep(ul) {
  list-style-type: disc;
}

.ai-message .message-content :deep(ol) {
  list-style-type: decimal;
}

.ai-message .message-content :deep(li) {
  margin-bottom: 0.3em;
  padding-left: 0.2em;
  word-wrap: break-word;
  display: list-item;
}

.ai-message .message-content :deep(ul ul),
.ai-message .message-content :deep(ol ul) {
  list-style-type: circle;
  margin-top: 0.2em;
  margin-bottom: 0.2em;
}

.ai-message .message-content :deep(ul ul ul),
.ai-message .message-content :deep(ol ul ul) {
  list-style-type: square;
}

/* 代码块样式 */
.ai-message .message-content :deep(pre) {
  background-color: #f6f8fa;
  border-radius: 6px;
  padding: 12px;
  overflow-x: auto;
  margin: 0.5em 0;
}

.ai-message .message-content :deep(code) {
  background-color: rgba(175, 184, 193, 0.2);
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-size: 85%;
  font-family: "Courier New", monospace;
}

.ai-message .message-content :deep(pre code) {
  background-color: transparent;
  padding: 0;
  font-size: 14px;
}

/* 引用块样式 */
.ai-message .message-content :deep(blockquote) {
  border-left: 4px solid #dfe2e5;
  padding-left: 1em;
  margin: 0.5em 0;
  color: #6a737d;
  font-style: italic;
}

/* 表格样式 */
.ai-message .message-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 0.5em 0;
}

.ai-message .message-content :deep(th),
.ai-message .message-content :deep(td) {
  border: 1px solid #dfe2e5;
  padding: 8px 12px;
  text-align: left;
}

.ai-message .message-content :deep(th) {
  background-color: #f6f8fa;
  font-weight: bold;
}

/* 链接样式 */
.ai-message .message-content :deep(a) {
  color: #0366d6;
  text-decoration: none;
}

.ai-message .message-content :deep(a:hover) {
  text-decoration: underline;
}

/* 强调样式 */
.ai-message .message-content :deep(strong) {
  font-weight: bold;
}

.ai-message .message-content :deep(em) {
  font-style: italic;
}

/* 水平线样式 */
.ai-message .message-content :deep(hr) {
  border: none;
  height: 1px;
  background-color: #dfe2e5;
  margin: 1em 0;
}

.message-time {
  font-size: 12px;
  opacity: 0.7;
  margin-top: 4px;
  text-align: right;
}

.chat-input-container {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  /* background-color: white; */
  /* border-top: 1px solid #e0e0e0; */
  z-index: 100;
  padding: 16px;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  background-color: #fff;
}

.chat-sender {
  width: 100%;
}

/* Sender 组件样式优化 */
.chat-sender :deep(.el-textarea__inner) {
  border-radius: 20px;
  border: 1px solid #ddd;
  padding: 12px 16px;
  font-size: 16px;
  resize: none;
  transition: border-color 0.3s;
}

.chat-sender :deep(.el-textarea__inner):focus {
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.1);
}

.chat-sender :deep(.el-button) {
  border-radius: 20px;
  padding: 0 20px;
  font-size: 16px;
}

.chat-sender :deep(.el-button--primary) {
  background-color: #007bff;
  border-color: #007bff;
}

.chat-sender :deep(.el-button--primary):hover {
  background-color: #0069d9;
  border-color: #0062cc;
}

/* 动画效果 */
.ai-answer {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .message {
    max-width: 95%;
  }

  .message-content {
    font-size: 15px;
  }

  .chat-input-container {
    padding: 12px;
  }

  .chat-messages {
    bottom: 84px;
    padding-bottom: 90px;
  }
}

@media (max-width: 480px) {
  .avatar {
    width: 32px;
    height: 32px;
  }

  .message-bubble {
    padding: 10px;
  }

  .message-content {
    font-size: 14px;
  }

  .chat-input-container {
    padding: 10px;
  }

  .chat-messages {
    bottom: 76px;
    padding-bottom: 80px;
  }
}

/* 连续消息气泡样式 */
.ai-message + .ai-message {
  margin-top: 4px;
}

.ai-message + .ai-message .avatar {
  visibility: hidden;
}

.ai-message + .ai-message .message-bubble {
  border-top-left-radius: 10px;
}
.isSelect {
  color: #626aef;
  border: 1px solid #626aef !important;
  border-radius: 15px;
  padding: 3px 12px;
  font-weight: 700;
}
</style>
