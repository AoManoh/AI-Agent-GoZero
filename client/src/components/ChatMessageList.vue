<template>
  <div class="mc-scroll" ref="messagesContainer">
    <div class="msg-container">
      <div v-if="readonly" class="replay-banner">
        <span>回放模式</span>
        <strong>当前会话已只读，可复盘对话或查看报告状态。</strong>
      </div>

      <div v-if="messages.length === 0" class="msg-empty">
        <strong>暂无对话内容</strong>
        <span>选择会话或新建面试后，对话会显示在这里。</span>
      </div>

      <article
        v-for="(msg, index) in messages"
        :key="index"
        class="msg"
        :class="msg.isUser ? 'usr' : 'ai'"
      >
        <div class="m-avatar">
          <span>{{ msg.isUser ? '你' : 'AI' }}</span>
        </div>
        <div class="m-body">
          <div class="m-head">
            <span class="m-name">{{ msg.isUser ? '你' : 'AI 面试官' }}</span>
            <time v-if="msg.time" class="m-time">{{ formatTime(msg.time) }}</time>
            <span v-if="msg.isStreaming" class="m-live">生成中</span>
          </div>
          <div class="m-content">
            <MarkdownMessage :content="processMessageContent(msg.content || '')" />
          </div>
        </div>
      </article>

      <article v-if="isConnecting && isPending" class="msg ai">
        <div class="m-avatar"><span>AI</span></div>
        <div class="m-body">
          <div class="m-head">
            <span class="m-name">AI 面试官</span>
            <span class="m-live">等待响应</span>
          </div>
          <div class="m-content pending">
            <span class="dcur"></span>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup>
import { nextTick, onMounted, ref, watch } from "vue";
import MarkdownMessage from "./MarkdownMessage.vue";

const props = defineProps({
  messages: {
    type: Array,
    default: () => [],
  },
  isConnecting: {
    type: Boolean,
    default: false,
  },
  isPending: {
    type: Boolean,
    default: false,
  },
  readonly: {
    type: Boolean,
    default: false,
  },
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

watch(() => props.messages, scrollToBottom, { deep: true });

onMounted(scrollToBottom);

const formatTime = (value) => {
  const ts = typeof value === "number" ? value : new Date(value).getTime();
  if (!ts || Number.isNaN(ts)) return "";
  return new Date(ts).toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
};

const processMessageContent = (content) => {
  if (!content) return "";
  let processedContent = String(content);
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

  processedContent = processedContent.replace(/\\u([0-9a-fA-F]{4})/g, (match, hex) =>
    String.fromCharCode(parseInt(hex, 16))
  );

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
  min-height: 0;
  overflow-y: auto;
  padding: 28px 0 clamp(180px, 18vh, 230px);
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.14) transparent;
  scroll-behavior: smooth;
}

.msg-container {
  width: min(100%, 900px);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 0 clamp(18px, 3vw, 34px);
}

.replay-banner,
.msg-empty {
  padding: 14px 16px;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: space-between;
  border: 1px solid rgba(220, 155, 90, 0.18);
  border-radius: var(--radius-md);
  background: rgba(220, 155, 90, 0.08);
  color: var(--t2);
  font-size: var(--fs-xs);
}

.replay-banner span {
  color: rgba(255, 224, 172, 0.95);
  font: 700 var(--fs-3xs) var(--mono);
  text-transform: uppercase;
}

.msg-empty {
  flex-direction: column;
  align-items: flex-start;
  border-color: rgba(255, 255, 255, 0.10);
  background: rgba(255, 255, 255, 0.035);
}

.msg-empty strong,
.replay-banner strong {
  color: var(--t);
  font-weight: 600;
}

.msg {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.m-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font: 800 var(--fs-3xs) var(--mono);
  flex-shrink: 0;
}

.msg.ai .m-avatar {
  background: rgba(220, 155, 90, 0.14);
  border: 1px solid rgba(220, 155, 90, 0.36);
  color: rgba(255, 224, 172, 0.96);
}

.msg.usr .m-avatar {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--t2);
}

.m-body {
  flex: 1;
  min-width: 0;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  background: rgba(11, 12, 16, 0.58);
}

.msg.usr .m-body {
  border-color: rgba(220, 155, 90, 0.20);
  background: rgba(220, 155, 90, 0.08);
}

.m-head {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.m-name {
  color: var(--t);
  font: 700 var(--fs-xs) var(--sans);
}

.m-time {
  color: rgba(255, 255, 255, 0.38);
  font: 500 var(--fs-3xs) var(--mono);
}

.m-live {
  margin-left: auto;
  padding: 2px 7px;
  border-radius: var(--radius-pill);
  background: rgba(220, 155, 90, 0.13);
  color: rgba(255, 224, 172, 0.92);
  font: 700 var(--fs-3xs) var(--mono);
}

.m-content {
  color: rgba(255, 255, 255, 0.86);
}

.m-content.pending {
  min-height: 20px;
}

.dcur {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: rgba(255, 224, 172, 0.92);
  vertical-align: middle;
  animation: cb 1s step-end infinite;
}

@keyframes cb {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

@media (max-width: 768px) {
  .msg {
    gap: 12px;
  }

  .m-avatar {
    width: 28px;
    height: 28px;
  }

  .replay-banner {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
