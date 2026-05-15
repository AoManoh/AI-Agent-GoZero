<template>
  <WorkbenchLayout>
    <div class="reports-page">
      <section class="reports-hero">
        <div class="reports-hero-copy">
          <div class="reports-eyebrow">
            <span class="reports-eyebrow-dot" aria-hidden="true"></span>
            <span>报告中心</span>
          </div>
          <h1 class="reports-title">面试复盘</h1>
          <p class="reports-sub">
            汇总已完成面试的评分、模式表现和复盘建议，选择一场报告查看摘要后再决定下一步练习。
          </p>
        </div>

        <div class="reports-hero-actions">
          <button
            type="button"
            class="reports-action reports-action-ghost"
            :disabled="loading"
            @click="loadReportCenter"
          >
            刷新
          </button>
          <router-link to="/workbench/new" class="reports-action reports-action-primary">
            新建面试
          </router-link>
        </div>
      </section>

      <section v-if="loading" class="reports-loading" aria-live="polite">
        <div class="reports-loading-dot" aria-hidden="true"></div>
        <span>正在加载报告中心...</span>
      </section>

      <section v-else-if="loadError" class="reports-state reports-state-error">
        <div class="reports-state-title">报告中心暂不可用</div>
        <p>{{ loadError }}</p>
        <button type="button" class="reports-action reports-action-primary" @click="loadReportCenter">
          重试加载
        </button>
      </section>

      <section v-else-if="isEmpty" class="reports-state reports-state-empty">
        <div class="reports-state-mark" aria-hidden="true">0</div>
        <div class="reports-state-title">完成首场面试后生成复盘</div>
        <p>报告中心不会展示示例数据。完成一场真实面试后，这里会出现评分、模式表现和复盘建议。</p>
        <router-link to="/workbench/new" class="reports-action reports-action-primary">
          开始首场面试
        </router-link>
      </section>

      <section v-else class="reports-grid">
        <aside class="reports-overview" aria-label="报告总览">
          <div class="reports-panel reports-score-panel">
            <div class="reports-panel-label">平均分</div>
            <div class="reports-score">
              <span>{{ formatScore(totals.averageScore) }}</span>
              <small>/ 100</small>
            </div>
            <div class="reports-meter" aria-hidden="true">
              <span :style="{ width: `${scorePercent(totals.averageScore)}%` }"></span>
            </div>
            <p>{{ scoreSummary }}</p>
          </div>

          <div class="reports-stat-grid">
            <article class="reports-stat">
              <span>{{ totals.totalSessions || 0 }}</span>
              <small>全部会话</small>
            </article>
            <article class="reports-stat">
              <span>{{ totals.readySessions || 0 }}</span>
              <small>可读报告</small>
            </article>
            <article class="reports-stat">
              <span>{{ totals.draftSessions || 0 }}</span>
              <small>草稿状态</small>
            </article>
            <article class="reports-stat">
              <span>{{ totals.resumeBackedSessions || 0 }}</span>
              <small>含简历</small>
            </article>
          </div>

          <div class="reports-panel">
            <div class="reports-panel-head">
              <h2>模式概览</h2>
              <span>{{ activeModeLabel }}</span>
            </div>
            <div class="reports-mode-list">
              <button
                type="button"
                class="reports-mode"
                :class="{ 'reports-mode-active': selectedModeKey === ALL_MODE_KEY }"
                @click="selectMode(ALL_MODE_KEY)"
              >
                <span>全部报告</span>
                <strong>{{ allReportsCount }}</strong>
              </button>
              <button
                v-for="mode in modes"
                :key="mode.modeKey"
                type="button"
                class="reports-mode"
                :class="{ 'reports-mode-active': selectedModeKey === mode.modeKey }"
                @click="selectMode(mode.modeKey)"
              >
                <span>{{ mode.mode }}</span>
                <strong>{{ mode.readySessions || mode.sessionCount || 0 }}</strong>
                <small>{{ attentionLabel(mode.attentionState) }}</small>
              </button>
            </div>
          </div>
        </aside>

        <main class="reports-list-panel" aria-label="报告列表">
          <div class="reports-panel reports-list-head">
            <div>
              <div class="reports-panel-label">最近报告</div>
              <h2>{{ reportListTitle }}</h2>
            </div>
            <span>{{ activeReports.length }} / {{ activeReportsTotal }} 份</span>
          </div>

          <div v-if="modeError" class="reports-inline-error">
            <span>{{ modeError }}</span>
            <button type="button" @click="retryMode">刷新该模式</button>
          </div>

          <div v-if="modeLoading" class="reports-list-loading">正在加载模式报告...</div>

          <div v-else-if="activeReports.length === 0" class="reports-panel reports-list-empty">
            <h3>当前筛选下暂无报告</h3>
            <p>换一个模式查看，或新建一场面试继续积累样本。</p>
          </div>

          <div v-else class="reports-list">
            <button
              v-for="report in activeReports"
              :key="reportKey(report)"
              type="button"
              class="reports-row"
              :class="{ 'reports-row-active': reportKey(report) === selectedReportKey }"
              @click="selectReport(report)"
            >
              <span class="reports-row-score" :class="scoreClass(report.overallScore)">
                {{ formatScore(report.overallScore) }}
              </span>
              <span class="reports-row-body">
                <span class="reports-row-title">{{ reportTitle(report) }}</span>
                <span class="reports-row-summary">{{ report.summary || fallbackSummary(report.status) }}</span>
                <span class="reports-row-meta">
                  {{ reportModeLabel(report) }}
                  <span aria-hidden="true">·</span>
                  {{ statusLabel(report.status) }}
                  <span aria-hidden="true">·</span>
                  {{ formatDate(report.lastActivityAt || report.generatedAt || report.session?.updatedAt) }}
                </span>
              </span>
              <span class="reports-row-tags">
                <span v-if="report.hasResume" class="reports-tag">简历 {{ report.resumeChunks || 0 }}</span>
                <span v-if="report.nextAction" class="reports-tag reports-tag-soft">{{ report.nextAction }}</span>
              </span>
            </button>
          </div>
        </main>

        <aside class="reports-detail" aria-label="报告摘要">
          <div v-if="!selectedReport" class="reports-panel reports-detail-empty">
            <div class="reports-panel-label">报告摘要</div>
            <h2>选择一场报告</h2>
            <p>点击左侧报告后，这里会读取真实摘要、维度评分、风险和下一步建议。</p>
          </div>

          <div v-else class="reports-panel reports-detail-card">
            <div class="reports-detail-head">
              <div>
                <div class="reports-panel-label">{{ reportModeLabel(selectedReport) }}</div>
                <h2>{{ reportTitle(selectedReport) }}</h2>
              </div>
              <button type="button" class="reports-close" aria-label="关闭报告摘要" @click="clearSelectedReport">
                ×
              </button>
            </div>

            <div v-if="detailLoading" class="reports-detail-loading">正在读取报告摘要...</div>

            <div v-else-if="detailError" class="reports-detail-error">
              <strong>摘要暂不可用</strong>
              <p>{{ detailError }}</p>
              <button type="button" class="reports-action reports-action-ghost" @click="reloadSelectedReport">
                重试
              </button>
            </div>

            <div v-else-if="readiness && !readiness.canReadReport" class="reports-readiness">
              <span class="reports-status-chip">{{ statusLabel(readiness.reportStatus) }}</span>
              <h3>报告需要刷新后查看</h3>
              <p>{{ readiness.reason || "当前报告快照未就绪，请先生成摘要。" }}</p>
              <button
                type="button"
                class="reports-action reports-action-primary"
                :disabled="prepareLoading"
                @click="prepareSelectedReport"
              >
                {{ prepareLoading ? "生成中..." : "生成报告摘要" }}
              </button>
            </div>

            <div v-else-if="reportSummary" class="reports-summary">
              <div class="reports-summary-score">
                <span>{{ formatScore(reportSummary.evaluation?.overallScore) }}</span>
                <small>{{ reportSummary.snapshot?.status || reportSummary.evaluation?.status }}</small>
              </div>

              <p class="reports-summary-text">
                {{ reportSummary.snapshot?.summary || reportSummary.evaluation?.summary || "本场报告暂无摘要。" }}
              </p>

              <div v-if="reportDimensions.length > 0" class="reports-dimensions">
                <div v-for="dim in reportDimensions" :key="dim.key" class="reports-dim">
                  <div class="reports-dim-head">
                    <span>{{ dim.label || dim.key }}</span>
                    <strong>{{ dim.score || 0 }}</strong>
                  </div>
                  <div class="reports-dim-bar" aria-hidden="true">
                    <span :style="{ width: `${dimensionPercent(dim)}%` }"></span>
                  </div>
                  <p v-if="dim.summary">{{ dim.summary }}</p>
                </div>
              </div>

              <div class="reports-summary-meta">
                <span>{{ reportSummary.conversation?.messageCount || 0 }} 条消息</span>
                <span>{{ reportSummary.conversation?.userTurns || 0 }} 次回答</span>
                <span>{{ reportSummary.assets?.hasResume ? `含简历 ${reportSummary.assets.resumeChunks || 0} 段` : "未绑定简历" }}</span>
              </div>

              <div v-if="reportSummary.evaluation?.strengths?.length" class="reports-note reports-note-good">
                <h3>强项</h3>
                <ul>
                  <li v-for="item in reportSummary.evaluation.strengths" :key="item">{{ item }}</li>
                </ul>
              </div>

              <div v-if="reportSummary.evaluation?.risks?.length" class="reports-note reports-note-risk">
                <h3>风险</h3>
                <ul>
                  <li v-for="item in reportSummary.evaluation.risks" :key="item">{{ item }}</li>
                </ul>
              </div>

              <div v-if="reportSummary.evaluation?.suggestions?.length" class="reports-note">
                <h3>下一步建议</h3>
                <ul>
                  <li v-for="item in reportSummary.evaluation.suggestions" :key="item">{{ item }}</li>
                </ul>
              </div>

              <router-link to="/workbench/new" class="reports-action reports-action-primary reports-detail-cta">
                按建议再练一场
              </router-link>
            </div>
          </div>
        </aside>
      </section>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

const ALL_MODE_KEY = "all";
const REPORT_LIMIT = 8;

const loading = ref(true);
const loadError = ref("");
const bootstrapData = ref(null);
const selectedModeKey = ref(ALL_MODE_KEY);
const modeDetail = ref(null);
const modeLoading = ref(false);
const modeError = ref("");
const selectedReport = ref(null);
const readiness = ref(null);
const reportSummary = ref(null);
const detailLoading = ref(false);
const detailError = ref("");
const prepareLoading = ref(false);

let modeRequestSeq = 0;
let detailRequestSeq = 0;

const totals = computed(() => bootstrapData.value?.overview?.totals || {});
const modes = computed(() => bootstrapData.value?.modes || []);
const overviewReports = computed(() => bootstrapData.value?.overview?.recentReports || []);
const allReportsCount = computed(() => totals.value.totalSessions || overviewReports.value.length);
const isEmpty = computed(() => allReportsCount.value === 0 && overviewReports.value.length === 0);

const activeReports = computed(() => {
  if (selectedModeKey.value === ALL_MODE_KEY) {
    return overviewReports.value;
  }
  return modeDetail.value?.reports || [];
});

const activeReportsTotal = computed(() => {
  if (selectedModeKey.value === ALL_MODE_KEY) {
    return allReportsCount.value;
  }
  return modeDetail.value?.total || activeReports.value.length;
});

const activeModeLabel = computed(() => {
  if (selectedModeKey.value === ALL_MODE_KEY) return "全部";
  return modes.value.find((mode) => mode.modeKey === selectedModeKey.value)?.mode || "当前模式";
});

const reportListTitle = computed(() =>
  selectedModeKey.value === ALL_MODE_KEY ? "最近报告列表" : `${activeModeLabel.value}报告列表`
);

const selectedReportKey = computed(() => reportKey(selectedReport.value));

const reportDimensions = computed(() => reportSummary.value?.evaluation?.dimensions || []);

const scoreSummary = computed(() => {
  const ready = totals.value.readySessions || 0;
  const total = totals.value.totalSessions || 0;
  if (!ready) return "完成面试并生成评估后，这里会显示平均分。";
  return `${ready} 份可读报告 / ${total} 场会话已进入复盘。`;
});

const loadReportCenter = async () => {
  loading.value = true;
  loadError.value = "";
  try {
    const res = await apiService.user.reportCenterBootstrap({ limit: REPORT_LIMIT });
    bootstrapData.value = res || null;
    selectedModeKey.value = ALL_MODE_KEY;
    modeDetail.value = null;
    modeError.value = "";
    clearSelectedReport();
  } catch (error) {
    loadError.value = error?.message || "报告中心加载失败，请稍后重试。";
  } finally {
    loading.value = false;
  }
};

const selectMode = async (modeKey) => {
  if (selectedModeKey.value === modeKey && modeKey === ALL_MODE_KEY) return;
  selectedModeKey.value = modeKey;
  modeError.value = "";
  clearSelectedReport();

  if (modeKey === ALL_MODE_KEY) {
    modeDetail.value = null;
    return;
  }

  const seq = ++modeRequestSeq;
  modeLoading.value = true;
  try {
    const res = await apiService.user.reportCenterModeDetail(modeKey, { limit: REPORT_LIMIT });
    if (seq !== modeRequestSeq) return;
    modeDetail.value = res || null;
  } catch (error) {
    if (seq !== modeRequestSeq) return;
    modeError.value = error?.message || "该模式报告加载失败，请稍后重试。";
    modeDetail.value = null;
  } finally {
    if (seq === modeRequestSeq) {
      modeLoading.value = false;
    }
  }
};

const retryMode = () => {
  if (selectedModeKey.value !== ALL_MODE_KEY) {
    selectMode(selectedModeKey.value);
  }
};

const selectReport = async (report) => {
  selectedReport.value = report;
  readiness.value = null;
  reportSummary.value = null;
  detailError.value = "";
  await loadSelectedReport();
};

const clearSelectedReport = () => {
  selectedReport.value = null;
  readiness.value = null;
  reportSummary.value = null;
  detailError.value = "";
  prepareLoading.value = false;
};

const reloadSelectedReport = () => {
  if (selectedReport.value) {
    loadSelectedReport();
  }
};

const loadSelectedReport = async () => {
  const sessionId = reportSessionId(selectedReport.value);
  if (!sessionId) {
    detailError.value = "报告缺少会话标识，无法读取摘要。";
    return;
  }

  const seq = ++detailRequestSeq;
  detailLoading.value = true;
  detailError.value = "";
  readiness.value = null;
  reportSummary.value = null;
  try {
    const readyResp = await apiService.user.sessionReportReadiness(sessionId);
    if (seq !== detailRequestSeq) return;
    readiness.value = readyResp;
    if (!readyResp?.canReadReport) {
      return;
    }
    const summary = await apiService.user.sessionReportSummary(sessionId);
    if (seq !== detailRequestSeq) return;
    reportSummary.value = summary || null;
  } catch (error) {
    if (seq !== detailRequestSeq) return;
    detailError.value = error?.message || "报告摘要读取失败，请稍后重试。";
  } finally {
    if (seq === detailRequestSeq) {
      detailLoading.value = false;
    }
  }
};

const prepareSelectedReport = async () => {
  const sessionId = reportSessionId(selectedReport.value);
  if (!sessionId || prepareLoading.value) return;

  prepareLoading.value = true;
  detailError.value = "";
  try {
    const res = await apiService.user.sessionReportPrepare(sessionId, { force: false });
    readiness.value = res?.readiness || readiness.value;
    reportSummary.value = res?.reportSummary || null;
  } catch (error) {
    detailError.value = error?.message || "报告生成失败，请稍后重试。";
  } finally {
    prepareLoading.value = false;
  }
};

const reportSessionId = (report) => report?.session?.sessionId || report?.sessionId || report?.id || "";

const reportKey = (report) => reportSessionId(report) || report?.generatedAt || report?.lastActivityAt || "";

const reportTitle = (report) => report?.session?.title || "未命名面试";

const reportModeLabel = (report) =>
  report?.session?.mode || report?.session?.modeKey || "未标注模式";

const formatScore = (score) => {
  const value = Number(score);
  if (!Number.isFinite(value) || value <= 0) return "—";
  return Math.round(value);
};

const scorePercent = (score) => {
  const value = Number(score);
  if (!Number.isFinite(value) || value <= 0) return 0;
  return Math.max(0, Math.min(100, value));
};

const dimensionPercent = (dimension) => {
  const score = Number(dimension?.score || 0);
  const maxScore = Number(dimension?.maxScore || 100) || 100;
  return Math.max(0, Math.min(100, (score / maxScore) * 100));
};

const scoreClass = (score) => {
  const value = Number(score);
  if (!Number.isFinite(value) || value <= 0) return "reports-score-empty";
  if (value >= 85) return "reports-score-high";
  if (value >= 70) return "reports-score-mid";
  return "reports-score-low";
};

const formatDate = (value) => {
  if (!value) return "暂无时间";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "暂无时间";
  return date.toLocaleDateString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const statusLabel = (status = "") => {
  const labels = {
    ready: "可读",
    readable: "可读",
    draft: "草稿",
    missing: "未生成",
    stale: "需刷新",
    insufficient_data: "样本不足",
  };
  return labels[status] || "待复盘";
};

const fallbackSummary = (status = "") => {
  if (status === "insufficient_data") return "样本不足，建议再完成一场面试。";
  if (status === "ready") return "报告已生成，可查看摘要。";
  return "报告仍处于草稿或待生成状态。";
};

const attentionLabel = (state = "") => {
  const labels = {
    ready: "可复盘",
    draft: "待生成",
    resume_only: "含简历",
    empty: "暂无",
  };
  return labels[state] || "待复盘";
};

onMounted(() => {
  loadReportCenter();
});
</script>

<style scoped>
.reports-page {
  width: min(100%, calc(100vw - clamp(32px, 6vw, 88px)));
  max-width: 1320px;
  margin: 0 auto;
  min-height: calc(100svh - 80px);
  display: flex;
  flex-direction: column;
  padding: clamp(20px, 3vw, 36px) 0 clamp(28px, 4vw, 48px);
  --reports-text-warm: #f3eee7;
  --reports-text-muted: rgba(243, 238, 231, 0.64);
  --reports-text-soft: rgba(243, 238, 231, 0.45);
  --reports-amber: rgba(240, 180, 60, 0.94);
  --reports-amber-soft: rgba(240, 180, 60, 0.12);
  --reports-panel: rgba(18, 18, 18, 0.78);
  --reports-border: rgba(255, 255, 255, 0.08);
}

.reports-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: clamp(18px, 3vw, 36px);
  flex: 0 0 auto;
  margin-bottom: clamp(18px, 2.4vw, 28px);
}

.reports-hero-copy {
  min-width: 0;
  max-width: 720px;
}

.reports-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
  padding: 6px 14px;
  border: 1px solid rgba(240, 180, 60, 0.18);
  border-radius: var(--radius-pill);
  background: rgba(240, 180, 60, 0.04);
  color: rgba(243, 230, 210, 0.78);
  font: var(--fs-xs) var(--mono);
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.reports-eyebrow-dot,
.reports-loading-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--reports-amber);
  box-shadow: 0 0 10px rgba(240, 180, 60, 0.45);
}

.reports-title {
  margin: 16px 0 0;
  color: var(--reports-text-warm);
  font: 800 var(--fs-display) var(--display);
  line-height: 1.16;
  letter-spacing: 0;
}

.reports-sub {
  max-width: 700px;
  margin: 12px 0 0;
  color: var(--reports-text-muted);
  font-size: var(--fs-lg);
  line-height: 1.65;
}

.reports-hero-actions,
.reports-card-foot {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.reports-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 10px 18px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.12);
  font: 700 var(--fs-sm) var(--sans);
  text-decoration: none;
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, transform 0.2s ease;
}

.reports-action:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.reports-action:not(:disabled):hover {
  transform: translateY(-1px);
}

.reports-action-primary {
  border-color: transparent;
  background: rgba(240, 180, 60, 0.92);
  color: rgba(20, 18, 14, 1);
}

.reports-action-primary:not(:disabled):hover {
  background: rgba(255, 205, 110, 1);
}

.reports-action-ghost {
  background: rgba(255, 255, 255, 0.03);
  color: var(--reports-text-warm);
}

.reports-action-ghost:not(:disabled):hover {
  border-color: rgba(240, 180, 60, 0.34);
  background: rgba(240, 180, 60, 0.06);
  color: var(--reports-amber);
}

.reports-loading,
.reports-state,
.reports-panel {
  border: 1px solid var(--reports-border);
  background:
    linear-gradient(180deg, rgba(24, 23, 22, 0.82), rgba(14, 14, 14, 0.76)) padding-box,
    linear-gradient(150deg, rgba(255, 255, 255, 0.09), rgba(255, 255, 255, 0.025)) border-box;
  border-radius: var(--radius-lg);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04), 0 22px 70px rgba(0, 0, 0, 0.32);
}

.reports-loading {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 18px 20px;
  color: var(--reports-text-muted);
}

.reports-loading-dot {
  animation: reports-pulse 1.6s ease-in-out infinite;
}

@keyframes reports-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.8); }
}

.reports-state {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: clamp(36px, 6vw, 72px);
  text-align: center;
}

.reports-state-mark {
  width: 72px;
  height: 72px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-circle);
  border: 1px solid rgba(240, 180, 60, 0.28);
  color: var(--reports-amber);
  font: 800 var(--fs-3xl) var(--display);
  background: rgba(240, 180, 60, 0.08);
}

.reports-state-title {
  color: var(--reports-text-warm);
  font: 800 var(--fs-2xl) var(--display);
}

.reports-state p {
  max-width: 560px;
  color: var(--reports-text-muted);
  line-height: 1.7;
}

.reports-state-error {
  align-items: flex-start;
  text-align: left;
  min-height: 260px;
}

.reports-grid {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(220px, 22%) minmax(0, 1fr) minmax(260px, 28%);
  gap: clamp(16px, 2vw, 28px);
  align-items: stretch;
}

.reports-overview,
.reports-list-panel,
.reports-detail {
  min-width: 0;
  min-height: 0;
}

.reports-overview,
.reports-list-panel {
  display: flex;
  flex-direction: column;
  gap: clamp(14px, 1.6vw, 20px);
}

.reports-list-panel,
.reports-detail {
  overflow: hidden;
}

.reports-detail {
  display: flex;
  flex-direction: column;
}

.reports-panel {
  padding: clamp(18px, 2vw, 24px);
}

.reports-panel-label {
  color: var(--reports-amber);
  font: var(--fs-2xs) var(--mono);
  letter-spacing: 0.16em;
}

.reports-panel h2,
.reports-panel h3 {
  margin: 0;
  color: var(--reports-text-warm);
}

.reports-panel p {
  margin: 0;
  color: var(--reports-text-muted);
  line-height: 1.65;
}

.reports-panel-head,
.reports-list-head,
.reports-detail-head,
.reports-dim-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.reports-list-head span,
.reports-panel-head span {
  color: var(--reports-text-soft);
  font: var(--fs-sm) var(--mono);
  white-space: nowrap;
}

.reports-score-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.reports-score {
  display: flex;
  align-items: baseline;
  gap: 8px;
  color: var(--reports-text-warm);
}

.reports-score span {
  font: 800 3rem var(--display);
  line-height: 1;
}

.reports-score small {
  color: var(--reports-text-soft);
  font-size: var(--fs-md);
}

.reports-meter,
.reports-dim-bar {
  height: 8px;
  overflow: hidden;
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.08);
}

.reports-meter span,
.reports-dim-bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(220, 155, 90, 0.95), rgba(255, 220, 130, 0.95));
}

.reports-stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.reports-stat {
  min-width: 0;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.025);
}

.reports-stat span {
  display: block;
  color: var(--reports-text-warm);
  font: 800 var(--fs-2xl) var(--display);
}

.reports-stat small,
.reports-mode small {
  color: var(--reports-text-soft);
  font-size: var(--fs-xs);
}

.reports-mode-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 16px;
}

.reports-mode {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 4px 10px;
  width: 100%;
  padding: 12px 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.025);
  color: var(--reports-text-muted);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.2s ease, background-color 0.2s ease, color 0.2s ease;
}

.reports-mode:hover,
.reports-mode-active {
  border-color: rgba(240, 180, 60, 0.34);
  background: rgba(240, 180, 60, 0.07);
  color: var(--reports-text-warm);
}

.reports-mode span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 700;
}

.reports-mode strong {
  color: var(--reports-amber);
}

.reports-mode small {
  grid-column: 1 / -1;
}

.reports-inline-error,
.reports-list-loading,
.reports-list-empty {
  color: var(--reports-text-muted);
}

.reports-inline-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(248, 113, 113, 0.24);
  border-radius: var(--radius-md);
  background: rgba(248, 113, 113, 0.07);
}

.reports-inline-error button {
  border: 0;
  background: transparent;
  color: rgba(255, 205, 120, 0.95);
  font-weight: 700;
  cursor: pointer;
}

.reports-list {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reports-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.025);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
}

.reports-row:hover,
.reports-row-active {
  border-color: rgba(240, 180, 60, 0.34);
  background: rgba(240, 180, 60, 0.06);
}

.reports-row:hover {
  transform: translateY(-1px);
}

.reports-row-score {
  width: 50px;
  height: 50px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-md);
  font: 800 var(--fs-xl) var(--display);
  background: rgba(255, 255, 255, 0.05);
}

.reports-score-high { color: rgba(134, 239, 172, 0.95); }
.reports-score-mid { color: rgba(255, 220, 130, 0.95); }
.reports-score-low { color: rgba(252, 165, 165, 0.95); }
.reports-score-empty { color: var(--reports-text-soft); }

.reports-row-body {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 0;
}

.reports-row-title,
.reports-row-summary,
.reports-row-meta {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reports-row-title {
  color: var(--reports-text-warm);
  font-weight: 800;
}

.reports-row-summary {
  color: var(--reports-text-muted);
  font-size: var(--fs-sm);
}

.reports-row-meta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--reports-text-soft);
  font-size: var(--fs-xs);
}

.reports-row-tags {
  display: flex;
  align-items: flex-end;
  flex-direction: column;
  gap: 8px;
  max-width: 150px;
}

.reports-tag,
.reports-status-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 4px 8px;
  border-radius: var(--radius-pill);
  border: 1px solid rgba(240, 180, 60, 0.22);
  color: rgba(255, 215, 145, 0.95);
  background: rgba(240, 180, 60, 0.07);
  font-size: var(--fs-xs);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reports-tag-soft {
  color: var(--reports-text-muted);
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
}

.reports-detail-card,
.reports-detail-empty {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.reports-detail-head h2 {
  margin-top: 6px;
  font-size: var(--fs-2xl);
}

.reports-close {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
  color: var(--reports-text-muted);
  cursor: pointer;
  font-size: var(--fs-xl);
}

.reports-close:hover {
  color: var(--reports-text-warm);
  border-color: rgba(240, 180, 60, 0.28);
}

.reports-detail-loading,
.reports-detail-error,
.reports-readiness {
  display: flex;
  flex-direction: column;
  gap: 12px;
  color: var(--reports-text-muted);
}

.reports-detail-error strong,
.reports-readiness h3 {
  color: var(--reports-text-warm);
}

.reports-summary {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.reports-summary-score {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border-radius: var(--radius-md);
  background: rgba(240, 180, 60, 0.06);
}

.reports-summary-score span {
  color: var(--reports-amber);
  font: 800 3rem var(--display);
  line-height: 1;
}

.reports-summary-score small {
  color: var(--reports-text-muted);
  font: var(--fs-sm) var(--mono);
}

.reports-summary-text {
  color: var(--reports-text-muted);
}

.reports-dimensions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reports-dim {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.reports-dim-head span {
  color: var(--reports-text-warm);
  font-weight: 700;
}

.reports-dim-head strong {
  color: var(--reports-amber);
}

.reports-dim p {
  color: var(--reports-text-soft);
  font-size: var(--fs-xs);
}

.reports-summary-meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.reports-summary-meta span {
  min-width: 0;
  padding: 10px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
  color: var(--reports-text-muted);
  font-size: var(--fs-xs);
  text-align: center;
}

.reports-note {
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.025);
}

.reports-note-good {
  border-color: rgba(134, 239, 172, 0.18);
}

.reports-note-risk {
  border-color: rgba(252, 165, 165, 0.18);
}

.reports-note h3 {
  margin-bottom: 10px;
  font-size: var(--fs-md);
}

.reports-note ul {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding-left: 18px;
  color: var(--reports-text-muted);
}

.reports-detail-cta {
  width: 100%;
}

@media (min-width: 1081px) {
  .reports-page {
    height: calc(100svh - 80px);
    min-height: 680px;
    overflow: hidden;
  }

  .reports-grid,
  .reports-list-panel,
  .reports-detail {
    overflow: hidden;
  }

  .reports-overview,
  .reports-list,
  .reports-detail-card,
  .reports-detail-empty {
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.16) transparent;
    scrollbar-gutter: stable;
  }

  .reports-overview::-webkit-scrollbar,
  .reports-list::-webkit-scrollbar,
  .reports-detail-card::-webkit-scrollbar,
  .reports-detail-empty::-webkit-scrollbar {
    width: 6px;
  }

  .reports-overview::-webkit-scrollbar-track,
  .reports-list::-webkit-scrollbar-track,
  .reports-detail-card::-webkit-scrollbar-track,
  .reports-detail-empty::-webkit-scrollbar-track {
    background: transparent;
  }

  .reports-overview::-webkit-scrollbar-thumb,
  .reports-list::-webkit-scrollbar-thumb,
  .reports-detail-card::-webkit-scrollbar-thumb,
  .reports-detail-empty::-webkit-scrollbar-thumb {
    border-radius: var(--radius-pill);
    background: rgba(255, 255, 255, 0.16);
  }

  .reports-overview::-webkit-scrollbar-thumb:hover,
  .reports-list::-webkit-scrollbar-thumb:hover,
  .reports-detail-card::-webkit-scrollbar-thumb:hover,
  .reports-detail-empty::-webkit-scrollbar-thumb:hover {
    background: rgba(240, 180, 60, 0.32);
  }
}

@media (max-width: 1080px) {
  .reports-page {
    height: auto;
    overflow: visible;
  }

  .reports-grid {
    grid-template-columns: minmax(210px, 30%) minmax(0, 1fr);
    align-items: start;
    overflow: visible;
  }

  .reports-detail {
    grid-column: 1 / -1;
    overflow: visible;
  }

  .reports-list-panel {
    overflow: visible;
  }
}

@media (max-width: 760px) {
  .reports-page {
    width: min(100%, calc(100vw - 32px));
    padding-top: 24px;
  }

  .reports-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .reports-hero-actions {
    justify-content: flex-start;
  }

  .reports-grid {
    grid-template-columns: 1fr;
  }

  .reports-row {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .reports-row-tags {
    grid-column: 1 / -1;
    flex-direction: row;
    align-items: center;
    max-width: 100%;
  }

  .reports-summary-meta {
    grid-template-columns: 1fr;
  }
}
</style>
