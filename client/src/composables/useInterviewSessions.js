import { computed, ref } from "vue";

const STORAGE_KEY = "gozero-ai-local-interview-sessions";
const ACTIVE_KEY = "gozero-ai-active-interview-session";
const MAX_SESSION_COUNT = 10;

const normalizeMessageContent = (content = "") => {
  const normalized = String(content || "");
  if (
    normalized.includes("GoZero-AI 工作伙伴") ||
    normalized.includes("模拟面试工作台。把岗位方向、简历或追问交给我。") ||
    normalized.includes("你好，这里是模拟面试工作台。把岗位方向、简历或追问交给我。")
  ) {
    return "输入岗位方向或上传简历，我会从第一轮追问开始。";
  }
  return normalized;
};

const cloneMessages = (messages) =>
  (messages || []).map((message) => ({
    role: message.role || (message.isUser ? "user" : "ai"),
    type: message.type || "text",
    content: normalizeMessageContent(message.content),
    time: message.time || Date.now(),
    isUser: Boolean(message.isUser),
  }));

const createSessionId = () => {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const readSessions = () => {
  if (typeof window === "undefined") {
    return [];
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    return [];
  }
};

const writeSessions = (sessions) => {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions));
};

const readActiveSessionId = () => {
  if (typeof window === "undefined") {
    return "";
  }
  return window.localStorage.getItem(ACTIVE_KEY) || "";
};

const writeActiveSessionId = (sessionId) => {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(ACTIVE_KEY, sessionId);
};

const normalizeMode = (value) => value || "Interview Studio";

const buildSession = ({
  id = createSessionId(),
  title,
  summary,
  focus,
  mode,
  messages,
  updatedAt,
  createdAt,
}) => {
  const normalizedMode = normalizeMode(mode || focus);
  const normalizedTitle =
    !title ||
    title === "新的 Agent 会话" ||
    title === "新的工作台会话" ||
    title === "新的模拟会话" ||
    title === "新的模拟面试"
      ? "未命名会话"
      : title;
  return {
    id,
    title: normalizedTitle,
    summary: summary || "从一个真实问题开始，逐步拉开上下文。",
    focus: focus || normalizedMode,
    mode: normalizedMode,
    updatedAt: updatedAt || Date.now(),
    createdAt: createdAt || Date.now(),
    messages: cloneMessages(messages),
  };
};

const normalizeStoredSessions = (storedSessions) =>
  (storedSessions || []).map((session) =>
    buildSession({
      ...session,
      id: session.id,
      title: session.title,
      summary: session.summary,
      focus: session.focus,
      mode: session.mode,
      messages: session.messages,
      updatedAt: session.updatedAt,
      createdAt: session.createdAt,
    })
  );

export function useInterviewSessions(defaultMessages = []) {
  const storedSessions = readSessions();
  const sessions = ref(
    storedSessions.length > 0
      ? normalizeStoredSessions(storedSessions)
      : [buildSession({ messages: defaultMessages, mode: "Interview Studio" })]
  );

  const initialActiveId = readActiveSessionId();
  const hasStoredActive = sessions.value.some((session) => session.id === initialActiveId);
  const activeSessionId = ref(hasStoredActive ? initialActiveId : sessions.value[0].id);

  const persist = () => {
    writeSessions(sessions.value.slice(0, MAX_SESSION_COUNT));
    writeActiveSessionId(activeSessionId.value);
  };

  const activeSession = computed(
    () => sessions.value.find((session) => session.id === activeSessionId.value) || sessions.value[0]
  );

  const activateSession = (sessionId) => {
    if (!sessions.value.some((session) => session.id === sessionId)) {
      return;
    }
    activeSessionId.value = sessionId;
    persist();
  };

  const createSession = (overrides = {}) => {
    const session = buildSession({
      id: overrides.id || createSessionId(),
      title: overrides.title,
      summary: overrides.summary,
      focus: overrides.focus,
      mode: overrides.mode,
      messages: overrides.messages || defaultMessages,
    });

    sessions.value = [session, ...sessions.value].slice(0, MAX_SESSION_COUNT);
    activeSessionId.value = session.id;
    persist();
    return session;
  };

  const updateActiveSession = (payload = {}) => {
    sessions.value = sessions.value.map((session) => {
      if (session.id !== activeSessionId.value) {
        return session;
      }

      return {
        ...session,
        ...payload,
        messages: payload.messages ? cloneMessages(payload.messages) : session.messages,
        updatedAt: payload.updatedAt || Date.now(),
      };
    });
    persist();
  };

  const removeSession = (sessionId) => {
    const remainingSessions = sessions.value.filter((session) => session.id !== sessionId);
    sessions.value =
      remainingSessions.length > 0
        ? remainingSessions
        : [buildSession({ messages: defaultMessages })];

    if (!sessions.value.some((session) => session.id === activeSessionId.value)) {
      activeSessionId.value = sessions.value[0].id;
    }
    persist();
  };

  persist();

  return {
    sessions,
    activeSessionId,
    activeSession,
    activateSession,
    createSession,
    updateActiveSession,
    removeSession,
    cloneMessages,
  };
}
