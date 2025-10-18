<template>
  <div class="page">
    <AuroraBackground />

    <SiteHeader>
      <template #actions>
        <ThemeToggle :theme="theme" @toggle="toggleTheme" />
        <!-- <router-link to="/login" class="nav-button">登录</router-link>
        <router-link to="/register" class="nav-button primary">注册</router-link> -->
      </template>
    </SiteHeader>

    <main class="main-content">
      <h1 class="hero-title">与真正的AI面试官对话</h1>
      <p class="hero-subtitle">
        深入代码级模拟面试，挖掘你的真实技术潜力。不再是题库，而是真实的思维碰撞。（项目会不断更新迭代，详细安排请见我的博客记录奥~）
      </p>

      <div class="product-showcase">
        <div class="window-header">
          <div class="window-dot dot-red"></div>
          <div class="window-dot dot-yellow"></div>
          <div class="window-dot dot-green"></div>
        </div>
        <pre class="code-area" v-pre><code><span class="token-comment">// AI: 你的解法很棒。现在，我们来聊聊它的空间复杂度。</span>
<span class="token-keyword">function</span> <span class="token-function">evaluate</span>(<span class="token-variable">code</span>) {
  <span class="token-keyword">const</span> <span class="token-variable">complexity</span> = <span class="token-function">analyze</span>(<span class="token-variable">code</span>);
  <span class="token-keyword">return</span> <span class="token-string">`空间复杂度为: ${complexity.space}`</span>;
}<span class="token-cursor"></span></code></pre>
      </div>

      <router-link to="/chat" class="cta-button">立即开始体验</router-link>
    </main>

    <section class="how-it-works main-content">
      <h2 class="section-title">工作原理</h2>
      <p class="section-description">
        只需三步，完成一次完整的模拟面试流程，获得真实且有深度的反馈。
      </p>
      <div class="steps-container">
        <div class="step-card">
          <div class="step-number">01</div>
          <h3 class="step-title">定制面试</h3>
          <p class="step-description">
            上传简历与JD，或选择技术方向，AI将为你生成高度定制化的面试流程。
          </p>
        </div>
        <div class="step-card">
          <div class="step-number">02</div>
          <h3 class="step-title">深度交互</h3>
          <p class="step-description">
            像与真人面试官一样进行对话、编写代码。AI会根据你的表现进行智能追问和引导。
          </p>
        </div>
        <div class="step-card">
          <div class="step-number">03</div>
          <h3 class="step-title">获取反馈</h3>
          <p class="step-description">
            面试结束后，你将获得一份全面的能力评估报告，包含知识点掌握情况和代码优化建议。
          </p>
        </div>
      </div>
    </section>

    <footer class="site-footer">
      <span>© 2025 AI 面试官</span>
      <div class="footer-links">
        <a href="http://aomanoh.com" target="_blank" rel="noopener noreferrer">我的个人博客</a>
      </div>
    </footer>
  </div>
</template>

<script setup>
import AuroraBackground from "../components/AuroraBackground.vue";
import SiteHeader from "../components/SiteHeader.vue";
import ThemeToggle from "../components/ThemeToggle.vue";
import { useTheme } from "../composables/useTheme";

const { theme, toggleTheme } = useTheme();
</script>

<style scoped>
.page {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  animation: fadeIn 0.5s ease-in-out;
}

.page > *:not(.aurora-background) {
  position: relative;
  z-index: 1;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.main-content {
  width: 100%;
  max-width: 900px;
  padding: 20px;
  text-align: center;
  margin: 5vh auto 0;
}

.hero-title {
  font-size: clamp(2.5rem, 5vw, 4rem);
  font-weight: 700;
  line-height: 1.1;
  margin-bottom: 24px;
  background: linear-gradient(90deg, #fff, #c0c0c0);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

body.light-mode .hero-title {
  background: linear-gradient(90deg, #111, #555);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-subtitle {
  font-size: 1.18rem;
  color: var(--color-text-secondary);
  max-width: 600px;
  margin: 0 auto 40px;
  line-height: 1.6;
}

.product-showcase {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  padding: 16px;
  text-align: left;
  margin-bottom: 50px;
  position: relative;
}

.product-showcase::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: 12px;
  padding: 1px;
  background: linear-gradient(135deg, var(--color-glow-1), var(--color-glow-2));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  opacity: 0.5;
  pointer-events: none;
}

.window-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-left: 4px;
}

.window-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.dot-red {
  background: #ff5f56;
}

.dot-yellow {
  background: #ffbd2e;
}

.dot-green {
  background: #27c93f;
}

.code-area {
  font-family: "JetBrains Mono", "SF Mono", "Fira Code", monospace;
  font-size: 0.9rem;
  line-height: 1.7;
  background: transparent;
  color: inherit;
  overflow: auto;
}

.token-comment {
  color: #6a9955;
}

.token-keyword {
  color: #569cd6;
}

.token-function {
  color: #dcdcaa;
}

.token-variable {
  color: #9cdcfe;
}

.token-string {
  color: #ce9178;
}

.token-cursor {
  display: inline-block;
  width: 8px;
  height: 1em;
  background: var(--color-text-secondary);
  animation: blink 1s step-end infinite;
  vertical-align: middle;
}

@keyframes blink {
  50% {
    opacity: 0;
  }
}

.cta-button {
  display: inline-block;
  padding: 14px 32px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(90deg, var(--color-glow-1), var(--color-glow-2));
  color: #fff;
  font-size: 1.1rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.cta-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 0 25px rgba(59, 130, 246, 0.5), 0 0 25px rgba(139, 92, 246, 0.5);
}


.how-it-works {
  padding: 80px 0;
}

.section-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 60px;
  text-align: center;
}

.section-description {
  text-align: center;
  max-width: 640px;
  margin: 0 auto 48px;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.steps-container {
  display: flex;
  justify-content: space-between;
  gap: 30px;
  flex-wrap: wrap;
}

.step-card {
  flex: 1;
  background: var(--color-input-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 30px;
  text-align: center;
  min-width: 250px;
}

.step-number {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-glow-1);
  margin-bottom: 16px;
}

.step-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 12px;
}

.step-description {
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.site-footer {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 30px 20px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  flex-shrink: 0;
}

.footer-links a {
  color: var(--color-text-secondary);
  text-decoration: none;
  margin-left: 20px;
  transition: color 0.2s ease;
}

.footer-links a:hover {
  color: var(--color-text-primary);
}

.nav-button {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 500;
  transition: background-color 0.2s;
  cursor: pointer;
}

body.light-mode .nav-button {
  background: rgba(0, 0, 0, 0.05);
}

.nav-button:hover {
  background: rgba(255, 255, 255, 0.2);
}

body.light-mode .nav-button:hover {
  background: rgba(0, 0, 0, 0.1);
}

.nav-button.primary {
  background: var(--color-glow-1) !important;
  border-color: transparent !important;
  color: #fff !important;
}

.nav-button.primary:hover {
  background: #4f8fff !important;
  color: #fff !important;
}

body.light-mode .nav-button.primary {
  background: var(--color-glow-1) !important;
  color: #fff !important;
}

body.light-mode .nav-button.primary:hover {
  background: #4f8fff !important;
  color: #fff !important;
}

@media (max-width: 768px) {
  .site-footer {
    flex-direction: column;
    gap: 12px;
  }

  .steps-container {
    flex-direction: column;
  }
}

@media (max-width: 480px) {
  .main-content {
    margin-top: 60px;
  }

  .hero-title {
    font-size: 2.5rem;
  }

  .hero-subtitle {
    font-size: 1rem;
  }

  .section-title {
    font-size: 2rem;
  }
}
</style>
