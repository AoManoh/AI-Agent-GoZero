<template>
  <div class="page chat-page" id="chat">
    <div class="workbench">
      <ChatSidebar
        :sessions="sessions"
        :active-session-id="activeSessionId"
        :username="username"
        @new-chat="handleNewChat"
        @select-session="handleSelectSession"
      />

      <!-- Main Chat -->
      <div class="main-chat">
        <div class="mc-header">
          <div class="mc-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
            {{ activeSession?.title || 'AI 模拟面试' }}
          </div>
          <button class="btn-ghost" :disabled="isFinishing" @click="goHome">
            &larr; {{ isFinishing ? '结束中…' : '结束面试' }}
          </button>
        </div>

        <ChatMessageList
          :messages="messages"
          :isConnecting="isConnecting"
          :isPending="isPending"
        />

        <ChatInputDock
          :isConnecting="isConnecting"
          :status-label="statusLabel"
          @send-message="handleSendMessage"
          @stop-stream="handleStopStream"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import ChatSidebar from "../components/ChatSidebar.vue";
import ChatInputDock from "../components/ChatInputDock.vue";
import ChatMessageList from "../components/ChatMessageList.vue";
import { useChatStream } from "../composables/useChatStream";
import { useInterviewSessions } from "../composables/useInterviewSessions";
import { useApi } from "../composables/useApi";
import { useAuth } from "../composables/useAuth";
import { buildSessionCreatePayload } from "../utils/interviewSession";

const router = useRouter();
const route = useRoute();
const api = useApi();
const { username } = useAuth();
const isFinishing = ref(false);
const sessionReady = ref(false);
const isCreatingSession = ref(false);
const flowState = ref(null);
let flowPollTimer = null;

const initialMessages = [
  {
    role: "ai",
    type: "text",
    content: "你好，我是你的AI面试官。请问有什么需要考察的？你可以附带你的简历并说明你想要面试的技术方向与岗位。",
    time: Date.now(),
    isUser: false,
  },
];

const {
  sessions,
  activeSessionId,
  activeSession,
  activateSession,
  createSession,
  updateActiveSession,
} = useInterviewSessions(initialMessages);

const getRouteSessionId = () =>
  String(route.query.sessionId || route.query.sid || "").trim();

const ensureLocalSession = (sessionId, overrides = {}) => {
  if (!sessionId) return;
  if (sessions.value.some((session) => session.id === sessionId)) {
    activateSession(sessionId);
    if (Object.keys(overrides).length > 0) {
      updateActiveSession(overrides);
    }
  } else {
    createSession({
      id: sessionId,
      title: overrides.title || "未命名会话",
      mode: overrides.mode || "Interview Studio",
      messages: overrides.messages || initialMessages,
    });
    if (Object.keys(overrides).length > 0) {
      updateActiveSession(overrides);
    }
  }
};

const buildFallbackSessionPayload = () =>
  buildSessionCreatePayload({
    title: "技术面试",
    directionKey: route.query.direction || "",
    difficulty: route.query.difficulty || "",
    focusKeys: route.query.focus || [],
    questionKey: route.query.questionId || route.query.questionKey || "",
  });

const normalizeServerMessages = (serverMessages = []) => {
  if (!Array.isArray(serverMessages) || serverMessages.length === 0) {
    return initialMessages;
  }
  return serverMessages.map((message) => {
    const role = String(message.role || "").toLowerCase();
    const isUser = role === "user";
    const createdAt = new Date(message.createdAt || Date.now()).getTime();
    return {
      role: isUser ? "user" : "ai",
      type: "text",
      content: message.content || "",
      time: Number.isNaN(createdAt) ? Date.now() : createdAt,
      isUser,
    };
  });
};

const applySessionBootstrap = (response, fallbackSessionId = "") => {
  const session = response?.session || {};
  const sessionId = session.sessionId || fallbackSessionId;
  if (!sessionId) return;
  const nextMessages = normalizeServerMessages(response?.messages);
  flowState.value = response?.flowState || flowState.value;
  ensureLocalSession(sessionId, {
    title: session.title || "未命名会话",
    mode: session.mode || session.modeKey || "Interview Studio",
    messages: nextMessages,
    completedAt: session.completedAt || "",
    isActive: session.isActive !== false,
    config: response?.config || null,
    reportSummary: response?.reportSummary || null,
  });
  setMessages(nextMessages);
};

const ensureBackendSession = async () => {
  const sessionId = getRouteSessionId();
  if (sessionId) {
    const response = await api.user.sessionBootstrap(sessionId);
    applySessionBootstrap(response, sessionId);
    return sessionId;
  }

  const response = await api.user.createSession(buildFallbackSessionPayload());
  const nextSessionId = response?.session?.sessionId;
  if (!nextSessionId) {
    throw new Error("后端未返回面试会话 ID");
  }

  await router.replace({
    path: route.path,
    query: {
      ...route.query,
      sessionId: nextSessionId,
    },
  });
  try {
    const bootstrap = await api.user.sessionBootstrap(nextSessionId);
    applySessionBootstrap(bootstrap, nextSessionId);
  } catch {
    ensureLocalSession(nextSessionId, {
      title: response?.session?.title || "未命名会话",
      mode: response?.session?.mode || "Interview Studio",
    });
  }
  return nextSessionId;
};

const {
  messages,
  isStreaming,
  phase,
  streamError,
  startStream,
  stopStream,
  pushUserMessage,
  setMessages,
  dispose,
} = useChatStream(activeSession.value?.messages || initialMessages);

// 用户主动点击"停止生成"按钮：中断当前 SSE 流，AI 消息会自动追加"已被用户中断"标记。
const handleStopStream = () => {
  stopStream();
  pollFlowStateUntilIdle();
};

const isPending = computed(() => {
  if (!isStreaming.value) return false;
  const lastMsg = messages.value[messages.value.length - 1];
  return Boolean(lastMsg && lastMsg.isUser);
});

const busyExecutionStates = new Set(["retrieving", "generating", "persisting"]);

const serverBusy = computed(() =>
  busyExecutionStates.has(String(flowState.value?.executionState || ""))
);

const isConnecting = computed(() => isStreaming.value || serverBusy.value);

const statusLabel = computed(() => {
  if (streamError.value) return streamError.value;
  if (phase.value === "retrieving") return "检索上下文中";
  if (phase.value === "generating") return "AI 正在生成";
  if (phase.value === "persisting") return "保存本轮对话中";
  if (phase.value === "connecting") return "连接面试服务中";
  const executionState = String(flowState.value?.executionState || "");
  if (executionState === "retrieving") return "服务端正在检索上下文";
  if (executionState === "generating") return "服务端仍在生成回复";
  if (executionState === "persisting") return "服务端正在保存对话";
  if (executionState === "failed") return "上一轮处理失败，可稍后重试";
  return "";
});

const deriveTitleFromMessage = (text) => {
  if (!text) return "";
  const trimmed = String(text).trim().replace(/\s+/g, " ");
  return trimmed.length > 24 ? `${trimmed.slice(0, 24)}…` : trimmed;
};

const refreshFlowState = async () => {
  const sessionId = activeSessionId.value || getRouteSessionId();
  if (!sessionId) return null;
  const response = await api.user.sessionFlowState(sessionId);
  flowState.value = response;
  const executionState = String(response?.executionState || "");
  if (executionState !== "failed" && !busyExecutionStates.has(executionState)) {
    streamError.value = "";
  }
  if (response?.session) {
    updateActiveSession({
      completedAt: response.session.completedAt || activeSession.value?.completedAt || "",
      isActive: response.session.isActive !== false,
    });
  }
  return response;
};

const clearFlowPollTimer = () => {
  if (flowPollTimer) {
    window.clearInterval(flowPollTimer);
    flowPollTimer = null;
  }
};

const pollFlowStateUntilIdle = () => {
  clearFlowPollTimer();
  let attempts = 0;
  flowPollTimer = window.setInterval(async () => {
    attempts += 1;
    try {
      const response = await refreshFlowState();
      const executionState = String(response?.executionState || "");
      if (!busyExecutionStates.has(executionState) || attempts >= 30) {
        clearFlowPollTimer();
      }
    } catch {
      if (attempts >= 3) {
        clearFlowPollTimer();
      }
    }
  }, 2000);
};

watch(
  () => activeSessionId.value,
  () => {
    setMessages(activeSession.value?.messages || initialMessages);
  }
);

watch(
  messages,
  (next) => {
    if (!activeSession.value) return;
    const updates = { messages: next };
    if (
      activeSession.value.title === "未命名会话" ||
      !activeSession.value.title
    ) {
      const firstUser = next.find((m) => m.isUser && m.content);
      if (firstUser) {
        updates.title = deriveTitleFromMessage(firstUser.content) || "未命名会话";
      }
    }
    updateActiveSession(updates);
  },
  { deep: true }
);

const handleSendMessage = (payload) => {
  if (!sessionReady.value) {
    ElMessage.warning("面试会话正在初始化，请稍候");
    return;
  }
  if (activeSession.value?.completedAt || flowState.value?.lifecycleState === "completed") {
    ElMessage.warning("本次面试已结束，可回看记录或新建面试");
    return;
  }
  if (serverBusy.value) {
    ElMessage.warning("服务端仍在处理上一轮回复，请稍候");
    pollFlowStateUntilIdle();
    return;
  }

  const formData = new FormData();
  formData.append("message", payload.message);

  if (payload.file) {
    formData.append("file", payload.file);
  }

  pushUserMessage(payload.message);
  startStream(formData, activeSessionId.value, {
    onDone: refreshFlowState,
    onError: () => {
      refreshFlowState().finally(() => {
        if (serverBusy.value) {
          pollFlowStateUntilIdle();
        }
      });
    },
  });
};

const handleNewChat = async () => {
  if (isCreatingSession.value) return;
  const previousReadyState = sessionReady.value;
  isCreatingSession.value = true;
  sessionReady.value = false;
  try {
    const response = await api.user.createSession(buildFallbackSessionPayload());
    const sessionId = response?.session?.sessionId;
    if (!sessionId) {
      throw new Error("后端未返回面试会话 ID");
    }
    createSession({
      id: sessionId,
      title: response?.session?.title || "未命名会话",
      mode: response?.session?.mode || "Interview Studio",
      messages: initialMessages,
    });
    router.replace({
      path: route.path,
      query: {
        ...route.query,
        sessionId,
      },
    });
    try {
      const bootstrap = await api.user.sessionBootstrap(sessionId);
      applySessionBootstrap(bootstrap, sessionId);
    } catch {
      setMessages(activeSession.value?.messages || initialMessages);
    }
    sessionReady.value = true;
  } catch (error) {
    ElMessage.error(error?.message || "创建面试失败，请稍后重试");
    sessionReady.value = previousReadyState;
  } finally {
    isCreatingSession.value = false;
  }
};

const handleSelectSession = (sessionId) => {
  activateSession(sessionId);
  router.replace({
    path: route.path,
    query: {
      ...route.query,
      sessionId,
    },
  });
  api.user.sessionBootstrap(sessionId)
    .then((response) => applySessionBootstrap(response, sessionId))
    .catch(() => {});
};

async function goHome() {
  if (isFinishing.value) return;
  if (!sessionReady.value) {
    ElMessage.warning("面试会话正在初始化，请稍候");
    return;
  }
  isFinishing.value = true;
  stopStream();
  try {
    const sessionId = activeSessionId.value || getRouteSessionId();
    if (sessionId) {
      await api.user.finishSession(sessionId);
    }
    ElMessage.success("面试已结束");
    router.push({ name: "Workbench" });
  } catch (error) {
    ElMessage.error(error?.message || "结束面试失败，请稍后重试");
  } finally {
    isFinishing.value = false;
  }
}

onMounted(async () => {
  try {
    await ensureBackendSession();
    await refreshFlowState().catch(() => null);
    if (serverBusy.value) {
      pollFlowStateUntilIdle();
    }
    sessionReady.value = true;
  } catch (error) {
    sessionReady.value = false;
    ElMessage.error(error?.message || "创建面试会话失败，请稍后重试");
    router.push({ name: "WorkbenchNew" });
  }
});

onBeforeUnmount(() => {
  clearFlowPollTimer();
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
  font-size: var(--fs-md);
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
  font-size: var(--fs-sm);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
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
