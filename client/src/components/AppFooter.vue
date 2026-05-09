<template>
  <footer class="app-footer">
    <div class="footer-bottom">
      <div class="footer-bottom-content">
        <div class="copyright">
          <span class="copyright-symbol">©</span>
          {{ currentYear }} AI 面试官 - 让每一次面试都成为成长的机会
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { computed } from "vue";

// 计算当前年份
const currentYear = computed(() => new Date().getFullYear());
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
   padding 0 + height 80 + flex 居中。 */
.footer-bottom-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0;
  height: 80px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.copyright {
  font-family: var(--sans);
  color: var(--t3);
  font-size: 0.9rem;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  text-align: center;
}

/* 版权符号用暖琥珀色，与 Home step-num 和 metrics 中的点缀琥珀色呼应 */
.copyright-symbol {
  font-weight: 600;
  color: rgba(220, 155, 90, 0.9);
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
}
</style>