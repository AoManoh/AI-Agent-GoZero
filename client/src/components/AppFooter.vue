<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-section brand-section">
        <div class="footer-logo">
          <div class="logo-icon">
            <img src="http://pic.aomanoh.com/note/20250903035824302.jpg" alt="AI面试官 Logo" class="logo-image" />
          </div>
          <div class="logo-text">
            <h3>AI 面试官</h3>
            <p>智能编程面试平台</p>
          </div>
        </div>
        <div class="brand-description">
          专业的AI驱动程序员面试模拟平台，<br />
          助力开发者提升技术面试能力，斩获心仪offer
        </div>
      </div>

      <!-- 屏蔽：Footer区域 - 面试类型 - 可恢复 -->
      <!-- <div class="footer-section">
        <h4>
          <span class="section-icon">💼</span>
          面试类型
        </h4>
        <div class="footer-links">
          <a href="#" class="footer-link">
            <span class="link-icon">⚡</span>
            算法面试
          </a>
          <a href="#" class="footer-link">
            <span class="link-icon">🏗️</span>
            系统设计
          </a>
          <a href="#" class="footer-link">
            <span class="link-icon">💻</span>
            编程实战
          </a>
          <a href="#" class="footer-link">
            <span class="link-icon">🎯</span>
            行为面试
          </a>
        </div>
      </div> -->

      <!-- 屏蔽：Footer区域 - 开发者资源 - 可恢复 -->
      <!-- <div class="footer-section">
        <h4>
          <span class="section-icon">🔗</span>
          开发者资源
        </h4>
        <div class="footer-links">
          <a href="https://leetcode.cn" target="_blank" class="footer-link">
            <span class="link-icon">📝</span>
            LeetCode
          </a>
          <a href="https://github.com" target="_blank" class="footer-link">
            <span class="link-icon">⭐</span>
            GitHub
          </a>
          <a
            href="https://stackoverflow.com"
            target="_blank"
            class="footer-link"
          >
            <span class="link-icon">❓</span>
            Stack Overflow
          </a>
          <a
            href="http://wangzhongyang.com/project/go/"
            target="_blank"
            class="footer-link"
          >
            <span class="link-icon">🦆</span>
            实战项目
          </a>
        </div>
      </div> -->

      <!-- 屏蔽：Footer区域 - 联系我们 - 可恢复 -->
      <!-- <div class="footer-section">
        <h4>
          <span class="section-icon">📞</span>
          联系我们
        </h4>
        <div class="footer-links">
          <a href="http://wangzhongyang.com/" class="footer-link">
            <span class="link-icon">📧</span>
            商务合作
          </a>
          <a href="http://wangzhongyang.com/" class="footer-link">
            <span class="link-icon">💬</span>
            技术交流群
          </a>
          <a href="#" class="footer-link">
            <span class="link-icon">📋</span>
            用户协议
          </a>
        </div>
      </div> -->
    </div>

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
@import url("https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600;700&display=swap");

.app-footer {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  position: relative;
  width: 100%;
  margin-top: auto;
  z-index: 2;
  overflow: hidden;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr; /* 只显示品牌区域 */
  gap: 40px;
  padding: 60px 20px 40px;
  position: relative;
  z-index: 1;
}

.footer-section {
  display: flex;
  flex-direction: column;
}

.brand-section {
  gap: 20px;
}

.footer-logo {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.logo-icon {
  width: 50px;
  height: 50px;
  background: linear-gradient(135deg, #6366f1, #0066ff);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 20px rgba(99, 102, 241, 0.4);
  overflow: hidden;
}

.logo-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 12px;
}

.code-brackets {
  font-family: "JetBrains Mono", monospace;
  font-size: 1.5rem;
  font-weight: 600;
  color: white;
}

.logo-text h3 {
  font-family: "Inter", sans-serif;
  font-size: 1.4rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 4px 0;
  background: linear-gradient(135deg, #ffffff, #c7d2fe);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.logo-text p {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.6);
  margin: 0;
  letter-spacing: 1px;
}

.brand-description {
  font-size: 0.95rem;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.6;
  font-family: "Inter", sans-serif;
}

.footer-section h4 {
  font-family: "Inter", sans-serif;
  font-size: 1.1rem;
  font-weight: 600;
  color: #ffffff;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  font-size: 1rem;
}

.footer-links {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.footer-link {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-family: "Inter", sans-serif;
  font-size: 0.9rem;
  padding: 6px 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 6px;
  position: relative;
}

.footer-link:hover {
  color: #6366f1;
  transform: translateX(4px);
}

.footer-link:hover .link-icon {
  transform: scale(1.1);
}

.link-icon {
  font-size: 0.9rem;
  transition: transform 0.3s;
  width: 16px;
  text-align: center;
}

.footer-bottom {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(10px);
}

.footer-bottom-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
}

.copyright {
  font-family: "Inter", sans-serif;
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.copyright-symbol {
  font-weight: 600;
  color: #6366f1;
}

.tech-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(
    135deg,
    rgba(99, 102, 241, 0.2),
    rgba(16, 185, 129, 0.2)
  );
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 20px;
  padding: 6px 12px;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.8rem;
  color: #6366f1;
  font-weight: 500;
}

.badge-icon {
  font-size: 0.9rem;
  animation: rocket 2s ease-in-out infinite;
}

@keyframes rocket {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .footer-content {
    grid-template-columns: 1fr 1fr;
    gap: 30px;
    padding: 50px 20px 30px;
  }

  .brand-section {
    grid-column: 1 / -1;
  }
}

@media (max-width: 768px) {
  .footer-content {
    grid-template-columns: 1fr;
    gap: 30px;
    padding: 40px 20px 30px;
  }

  .footer-logo {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }

  .brand-description {
    text-align: center;
  }

  .footer-section h4 {
    justify-content: center;
  }

  .footer-links {
    align-items: center;
  }

  .footer-bottom-content {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }
}

@media (max-width: 480px) {
  .footer-content {
    padding: 30px 15px 20px;
    gap: 25px;
  }

  .logo-icon {
    width: 40px;
    height: 40px;
  }

  .code-brackets {
    font-size: 1.2rem;
  }

  .logo-text h3 {
    font-size: 1.2rem;
  }

  .footer-section h4 {
    font-size: 1rem;
    margin-bottom: 15px;
  }

  .footer-link {
    font-size: 0.85rem;
    justify-content: center;
  }

  .copyright {
    font-size: 0.8rem;
    text-align: center;
  }

  .tech-badge {
    font-size: 0.75rem;
    padding: 5px 10px;
  }
}

/* 暗色主题增强 */
@media (prefers-color-scheme: dark) {
  .app-footer {
    background: linear-gradient(135deg, #0a0f1c 0%, #1a1f2e 100%);
  }
}

/* 打印样式 */
@media print {
  .app-footer {
    background: white !important;
    color: black !important;
  }

  .footer-link,
  .copyright,
  .tech-badge {
    color: black !important;
  }
}
</style>