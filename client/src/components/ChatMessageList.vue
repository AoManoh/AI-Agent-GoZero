<template>
  <div class="mc-scroll" ref="messagesContainer">
    <div class="msg-container">
      <div v-for="(msg, index) in messages" :key="index" class="msg" :class="msg.isUser ? 'usr' : 'ai'">
        <div class="m-avatar">
          <span>{{ msg.isUser ? 'U' : 'AI' }}</span>
        </div>
        <div class="m-body">
          <div class="m-name">
            {{ msg.isUser ? 'You' : 'AI 面试官' }}
          </div>
          <div class="m-content">
            <p v-if="msg.isUser">{{ msg.content }}</p>
            <MarkdownMessage v-else :content="processMessageContent(msg.content || '')" />
          </div>
        </div>
      </div>

      <!-- Loading indicator / Streaming cursor indicator -->
      <div v-if="isConnecting && isPending" class="msg ai">
        <div class="m-avatar"><span>AI</span></div>
        <div class="m-body">
          <div class="m-name">AI 面试官</div>
          <div class="m-content">
            <span class="dcur"></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from 'vue';
import MarkdownMessage from "./MarkdownMessage.vue";

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  },
  isConnecting: {
    type: Boolean,
    default: false
  },
  isPending: {
    type: Boolean,
    default: false
  }
});

const messagesContainer = ref(null);

const scrollToBottom = () => {
  nextTick(() => {
    const container = messagesContainer.value;
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  });
};

watch(() => props.messages, () => {
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
</script>

<style scoped>
.mc-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 40px 0 clamp(190px, 18vh, 240px);
  scrollbar-width: none;
  scroll-behavior: smooth;
}

.mc-scroll::-webkit-scrollbar {
  display: none;
}

.msg-container {
  max-width: 840px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 48px;
  padding: 0 32px;
}

/* Document Flow Messages */
.msg {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.m-avatar {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font: 700 12px var(--mono);
  flex-shrink: 0;
}

.msg.ai .m-avatar {
  background: var(--t);
  color: var(--bg);
}

.msg.usr .m-avatar {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--b);
  color: var(--t3);
}

.m-body {
  flex: 1;
  padding-top: 2px;
  min-width: 0;
}

.m-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--t);
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.m-name span.badge {
  font-size: 11px;
  padding: 2px 8px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-xs);
  font-weight: normal;
  color: var(--t2);
  font-family: var(--mono);
}

.m-content {
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.85);
}

.m-content p {
  margin: 0 0 1em 0;
}

.m-content :deep(p:last-child) {
  margin-bottom: 0;
}

.dcur {
  display: inline-block;
  width: 8px;
  height: 15px;
  background: var(--t);
  vertical-align: middle;
  animation: cb 1s step-end infinite;
}

@keyframes cb {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

@media (max-width: 768px) {
  .msg-container {
    padding: 0 16px;
  }
}
</style>
