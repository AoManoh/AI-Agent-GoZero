<template>
  <footer class="app-footer">
    <div class="footer-bottom">
      <div class="footer-bottom-content">
        <div class="copyright">
          <span class="copyright-symbol">©</span>
          {{ currentYear }} AI 面试官
        </div>
        <div v-if="isDev" class="footer-build-tag" title="开发环境构建、仅 dev 可见">
          dev build
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { computed } from "vue";

// 计算当前年份
const currentYear = computed(() => new Date().getFullYear());

/**
 * 视觉打磨 Phase 1 包 3（2026-05-11）：dev build 标签。
 * import.meta.env.MODE 是 Vite 内置环境标识（'development' / 'production' / 'test'），
 * 仅在 dev server 下为 true。产环境不显示该标签，用户看到只是单一 © 版权。
 * 避免显示「假版本号」「假部署时间」等与 AGENTS 原则 5 冲突的信息。
 */
const isDev = import.meta.env.MODE === "development";
</script>

<style scoped>
/* 与 SiteHeader 镜像一致：透明 bg + 顶部细分隔线 + outer 0 44 padding +
   inner max-width 1320 + height 80。这样 footer 看起来就是 "底部 nav"，
   与顶部 SiteHeader 视觉重量对称。 */
.app-footer {
  width: 100%;
  margin-top: auto;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  position: relative;
  z-index: 2;
}

/* outer：调 padding、不控高，在包裹中仅提供左右 44px 边距 */
.footer-bottom {
  padding: 0 44px;
}

/* inner：与 .site-header-inner 同构：max-width 1320 + margin auto +
   padding 0 + height 80 + flex space-between（左 © / 右 dev tag 对称）。
   视觉打磨 Phase 1 包 3（2026-05-11）：justify-content center → space-between，
   与顶部 SiteHeader「logo + actions」两端布局镜像对称。 */
.footer-bottom-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0;
  height: 80px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.copyright {
  font-family: var(--sans);
  color: var(--t3);
  font-size: 0.9rem;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  text-align: left;
}

/* 版权符号用暖琥珀色，与 Home step-num 和 metrics 中的点缀琥珀色呼应 */
.copyright-symbol {
  font-weight: 600;
  color: rgba(220, 155, 90, 0.9);
}

/* 视觉打磨 Phase 1 包 3（2026-05-11）：dev build 标签。
   微迷你风 monospace + 35% 白文字 + 8% 白边框 + 12px 字号，
   存在感低但提醒 dev 环境。产环境 v-if="false" 不渲染。 */
.footer-build-tag {
  font: 12px/1.4 var(--mono, ui-monospace, "SF Mono", Monaco, Consolas, monospace);
  color: rgba(255, 255, 255, 0.4);
  letter-spacing: 0.06em;
  padding: 3px 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  user-select: none;
}

@media (max-width: 768px) {
  .footer-bottom {
    padding: 0 20px;
  }

  .footer-bottom-content {
    height: 64px;
  }

  .copyright {
    font-size: 0.85rem;
  }

  .footer-build-tag {
    font-size: 11px;
    padding: 2px 6px;
  }
}
</style>