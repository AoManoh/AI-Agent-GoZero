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

  return {
    messages,
    isStreaming,
    appendAIChunk,
    setStreaming,
    startStream,
    pushUserMessage,
    dispose,
  };
}
