<template>
  <!--
    报告中心占位 SFC（v0.1）
    定位：本轮工作台 4 卡重构（决策 5=A）只新增路由 + 占位页面，让 4 卡的"报告中心"
         CTA 跳到这里不报错。完整 SFC 由独立需求阶段（推荐 2026-05-XX-report-center-frontend.md）实现。

    当前可用替代路径：
    - 上次会话回看：/chat?sessionId=xxx&mode=replay（如果用户有 recentSessions）
    - 工作台首页：/workbench

    后端依赖（独立阶段实现时消费）：
    - GET /api/users/report-center/overview
    - GET /api/users/report-center/sessions
    - GET /api/users/report-center/modes
    - GET /api/users/report-center/modes/:modeKey
    - GET /api/users/report-center/bootstrap
  -->
  <WorkbenchLayout>
    <div class="reports-page">
      <section class="reports-hero">
        <div class="reports-eyebrow">
          <span class="reports-eyebrow-dot" aria-hidden="true"></span>
          <span>建设中 · v0.1 占位</span>
        </div>
        <h1 class="reports-title">
          <span class="reports-title-greet">报告中心</span>
        </h1>
        <p class="reports-sub">
          完整版报告中心正在建设中，将提供跨会话能力分析、模式分布、单场报告与复盘建议。
        </p>

        <div class="reports-card">
          <div class="reports-card-eyebrow">当前可用</div>
          <h2 class="reports-card-title">{{ replaceCardTitle }}</h2>
          <p class="reports-card-desc">{{ replaceCardDesc }}</p>
          <div class="reports-card-foot">
            <router-link
              v-if="hasReplayLink"
              :to="replayLink"
              class="reports-cta reports-cta-amber"
            >
              回看上次对话
              <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M5 3l5 5-5 5" />
              </svg>
            </router-link>
            <router-link to="/workbench/new" class="reports-cta reports-cta-text">
              {{ hasReplayLink ? '新建一场' : '开始首场练习' }}
              <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M5 3l5 5-5 5" />
              </svg>
            </router-link>
          </div>
        </div>

        <div class="reports-roadmap">
          <div class="reports-roadmap-title">即将上线</div>
          <ul class="reports-roadmap-list">
            <li v-for="item in roadmap" :key="item.key" class="reports-roadmap-item">
              <span class="reports-roadmap-dot" aria-hidden="true"></span>
              <div class="reports-roadmap-text">
                <strong>{{ item.label }}</strong>
                <span>{{ item.desc }}</span>
              </div>
            </li>
          </ul>
        </div>

        <router-link to="/workbench" class="reports-back">
          ← 返回工作台
        </router-link>
      </section>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";
import { apiService } from "../composables/useApi";

// 复用 bootstrap：拿最近一场会话用于"回看上次对话"备选 CTA
const recentSessionId = ref("");

const hasReplayLink = computed(() => Boolean(recentSessionId.value));

const replayLink = computed(() =>
  recentSessionId.value
    ? `/chat?sessionId=${encodeURIComponent(recentSessionId.value)}&mode=replay`
    : "/workbench/new"
);

const replaceCardTitle = computed(() =>
  hasReplayLink.value ? "暂时回看上次会话作为复盘" : "完成首场面试后开启复盘"
);

const replaceCardDesc = computed(() =>
  hasReplayLink.value
    ? "完整报告中心建设中。在此期间，你可以回看上次的对话内容、查看 AI 面试官给出的实时评估和建议。"
    : "完成首场面试后，这里会显示你的能力分析、模式表现与复盘建议。先去新建一场试试。"
);

// roadmap 描述将来会有的能力，给用户一个预期边界
const roadmap = [
  {
    key: "ability",
    label: "跨会话能力雷达",
    desc: "汇总所有会话的能力维度，看 30 天 / 90 天能力曲线。",
  },
  {
    key: "modes",
    label: "练习模式分布",
    desc: "按方向 / 难度 / 侧重点统计你的练习偏好与表现。",
  },
  {
    key: "session-detail",
    label: "单场报告深度看板",
    desc: "总分、维度、证据、逐题卡片与建议追问题，一页可查。",
  },
  {
    key: "actions",
    label: "下一步建议",
    desc: "基于弱项推荐下一场练习方向和重点 focus。",
  },
];

onMounted(async () => {
  // 占位 SFC 也接 /api/users/report-center/bootstrap：
  // - 让占位页对接真实接口契约，未来升级到完整 SFC 时数据来源不需要换
  // - bootstrap 同时拿到 overview + modes + 当前 mode 的 reports[]，里面就有最近会话 sessionId
  // - 失败静默降级到 "完成首场面试后开启复盘" 的引导文案
  try {
    const res = await apiService.user.reportCenterBootstrap();
    // ModeDetail.Reports 是 ReportCenterRecentReport[]，第 1 条 = 最近完成的会话
    const reports = Array.isArray(res?.modeDetail?.reports) ? res.modeDetail.reports : [];
    if (reports[0]) {
      recentSessionId.value = reports[0].sessionId || reports[0].id || "";
    }
  } catch (error) {
    // 接口失败 / 401 / 用户未完成任何会话：保持 hasReplayLink=false，
    // 模板会自动切到「完成首场面试后开启复盘」+「开始首场练习」CTA。
  }
});
</script>

<style scoped>
.reports-page {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 44px 80px;
}

.reports-hero {
  display: flex;
  flex-direction: column;
  gap: 26px;
  padding: 60px 0 0;
  /* hero-scoped token 与 Workbench.vue 同源，让占位页与主页同质感 */
  --hero-text-warm: #f3eee7;
  --hero-text-muted: rgba(243, 238, 231, 0.62);
  --hero-text-soft: rgba(243, 238, 231, 0.46);
  --hero-amber: rgba(240, 180, 60, 0.95);
  --hero-amber-soft: rgba(240, 180, 60, 0.14);
}

.reports-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  color: rgba(243, 230, 210, 0.78);
  border: 1px solid rgba(240, 180, 60, 0.18);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  letter-spacing: .04em;
  background: rgba(240, 180, 60, 0.04);
  backdrop-filter: blur(8px);
  white-space: nowrap;
  width: fit-content;
}

.reports-eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(240, 180, 60, 0.95);
  box-shadow: 0 0 6px rgba(240, 180, 60, 0.45);
  animation: reports-edot 2.6s ease-in-out infinite;
}

@keyframes reports-edot {
  0%, 100% { opacity: 1; }
  50% { opacity: .35; }
}

.reports-title {
  font-size: clamp(36px, 3.6vw, 56px);
  font-weight: 800;
  font-family: var(--display);
  line-height: 1.18;
  letter-spacing: -.02em;
  color: var(--t);
  margin: 0;
}

.reports-title-greet {
  /* amber gradient text + 流光：与 Workbench Hero 的 username 同质感 */
  background: linear-gradient(
    100deg,
    rgba(220, 155, 90, 0.95) 0%,
    rgba(255, 230, 200, 0.85) 18%,
    rgba(240, 180, 60, 0.92) 35%,
    rgba(255, 230, 200, 0.85) 52%,
    rgba(240, 180, 60, 0.92) 70%,
    rgba(220, 155, 90, 0.95) 100%
  );
  background-size: 200% 100%;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  -webkit-text-stroke: 1px rgba(240, 180, 60, 0.28);
  animation: reports-shimmer 8s linear infinite;
  filter: drop-shadow(0 0 12px rgba(240, 180, 60, 0.18));
}

@keyframes reports-shimmer {
  0% { background-position: 0% 50%; }
  100% { background-position: 200% 50%; }
}

.reports-sub {
  font-size: 15px;
  line-height: 1.65;
  color: var(--hero-text-muted);
  max-width: 680px;
  margin: 0;
}

/* 主卡：与 Workbench Hero card 同质感 */
.reports-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 36px 38px 30px;
  margin-top: 12px;
  max-width: 720px;
  background:
    linear-gradient(180deg,
      rgba(22, 20, 18, 1) 0%,
      rgba(20, 18, 16, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.10) 0%,
      rgba(255, 255, 255, 0.03) 50%,
      rgba(255, 255, 255, 0.06) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: 20px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    0 24px 70px rgba(0, 0, 0, 0.42),
    0 2px 10px rgba(0, 0, 0, 0.18);
  isolation: isolate;
}

.reports-card-eyebrow {
  font: 11px var(--mono);
  color: rgba(240, 180, 60, 0.95);
  letter-spacing: .24em;
  text-transform: uppercase;
}

.reports-card-title {
  font: 700 26px var(--display);
  color: var(--hero-text-warm);
  letter-spacing: -.01em;
  margin: 0;
}

.reports-card-desc {
  font-size: 14px;
  line-height: 1.7;
  color: var(--hero-text-muted);
  margin: 0;
}

.reports-card-foot {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 10px;
}

.reports-cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 600 14px var(--sans);
  text-decoration: none;
  padding: 12px 20px;
  border-radius: var(--radius-md);
  transition: background-color .25s ease, border-color .25s ease, color .25s ease, transform .25s ease;
  white-space: nowrap;
}

.reports-cta-amber {
  background: rgba(240, 180, 60, 0.92);
  color: rgba(20, 18, 14, 1);
  border: 1px solid transparent;
  font-weight: 700;
}

.reports-cta-amber:hover {
  background: rgba(255, 200, 100, 1);
  transform: translateY(-1px);
}

.reports-cta-text {
  background: transparent;
  color: var(--hero-text-warm);
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.reports-cta-text:hover {
  border-color: rgba(240, 180, 60, 0.40);
  color: rgba(240, 180, 60, 0.95);
  background: rgba(240, 180, 60, 0.05);
}

/* roadmap：横向 list，同色族 */
.reports-roadmap {
  margin-top: 18px;
  max-width: 720px;
}

.reports-roadmap-title {
  font: 11px var(--mono);
  color: var(--hero-text-soft);
  letter-spacing: .24em;
  text-transform: uppercase;
  margin-bottom: 14px;
}

.reports-roadmap-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px;
}

.reports-roadmap-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 18px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  transition: border-color .25s ease, background-color .25s ease;
}

.reports-roadmap-item:hover {
  border-color: rgba(240, 180, 60, 0.22);
  background: rgba(240, 180, 60, 0.03);
}

.reports-roadmap-dot {
  flex-shrink: 0;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(240, 180, 60, 0.65);
  margin-top: 7px;
}

.reports-roadmap-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.reports-roadmap-text strong {
  font: 600 14px var(--sans);
  color: var(--hero-text-warm);
}

.reports-roadmap-text span {
  font-size: 12.5px;
  line-height: 1.55;
  color: var(--hero-text-muted);
}

.reports-back {
  font: 500 13px var(--sans);
  color: var(--hero-text-soft);
  text-decoration: none;
  margin-top: 8px;
  width: fit-content;
  transition: color .2s ease;
}

.reports-back:hover {
  color: rgba(240, 180, 60, 0.95);
}

/* 响应式：< 720px 主卡 padding 收敛 */
@media (max-width: 720px) {
  .reports-page {
    padding: 0 24px 60px;
  }
  .reports-card {
    padding: 28px 24px 22px;
  }
  .reports-card-foot {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
