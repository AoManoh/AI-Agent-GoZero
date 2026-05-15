<template>
  <!--
    WorkbenchNew：新建面试配置向导（对应设计图 2）。
    布局：顶部步骤条 + 纵向配置区 + 底部确认条。
    交互：用户在一页完成方向 / 简历 / 难度 / 侧重点选择，点“开始面试”创建后端会话。
  -->
  <WorkbenchLayout>
    <div class="wb-new-content">
      <section class="wb-new-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>新建面试</span>
        </div>
        <p class="wb-new-sub">方向、难度、简历和侧重点会作为会话配置传给后端面试官策略。</p>
      </section>

      <ol class="wb-stepper" aria-label="配置步骤">
        <li
          v-for="(step, idx) in steps"
          :key="step.key"
          class="wb-step"
          :class="{ 'wb-step-active': currentStep === idx, 'wb-step-done': currentStep > idx }"
        >
          <span class="wb-step-num">{{ idx + 1 }}</span>
          <span class="wb-step-label">{{ step.label }}</span>
        </li>
      </ol>

      <div class="wb-new-shell">
        <section class="wb-config-panel" aria-label="面试配置">
          <div class="wb-new-form">
            <fieldset class="wb-block" data-step="direction">
              <legend class="wb-visually-hidden">选择方向</legend>
              <div class="wb-block-headline">
                <span class="wb-block-tag">01</span>
                <div>
                  <h2 class="wb-block-title">面试方向</h2>
                  <p class="wb-block-desc">题库范围与面试官策略</p>
                </div>
              </div>
              <div class="wb-block-body wb-direction-grid">
                <button
                  v-for="dir in directions"
                  :key="dir.key"
                  type="button"
                  class="wb-dir"
                  :class="{ 'wb-dir-active': form.direction === dir.key }"
                  @click="selectDirection(dir)"
                >
                  <span class="wb-dir-dot" :style="{ background: dir.color }" aria-hidden="true"></span>
                  <span class="wb-dir-name">{{ dir.label }}</span>
                  <span class="wb-dir-tags">{{ formatDirectionMeta(dir) }}</span>
                  <span v-if="form.direction === dir.key" class="wb-choice-check" aria-hidden="true">✓</span>
                </button>
              </div>
            </fieldset>

            <fieldset class="wb-block">
              <legend class="wb-visually-hidden">关联简历</legend>
              <div class="wb-block-headline">
                <span class="wb-block-tag">02</span>
                <div>
                  <h2 class="wb-block-title">简历</h2>
                  <p class="wb-block-desc">可选上下文</p>
                </div>
              </div>
              <div class="wb-block-body wb-resume-row">
                <button
                  v-for="resume in resumes"
                  :key="resume.id"
                  type="button"
                  class="wb-resume"
                  :class="{ 'wb-resume-active': form.resumeArtifactId === resume.id }"
                  @click="form.resumeArtifactId = resume.id"
                >
                  <div class="wb-resume-icon" aria-hidden="true">
                    <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                      <path d="M3.5 1.5h6.5L13 4.5v10H3.5z" />
                      <path d="M10 1.5V4.5h3" />
                      <line x1="5.5" y1="8" x2="10.5" y2="8" stroke-linecap="round" />
                      <line x1="5.5" y1="10.5" x2="10.5" y2="10.5" stroke-linecap="round" />
                    </svg>
                  </div>
                  <div class="wb-resume-meta">
                    <div class="wb-resume-name">{{ resume.name }}</div>
                    <div class="wb-resume-info">{{ resume.info }}</div>
                  </div>
                  <div class="wb-resume-tag" v-if="resume.primary">主用</div>
                </button>
                <button type="button" class="wb-resume wb-resume-add" @click="goToResumePage">
                  <div class="wb-resume-icon wb-resume-icon-add" aria-hidden="true">+</div>
                  <div class="wb-resume-meta">
                    <div class="wb-resume-name">添加新简历</div>
                    <div class="wb-resume-info">PDF，最大 10MB</div>
                  </div>
                </button>
              </div>
            </fieldset>

            <fieldset class="wb-block">
              <legend class="wb-visually-hidden">难度等级</legend>
              <div class="wb-block-headline">
                <span class="wb-block-tag">03</span>
                <div>
                  <h2 class="wb-block-title">难度</h2>
                  <p class="wb-block-desc">追问强度</p>
                </div>
              </div>
              <div class="wb-block-body wb-difficulty">
                <button
                  v-for="(diff, idx) in difficulties"
                  :key="diff.key"
                  type="button"
                  class="wb-diff-btn"
                  :class="{ 'wb-diff-btn-active': form.difficultyIdx === idx }"
                  @click="form.difficultyIdx = idx"
                >
                  <span class="wb-diff-name">{{ diff.label }}</span>
                  <span class="wb-diff-desc">{{ diff.desc }}</span>
                </button>
              </div>
            </fieldset>

            <fieldset class="wb-block">
              <legend class="wb-visually-hidden">考察侧重</legend>
              <div class="wb-block-headline">
                <span class="wb-block-tag">04</span>
                <div>
                  <h2 class="wb-block-title">考察侧重</h2>
                  <p class="wb-block-desc">多选能力维度</p>
                </div>
              </div>
              <div class="wb-block-body wb-focus-grid">
                <label
                  v-for="focus in focusOptions"
                  :key="focus.key"
                  class="wb-focus"
                  :class="{ 'wb-focus-active': form.focus.includes(focus.key) }"
                >
                  <input
                    type="checkbox"
                    :value="focus.key"
                    :checked="form.focus.includes(focus.key)"
                    class="wb-focus-input"
                    @change="toggleFocus(focus.key)"
                  />
                  <span class="wb-focus-check" aria-hidden="true">
                    <svg viewBox="0 0 16 16" fill="none">
                      <polyline points="3,8 7,12 13,4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                  </span>
                  <span class="wb-focus-label">{{ focus.label }}</span>
                </label>
              </div>
            </fieldset>
          </div>
        </section>
      </div>

      <!-- 底部 summary bar -->
      <div class="wb-summary-bar">
        <div class="wb-summary-text">
          <div class="wb-summary-line">
            <span class="wb-summary-label">方向</span>
            <span class="wb-summary-value">{{ selectedDirectionLabel || '未选择' }}</span>
          </div>
          <div class="wb-summary-sep" aria-hidden="true"></div>
          <div class="wb-summary-line">
            <span class="wb-summary-label">难度</span>
            <span class="wb-summary-value">{{ selectedDifficultyLabel }}</span>
          </div>
          <div class="wb-summary-sep" aria-hidden="true"></div>
          <div class="wb-summary-line">
            <span class="wb-summary-label">重点</span>
            <span class="wb-summary-value">{{ focusSummaryLabel }}</span>
          </div>
          <div class="wb-summary-sep" aria-hidden="true"></div>
          <div class="wb-summary-line">
            <span class="wb-summary-label">简历</span>
            <span class="wb-summary-value">{{ selectedResumeLabel }}</span>
          </div>
          <div class="wb-summary-sep" aria-hidden="true"></div>
          <div class="wb-summary-line">
            <span class="wb-summary-label">预计</span>
            <span class="wb-summary-value">{{ selectedEstimatedMinutes }} 分钟</span>
          </div>
        </div>
        <button
          type="button"
          class="wb-summary-cta"
          :disabled="!canStart || isStarting"
          @click="startInterview"
        >
          <span>{{ isStarting ? "创建中…" : "开始面试" }}</span>
          <span class="wb-summary-arrow" aria-hidden="true">→</span>
        </button>
      </div>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";
import { buildSessionCreatePayload } from "../utils/interviewSession";

const router = useRouter();
const route = useRoute();

// === 步骤定义（视觉性 stepper，仅展示进度，不强制顺序） ===
const steps = [
  { key: "direction", label: "方向", desc: "技术方向" },
  { key: "resume", label: "简历", desc: "可选关联" },
  { key: "difficulty", label: "难度", desc: "追问强度" },
  { key: "focus", label: "侧重", desc: "能力维度" },
];

// 当前所在步骤：根据已填字段推断（任何用户操作都会推进 step）
const currentStep = computed(() => {
  if (!form.value.direction) return 0;
  if (form.value.difficultyIdx === null) return 2;
  return 3;
});

// === 方向色板（key → 语义色点） ===
// 后端 interviewPresets 返回的 InterviewDirectionPreset 不包含颜色，这里本地维护 key → color 映射。
// 未命中时退中性灰，与 WorkbenchKnowledge 色点调色板同源。
const DIRECTION_COLOR_MAP = {
  "go-backend": "#4cd6a8",
  go_backend: "#4cd6a8",
  java_backend: "#f59f68",
  frontend: "#6eb6ff",
  frontend_vue: "#6eb6ff",
  fullstack: "#b599ff",
  devops: "#ffd770",
  data: "#ff9966",
  system: "rgba(255, 255, 255, 0.55)",
  system_design: "#b599ff",
  algorithm: "#ff9966",
};
const pickDirectionColor = (key) => DIRECTION_COLOR_MAP[key] || "rgba(255, 255, 255, 0.55)";

// === 方向选项（本地 fallback；onMounted 异步接入 interviewPresets 后覆盖）===
const directions = ref([
  {
    key: "go_backend",
    label: "Go 后端",
    color: "#4cd6a8",
    tags: ["concurrency", "database", "system_design"],
    focusKeys: ["concurrency", "database", "system_design", "engineering"],
  },
  {
    key: "java_backend",
    label: "Java 后端",
    color: "#f59f68",
    tags: ["system_design", "database", "network"],
    focusKeys: ["system_design", "database", "network", "engineering"],
  },
  {
    key: "frontend_vue",
    label: "前端 Vue",
    color: "#6eb6ff",
    tags: ["frontend_arch", "performance", "engineering"],
    focusKeys: ["frontend_arch", "performance", "engineering"],
  },
  {
    key: "system_design",
    label: "系统设计",
    color: "#b599ff",
    tags: ["system_design", "database", "network"],
    focusKeys: ["system_design", "database", "network", "observability"],
  },
  {
    key: "algorithm",
    label: "算法基础",
    color: "#ff9966",
    tags: ["algorithm", "communication"],
    focusKeys: ["algorithm", "communication"],
  },
]);

const selectedDirectionLabel = computed(() => {
  return directions.value.find((d) => d.key === form.value.direction)?.label || "";
});

// === 简历选项（初值空数组，loadResumes 从后端拉真实数据）===
// 历史上这里有两条硬编码示例简历，新用户没上传也会看到假简历。
// 现在严格走「0 简历 → step 2 只显示『添加新简历』按钮 → 引导跳 /workbench/resume 上传」。
const resumes = ref([]);

const goToResumePage = () => {
  router.push({ path: "/workbench/resume" });
};

// === 难度等级（本地 fallback；onMounted 异步接入 interviewPresets 后覆盖）===
// 后端 InterviewDifficultyPreset 是 5 级（Level 1-5），label 分别为 入门/初级/中级/资深/专家。
// 本地 string key 仅用于 UI 状态与 createSession 传参，最终通过
// utils/interviewSession.js 的 normalizeDifficultyLevel(key) 反向转回 int 1-5。
// 必须与 DIFFICULTY_LEVEL_ALIASES (utils/interviewSession.js) 严格反向一致，
// 否则 UI 显示的难度与提交给后端的难度会错位（详见 docs/code-review/2026-05-10-home-auth-workbench-e2e.md #5）。
const DIFFICULTY_KEY_BY_LEVEL = {
  1: "entry",   // 入门
  2: "junior",  // 初级
  3: "mid",     // 中级
  4: "senior",  // 资深
  5: "expert",  // 专家
};
const difficulties = ref([
  { key: "entry", level: 1, label: "入门", desc: "确认基本概念和术语理解" },
  { key: "junior", level: 2, label: "初级", desc: "常见场景与基础工程经验" },
  { key: "mid", level: 3, label: "中级", desc: "机制、取舍与故障复盘" },
  { key: "senior", level: 4, label: "资深", desc: "架构判断与工程边界" },
  { key: "expert", level: 5, label: "专家", desc: "高压技术面 · 系统化论证" },
]);

const selectedDifficultyLabel = computed(() => {
  if (form.value.difficultyIdx === null) return "未选择";
  return difficulties.value[form.value.difficultyIdx]?.label || "";
});

const selectedResumeLabel = computed(() => {
  const resume = resumes.value.find((item) => item.id === form.value.resumeArtifactId);
  return resume?.name || "未关联";
});

const focusSummaryLabel = computed(() =>
  form.value.focus.length > 0 ? `${form.value.focus.length} 项` : "默认侧重"
);

const selectedEstimatedMinutes = ref(30);

const formatDirectionMeta = (dir) => {
  const count = Number(dir?.questionCount) || 0;
  const range = Array.isArray(dir?.difficultyRange) && dir.difficultyRange.length >= 2
    ? `难度 ${dir.difficultyRange[0]}-${dir.difficultyRange[1]}`
    : "难度 1-5";
  if (count > 0) {
    return `题库 ${count} 题 / ${range}`;
  }
  return Array.isArray(dir?.tags) ? dir.tags.join(" · ") : range;
};

// === 重点方向（本地 fallback；onMounted 异步接入 interviewPresets.focusOptions 后覆盖）===
const focusOptions = ref([
  { key: "project", label: "项目深度追问" },
  { key: "algo", label: "算法 + 数据结构" },
  { key: "design", label: "系统设计" },
  { key: "lang", label: "语言基础" },
  { key: "framework", label: "框架原理" },
  { key: "behavior", label: "软实力 / STAR" },
]);

const toggleFocus = (key) => {
  const i = form.value.focus.indexOf(key);
  if (i >= 0) {
    form.value.focus.splice(i, 1);
  } else {
    form.value.focus.push(key);
  }
};

const defaultFocusForDirection = (dir) => {
  return Array.isArray(dir?.focusKeys) ? dir.focusKeys.slice(0, 3) : [];
};

const selectDirection = (dir) => {
  if (!dir?.key) return;
  form.value.direction = dir.key;
  form.value.focus = defaultFocusForDirection(dir);
};

// === 表单状态 ===
const form = ref({
  direction: "",
  resumeArtifactId: String(route.query.resumeArtifactId || route.query.resumeId || ""),
  difficultyIdx: null,
  focus: [],
});
const isStarting = ref(false);

const canStart = computed(() => {
  return Boolean(form.value.direction) && form.value.difficultyIdx !== null;
});

// === 启动面试 ===
const startInterview = async () => {
  if (!canStart.value || isStarting.value) return;
  isStarting.value = true;
  try {
    const direction = directions.value.find((item) => item.key === form.value.direction);
    const difficulty = difficulties.value[form.value.difficultyIdx];
    const response = await apiService.user.createSession(
      buildSessionCreatePayload({
        title: `${direction?.label || "技术"}面试`,
        directionKey: form.value.direction,
        difficulty: difficulty?.key || "mid",
        focusKeys: form.value.focus,
        estimatedMinutes: selectedEstimatedMinutes.value,
        resumeArtifactId: form.value.resumeArtifactId,
      })
    );
    const sessionId = response?.session?.sessionId;
    if (!sessionId) {
      throw new Error("后端未返回面试会话 ID");
    }
    router.push({
      path: "/chat",
      query: { sessionId },
    });
  } catch (error) {
    ElMessage.error(error?.message || "创建面试失败，请稍后重试");
  } finally {
    isStarting.value = false;
  }
};

// === 绝对时间戳 → 相对时间 ===
const formatRelativeTime = (timestamp) => {
  if (!timestamp) return "近期";
  const ts = typeof timestamp === "number" ? timestamp : new Date(timestamp).getTime();
  if (Number.isNaN(ts)) return "近期";
  const diff = Date.now() - ts;
  const min = 60 * 1000;
  const hour = 60 * min;
  const day = 24 * hour;
  if (diff < hour) return `${Math.max(1, Math.floor(diff / min))} 分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`;
  if (diff < 2 * day) return "昨天";
  if (diff < 7 * day) return `${Math.floor(diff / day)} 天前`;
  if (diff < 30 * day) return `${Math.floor(diff / (7 * day))} 周前`;
  return new Date(ts).toLocaleDateString("zh-CN");
};

// === 异步加载预设（方向 / 难度 / focus）===
const loadPresets = async () => {
  try {
    const res = await apiService.user.interviewPresets();
    if (!res) return;

    if (Array.isArray(res.directions) && res.directions.length > 0) {
      directions.value = res.directions.map((d) => ({
        key: d.key,
        label: d.label,
        color: pickDirectionColor(d.key),
        focusKeys: Array.isArray(d.focusKeys) ? [...d.focusKeys] : [],
        questionCount: Number(d.questionCount) || 0,
        difficultyRange: Array.isArray(d.difficultyRange) ? [...d.difficultyRange] : [],
        // 后端 focusKeys 是该方向的重点 chip。用作 tags 显示；限高 3 项避免溢出。
        tags: Array.isArray(d.focusKeys) ? d.focusKeys.slice(0, 3) : [],
      }));
    }

    if (Array.isArray(res.difficulties) && res.difficulties.length > 0) {
      difficulties.value = res.difficulties.map((dif) => {
        const key = DIFFICULTY_KEY_BY_LEVEL[dif.level];
        if (!key) {
          // 后端新增了等级但前端常量未更新；显式 warn 而非悄悄降级到 mid。
          // 兜底用 "mid" 保证 UI 不崩，同时通过 console 让开发者立刻发现契约漂移。
          console.warn(
            `[WorkbenchNew] 未知难度 level=${dif.level}，请同步更新 DIFFICULTY_KEY_BY_LEVEL` +
              ` 与 utils/interviewSession.js 的 DIFFICULTY_LEVEL_ALIASES`,
          );
        }
        return {
          key: key || "mid",
          level: dif.level,
          label: dif.label,
          desc: dif.description,
        };
      });
    }

    if (Array.isArray(res.focusOptions) && res.focusOptions.length > 0) {
      focusOptions.value = res.focusOptions.map((f) => ({
        key: f.key,
        label: f.label,
        desc: f.description,
      }));
    }

    // 如果后端提供了默认配置（且用户还未选中任何内容），预填表单。
    const dc = res.defaultConfig;
    if (dc) {
      if (!form.value.direction && dc.directionKey) {
        selectDirection(
          directions.value.find((d) => d.key === dc.directionKey) || {
            key: dc.directionKey,
            focusKeys: [],
          },
        );
      }
      if (form.value.difficultyIdx === null && typeof dc.difficultyLevel === "number") {
        const idx = difficulties.value.findIndex(
          (d) => d.key === DIFFICULTY_KEY_BY_LEVEL[dc.difficultyLevel],
        );
        if (idx >= 0) form.value.difficultyIdx = idx;
      }
      if (form.value.focus.length === 0 && Array.isArray(dc.focusAreas)) {
        form.value.focus = dc.focusAreas.map((f) => f.key).filter(Boolean);
      }
      if (typeof dc.estimatedMinutes === "number" && dc.estimatedMinutes > 0) {
        selectedEstimatedMinutes.value = dc.estimatedMinutes;
      }
    }
  } catch (error) {
    // 静默降级；本地 fallback 预设继续可用
  }
};

// === 异步加载简历资料列表 ===
// 后端返回 0 条 = 新用户从未上传，直接保持 resumes.value=[]，step 2 只显示「添加新简历」按钮兜底。
// 不允许 fallback 到本地示例，那会让新用户误选不存在的 artifactId，提交时 createSession 会失败。
const loadResumes = async () => {
  try {
    const res = await apiService.user.resumeArtifacts();
    const list = Array.isArray(res?.artifacts) ? res.artifacts : [];
    resumes.value = list.map((it, i) => ({
      id: it.artifactId,
      name: it.title || it.filename || `简历 v${it.version}`,
      info: `${it.chunkCount || 0} 片段 · ${formatRelativeTime(it.updatedAt || it.uploadedAt)}`,
      primary: i === 0,
    }));
  } catch (error) {
    // 静默降级为空列表：关联简历区域只保留“添加新简历”入口。
    resumes.value = [];
  }
};

onMounted(async () => {
  // 并发拉两个接口；简历必须来自后端，空响应保持空态。
  await Promise.all([loadPresets(), loadResumes()]);
});
</script>

<style scoped>
/* ============ Layout ============ */
.wb-new-content {
  width: 100%;
  max-width: min(1440px, 100%);
  margin: 0 auto;
  padding: 0 44px 120px;
  /* 底部 120px 留白：summary bar 自然布局到页面底部后的呼吸空间（原 sticky 避让
     语义已废弃，方案 A：summary 不再 sticky，改为普通页面底部块）。 */
}

/* === Hero === */
.wb-new-hero {
  padding: 0 0 48px;
}

.wb-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: var(--fs-xs) var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.025);
  backdrop-filter: blur(8px);
  width: fit-content;
  flex-shrink: 0;
}

.wb-eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(220, 155, 90, 0.9);
  animation: wb-edot 2.6s ease-in-out infinite;
}

@keyframes wb-edot {
  0%, 100% { opacity: 1; }
  50% { opacity: .35; }
}

.wb-new-sub {
  font-size: var(--fs-lg);
  color: var(--t3);
  line-height: 1.7;
  margin: 0;
  max-width: 560px;
}

/* ============ Stepper ============ */
.wb-stepper {
  list-style: none;
  display: flex;
  gap: 12px;
  margin: 0 0 40px;
  padding: 0;
  flex-wrap: wrap;
}

.wb-step {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 18px 10px 14px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  font: var(--fs-sm) var(--sans);
  color: var(--t3);
  transition: color .25s ease, background-color .25s ease, border-color .25s ease;
}

.wb-step-num {
  font: 600 var(--fs-xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  padding: 3px 8px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-step-active {
  color: var(--t);
  background: rgba(220, 155, 90, 0.06);
  border-color: rgba(220, 155, 90, 0.3);
}

.wb-step-active .wb-step-num {
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.12);
  border-color: rgba(220, 155, 90, 0.35);
}

.wb-step-done {
  color: var(--t2);
  border-color: rgba(255, 255, 255, 0.1);
}

.wb-step-done .wb-step-num {
  color: var(--t);
}

/* ============ Form Block ============ */
.wb-new-form {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.wb-block {
  border: none;
  padding: 0;
  margin: 0;
}

.wb-block-legend {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font: 600 var(--fs-md) var(--display);
  color: var(--t);
  margin-bottom: 16px;
  letter-spacing: 0;
}

.wb-block-tag {
  font: 600 var(--fs-2xs) var(--mono);
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border: 1px solid rgba(220, 155, 90, 0.25);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  letter-spacing: .04em;
}

.wb-block-hint {
  font: var(--fs-xs) var(--mono);
  color: var(--t3);
  font-weight: 400;
  margin-left: 4px;
}

/* === 方向 grid === */
.wb-direction-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.wb-dir {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 18px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 1) 0%, rgba(11, 12, 16, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  text-align: left;
  transition: transform .25s ease, border-color .25s ease, box-shadow .25s ease;
  font-family: inherit;
  color: inherit;
}

.wb-dir:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.15);
}

.wb-dir-active {
  background:
    linear-gradient(180deg, rgba(28, 22, 18, 1) 0%, rgba(18, 14, 11, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 155, 90, 0.45) 0%, rgba(220, 155, 90, 0.15) 100%) border-box;
  box-shadow: 0 0 0 1px rgba(220, 155, 90, 0.4) inset, 0 8px 24px rgba(0, 0, 0, 0.3);
}

.wb-dir-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 8px currentColor;
  /* 阴影色继承 background；通过 currentColor 触发 halo 效果，需 element 自带 color */
}

.wb-dir-active .wb-dir-dot {
  box-shadow: 0 0 0 3px rgba(220, 155, 90, 0.18);
}

.wb-dir-name {
  font: 700 var(--fs-md) var(--display);
  color: var(--t);
  letter-spacing: 0;
}

.wb-dir-tags {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-dir-active .wb-dir-tags {
  color: rgba(220, 155, 90, 0.85);
}

/* === 简历 row === */
.wb-resume-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.wb-resume {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  color: inherit;
  transition: border-color .2s ease, background-color .2s ease;
  position: relative;
}

.wb-resume:hover {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.04);
}

.wb-resume-active {
  border-color: rgba(220, 155, 90, 0.4);
  background: rgba(220, 155, 90, 0.05);
}

.wb-resume-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: var(--t2);
  flex-shrink: 0;
}

.wb-resume-icon svg {
  width: 18px;
  height: 18px;
  display: block;
}

.wb-resume-icon-add {
  font-size: 22px;
  font-weight: 600;
  color: var(--t3);
}

.wb-resume-meta {
  flex: 1;
  min-width: 0;
}

.wb-resume-name {
  font: 600 var(--fs-sm) var(--sans);
  color: var(--t);
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-resume-info {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-resume-tag {
  font: var(--fs-3xs) var(--mono);
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.1);
  border: 1px solid rgba(220, 155, 90, 0.3);
  border-radius: var(--radius-pill);
  padding: 2px 8px;
  letter-spacing: .04em;
  flex-shrink: 0;
}

.wb-resume-add {
  border-style: dashed;
}

/* === 难度按钮组 === */
.wb-difficulty {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.wb-diff-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 14px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-family: inherit;
  color: inherit;
  transition: border-color .2s ease, background-color .2s ease;
}

.wb-diff-btn:hover {
  border-color: rgba(255, 255, 255, 0.14);
}

.wb-diff-btn-active {
  border-color: rgba(220, 155, 90, 0.4);
  background: rgba(220, 155, 90, 0.05);
}

.wb-diff-name {
  font: 700 var(--fs-md) var(--display);
  color: var(--t);
}

.wb-diff-btn-active .wb-diff-name {
  color: rgba(220, 155, 90, 0.95);
}

.wb-diff-desc {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  text-align: center;
}

/* === Focus checkboxes === */
.wb-focus-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.wb-focus {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color .2s ease, background-color .2s ease;
}

.wb-focus:hover {
  border-color: rgba(255, 255, 255, 0.14);
}

.wb-focus-active {
  border-color: rgba(220, 155, 90, 0.4);
  background: rgba(220, 155, 90, 0.05);
}

.wb-focus-input {
  /* 视觉隐藏，仍参与表单语义 */
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.wb-focus-check {
  width: 20px;
  height: 20px;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: transparent;
  transition: background-color .2s ease, border-color .2s ease, color .2s ease;
}

.wb-focus-check svg {
  width: 14px;
  height: 14px;
}

.wb-focus-active .wb-focus-check {
  background: rgba(220, 155, 90, 0.95);
  border-color: rgba(220, 155, 90, 0.95);
  color: var(--bg);
}

.wb-focus-label {
  font: var(--fs-md) var(--sans);
  color: var(--t2);
}

.wb-focus-active .wb-focus-label {
  color: var(--t);
}

/* ============ 底部 summary bar ============ */
/* 设计决策：原 position:sticky+bottom:24 在 1440x900 桌面视口下会浮在视口底部覆盖
   「关联简历」+「难度等级」区域（plan §4 Task 3 / D-1 决策方案 A）。
   改为普通页面底部块：summary 自然出现在所有 fieldset 之后，用户必须滚动到页面底部
   才能看到，不再阻挡首屏阅读路径。响应式 @media 查询保留 flex-direction:column 不变。 */
.wb-summary-bar {
  margin-top: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 18px 24px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.92) 0%, rgba(11, 12, 16, 0.92) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.14) 0%, rgba(255, 255, 255, 0.04) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  backdrop-filter: blur(20px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 16px 40px rgba(0, 0, 0, 0.4);
  z-index: 5;
}

.wb-summary-text {
  display: flex;
  align-items: center;
  gap: 24px;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.wb-summary-line {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.wb-summary-label {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-summary-value {
  font: 600 var(--fs-md) var(--sans);
  color: var(--t);
}

.wb-summary-sep {
  width: 1px;
  height: 32px;
  background: rgba(255, 255, 255, 0.08);
}

.wb-summary-cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 600 var(--fs-lg) var(--sans);
  color: var(--bg);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 13px 28px;
  border-radius: var(--radius-md);
  transition: opacity .2s ease, transform .2s ease, box-shadow .2s ease;
  box-shadow: 0 4px 16px rgba(255, 255, 255, 0.1);
}

.wb-summary-cta:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(255, 255, 255, 0.18);
}

.wb-summary-cta:disabled {
  opacity: .4;
  cursor: not-allowed;
}

.wb-summary-arrow {
  font-weight: 400;
  font-size: 18px;
  line-height: 1;
}

/* ============ 响应式 ============ */
@media (max-width: 1024px) {
  .wb-direction-grid,
  .wb-resume-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .wb-difficulty {
    grid-template-columns: repeat(2, 1fr);
  }
  .wb-focus-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .wb-new-content {
    padding: 0 20px 100px;
  }
  .wb-direction-grid,
  .wb-resume-row,
  .wb-focus-grid {
    grid-template-columns: 1fr;
  }
  .wb-summary-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  .wb-summary-cta {
    width: 100%;
    justify-content: center;
  }
}

/* ============ C1 双栏工作台重构 ============ */
.wb-new-content {
  width: 100%;
  max-width: 1680px;
  margin: 0 auto;
  padding: 0 clamp(20px, 4vw, 56px) 96px;
}

.wb-new-hero {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0 0 clamp(24px, 3vw, 40px);
}

.wb-new-sub {
  max-width: 720px;
}

.wb-new-shell {
  display: grid;
  grid-template-columns: minmax(260px, 34%) minmax(0, 1fr);
  gap: clamp(18px, 2vw, 32px);
  align-items: start;
}

.wb-config-panel {
  min-width: 0;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.86) 0%, rgba(8, 9, 13, 0.90) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.12) 0%, rgba(255, 255, 255, 0.035) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  backdrop-filter: blur(18px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.055),
    0 18px 46px rgba(0, 0, 0, 0.34);
  padding: clamp(18px, 2vw, 28px);
}

.wb-stepper {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin: 0 0 28px;
}

.wb-step {
  width: 100%;
  justify-content: flex-start;
  padding: 10px 12px;
  border-radius: var(--radius-md);
}

.wb-step-copy {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  min-width: 0;
}

.wb-step-label {
  color: inherit;
  white-space: nowrap;
}

.wb-step-desc {
  font: var(--fs-2xs) var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  white-space: nowrap;
}

.wb-new-form {
  gap: clamp(24px, 2.4vw, 34px);
}

.wb-direction-grid,
.wb-resume-row,
.wb-focus-grid {
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.wb-difficulty {
  grid-template-columns: repeat(auto-fit, minmax(118px, 1fr));
}

.wb-summary-bar {
  margin-top: clamp(20px, 2.4vw, 36px);
}

.wb-summary-value {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .wb-new-shell {
    grid-template-columns: 1fr;
  }

  .wb-stepper {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .wb-stepper {
    grid-template-columns: 1fr;
  }

  .wb-summary-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .wb-summary-sep {
    display: none;
  }
}

/* ============ C4 配置型新建面试页语义修正 ============ */
.wb-new-content {
  max-width: min(1440px, 100%);
  min-height: calc(100svh - 80px);
  display: flex;
  flex-direction: column;
  padding: 0 clamp(20px, 4vw, 56px) 96px;
}

.wb-new-hero {
  width: 100%;
  max-width: 1120px;
  margin: 0 auto;
  display: flex;
  flex-direction: row;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 18px;
  text-align: left;
  padding-bottom: clamp(20px, 2.6vw, 34px);
}

.wb-new-sub {
  flex: 1 1 560px;
  max-width: none;
  color: var(--t2);
  line-height: 1.65;
}

.wb-stepper {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: clamp(10px, 2vw, 24px);
  margin: 0 auto clamp(22px, 3vw, 38px);
  max-width: 720px;
}

.wb-step {
  position: relative;
  width: auto;
  min-width: 0;
  padding: 0;
  gap: 10px;
  border: 0;
  background: transparent;
  color: var(--t3);
}

.wb-step:not(:last-child)::after {
  content: "";
  width: clamp(28px, 4vw, 58px);
  height: 1px;
  margin-left: clamp(4px, 1vw, 14px);
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.20), rgba(255, 255, 255, 0.06));
}

.wb-step-active:not(:last-child)::after,
.wb-step-done:not(:last-child)::after {
  background: linear-gradient(90deg, rgba(220, 155, 90, 0.85), rgba(255, 255, 255, 0.08));
}

.wb-step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  padding: 0;
  border-radius: 50%;
  background: rgba(8, 10, 14, 0.72);
  border-color: rgba(255, 255, 255, 0.18);
}

.wb-step-active .wb-step-num {
  color: rgba(255, 224, 172, 0.98);
  box-shadow: 0 0 0 5px rgba(220, 155, 90, 0.10), 0 0 18px rgba(220, 155, 90, 0.24);
}

.wb-step-label {
  font: 600 var(--fs-sm) var(--sans);
}

.wb-new-shell {
  display: block;
  width: 100%;
  max-width: 1120px;
  margin: 0 auto;
  min-height: 0;
}

.wb-config-panel {
  padding: 0;
  background: none;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  backdrop-filter: none;
}

.wb-new-form {
  gap: clamp(14px, 1.8vw, 20px);
}

.wb-block {
  display: grid;
  grid-template-columns: minmax(180px, 23%) minmax(0, 1fr);
  align-items: center;
  gap: clamp(18px, 3vw, 36px);
  padding: clamp(20px, 2.2vw, 28px);
  background:
    linear-gradient(180deg, rgba(18, 22, 29, 0.86) 0%, rgba(8, 11, 16, 0.92) 100%) padding-box,
    linear-gradient(145deg, rgba(255, 255, 255, 0.16) 0%, rgba(255, 255, 255, 0.035) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    0 14px 42px rgba(0, 0, 0, 0.25);
}

.wb-visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
  border: 0;
}

.wb-block-headline {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  min-width: 0;
}

.wb-block-headline > div {
  min-width: 0;
}

.wb-block-title {
  margin: 0 0 8px;
  color: var(--t);
  font: 800 clamp(20px, 1.6vw, 26px) var(--display);
  letter-spacing: 0;
}

.wb-block-desc {
  margin: 0;
  color: var(--t3);
  line-height: 1.65;
}

.wb-block-tag {
  flex-shrink: 0;
  min-width: 34px;
  text-align: center;
}

.wb-block-body {
  min-width: 0;
}

.wb-direction-grid {
  grid-template-columns: repeat(auto-fit, minmax(min(250px, 100%), 1fr));
}

.wb-dir {
  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  grid-template-areas:
    "dot name"
    "dot meta";
  column-gap: 14px;
  row-gap: 5px;
  align-items: center;
  min-height: 96px;
  padding: 18px;
}

.wb-dir-dot {
  grid-area: dot;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.16), 0 0 18px rgba(255, 255, 255, 0.10);
}

.wb-dir-name {
  grid-area: name;
  min-width: 0;
  padding-right: 32px;
}

.wb-dir-tags {
  grid-area: meta;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-choice-check {
  position: absolute;
  top: 18px;
  right: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(245, 180, 92, 0.96);
  color: #15110c;
  font: 800 var(--fs-xs) var(--sans);
}

.wb-resume-row {
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
}

.wb-resume {
  min-height: 66px;
}

.wb-difficulty {
  grid-template-columns: repeat(auto-fit, minmax(128px, 1fr));
}

.wb-diff-btn {
  min-height: 88px;
}

.wb-focus-grid {
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.wb-summary-bar {
  width: 100%;
  max-width: 1120px;
  margin: clamp(20px, 2.8vw, 36px) auto 0;
  border-radius: var(--radius-lg);
}

.wb-summary-cta {
  min-width: 180px;
  border-radius: var(--radius-pill);
  background:
    linear-gradient(180deg, #ffd98c 0%, #f1af4f 100%);
  color: #17110a;
  box-shadow: 0 12px 30px rgba(220, 155, 90, 0.20), inset 0 1px 0 rgba(255, 255, 255, 0.55);
}

@media (max-width: 900px) {
  .wb-new-hero {
    align-items: flex-start;
    text-align: left;
  }

  .wb-stepper {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .wb-step:not(:last-child)::after {
    width: clamp(18px, 6vw, 42px);
  }

  .wb-block {
    grid-template-columns: 1fr;
    align-items: stretch;
  }
}

@media (min-width: 900px) {
  .wb-new-content {
    height: calc(100svh - 80px);
    min-height: 640px;
    overflow: hidden;
    padding-bottom: clamp(28px, 4vw, 48px);
  }

  .wb-new-hero {
    flex: 0 0 auto;
    padding: clamp(20px, 3vw, 28px) 0 clamp(16px, 2vw, 24px);
  }

  .wb-stepper {
    flex: 0 0 auto;
    margin-bottom: clamp(16px, 2vw, 24px);
  }

  .wb-new-shell {
    flex: 1 1 auto;
    overflow: hidden;
  }

  .wb-config-panel {
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }

  .wb-new-form {
    height: 100%;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.16) transparent;
    scrollbar-gutter: stable;
    padding-right: 6px;
  }

  .wb-new-form::-webkit-scrollbar {
    width: 6px;
  }

  .wb-new-form::-webkit-scrollbar-track {
    background: transparent;
  }

  .wb-new-form::-webkit-scrollbar-thumb {
    border-radius: var(--radius-pill);
    background: rgba(255, 255, 255, 0.16);
  }

  .wb-new-form::-webkit-scrollbar-thumb:hover {
    background: rgba(220, 155, 90, 0.32);
  }

  .wb-summary-bar {
    flex: 0 0 auto;
    margin-top: clamp(16px, 2vw, 24px);
  }
}

@media (max-width: 620px) {
  .wb-stepper {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .wb-step::after {
    display: none;
  }

  .wb-step {
    padding: 9px 10px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.025);
  }

  .wb-dir {
    grid-template-columns: auto minmax(0, 1fr);
    grid-template-areas:
      "dot name"
      "dot meta";
  }
}
</style>
