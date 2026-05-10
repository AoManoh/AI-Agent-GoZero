<template>
  <div class="page visible" id="home">
    <!-- Aurora 背景层：mix-blend-mode: screen 让黑色透明，只把冷色光晕叠加到星空上 -->
    <div class="aurora-bg" aria-hidden="true">
      <div class="aurora-blob aurora-a"></div>
      <div class="aurora-blob aurora-b"></div>
    </div>
    <SiteHeader :fixed="true">
      <template #actions>
        <button v-if="isAuthenticated" class="btn-ng" @click="goToWorkbench">工作台</button>
        <button v-else class="btn-ng" @click="goToLogin">登录</button>
        <button v-if="isAuthenticated" class="btn-ns" @click="handleLogout">退出</button>
        <button v-else class="btn-ns" @click="goToChat">开始体验</button>
      </template>
    </SiteHeader>

    <div class="hero-outer">
      <section class="hero">
        <div class="hero-left">
          <div class="eyebrow"><span class="edot"></span>GoZero · pgvector · gRPC · Redis</div>
          <h1 class="hero-title">
            真正会追问的<br><span class="dim">AI 面试官</span>
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
              <span class="win-lbl">{{ demoSceneTitle }}</span>
            </div>
            <div class="demo-msgs" id="demo-msgs" ref="demoMsgsRef">
            </div>
            <div class="demo-foot">
              <span class="live-dot"></span>
              <span class="demo-foot-text">{{ demoSceneFoot }}</span>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 4 步流程区块：如何工作 → 上传简历 → 选择方向 → 开始面试 → 查看评估 -->
    <div class="steps-outer">
      <div class="steps-header">
        <h2 class="steps-title">四步开启你的第一场面试</h2>
      </div>
      <div class="steps">
        <div class="step-card">
          <div class="step-num">01</div>
          <div class="step-icon">
            <svg width="44" height="44" viewBox="0 0 44 44" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
              <rect x="11" y="7" width="22" height="30" rx="2"/>
              <line x1="16" y1="15" x2="28" y2="15"/>
              <line x1="16" y1="20" x2="28" y2="20"/>
              <line x1="16" y1="25" x2="24" y2="25"/>
            </svg>
          </div>
          <h3 class="step-title">上传简历</h3>
          <p class="step-desc">上传你的简历，AI 将自动提取关键信息，生成面试档案。</p>
        </div>
        <div class="step-arrow">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="13 6 19 12 13 18"/>
          </svg>
        </div>
        <div class="step-card">
          <div class="step-num">02</div>
          <div class="step-icon">
            <svg width="44" height="44" viewBox="0 0 44 44" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="9" width="11" height="11" rx="1.5"/>
              <rect x="24" y="9" width="11" height="11" rx="1.5"/>
              <rect x="9" y="24" width="11" height="11" rx="1.5"/>
              <path d="M24 30 l4 4 l7 -8"/>
            </svg>
          </div>
          <h3 class="step-title">选择方向</h3>
          <p class="step-desc">选择职位方向与面试类型，定制适合你的面试内容。</p>
        </div>
        <div class="step-arrow">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="13 6 19 12 13 18"/>
          </svg>
        </div>
        <div class="step-card">
          <div class="step-num">03</div>
          <div class="step-icon">
            <svg width="44" height="44" viewBox="0 0 44 44" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
              <path d="M8 12 a3 3 0 0 1 3 -3 h12 a3 3 0 0 1 3 3 v8 a3 3 0 0 1 -3 3 h-6 l-5 4 v-4 h-1 a3 3 0 0 1 -3 -3 z"/>
              <path d="M20 21 v2 a3 3 0 0 0 3 3 h6 l5 4 v-4 h1 a3 3 0 0 0 3 -3 v-8 a3 3 0 0 0 -3 -3 h-3"/>
            </svg>
          </div>
          <h3 class="step-title">开始面试</h3>
          <p class="step-desc">与 AI 面试官实时对话，获得沉浸式的面试体验。</p>
        </div>
        <div class="step-arrow">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="13 6 19 12 13 18"/>
          </svg>
        </div>
        <div class="step-card">
          <div class="step-num">04</div>
          <div class="step-icon">
            <svg width="44" height="44" viewBox="0 0 44 44" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="9" width="26" height="26" rx="2"/>
              <line x1="14" y1="29" x2="14" y2="22"/>
              <line x1="20" y1="29" x2="20" y2="17"/>
              <line x1="26" y1="29" x2="26" y2="14"/>
              <circle cx="32" cy="14" r="2"/>
            </svg>
          </div>
          <h3 class="step-title">查看评估</h3>
          <p class="step-desc">查看 AI 生成的面试评估报告，了解优势与改进建议。</p>
        </div>
      </div>
    </div>

    <!-- 技术亮点 + 立即开始 CTA（合并版）：原 metrics 4 个技术 keyword + 原 bottom-cta 文案+按钮 -->
    <div class="metrics-outer">
      <div class="metrics">
        <div class="metrics-row">
          <div class="met"><div class="met-n">N+1</div><div class="met-l">层深度追问，不接受表面答案</div></div>
          <div class="met"><div class="met-n">RAG</div><div class="met-l">知识库增强，题库与技术同步更新</div></div>
          <div class="met"><div class="met-n">SSE</div><div class="met-l">实时流式输出，零等待感知</div></div>
          <div class="met"><div class="met-n">OSS</div><div class="met-l">完整开源，GoZero 微服务架构</div></div>
        </div>
        <div class="metrics-cta">
          <div class="metrics-cta-l">
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

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useApi } from "../composables/useApi";
import { useTheme } from "../composables/useTheme";
import { useAuth } from "../composables/useAuth";
import SiteHeader from "../components/SiteHeader.vue";
import AppFooter from "../components/AppFooter.vue";

const router = useRouter();
const api = useApi();
const { theme, toggleTheme } = useTheme();
const { isAuthenticated, logout } = useAuth();

const demoMsgsRef = ref(null);
const demoSceneTitle = ref("AI Interview — Go 后端");
const demoSceneFoot = ref("正在抽取演示片段");
let demoTimer = null;
let typeRafId = null;
let activeScript = [];
let isUnmounted = false;

const goToChat = () => {
  if (isAuthenticated.value) {
    router.push("/workbench/new");
    return;
  }
  router.push({ path: "/login", query: { redirect: "/workbench/new" } });
};
const goToLogin = () => router.push("/login");
const goToWorkbench = () => router.push("/workbench");
const handleLogout = async () => {
  await logout();
  router.push("/");
};

// --- Demo Terminal Animation Logic ---
const FALLBACK_SCRIPT = [
  { r: 'ai',  name: 'AI 面试官',    t: '请解释 Go 的 goroutine 调度器是如何工作的？' },
  { r: 'usr', name: '你',           t: 'goroutine 采用 M:N 模型，由 Go runtime 调度，将 M 个 goroutine 映射到 N 个线程...' },
  { r: 'ai',  name: 'AI · 追问 #1', t: '好。当 goroutine 触发阻塞 syscall 时，调度器如何处理？' },
];
activeScript = FALLBACK_SCRIPT;

const FALLBACK_SCENE_META = {
  title: "AI Interview — Go 后端",
  foot: "本地兜底 · Go 后端 · 中级 · 资深技术官 · N+3",
};

const normalizeDemoFact = (value) => String(value || "").trim();

const setFallbackDemoScene = () => {
  activeScript = FALLBACK_SCRIPT;
  demoSceneTitle.value = FALLBACK_SCENE_META.title;
  demoSceneFoot.value = FALLBACK_SCENE_META.foot;
};

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

const applyDemoSceneFacts = (scene) => {
  const title = normalizeDemoFact(scene?.title) || normalizeDemoFact(scene?.directionLabel) || "Go 后端";
  const sourceLabel = normalizeDemoFact(scene?.sourceLabel) || "真实演示";
  const facts = [
    normalizeDemoFact(scene?.directionLabel),
    normalizeDemoFact(scene?.difficultyLabel),
    normalizeDemoFact(scene?.interviewerStyleLabel),
    normalizeDemoFact(scene?.followUpDepth),
  ].filter(Boolean);

  demoSceneTitle.value = `AI Interview — ${title}`;
  demoSceneFoot.value = facts.length ? `${sourceLabel} · ${facts.join(" · ")}` : `${sourceLabel} · 实时追问中`;
};

const loadDemoSceneScript = async () => {
  try {
    const scene = await api.user.demoInterviewSceneRandom({ limit: 3 });
    const nextScript = mapDemoSceneToScript(scene);
    if (!isUnmounted && nextScript.length >= 2) {
      activeScript = nextScript;
      applyDemoSceneFacts(scene);
      return;
    }
  } catch (error) {
    // 首页演示不能被接口失败阻断；真实数据不可用时播放本地兜底脚本。
  }
  if (!isUnmounted) setFallbackDemoScene();
};

// 每个字符的最小渲染间隔（毫秒）。
// 60ms ≈ 16-17 字/秒，肉眼能清晰感知到"逐字"节奏；
// 低于 35ms 会被浏览器合并到同一帧 paint，视觉上变成"一坨一坨"出现。
const CHAR_DELAY_MS = 60;

function appendMsg(cfg) {
  const container = demoMsgsRef.value;
  if (!container) return null;
  const role = cfg.r === "usr" ? "usr" : "ai";
  const div = document.createElement('div');
  div.className = `dmsg ${role}`;
  // 关键结构：
  //   .dbody 是包裹 dlbl + dbbl 的列容器，必须明确 flex:1 1 auto 让它 stretch 到
  //   dmsg 剩余空间（固定宽度），否则它默认 hug-content，会让 .dbbl max-width 85%
  //   的百分比基数跟着 content 每字变化 → 换行位置每字符重算 → 视觉"上下乱吐"。
  //   .dbbl 自己保持 hug-content（inline-block），跟 content 一字字增长，
  //   触达固定 max-width 后自然换行，换行位置稳定。
  //   所有消息（含用户消息）都带打字光标，作为"逐字输入"的视觉锚点。
  const avatar = document.createElement("div");
  avatar.className = "dav";
  avatar.textContent = role === "ai" ? "AI" : "U";

  const body = document.createElement("div");
  body.className = "dbody";

  const label = document.createElement("div");
  label.className = "dlbl";
  label.textContent = normalizeDemoFact(cfg.name) || (role === "ai" ? "AI 面试官" : "你");

  const bubble = document.createElement("div");
  bubble.className = "dbbl";

  const content = document.createElement("span");
  content.className = "content";

  const cursor = document.createElement("span");
  cursor.className = "dcur";

  bubble.append(content, cursor);
  body.append(label, bubble);
  div.append(avatar, body);
  container.appendChild(div);
  requestAnimationFrame(() => div.classList.add('show'));
  return content;
}

// typeText 用 requestAnimationFrame + 时间戳累计来逐字吐字。
// 相比 setInterval：
// 1. RAF 与浏览器 vsync 同步，每帧最多吐 1 个字符，避免多字符在同一帧合并渲染；
// 2. 后台 tab 时 RAF 暂停，回到前台不会"补吐"一大段文字；
// 3. 时间戳节流让节奏不受单帧延迟影响，整体打字节奏稳定。
function typeText(el, text, cb) {
  const content = String(text || "");
  if (!content) {
    if (cb) cb();
    return;
  }
  let i = 0;
  let lastTs = 0;
  function step(ts) {
    if (isUnmounted) return;
    if (!lastTs) lastTs = ts;
    if (ts - lastTs >= CHAR_DELAY_MS) {
      el.textContent += content[i++];
      lastTs = ts;
      if (i >= content.length) {
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
    if (isUnmounted || !demoMsgsRef.value) return;
    const script = activeScript.length ? activeScript : FALLBACK_SCRIPT;
    if (si >= script.length) {
      si = 0;
      demoMsgsRef.value.replaceChildren();
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

onMounted(async () => {
  await loadDemoSceneScript();
  if (isUnmounted) return;
  runDemo();
});

onBeforeUnmount(() => {
  isUnmounted = true;
  clearTimeout(demoTimer);
  if (typeRafId !== null) cancelAnimationFrame(typeRafId);
});
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

/* Aurora 背景层：在星空和 Home 内容之间叠一层冷色光晕。
   关键技巧：mix-blend-mode: screen 让 blob 容器的"黑"被透明化，
   只把彩色叠加上去，星空和 ripple 完全不被遮挡。
   z-index: -1 让它在 #home 自身 stacking context 里垫底，
   仍处于 #home (z-index 10) 之上、CosmicCanvas (z-index 0) 与
   ripple overlay (z-index 20) 的合成栈中。 */
.aurora-bg {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  overflow: hidden;
  mix-blend-mode: screen;
}

.aurora-blob {
  position: absolute;
  border-radius: 50%;
  /* 重度 blur 让 blob 边缘极柔，肉眼看不到圆形边界，
     只剩一片"雾感"色彩。 */
  filter: blur(90px);
  will-change: transform;
}

/* 左上：主调冷紫，超大尺寸 + 重度 blur，让色彩屁能蔓到半屏。
   alpha 压低到 0.18：在 mix-blend-mode: screen 下，实际叠加到
   星空上的亮度增量几乎肉眼不可辨，但多帧叠加后形成
   "屏幕被冷光染过一点"的气质，避免上一版的 "AI splash 色块"。 */
.aurora-blob.aurora-a {
  top: -25vw;
  left: -15vw;
  width: 80vw;
  height: 80vw;
  /* 冷月光近白，alpha 压到 0.07：screen 模式下混合后背景只微微发亮，
     不会让整体“发灰”。 */
  background: rgba(225, 230, 245, 0.07);
  animation: aurora-drift-a 80s ease-in-out infinite;
}

/* 右下：同样冷调但偏青，与 左上冷紫 同属一个色系，
   避免上一版的 "紫/青/琥珀 三色花斑"。 */
.aurora-blob.aurora-b {
  bottom: -20vw;
  right: -15vw;
  width: 70vw;
  height: 70vw;
  /* 暖琥珀近白，alpha 压到 0.05，同样避免背景“发灰”。 */
  background: rgba(245, 230, 210, 0.05);
  animation: aurora-drift-b 95s ease-in-out infinite;
}

/* 两个 blob 永不同步漂移：振幅 4~5vw/vh（上版 6~8 太大），
   周期错开到8 0s / 95s，最大公约数 8min，人眼肉眼看不出循环，
   只会觉得 "雾在隆隆地动"。 */
@keyframes aurora-drift-a {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(5vw, 4vh); }
}
@keyframes aurora-drift-b {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-4vw, -5vh); }
}

/* Nav 按钮样式（内嵌 .nav / .nav-brand / .nav-r 已提到 SiteHeader 组件；
   .btn-ng / .btn-ns 保留为 Home 私有，通过 SiteHeader actions slot 传入，
   Vue scoped CSS 下 slot 内容仍属父组件作用域，样式仍能作用于当前 button。 */
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
  /* 与 SiteHeader fixed 80px 高度配合，避免 hero 内容被覆盖 */
  margin-top: 80px;
}
.hero {
  position: relative;
  z-index: 1;
  display: grid;
  /* 画面 P：列比 2:3 让演示窗成为右半主角。1320px 容器、44px 左右 padding、
     64px gap 下，左列 ~467px、右列 ~701px；与上一轮 .demo-win height:100%
     min-height:480px 叠加后，演示窗约 1.46:1，接近 Mac 半屏工作站窗。
     左侧 hero-title 在 clamp(46px,4.4vw,66px) 下仍能容纳"真正会追问的"
     6 字单行，视觉重心略右倾但左侧叙事仍主导。 */
  grid-template-columns: 2fr 3fr;
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

.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font: 13px var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-pill);
  padding: 6px 16px;
  margin-bottom: 40px;
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.025);
  backdrop-filter: blur(8px);
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
  /* 等比放大：46~66px，有仪式感。 */
  font-size: clamp(46px, 4.4vw, 66px);
  font-weight: 900;
  font-family: var(--display);
  line-height: 1.12;
  letter-spacing: -.02em;
  margin-bottom: 26px;
  position: relative;
  /* 描边加强（上版 0.6px / 0.22 alpha 太弱看不见）：
     1px width【｜】0.5 alpha = 字形外缘有清晰白细线，立体感明显。 */
  -webkit-text-stroke: 1px rgba(255, 255, 255, 0.5);
  /* Shimmer 改 "激光扫光" 模式：
     - 基础色统一 0.35 alpha 浅灰（上版 0.62~0.82 渐变太柔看不出对比）
     - 中央 48%~52% 是纯白窄亮带（4% 宽，上版 30%~70% 太宽看不出扫过）
     - 38%~48% 和 52%~62% 是过渡段：羽化亮带边缘，避免锯齿
     - 周期 4s（更快）：扫过更频繁。视觉感受：一道激光每 4s 划过文字。 */
  background: linear-gradient(
    100deg,
    rgba(180, 180, 200, 0.35) 0%,
    rgba(180, 180, 200, 0.35) 38%,
    rgba(255, 255, 255, 1) 48%,
    rgba(255, 255, 255, 1) 52%,
    rgba(180, 180, 200, 0.35) 62%,
    rgba(180, 180, 200, 0.35) 100%
  );
  background-size: 200% 100%;
  background-position: 200% 0;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: title-shimmer 4s linear infinite;
  /* 双层 drop-shadow：
     - 0 2px 4px 黑色：让字“凸起”于背景，立体感更强
     - 0 0 20px 蓝白：远距漫射光晕。 */
  filter:
    drop-shadow(0 2px 4px rgba(0, 0, 0, 0.4))
    drop-shadow(0 0 20px rgba(220, 230, 255, 0.12));
}

@keyframes title-shimmer {
  /* 5s 一轮永不停循环：200% → -100% 匀速扫。
     循环点从 -100% 跳回 200% 时高光却在文字外，
     肉眼看不到跳变。 */
  0% { background-position: 200% 0; }
  100% { background-position: -100% 0; }
}

.hero-title .dim {
  /* 不继承父级 transparent fill；同时取消描边让暗色文字保持简洁。 */
  -webkit-text-fill-color: rgba(220, 220, 212, 0.55);
  color: rgba(220, 220, 212, 0.55);
  -webkit-text-stroke: 0;
}

.hero-sub {
  font-size: 17px;
  color: var(--t2);
  line-height: 1.75;
  margin-bottom: 44px;
  max-width: 480px;
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
  gap: 9px;
  font: 600 17px var(--sans);
  color: var(--bg, #020204);
  background: var(--t);
  border: none;
  cursor: pointer;
  padding: 16px 34px;
  border-radius: var(--radius-md);
  text-decoration: none;
  overflow: hidden;
  transition: transform .25s ease, opacity .25s ease;
  /* 持续呼吸的中性暖白光晕（与 aurora 的冷暖调同），每 6s 一次。 */
  animation: btn-breathing 6s ease-in-out infinite;
}

/* 内部持续扫光：::before 是一道半透明白带，从左到右无限循环慢扫。
   叠在白底上让按钮中段“持续微亮”，像有一束光在按钮表面流过。
   速度 4s/cycle，自然不刺眼。 */
.btn-p::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: linear-gradient(
    110deg,
    transparent 0%,
    transparent 40%,
    rgba(255, 255, 255, 0.5) 50%,
    transparent 60%,
    transparent 100%
  );
  background-size: 200% 100%;
  background-position: 200% 0;
  animation: btn-shimmer 4s linear infinite;
  pointer-events: none;
}

@keyframes btn-breathing {
  /* 从冷紫改中性暖白：与新 aurora 的“冷月光+暖琥珀”调同，
     不再出现“紫色 AI splash 调”。 */
  0%, 100% {
    box-shadow:
      0 0 28px rgba(245, 240, 230, 0.18),
      0 0 60px rgba(245, 240, 230, 0.08),
      0 6px 18px rgba(0, 0, 0, 0.25);
  }
  50% {
    box-shadow:
      0 0 44px rgba(255, 250, 240, 0.30),
      0 0 90px rgba(255, 250, 240, 0.14),
      0 6px 18px rgba(0, 0, 0, 0.25);
  }
}

@keyframes btn-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -100% 0; }
}

.btn-p:hover {
  opacity: .95;
  transform: translateY(-1px);
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
  gap: 8px;
  font: 500 16px var(--sans);
  color: var(--t2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(8px);
  cursor: pointer;
  padding: 15px 30px;
  border-radius: var(--radius-md);
  text-decoration: none;
  transition: all .25s ease;
}

.btn-g:hover {
  color: var(--t);
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.06);
}

.hero-right {
  position: relative;
  height: 100%;
  display: flex;
  /* 由 center 改为 stretch：让唯一的 flex 子项 .demo-win 纵向拉满
     hero 行高（≥480px），不再让窗体居中后上下留出黑边。 */
  align-items: stretch;
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
  /* 画面 A：去除 16:9 比例锁，改为 height: 100% 贴满 hero 行高。
     min-height 与 .hero 的 grid-template-rows: minmax(480px, auto) 同步，
     避免左列内容极少时窗体塌缩。视觉上从扁宽视频窗变为约 6:5 的工作站窗，
     窗内 demo-msgs 呼吸更松，与底栏"实时追问中"的待命叙事呼应。
     小屏（≤1024px）下另行还原 aspect-ratio，见 media query 兜底。 */
  height: 100%;
  min-height: 480px;
  /* 哑光高级黑（MacBook Pro 太空黑质感）：
     - padding-box 改垂直渐变冷石墨黑：顶 #121318 → 中 #0b0c10 → 底 #07080b。
       从顶到底的极淡明暗梯度模拟“光从顶部漫射照射”，让平面有体积感。
     - border 顶高光从 0.22 降到 0.16，避免贼亮玻璃描边，更内敛哑光。 */
  background:
    linear-gradient(180deg,
      rgba(18, 19, 24, 1) 0%,
      rgba(11, 12, 16, 1) 50%,
      rgba(7, 8, 11, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.16) 0%,
      rgba(255, 255, 255, 0.04) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.08) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  backdrop-filter: blur(20px);
  display: flex;
  flex-direction: column;
  /* 哑光质感的 shadow 公式：
     - inset 1px 顶柔光（0.06 alpha）：替代镕利镜面 1px 白线
     - inset 80px 顶部漫射（0.03 alpha）：让顶反光“散开”，不是单一锐线
     - 主投影 0.5：卡片“重”
     - 外阴影改黑色（不再是白辉光）：强化“低光环境”感 */
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    inset 0 60px 80px -60px rgba(255, 255, 255, 0.03),
    0 30px 60px rgba(0, 0, 0, 0.5),
    0 0 100px rgba(0, 0, 0, 0.2);
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
  max-width: calc(100% - 80px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.demo-foot-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4ade80;
  animation: edot 2s ease-in-out infinite;
}

/* === 4 步流程区块：如何工作 ===
   位于 hero 之后、metrics 之前，承担“如何工作”叙事。
   标题居中 + 4 张哑光黑卡片 + 3 个连接箭头。 */
.steps-outer {
  position: relative;
  z-index: 1;
  padding: 80px 44px 60px;
  max-width: 1320px;
  margin: 0 auto;
}

.steps-header {
  text-align: center;
  margin-bottom: 56px;
}

.steps-title {
  font-size: clamp(32px, 3vw, 44px);
  font-weight: 800;
  font-family: var(--display);
  color: var(--t);
  letter-spacing: -.02em;
  line-height: 1.2;
  margin: 0;
}

.steps {
  display: flex;
  align-items: stretch;
  justify-content: center;
  gap: 16px;
}

.step-card {
  /* 画面 S1：flex: 1 1 220px 让 4 张卡片响应式撑满 .steps 内容区（与 hero 同
     max-width: 1320 + padding 44 约束），左右缘自动对齐 hero-title / demo-win。
     基础 220px，小屏（viewport < 1408）自动收缩避免溢出引发横向滚动；
     大屏（viewport ≥ 1408）max-width 1320 触发后单卡片自然扩展到 ~266px。
     min-height 340 保住"纵向偏长"长方形状，不让卡片在大屏被压扁。 */
  flex: 1 1 220px;
  min-height: 340px;
  /* 同 demo-win 哑光黑卡质感（缩小版）。 */
  background:
    linear-gradient(180deg,
      rgba(18, 19, 24, 1) 0%,
      rgba(11, 12, 16, 1) 50%,
      rgba(7, 8, 11, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.14) 0%,
      rgba(255, 255, 255, 0.03) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.07) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  padding: 28px 22px 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 18px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    inset 0 50px 60px -50px rgba(255, 255, 255, 0.025),
    0 16px 40px rgba(0, 0, 0, 0.35);
  /* 悬浮上抬 + 阴影加深：提供“可探索”的视觉暗示。 */
  transition: transform .35s ease, box-shadow .35s ease;
}

.step-card:hover {
  transform: translateY(-6px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    inset 0 50px 60px -50px rgba(255, 255, 255, 0.04),
    0 24px 60px rgba(0, 0, 0, 0.5);
}

.step-num {
  /* 参考图：序号靠在卡片左上角，暖琥珀色（与右下 aurora-b 同调）。
     不是全白 0.5 alpha 。 */
  align-self: flex-start;
  font: 600 14px var(--mono);
  color: rgba(220, 155, 90, 0.9);
  letter-spacing: .08em;
}

.step-icon {
  /* 图标变大到1:1 容器 80px，以及 SVG fill 到容器宽高。
     参考图图标明显比之前大，是卡片中的视觉重心之一。 */
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--t);
  opacity: 0.92;
  margin-top: 8px;
}

/* SVG 自适应容器 + 细化 stroke（参考图图标 stroke 偏细）。
   原 SVG attribute stroke-width="1.4" 被 CSS 覆盖为 1.0，
   结合 80/44 宽度缩放，最终视觉 stroke 约 1.8px。 */
.step-icon svg {
  width: 100%;
  height: 100%;
  stroke-width: 1;
}

.step-title {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--display);
  color: var(--t);
  margin: 0;
  letter-spacing: -.01em;
}

.step-desc {
  font-size: 13px;
  line-height: 1.75;
  color: var(--t2);
  margin: 0;
}

.step-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  /* 箭头颜色从 0.18 提到 0.35，参考图箭头是可辨识的但不抢镜的中调灰。 */
  color: rgba(255, 255, 255, 0.35);
  flex-shrink: 0;
}

.metrics-outer {
  position: relative;
  z-index: 1;
  /* 画面 Y：padding 从 0 44px 改为 0 11px，让 metrics 容器整体外扩 66px。
     配合下方 .metrics max-width: 1386px，使首尾 met 的内容左右缘
     视觉对齐 hero 容器的内容左右缘——容器实际宽 1386 比 hero 1320 宽 66px，
     但 met-n 内容左缘与 hero-title 左缘重合，响应式下任意 viewport 都生效。 */
  padding: 0 11px;
  /* steps 之后与 metrics 之间需要间距。 */
  margin-top: 40px;
  margin-bottom: 100px;
}

.metrics {
  /* 1386 = 1320（hero max-width） + 66（= 33 × 2）。
     33px = 1px metrics 卡片 border + 32px 第一个 .met 的 padding-left。
     通过容器外扩补偿首尾 met 的内边距，使 met-n / met-l 的视觉左右缘
     与 hero 内容（hero-title 左缘 / demo-win 右缘）的视觉左右缘对齐。 */
  max-width: 1386px;
  margin: 0 auto;
  display: flex;
  /* 改 column：内部分为 metrics-row（4 列指标）+ metrics-cta（CTA 行）。 */
  flex-direction: column;
  /* 哑光高级黑（同 demo-win 样质感），垂直渐变冷石墨黑。 */
  background:
    linear-gradient(180deg,
      rgba(16, 17, 22, 1) 0%,
      rgba(10, 11, 14, 1) 50%,
      rgba(6, 7, 10, 1) 100%
    ) padding-box,
    linear-gradient(160deg,
      rgba(255, 255, 255, 0.14) 0%,
      rgba(255, 255, 255, 0.03) 30%,
      rgba(255, 255, 255, 0.02) 70%,
      rgba(255, 255, 255, 0.07) 100%
    ) border-box;
  border: 1px solid transparent;
  border-radius: 24px;
  overflow: hidden;
  backdrop-filter: blur(12px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    inset 0 60px 80px -60px rgba(255, 255, 255, 0.025),
    0 24px 50px rgba(0, 0, 0, 0.45),
    0 0 80px rgba(0, 0, 0, 0.15);
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
  padding: 36px 0 36px 40px;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.met:first-child { padding-left: 32px; }
/* :last-child 加 padding-right: 32px 与 :first-child 的 padding-left: 32px 对称，
   补偿 metrics 容器右侧外扩的 33px（= 1px metrics border + 32px 此 padding），
   让最后一个 met 的内容右缘视觉对齐 demo-win 右缘。 */
.met:last-child { border-right: none; padding-right: 32px; }

.met-n {
  font: 700 40px var(--mono);
  color: var(--t);
  line-height: 1;
  letter-spacing: -.04em;
}

.met-l {
  font-size: 15px;
  color: var(--t2);
  margin-top: 4px;
}

/* metrics 内部顶层容器：4 列 met 横排（原 .metrics 直接是 row，
   现在 .metrics 改 column 后由这一层接管 row 布局）。 */
.metrics-row {
  display: flex;
}

/* metrics 内部底部 CTA 行：左侧文案（eyebrow + 标语）+ 右侧开始体验按钮。
   与 metrics-row 之间用 1px 极淡分隔线隔开。 */
.metrics-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
  /* padding 从 32px 40px 改为 32px 32px：让 metrics-cta 内容（左侧 eyebrow +
     标语、右侧"开始体验"按钮）的视觉左右缘与 hero 内容对齐。
     32 = metrics 容器外扩 33px - 1px border，与 .met 首尾 padding 32 同源补偿。 */
  padding: 32px 32px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.metrics-cta-l {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  font: 13px var(--mono);
  color: var(--t2);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-pill);
  padding: 6px 14px;
  letter-spacing: .04em;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(8px);
}

/* CTA 行内复用的小元素：eyebrow 提示行 + 大字标语。
   原独立的 .bottom-cta 已合并进 metrics-cta，这两个 class
   现在只在 metrics-cta 内部使用。字号从 clamp(24,2.6vw,36) 缩到
   clamp(22,2.4vw,32)，因为嵌在 metrics 内部不应过大。 */
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
  font-size: clamp(22px, 2.4vw, 32px);
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
  .metrics-cta {
    flex-direction: column;
    align-items: stretch;
    gap: 20px;
    padding: 24px;
  }
  .metrics-cta-l {
    align-items: center;
    text-align: center;
  }
  .steps {
    flex-direction: column;
    gap: 16px;
  }
  .step-arrow {
    transform: rotate(90deg);
    align-self: center;
  }
  .steps-outer {
    padding: 60px 24px 40px;
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
  /* 单列时 .hero 的 grid-template-rows 改为 auto，.demo-win 失去父级高度参考。
     还原 aspect-ratio 让窗体按 16:9 自适应宽度，避免 height:100% 在 auto 行下塌缩。 */
  .demo-win {
    height: auto;
    min-height: 0;
    aspect-ratio: 16 / 9;
  }
}
</style>
