<template>
  <header :class="['site-header', { 'site-header-fixed': fixed, 'site-header-scrolled': fixed && isScrolled }]">
    <div class="site-header-outer">
      <div class="site-header-inner">
        <slot name="left">
          <router-link class="nav-brand" to="/" aria-label="AI 面试官 首页">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-hidden="false" role="img">
              <title>AI 面试官 logo</title>
              <rect x="0.5" y="3" width="23" height="18" rx="3.5" fill="rgba(220, 155, 90, 0.15)" stroke="rgba(220, 155, 90, 0.45)" stroke-width="1"/>
              <text x="3.5" y="17" font-family="'DM Sans', system-ui, sans-serif" font-size="13" font-weight="800" fill="rgba(255,255,255,0.95)">A</text>
              <text x="13" y="17" font-family="'DM Sans', system-ui, sans-serif" font-size="13" font-weight="800" fill="rgba(220, 155, 90, 0.95)">I</text>
            </svg>
            <span>面试官</span>
          </router-link>
        </slot>

        <slot name="center"></slot>

        <nav class="main-nav">
          <slot name="actions"></slot>
        </nav>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";

const props = defineProps({
  /**
   * Home 等营销页需要 fixed 导航（滚动时贴顶），配合调用方 hero-outer margin-top: 80px
   * 的空间补偿；Login / Register 等子页面默认 non-fixed，作为 auth-page flex 列内的
   * 第一行占据文档流高度。保持原三槽位 (left / center / actions) 调用契约不变。
   */
  fixed: {
    type: Boolean,
    default: false,
  },
});

/**
 * 视觉打磨 Phase 1 包 1（2026-05-11）：滚动状态生效。
 * 此前 .site-header-fixed 写了 transition 但无 scroll listener，状态永远不切换 —
 * 现引入 isScrolled ref + window scroll 监听，超过 SCROLL_THRESHOLD（8px）后切换
 * .site-header-scrolled class，让 fixed header 从 100% 透明过渡到毛玻璃 + 底边线，
 * 滚回顶部时回到透明。non-fixed 模式（Login / Register / Workbench layout 内嵌
 * 调用方）不挂监听，避免无效订阅。
 */
const isScrolled = ref(false);
const SCROLL_THRESHOLD = 8;

const handleScroll = () => {
  isScrolled.value = window.scrollY > SCROLL_THRESHOLD;
};

onMounted(() => {
  if (!props.fixed) return;
  window.addEventListener("scroll", handleScroll, { passive: true });
  handleScroll();
});

onBeforeUnmount(() => {
  if (!props.fixed) return;
  window.removeEventListener("scroll", handleScroll);
});
</script>

<style scoped>
.site-header {
  width: 100%;
  flex-shrink: 0;
  z-index: 10;
}

/* Home 场景：fixed 在视口顶部，与 hero-outer margin-top: 80px 配合。
   初始 100% 透明 — 视觉上 hero aurora 不被横条切断；滚动后通过
   .site-header-scrolled 切换出毛玻璃背景 + 1px 底边线，与下方内容形成分层。 */
.site-header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background-color: transparent;
  border-bottom: 1px solid transparent;
  transition: background-color 0.35s ease, border-color 0.35s ease,
    backdrop-filter 0.35s ease, -webkit-backdrop-filter 0.35s ease;
}

/* 滚动 > 8px 后激活：72% 不透明深底 + saturate 140% / blur 12px 毛玻璃 +
   8% 白色底边线。色值与 page-bg #020204 同色系，避免突兀色块。 */
.site-header-scrolled {
  background-color: rgba(2, 2, 4, 0.72);
  backdrop-filter: saturate(140%) blur(12px);
  -webkit-backdrop-filter: saturate(140%) blur(12px);
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

/* outer 与 hero-outer 同构：padding 0 44 推进子元素。
   配合全局 html { scrollbar-gutter: stable }，fixed/static 两种模式下
   内层容器的可用宽度基准一致。 */
.site-header-outer {
  padding: 0 44px;
}

/* inner 与 .hero 同构：max-width 1320 + margin auto + padding 0。
   height 80 提升垂直 breathing room；logo 左缘与 hero-title 左缘、
   按钮右缘与 demo-win 右缘像素级对齐。 */
.site-header-inner {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0;
  height: 80px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-brand {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  font: 600 15px var(--sans);
  color: var(--t);
  text-decoration: none;
  letter-spacing: -0.01em;
  transition: color 0.2s ease;
}

/* 视觉打磨 Phase 1 包 4（2026-05-11）：nav-brand hover 暖琥珀反馈。
   与 AppFooter copyright-symbol 的 rgba(220, 155, 90, 0.9) 同色系呼应。
   SVG 内部 stroke 用 rgba 硬编码不响应 currentColor，故仅文字变色，
   icon 保持稳定，避免 hover 时图标抖动。 */
.nav-brand:hover {
  color: rgba(220, 155, 90, 0.95);
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* 小屏收敛：outer padding 内缩到 20px，inner height 80 → 64 减少手机端竖向占用 */
@media (max-width: 768px) {
  .site-header-outer {
    padding: 0 20px;
  }
  .site-header-inner {
    height: 64px;
  }
}
</style>
