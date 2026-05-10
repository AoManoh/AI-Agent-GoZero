<template>
  <div class="input-dock">
    <div class="input-hints">
      <span v-if="fileName" class="file-tag">已附加：{{ fileName }}</span>
      <span v-if="uploadingKnowledge" class="hint">知识库文件上传中...</span>
    </div>
    
    <div class="floating-inp" :class="{ disabled: isConnecting }">
      <div class="chat-actions">
        <el-upload
          v-model:file-list="fileList"
          :limit="1"
          :auto-upload="false"
          :show-file-list="false"
          :on-exceed="handleExceed"
          :disabled="isConnecting"
        >
          <button class="action-btn" type="button" title="上传附件" :disabled="isConnecting">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="m18.375 12.739-7.693 7.693a4.5 4.5 0 0 1-6.364-6.364l10.94-10.94A3.375 3.375 0 1 1 18.374 7.5l-1.497 1.498a1.5 1.5 0 0 1-2.121-2.121l1.497-1.498 2.121-2.121a3.375 3.375 0 0 1 4.773 4.773l-10.94 10.94a4.5 4.5 0 1 1-6.364-6.364l7.693-7.693a.75.75 0 0 1 1.06 1.06z" />
            </svg>
          </button>
        </el-upload>
        <el-upload
          v-model:file-list="knowledgeFileList"
          :limit="1"
          accept=".pdf"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleKnowledgeFileChange"
          :on-exceed="handleExceedKnowledge"
          :disabled="isConnecting || uploadingKnowledge"
        >
          <button class="action-btn" type="button" title="上传到知识库" :disabled="isConnecting || uploadingKnowledge">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6-2.292m0 0V3.75m0 12.558a8.966 8.966 0 0 1-6-2.292V3.75a8.966 8.966 0 0 1 6 2.292m0 12.558a8.966 8.966 0 0 0 6-2.292V3.75a8.966 8.966 0 0 0-6 2.292" />
            </svg>
          </button>
        </el-upload>
      </div>

      <textarea
        ref="textareaRef"
        v-model="inputMessage"
        class="inp-ta"
        placeholder="输入你的回答或追问... (Shift+Enter换行)"
        :disabled="isConnecting"
        @input="autoResizeTextarea"
        @keydown="handleKeydown"
      ></textarea>

      <!-- 流式中：显示停止按钮（红色方块），点击中断 SSE -->
      <button
        v-if="isConnecting"
        class="inp-send inp-stop"
        type="button"
        title="停止生成"
        @click="$emit('stop-stream')"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
      </button>
      <!-- 空闲时：显示发送按钮 -->
      <button
        v-else
        class="inp-send"
        type="button"
        title="发送"
        @click="handleSendMessage"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { ElMessage } from "element-plus";
import { apiService } from "../composables/useApi";

const props = defineProps({
  isConnecting: {
    type: Boolean,
    default: false,
  }
});

const emit = defineEmits(["send-message", "stop-stream"]);

const fileList = ref([]);
const knowledgeFileList = ref([]);
const inputMessage = ref("");
const textareaRef = ref(null);
const uploadingKnowledge = ref(false);

const fileName = computed(() => (fileList.value[0]?.name ? fileList.value[0].name : ""));

const autoResizeTextarea = () => {
  const textarea = textareaRef.value;
  if (!textarea) return;
  textarea.style.height = '24px';
  const newHeight = Math.min(textarea.scrollHeight, 200);
  textarea.style.height = `${newHeight}px`;
  textarea.style.overflowY = textarea.scrollHeight > 200 ? 'auto' : 'hidden';
};

const handleExceed = (files) => {
  fileList.value = [files[0]];
};

const handleExceedKnowledge = (files) => {
  knowledgeFileList.value = [files[0]];
};

const handleKnowledgeFileChange = async (uploadFile) => {
  if (uploadFile.status !== "ready" || !uploadFile.raw) return;

  if (uploadFile.raw.type !== "application/pdf") {
    ElMessage.error("只支持PDF格式的文件");
    knowledgeFileList.value = [];
    return;
  }

  if (uploadFile.raw.size > 50 * 1024 * 1024) {
    ElMessage.error("文件大小不能超过50MB");
    knowledgeFileList.value = [];
    return;
  }

  uploadingKnowledge.value = true;
  const loadingMessage = ElMessage({ message: "正在上传知识库文件...", type: "info", duration: 0 });

  try {
    const formData = new FormData();
    formData.append("file", uploadFile.raw);
    await apiService.chat.knowledgeUpload(formData);
    loadingMessage.close();
    ElMessage.success("知识库文件上传成功");
    knowledgeFileList.value = [];
  } catch (error) {
    loadingMessage.close();
    // core.js 拦截器已把后端 message 规范化到 error.message，无需再读 error.response.data。
    ElMessage.error(error?.message || "知识库文件上传失败，请重试");
    knowledgeFileList.value = [];
  } finally {
    uploadingKnowledge.value = false;
  }
};

const handleSendMessage = () => {
  if (props.isConnecting) {
    ElMessage.warning("AI 正在回复，请稍候");
    return;
  }

  const content = inputMessage.value.trim();
  if (!content) {
    ElMessage.warning("请输入内容");
    return;
  }

  const payload = {
    message: content,
    file: fileList.value.length > 0 ? fileList.value[0].raw : null
  };

  emit("send-message", payload);

  fileList.value = [];
  inputMessage.value = "";
  
  if (textareaRef.value) {
    textareaRef.value.style.height = `24px`;
  }
};

const handleKeydown = (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    handleSendMessage();
  }
};
</script>

<style scoped>
.input-dock {
  position: absolute;
  bottom: 32px;
  left: 0;
  right: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  pointer-events: none;
  padding: 0 32px;
}

.input-hints {
  width: 100%;
  max-width: 840px;
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
  pointer-events: auto;
}

.file-tag {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--t2);
}

.hint {
  font-size: 12px;
  color: var(--t3);
  display: flex;
  align-items: center;
}

.floating-inp {
  width: 100%;
  max-width: 840px;
  background: rgba(18, 18, 24, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-lg);
  padding: 14px 16px;
  display: flex;
  gap: 12px;
  align-items: flex-end;
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  pointer-events: auto;
  transition: border-color 0.3s;
}

.floating-inp.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.floating-inp:focus-within {
  border-color: rgba(255, 255, 255, 0.4);
}

.chat-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-bottom: 2px;
}

.action-btn {
  background: transparent;
  border: none;
  color: var(--t2);
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn:hover {
  color: var(--t);
}

.action-btn svg {
  width: 20px;
  height: 20px;
}

.inp-ta {
  flex: 1;
  background: none;
  border: none;
  resize: none;
  font: 15px var(--sans);
  color: var(--t);
  line-height: 1.6;
  height: 24px;
  max-height: 200px;
  outline: none;
  padding: 0;
  margin: 0;
}

.inp-ta::placeholder {
  color: rgba(255, 255, 255, 0.3);
}

.inp-send {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--t);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #000;
  transition: transform 0.2s;
  flex-shrink: 0;
}

.inp-send:hover:not(:disabled) {
  transform: scale(1.05);
}

.inp-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 停止按钮：红色背景表明"中断"语义，与发送的白色按钮在视觉上明显区分 */
.inp-stop {
  background: #ef4444;
  color: #fff;
}

.inp-stop:hover {
  background: #dc2626;
  transform: scale(1.05);
}

@media (max-width: 768px) {
  .input-dock {
    padding: 0 16px;
  }
}
</style>
