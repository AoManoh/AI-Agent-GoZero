<template>
  <!--
    WorkbenchResumeDetail：简历完整详情钻深页（设计图 241 / 261）。
    布局：顶部 banner（候选人 + 评分 + 关闭）+ 主区双栏（左 60% chunks 网格 / 右 40% 5 面板）+ 底部 CTA。
    路由：/workbench/resume/:id（D-Q3 独立路由策略）。
    入口：主面板右栏 [看完整详情 →] router.push（D-U3 保留历史）。
    本文件由 C1 commit 创建为占位骨架，C8 commit 填充完整内容。
    详见：docs/requirements/2026-05-12-workbench-resume-redesign.md §6.3 + §7.3 + 分镜 5/5。
  -->
  <WorkbenchLayout>
    <div class="wb-rd-page">
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

      <section class="wb-rd-placeholder" aria-live="polite">
        <div class="wb-rd-placeholder-icon" aria-hidden="true">
          <svg viewBox="0 0 64 64" fill="none">
            <rect x="14" y="10" width="36" height="44" rx="3" stroke="currentColor" stroke-width="1.5" />
            <line x1="22" y1="22" x2="42" y2="22" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            <line x1="22" y1="30" x2="42" y2="30" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            <line x1="22" y1="38" x2="34" y2="38" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
        </div>
        <h2 class="wb-rd-placeholder-title">钻深页占位</h2>
        <p class="wb-rd-placeholder-text">
          此页为简历完整详情钻深页骨架，将在 C8 commit 填充完整内容（18 chunks 网格 + 5 个评估面板）。
        </p>
        <p class="wb-rd-placeholder-meta">
          当前简历 ID：<code class="wb-rd-id">{{ resumeId || "未指定" }}</code>
        </p>
        <RouterLink :to="{ name: 'WorkbenchResume' }" class="wb-rd-back-btn">
          返回简历库
        </RouterLink>
      </section>
    </div>
  </WorkbenchLayout>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import WorkbenchLayout from "../components/dashboard/WorkbenchLayout.vue";

const route = useRoute();
const router = useRouter();

// 简历 artifact ID（从动态路由 :id 取）。
// 当前为占位渲染；C8 commit 时改为：
//   - onMounted 调 apiService.user.resumeArtifactDetail(resumeId.value) 拉 chunks
//   - 调 apiService.user.resumeArtifactAnalysis(resumeId.value) 拉 5 面板数据
const resumeId = computed(() => String(route.params.id || ""));

// 关闭按钮：优先 router.back() 回到主面板（保留 D-U3 push 历史）；
// 浏览器无历史时回退到 /workbench/resume。
const goBack = () => {
  if (window.history.length > 1) {
    router.back();
  } else {
    router.push({ name: "WorkbenchResume" });
  }
};
</script>

<style scoped>
.wb-rd-page {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 44px 80px;
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

.wb-rd-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  padding: 80px 40px;
  margin-top: 40px;
  text-align: center;
  border: 1px dashed rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.018);
}

.wb-rd-placeholder-icon {
  width: 72px;
  height: 72px;
  color: var(--t3);
  opacity: 0.6;
}

.wb-rd-placeholder-icon svg {
  width: 100%;
  height: 100%;
}

.wb-rd-placeholder-title {
  font: 700 18px var(--display);
  color: var(--t2);
  margin: 0;
  letter-spacing: -.01em;
}

.wb-rd-placeholder-text {
  font: 14px/1.7 var(--sans);
  color: var(--t3);
  margin: 0;
  max-width: 480px;
}

.wb-rd-placeholder-meta {
  font: 12px var(--mono);
  color: var(--t3);
  margin: 0;
  letter-spacing: .04em;
}

.wb-rd-id {
  display: inline-block;
  padding: 2px 10px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 4px;
  color: var(--t2);
  font: 12px var(--mono);
  margin-left: 4px;
}

.wb-rd-back-btn {
  display: inline-flex;
  align-items: center;
  margin-top: 12px;
  padding: 10px 22px;
  background: rgba(220, 155, 90, 0.12);
  border: 1px solid rgba(220, 155, 90, 0.4);
  border-radius: var(--radius-sm);
  color: rgba(255, 224, 190, 0.98);
  font: 600 13px var(--sans);
  text-decoration: none;
  transition: background-color .2s ease, border-color .2s ease;
}

.wb-rd-back-btn:hover {
  background: rgba(220, 155, 90, 0.22);
  border-color: rgba(220, 155, 90, 0.6);
}

@media (max-width: 768px) {
  .wb-rd-page {
    padding: 0 20px 60px;
  }
  .wb-rd-placeholder {
    padding: 48px 20px;
  }
}
</style>
