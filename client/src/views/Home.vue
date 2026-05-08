<template>
  <div class="page visible" id="home">
    <nav class="nav">
      <a class="nav-brand" href="#">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><rect width="24" height="24" rx="5" fill="rgba(255,255,255,.08)" stroke="rgba(255,255,255,.12)" stroke-width="1"/><circle cx="12" cy="12" r="7" stroke="rgba(255,255,255,.45)" stroke-width="1"/><line x1="7" y1="10" x2="13" y2="10" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/><line x1="11" y1="14" x2="17" y2="14" stroke="rgba(255,255,255,.9)" stroke-width="1.5" stroke-linecap="round"/></svg>
        AI 面试官
      </a>
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
          <div class="feat-pills">
            <span class="feat-pill">无限追问</span>
            <span class="feat-pill">真题召回</span>
            <span class="feat-pill">实时流式</span>
          </div>
          <div class="hero-cta">
            <a class="btn-p" href="#" @click.prevent="goToChat">
              <span class="btn-ring"></span>
              立即开始体验
              <svg width="13" height="13" viewBox="0 0 13 13" fill="none"><path d="M2 6.5h9M8 4l3 2.5L8 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </a>
            <a class="btn-g" href="#">了解工作原理</a>
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

    <div class="bottom-cta-outer">
      <div class="bottom-cta">
        <div class="bcta-l">
          <div class="bcta-eyebrow"><span class="edot"></span>准备好了吗？</div>
          <div class="bcta-text">现在就开始你的第一场 AI 面试</div>
        </div>
        <a class="btn-p" href="#" @click.prevent="goToChat">
          <span class="btn-ring"></span>
          开始体验
          <svg width="13" height="13" viewBox="0 0 13 13" fill="none"><path d="M2 6.5h9M8 4l3 2.5L8 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useApi } from "../composables/useApi";
import { useTheme } from "../composables/useTheme";

const router = useRouter();
const api = useApi();
const { theme, toggleTheme } = useTheme();

const demoMsgsRef = ref(null);
let demoTimer = null;
let typeRafId = null;
let activeScript = [];
let isUnmounted = false;

const goToChat = () => router.push("/chat");
const goToLogin = () => router.push("/login");

// --- Mock Terminal Animation Logic ---
const FALLBACK_SCRIPT = [
  { r: 'ai',  name: 'AI 面试官',    t: '请解释 Go 的 goroutine 调度器是如何工作的？' },
  { r: 'usr', name: '你',           t: 'goroutine 采用 M:N 模型，由 Go runtime 调度，将 M 个 goroutine 映射到 N 个线程...' },
  { r: 'ai',  name: 'AI · 追问 #1', t: '好。当 goroutine 触发阻塞 syscall 时，调度器如何处理？' },
];
activeScript = FALLBACK_SCRIPT;

const mapDemoSceneToScript = (scene) => {
  if (!scene?.available || !Array.isArray(scene.messages)) {
    return [];
  }

  return scene.messages
    .map((message) => {
      const content = String(message?.content || "").trim();
      if (!content) return null;

      const isUser = message?.role === "user";
      return {
        r: isUser ? "usr" : "ai",
        name: message?.name || (isUser ? "你" : "AI 面试官"),
        t: content,
      };
    })
    .filter(Boolean);
};

const loadDemoSceneScript = async () => {
  try {
    const scene = await api.user.demoInterviewSceneRandom({ limit: 3 });
    const nextScript = mapDemoSceneToScript(scene);
    if (!isUnmounted && nextScript.length >= 2) {
      activeScript = nextScript;
    }
  } catch (error) {
    if (!isUnmounted) {
      activeScript = FALLBACK_SCRIPT;
    }
  }
};

// 每个字符的最小渲染间隔（毫秒）。
// 60ms ≈ 16-17 字/秒，肉眼能清晰感知到"逐字"节奏；
// 低于 35ms 会被浏览器合并到同一帧 paint，视觉上变成"一坨一坨"出现。
const CHAR_DELAY_MS = 60;

function appendMsg(cfg) {
  const container = demoMsgsRef.value;
  if (!container) return null;
  const div = document.createElement('div');
  div.className = `dmsg ${cfg.r}`;
  // 关键结构：
  //   .dbody 是包裹 dlbl + dbbl 的列容器，必须明确 flex:1 1 auto 让它 stretch 到
  //   dmsg 剩余空间（固定宽度），否则它默认 hug-content，会让 .dbbl max-width 85%
  //   的百分比基数跟着 content 每字变化 → 换行位置每字符重算 → 视觉"上下乱吐"。
  //   .dbbl 自己保持 hug-content（inline-block），跟 content 一字字增长，
  //   触达固定 max-width 后自然换行，换行位置稳定。
  //   所有消息（含用户消息）都带打字光标，作为"逐字输入"的视觉锚点。
  div.innerHTML = `<div class="dav">${cfg.r === 'ai' ? 'AI' : 'U'}</div>
    <div class="dbody"><div class="dlbl">${cfg.name}</div>
    <div class="dbbl"><span class="content"></span><span class="dcur"></span></div></div>`;
  container.appendChild(div);
  requestAnimationFrame(() => div.classList.add('show'));
  return div.querySelector('.content');
}

// typeText 用 requestAnimationFrame + 时间戳累计来逐字吐字。
// 相比 setInterval：
// 1. RAF 与浏览器 vsync 同步，每帧最多吐 1 个字符，避免多字符在同一帧合并渲染；
// 2. 后台 tab 时 RAF 暂停，回到前台不会"补吐"一大段文字；
// 3. 时间戳节流让节奏不受单帧延迟影响，整体打字节奏稳定。
function typeText(el, text, cb) {
  let i = 0;
  let lastTs = 0;
  function step(ts) {
    if (!lastTs) lastTs = ts;
    if (ts - lastTs >= CHAR_DELAY_MS) {
      el.textContent += text[i++];
      lastTs = ts;
      if (i >= text.length) {
        typeRafId = null;
        if (cb) cb();
        return;
      }
    }
    typeRafId = requestAnimationFrame(step);
  }
  typeRafId = requestAnimationFrame(step);
}

function runDemo() {
  let si = 0;
  function runNext() {
    if (!demoMsgsRef.value) return;
    const script = activeScript.length ? activeScript : FALLBACK_SCRIPT;
    if (si >= script.length) {
      si = 0;
      demoMsgsRef.value.innerHTML = '';
      demoTimer = setTimeout(runNext, 2000);
      return;
    }
    const cfg = script[si++];
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
  loadDemoSceneScript();
});

onBeforeUnmount(() => {
  isUnmounted = true;
  clearTimeout(demoTimer);
  if (typeRafId !== null) cancelAnimationFrame(typeRafId);
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
  border-radius: var(--radius-sm);
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
  border-radius: var(--radius-sm);
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
  /* 列比 1.05:1 让文字主角回归：原 1:1.6 下左列只有 ~360px，三行汉字
     52px 字号被挤碎、视觉重心被推到右侧 demo。调为 1.05:1 后左列
     ~446px、右列 ~426px，两侧视觉对等但左侧略主导，文字主角回归。 */
  grid-template-columns: 1.05fr 1fr;
  /* hero 高度由 content 撑出，标题缩小后不再强制 540px 留白；
     min 480 让右侧 demo-win 区域有保底高度。 */
  grid-template-rows: minmax(480px, auto);
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
  border-radius: var(--radius-pill);
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
  /* 字号克制：避免大屏 78px × 三行汉字带来的视觉压迫感，
     收到 52px 上限把视觉重心留给右侧假终端 demo。 */
  font-size: clamp(36px, 3.6vw, 52px);
  font-weight: 900;
  font-family: var(--display);
  line-height: 1.12;
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
  border-radius: var(--radius-md);
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
  border-radius: 10px;
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
  border-radius: var(--radius-md);
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
  border-radius: var(--radius-lg);
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
  border-radius: var(--radius-lg);
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

:deep(.dbody) {
  /* stretch 到 dmsg 剩余空间（固定宽度），切断 dbbl max-width 85% 百分比
     与 content max-content 之间的循环依赖，这是换行位置稳定的前提。
     min-width: 0 让它可以缩小到不被 content 撑破 flex 容器。 */
  flex: 1 1 auto;
  min-width: 0;
}

:deep(.dbbl) {
  /* inline-block 让气泡自己 hug-content 随字增长（视觉上边吐边长）；
     父 .dbody 已 stretch，所以 max-width: 85% 的基数是固定的，
     气泡触达这个固定阈值才换行，换行点永远稳定。 */
  display: inline-block;
  vertical-align: top;
  border-radius: var(--radius-md);
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.65;
  max-width: 85%;
  /* 只在长英文 token 撑破边界时才断词，不用 overflow-wrap:anywhere，
     避免每个字符位置都成为潜在断点，减少 word-break 重算时的跳变。 */
  word-break: break-word;
  /* text-wrap: stable 是现代浏览器的稳定换行算法：
     逐字增量时优先保留既有换行位置，只对新增内容做局部换行决策。
     Chrome 120+ / Safari 17.4+ 原生支持；老浏览器忽略不影响。 */
  text-wrap: stable;
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
  border-radius: 24px;
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

/* Hero 左侧产品能力 pills：与 eyebrow 同视觉语言（mono / pill / 半透明边框），
   填补标题缩小后 hero 左侧的视觉空旷。 */
.feat-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 32px;
}

.feat-pill {
  font: 12px var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255,255,255,.12);
  border-radius: var(--radius-pill);
  padding: 5px 12px;
  letter-spacing: .04em;
  background: rgba(255,255,255,.025);
}

/* Metrics 后的底部 CTA 条：坐何页面收束、避免 #home min-height:100vh
   在较高屏上让 metrics 之后出现大片黑空。与 metrics 同等
   max-width 让视觉对位。 */
.bottom-cta-outer {
  position: relative;
  z-index: 1;
  padding: 0 44px;
  margin-top: 64px;
  margin-bottom: 80px;
}

.bottom-cta {
  max-width: 1320px;
  margin: 0 auto;
  padding: 36px 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
  border: 1px solid rgba(255,255,255,.08);
  border-radius: var(--radius-lg);
  background: linear-gradient(135deg, rgba(255,255,255,.025) 0%, rgba(255,255,255,.01) 100%);
}

.bcta-l {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bcta-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 12px var(--mono);
  color: var(--t2);
  letter-spacing: .04em;
  width: fit-content;
}

.bcta-text {
  font-family: var(--display);
  font-size: clamp(20px, 2.2vw, 30px);
  font-weight: 700;
  letter-spacing: -.01em;
  color: var(--t);
  line-height: 1.2;
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
  .bottom-cta {
    flex-direction: column;
    align-items: stretch;
    gap: 20px;
    padding: 28px 24px;
  }
  .bcta-l {
    align-items: center;
    text-align: center;
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
  .metrics {
    flex-wrap: wrap;
    border-radius: var(--radius-lg);
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
