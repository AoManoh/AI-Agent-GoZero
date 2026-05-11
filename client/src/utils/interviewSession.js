const DIRECTION_KEY_ALIASES = {
  "go-backend": "go_backend",
  go_backend: "go_backend",
  frontend: "frontend_vue",
  frontend_vue: "frontend_vue",
  fullstack: "system_design",
  devops: "system_design",
  data: "algorithm",
  system: "system_design",
  system_design: "system_design",
  algorithm: "algorithm",
  java_backend: "java_backend",
};

const FOCUS_KEY_ALIASES = {
  project: "engineering",
  algo: "algorithm",
  design: "system_design",
  lang: "engineering",
  framework: "frontend_arch",
  behavior: "communication",
  concurrency: "concurrency",
  database: "database",
  system_design: "system_design",
  engineering: "engineering",
  network: "network",
  performance: "performance",
  algorithm: "algorithm",
  communication: "communication",
  frontend_arch: "frontend_arch",
  observability: "observability",
};

const DIFFICULTY_LEVEL_ALIASES = {
  entry: 1,
  junior: 2,
  low: 2,
  mid: 3,
  medium: 3,
  senior: 4,
  high: 4,
  expert: 5,
};

const unique = (items) => [...new Set(items)];

export const normalizeDirectionKey = (value) => {
  const key = String(value || "").trim();
  if (!key) return "";
  return DIRECTION_KEY_ALIASES[key] || key;
};

export const normalizeDifficultyLevel = (value) => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return Math.min(5, Math.max(1, Math.round(value)));
  }
  const text = String(value || "").trim().toLowerCase();
  const numeric = Number(text);
  if (Number.isFinite(numeric) && numeric > 0) {
    return Math.min(5, Math.max(1, Math.round(numeric)));
  }
  return DIFFICULTY_LEVEL_ALIASES[text] || 3;
};

export const normalizeFocusKeys = (value) => {
  const raw = Array.isArray(value)
    ? value
    : String(value || "")
        .split(",")
        .map((item) => item.trim());

  return unique(
    raw
      .map((key) => {
        const normalized = String(key || "").trim();
        if (!normalized) return "";
        return FOCUS_KEY_ALIASES[normalized] || normalized;
      })
      .filter(Boolean)
  );
};

export const buildSessionCreatePayload = ({
  title = "",
  mode = "Interview",
  directionKey = "",
  difficulty = "",
  focusKeys = [],
  interviewerStyle = "senior",
  estimatedMinutes = 30,
  questionKey = "",
} = {}) => {
  const payload = {
    mode,
    directionKey: normalizeDirectionKey(directionKey),
    difficulty: normalizeDifficultyLevel(difficulty),
    interviewerStyle,
    estimatedMinutes,
  };

  const normalizedFocusKeys = normalizeFocusKeys(focusKeys);
  if (normalizedFocusKeys.length > 0) {
    payload.focusKeys = normalizedFocusKeys;
  }

  const normalizedTitle = String(title || "").trim();
  if (normalizedTitle) {
    payload.title = normalizedTitle;
  }

  const normalizedQuestionKey = String(questionKey || "").trim();
  if (normalizedQuestionKey) {
    payload.questionKey = normalizedQuestionKey;
  }

  return payload;
};
