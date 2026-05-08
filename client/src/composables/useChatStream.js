import { ref } from "vue";
import { ElMessage } from "element-plus";
import { chatWithLoveApp } from "../api/index.js";

export function useChatStream(initialMessages = []) {
  const messages = ref([...initialMessages]);
  const isStreaming = ref(false);
  const streamRef = ref(null);

  const appendAIChunk = (content) => {
    if (!content) return;
    const last = messages.value[messages.value.length - 1];
    if (last && !last.isUser && last.isStreaming) {
      last.content += content;
      return;
    }
    messages.value.push({
      role: "ai",
      type: "text",
      content,
      time: Date.now(),
      isUser: false,
      isStreaming: true,
    });
  };

  const setStreaming = (state) => {
    isStreaming.value = state;
    if (!state) {
      const last = messages.value[messages.value.length - 1];
      if (last && last.isStreaming) {
        delete last.isStreaming;
      }
    }
  };

  const normalizeChunk = (raw) =>
    raw.replaceAll("\\n", "\n").replaceAll("\\r", "\r");

  const startStream = (formData, chatId) => {
    if (isStreaming.value) {
      ElMessage.warning("AI 正在回复，请稍候");
      return;
    }

    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }

    formData.append("chatId", chatId);
    setStreaming(true);

    const stream = chatWithLoveApp(formData, chatId);
    streamRef.value = stream;

    stream.onmessage = (data) => {
      if (!data) return;
      if (data === "[DONE]") {
        setStreaming(false);
        return;
      }
      appendAIChunk(normalizeChunk(String(data)));
    };

    stream.onerror = () => {
      setStreaming(false);
      ElMessage.error("对话连接异常，请稍后再试");
      if (streamRef.value) {
        streamRef.value.close();
        streamRef.value = null;
      }
    };
  };

  const pushUserMessage = (content) => {
    messages.value.push({
      role: "user",
      type: "text",
      content,
      time: Date.now(),
      isUser: true,
    });
  };

  const dispose = () => {
    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }
  };

  // stopStream 由用户主动触发（点击"停止"按钮），与 dispose 区别在于：
  // - 关闭浏览器侧 SSE 连接（后端 net/http 检测到 client gone 后会停止写入并关闭 OpenAI 上游 stream）
  // - 给最后一条 AI 消息追加"已被用户中断"标记，便于用户区分截断状态
  // - 重置 isStreaming，让发送按钮重新可用
  const stopStream = () => {
    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }
    const last = messages.value[messages.value.length - 1];
    if (last && !last.isUser && last.isStreaming) {
      last.content = (last.content || "") + "\n\n_（已被用户中断）_";
      delete last.isStreaming;
    }
    isStreaming.value = false;
  };

  const setMessages = (nextMessages) => {
    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }
    isStreaming.value = false;
    messages.value = Array.isArray(nextMessages) ? [...nextMessages] : [];
  };

  return {
    messages,
    isStreaming,
    appendAIChunk,
    setStreaming,
    startStream,
    stopStream,
    pushUserMessage,
    setMessages,
    dispose,
  };
}
