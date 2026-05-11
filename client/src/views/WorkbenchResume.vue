<template>
  <!--
    WorkbenchResume：简历管理。
    布局：顶部 hero + 上传 dropzone + 已上传简历卡片列表 + 选中简历的解析详情面板。
    后端契约：上传走 /users/resume/upload (multipart/form-data)；
    上传成功后自动触发 /users/resume/artifacts/:id/analysis/prepare 生成持久化评估。
  -->
  <WorkbenchLayout>
    <div class="wb-resume-content">
      <section class="wb-resume-hero">
        <div class="wb-eyebrow">
          <span class="wb-eyebrow-dot" aria-hidden="true"></span>
          <span>简历管理</span>
        </div>
        <h1 class="wb-resume-title">让 AI 读懂你的项目</h1>
        <p class="wb-resume-sub">上传简历后，AI 会基于项目细节做深度追问，模拟真实面试节奏。</p>
      </section>

      <!-- 上传区 -->
      <section class="wb-upload">
        <div
          class="wb-dropzone"
          :class="{ 'wb-dropzone-dragging': isDragging, 'wb-dropzone-uploading': uploading }"
          @dragenter.prevent="handleDragEnter"
          @dragover.prevent
          @dragleave.prevent="handleDragLeave"
          @drop.prevent="handleDrop"
          @click="triggerFileInput"
        >
          <input
            ref="fileInputRef"
            type="file"
            accept=".pdf,application/pdf"
            class="wb-file-input"
            @change="handleFileChange"
          />
          <div class="wb-dropzone-inner">
            <div class="wb-dropzone-icon" aria-hidden="true">
              <svg viewBox="0 0 64 64" fill="none">
                <rect x="14" y="10" width="36" height="44" rx="3" stroke="currentColor" stroke-width="1.5" />
                <line x1="22" y1="22" x2="42" y2="22" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                <line x1="22" y1="30" x2="42" y2="30" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                <line x1="22" y1="38" x2="34" y2="38" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                <circle cx="48" cy="48" r="10" fill="rgba(220,155,90,0.12)" stroke="rgba(220,155,90,0.85)" stroke-width="1.5" />
                <line x1="48" y1="44" x2="48" y2="52" stroke="rgba(220,155,90,0.95)" stroke-width="1.5" stroke-linecap="round" />
                <line x1="44" y1="48" x2="52" y2="48" stroke="rgba(220,155,90,0.95)" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </div>
            <div class="wb-dropzone-text">
              <div class="wb-dropzone-main">{{ uploading ? '正在上传...' : '拖拽简历到此处' }}</div>
              <div class="wb-dropzone-sub">
                <span v-if="!uploading">或</span>
                <span class="wb-dropzone-action" v-if="!uploading">点击选择文件</span>
                <span class="wb-dropzone-meta">支持 PDF，最大 10 MB</span>
              </div>
            </div>
          </div>
          <div v-if="uploading" class="wb-upload-progress">
            <div class="wb-upload-bar" :style="{ width: uploadProgress + '%' }"></div>
          </div>
        </div>
        <p v-if="uploadError" class="wb-upload-error" role="alert">{{ uploadError }}</p>
      </section>

      <!-- 已上传简历列表 -->
      <section class="wb-resumes">
        <header class="wb-block-head">
          <h3 class="wb-block-title">我的简历</h3>
          <span class="wb-block-meta">{{ resumes.length }} 份</span>
        </header>

        <div v-if="resumes.length > 0" class="wb-resumes-grid">
          <article
            v-for="resume in resumes"
            :key="resume.id"
            class="wb-resume-card"
            :class="{ 'wb-resume-card-active': selectedId === resume.id }"
            @click="selectedId = resume.id"
          >
            <div class="wb-resume-card-head">
              <div class="wb-resume-card-icon" aria-hidden="true">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                  <path d="M3.5 1.5h6.5L13 4.5v10H3.5z" />
                  <path d="M10 1.5V4.5h3" />
                  <line x1="5.5" y1="8" x2="10.5" y2="8" stroke-linecap="round" />
                  <line x1="5.5" y1="10.5" x2="10.5" y2="10.5" stroke-linecap="round" />
                </svg>
              </div>
              <span v-if="resume.primary" class="wb-resume-card-tag">主用</span>
            </div>
            <h4 class="wb-resume-card-name">{{ resume.name }}</h4>
            <div class="wb-resume-card-meta">
              <span>{{ resume.size }}</span>
              <span aria-hidden="true">·</span>
              <span>{{ resume.uploadedAt }}</span>
            </div>
            <div class="wb-resume-card-stats">
              <div class="wb-resume-stat">
                <span class="wb-resume-stat-num">{{ resume.projectCount }}</span>
                <span class="wb-resume-stat-lb">项目</span>
              </div>
              <div class="wb-resume-stat">
                <span class="wb-resume-stat-num">{{ resume.skillCount }}</span>
                <span class="wb-resume-stat-lb">技能</span>
              </div>
            </div>
            <div class="wb-resume-card-foot">
              <span class="wb-resume-card-status">
                <span class="wb-resume-card-dot" :class="`wb-status-${resume.status}`"></span>
                {{ getStatusLabel(resume.status) }}
              </span>
              <span class="wb-resume-card-action">查看 →</span>
            </div>
          </article>
        </div>

        <div v-else class="wb-empty">
          <div class="wb-empty-title">还没有简历</div>
          <div class="wb-empty-sub">上传简历后，AI 会基于项目内容深度追问。</div>
        </div>
      </section>

      <!-- 选中简历详情 -->
      <section v-if="selectedResume" class="wb-resume-detail">
        <header class="wb-block-head">
          <h3 class="wb-block-title">解析详情 · {{ selectedResume.name }}</h3>
          <div class="wb-detail-actions">
            <button
              type="button"
              class="wb-refresh-btn"
              :disabled="selectedResume.evaluationLoading"
              @click="prepareResumeEvaluation(selectedResume.id, { force: true })"
            >
              {{ selectedResume.evaluationLoading ? '评估中' : '刷新评估' }}
            </button>
            <button type="button" class="wb-block-close" @click="selectedId = ''" aria-label="关闭详情">×</button>
          </div>
        </header>

        <div class="wb-eval-overview">
          <div class="wb-eval-score">
            <span class="wb-eval-score-num">{{ formatScore(selectedResume.overallScore) }}</span>
            <span class="wb-eval-score-label">面试准备度</span>
          </div>
          <div class="wb-eval-summary">
            <div class="wb-eval-status" :class="`wb-eval-${selectedResume.evaluationStatus}`">
              {{ getEvaluationStatusLabel(selectedResume.evaluationStatus) }}
            </div>
            <p>{{ selectedResume.summary || '已完成基础解析，等待生成简历评估。' }}</p>
          </div>
        </div>

        <div v-if="selectedResume.dimensions.length" class="wb-dimensions">
          <div v-for="dimension in selectedResume.dimensions" :key="dimension.key" class="wb-dimension">
            <div class="wb-dimension-head">
              <span>{{ dimension.label }}</span>
              <strong>{{ dimension.score }}</strong>
            </div>
            <div class="wb-dimension-bar" aria-hidden="true">
              <span :style="{ width: `${Math.min(100, Math.max(0, dimension.score || 0))}%` }"></span>
            </div>
            <p>{{ dimension.summary }}</p>
          </div>
        </div>

        <div class="wb-detail-grid">
          <div class="wb-detail-col">
            <div class="wb-detail-label">关键技能</div>
            <div class="wb-tags">
              <span v-for="tag in selectedResume.skills" :key="tag" class="wb-tag">{{ tag }}</span>
            </div>
          </div>
          <div class="wb-detail-col">
            <div class="wb-detail-label">项目（{{ selectedResume.projects.length }}）</div>
            <ol class="wb-projects">
              <li v-for="proj in selectedResume.projects" :key="proj.name" class="wb-project">
                <div class="wb-project-name">{{ proj.name }}</div>
                <div class="wb-project-stack">{{ proj.stack }}</div>
              </li>
            </ol>
          </div>
        </div>

        <div class="wb-eval-lists">
          <div class="wb-eval-list">
            <div class="wb-detail-label">优势</div>
            <ul>
              <li v-for="item in selectedResume.strengths" :key="item">{{ item }}</li>
            </ul>
          </div>
          <div class="wb-eval-list">
            <div class="wb-detail-label">建议</div>
            <ul>
              <li v-for="item in selectedResume.suggestions" :key="item">{{ item }}</li>
            </ul>
          </div>
          <div class="wb-eval-list">
            <div class="wb-detail-label">风险</div>
            <ul>
              <li v-for="risk in selectedResume.risks" :key="risk.key || risk.label">
                {{ risk.label }}：{{ risk.suggestion }}
              </li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

// === 上传 ===
const fileInputRef = ref(null);
const isDragging = ref(false);
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadError = ref("");

const handleDragEnter = () => {
  isDragging.value = true;
};

const handleDragLeave = (e) => {
  // 只在真正离开 dropzone 时才取消高亮（不算冒泡到子元素）
  if (e.currentTarget === e.target) {
    isDragging.value = false;
  }
};

const handleDrop = (e) => {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) {
    uploadFile(files[0]);
  }
};

const triggerFileInput = () => {
  if (uploading.value) return;
  fileInputRef.value?.click();
};

const handleFileChange = (e) => {
  const file = e.target.files?.[0];
  if (file) {
    uploadFile(file);
  }
};

const validateFile = (file) => {
  const maxBytes = 10 * 1024 * 1024;
  if (file.size > maxBytes) {
    return "文件过大，最大支持 10 MB";
  }
  if (!/\.pdf$/i.test(file.name || "")) {
    return "仅支持 PDF 格式";
  }
  return "";
};

const createResumeChatId = () => {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return `resume-${crypto.randomUUID()}`;
  }
  return `resume-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
};

const uploadFile = async (file) => {
  uploadError.value = "";
  const validationError = validateFile(file);
  if (validationError) {
    uploadError.value = validationError;
    return;
  }

  uploading.value = true;
  uploadProgress.value = 10;

  // 视觉伪进度：真实进度需要 axios upload progress event，
  // 当前 apiService.user.resumeUpload 直接返回 Promise，无 onUploadProgress 接入。
  // 后续可改造 endpoint 工厂支持 progress callback。
  const fakeTimer = setInterval(() => {
    if (uploadProgress.value < 85) {
      uploadProgress.value += 5;
    }
  }, 200);

  try {
    const formData = new FormData();
    formData.append("file", file);
    // ResumeUploadReq 要求 chatId 必填；每次上传使用唯一绑定 ID，避免不同用户共享 workbench-default。
    formData.append("chatId", createResumeChatId());
    formData.append("title", file.name);
    formData.append("mode", "Memory");
    const res = await apiService.user.resumeUpload(formData);
    uploadProgress.value = 100;

    // 成功后立即拉列表，让后端返回的 artifactId / status 为准。
    await loadResumes();
    const targetId = res?.artifactId || resumes.value[0]?.id || "";
    if (targetId) {
      selectedId.value = targetId;
      void prepareResumeEvaluation(targetId, { force: true });
    }
  } catch (error) {
    uploadError.value = error?.message || "上传失败，请稍后再试";
  } finally {
    clearInterval(fakeTimer);
    uploading.value = false;
    uploadProgress.value = 0;
    if (fileInputRef.value) {
      fileInputRef.value.value = "";
    }
  }
};

const formatBytes = (bytes) => {
  if (!bytes || bytes < 0) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

const createEmptyEvaluationState = () => ({
  evaluationStatus: "missing",
  evaluationLoading: false,
  evaluationLoaded: false,
  overallScore: null,
  level: "",
  summary: "",
  dimensions: [],
  strengths: [],
  risks: [],
  suggestions: [],
  evidence: [],
});

// === 简历列表（mock first，onMounted 异步接入 resumeArtifacts 覆盖） ===
const resumes = ref([
  {
    id: "r-v3",
    name: "Resume_v3.pdf",
    size: "1.2 MB",
    uploadedAt: "2 天前",
    projectCount: 12,
    skillCount: 28,
    status: "parsed",
    primary: true,
    skills: ["Go", "Vue", "Postgres", "Redis", "Docker", "ETCD", "gRPC", "RAG"],
    projects: [
      { name: "GoZero-AI 个人面试官", stack: "Go-Zero · Vue 3 · pgvector" },
      { name: "微服务订单系统", stack: "Go · Kafka · MySQL" },
      { name: "实时协作白板", stack: "WebSocket · CRDT · Redis" },
    ],
    ...createEmptyEvaluationState(),
    evaluationStatus: "ready",
    overallScore: 82,
    summary: "项目素材完整，适合围绕微服务、RAG 和工程实践展开追问。",
    strengths: ["技术栈清晰", "项目素材丰富"],
    suggestions: ["补充核心项目的量化指标。"],
    risks: [],
  },
  {
    id: "r-v2",
    name: "Resume_v2.pdf",
    size: "1.0 MB",
    uploadedAt: "1 周前",
    projectCount: 10,
    skillCount: 22,
    status: "parsed",
    primary: false,
    skills: ["Go", "Vue", "Postgres", "Docker"],
    projects: [
      { name: "面试系统 v1", stack: "Go · MySQL · Vue" },
      { name: "blog 后端", stack: "Go · MongoDB" },
    ],
    ...createEmptyEvaluationState(),
    evaluationStatus: "ready",
    overallScore: 74,
    summary: "已有基础项目线索，仍需补充职责边界和结果证据。",
    strengths: ["方向明确"],
    suggestions: ["补充项目职责和优化结果。"],
    risks: [],
  },
]);

const selectedId = ref("");

const selectedResume = computed(() => {
  return resumes.value.find((r) => r.id === selectedId.value) || null;
});

// 后端 status (string) → 本地状态表（UI：parsed/parsing/failed）
const mapArtifactStatus = (raw) => {
  if (!raw) return "parsing";
  const s = String(raw).toLowerCase();
  if (s.includes("ready") || s.includes("parsed") || s.includes("success")) return "parsed";
  if (s.includes("fail") || s.includes("error")) return "failed";
  return "parsing";
};

// 绝对时间戳 → 相对时间
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

// 列表拉取：后端只返回基础信息（不含 skills/projects），详情在选中时 lazy load。
const loadResumes = async () => {
  try {
    const res = await apiService.user.resumeArtifacts();
    const list = Array.isArray(res?.artifacts) ? res.artifacts : [];
    if (list.length === 0) return; // 保留 mock

    resumes.value = list.map((it, i) => ({
      id: it.artifactId,
      name: it.title || it.filename || `简历 v${it.version}`,
      size: it.chunkCount > 0 ? `${it.chunkCount} 片段` : "—",
      uploadedAt: formatRelativeTime(it.updatedAt || it.uploadedAt),
      projectCount: 0,
      skillCount: 0,
      status: mapArtifactStatus(it.status),
      primary: i === 0,
      skills: [],
      projects: [],
      ...createEmptyEvaluationState(),
    }));
  } catch (error) {
    // 静默降级；mock 列表已可用
  }
};

const updateResume = (id, patch) => {
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0) return;
  resumes.value[idx] = {
    ...resumes.value[idx],
    ...patch,
  };
};

const applyResumeAnalysis = (id, res) => {
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0 || !res) return;
  const target = resumes.value[idx];

  const skills = Array.isArray(res.skills) ? res.skills.map((s) => s.label).filter(Boolean) : [];
  const projects = Array.isArray(res.projects)
    ? res.projects.map((p) => ({
      name: p.title || "未命名项目",
      stack: p.summary || (Array.isArray(p.evidence) ? p.evidence.slice(0, 2).join(" · ") : ""),
    }))
    : [];

  resumes.value[idx] = {
    ...target,
    skills,
    projects,
    skillCount: skills.length,
    projectCount: projects.length,
    evaluationStatus: res.evaluationStatus || target.evaluationStatus || "missing",
    overallScore: typeof res.overallScore === "number" ? res.overallScore : target.overallScore,
    level: res.level || target.level || "",
    summary: res.summary || target.summary || "",
    dimensions: Array.isArray(res.dimensions) ? res.dimensions : [],
    strengths: Array.isArray(res.strengths) ? res.strengths : [],
    risks: Array.isArray(res.risks) ? res.risks : [],
    suggestions: Array.isArray(res.suggestions) ? res.suggestions : [],
    evidence: Array.isArray(res.evidence) ? res.evidence : [],
    evaluationLoaded: true,
  };
};

// 选中某份简历后 lazy 拉 analysis 覆盖 skills/projects。
const loadResumeAnalysis = async (id) => {
  if (!id) return;
  const idx = resumes.value.findIndex((r) => r.id === id);
  if (idx < 0) return;
  const target = resumes.value[idx];
  if (target.evaluationLoaded) return; // 已拉过不重复

  try {
    const res = await apiService.user.resumeArtifactAnalysis(id, { limit: 6 });
    applyResumeAnalysis(id, res);
  } catch (error) {
    // 静默降级；mock 字段保留
  }
};

const prepareResumeEvaluation = async (id, options = {}) => {
  if (!id) return;
  updateResume(id, {
    evaluationLoading: true,
    evaluationStatus: "evaluating",
  });
  try {
    const res = await apiService.user.resumeArtifactAnalysisPrepare(id, {
      force: Boolean(options.force),
      limit: 6,
    });
    applyResumeAnalysis(id, res);
  } catch (error) {
    updateResume(id, {
      evaluationStatus: "failed",
      suggestions: [error?.message || "评估失败，请稍后重试"],
    });
  } finally {
    updateResume(id, {
      evaluationLoading: false,
      evaluationLoaded: true,
    });
  }
};

// 选中变化时拉取详情
watch(selectedId, (id) => {
  if (id) loadResumeAnalysis(id);
});

onMounted(() => {
  loadResumes();
});

const getStatusLabel = (status) => {
  switch (status) {
    case "parsed":
      return "已解析";
    case "parsing":
      return "解析中";
    case "failed":
      return "解析失败";
    default:
      return "待处理";
  }
};

const getEvaluationStatusLabel = (status) => {
  switch (status) {
    case "ready":
      return "评估完成";
    case "stale":
      return "需刷新";
    case "evaluating":
      return "评估中";
    case "insufficient_data":
      return "资料不足";
    case "failed":
      return "评估失败";
    case "missing":
      return "待评估";
    default:
      return "待评估";
  }
};

const formatScore = (score) => {
  if (typeof score !== "number" || Number.isNaN(score)) return "—";
  return Math.round(score);
};
</script>

<style scoped>
/* ============ Layout ============ */
.wb-resume-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 44px 80px;
}

.wb-resume-hero {
  padding: 0 0 40px;
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

.wb-resume-title {
  font: 800 clamp(30px, 2.8vw, 42px) var(--display);
  color: var(--t);
  letter-spacing: -.02em;
  margin: 0 0 14px;
}

.wb-resume-sub {
  font-size: 15px;
  color: var(--t3);
  line-height: 1.7;
  margin: 0;
  max-width: 560px;
}

/* ============ Dropzone ============ */
.wb-upload {
  margin-bottom: 48px;
}

.wb-dropzone {
  position: relative;
  display: block;
  padding: 56px 32px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.7) 0%, rgba(11, 12, 16, 0.7) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1.5px dashed rgba(255, 255, 255, 0.14);
  border-radius: var(--radius-lg);
  cursor: pointer;
  text-align: center;
  transition: border-color .25s ease, background-color .25s ease;
  overflow: hidden;
}

.wb-dropzone:hover {
  border-color: rgba(220, 155, 90, 0.5);
}

.wb-dropzone-dragging {
  border-color: rgba(220, 155, 90, 0.85);
  background: rgba(220, 155, 90, 0.05);
}

.wb-dropzone-uploading {
  cursor: progress;
}

.wb-file-input {
  display: none;
}

.wb-dropzone-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.wb-dropzone-icon {
  width: 80px;
  height: 80px;
  color: var(--t3);
}

.wb-dropzone-icon svg {
  width: 100%;
  height: 100%;
}

.wb-dropzone-main {
  font: 600 18px var(--display);
  color: var(--t);
  letter-spacing: -.01em;
  margin-bottom: 6px;
}

.wb-dropzone-sub {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 13px var(--sans);
  color: var(--t3);
  flex-wrap: wrap;
  justify-content: center;
}

.wb-dropzone-action {
  color: rgba(220, 155, 90, 0.95);
  font-weight: 600;
}

.wb-dropzone-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
  margin-left: 6px;
  padding-left: 8px;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
}

.wb-upload-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: rgba(255, 255, 255, 0.06);
}

.wb-upload-bar {
  height: 100%;
  background: linear-gradient(90deg, rgba(220, 155, 90, 0.6), rgba(220, 155, 90, 0.95));
  transition: width .3s ease;
}

.wb-upload-error {
  margin-top: 12px;
  font: 13px var(--sans);
  color: #ef6660;
}

/* ============ Resume Cards ============ */
.wb-resumes {
  margin-bottom: 32px;
}

.wb-block-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.wb-block-title {
  font: 700 17px var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
}

.wb-block-meta {
  font: 12px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-block-close {
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--t3);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color .2s ease, border-color .2s ease;
}

.wb-block-close:hover {
  color: var(--t);
  border-color: rgba(255, 255, 255, 0.25);
}

.wb-resumes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
}

.wb-resume-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 1) 0%, rgba(11, 12, 16, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: transform .25s ease, box-shadow .25s ease, border-color .25s ease;
  isolation: isolate;
}

.wb-resume-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.35);
}

.wb-resume-card-active {
  border-color: rgba(220, 155, 90, 0.5);
  background:
    linear-gradient(180deg, rgba(28, 22, 18, 1) 0%, rgba(18, 14, 11, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 155, 90, 0.4) 0%, rgba(220, 155, 90, 0.1) 100%) border-box;
}

.wb-resume-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wb-resume-card-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--t2);
}

.wb-resume-card-icon svg {
  width: 16px;
  height: 16px;
  display: block;
}

.wb-resume-card-tag {
  font: 10px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  background: rgba(220, 155, 90, 0.1);
  border: 1px solid rgba(220, 155, 90, 0.3);
  border-radius: var(--radius-pill);
  padding: 2px 8px;
  letter-spacing: .04em;
}

.wb-resume-card-name {
  font: 600 14px var(--sans);
  color: var(--t);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-resume-card-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
  display: flex;
  gap: 8px;
}

.wb-resume-card-stats {
  display: flex;
  gap: 16px;
  padding: 8px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.wb-resume-stat {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.wb-resume-stat-num {
  font: 700 18px var(--mono);
  color: var(--t);
}

.wb-resume-stat-lb {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-resume-card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wb-resume-card-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

.wb-resume-card-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.wb-status-parsed {
  background: rgba(155, 209, 168, 0.85);
  box-shadow: 0 0 6px rgba(155, 209, 168, 0.45);
}

.wb-status-parsing {
  background: rgba(220, 155, 90, 0.85);
  animation: wb-edot 1.4s ease-in-out infinite;
}

.wb-status-failed {
  background: #ef6660;
}

.wb-resume-card-action {
  font: 12px var(--mono);
  color: var(--t2);
  letter-spacing: .04em;
}

.wb-resume-card-active .wb-resume-card-action {
  color: rgba(220, 155, 90, 0.95);
}

/* === Empty === */
.wb-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 40px 20px;
  gap: 8px;
}

.wb-empty-icon {
  font-size: 32px;
  opacity: .5;
}

.wb-empty-title {
  font: 600 15px var(--display);
  color: var(--t);
}

.wb-empty-sub {
  font-size: 13px;
  color: var(--t3);
}

/* ============ Detail ============ */
.wb-resume-detail {
  padding: 24px 26px;
  background:
    linear-gradient(180deg, rgba(16, 17, 22, 1) 0%, rgba(10, 11, 14, 1) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.03) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  isolation: isolate;
}

.wb-detail-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.wb-refresh-btn {
  height: 30px;
  padding: 0 12px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(220, 155, 90, 0.35);
  background: rgba(220, 155, 90, 0.10);
  color: rgba(255, 224, 190, 0.95);
  font: 600 12px var(--sans);
  cursor: pointer;
}

.wb-refresh-btn:disabled {
  cursor: wait;
  opacity: .62;
}

.wb-eval-overview {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 22px;
  align-items: stretch;
  margin-top: 20px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
}

.wb-eval-score {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 92px;
}

.wb-eval-score-num {
  font: 700 42px/1 var(--sans);
  color: var(--t);
  letter-spacing: 0;
}

.wb-eval-score-label {
  margin-top: 8px;
  font: 12px var(--mono);
  color: var(--t3);
}

.wb-eval-summary {
  min-width: 0;
}

.wb-eval-summary p {
  margin: 10px 0 0;
  color: var(--t2);
  font: 14px/1.7 var(--sans);
}

.wb-eval-status {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 9px;
  border-radius: var(--radius-pill);
  font: 600 12px var(--sans);
  color: rgba(255, 255, 255, 0.88);
  background: rgba(255, 255, 255, 0.07);
}

.wb-eval-ready {
  background: rgba(76, 214, 168, 0.12);
  color: rgba(155, 242, 213, 0.95);
}

.wb-eval-stale,
.wb-eval-missing,
.wb-eval-evaluating {
  background: rgba(255, 215, 112, 0.12);
  color: rgba(255, 230, 160, 0.95);
}

.wb-eval-failed,
.wb-eval-insufficient_data {
  background: rgba(255, 120, 120, 0.12);
  color: rgba(255, 185, 185, 0.95);
}

.wb-dimensions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;
}

.wb-dimension {
  min-width: 0;
  padding: 14px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-sm);
}

.wb-dimension-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font: 600 13px var(--sans);
  color: var(--t);
}

.wb-dimension-head strong {
  font: 700 13px var(--mono);
  color: rgba(220, 155, 90, 0.95);
}

.wb-dimension-bar {
  height: 5px;
  margin-top: 10px;
  overflow: hidden;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.08);
}

.wb-dimension-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(76, 214, 168, 0.9), rgba(220, 155, 90, 0.95));
}

.wb-dimension p {
  margin: 9px 0 0;
  color: var(--t3);
  font: 12px/1.6 var(--sans);
}

.wb-detail-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
  gap: 32px;
  margin-top: 20px;
}

.wb-detail-col {
  min-width: 0;
}

.wb-detail-label {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .06em;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.wb-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wb-tag {
  font: 12px var(--sans);
  color: var(--t2);
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.wb-projects {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-project {
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
}

.wb-project-name {
  font: 600 13px var(--sans);
  color: var(--t);
  margin-bottom: 2px;
}

.wb-project-stack {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-eval-lists {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-top: 24px;
}

.wb-eval-list ul {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-eval-list li {
  color: var(--t2);
  font: 13px/1.6 var(--sans);
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.022);
}

@media (max-width: 1024px) {
  .wb-detail-grid {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .wb-dimensions,
  .wb-eval-lists {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .wb-resume-content {
    padding: 0 20px 60px;
  }
  .wb-dropzone {
    padding: 36px 20px;
  }
  .wb-resumes-grid {
    grid-template-columns: 1fr;
  }

  .wb-eval-overview {
    grid-template-columns: 1fr;
  }
}
</style>
