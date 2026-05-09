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
/* 简化版：删除品牌区，只保留底部一行 copyright 居中。
   仍与 Home 哑光黑视觉语言连续，顶部一条细分隔线与 metrics 结束过渡。 */
.app-footer {
  background: rgba(2, 2, 4, 0.6);
  width: 100%;
  margin-top: auto;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  position: relative;
  z-index: 2;
}

.footer-bottom {
  background: rgba(2, 2, 4, 0.4);
}

/* 与 hero-outer 同构的对齐基准：max-width 1320 + margin auto + padding 24 44。
   justify-content: center 让版权文本水平居中于整个 footer。 */
.footer-bottom-content {
  max-width: 1320px;
  margin: 0 auto;
  padding: 24px 44px;
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
  .footer-bottom-content {
    padding: 18px 20px;
  }

  .copyright {
    font-size: 0.85rem;
  }
}
</style>