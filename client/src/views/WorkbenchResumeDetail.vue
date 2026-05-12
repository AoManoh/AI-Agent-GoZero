<template>
  <!--
    WorkbenchResumeDetail：简历完整详情钻深页（设计图 241 / 261）。
    布局：顶部 banner（候选人 + 评分 + 关闭）+ 主区双栏（左 60% chunks 列表 / 右 40% 5 面板）+ 底部 CTA。
    路由：/workbench/resume/:id（D-Q3 独立路由策略）。
    入口：主面板右栏 [看完整详情 →] router.push（D-U3 保留历史）。
    本文件由 C1 commit 创建为占位骨架，C8 commit 填充完整内容。
    详见：docs/requirements/2026-05-12-workbench-resume-redesign.md §6.3 + §7.3 + 分镜 5/5。
  -->
  <WorkbenchLayout>
    <div class="wb-rd-page">
      <!-- 顶部 banner：返回简历库 + 关闭 × -->
      <header class="wb-rd-banner">
        <div class="wb-rd-breadcrumb">
          <RouterLink :to="{ name: 'WorkbenchResume' }" class="wb-rd-back-link">
            ← 返回简历库
          </RouterLink>
          <span class="wb-rd-sep" aria-hidden="true">/</span>
          <span class="wb-rd-current">简历完整详情</span>
        </div>
        <button
          type="button"
          class="wb-rd-close"
          @click="goBack"
          aria-label="关闭详情"
          title="返回简历库"
        >×</button>
      </header>

      <!-- 加载中 -->
      <div v-if="loading" class="wb-rd-loading">
        <div class="wb-rd-spinner" aria-hidden="true"></div>
        <p>加载完整详情中…</p>
      </div>

      <!-- 错误 / 简历不存在 -->
      <div v-else-if="error" class="wb-rd-error">
        <p class="wb-rd-error-title">无法加载简历详情</p>
        <p class="wb-rd-error-meta">{{ error }}</p>
        <button type="button" class="wb-rd-retry" @click="loadDetail">重试</button>
      </div>

      <!-- 主体内容 -->
      <template v-else-if="detail">
        <!-- 简历元数据头 -->
        <div class="wb-rd-meta-head">
          <div class="wb-rd-meta-text">
            <h1 class="wb-rd-title">{{ detail.name || '未命名简历' }}</h1>
            <div class="wb-rd-meta-stats">
              <span v-if="detail.skillCount">{{ detail.skillCount }} 技能</span>
              <span v-if="detail.skillCount" aria-hidden="true">·</span>
              <span>{{ chunks.length }} chunks</span>
              <span aria-hidden="true">·</span>
              <span>{{ formatTimestamp(detail.uploadedAt) }}</span>
              <span v-if="detail.evaluationMeta?.scoreSource" aria-hidden="true">·</span>
              <span v-if="detail.evaluationMeta?.scoreSource" class="wb-rd-meta-source">{{ scoreSourceLabel(detail.evaluationMeta.scoreSource) }}</span>
            </div>
          </div>
          <div class="wb-rd-overall" v-if="typeof detail.overallScore === 'number'">
            <span class="wb-rd-overall-num">{{ Math.round(detail.overallScore) }}</span>
            <span class="wb-rd-overall-meta">/ 100</span>
          </div>
        </div>

        <!-- AI 总结 banner -->
        <p v-if="detail.summary" class="wb-rd-summary">{{ detail.summary }}</p>

        <!-- 双栏 body：左 chunks / 右评估 -->
        <div class="wb-rd-body">
          <!-- 左 60%：原文 chunks -->
          <main class="wb-rd-chunks-col">
            <h2 class="wb-rd-section-title">
              <span>简历原文</span>
              <span class="wb-rd-section-meta">{{ chunks.length }} chunks</span>
            </h2>
            <ol v-if="chunks.length > 0" class="wb-rd-chunk-list">
              <li
                v-for="chunk in chunks"
                :key="chunk.index"
                class="wb-rd-chunk-card"
                :data-chunk-index="chunk.index"
              >
                <header class="wb-rd-chunk-head">
                  <span class="wb-rd-chunk-num">chunk #{{ String(chunk.index ?? '?').padStart(2, '0') }}</span>
                </header>
                <p class="wb-rd-chunk-content">{{ chunk.content }}</p>
              </li>
            </ol>
            <p v-else class="wb-rd-chunk-empty">暂无可显示的原文片段</p>
          </main>

          <!-- 右 40%：评估面板 -->
          <aside class="wb-rd-eval-col">
            <!-- 5 维评估 -->
            <section v-if="filteredDimensions.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">能力评估</h3>
              <div class="wb-rd-dimensions">
                <div v-for="dim in filteredDimensions" :key="dim.key" class="wb-rd-dim">
                  <div class="wb-rd-dim-head">
                    <span class="wb-rd-dim-label">{{ dim.label }}</span>
                    <span class="wb-rd-dim-score">{{ dim.score }}</span>
                  </div>
                  <div class="wb-rd-dim-bar" aria-hidden="true">
                    <span :style="{ width: `${Math.min(100, Math.max(0, dim.score || 0))}%` }"></span>
                  </div>
                  <p v-if="dim.summary" class="wb-rd-dim-summary">{{ dim.summary }}</p>
                </div>
              </div>
            </section>

            <!-- 强项 -->
            <section v-if="detail.strengths?.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">强项 <span class="wb-rd-panel-count">{{ detail.strengths.length }}</span></h3>
              <ul class="wb-rd-list">
                <li v-for="s in detail.strengths" :key="s">{{ s }}</li>
              </ul>
            </section>

            <!-- 风险（三色 left border） -->
            <section v-if="detail.risks?.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">风险 <span class="wb-rd-panel-count">{{ detail.risks.length }}</span></h3>
              <ul class="wb-rd-risks">
                <li
                  v-for="r in detail.risks"
                  :key="r.key || r.label"
                  :class="`wb-rd-risk-${r.severity || 'medium'}`"
                >
                  <span class="wb-rd-risk-label">{{ r.label }}</span>
                  <span v-if="r.suggestion" class="wb-rd-risk-suggest">{{ r.suggestion }}</span>
                </li>
              </ul>
            </section>

            <!-- 建议 -->
            <section v-if="detail.suggestions?.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">建议 <span class="wb-rd-panel-count">{{ detail.suggestions.length }}</span></h3>
              <ul class="wb-rd-list">
                <li v-for="s in detail.suggestions" :key="s">{{ s }}</li>
              </ul>
            </section>

            <!-- AI 追问全部（不限 6 题） -->
            <section v-if="detail.suggestedQuestions?.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">AI 追问 <span class="wb-rd-panel-count">{{ detail.suggestedQuestions.length }}</span></h3>
              <div class="wb-rd-questions">
                <article
                  v-for="(q, i) in detail.suggestedQuestions"
                  :key="q.key || `q-${i}`"
                  class="wb-rd-question"
                >
                  <header class="wb-rd-question-head">
                    <span class="wb-rd-question-num">#{{ String(i + 1).padStart(2, '0') }}</span>
                    <span v-if="q.difficultyLabel" class="wb-rd-question-chip wb-rd-question-chip-difficulty">{{ q.difficultyLabel }}</span>
                    <span v-if="q.focusLabel" class="wb-rd-question-chip wb-rd-question-chip-focus">{{ q.focusLabel }}</span>
                  </header>
                  <p class="wb-rd-question-title">{{ q.title || q.prompt }}</p>
                  <p
                    v-if="q.expectedSignals?.length"
                    class="wb-rd-question-signals"
                  >期望信号：{{ q.expectedSignals.join(' / ') }}</p>
                </article>
              </div>
            </section>

            <!-- 评估证据 evidence -->
            <section v-if="detail.evidence?.length > 0" class="wb-rd-panel">
              <h3 class="wb-rd-panel-title">评估证据 <span class="wb-rd-panel-count">{{ detail.evidence.length }}</span></h3>
              <ul class="wb-rd-evidence-list">
                <li v-for="(ev, i) in detail.evidence" :key="i">
                  <span v-if="ev.label" class="wb-rd-evidence-label">{{ ev.label }}</span>
                  <span class="wb-rd-evidence-text">{{ ev.text || ev.content || JSON.stringify(ev) }}</span>
                </li>
              </ul>
            </section>
          </aside>
        </div>

        <!-- 底部 CTA bar -->
        <div class="wb-rd-cta-bar">
          <button
            type="button"
            class="wb-rd-cta-secondary"
            :disabled="regenerating"
            @click="handleRegenerate"
          >
            <span aria-hidden="true">↻</span>
            <span>{{ regenerating ? '生成中…' : '重新生成 AI 画像' }}</span>
          </button>
          <!-- TODO(phase2-resume-cta-handler): 后端 CreateSession 接受 resumeId 后改为
               apiService.user.createSession({ resumeId, ... }) 直接进面试。
               触发条件：@d:\Go-Project\GoZero-AI\api\user\user.api 中 CreateSessionReq 出现 resumeId 字段。 -->
          <button
            type="button"
            class="wb-rd-cta-primary"
            @click="handleStartInterview"
          >
            <span>用这份简历开始面试</span>
            <span aria-hidden="true">→</span>
          </button>
        </div>
      </template>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

const route = useRoute();
const router = useRouter();

// 简历 artifact ID（从动态路由 :id 取）。
const resumeId = computed(() => String(route.params.id || ""));

// === 状态 ===
const loading = ref(true);
const error = ref("");
const detail = ref(null);
const chunks = ref([]);
const regenerating = ref(false);

// 与主页 WorkbenchResume.vue 同步的过滤策略，保持视觉与计算一致。
const DIMENSION_OMITTED_KEYS = new Set(["target_alignment", "interview_readiness"]);

const filteredDimensions = computed(() => {
  const dims = detail.value?.dimensions || [];
  return dims.filter((d) => !DIMENSION_OMITTED_KEYS.has(d.key));
});

// ISO 时间 → 本地易读格式（YYYY-MM-DD HH:mm）
const formatTimestamp = (iso) => {
  if (!iso) return "—";
  try {
    const date = new Date(iso);
    if (Number.isNaN(date.getTime())) return iso;
    return date.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch (e) {
    return iso;
  }
};

const scoreSourceLabel = (source) => {
  switch ((source || "").toLowerCase()) {
    case "llm":
      return "LLM 评估";
    case "heuristic":
    case "fallback":
      return "规则降级";
    default:
      return "未知来源";
  }
};

// 加载详情：并发拉 detail（chunks）和 analysis（评估）。
// 容错：任一接口失败仍然显示对方数据；都失败时进 error 状态。
const loadDetail = async () => {
  if (!resumeId.value) {
    error.value = "未指定简历 ID";
    loading.value = false;
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    const [detailRes, analysisRes] = await Promise.allSettled([
      apiService.user.resumeArtifactDetail(resumeId.value),
      apiService.user.resumeArtifactAnalysis(resumeId.value, { limit: 50 }),
    ]);

    let detailData = null;
    let chunksData = [];

    if (detailRes.status === "fulfilled" && detailRes.value) {
      const r = detailRes.value;
      // 后端 ResumeArtifactDetailResp.Artifact 是嵌套 ResumeArtifactItem 对象。
      // 元信息字段（title / artifactId / uploadedAt）都在 artifact 子对象里，不是顶层。
      const artifact = r.artifact || {};
      detailData = {
        id: artifact.artifactId || resumeId.value,
        name: artifact.title || "未命名简历",
        uploadedAt: artifact.uploadedAt || "",
        status: artifact.status || "",
      };
      chunksData = Array.isArray(r.chunks) ? r.chunks : [];
    }

    if (analysisRes.status === "fulfilled" && analysisRes.value) {
      const a = analysisRes.value;
      detailData = {
        ...(detailData || { id: resumeId.value }),
        ...detailData,
        // analysis 字段补充 detail（detail 优先，因为含上传元信息）
        skillCount: Array.isArray(a.skills) ? a.skills.length : 0,
        overallScore: typeof a.overallScore === "number" ? a.overallScore : null,
        level: a.level || "",
        summary: a.summary || "",
        dimensions: Array.isArray(a.dimensions) ? a.dimensions : [],
        strengths: Array.isArray(a.strengths) ? a.strengths : [],
        risks: Array.isArray(a.risks) ? a.risks : [],
        suggestions: Array.isArray(a.suggestions) ? a.suggestions : [],
        focusMatches: Array.isArray(a.focusMatches) ? a.focusMatches : [],
        suggestedQuestions: Array.isArray(a.suggestedQuestions) ? a.suggestedQuestions : [],
        evidence: Array.isArray(a.evidence) ? a.evidence : [],
        evaluationMeta: a.evaluationMeta || null,
        evaluationStatus: a.evaluationStatus || "",
      };
    }

    if (!detailData) {
      error.value = "简历详情接口返回为空，可能简历已被删除或权限不足";
    } else {
      detail.value = detailData;
      chunks.value = chunksData;
    }
  } catch (err) {
    console.warn('[resume-detail] load failed:', err);
    error.value = err?.message || "加载失败，请稍后重试";
  } finally {
    loading.value = false;
  }
};

// 重新生成 AI 画像。完成后重新 loadDetail 拉最新数据。
const handleRegenerate = async () => {
  if (!resumeId.value || regenerating.value) return;
  regenerating.value = true;
  try {
    await apiService.user.resumeArtifactAnalysisPrepare(resumeId.value, {
      force: true,
      limit: 50,
    });
    await loadDetail();
  } catch (err) {
    console.warn('[resume-detail] regenerate failed:', err);
    error.value = err?.message || "重新生成失败，请稍后再试";
  } finally {
    regenerating.value = false;
  }
};

// 用这份简历开始面试。当前桥接到 /workbench/new?resumeId=:id（WorkbenchNew 已支持
// 该 query 预填表单）。Phase 2 后端 CreateSession 接受 resumeId 后会改为直接调用。
const handleStartInterview = () => {
  if (!resumeId.value) return;
  router.push({ path: '/workbench/new', query: { resumeId: resumeId.value } });
};

// 关闭按钮：优先 router.back() 回到主面板（保留 D-U3 push 历史）；
// 浏览器无历史时回退到 /workbench/resume。
const goBack = () => {
  if (window.history.length > 1) {
    router.back();
  } else {
    router.push({ name: "WorkbenchResume" });
  }
};

onMounted(() => {
  loadDetail();
});
</script>

<style scoped>
.wb-rd-page {
  width: 100%;
  max-width: 1680px;
  margin: 0 auto;
  padding: 0 clamp(20px, 4vw, 56px) 80px;
}

.wb-rd-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28px 0 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-rd-breadcrumb {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font: 13px var(--sans);
  color: var(--t3);
  letter-spacing: .01em;
}

.wb-rd-back-link {
  color: rgba(220, 155, 90, 0.95);
  text-decoration: none;
  font-weight: 600;
  transition: color .2s ease;
}

.wb-rd-back-link:hover {
  color: rgba(255, 200, 140, 1);
  text-decoration: underline;
}

.wb-rd-sep {
  color: rgba(255, 255, 255, 0.2);
}

.wb-rd-current {
  color: var(--t);
  font-weight: 500;
}

.wb-rd-close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: var(--t2);
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color .2s ease, background-color .2s ease, border-color .2s ease;
}

.wb-rd-close:hover {
  color: var(--t);
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.18);
}

/* ============ 加载中 / 错误 ============ */
.wb-rd-loading,
.wb-rd-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 80px 40px;
  margin-top: 40px;
  text-align: center;
  border-radius: var(--radius-lg);
}

.wb-rd-loading {
  background: rgba(255, 255, 255, 0.018);
  border: 1px solid rgba(255, 255, 255, 0.06);
  color: var(--t3);
}

.wb-rd-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(220, 155, 90, 0.20);
  border-top-color: rgba(220, 155, 90, 0.95);
  border-radius: 50%;
  animation: wb-rd-spin 0.9s linear infinite;
}

@keyframes wb-rd-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.wb-rd-error {
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 100, 80, 0.25) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
}

.wb-rd-error-title {
  font: 600 14px var(--sans);
  color: rgba(255, 195, 180, 0.95);
  margin: 0;
}

.wb-rd-error-meta {
  font: 12px/1.55 var(--sans);
  color: var(--t3);
  margin: 0;
  max-width: 480px;
}

.wb-rd-retry {
  margin-top: 6px;
  padding: 8px 18px;
  background: rgba(220, 155, 90, 0.12);
  border: 1px solid rgba(220, 155, 90, 0.40);
  border-radius: var(--radius-sm);
  color: rgba(255, 224, 190, 0.98);
  font: 600 12px var(--sans);
  cursor: pointer;
  transition: background-color .2s ease;
}

.wb-rd-retry:hover {
  background: rgba(220, 155, 90, 0.22);
}

/* ============ 元数据头 + 总评分 ============ */
.wb-rd-meta-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 28px 0 18px;
}

.wb-rd-title {
  font: 700 24px var(--display);
  color: var(--t);
  margin: 0 0 8px;
  letter-spacing: -.02em;
}

.wb-rd-meta-stats {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  color: var(--t3);
  letter-spacing: .03em;
}

.wb-rd-meta-source {
  color: rgba(220, 155, 90, 0.85);
}

.wb-rd-overall {
  display: flex;
  align-items: baseline;
  gap: 4px;
  padding: 8px 18px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(220, 155, 90, 0.30) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
}

.wb-rd-overall-num {
  font: 700 32px/1 var(--display);
  color: rgba(255, 224, 190, 0.98);
  letter-spacing: -.02em;
}

.wb-rd-overall-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

/* AI 总结 banner */
.wb-rd-summary {
  margin: 0 0 24px;
  padding: 16px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.10) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  font: 14px/1.7 var(--sans);
  color: var(--t2);
  isolation: isolate;
}

/* ============ 双栏 body ============ */
.wb-rd-body {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(320px, 1fr);
  gap: clamp(20px, 2vw, 36px);
  align-items: start;
}

@media (max-width: 1024px) {
  .wb-rd-body {
    grid-template-columns: 1fr;
  }
}

.wb-rd-section-title {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  font: 600 14px var(--sans);
  color: var(--t);
  margin: 0 0 14px;
}

.wb-rd-section-meta {
  font: 11px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

/* === 左栏 chunks 列表 === */
.wb-rd-chunk-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-rd-chunk-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
  transition: border-color .2s ease;
}

.wb-rd-chunk-card:hover {
  border-color: rgba(220, 155, 90, 0.25);
}

.wb-rd-chunk-num {
  font: 600 11px var(--mono);
  color: rgba(220, 155, 90, 0.95);
  letter-spacing: .04em;
}

.wb-rd-chunk-content {
  margin: 0;
  font: 13px/1.7 var(--sans);
  color: var(--t2);
  white-space: pre-wrap;
  word-break: break-word;
}

.wb-rd-chunk-empty {
  font: 12px var(--sans);
  color: var(--t3);
  text-align: center;
  padding: 40px 20px;
  border: 1px dashed rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-md);
}

/* === 右栏评估面板 === */
.wb-rd-eval-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.wb-rd-panel {
  padding: 18px 20px;
  background:
    linear-gradient(180deg, rgba(18, 19, 24, 0.85) 0%, rgba(11, 12, 16, 0.85) 100%) padding-box,
    linear-gradient(160deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.025) 100%) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  isolation: isolate;
}

.wb-rd-panel-title {
  font: 600 13px var(--sans);
  color: var(--t);
  margin: 0 0 12px;
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.wb-rd-panel-count {
  font: 600 10px var(--mono);
  color: var(--t3);
  letter-spacing: .04em;
}

/* 5 维评估 */
.wb-rd-dimensions {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.wb-rd-dim {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wb-rd-dim-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.wb-rd-dim-label {
  font: 600 12px var(--sans);
  color: var(--t);
}

.wb-rd-dim-score {
  font: 700 12px var(--mono);
  color: rgba(220, 155, 90, 0.95);
}

.wb-rd-dim-bar {
  height: 6px;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.06);
  overflow: hidden;
}

.wb-rd-dim-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(76, 214, 168, 0.85), rgba(220, 155, 90, 0.95));
  transition: width .5s ease;
}

.wb-rd-dim-summary {
  margin: 0;
  font: 11px/1.55 var(--sans);
  color: var(--t3);
}

/* 强项 / 建议 list */
.wb-rd-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-rd-list li {
  font: 12px/1.6 var(--sans);
  color: var(--t2);
  padding: 7px 10px;
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
}

/* 风险 三色 left border */
.wb-rd-risks {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-rd-risks li {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 8px 11px;
  background: rgba(255, 255, 255, 0.025);
  border-radius: var(--radius-sm);
  border-left: 3px solid rgba(255, 255, 255, 0.15);
}

.wb-rd-risk-high { border-left-color: rgba(220, 100, 80, 0.85); }
.wb-rd-risk-medium { border-left-color: rgba(230, 165, 100, 0.85); }
.wb-rd-risk-low { border-left-color: rgba(220, 200, 130, 0.80); }

.wb-rd-risk-label {
  font: 600 12px var(--sans);
  color: var(--t);
}

.wb-rd-risk-suggest {
  font: 11px/1.55 var(--sans);
  color: var(--t3);
}

/* AI 追问 */
.wb-rd-questions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wb-rd-question {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
}

.wb-rd-question-head {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.wb-rd-question-num {
  font: 600 11px var(--mono);
  color: var(--t3);
}

.wb-rd-question-chip {
  display: inline-flex;
  height: 18px;
  padding: 0 7px;
  font: 600 10px var(--sans);
  border-radius: var(--radius-pill);
  border: 1px solid transparent;
  align-items: center;
}

.wb-rd-question-chip-difficulty {
  background: rgba(76, 174, 230, 0.10);
  border-color: rgba(76, 174, 230, 0.30);
  color: rgba(165, 215, 250, 0.95);
}

.wb-rd-question-chip-focus {
  background: rgba(220, 155, 90, 0.10);
  border-color: rgba(220, 155, 90, 0.30);
  color: rgba(255, 224, 190, 0.95);
}

.wb-rd-question-title {
  margin: 0;
  font: 13px/1.55 var(--sans);
  color: var(--t);
}

.wb-rd-question-signals {
  margin: 0;
  font: 11px/1.5 var(--sans);
  color: var(--t3);
}

/* evidence */
.wb-rd-evidence-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.wb-rd-evidence-list li {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 7px 10px;
  background: rgba(255, 255, 255, 0.022);
  border-radius: var(--radius-sm);
}

.wb-rd-evidence-label {
  font: 600 11px var(--mono);
  color: rgba(220, 155, 90, 0.85);
  letter-spacing: .04em;
}

.wb-rd-evidence-text {
  font: 12px/1.55 var(--sans);
  color: var(--t2);
  white-space: pre-wrap;
  word-break: break-word;
}

/* ============ 底部 CTA bar ============ */
.wb-rd-cta-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 36px;
  padding-top: 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.wb-rd-cta-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: var(--radius-sm);
  color: var(--t2);
  font: 600 13px var(--sans);
  cursor: pointer;
  transition: background-color .2s ease, border-color .2s ease, color .2s ease;
}

.wb-rd-cta-secondary:hover:not(:disabled) {
  color: rgba(255, 224, 190, 0.95);
  border-color: rgba(220, 155, 90, 0.40);
  background: rgba(220, 155, 90, 0.08);
}

.wb-rd-cta-secondary:disabled {
  cursor: wait;
  opacity: 0.6;
}

.wb-rd-cta-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 26px;
  background: linear-gradient(135deg, rgba(220, 155, 90, 0.95), rgba(200, 130, 65, 0.95));
  border: 1px solid rgba(220, 155, 90, 0.6);
  border-radius: var(--radius-md);
  color: rgba(20, 12, 6, 0.95);
  font: 700 14px var(--sans);
  letter-spacing: .01em;
  cursor: pointer;
  transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease;
  box-shadow: 0 4px 14px rgba(220, 155, 90, 0.20);
}

.wb-rd-cta-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(220, 155, 90, 0.30);
  opacity: 0.96;
}

@media (max-width: 768px) {
  .wb-rd-meta-head {
    flex-direction: column;
    gap: 14px;
  }
  .wb-rd-cta-bar {
    flex-direction: column-reverse;
    align-items: stretch;
  }
}
</style>
