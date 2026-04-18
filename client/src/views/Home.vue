<template>
  <div class="page visible" id="home">
    <nav class="nav">
      <a class="nav-brand" href="#">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><rect width="24" height="24" rx="5" fill="rgba(255,255,255,.08)" stroke="rgba(255,255,255,.12)" stroke-width="1"/><circle cx="12" cy="12" r="7" stroke="rgba(255,255,255,.45)" stroke-width="1"/><line x1="7" y1="10" x2="13" y2="10" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/><line x1="11" y1="14" x2="17" y2="14" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/></svg>
        AI 面试官
      </a>
      <ul class="nav-links">
        <li><a href="#">功能</a></li>
        <li><a href="#">工作原理</a></li>
        <li><a href="#">技术栈</a></li>
      </ul>
      <div class="nav-r">
        <!-- 保留原有ThemeToggle逻辑，这里暂时写死为文字或者用组件 -->
        <!-- 如果后续还需要浅色模式，可以将 ThemeToggle 放回来 -->
        <button class="btn-ng" @click="goToLogin">登录</button>
        <button class="btn-ns" @click="goToChat">开始体验</button>
      </div>
    </nav>

    <div class="hero-outer">
      <section class="hero">
        <div class="hero-left">
          <div class="eyebrow"><span class="edot"></span>GoZero · pgvector · gRPC · Redis</div>
          <h1 class="hero-title">
            真正会<br>追问的<br><span class="dim">AI 面试官</span>
          </h1>
          <p class="hero-sub">每一轮回答都触发深度追问。AI 感知你的知识边界，复现真实技术面试的节奏与压力。</p>
          <div class="hero-cta">
            <a class="btn-p" href="#" @click.prevent="goToChat">
              <span class="btn-ring"></span>
              立即开始体验
              <svg width="13" height="13" viewBox="0 0 13 13" fill="none"><path d="M2 6.5h9M8 4l3 2.5L8 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </a>
            <a class="btn-g" href="#">了解工作原理</a>
          </div>
          <div class="hero-meta">
            <span>Go 后端</span><span>·</span><span>系统设计</span><span>·</span><span>分布式</span><span>·</span><span>数据库</span>
          </div>
        </div>

        <div class="hero-right">
          <div class="vortex-wrap">
            <div class="vring"></div><div class="vring"></div><div class="vring"></div><div class="vring"></div><div class="vring"></div>
          </div>
          <div class="demo-win">
            <div class="win-bar">
              <span class="wd r"></span><span class="wd y"></span><span class="wd g"></span>
              <span class="win-lbl">AI Interview — Go 后端</span>
            </div>
            <div class="demo-msgs" id="demo-msgs" ref="demoMsgsRef">
            </div>
            <div class="demo-foot"><span class="live-dot"></span>实时追问中</div>
          </div>
        </div>
      </section>
    </div>

    <div class="metrics-outer">
      <div class="metrics">
        <div class="met"><div class="met-n">N+1</div><div class="met-l">层深度追问，不接受表面答案</div></div>
        <div class="met"><div class="met-n">RAG</div><div class="met-l">知识库增强，题库与技术同步更新</div></div>
        <div class="met"><div class="met-n">SSE</div><div class="met-l">实时流式输出，零等待感知</div></div>
        <div class="met"><div class="met-n">OSS</div><div class="met-l">完整开源，GoZero 微服务架构</div></div>
      </div>
    </div>

    <div class="bottom-phrase">
      不是题库，<span class="dim" style="color:rgba(255,255,255,0.7)">是真实的追问。</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useTheme } from "../composables/useTheme";

const router = useRouter();
const { theme, toggleTheme } = useTheme();

const demoMsgsRef = ref(null);
let demoTimer = null;
let typeTimer = null;

const goToChat = () => router.push("/chat");
const goToLogin = () => router.push("/login");

// --- Mock Terminal Animation Logic ---
const SCRIPT = [
  { r: 'ai',  name: 'AI 面试官',    t: '请解释 Go 的 goroutine 调度器是如何工作的？' },
  { r: 'usr', name: '你',           t: 'goroutine 采用 M:N 模型，由 Go runtime 调度，将 M 个 goroutine 映射到 N 个线程...' },
  { r: 'ai',  name: 'AI · 追问 #1', t: '好。当 goroutine 触发阻塞 syscall 时，调度器如何处理？' },
];

function appendMsg(cfg) {
  const container = demoMsgsRef.value;
  if (!container) return null;
  const div = document.createElement('div');
  div.className = `dmsg ${cfg.r}`;
  div.innerHTML = `<div class="dav">${cfg.r === 'ai' ? 'AI' : 'U'}</div>
    <div><div class="dlbl">${cfg.name}</div>
    <div class="dbbl"><span class="content"></span>${cfg.r === 'ai' ? '<span class="dcur"></span>' : ''}</div></div>`;
  container.appendChild(div);
  requestAnimationFrame(() => div.classList.add('show'));
  return div.querySelector('.content');
}

function typeText(el, text, cb) {
  let i = 0;
  typeTimer = setInterval(() => {
    el.textContent += text[i++];
    if (i >= text.length) {
      clearInterval(typeTimer);
      if (cb) cb();
    }
  }, 35); // Adjusted speed
}

function runDemo() {
  let si = 0;
  function runNext() {
    if (!demoMsgsRef.value) return;
    if (si >= SCRIPT.length) {
      si = 0;
      demoMsgsRef.value.innerHTML = '';
      demoTimer = setTimeout(runNext, 2000);
      return;
    }
    const cfg = SCRIPT[si++];
    const el = appendMsg(cfg);
    if (!el) return;
    const cur = el.parentElement.querySelector('.dcur');
    demoTimer = setTimeout(() => {
      typeText(el, cfg.t, () => {
        if (cur) cur.style.display = 'none';
        demoTimer = setTimeout(runNext, cfg.r === 'ai' ? 1200 : 800);
      });
    }, cfg.r === 'ai' ? 600 : 300);
  }
  demoTimer = setTimeout(runNext, 500);
}

onMounted(() => {
  runDemo();
});

onBeforeUnmount(() => {
  clearTimeout(demoTimer);
  clearInterval(typeTimer);
});

const startChat = () => {
  router.push({ name: "Chat" });
};
</script>

<style scoped>
/* Home Layer Setup */
#home {
  position: relative;
  z-index: 10;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Nav */
.nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 48px;
  height: 64px;
  border-bottom: 1px solid transparent;
  transition: all .3s;
}
.nav-brand {
  display: flex;
  align-items: center;
  gap: 9px;
  font: 600 15px var(--sans);
  color: var(--t);
  text-decoration: none;
}
.nav-links {
  display: flex;
  gap: 32px;
  list-style: none;
  margin: 0;
  padding: 0;
}
.nav-links a {
  font-size: 14px;
  color: var(--t2);
  text-decoration: none;
  transition: color .2s;
  opacity: .9;
}
.nav-links a:hover {
  color: var(--t);
  opacity: 1;
}
.nav-r {
  display: flex;
  gap: 12px;
}
.btn-ng {
  font: 14px var(--sans);
  color: var(--t);
  background: none;
  border: none;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 8px;
  transition: color .2s;
  opacity: .75;
}
.btn-ng:hover {
  opacity: 1;
}
.btn-ns {
  font: 600 14px var(--sans);
  color: var(--bg, #020204);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 8px 20px;
  border-radius: 8px;
  transition: opacity .2s;
}
.btn-ns:hover {
  opacity: .85;
}

/* Hero Section */
.hero-outer {
  padding: 0 44px;
  margin-top: 64px;
}
.hero {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1fr 1.6fr;
  grid-template-rows: 540px;
  align-items: center;
  gap: 64px;
  padding: 60px 0;
  max-width: 1320px;
  margin: 0 auto;
}
.hero-left {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
}

.hero-left::before {
  content: '';
  position: absolute;
  inset: -60px -40px -60px -120px;
  background: radial-gradient(ellipse at 20% 50%, rgba(4,4,6,.98) 0%, rgba(4,4,6,.9) 40%, rgba(4,4,6,.6) 65%, transparent 85%);
  pointer-events: none;
  z-index: -1;
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255,255,255,.14);
  border-radius: 100px;
  padding: 5px 14px;
  margin-bottom: 36px;
  letter-spacing: .04em;
  background: rgba(4,4,6,.6);
  white-space: nowrap;
  width: fit-content;
}

.edot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--t);
  animation: edot 2.6s ease-in-out infinite;
}

@keyframes edot {
  0%, 100% { opacity: 1; }
  50% { opacity: .22; }
}

.hero-title {
  font-size: clamp(46px, 5.5vw, 78px);
  font-weight: 900;
  font-family: var(--display);
  line-height: 1.08;
  letter-spacing: -.02em;
  margin-bottom: 22px;
  text-shadow: 0 0 0 transparent;
}

.hero-title .dim {
  color: rgba(220,220,212,.82);
}

.hero-sub {
  font-size: 15px;
  color: var(--t2);
  line-height: 1.75;
  margin-bottom: 40px;
  max-width: 440px;
}

.hero-cta {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-top: auto;
}

.hero-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 28px;
  border-top: 1px solid rgba(255,255,255,.07);
  overflow: hidden;
  flex-wrap: nowrap;
  margin-top: 32px;
}

.hero-meta span {
  font: 11px var(--mono);
  color: rgba(255,255,255,.4);
  letter-spacing: .08em;
  white-space: nowrap;
}

.btn-p {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font: 600 15px var(--sans);
  color: var(--bg, #020204);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 14px 28px;
  border-radius: 11px;
  text-decoration: none;
  transition: opacity .2s, transform .15s;
  box-shadow: 0 0 40px rgba(237,237,235,.1), 0 0 80px rgba(237,237,235,.04);
}

.btn-p:hover {
  opacity: .9;
  transform: translateY(-1px);
  box-shadow: 0 0 52px rgba(237,237,235,.16), 0 0 100px rgba(237,237,235,.07);
}

.btn-ring {
  position: absolute;
  inset: -6px;
  border: 1px solid rgba(237,237,235,.2);
  border-radius: 16px;
  pointer-events: none;
  animation: ring 3.6s ease-in-out infinite;
}

@keyframes ring {
  0%, 100% { transform: scale(1); opacity: .55; }
  50% { transform: scale(1.06); opacity: .08; }
}

.btn-g {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font: 500 15px var(--sans);
  color: var(--t2);
  background: none;
  border: 1px solid var(--b);
  cursor: pointer;
  padding: 13px 24px;
  border-radius: 10px;
  text-decoration: none;
  transition: all .2s;
}

.btn-g:hover {
  color: var(--t);
  border-color: var(--bh);
  background: rgba(255,255,255,0.05);
}

.hero-right {
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
}

.vortex-wrap {
  position: absolute;
  inset: 0;
  overflow: hidden;
  border-radius: 18px;
  pointer-events: none;
  z-index: 0;
}

.vring {
  position: absolute;
  top: 50%;
  left: 50%;
  border-radius: 50%;
  border: 1px solid rgba(237,237,235,.055);
  transform: translate(-50%,-50%);
  animation: vring-out 5s ease-out infinite;
}

.vring:nth-child(1) { width: 60px; height: 60px; animation-delay: 0s; }
.vring:nth-child(2) { width: 150px; height: 150px; animation-delay: 1s; }
.vring:nth-child(3) { width: 280px; height: 280px; animation-delay: 2s; }
.vring:nth-child(4) { width: 420px; height: 420px; animation-delay: 3s; }
.vring:nth-child(5) { width: 580px; height: 580px; animation-delay: 4s; }

@keyframes vring-out {
  0% { opacity: .55; transform: translate(-50%,-50%) scale(.2); }
  100% { opacity: 0; transform: translate(-50%,-50%) scale(1); }
}

.demo-win {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 440px;
  background: rgba(8,8,14,.85);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 16px;
  backdrop-filter: blur(20px);
  display: flex;
  flex-direction: column;
  box-shadow: 0 40px 100px rgba(0,0,0,.7), 0 0 0 1px rgba(255,255,255,.04) inset;
}

.win-bar {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--b);
  flex-shrink: 0;
  background: rgba(255,255,255,.02);
}

.wd {
  width: 11px;
  height: 11px;
  border-radius: 50%;
}
.wd.r { background: #ff5f56; }
.wd.y { background: #ffbd2e; }
.wd.g { background: #27c93f; }

.win-lbl {
  font: 13px var(--mono);
  color: var(--t2);
  margin: 0 auto;
  letter-spacing: .03em;
  opacity: .7;
}

.demo-msgs {
  flex: 1;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow: hidden;
}

:deep(.dmsg) {
  display: flex;
  gap: 12px;
  opacity: 0;
  transform: translateY(6px);
  transition: opacity .4s ease, transform .4s ease;
}

:deep(.dmsg.show) {
  opacity: 1;
  transform: none;
}

:deep(.dav) {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font: 700 11px var(--mono);
}

:deep(.dmsg.ai .dav) {
  background: var(--t);
  color: var(--bg, #020204);
}

:deep(.dmsg.usr .dav) {
  background: rgba(255,255,255,.06);
  border: 1px solid var(--b);
  color: var(--t3);
}

:deep(.dbbl) {
  border-radius: 10px;
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.65;
  max-width: 85%;
}

:deep(.dmsg.ai .dbbl) {
  background: rgba(255,255,255,.05);
  border: 1px solid rgba(255,255,255,.09);
  color: var(--t);
}

:deep(.dmsg.usr .dbbl) {
  background: rgba(255,255,255,.025);
  border: 1px solid rgba(255,255,255,.05);
  color: var(--t2);
}

:deep(.dlbl) {
  font: 11px var(--mono);
  color: rgba(255,255,255,.4);
  letter-spacing: .07em;
  text-transform: uppercase;
  margin-bottom: 6px;
}

:deep(.dcur) {
  display: inline-block;
  width: 2px;
  height: 13px;
  background: var(--t);
  vertical-align: middle;
  margin-left: 2px;
  animation: cb 1s step-end infinite;
}

@keyframes cb {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.demo-foot {
  padding: 12px 18px;
  border-top: 1px solid var(--b);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  color: rgba(255,255,255,.65);
}

.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4ade80;
  animation: edot 2s ease-in-out infinite;
}

.metrics-outer {
  position: relative;
  z-index: 1;
  padding: 0 44px;
  margin-top: -20px;
}

.metrics {
  max-width: 1320px;
  margin: 0 auto;
  display: flex;
  background: var(--bg3);
  border-radius: 48px;
  overflow: hidden;
  border: 2px solid rgba(255,255,255,.15);
  box-shadow: 0 0 0 2px rgba(255,255,255,.05) inset;
  position: relative;
}

.metrics::before, .metrics::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--sep);
  pointer-events: none;
}

.metrics::before { top: 0; }
.metrics::after { bottom: 0; }

.met {
  flex: 1;
  padding: 28px 0 28px 32px;
  border-right: 1px solid var(--b);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.met:first-child { padding-left: 32px; }
.met:last-child { border-right: none; }

.met-n {
  font: 700 32px var(--mono);
  color: var(--t);
  line-height: 1;
  letter-spacing: -.04em;
}

.met-l {
  font-size: 14px;
  color: var(--t2);
  margin-top: 2px;
}

.bottom-phrase {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: 80px 0;
  font-family: var(--display);
  font-size: clamp(32px, 4vw, 56px);
  font-weight: 900;
  letter-spacing: -.01em;
}

@media (max-width: 1024px) {
  .hero {
    grid-template-columns: 1fr;
    grid-template-rows: auto;
    text-align: center;
  }
  .hero-left::before {
    inset: -20px;
  }
  .eyebrow {
    margin: 0 auto 24px;
  }
  .hero-sub {
    margin: 0 auto 32px;
  }
  .hero-cta {
    justify-content: center;
  }
  .hero-meta {
    justify-content: center;
  }
  .metrics {
    flex-wrap: wrap;
    border-radius: 24px;
  }
  .met {
    min-width: 50%;
    border-bottom: 1px solid var(--b);
  }
  .met:nth-child(even) {
    border-right: none;
  }
}
</style>
