<template>
  <WorkbenchLayout content-mode="immersive">
    <div class="chat-page" id="chat">
      <div class="chat-frame">
        <nav class="chat-mobile-tabs" aria-label="聊天工作区">
          <button
            v-for="tab in mobileTabs"
            :key="tab.key"
            type="button"
            :class="{ active: mobileTab === tab.key }"
            @click="mobileTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </nav>

        <div class="chat-grid">
          <aside class="chat-column chat-left" :class="{ 'mobile-active': mobileTab === 'sessions' }">
            <section class="chat-panel" aria-label="面试会话窗口">
              <header class="panel-head">
                <div>
                  <h2>面试会话</h2>
                </div>
                <button class="icon-btn" type="button" title="新建面试会话" :disabled="isCreatingSession" @click="handleNewChat">
                  <span aria-hidden="true">+</span>
                </button>
              </header>

            <div class="panel-scroll">
              <div v-if="loadingSessions" class="panel-empty">正在加载会话...</div>
              <template v-else-if="groupedSessions.some((group) => group.items.length)">
                <div v-for="group in groupedSessions" :key="group.label" class="session-group">
                  <div v-if="group.items.length" class="session-group-label">{{ group.label }}</div>
                  <button
                    v-for="session in group.items"
                    :key="session.sessionId"
                    type="button"
                    class="session-item"
                    :class="{ active: session.sessionId === activeSessionId }"
                    @click="handleSelectSession(session.sessionId)"
                  >
                    <span class="session-dot" :class="{ done: Boolean(session.completedAt) }"></span>
                    <span class="session-main">
                      <strong>{{ session.title || "未命名会话" }}</strong>
                      <small>{{ sessionModeLabel(session) }} · {{ formatRelativeTime(session.lastMessageAt || session.updatedAt) }}</small>
                    </span>
                    <span class="session-count">{{ session.messageCount || 0 }}</span>
                  </button>
                </div>
              </template>
              <div v-else class="panel-empty">暂无会话，点击右上角 + 新建</div>
            </div>
          </section>

          <section class="chat-panel" aria-label="当前会话详细信息">
            <header class="panel-head">
              <div>
                <h2>会话详情</h2>
              </div>
              <span class="status-pill">{{ sessionScenarioLabel }}</span>
            </header>

            <div class="panel-scroll detail-stack">
              <div class="detail-title">
                <strong>{{ activeSession?.title || "未选择会话" }}</strong>
                <span>{{ activeSession?.sessionId ? shortId(activeSession.sessionId) : "等待初始化" }}</span>
              </div>

              <dl class="detail-list">
                <div v-for="row in sessionDetailRows" :key="row.label">
                  <dt>{{ row.label }}</dt>
                  <dd>{{ row.value }}</dd>
                </div>
              </dl>

              <div class="detail-block">
                <div class="detail-block-title">介入数据源</div>
                <div class="source-list">
                  <span v-for="source in linkedSources" :key="source.key" :class="{ active: source.active }">
                    {{ source.label }}
                  </span>
                </div>
              </div>

              <div class="detail-block">
                <div class="detail-block-title">考察侧重</div>
                <div v-if="focusLabels.length" class="focus-chips">
                  <span v-for="focus in focusLabels" :key="focus">{{ focus }}</span>
                </div>
                <p v-else class="muted">暂无侧重配置</p>
              </div>
            </div>
          </section>
        </aside>

        <main class="main-chat" :class="{ 'mobile-active': mobileTab === 'dialog' }" aria-label="对话内容">
          <header class="chat-main-head">
            <div class="chat-title-block">
              <h1>{{ activeSession?.title || "AI 模拟面试" }}</h1>
              <div class="chat-subline">
                <span>{{ sessionScenarioLabel }}</span>
                <span>{{ selectedDirectionLabel }}</span>
                <span>{{ selectedDifficultyLabel }}</span>
              </div>
            </div>
            <div class="chat-head-actions">
              <span class="status-pill" :class="executionToneClass">{{ statusLabel || "待输入" }}</span>
              <button class="text-btn" type="button" @click="goWorkbench">返回工作台</button>
              <button
                v-if="!isReadOnly"
                class="text-btn danger"
                type="button"
                :disabled="isFinishing || !sessionReady"
                @click="finishInterview"
              >
                {{ isFinishing ? "结束中..." : "结束面试" }}
              </button>
              <button v-else class="text-btn primary" type="button" @click="handleNewChat">新建面试</button>
            </div>
          </header>

          <ChatMessageList
            :messages="messages"
            :is-connecting="isConnecting"
            :is-pending="isPending"
            :readonly="isReadOnly"
          />

          <ChatInputDock
            v-if="!isReadOnly"
            :is-connecting="isStreaming"
            :disabled="serverBusy || loadingBundle || !sessionReady"
            :status-label="statusLabel"
            @send-message="handleSendMessage"
            @stop-stream="handleStopStream"
          />
          <div v-else class="chat-readonly-dock">
            <span>当前会话为回看状态</span>
            <button type="button" @click="handleNewChat">开始新面试</button>
          </div>
        </main>

        <aside class="chat-column chat-right" :class="{ 'mobile-active': mobileTab === 'status' }">
          <section class="chat-panel" aria-label="当前面试进度">
            <header class="panel-head">
              <div>
                <h2>面试进度</h2>
              </div>
              <strong class="progress-number">{{ progressPercent }}%</strong>
            </header>

            <div class="panel-scroll progress-stack">
              <div class="progress-track" aria-hidden="true">
                <span :style="{ width: `${progressPercent}%` }"></span>
              </div>

              <div class="progress-metrics">
                <div>
                  <span>已完成</span>
                  <strong>{{ completedQuestionsLabel }}</strong>
                </div>
                <div>
                  <span>对话轮次</span>
                  <strong>{{ turnCountLabel }}</strong>
                </div>
              </div>

              <div class="detail-block">
                <div class="detail-block-title">下一题</div>
                <p class="next-question">{{ nextQuestionLabel }}</p>
              </div>

              <div class="focus-progress-list">
                <div v-for="item in focusProgressItems" :key="item.key" class="focus-progress-item">
                  <div>
                    <span>{{ item.label }}</span>
                    <strong>{{ item.completedQuestions }}/{{ item.plannedQuestions }}</strong>
                  </div>
                  <i><b :style="{ width: `${item.progressPercent}%` }"></b></i>
                </div>
                <p v-if="!focusProgressItems.length" class="muted">暂无题目进度，开始对话后同步更新。</p>
              </div>
            </div>
          </section>

          <section class="chat-panel" aria-label="流程状态">
            <header class="panel-head">
              <div>
                <h2>流程状态</h2>
              </div>
              <span class="status-pill" :class="executionToneClass">{{ executionStateLabel }}</span>
            </header>

            <div class="panel-scroll flow-stack">
              <dl class="detail-list">
                <div v-for="row in flowRows" :key="row.label">
                  <dt>{{ row.label }}</dt>
                  <dd>{{ row.value }}</dd>
                </div>
              </dl>

              <div class="flow-events">
                <div class="detail-block-title">最近事件</div>
                <ol v-if="flowEvents.length">
                  <li v-for="(event, index) in flowEvents" :key="`${event.at}-${index}`">
                    <span>{{ flowEventLabel(event) }}</span>
                  </li>
                </ol>
                <p v-else class="muted">暂无流程事件</p>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>
  </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import ChatInputDock from "../components/ChatInputDock.vue";
import ChatMessageList from "../components/ChatMessageList.vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { useChatStream } from "../composables/useChatStream";
import { useApi } from "../composables/useApi";
import { buildSessionCreatePayload } from "../utils/interviewSession";

const router = useRouter();
const route = useRoute();
const api = useApi();

const sessions = ref([]);
const activeSessionId = ref("");
const sessionConfig = ref(null);
const progressState = ref(null);
const flowState = ref(null);
const reportSummary = ref(null);
const loadingSessions = ref(false);
const loadingBundle = ref(false);
const sessionReady = ref(false);
const isCreatingSession = ref(false);
const isFinishing = ref(false);
const mobileTab = ref("dialog");
const progressError = ref("");
let flowPollTimer = null;

const mobileTabs = [
  { key: "sessions", label: "会话" },
  { key: "dialog", label: "对话" },
  { key: "status", label: "状态" },
];

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
} = useChatStream([]);

const DAY_MS = 24 * 60 * 60 * 1000;
const busyExecutionStates = new Set(["retrieving", "generating", "persisting"]);

const getRouteSessionId = () => String(route.query.sessionId || route.query.sid || "").trim();

const normalizeServerMessages = (serverMessages = []) => {
  if (!Array.isArray(serverMessages)) return [];
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

const normalizeSessionItem = (session = {}) => ({
  sessionId: session.sessionId || session.id || "",
  title: session.title || "未命名会话",
  mode: session.mode || "",
  modeKey: session.modeKey || "Interview",
  messageCount: Number(session.messageCount) || 0,
  isActive: session.isActive !== false,
  createdAt: session.createdAt || "",
  updatedAt: session.updatedAt || "",
  lastMessageAt: session.lastMessageAt || "",
  completedAt: session.completedAt || "",
});

const upsertSession = (session) => {
  const item = normalizeSessionItem(session);
  if (!item.sessionId) return;
  const index = sessions.value.findIndex((entry) => entry.sessionId === item.sessionId);
  if (index >= 0) {
    sessions.value.splice(index, 1, { ...sessions.value[index], ...item });
  } else {
    sessions.value.unshift(item);
  }
};

const activeSession = computed(
  () => sessions.value.find((session) => session.sessionId === activeSessionId.value) || null
);

const buildFallbackSessionPayload = () =>
  buildSessionCreatePayload({
    title: "技术面试",
    directionKey: route.query.direction || "",
    difficulty: route.query.difficulty || "",
    focusKeys: route.query.focus || [],
    questionKey: route.query.questionId || route.query.questionKey || "",
    resumeArtifactId: route.query.resumeArtifactId || route.query.resumeId || "",
  });

const loadSessions = async () => {
  loadingSessions.value = true;
  try {
    const response = await api.user.sessions();
    const list = Array.isArray(response?.sessions) ? response.sessions : [];
    sessions.value = list.map(normalizeSessionItem);
  } finally {
    loadingSessions.value = false;
  }
};

const loadSessionBundle = async (sessionId, options = {}) => {
  const { quiet = false, refreshSessions = false } = options;
  if (!sessionId) return;
  if (!quiet) sessionReady.value = false;
  loadingBundle.value = true;
  progressError.value = "";

  try {
    const [bootstrap, progress] = await Promise.all([
      api.user.sessionBootstrap(sessionId),
      api.user.sessionProgress(sessionId, { planLimit: 8 }).catch((error) => {
        progressError.value = error?.message || "进度暂不可用";
        return null;
      }),
    ]);

    activeSessionId.value = sessionId;
    upsertSession(bootstrap?.session);
    sessionConfig.value = bootstrap?.config || null;
    flowState.value = bootstrap?.flowState || flowState.value;
    reportSummary.value = bootstrap?.reportSummary || null;
    progressState.value = progress;
    setMessages(normalizeServerMessages(bootstrap?.messages));

    if (progress?.session) upsertSession(progress.session);
    if (refreshSessions) await loadSessions().catch(() => null);
    sessionReady.value = true;
  } finally {
    loadingBundle.value = false;
  }
};

const ensureBackendSession = async () => {
  await loadSessions().catch(() => null);
  const routeSessionId = getRouteSessionId();
  if (routeSessionId) {
    await loadSessionBundle(routeSessionId);
    return routeSessionId;
  }

  const response = await api.user.createSession(buildFallbackSessionPayload());
  const nextSessionId = response?.session?.sessionId;
  if (!nextSessionId) throw new Error("后端未返回面试会话 ID");

  upsertSession(response.session);
  sessionConfig.value = response?.config || null;
  await router.replace({ path: route.path, query: { sessionId: nextSessionId } });
  await loadSessionBundle(nextSessionId, { refreshSessions: true });
  return nextSessionId;
};

const refreshFlowState = async () => {
  const sessionId = activeSessionId.value || getRouteSessionId();
  if (!sessionId) return null;
  const response = await api.user.sessionFlowState(sessionId);
  flowState.value = response;
  if (response?.session) upsertSession(response.session);
  return response;
};

const refreshProgress = async () => {
  const sessionId = activeSessionId.value || getRouteSessionId();
  if (!sessionId) return null;
  const response = await api.user.sessionProgress(sessionId, { planLimit: 8 });
  progressState.value = response;
  if (response?.session) upsertSession(response.session);
  return response;
};

const refreshAfterTurn = () => {
  const sessionId = activeSessionId.value;
  if (!sessionId) return;
  loadSessionBundle(sessionId, { quiet: true, refreshSessions: true }).catch((error) => {
    ElMessage.warning(error?.message || "会话状态刷新失败");
  });
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
        refreshProgress().catch(() => null);
        loadSessions().catch(() => null);
      }
    } catch {
      if (attempts >= 3) clearFlowPollTimer();
    }
  }, 2000);
};

const groupedSessions = computed(() => {
  const todayStart = new Date();
  todayStart.setHours(0, 0, 0, 0);
  const todayStartTs = todayStart.getTime();
  const sevenDaysAgoTs = todayStartTs - 7 * DAY_MS;
  const groups = [
    { label: "今天", items: [] },
    { label: "最近 7 天", items: [] },
    { label: "更早", items: [] },
  ];

  [...sessions.value]
    .sort((a, b) => dateValue(b.lastMessageAt || b.updatedAt) - dateValue(a.lastMessageAt || a.updatedAt))
    .forEach((session) => {
      const ts = dateValue(session.lastMessageAt || session.updatedAt || session.createdAt);
      if (ts >= todayStartTs) groups[0].items.push(session);
      else if (ts >= sevenDaysAgoTs) groups[1].items.push(session);
      else groups[2].items.push(session);
    });

  return groups;
});

const serverBusy = computed(() => busyExecutionStates.has(String(flowState.value?.executionState || "")));
const isConnecting = computed(() => isStreaming.value || serverBusy.value || loadingBundle.value);
const isPending = computed(() => {
  if (!isStreaming.value) return false;
  const lastMsg = messages.value[messages.value.length - 1];
  return Boolean(lastMsg && lastMsg.isUser);
});

const isReadOnly = computed(() => {
  const routeMode = String(route.query.mode || "").toLowerCase();
  return (
    routeMode === "replay" ||
    Boolean(activeSession.value?.completedAt) ||
    Boolean(sessionConfig.value?.completedAt) ||
    flowState.value?.lifecycleState === "completed" ||
    flowState.value?.interviewState === "end"
  );
});

const statusLabel = computed(() => {
  if (streamError.value) return streamError.value;
  if (!sessionReady.value) return "会话初始化中";
  if (isReadOnly.value) return "回看模式";
  if (phase.value === "retrieving") return "检索上下文中";
  if (phase.value === "generating") return "AI 正在生成";
  if (phase.value === "persisting") return "保存本轮对话中";
  if (phase.value === "connecting") return "连接面试服务中";
  const executionState = String(flowState.value?.executionState || "");
  if (executionState === "retrieving") return "服务端正在检索";
  if (executionState === "generating") return "服务端正在生成";
  if (executionState === "persisting") return "服务端正在保存";
  if (executionState === "failed") return "上一轮处理失败";
  return messages.value.length ? "等待下一轮输入" : "等待第一轮输入";
});

const executionToneClass = computed(() => {
  const executionState = String(flowState.value?.executionState || "");
  if (streamError.value || executionState === "failed") return "danger";
  if (isStreaming.value || busyExecutionStates.has(executionState)) return "busy";
  if (isReadOnly.value) return "done";
  return "";
});

const sessionScenarioLabel = computed(() =>
  sessionConfig.value?.scenarioLabel || scenarioLabel(sessionConfig.value?.scenarioType)
);
const selectedDirectionLabel = computed(() => sessionConfig.value?.directionLabel || "未指定方向");
const selectedDifficultyLabel = computed(() => sessionConfig.value?.difficultyLabel || "未指定难度");
const focusLabels = computed(() =>
  (sessionConfig.value?.focusAreas || []).map((item) => item.label || item.key).filter(Boolean)
);

const sessionDetailRows = computed(() => [
  { label: "模式", value: sessionModeLabel(activeSession.value) },
  { label: "场景", value: sessionScenarioLabel.value },
  { label: "方向", value: selectedDirectionLabel.value },
  { label: "难度", value: selectedDifficultyLabel.value },
  { label: "面试官", value: sessionConfig.value?.interviewerStyleLabel || "资深技术官" },
  { label: "追问深度", value: sessionConfig.value?.followUpDepth || "N+3" },
  { label: "预计时长", value: `${sessionConfig.value?.estimatedMinutes || 30} 分钟` },
  { label: "消息数", value: `${activeSession.value?.messageCount || messages.value.length || 0} 条` },
]);

const linkedSources = computed(() => {
  const starter = sessionConfig.value?.starterSource || "none";
  return [
    { key: "config", label: "会话配置", active: true },
    { key: "bank", label: "题库题目", active: starter === "bank" },
    { key: "resume", label: "简历画像", active: starter === "resume_plan" || Boolean(sessionConfig.value?.resumeArtifactId) },
    { key: "knowledge", label: "知识库检索", active: true },
  ];
});

const progressPercent = computed(() => {
  const value = Number(progressState.value?.progressPercent ?? sessionConfig.value?.progressPercent ?? 0);
  return Math.max(0, Math.min(100, Number.isFinite(value) ? Math.round(value) : 0));
});
const completedQuestionsLabel = computed(() => {
  const completed = Number(progressState.value?.completedQuestions) || 0;
  const total = Number(progressState.value?.totalQuestions) || 0;
  if (!total) return progressError.value ? "暂不可用" : "0/0";
  return `${completed}/${total}`;
});
const turnCountLabel = computed(() => {
  const userTurns = Number(progressState.value?.userTurns) || 0;
  const assistantTurns = Number(progressState.value?.assistantTurns) || 0;
  const flowTurns = Number(flowState.value?.turnCount) || 0;
  return `${Math.max(userTurns, flowTurns)}/${assistantTurns}`;
});
const nextQuestionLabel = computed(() => {
  if (isReadOnly.value) return "本次会话已结束，可回看完整对话。";
  const next = progressState.value?.nextQuestion;
  if (next?.title || next?.prompt) return next.title || next.prompt;
  if (messages.value.length === 0) return "等待候选人输入第一轮背景或回答。";
  return "等待 AI 根据当前回答给出下一轮追问。";
});
const focusProgressItems = computed(() =>
  Array.isArray(progressState.value?.focusProgress) ? progressState.value.focusProgress : []
);
const executionStateLabel = computed(() => executionLabel(flowState.value?.executionState));
const flowRows = computed(() => [
  { label: "面试阶段", value: interviewStateLabel(flowState.value?.interviewState) },
  { label: "生命周期", value: lifecycleLabel(flowState.value?.lifecycleState) },
  { label: "执行状态", value: executionStateLabel.value },
  { label: "记忆范围", value: memoryScopeLabel(flowState.value?.memoryScope) },
  { label: "通道", value: laneLabel(flowState.value?.lane) },
  { label: "最近原因", value: flowState.value?.lastReason || "暂无" },
]);
const flowEvents = computed(() =>
  Array.isArray(flowState.value?.events) ? flowState.value.events.slice(0, 8) : []
);

const handleStopStream = () => {
  stopStream();
  pollFlowStateUntilIdle();
};

const handleSendMessage = (payload) => {
  if (!sessionReady.value || !activeSessionId.value) {
    ElMessage.warning("面试会话正在初始化，请稍候");
    return;
  }
  if (isReadOnly.value) {
    ElMessage.warning("当前会话为回看状态，请新建面试后继续");
    return;
  }
  if (serverBusy.value) {
    ElMessage.warning("服务端仍在处理上一轮回复，请稍候");
    pollFlowStateUntilIdle();
    return;
  }

  const formData = new FormData();
  formData.append("message", payload.message);
  if (payload.file) formData.append("file", payload.file);

  pushUserMessage(payload.message);
  mobileTab.value = "dialog";
  startStream(formData, activeSessionId.value, {
    onDone: refreshAfterTurn,
    onError: () => {
      refreshFlowState().finally(() => {
        if (serverBusy.value) pollFlowStateUntilIdle();
      });
    },
  });
};

const handleNewChat = async () => {
  if (isCreatingSession.value) return;
  isCreatingSession.value = true;
  sessionReady.value = false;
  try {
    stopStream();
    const response = await api.user.createSession(buildSessionCreatePayload({ title: "技术面试" }));
    const sessionId = response?.session?.sessionId;
    if (!sessionId) throw new Error("后端未返回面试会话 ID");
    upsertSession(response.session);
    sessionConfig.value = response?.config || null;
    await router.replace({ path: "/chat", query: { sessionId } });
    await loadSessionBundle(sessionId, { refreshSessions: true });
    mobileTab.value = "dialog";
  } catch (error) {
    ElMessage.error(error?.message || "创建面试失败，请稍后重试");
    sessionReady.value = true;
  } finally {
    isCreatingSession.value = false;
  }
};

const handleSelectSession = async (sessionId) => {
  if (!sessionId || sessionId === activeSessionId.value) return;
  stopStream();
  activeSessionId.value = sessionId;
  mobileTab.value = "dialog";
  await router.replace({ path: "/chat", query: { sessionId } });
  loadSessionBundle(sessionId).catch((error) => {
    ElMessage.error(error?.message || "加载会话失败");
  });
};

const finishInterview = async () => {
  if (isFinishing.value || !activeSessionId.value) return;
  isFinishing.value = true;
  stopStream();
  try {
    await api.user.finishSession(activeSessionId.value);
    await router.replace({ path: "/chat", query: { sessionId: activeSessionId.value, mode: "replay" } });
    await loadSessionBundle(activeSessionId.value, { quiet: true, refreshSessions: true });
    ElMessage.success("面试已结束");
  } catch (error) {
    ElMessage.error(error?.message || "结束面试失败，请稍后重试");
  } finally {
    isFinishing.value = false;
  }
};

const goWorkbench = () => {
  router.push({ name: "Workbench" });
};

watch(
  () => getRouteSessionId(),
  (nextSessionId) => {
    if (!nextSessionId || nextSessionId === activeSessionId.value || !sessionReady.value) return;
    loadSessionBundle(nextSessionId).catch((error) => {
      ElMessage.error(error?.message || "加载会话失败");
    });
  }
);

onMounted(async () => {
  try {
    await ensureBackendSession();
    await refreshFlowState().catch(() => null);
    if (serverBusy.value) pollFlowStateUntilIdle();
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

function dateValue(value) {
  if (!value) return 0;
  const ts = new Date(value).getTime();
  return Number.isNaN(ts) ? 0 : ts;
}

function formatRelativeTime(value) {
  const ts = dateValue(value);
  if (!ts) return "暂无时间";
  const diff = Date.now() - ts;
  if (diff < 60 * 1000) return "刚刚";
  if (diff < 60 * 60 * 1000) return `${Math.max(1, Math.floor(diff / 60000))} 分钟前`;
  if (diff < DAY_MS) return `${Math.floor(diff / (60 * 60 * 1000))} 小时前`;
  if (diff < 2 * DAY_MS) return "昨天";
  return new Date(ts).toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}

function shortId(value = "") {
  const text = String(value || "");
  if (text.length <= 12) return text || "N/A";
  return `${text.slice(0, 8)}...${text.slice(-4)}`;
}

function scenarioLabel(value = "") {
  return value === "question_practice" ? "题库练习" : "模拟面试";
}

function sessionModeLabel(session = {}) {
  const raw = String(session?.mode || session?.modeKey || "").trim();
  const normalized = raw.toLowerCase().replace(/\s+/g, "_");
  const labels = {
    interview: "模拟面试",
    interview_studio: "模拟面试",
    mock_interview: "模拟面试",
    practice: "练习模式",
    question_practice: "题库练习",
    resume_plan: "简历面试",
  };
  if (!raw) return sessionScenarioLabel.value || "模拟面试";
  if (labels[normalized]) return labels[normalized];
  return /[A-Za-z]/.test(raw) ? sessionScenarioLabel.value || "模拟面试" : raw;
}

function interviewStateLabel(value = "") {
  const labels = { start: "开场", question: "核心问题", follow_up: "追问", evaluate: "评估", end: "结束" };
  return labels[value] || "开场";
}

function lifecycleLabel(value = "") {
  const labels = { active: "进行中", completed: "已完成", errored: "异常" };
  return labels[value] || "进行中";
}

function executionLabel(value = "") {
  const labels = { idle: "空闲", retrieving: "检索中", generating: "生成中", persisting: "保存中", failed: "失败" };
  return labels[value] || "空闲";
}

function memoryScopeLabel(value = "") {
  return value === "user" ? "用户级" : "会话级";
}

function laneLabel(value = "") {
  const labels = { interview: "面试", research: "研究", memory: "记忆", coach: "教练" };
  return labels[value] || "面试";
}

function flowEventLabel(event = {}) {
  if (event.type === "state.transition") {
    return `${interviewStateLabel(event.from)} -> ${interviewStateLabel(event.to)}`;
  }
  if (event.type === "turn.recorded") return event.role === "user" ? "候选人消息已记录" : "AI 回复已记录";
  if (event.type) return event.type;
  return event.reason || "流程事件";
}
</script>

<style scoped>
.chat-page {
  position: relative;
  z-index: 1;
  width: calc(100% - clamp(32px, 5vw, 96px));
  max-width: none;
  height: calc(100svh - 80px - 80px - clamp(72px, 8vh, 108px));
  min-height: 620px;
  margin: clamp(16px, 2.2vh, 28px) auto clamp(22px, 3vh, 34px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
  background: transparent;
}

.chat-frame {
  width: 100%;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.chat-grid {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(240px, 20%) minmax(520px, 1fr) minmax(280px, 22%);
  gap: clamp(14px, 1.35vw, 24px);
}

.chat-column {
  min-width: 0;
  min-height: 0;
  display: grid;
  grid-template-rows: minmax(0, 1fr) minmax(0, 1fr);
  gap: clamp(14px, 1.4vw, 22px);
}

.chat-panel,
.main-chat {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.52), rgba(7, 8, 11, 0.44)) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.035), rgba(255, 255, 255, 0.08)) border-box;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 18px 50px rgba(0, 0, 0, 0.16);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.chat-panel:hover,
.main-chat:hover {
  border-color: rgba(220, 155, 90, 0.15);
}

.chat-panel {
  display: flex;
  flex-direction: column;
}

.main-chat {
  position: relative;
  display: flex;
  flex-direction: column;
}

.panel-head,
.chat-main-head {
  flex-shrink: 0;
  min-height: 58px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: clamp(14px, 1.4vw, 20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.038), rgba(255, 255, 255, 0.012));
}

.panel-head > div {
  min-width: 0;
}

.panel-head h2,
.chat-main-head h1 {
  margin: 0;
  color: var(--t);
  font: 700 var(--fs-md) var(--sans);
  letter-spacing: 0;
}

.chat-main-head h1 {
  max-width: min(52vw, 680px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--fs-xl);
}

.panel-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: clamp(12px, 1.3vw, 18px);
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.18) transparent;
}

.icon-btn,
.text-btn,
.session-item,
.chat-mobile-tabs button,
.chat-readonly-dock button {
  font-family: var(--sans);
}

.icon-btn {
  width: 32px;
  height: 32px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.06);
  color: var(--t);
  cursor: pointer;
  transition: border-color 0.18s ease, background-color 0.18s ease, transform 0.18s ease;
}

.icon-btn:hover:not(:disabled) {
  border-color: rgba(220, 155, 90, 0.26);
  background: rgba(220, 155, 90, 0.10);
  transform: translateY(-1px);
}

.icon-btn:disabled,
.text-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.text-btn {
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.04);
  color: var(--t2);
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.18s ease, border-color 0.18s ease, background-color 0.18s ease, transform 0.18s ease;
}

.text-btn:hover:not(:disabled) {
  color: var(--t);
  border-color: rgba(220, 155, 90, 0.24);
  background: rgba(220, 155, 90, 0.08);
  transform: translateY(-1px);
}

.text-btn.primary {
  border-color: rgba(245, 180, 92, 0.30);
  color: #ffd98c;
}

.text-btn.danger {
  border-color: rgba(239, 68, 68, 0.24);
  color: #fecaca;
}

.session-group + .session-group {
  margin-top: 18px;
}

.session-group-label {
  margin: 0 0 8px;
  color: rgba(255, 255, 255, 0.42);
  font: 600 var(--fs-3xs) var(--mono);
}

.session-item {
  width: 100%;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--t2);
  text-align: left;
  cursor: pointer;
}

.session-item:hover,
.session-item.active {
  background: rgba(255, 255, 255, 0.055);
  border-color: rgba(255, 255, 255, 0.08);
  color: var(--t);
}

.session-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #8bd49c;
  box-shadow: 0 0 0 4px rgba(139, 212, 156, 0.08);
}

.session-dot.done {
  background: #82aaff;
  box-shadow: 0 0 0 4px rgba(130, 170, 255, 0.08);
}

.session-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.session-main strong,
.detail-title strong {
  overflow: hidden;
  color: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--fs-sm);
}

.session-main small,
.detail-title span {
  overflow: hidden;
  color: var(--t3);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--fs-xs);
}

.session-count {
  color: var(--t3);
  font: 600 var(--fs-2xs) var(--mono);
}

.panel-empty {
  min-height: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--t3);
  text-align: center;
  font-size: var(--fs-sm);
}

.detail-stack,
.progress-stack,
.flow-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.detail-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.detail-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-list div {
  display: grid;
  grid-template-columns: minmax(76px, 32%) minmax(0, 1fr);
  gap: 10px;
}

.detail-list dt {
  color: var(--t3);
  font-size: var(--fs-xs);
}

.detail-list dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--t);
  font-size: var(--fs-sm);
}

.detail-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-block-title {
  color: rgba(255, 255, 255, 0.50);
  font: 600 var(--fs-xs) var(--sans);
}

.source-list,
.focus-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.source-list span,
.focus-chips span {
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: var(--t3);
  background: rgba(255, 255, 255, 0.035);
  font-size: var(--fs-xs);
}

.source-list span.active {
  border-color: rgba(245, 180, 92, 0.28);
  color: #ffd98c;
  background: rgba(245, 180, 92, 0.08);
}

.muted {
  margin: 0;
  color: var(--t3);
  font-size: var(--fs-sm);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: var(--radius-pill);
  border: 1px solid rgba(255, 255, 255, 0.10);
  color: var(--t2);
  background: rgba(255, 255, 255, 0.045);
  font: 600 var(--fs-xs) var(--sans);
  white-space: nowrap;
}

.status-pill.busy {
  border-color: rgba(130, 170, 255, 0.26);
  color: #bfdbfe;
  background: rgba(130, 170, 255, 0.08);
}

.status-pill.done {
  border-color: rgba(139, 212, 156, 0.24);
  color: #bbf7d0;
  background: rgba(139, 212, 156, 0.08);
}

.status-pill.danger {
  border-color: rgba(239, 68, 68, 0.24);
  color: #fecaca;
  background: rgba(239, 68, 68, 0.08);
}

.chat-title-block {
  min-width: 0;
}

.chat-subline {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--t3);
  font-size: var(--fs-xs);
}

.chat-subline span:not(:last-child)::after {
  content: "/";
  margin-left: 8px;
  color: rgba(255, 255, 255, 0.22);
}

.chat-head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.chat-readonly-dock {
  flex-shrink: 0;
  min-height: 78px;
  margin: 0 clamp(18px, 3vw, 42px) clamp(18px, 3vw, 28px);
  padding: 14px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.045);
  color: var(--t2);
}

.chat-readonly-dock button {
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid rgba(245, 180, 92, 0.28);
  border-radius: var(--radius-sm);
  background: rgba(245, 180, 92, 0.08);
  color: #ffd98c;
  cursor: pointer;
}

.progress-number {
  color: var(--t);
  font: 800 var(--fs-2xl) var(--mono);
}

.progress-track,
.focus-progress-item i {
  position: relative;
  overflow: hidden;
  height: 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
}

.progress-track span,
.focus-progress-item b {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #ffd98c, #82aaff, #ffd98c);
  background-size: 200% 100%;
  animation: progress-flow 3.6s linear infinite;
}

.progress-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.progress-metrics div {
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.035);
}

.progress-metrics span {
  display: block;
  color: var(--t3);
  font-size: var(--fs-xs);
}

.progress-metrics strong {
  color: var(--t);
  font: 800 var(--fs-lg) var(--mono);
}

.next-question {
  margin: 0;
  color: var(--t2);
  font-size: var(--fs-sm);
  line-height: 1.65;
}

.focus-progress-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.focus-progress-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.focus-progress-item div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--t2);
  font-size: var(--fs-xs);
}

.flow-events li {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--t2);
  font-size: var(--fs-xs);
}

.flow-events li::before {
  content: "";
  width: 6px;
  height: 6px;
  flex-shrink: 0;
  border-radius: 999px;
  background: rgba(220, 155, 90, 0.85);
  box-shadow: 0 0 0 4px rgba(220, 155, 90, 0.08);
}

.focus-progress-item strong {
  color: var(--t);
  font-family: var(--mono);
}

.flow-events ol {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.chat-mobile-tabs {
  display: none;
}

@keyframes progress-flow {
  from {
    background-position: 0% 50%;
  }
  to {
    background-position: 200% 50%;
  }
}

@media (max-width: 1180px) {
  .chat-grid {
    grid-template-columns: minmax(220px, 23%) minmax(420px, 1fr) minmax(240px, 24%);
  }

  .chat-head-actions {
    flex-wrap: wrap;
    justify-content: flex-end;
  }
}

@media (max-width: 900px) {
  .chat-page {
    width: calc(100% - 20px);
    height: min(760px, calc(100svh - 178px));
    min-height: 560px;
    margin-top: 14px;
    margin-bottom: 24px;
  }

  .chat-frame {
    padding: 0;
  }

  .chat-mobile-tabs {
    flex-shrink: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 6px;
    margin-bottom: 8px;
  }

  .chat-mobile-tabs button {
    min-height: 36px;
    border: 1px solid rgba(255, 255, 255, 0.09);
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.04);
    color: var(--t3);
  }

  .chat-mobile-tabs button.active {
    color: var(--t);
    border-color: rgba(245, 180, 92, 0.28);
    background: rgba(245, 180, 92, 0.08);
  }

  .chat-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .chat-column,
  .main-chat {
    display: none;
  }

  .chat-column.mobile-active {
    display: grid;
  }

  .main-chat.mobile-active {
    display: flex;
  }

  .chat-main-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .chat-main-head h1 {
    max-width: 100%;
  }

  .chat-head-actions {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 560px) {
  .panel-head,
  .chat-main-head {
    padding: 12px;
  }

  .panel-scroll {
    padding: 12px;
  }

  .progress-metrics {
    grid-template-columns: 1fr;
  }

  .chat-readonly-dock {
    align-items: flex-start;
    flex-direction: column;
    margin: 0 12px 12px;
  }
}
</style>
