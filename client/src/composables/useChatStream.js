import { ref } from "vue";
import { ElMessage } from "element-plus";
import { apiService } from "./useApi.js";

const PHASE_IDLE = "idle";
const PHASE_CONNECTING = "connecting";
const PHASE_ERROR = "error";

export function useChatStream(initialMessages = []) {
  const messages = ref([...initialMessages]);
  const isStreaming = ref(false);
  const streamRef = ref(null);
  const phase = ref(PHASE_IDLE);
  const streamError = ref("");

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
    if (state && phase.value === PHASE_IDLE) {
      phase.value = PHASE_CONNECTING;
    }
    if (!state && phase.value !== PHASE_ERROR) {
      phase.value = PHASE_IDLE;
    }
    if (!state) {
      const last = messages.value[messages.value.length - 1];
      if (last && last.isStreaming) {
        delete last.isStreaming;
      }
    }
  };

  const normalizeChunk = (raw) =>
    raw.replaceAll("\\n", "\n").replaceAll("\\r", "\r");

  const startStream = (formData, chatId, options = {}) => {
    if (isStreaming.value) {
      ElMessage.warning("AI 正在回复，请稍候");
      return;
    }

    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }

    streamError.value = "";
    phase.value = PHASE_CONNECTING;
    setStreaming(true);

    const stream = apiService.chat.interviewStream(formData, chatId);
    streamRef.value = stream;

    stream.onmessage = (data, event = "message") => {
      if (!data) return;
      if (event === "phase") {
        phase.value = String(data || PHASE_CONNECTING);
        return;
      }
      if (data === "[DONE]") {
        setStreaming(false);
        options.onDone?.();
        return;
      }
      appendAIChunk(normalizeChunk(String(data)));
    };

    stream.onerror = (error) => {
      setStreaming(false);
      phase.value = PHASE_ERROR;
      streamError.value = error?.message || "对话连接异常";
      if (error?.response?.status === 409) {
        ElMessage.warning(streamError.value || "该面试仍在生成中，请稍后再试");
      } else {
        ElMessage.error("对话连接异常，请稍后再试");
      }
      if (streamRef.value) {
        streamRef.value.close();
        streamRef.value = null;
      }
      options.onError?.(error);
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
    phase.value = PHASE_IDLE;
  };

  const setMessages = (nextMessages) => {
    if (streamRef.value) {
      streamRef.value.close();
      streamRef.value = null;
    }
    isStreaming.value = false;
    phase.value = PHASE_IDLE;
    messages.value = Array.isArray(nextMessages) ? [...nextMessages] : [];
  };

  return {
    messages,
    isStreaming,
    phase,
    streamError,
    appendAIChunk,
    setStreaming,
    startStream,
    stopStream,
    pushUserMessage,
    setMessages,
    dispose,
  };
}
