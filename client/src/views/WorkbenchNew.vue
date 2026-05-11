<template>
  <!--
    WorkbenchNew：新建面试配置向导（对应设计图 2）。
    布局：stepper 进度条 + 方向 grid + 简历选择 + 难度滑杆 + 重点 checkbox + 底部 summary bar
    交互：单页配置（非分步路由），用户在一页完成所有选择，点 "开始面试" 跳转 /chat?direction=...&difficulty=...
    后端契约：当前接 /users/demo/interview-scenes/random 仅作为种子参考；
            真实开练通过 /chat 页面的 SSE chat 接口（携带 direction/difficulty/focus 参数）。
  -->
  <WorkbenchLayout>
    <div class="wb-new-content">
      <section class="wb-new-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>新建面试</span>
        </div>
        <h1 class="wb-new-title">配置一场专属面试</h1>
        <p class="wb-new-sub">选择方向、难度和重点方向，AI 会基于你的简历定制题目。</p>
      </section>

      <!-- Stepper -->
      <ol class="wb-stepper" aria-label="配置步骤">
        <li
          v-for="(step, idx) in steps"
          :key="step.key"
          class="wb-step"
          :class="{ 'wb-step-active': currentStep === idx, 'wb-step-done': currentStep > idx }"
        >
          <span class="wb-step-num">{{ String(idx + 1).padStart(2, '0') }}</span>
          <span class="wb-step-label">{{ step.label }}</span>
        </li>
      </ol>

      <div class="wb-new-form">
        <!-- 方向选择 grid -->
        <fieldset class="wb-block" data-step="direction">
          <legend class="wb-block-legend">
            <span class="wb-block-tag">01</span>
            选择方向
          </legend>
          <div class="wb-direction-grid">
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
              <span class="wb-dir-tags">{{ dir.tags.join(' · ') }}</span>
            </button>
          </div>
        </fieldset>

        <!-- 简历选择 row -->
        <fieldset class="wb-block">
          <legend class="wb-block-legend">
            <span class="wb-block-tag">02</span>
            关联简历
          </legend>
          <div class="wb-resume-row">
            <button
              v-for="resume in resumes"
              :key="resume.id"
              type="button"
              class="wb-resume"
              :class="{ 'wb-resume-active': form.resumeId === resume.id }"
              @click="form.resumeId = resume.id"
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
                <div class="wb-resume-info">PDF / DOCX，最大 10MB</div>
              </div>
            </button>
          </div>
        </fieldset>

        <!-- 难度滑杆 -->
        <fieldset class="wb-block">
          <legend class="wb-block-legend">
            <span class="wb-block-tag">03</span>
            难度等级
          </legend>
          <div class="wb-difficulty">
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

        <!-- 重点方向 -->
        <fieldset class="wb-block">
          <legend class="wb-block-legend">
            <span class="wb-block-tag">04</span>
            重点方向 <span class="wb-block-hint">（可多选）</span>
          </legend>
          <div class="wb-focus-grid">
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
            <span class="wb-summary-value">{{ form.focus.length > 0 ? `${form.focus.length} 项` : '未选择' }}</span>
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
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";
import { buildSessionCreatePayload } from "../utils/interviewSession";

const router = useRouter();

// === 步骤定义（视觉性 stepper，仅展示进度，不强制顺序） ===
const steps = [
  { key: "direction", label: "选择方向" },
  { key: "resume", label: "关联简历" },
  { key: "difficulty", label: "难度等级" },
  { key: "focus", label: "重点方向" },
];

// 当前所在步骤：根据已填字段推断（任何用户操作都会推进 step）
const currentStep = computed(() => {
  if (!form.value.direction) return 0;
  if (!form.value.resumeId) return 1;
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

// === 方向选项（mock first，onMounted 异步接入 interviewPresets 覆盖）===
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

// === 简历选项（mock first，onMounted 异步接入 resumeArtifacts 覆盖）===
const resumes = ref([
  { id: "r-v3", name: "Resume_v3.pdf", info: "12 项目 · 上次更新 2 天前", primary: true },
  { id: "r-v2", name: "Resume_v2.pdf", info: "10 项目 · 1 周前", primary: false },
]);

const goToResumePage = () => {
  router.push({ path: "/workbench/resume" });
};

// === 难度等级（mock first，onMounted 异步接入 interviewPresets 覆盖）===
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
  { key: "entry", label: "入门", desc: "确认基本概念和术语理解" },
  { key: "junior", label: "初级", desc: "常见场景与基础工程经验" },
  { key: "mid", label: "中级", desc: "机制、取舍与故障复盘" },
  { key: "senior", label: "资深", desc: "架构判断与工程边界" },
  { key: "expert", label: "专家", desc: "高压技术面 · 系统化论证" },
]);

const selectedDifficultyLabel = computed(() => {
  if (form.value.difficultyIdx === null) return "未选择";
  return difficulties.value[form.value.difficultyIdx]?.label || "";
});

// === 重点方向（mock first，onMounted 异步接入 interviewPresets.focusOptions 覆盖）===
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
  resumeId: "",
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
    }
  } catch (error) {
    // 静默降级；mock 预设已然可用
  }
};

// === 异步加载简历资料列表 ===
const loadResumes = async () => {
  try {
    const res = await apiService.user.resumeArtifacts();
    const list = Array.isArray(res?.artifacts) ? res.artifacts : [];
    if (list.length === 0) return; // 保留 mock
    resumes.value = list.map((it, i) => ({
      id: it.artifactId,
      name: it.title || it.filename || `简历 v${it.version}`,
      info: `${it.chunkCount || 0} 片段 · ${formatRelativeTime(it.updatedAt || it.uploadedAt)}`,
      primary: i === 0,
    }));
  } catch (error) {
    // 静默降级
  }
};

onMounted(() => {
  // 并发拉两个接口；mock 已组装，成功后覆盖。
  loadPresets();
  loadResumes();
});
</script>

<style scoped>
/* ============ Layout ============ */
.wb-new-content {
  max-width: 1320px;
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
  font: 12px var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  margin-bottom: 22px;
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.025);
  backdrop-filter: blur(8px);
  width: fit-content;
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

.wb-new-title {
  font: 800 clamp(30px, 2.8vw, 42px) var(--display);
  color: var(--t);
  letter-spacing: -.02em;
  margin: 0 0 14px;
}

.wb-new-sub {
  font-size: 15px;
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
  font: 13px var(--sans);
  color: var(--t3);
  transition: color .25s ease, background-color .25s ease, border-color .25s ease;
}

.wb-step-num {
  font: 600 12px var(--mono);
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
  font: 600 16px var(--display);
  color: var(--t);
  margin-bottom: 16px;
  letter-spacing: -.01em;
}

.wb-block-tag {
  font: 600 11px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.08);
  border: 1px solid rgba(220, 155, 90, 0.25);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  letter-spacing: .04em;
}

.wb-block-hint {
  font: 12px var(--mono);
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
  font: 700 16px var(--display);
  color: var(--t);
  letter-spacing: -.01em;
}

.wb-dir-tags {
  font: 11px var(--mono);
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
  font: 600 13px var(--sans);
  color: var(--t);
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-resume-info {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-resume-tag {
  font: 10px var(--mono);
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
  font: 700 14px var(--display);
  color: var(--t);
}

.wb-diff-btn-active .wb-diff-name {
  color: rgba(220, 155, 90, 0.95);
}

.wb-diff-desc {
  font: 11px var(--mono);
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
  font: 14px var(--sans);
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
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
}

.wb-summary-value {
  font: 600 14px var(--sans);
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
  font: 600 15px var(--sans);
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
</style>
