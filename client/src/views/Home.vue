<template>
  <div class="home-container">
    <!-- Header -->
    <header class="main-header">
      <nav class="navbar">
        <div class="nav-brand">
          <div class="nav-logo">
            <div class="nav-logo-icon">
              <img src="http://pic.aomanoh.com/note/20250903035824302.jpg" alt="AI面试官 Logo" class="nav-logo-image" />
            </div>
            <div class="nav-logo-text">
              <h3>AI 面试官</h3>
              <p>智能编程面试平台</p>
            </div>
          </div>
        </div>
        <router-link to="/interview-master">
        <div class="nav-actions">
          <button class="nav-btn primary" @click="navigateTo('/interview-master')">
            <span>进入体验</span>
          </button>
        </div>
        </router-link>
      </nav>
    </header>

    <!-- 动态背景 -->
    <div class="dynamic-background">
      <div class="gradient-overlay"></div>
      <div class="grid-pattern"></div>
      <div class="floating-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
        <div class="shape shape-4"></div>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <main class="main-section">
      <div class="hero-container">
        <!-- 左侧主要内容 -->
        <div class="hero-content">
          <!-- 品牌介绍区域 -->
          <div class="brand-intro">
            <div class="brand-title-wrapper">
              <h1 class="brand-title">
                <span class="title-main">AI驱动的</span>
                <span class="title-highlight">智能面试平台</span>
              </h1>
              <div class="title-status">
                <span class="status-badge">持续更新中</span>
                <!-- 屏蔽：内容统计信息 - 可恢复 -->
                <!-- <span class="content-count"
                  >已帮助 <strong>10k+</strong> 位求职者</span
                > -->
              </div>
            </div>

            <!-- 屏蔽：品牌描述区域 - 可恢复 -->
            <!-- <div class="brand-description">
              <p class="main-desc">
                在这里，我分享
                <span class="highlight-text">技术面试</span>、<span
                  class="highlight-text"
                  >算法突破</span
                >
                和 <span class="highlight-text">职场进阶</span>。每一次练习，
                每一个反馈，都是成长路上的足迹。
              </p>
            </div> -->

            <!-- 屏蔽：特色标签区域 - 可恢复 -->
            <!-- <div class="feature-tags">
              <div class="tag-item frontend">
                <a href="http://wangzhongyang.com/" target="_blank" rel="noopener noreferrer">
                <span class="tag-icon">🔧</span>
                <span class="tag-text">就业陪跑</span>
              </a>
              </div>
              <div class="tag-item backend">
                <a href="https://mp.weixin.qq.com/s/5fxOGqtAQlFg0YDmoW2MXA" target="_blank" rel="noopener noreferrer">
                <span class="tag-icon">🔮</span>
                <span class="tag-text">知识星球</span>
              </a>
              </div>
              <div class="tag-item algorithm">
                <a href="https://mp.weixin.qq.com/s/tKbBrdwOzLCvQv_FqkVvQg" target="_blank" rel="noopener noreferrer">
                <span class="tag-icon">💡</span>
                <span class="tag-text">Go训练营</span>
              </a>
              </div>
              <div class="tag-item career">
                <a href="https://mp.weixin.qq.com/s/SP-f9HZBdEy6XdIJJN-W_A" target="_blank" rel="noopener noreferrer">
                <span class="tag-icon">🎯</span>
                <span class="tag-text">Java训练营</span>
              </a>
              </div>
              <div class="tag-item career">
                <a href="https://clin.pro/" target="_blank" rel="noopener noreferrer">
                  <span class="tag-icon">🎯</span>
                  <span class="tag-text">前端训练营</span>
                </a>
              </div>
            </div> -->
          </div>

          <!-- 主要行动区域 -->
          <div class="action-section">
            <div class="primary-actions">
              <div class="cta-wrapper">
                <div class="cta-glow"></div>
                <router-link class="primary-cta" to="/interview-master">
                  <div class="cta-content">
                    <span class="cta-icon">📖</span>
                    <span class="cta-text">开始体验</span>
                    <span class="cta-arrow">→</span>
                  </div>
                </router-link>
              </div>
            </div>

            <!-- 屏蔽：行动信息区域 - 可恢复 -->
            <!-- <div class="action-info">
              <div class="info-item">
                <a href="https://mp.weixin.qq.com/s/tKbBrdwOzLCvQv_FqkVvQg" target="_blank" rel="noopener noreferrer">
                  <span class="tag-icon">💡</span>
                  <span class="tag-text">Go训练营</span>
                </a>
              </div>
              <div class="info-item">
                <a href="http://www.bytemyth.cn/about-us.html" target="_blank" rel="noopener noreferrer">
                  <span class="tag-icon">💡</span>
                  <span class="tag-text">关于团队</span>
                </a>
              </div>
              <div class="info-item">
                <a href="https://mp.weixin.qq.com/s/5fxOGqtAQlFg0YDmoW2MXA" target="_blank" rel="noopener noreferrer">
                  <span class="tag-icon">💡</span>
                  <span class="tag-text">与我交流</span>
                </a>
              </div>
            </div> -->
          </div>
        </div>

        <!-- 右侧视觉演示 -->
        <div class="hero-visual">
          <div class="visual-container">
            <div class="code-window">
              <div class="window-header">
                <div class="window-controls">
                  <span class="control red"></span>
                  <span class="control yellow"></span>
                  <span class="control green"></span>
                </div>
                <span class="window-title">
                  <span class="title-icon">💻</span>
                  AI面试进行中...
                </span>
                <div class="window-status">
                  <span class="status-dot"></span>
                  <span class="status-text">正在评估</span>
                </div>
              </div>
              <div class="code-content">
                <div class="code-line">
                  <span class="line-number">1</span>
                  <span class="code-text">// 算法题：两数之和</span>
                </div>
                <div class="code-line">
                  <span class="line-number">2</span>
                  <span class="code-text">function twoSum(nums, target) {</span>
                </div>
                <div class="code-line">
                  <span class="line-number">3</span>
                  <span class="code-text"> const map = new Map();</span>
                </div>
                <div class="code-line active">
                  <span class="line-number">4</span>
                  <span class="code-text">
                    for (let i = 0; i < nums.length; i++) {</span
                  >
                  <span class="cursor"></span>
                </div>
                <div class="code-line">
                  <span class="line-number">5</span>
                  <span class="code-text">
                    const complement = target - nums[i];</span
                  >
                </div>
                <div class="code-line">
                  <span class="line-number">6</span>
                  <span class="code-text"> if (map.has(complement)) {</span>
                </div>
                <div class="code-line">
                  <span class="line-number">7</span>
                  <span class="code-text">
                    return [map.get(complement), i];</span
                  >
                </div>
                <div class="code-line">
                  <span class="line-number">8</span>
                  <span class="code-text"> }</span>
                </div>
                <div class="code-line">
                  <span class="line-number">9</span>
                  <span class="code-text"> map.set(nums[i], i);</span>
                </div>
                <div class="code-line">
                  <span class="line-number">10</span>
                  <span class="code-text"> }</span>
                </div>
                <div class="code-line">
                  <span class="line-number">11</span>
                  <span class="code-text">}</span>
                </div>
              </div>
            </div>

            <div class="ai-feedback">
              <div class="feedback-header">
                <span class="ai-avatar">🤖</span>
                <div class="feedback-info">
                  <span class="feedback-title">AI面试官</span>
                  <span class="feedback-subtitle">实时评估反馈</span>
                </div>
                <div class="feedback-score">
                  <span class="score">92</span>
                  <span class="score-label">分</span>
                </div>
              </div>
              <div class="feedback-content">
                <div class="feedback-item excellent">
                  <span class="feedback-icon">🎉</span>
                  <span class="feedback-text">算法思路清晰，逻辑正确</span>
                </div>
                <div class="feedback-item good">
                  <span class="feedback-icon">⚡</span>
                  <span class="feedback-text">时间复杂度：O(n)，表现优秀</span>
                </div>
                <div class="feedback-item tip">
                  <span class="feedback-icon">💡</span>
                  <span class="feedback-text">建议添加输入参数验证</span>
                </div>
                <div class="feedback-item insight">
                  <span class="feedback-icon">🔍</span>
                  <span class="feedback-text">可以考虑空间复杂度优化</span>
                </div>
              </div>
              <div class="feedback-action">
                <router-link class="feedback-btn" to="/interview-master">
                  立即体验完整面试
                </router-link>
              </div>
            </div>

            <!-- 数据统计卡片 -->
            <!-- <div class="stats-cards">
              <div class="stat-card">
                <div class="stat-icon">👥</div>
                <div class="stat-content">
                  <span class="stat-number">10k+</span>
                  <span class="stat-label">用户</span>
                </div>
              </div>
              <div class="stat-card">
                <div class="stat-icon">📈</div>
                <div class="stat-content">
                  <span class="stat-number">95%</span>
                  <span class="stat-label">提升率</span>
                </div>
              </div>
              <div class="stat-card">
                <div class="stat-icon">⭐</div>
                <div class="stat-content">
                  <span class="stat-number">4.9</span>
                  <span class="stat-label">评分</span>
                </div>
              </div>
            </div> -->
          </div>
        </div>
      </div>
    </main>

    <!-- 特性展示区域 -->
    <!-- <section class="features-section">
      <div class="features-container">
        <h2 class="features-title">为什么选择我们？</h2>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">🧠</span>
            </div>
            <h3 class="feature-title">智能算法面试</h3>
            <p class="feature-desc">
              涵盖数据结构、算法设计等核心知识点，难度分级，循序渐进
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">🏗️</span>
            </div>
            <h3 class="feature-title">系统设计模拟</h3>
            <p class="feature-desc">
              真实项目场景，从需求分析到架构设计，全方位能力提升
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">📈</span>
            </div>
            <h3 class="feature-title">个性化反馈</h3>
            <p class="feature-desc">
              AI深度分析你的答题过程，提供针对性建议和学习路径
            </p>
          </div>
        </div>
      </div>
    </section> -->

    <AppFooter />
  </div>
</template>

<script>
import AppFooter from "../components/AppFooter.vue";

export default {
  name: "Home",
  components: {
    AppFooter,
  },
};
</script>

<style scoped>
@import url("https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap");

.home-container {
  min-height: 100vh;
  background: var(--bg-primary);
  font-family: "Inter", sans-serif;
  position: relative;
  overflow-x: hidden;
}

/* 动态背景 */
.dynamic-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: -1;
  overflow: hidden;
}

.gradient-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.03) 0%,
    rgba(124, 58, 237, 0.05) 25%,
    rgba(16, 185, 129, 0.03) 50%,
    rgba(239, 68, 68, 0.02) 75%,
    rgba(245, 158, 11, 0.04) 100%
  );
  animation: gradientShift 15s ease-in-out infinite alternate;
}

@keyframes gradientShift {
  0% {
    background: linear-gradient(
      135deg,
      rgba(37, 99, 235, 0.03) 0%,
      rgba(124, 58, 237, 0.05) 25%,
      rgba(16, 185, 129, 0.03) 50%,
      rgba(239, 68, 68, 0.02) 75%,
      rgba(245, 158, 11, 0.04) 100%
    );
  }
  100% {
    background: linear-gradient(
      225deg,
      rgba(124, 58, 237, 0.04) 0%,
      rgba(16, 185, 129, 0.03) 25%,
      rgba(37, 99, 235, 0.05) 50%,
      rgba(245, 158, 11, 0.03) 75%,
      rgba(239, 68, 68, 0.02) 100%
    );
  }
}

.grid-pattern {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: linear-gradient(rgba(37, 99, 235, 0.1) 1px, transparent 1px),
    linear-gradient(90deg, rgba(37, 99, 235, 0.1) 1px, transparent 1px);
  background-size: 60px 60px;
  mask: radial-gradient(
    ellipse 80% 50% at 50% 0%,
    black 0%,
    black 50%,
    transparent 100%
  );
  animation: gridFloat 20s ease-in-out infinite;
}

@keyframes gridFloat {
  0%,
  100% {
    transform: translate(0, 0) rotate(0deg);
    opacity: 0.3;
  }
  50% {
    transform: translate(10px, -10px) rotate(1deg);
    opacity: 0.5;
  }
}

.floating-shapes {
  position: absolute;
  width: 100%;
  height: 100%;
}

.shape {
  position: absolute;
  border-radius: 50%;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.1),
    rgba(124, 58, 237, 0.1)
  );
  filter: blur(1px);
}

.shape-1 {
  width: 200px;
  height: 200px;
  top: 10%;
  right: 20%;
  animation: float1 25s ease-in-out infinite;
}

.shape-2 {
  width: 150px;
  height: 150px;
  bottom: 20%;
  left: 10%;
  animation: float2 20s ease-in-out infinite reverse;
}

.shape-3 {
  width: 100px;
  height: 100px;
  top: 50%;
  left: 5%;
  animation: float3 30s ease-in-out infinite;
}

.shape-4 {
  width: 120px;
  height: 120px;
  bottom: 10%;
  right: 5%;
  animation: float4 22s ease-in-out infinite reverse;
}

@keyframes float1 {
  0%,
  100% {
    transform: translate(0, 0) rotate(0deg);
    opacity: 0.4;
  }
  33% {
    transform: translate(30px, -20px) rotate(120deg);
    opacity: 0.7;
  }
  66% {
    transform: translate(-20px, 15px) rotate(240deg);
    opacity: 0.5;
  }
}

@keyframes float2 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
    opacity: 0.3;
  }
  50% {
    transform: translate(40px, -30px) scale(1.1);
    opacity: 0.6;
  }
}

@keyframes float3 {
  0%,
  100% {
    transform: translate(0, 0) rotate(0deg);
    opacity: 0.5;
  }
  25% {
    transform: translate(20px, -25px) rotate(90deg);
    opacity: 0.3;
  }
  75% {
    transform: translate(-15px, 20px) rotate(270deg);
    opacity: 0.7;
  }
}

@keyframes float4 {
  0%,
  100% {
    transform: translate(0, 0) rotate(0deg) scale(1);
    opacity: 0.4;
  }
  50% {
    transform: translate(-25px, -35px) rotate(180deg) scale(0.9);
    opacity: 0.6;
  }
}

/* 主要内容区域 */
.main-section {
  /* min-height: 100vh; */
  display: flex;
  align-items: center;
  padding: 120px 40px 80px;
  max-width: 1200px;
  margin: 0 auto;
}

.hero-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 80px;
  align-items: center;
  width: 100%;
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

/* 品牌介绍区域 */
.brand-intro {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.brand-title-wrapper {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.brand-title {
  font-size: 3.2rem;
  font-weight: 800;
  line-height: 1.1;
  margin: 0;
}

.title-main {
  display: block;
  font-size: 2.4rem;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.title-highlight {
  display: block;
  font-size: 3.2rem;
  background: linear-gradient(135deg, #2563eb 0%, #8b5cf6 100%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 8px;
}

.title-status {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.status-badge {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.3px;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}

.content-count {
  color: var(--text-secondary);
  font-size: 0.95rem;
}

.content-count strong {
  color: var(--primary-blue);
  font-weight: 700;
}

.brand-description {
  margin: 4px 0;
}

.main-desc {
  font-size: 1.1rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}

.highlight-text {
  color: var(--primary-blue);
  font-weight: 600;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.1),
    rgba(124, 58, 237, 0.1)
  );
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid rgba(37, 99, 235, 0.2);
}

/* 特色标签 */
.feature-tags {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.tag-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(37, 99, 235, 0.15);
  border-radius: 24px;
  padding: 10px 16px;
  transition: all 0.3s ease;
  cursor: pointer;
  backdrop-filter: blur(10px);
}

.tag-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.2);
  border-color: rgba(37, 99, 235, 0.3);
  background: rgba(255, 255, 255, 0.8);
}

.tag-item.frontend:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.2);
}

.tag-item.backend:hover {
  border-color: rgba(124, 58, 237, 0.4);
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.2);
}

.tag-item.algorithm:hover {
  border-color: rgba(245, 158, 11, 0.4);
  box-shadow: 0 4px 16px rgba(245, 158, 11, 0.2);
}

.tag-item.career:hover {
  border-color: rgba(16, 185, 129, 0.4);
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.2);
}

.tag-icon {
  font-size: 1.1rem;
}

.tag-text {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
  white-space: nowrap;
}

/* 价值主张区域 */
/* AI技术生态展示 */
.ai-ecosystem {
  display: flex;
  flex-direction: column;
  gap: 20px;
  position: relative;
}

.tech-showcase {
  background: linear-gradient(
    135deg,
    rgba(99, 102, 241, 0.05) 0%,
    rgba(139, 92, 246, 0.05) 100%
  );
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(99, 102, 241, 0.1);
  overflow: hidden;
}

.tech-scroll-container {
  height: 40px;
  overflow: hidden;
  position: relative;
  mask: linear-gradient(
    90deg,
    transparent 0%,
    black 10%,
    black 90%,
    transparent 100%
  );
}

.tech-scroll-track {
  display: flex;
  gap: 16px;
  animation: techScroll 20s linear infinite;
  white-space: nowrap;
}

@keyframes techScroll {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-50%);
  }
}

.tech-tag {
  font-family: "Inter", sans-serif;
  font-size: 0.85rem;
  font-weight: 600;
  color: #4338ca;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 20px;
  padding: 8px 16px;
  white-space: nowrap;
  display: inline-block;
  transition: all 0.3s ease;
  cursor: pointer;
}

.tech-tag:hover {
  background: rgba(99, 102, 241, 0.2);
  transform: scale(1.05);
  border-color: rgba(99, 102, 241, 0.4);
}

/* 平台优势展示 */
.platform-advantages {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 8px;
}

.advantage-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(37, 99, 235, 0.1);
  border-radius: 12px;
  transition: all 0.3s ease;
  cursor: pointer;
  backdrop-filter: blur(10px);
}

.advantage-item:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.7);
  border-color: rgba(37, 99, 235, 0.2);
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.1);
}

.advantage-icon {
  font-size: 1.5rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.1),
    rgba(124, 58, 237, 0.1)
  );
  border-radius: 10px;
  flex-shrink: 0;
  border: 1px solid rgba(37, 99, 235, 0.15);
}

.advantage-content {
  flex: 1;
}

.advantage-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 4px 0;
  line-height: 1.3;
}

.advantage-desc {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

.ai-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(99, 102, 241, 0.1);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
}

.status-dot.pulsing {
  animation: statusPulse 2s ease-in-out infinite;
}

@keyframes statusPulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.2);
  }
}

.status-text {
  font-size: 0.8rem;
  color: #374151;
  font-weight: 500;
}

.tech-metrics {
  display: flex;
  gap: 16px;
}

.metric {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  color: #6b7280;
  font-weight: 500;
}

.ai-capabilities-compact {
  margin-top: 8px;
}

.capability-pills {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.pill-item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(
    135deg,
    rgba(16, 185, 129, 0.1) 0%,
    rgba(59, 130, 246, 0.1) 100%
  );
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 24px;
  padding: 8px 16px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.pill-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.2);
  border-color: rgba(16, 185, 129, 0.3);
}

.pill-icon {
  font-size: 1rem;
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, #10b981, #3b82f6);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.pill-text {
  font-size: 0.85rem;
  color: #374151;
  font-weight: 500;
  white-space: nowrap;
}

/* 行动区域重组 */
.action-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
  align-items: flex-start;
}

.primary-actions {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.secondary-action {
  flex-shrink: 0;
}

.secondary-cta {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(37, 99, 235, 0.15);
  border-radius: 10px;
  padding: 12px 20px;
  color: var(--text-primary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
  font-size: 0.95rem;
}

.secondary-cta:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(37, 99, 235, 0.3);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15);
}

.secondary-icon {
  font-size: 1rem;
}

.secondary-text {
  font-weight: 500;
}

.secondary-arrow {
  font-weight: 600;
  color: var(--primary-blue);
  transition: transform 0.3s ease;
}

.secondary-cta:hover .secondary-arrow {
  transform: translateX(2px);
}

.action-info {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 4px 8px;
  border-radius: 6px;
}

.info-item:hover {
  color: var(--primary-blue);
  background: rgba(37, 99, 235, 0.05);
  transform: translateY(-1px);
}

.info-icon {
  font-size: 1rem;
}

/* 统计数据紧凑化 */
.stats-section {
  background: rgba(255, 255, 255, 0.4);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid rgba(37, 99, 235, 0.08);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 8px;
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-icon {
  font-size: 1.5rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(
    135deg,
    var(--primary-blue),
    var(--primary-purple)
  );
  border-radius: 8px;
  color: white;
  flex-shrink: 0;
}

.stat-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-number {
  font-family: "JetBrains Mono", monospace;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.stat-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.stat-trend {
  font-size: 0.75rem;
  color: var(--success-green);
  font-weight: 600;
}

/* 右侧视觉元素 */
.hero-visual {
  display: flex;
  justify-content: center;
  align-items: center;
}

.visual-container {
  position: relative;
  width: 100%;
  max-width: 480px;
}

.code-window {
  background: var(--bg-card);
  border-radius: 12px;
  box-shadow: var(--shadow-xl), 0 0 0 1px rgba(37, 99, 235, 0.1);
  border: 1px solid var(--border-light);
  overflow: hidden;
  margin-bottom: 20px;
  transition: all 0.3s ease;
}

.code-window:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-xl), 0 0 0 1px rgba(37, 99, 235, 0.2),
    0 0 20px rgba(37, 99, 235, 0.1);
}

.window-header {
  background: linear-gradient(135deg, #f6f8fa, #e7e9eb);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-light);
}

.window-controls {
  display: flex;
  gap: 6px;
}

.control {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.control:hover {
  transform: scale(1.1);
}

.control.red {
  background: linear-gradient(135deg, #ff5f57, #ff3b30);
  box-shadow: 0 0 10px rgba(255, 95, 87, 0.3);
}
.control.yellow {
  background: linear-gradient(135deg, #ffbd2e, #ff9500);
  box-shadow: 0 0 10px rgba(255, 189, 46, 0.3);
}
.control.green {
  background: linear-gradient(135deg, #28ca42, #34c759);
  box-shadow: 0 0 10px rgba(40, 202, 66, 0.3);
}

.window-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.title-icon {
  font-size: 1rem;
}

.window-status {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: var(--success-green);
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.1);
  }
}

.status-text {
  font-size: 0.75rem;
  color: var(--success-green);
  font-weight: 500;
}

.code-content {
  padding: 20px;
  background: linear-gradient(135deg, #1e293b, #334155);
  font-family: "JetBrains Mono", monospace;
  font-size: 0.9rem;
  line-height: 1.6;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 2px 0;
  position: relative;
  transition: all 0.2s ease;
}

.code-line:hover {
  background: rgba(59, 130, 246, 0.05);
  margin: 0 -20px;
  padding: 2px 20px;
}

.code-line.active {
  background: rgba(59, 130, 246, 0.1);
  margin: 0 -20px;
  padding: 2px 20px;
  border-left: 3px solid #3b82f6;
  padding-left: 17px;
}

.line-number {
  color: #64748b;
  width: 20px;
  text-align: right;
  user-select: none;
}

.code-text {
  color: #e2e8f0;
}

.code-line:nth-of-type(3) .code-text,
.code-line:nth-of-type(4) .code-text,
.code-line:nth-of-type(10) .code-text {
  padding-inline-start: 26px;
}

.code-line:nth-of-type(5) .code-text,
.code-line:nth-of-type(6) .code-text,
.code-line:nth-of-type(8) .code-text,
.code-line:nth-of-type(9) .code-text {
  padding-inline-start: 52px;
}

.code-line:nth-of-type(7) .code-text {
  padding-inline-start: 78px;
}

.cursor {
  width: 2px;
  height: 20px;
  background: #3b82f6;
  animation: blink 1s infinite;
  margin-left: 2px;
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.5);
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}

.ai-feedback {
  background: var(--bg-card);
  border-radius: 12px;
  box-shadow: var(--shadow-lg), 0 0 0 1px rgba(16, 185, 129, 0.1);
  border: 1px solid var(--border-light);
  padding: 20px;
  transition: all 0.3s ease;
}

.ai-feedback:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-xl), 0 0 0 1px rgba(16, 185, 129, 0.2),
    0 0 20px rgba(16, 185, 129, 0.1);
}

.feedback-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.feedback-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.feedback-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.feedback-subtitle {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.feedback-score {
  display: flex;
  align-items: baseline;
  gap: 2px;
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  padding: 6px 14px;
  border-radius: 20px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.score {
  font-size: 1.2rem;
  font-family: "JetBrains Mono", monospace;
}

.score-label {
  font-size: 0.8rem;
}

.feedback-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.feedback-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.feedback-item.excellent {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.feedback-item.good {
  background: rgba(37, 99, 235, 0.1);
  border: 1px solid rgba(37, 99, 235, 0.2);
}

.feedback-item.tip {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.feedback-item.insight {
  background: rgba(124, 58, 237, 0.1);
  border: 1px solid rgba(124, 58, 237, 0.2);
}

.feedback-item:hover {
  transform: translateX(3px);
}

.feedback-icon {
  font-size: 1rem;
}

.feedback-text {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.feedback-action {
  margin-top: 8px;
}

.feedback-btn {
  background: linear-gradient(
    135deg,
    var(--primary-blue),
    var(--primary-purple)
  );
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;
  display: block;
  text-align: center;
}

.feedback-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

/* 统计卡片 */
.stats-cards {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.stats-cards .stat-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(37, 99, 235, 0.1);
  border-radius: 12px;
  padding: 16px 12px;
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.stats-cards .stat-card:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.8);
  border-color: rgba(37, 99, 235, 0.2);
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.15);
}

.stats-cards .stat-icon {
  font-size: 1.5rem;
  width: 36px;
  height: 36px;
  background: linear-gradient(
    135deg,
    var(--primary-blue),
    var(--primary-purple)
  );
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.3);
}

.stats-cards .stat-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stats-cards .stat-number {
  font-family: "JetBrains Mono", monospace;
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.stats-cards .stat-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}

/* 屏蔽：注意事项区域样式 - 可恢复 */
/* .notice-section {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  padding: 40px 20px;
  position: relative;
  overflow: hidden;
}

.notice-container {
  max-width: 800px;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

.notice-card {
  background: rgba(255, 255, 255, 0.9);
  border: 2px solid #f59e0b;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 8px 32px rgba(245, 158, 11, 0.2);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.notice-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(245, 158, 11, 0.3);
}

.notice-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.notice-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  color: white;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.3);
}

.notice-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  background: linear-gradient(135deg, #92400e, #b45309);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.notice-content {
  margin: 0;
}

.notice-text {
  font-size: 1.1rem;
  color: var(--text-primary);
  line-height: 1.7;
  margin: 0;
  font-weight: 500;
  text-align: left;
} */

/* 屏蔽：注意事项响应式设计 - 可恢复 */
/* @media (max-width: 768px) {
  .notice-section {
    padding: 30px 15px;
  }
  
  .notice-card {
    padding: 24px;
  }
  
  .notice-header {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .notice-title {
    font-size: 1.3rem;
  }
  
  .notice-text {
    font-size: 1rem;
    text-align: center;
  }
} */

/* 特性展示区域 */
.features-section {
  background: linear-gradient(135deg, var(--bg-secondary), #f1f5f9);
  padding: 80px 20px;
  position: relative;
  overflow: hidden;
}

.features-section::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: radial-gradient(
      circle at 20% 20%,
      rgba(37, 99, 235, 0.05) 0%,
      transparent 50%
    ),
    radial-gradient(
      circle at 80% 80%,
      rgba(124, 58, 237, 0.05) 0%,
      transparent 50%
    );
  pointer-events: none;
}

.features-container {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
  position: relative;
  z-index: 1;
}

.features-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 60px;
  letter-spacing: -0.02em;
  position: relative;
}

.features-title::after {
  content: "";
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 4px;
  background: linear-gradient(
    135deg,
    var(--primary-blue),
    var(--primary-purple)
  );
  border-radius: 2px;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 40px;
}

.feature-card {
  background: var(--bg-card);
  padding: 40px 30px;
  border-radius: 16px;
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-light);
  text-align: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.feature-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.02),
    rgba(124, 58, 237, 0.02)
  );
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-xl), 0 0 0 1px rgba(37, 99, 235, 0.1);
}

.feature-card:hover::before {
  opacity: 1;
}

.feature-icon-wrapper {
  width: 80px;
  height: 80px;
  margin: 0 auto 24px;
  background: linear-gradient(135deg, #ddd6fe, #bfdbfe);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  position: relative;
  z-index: 1;
}

.feature-card:hover .feature-icon-wrapper {
  transform: scale(1.1) rotate(5deg);
  box-shadow: 0 10px 25px rgba(124, 58, 237, 0.2);
}

.feature-icon {
  font-size: 2.5rem;
  transition: all 0.3s ease;
}

.feature-card:hover .feature-icon {
  transform: scale(1.1);
}

.feature-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
  position: relative;
  z-index: 1;
}

.feature-desc {
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
  font-size: 1rem;
  position: relative;
  z-index: 1;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .hero-container {
    grid-template-columns: 1fr;
    gap: 40px;
    text-align: center;
  }

  .hero-visual {
    order: -1;
  }

  .feature-tags {
    justify-content: center;
  }

  .primary-actions {
    justify-content: center;
  }

  .action-info {
    justify-content: center;
  }

  .platform-advantages {
    max-width: 500px;
    margin: 0 auto;
  }
}

@media (max-width: 768px) {
  .main-section {
    padding: 100px 20px 60px;
  }

  .nav-logo {
    gap: 8px;
  }

  .nav-logo-icon {
    width: 32px;
    height: 32px;
  }

  .code-brackets {
    font-size: 1rem;
  }

  .nav-logo-text h3 {
    font-size: 1rem;
    color: #000;
  }

  .nav-logo-text p {
    font-size: 0.6rem;
  }

  .brand-title {
    font-size: 2.5rem;
  }

  .title-main {
    font-size: 2rem;
  }

  .title-highlight {
    font-size: 2.5rem;
  }

  .title-status {
    justify-content: center;
  }

  .feature-tags {
    gap: 8px;
  }

  .tag-item {
    padding: 8px 12px;
  }

  .tag-text {
    font-size: 0.8rem;
  }

  .primary-actions {
    flex-direction: column;
    gap: 12px;
    width: 100%;
  }

  .primary-cta {
    width: 100%;
    padding: 18px 32px;
  }

  .secondary-cta {
    width: 100%;
    justify-content: center;
  }

  .action-info {
    flex-direction: column;
    gap: 12px;
  }

  .platform-advantages {
    gap: 12px;
  }

  .advantage-item {
    padding: 12px;
  }

  .advantage-title {
    font-size: 0.9rem;
  }

  .advantage-desc {
    font-size: 0.8rem;
  }

  .stats-cards {
    gap: 8px;
  }

  .stats-cards .stat-card {
    padding: 12px 8px;
  }

  .stats-cards .stat-number {
    font-size: 1.1rem;
  }

  .stats-cards .stat-label {
    font-size: 0.75rem;
  }
}

/* Header样式 */
.main-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(-24px) saturate(150%) brightness(105%);
  -webkit-backdrop-filter: blur(-24px) saturate(150%) brightness(105%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 4px 32px rgba(0, 0, 0, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
  z-index: 1000;
  padding: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.main-header::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.25) 0%,
    rgba(248, 250, 252, 0.15) 50%,
    rgba(240, 245, 255, 0.2) 100%
  );
  backdrop-filter: blur(32px) saturate(180%);
  -webkit-backdrop-filter: blur(32px) saturate(180%);
  z-index: -1;
  border-radius: 0 0 0 0;
}

.main-header.scrolled {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(40px) saturate(200%) brightness(110%);
  -webkit-backdrop-filter: blur(40px) saturate(200%) brightness(110%);
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  border-bottom-color: rgba(255, 255, 255, 0.3);
}

.main-header.scrolled::before {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.3) 0%,
    rgba(248, 250, 252, 0.2) 50%,
    rgba(240, 245, 255, 0.25) 100%
  );
  backdrop-filter: blur(48px) saturate(220%);
  -webkit-backdrop-filter: blur(48px) saturate(220%);
}

.navbar {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  position: relative;
  z-index: 1;
}

.nav-brand {
  display: flex;
  align-items: center;
}

.nav-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.nav-logo:hover {
  transform: translateY(-1px);
}

.nav-logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #6366f1 0%, #0066ff 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
  position: relative;
  overflow: hidden;
}

.nav-logo-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

.nav-logo-icon::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.1) 0%,
    rgba(255, 255, 255, 0.05) 50%,
    rgba(255, 255, 255, 0.1) 100%
  );
  border-radius: 8px;
}

.code-brackets {
  font-family: "JetBrains Mono", monospace;
  font-size: 1.2rem;
  font-weight: 600;
  color: white;
  position: relative;
  z-index: 1;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.nav-logo-text h3 {
  font-family: "Inter", sans-serif;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 2px 0;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-blue));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1.2;
}

.nav-logo-text p {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.7rem;
  color: var(--text-secondary);
  margin: 0;
  letter-spacing: 0.5px;
  opacity: 0.8;
  line-height: 1;
}

.nav-actions {
  display: flex;
  gap: 12px;
}

.nav-btn {
  padding: 12px 24px;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: none;
  font-size: 0.95rem;
  letter-spacing: 0.3px;
  position: relative;
  overflow: hidden;
}

.nav-btn.primary {
  background: linear-gradient(
    135deg,
    #2563eb 0%,
    #3b82f6 25%,
    #6366f1 75%,
    #8b5cf6 100%
  );
  color: white;
  box-shadow: 0 4px 16px rgba(37, 99, 235, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.2);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  font-weight: 700;
}

.nav-btn.primary::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.1) 0%,
    rgba(255, 255, 255, 0.05) 50%,
    rgba(255, 255, 255, 0.1) 100%
  );
  z-index: 1;
  border-radius: 9px;
}

.nav-btn.primary span {
  position: relative;
  z-index: 2;
  color: white !important;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  font-weight: 700;
}

.nav-btn.primary:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 8px 24px rgba(37, 99, 235, 0.5),
    0 0 0 2px rgba(37, 99, 235, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.nav-btn.primary:active {
  transform: translateY(-1px) scale(1.01);
}

/* CTA按钮样式 */
.cta-wrapper {
  position: relative;
  display: inline-block;
}

.cta-glow {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  background: linear-gradient(
    135deg,
    rgba(37, 99, 235, 0.2) 0%,
    rgba(124, 58, 237, 0.2) 50%,
    rgba(16, 185, 129, 0.2) 100%
  );
  border-radius: 20px;
  z-index: -1;
  animation: ctaGlow 4s ease-in-out infinite;
  filter: blur(4px);
}

@keyframes ctaGlow {
  0%,
  100% {
    opacity: 0.3;
    transform: scale(0.98);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.02);
  }
}

.cta-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    var(--primary-blue),
    var(--primary-purple)
  );
  border-radius: 16px;
  transform: translate(-50%, -50%);
  z-index: -2;
  animation: ctaPulse 3s ease-in-out infinite;
  opacity: 0.15;
}

@keyframes ctaPulse {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.15;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.05);
    opacity: 0.25;
  }
  100% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.15;
  }
}

.primary-cta {
  background: linear-gradient(
    135deg,
    #2563eb 0%,
    #3b82f6 25%,
    #6366f1 75%,
    #8b5cf6 100%
  );
  color: white;
  border: none;
  padding: 20px 40px;
  border-radius: 12px;
  font-size: 1.2rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 8px 24px rgba(37, 99, 235, 0.3), 0 0 0 0 rgba(37, 99, 235, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  position: relative;
  overflow: hidden;
  min-width: 260px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  letter-spacing: 0.5px;
}

.primary-cta::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.08) 0%,
    rgba(255, 255, 255, 0.03) 50%,
    rgba(255, 255, 255, 0.08) 100%
  );
  z-index: 1;
  border-radius: 11px;
}

.cta-background {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  transition: left 0.8s ease;
  z-index: 2;
}

.cta-shine {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(
    45deg,
    transparent 45%,
    rgba(255, 255, 255, 0.15) 50%,
    transparent 55%
  );
  transform: rotate(45deg);
  z-index: 2;
  animation: shine 6s ease-in-out infinite;
}

@keyframes shine {
  0% {
    transform: translateX(-100%) translateY(-100%) rotate(45deg);
  }
  50% {
    transform: translateX(100%) translateY(100%) rotate(45deg);
  }
  100% {
    transform: translateX(-100%) translateY(-100%) rotate(45deg);
  }
}

.primary-cta:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 12px 32px rgba(37, 99, 235, 0.4),
    0 0 0 3px rgba(37, 99, 235, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.2);
}

.primary-cta:hover .cta-background {
  left: 100%;
}

.primary-cta:active {
  transform: translateY(-2px) scale(1.01);
}

.cta-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  position: relative;
  z-index: 3;
}

.cta-text {
  position: relative;
  z-index: 3;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.cta-icon {
  font-size: 1.2rem;
  transition: transform 0.3s ease;
  position: relative;
  z-index: 3;
}

.cta-arrow {
  font-size: 1.1rem;
  font-weight: 600;
  transition: transform 0.3s ease;
  position: relative;
  z-index: 3;
}

.primary-cta:hover .cta-arrow {
  transform: translateX(3px);
}

.free-badge {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.3px;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
  animation: badgePulse 4s ease-in-out infinite;
}

@keyframes badgePulse {
  0%,
  100% {
    transform: scale(1);
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
  }
  50% {
    transform: scale(1.02);
    box-shadow: 0 3px 10px rgba(16, 185, 129, 0.25);
  }
}
</style>
