<template>
  <header :class="['site-header', { 'site-header-fixed': fixed }]">
    <div class="site-header-outer">
      <div class="site-header-inner">
        <slot name="left">
          <router-link class="nav-brand" to="/" aria-label="AI 面试官 首页">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect width="24" height="24" rx="5" fill="rgba(255,255,255,.08)" stroke="rgba(255,255,255,.12)" stroke-width="1"/>
              <circle cx="12" cy="12" r="7" stroke="rgba(255,255,255,.45)" stroke-width="1"/>
              <line x1="7" y1="10" x2="13" y2="10" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/>
              <line x1="11" y1="14" x2="17" y2="14" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <span>AI 面试官</span>
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
defineProps({
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
</script>

<style scoped>
.site-header {
  width: 100%;
  flex-shrink: 0;
  z-index: 10;
}

/* Home 场景：fixed 在视口顶部，与 hero-outer margin-top: 80px 配合。 */
.site-header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  border-bottom: 1px solid transparent;
  transition: border-color 0.3s ease, background-color 0.3s ease;
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
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* 小屏收敛：outer padding 内缩到 20px，inner 不变 */
@media (max-width: 768px) {
  .site-header-outer {
    padding: 0 20px;
  }
}
</style>
